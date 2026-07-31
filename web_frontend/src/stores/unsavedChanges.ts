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

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/** useUnsavedChangesStore remembers which open forms contain unsaved edits. */
export const useUnsavedChangesStore = defineStore('unsavedChanges', () => {
  // Store each dirty form once under its unique identifier
  const dirtySources = ref<Set<string>>(new Set())

  // Tell navigation guards whether any registered form is dirty
  const isDirty = computed(() => dirtySources.value.size > 0)

  // Add or remove one form from the dirty registry
  function setDirty(id: string, dirty: boolean): void {
    // Keep the registry synchronized with the form's latest state
    if (dirty) dirtySources.value.add(id)
    else dirtySources.value.delete(id)
  }

  // Forced authentication loss discards every open form without prompting
  function discardAll(): void {
    dirtySources.value = new Set()
  }

  // Keep registry details private so consumers cannot bypass setDirty
  return { discardAll, isDirty, setDirty }
})
