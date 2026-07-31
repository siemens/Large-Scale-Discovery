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
import InputNumber from 'primevue/inputnumber'

import FieldNumber from '@/components/elements/FieldNumber.vue'

describe('FieldNumber', () => {
  it('forwards numeric constraints and associates the supplied id with the input', () => {
    // Arrange and act
    const wrapper = mount(FieldNumber, {
      props: {
        modelValue: 5,
        id: 'timeout',
        min: 1,
        max: 10,
        step: 2,
        invalid: true,
        unit: 'Seconds',
      },
      global: { plugins: [[PrimeVue, { theme: 'none' }]] },
    })

    // Assert
    const inputNumber = wrapper.getComponent(InputNumber)
    expect(inputNumber.props()).toMatchObject({ min: 1, max: 10, step: 2, invalid: true })
    expect(wrapper.get('input').attributes('id')).toBe('timeout')
    expect(wrapper.get('.p-inputgroupaddon').text()).toBe('Seconds')
  })
})
