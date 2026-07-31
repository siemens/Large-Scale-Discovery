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
 * 3-color progress bar matching the v1 "multiple progress" style.
 * Segments: teal=done, red=failed, yellow=active (order: [done, failed, active]).
 *
 * Two sizes (both show percentage labels inside each band):
 *   - "md" (default) — taller bar.
 *   - "sm"           — thinner bar (used in the scope list page).
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    ariaLabel: 'Done: {done}, Failed: {failed}, Active: {active}',
  },
}

type ProgressValues = [number, number, number]

const segmentIndexes = [0, 1, 2] as const

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** [done, failed, active] counts. */
    progress: [number, number, number]
    /** Denominator override. Defaults to sum of the three values. */
    total?: number
    /** Visual size variant. */
    size?: 'sm' | 'md'
  }>(),
  { size: 'md' },
)

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Computed ──────────────────────────────────────────────────
const values = computed<ProgressValues>(() => [
  normalizeProgressValue(props.progress[0]),
  normalizeProgressValue(props.progress[1]),
  normalizeProgressValue(props.progress[2]),
])

const denominator = computed(() => {
  const sum = values.value[0] + values.value[1] + values.value[2]
  const total = props.total
  const configuredTotal =
    typeof total === 'number' && Number.isFinite(total) && total > 0 ? total : 0
  return Math.max(configuredTotal, sum)
})

// Visual widths for all three segments, computed together so constraints are
// applied globally:
//   1. Total fill is capped by the nonlinear scale (ratio^0.4): small sums
//      leave lots of gray, large sums leave little, 100% fills exactly.
//   2. Any non-zero segment is boosted to a minimum visual width of 5% so its
//      label is always legible.
//   3. The budget for that boost is taken proportionally from larger segments,
//      so the total never exceeds the nonlinear cap.
const segmentWidths = computed((): ProgressValues => {
  if (denominator.value === 0) return [0, 0, 0]

  const sum = values.value[0] + values.value[1] + values.value[2]
  const ratio = Math.min(sum / denominator.value, 1)
  const budget = Math.pow(ratio, 0.4) * 100

  // Raw proportional widths within the budget
  const rawWidths: ProgressValues = [
    progressWidth(values.value[0], budget),
    progressWidth(values.value[1], budget),
    progressWidth(values.value[2], budget),
  ]

  // Define min segment widths
  const minSingleDigit = 16
  const minDoubleDigit = 24

  // Boost segments below their minimum, accumulate the deficit.
  // Double-digit labels (≥10%) need more room than single-digit ones.
  let deficit = 0
  const boosted: ProgressValues = [0, 0, 0]
  const mins: ProgressValues = [0, 0, 0]
  for (const index of segmentIndexes) {
    const width = rawWidths[index]
    const minimum = toRawPercent(values.value[index]) >= 10 ? minDoubleDigit : minSingleDigit
    mins[index] = minimum

    if (width === 0) continue
    if (width < minimum) {
      deficit += minimum - width
      boosted[index] = minimum
      continue
    }
    boosted[index] = width
  }

  // Recover the deficit from segments larger than their minimum, proportionally
  let reducible = 0
  for (const index of segmentIndexes) {
    reducible += Math.max(0, boosted[index] - mins[index])
  }
  if (reducible > 0) {
    const scale = Math.max(0, 1 - deficit / reducible)
    for (const index of segmentIndexes) {
      if (boosted[index] > mins[index]) {
        boosted[index] = mins[index] + (boosted[index] - mins[index]) * scale
      }
    }
  }

  return boosted
})

// ── Functions ─────────────────────────────────────────────────
function normalizeProgressValue(value: number): number {
  return Number.isFinite(value) && value > 0 ? value : 0
}

function progressWidth(value: number, budget: number): number {
  return value > 0 ? (value / denominator.value) * budget : 0
}

function toRawPercent(value: number): number {
  if (denominator.value === 0) return 0
  return Math.min((value / denominator.value) * 100, 100)
}

// Rounded percentage label using the true data ratio, blank for empty segments
function toPercentLabel(value: number): string {
  const percentage = toRawPercent(value)
  if (percentage === 0) return ''
  return `${Math.round(percentage)}%`
}
</script>

<template>
  <div
    class="progress-bar-3"
    data-testid="progress-bar-3"
    :class="{ 'progress-bar-3--sm': size === 'sm' }"
    role="progressbar"
    :aria-valuemin="0"
    :aria-valuemax="denominator"
    :aria-valuenow="values[0]"
    :aria-label="t('ariaLabel', { done: values[0], failed: values[1], active: values[2] })"
  >
    <div
      v-if="segmentWidths[0] > 0"
      class="progress-bar-3__bar progress-bar-3__bar--teal"
      data-testid="progress-bar-segment"
      :style="{ width: segmentWidths[0].toFixed(1) + '%' }"
    >
      <span class="progress-bar-3__label">{{ toPercentLabel(values[0]) }}</span>
    </div>
    <div
      v-if="segmentWidths[1] > 0"
      class="progress-bar-3__bar progress-bar-3__bar--red"
      data-testid="progress-bar-segment"
      :style="{ width: segmentWidths[1].toFixed(1) + '%' }"
    >
      <span class="progress-bar-3__label">{{ toPercentLabel(values[1]) }}</span>
    </div>
    <div
      v-if="segmentWidths[2] > 0"
      class="progress-bar-3__bar progress-bar-3__bar--yellow"
      data-testid="progress-bar-segment"
      :style="{ width: segmentWidths[2].toFixed(1) + '%' }"
    >
      <span class="progress-bar-3__label">{{ toPercentLabel(values[2]) }}</span>
    </div>
  </div>
</template>

<style scoped>
.progress-bar-3 {
  display: flex;
  width: 100%;
  height: 1.8rem;
  border-radius: var(--p-border-radius-sm);
  overflow: hidden;
  background: var(--p-surface-200);
}

.progress-bar-3--sm {
  height: 1rem;
  min-width: 6rem;
}

.progress-bar-3__bar {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 2px;
  transition: width 0.2s ease;
  overflow: hidden;
}

.progress-bar-3__label {
  font-size: var(--app-font-size-small);
  font-weight: 700;
  color: var(--app-color-offwhite);
  white-space: nowrap;
}

/* Segment colors: teal = done, red = failed, yellow = active. */
.progress-bar-3__bar--teal {
  background: var(--p-primary-400);
}
.progress-bar-3__bar--red {
  background: var(--p-red-500);
}
.progress-bar-3__bar--yellow {
  background: var(--p-yellow-500);
}
</style>
