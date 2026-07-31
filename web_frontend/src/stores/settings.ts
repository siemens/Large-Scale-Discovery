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
import { ref } from 'vue'

import http from '@/api/http'
import { useAsyncState } from '@/composables/useAsyncState'
import type { FrontendLink } from '@/types/settings'
import { parseSettingsResponse } from '@/types/settings'

/** Owns backend-provided frontend configuration such as login modes and external links. */
export const useSettingsStore = defineStore('settings', () => {
  // Store whether the backend permits password-free development login
  const isDevelopmentLogin = ref(false)

  // Store whether users may register credentials-based accounts
  const isCredentialsRegistration = ref(false)

  // Store links that the navigation and footer can render
  const frontendLinks = ref<Record<string, FrontendLink[]>>({})

  // Store the largest token lifetime accepted by the backend
  const tokenExpiryMaxDays = ref(365)

  // Remember whether the initial settings request completed
  const isLoaded = ref(false)

  // Track the settings request for loading and error feedback
  const fetchState = useAsyncState()

  // Reuse one promise while the settings request is already running
  let inFlightFetch: Promise<void> | null = null

  // Load public frontend settings once for all interested consumers
  function fetchSettings(endpoint: string): Promise<void> {
    // Return the active request instead of sending a duplicate
    if (inFlightFetch) return inFlightFetch

    // Fetch and copy the backend settings into their focused state fields
    inFlightFetch = fetchState.run(async () => {
      // Request the settings that are safe to expose to the frontend
      const response = await http.get<{ body?: unknown }>(endpoint)

      // Validate and normalize untrusted public settings before exposing them to consumers
      const body = parseSettingsResponse(response.data.body ?? {})

      // Copy login and registration settings into reactive state
      isDevelopmentLogin.value = body.development_login ?? false
      isCredentialsRegistration.value = body.credentials_registration ?? false

      // Copy navigation and token settings into reactive state
      frontendLinks.value = body.frontend_links ?? {}
      tokenExpiryMaxDays.value = body.token_expiry_max_days ?? 365

      // Mark the store ready after every setting has been copied
      isLoaded.value = true
    })

    // Release the shared promise after success or failure
    void inFlightFetch.then(
      () => {
        inFlightFetch = null
      },
      () => {
        inFlightFetch = null
      },
    )

    // Give every concurrent caller the same request promise
    return inFlightFetch
  }

  // Expose settings and their shared request lifecycle
  return {
    isDevelopmentLogin,
    isCredentialsRegistration,
    frontendLinks,
    tokenExpiryMaxDays,
    isLoaded,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    fetchSettings,
  }
})
