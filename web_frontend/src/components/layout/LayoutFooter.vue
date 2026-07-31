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
 * <LayoutFooter> — app footer with tooling and repository link columns and a
 * feedback call to action. Link groups are backend-configured via the settings store.
 */
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import { useSettingsStore } from '@/stores/settings'
import { isSafeFrontendHref } from '@/types/settings'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    footer: {
      tooling: {
        heading: 'Security Tooling',
      },
      feedback: {
        heading: 'Help Improve This Project!',
        intro:
          'Large-Scale Discovery is a service from security professionals for security professionals. Let us know what you think!',
      },
      repositories: {
        heading: 'Repositories',
      },
    },
  },
}

// ── Props and emits ──────────────────────────────────────────
const emit = defineEmits<{
  openFeedback: []
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const settingsStore = useSettingsStore()
const { frontendLinks } = storeToRefs(settingsStore)

// ── Computed ──────────────────────────────────────────────────
const toolingLinks = computed(() =>
  (frontendLinks.value.tooling ?? []).filter((link) => isSafeFrontendHref(link.href)),
)
const repositoryLinks = computed(() =>
  (frontendLinks.value.repositories ?? []).filter((link) => isSafeFrontendHref(link.href)),
)
</script>

<template>
  <footer class="layout-footer" data-testid="layout-footer">
    <div class="layout-footer__inner">
      <div class="layout-footer__section">
        <SectionHeader level="h5" variant="plain">{{ t('footer.tooling.heading') }}</SectionHeader>
        <nav :aria-label="t('footer.tooling.heading')">
          <a
            v-for="link in toolingLinks"
            :key="link.href"
            :href="link.href"
            class="layout-footer__link"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ link.text }}
          </a>
        </nav>
      </div>

      <div class="layout-footer__section layout-footer__section--wide">
        <SectionHeader level="h5" variant="plain">{{ t('footer.feedback.heading') }}</SectionHeader>
        <p class="layout-footer__text">{{ t('footer.feedback.intro') }}</p>
        <SegmentedButton
          data-testid="footer-open-feedback"
          icon="pi pi-lightbulb"
          :label="t('common.labels.contact')"
          class="layout-footer__feedback-btn"
          @click="emit('openFeedback')"
        />
      </div>

      <div class="layout-footer__section">
        <SectionHeader level="h5" variant="plain">{{
          t('footer.repositories.heading')
        }}</SectionHeader>
        <nav :aria-label="t('footer.repositories.heading')">
          <a
            v-for="link in repositoryLinks"
            :key="link.href"
            :href="link.href"
            class="layout-footer__link"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ link.text }}
          </a>
        </nav>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.layout-footer {
  background-color: var(--p-surface-50);
  border-top: 1px solid var(--p-surface-300);
  /* Legacy footer inset shadow (css/styles.css `.footer`). */
  box-shadow: 0 2px 3px rgba(0, 0, 0, 0.08) inset;
  padding: 2rem 1rem 3rem;
  margin-top: auto;
}

.layout-footer__inner {
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  gap: 2rem;
  max-width: 85.714286rem;
  margin: 0 auto;
  text-align: center;
}

@media (max-width: 48rem) {
  .layout-footer__inner {
    grid-template-columns: 1fr;
  }
}

/* Footer column headings use the shared <SectionHeader variant="plain"> (legacy
   `h4.ui.header`: borderless, mixed case, no transform). The plain variant emits
   margin:0 and inherits the centered alignment from `.layout-footer__inner`. We only
   add the legacy 0.75rem gap below it here. */
.layout-footer__section :deep(.section-header) {
  margin-bottom: 0.75rem;
}

.layout-footer__link {
  display: block;
  padding: 0.25rem 0;
  font-size: var(--app-font-size-small);
  /* Link-list item color: gray, not teal. The global `a` rule in styles.css
     applied #009590 only to anchors NOT `.item`. Link-list items default to
     rgba(0,0,0,.4) with hover darkening. */
  color: rgba(0, 0, 0, 0.4);
  text-decoration: none;
}

.layout-footer__link:hover {
  color: rgba(0, 0, 0, 0.6);
  text-decoration: none;
}

.layout-footer__text {
  margin-bottom: 1rem;
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
  line-height: 1.5;
}

/* Shrink the button text below the PrimeVue default (important needed to beat
   the framework's own button font-size rule) */
.layout-footer__feedback-btn {
  font-size: var(--app-font-size-small) !important;
}
</style>
