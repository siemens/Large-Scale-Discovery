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
import { createI18n } from 'vue-i18n'

import http from '@/api/http'
import { useAsyncState } from '@/composables/useAsyncState'

import type {
  AssetsRequest,
  CustomRequest,
  CustomScopeResponse,
  NetworksRequest,
  ScopeIdRequest,
  ScopeResponse,
  ScopeSecretResponse,
  ScopeTarget,
  ScopeTargetResetRequest,
  ScopeTargetsResponse,
  SettingsRequest,
} from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    configuration: {
      scopes: {
        actions: {
          delete: {
            error: 'Could not delete scope.',
          },
          generateScopeSecret: {
            error: 'Could not reset secret.',
          },
          newCycle: {
            error: 'Could not start new cycle.',
          },
          pause: {
            error: 'Could not update pause state',
          },
          rescanFailed: {
            error: 'Could not rescan failed targets.',
          },
        },
        error: 'Could not load scopes.',
      },
    },
    scopeAddAssetsPanel: {
      error: 'Could not save assets scope.',
    },
    scopeAddCustomPanel: {
      error: 'Could not save custom scope.',
    },
    scopeAddNetworksPanel: {
      error: 'Could not save networks scope.',
    },
    scopeSettingsPanel: {
      error: 'Could not save scan settings.',
    },
  },
}

// Create a tiny translator for messages owned by this store
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

