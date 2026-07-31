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

interface ZodGlobalConfig {
  jitless?: boolean
}

interface ZodGlobalThis {
  __zod_globalConfig?: ZodGlobalConfig
}

// Zod reads this shared global object while its modules are evaluated. Setting
// jitless before Zod is imported avoids the runtime eval-support probe under CSP.
const zodGlobal = globalThis as typeof globalThis & ZodGlobalThis
zodGlobal.__zod_globalConfig = {
  ...zodGlobal.__zod_globalConfig,
  jitless: true,
}
