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
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'

import LayoutNavTop from '@/components/layout/LayoutNavTop.vue'
import { i18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'

import type { UserProfile } from '@/types/auth'

// Reset shared application state before each role-specific navigation check
beforeEach(() => {
  setActivePinia(createPinia())
})

async function createTestRouter(initialPath: string) {
  // Keep navigation tests independent from the application's lazy route modules
  const EmptyPage = { template: '<div />' }
  const testRouter = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/home', component: EmptyPage },
      { path: '/configuration/scopes', component: EmptyPage },
      { path: '/configuration/views', component: EmptyPage },
      { path: '/admin/users', component: EmptyPage },
      { path: '/admin/groups', component: EmptyPage },
      { path: '/api', component: EmptyPage },
      { path: '/installation', component: EmptyPage },
      { path: '/login', component: EmptyPage },
    ],
  })

  // Start each assertion on an explicit route before mounting the navigation
  await testRouter.push(initialPath)
  await testRouter.isReady()
  return testRouter
}

describe('LayoutNavTop role access', () => {
  it.each([
    ['admin', true],
    ['owner', true],
    ['user', false],
    ['demo', false],
  ] as const)(
    'shows Configuration for the %s role only when authorized',
    async (role, expected) => {
      // Arrange: authenticate one account with the role under test
      const authStore = useAuthStore()
      authStore.setProfile({
        id: '1',
        fullName: 'Example User',
        email: 'user@domain.tld',
        role: role satisfies UserProfile['role'],
      })
      const testRouter = await createTestRouter('/home')

      // Act: render the real navigation with its application dependencies
      const wrapper = mount(LayoutNavTop, {
        global: {
          plugins: [testRouter, [PrimeVue, { theme: 'none' }], ToastService, i18n],
        },
      })

      // Assert: the visible link follows the same admin-or-owner rule as the route
      expect(wrapper.find('[data-testid="nav-link-configuration"]').exists()).toBe(expected)

      // API navigation follows the same owner-or-admin rule as configuration
      expect(wrapper.find('[data-testid="nav-link-api"]').exists()).toBe(expected)

      // Demo accounts retain logout access but cannot request service credentials
      expect(wrapper.find('[data-testid="nav-action-generate-database-password"]').exists()).toBe(
        role !== 'demo',
      )
      expect(wrapper.find('[data-testid="nav-link-logout"]').exists()).toBe(true)

      wrapper.unmount()
    },
  )

  it('keeps parent navigation entries active across their flat child routes', async () => {
    // Arrange: admins can see both navigation sections covered by this behavior
    const authStore = useAuthStore()
    authStore.setProfile({
      id: '1',
      fullName: 'Example Admin',
      email: 'admin@domain.tld',
      role: 'admin',
    })
    const testRouter = await createTestRouter('/configuration/views')

    // Act: render the navigation on a non-default configuration route
    const wrapper = mount(LayoutNavTop, {
      global: {
        plugins: [testRouter, [PrimeVue, { theme: 'none' }], ToastService, i18n],
      },
    })

    // Assert: the section remains active even though the RouterLink targets scopes
    expect(wrapper.get('[data-testid="nav-link-configuration"]').attributes('aria-current')).toBe(
      'page',
    )

    // Assert: administration uses the same section-prefix behavior
    await testRouter.push('/admin/groups')
    expect(wrapper.get('[data-testid="nav-link-admin"]').attributes('aria-current')).toBe('page')

    wrapper.unmount()
  })
})
