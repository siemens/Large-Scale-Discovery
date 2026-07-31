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
import { useAsyncState, useListFetcher } from '@/composables/useAsyncState'
import { toCycleProgress } from '@/lib/collections'

import type {
  CreateViewRequest,
  DeleteViewRequest,
  GrantedViewResponse,
  TokenRow,
  UpdateViewRequest,
  ViewGrantTokenRequest,
  ViewGrantTokenResponse,
  ViewGrantUsersRequest,
  ViewResponse,
  ViewRevokeGrantRequest,
} from '@/types/views'
import type { AxiosRequestConfig } from 'axios'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    configuration: {
      views: {
        actions: {
          delete: {
            error: 'Could not delete view.',
          },
          revokeUser: {
            error: 'Could not revoke access.',
          },
        },
        error: 'Could not load views.',
      },
    },
    grantedPanel: {
      error: 'Could not load granted views.',
    },
    viewAddPanel: {
      error: 'Could not create view.',
    },
    viewEditPanel: {
      error: 'Could not update view.',
    },
    viewGrantPanel: {
      error: 'Could not update access rights.',
    },
    viewTokenPanel: {
      error: 'Could not create access token.',
    },
  },
}

// Create a tiny translator for messages owned by this store
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

/** Owns the administrable scan views, their grants, and token/user access management. */
export const useViewsStore = defineStore('views', () => {
  // Store views the current account may administer
  const views = ref<ViewResponse[]>([])

  // Track the administration list request through the shared async envelope
  const fetchState = useAsyncState()

  // Load granted views while preserving the list reference when data stays unchanged
  const {
    data: grantedViews,
    isFetching: isFetchingGranted,
    fetchError: fetchGrantedViewsError,
    fetch: fetchGrantedViews,
    fetchSilent: fetchGrantedViewsSilent,
    reset: resetGrantedViews,
  } = useListFetcher<GrantedViewResponse[]>(async (config?: AxiosRequestConfig) => {
    // Ask the backend for every view currently accessible to the account
    const response = await http.get<{ body: { views: GrantedViewResponse[] } }>(
      '/views/granted',
      config,
    )

    // Normalize a missing backend list before enriching its scopes
    const grantedViewList = response.data.body?.views ?? []

    // Add the cycle summary consumed by each granted-view card
    for (const view of grantedViewList) {
      // Update the nested scope referenced by this view
      const scanScope = view.scan_scope
      scanScope.cycle_progress = toCycleProgress(
        scanScope.cycle_done,
        scanScope.cycle_failed,
        scanScope.cycle_active,
      )
    }

    // Publish the enriched granted-view list
    return grantedViewList
  }, localMessages['en-US'].grantedPanel.error)

  // Track creating and updating views
  const submitState = useAsyncState()

  // Track deleting views and individual grants
  const deleteState = useAsyncState()

  // Flatten non-user grants into the token rows rendered by administration
  const tokens = computed<TokenRow[]>(() =>
    views.value.flatMap((view) =>
      view.grants
        .filter((grant) => !grant.is_user)
        .map((grant) => ({
          view_id: view.id,
          view_name: view.name,
          scope_name: view.scan_scope.name,
          group_name: view.scan_scope.group_name,
          username: grant.username,
          description: grant.description ?? '',
          created_by: grant.created_by ?? '',
          created: grant.created ?? '',
          expiry: grant.expiry ?? '',
        })),
    ),
  )

  // Tell access-gated layouts whether the account has at least one granted view
  const hasAccess = computed(() => (grantedViews.value?.length ?? 0) > 0)

  // Load every view the current account may administer
  async function fetchViews(): Promise<void> {
    await fetchState.run(
      async () => {
        // Ask the backend for the complete administrable view registry
        const response = await http.get<{ body: { views: ViewResponse[] } }>('/views')

        // Publish an empty list when the backend has no administrable views
        views.value = response.data.body?.views ?? []
      },
      { fallback: localI18n.global.t('configuration.views.error') },
    )
  }

  // Create one view
  async function createView(payload: CreateViewRequest): Promise<string> {
    return submitState.run(
      async () => {
        // Send the complete view request to the backend
        const response = await http.post<{ message?: string }>('/view/create', payload)

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'View created.'
      },
      { fallback: localI18n.global.t('viewAddPanel.error') },
    )
  }

  // Rename one existing view
  async function updateView(id: string, name: string): Promise<string> {
    return submitState.run(
      async () => {
        // Identify the view and its replacement name
        const response = await http.post<{ message?: string }>('/view/update', {
          id,
          name,
        } satisfies UpdateViewRequest)

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'View updated.'
      },
      { fallback: localI18n.global.t('viewEditPanel.error') },
    )
  }

  // Delete one existing view
  async function deleteView(id: string): Promise<string> {
    return deleteState.run(
      async () => {
        // Identify the view that should be deleted
        const response = await http.post<{ message?: string }>('/view/delete', {
          id,
        } satisfies DeleteViewRequest)

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'View deleted.'
      },
      { fallback: localI18n.global.t('configuration.views.actions.delete.error') },
    )
  }

  // Grant one view to the supplied user addresses
  async function grantUsers(viewId: string, emails: string[]): Promise<string> {
    return submitState.run(
      async () => {
        // Identify the view and the complete user-address list
        const response = await http.post<{ message?: string }>('/view/grant/users', {
          view_id: viewId,
          users: emails,
        } satisfies ViewGrantUsersRequest)

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'Access rights updated.'
      },
      { fallback: localI18n.global.t('viewGrantPanel.error') },
    )
  }

  // Revoke one user or token grant from a view
  async function revokeGrant(viewId: string, username: string): Promise<string> {
    return deleteState.run(
      async () => {
        // Identify the view and grant that should be removed
        const response = await http.post<{ message?: string }>('/view/grant/revoke', {
          view_id: viewId,
          username,
        } satisfies ViewRevokeGrantRequest)

        // Return the backend confirmation or the local success text
        return response.data.message ?? 'Access revoked.'
      },
      { fallback: localI18n.global.t('configuration.views.actions.revokeUser.error') },
    )
  }

  // Create credentials that grant one external client access to a view
  async function generateAccessToken(
    viewId: string,
    description: string,
    expiryDays: number,
  ): Promise<{ message: string; credentials: ViewGrantTokenResponse | null }> {
    return submitState.run(
      async () => {
        // Send the view, description, and token lifetime to the backend
        const response = await http.post<{ message?: string; body?: ViewGrantTokenResponse }>(
          '/view/grant/token',
          { view_id: viewId, description, expiry_days: expiryDays } satisfies ViewGrantTokenRequest,
        )

        // Read the one-time credential response before it leaves this action
        const body = response.data.body

        // Keep credentials only when the backend returned at least one secret field
        const credentials =
          body && (body.username || body.password)
            ? { username: body.username, password: body.password }
            : null

        // Return the confirmation and one-time credentials directly to the caller
        return { message: response.data.message ?? 'Access token created.', credentials }
      },
      { fallback: localI18n.global.t('viewTokenPanel.error') },
    )
  }

  // Clear all view data and every request lifecycle
  function reset(): void {
    // Remove administration and granted-view data from the previous account
    views.value = []
    resetGrantedViews()

    // Reset each independent asynchronous action
    fetchState.reset()
    submitState.reset()
    deleteState.reset()
  }

  // Expose view data together with independent request lifecycles
  return {
    views,
    grantedViews,
    hasAccess,
    isFetchingViews: fetchState.isLoading,
    isFetchingGranted,
    isSubmitting: submitState.isLoading,
    isDeleting: deleteState.isLoading,
    fetchError: fetchState.error,
    fetchGrantedViewsError,
    submitError: submitState.error,
    tokens,
    fetchViews,
    fetchGrantedViews,
    fetchGrantedViewsSilent,
    resetGrantedViews,
    reset,
    createView,
    updateView,
    deleteView,
    grantUsers,
    revokeGrant,
    generateAccessToken,
  }
})
