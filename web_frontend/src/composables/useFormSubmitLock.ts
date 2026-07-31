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
import { readonly, ref } from 'vue'

/** FormSubmitLock exposes one guarded submit action and its running state. */
export interface FormSubmitLock {
  readonly isSubmitting: Readonly<Ref<boolean>>
  run: (action: () => Promise<void> | void) => Promise<void>
}

/** Run at most one submit action until the accepted action has settled. */
export function useFormSubmitLock(): FormSubmitLock {
  // Track whether a submission is already running
  const isSubmitting = ref(false)

  // Run one submit action while blocking duplicate attempts
  async function run(action: () => Promise<void> | void): Promise<void> {
    if (isSubmitting.value) return
    isSubmitting.value = true
    try {
      await action()
    } finally {
      isSubmitting.value = false
    }
  }

  // Prevent consumers from changing the running state directly
  return { isSubmitting: readonly(isSubmitting), run }
}
