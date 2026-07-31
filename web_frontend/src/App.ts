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

import '@/App.css'
import LayoutFooter from '@/components/layout/LayoutFooter.vue'
import LayoutNavSide from '@/components/layout/LayoutNavSide.vue'
import LayoutNavTop from '@/components/layout/LayoutNavTop.vue'
import { useAuthStore } from '@/stores/auth'

/** appShell exposes the application-specific pieces consumed by the shared root shell. */
export const appShell = {
  loadingBarColor: 'var(--p-primary-500)',
  showFeedbackFab: false,
  hasSideNav(path: string): boolean {
    return path.startsWith('/configuration') || path.startsWith('/admin')
  },
} as const

export { LayoutFooter, LayoutNavSide, LayoutNavTop, useAuthStore }
