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

/**
 * Uniform error model shared by the three frontend applications.
 *
 * Every failure that leaves `http.ts` — network failure, HTTP error status, or
 * an application-level error envelope (`{ error: true, message }` returned with
 * HTTP 200 by the legacy `web_backend`) — is normalized into a single
 * `NormalizedError` shape so pages render errors uniformly and never
 * have to branch on Axios internals.
 */

/** Discriminates how a request failed, so callers can react without parsing strings. */
export type ApiErrorKind =
  | 'network' // request never reached the server (offline, CORS, DNS, timeout)
  | 'auth' // 401 — session invalid/expired
  | 'forbidden' // 403 — authenticated but not allowed
  | 'not_found' // 404
  | 'rate_limited' // 429
  | 'unavailable' // 503 — component temporarily down
  | 'server' // other 5xx
  | 'app' // HTTP 200 but `{ error: true }` envelope
  | 'client' // other 4xx
  | 'unknown' // anything that does not fit the above

/** NormalizedError is the stable failure shape produced by the HTTP layer. */
export interface NormalizedError {
  /** Always true — lets callers type-guard with `isNormalizedError`. */
  readonly normalized: true
  /** Coarse machine-readable category. */
  readonly kind: ApiErrorKind
  /** HTTP status code, or 0 when the request never completed. */
  readonly status: number
  /** Human-readable, already-localized-enough message safe to surface in a toast. */
  readonly message: string
  /** The parsed response body, when one was returned (the legacy envelope or raw). */
  readonly body: unknown
  /** The HTTP method + path that failed, for logging/telemetry. */
  readonly request: { method: string; url: string }
}

/** HttpError is the concrete normalized error rejected by the shared HTTP client. */
export class HttpError extends Error implements NormalizedError {
  readonly normalized = true

  constructor(
    readonly kind: ApiErrorKind,
    readonly status: number,
    message: string,
    readonly body: unknown,
    readonly request: { method: string; url: string },
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

/** errorDetail extracts optional technical detail without substituting a fallback message. */
export function errorDetail(error: unknown): string | undefined {
  if (isNormalizedError(error)) return error.message
  if (error instanceof Error) return error.message
  if (error && typeof error === 'object' && 'message' in error) {
    const message = error.message
    if (typeof message === 'string' && message) return message
  }
  return undefined
}

/** errorMessage extracts a human-readable message, falling back to a caller-supplied or generic string. */
export function errorMessage(error: unknown, fallback = 'Unexpected error'): string {
  return errorDetail(error) ?? fallback
}

/** isNormalizedError identifies the stable failure shape produced by the HTTP layer. */
export function isNormalizedError(value: unknown): value is NormalizedError {
  // Only our normalized errors carry the literal `normalized: true` marker
  return (
    typeof value === 'object' &&
    value !== null &&
    'normalized' in value &&
    value.normalized === true
  )
}

/** Map an HTTP status code to a coarse {@link ApiErrorKind}. */
export function kindForStatus(status: number): ApiErrorKind {
  if (status === 0) return 'network'
  if (status === 401) return 'auth'
  if (status === 403) return 'forbidden'
  if (status === 404) return 'not_found'
  if (status === 429) return 'rate_limited'
  if (status === 503) return 'unavailable'
  if (status >= 500) return 'server'
  if (status >= 400) return 'client'
  return 'unknown'
}
