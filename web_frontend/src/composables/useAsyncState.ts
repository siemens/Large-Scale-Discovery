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
import { type Ref, ref, shallowRef } from 'vue'

import { errorMessage, isNormalizedError } from '@/api/errors'
import { TOAST_LIFE } from '@/composables/useToast'
import { sameJson } from '@/lib/collections'

import type { AxiosRequestConfig } from 'axios'

/** Options for a single {@link AsyncState.run} invocation. */
interface RunOptions {
  /** Fallback message when the caught error carries none. Forwarded to `errorMessage`. */
  fallback?: string
  /**
   * Re-throw the caught error after recording it (default `true`). Set `false`
   * for fire-and-forget loaders whose callers only read `error`/`isLoading`
   * (e.g. background polling on the home dashboard).
   */
  rethrow?: boolean
  /**
   * Follow-up to await on success, inside the same loading window. Typically a
   * re-fetch after a mutation (e.g. `() => fetchScopes()`). A rejection here is
   * treated like the main operation failing.
   */
  after?: () => Promise<void> | void
  /** Keep the operation out of the local loading state, for quiet background polling. */
  trackLoading?: boolean
  /** Leave non-HTTP failure feedback to a specialized caller. */
  suppressErrorToast?: boolean
}

/** The reactive loading/error pair plus the action runner returned by {@link useAsyncState}. */
export interface AsyncState {
  /** True while a `run` call is in flight. */
  readonly isLoading: Ref<boolean>
  /** Localized message of the last failure, or `null` when the last run succeeded. */
  readonly error: Ref<string | null>
  run: {
    <T>(fn: () => Promise<T>, opts?: RunOptions & { rethrow?: true }): Promise<T>
    <T>(fn: () => Promise<T>, opts: RunOptions & { rethrow: false }): Promise<T | undefined>
  }

  /** Reset the loading flag and clear the error (e.g. for a store's `reset()`). */
  reset(): void
}

/** ListFetcher contains stable list state backed by useAsyncState. */
export interface ListFetcher<T> {
  data: Ref<T | null>
  isFetching: Ref<boolean>
  fetchError: Ref<string | null>
  fetch: (config?: AxiosRequestConfig) => Promise<void>
  fetchSilent: () => Promise<void>
  reset: () => void
}

/** InitializedListFetcher contains non-null list state created by a caller-supplied factory. */
export interface InitializedListFetcher<T> extends Omit<ListFetcher<T>, 'data'> {
  data: Ref<T>
}

/** ListFetcherOptions customizes list initialization and reference-preserving reconciliation. */
export interface ListFetcherOptions<T> {
  initialData?: () => T
  reconcile?: (current: T | null, next: T) => T
}

/** InitializedListFetcherOptions guarantees that a list fetcher starts with non-null data. */
export interface InitializedListFetcherOptions<T> extends ListFetcherOptions<T> {
  initialData: () => T
}

/**
 * useAsyncState provides loading and error state for one async store operation.
 * Replaces the hand-rolled
 * `isX = true; xError = null; try { ... } catch { xError = errorMessage(err); throw } finally { isX = false }`
 * envelope that every store action repeated.
 *
 * Each operation gets its own `useAsyncState()`; expose its refs under the
 * store's established names so the public API is unchanged, e.g.:
 *
 * ```ts
 * const fetchState = useAsyncState()
 * async function fetchGroups() {
 *   return fetchState.run(async () => {
 *     const res = await http.get<...>('/groups')
 *     groups.value = res.data.body?.groups ?? []
 *   })
 * }
 * return { isFetching: fetchState.isLoading, fetchError: fetchState.error, fetchGroups }
 * ```
 */
