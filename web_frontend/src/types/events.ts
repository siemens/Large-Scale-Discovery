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

/** Payload for `POST /admin/events`. */
export interface EventRequest {
  event: string
  since?: string
  group?: boolean
  limit?: number
}

// ── Response interfaces ─────────────────────────────────────────────────────

/** The user details that may accompany an audit event. */
interface EventUser {
  email: string
  name?: string
  surname?: string
  company?: string | null
  department?: string | null
  gender?: string | null
  last_login?: string | null
}

/** A single audit event as returned by `POST /admin/events`. */
export interface EventResponse {
  timestamp: string
  email?: string
  event: string
  event_type?: string
  event_detail?: string
  source?: string
  user?: EventUser | null
}
