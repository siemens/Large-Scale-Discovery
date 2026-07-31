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

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for `POST /admin/database/update` (create or update). */
export interface DatabaseRequest {
  id?: number
  name: string
  dialect: string
  host: string
  host_public: string
  port: number
  admin: string
  password: string
  args: string
}

/** Payload for `POST /admin/database/remove`. */
export interface DeleteDatabaseRequest {
  id: number
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** A database server as returned by `GET /admin/databases`. */
export interface DatabaseResponse {
  id: number
  name: string
  dialect: string
  host: string
  host_public: string
  port: number
  admin: string
  args: string
}
