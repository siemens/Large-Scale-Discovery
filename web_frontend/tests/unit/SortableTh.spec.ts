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

import SortableTh from '@/components/elements/SortableTh.vue'

describe('SortableTh', () => {
  it('exposes the active direction and emits its field from a native button', async () => {
    // Arrange
    const wrapper = mount(SortableTh, {
      props: { field: 'name', active: 'name', direction: 'desc' },
      slots: { default: 'Name' },
    })

    // Act
    await wrapper.get('button').trigger('click')

    // Assert
    expect(wrapper.attributes('aria-sort')).toBe('descending')
    expect(wrapper.emitted('toggle')).toEqual([['name']])
  })

  it('shows the neutral state for an inactive column', () => {
    // Arrange
    const wrapper = mount(SortableTh, {
      props: { field: 'name', active: 'created', direction: 'asc' },
      slots: { default: 'Name' },
    })

    // Assert
    expect(wrapper.attributes('aria-sort')).toBe('none')
  })
})
