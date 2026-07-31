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

import { defineStore } from 'pinia'
import { ref } from 'vue'

import http from '@/api/http'
import { useListFetcher } from '@/composables/useAsyncState'
import { sameJson } from '@/lib/collections'

import type { EventRequest, EventResponse } from '@/types/events'
import type { WeeklyUsage } from '@/types/usage'
import type { AxiosRequestConfig } from 'axios'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    eventsError: 'Could not load activity.',
    weeklyUsageError: 'Could not load usage.',
  },
}

/** oneYearAgo returns the lower time boundary shared by dashboard event requests. */
function oneYearAgo(): string {
  return new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString()
}

/** useEventsStore owns raw activity events and their weekly dashboard summaries. */
export const useEventsStore = defineStore('events', () => {
  // Remember which event name the next ungrouped request must retrieve
  let requestedEventName = '*'

  // Load raw activity while preserving its list reference when polling returns unchanged data
  const {
    data: events,
    isFetching,
    fetchError,
    fetch: fetchEventList,
    reset: resetEvents,
  } = useListFetcher<EventResponse[]>(
    async (config?: AxiosRequestConfig) => {
      // Describe the bounded event query accepted by both backends
      const payload: EventRequest = {
        event: requestedEventName,
        since: oneYearAgo(),
        group: false,
      }

      // Ask the backend for matching ungrouped activity events
      const response = await http.post<{ body?: { result: EventResponse[] } }>(
        '/admin/events',
        payload,
        config,
      )

      // Return an empty activity list when the backend has no events
      return response.data.body?.result ?? []
    },
    localMessages['en-US'].eventsError,
    { initialData: () => [] },
  )

  // Remember which event names the next grouped request must retrieve
  let requestedEventNames: readonly string[] = []

  // Load one ungrouped event kind without coupling the store to an application page
  async function fetchEvents(eventName: string, config?: AxiosRequestConfig): Promise<void> {
    // Capture the requested event before starting the shared list fetcher
    requestedEventName = eventName
    await fetchEventList(config)
  }

  // Keep weekly aggregates separate because they have an independent request lifecycle
  const weeklyUsage = ref<Record<string, WeeklyUsage>>({})
  const isFetchingWeeklyUsage = ref(false)
  const weeklyUsageError = ref<unknown>(null)

  // Load all requested weekly series, publishing each event's data as soon as it arrives
  async function fetchWeeklyUsage(eventNames: readonly string[]): Promise<void> {
    // Copy the input so a caller cannot mutate an in-flight request definition
    requestedEventNames = [...eventNames]
    isFetchingWeeklyUsage.value = true
    weeklyUsageError.value = null

    try {
      await Promise.all(
        requestedEventNames.map(async (eventName) => {
          try {
            const response = await http.post<{ body?: { result: WeeklyUsage } }>(
              '/admin/events',
              { event: eventName, since: oneYearAgo(), group: true },
              { suppressLoader: true },
            )

            // Publish only changed data so rendered charts keep their current instance
            publishWeeklyUsage(eventName, response.data.body?.result ?? {})
          } catch (error) {
            // Fall back to an empty series so the chart shows the empty state instead of spinning
            publishWeeklyUsage(eventName, {})
            weeklyUsageError.value = error
          }
        }),
      )
    } finally {
      isFetchingWeeklyUsage.value = false
    }
  }

  // Replace one chart series without disturbing charts whose data did not change
  function publishWeeklyUsage(eventName: string, usage: WeeklyUsage): void {
    if (sameJson(weeklyUsage.value[eventName], usage)) return

    weeklyUsage.value = {
      ...weeklyUsage.value,
      [eventName]: usage,
    }
  }

  // Clear both projections together when the active account changes
  function reset(): void {
    // Reset raw event state
    requestedEventName = '*'
    resetEvents()

    // Reset grouped event state and its pending query definition
    requestedEventNames = []
    weeklyUsage.value = {}
    isFetchingWeeklyUsage.value = false
    weeklyUsageError.value = null
  }

  // Expose both event projections and their independent request state
  return {
    events,
    weeklyUsage,
    isFetching,
    fetchError,
    isFetchingWeeklyUsage,
    weeklyUsageError,
    fetchEvents,
    fetchWeeklyUsage,
    reset,
  }
})
