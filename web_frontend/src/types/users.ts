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

/** A user account as managed in the admin users panel. */
export interface UserResponse {
  id: number
  email: string
  active?: boolean
  admin?: boolean
  demo?: boolean
  name: string
  surname: string
  company: string | null
  department: string | null
  gender: string | null
  last_login: string
}

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for `POST /admin/user/delete`. */
export interface DeleteUserRequest {
  id: number
}

// ── Response interfaces ─────────────────────────────────────────────────────
