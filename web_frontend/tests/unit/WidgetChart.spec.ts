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

import { flushPromises, shallowMount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { i18n } from '@/i18n'

afterEach(() => {
  vi.doUnmock('chart.js/auto')
  vi.resetModules()
  vi.restoreAllMocks()
})

describe('WidgetChart', () => {
  it('shows a localized error when the chart module cannot load', async () => {
    // Arrange
    vi.doMock('chart.js/auto', () => {
      throw new Error('Chart module unavailable')
    })
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined)
    const { default: WidgetChart } = await import('@/components/widgets/WidgetChart.vue')

    // Act
    const wrapper = shallowMount(WidgetChart, {
      props: {
        event: 'Logins',
        usage: { Example: [{ week: 'W01', count: 1 }] },
      },
      global: {
        plugins: [i18n],
        directives: { tooltip: {} },
      },
    })
    await flushPromises()

    // Assert
    expect(wrapper.get('[data-testid="usage-chart-error"]').attributes('role')).toBe('alert')
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    expect(consoleError).toHaveBeenCalledOnce()
  })
})
