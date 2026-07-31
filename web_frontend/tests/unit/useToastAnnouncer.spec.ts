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
import ToastEventBus from 'primevue/toasteventbus'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'

import { type ToastAnnouncer, useToastAnnouncer } from '@/composables/useToast'

function mountHook(): { announcer: ToastAnnouncer; unmount: () => void } {
  let announcer!: ToastAnnouncer
  const wrapper = mount(
    defineComponent({
      setup() {
        announcer = useToastAnnouncer()
        return () => h('div')
      },
    }),
  )
  return { announcer, unmount: () => wrapper.unmount() }
}

function emitToast(message: Record<string, unknown>): void {
  ToastEventBus.emit('add', message)
}

describe('useToastAnnouncer', () => {
  afterEach(() => {
    ToastEventBus.clear()
  })

  it('mirrors a success toast into the polite region', async () => {
    const { announcer } = mountHook()
    emitToast({ severity: 'success', summary: 'Saved' })
    await nextTick()
    expect(announcer.polite.value).toBe('Saved')
    expect(announcer.assertive.value).toBe('')
  })

  it('routes error severity to the assertive region', async () => {
    const { announcer } = mountHook()
    emitToast({ severity: 'error', summary: 'Boom', detail: 'Try again' })
    await nextTick()
    expect(announcer.assertive.value).toBe('Boom. Try again')
    expect(announcer.polite.value).toBe('')
  })

  it('re-announces identical consecutive messages by clearing first', async () => {
    const { announcer } = mountHook()
    emitToast({ severity: 'info', summary: 'Copied' })
    await nextTick()
    expect(announcer.polite.value).toBe('Copied')

    emitToast({ severity: 'info', summary: 'Copied' })

    // Synchronously reset to empty so the live region detects a change.
    expect(announcer.polite.value).toBe('')
    await nextTick()
    expect(announcer.polite.value).toBe('Copied')
  })

  it('ignores empty toasts', async () => {
    const { announcer } = mountHook()
    emitToast({ severity: 'success' })
    await nextTick()
    expect(announcer.polite.value).toBe('')
  })

  it('stops listening after unmount', async () => {
    const { announcer, unmount } = mountHook()
    unmount()
    emitToast({ severity: 'success', summary: 'Late' })
    await nextTick()
    expect(announcer.polite.value).toBe('')
  })
})
