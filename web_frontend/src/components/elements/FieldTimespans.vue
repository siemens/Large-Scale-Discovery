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
 * FieldTimespans — editable list of scan-time windows (day + time ranges).
 *
 * Layout: 2 timespans per row (grid). Each timespan card shows start-day,
 * start-time, end-day, end-time with icons and FloatLabel "on" labels.
 * A fully filled timespan is highlighted teal; incomplete ones appear muted.
 * A trailing empty/dummy row always exists so the user can add without a button.
 */
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import FloatLabel from 'primevue/floatlabel'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import Select from 'primevue/select'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import FieldHint from '@/components/elements/FieldHint.vue'
import { registerFormValidator } from '@/composables/useFormValidation'

import type { DatePickerBlurEvent } from 'primevue/datepicker'
import type { EditableTimespan, ScanTimespan } from '@/types/scopes'

// Local English defaults keep invariant timespan controls with the component that renders them
const localMessages = {
  'en-US': {
    day: 'Weekday',
    empty: 'Without timespan definition scans will run 24/7.',
    endTime: 'End',
    invalid: 'Start and end day/time must not be identical.',
    remove: 'Remove',
    startTime: 'Start',
  },
}

// ── Props ─────────────────────────────────────────────────────
const props = defineProps<{
  /** Raw timespans from the API (no client-side ids). Reassign to re-seed. */
  initialTimespans: readonly ScanTimespan[] | null
  textDayOptions: ReadonlyArray<{ label: string; value: number }>
}>()

const { t } = useI18n({ inheritLocale: true, messages: localMessages })

// ── Composables ───────────────────────────────────────────────

// ── Reactive state ────────────────────────────────────────────

let timespanIdCounter = 0
const timespans = ref<EditableTimespan[]>([])

// ── Computed ──────────────────────────────────────────────────
// Rows where all four fields are filled — only these are submitted.
const validTimespans = computed(() =>
  timespans.value.filter(
    (ts) => ts.startDay !== -1 && ts.startTime !== '' && ts.endDay !== -1 && ts.endTime !== '',
  ),
)

// PrimeVue's Select expects a mutable options array. Keep the public prop
// readonly and provide the control with an isolated shallow copy.
const dayOptions = computed(() => [...props.textDayOptions])

// True when a complete timespan has identical start and end (day + time).
const hasInvalidTimespans = computed(() =>
  validTimespans.value.some((ts) => ts.startDay === ts.endDay && ts.startTime === ts.endTime),
)

// Let the surrounding form validate this composite field without inspecting child state
registerFormValidator(() => !hasInvalidTimespans.value)

// ── Functions ─────────────────────────────────────────────────
function timeStringToDate(value: string): Date | null {
  const match = /^(\d{2}):(\d{2})$/.exec(value)
  if (!match) return null

  const hours = Number(match[1])
  const minutes = Number(match[2])
  if (hours > 23 || minutes > 59) return null

  const date = new Date()
  date.setHours(hours, minutes, 0, 0)
  return date
}

function dateToTimeString(value: Date | null): string {
  if (!value) return ''
  const h = String(value.getHours()).padStart(2, '0')
  const m = String(value.getMinutes()).padStart(2, '0')
  return `${h}:${m}`
}

function datePickerValueToTimeString(value: unknown): string {
  return dateToTimeString(value instanceof Date ? value : null)
}

// DatePicker with manualInput clears partial typed values on blur (internal validation).
// Recover the raw text ourselves so typed values like "08:30" are not silently lost.
function onTimeBlur(id: number, field: 'startTime' | 'endTime', event: DatePickerBlurEvent): void {
  const input = event.originalEvent.target
  if (!(input instanceof HTMLInputElement)) return
  const raw = input.value.trim()
  const match = raw.match(/^(\d{1,2}):(\d{2})$/)
  if (!match) return
  const h = Number(match[1])
  const m = Number(match[2])
  if (h > 23 || m > 59) return
  updateTimespan(id, field, `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`)
}

function isComplete(ts: EditableTimespan): boolean {
  return ts.startDay !== -1 && ts.startTime !== '' && ts.endDay !== -1 && ts.endTime !== ''
}

function isInvalid(ts: EditableTimespan): boolean {
  return isComplete(ts) && ts.startDay === ts.endDay && ts.startTime === ts.endTime
}

function generateTimespanId(): number {
  return ++timespanIdCounter
}

function createEmptyTimespan(): EditableTimespan {
  return {
    id: generateTimespanId(),
    startDay: -1,
    startTime: '',
    endDay: -1,
    endTime: '',
  }
}

function ensureEmptyTimespan(): void {
  const hasDummy = timespans.value.some((ts) => !isComplete(ts))
  if (!hasDummy) timespans.value.push(createEmptyTimespan())
}

