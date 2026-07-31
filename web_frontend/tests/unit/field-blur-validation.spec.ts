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

import { mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import InputText from 'primevue/inputtext'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'

import Field from '@/components/elements/Field.vue'

const InputFieldHost = defineComponent({
  components: { Field, InputText },
  props: { invalid: { type: Boolean, required: true } },
  template: `
    <Field label="Email address" :error="invalid ? 'Invalid address' : undefined">
      <template #default="{ id }">
        <InputText :id="id" :invalid="invalid" />
      </template>
    </Field>
  `,
})

describe('Field leave validation', () => {
  it('validates only when focus leaves the complete field', async () => {
    const validate = vi.fn()
    const wrapper = mount(Field, {
      props: { label: 'Composite field', validate },
      slots: {
        default: '<button id="first">First</button><button id="second">Second</button>',
      },
    })
    const field = wrapper.get('.field')

    await field.trigger('focusout', { relatedTarget: wrapper.get('#second').element })
    expect(validate).not.toHaveBeenCalled()

    await field.trigger('focusout', { relatedTarget: document.body })
    expect(validate).toHaveBeenCalledOnce()
  })

  it('keeps an invalid control border visible through focus and blur until corrected', async () => {
    // Arrange — render a field whose current content has already failed validation
    const wrapper = mount(InputFieldHost, {
      props: { invalid: true },
      attachTo: document.body,
      global: { plugins: [[PrimeVue, { theme: 'none' }]] },
    })
    const input = wrapper.get<HTMLInputElement>('input')

    // Act — focus and leave the still-invalid field
    input.element.focus()
    expect(input.attributes('aria-invalid')).toBe('true')
    input.element.blur()

    // Assert — invalid state survives blur, then disappears only after correction
    expect(input.attributes('aria-invalid')).toBe('true')
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    await wrapper.setProps({ invalid: false })
    input.element.focus()
    expect(input.attributes('aria-invalid')).toBeUndefined()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)

    wrapper.unmount()
  })
})
