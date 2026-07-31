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

import Toast from 'primevue/toast'

import { useToastAnnouncer } from '@/composables/useToast'

// Map generic toast severities to their default visual indicators
const TOAST_ICON_BY_SEVERITY: Readonly<Record<string, string>> = {
  success: 'pi pi-check-circle',
  warn: 'pi pi-exclamation-triangle',
  error: 'pi pi-times-circle',
  info: 'pi pi-info-circle',
}

// Mirror toast content into dedicated screen-reader live regions
const toastAnnouncer = useToastAnnouncer()

// Respect explicit per-message icons before falling back to the severity indicator
function resolveToastIcon(message: { severity?: string; icon?: string | null }): string | null {
  if ('icon' in message) return message.icon ?? null

  return TOAST_ICON_BY_SEVERITY[message.severity ?? ''] ?? null
}
</script>

<template>
  <div
    class="app-toast__announcer"
    data-testid="toast-announcer-polite"
    aria-live="polite"
    aria-atomic="true"
  >
    {{ toastAnnouncer.polite }}
  </div>
  <div
    class="app-toast__announcer"
    data-testid="toast-announcer-assertive"
    role="alert"
    aria-live="assertive"
    aria-atomic="true"
  >
    {{ toastAnnouncer.assertive }}
  </div>

  <Toast position="bottom-right" :on-mouse-enter="() => {}" :on-mouse-leave="() => {}">
    <!-- Preserve per-message icon control while providing severity defaults -->
    <template #message="slotProps">
      <i
        v-if="resolveToastIcon(slotProps.message)"
        :class="['p-toast-message-icon', resolveToastIcon(slotProps.message)]"
        aria-hidden="true"
      />
      <div class="p-toast-message-text">
        <span class="p-toast-summary">{{ slotProps.message.summary }}</span>
        <span v-if="slotProps.message.detail" class="p-toast-detail">
          {{ slotProps.message.detail }}
        </span>
      </div>
    </template>
  </Toast>
</template>

<style scoped>
/* Keep toast announcements available to assistive technology without displaying duplicate text */
.app-toast__announcer {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
