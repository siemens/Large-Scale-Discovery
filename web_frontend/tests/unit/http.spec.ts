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
import { afterEach, describe, expect, it, vi } from 'vitest'

import http, { resolveApiBaseUrl } from '@/api/http'

import type { AxiosAdapter } from 'axios'

// Restore the production adapter after tests that exercise the real response interceptor
const originalAdapter = http.defaults.adapter

afterEach(() => {
  http.defaults.adapter = originalAdapter
  ToastEventBus.clear()
})

describe('resolveApiBaseUrl', () => {
  it('supports root-relative API paths', () => {
    // Arrange
    const configuredValue = '/services/api/'

    // Act
    const baseUrl = resolveApiBaseUrl(configuredValue, 'https://application.example')

    // Assert
    expect(baseUrl).toBe('/services/api')
  })

  it('supports an absolute same-origin API URL in browsers', () => {
    // Arrange
    const configuredValue = 'https://application.example/services/api/'

    // Act
    const baseUrl = resolveApiBaseUrl(configuredValue, 'https://application.example')

    // Assert
    expect(baseUrl).toBe('https://application.example/services/api')
  })

  it('rejects an absolute cross-origin API URL in browsers', () => {
    // Arrange
    const configuredValue = 'https://api.example/services/api'

    // Act
    const resolveConfiguredUrl = () =>
      resolveApiBaseUrl(configuredValue, 'https://application.example')

    // Assert
    expect(resolveConfiguredUrl).toThrow('same-origin')
  })

  it('rejects protocol-relative API URLs', () => {
    // Arrange
    const configuredValue = '//api.example/services/api'

    // Act
    const resolveConfiguredUrl = () =>
      resolveApiBaseUrl(configuredValue, 'https://application.example')

    // Assert
    expect(resolveConfiguredUrl).toThrow('root-relative')
  })
})

describe('API error feedback', () => {
  it('emits an error toast for a failed backend envelope', async () => {
    // Arrange: run the real response interceptor against an application-level failure
    const toastListener = vi.fn()
    const errorToastMessage = 'Could not complete the request.'
    ToastEventBus.on('add', toastListener)
    http.defaults.adapter = applicationErrorAdapter

    // Act
    await expect(http.get('/automatic-error-toast', { errorToastMessage })).rejects.toBeDefined()

    // Assert: the client owns one notification and honors its contextual message
    expect(toastListener).toHaveBeenCalledWith(
      expect.objectContaining({ summary: errorToastMessage }),
    )
  })

  it('allows a specialized caller to suppress the automatic error toast', async () => {
    // Arrange: observe the same client behavior through an explicitly suppressed request
    const toastListener = vi.fn()
    ToastEventBus.on('add', toastListener)
    http.defaults.adapter = applicationErrorAdapter

    // Act
    await expect(
      http.get('/suppressed-error-toast', { suppressErrorToast: true }),
    ).rejects.toBeDefined()

    // Assert
    expect(toastListener).not.toHaveBeenCalled()
  })
})

/** applicationErrorAdapter returns the legacy HTTP-200 error envelope. */
const applicationErrorAdapter: AxiosAdapter = (config) =>
  Promise.resolve({
    data: { error: true, message: 'Request rejected.' },
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
  })
