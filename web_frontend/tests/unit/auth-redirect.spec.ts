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

import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  clearPendingRedirect,
  consumePendingRedirect,
  createAuthRedirectFlow,
  savePendingRedirect,
} from '@/lib/redirect'

const sessionMock = vi.hoisted(() => ({
  getAccessToken: vi.fn<() => string | null>(),
  hasValidSession: vi.fn<() => boolean>(),
}))

const routerMock = vi.hoisted(() => ({
  currentRoute: { value: { fullPath: '/home' } },
  isReady: vi.fn<() => Promise<void>>(),
  replace: vi.fn<(target: string) => Promise<void>>(),
}))

const toastEventBusMock = vi.hoisted(() => ({ emit: vi.fn() }))

vi.mock('@/api/session', () => sessionMock)
vi.mock('primevue/toasteventbus', () => ({ default: toastEventBusMock }))

describe('authentication redirect flow', () => {
  beforeEach(() => {
    sessionStorage.clear()
    window.history.replaceState({}, '', '/')
    sessionMock.getAccessToken.mockReturnValue(null)
    sessionMock.hasValidSession.mockReturnValue(true)
    routerMock.currentRoute.value.fullPath = '/home'
    routerMock.isReady.mockResolvedValue()
    routerMock.replace.mockResolvedValue()
    vi.clearAllMocks()
  })

  it('restores a same-origin deep link after authentication', async () => {
    savePendingRedirect('/admin/users?filter=active#details')

    const flow = createAuthRedirectFlow(routerMock)
    await flow.restorePendingRedirect()

    expect(routerMock.isReady).toHaveBeenCalledOnce()
    expect(routerMock.replace).toHaveBeenCalledWith('/admin/users?filter=active#details')
    expect(sessionStorage.getItem('redirect')).toBeNull()
  })

  it.each(['//outside.domain.tld/path', '/login', '/register'])(
    'rejects unsafe or authentication-page redirect %s',
    async (target) => {
      savePendingRedirect(target)

      const flow = createAuthRedirectFlow(routerMock)
      await flow.restorePendingRedirect()

      expect(routerMock.replace).not.toHaveBeenCalled()
      expect(sessionStorage.getItem('redirect')).toBeNull()
    },
  )

  it('shares redirect consumption and cleanup with credential authentication', () => {
    savePendingRedirect('/api?section=tokens')
    expect(consumePendingRedirect()).toBe('/api?section=tokens')
    expect(consumePendingRedirect()).toBeNull()

    savePendingRedirect('/home')
    clearPendingRedirect()
    expect(consumePendingRedirect()).toBeNull()
  })

  it('removes and reports an OAuth callback failure exactly once', () => {
    sessionMock.hasValidSession.mockReturnValue(false)
    window.history.replaceState({}, '', '/?error=Unauthorized')

    const flow = createAuthRedirectFlow(routerMock)
    flow.showFailure()

    expect(window.location.search).toBe('')
    expect(toastEventBusMock.emit).toHaveBeenCalledOnce()
    expect(toastEventBusMock.emit).toHaveBeenCalledWith(
      'add',
      expect.objectContaining({ severity: 'error' }),
    )
  })
})
