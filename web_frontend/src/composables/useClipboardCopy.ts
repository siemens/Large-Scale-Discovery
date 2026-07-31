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

import { stringifyUnknown } from '@/lib/collections'

/** Column used by tab-separated clipboard exports. */
interface ClipboardColumn {
  field: string
  title?: string
}

/** Clipboard-ready table serialization and copying. */
interface ClipboardCopyActions {
  toTsv: <T extends object>(rows: readonly T[], columns: readonly ClipboardColumn[]) => string
  copyRows: <T extends object>(
    rows: readonly T[],
    columns: readonly ClipboardColumn[],
  ) => Promise<boolean>
}

/** Builds and copies tab-separated rows for grid-style exports. */
export function useClipboardCopy(): ClipboardCopyActions {
  // Normalize one cell before writing TSV text
  function clean(value: unknown): string {
    if (value === null || value === undefined) return ''
    return stringifyUnknown(value).replace(/[\t\r\n]+/g, ' ')
  }

  // Convert selected rows into tab-separated text
  function toTsv<T extends object>(
    rows: readonly T[],
    columns: readonly ClipboardColumn[],
  ): string {
    // Collect output lines before joining them
    const lines: string[] = [
      columns.map((column) => clean(column.title ?? column.field)).join('\t'),
    ]

    for (const row of rows) {
      // Build one tab-separated output row
      let line = ''
      for (let i = 0; i < columns.length; i += 1) {
        if (i > 0) line += '\t'

        // Match this cell to its configured table column
        const column = columns[i]
        if (column) line += clean(Reflect.get(row, column.field))
      }
      lines.push(line)
    }

    return lines.join('\n')
  }

  // Copy selected rows and report whether the browser accepted them
  async function copyRows<T extends object>(
    rows: readonly T[],
    columns: readonly ClipboardColumn[],
  ): Promise<boolean> {
    return copyToClipboard(toTsv(rows, columns))
  }

  return { toTsv, copyRows }
}

/**
 * Copy `text` to the clipboard, returning `true` on success and `false` if both
 * the async Clipboard API and the legacy fallback fail.
 *
 * Uses `navigator.clipboard.writeText` when available (secure contexts), and
 * otherwise falls back to a hidden, off-screen `<textarea>` + `execCommand`
 * (insecure contexts / older browsers). Callers decide what to show on each
 * outcome (success / warning toast, or nothing). This helper never toasts.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // Fall through to the legacy path (e.g. permission denied)
    }
  }
  return legacyCopy(text)
}

/** Hidden-textarea + `execCommand('copy')` fallback for insecure contexts. */
function legacyCopy(text: string): boolean {
  // Create a temporary textarea for the legacy clipboard fallback
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')

  // Keep it out of the viewport so focusing/selecting it doesn't scroll the page
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'

  document.body.appendChild(textarea)
  textarea.select()
  try {
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    textarea.remove()
  }
}
