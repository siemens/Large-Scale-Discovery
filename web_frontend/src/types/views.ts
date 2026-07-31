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

import type { CycleProgressTuple } from '@/lib/collections'
import type { ScopeResponse } from '@/types/scopes'

// ── Generic interfaces ──────────────────────────────────────────────────────

/** Database connection details for a scope's result database. */
export interface ScopeConnection {
  host: string
  port: number
  database: string
}

/** A user or token grant on a view. */
export interface ViewGrant {
  is_user: boolean
  username: string
  description?: string
  created_by?: string
  created?: string
  expiry?: string
  user_company?: string
  user_department?: string
  user_last_login?: string
  user_created?: string
  user_is_admin?: boolean
  user_is_owner?: boolean
}

/** Scope dropdown option used when creating a view. */
/** Flattened token row for the admin token overview table. */
export interface TokenRow {
  view_id: string
  view_name: string
  scope_name: string
  group_name: string
  username: string
  description: string
  created_by: string
  created: string
  expiry: string
}

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for `POST /view/create`. */
export interface CreateViewRequest {
  scope_id: number
  view_name: string
  filter_input_targets: string[]
  filter_input_countries: string[]
  filter_input_locations: string[]
  filter_input_routing_domains: string[]
  filter_input_zones: string[]
  filter_input_purposes: string[]
  filter_input_companies: string[]
  filter_input_departments: string[]
  filter_input_managers: string[]
  filter_input_contacts: string[]
  filter_input_comments: string[]
}

/** Payload for `POST /view/update`. */
export interface UpdateViewRequest {
  id: string
  name: string
}

/** Payload for `POST /view/delete`. */
export interface DeleteViewRequest {
  id: string
}

/** Payload for `POST /view/grant/users`. */
export interface ViewGrantUsersRequest {
  view_id: string
  users: string[]
}

/** Payload for `POST /view/grant/revoke`. */
export interface ViewRevokeGrantRequest {
  view_id: string
  username: string
}

/** Payload for `POST /view/grant/token`. */
export interface ViewGrantTokenRequest {
  view_id: string
  description: string
  expiry_days: number
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** A view as returned by `GET /views`. */
export interface ViewResponse {
  id: string
  name: string
  scan_scope: ScopeResponse & { cycle_progress?: CycleProgressTuple }
  filters: Record<string, string[]>
  grants: ViewGrant[]
}

/** A view granted to the current user, as returned by `GET /views/granted`. */
export interface GrantedViewResponse {
  id: number
  name: string
  view_names: string

  // The granted-views endpoint always populates the database connection (unlike
  // the configuration-views endpoint), so connection is required here.
  scan_scope: ScopeResponse & { cycle_progress?: CycleProgressTuple; connection: ScopeConnection }
}

/** Body of `POST /view/grant/token` response (generated credentials). */
export interface ViewGrantTokenResponse {
  username: string
  password: string
}
