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

import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'
import { vueTsConfigs, withVueTs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'

export default withVueTs(
  { rootDir: import.meta.dirname },
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      'tests/artifacts/**',
      '.cache/**',
      '.repo-edits/**',
      '.agents/**',
    ],
  },
  pluginVue.configs['flat/recommended'],
  vueTsConfigs.recommendedTypeChecked,
  {
    files: ['src/**/*.{ts,vue}', 'tests/**/*.ts', '*.config.ts'],
    rules: {
      '@typescript-eslint/consistent-type-imports': ['error', { prefer: 'type-imports' }],
      '@typescript-eslint/no-confusing-void-expression': ['error', { ignoreArrowShorthand: true }],
      '@typescript-eslint/no-deprecated': 'off',
      '@typescript-eslint/no-floating-promises': ['error', { ignoreVoid: true }],
      '@typescript-eslint/no-misused-promises': ['error', { checksVoidReturn: false }],
      '@typescript-eslint/no-unnecessary-condition': 'off',
      '@typescript-eslint/prefer-nullish-coalescing': 'off',
      '@typescript-eslint/return-await': ['error', 'in-try-catch'],
      'sort-imports': [
        'error',
        {
          allowSeparatedGroups: true,
          ignoreDeclarationSort: true,
          ignoreMemberSort: false,
        },
      ],
      'vue/block-lang': ['error', { script: { lang: 'ts' } }],
      'vue/component-api-style': ['error', ['script-setup']],
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      'vue/custom-event-name-casing': ['error', 'camelCase'],
      'vue/define-emits-declaration': ['error', 'type-literal'],
      'vue/define-props-declaration': ['error', 'type-based'],
      'vue/html-button-has-type': 'error',
      'vue/no-v-html': 'error',
      'vue/no-deprecated-filter': 'off',
      'vue/prefer-true-attribute-shorthand': 'error',
      'vue/require-default-prop': 'off',
      'vue/require-explicit-emits': 'error',
    },
  },
  {
    files: ['src/pages/**/*.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
  {
    files: [
      'src/api/**/*.{ts,vue}',
      'src/config/**/*.{ts,vue}',
      'src/lib/**/*.{ts,vue}',
      'src/types/**/*.{ts,vue}',
    ],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/components/**', '@/pages/**', '@/stores/**'],
              message:
                'Infrastructure and domain modules must not depend on UI or application state.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/stores/**/*.{ts,vue}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/components/**', '@/pages/**'],
              message: 'Stores must not depend on presentation modules.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/components/elements/**/*.{ts,vue}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/pages/**', '@/components/forms/**', '@/components/widgets/**'],
              message:
                'Shared UI primitives must not depend on feature-level presentation modules.',
            },
          ],
        },
      ],
    },
  },
  skipFormatting,
)
