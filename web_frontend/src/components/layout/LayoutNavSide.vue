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
 * <LayoutNavSide> — section sidebar listing the sub-navigation links for the active
 * area (configuration or admin). Emits `navigate` so a parent mobile drawer can
 * close itself after a link is tapped.
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    nav: {
      side: {
        logs: 'Logs',
      },
    },
    a11y: {
      sectionNav: 'Section navigation',
    },
  },
}

// ── Props and emits ──────────────────────────────────────────
const emit = defineEmits<{
  navigate: []
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const route = useRoute()

// ── Reactive state ────────────────────────────────────────────
// Link sets for each section, labels kept reactive so they follow locale changes
const configurationRoutes = [
  {
    to: '/configuration/scopes',
    label: computed(() => t('common.labels.scopes')),
  },
  {
    to: '/configuration/views',
    label: computed(() => t('common.labels.views')),
  },
]

const adminRoutes = [
  {
    to: '/admin/users',
    label: computed(() => t('common.labels.users')),
  },
  {
    to: '/admin/groups',
    label: computed(() => t('common.labels.groups')),
  },
  {
    to: '/admin/databases',
    label: computed(() => t('common.labels.databases')),
  },
  { to: '/admin/logs', label: computed(() => t('nav.side.logs')) },
]

// ── Computed ──────────────────────────────────────────────────
const isConfigurationSection = computed(() => route.path.startsWith('/configuration'))
const isAdminSection = computed(() => route.path.startsWith('/admin'))

const sideLinks = computed(() => {
  if (isConfigurationSection.value) return configurationRoutes
  if (isAdminSection.value) return adminRoutes
  return []
})
</script>

<template>
  <nav
    v-if="sideLinks.length > 0"
    class="layout-nav-side"
    data-testid="nav-side"
    :aria-label="t('a11y.sectionNav')"
  >
    <router-link
      v-for="link in sideLinks"
      :key="link.to"
      :to="link.to"
      class="nav-item"
      :data-testid="`sidenav-${link.to.split('/').pop()}`"
      @click="emit('navigate')"
    >
      {{ link.label.value }}
    </router-link>
  </nav>
</template>

<style scoped>
.layout-nav-side {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  /* Narrow section nav so the main content column gets more width (was 180px,
     reduced ~30%). */
  min-width: 9rem;
  /* Align the first nav entry with the first section header's text. The page
     header (SectionHeader h2, dividing) carries a 2em top margin (≈3.6rem at its
     1.8rem font size); each nav item adds 0.625rem of its own top padding, so
     offset the column by the difference and the first button lines up with the
     heading text. */
  margin-top: 5.8rem;
  /* Keep the section nav in view while the content scrolls (legacy `initSticky`
     on the vertical menu). Stick so its distance to the docked top-nav bar
     equals the initial gap and never shrinks as the page scrolls. Initial gap
     below the bar = body padding-top (1.25rem) + this column's margin-top
     (6.6rem − 0.1rem). Add the docked bar height (40px) so the column pins
     exactly that far below the pinned bar. */
  position: sticky;
  top: calc(40px + 1.25rem + 6.6rem - 0.1rem);
  /* Legacy `secondary vertical pointing menu fluid right`: one continuous gray line
     runs down the whole nav's right edge; each item's active state paints a teal
     segment over it via box-shadow (see below). */
  border-right: 3px solid var(--p-surface-300);
}

/* Sidebar-specific overrides: right-aligned text, no rounding (so the active
   shadow lands flush on the gray border-right), deeper padding, and a right-edge
   box-shadow active marker instead of the default left-border. */
.layout-nav-side :deep(.nav-item) {
  justify-content: flex-end;
  text-align: right;
  padding: 0.8rem;
  border-radius: 0;
  border-left: none;
}

/* Active entry: teal 3px right-edge shadow that paints over the container's gray line */
.layout-nav-side :deep(.nav-item--active),
.layout-nav-side :deep(.nav-item.router-link-active) {
  box-shadow: 3px 0 0 0 var(--p-primary-600);
}
</style>
