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

import { defineConfig, devices } from '@playwright/test'
import { join } from 'node:path'
import { fileURLToPath, URL } from 'node:url'

const projectRoot = fileURLToPath(new URL('..', import.meta.url))
const artifactRoot = join(projectRoot, 'tests', 'artifacts', 'e2e')
const serverUrl = 'http://127.0.0.1:4173'

export default defineConfig({
  testDir: './e2e',

  // Playwright clears this directory at the start of each run. Keeping all
  // generated screenshots, traces, and reports below one ignored root also
  // lets CI upload the complete run as a single artifact.
  outputDir: join(artifactRoot, 'results'),
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,

  // Firefox became unreliable when Playwright used every available logical
  // CPU in CI. Two workers retain parallelism without starving browser setup.
  workers: 2,
  reporter: [['html', { outputFolder: join(artifactRoot, 'report') }]],
  webServer: {
    // CI builds first and exercises the production bundle; local runs keep the
    // faster Vite development server. Both expose the same explicit endpoint.
    command: process.env.CI
      ? 'npm run preview -- --host 127.0.0.1 --port 4173 --strictPort'
      : 'npm run dev -- --host 127.0.0.1 --port 4173 --strictPort',

    // Playwright otherwise uses the moved config's directory as the process cwd.
    cwd: projectRoot,
    url: serverUrl,
    reuseExistingServer: !process.env.CI,

    // Surface Vite startup output in CI and allow slow shared runners enough
    // time to expose the already-built preview bundle.
    stdout: 'pipe',
    stderr: 'pipe',
    timeout: 120_000,
  },
  use: {
    baseURL: serverUrl,
    locale: 'en-US',
    timezoneId: 'UTC',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
  ],
})
