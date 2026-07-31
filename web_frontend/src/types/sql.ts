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

/** Payload for `POST /admin/sql`. */
export interface SqlRequest {
  db_name: string
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** A SQL query log entry as returned by `POST /admin/sql`. */
export interface SqlResponse {
  id: number
  query: string
  db_table: string
  db_user: string
  db_name: string
  query_timestamp: string
  query_duration_string: string
  total_duration_string: string
  query_results: string | number
  client_name: string
}
