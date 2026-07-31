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

import { effectScope, nextTick, ref } from 'vue'

import { useSectionScroll } from '@/composables/useSectionScroll'

// Browser scheduling is controlled so every scroll assertion is deterministic
const requestAnimationFrameMock = vi.fn((callback: FrameRequestCallback): number => {
  callback(0)
  return 1
})

describe('useSectionScroll', () => {
  beforeEach(() => {
    // Each test starts with isolated timers and browser scrolling spies
    vi.useFakeTimers()
    vi.stubGlobal('requestAnimationFrame', requestAnimationFrameMock)
    vi.stubGlobal('cancelAnimationFrame', vi.fn())
    vi.spyOn(window, 'scrollTo').mockImplementation(() => undefined)
    Object.defineProperty(window, 'scrollY', { configurable: true, value: 30 })
  })

  afterEach(() => {
    // Browser globals and timers must not leak into another test
    vi.useRealTimers()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    requestAnimationFrameMock.mockClear()
  })

  it('smooth-scrolls to the section with the sticky navigation offset', () => {
    // Arrange — expose a section below the current viewport position
    const section = document.createElement('section')
    vi.spyOn(section, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 200))
    const scope = effectScope()

    // Act — request an immediate section scroll
    scope.run(() => useSectionScroll(() => section).scrollToSection())
    vi.runAllTimers()

    // Assert — scroll was called with smooth behavior
    expect(window.scrollTo).toHaveBeenCalledWith(expect.objectContaining({ behavior: 'smooth' }))
    scope.stop()
  })

  it('waits for the reveal transition before scrolling to a newly opened section', async () => {
    // Arrange — watch an initially closed action section
    const activeSection = ref<string | null>(null)
    const openCount = ref(0)
    const section = document.createElement('section')
    const scope = effectScope()
    scope.run(() =>
      useSectionScroll(() => section, {
        activeSection,
        openCount,
      }),
    )

    // Act — open the action section, wait for timers to run
    activeSection.value = 'add'
    openCount.value += 1
    await nextTick()

    // Assert — scrolling starts only after the transition delay has elapsed
    expect(window.scrollTo).not.toHaveBeenCalled()
    vi.runAllTimers()
    expect(window.scrollTo).toHaveBeenCalledTimes(1)
    scope.stop()
  })

  it('replaces a pending opening scroll when the active target changes', async () => {
    // Arrange — open one action target while its reveal scroll is pending
    const activeSection = ref<string | null>(null)
    const target = ref({ id: 1 })
    const section = document.createElement('section')
    const scope = effectScope()
    scope.run(() =>
      useSectionScroll(() => section, {
        activeSection,
        target,
      }),
    )
    activeSection.value = 'edit'
    await nextTick()

    // Act — switch targets before the first delayed scroll executes
    target.value = { id: 2 }
    await nextTick()
    vi.advanceTimersByTime(50)

    // Assert — only the latest section state causes a scroll
    expect(window.scrollTo).toHaveBeenCalledTimes(1)
    vi.runAllTimers()
    expect(window.scrollTo).toHaveBeenCalledTimes(1)
    scope.stop()
  })

  it('cancels a pending scroll when the action section closes', async () => {
    // Arrange — schedule scrolling for a newly opened action section
    const activeSection = ref<string | null>(null)
    const section = document.createElement('section')
    const scope = effectScope()
    scope.run(() =>
      useSectionScroll(() => section, {
        activeSection,
      }),
    )
    activeSection.value = 'add'
    await nextTick()

    // Act — close the section before the reveal transition completes
    activeSection.value = null
    await nextTick()
    vi.runAllTimers()

    // Assert — a removed section never causes a delayed page jump
    expect(window.scrollTo).not.toHaveBeenCalled()
    scope.stop()
  })
})
