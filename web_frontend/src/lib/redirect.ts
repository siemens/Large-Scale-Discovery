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

import ToastEventBus from 'primevue/toasteventbus'
import { watch } from 'vue'
import { createI18n } from 'vue-i18n'

import { getAccessToken, hasValidSession } from '@/api/session'
import { TOAST_LIFE } from '@/composables/useToast'

// Local English defaults keep OAuth startup feedback independent from the global language catalog
const localMessages = {
  'en-US': {
    unauthorized: 'Unauthorized.',
    temporary: 'Component temporarily not available.',
    unexpected: 'Unexpected Error.',
  },
}

// Translate module-owned messages without importing the global catalog
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

// Store the protected destination under one stable session-storage key
const REDIRECT_STORAGE_KEY = 'redirect'

// AuthenticationFailure lists errors accepted from an OAuth callback
type AuthenticationFailure = 'Unauthorized' | 'Temporary' | 'Unexpected'

/** AuthRedirectFlow owns the one-shot SSO result and pending redirect lifecycle. */
interface AuthRedirectFlow {
  observeAuthentication: (getIsAuthenticated: () => boolean) => void
  handleUnauthenticatedRestore: () => void
  restorePendingRedirect: () => Promise<void>
  showFailure: () => void
}

/** Minimal router contract needed to restore an authenticated destination. */
interface AuthRedirectRouter {
  readonly currentRoute: { readonly value: { readonly fullPath: string } }
  isReady: () => Promise<void>
  replace: (target: string) => Promise<unknown>
}

/** Remember the protected route that authentication should return to. */
export function savePendingRedirect(target: string): void {
  try {
    sessionStorage.setItem(REDIRECT_STORAGE_KEY, target)
  } catch {
    // Session storage can be unavailable in locked-down browsers
  }
}

/** Remove a redirect that must no longer survive the current authentication attempt. */
export function clearPendingRedirect(): void {
  try {
    sessionStorage.removeItem(REDIRECT_STORAGE_KEY)
  } catch {
    // Session storage can be unavailable in locked-down browsers
  }
}

/** Consume a router-owned, same-origin path without allowing an external redirect. */
export function consumePendingRedirect(): string | null {
  // Read the requested return path from the URL
  let rawTarget: string | null = null
  try {
    rawTarget = sessionStorage.getItem(REDIRECT_STORAGE_KEY)
    sessionStorage.removeItem(REDIRECT_STORAGE_KEY)
  } catch {
    return null
  }

  if (!rawTarget || !rawTarget.startsWith('/') || rawTarget.startsWith('//')) return null

  try {
    // Resolve the stored path and reject any origin-changing normalization
    const target = new URL(rawTarget, window.location.origin)
    if (
      target.origin !== window.location.origin ||
      target.pathname === '/login' ||
      target.pathname === '/register'
    )
      return null
    return `${target.pathname}${target.search}${target.hash}`
  } catch {
    return null
  }
}

/** createAuthRedirectFlow captures callback state before the router can consume its query string. */
export function createAuthRedirectFlow(router: AuthRedirectRouter): AuthRedirectFlow {
  // Capture the session and callback result once so neither can be replayed by later navigation
  const sessionWasPresent = getAccessToken() !== null

  // Capture an authentication error encoded in the callback URL
  const failure = hasValidSession() ? null : consumeOAuthFailure()
  if (failure) clearPendingRedirect()

  // Explicit logout must discard a redirect saved before an abandoned OAuth handoff
  function observeAuthentication(getIsAuthenticated: () => boolean): void {
    watch(getIsAuthenticated, (isAuthenticated) => {
      if (!isAuthenticated) clearPendingRedirect()
    })
  }

  // A rejected persisted session cannot retain an earlier redirect target
  function handleUnauthenticatedRestore(): void {
    if (sessionWasPresent) clearPendingRedirect()
  }

  // Restore the validated same-origin target only after authentication succeeds
  async function restorePendingRedirect(): Promise<void> {
    // Consume once so reloads cannot replay a completed authentication redirect
    const target = consumePendingRedirect()
    if (!target) return

    await router.isReady()
    if (router.currentRoute.value.fullPath !== target) await router.replace(target)
  }

  // Toast only after the application mounted its global toast host
  function showFailure(): void {
    if (failure) emitOAuthFailure(failure)
  }

  return {
    observeAuthentication,
    handleUnauthenticatedRestore,
    restorePendingRedirect,
    showFailure,
  }
}

/** Read and remove the backend OAuth callback error before the router drops its query. */
function consumeOAuthFailure(): AuthenticationFailure | null {
  // Read the callback result before the router consumes its query string
  const url = new URL(window.location.href)

  // Ignore ordinary application URLs without an OAuth error
  const error = url.searchParams.get('error')
  if (!error) return null

  // Remove the complete query so navigation and reloads cannot replay the failure toast
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.hash}`)

  if (error === 'Unauthorized' || error === 'Temporary') return error
  return 'Unexpected'
}

/** Emit the captured OAuth failure with its matching severity icon. */
function emitOAuthFailure(failure: AuthenticationFailure): void {
  // Translate the captured failure without importing the global language catalog
  const { t } = localI18n.global

  // Map the backend result to its established user-facing message and icon
  const message =
    failure === 'Unauthorized'
      ? { summary: t('unauthorized'), icon: 'pi pi-lock' }
      : failure === 'Temporary'
        ? { summary: t('temporary'), icon: 'pi pi-sitemap' }
        : { summary: t('unexpected'), icon: 'pi pi-exclamation-triangle' }

  ToastEventBus.emit('add', { severity: 'error', life: TOAST_LIFE.error, ...message })
}
