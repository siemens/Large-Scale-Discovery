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

import { type Ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

/** Reactive state for a page that renders at most one inline action panel. */
export interface PanelDeeplinkState<Kind extends string, Target> {
  activeKind: Ref<Kind | null>
  target: Ref<Target | null>
}

/** Page-specific mapping between inline panel state and the URL hash body. */
export interface PanelDeeplinkOptions<Kind extends string, Target> {
  encode: (kind: Kind, target: Target | null) => string
  restore: (raw: string) => void
}

/** PanelDeeplink converts action-panel state to and from a URL hash. */
export interface PanelDeeplink {
  restoreFromHash: () => void
}

/**
 * Keep an inline action panel synchronized with the URL hash.
 *
 * This follows the LSD frontend contract: writes replace the current history
 * entry, restore happens after page data has loaded, and stale hashes are
 * removed when their action or target cannot be resolved.
 */
export function usePanelDeeplink<Kind extends string, Target>(
  panel: PanelDeeplinkState<Kind, Target>,
  options: PanelDeeplinkOptions<Kind, Target>,
): PanelDeeplink {
  // Use the router to update deep links without reloading the page
  const router = useRouter()

  // Observe the current route hash for incoming deep links
  const route = useRoute()

  watch([panel.activeKind, panel.target], ([kind, target]) => {
    // Replace the current hash so opening inline forms does not pollute browser history
    const next = kind ? `#${options.encode(kind, target)}` : ''
    if (next !== route.hash) {
      void router.replace({ query: route.query, hash: next })
    }
  })

  // Restore the action panel described by the current URL hash
  function restoreFromHash(): void {
    // Read the untrusted text before parsing it
    const raw = route.hash.replace(/^#/, '')
    if (!raw) return

    options.restore(raw)

    if (panel.activeKind.value === null) {
      void router.replace({ query: route.query, hash: '' })
    }
  }

  return { restoreFromHash }
}
