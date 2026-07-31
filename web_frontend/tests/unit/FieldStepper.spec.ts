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
import { describe, expect, it } from 'vitest'

import FieldStepper from '@/components/elements/FieldStepper.vue'
import { i18n } from '@/i18n'

const props = {
  label: 'Retention',
}

describe('FieldStepper', () => {
  it('cycles retention between all and positive values', async () => {
    // Arrange
    const wrapper = mount(FieldStepper, {
      props: { ...props, modelValue: 'All' },
      global: { plugins: [[PrimeVue, { theme: 'none' }], i18n] },
    })

    // Act and assert
    await wrapper.get('[aria-label="Increase cycle retention"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([1])
    await wrapper.setProps({ modelValue: 1 })
    await wrapper.get('[aria-label="Decrease cycle retention"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['All'])
  })

  it('steps through valid split sizes', async () => {
    // Arrange
    const wrapper = mount(FieldStepper, {
      props: {
        kind: 'splitSize',
        label: 'Split Size',
        modelValue: 1_024,
      },
      global: { plugins: [[PrimeVue, { theme: 'none' }], i18n] },
    })

    // Act
    await wrapper.get('[aria-label="Increase network size"]').trigger('click')

    // Assert
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([2_048])

    // Act
    await wrapper.setProps({ modelValue: 2_048 })
    await wrapper.get('[aria-label="Decrease network size"]').trigger('click')

    // Assert
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([1_024])
  })

  it('stops split sizes at the Nmap maximum', () => {
    // Arrange
    const wrapper = mount(FieldStepper, {
      props: {
        kind: 'splitSize',
        label: 'Split Size',
        modelValue: 2_048,
      },
      global: { plugins: [[PrimeVue, { theme: 'none' }], i18n] },
    })

    // Act
    const increaseButton = wrapper.get('[aria-label="Increase network size"]')

    // Assert
    expect(increaseButton.attributes('disabled')).toBeDefined()
  })

  it('stops split sizes at one address', () => {
    // Arrange
    const wrapper = mount(FieldStepper, {
      props: {
        kind: 'splitSize',
        label: 'Split Size',
        modelValue: 1,
      },
      global: { plugins: [[PrimeVue, { theme: 'none' }], i18n] },
    })

    // Act
    const decreaseButton = wrapper.get('[aria-label="Decrease network size"]')

    // Assert
    expect(decreaseButton.attributes('disabled')).toBeDefined()
  })

  it('disables both actions without changing the value', async () => {
    // Arrange
    const wrapper = mount(FieldStepper, {
      props: { ...props, modelValue: 2, disabled: true },
      global: { plugins: [[PrimeVue, { theme: 'none' }], i18n] },
    })

    // Act
    await wrapper.get('[aria-label="Increase cycle retention"]').trigger('click')
    await wrapper.get('[aria-label="Decrease cycle retention"]').trigger('click')

    // Assert
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(
      wrapper.findAll('button').every((button) => button.attributes('disabled') !== undefined),
    ).toBe(true)
    expect(wrapper.get('[aria-disabled="true"]').get('label').text()).toBe('Retention')
  })
})
