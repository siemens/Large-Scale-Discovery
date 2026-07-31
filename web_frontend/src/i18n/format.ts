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

import { createI18n } from 'vue-i18n'

// Local English defaults keep formatting helpers independent from the global language catalog
const localMessages = {
  'en-US': {
    daysAgo: '1 day ago | {count} days ago',
    justNow: 'just now',
  },
}

// Translate module-owned messages without importing the global catalog
const localI18n = createI18n({ legacy: false, locale: 'en-US', messages: localMessages })

// Keep fixed-format output consistent with the currently supported application locale
const FORMAT_LOCALE = 'en-US'
const ZERO_DATE_PLACEHOLDER = '-'

/** toDate coerces supported date inputs into a valid Date or returns null. */
function toDate(input: Date | string | number | null | undefined): Date | null {
  // A nullish input has no date to coerce
  if (input === null || input === undefined) return null

  // Pass through an existing Date, otherwise let the Date constructor parse it
  const parsed = input instanceof Date ? input : new Date(input)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

// Pad date and time fields to the fixed two-character display width
function pad2(value: number): string {
  return value.toString().padStart(2, '0')
}

/**
 * Render a full local datetime using the established `DD.MM.YYYY | HH:mm:ss` pattern.
 * @example formatDateTime('2026-06-28T14:30:55+02:00') // "28.06.2026 | 14:30:55"
 */
export function formatDateTime(input: Date | string | number | null | undefined): string {
  // An unparseable input renders as an empty string rather than "Invalid Date"
  const date = toDate(input)
  if (!date) return ''

  // Go's zero time is a valid Date object, but should render as an empty marker
  if (date.getFullYear() < 1000) return ZERO_DATE_PLACEHOLDER

  // Build the hardcoded European pattern with a literal `|` separator
  return (
    `${pad2(date.getDate())}.${pad2(date.getMonth() + 1)}.${date.getFullYear()}` +
    ` | ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`
  )
}

/** Render a local date using the established `DD.MM.YYYY` pattern. */
export function formatDate(input: Date | string | number | null | undefined): string {
  // An unparseable input renders as an empty string rather than "Invalid Date"
  const date = toDate(input)
  if (!date) return ''

  // Go's zero time is a valid Date object, but should render as an empty marker
  if (date.getFullYear() < 1000) return ZERO_DATE_PLACEHOLDER

  // Build the hardcoded European date pattern
  return `${pad2(date.getDate())}.${pad2(date.getMonth() + 1)}.${date.getFullYear()}`
}

/** Smallest-to-largest unit thresholds (in seconds) for {@link formatRelative}. */
const RELATIVE_DIVISIONS: Array<{ amount: number; unit: Intl.RelativeTimeFormatUnit }> = [
  { amount: 60, unit: 'second' },
  { amount: 60, unit: 'minute' },
  { amount: 24, unit: 'hour' },
  { amount: 7, unit: 'day' },
  { amount: 4.34524, unit: 'week' },
  { amount: 12, unit: 'month' },
  { amount: Number.POSITIVE_INFINITY, unit: 'year' },
]

/**
 * formatRelative formats a timestamp relative to now ("3 minutes ago", "in 2 days").
 *
 * @example formatRelative(Date.now() - 60_000) // "1 minute ago"
 */
export function formatRelative(
  input: Date | string | number | null | undefined,
  opts: Intl.RelativeTimeFormatOptions = { numeric: 'auto' },
): string {
  // An unparseable input renders as an empty string
  const date = toDate(input)
  if (!date) return ''

  // Go's zero time is a valid Date object, but should render as an empty marker
  if (date.getFullYear() < 1000) return ZERO_DATE_PLACEHOLDER

  // Format relative values with the locale supported by the application
  const formatter = new Intl.RelativeTimeFormat(FORMAT_LOCALE, opts)

  // Calculate the elapsed duration in the selected unit
  let duration = (date.getTime() - Date.now()) / 1000

  // Anything within the last minute is shown as a generic "just now" label
  // rather than an exact second count, which would be noisy and misleading.
  if (Math.abs(duration) < 60) return localI18n.global.t('justNow')

  for (const division of RELATIVE_DIVISIONS) {
    if (Math.abs(duration) < division.amount) {
      return formatter.format(Math.round(duration), division.unit)
    }
    duration /= division.amount
  }
  return formatter.format(Math.round(duration), 'year')
}

/** Options for the shared byte-size formatter. */
export interface ByteFormatOptions {
  /** Unit divisor. Use 1000 for SI values or 1024 for legacy binary-sized values. */
  base?: 1000 | 1024
  /** Remove a trailing `.0` from scaled values. */
  trimTrailingZero?: boolean
}

/** formatBytes renders a byte count using the largest supported unit. */
export function formatBytes(bytes: number, options: ByteFormatOptions = {}): string {
  // Apply caller overrides while keeping the shared SI defaults
  const { base = 1000, trimTrailingZero = false } = options

  // List byte units from smallest to largest
  const units = ['B', 'KB', 'MB', 'GB', 'TB']

  // Start with a finite non-negative byte count
  let amount = Number.isFinite(bytes) && bytes > 0 ? bytes : 0

  // Track which byte unit currently fits the amount
  let unitIndex = 0

  while (amount >= base && unitIndex < units.length - 1) {
    amount /= base
    unitIndex += 1
  }

  if (unitIndex === 0) return `${amount} B`

  // Format the scaled byte value with one decimal place
  const formatted = amount.toFixed(1)
  return `${trimTrailingZero ? formatted.replace(/\.0$/u, '') : formatted} ${units[unitIndex]}`
}

/** isZeroDate identifies missing, invalid, and Go zero-time values. */
export function isZeroDate(value: Date | string | number | undefined | null): boolean {
  if (value === null || value === undefined || value === '') return true

  // Parse the supplied value once before applying it
  const parsed = toDate(value)
  return !parsed || parsed.getFullYear() < 1000
}

/** daysAgo renders a whole-day elapsed label or an empty string for zero dates. */
export function daysAgo(input: string | null | undefined): string {
  if (isZeroDate(input)) return ''

  // Convert the accepted value into a Date instance
  const timestamp = toDate(input)
  if (!timestamp) return ''

  // Calculate the whole-day difference from the current time
  const days = Math.max(0, Math.floor((Date.now() - timestamp.getTime()) / 86_400_000))
  return localI18n.global.t('daysAgo', { count: days }, days)
}
