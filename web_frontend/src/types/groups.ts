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

/** The user sub-object embedded in an ownership entry. */
interface OwnershipUser {
  email: string
}

/** A user-to-group ownership entry. */
interface Ownership {
  id: number
  user: OwnershipUser
}

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for `POST /admin/group/create`. */
export interface CreateGroupRequest {
  name: string
  db_server_id: number
  max_scopes: number
  max_views: number
  max_targets: number
  max_owners: number
  allow_custom: boolean
  allow_network: boolean
  allow_asset: boolean
}

/** Payload for `POST /admin/group/update`. */
export interface UpdateGroupRequest {
  id: number
  name?: string
  db_server_id?: number
  max_scopes?: number
  max_views?: number
  max_targets?: number
  max_owners?: number
  allow_custom?: boolean
  allow_network?: boolean
  allow_asset?: boolean
}

/** Payload for `POST /admin/group/delete`. */
export interface DeleteGroupRequest {
  id: number
}

/** Payload for `POST /admin/group/assign`. */
export interface AssignOwnersRequest {
  id: number
  owners: string[]
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** A scan group as returned by `GET /admin/groups`. */
export interface GroupResponse {
  id: number
  name: string
  created: string
  created_by: string
  db_server_id: number
  max_scopes: number
  max_views: number
  max_targets: number
  max_owners: number
  allow_custom: boolean
  allow_network: boolean
  allow_asset: boolean
  ownerships: Ownership[]
}
