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
 * Unified message banner component.
 *
 * Severity drives the color scheme and icon automatically. An optional close
 * button can be enabled via the `closable` prop. Structured content can be
 * supplied through the default slot without rendering raw HTML.
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    close: 'Close',
  },
}

// ── Types ─────────────────────────────────────────────────────
interface AppMessageProps {
  /** Color intent; the matching icon is chosen automatically. */
  severity: 'info' | 'success' | 'warn' | 'error'
  /** Simple message text. Use the default slot for rich HTML content. */
  message?: string
  /** Override the severity's default PrimeIcon. */
  icon?: string
  /** Show the severity icon in the banner's dedicated icon position. */
  showIcon?: boolean
  /** ARIA role. Persistent informational content should use `note`; urgent feedback uses `alert`. */
  role?: 'alert' | 'status' | 'note'
  /** Show a user-operated close action. */
  closable?: boolean
  /** Accessible label for the close action. */
  textCloseLabel?: string
}

// ── Props and emits ───────────────────────────────────────────
const props = withDefaults(defineProps<AppMessageProps>(), { role: 'alert', showIcon: true })

const emit = defineEmits<{
  close: []
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Computed ──────────────────────────────────────────────────
const resolvedIcon = computed(() => {
  if (props.icon) return props.icon
  if (props.severity === 'info') return 'pi pi-info-circle'
  if (props.severity === 'success') return 'pi pi-check-circle'
  if (props.severity === 'warn') return 'pi pi-exclamation-triangle'
  return 'pi pi-times-circle'
})
</script>

<template>
  <div :class="['app-message', `app-message--${severity}`]" :role="role">
    <div class="app-message__body">
      <span v-if="showIcon" :class="['app-message__icon', resolvedIcon]" aria-hidden="true" />
      <div class="app-message__content">
        <p v-if="message">{{ message }}</p>
        <slot />
      </div>
      <button
        v-if="closable"
        type="button"
        class="app-message__close"
        :aria-label="props.textCloseLabel ?? t('close')"
        @click="emit('close')"
      >
        <i class="pi pi-times" aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<style scoped>
/* Own the space after every message without extending its colored banner. */
.app-message {
  padding-bottom: var(--app-message-padding-bottom, 1rem);
}

/* A terminal message does not need trailing space inside its parent. */
.app-message:last-child {
  padding-bottom: var(--app-message-terminal-padding-bottom, 0);
}

/* Keep visual padding inside a dedicated body so outer spacing stays transparent. */
.app-message__body {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--p-content-border-radius);
  border: 1px solid;
}

.app-message--info .app-message__body {
  background: var(--app-info-bg);
  border-color: var(--app-info-border);
  color: var(--app-info-text);
}

.app-message--success .app-message__body {
  background: var(--app-success-bg);
  border-color: var(--app-success-border);
  color: var(--app-success-text);
}

.app-message--warn .app-message__body {
  background: var(--app-warn-bg);
  border-color: var(--app-warn-border);
  color: var(--app-warn-text);
}

.app-message--error .app-message__body {
  background: var(--app-error-bg);
  border-color: var(--app-error-border);
  color: var(--app-error-text);
}

.app-message__icon {
  font-size: 1rem;
  flex-shrink: 0;
}

.app-message__content {
  flex: 1;
  min-width: 0;
  font-size: var(--app-font-size-small);
}

.app-message__content :deep(p),
.app-message__content :deep(ul),
.app-message__content :deep(ol) {
  margin: 0;
}

.app-message__content :deep(p + p),
.app-message__content :deep(p + ul),
.app-message__content :deep(p + ol),
.app-message__content :deep(ul + p),
.app-message__content :deep(ol + p) {
  padding-top: 0.25rem;
}

.app-message__close {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 50%;
  color: inherit;
  opacity: 0.7;
  transition: opacity 0.15s;
}

.app-message__close:hover {
  opacity: 1;
}
</style>
