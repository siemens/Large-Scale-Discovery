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
import Slider from 'primevue/slider'

import FieldNumber from '@/components/elements/FieldNumber.vue'
import FieldSlider from '@/components/elements/FieldSlider.vue'

describe('FieldSlider', () => {
  it('keeps the slider and optional numeric control on one model contract', async () => {
    // Arrange
    const wrapper = mount(FieldSlider, {
      props: { modelValue: 10, label: 'Expiry', min: 1, max: 30, showValue: true },
      global: { plugins: [[PrimeVue, { theme: 'none' }]] },
    })

    // Act
    wrapper.getComponent(Slider).vm.$emit('update:modelValue', 20)
    wrapper.getComponent(FieldNumber).vm.$emit('update:modelValue', null)
    await wrapper.vm.$nextTick()

    // Assert
    expect(wrapper.emitted('update:modelValue')).toEqual([[20], [1]])
  })

  it('exposes an invalid state and validates only after focus leaves the field', async () => {
    // Arrange
    const validate = vi.fn()
    const wrapper = mount(FieldSlider, {
      props: { modelValue: 10, label: 'Expiry', error: 'Invalid expiry', validate },
      global: { plugins: [[PrimeVue, { theme: 'none' }]] },
    })

    // Act
    await wrapper.trigger('focusout')

    // Assert
    expect(validate).toHaveBeenCalledOnce()
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
  })
})
