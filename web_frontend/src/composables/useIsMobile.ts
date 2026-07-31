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

import { onBeforeUnmount, onMounted, readonly, type Ref, ref } from 'vue'

// `md` breakpoint (px), the only breakpoint consumed in JS. Lives here — its
// sole consumer — rather than as an exported constant.
const MEDIUM_BREAKPOINT_PX = 768

/**
 * Reactive `matchMedia` wrapper. Returns a readonly ref that tracks whether the
 * given media query currently matches, updating on viewport changes.
 *
 * SSR/no-`matchMedia` safe: falls back to `false` when the API is unavailable.
 *
 * @param query a CSS media query string, e.g. `(max-width: 768px)`
 */
function useMediaQuery(query: string): Readonly<Ref<boolean>> {
  // Store whether the mobile media query currently matches
  const matches = ref(false)

  // matchMedia is undefined in non-browser environments (unit tests / SSR)
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return readonly(matches)
  }

  // Create the browser media-query listener
  const mediaQueryList = window.matchMedia(query)

  // Copy media-query changes into Vue state
  const onMediaQueryChange = (event: MediaQueryList | MediaQueryListEvent): void => {
    matches.value = event.matches
  }

  onMounted(() => {
    matches.value = mediaQueryList.matches
    mediaQueryList.addEventListener('change', onMediaQueryChange)
  })

  onBeforeUnmount(() => {
    mediaQueryList.removeEventListener('change', onMediaQueryChange)
  })

  // Seed synchronously too so the value is correct before mount (e.g. when the
  // composable is read in setup before the first paint)
  matches.value = mediaQueryList.matches

  return readonly(matches)
}

/**
 * Canonical "is this a mobile/tablet viewport?" check, tied to the `md`
 * breakpoint (768px) documented in docs/responsive.md. At or below 768px the
 * app uses its compact chrome (off-canvas drawer nav + hamburger).
 */
export function useIsMobile(): Readonly<Ref<boolean>> {
  return useMediaQuery(`(max-width: ${MEDIUM_BREAKPOINT_PX}px)`)
}