/** Owns scan scopes, their creation rights, and all scope lifecycle and target actions. */
export const useScopesStore = defineStore('scopes', () => {
  // Store the scan scopes available to the current account
  const scopes = ref<ScopeResponse[]>([])

  // Store whether the account may create custom scopes
  const allowCustom = ref(false)

  // Store whether the account may create network scopes
  const allowNetwork = ref(false)

  // Store whether the account may create asset scopes
  const allowAsset = ref(false)

  // Track loading the scope list
  const fetchState = useAsyncState()

  // Track loading targets for one scope
  const fetchTargetsState = useAsyncState()

  // Track pausing or resuming a scope
  const pauseState = useAsyncState()

  // Track starting a new scan cycle
  const startCycleState = useAsyncState()

  // Track rescanning failed targets
  const rescanState = useAsyncState()

  // Track generating a replacement scope secret
  const generateScopeSecretState = useAsyncState()

  // Track deleting a scope
  const deleteState = useAsyncState()

  // Track resetting one target
  const resetTargetState = useAsyncState()

  // Track creating a custom scope
  const submitCustomState = useAsyncState()

  // Track creating an asset scope
  const submitAssetsState = useAsyncState()

  // Track creating a network scope
  const submitNetworksState = useAsyncState()

  // Track saving scope settings
  const submitSettingsState = useAsyncState()

  // Load scopes and the current account's creation permissions
  async function fetchScopes(): Promise<void> {
    await fetchState.run(
      async () => {
        // Deny creation until the backend explicitly grants each permission
        allowCustom.value = false
        allowNetwork.value = false
        allowAsset.value = false

        // Ask the backend for scopes and effective creation rights
        const response = await http.get<{
          body: {
            scopes: ScopeResponse[]
            allow_custom: boolean
            allow_network: boolean
            allow_asset: boolean
          }
        }>('/scopes')

        // Publish the returned scope registry
        scopes.value = response.data.body?.scopes ?? []

        // Publish each creation permission separately for simple consumers
        allowCustom.value = response.data.body?.allow_custom ?? false
        allowNetwork.value = response.data.body?.allow_network ?? false
        allowAsset.value = response.data.body?.allow_asset ?? false
      },
      { fallback: localI18n.global.t('configuration.scopes.error') },
    )
  }

  // Load targets and synchronization state for one scope
  async function fetchTargets(
    scopeId: number,
  ): Promise<{ targets: ScopeTarget[]; synchronization: boolean }> {
    return fetchTargetsState.run(
      async () => {
        // Ask the backend for the selected scope's targets
        const response = await http.post<{ body: ScopeTargetsResponse }>('/scope/targets', {
          id: scopeId,
        } satisfies ScopeIdRequest)

        // Normalize optional target data for the expanded table row
        return {
          targets: response.data.body?.targets ?? [],
          synchronization: response.data.body?.synchronization ?? false,
        }
      },
      { fallback: 'Could not load targets.' },
    )
  }

  // Toggle whether one scope is paused
  async function pauseScope(scopeId: number): Promise<string> {
    return pauseState.run(
      async () => {
        // Ask the backend to toggle the selected scope
        const response = await http.post<{ message: string }>('/scope/pause', {
          id: scopeId,
        } satisfies ScopeIdRequest)

        // Find the local scope that mirrors the successful backend action
        const scope = scopes.value.find((candidate) => candidate.id === scopeId)

        // Reflect the toggle immediately without another list request
        if (scope) scope.enabled = !scope.enabled

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      { fallback: localI18n.global.t('configuration.scopes.actions.pause.error') },
    )
  }

  // Start a new scan cycle for one scope
  async function startCycle(scopeId: number): Promise<string> {
    return startCycleState.run(
      async () => {
        // Ask the backend to start the selected scope's next cycle
        const response = await http.post<{ message: string }>('/scope/cycle', {
          id: scopeId,
        } satisfies ScopeIdRequest)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('configuration.scopes.actions.newCycle.error'),
        after: fetchScopes,
      },
    )
  }

  // Schedule another attempt for failed targets in one scope
  async function rescanFailedTargets(scopeId: number): Promise<string> {
    return rescanState.run(
      async () => {
        // Ask the backend to queue failed targets again
        const response = await http.post<{ message: string }>('/scope/rescan', {
          id: scopeId,
        } satisfies ScopeIdRequest)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('configuration.scopes.actions.rescanFailed.error'),
        after: fetchScopes,
      },
    )
  }

  // Replace one scope secret and return it only when manual display is required
  async function generateScopeSecret(scopeId: number): Promise<string> {
    return generateScopeSecretState.run(
      async () => {
        // Ask the backend to replace and deliver the selected scope secret
        const response = await http.post<{ message: string; body?: ScopeSecretResponse }>(
          '/scope/secret',
          { id: scopeId } satisfies ScopeIdRequest,
        )

        // Return a secret only when the backend could not deliver it automatically
        return response.data.body?.secret ?? ''
      },
      {
        fallback: localI18n.global.t('configuration.scopes.actions.generateScopeSecret.error'),
        after: fetchScopes,
      },
    )
  }

  // Delete one scope and refresh the registry after success
  async function deleteScope(scopeId: number): Promise<string> {
    return deleteState.run(
      async () => {
        // Ask the backend to delete the selected scope
        const response = await http.post<{ message: string }>('/scope/delete', {
          id: scopeId,
        } satisfies ScopeIdRequest)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('configuration.scopes.actions.delete.error'),
        after: fetchScopes,
      },
    )
  }

  // Reset one target so it can be processed again
  async function resetTarget(scopeId: number, input: string): Promise<string> {
    return resetTargetState.run(
      async () => {
        // Identify the scope and target that should be reset
        const response = await http.post<{ message: string }>('/scope/target/reset', {
          scope_id: scopeId,
          input,
        } satisfies ScopeTargetResetRequest)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      { fallback: 'Could not reset target.' },
    )
  }

  // Create or update a custom scope and return any backend warnings
  async function submitCustom(
    payload: CustomRequest,
  ): Promise<{ message: string; warnings?: string[] }> {
    return submitCustomState.run(
      async () => {
        // Send the complete custom-scope request to the backend
        const response = await http.post<{ message: string; body?: CustomScopeResponse }>(
          '/scope/update/custom',
          payload,
        )

        // Return the confirmation and optional validation warnings
        return {
          message: response.data.message ?? '',
          warnings: response.data.body?.warnings ?? undefined,
        }
      },
      {
        fallback: localI18n.global.t('scopeAddCustomPanel.error'),
        after: fetchScopes,
      },
    )
  }

  // Create or update an asset-based scope
  async function submitAssets(payload: AssetsRequest): Promise<string> {
    return submitAssetsState.run(
      async () => {
        // Send the complete asset-scope request to the backend
        const response = await http.post<{ message: string }>('/scope/update/assets', payload)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('scopeAddAssetsPanel.error'),
        after: fetchScopes,
      },
    )
  }

  // Create or update a network-based scope
  async function submitNetworks(payload: NetworksRequest): Promise<string> {
    return submitNetworksState.run(
      async () => {
        // Send the complete network-scope request to the backend
        const response = await http.post<{ message: string }>('/scope/update/networks', payload)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('scopeAddNetworksPanel.error'),
        after: fetchScopes,
      },
    )
  }

  // Save scan settings for one existing scope
  async function submitSettings(payload: SettingsRequest): Promise<string> {
    return submitSettingsState.run(
      async () => {
        // Send the complete scan-settings request to the backend
        const response = await http.post<{ message: string }>('/scope/update/settings', payload)

        // Return the backend confirmation to the calling page
        return response.data.message ?? ''
      },
      {
        fallback: localI18n.global.t('scopeSettingsPanel.error'),
        after: fetchScopes,
      },
    )
  }

  // Expose scope data together with every independent request lifecycle
  return {
    scopes,
    allowCustom,
    allowNetwork,
    allowAsset,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    isFetchingTargets: fetchTargetsState.isLoading,
    isPausing: pauseState.isLoading,
    isStartingCycle: startCycleState.isLoading,
    isRescanning: rescanState.isLoading,
    isResettingSecret: generateScopeSecretState.isLoading,
    isDeleting: deleteState.isLoading,
    isResettingTarget: resetTargetState.isLoading,
    isSubmittingCustom: submitCustomState.isLoading,
    isSubmittingAssets: submitAssetsState.isLoading,
    isSubmittingNetworks: submitNetworksState.isLoading,
    isSubmittingSettings: submitSettingsState.isLoading,
    fetchScopes,
    fetchTargets,
    pauseScope,
    startCycle,
    rescanFailedTargets,
    generateScopeSecret,
    deleteScope,
    resetTarget,
    submitCustom,
    submitAssets,
    submitNetworks,
    submitSettings,
  }
})
