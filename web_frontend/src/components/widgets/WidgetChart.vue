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

/** Renders a rolling weekly-usage area chart with ordering and accumulation controls. */
import Button from 'primevue/button'
import Chart from 'primevue/chart'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { APP_CHART_COLORS, APP_CHART_OTHERS_COLOR } from '@/lib/colors'

import type { WeeklyPoint, WeeklyUsage } from '@/types/usage'

interface WeekSlot {
  key: string
  label: string
}

type UsageOrder = 'key' | 'count'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    usageChart: {
      axes: {
        events: 'Events',
        calendarWeek: 'Calendar Week',
      },
      loading: 'Loading chart data...',
      loadError: 'Chart data could not be displayed.',
      empty: 'No usage recorded.',
      actions: {
        orderByCompany: 'Order by company',
        orderByAmount: 'Order by amount',
        disableAccumulation: "Don't accumulate",
        accumulate: 'Accumulate',
      },
    },
  },
}

// ── Props ─────────────────────────────────────────────────────────────────────

const props = withDefaults(
  defineProps<{
    event: string
    usage?: WeeklyUsage
    title?: string | null
    loading?: boolean
    height?: string
    yAxisLabel?: string
    testIdPrefix?: string
  }>(),
  {
    title: undefined,
    loading: false,
    height: '14rem',
    yAxisLabel: undefined,
  },
)

// ── i18n ──────────────────────────────────────────────────────────────────────

const { t } = useI18n({ inheritLocale: true, messages: localMessages })

// ── Chart readiness (dynamic import guard) ────────────────────────────────────

// chart.js/auto stays out of the initial app bundle (lazy chunk), but we kick
// off the import at module-parse time — not deferred to onMounted — so the
// download runs in parallel with the first API requests rather than being
// blocked until the component mounts.  By the time onMounted awaits the
// promise the chunk is usually already cached, eliminating the ~5s stall that
// occurred when all WidgetChart instances awaited the same large chunk serially.
const chartModulePromise = import('chart.js/auto')

const isChartReady = ref(false)
const hasChartLoadFailed = ref(false)
const order = ref<UsageOrder>('key')
const accumulate = ref(false)
let isUnmounted = false

// ── Derived display state ─────────────────────────────────────────────────────

const hasUsage = computed(() => Object.keys(props.usage ?? {}).length > 0)

// Resolved y-axis label: prefer the explicit prop, fall back to the i18n default.
const resolvedYAxisLabel = computed(() => props.yAxisLabel ?? t('usageChart.axes.events'))
const orderActionLabel = computed(() =>
  order.value === 'count'
    ? t('usageChart.actions.orderByCompany')
    : t('usageChart.actions.orderByAmount'),
)
const accumulationActionLabel = computed(() =>
  accumulate.value
    ? t('usageChart.actions.disableAccumulation')
    : t('usageChart.actions.accumulate'),
)

// ── ISO calendar-week helpers ─────────────────────────────────────────────────

/** Returns the ISO week label (e.g. "W07") for any given date. */
function toWeekSlot(date: Date): WeekSlot {
  const working = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()))
  const day = working.getUTCDay() || 7
  working.setUTCDate(working.getUTCDate() + 4 - day)
  const yearStart = new Date(Date.UTC(working.getUTCFullYear(), 0, 1))
  const week = Math.ceil(((working.getTime() - yearStart.getTime()) / 86_400_000 + 1) / 7)
  const label = `W${String(week).padStart(2, '0')}`

  return {
    key: `${working.getUTCFullYear()}-${label}`,
    label,
  }
}

/** Rolling 51-week label array ending at the current week. */
const weekSlots = computed<WeekSlot[]>(() => {
  const result: WeekSlot[] = []
  const now = new Date()
  const start = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()))
  const day = start.getUTCDay() || 7
  start.setUTCDate(start.getUTCDate() - day + 1 - 50 * 7)

  for (let index = 0; index < 51; index++) {
    const date = new Date(start)
    date.setUTCDate(start.getUTCDate() + index * 7)
    result.push(toWeekSlot(date))
  }

  return result
})

/** Looks up a weekly count by label, accepting both "W07" and "2026-W07" formats. */
function pointCount(points: WeeklyPoint[], weekSlot: WeekSlot): number {
  return (
    points.find((point) => point.week === weekSlot.key)?.count ??
    points.find((point) => point.week === weekSlot.label)?.count ??
    0
  )
}

// ── Chart.js data and options ─────────────────────────────────────────────────

const chartData = computed(() => {
  const entries = Object.entries(props.usage ?? {}).map(([key, points]) => ({
    key,
    points,
    total: points.reduce((total, point) => total + point.count, 0),
  }))

  if (order.value === 'key') {
    entries.sort((left, right) => left.key.localeCompare(right.key))
  } else {
    entries.sort((left, right) => right.total - left.total)
  }

  return {
    labels: weekSlots.value.map(({ label }) => label),
    datasets: entries.map((entry, index) => {
      let running = 0
      const data = weekSlots.value.map((weekSlot) => {
        const count = pointCount(entry.points, weekSlot)
        running += count
        return accumulate.value ? running : count
      })

      const color =
        entry.key === 'Others'
          ? APP_CHART_OTHERS_COLOR
          : APP_CHART_COLORS[index % APP_CHART_COLORS.length]

      return {
        label: entry.key,
        data,
        borderColor: color,
        backgroundColor: color,
        borderWidth: 0,
        fill: true,
        tension: 0.3,
        pointRadius: 0,
        pointHoverRadius: 3,
        pointBorderWidth: 0,
      }
    }),
  }
})

