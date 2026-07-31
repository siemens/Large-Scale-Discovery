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

import type { Ref } from 'vue'
import { ref, watch } from 'vue'

// Persistence key and body class, kept identical to the legacy app for parity
const PRESENTATION_STORAGE_KEY = 'presentation'

// Toggle one body class so all presentation-sensitive styles share the same state
const PRESENTATION_BODY_CLASS = 'presentation-mode'

/** PresentationModeControls exposes the shared presentation state and its toggle. */
interface PresentationModeControls {
  isPresentationMode: Ref<boolean>
  toggle(): void
}

/** readInitialPresentationMode reads the persisted flag, defaulting to off. */
function readInitialPresentationMode(): boolean {
  // localStorage can throw in private-mode or sandboxed contexts, so treat any failure as off
  try {
    return localStorage.getItem(PRESENTATION_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

// Module-level singleton state so every caller stays in sync
const isPresentationMode = ref(readInitialPresentationMode())

// Reflect the current state onto <body> and persist it. Runs immediately so the
// body class is correct on first import (e.g. after a reload with mode on)
watch(
  isPresentationMode,
  (active) => {
    document.body.classList.toggle(PRESENTATION_BODY_CLASS, active)

    // Persistence is best-effort because the in-memory state remains functional
    try {
      localStorage.setItem(PRESENTATION_STORAGE_KEY, String(active))
    } catch {
      // Keep presentation mode active when storage is unavailable
    }
  },
  { immediate: true },
)

/**
 * Shared blur/presentation toggle backed by a body class and localStorage.
 *
 * Restores the v1 "eye" toggle: toggles `presentation-mode` on `<body>` so the
 * global rule in App.vue blurs any descendant marked `.sensitive`. Persists
 * under the same `"presentation"` key the legacy app used, so the choice
 * survives reloads.
 */
export function usePresentationMode(): PresentationModeControls {
  // Switch presentation mode on or off
  function toggle(): void {
    isPresentationMode.value = !isPresentationMode.value
  }

  // Expose the reactive flag plus its toggle
  return { isPresentationMode, toggle }
}
