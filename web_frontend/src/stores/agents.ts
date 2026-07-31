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

import http from '@/api/http'
import { useListFetcher } from '@/composables/useAsyncState'

import type { AgentGroupResponse } from '@/types/agents'
import type { AxiosRequestConfig } from 'axios'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    error: 'Could not load agents.',
  },
}

/** Owns the scope-grouped agent list shown on the home dashboard */
export const useAgentsStore = defineStore('agents', () => {
  // Load agent groups while keeping unchanged polling results on the same reference
  const {
    data: agentGroups,
    isFetching,
    fetchError,
    fetch: fetchAgents,
    fetchSilent: fetchAgentsSilent,
    reset,
  } = useListFetcher<AgentGroupResponse[]>(async (config?: AxiosRequestConfig) => {
    // Ask the backend for every agent grouped by its scan scope
    const response = await http.get<{ body: { scope_agents: AgentGroupResponse[] } }>(
      '/agents',
      config,
    )

    // Return an empty list when the backend has no registered agents
    return response.data.body?.scope_agents ?? []
  }, localMessages['en-US'].error)

  // Expose the agent data together with its request lifecycle
  return {
    agentGroups,
    isFetching,
    fetchError,
    fetchAgents,
    fetchAgentsSilent,
    reset,
  }
})