function removeTimespan(id: number): void {
  timespans.value = timespans.value.filter((entry) => entry.id !== id)
  ensureEmptyTimespan()
}

function updateTimespan<K extends Exclude<keyof EditableTimespan, 'id'>>(
  id: number,
  field: K,
  value: EditableTimespan[K],
): void {
  const timespan = timespans.value.find((entry) => entry.id === id)
  if (timespan) {
    timespan[field] = value
    ensureEmptyTimespan()
  }
}

// ── Watchers ──────────────────────────────────────────────────
watch(
  () => props.initialTimespans,
  (newVal) => {
    timespanIdCounter = 0

    // Backend sends startDay/endDay as strings; coerce to numbers so Select option matching works.
    timespans.value = (newVal ?? []).map((ts) => ({
      ...ts,
      startDay: Number(ts.startDay),
      endDay: Number(ts.endDay),
      id: generateTimespanId(),
    }))
    ensureEmptyTimespan()
  },
  { immediate: true, flush: 'sync' },
)

// ── Expose ────────────────────────────────────────────────────
defineExpose({ timespans, validTimespans, hasInvalidTimespans })
</script>

<template>
  <!-- 2-column grid: each wrapper occupies one cell, card + hint stack inside it -->
  <div class="field-timespans__grid">
    <div v-for="ts in timespans" :key="ts.id" class="field-timespans__item">
      <div
        class="field-timespans__card"
        :class="
          isInvalid(ts)
            ? 'field-timespans__card--invalid'
            : isComplete(ts)
              ? 'field-timespans__card--set'
              : 'field-timespans__card--unset'
        "
      >
        <!-- Start group: weekday + time -->
        <div class="field-timespans__group">
          <!-- Start weekday -->
          <FloatLabel variant="on" class="field-timespans__float">
            <div class="field-timespans__select-wrap">
              <i class="pi pi-hourglass field-timespans__select-icon" aria-hidden="true" />
              <Select
                :input-id="`ts-startday-${ts.id}`"
                :model-value="ts.startDay === -1 ? null : ts.startDay"
                :options="dayOptions"
                option-label="label"
                option-value="value"
                scroll-height="20rem"
                :aria-label="t('day')"
                class="field-timespans__select"
                @update:model-value="updateTimespan(ts.id, 'startDay', $event ?? -1)"
              />
            </div>
            <label :for="`ts-startday-${ts.id}`">{{ t('day') }}</label>
          </FloatLabel>

          <!-- Start time -->
          <FloatLabel variant="on" class="field-timespans__float field-timespans__float--time">
            <IconField>
              <InputIcon class="pi pi-clock" />
              <DatePicker
                :input-id="`ts-starttime-${ts.id}`"
                :model-value="timeStringToDate(ts.startTime)"
                time-only
                hour-format="24"
                manual-input
                :aria-label="t('startTime')"
                class="field-timespans__time"
                @update:model-value="
                  updateTimespan(ts.id, 'startTime', datePickerValueToTimeString($event))
                "
                @blur="onTimeBlur(ts.id, 'startTime', $event)"
              />
            </IconField>
            <label :for="`ts-starttime-${ts.id}`">{{ t('startTime') }}</label>
          </FloatLabel>
        </div>

        <!-- Separator arrow -->
        <i class="pi pi-arrow-right field-timespans__arrow" aria-hidden="true" />

        <!-- End group: weekday + time -->
        <div class="field-timespans__group">
          <!-- End weekday -->
          <FloatLabel variant="on" class="field-timespans__float">
            <div class="field-timespans__select-wrap">
              <i class="pi pi-hourglass field-timespans__select-icon" aria-hidden="true" />
              <Select
                :input-id="`ts-endday-${ts.id}`"
                :model-value="ts.endDay === -1 ? null : ts.endDay"
                :options="dayOptions"
                option-label="label"
                option-value="value"
                scroll-height="20rem"
                :aria-label="t('day')"
                class="field-timespans__select"
                @update:model-value="updateTimespan(ts.id, 'endDay', $event ?? -1)"
              />
            </div>
            <label :for="`ts-endday-${ts.id}`">{{ t('day') }}</label>
          </FloatLabel>

          <!-- End time -->
          <FloatLabel variant="on" class="field-timespans__float field-timespans__float--time">
            <IconField>
              <InputIcon class="pi pi-clock" />
              <DatePicker
                :input-id="`ts-endtime-${ts.id}`"
                :model-value="timeStringToDate(ts.endTime)"
                time-only
                hour-format="24"
                manual-input
                :aria-label="t('endTime')"
                class="field-timespans__time"
                @update:model-value="
                  updateTimespan(ts.id, 'endTime', datePickerValueToTimeString($event))
                "
                @blur="onTimeBlur(ts.id, 'endTime', $event)"
              />
            </IconField>
            <label :for="`ts-endtime-${ts.id}`">{{ t('endTime') }}</label>
          </FloatLabel>
        </div>

        <!-- Remove button -->
        <div class="field-timespans__remove-wrap">
          <Button
            type="button"
            icon="pi pi-times"
            text
            :severity="isComplete(ts) ? 'contrast' : 'danger'"
            size="small"
            :aria-label="t('remove')"
            @click="removeTimespan(ts.id)"
          />
        </div>
      </div>
      <FieldHint v-if="isInvalid(ts)" :error="t('invalid')" />
      <FieldHint
        v-else-if="!isComplete(ts) && validTimespans.length === 0"
        :text="t('empty')"
        severity="info"
      />
    </div>
  </div>
