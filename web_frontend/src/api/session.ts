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
 * Session/token bridge.
 *
 * Single source of truth for the JWT access token, shared between `http.ts`
 * (which attaches it and refreshes it on every response) and `useAuthStore`
 * (which drives login/logout). Keeping it in a standalone module breaks the
 * store ↔ http circular import and matches the legacy `sessionStorage("token")`
 * persistence so a page reload keeps the user signed in.
 *
 * Legacy parity: the backend re-issues a fresh `{ access_token, expire }` on
 * (almost) every successful response — a sliding window — which we persist via
 * {@link setToken}. A 401 anywhere clears frontend-owned authentication state
 * and fires {@link onUnauthorized} before returning to login.
 */

/** AuthToken is the JWT access token plus its expiry, as issued by the backend. */
export interface AuthToken {
  access_token: string
  /** RFC3339 timestamp string from the backend. */
  expire: string
}

// Persist the token under the key shared by the authentication flow
const SESSION_STORAGE_KEY = 'token'

// Cap timers at the largest delay supported reliably by browsers
const MAX_TIMER_DELAY_MS = 2_147_483_647

// In-memory session state shared across the app for the current tab
let token: AuthToken | null = null

// Store the callback that signs the application out after a 401 response
let unauthorizedHandler: (() => void) | null = null

// Prevent parallel 401 responses from producing duplicate redirects and toasts
let isUnauthorizedTriggered = false

// Handle for the passive-expiry timer. Cleared when the session is torn down.
let expiryTimer: ReturnType<typeof setTimeout> | null = null

// Remember whether the document visibility listener is installed
let isVisibilityListenerRegistered = false

// Recheck the session when the browser tab becomes visible
function onVisibilityChange(): void {
  if (document.visibilityState === 'visible' && token !== null && !hasValidSession()) {
    triggerUnauthorized()
  }
}

// Track sliding-window refreshes so the latest token expiry owns the redirect timer
function scheduleExpiry(): void {
  if (expiryTimer !== null) {
    clearTimeout(expiryTimer)
    expiryTimer = null
  }
  if (!token?.expire) return

  // Calculate how long the current token remains valid
  const delay = new Date(token.expire).getTime() - Date.now()

  // Already expired tokens are handled by the router guard / first API 401 — no
  // need to fire immediately here (avoid a synchronous redirect on loadSession).
  if (delay <= 0) return

  // Timer delays are signed 32-bit integers in browsers. Schedule long-lived
  // sessions in safe chunks and re-evaluate when the current chunk elapses.
  expiryTimer = setTimeout(
    () => {
      expiryTimer = null
      if (hasValidSession()) {
        scheduleExpiry()
        return
      }
      triggerUnauthorized()
    },
    Math.min(delay, MAX_TIMER_DELAY_MS),
  )
}

/** loadSession hydrates the in-memory token from sessionStorage (call once at app start). */
export function loadSession(): void {
  // Validate persisted state before allowing it to become the active session
  try {
    const raw = sessionStorage.getItem(SESSION_STORAGE_KEY)
    const storedToken: unknown = raw ? JSON.parse(raw) : null
    token = isAuthToken(storedToken) ? storedToken : null
  } catch {
    token = null
  }

  // Arm the passive-expiry timer so a tab that stays open past the token deadline
  // redirects to login without waiting for user interaction.
  scheduleExpiry()

  // setTimeout does not survive OS sleep/hibernate — the JS engine is frozen along
  // with the tab, so the expiry timer silently elapses without firing. On visibility
  // restore, check whether the token expired while the tab was frozen and redirect
  // immediately if so. This is a no-op on normal focus changes where the token is
  // still valid.
  if (!isVisibilityListenerRegistered) {
    document.addEventListener('visibilitychange', onVisibilityChange)
    isVisibilityListenerRegistered = true
  }
}

/** getAccessToken returns the current JWT access token, or null when signed out. */
export function getAccessToken(): string | null {
  return token?.access_token ?? null
}

// getExpiry parses the token expiry into a Date, or null when there is no token
function getExpiry(): Date | null {
  return token?.expire ? new Date(token.expire) : null
}

/** hasValidSession is true when there is a token and it has not yet expired. */
export function hasValidSession(): boolean {
  // Require both a token and a future expiry to consider the session valid
  const expiry = getExpiry()
  return token !== null && expiry !== null && expiry.getTime() > Date.now()
}

/** setToken persists a (re-issued) token, used by login and the sliding-window refresh. */
export function setToken(next: AuthToken | null): void {
  // A fresh token means a fresh session — lift the terminated flag so new requests
  // are allowed through and reschedule the passive-expiry timer.
  if (next) {
    isUnauthorizedTriggered = false
  }

  // Update the in-memory token first so it takes effect even if storage fails
  token = next

  // Mirror the change into sessionStorage for reload persistence.
  // Swallow storage failures because the in-memory token still works.
  try {
    if (next) {
      sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(next))
    } else {
      sessionStorage.removeItem(SESSION_STORAGE_KEY)
    }
  } catch {
    // Keep the in-memory token when storage is unavailable
  }

  // Arm (or re-arm after sliding-window refresh) the passive-expiry timer
  scheduleExpiry()
}

/** clearSession removes the frontend-owned authentication state. */
export function clearSession(): void {
  // Cancel the passive-expiry timer — no token means no expiry to schedule
  if (expiryTimer !== null) {
    clearTimeout(expiryTimer)
    expiryTimer = null
  }

  // Clear in-memory state first, then best-effort clear persisted keys
  token = null
  try {
    sessionStorage.removeItem(SESSION_STORAGE_KEY)
    sessionStorage.removeItem('user')
  } catch {
    // In-memory state is already clear when storage is unavailable
  }
}

/**
 * onUnauthorized registers the handler invoked when the server rejects us with
 * 401 (or the session expires). The router wires this to a redirect-to-login so
 * `http.ts` stays free of router/store imports.
 */
export function onUnauthorized(handler: () => void): void {
  unauthorizedHandler = handler
}

/** triggerUnauthorized clears the session and notifies the registered handler. */
export function triggerUnauthorized(): void {
  // Deduplicate: parallel 401 responses or a racing expiry timer could each call
  // this. Only the first invocation triggers the redirect; the rest are no-ops.
  if (isUnauthorizedTriggered) return
  isUnauthorizedTriggered = true

  // Clear session state before notifying so the handler sees a signed-out state
  clearSession()
  unauthorizedHandler?.()
}

/** isAuthToken validates session-storage data before it reaches authentication state. */
function isAuthToken(value: unknown): value is AuthToken {
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
