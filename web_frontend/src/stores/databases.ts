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

import type { DatabaseRequest, DatabaseResponse, DeleteDatabaseRequest } from '@/types/databases'

/** useDatabasesStore owns result databases without retaining their credentials. */
export const useDatabasesStore = defineStore('databases', () => {
  // Store database metadata returned without passwords by the backend
  const databases = ref<DatabaseResponse[]>([])

  // Track loading the database list
  const fetchState = useAsyncState()

  // Track database creation and updates
  const saveState = useAsyncState()

  // Track database deletion
  const deleteState = useAsyncState()

  // Load the complete result-database registry
  async function fetchDatabases(): Promise<void> {
    await fetchState.run(async () => {
      // Ask the backend for database metadata that is safe to display
      const response = await http.get<{ body?: { databases: DatabaseResponse[] } }>(
        '/admin/databases',
      )

      // Keep the newest database entries at the front of the table
      databases.value = (response.data.body?.databases ?? []).slice().reverse()
    })
  }

  // Create or update a database and refresh the registry after success
  async function saveDatabase(payload: DatabaseRequest): Promise<void> {
    await saveState.run(() => http.post('/admin/database/update', payload), {
      after: fetchDatabases,
    })
  }

  // Delete one database and remove it from local state
  async function deleteDatabase(id: number): Promise<void> {
    await deleteState.run(async () => {
      // Ask the backend to delete the selected database
      await http.post('/admin/database/remove', { id } satisfies DeleteDatabaseRequest)

      // Remove the deleted database without another list request
      databases.value = databases.value.filter((database) => database.id !== id)
    })
  }

  // Clear database metadata and every request lifecycle
  function reset(): void {
    // Remove domain data from the previous account
    databases.value = []

    // Reset each independent asynchronous action
    fetchState.reset()
    saveState.reset()
    deleteState.reset()
  }

  // Expose database metadata together with independent request lifecycles
  return {
    databases,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    isSaving: saveState.isLoading,
    saveError: saveState.error,
    isDeleting: deleteState.isLoading,
    deleteError: deleteState.error,
    fetchDatabases,
    saveDatabase,
    deleteDatabase,
    reset,
  }
})
