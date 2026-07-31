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

import FieldToggle from '@/components/elements/FieldToggle.vue'

describe('FieldToggle', () => {
  it('associates its label, renders required state, and validates on focus leave', async () => {
    // Arrange
    const validate = vi.fn()
    const wrapper = mount(FieldToggle, {
      props: { modelValue: false, label: 'Synchronize', required: true, validate },
      global: { plugins: [[PrimeVue, { theme: 'none' }]] },
    })

    // Act
    await wrapper.trigger('focusout')

    // Assert
    const input = wrapper.get('input')
    expect(wrapper.get('label').attributes('for')).toBe(input.attributes('id'))
    expect(wrapper.get('.field-toggle__required').text()).toBe('*')
    expect(validate).toHaveBeenCalledOnce()
  })
})
