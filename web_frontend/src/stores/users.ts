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

import type { DeleteUserRequest, UserResponse } from '@/types/users'
import type { AxiosRequestConfig } from 'axios'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    admin: {
      users: {
        fetchError: 'Could not load users.',
        actions: {
          delete: {
            success: 'User deleted.',
          },
          save: {
            success: 'User saved.',
          },
        },
      },
    },
  },
}

// Create a tiny translator for messages owned by this store
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

// Prefer a useful backend message and fall back to the local action text
function messageOf(response: { data: { message?: string } }, fallback: string): string {
  return response.data.message || fallback
}

/** Owns the user registry and its update/delete actions. */
export const useUsersStore = defineStore('users', () => {
  // Remember whether at least one user request completed
  const isLoaded = ref(false)

  // Load users while preserving the list reference when polling returns unchanged data
  const {
    data: users,
    isFetching,
    fetchError,
    fetch: fetchUserList,
    reset: resetUserList,
  } = useListFetcher<UserResponse[]>(
    async (config?: AxiosRequestConfig) => {
      // Ask the LSD user endpoint for its account records
      const response = await http.get<{ body: { users: UserResponse[] } }>('/users', config)

      // Mark the store ready after receiving a valid response
      isLoaded.value = true

      // Keep the newest backend entries at the front of the administration list
      return [...(response.data.body?.users ?? [])].reverse()
    },
    localI18n.global.t('admin.users.fetchError'),
    { initialData: () => [] },
  )

  // Restore backend order for dashboard consumers that show recent accounts
  const recentUsers = computed(() => [...users.value].reverse())

  // Track update and deletion requests independently
  const updateState = useAsyncState()

  // Track account deletion requests
  const deleteState = useAsyncState()

  // Refresh the shared user registry
  async function fetchUsers(config?: AxiosRequestConfig): Promise<void> {
    await fetchUserList(config)
  }

  async function fetchUsersSilent(): Promise<void> {
    await fetchUserList({ suppressLoader: true })
  }

  // Save the editable LSD account fields
  async function updateUser(user: UserResponse): Promise<string> {
    return updateState.run(async () => {
      // Send the typed LSD user payload to the update endpoint
      const response = await http.post<{ message?: string }>('/admin/user/update', user)

      // Return a user-facing success message to the calling page
      return messageOf(response, localI18n.global.t('admin.users.actions.save.success'))
    })
  }

  // Delete one account and remove it from local state immediately
  async function deleteUser(id: number): Promise<string> {
    return deleteState.run(async () => {
      // Ask the backend to delete the selected account
      const response = await http.post<{ message?: string }>('/admin/user/delete', {
        id,
      } satisfies DeleteUserRequest)

      // Remove the deleted account without requiring another list request
      users.value = users.value.filter((user) => user.id !== id)

      // Return a user-facing success message to the calling page
      return messageOf(response, localI18n.global.t('admin.users.actions.delete.success'))
    })
  }

  // Clear user data and request state when the active account changes
  function reset(): void {
    resetUserList()
    isLoaded.value = false
  }

  // Expose user data together with independent request lifecycles
  return {
    users,
    recentUsers,
    isLoaded,
    isFetching,
    fetchError,
    isUpdating: updateState.isLoading,
    updateError: updateState.error,
    isDeleting: deleteState.isLoading,
    deleteError: deleteState.error,
    fetchUsers,
    fetchUsersSilent,
    updateUser,
    deleteUser,
    reset,
  }
})
