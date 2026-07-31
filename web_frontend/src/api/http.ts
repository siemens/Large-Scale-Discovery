/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2025-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

import type { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig, } from 'axios'
import axios from 'axios'
import ToastEventBus from 'primevue/toasteventbus'
import { computed, ref } from 'vue'

import type { ApiErrorKind, NormalizedError } from './errors'
import { HttpError, kindForStatus } from './errors'
import { getAccessToken, setToken, triggerUnauthorized } from './session'
import { TOAST_LIFE } from '@/composables/useToast'

// Use the versioned same-origin API path when deployment provides no override
const DEFAULT_API_BASE_URL = '/api/v1'

/**
 * Resolves the optional deployment-specific API URL.
 *
 * Browser deployments may use an absolute override only when it has the same
 * origin as the application. This prevents credentials from being attached to
 * a backend selected by injected or accidentally unsafe runtime configuration.
 */
export function resolveApiBaseUrl(
  configuredValue: string | undefined,
  browserOrigin?: string,
): string {
  // Fall back to the same-origin API path when no override exists
  const value = configuredValue?.trim() || DEFAULT_API_BASE_URL

  // Protocol-relative URLs are external-capable and must not pass as root-relative paths
  if (value.startsWith('//')) {
    throw new Error('VITE_API_BASE_URL must be an absolute HTTP(S) URL or a root-relative path')
  }

  // Root-relative paths are always resolved by the current application origin
  if (value.startsWith('/')) return value.replace(/\/+$/u, '') || '/'

  // Parse absolute values before enforcing their protocol and browser origin
  let parsed: URL
  try {
    parsed = new URL(value)
  } catch {
    throw new Error('VITE_API_BASE_URL must be an absolute HTTP(S) URL or a root-relative path')
  }

  if (parsed.protocol !== 'https:' && parsed.protocol !== 'http:') {
    throw new Error('VITE_API_BASE_URL must be an absolute HTTP(S) URL or a root-relative path')
  }
  if (parsed.username !== '' || parsed.password !== '') {
    throw new Error('VITE_API_BASE_URL must not contain URL credentials')
  }
  if (browserOrigin !== undefined && parsed.origin !== new URL(browserOrigin).origin) {
    throw new Error('VITE_API_BASE_URL must be same-origin with the browser application')
  }

  return value.replace(/\/+$/u, '')
}

/**
 * Reads the configured API base URL for the shared Axios instance.
 */
function apiBaseUrl(): string {
  // Supply the browser origin only when the module runs in a browser environment
  const browserOrigin = typeof window === 'undefined' ? undefined : window.location.origin
  return resolveApiBaseUrl(import.meta.env.VITE_API_BASE_URL, browserOrigin)
}

/**
 * Shared Axios HTTP client used by all three frontend applications.
 *
 * This module creates and exports the single Axios instance used by every API
 * call in the application. No component should create its own Axios instance or
 * use raw `fetch` — all requests must go through this client so the interceptor
 * pipeline is applied consistently.
 *
 * ## Interceptor pipeline (applied in order)
 *
 * | #  | Phase    | What it does                                                  |
 * |----|----------|---------------------------------------------------------------|
 * | 1  | Request  | Increments the global loading bar counter (skippable).        |
 * | 2  | Request  | Attaches the Bearer token and CSRF cookie to every request.   |
 * | 3  | Response | Refreshes the session token (sliding-window) on success.      |
 * |    |          | Detects legacy `{ error: true }` envelopes and rejects them.  |
 * |    | Error    | Retries failed idempotent requests with exponential backoff.  |
 * |    |          | Triggers logout on 401. Normalizes all errors uniformly.      |
 *
 * ## Usage
 *
 * ```ts
 * import http from '@/api/http'
 *
 * // Standard GET — loading bar shown, auth attached, errors normalized.
 * const res = await http.get<{ body: SomeType }>('/some/endpoint')
 *
 * // Suppress the loading bar (e.g. background polling).
 * const res = await http.get('/status', { suppressLoader: true })
 * ```
 */

// ─── Custom Axios config properties ─────────────────────────────────────────
//
// These are injected into Axios's type system so callers can pass them in the
// request config object. TypeScript's module augmentation makes this safe:
// any `http.get('/url', { suppressLoader: true })` call is fully typed.

declare module 'axios' {
  // Extend Axios requests with loading-bar bookkeeping used by the interceptors
  interface AxiosRequestConfig {
    /** When true, this request will not trigger the global loading bar. */
    suppressLoader?: boolean

