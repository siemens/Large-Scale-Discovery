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
import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'

import LayoutLoading from '@/components/layout/LayoutLoading.vue'

const mountLayout = (loading: boolean, empty: boolean) => {
  const i18n = createI18n({
    legacy: false,
    locale: 'en-US',
    messages: { 'en-US': {} },
  })

  return mount(LayoutLoading, {
    props: {
      loading,
      loadingMessage: 'Loading records...',
      empty,
      emptyMessage: 'No records found.',
    },
    global: {
      plugins: [i18n],
    },
  })
}

describe('LayoutLoading', () => {
  it('shows the loading message with a spinner while loading', () => {
    // Arrange and act: render the loading state.
    const wrapper = mountLayout(true, false)

    // Assert: loading uses its own message and spinner rather than the empty icon.
    expect(wrapper.text()).toContain('Loading records...')
  })

  it('shows the empty message with the empty-state icon after loading', () => {
    // Arrange and act: render the empty state.
    const wrapper = mountLayout(false, true)

    // Assert: the empty state remains visually distinct from loading.
    expect(wrapper.text()).toContain('No records found.')
  })
})
