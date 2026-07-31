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
import Button from 'primevue/button'
import SplitButton from 'primevue/splitbutton'

/**
 * PrimeVue action button with a fixed, subtly shaded leading icon rail.
 *
 * Attributes, listeners, severity, loading state, ARIA state, and test hooks
 * fall through to the underlying PrimeVue component. Set split for actions
 * that also expose a PrimeVue SplitButton menu.
 */
defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{ split?: boolean }>(), { split: false })
</script>

<template>
  <SplitButton
    v-if="props.split"
    v-bind="$attrs"
    class="segmented-button segmented-button--split"
  />
  <Button v-else v-bind="$attrs" class="segmented-button segmented-button--button" />
</template>

<style scoped>
.segmented-button--button,
.segmented-button--split :deep(.p-splitbutton-button) {
  position: relative;
  overflow: hidden;
  justify-content: center;
  padding-inline: 4.07142857em 1.5em;
}

.segmented-button--button :deep(.p-button-icon),
.segmented-button--button :deep(.p-button-loading-icon),
.segmented-button--split :deep(.p-splitbutton-button > .p-button-icon),
.segmented-button--split :deep(.p-splitbutton-button > .p-button-loading-icon) {
  position: absolute;
  inset-block: 0;
  inset-inline-start: 0;
  display: grid;
  width: 2.57142857em;
  margin: 0;
  place-items: center;
  background: color-mix(in srgb, currentcolor 6%, transparent);
}

.segmented-button--button :deep(.p-button-label),
.segmented-button--split :deep(.p-splitbutton-button > .p-button-label) {
  width: auto;
  text-align: center;
}
</style>
