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

/** Guards route changes, reloads, and window closes while a form reports unsaved changes. */
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import DialogConfirm from './DialogConfirm.vue'
import { useUnsavedChangesStore } from '@/stores/unsavedChanges'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    title: 'Discard unsaved changes?',
    message: 'You have unsaved changes. If you leave this page, they will be discarded.',
    keepEditing: 'Keep editing',
    discard: 'Discard',
  },
}

const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Composables and stores ────────────────────────────────────
const store = useUnsavedChangesStore()
const router = useRouter()

// ── Reactive state ────────────────────────────────────────────
const promptVisible = ref(false)

// Holds the pending navigation's resolve callback while the confirm is open
let resolveLeave: ((allow: boolean) => void) | null = null
let removeRouteGuard: (() => void) | null = null

// ── Functions ─────────────────────────────────────────────────
function resolvePendingLeave(allow: boolean): void {
  resolveLeave?.(allow)
  resolveLeave = null
}

function onDiscard(): void {
  resolvePendingLeave(true)
}

function onKeepEditing(): void {
  resolvePendingLeave(false)
}

// Native browser prompt for hard close or reload while the page is dirty
function onBeforeUnload(event: BeforeUnloadEvent): void {
  if (!store.isDirty) return

  // The standard incantation to trigger the browser's native "Leave page?"
  // prompt. The custom string is ignored by modern browsers.
  event.preventDefault()
  event.returnValue = ''
}

// ── Watchers and lifecycle ────────────────────────────────────
watch(
  () => store.isDirty,
  (dirty) => {
    if (dirty || !promptVisible.value) return

    // Forced session loss clears dirty state elsewhere, so release any pending
    // navigation instead of leaving an impossible discard dialog on screen.
    promptVisible.value = false
    resolvePendingLeave(true)
  },
)

onMounted(() => {
  removeRouteGuard = router.beforeEach(() => {
    if (!store.isDirty) return true

    // Do not replace the resolver of an already-open prompt. A second navigation
    // attempt is cancelled while the user decides about the first one.
    if (resolveLeave) return false

    return new Promise<boolean>((resolve) => {
      resolveLeave = resolve
      promptVisible.value = true
    })
  })
  window.addEventListener('beforeunload', onBeforeUnload)
})
onBeforeUnmount(() => {
  removeRouteGuard?.()
  removeRouteGuard = null
  window.removeEventListener('beforeunload', onBeforeUnload)
  resolvePendingLeave(false)
})
</script>

<template>
  <DialogConfirm
    v-model:visible="promptVisible"
    :title="t('title')"
    :text-cancel-label="t('keepEditing')"
    :text-confirm-label="t('discard')"
    :message="t('message')"
    @confirm="onDiscard"
    @cancel="onKeepEditing"
  />
</template>
