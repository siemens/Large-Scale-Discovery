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

import type {
  AssignOwnersRequest,
  CreateGroupRequest,
  DeleteGroupRequest,
  GroupResponse,
  UpdateGroupRequest,
} from '@/types/groups'
import type { AxiosRequestConfig } from 'axios'

/** Owns the group registry and its create/update/delete/owner-assignment actions. */
export const useGroupsStore = defineStore('groups', () => {
  // Store every group the current account may administer
  const groups = ref<GroupResponse[]>([])

  // Track list and every mutation request independently
  const fetchState = useAsyncState()

  // Track group creation requests
  const createState = useAsyncState()

  // Track group update requests
  const updateState = useAsyncState()

  // Track group deletion requests
  const deleteState = useAsyncState()

  // Track owner-assignment requests
  const assignState = useAsyncState()

  // Load every group the current account may administer
  async function fetchGroups(config?: AxiosRequestConfig): Promise<void> {
    await fetchState.run(
      async () => {
        // Ask the backend for the complete group registry
        const response = await http.get<{ body: { groups: GroupResponse[] } }>('/groups', config)

        // Keep the newest backend entries at the front of the administration list
        groups.value = [...(response.data.body?.groups ?? [])].reverse()
      },
      { trackLoading: config?.suppressLoader !== true },
    )
  }

  // Create one group and refresh the registry after success
  async function createGroup(payload: CreateGroupRequest): Promise<void> {
    await createState.run(() => http.post('/admin/group/create', payload), { after: fetchGroups })
  }

  // Update the editable scalar fields of one group
  async function updateGroup(payload: UpdateGroupRequest): Promise<void> {
    await updateState.run(() => http.post('/admin/group/update', payload))
  }

  // Delete one group and remove it from local state
  async function deleteGroup(id: number): Promise<void> {
    await deleteState.run(async () => {
      // Ask the backend to delete the selected group
      await http.post('/admin/group/delete', { id } satisfies DeleteGroupRequest)

      // Remove the deleted group without another list request
      groups.value = groups.value.filter((group) => group.id !== id)
    })
  }

  // Replace the owners of one group and refresh the registry
  async function assignOwners(payload: AssignOwnersRequest): Promise<void> {
    await assignState.run(() => http.post('/admin/group/assign', payload), { after: fetchGroups })
  }

  // Clear groups and every request lifecycle
  function reset(): void {
    // Remove domain data from the previous account
    groups.value = []

    // Reset each independent asynchronous action
    fetchState.reset()
    createState.reset()
    updateState.reset()
    deleteState.reset()
    assignState.reset()
  }

  // Expose group data together with independent request lifecycles
  return {
    groups,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    isCreating: createState.isLoading,
    createError: createState.error,
    isUpdating: updateState.isLoading,
    updateError: updateState.error,
    isDeleting: deleteState.isLoading,
    deleteError: deleteState.error,
    isAssigning: assignState.isLoading,
    assignError: assignState.error,
    fetchGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    assignOwners,
    reset,
  }
})
