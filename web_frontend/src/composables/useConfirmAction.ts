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

import type { ComputedRef, Ref } from 'vue'
import { computed, ref, shallowRef } from 'vue'

/**
 * A pending confirm-gated operation: the target item plus the async action to run
 * against it when the user confirms.
 */
interface PendingAction<T, R> {
  /** The item the action operates on (a scope, view, grant...). */
  item: T
  /** Performs the operation; its resolved value is forwarded to `onSuccess`. */
  run: (item: T) => Promise<R>
  /** Called after `run` resolves. */
  onSuccess?: (result: R, item: T) => void
  /** Called if `run` rejects. */
  onError?: (error: unknown, item: T) => void
}

/** The reactive confirm state plus its ask/confirm handlers from {@link useConfirmAction}. */
export interface ConfirmAction<T, R> {
  /** Bind to the dialog's `v-model:visible`. */
  visible: Ref<boolean>
  /** The currently-pending action, or `null` when none is queued. */
  pending: Ref<PendingAction<T, R> | null>
  /** Queue an action and open the confirm dialog. */
  ask: (action: PendingAction<T, R>) => void
  /** Run the pending action (bind to the dialog's confirm event). */
  confirm: () => Promise<void>
}

/** ConfirmDeleteOptions configures the fixed operation used by useConfirmDelete. */
export interface ConfirmDeleteOptions<T, R> {
  remove: (item: T) => Promise<R>
  onSuccess?: (result: R, item: T) => void
  onError?: (error: unknown, item: T) => void
}

/** ConfirmDelete exposes delete-oriented aliases for the generic confirmation flow. */
export interface ConfirmDelete<T> {
  deleteConfirmVisible: Ref<boolean>
  pendingDeleteTarget: ComputedRef<T | null>
  openDeleteConfirm: (item: T) => void
  confirmDelete: () => Promise<void>
}

/**
 * Generic "ask, then run an action" confirm flow. Unlike {@link useConfirmDelete}
 * (one fixed `remove` operation), the operation is supplied per-invocation via
 * `ask`, so a single dialog can gate several distinct destructive actions
 * (e.g. a scope's delete and reset-secret) by passing a different `run` /
 * `onSuccess` each time.
 *
 * The dialog is expected to close itself on confirm; `confirm` only clears the
 * pending action when done, and `ask` opens it.
 */
export function useConfirmAction<T = unknown, R = unknown>(): ConfirmAction<T, R> {
  // Control the confirmation dialog independently from the queued action
  const visible = ref(false)

  // Avoid deep unwrapping the generic callback-bearing action
  const pending = shallowRef<PendingAction<T, R> | null>(null)

  // Store the requested action and open its confirmation dialog
  function ask(action: PendingAction<T, R>): void {
    pending.value = action
    visible.value = true
  }

  // Run the pending action after the user confirms it
  async function confirm(): Promise<void> {
    // Take the pending action before closing the dialog
    const action = pending.value
    if (!action) return

    try {
      // Forward the resolved value so callers can update their surrounding view
      const result = await action.run(action.item)
      action.onSuccess?.(result, action.item)
    } catch (error) {
      // The API client reports request failures; invoke only specialized callbacks here
      action.onError?.(error, action.item)
    } finally {
      pending.value = null
    }
  }

  // Expose the reactive state plus the ask/confirm handlers
  return { visible, pending, ask, confirm }
}

/** useConfirmDelete specializes useConfirmAction for one fixed remove operation. */
export function useConfirmDelete<T, R = unknown>(
  options: ConfirmDeleteOptions<T, R>,
): ConfirmDelete<T> {
  // Reuse the generic execution and error-handling flow under delete-oriented aliases
  const action = useConfirmAction<T, R>()

  // Store the item waiting for delete confirmation
  const pendingDeleteTarget = computed(() => action.pending.value?.item ?? null)

  // Remember the item and open its delete confirmation
  function openDeleteConfirm(item: T): void {
    action.ask({
      item,
      run: options.remove,
      onSuccess: options.onSuccess,
      onError: options.onError,
    })
  }

  return {
    deleteConfirmVisible: action.visible,
    pendingDeleteTarget,
    openDeleteConfirm,
    confirmDelete: action.confirm,
  }
}
