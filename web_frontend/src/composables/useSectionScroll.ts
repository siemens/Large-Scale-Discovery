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
import { onScopeDispose, watch } from 'vue'

// Leave room for the fixed navigation when scrolling to a section
const NAVIGATION_HEIGHT_PX = 40

// Wait for reveal transitions before measuring the destination section
const REVEAL_DELAY_MS = 220

// Briefly debounce follow-up scrolls while panel data changes
const UPDATE_DELAY_MS = 50

/** scrollToTop smoothly returns the viewport to the top after a completed workflow. */
export function scrollToTop(): void {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

/**
 * useSectionScroll aligns a revealed page section below the sticky navigation after its layout has settled.
 *
 * Pass `activeSection` to react automatically when a panel opens, switches target, or reopens.
 * Use the returned `scrollToSection` for imperative scrolling (e.g. a button click).
 */
export function useSectionScroll(
  getSectionElement: () => HTMLElement | null,
  options: {
    activeSection?: Readonly<Ref<unknown>>
    target?: Readonly<Ref<unknown>>
    openCount?: Readonly<Ref<number>>
  } = {},
): { scrollToSection: (delayMs?: number) => void } {
  // Hold the delayed-scroll timer so stale work can be canceled
  let scrollTimer: ReturnType<typeof setTimeout> | null = null

  // Hold the animation frame used for the final scroll
  let scrollFrame: number | null = null

  // Cancel stale work so rapid section changes produce one final page movement
  function cancelPendingScroll(): void {
    // Clear both scheduling stages because either one may currently be pending
    if (scrollTimer !== null) {
      clearTimeout(scrollTimer)
      scrollTimer = null
    }
    if (scrollFrame !== null) {
      cancelAnimationFrame(scrollFrame)
      scrollFrame = null
    }
  }

  // Measure after the requested delay and the browser's next layout frame
  function scrollToSection(delayMs = 0): void {
    // A newer request always supersedes an older scheduled scroll
    cancelPendingScroll()
    scrollTimer = setTimeout(() => {
      scrollTimer = null
      scrollFrame = requestAnimationFrame(() => {
        scrollFrame = null

        // Resolve the section that should be brought into view
        const sectionElement = getSectionElement()
        if (!sectionElement) return

        // Convert the viewport-relative position to an absolute document offset
        const sectionTop = sectionElement.getBoundingClientRect().top + window.scrollY
        window.scrollTo({ top: sectionTop - NAVIGATION_HEIGHT_PX, behavior: 'smooth' })
      })
    }, delayMs)
  }

  // Reactive action sections scroll after opening, switching target, or reopening
  const { activeSection, target, openCount } = options
  if (activeSection) {
    watch(
      () => [activeSection.value, target?.value ?? null, openCount?.value ?? null],
      ([section], [previousSection]) => {
        if (section === null || section === undefined) {
          cancelPendingScroll()
          return
        }

        // A new reveal needs the transition delay; an already open section only needs a reflow
        const wasClosed = previousSection === null || previousSection === undefined
        scrollToSection(wasClosed ? REVEAL_DELAY_MS : UPDATE_DELAY_MS)
      },
      { flush: 'post' },
    )
  }

  // Component disposal must never leave a delayed page jump behind
  onScopeDispose(cancelPendingScroll)

  return { scrollToSection }
}
