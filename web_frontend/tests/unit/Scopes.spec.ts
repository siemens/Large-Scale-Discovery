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
import ToastService from 'primevue/toastservice'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it, vi } from 'vitest'

import Scopes from '@/pages/configuration/Scopes.vue'

import type { ScopeResponse } from '@/types/scopes'
import type { Directive } from 'vue'

// Keep HTTP as the only mocked boundary while the page and store use their production wiring
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

// Capture the third-party tooltip contract without depending on visual positioning in JSDOM
const tooltipDirective: Directive<HTMLElement, string> = {
  mounted(element, binding) {
    element.dataset.tooltip = binding.value
    element.dataset.tooltipPosition = binding.modifiers.left ? 'left' : ''
  },
}

// Build the scope data exercised by the configuration grid
function scopeFixture(): ScopeResponse {
  return {
    id: 7,
    name: 'Example Scope',
    group_id: 3,
    group_name: 'Example Group',
    type: 'custom',
    size: 42,
    enabled: true,
    cycles: false,
    cycles_retention: -1,
    split_size: 2_048,
    created: '2026-07-30T08:00:00Z',
    created_by: 'owner@example.test',
    attributes: {},
    scan_settings: {} as ScopeResponse['scan_settings'],
    scan_agents: [],
    cycle: 1,
    cycle_started: '2026-07-29T08:00:00Z',
    cycle_done: 75,
    cycle_failed: 5,
    cycle_active: 10,
    cycle_queue: {},
    last_sync: '2026-07-30T07:00:00Z',
    secret: '',
    synchronization: false,
    connection: {
      host: 'database.example.test',
      port: 5_432,
      database: 'lsd_scope_0123456789',
    },
  }
}

// Mount the production page with its router and application services
async function mountPage() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/configuration/scopes', component: Scopes }],
  })
  await router.push('/configuration/scopes')
  await router.isReady()

  const wrapper = mount(Scopes, {
    global: {
      plugins: [createPinia(), [PrimeVue, { theme: 'none' }], ToastService, router],
      directives: { tooltip: tooltipDirective },
    },
  })
  await flushPromises()

  return wrapper
}

describe('Scopes', () => {
  it('shows the physical database name in a left-positioned scope-name tooltip', async () => {
    // Arrange
    mocks.get.mockResolvedValue({
      data: {
        body: {
          scopes: [scopeFixture()],
          allow_custom: false,
          allow_network: false,
          allow_asset: false,
        },
      },
    })

    // Act
    const wrapper = await mountPage()
    const scopeName = wrapper.get('[data-testid="sl-scope-name"]')

    // Assert
    expect(scopeName.text()).toBe('Example Scope')
    expect(scopeName.attributes('data-tooltip')).toBe('lsd_scope_0123456789')
    expect(scopeName.attributes('data-tooltip-position')).toBe('left')
  })
})