    /** When true, callers own the request's user-facing error feedback. */
    suppressErrorToast?: boolean

    /** Localized message that replaces the normalized error in the automatic toast. */
    errorToastMessage?: string

    /**
     * Internal flag set by the loading-bar interceptor to track whether
     * `startLoading()` was called for this request. Used to ensure the
     * matching `stopLoading()` is called exactly once.
     */
    loaderCounted?: boolean

    /** Internal retry count maintained by the response interceptor. */
    retryCount?: number
  }
}

// ─── Types ──────────────────────────────────────────────────────────────────

/**
 * Shape of the legacy backend response envelope.
 *
 * The `web_backend` wraps every response in `{ body, token }`. When the
 * business logic encounters an error, it returns HTTP 200 with
 * `{ error: true, message: "..." }` instead of an HTTP error status. This
 * interceptor chain detects that pattern and rejects it as an error.
 */
interface BackendEnvelope {
  error?: boolean
  message?: string
  body?: unknown
  token?: unknown
}

/** RefreshedToken is the validated session grant accepted from a response envelope. */
interface RefreshedToken {
  access_token: string
  expire: string
}

// ─── Retry policy constants ─────────────────────────────────────────────────

/** Maximum number of automatic retry attempts (total tries = 1 + MAX_RETRIES). */
const MAX_RETRIES = 2

/** Starting backoff delay in milliseconds. Doubles with each attempt: 300 → 600. */
const BASE_BACKOFF_MS = 300

/** HTTP methods that are safe to retry (all idempotent per RFC 7231 §4.2.2). */
const RETRYABLE_METHODS = new Set(['get', 'head', 'options', 'put', 'delete'])

/** Default request timeout in milliseconds. */
const HTTP_DEFAULT_TIMEOUT_MS = 30_000

// Remember the last failure per request so background polling cannot repeat the same toast
const lastErrorToastByRequest = new Map<string, string>()

/** ERROR_ICON_BY_KIND maps infrastructure failures to their established visual cue. */
const ERROR_ICON_BY_KIND: Partial<Record<ApiErrorKind, string>> = {
  network: 'pi pi-wifi',
  rate_limited: 'pi pi-ban',
  server: 'pi pi-times-circle',
  unavailable: 'pi pi-sitemap',
}

// ─── Error message providers ─────────────────────────────────────────────────
//
// Error builders use stable local English defaults so this low-level API module
// stays independent from component and application localization state.

interface HttpErrorMessages {
  requestFailed(): string
  network(): string
  unauthorized(): string
  rateLimited(): string
  unavailable(): string
  http(params: { status: number; statusText: string }): string
}

// Define stable error messages for this low-level HTTP module
const httpErrorMessages: HttpErrorMessages = {
  requestFailed: () => 'Request failed.',
  network: () => 'Service currently not reachable.',
  unauthorized: () => 'You are not authorized to execute this action.',
  rateLimited: () => 'Too many requests. Please try again later.',
  unavailable: () => 'Component temporarily not available.',
  http: ({ status, statusText }) => `Unexpected Error: ${status}${statusText}.`,
}

// ─── Create the Axios instance ──────────────────────────────────────────────

const http: AxiosInstance = axios.create({
  baseURL: apiBaseUrl(),
  headers: { 'Content-Type': 'application/json' },
  timeout: HTTP_DEFAULT_TIMEOUT_MS,

  // Cookies are sent alongside the Bearer token. The backend currently relies
  // on Bearer only, but `withCredentials: true` ensures a future switch to
  // session cookies or CSRF-via-cookie works without a frontend change.
  withCredentials: true,
})

// ─── Loading bar state ──────────────────────────────────────────────────────
//
// Tracks the number of in-flight requests. The `isLoading` computed is exported
// for `LoadingBar.vue` to bind against. The interceptors below increment and
// decrement the counter automatically.

const activeRequestCount = ref(0)

/** Reactive flag: true while at least one HTTP request is in flight. */
export const isLoading = computed(() => activeRequestCount.value > 0)

// Count one more active request
export function startLoading(): void {
  activeRequestCount.value++
}

// Remove one completed request from the active count
export function stopLoading(): void {
  activeRequestCount.value = Math.max(0, activeRequestCount.value - 1)
}

// ─── Interceptor 1: Loading bar (request) ───────────────────────────────────
//
// Increments the global in-flight request counter, which drives the loading
// bar visibility in App.vue. Callers can opt out with `{ suppressLoader: true }`
// — useful for background polling that shouldn't flash the loading indicator.

