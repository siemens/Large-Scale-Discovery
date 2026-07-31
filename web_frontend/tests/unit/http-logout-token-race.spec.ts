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

import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import http from '@/api/http'
import { clearSession, getAccessToken, setToken } from '@/api/session'

import type { AxiosAdapter } from 'axios'

/**
 * Regression for THEA-631: a sliding-window token refresh on an in-flight
 * response must NOT restore the JWT after the user has logged out.
 *
 * We drive the real `http` instance (so the real response interceptor runs)
 * through a fake adapter that returns a refreshed `{ token }` envelope, then
 * assert the interceptor only persists it when the session is still active.
 */

const REFRESHED = { access_token: 'refreshed-jwt', expire: '2099-01-01T00:00:00Z' }

/** Adapter that mimics the backend re-issuing a token on a successful response. */
const refreshAdapter: AxiosAdapter = (config) =>
  Promise.resolve({
    data: { token: REFRESHED },
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
  })

const originalAdapter = http.defaults.adapter

describe('http response interceptor — sliding-window refresh after logout', () => {
  beforeEach(() => {
    http.defaults.adapter = refreshAdapter
    clearSession()
    sessionStorage.clear()
  })

  afterEach(() => {
    http.defaults.adapter = originalAdapter
    clearSession()
    sessionStorage.clear()
  })

  it('persists the refreshed token while the session is active', async () => {
    setToken({ access_token: 'active-jwt', expire: '2099-01-01T00:00:00Z' })

    await http.get('/whatever')

    expect(getAccessToken()).toBe(REFRESHED.access_token)
    expect(sessionStorage.getItem('token')).toContain(REFRESHED.access_token)
  })

  it('does NOT restore a token when the session was cleared (post-logout)', async () => {
    // No active session: simulates a request that resolves after clearSession().
    expect(getAccessToken()).toBeNull()

    await http.get('/whatever')

    expect(getAccessToken()).toBeNull()
    expect(sessionStorage.getItem('token')).toBeNull()
  })
})
