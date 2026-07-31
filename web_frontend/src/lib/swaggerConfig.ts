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

import type { SwaggerConfig } from '@/lib/swagger'

/** Display order for endpoints within the API documentation page. */
const endpointOrder: string[] = [
  '/scopes',
  '/scope/targets',
  '/scope/update/custom',
  '/scope/update/networks',
  '/scope/update/assets',
  '/scope/update/settings',
]

/**
 * Fields to hide from the API documentation. Keys are swagger definition names
 * (without package prefix). Listed fields are completely removed from parameter
 * tables, request forms, and example responses.
 */
const fieldHidden: Record<string, string[]> = {
  T_discovery: ['scan_started', 'scan_finished'],
  Scope: ['scan_settings', 'scan_agents', 'connection', 'cycle_queue', 'attributes'],
}

/**
 * Field display order for API response documentation. Keys are swagger
 * definition names (without package prefix). Fields not listed are appended.
 */
const fieldOrder: Record<string, string[]> = {
  ScopeTargetsRequest: ['id'],
  ScopeTargetsResponse: ['synchronization', 'targets'],
  ScopeCreateUpdateCustomRequest: [
    'scope_id',
    'group_id',
    'name',
    'ot',
    'cycles',
    'cycles_retention',
    'targets',
  ],
  ScopeCreateUpdateCustomResponse: ['warnings'],
  T_discovery: [
    'input',
    'enabled',
    'priority',
    'input_network',
    'input_country',
    'input_location',
    'input_zone',
    'input_purpose',
    'input_company',
    'input_department',
    'input_manager',
    'input_contact',
    'input_comment',
  ],
}

// ── Types ─────────────────────────────────────────────────────────────────────

/** SWAGGER_CONFIG defines application-specific endpoint and schema presentation. */
export const SWAGGER_CONFIG: SwaggerConfig = {
  endpointOrder,
  fieldHidden,
  fieldOrder,
  defaultTitle: 'API Documentation',
}
