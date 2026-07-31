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

import ToastEventBus from 'primevue/toasteventbus'
import { useToast as usePrimeToast } from 'primevue/usetoast'
import { nextTick, onBeforeUnmount, onMounted, type Ref, ref } from 'vue'

import type { ToastMessageOptions } from 'primevue/toast'

/** Extends PrimeVue's message options with a per-message icon override for the custom message slot in App.vue. */
type ToastMessage = ToastMessageOptions & {
  /** PrimeIcon class string, null for no icon, or undefined to use the default severity icon. */
  icon?: string | null
}

/** ToastControls exposes the standard application toast intents and the raw add escape hatch. */
interface ToastControls {
  success(summary: string, detail?: string): void
  error(summary: string, detail?: string): void
  warn(summary: string, detail?: string): void
  brief(summary: string, detail?: string): void
  add(message: ToastMessage): void
}

/** ToastAnnouncementPayload is the untrusted subset read from PrimeVue's event bus. */
interface ToastAnnouncementPayload {
  summary?: unknown
  detail?: unknown
  severity?: unknown
}

/**
 * TOAST_LIFE defines display durations in milliseconds, keyed by intent.
 *
 * Single source of truth for how long each kind of toast stays on screen.
 * Exported for callers that configure custom toast durations.
 */
export const TOAST_LIFE = {
  success: 3500,
  error: 4500,
  warn: 4000,
  /** Brief acknowledgments (copy-to-clipboard, quick confirmations). */
  short: 1500,
} as const

/** Reactive text rendered by App.vue's polite and assertive ARIA live regions. */
export interface ToastAnnouncer {
  polite: Ref<string>
  assertive: Ref<string>
}

/** toastAnnouncementText joins the textual parts of a toast into one announcement. */
function toastAnnouncementText(message: ToastAnnouncementPayload): string {
  return [message.summary, message.detail]
    .filter((part): part is string => typeof part === 'string' && part.trim().length > 0)
    .map((part) => part.trim())
    .join('. ')
}

/**
 * Mirrors PrimeVue's global toast events into screen-reader live regions.
 *
 * PrimeVue renders visible toast content but its custom message slot does not
 * provide reliable live-region behavior. App.vue calls this once at the shell
 * level: errors use the assertive region, while other severities use the polite
 * region. Clearing before the next tick makes repeated identical messages
 * announce again.
 */
export function useToastAnnouncer(): ToastAnnouncer {
  // Store screen-reader announcements that should not interrupt
  const polite = ref('')

  // Store urgent screen-reader announcements
  const assertive = ref('')

  // Mirror visible toast text into the correct live region
  function announce(payload: unknown): void {
    // Ignore event-bus payloads that cannot contain toast fields
    if (!isToastAnnouncementPayload(payload)) return

    // Announce only payloads with visible summary or detail text
    const text = toastAnnouncementText(payload)
    if (!text) return

    // Route errors to the interrupting region and everything else to polite output
    const target = payload.severity === 'error' ? assertive : polite
    const other = payload.severity === 'error' ? polite : assertive
    target.value = ''
    other.value = ''

    // Clear first so screen readers repeat an identical consecutive message
    void nextTick(() => {
      target.value = text
    })
  }

  onMounted(() => ToastEventBus.on('add', announce))
  onBeforeUnmount(() => ToastEventBus.off('add', announce))

  return { polite, assertive }
}

/**
 * Thin wrapper around PrimeVue's `useToast` that pairs each severity with its
 * standard display duration, so call sites no longer repeat the
 * `{ severity, summary, life }` shape (or import the duration constants).
 *
 * Returns one helper per intent plus `add`, the raw PrimeVue method, as an
 * escape hatch for anything the helpers don't cover.
 */
export function useToast(): ToastControls {
  // Use the shared toast service for user feedback
  const toast = usePrimeToast()

  // Wrapper that accepts our extended type; PrimeVue passes extra fields through
  // to the message slot, so the runtime sees them even though they're not declared.
  const add = (msg: ToastMessage): void => toast.add(msg)

  /** Green success toast (standard duration). */
  function success(summary: string, detail?: string): void {
    toast.add({ severity: 'success', summary, detail, life: TOAST_LIFE.success })
  }

  /** Red error toast (longer duration — errors deserve more reading time). */
  function error(summary: string, detail?: string): void {
    toast.add({ severity: 'error', summary, detail, life: TOAST_LIFE.error })
  }

  /** Yellow warning toast. */
  function warn(summary: string, detail?: string): void {
    toast.add({ severity: 'warn', summary, detail, life: TOAST_LIFE.warn })
  }

  /** Brief success acknowledgment (copy-to-clipboard etc.) — success severity, short duration. */
  function brief(summary: string, detail?: string): void {
    toast.add({ severity: 'success', summary, detail, life: TOAST_LIFE.short })
  }

  // Expose the intent helpers plus the raw `add` escape hatch
  return { success, error, warn, brief, add }
}

/** isToastAnnouncementPayload narrows object-shaped event-bus payloads. */
function isToastAnnouncementPayload(value: unknown): value is ToastAnnouncementPayload {
  return typeof value === 'object' && value !== null
}
