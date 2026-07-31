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

import { config, mount } from '@vue/test-utils'
import ToastService from 'primevue/toastservice'
import { afterEach, describe, expect, it, vi } from 'vitest'

import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import { i18n } from '@/i18n'

const plugins = config.global.plugins.includes(i18n) ? [ToastService] : [i18n, ToastService]

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('CodeSnippet', () => {
  it('renders a caller-owned footer link', () => {
    // Arrange
    const wrapper = mount(CodeSnippet, {
      props: { text: 'SELECT address FROM assets', testId: 'query' },
      slots: {
        footer:
          '<a href="https://documentation.example/query" rel="noopener noreferrer">Query language documentation</a>',
      },
      global: {
        plugins,
        directives: { tooltip: {} },
      },
    })

    // Act
    const footerLink = wrapper.get('a')

    // Assert
    expect(footerLink.attributes('href')).toBe('https://documentation.example/query')
  })

  it('copies only the snippet text when footer content is present', async () => {
    // Arrange
    const writeText = vi.fn<(text: string) => Promise<void>>().mockResolvedValue()
    vi.stubGlobal('navigator', { clipboard: { writeText } })
    const wrapper = mount(CodeSnippet, {
      props: { text: 'SELECT address FROM assets', testId: 'query' },
      slots: {
        footer: '<span>Query language documentation</span>',
      },
      global: {
        plugins,
        directives: { tooltip: {} },
      },
    })

    // Act
    await wrapper.get('button[aria-label="Copy to clipboard"]').trigger('click')

    // Assert
    expect(writeText.mock.calls).toEqual([['SELECT address FROM assets']])
  })
})
