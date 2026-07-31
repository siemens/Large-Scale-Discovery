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

import { defineConfig } from 'vite'

import baseConfig from '../vite.config'

// Smoke-test override: proxy to the DevMode backend on port 8090.
export default defineConfig({
  ...baseConfig,
  server: {
    ...baseConfig.server,
    port: 5210,
    strictPort: true,
    host: '127.0.0.1',
    proxy: {
      '/api/v1': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path,
      },
    },
  },
})
