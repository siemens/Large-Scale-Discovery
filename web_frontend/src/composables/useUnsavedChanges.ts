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
import { onBeforeUnmount, watch } from 'vue'

import { useUnsavedChangesStore } from '@/stores/unsavedChanges'

// Issue one unique registry key to each mounted form
let registrationCounter = 0

/**
 * Publish a form's dirty state into the central guard store.
 *
 * Called by an editable form to register its dirty state with
 * {@link useUnsavedChangesStore}. The application-level
 * `<UnsavedDialog>` then prompts before navigation while any registered
 * form is dirty.
 *
 * The registration is cleared automatically on unmount, so a panel that is torn
 * down after a successful submit stops contributing to the dirty state without
 * the component having to reset it explicitly.
 *
 * @param isDirty Reactive boolean (Ref or getter) — true while the form holds
 *                unsaved user input. Single-slot panels unmount on close, which
 *                auto-clears their dirty state. A component that stays mounted
 *                while hidden should fold its `visible` flag into this so it
 *                never reports dirty while closed.
 */
export function useUnsavedChanges(isDirty: Ref<boolean> | (() => boolean)): void {
  // Each instance gets a unique id so multiple forms on one page don't collide
  const store = useUnsavedChangesStore()

  // Give this instance a stable registry identifier
  const registrationId = `unsaved-${++registrationCounter}`

  // Keep the cleanup callback returned by the registry
  const stopWatching = watch(isDirty, (dirty) => store.setDirty(registrationId, dirty), {
    immediate: true,
  })

  onBeforeUnmount(() => {
    stopWatching()
    store.setDirty(registrationId, false)
  })
}
