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

import type { Ref } from 'vue'
import { onScopeDispose, ref } from 'vue'

/** Configuration for {@link usePolling}. */
export interface UsePollingOptions {
  /** Poll period in ms. */
  intervalMs: number
  /** Run `fn` immediately on start. Default true. */
  immediate?: boolean
  /** Start polling right away. Default true. */
  autoStart?: boolean
  /** Pause polling while the document is hidden. Default true. */
  pauseWhenHidden?: boolean
}

/** Reactive polling state plus its lifecycle controls. */
export interface UsePollingReturn {
  /** True while the polling loop is active. */
  isPolling: Ref<boolean>
  /** Last error thrown by `fn`, if any. */
  error: Ref<unknown>
  /** Start the polling loop. */
  start: () => void
  /** Stop the polling loop. */
  stop: () => void
  /** Trigger one poll immediately without waiting for the next tick. */
  poll: () => Promise<void>
}

/**
 * Lifecycle-bound, pause-aware interval polling primitive.
 *
 * Every "live" view in the app polls the REST API on an interval rather than
 * using a server-push channel. This composable is the shared primitive replacing
 * ad-hoc `setInterval` calls: runs `fn` immediately (optional), then every
 * `intervalMs`; pauses automatically while the tab is hidden and resumes on
 * focus; tears down on scope dispose.
 */
export function usePolling(
  fn: () => void | Promise<void>,
  options: UsePollingOptions,
): UsePollingReturn {
  // Resolve defaults once so every lifecycle branch follows the same policy
  const { intervalMs, immediate = true, autoStart = true, pauseWhenHidden = true } = options

  // Distinguish an enabled poller from one temporarily paused with the document
  const isPolling = ref(false)

  // Preserve the last failure for consumers that choose to render it
  const error = ref<unknown>(null)

  // Keep one interval and prevent overlapping requests
  let timer: ReturnType<typeof setInterval> | null = null
  let isRequestRunning = false

  // Run one request while skipping ticks that overlap a slow response
  async function poll(): Promise<void> {
    if (isRequestRunning) return
    isRequestRunning = true

    try {
      await fn()
      error.value = null
    } catch (pollError) {
      error.value = pollError
    } finally {
      isRequestRunning = false
    }
  }

  // Remove the current interval without changing the enabled state
  function clearTimer(): void {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
  }

  // Schedule ticks only while the poller is allowed to run in this document state
  function scheduleTimer(): void {
    clearTimer()
    if (pauseWhenHidden && document.hidden) return

    timer = setInterval(() => void poll(), intervalMs)
  }

  // Start polling unless it is already active
  function start(): void {
    if (isPolling.value) return
    isPolling.value = true

    // A hidden document starts paused and refreshes when it becomes visible
    if ((!pauseWhenHidden || !document.hidden) && immediate) void poll()
    scheduleTimer()
  }

  // Stop polling and cancel the pending timer
  function stop(): void {
    isPolling.value = false
    clearTimer()
  }

  // Pause hidden tabs and refresh immediately when they return
  function onVisibilityChange(): void {
    if (!pauseWhenHidden) return

    // Resume with fresh data instead of waiting one complete interval after reveal
    if (document.hidden) {
      clearTimer()
    } else if (isPolling.value && timer === null) {
      void poll()
      scheduleTimer()
    }
  }

  // Register the visibility listener and auto-start if configured
  if (pauseWhenHidden) {
    document.addEventListener('visibilitychange', onVisibilityChange)
  }
  if (autoStart) start()

  onScopeDispose(() => {
    stop()
    if (pauseWhenHidden) {
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  })

  // Expose the reactive state plus the lifecycle controls
  return { isPolling, error, start, stop, poll }
}
