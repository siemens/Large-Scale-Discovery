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
import { defineComponent, h, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import FormScopeAddCustom from '@/components/forms/FormScopeAddCustom.vue'
import LayoutForm from '@/components/layout/LayoutForm.vue'

interface CustomFormHandle {
  submit: () => Promise<void>
}

// Keep HTTP as the only mocked boundary while the form and stores use their production wiring
const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/http', () => ({
  default: {
    get: mocks.get,
    post: mocks.post,
  },
}))

// WidgetTargets is a heavy grid and is unrelated to the scalar scope fields under test
// eslint-disable-next-line vue/one-component-per-file
const WidgetTargetsStub = defineComponent({
  setup(_, { expose }) {
    expose({
      clearTargetScanData: () => undefined,
      getItems: () => [],
      setItems: () => undefined,
    })
    return () => h('div')
  },
})

// FormHost reproduces the production LayoutForm submission boundary
// eslint-disable-next-line vue/one-component-per-file
const FormHost = defineComponent({
  components: { FormScopeAddCustom, LayoutForm },
  setup() {
    const form = ref<CustomFormHandle | null>(null)
    const submit = (): Promise<void> | undefined => form.value?.submit()
    return { form, submit }
  },
  template: `
    <LayoutForm
      :footer="false"
      text-submit-label="Save"
      :submit-handler="submit"
    >
      <FormScopeAddCustom ref="form" />
    </LayoutForm>
  `,
})

// Mount the production form boundary with only its heavy target grid replaced
function mountForm() {
  return mount(FormHost, {
    global: {
      plugins: [createPinia(), [PrimeVue, { theme: 'none' }]],
      stubs: {
        Select: true,
        WidgetTargets: WidgetTargetsStub,
      },
    },
  })
}

describe('FormScopeAddCustom', () => {
  beforeEach(() => {
    // Reset API behavior for each independently mounted form
    mocks.get.mockReset()
    mocks.post.mockReset()
    mocks.get.mockImplementation((url: string) => {
      if (url === '/groups') {
        return Promise.resolve({
          data: {
            body: {
              groups: [
                {
                  id: 7,
                  name: 'Example Group',
                  created: '2026-01-01T00:00:00Z',
                  created_by: 'user@domain.tld',
                  db_server_id: 1,
                  max_scopes: -1,
                  max_views: -1,
                  max_targets: -1,
                  max_owners: -1,
                  allow_custom: true,
                  allow_network: true,
                  allow_asset: true,
                  ownerships: [],
                },
              ],
            },
          },
        })
      }
      return Promise.resolve({
        data: {
          body: {
            scopes: [],
            allow_custom: true,
            allow_network: true,
            allow_asset: true,
          },
        },
      })
    })
    mocks.post.mockResolvedValue({ data: { message: 'Scope saved.', body: { warnings: [] } } })
  })

  it('submits the selected network segment size', async () => {
    // Arrange
    const wrapper = mountForm()
    await flushPromises()
    await wrapper.get('[data-testid="sac-name"]').setValue('Example Scope')

    // Act
    await wrapper.get('[aria-label="Decrease network size"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    // Assert
    expect(mocks.post).toHaveBeenCalledWith(
      '/scope/update/custom',
      expect.objectContaining({
        group_id: 7,
        split_size: 1_024,
      }),
    )
  })

  it('forces and disables the maximum network segment size when OT discovery is enabled', async () => {
    // Arrange
    const wrapper = mountForm()
    await flushPromises()
    await wrapper.get('[data-testid="sac-name"]').setValue('Example Scope')
    await wrapper.get('[aria-label="Decrease network size"]').trigger('click')

    // Act
    await wrapper.get('[data-testid="sac-ot"] input').setValue(true)
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    // Assert
    expect(wrapper.get('[data-testid="sac-network-size-value"]').text()).toBe('2048 IPs')
    expect(wrapper.get('[aria-label="Decrease network size"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[aria-label="Increase network size"]').attributes('disabled')).toBeDefined()
    expect(mocks.post).toHaveBeenCalledWith(
      '/scope/update/custom',
      expect.objectContaining({
        ot: true,
        split_size: 2_048,
      }),
    )
  })
})
