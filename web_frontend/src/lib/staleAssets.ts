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

import type { RouteLocationNormalized, Router } from 'vue-router'

const RELOAD_STORAGE_KEY = 'app:stale-assets:last-reload'
const RELOAD_LOOP_WINDOW_MS = 10_000

interface StaleAssetReloadOptions {
  beforeReload?: () => void
}

interface VitePreloadErrorEvent extends Event {
  payload?: unknown
}

/** installStaleAssetReload reloads the SPA once when old hashed chunks disappeared after deployment. */
export function installStaleAssetReload(
  router: Router,
  options: StaleAssetReloadOptions = {},
): void {
  // Router lazy-load errors are the most visible stale-asset failure on link clicks
  router.onError((error, to) => {
    if (isStaleAssetError(error)) reloadFreshApplication(options, to)
  })

  // Vite emits a dedicated event for failed modulepreload or dynamic import fetches
  window.addEventListener('vite:preloadError', (event: Event) => {
    const preloadEvent = event as VitePreloadErrorEvent
    if (!isStaleAssetError(preloadEvent.payload)) return

    event.preventDefault()
    reloadFreshApplication(options)
  })

  // Some Vite preload failures surface only as rejected lazy-import promises
  window.addEventListener('unhandledrejection', (event: PromiseRejectionEvent) => {
    if (!isStaleAssetError(event.reason)) return

    event.preventDefault()
    reloadFreshApplication(options)
  })
}

function reloadFreshApplication(
  options: StaleAssetReloadOptions,
  target?: RouteLocationNormalized,
): void {
  // Avoid reload loops when the deployment is still inconsistent or cached badly
  if (hasReloadedRecently()) return

  // A stale bundle cannot safely keep local form edits, so clear app guards first
  options.beforeReload?.()

  // Store the reload moment before navigating away
  sessionStorage.setItem(RELOAD_STORAGE_KEY, Date.now().toString())

  // Router failures know the intended destination, so preserve the user's click
  if (target?.fullPath) {
    window.location.assign(target.fullPath)
    return
  }

  // Generic preload failures have no route target and can only refresh in place
  window.location.reload()
}

function hasReloadedRecently(): boolean {
  // Missing sessionStorage should not block the recovery reload
  const lastReload = Number(sessionStorage.getItem(RELOAD_STORAGE_KEY) ?? 0)

  // Treat invalid stored values as if no reload happened
  if (!Number.isFinite(lastReload)) return false

  // A second stale-asset error shortly after reload probably means deployment is incomplete
  return Date.now() - lastReload < RELOAD_LOOP_WINDOW_MS
}

function isStaleAssetError(error: unknown): boolean {
  // Normalize browser-specific chunk-load failures into one comparable string
  const message = errorMessage(error)

  // Match Vite, native ESM, and common bundler wording for missing hashed chunks
  return (
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Importing a module script failed') ||
    message.includes('error loading dynamically imported module') ||
    message.includes('Unable to preload CSS') ||
    message.includes('ChunkLoadError')
  )
}

function errorMessage(error: unknown): string {
  // Native failures and Vite preload payloads usually arrive as Error instances
  if (error instanceof Error) return error.message

  // Some browsers surface preload failures as plain event detail strings
  if (typeof error === 'string') return error

  // Non-string payloads do not carry useful stale-asset wording
  return ''
}
