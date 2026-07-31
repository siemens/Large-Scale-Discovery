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
 * Segmented inline label.
 *
 * Renders a compact train of attached segments: optional left icon, primary
 * text, optional detail text, and optional right icon button.
 */

type LabelVariant = 'info' | 'primary' | 'success' | 'warn' | 'orange' | 'error'

withDefaults(
  defineProps<{
    /** Primary label text. */
    text: string
    /** Optional title/aria text for the primary text segment. Defaults to text. */
    textAlt?: string
    /** Color intent for the label train. */
    variant?: LabelVariant
    /** Optional left PrimeIcon class. */
    icon?: string
    /** Optional secondary text rendered to the right of the primary text. */
    detail?: string
    /** Optional secondary segments appended after detail, in the given order. */
    details?: string[]
    /** Optional title/aria text for the secondary detail text segment. Defaults to detail. */
    detailAlt?: string
    /** Optional max-width for the primary text segment. */
    textMaxWidth?: string
    /** Optional max-width for the secondary detail text segment. */
    detailMaxWidth?: string
    /** Optional right icon-button PrimeIcon class. */
    buttonIcon?: string
    /** Accessible label for the optional right icon button. */
    buttonLabel?: string
    /** Marks the label text as sensitive. */
    sensitive?: boolean
    /** Marks only the secondary detail segments as sensitive. */
    detailSensitive?: boolean
  }>(),
  { sensitive: false, detailSensitive: false },
)

const emit = defineEmits<{
  action: [event: MouseEvent]
}>()
</script>

<template>
  <span
    class="tag-label"
    :class="[
      variant ? `tag-label--${variant}` : null,
      { 'tag-label--has-detail': detail || details?.length },
    ]"
  >
    <span v-if="icon" class="tag-label__segment tag-label__icon" aria-hidden="true">
      <i :class="icon" />
    </span>
    <span
      class="tag-label__segment tag-label__text"
      :class="{ sensitive }"
      :style="{ maxWidth: textMaxWidth ?? '100%' }"
    >
      <span class="tag-label__value" :title="textAlt ?? text" :aria-label="textAlt ?? text">{{
        text
      }}</span>
    </span>
    <span
      v-if="detail"
      class="tag-label__segment tag-label__detail"
      :class="{ sensitive: sensitive || detailSensitive }"
      :style="{ maxWidth: detailMaxWidth ?? '100%' }"
    >
      <span
        class="tag-label__value"
        :title="detailAlt ?? detail"
        :aria-label="detailAlt ?? detail"
        >{{ detail }}</span
      >
    </span>
    <span
      v-for="(detailValue, index) in details"
      :key="`${detailValue}-${index}`"
      class="tag-label__segment tag-label__detail"
      :class="{ sensitive: sensitive || detailSensitive }"
      :style="{ maxWidth: detailMaxWidth ?? '100%' }"
    >
      <span class="tag-label__value" :title="detailValue" :aria-label="detailValue">{{
        detailValue
      }}</span>
    </span>
    <button
      v-if="buttonIcon"
      type="button"
      class="tag-label__segment tag-label__button"
      :aria-label="buttonLabel ?? text"
      @click="emit('action', $event)"
    >
      <i :class="buttonIcon" aria-hidden="true" />
    </button>
  </span>
</template>

<style scoped>
.tag-label {
  --tag-label-background: var(--p-surface-400);
  --tag-label-background-light: var(--p-surface-300);
  --tag-label-background-dark: var(--p-surface-500);
  --tag-label-foreground: var(--p-surface-0);
  --tag-label-focus: var(--p-focus-ring-color);

  display: inline-flex;
  align-items: stretch;
  max-width: 100%;
  margin: 1px 2px 1px 0;
  overflow: hidden;
  border-radius: 0.28571429rem;
  font-size: var(--app-font-size-tiny);
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
  vertical-align: middle;
}

.tag-label--primary {
  --tag-label-background: var(--p-primary-500);
  --tag-label-background-light: var(--p-primary-400);
  --tag-label-background-dark: var(--p-primary-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label--info {
  --tag-label-background: var(--p-blue-500);
  --tag-label-background-light: var(--p-blue-400);
  --tag-label-background-dark: var(--p-blue-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label--success {
  --tag-label-background: var(--p-green-500);
  --tag-label-background-light: var(--p-green-400);
  --tag-label-background-dark: var(--p-green-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label--warn {
  --tag-label-background: var(--p-yellow-500);
  --tag-label-background-light: var(--p-yellow-400);
  --tag-label-background-dark: var(--p-yellow-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label--orange {
  --tag-label-background: var(--p-orange-500);
  --tag-label-background-light: var(--p-orange-400);
  --tag-label-background-dark: var(--p-orange-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label--error {
  --tag-label-background: var(--p-red-500);
  --tag-label-background-light: var(--p-red-400);
  --tag-label-background-dark: var(--p-red-600);
  --tag-label-foreground: var(--p-primary-contrast-color);
}

.tag-label__segment {
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  min-width: 0;
  padding: 0.3em 0.5em;
  color: var(--tag-label-foreground);
  line-height: 1.2;
}

.tag-label__icon {
  justify-content: center;
  background: var(--tag-label-background-light);
}

.tag-label__text {
  background: var(--tag-label-background);
  overflow: hidden;
}

.tag-label__detail {
  background: var(--tag-label-background-dark);
  overflow: hidden;
}

.tag-label__value {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag-label__button {
  justify-content: center;
  border: 0;
  background: var(--tag-label-background);
  cursor: pointer;
  font: inherit;
  line-height: 1.2;
  padding-left: 0;
}

.tag-label--has-detail .tag-label__button {
  background: var(--tag-label-background-dark);
  padding-left: 2px;
}

.tag-label__button:hover,
.tag-label__button:focus-visible {
  filter: brightness(1.08);
}

.tag-label__button:focus-visible {
  outline: 2px solid var(--tag-label-focus);
  outline-offset: -2px;
}

.tag-label__icon .pi,
.tag-label__button .pi {
  font-size: 0.75em;
  line-height: 1;
}
</style>
