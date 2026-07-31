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

/** Layout-facing user shape consumed by NavTop and NavSide. */
export interface UserProfile {
  id: string
  fullName: string
  email: string
  role: 'owner' | 'admin' | 'user' | 'demo'
}

// ── Request interfaces ──────────────────────────────────────────────────────

/** Payload for `POST /backend/authenticator` (pre-auth step). */
export interface AuthenticatorRequest {
  email: string
}

/** Payload for `POST /auth/login`. */
export interface LoginRequest {
  email: string
  password?: string
}

/** Payload for `POST /auth/register`. */
export interface RegisterRequest {
  email: string
  password: string
  name: string
  surname: string
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** Body of `POST /backend/authenticator` response. */
export interface AuthenticatorResponse {
  entry_url: string
}

/** Body of `GET /user/details` response. */
export interface UserDetailsResponse {
  id: number
  email: string
  name: string
  surname: string
  admin: boolean
  owner: boolean
  demo: boolean
}