http.interceptors.request.use((config) => {
  // Mark this request as counted so the matching stopLoading() fires exactly once
  if (!config.suppressLoader && !config.loaderCounted) {
    config.loaderCounted = true
    startLoading()
  }
  return config
})

// ─── Interceptor 2: Auth & CSRF (request) ───────────────────────────────────
//
// Attaches the session token and CSRF cookie to every outgoing request:
//   - Authorization: Bearer <token>  — primary auth mechanism
//   - X-XSRF-TOKEN: <cookie-value>   — CSRF double-submit (future-proofing)

http.interceptors.request.use((config) => {
  // Attach the JWT access token from the session bridge.
  const accessToken = getAccessToken()
  if (accessToken) {
    config.headers.set('Authorization', `Bearer ${accessToken}`)
  }

  // Mirror the XSRF-TOKEN cookie as a header (double-submit pattern).
  // This is a no-op today — the backend doesn't set the cookie — but it's
  // wired so enabling CSRF protection server-side needs zero frontend work.
  const csrfToken = readCookie('XSRF-TOKEN')
  if (csrfToken) {
    config.headers.set('X-XSRF-TOKEN', csrfToken)
  }

  return config
})

// ─── Interceptor 3: Response handling ───────────────────────────────────────
//
// Success path: refresh the session token, detect envelope errors, stop loader.
// Error path: retry if eligible, handle 401, normalize all errors.
//
// The loading bar is decremented here (not in a separate interceptor) because
// both exit paths replace the Axios error with a plain NormalizedError that
// has no `config` — a downstream interceptor would never see `loaderCounted`.

http.interceptors.response.use(
  // ── Success handler ───────────────────────────────────────────────────────
  (response: AxiosResponse<BackendEnvelope>) => {
    logBackendResponse(response)
    refreshSessionToken(response)

    // The backend sometimes signals errors via HTTP 200 with `{ error: true }`.
    // Reject these as errors so callers don't have to check the envelope.
    if (response.data?.error === true) {
      if (response.config.loaderCounted) stopLoading()

      // Normalize and report application-level errors just like protocol failures
      const normalizedError = buildNormalizedError({
        kind: 'app',
        status: response.status,
        message: response.data.message ?? httpErrorMessages.requestFailed(),
        body: response.data,
        config: response.config,
      })
      reportApiError(normalizedError, response.config)
      return Promise.reject(normalizedError)
    }

    // A successful response allows the same request to report a future failure again
    lastErrorToastByRequest.delete(requestToastKey(response.config))
    if (response.config.loaderCounted) stopLoading()
    return response
  },

  // ── Error handler ─────────────────────────────────────────────────────────
  async (error: AxiosError<BackendEnvelope>) => {
    // Keep retry and loading state on the original Axios request config
    const config = error.config

    // Retry eligible requests before giving up.
    if (config && canRetry(error, config)) {
      config.retryCount = (config.retryCount ?? 0) + 1
      await sleep(calculateBackoff(config.retryCount, error))
      return http(config)
    }

    logProtocolError(error)

    if (config?.loaderCounted) stopLoading()

    // A 401 normally means the active session is invalid: clear it and redirect
    // to login. A rejected login attempt has no session to invalidate, so let
    // that rejection reach the form and render its inline validation feedback.
    const status = error.response?.status ?? 0
    if (status === 401) {
      if (config?.url === '/auth/login') {
        // Let the localized login flow own feedback when it explicitly suppresses the client toast
        const normalizedError = buildNormalizedError({
          kind: 'auth',
          status,
          message: extractErrorMessage(error),
          body: error.response?.data,
          config: error.config,
        })
        reportApiError(normalizedError, config)
        return Promise.reject(normalizedError)
      }

      // Preserve the legacy unauthorized toast before invalidating the browser session
      const normalizedError = buildNormalizedError({
        kind: 'auth',
        status,
        message: extractErrorMessage(error),
        body: error.response?.data,
        config: error.config,
      })
      reportApiError(normalizedError, config)
      triggerUnauthorized()
      return new Promise(() => {
        /* intentionally never resolves */
      })
    }

    // Normalize and report the terminal failure after all eligible retries are exhausted
    const normalizedError = buildNormalizedError({
      kind: kindForStatus(status),
      status,
      message: extractErrorMessage(error),
      body: error.response?.data,
      config: error.config,
    })
    reportApiError(normalizedError, config)
    return Promise.reject(normalizedError)
  },
)

// ─── Error feedback ─────────────────────────────────────────────────────────

