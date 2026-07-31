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

import 'vue-router'

declare module 'vue-router' {
  /** Authorization and layout flags consumed by the global route guard and App.vue. */
  interface RouteMeta {
    /** Route is available without an authenticated session. */
    public?: boolean
    /** Route uses the full-bleed authentication layout. */
    auth?: boolean
    /** Route requires an administrator profile. */
    admin?: boolean
    /** Route requires an owner or administrator profile. */
    owner?: boolean
  }
}

export {}
