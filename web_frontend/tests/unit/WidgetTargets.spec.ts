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

import { mount, type VueWrapper } from '@vue/test-utils'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import { defineComponent, h } from 'vue'

import WidgetTargets from '@/components/widgets/WidgetTargets.vue'
import { i18n } from '@/i18n'

import type { ScopeTarget } from '@/types/scopes'

// WidgetTargetsHandle describes the public target-list methods used by its parent form
interface WidgetTargetsHandle {
  getItems: () => ScopeTarget[]
}

// MountOptions controls the interaction guards exercised by these tests
interface MountOptions {
  isModalOpen?: () => boolean
}

// Replace the heavy virtualized table while preserving the methods used by the widget
const DataTableStub = defineComponent({
  name: 'DataTable',
  setup(_, { expose }) {
    expose({
      clearAllFilters: vi.fn(),
      getVisibleRowCapacity: () => 1,
      isScrolledToBottom: () => false,
      nudgeScrollAboveBottom: vi.fn(),
      preserveScrollPosition: vi.fn(),
      resetScrollPosition: vi.fn(),
      scrollToBottomPosition: vi.fn(),
    })
    return () => h('div', { class: 'data-table-stub' })
  },
})

// Mount the real widget so clipboard behavior is verified at its owning boundary
function mountWidget(options: MountOptions = {}): VueWrapper {
  return mount(WidgetTargets, {
    props: options,
    global: {
      plugins: [i18n, [PrimeVue, { theme: 'none' }], ToastService],
      stubs: { DataTable: DataTableStub },
      directives: { tooltip: () => undefined },
    },
  })
}

// Dispatch a browser-like paste event with caller-provided plain text
function paste(wrapper: VueWrapper, text: string, target?: HTMLElement): ClipboardEvent {
  const event = new Event('paste', { bubbles: true, cancelable: true }) as ClipboardEvent
  Object.defineProperty(event, 'clipboardData', {
    value: { getData: () => text },
  })
  ;(target ?? wrapper.get('[data-testid="scope-targets-grid"]').element).dispatchEvent(event)
  return event
}

// Return only populated targets because the editor deliberately retains blank input rows
function populatedTargets(wrapper: VueWrapper): ScopeTarget[] {
  const component = wrapper.vm as unknown as WidgetTargetsHandle
  return component.getItems().filter((target) => target.input?.trim())
}

describe('WidgetTargets clipboard behavior', () => {
  it('maps title headers and replaces the current target rows', () => {
    const wrapper = mountWidget()

    // Paste two Excel-style rows under translated column titles
    const event = paste(wrapper, 'Scan Input\tEnabled\n192.0.2.1\ttrue\n198.51.100.2\tfalse')

    // Verify that the widget owns the paste and preserves typed field values
    expect(event.defaultPrevented).toBe(true)
    expect(populatedTargets(wrapper)).toMatchObject([
      { input: '192.0.2.1', enabled: true },
      { input: '198.51.100.2', enabled: false },
    ])
  })

  it('accepts backend field headers and headerless positional rows', () => {
    const wrapper = mountWidget()

    // Stable field names are accepted when data comes from another application
    paste(wrapper, 'input\tenabled\n192.0.2.3\tfalse')
    expect(populatedTargets(wrapper)).toMatchObject([{ input: '192.0.2.3', enabled: false }])

    // Headerless values map directly to the configured data-column order
    paste(wrapper, '198.51.100.4\ttrue')
    expect(populatedTargets(wrapper)).toMatchObject([{ input: '198.51.100.4', enabled: true }])
  })

  it('normalizes line endings and ignores trailing blank rows', () => {
    const wrapper = mountWidget()

    // Windows clipboard text commonly ends in an additional CRLF
    paste(wrapper, 'Scan Input\tEnabled\r\n192.0.2.5\ttrue\r\n')
    expect(populatedTargets(wrapper)).toMatchObject([{ input: '192.0.2.5', enabled: true }])
  })

  it('leaves ordinary text-input paste behavior untouched', () => {
    const wrapper = mountWidget()
    const input = document.createElement('input')
    wrapper.get('[data-testid="scope-targets-grid"]').element.appendChild(input)

    // A cell editor or filter input remains the owner of its paste event
    const event = paste(wrapper, '192.0.2.6', input)
    expect(event.defaultPrevented).toBe(false)
    expect(populatedTargets(wrapper)).toEqual([])
  })
})

describe('WidgetTargets paste shortcut focus', () => {
  it('focuses the widget for an unclaimed Ctrl+V shortcut', () => {
    const wrapper = mountWidget()
    const root = wrapper.get('[data-testid="scope-targets-grid"]').element as HTMLElement
    const focus = vi.spyOn(root, 'focus')

    // Global paste shortcuts should route the following paste event to the widget
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'v', ctrlKey: true }))
    expect(focus).toHaveBeenCalledOnce()
  })

  it('does not steal focus from text fields, dialogs, or PrimeVue overlays', () => {
    const wrapper = mountWidget({ isModalOpen: () => true })
    const root = wrapper.get('[data-testid="scope-targets-grid"]').element as HTMLElement
    const focus = vi.spyOn(root, 'focus')

    // A consumer-owned dialog blocks the global shortcut
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'v', ctrlKey: true }))
    expect(focus).not.toHaveBeenCalled()

    // Native text input focus also takes priority over the target widget
    const input = document.createElement('input')
    document.body.appendChild(input)
    input.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'v', ctrlKey: true }))
    expect(focus).not.toHaveBeenCalled()
  })

  it('removes its global shortcut listener when unmounted', () => {
    const wrapper = mountWidget()
    const root = wrapper.get('[data-testid="scope-targets-grid"]').element as HTMLElement
    const focus = vi.spyOn(root, 'focus')
    wrapper.unmount()

    // An unmounted widget must not retain a document-level event listener
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'v', ctrlKey: true }))
    expect(focus).not.toHaveBeenCalled()
  })
})
