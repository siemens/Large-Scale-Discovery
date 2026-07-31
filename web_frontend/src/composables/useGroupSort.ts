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

import type { Ref } from 'vue'
import { ref } from 'vue'

import { stringifyUnknown } from '@/lib/collections'

/** Sort direction. */
export type SortDirection = 'asc' | 'desc'

/** Reactive sort state plus its toggle/sort helpers, from {@link useGroupSort}. */
export interface GroupSort {
  /** The column currently sorted on. */
  field: Ref<string>
  /** The active sort direction. */
  direction: Ref<SortDirection>
  /**
   * Sort on `field`. Clicking a new column selects it ascending; re-clicking the
   * active column cycles asc → desc → none (cleared back to unsorted).
   */
  toggle: (field: string) => void
  /**
   * Return a new array of `rows` sorted by the active field/direction
   * (case-insensitive). When no column is active, the rows are returned in their
   * original order (a fresh copy).
   */
  sorted: <T>(rows: T[]) => T[]
}

/**
 * Per-table sort state for the hand-rolled sortable tables (one instance per
 * table). Replaces the repeated field/dir ref pairs + toggle + comparator.
 *
 * @param defaultField  the column to sort on initially. Omit (or pass '') to start
 *                      with no column sorted — nothing is highlighted until clicked.
 */
function useGroupSort(defaultField = ''): GroupSort {
  // Store the column currently used for sorting
  const field = ref<string>(defaultField)

  // Store the current sort direction
  const direction = ref<SortDirection>('asc')

  // Switch between ascending and descending order
  function toggle(column: string): void {
    // A new column starts ascending. Re-clicking the active column cycles through
    // three states: asc → desc → none (cleared back to unsorted), then asc again.
    if (field.value !== column) {
      field.value = column
      direction.value = 'asc'
    } else if (direction.value === 'asc') {
      direction.value = 'desc'
    } else {
      field.value = ''
      direction.value = 'asc'
    }
  }

  // Return a sorted copy without changing the source list
  function sorted<T>(rows: T[]): T[] {
    if (field.value === '') return [...rows]

    // Rows are cast to an indexable record because the sort field is a dynamic string
    return [...rows].sort((first, second) => {
      // Read the first row value used by this comparison
      const firstValue = stringifyUnknown(readProperty(first, field.value)).toLowerCase()

      // Read the second row value used by this comparison
      const secondValue = stringifyUnknown(readProperty(second, field.value)).toLowerCase()

      // Compare normalized values before applying the direction
      const comparison = firstValue.localeCompare(secondValue)
      return direction.value === 'asc' ? comparison : -comparison
    })
  }

  // Expose the reactive state plus the toggle/sort helpers
  return { field, direction, toggle, sorted }
}

function readProperty(value: unknown, property: string): unknown {
  // Reflective access keeps the generic row type intact while supporting dynamic columns
  if (typeof value !== 'object' || value === null) return undefined
  return Reflect.get(value, property)
}

/** A family of independent {@link GroupSort} states, one per key. */
export interface KeyedGroupSort {
  /**
   * Return the {@link GroupSort} for `key`, creating it on first access. Each key
   * gets its own field/direction, so sorting one group's table never affects another.
   */
  forGroup: (key: string) => GroupSort
}

/**
 * Independent sort state per group key, for pages that render several sibling
 * tables from one grouped list (e.g. the Views page's per-group view and token
 * tables). A single `useGroupSort` shared across those tables would sort them all
 * together; this keeps each group's sort isolated.
 *
 * @param defaultField  the column each group's table sorts on initially. Omit to
 *                      start with no column sorted.
 */
export function useKeyedGroupSort(defaultField = ''): KeyedGroupSort {
  // Plain (non-reactive) cache so each group key resolves to one stable GroupSort
  const byKey = new Map<string, GroupSort>()

  // Reuse one independent sorter for each group
  function forGroup(key: string): GroupSort {
    // Create this group sorter only when it is first requested
    let sort = byKey.get(key)
    if (!sort) {
      sort = useGroupSort(defaultField)
      byKey.set(key, sort)
    }
    return sort
  }

  // Expose the per-key accessor
  return { forGroup }
}