/** reportApiError emits one legacy-style toast for a terminal API failure. */
function reportApiError(error: HttpError, config?: AxiosRequestConfig): void {
  // Leave specialized localized flows in control when they explicitly opt out
  if (config?.suppressErrorToast) return

  // Prefer a caller's contextual message over backend or protocol wording
  const message = config?.errorToastMessage ?? error.message

  // Suppress an unchanged polling failure until this request succeeds again
  const requestKey = requestToastKey(config)
  if (lastErrorToastByRequest.get(requestKey) === message) return
  lastErrorToastByRequest.set(requestKey, message)

  // Preserve the established icon semantics for infrastructure failures
  const icon = ERROR_ICON_BY_KIND[error.kind] ?? null

  ToastEventBus.emit('add', {
    severity: 'error',
    summary: message,
    life: TOAST_LIFE.error,
    icon,
  })
}

/** requestToastKey identifies one backend operation for repeated-error suppression. */
function requestToastKey(config?: { method?: string; url?: string }): string {
  return `${config?.method?.toUpperCase() ?? 'GET'} ${config?.url ?? ''}`
}

// ─── Retry helpers ──────────────────────────────────────────────────────────

/**
 * Determines whether a failed request should be retried.
 *
 * A request is retryable when ALL of these are true:
 *   1. The retry count hasn't reached `MAX_RETRIES`.
 *   2. The HTTP method is idempotent (GET, HEAD, OPTIONS, PUT, DELETE).
 *   3. The failure is transient: network error, 429, 503, or any 5xx.
 */
function canRetry(error: AxiosError, config: InternalAxiosRequestConfig): boolean {
  // Short-circuit on retry cap and non-idempotent methods before inspecting the response
  const attempts = config.retryCount ?? 0
  if (attempts >= MAX_RETRIES) return false

  // Normalize the request method before checking whether retry is safe
  const method = (config.method ?? 'get').toLowerCase()
  if (!RETRYABLE_METHODS.has(method)) return false

  // No response at all = network failure — always retryable for safe methods.
  if (!error.response) return true

  // Read the HTTP status used to classify the failure
  const status = error.response.status
  return status === 429 || status === 503 || status >= 500
}

/**
 * Calculates the backoff delay for a retry attempt.
 *
 * Uses exponential backoff (BASE_BACKOFF_MS * 2^(attempt-1)):
 *   attempt 1 → 300ms, attempt 2 → 600ms
 *
 * If the server sends a `Retry-After` header (in seconds), that value takes
 * precedence over the calculated backoff.
 */
function calculateBackoff(attempt: number, error: AxiosError): number {
  // Prefer the server's own Retry-After directive over our exponential schedule
  const retryAfterHeader = error.response?.headers?.['retry-after']
  if (retryAfterHeader) {
    // Accept the common delay-seconds form first
    const seconds = Number(retryAfterHeader)
    if (Number.isFinite(seconds) && seconds >= 0) return seconds * 1_000

    // HTTP-date is the protocol's alternative Retry-After representation
    const retryAt = Date.parse(String(retryAfterHeader))
    if (Number.isFinite(retryAt)) return Math.max(0, retryAt - Date.now())
  }
  return BASE_BACKOFF_MS * 2 ** (attempt - 1)
}

// ─── Token refresh ──────────────────────────────────────────────────────────

/**
 * Persists the refreshed token from a successful response.
 *
 * The backend re-issues a JWT on most successful responses (sliding-window
 * pattern). We persist it so the session stays alive as long as the user is
 * active.
 *
 * Guard: if `getAccessToken()` returns null, the user has logged out while
 * this request was in flight. In that case, discard the refreshed token to
 * avoid silently re-authenticating a logged-out user.
 */
function refreshSessionToken(response: AxiosResponse<BackendEnvelope>): void {
  // Extract and validate the refreshed token from the response envelope
  const refreshedToken = response.data?.token
  if (!isRefreshedToken(refreshedToken)) return

  // Accept the token when it's the initial grant (login) OR a sliding-window
  // refresh of an active session.  The old `getAccessToken() !== null` guard
  // blocked the login response because no prior session existed yet.
  const isLoginGrant = response.config.url === '/auth/login'
  if (isLoginGrant || getAccessToken() !== null) {
    setToken({
      access_token: refreshedToken.access_token,
      expire: refreshedToken.expire,
    })
  }
}

// ─── Response logging ───────────────────────────────────────────────────────

