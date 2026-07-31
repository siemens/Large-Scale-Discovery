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
import ToastService from 'primevue/toastservice'
import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'

import DataTable from '@/components/datatable/DataTable.vue'

function mountEmptyTable(hideWhenEmpty: boolean) {
  const i18n = createI18n({ legacy: false, locale: 'en-US', messages: { 'en-US': {} } })
  return mount(DataTable, {
    props: {
      rows: [],
      columns: [],
      loading: true,
      loadingMessage: 'Loading jobs...',
      emptyMessage: 'No jobs found.',
      hideWhenEmpty,
    },
    global: {
      plugins: [i18n, ToastService],
    },
  })
}

describe('DataTable empty-state handling', () => {
  it('replaces the table with the shared placeholder when an empty source opts in', async () => {
    // Arrange — begin with the initial loading placeholder
    const wrapper = mountEmptyTable(true)
    expect(wrapper.text()).toContain('Loading jobs...')

    // Act — settle the initial request without adding rows
    await wrapper.setProps({ loading: false })

    // Assert — the message remains outside PrimeVue's table markup
    expect(wrapper.text()).toContain('No jobs found.')
    expect(wrapper.find('.p-datatable').exists()).toBe(false)
  })

  it('keeps the regular table empty row when external replacement is disabled', async () => {
    // Arrange — use the default in-table empty-state behavior
    const wrapper = mountEmptyTable(false)

    // Act — settle the initial request without adding rows
    await wrapper.setProps({ loading: false })

    // Assert — existing consumers retain their headers and table container
    expect(wrapper.text()).toContain('No jobs found.')
    expect(wrapper.find('.p-datatable').exists()).toBe(true)
  })
})
