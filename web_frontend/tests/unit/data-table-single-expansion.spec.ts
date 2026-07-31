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

/* eslint-disable vue/one-component-per-file -- focused inline stubs exercise the wrapper contract */
import { mount, shallowMount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'

import DataTable from '@/components/datatable/DataTable.vue'
import { i18n } from '@/i18n'

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))

const rows = [
  { id: 1, name: 'First' },
  { id: 2, name: 'Second' },
]
const columns = [{ field: 'name', header: 'Name' }]

const PrimeDataTableStub = defineComponent({
  name: 'PrimeDataTableStub',
  inheritAttrs: false,
  props: {
    expandedRows: { type: [Object, Array], default: undefined },
    value: { type: Array, default: () => [] },
    pt: { type: Object, default: () => ({}) },
  },
  emits: ['update:expandedRows', 'row-click'],
  template: `
    <table>
      <tbody>
        <tr
          v-for="(row, index) in value"
          :key="row.id"
          data-testid="data-row"
          @click="$emit('row-click', { originalEvent: $event, data: row, index })"
        >
          <td data-testid="row-space" />
          <td>
            <span data-testid="row-content">{{ row.name }}</span>
            <input data-testid="row-editor" :value="row.name" />
          </td>
          <td><button data-testid="row-action">Action</button></td>
        </tr>
        <tr class="p-datatable-row-expansion">
          <td
            :ref="pt.rowExpansionCell?.ref"
            :class="pt.rowExpansionCell?.class"
            colspan="3"
          >
            <slot name="expansion" :data="value[0]" />
          </td>
        </tr>
      </tbody>
    </table>
  `,
})

const PrimeDataTableReorderStub = defineComponent({
  name: 'PrimeDataTableReorderStub',
  inheritAttrs: false,
  emits: ['columnReorder'],
  template: '<div><slot /></div>',
})

const ColumnStub = defineComponent({
  name: 'ColumnStub',
  inheritAttrs: false,
  props: {
    columnKey: { type: String, default: undefined },
    reorderableColumn: { type: Boolean, default: true },
  },
  template: `
    <div
      data-testid="column-stub"
      :data-column-key="columnKey"
      :data-reorderable="String(reorderableColumn)"
    />
  `,
})

function mountTable(expandedRows: Record<string, boolean> | typeof rows) {
  return shallowMount(DataTable, {
    props: { rows, columns, expandable: true, expandedRows },
    global: {
      plugins: [i18n],
      stubs: { DataTable: PrimeDataTableStub },
    },
  })
}

