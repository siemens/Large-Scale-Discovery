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

/** Renders a circular “or” badge between two adjacent actions. */

import { useI18n } from 'vue-i18n'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    or: 'or',
  },
}

// ── Props ─────────────────────────────────────────────────────
const props = defineProps<{ textOr?: string }>()

const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
</script>

<template>
  <span class="or-connector" data-testid="or-connector" aria-hidden="true">
    <span class="or-connector__badge">{{ props.textOr ?? t('or') }}</span>
  </span>
</template>

<style scoped>
/* Zero-width anchor stretched to the button row so the badge can center over the seam */
.or-connector {
  position: relative;
  flex: 0 0 0;
  align-self: stretch;
  z-index: 3;
  pointer-events: none;
  background-color: var(--p-surface-0);
  padding: 0.35rem;
}

.or-connector__badge {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.78571429em;
  height: 1.78571429em;
  border-radius: 50%;
  background-color: var(--p-content-background);
  color: var(--p-text-muted-color);
  font-weight: 600;
  line-height: 1;
  text-transform: lowercase;
  padding-bottom: 0.1rem;
  pointer-events: auto;
}
</style>
