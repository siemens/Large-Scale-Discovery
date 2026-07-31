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
import { createPinia, setActivePinia } from 'pinia'

import UnsavedChangesDialog from '@/components/elements/UnsavedDialog.vue'
import { i18n } from '@/i18n'
import { useUnsavedChangesStore } from '@/stores/unsavedChanges'

const routeGuard = vi.hoisted(() => ({
  callback: null as null | (() => boolean | Promise<boolean>),
  remove: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    beforeEach: (callback: () => boolean | Promise<boolean>) => {
      routeGuard.callback = callback
      return routeGuard.remove
    },
  }),
}))

describe('UnsavedChangesDialog', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    routeGuard.callback = null
    routeGuard.remove.mockClear()
  })

  it('allows clean navigation without opening a prompt', () => {
    // Arrange
    const wrapper = mount(UnsavedChangesDialog, {
      global: { plugins: [i18n], stubs: { DialogConfirm: true } },
    })

    // Act and assert
    expect(routeGuard.callback?.()).toBe(true)
    expect(wrapper.getComponent({ name: 'DialogConfirm' }).props('visible')).toBe(false)
  })

  it('resolves dirty navigation from the confirmation result', async () => {
    // Arrange
    const wrapper = mount(UnsavedChangesDialog, {
      global: { plugins: [i18n], stubs: { DialogConfirm: true } },
    })
    useUnsavedChangesStore().setDirty('form', true)

    // Act
    const decision = routeGuard.callback?.()
    await wrapper.vm.$nextTick()
    wrapper.getComponent({ name: 'DialogConfirm' }).vm.$emit('cancel')

    // Assert
    await expect(decision).resolves.toBe(false)
  })

  it('removes the application guard when the app shell unmounts', () => {
    const wrapper = mount(UnsavedChangesDialog, {
      global: { plugins: [i18n], stubs: { DialogConfirm: true } },
    })

    wrapper.unmount()

    expect(routeGuard.remove).toHaveBeenCalledOnce()
  })
})
