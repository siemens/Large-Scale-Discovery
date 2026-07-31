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

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { createI18n } from 'vue-i18n'

import http from '@/api/http'
import { clearSession, getAccessToken, hasValidSession } from '@/api/session'
import { useAsyncState } from '@/composables/useAsyncState'
import { useAgentsStore } from '@/stores/agents'
import { useEventsStore } from '@/stores/events'
import { useSwaggerStore } from '@/stores/swagger'
import { useUsersStore } from '@/stores/users'
import { useViewsStore } from '@/stores/views'

import type {
  AuthenticatorRequest,
  AuthenticatorResponse,
  LoginRequest,
  RegisterRequest,
  UserDetailsResponse,
  UserProfile,
} from '@/types/auth'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    auth: {
      login: {
        actions: {
          submit: {
            error: {
              invalidCredentials: 'Invalid credentials.',
              missingToken: 'Authentication response did not contain an access token.',
              startFailed: 'Could not start authentication.',
            },
          },
        },
      },
      register: {
        actions: {
          submit: {
            error: 'Registration failed. Please check your details and try again.',
          },
        },
      },
    },
  },
}

// Create a tiny translator for messages owned by this store
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

// Convert the backend account response into the small profile used by the UI
function toProfile(body: UserDetailsResponse): UserProfile {
  // Demo status overrides stored privileges because demo sessions must remain read-only
  const role: UserProfile['role'] = body.demo
    ? 'demo'
    : body.admin
      ? 'admin'
      : body.owner
        ? 'owner'
        : 'user'

  // Copy account data and add the display name expected by consumers
  return {
    id: String(body.id),
    fullName: `${body.name} ${body.surname}`.trim(),
    email: body.email,
    role,
  }
}

