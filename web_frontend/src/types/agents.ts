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

// ── Generic interfaces ──────────────────────────────────────────────────────

/** Per-module task counts currently assigned to an agent. */
interface AgentTask {
  Discovery?: number
  Banner?: number
  Nfs?: number
  Nuclei?: number
  Smb?: number
  Ssh?: number
  Ssl?: number
  Webcrawler?: number
  Webenum?: number
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** A scan agent as returned by the backend API. */
export interface AgentResponse {
  id: number
  name: string
  host: string
  ip: string
  platform: string
  platform_family: string
  platform_version: string
  cpu_rate: number
  cpu_cores: number
  cpu_mhz: number
  memory_rate: number
  memory_total: number
  memory_bytes?: number
  last_seen: string
  shared: boolean
  tasks: AgentTask | null
  api_version: string
  build_commit: string
  build_timestamp: string
  version_nmap?: string
  version_npcap?: string
  version_sslyze?: string
  version_python?: string
  limits?: boolean
}

/** Agents grouped by the scope they are assigned to. */
export interface AgentGroupResponse {
  scope_id: number
  scope_name: string
  scope_db: string
  scope_ownership: boolean
  agents: AgentResponse[]
}