const options = computed(() => ({
  elements: { point: { radius: 0 } },
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'nearest' as const, axis: 'x' as const },
  plugins: {
    title: { display: false },
    legend: { position: 'top' as const },
    tooltip: {
      mode: 'index' as const,
      filter: (item: { raw: number }) => item.raw > 0,
    },
  },
  scales: {
    x: {
      stacked: !accumulate.value,
      ticks: { maxTicksLimit: 10 },
      title: { display: true, text: t('usageChart.axes.calendarWeek') },
    },
    y: {
      stacked: true,
      beginAtZero: true,
      ticks: { precision: 0, stepSize: 1 },
      title: { display: true, text: resolvedYAxisLabel.value },
    },
  },
}))

// ── Functions ────────────────────────────────────────────────────────────────

/** Toggles between alphabetical and total-count ordering. */
function toggleOrder(): void {
  order.value = order.value === 'key' ? 'count' : 'key'
}

/** Toggles cumulative rendering for each usage series. */
function toggleAccumulation(): void {
  accumulate.value = !accumulate.value
}

// ── Lifecycle ────────────────────────────────────────────────────────────────

onMounted(async () => {
  try {
    await chartModulePromise

    if (!isUnmounted) {
      isChartReady.value = true
    }
  } catch (error: unknown) {
    if (!isUnmounted) {
      hasChartLoadFailed.value = true
      console.error('Could not load the usage chart', error)
    }
  }
})

onBeforeUnmount(() => {
  isUnmounted = true
})
</script>

<template>
  <article class="widget-chart" :data-testid="testIdPrefix ? `${testIdPrefix}-chart` : undefined">
    <!-- Header: hidden when title is explicitly null, shows title or event name otherwise -->
    <header v-if="title !== null">
      <h4>{{ title ?? event }}</h4>
    </header>

    <!-- Loading state: chart.js not yet ready, or parent signals data is still in flight -->
    <div
      v-if="loading || (!isChartReady && !hasChartLoadFailed)"
      class="widget-chart__placeholder"
      role="status"
    >
      <i class="pi pi-spinner pi-spin" aria-hidden="true" />
      <em>{{ t('usageChart.loading') }}</em>
    </div>

    <!-- Error state: chart.js failed to load and retrying requires a page reload -->
    <div
      v-else-if="hasChartLoadFailed"
      class="widget-chart__error"
      data-testid="usage-chart-error"
      role="alert"
    >
      <i class="pi pi-exclamation-triangle" aria-hidden="true" />
      <em>{{ t('usageChart.loadError') }}</em>
    </div>

    <!-- Populated state: chart canvas + toggle controls -->
    <template v-else-if="hasUsage">
      <div class="widget-chart__canvas">
        <Chart
          type="line"
          :data="chartData"
          :options="options"
          :style="{ height }"
          class="sensitive"
        />
      </div>
      <div class="widget-chart__controls">
        <Button
          v-tooltip.top="orderActionLabel"
          :data-testid="testIdPrefix ? `${testIdPrefix}-toggle-order` : 'usage-toggle-order'"
          type="button"
          :icon="order === 'count' ? 'pi pi-circle-fill' : 'pi pi-circle'"
          class="widget-chart__button"
          :aria-label="orderActionLabel"
          :aria-pressed="order === 'count'"
          @click="toggleOrder"
        />
        <Button
          v-tooltip.top="accumulationActionLabel"
          :data-testid="
            testIdPrefix ? `${testIdPrefix}-toggle-accumulate` : 'usage-toggle-accumulate'
          "
          type="button"
          :icon="accumulate ? 'pi pi-circle-fill' : 'pi pi-circle'"
          class="widget-chart__button"
          :aria-label="accumulationActionLabel"
          :aria-pressed="accumulate"
          @click="toggleAccumulation"
        />
      </div>
    </template>

    <!-- Empty state: data fetched but no usage recorded -->
    <div v-else class="widget-chart__empty" data-testid="usage-empty">
      <i class="pi pi-chart-line" aria-hidden="true" />
      <em>{{ t('usageChart.empty') }}</em>
    </div>
  </article>
</template>

<style scoped>
/* ── Article wrapper ──────────────────────────────────────────── */
.widget-chart {
  min-width: 0;
  margin-bottom: 1rem;
}

.widget-chart:last-child {
  margin-bottom: 0;
}

/* ── Header row ───────────────────────────────────────────────── */
.widget-chart header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.widget-chart h4 {
  margin: 0;
}

/* ── Chart canvas ─────────────────────────────────────────────── */
.widget-chart__canvas {
  width: 100%;
}

/* ── Control strip (order / accumulate toggles) ───────────────── */
.widget-chart__controls {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding-right: 1.5rem;
  margin-top: -0.5rem;
}

/* Compact icon-only toggle buttons.
   !important overrides the PrimeVue Button preset's default sizing and borders,
   which cannot be targeted from here without it. */
.widget-chart__button {
  width: 1rem !important;
  height: 1rem !important;
  padding: 0 !important;
  border: none !important;
  background: none !important;
  font-size: 0.7rem !important;
  color: var(--p-text-muted-color) !important;
}

.widget-chart__button:hover {
  background: none !important;
  color: var(--p-primary-500) !important;
}

.widget-chart__button[aria-pressed='true'] {
  color: var(--p-primary-500) !important;
}

.widget-chart__button[aria-pressed='true']:hover {
  color: var(--p-primary-600) !important;
}

/* ── Loading / empty placeholders ────────────────────────────── */
.widget-chart__placeholder,
.widget-chart__error,
.widget-chart__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 4rem;
  color: var(--p-text-muted-color);
  text-align: center;
}
</style>
