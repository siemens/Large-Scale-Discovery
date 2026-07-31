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

import ProgressBar3 from '@/components/elements/ProgressBar3.vue'
import { i18n } from '@/i18n'

describe('ProgressBar3', () => {
  it('renders normalized segment widths and an accessible summary', () => {
    // Arrange
    // Act
    const wrapper = mount(ProgressBar3, {
      props: { progress: [50, 25, 25], total: 100 },
      global: { plugins: [i18n] },
    })

    // Assert
    expect(wrapper.attributes('aria-valuemax')).toBe('100')
    expect(wrapper.attributes('aria-valuenow')).toBe('50')
    expect(
      wrapper
        .findAll('[data-testid="progress-bar-segment"]')
        .map((segment) => segment.attributes('style')),
    ).toEqual(['width: 50%;', 'width: 25%;', 'width: 25%;'])
  })

  it('ignores invalid values and keeps stale totals from overflowing the bar', () => {
    // Arrange and act
    const wrapper = mount(ProgressBar3, {
      props: {
        progress: [-5, Number.NaN, 30],
        total: 10,
      },
      global: { plugins: [i18n] },
    })

    // Assert
    expect(wrapper.attributes('aria-valuemax')).toBe('30')
    expect(wrapper.findAll('[data-testid="progress-bar-segment"]')).toHaveLength(1)
    expect(wrapper.get('[data-testid="progress-bar-segment"]').attributes('style')).toBe(
      'width: 100%;',
    )
  })
})
