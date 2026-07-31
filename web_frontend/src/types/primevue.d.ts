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

declare module 'primevue/toasteventbus' {
  // Describe the global PrimeVue toast event bus used by the announcer
  const ToastEventBus: {
    emit: (event: string, payload?: unknown) => void
    on: (event: string, callback: (payload: unknown) => void) => void
    off: (event: string, callback: (payload: unknown) => void) => void
    clear: () => void
  }

  export default ToastEventBus
}
