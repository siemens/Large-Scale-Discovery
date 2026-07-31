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

import { type DeepReadonly, readonly, type Ref, ref, shallowRef } from 'vue'

import { useUnsavedChangesStore } from '@/stores/unsavedChanges'

/**
 * Reactive single-slot panel state plus its guarded open/close handlers,
 * returned by {@link useGuardedPanel}.
 */
export interface GuardedPanel<Kind extends string, Target> {
  /** Which form is currently shown, or `null` when the slot is empty. */
  activeKind: Ref<Kind | null>
  /** The row/context the active form is bound to (e.g. the scope/view/group). */
  target: Ref<Target | null>
  /** Whether the discard-changes confirm is currently open. */
  confirmVisible: Ref<boolean>
  /**
   * Monotonically increasing counter — incremented on every successful open,
   * including re-opens of the same kind + target. Include in a `watch()`
   * dependency array alongside `activeKind` / `target` so the watcher fires
   * even when only the generation changes (e.g. scroll-to-panel on re-open).
   */
  openCount: DeepReadonly<Ref<number>>
  /**
   * Builds a stable `:key` string for a hosted form component. The key
   * embeds a monotonically increasing generation counter so that re-opening
   * the same form (same kind + target) always forces a fresh mount and a
   * clean `resetForm()` run — even when Vue would otherwise reuse the element.
   *
   * @example
   * ```vue
   * <FormViewEdit :key="panelKey('edit', selectedView?.id)" ... />
   * <ScopeAddForm :key="panelKey('networks')" ... />
   * ```
   */
  keyFor: (kind: Kind, id?: string | number | null) => string
  /**
   * Show a form in the single slot. If another form is open with unsaved edits,
   * the switch is deferred behind the discard confirm; otherwise it happens at
   * once. Passing the same kind + target simply re-asserts the current panel.
   */
  open: (kind: Kind, target?: Target | null) => void
  /**
   * Close the slot directly. Cancel and post-save both close without a prompt.
   * The discard guard applies only to switching/opening another panel while the
   * current one is dirty (see {@link GuardedPanel.open}).
   */
  close: () => void
  /** Confirm handler for the discard dialog — applies the deferred switch. */
  onConfirmDiscard: () => void
}

/**
 * Page-side state machine for the single inline-form slot.
 *
 * A page renders at most one form at a time above its content. Opening a form
 * while another is open replaces it; opening the same form for a different row
 * re-targets it. When the open form has unsaved edits (tracked centrally by
 * {@link useUnsavedChangesStore}), replacing it is gated behind the shared
 * "discard unsaved changes?" confirmation. Explicit cancel and post-save close
 * actions remain immediate because the caller has already resolved the form.
 */
export function useGuardedPanel<Kind extends string, Target = unknown>(): GuardedPanel<
  Kind,
  Target
> {
  // Read the shared registry of forms with unsaved changes
  const unsaved = useUnsavedChangesStore()

  // Store which action panel is currently open
  const activeKind: Ref<Kind | null> = shallowRef<Kind | null>(null)

  // Keep the selected row opaque so Vue does not deeply unwrap generic targets
  const target: Ref<Target | null> = shallowRef<Target | null>(null)

  // Control the discard-confirmation dialog
  const confirmVisible = ref(false)

  // Increment the render key whenever panel content changes
  const generation = ref(0)

  // Retain the requested replacement while the discard confirmation is open
  const pending = shallowRef<{ kind: Kind; target: Target | null } | null>(null)

  // Replace the currently open panel without another guard check
  function apply(kind: Kind | null, nextTarget: Target | null): void {
    activeKind.value = kind
    target.value = nextTarget
    generation.value++
  }

  // Build a stable render key for the active panel and target
  function keyFor(kind: Kind, id?: string | number | null): string {
    return `${kind}-${id ?? 'new'}-${generation.value}`
  }

  // Open immediately or ask before discarding unsaved work
  function requestOpen(kind: Kind, nextTarget: Target | null): void {
    // Defer replacement when the current form still owns unsaved input
    if (activeKind.value !== null && unsaved.isDirty) {
      pending.value = { kind, target: nextTarget }
      confirmVisible.value = true
      return
    }

    apply(kind, nextTarget)
  }

  // Request a different action panel
  function open(kind: Kind, nextTarget: Target | null = null): void {
    requestOpen(kind, nextTarget)
  }

  // Request closing the current action panel
  function close(): void {
    // Cancel and post-save close only after the caller has resolved the current form
    apply(null, null)
  }

  // Continue the pending panel change after discard confirmation
  function onConfirmDiscard(): void {
    if (pending.value) {
      apply(pending.value.kind, pending.value.target)
      pending.value = null
    }
  }

  // Expose the reactive slot state plus the guarded handlers
  return {
    activeKind,
    target,
    confirmVisible,
    openCount: readonly(generation),
    keyFor,
    open,
    close,
    onConfirmDiscard,
  }
}
