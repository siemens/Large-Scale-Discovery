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

import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { PropType } from 'vue'
import { defineComponent, h } from 'vue'
import { createI18n } from 'vue-i18n'

import LayoutFeedback from '@/components/layout/LayoutFeedback.vue'

const mocks = vi.hoisted(() => ({
  post: vi.fn(),
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('@/api/http', () => ({
  default: {
    post: mocks.post,
  },
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: mocks.success,
    error: mocks.error,
  }),
}))

const DialogFormStub = defineComponent({
  name: 'DialogForm',
  props: {
    submitHandler: {
      type: Function as PropType<() => Promise<void> | void>,
      required: true,
    },
  },
  setup(props, { slots }) {
    return () =>
      h(
        'form',
        {
          'data-testid': 'feedback-form',
          onSubmit: (event: Event) => {
            event.preventDefault()
            void props.submitHandler()
          },
        },
        [slots.default?.(), h('button', { type: 'submit' }, 'Submit')],
      )
  },
})

function mountFeedback() {
  const i18n = createI18n({
    legacy: false,
    locale: 'en-US',
  })

  return mount(LayoutFeedback, {
    props: {
      visible: true,
      userDisplayName: 'Example User',
    },
    global: {
      plugins: [createPinia(), i18n, [PrimeVue, { theme: 'none' }]],
      stubs: {
        DialogForm: DialogFormStub,
      },
    },
  })
}

describe('LayoutFeedback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.post.mockResolvedValue({ data: { body: {} } })
  })

  it('submits feedback unchanged and resets after success', async () => {
    // Arrange — mount with ordinary feedback content
    const subject = 'Example subject'
    const message = 'Example message'
    const wrapper = mountFeedback()

    // Act — complete and submit the feedback form
    await wrapper.get('[data-testid="feedback-subject"]').setValue(subject)
    await wrapper.get('[data-testid="feedback-message"]').setValue(message)
    await wrapper.get('[data-testid="feedback-form"]').trigger('submit')
    await flushPromises()

    // Assert — submit the original content and reset the successful dialog
    expect(mocks.post).toHaveBeenCalledWith(
      '/user/feedback',
      { subject, message },
      { errorToastMessage: expect.any(String) },
    )
    expect(wrapper.emitted('update:visible')).toContainEqual([false])
    expect(
      (wrapper.get('[data-testid="feedback-subject"]').element as HTMLInputElement).value,
    ).toBe('')
    expect(
      (wrapper.get('[data-testid="feedback-message"]').element as HTMLTextAreaElement).value,
    ).toBe('')
    expect(mocks.success).toHaveBeenCalledOnce()
  })

  it('rejects empty required fields without submitting or closing', async () => {
    // Arrange — mount with the initially empty form
    const wrapper = mountFeedback()

    // Act — submit without filling either required field
    await wrapper.get('[data-testid="feedback-form"]').trigger('submit')
    await flushPromises()

    // Assert — retain the dialog and expose both validation messages
    expect(mocks.post).not.toHaveBeenCalled()
    expect(wrapper.findAll('[role="alert"]')).toHaveLength(2)
    expect(wrapper.emitted('update:visible')).toBeUndefined()
  })
})
