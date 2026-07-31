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

import { type GlobalMountOptions, mount } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import SplitButton from 'primevue/splitbutton'

import SegmentedButton from '@/components/elements/SegmentedButton.vue'

const global: GlobalMountOptions = {
  plugins: [[PrimeVue, { theme: 'none' }]],
}

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

afterAll(() => {
  vi.unstubAllGlobals()
})

describe('SegmentedButton', () => {
  it('renders a real PrimeVue button and forwards attrs, classes, ARIA, and clicks', async () => {
    const onClick = vi.fn()
    const wrapper = mount(SegmentedButton, {
      props: { split: false },
      attrs: {
        class: 'caller-button',
        type: 'button',
        label: 'API Token',
        icon: 'pi pi-lock',
        severity: 'warn',
        'aria-label': 'Reset API token',
        'data-testid': 'ordinary-segmented',
        onClick,
      },
      global,
    })

    const button = wrapper.get('button')
    expect(button.attributes('aria-label')).toBe('Reset API token')
    expect(button.attributes('data-testid')).toBe('ordinary-segmented')

    await button.trigger('click')
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders a real PrimeVue split button and forwards its model, ARIA, and disabled state', () => {
    const items = [{ label: 'Add Network Scope', command: vi.fn() }]
    const wrapper = mount(SegmentedButton, {
      props: { split: true },
      attrs: {
        class: 'caller-split',
        label: 'Add Custom Scope',
        icon: 'pi pi-plus',
        severity: 'secondary',
        model: items,
        disabled: true,
        'data-testid': 'split-segmented',
        buttonProps: {
          'aria-label': 'Add Custom Scope',
          'data-testid': 'split-main',
        },
        menuButtonProps: {
          'aria-label': 'More scope types',
          'data-testid': 'split-menu',
        },
      },
      global,
    })

    const split = wrapper.getComponent(SplitButton)
    const main = wrapper.get('[data-testid="split-main"]')
    const menu = wrapper.get('[data-testid="split-menu"]')

    expect(split.props('model')).toEqual(items)
    expect(main.attributes('aria-label')).toBe('Add Custom Scope')
    expect(menu.attributes('aria-label')).toBe('More scope types')
    expect(main.attributes('disabled')).toBeDefined()
    expect(menu.attributes('disabled')).toBeDefined()
  })
})