export function useAsyncState(): AsyncState {
  // Expose whether at least one tracked action in the current generation is running
  const isLoading = ref(false)

  // Count overlapping runs and invalidate stale cleanup after a reset
  let activeLoadingRunCount = 0
  let loadingGeneration = 0

  // Store the readable error from the latest failed action
  const error = ref<string | null>(null)

  // Run one action while maintaining its loading and error state
  async function run<T>(fn: () => Promise<T>, opts: RunOptions = {}): Promise<T | undefined> {
    // Join this action to the current loading generation before executing
    const trackLoading = opts.trackLoading !== false
    const runLoadingGeneration = loadingGeneration
    if (trackLoading) {
      activeLoadingRunCount += 1
      isLoading.value = true
    }
    error.value = null

    try {
      // Wait for the main operation and keep its returned result
      const result = await fn()
      if (opts.after) await opts.after()
      return result
    } catch (caught) {
      error.value = errorMessage(caught, opts.fallback)

      // HTTP failures are reported by the API client; report later processing failures here
      if (!isNormalizedError(caught) && !opts.suppressErrorToast) {
        ToastEventBus.emit('add', {
          severity: 'error',
          summary: error.value,
          life: TOAST_LIFE.error,
          icon: null,
        })
      }

      if (opts.rethrow === false) return undefined
      throw caught
    } finally {
      // Keep loading active for concurrent runs and ignore cleanup invalidated by reset
      if (trackLoading && runLoadingGeneration === loadingGeneration) {
        activeLoadingRunCount = Math.max(0, activeLoadingRunCount - 1)
        isLoading.value = activeLoadingRunCount > 0
      }
    }
  }

  // Invalidate active cleanup and clear state when a consumer discards its data
  function reset(): void {
    loadingGeneration += 1
    activeLoadingRunCount = 0
    isLoading.value = false
    error.value = null
  }

  // Expose the reactive state and the typed action runner
  return { isLoading, error, run, reset }
}

/** useListFetcher creates an initialized list whose value is never null. */
export function useListFetcher<T>(
  fetcher: (config?: AxiosRequestConfig) => Promise<T>,
  fallbackMessage: string,
  options: InitializedListFetcherOptions<T>,
): InitializedListFetcher<T>

/** useListFetcher creates a list that stays null until its first successful request. */
export function useListFetcher<T>(
  fetcher: (config?: AxiosRequestConfig) => Promise<T>,
  fallbackMessage: string,
  options?: ListFetcherOptions<T>,
): ListFetcher<T>

/** useListFetcher loads reference-stable data without rejecting background polling. */
export function useListFetcher<T>(
  fetcher: (config?: AxiosRequestConfig) => Promise<T>,
  fallbackMessage: string,
  options: ListFetcherOptions<T> = {},
): ListFetcher<T> | InitializedListFetcher<T> {
  // Keep generic payloads intact instead of deeply unwrapping their nested refs
  const data = shallowRef<T | null>(options.initialData ? options.initialData() : null)

  // Reuse the standard loading and error lifecycle
  const state = useAsyncState()

  // Count resets so late responses can be ignored
  let generation = 0

  // Load the next value and publish it only when it is still current
  async function fetch(config?: AxiosRequestConfig): Promise<void> {
    // Capture the reset generation that owns this request
    const fetchGeneration = generation
    await state.run(
      async () => {
        // Ignore responses invalidated by a reset and avoid unchanged reactive assignments
        let next: T
        try {
          next = await fetcher(config)
        } catch (error) {
          if (fetchGeneration !== generation) return
          throw error
        }
        if (fetchGeneration !== generation) return

        // Reconcile custom structures before checking whether data changed
        const reconciled = options.reconcile ? options.reconcile(data.value, next) : next
        if (!sameJson(reconciled, data.value)) data.value = reconciled
      },
      {
        rethrow: false,
        fallback: fallbackMessage,
        trackLoading: config?.suppressLoader !== true,
      },
    )
  }

  // Clear data and invalidate every response still in flight
  function reset(): void {
    generation += 1
    data.value = options.initialData ? options.initialData() : null
    state.reset()
  }

  // Suppress loading indicators for background polling while keeping the sameJson guard
  function fetchSilent(): Promise<void> {
    return fetch({ suppressLoader: true })
  }

  // Expose stable data together with its request lifecycle
  return { data, isFetching: state.isLoading, fetchError: state.error, fetch, fetchSilent, reset }
}
