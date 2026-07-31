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

import type { AgentResponse } from '@/types/agents'
import type { ScopeConnection } from '@/types/views'

/** NETWORK_SIZE_MAX is the hard Nmap-compatible target size used for new scopes. */
export const NETWORK_SIZE_MAX = 2_048

// ── Generic interfaces ──────────────────────────────────────────────────────

/** A discovery scan time window (day + time range). */
export interface ScanTimespan {
  startDay: string
  startTime: string
  endDay: string
  endTime: string
}

/** EditableTimespan is the user-editable form of a weekly scan window. */
export interface EditableTimespan {
  id: number
  startDay: number
  startTime: string
  endDay: number
  endTime: string
}

/** Per-scope scan module configuration (timeouts, concurrency, filters). */
export interface ScanSettings {
  max_instances_banner: number
  max_instances_discovery: number
  max_instances_nfs: number
  max_instances_nuclei: number
  max_instances_smb: number
  max_instances_ssh: number
  max_instances_ssl: number
  max_instances_webcrawler: number
  max_instances_webenum: number
  sensitive_ports: number[]
  network_timeout_seconds: number
  discovery_timespans: ScanTimespan[] | null
  discovery_nmap_args: string
  discovery_nmap_args_prescan: string
  discovery_exclude_domains: string
  nfs_scan_timeout_minutes: number
  nfs_depth: number
  nfs_threads: number
  nfs_exclude_shares: string
  nfs_exclude_folders: string
  nfs_exclude_extensions: string
  nfs_exclude_file_size_below: number
  nfs_exclude_last_modified_below: string
  nfs_accessible_only: number
  smb_scan_timeout_minutes: number
  smb_depth: number
  smb_threads: number
  smb_forced_shares: string
  smb_exclude_shares: string
  smb_exclude_folders: string
  smb_exclude_extensions: string
  smb_exclude_file_size_below: number
  smb_exclude_last_modified_below: string
  smb_accessible_only: number
  nuclei_scan_timeout_minutes: number
  nuclei_include_severities: string
  nuclei_exclude_severities: string
  nuclei_include_tags: string
  nuclei_exclude_tags: string
  nuclei_include_ids: string
  nuclei_exclude_ids: string
  nuclei_include_protocols: string
  nuclei_exclude_protocols: string
  ssh_scan_timeout_minutes: number
  ssl_scan_timeout_minutes: number
  http_user_agent: string
  webcrawler_scan_timeout_minutes: number
  webcrawler_depth: number
  webcrawler_max_threads: number
  webcrawler_follow_query_strings: boolean
  webcrawler_always_store_root: boolean
  webcrawler_follow_types: string
  webenum_scan_timeout_minutes: number
  webenum_probe_robots: boolean
  ot: boolean
}

/** Tri-state criticality filter — the backend stores this as a string, not a boolean. */
export type AssetCritical = 'Any' | 'Yes' | 'No'

/** Asset filter attributes attached to a scope. */
interface ScopeAttributes {
  targets?: ScopeTarget[]
  sync?: boolean
  asset_type?: string
  asset_companies?: string[] | null
  asset_departments?: string[] | null
  asset_countries?: string[] | null
  asset_locations?: string[] | null
  asset_contacts?: string[] | null
  asset_critical?: AssetCritical
  asset_routing_domains?: string[] | null
  asset_zones?: string[] | null
  asset_purposes?: string[] | null
  asset_exclude_keywords?: string[] | null
}

/** A single scan target entry within a scope. */
export interface ScopeTarget {
  input?: string
  enabled?: boolean
  priority?: boolean
  timezone?: number | null
  lat?: string | null
  lng?: string | null
  postal_address?: string
  input_network?: string
  input_country?: string
  input_location?: string
  input_routing_domain?: string
  input_zone?: string
  input_purpose?: string
  input_company?: string
  input_department?: string
  input_manager?: string
  input_contact?: string
  input_comment?: string
  scan_started?: string | null
  scan_finished?: string | null
}

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for scope-id-only endpoints: `/scope/targets`, `/scope/pause`, `/scope/cycle`, `/scope/rescan`, `/scope/secret`, `/scope/delete`. */
export interface ScopeIdRequest {
  id: number
}

/** Payload for `POST /scope/target/reset`. */
export interface ScopeTargetResetRequest {
  scope_id: number
  input: string
}

/** Payload for `POST /scope/update/custom` (create or update). */
export interface CustomRequest {
  name: string
  cycles: boolean
  ot: boolean
  cycles_retention: number
  split_size: number
  targets?: ScopeTarget[] // omit to leave existing targets unchanged; send [] to clear all
  scope_id?: number
  group_id?: number
}

/** Payload for `POST /scope/assets` (create or update). */
export interface AssetsRequest {
  type: string
  name: string
  cycles: boolean
  cycles_retention: number
  split_size: number
  sync: boolean
  asset_type: string
  asset_companies: string[]
  asset_departments: string[]
  asset_countries: string[]
  asset_locations: string[]
  asset_contacts: string[]
  asset_critical: AssetCritical
  scope_id?: number
  group_id?: number
}

/** Payload for `POST /scope/networks` (create or update). */
export interface NetworksRequest {
  type: string
  name: string
  cycles: boolean
  cycles_retention: number
  split_size: number
  sync: boolean
  asset_companies: string[]
  asset_departments: string[]
  asset_routing_domains: string[]
  asset_zones: string[]
  asset_purposes: string[]
  asset_countries: string[]
  asset_locations: string[]
  asset_contacts: string[]
  asset_exclude_keywords: string[]
  scope_id?: number
  group_id?: number
}

/** Payload for `POST /scope/settings`. */
export interface SettingsRequest {
  id: number
  scan_settings: ScanSettings
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** Body of `POST /scope/targets` response. */
export interface ScopeTargetsResponse {
  targets: ScopeTarget[] | null
  synchronization: boolean
}

/** Body of `POST /scope/update/custom` response. */
export interface CustomScopeResponse {
  warnings: string[] | null
}

/**
 * Body of `POST /scope/secret` response. The backend only populates `secret`
 * when it could NOT deliver the new secret by encrypted e-mail (the user has no
 * certificate configured), so the web UI can reveal it once. When e-mail
 * delivery succeeded, `secret` is empty.
 */
export interface ScopeSecretResponse {
  secret: string
}

/** A scan scope as returned by `GET /scopes`. */
export interface ScopeResponse {
  id: number
  name: string
  group_id: number
  group_name: string
  type: string
  size: number
  enabled: boolean
  cycles: boolean
  cycles_retention: number
  split_size: number
  created: string
  created_by: string
  attributes: ScopeAttributes
  scan_settings: ScanSettings
  scan_agents: AgentResponse[] | null
  cycle: number
  cycle_started: string | null
  cycle_done: number
  cycle_failed: number
  cycle_active: number
  cycle_queue?: Record<string, number>
  last_sync: string
  secret: string
  synchronization: boolean
  connection?: ScopeConnection
}
