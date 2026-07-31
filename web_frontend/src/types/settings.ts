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

import { z } from 'zod'

/** isSafeFrontendHref validates a configured browser link without using the application origin. */
export function isSafeFrontendHref(href: string): boolean {
  // Protocol-relative URLs can silently switch origins and are not accepted
  if (href.startsWith('//')) return false

  try {
    // Resolve application-relative links against a fixed safe origin for protocol validation
    const parsed = new URL(href, 'https://application.invalid')
    if (parsed.username !== '' || parsed.password !== '') return false
    return parsed.protocol === 'https:' || parsed.protocol === 'http:'
  } catch {
    return false
  }
}

// Validate and normalize each backend-configured frontend link
const frontendLinkSchema = z.object({
  icon: z.string(),
  text: z.string(),
  href: z.string().trim().min(1).refine(isSafeFrontendHref),
  target: z
    .string()
    .transform((target): '_self' | '_blank' => (target === '_blank' ? '_blank' : '_self')),
})

/** FrontendLink describes one validated backend-configured application link. */
export type FrontendLink = z.infer<typeof frontendLinkSchema>

// Validate the public settings envelope before it crosses into reactive application state
const settingsResponseSchema = z.object({
  development_login: z.boolean().optional(),
  credentials_registration: z.boolean().optional(),
  frontend_links: z.record(z.string(), z.array(frontendLinkSchema)).optional(),
  token_expiry_max_days: z.number().int().positive().optional(),
})

/** SettingsResponse is the validated body of the public settings endpoint. */
export type SettingsResponse = z.infer<typeof settingsResponseSchema>

/** parseSettingsResponse validates untrusted public settings and normalizes safe link targets. */
export function parseSettingsResponse(value: unknown): SettingsResponse {
  return settingsResponseSchema.parse(value)
}
