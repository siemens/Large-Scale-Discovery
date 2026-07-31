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

/** CycleProgressTuple represents the done, failed, and active counts of a scan cycle. */
export type CycleProgressTuple = [number, number, number]

/** toCycleProgress normalizes raw cycle counts for the shared progress bar. */
export function toCycleProgress(
  done: number | null | undefined,
  failed: number | null | undefined,
  active: number | null | undefined,
): CycleProgressTuple {
  return [Math.round(done ?? 0), Math.round(failed ?? 0), Math.round(active ?? 0)]
}

/**
 * sameJson compares JSON-shaped values structurally without depending on object key order.
 *
 * Arrays remain order-sensitive. The function intentionally supports only the
 * acyclic, JSON-compatible objects and primitives returned by the REST API.
 */
export function sameJson(left: unknown, right: unknown): boolean {
  // Equal primitives and shared references need no further traversal
  if (Object.is(left, right)) return true

  // Arrays must have the same length and values in the same order
  if (Array.isArray(left) || Array.isArray(right)) {
    if (!Array.isArray(left) || !Array.isArray(right) || left.length !== right.length) return false
    return left.every((value, index) => sameJson(value, right[index]))
  }

  // Non-null objects compare by their own keys, regardless of insertion order
  if (!isJsonObject(left) || !isJsonObject(right)) return false
  const leftKeys = Object.keys(left)
  const rightKeys = Object.keys(right)
  if (leftKeys.length !== rightKeys.length) return false

  return leftKeys.every((key) => Object.hasOwn(right, key) && sameJson(left[key], right[key]))
}

/** stringifyUnknown converts an unknown value into stable display text without throwing. */
export function stringifyUnknown(value: unknown, fallback = ''): string {
  if (value === null || value === undefined) return fallback
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean' || typeof value === 'bigint') {
    return `${value}`
  }
  if (typeof value === 'symbol') return value.description ?? fallback
  if (typeof value === 'function') return fallback

  try {
    return JSON.stringify(value) ?? fallback
  } catch {
    return fallback
  }
}

/** sameStringSet compares string lists without treating selection order as meaningful. */
export function sameStringSet(left: readonly string[], right: readonly string[]): boolean {
  // Compare unique values so duplicate input cannot distort set equality
  const leftValues = new Set(left)
  const rightValues = new Set(right)
  if (leftValues.size !== rightValues.size) return false

  return [...leftValues].every((value) => rightValues.has(value))
}

/** compareDateTimes orders date-time strings chronologically with a stable lexical fallback. */
export function compareDateTimes(left: string, right: string): number {
  // Preserve deterministic ordering when either value is not a parseable date
  const leftTime = Date.parse(left)
  const rightTime = Date.parse(right)
  if (Number.isNaN(leftTime) || Number.isNaN(rightTime)) return left.localeCompare(right)
  return leftTime - rightTime
}

/** groupByKey groups items while preserving the first-seen key order. */
export function groupByKey<T>(
  items: readonly T[],
  keyForItem: (item: T) => string,
  transform?: (group: T[], key: string) => T[],
): Array<readonly [string, T[]]> {
  const groups = new Map<string, T[]>()

  for (const item of items) {
    const key = keyForItem(item)
    const group = groups.get(key)

    if (group) {
      group.push(item)
    } else {
      groups.set(key, [item])
    }
  }

  return Array.from(groups, ([key, group]) => [key, transform ? transform(group, key) : group])
}

/** splitCsv converts comma-separated legacy API values into trimmed entries. */
export function splitCsv(value?: string | null): string[] {
  return (value ?? '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

/** isJsonObject narrows a non-array JSON object for recursive comparison. */
function isJsonObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}
