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

import type { ComputedRef, InjectionKey, MaybeRefOrGetter, Ref } from 'vue'
import { computed, inject, onBeforeUnmount, provide, shallowReactive, toValue } from 'vue'

// FormValidationRegistry coordinates child errors and validators with one parent form
interface FormValidationRegistry {
  hasVisibleErrors: ComputedRef<boolean>
  submitDisabled: ComputedRef<boolean>
  submitLoading: ComputedRef<boolean>
  register: (showError: Readonly<Ref<boolean>>) => () => void
  registerValidator: (validator: FormValidator) => () => void
  validate: () => Promise<boolean>
}

/** FormValidator reveals its own errors and reports whether submission may continue. */
export type FormValidator = () => boolean | Promise<boolean>

// FormValidationOptions supplies external loading and disabled state
interface FormValidationOptions {
  /** Non-validation reasons to disable submission, including an in-flight request. */
  submitDisabled?: MaybeRefOrGetter<boolean>
  /** Whether the shared submit control should show its loading state. */
  submitLoading?: MaybeRefOrGetter<boolean>
}

/** FormSubmitState exposes the final loading and disabled state for a submit control. */
export interface FormSubmitState {
  submitDisabled: ComputedRef<boolean>
  submitLoading: ComputedRef<boolean>
}

/** FormValidationState adds descendant validation to the shared submit state. */
export interface FormValidationState extends FormSubmitState {
  validate: () => Promise<boolean>
}

// Identify the validation registry shared through Vue dependency injection
const formValidationKey: InjectionKey<FormValidationRegistry> = Symbol('form-validation')

/**
 * Create the visible-error registry owned by a shared form container.
 * Descendant FieldHints register themselves, so forms do not need to wire
 * their validation implementation to the submit button individually.
 */
export function provideFormValidation(options: FormValidationOptions = {}): FormValidationState {
  // Collect visible-error flags registered by child fields
  const errorSources = shallowReactive(new Set<Readonly<Ref<boolean>>>())

  // Collect validation callbacks registered by child fields
  const validators = new Set<FormValidator>()

  // Summarize errors that child fields have chosen to reveal
  const hasVisibleErrors = computed(() => [...errorSources].some((source) => source.value))

  // Reflect the parent action loading state
  const submitLoading = computed(() => toValue(options.submitLoading) ?? false)

  // Disable submission while loading or while visible errors remain
  const submitDisabled = computed(
    () =>
      (toValue(options.submitDisabled) ?? false) || submitLoading.value || hasVisibleErrors.value,
  )

  // Run every registered validator and combine their results
  async function validate(): Promise<boolean> {
    // Wait for every child validator before deciding whether the form is valid
    const results = await Promise.all([...validators].map(async (validator) => validator()))
    return results.every(Boolean)
  }

  provide(formValidationKey, {
    hasVisibleErrors,
    submitDisabled,
    submitLoading,
    register(showError): () => void {
      errorSources.add(showError)
      return () => errorSources.delete(showError)
    },
    registerValidator(validator): () => void {
      validators.add(validator)
      return () => validators.delete(validator)
    },
    validate,
  })

  return { submitDisabled, submitLoading, validate }
}

/** useFormSubmitState reads the state supplied by the nearest shared form container. */
export function useFormSubmitState(): FormSubmitState {
  // Find the registry supplied by the nearest parent component
  const registry = inject(formValidationKey, null)
  if (registry) {
    return {
      submitDisabled: registry.submitDisabled,
      submitLoading: registry.submitLoading,
    }
  }

  return {
    submitDisabled: computed(() => false),
    submitLoading: computed(() => false),
  }
}

/** registerFormValidationError links one rendered validation hint to its nearest form. */
export function registerFormValidationError(showError: Readonly<Ref<boolean>>): void {
  // Find the registry supplied by the nearest parent component
  const registry = inject(formValidationKey, null)
  if (!registry) return

  // Remove this child registration when its component unmounts
  const unregister = registry.register(showError)
  onBeforeUnmount(unregister)
}

/** registerFormValidator links one child-owned validator to its nearest shared form. */
export function registerFormValidator(validator: FormValidator): void {
  // Find the registry supplied by the nearest parent component
  const registry = inject(formValidationKey, null)
  if (!registry) return

  // Remove this child registration when its component unmounts
  const unregister = registry.registerValidator(validator)
  onBeforeUnmount(unregister)
}
