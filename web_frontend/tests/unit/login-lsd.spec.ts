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
import { createPinia, setActivePinia } from 'pinia'
import PrimeVue from 'primevue/config'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'

import LoginPage from '@/pages/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

// Replace visible toast delivery while retaining the page's production flow
const toastMock = vi.hoisted(() => ({
  error: vi.fn(),
  success: vi.fn(),
  info: vi.fn(),
  warn: vi.fn(),
}))

vi.mock('@/composables/useToast', () => ({ useToast: () => toastMock }))

// Mount one independently configured login page for each scenario
async function mountLogin(options: { development?: boolean; entryUrl?: string } = {}) {
  // Isolate the stores used by this login instance
  const pinia = createPinia()
  setActivePinia(pinia)

  // Control authentication outcomes at the page's store boundary
  const auth = useAuthStore()
  auth.resolveAuthenticator = vi.fn().mockResolvedValue(options.entryUrl ?? '')
  auth.login = vi.fn().mockResolvedValue(undefined)

  // Resolve public settings without issuing a network request
  const settings = useSettingsStore()
  settings.isDevelopmentLogin = options.development ?? false
  settings.isLoaded = true
  settings.fetchSettings = vi.fn().mockResolvedValue(undefined)

  // Provide the destinations used after successful authentication
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'login', component: { template: '<div />' } },
      { path: '/home', name: 'home', component: { template: '<div />' } },
    ],
  })
  await router.push('/login')
  await router.isReady()

  // Render with the same framework plugins as the application
  const wrapper = mount(LoginPage, {
    global: {
      plugins: [pinia, router, [PrimeVue, { theme: 'none' }]],
    },
  })
  await flushPromises()

  // Expose the observable dependencies and rendered page
  return { auth, router, wrapper }
}

describe('Large-Scale Discovery login', () => {
  beforeEach(() => {
    // Reset browser and mock state between independent login attempts
    vi.clearAllMocks()
    window.history.replaceState(null, '', '/')
    sessionStorage.clear()
  })

  it('keeps the username in one persistent form when credentials login reveals the password', async () => {
    // Arrange — mount the e-mail step and retain the browser-facing form element
    const { wrapper } = await mountLogin()
    const loginForm = wrapper.get('[data-testid="login-form"]')
    const emailInput = wrapper.get('[data-testid="login-email-input"]')

    // Act — resolve the submitted address to credentials authentication
    await emailInput.setValue('user@example.test')
    await loginForm.trigger('submit')
    await flushPromises()

    // Assert — password managers see the same form, username, and current password together
    expect(wrapper.findAll('form')).toHaveLength(1)
    expect(wrapper.get('[data-testid="login-form"]').element).toBe(loginForm.element)
    expect(emailInput.attributes('type')).toBe('email')
    expect(emailInput.attributes('autocomplete')).toBe('username')
    expect(wrapper.get('[data-testid="login-email-input"]').element).toBe(emailInput.element)
    expect(wrapper.get('[data-testid="login-password-input"]').attributes('autocomplete')).toBe(
      'current-password',
    )
  })

  it('does not pass a password to development login', async () => {
    // Arrange — mount address-only development authentication
    const { auth, router, wrapper } = await mountLogin({ development: true })
    await wrapper.get('[data-testid="login-email-input"]').setValue('dev@example.test')

    // Act — submit the address-only development login
    await wrapper.get('[data-testid="login-form"]').trigger('submit')
    await flushPromises()

    // Assert — development authentication never receives a password value
    expect(auth.login).toHaveBeenCalledWith('dev@example.test', undefined)
    expect(wrapper.find('[data-testid="login-password-input"]').exists()).toBe(false)
    expect(router.currentRoute.value.fullPath).toBe('/home')
  })

  it('hands an SSO entry URL to the browser without rendering or sending a password', async () => {
    // Arrange — resolve the submitted account to an external authenticator
    const { auth, wrapper } = await mountLogin({ entryUrl: '#lsd-sso' })
    await wrapper.get('[data-testid="login-email-input"]').setValue('sso@example.test')

    // Act — submit the address-only SSO discovery step
    await wrapper.get('[data-testid="login-form"]').trigger('submit')
    await flushPromises()

    // Assert — the external flow starts without a local password field or login request
    expect(window.location.hash).toBe('#lsd-sso')
    expect(wrapper.find('[data-testid="login-password-input"]').exists()).toBe(false)
    expect(auth.login).not.toHaveBeenCalled()
  })
})
