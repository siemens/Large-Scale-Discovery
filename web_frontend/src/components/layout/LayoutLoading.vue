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

/** Renders a shared loading state, empty state, or the supplied content. */

import { useI18n } from 'vue-i18n'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    loading: 'Loading...',
  },
}

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    loading: boolean
    loadingIcon?: string
    loadingMessage?: string
    empty: boolean
    emptyIcon?: string
    emptyMessage: string
    emptyTestId?: string
  }>(),
  {
    loadingIcon: 'pi pi-spinner pi-spin',
    loadingMessage: undefined,
    emptyIcon: 'pi pi-inbox',
  },
)

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
</script>

<template>
  <div
    v-if="loading"
    class="layout-loading__placeholder"
    role="status"
    :aria-label="props.loadingMessage ?? t('loading')"
  >
    <i :class="loadingIcon" aria-hidden="true" />
    <span>{{ props.loadingMessage ?? t('loading') }}</span>
  </div>

  <div
    v-else-if="empty"
    class="layout-loading__placeholder"
    :data-testid="emptyTestId"
    role="status"
  >
    <i :class="emptyIcon" aria-hidden="true" />
    <span>{{ emptyMessage }}</span>
  </div>

  <slot v-else />
</template>

<style scoped>
.layout-loading__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem;
  color: var(--p-text-muted-color);
  font-style: italic;
}
</style>