</template>

<style scoped>
/* 2-per-row grid; on narrow viewports collapses to 1 column */
.field-timespans__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

@media (max-width: 860px) {
  .field-timespans__grid {
    grid-template-columns: 1fr;
  }
}

/* Card: styled like an input group — border color communicates state */
.field-timespans__card {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 0.6rem 0.4rem 0.6rem;
  border-radius: var(--p-content-border-radius);
  border: 1px solid var(--p-content-border-color);
  transition:
    border-color 0.15s,
    background-color 0.15s;
}

/* Set (fully filled) — teal accent border */
.field-timespans__card--set {
  border-color: var(--p-primary-500);
  background: color-mix(in srgb, var(--p-primary-500) 6%, var(--p-surface-0));
}

/* Unset/incomplete — muted surface background */
.field-timespans__card--unset {
  background: var(--p-surface-50);
}

/* Invalid (start === end) — red accent border */
.field-timespans__card--invalid {
  border-color: var(--p-red-500);
  background: color-mix(in srgb, var(--p-red-500) 6%, var(--p-surface-0));
}

/* Start group and end group: weekday + time side by side with tight gap */
.field-timespans__group {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex: 1;
  min-width: 0;
}

/* Each FloatLabel stretches to fill its slot */
.field-timespans__float {
  flex: 1;
  min-width: 0;
}

/* Time field is narrower than the weekday select */
.field-timespans__float--time {
  flex: 0 0 5.8rem;
}

/* Select fills its FloatLabel container */
.field-timespans__select {
  width: 100%;
}

/* Time input fills its FloatLabel container */
.field-timespans__time {
  width: 100%;
}

/* Remove button: compact size so it doesn't dominate the card */
.field-timespans__remove-wrap {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.field-timespans__remove-wrap :deep(.p-button) {
  padding: 0.2rem;
  width: 1.4rem;
  height: 1.4rem;
  color: var(--p-primary-500) !important;
}

.field-timespans__remove-wrap :deep(.p-button:hover) {
  color: var(--p-primary-500) !important;
  background: color-mix(in srgb, var(--p-primary-500) 10%, transparent) !important;
}

/* Arrow separator between start and end groups */
.field-timespans__arrow {
  flex-shrink: 0;
  color: var(--p-text-muted-color);
  font-size: 0.75rem;
}

/* Wrapper that stacks the card and its optional error hint vertically within the grid cell */
.field-timespans__item {
  display: flex;
  flex-direction: column;
}

/* FloatLabel renders a bare <label> with no PrimeVue class.
   PrimeVue uses inset-inline-start (logical property) for horizontal position — override to
   align the label with the icon-padded input text (2.2rem). Vertical nudge corrects a 1px
   optical offset in PrimeVue's default top: 50% baseline. */
.field-timespans__float :deep(label) {
  inset-inline-start: 2.2rem !important;
  top: calc(50% - 1px);
}

/* Select icon wrapper: positions the icon absolutely inside the Select trigger,
   mirroring how PrimeVue IconField works for InputText/InputMask. */
.field-timespans__select-wrap {
  position: relative;
  display: flex;
  width: 100%;
}

.field-timespans__select-icon {
  position: absolute;
  top: 50%;
  left: 0.75rem;
  transform: translateY(-50%);
  z-index: 1;
  color: var(--p-iconfield-icon-color);
  pointer-events: none;
}

/* Indent the Select's label text to clear the icon. */
.field-timespans__select :deep(.p-select-label) {
  padding-left: 2.2rem;
}

/* Indent the time input text to clear the icon.
   IconField applies padding-inline-start via its own stylesheet; override it explicitly. */
.field-timespans__time :deep(.p-datepicker-input) {
  padding-inline-start: 2.2rem !important;
}
</style>
