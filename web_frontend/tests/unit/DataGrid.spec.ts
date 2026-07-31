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
import { h, ref } from 'vue'

import DataGrid from '@/components/elements/DataGrid.vue'

interface TestRow {
  id: number
  name: string
}

describe('DataGrid', () => {
  it('renders each group with stable rows and resolved sort context', () => {
    // Arrange
    const groupSort = {
      field: ref('name'),
      direction: ref<'asc' | 'desc'>('asc'),
      toggle: vi.fn(),
      sorted: <T>(rows: T[]) => [...rows],
    }
    const forGroup = vi.fn(() => groupSort)

    // Act
    const wrapper = mount(DataGrid, {
      props: {
        groups: [['Administrators', [{ id: 1, name: 'Ada' }]]],
        sort: { forGroup },
        rowKey: (row: unknown) => (row as TestRow).id,
      },
      slots: {
        header: ({ groupSort: resolvedSort }) =>
          h('th', { 'data-sort-field': resolvedSort.field.value }, 'Name'),
        row: ({ row }) => h('td', (row as TestRow).name),
      },
    })

    // Assert
    expect(forGroup).toHaveBeenCalledWith('Administrators')
    expect(wrapper.get('caption').text()).toBe('Administrators')
    expect(wrapper.get('thead th').attributes('data-sort-field')).toBe('name')
    expect(wrapper.get('tbody td').text()).toBe('Ada')
  })

  it('allows consumers to replace the group heading and accessible caption', () => {
    // Arrange and act
    const wrapper = mount(DataGrid, {
      props: {
        groups: [['Internal', []]],
        sort: {
          forGroup: () => ({
            field: ref(''),
            direction: ref<'asc' | 'desc'>('asc'),
            toggle: vi.fn(),
            sorted: <T>(rows: T[]) => [...rows],
          }),
        },
        rowKey: () => 'unused',
      },
      slots: {
        'group-header': ({ groupName }) => h('h3', `Group ${groupName}`),
        caption: ({ groupName }) => `Members of ${groupName}`,
        header: () => h('th', 'Name'),
      },
    })

    // Assert
    expect(wrapper.get('.data-grid__group-header').text()).toBe('Group Internal')
    expect(wrapper.get('caption').text()).toBe('Members of Internal')
  })
})
