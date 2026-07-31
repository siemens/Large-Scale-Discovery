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

import type { SqlRequest, SqlResponse } from '@/types/sql'

/** Owns the SQL execution log shown in the admin area. */
export const useSqlLogsStore = defineStore('sqlLogs', () => {
  // Store the newest SQL executions shown in administration
  const sqlLogs = ref<SqlResponse[]>([])

  // Track the SQL log request for loading and error feedback
  const fetchState = useAsyncState()

  // Load recent executions across every configured result database
  async function fetchLogs(): Promise<void> {
    await fetchState.run(async () => {
      // Use an empty database name to request logs across all databases
      const response = await http.post<{ body: { logs: SqlResponse[] | null } }>('/admin/sql', {
        db_name: '',
      } satisfies SqlRequest)

      // Normalize a missing backend list before applying the browser limit
      const allLogs = response.data.body?.logs ?? []

      // Keep only the newest 2,000 rows so client-side sorting remains lightweight
      sqlLogs.value = allLogs.slice(-2_000).reverse()
    })
  }

  // Clear SQL logs and their request lifecycle
  function reset(): void {
    sqlLogs.value = []
    fetchState.reset()
  }

  // Expose SQL logs together with their request lifecycle
  return {
    sqlLogs,
    isFetching: fetchState.isLoading,
    fetchError: fetchState.error,
    fetchLogs,
    reset,
  }
})
