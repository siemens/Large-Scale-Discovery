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
import { createI18n } from 'vue-i18n'

import FieldTimespans from '@/components/elements/FieldTimespans.vue'

const textProps = {
  textDayOptions: [{ label: 'Monday', value: 1 }],
}

const i18n = createI18n({ legacy: false, locale: 'en-US' })

interface TimespansHandle {
  timespans: Array<{
    id: number
    startDay: number
    startTime: string
    endDay: number
    endTime: string
  }>
  validTimespans: Array<{ startDay: number; startTime: string; endDay: number; endTime: string }>
  hasInvalidTimespans: boolean
}

describe('FieldTimespans', () => {
  beforeAll(() => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    )
  })

  afterAll(() => vi.unstubAllGlobals())

  it('normalizes API days and keeps one empty row available for additions', () => {
    // Arrange and act
    const wrapper = mount(FieldTimespans, {
      props: {
        ...textProps,
        initialTimespans: [{ startDay: '1', startTime: '08:00', endDay: '1', endTime: '09:00' }],
      },
      global: { plugins: [i18n, [PrimeVue, { theme: 'none' }]] },
    })
    const component = wrapper.vm as unknown as TimespansHandle

    // Assert
    expect(component.validTimespans).toHaveLength(1)
    expect(component.validTimespans[0]?.startDay).toBe(1)
    expect(component.timespans).toHaveLength(2)
    expect(component.timespans.at(-1)).toMatchObject({
      startDay: -1,
      startTime: '',
      endDay: -1,
      endTime: '',
    })
  })

  it('flags equal complete endpoints and removes a completed window', async () => {
    // Arrange
    const wrapper = mount(FieldTimespans, {
      props: {
        ...textProps,
        initialTimespans: [{ startDay: '1', startTime: '08:00', endDay: '1', endTime: '08:00' }],
      },
      global: { plugins: [i18n, [PrimeVue, { theme: 'none' }]] },
    })
    const component = wrapper.vm as unknown as TimespansHandle

    // Act and assert
    expect(component.hasInvalidTimespans).toBe(true)
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    await wrapper.get('button[aria-label="Remove"]').trigger('click')
    expect(component.validTimespans).toHaveLength(0)
    expect(component.timespans).toHaveLength(1)
  })
})
