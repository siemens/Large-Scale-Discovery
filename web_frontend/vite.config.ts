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

import vueI18n from '@intlify/unplugin-vue-i18n/vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig, type Plugin } from 'vite'

// Ensure all <link rel="stylesheet"> tags appear before any <script> tags in
// the built index.html. Vite injects the entry script before stylesheets by
// default, which causes the browser to force layout before stylesheets are
// applied and emit a "Layout was forced before the page was fully loaded" warning.
function cssBeforeScripts(): Plugin {
  return {
    name: 'css-before-scripts',
    transformIndexHtml(html) {
      const stylePattern = /<link[^>]+rel="stylesheet"[^>]*>\n?/g
      const styles = html.match(stylePattern) ?? []
      const withoutStyles = html.replace(stylePattern, '')
      return withoutStyles.replace(/(<script\b)/, styles.join('') + '$1')
    },
  }
}

export default defineConfig({
  plugins: [
    vue(),
    cssBeforeScripts(),

    // Pre-compile all translation messages at build time so the runtime
    // compiler (which requires 'unsafe-eval') is never loaded.
    vueI18n({ include: fileURLToPath(new URL('./src/i18n/en-US.ts', import.meta.url)) }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),

      // Runtime-only Vue build: templates are pre-compiled by @vitejs/plugin-vue,
      // so the runtime template compiler (which needs 'unsafe-eval') is not needed.
      vue: 'vue/dist/vue.runtime.esm-bundler.js',

      // Runtime-only vue-i18n build: messages are pre-compiled at build time by
      // @intlify/unplugin-vue-i18n, so the runtime message compiler is not needed.
      'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js',
    },
  },
  server: {
    proxy: {
      // Scope the dev proxy to the REST prefix only. All API calls live under
      // `/api/v1` (see src/api/http.ts baseURL), so proxying `/api` would also
      // swallow the SPA page route `/api` (B3 / THEA-499 API docs page) and make
      // it unreachable in dev. Proxy `/api/v1` so `/api` falls through to the SPA.
      '/api/v1': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false, // accept self-signed certs
      },
    },
  },
})
