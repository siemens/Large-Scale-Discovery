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

import { flushPromises, mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { z } from 'zod'

import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import DialogForm from '@/components/elements/DialogForm.vue'
import FieldHint from '@/components/elements/FieldHint.vue'
import LayoutForm from '@/components/layout/LayoutForm.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormSubmitLock } from '@/composables/useFormSubmitLock'
import { useZodForm } from '@/composables/useZodForm'
import { i18n } from '@/i18n'

describe('useZodForm visible-error state', () => {
  it('keeps pristine invalid forms submittable, then tracks revealed errors live', () => {
    const form = ref({ name: '', count: 0 })
    const schema = z.object({ name: z.string().min(3), count: z.number().min(1) })
    const { errors, validateAll, clearErrors } = useZodForm(schema, form)

    expect(errors.value).toEqual({})

    expect(validateAll()).toBe(false)
    expect(errors.value.name).toBeTruthy()
    expect(errors.value.count).toBeTruthy()

    form.value.name = 'Valid name'
    expect(errors.value.name).toBeUndefined()
    expect(errors.value.count).toBeTruthy()

    form.value.count = 1
    expect(Object.values(errors.value).some(Boolean)).toBe(false)

    form.value.name = 'x'
    expect(errors.value.name).toBeTruthy()

    clearErrors()
    expect(errors.value).toEqual({})
  })
})

describe('shared form visible-error aggregation', () => {
  const ButtonStub = {
    inheritAttrs: false,
    props: ['type', 'form', 'label', 'disabled', 'loading'],
    emits: ['click'],
    template: `
      <button
        :type="type"
        :form="form"
        :disabled="disabled"
        :data-loading="loading ? 'true' : undefined"
        v-bind="$attrs"
        @click="$emit('click')"
      >{{ label }}</button>
    `,
  }

  it('updates standard and integrated SubmitFooters from rendered errors', async () => {
    const error = ref<string>()
    const submitHandler = vi.fn()

    // Inline harness keeps the provider/slot relationship under test in one place.
    // eslint-disable-next-line vue/one-component-per-file
    const Host = defineComponent({
      components: { FieldHint, LayoutForm, LayoutSubmitFooter },
      setup: () => ({ error, submitHandler }),
      template: `
        <LayoutForm
          title="Test form"
          text-submit-label="Submit"
          submit-test-id="standard-submit"
          :submit-handler="submitHandler"
        >
          <FieldHint :error="error" />
          <LayoutSubmitFooter
            text-cancel-label="Cancel"
            text-submit-label="Submit"
            submit-test-id="integrated-submit"
          />
        </LayoutForm>
      `,
    })
    const wrapper = mount(Host, {
      global: {
        plugins: [i18n],
        stubs: {
          SectionHeader: { template: '<header><slot /></header>' },
          Button: ButtonStub,
          OrConnector: { template: '<span />' },
        },
      },
    })

    expect(wrapper.get('[data-testid="standard-submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('[data-testid="integrated-submit"]').attributes('disabled')).toBeUndefined()

    error.value = 'Required'
    await nextTick()
    expect(wrapper.get('[data-testid="standard-submit"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="integrated-submit"]').attributes('disabled')).toBeDefined()

    error.value = undefined
    await nextTick()
    expect(wrapper.get('[data-testid="standard-submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('[data-testid="integrated-submit"]').attributes('disabled')).toBeUndefined()
  })

  it('automatically updates DialogForm submit buttons from rendered errors', async () => {
    const error = ref<string>()
    const submitHandler = vi.fn()

    // Inline harness keeps the provider/teleported-footer relationship under test.
    // eslint-disable-next-line vue/one-component-per-file
    const Host = defineComponent({
      components: { DialogForm, FieldHint },
      setup: () => ({ error, submitHandler }),
      template: `
        <DialogForm
          :visible="true"
          title="Test dialog"
          text-submit-label="Submit"
          submit-test-id="dialog-submit"
          :submit-handler="submitHandler"
        >
          <FieldHint :error="error" />
        </DialogForm>
      `,
    })
    const wrapper = mount(Host, {
      global: {
        plugins: [i18n],
        stubs: {
          Dialog: { template: '<section><slot /><slot name="footer" /></section>' },
          Button: ButtonStub,
          OrConnector: { template: '<span />' },
          DialogConfirm: true,
        },
      },
    })

    const submit = wrapper.get('[data-testid="dialog-submit"]')
    expect(submit.attributes('disabled')).toBeUndefined()

    error.value = 'Required'
    await nextTick()
    expect(submit.attributes('disabled')).toBeDefined()
  })

  it('blocks duplicate LayoutForm submissions and exposes pending state to its footer', async () => {
    let resolveSubmit: (() => void) | undefined
    const submitHandler = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveSubmit = resolve
        }),
    )

    // eslint-disable-next-line vue/one-component-per-file
    const Host = defineComponent({
      components: { LayoutForm },
      setup: () => ({ submitHandler }),
      template: `
        <LayoutForm
          title="Pending form"
          text-submit-label="Submit"
          submit-test-id="pending-submit"
          :submit-handler="submitHandler"
        />
      `,
    })
    const wrapper = mount(Host, {
      global: {
        plugins: [i18n],
        stubs: {
          SectionHeader: { template: '<header><slot /></header>' },
          Button: ButtonStub,
          OrConnector: { template: '<span />' },
        },
      },
    })

    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')

    expect(submitHandler).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="pending-submit"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="pending-submit"]').attributes('data-loading')).toBe('true')

    resolveSubmit?.()
    await flushPromises()

    expect(wrapper.get('[data-testid="pending-submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('[data-testid="pending-submit"]').attributes('data-loading')).toBeUndefined()
  })
})

describe('useFormSubmitLock', () => {
  it('ignores repeated actions without requiring a disabled button', async () => {
    let resolveSubmit: (() => void) | undefined
    const action = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveSubmit = resolve
        }),
    )
    const { isSubmitting, run } = useFormSubmitLock()

    const first = run(action)
    const duplicate = run(action)

    expect(action).toHaveBeenCalledTimes(1)
    expect(isSubmitting.value).toBe(true)

    resolveSubmit?.()
    await Promise.all([first, duplicate])
    expect(isSubmitting.value).toBe(false)
  })
})

