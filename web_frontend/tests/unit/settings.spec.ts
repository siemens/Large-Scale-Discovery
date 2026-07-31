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

import { parseSettingsResponse } from '@/types/settings'

describe('parseSettingsResponse', () => {
  it('accepts safe links and normalizes their browsing context', () => {
    // Arrange
    const settings = {
      frontend_links: {
        menu: [
          {
            icon: 'external',
            text: 'Documentation',
            href: 'https://docs.example/guide',
            target: 'documentation-window',
          },
          {
            icon: 'home',
            text: 'Home',
            href: '/home',
            target: '_blank',
          },
        ],
      },
    }

    // Act
    const parsed = parseSettingsResponse(settings)

    // Assert
    expect(parsed.frontend_links?.menu?.map((link) => link.target)).toEqual(['_self', '_blank'])
  })

  it.each([
    ['URL credentials', 'https://user:password@docs.example/guide'],
    ['protocol-relative URL', '//docs.example/guide'],
  ])('rejects a link containing %s', (_caseName, href) => {
    // Arrange
    const settings = {
      frontend_links: {
        menu: [{ icon: 'external', text: 'Documentation', href, target: '_blank' }],
      },
    }

    // Act
    const parseUnsafeSettings = () => parseSettingsResponse(settings)

    // Assert
    expect(parseUnsafeSettings).toThrow()
  })
})