describe('DataTable single-row expansion', () => {
  it('replaces the previously expanded keyed row in one update', async () => {
    const wrapper = mountTable({ 1: true })
    const table = wrapper.getComponent(PrimeDataTableStub)

    table.vm.$emit('update:expandedRows', { 1: true, 2: true })
    expect(wrapper.emitted('update:expandedRows')?.at(-1)?.[0]).toEqual({ 2: true })

    await wrapper.setProps({ expandedRows: { 2: true } })
    table.vm.$emit('update:expandedRows', { 1: true, 2: true })
    expect(wrapper.emitted('update:expandedRows')?.at(-1)?.[0]).toEqual({ 1: true })

    table.vm.$emit('update:expandedRows', {})
    expect(wrapper.emitted('update:expandedRows')?.at(-1)?.[0]).toEqual({})
  })

  it('normalizes initial and updated row-array models to one row', () => {
    const wrapper = mountTable(rows)
    const table = wrapper.getComponent(PrimeDataTableStub)

    expect(table.props('expandedRows')).toEqual([rows[1]])

    table.vm.$emit('update:expandedRows', rows)
    expect(wrapper.emitted('update:expandedRows')?.at(-1)?.[0]).toEqual([rows[1]])
  })

  it('does not toggle details from row clicks', async () => {
    const wrapper = mountTable({})

    await wrapper.find('[data-testid="row-space"]').trigger('click')
    await wrapper.find('[data-testid="row-action"]').trigger('click')
    await wrapper.find('[data-testid="row-editor"]').trigger('click')
    await wrapper.find('[data-testid="row-content"]').trigger('click')

    expect(wrapper.emitted('update:expandedRows')).toBeUndefined()
  })

  it('uses a real leading cell and spans the detail cell across the remaining columns', () => {
    const wrapper = mountTable({ 1: true })
    const expansionRow = wrapper.get('tr.p-datatable-row-expansion')
    const cells = expansionRow.findAll(':scope > td')

    expect(cells).toHaveLength(2)
    expect(cells[0]?.attributes('aria-hidden')).toBe('true')
    expect(cells[1]?.attributes('colspan')).toBe('2')
  })

  it('splits the expansion row rendered by PrimeVue into sibling table cells', () => {
    const wrapper = mount(DataTable, {
      props: { rows, columns, expandable: true, expandedRows: { 1: true } },
      slots: { expansion: '<span data-testid="row-details">Details</span>' },
      global: {
        plugins: [i18n, PrimeVue],
        directives: { tooltip: {} },
      },
    })
    const expansionRow = wrapper.get('tr.p-datatable-row-expansion')
    const cells = expansionRow.findAll(':scope > td')

    expect(cells).toHaveLength(2)
    expect(cells[0]?.attributes('aria-hidden')).toBe('true')
    expect(cells[1]?.attributes('colspan')).toBe('1')
    expect(cells[1]?.find('[data-testid="row-details"]').exists()).toBe(true)
  })

  it('leaves an empty loading table quiet instead of rendering a local spinner', () => {
    const wrapper = mount(DataTable, {
      props: { rows: [], columns, loading: true },
      global: {
        plugins: [i18n, PrimeVue],
        directives: { tooltip: {} },
      },
    })

    expect(wrapper.find('[data-testid="adt-loading"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="adt-empty"]').exists()).toBe(false)
  })

  it('keeps action columns fixed and rejects drops beyond a trailing action column', async () => {
    const actionColumns = [
      ...columns,
      { field: 'status', header: 'Status' },
      { field: 'actions', header: 'Actions', formatter: 'actions' as const },
    ]
    const mountReorderTable = (actionsPosition: 'start' | 'end') =>
      shallowMount(DataTable, {
        props: { rows, columns: actionColumns, actionsPosition, reorderableColumns: true },
        global: {
          plugins: [i18n],
          stubs: { DataTable: PrimeDataTableReorderStub, Column: ColumnStub },
        },
      })

    for (const position of ['start', 'end'] as const) {
      const wrapper = mountReorderTable(position)
      expect(
        wrapper.get('[data-column-key="__actions_actions"]').attributes('data-reorderable'),
      ).toBe('false')
      wrapper.unmount()
    }

    const wrapper = mountReorderTable('end')
    let table = wrapper.getComponent(PrimeDataTableReorderStub)
    const beforeTrailingDrop = table.vm
    table.vm.$emit('columnReorder', { dragIndex: 0, dropIndex: 2 })
    await wrapper.vm.$nextTick()
    table = wrapper.getComponent(PrimeDataTableReorderStub)
    expect(table.vm).not.toBe(beforeTrailingDrop)

    const beforeValidDrop = table.vm
    table.vm.$emit('columnReorder', { dragIndex: 0, dropIndex: 1 })
    await wrapper.vm.$nextTick()
    table = wrapper.getComponent(PrimeDataTableReorderStub)
    expect(table.vm).toBe(beforeValidDrop)

    const beforeActionDrag = table.vm
    table.vm.$emit('columnReorder', { dragIndex: 2, dropIndex: 0 })
    await wrapper.vm.$nextTick()
    expect(wrapper.getComponent(PrimeDataTableReorderStub).vm).not.toBe(beforeActionDrag)
  })
})
