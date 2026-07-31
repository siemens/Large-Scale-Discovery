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

/** One data point in a weekly usage series. */
export interface WeeklyPoint {
  week: string
  count: number
}

/** A map of group key → weekly data points, as expected by WidgetChart. */
export type WeeklyUsage = Record<string, WeeklyPoint[]>
