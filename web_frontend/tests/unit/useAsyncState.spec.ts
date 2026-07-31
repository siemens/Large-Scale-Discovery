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

import { describe, expect, it } from 'vitest'

import { useAsyncState, useListFetcher } from '@/composables/useAsyncState'

describe('useAsyncState', () => {
  it('stays loading until every concurrent tracked run settles', async () => {
    // Arrange — start two independently controlled operations through one shared state
    const state = useAsyncState()
    let resolveFirst: (() => void) | undefined
    let resolveSecond: (() => void) | undefined
    const first = state.run(
      () =>
        new Promise<void>((resolve) => {
          resolveFirst = resolve
        }),
    )
    const second = state.run(
      () =>
        new Promise<void>((resolve) => {
          resolveSecond = resolve
        }),
    )

    // Act — finish only the first operation
    resolveFirst?.()
    await first

    // Assert — the second operation continues to own the shared loading state
    expect(state.isLoading.value).toBe(true)
    resolveSecond?.()
    await second
    expect(state.isLoading.value).toBe(false)
  })

  it('ignores a stale completion after reset starts a new loading generation', async () => {
    // Arrange — retain control of an old operation across a reset
    const state = useAsyncState()
    let resolveOld: (() => void) | undefined
    let resolveNew: (() => void) | undefined
    const oldRun = state.run(
      () =>
        new Promise<void>((resolve) => {
          resolveOld = resolve
        }),
    )
    state.reset()
    const newRun = state.run(
      () =>
        new Promise<void>((resolve) => {
          resolveNew = resolve
        }),
    )

    // Act — settle the invalidated operation while the new one remains active
    resolveOld?.()
    await oldRun

    // Assert — stale cleanup cannot hide the new generation's loading state
    expect(state.isLoading.value).toBe(true)
    resolveNew?.()
    await newRun
    expect(state.isLoading.value).toBe(false)
  })
})

describe('useListFetcher', () => {
  it('publishes a silent refresh without activating the loading indicator', async () => {
    // Arrange — create an initialized list backed by a successful request
    const listFetcher = useListFetcher(
      () => Promise.resolve(['entry']),
      'Could not load the list.',
      { initialData: () => [] },
    )

    // Act — refresh in the background and observe state before awaiting completion
    const pendingFetch = listFetcher.fetchSilent()
    expect(listFetcher.isFetching.value).toBe(false)
    await pendingFetch

    // Assert — successful data and error state match a foreground fetch
    expect(listFetcher.data.value).toEqual(['entry'])
    expect(listFetcher.fetchError.value).toBeNull()
  })
})
