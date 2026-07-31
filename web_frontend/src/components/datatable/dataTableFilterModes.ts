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

import { FilterService } from '@primevue/core/api'
import { formatDateTime } from '@/i18n/format'
import { stringifyUnknown } from '@/lib/collections'

/** Names the custom PrimeVue filters registered by the shared table module. */
export const DATA_TABLE_FILTER_MODE = {
  dateText: 'dateText',
  deepContains: 'deepContains',
  noop: 'noop',
} as const

FilterService.register(
  DATA_TABLE_FILTER_MODE.dateText,
  (value: unknown, filter: string | null): boolean => {
    if (filter === null || filter === undefined || filter === '') return true
    if (value === null || value === undefined) return false
    return formatDateTime(toDateValue(value)).toLowerCase().includes(filter.toLowerCase())
  },
)

/**
 * "noop" — always matches, so a header input renders for visual parity but never
 * hides rows. Used for columns whose real value is intentionally absent client-
 * side (masked password), where a CONTAINS filter would zero the grid.
 */
FilterService.register(DATA_TABLE_FILTER_MODE.noop, (): boolean => true)

/**
 * "deepContains" — substring-match API values that are arrays or objects
 * without flattening them into renamed row fields first.
 */
function deepContains(value: unknown, filter: string | null): boolean {
  if (filter === null || filter === undefined || filter.trim() === '') return true
  if (value === null || value === undefined) return false
  return stringifyUnknown(value).toLowerCase().includes(filter.toLowerCase())
}

FilterService.register(DATA_TABLE_FILTER_MODE.deepContains, deepContains)

/**
 * Override the PrimeVue built-in negative match modes ("notContains",
 * "notEquals") which incorrectly return false (hide the row) when the cell
 * value is null or undefined. A null cell does NOT contain the filter string
 * and is NOT equal to it, so it must show through. Registering under the same
 * built-in name is a harmless, documented overwrite (PrimeVue re-uses
 * FilterService.filters internally for row filtering).
 */
FilterService.register('notContains', (value: unknown, filter: string | null): boolean => {
  if (filter === null || filter === undefined || filter.trim() === '') return true

  // Null/undefined cell value: the cell contains nothing, so it passes.
  if (value === null || value === undefined) return true

  return !stringifyUnknown(value).toLowerCase().includes(filter.toLowerCase())
})

FilterService.register('notEquals', (value: unknown, filter: string | null): boolean => {
  if (filter === null || filter === undefined || filter.trim() === '') return true

  // Null/undefined cell value: it is not equal to the filter string, so it passes.
  if (value === null || value === undefined) return true

  return stringifyUnknown(value).toLowerCase() !== filter.toLowerCase()
})

function toDateValue(value: unknown): Date | string | number | null | undefined {
  if (value instanceof Date || typeof value === 'string' || typeof value === 'number') return value

  return undefined
}
