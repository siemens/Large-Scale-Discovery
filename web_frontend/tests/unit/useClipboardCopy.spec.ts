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

import { describe, expect, it, vi } from 'vitest'

import { useClipboardCopy } from '@/composables/useClipboardCopy'

interface Row {
  input: string
  enabled: boolean
}

const columns = [
  { field: 'input', title: 'Target' },
  { field: 'enabled', title: 'Enabled' },
]

describe('useClipboardCopy', () => {
  it('builds TSV with a header row', () => {
    const { toTsv } = useClipboardCopy()
    const tsv = toTsv<Row>(
      [
        { input: '10.0.0.1', enabled: true },
        { input: '10.0.0.2', enabled: false },
      ],
      columns,
    )
    expect(tsv).toBe('Target\tEnabled\n10.0.0.1\ttrue\n10.0.0.2\tfalse')
  })

  it('renders null/undefined cells as empty strings', () => {
    const { toTsv } = useClipboardCopy()
    const tsv = toTsv([{ input: null, enabled: undefined }] as unknown as Row[], columns)
    expect(tsv).toBe('Target\tEnabled\n\t')
  })

  it('strips embedded tabs/newlines so a cell cannot break the grid', () => {
    const { toTsv } = useClipboardCopy()
    const tsv = toTsv([{ input: 'a\tb\nc', enabled: true }], columns)
    expect(tsv).toBe('Target\tEnabled\na b c\ttrue')
  })

  it('falls back to the field name when a column has no title', () => {
    const { toTsv } = useClipboardCopy()
    const tsv = toTsv([{ input: 'x', enabled: true }], [{ field: 'input' }])
    expect(tsv).toBe('input\nx')
  })

  it('copyRows writes the TSV to the clipboard', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText } })
    const { copyRows } = useClipboardCopy()
    await copyRows([{ input: 'x', enabled: true }], columns)
    expect(writeText).toHaveBeenCalledWith('Target\tEnabled\nx\ttrue')
  })
})
