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

import type { MaybeRefOrGetter, Ref } from 'vue'
import { shallowRef, toValue, watch } from 'vue'
import { z } from 'zod'

/** FormField is the string-keyed portion of a form model. */
type FormField<T> = Extract<keyof T, string>

/** ZodFormControls exposes field errors and validation actions for a form model. */
export interface ZodFormControls<T> {
  errors: Ref<Partial<Record<FormField<T>, string>>>
  validateField: (field: FormField<T>) => boolean
  validateAll: () => boolean
  clearErrors: () => void
}

/**
 * useZodForm provides shared Zod validation mechanics used by the
 * create/edit forms and panels (groups, databases add, feedback).
 *
 * Owns the `errors` map and the two operations every caller re-implemented:
 *   - `validateField(field)` — re-parse one field (on blur, or per step in a
 *     multi-step form); set/clear its error and return whether it passed.
 *   - `validateAll()` — parse the whole form on submit; on failure fill `errors`
 *     (first issue per field wins) and return false, else clear errors → true.
 *
 * Both the schema and the form value may be static or a ref/getter — groups
 * builds its schema from `computed` (so i18n stays reactive), and forms that
 * keep fields as separate refs can pass a getter that assembles the form object.
 */
export function useZodForm<T extends Record<string, unknown>>(
  schema: MaybeRefOrGetter<z.ZodObject<z.ZodRawShape>>,
  form: MaybeRefOrGetter<T>,
): ZodFormControls<T> {
  // Keep the generic error map shallow so Vue does not widen its field keys
  const errors = shallowRef<Partial<Record<FormField<T>, string>>>({})

  // Track fields that should validate immediately while the user edits
  const liveValidationFields = new Set<FormField<T>>()

  // Validate one field without revealing errors on untouched siblings
  function validateField(field: FormField<T>): boolean {
    // Look up the single-field schema from the object shape and parse only that value
    const fieldSchema = toValue(schema).shape[field]
    if (!fieldSchema) return true

    // Preserve the parser result so the error map and return value cannot diverge
    const result = z.safeParse(fieldSchema, toValue(form)[field])
    if (!result.success) liveValidationFields.add(field)
    errors.value = {
      ...errors.value,
      [field]: result.success ? undefined : result.error.issues[0]?.message,
    }
    return result.success
  }

  // Validate the complete model and reveal the first issue for each field
  function validateAll(): boolean {
    // Parse the entire form before replacing any visible errors
    const result = toValue(schema).safeParse(toValue(form))
    if (!result.success) {
      // Convert validation issues into one message per field
      const formValue = toValue(form)
      const fieldErrors: Partial<Record<FormField<T>, string>> = {}
      for (const issue of result.error.issues) {
        // Ignore object-level issues and paths that do not identify a form field
        const field = issue.path[0]
        if (!isFormField(field, formValue)) continue

        if (!fieldErrors[field]) {
          fieldErrors[field] = issue.message
          liveValidationFields.add(field)
        }
      }
      errors.value = fieldErrors
      return false
    }

    errors.value = {}
    return true
  }

  // Return the form to its pristine validation state
  function clearErrors(): void {
    errors.value = {}
    liveValidationFields.clear()
  }

  // Once an error has been revealed by blur or submit, keep that field's visible
  // state current while the user edits it. Untouched fields remain silent until
  // their first validation, preserving the pristine-submit discovery flow.
  watch(
    () => toValue(form),
    () => {
      for (const field of liveValidationFields) validateField(field)
    },
    { deep: true, flush: 'sync' },
  )

  return { errors, validateField, validateAll, clearErrors }
}

/** isFormField checks whether a Zod issue path begins with a field from the form model. */
function isFormField<T extends Record<string, unknown>>(
  value: PropertyKey | undefined,
  form: T,
): value is FormField<T> {
  return typeof value === 'string' && value in form
}
