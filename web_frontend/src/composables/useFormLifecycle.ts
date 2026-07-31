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

import { onMounted, watch } from 'vue'

import { useUnsavedChanges } from '@/composables/useUnsavedChanges'

/** Configuration for {@link useFormLifecycle}. */
export interface FormLifecycleOptions {
  /**
   * Optional target getter (e.g. `() => props.scope`). When it changes, the form
   * re-initializes to the new row via `reset()`. Omit for add-only forms with no
   * target.
   */
  target?: () => unknown
  /** Initialize the form to its target. Runs on mount and on every target change. */
  reset: () => void
  /**
   * Dirty-state getter. Registered with the unsaved-changes guard so the page's
   * `<UnchangedGuard>` and the guarded panel switch both see this form's edits.
   */
  dirty: () => boolean
  /**
   * Perform the submit. Return `true` to close the panel on success; `false` /
   * `undefined` keeps it open (e.g. validation failed). Throwing routes to
   * `onError`. Do the success toast and `emit('saved')` inside here — those vary
   * per form.
   */
  submit: () => Promise<boolean | void>
  /** Handle a thrown submit error (typically a toast). The panel stays open. */
  onError?: (error: unknown) => void
  /** Close the panel (typically `() => emit('close')`). */
  close: () => void
}

/** FormLifecycle contains the submit and cancel handlers returned by {@link useFormLifecycle}. */
export interface FormLifecycle {
  /** Run the submit flow; closes the panel when {@link FormLifecycleOptions.submit} resolves truthy. */
  submit: () => Promise<void>
  /** Close the panel without submitting (Cancel). */
  cancel: () => void
}

/**
 * useFormLifecycle provides the shared lifecycle for single-slot inline form panels: reset-on-open (and on
 * target change), unsaved-changes registration, `cancel()`, and the submit
 * try/catch skeleton.
 *
 * Unlike a modal dialog that watches a `visible` prop, a single-slot panel is
 * mounted only while open, so this resets on mount and re-resets whenever the
 * bound target changes (re-targeting the same form to another row). Validation,
 * toasts, and `emit('saved')` stay in the caller's `submit` closure — that is
 * where the forms genuinely differ.
 */
export function useFormLifecycle(options: FormLifecycleOptions): FormLifecycle {
  // Initialize on mount — the panel is rendered only when open, so mount == open
  onMounted(() => {
    options.reset()
  })

  // Re-initialize when the bound target changes (same form, different row)
  if (options.target) {
    watch(options.target, () => {
      options.reset()
    })
  }

  // Publish this form's dirty state into the central unsaved-changes guard
  useUnsavedChanges(options.dirty)

  // Cancel the form and return to its parent view
  function cancel(): void {
    options.close()
  }

  // Submit once and report success or failure to the user
  async function submit(): Promise<void> {
    try {
      if (await options.submit()) options.close()
    } catch (error) {
      // The API client reports request failures; invoke only specialized callbacks here
      options.onError?.(error)
    }
  }

  // Expose the two handlers the template binds to
  return { submit, cancel }
}

/** FormSnapshot tracks a serialized baseline for unsaved-change detection. */
export interface FormSnapshot {
  capture: () => void
  isDirty: () => boolean
}

/** useFormSnapshot tracks a serialized baseline for unsaved-change detection. */
export function useFormSnapshot(snapshot: () => string, initial?: string): FormSnapshot {
  // Store the last known clean form snapshot
  let baseline = initial ?? snapshot()

  // Save the current form state as the clean baseline
  function capture(): void {
    baseline = snapshot()
  }

  // Compare against the last captured state without mutating the baseline
  function isDirty(): boolean {
    return snapshot() !== baseline
  }

  return { capture, isDirty }
}
