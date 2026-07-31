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

import { mount } from '@vue/test-utils'

import LinuxIcon from '@/components/elements/LinuxIcon.vue'

describe('LinuxIcon', () => {
  it('is decorative by default', () => {
    // Arrange and act
    const wrapper = mount(LinuxIcon)
    const icon = wrapper.get('svg')

    // Assert
    expect(icon.attributes('aria-hidden')).toBe('true')
    expect(icon.attributes('role')).toBeUndefined()
  })

  it('can be named when it conveys information on its own', () => {
    // Arrange and act
    const wrapper = mount(LinuxIcon, { props: { label: 'Linux' } })
    const icon = wrapper.get('svg')

    // Assert
    expect(icon.attributes('role')).toBe('img')
    expect(icon.attributes('aria-label')).toBe('Linux')
    expect(icon.attributes('aria-hidden')).toBeUndefined()
  })
})
