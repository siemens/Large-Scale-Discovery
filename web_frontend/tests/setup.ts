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

import { config, enableAutoUnmount } from '@vue/test-utils'
import { afterEach } from 'vitest'

import { i18n } from '@/i18n'

// Dispose every wrapper after its test, including wrappers from failed assertions
enableAutoUnmount(afterEach)

// Components use the same app-wide translator in production and unit tests
config.global.plugins = [i18n]

// Component tests do not need PrimeVue's visual tooltip behavior
config.global.directives = { tooltip: {} }

// JSDOM exposes scrollTo but reports it as unimplemented when components use it
window.scrollTo = () => undefined
