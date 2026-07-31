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

// Ambient types for Vite-specific features: typed `import.meta.env`, the `*.vue`
// SFC module shim, and suffixed asset imports (`?url`, `?raw`, `?worker`).
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
}

// Add the optional deployment API address to Vite environment typing
interface ImportMeta {
  readonly env: ImportMetaEnv
}
