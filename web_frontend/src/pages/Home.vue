<script setup lang="ts">
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

/**
 * Home page.
 *
 * Read-only dashboard. Sections:
 *   1. Logins            (admin only)         — polls /users every 10 000 ms
 *   2. Scan Agents       (all users)          — polls /agents every 19 549 ms
 *   3. Accessible Views  (all users)          — one-shot on mount via /views/granted
 *   4. Access Requests   (admin only)         — one-shot on mount via /admin/events
 *
 * Live data uses REST polling via the shared `usePolling` composable. Granted
 * views are loaded once because access assignments rarely change during a session.
 */

import { storeToRefs } from 'pinia'
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import SectionHeader from '@/components/elements/SectionHeader.vue'
import WidgetAgents from '@/components/widgets/WidgetAgents.vue'
import WidgetChart from '@/components/widgets/WidgetChart.vue'
import WidgetGranted from '@/components/widgets/WidgetGranted.vue'
import WidgetLogins from '@/components/widgets/WidgetLogins.vue'
import { usePolling } from '@/composables/usePolling'
import { useAgentsStore } from '@/stores/agents'
import { useAuthStore } from '@/stores/auth'
import { useEventsStore } from '@/stores/events'
import { useUsersStore } from '@/stores/users'
import { useViewsStore } from '@/stores/views'

import type { EventResponse } from '@/types/events'
import type { WeeklyPoint, WeeklyUsage } from '@/types/usage'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    accessRequestsChart: {
      title: 'Access Requests',
    },
    grantedPanel: {
      title: 'Accessible Views',
      help: 'This section lists all accessible database views. A view is a (possibly limited) representation of a scan.',
    },
    agentsPanel: {
      title: 'Scan Agents',
    },
    recentLogins: {
      title: 'Logins',
    },
    usageChart: {
      axes: {
        accesses: 'Accesses',
      },
    },
  },
}

// Home dashboard polling cadences (ms). The agents interval is deliberately
// uneven to lower the probability of running in parallel with the activity
// update (per the legacy agents.js comment).
const HOME_POLL_RECENT_USERS_MS = 10_000
const HOME_POLL_AGENTS_MS = 19_549

// Select the audit-event family represented by the access chart
const ACCESS_REQUESTS_EVENT = 'Database Password'

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })

const authStore = useAuthStore()
const { isAuthenticated, isAdmin, isDemo, profile } = storeToRefs(authStore)

const usersStore = useUsersStore()
const { recentUsers, isLoaded: areUsersLoaded } = storeToRefs(usersStore)
const agentsStore = useAgentsStore()
const eventsStore = useEventsStore()
const { events } = storeToRefs(eventsStore)
const viewsStore = useViewsStore()

const recentUsersPoller = usePolling(() => usersStore.fetchUsersSilent(), {
  intervalMs: HOME_POLL_RECENT_USERS_MS,
  autoStart: false,
  pauseWhenHidden: true, // The agent poller remains active and keeps the session current
})
const agentsPoller = usePolling(() => agentsStore.fetchAgentsSilent(), {
  intervalMs: HOME_POLL_AGENTS_MS,
  autoStart: false,
  pauseWhenHidden: false,
})

// Group access requests by company and ISO calendar week for the shared chart
const accessRequestsUsage = computed<WeeklyUsage>(() => {
  const eventList = events.value
  if (!eventList || eventList.length === 0) return {}

  // Count requests in the matching company and week bucket
  const seriesByCompany: Record<string, Record<string, number>> = {}
  for (const event of eventList) {
    const company = companyLabel(event)
    const companySeries = (seriesByCompany[company] ??= {})
    const label = isoWeekLabel(event.timestamp)
    companySeries[label] = (companySeries[label] ?? 0) + 1
  }

  // Convert the internal lookup into the chart's array-based series shape
  const result: WeeklyUsage = {}
  for (const [company, weeks] of Object.entries(seriesByCompany)) {
    result[company] = Object.entries(weeks).map(([week, count]): WeeklyPoint => ({ week, count }))
  }
  return result
})

function companyLabel(event: EventResponse): string {
  if (!event.user) return ''

  // Preserve the legacy department grouping for internal accounts
  const { company, department } = event.user
  if (company === 'SIEMENS' && department) return `SIEMENS (${department.split(' ')[0]})`
  return company ?? ''
}

function isoWeekLabel(timestamp: string): string {
  // Shift the UTC date to the Thursday that determines its ISO week year
  const date = new Date(timestamp)
  const workingDate = new Date(
    Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()),
  )
  const day = workingDate.getUTCDay() || 7
  workingDate.setUTCDate(workingDate.getUTCDate() + 4 - day)

  // Count full week boundaries from the first day of the ISO week year
  const yearStart = new Date(Date.UTC(workingDate.getUTCFullYear(), 0, 1))
  const week = Math.ceil(((workingDate.getTime() - yearStart.getTime()) / 86_400_000 + 1) / 7)
  return `W${String(week).padStart(2, '0')}`
}

// ── Watchers and lifecycle ────────────────────────────────────
// Start live agent updates and load stable view grants once for authenticated non-demo users
watch(
  [isAuthenticated, isDemo],
  ([authenticated, demo]) => {
    if (authenticated && !demo) {
      agentsPoller.start()
      void viewsStore.fetchGrantedViews()
    } else {
      agentsPoller.stop()
    }
  },
  { immediate: true },
)

watch(
  isAdmin,
  (admin) => {
    if (admin) {
      recentUsersPoller.start()
      void eventsStore.fetchEvents('Database Password')
    } else {
      recentUsersPoller.stop()
    }
  },
  { immediate: true },
)
</script>

<template>
  <!-- 1. Logins (admin only) -->
  <section v-if="isAdmin" class="page-section" data-testid="section-logins">
    <SectionHeader>{{ t('recentLogins.title') }}</SectionHeader>
    <WidgetLogins :users="recentUsers" :current-email="profile?.email" :loading="!areUsersLoaded" />
  </section>

  <!-- 2. Scan Agents (all users — backend filters by role) -->
  <section class="page-section" data-testid="section-agents">
    <SectionHeader>{{ t('agentsPanel.title') }}</SectionHeader>
    <WidgetAgents />
  </section>

  <!-- 3. Accessible Views (all users) -->
  <section class="page-section" data-testid="section-views">
    <SectionHeader
      :title="t('grantedPanel.title')"
      :help-text="t('grantedPanel.help')"
      help-test-id="views-tooltip"
    />
    <WidgetGranted />
  </section>

  <!-- 4. Access Requests chart (admin only) -->
  <section v-if="isAdmin" class="page-section" data-testid="section-access-chart">
    <SectionHeader>{{ t('accessRequestsChart.title') }}</SectionHeader>
    <WidgetChart
      :event="ACCESS_REQUESTS_EVENT"
      :usage="accessRequestsUsage"
      :title="null"
      :loading="events === null"
      :y-axis-label="t('usageChart.axes.accesses')"
      test-id-prefix="access"
    />
  </section>
</template>