describe('DialogConfirm message emphasis', () => {
  function mountDialog(
    message: string,
    secondaryMessage?: string,
    severity: 'danger' | 'warn' | 'primary' | 'secondary' = 'danger',
  ) {
    return mount(DialogConfirm, {
      props: {
        visible: true,
        title: 'Confirm',
        message,
        secondaryMessage,
        severity,
        textCancelLabel: 'Cancel',
        textConfirmLabel: 'Confirm',
        textOr: 'or',
      },
      global: {
        plugins: [[PrimeVue, { theme: 'none' }], i18n],
        stubs: {
          Dialog: {
            template: '<section><slot name="header" /><slot /><slot name="footer" /></section>',
          },
        },
      },
    })
  }

  it('renders every value between paired quotes as strong text', () => {
    const wrapper = mountDialog("Delete 'scope-a' from 'group-b'?", "Revoke 'token-c' too.")

    // Three quoted values → three strong elements
    expect(wrapper.findAll('strong')).toHaveLength(3)
  })

  it('renders custom body content through the documented default slot', () => {
    const wrapper = mount(DialogConfirm, {
      props: {
        visible: true,
        title: 'Confirm',
        textCancelLabel: 'Cancel',
        textConfirmLabel: 'Confirm',
        textOr: 'or',
      },
      slots: { default: '<p data-testid="custom-confirm-copy">Custom copy</p>' },
      global: {
        plugins: [[PrimeVue, { theme: 'none' }], i18n],
        stubs: {
          Dialog: { template: '<section><slot /></section>' },
        },
      },
    })

    expect(wrapper.find('[data-testid="custom-confirm-copy"]').exists()).toBe(true)
  })
})