/** Owns the authenticated user's session, profile, and login/register/logout flow. */
export const useAuthStore = defineStore('auth', () => {
  // Store the authenticated account profile
  const profile = ref<UserProfile | null>(null)

  // Tell route guards whether authentication has completed successfully
  const isAuthenticated = ref(false)

  // Track authentication steps independently so the UI can show precise progress
  const resolveState = useAsyncState()

  // Track the local-login request
  const loginState = useAsyncState()

  // Track account-registration requests
  const registerState = useAsyncState()

  // Track loading the authenticated profile
  const profileState = useAsyncState()

  // Tell administration controls whether the account is an administrator
  const isAdmin = computed(() => profile.value?.role === 'admin')

  // Tell configuration controls whether the account owns scopes
  const isOwner = computed(() => profile.value?.role === 'owner')

  // Tell presentation guards whether the account is a demo user
  const isDemo = computed(() => profile.value?.role === 'demo')

  // Present resolver and login failures through one login-facing error
  const loginError = computed(() => loginState.error.value ?? resolveState.error.value)

  // Clear every store whose contents belong to one account
  function resetUserDashboardData(): void {
    // Clear every account-specific dashboard registry
    useViewsStore().reset()
    useAgentsStore().reset()
    useEventsStore().reset()
    useUsersStore().reset()
    useSwaggerStore().reset()
  }

  // Publish a profile and clear stale dashboard data after account changes
  function setProfile(profileValue: UserProfile): void {
    // Reset cached data only when a different account becomes active
    if (profile.value?.id !== profileValue.id) resetUserDashboardData()

    // Mark the new profile as authenticated
    profile.value = profileValue
    isAuthenticated.value = true
  }

  // Remove the active profile and all account-specific dashboard data
  function clearProfile(): void {
    // Clear dependent stores before dropping the account identity
    resetUserDashboardData()

    // Return authentication state to its signed-out values
    profile.value = null
    isAuthenticated.value = false
  }

  // Resolve whether an address uses SSO or credentials login
  async function resolveAuthenticator(email: string): Promise<string> {
    // Clear an earlier login failure before resolving a new address
    loginState.reset()

    // Ask the backend which authentication method owns the address
    return resolveState.run(
      async () => {
        // Ask the backend which authentication method is configured for this address
        const response = await http.post<{ body: AuthenticatorResponse }>(
          '/backend/authenticator',
          {
            email: email.trim(),
          } satisfies AuthenticatorRequest,
          { suppressErrorToast: true },
        )

        // Return an SSO entry URL or an empty value for local login
        return response.data.body?.entry_url ?? ''
      },
      {
        fallback: localI18n.global.t('auth.login.actions.submit.error.startFailed'),
        suppressErrorToast: true,
      },
    )
  }

  // Sign in through a local or development authenticator and load the resulting account profile
  async function login(email: string, password?: string): Promise<void> {
    // Clear an earlier resolver failure before attempting local login
    resolveState.reset()

    // Run login and profile loading under one visible action state
    await loginState.run(
      async () => {
        try {
          // Build the authenticator payload without adding an unused development-login password
          const payload: LoginRequest = { email: email.trim() }
          if (password !== undefined) payload.password = password.trim()

          // Exchange the authenticator input for the session token handled by the HTTP layer
          await http.post('/auth/login', payload, { suppressErrorToast: true })

          // Reject incomplete authentication responses before loading the profile
          if (!getAccessToken())
            throw new Error(localI18n.global.t('auth.login.actions.submit.error.missingToken'))

          // Populate the authenticated account profile
          await fetchProfile()
        } catch (error) {
          // Remove partial session state before forwarding the failure
          clearSession()
          clearProfile()
          throw error
        }
      },
      {
        fallback: localI18n.global.t('auth.login.actions.submit.error.invalidCredentials'),
        suppressErrorToast: true,
      },
    )
  }

  // Register one credentials-based account
  async function register(payload: RegisterRequest): Promise<string> {
    // Run registration under its dedicated loading and error state
    return registerState.run(
      async () => {
        // Normalize text fields before sending the registration request
        const response = await http.post<{ message?: string }>('/auth/register', {
          email: payload.email.trim(),
          password: payload.password,
          name: payload.name.trim(),
          surname: payload.surname.trim(),
        })

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'Registration successful.'
      },
      { fallback: localI18n.global.t('auth.register.actions.submit.error') },
    )
  }

  // Load the profile represented by the current access token
  async function fetchProfile(): Promise<void> {
    // Treat a missing token as a signed-out session
    if (!getAccessToken()) {
      clearProfile()
      return
    }

    // Retrieve and publish the authenticated account details
    await profileState.run(async () => {
      // Ask the backend which account belongs to the current access token
      const response = await http.get<{ body: UserDetailsResponse }>('/user/details')
      setProfile(toProfile(response.data.body))
    })
  }

  // End the session locally and clear every account-specific store
  function logout(): void {
    clearSession()
    clearProfile()
  }

  // Reuse one promise while session restoration is already running
  let restoreInFlight: Promise<void> | null = null

  // Restore a valid browser session once for all concurrent callers
  function restore(): Promise<void> {
    // Return the active restore instead of loading the profile twice
    if (restoreInFlight) return restoreInFlight

    // Validate the session and quietly fall back to signed-out state on failure
    const restorePromise = (async () => {
      // Reject missing or expired sessions before making a request
      if (!hasValidSession()) {
        logout()
        return
      }

      // Keep startup resilient when a stored session is rejected by the backend
      try {
        await fetchProfile()
      } catch {
        logout()
      }
    })()

    restoreInFlight = restorePromise

    // Release the shared restore promise after every caller has observed it
    void restorePromise.then(() => {
      if (restoreInFlight === restorePromise) restoreInFlight = null
    })

    // Give every concurrent caller the same restore promise
    return restorePromise
  }

  // Expose account data together with each authentication action state
  return {
    profile,
    isAuthenticated,
    isResolvingAuth: resolveState.isLoading,
    isLoggingIn: loginState.isLoading,
    loginError,
    isRegistering: registerState.isLoading,
    registerError: registerState.error,
    isFetchingProfile: profileState.isLoading,
    isAdmin,
    isOwner,
    isDemo,
    setProfile,
    clearProfile,
    resolveAuthenticator,
    login,
    register,
    fetchProfile,
    logout,
    restore,
  }
})