/** logBackendResponse writes one backend response envelope to the browser console. */
function logBackendResponse(response: AxiosResponse<BackendEnvelope>): void {
  // Build the same readable request label used by the legacy ajax helper
  const label = requestLogLabel(response.config)

  // Keep backend-declared failures visually distinct in the console
  if (response.data?.error === true) {
    console.log(
      `REQUEST FAILED: ${response.data.message ?? httpErrorMessages.requestFailed()} ${label}`,
    )
  } else {
    console.log(`REQUEST SUCCESS ${label}`)
  }

  // Log only the backend response object on the next line, matching the old UI
  console.log(response.data)
}

/** logProtocolError writes the legacy protocol-error line for failed HTTP requests. */
function logProtocolError(error: AxiosError<BackendEnvelope>): void {
  // Build the same readable request label used by the legacy ajax helper
  const label = requestLogLabel(error.config)

  // Network failures have no HTTP response from the backend
  if (!error.response) {
    console.error(`PROTOCOL ERROR: Connection failed ${label}`)
    return
  }

  // HTTP failures are protocol errors in the legacy console vocabulary
  console.error(`PROTOCOL ERROR: ${error.response.status} ${error.response.statusText} ${label}`)
}

/** requestLogLabel formats the method and URL like the legacy ajax helper. */
function requestLogLabel(config?: { method?: string; url?: string }): string {
  // Normalize missing Axios metadata to the same practical defaults as a GET request
  const method = config?.method?.toUpperCase() ?? 'GET'
  const url = config?.url ?? ''

  return `(${method} ${url})`
}

// ─── Error message extraction ───────────────────────────────────────────────

/**
 * Builds a human-readable error message for a failed request.
 *
 * Priority:
 *   1. The backend's own message from the response envelope (`data.message`).
 *   2. A status-specific fallback for well-known codes (0, 401, 429, 503).
 *   3. "Unexpected Error: N StatusText." for all other HTTP errors — matching
 *      the legacy `apiCall` behavior in `web_frontend_old/src/js/utils-app.js`.
 */
function extractErrorMessage(error: AxiosError<BackendEnvelope>): string {
  // Prefer the backend's own error message — it's the most specific.
  const backendMessage = error.response?.data?.message
  if (backendMessage) return backendMessage

  // Read the HTTP status used to classify the failure
  const status = error.response?.status ?? 0

  if (status === 0) return httpErrorMessages.network()
  if (status === 401) return httpErrorMessages.unauthorized()
  if (status === 429) return httpErrorMessages.rateLimited()
  if (status === 503) return httpErrorMessages.unavailable()

  // All other HTTP errors: include the numeric code and status text so the
  // user and support staff can identify the exact failure without opening DevTools.
  const statusText = error.response?.statusText ?? ''
  return httpErrorMessages.http({ status, statusText: statusText ? ` ${statusText}` : '' })
}

// ─── Error normalization ────────────────────────────────────────────────────

/**
 * Input shape for building a {@link NormalizedError}.
 * Separates the construction concern from the Axios-specific error parsing.
 */
interface NormalizedErrorInput {
  kind: NormalizedError['kind']
  status: number
  message: string
  body: unknown
  config?: { method?: string; url?: string }
}

/**
 * Converts raw error data into the uniform {@link NormalizedError} shape.
 *
 * Every rejected promise from this HTTP client carries a `NormalizedError`,
 * so callers can handle errors with a single consistent interface — no need
 * to check for Axios internals, network failures, or envelope quirks.
 */
function buildNormalizedError(input: NormalizedErrorInput): HttpError {
  return new HttpError(input.kind, input.status, input.message, input.body, {
    method: (input.config?.method ?? 'get').toUpperCase(),
    url: input.config?.url ?? '',
  })
}

// ─── Utility functions ──────────────────────────────────────────────────────

/** Returns a promise that resolves after `ms` milliseconds. */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** Reads a cookie value by name from `document.cookie`. Returns null if not found. */
function readCookie(name: string): string | null {
  // Ignore absent or malformed cookie encoding instead of failing the whole request
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
  const encodedValue = match?.[1]
  if (encodedValue === undefined) return null

  try {
    return decodeURIComponent(encodedValue)
  } catch {
    return null
  }
}

/** isRefreshedToken validates token data before it reaches session state. */
function isRefreshedToken(value: unknown): value is RefreshedToken {
  return (
    typeof value === 'object' &&
    value !== null &&
    'access_token' in value &&
    typeof value.access_token === 'string' &&
    value.access_token.length > 0 &&
    'expire' in value &&
    typeof value.expire === 'string' &&
    Number.isFinite(Date.parse(value.expire))
  )
}

// ─── Export ─────────────────────────────────────────────────────────────────

export default http
