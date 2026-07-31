<script setup lang="ts">
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

/**
 * WidgetTargets — spreadsheet-style scope-targets editor built on the shared
 * <DataTable> wrapper (columns config + slots), adding the target-specific kit on
 * top: clipboard paste, copy-table, the trailing blank row, OT collapse, per-row
 * delete/reset, and the read-only status columns.
 */
import { FilterMatchMode } from '@primevue/core/api'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  shallowRef,
  toRef,
  watch,
  watchEffect,
} from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable from '@/components/datatable/DataTable.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import FieldHint from '@/components/elements/FieldHint.vue'
import { startLoading, stopLoading } from '@/api/http'
import { copyToClipboard } from '@/composables/useClipboardCopy'
import { registerFormValidator } from '@/composables/useFormValidation'
import { useToast } from '@/composables/useToast'
import { formatDateTime, isZeroDate } from '@/i18n/format'
import { stringifyUnknown } from '@/lib/collections'
import { validateHostOrNetwork } from '@/lib/validators'

import type { CellEditPayload, ColumnDef } from '@/components/datatable/dataTableTypes'
import type { ScopeTarget } from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    invalidInputsHint:
      'Invalid scan inputs. Each entry must be either a CIDR range, IP address or FQDN.',
    grid: {
      columns: {
        rowNum: '#',
        input: 'Scan Input',
        enabled: 'Enabled',
        priority: 'Priority',
        timezone: 'Timezone',
        lat: 'Lat',
        lng: 'Lng',
        postalAddress: 'Postal Address',
        inputNetwork: 'Input Network',
        inputCountry: 'Country',
        inputLocation: 'Location',
        scanStarted: 'Scan Started',
        scanFinished: 'Scan Finished',
      },
      otInputPlaceholder: 'Network auto-detected by the agent',
    },
    actions: {
      clearFilters: 'Clear all filters',
      copy: {
        tooltip:
          'Use the clipboard to copy this table between here and Excel. You can use CTRL+V to paste your Excel changes back here.',
        success: 'Copied to clipboard.',
        error: 'Could not copy to clipboard.',
      },
      deleteRow: {
        tooltip: 'Delete row',
      },
      resetTarget: {
        tooltip: 'Reset target',
        title: 'Reset Scan Target',
        message:
          'The scan target will be queued again. A second result set may appear for this scan cycle.',
      },
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
// ScopeTarget extended with a client-side identity key for stable DataTable keying.
// timezone is widened to string | number | null | undefined here so the InputText
// editor can bind a string while the user types; getItems() coerces it back to number.
interface TargetRow extends Omit<ScopeTarget, 'timezone'> {
  rowId: string
  timezone?: string | number | null
}

// PasteColumn maps one clipboard column to a target field
interface PasteColumn {
  field: keyof TargetRow
  title: string
}

const props = withDefaults(
  defineProps<{
    otMode?: boolean
    readonly?: boolean
    scopeId?: number
    isModalOpen?: () => boolean
  }>(),
  {
    otMode: false,
    readonly: false,
    scopeId: undefined,
    isModalOpen: () => false,
  },
)

const emit = defineEmits<{
  edited: []
  resetTarget: [input: string]
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()

// ── Row identity ──────────────────────────────────────────────
let rowIdCounter = 0
function generateRowId(): string {
  return String(++rowIdCounter)
}

function makeRow(): TargetRow {
  return { enabled: true, rowId: generateRowId() }
}

// ── Items ─────────────────────────────────────────────────────
const targetRows = shallowRef<TargetRow[]>([makeRow()])

// True while waiting for the parent's first setItems() call (edit mode only).
// scopeId being defined means the parent will fetch targets before calling setItems().
const isLoadingTargets = ref(props.scopeId != null)
const preparedRowCapacity = ref(1)

function refreshRows(): void {
  targetRows.value = targetRows.value.slice()
}

/** Called by parent to load targets from the API response. */
function setItems(targets: ScopeTarget[]): void {
  isLoadingTargets.value = false
  const rows = targets.map((target) => ({ ...target, rowId: generateRowId() }))
  maintainEmptyRows(rows)
  targetRows.value = rows
  runRevalidate()
  clearTableFilters()
  scrollTableToTop()
  schedulePreparedRowSync()
}

/** Called by parent on submit to get clean target list. */
function getItems(): ScopeTarget[] {
  return targetRows.value.map((row) => {
    const target: ScopeTarget = { ...row, timezone: sanitizeTimezone(row.timezone) }
    Reflect.deleteProperty(target, 'rowId')
    return target
  })
}

// Coerce a raw timezone cell value (string from InputText, or existing number) to a
// clamped float, matching the legacy sanitizeTarget() in utils-app.js. Float because
// some timezones are fractional (e.g. UTC+5:30 = 5.5). Go backend expects float32.
function sanitizeTimezone(value: string | number | null | undefined): number {
  const n = typeof value === 'string' ? parseFloat(value) : Number(value ?? NaN)
  if (isNaN(n)) return 0
  if (n > 12) return 12
  if (n < -12) return -12
  return n
}

// Row IDs of non-empty input cells that fail scan-target validation. Bulk
// operations revalidate all rows; individual cell edits update exactly one row.
const invalidRowIds = ref<Set<string>>(new Set())
const hasInvalidInputs = computed(() => invalidRowIds.value.size > 0)

function runRevalidate(): void {
  const bad = new Set<string>()
  for (const row of targetRows.value) {
    const v = row.input?.trim()
    if (v && validateHostOrNetwork(v) !== undefined) bad.add(row.rowId)
  }
  invalidRowIds.value = bad
}

function validateRowInput(row: TargetRow): void {
  const next = new Set(invalidRowIds.value)
  const value = row.input?.trim()
  if (value && validateHostOrNetwork(value) !== undefined) next.add(row.rowId)
  else next.delete(row.rowId)
  invalidRowIds.value = next
}

function validateInputs(): boolean {
  runRevalidate()
  return invalidRowIds.value.size === 0
}

registerFormValidator(validateInputs)

defineExpose({ setItems, getItems, clearTargetScanData })

// ── Empty-row maintenance ─────────────────────────────────────
// In OT mode: isEmptyRow always returns true so the maintenance logic never appends.
// It still seeds 1 row when the list is empty (desired for OT).
const otModeRef = toRef(props, 'otMode')

const isEmptyRow = (row: TargetRow): boolean => {
  if (otModeRef.value) return true
  return !row.input?.trim()
}

// Keep enough prepared blank rows to fill the current visible editor viewport
// when the list is short. Once real rows fill the viewport, keep only the usual
// trailing blank row so the user can keep typing without an explicit "add row".
function maintainEmptyRows(list: TargetRow[]): boolean {
  if (otModeRef.value) {
    if (list.length > 0) return false
    list.push(makeRow())
    return true
  }

  let changed = false

  let lastFilledIndex = -1
  for (let i = list.length - 1; i >= 0; i -= 1) {
    const row = list[i]
    if (row && !isEmptyRow(row)) {
      lastFilledIndex = i
      break
    }
  }

  const requiredLength = Math.max(preparedRowCapacity.value, lastFilledIndex + 2)

  while (list.length > requiredLength) {
    const last = list[list.length - 1]
    if (!last || !isEmptyRow(last)) break
    list.pop()
    changed = true
  }

  while (list.length < requiredLength) {
    list.push(makeRow())
    changed = true
  }

  return changed
}

watch(
  // Watch length AND the last row's emptiness so editing the final blank row
  // (without changing length) still triggers a new trailing blank.
  () => {
    const list = targetRows.value
    const last = list.length > 0 ? list[list.length - 1] : undefined
    return [list.length, last ? isEmptyRow(last) : true] as const
  },
  () => {
    if (maintainEmptyRows(targetRows.value)) refreshRows()
  },
  { immediate: true },
)

// In OT mode: also ensure we never have more than 1 row (paste guard)
watchEffect(() => {
  // length > 1 guarantees a first element; keep only it in OT mode.
  const first = targetRows.value[0]
  if (props.otMode && targetRows.value.length > 1 && first) {
    targetRows.value = [first]
    runRevalidate()
  }
})

// When OT mode is toggled ON: keep only first row with input cleared
watch(otModeRef, (isOt) => {
  if (isOt) {
    const first = targetRows.value[0] ?? makeRow()
    first.input = ''
    first.enabled = true
    targetRows.value = [first]
    runRevalidate()
    scrollTableToTop()
    return
  }

  if (maintainEmptyRows(targetRows.value)) refreshRows()
  schedulePreparedRowSync()
})

// ── Paste columns ────────────────────────────────────────────

const PASTE_COLUMNS: PasteColumn[] = [
  { field: 'input', title: t('grid.columns.input') },
  { field: 'enabled', title: t('grid.columns.enabled') },
  { field: 'priority', title: t('grid.columns.priority') },
  { field: 'timezone', title: t('grid.columns.timezone') },
  { field: 'lat', title: t('grid.columns.lat') },
  { field: 'lng', title: t('grid.columns.lng') },
  {
    field: 'postal_address',
    title: t('grid.columns.postalAddress'),
  },
  {
    field: 'input_network',
    title: t('grid.columns.inputNetwork'),
  },
  {
    field: 'input_country',
    title: t('grid.columns.inputCountry'),
  },
  {
    field: 'input_location',
    title: t('grid.columns.inputLocation'),
  },
  {
    field: 'input_routing_domain',
    title: t('common.labels.routingDomain'),
  },
  { field: 'input_zone', title: t('common.labels.zone') },
  {
    field: 'input_purpose',
    title: t('common.labels.purpose'),
  },
  {
    field: 'input_company',
    title: t('common.labels.company'),
  },
  {
    field: 'input_department',
    title: t('common.labels.department'),
  },
  {
    field: 'input_manager',
    title: t('common.labels.manager'),
  },
  {
    field: 'input_contact',
    title: t('common.labels.contact'),
  },
  {
    field: 'input_comment',
    title: t('common.labels.comment'),
  },
]

// ── Paste handler ────────────────────────────────────────────
// Fields the parser yields as strings but the grid stores as other types. Empty
// pasted cells are dropped so the row keeps its default (e.g. enabled = true).
const BOOL_FIELDS = new Set(['enabled', 'priority'])

// Fields that are numeric in the backend (float32) and must be coerced from the pasted string.
const FLOAT_FIELDS = new Set(['timezone'])

// Parse a pasted string into a boolean (Excel/TSV variants: true/yes/1/x/✓).
function parseBooleanCell(value: string): boolean {
  const v = value.trim().toLowerCase()
  return v === 'true' || v === 'yes' || v === '1' || v === 'x' || v === '✓'
}

// Coerce one parsed row's string cells to the grid's field types, dropping blank
// cells so column defaults survive (e.g. an unset enabled stays true).
function coercePastedRow(raw: Partial<TargetRow>): Partial<TargetRow> {
  const out: Record<string, unknown> = {}
  for (const [field, value] of Object.entries(raw)) {
    // The parser only ever yields strings; skip blanks so defaults apply
    if (typeof value !== 'string' || value.trim() === '') continue
    if (BOOL_FIELDS.has(field)) out[field] = parseBooleanCell(value)
    else if (FLOAT_FIELDS.has(field)) {
      out[field] = sanitizeTimezone(value)
    } else {
      out[field] = value
    }
  }
  return out
}

// Identify native text-entry controls whose ordinary paste behavior must win
function isTextEntryElement(element: EventTarget | null): boolean {
  if (element instanceof HTMLElement && element.isContentEditable) return true
  if (element instanceof HTMLTextAreaElement) return true
  if (!(element instanceof HTMLInputElement)) return false

  // Exclude input types that do not accept free-form text
  const type = element.type
  return type !== 'checkbox' && type !== 'radio' && type !== 'button' && type !== 'submit'
}

// Detect PrimeVue overlays that should retain the active keyboard interaction
function hasOpenOverlay(): boolean {
  return document.querySelector('.p-dialog-mask, .p-overlay-mask, .p-overlay') !== null
}

// Parse Excel-compatible tab-separated clipboard text into target field values
function parsePastedRows(clipboard: string): Array<Record<string, string>> {
  if (!clipboard) return []

  // Split all common line endings and inspect the first row as a possible header
  const lines = clipboard.split(/\r\n|\n|\r/)
  const header = lines[0]?.split('\t')
  if (!header) return []

  // Normalize labels once so translated titles and stable field names both work
  const normalize = (value: string | undefined): string => value?.trim() ?? ''
  const columnsByTitle = new Map(PASTE_COLUMNS.map((column) => [normalize(column.title), column]))
  const columnsByField = new Map(PASTE_COLUMNS.map((column) => [String(column.field), column]))

  // Prefer the user-visible header and fall back to backend field names
  let columnMap = header.map((cell) => columnsByTitle.get(normalize(cell)))
  let hasHeader = columnMap.every((column) => column !== undefined)
  if (!hasHeader) {
    columnMap = header.map((cell) => columnsByField.get(normalize(cell)))
    hasHeader = columnMap.every((column) => column !== undefined)
  }

  // Headerless Excel data maps positionally to the configured data columns
  if (!hasHeader) columnMap = PASTE_COLUMNS.slice()

  // Convert every non-empty clipboard line into one partial target record
  const rows: Array<Record<string, string>> = []
  for (let lineIndex = hasHeader ? 1 : 0; lineIndex < lines.length; lineIndex += 1) {
    const cells = (lines[lineIndex] ?? '').split('\t')
    if (cells.length === 1 && cells[0] === '') continue

    // Preserve only cells that have a matching target column
    const row: Record<string, string> = {}
    for (let cellIndex = 0; cellIndex < cells.length; cellIndex += 1) {
      const column = columnMap[cellIndex]
      if (column) row[column.field] = cells[cellIndex] ?? ''
    }
    rows.push(row)
  }
  return rows
}

// Replace the editor contents with normalized clipboard rows
function applyPastedRows(rows: Array<Record<string, string>>): void {
  if (props.readonly || props.otMode) return

  // Keep defaults for omitted cells and discard rows without a scan input
  const nextRows: TargetRow[] = []
  for (const row of rows) {
    const targetRow = { ...makeRow(), ...coercePastedRow(row) }
    if (targetRow.input?.trim()) nextRows.push(targetRow)
  }

  // Restore editor invariants and make every replacement row visible
  maintainEmptyRows(nextRows)
  targetRows.value = nextRows
  runRevalidate()
  clearTableFilters()
  scrollTableToTop()
  schedulePreparedRowSync()
  emit('edited')
}

// Own paste events on the widget without hijacking its text editors or filters
function onPaste(event: ClipboardEvent): void {
  if (isTextEntryElement(event.target)) return

  // Read plain text and leave empty clipboard events to the browser
  const clipboard = event.clipboardData?.getData('text/plain') ?? ''
  if (!clipboard) return

  // Prevent raw insertion because the widget consumes the structured rows
  event.preventDefault()
  applyPastedRows(parsePastedRows(clipboard))
}

// ── Ctrl+V focus ──────────────────────────────────────────────
// Focus the grid on Ctrl/Cmd+V so paste lands in the table, not outside it.
const tableFocusRef = ref<HTMLElement | null>(null)
const tableScrollResetKey = ref(0)

// Focus the widget when a global paste shortcut has no more specific owner
function onPasteShortcut(event: KeyboardEvent): void {
  if (!(event.ctrlKey || event.metaKey) || event.key.toLowerCase() !== 'v') return
  if (isTextEntryElement(document.activeElement)) return
  if (props.isModalOpen() || hasOpenOverlay()) return

  // Route the following browser paste event to the widget wrapper
  tableFocusRef.value?.focus()
}

// ── Grid ref and filter reset ─────────────────────────────────
// The <DataTable> wrapper owns the per-column filter state. We hold a ref to it so
// a paste can clear all filters (freshly-pasted rows must be visible, not hidden
// by a stale filter). The wrapper exposes clearAllFilters().
interface DataTableExposed {
  clearAllFilters: () => void
  getVisibleRowCapacity: () => number
  isScrolledToBottom: () => boolean
  nudgeScrollAboveBottom: () => void
  preserveScrollPosition: () => void
  resetScrollPosition: () => void
  scrollToBottomPosition: () => void
}

const dataTableRef = ref<DataTableExposed | null>(null)

function clearTableFilters(): void {
  dataTableRef.value?.clearAllFilters()
}

function scrollTableToTop(): void {
  tableScrollResetKey.value += 1
}

function scrollTableToBottom(): void {
  dataTableRef.value?.scrollToBottomPosition()
}

function onVisibleRowCapacityChange(capacity: number): void {
  const nextCapacity = Math.max(1, Math.floor(capacity))
  if (nextCapacity === preparedRowCapacity.value) return

  preparedRowCapacity.value = nextCapacity
  if (maintainEmptyRows(targetRows.value)) refreshRows()
}

function measureRenderedRowCapacity(): number {
  const root = tableFocusRef.value
  if (!root) return 0

  const rows = root.querySelectorAll<HTMLElement>('.p-datatable-tbody > tr')
  let row: HTMLElement | null = null
  for (const candidate of rows) {
    if (
      candidate.classList.contains('p-datatable-emptymessage') ||
      candidate.classList.contains('p-datatable-empty-message') ||
      candidate.querySelector('.data-table__empty')
    ) {
      continue
    }

    row = candidate
    break
  }

  const rowHeight = row?.getBoundingClientRect().height ?? 0
  if (rowHeight <= 0) return 0

  const scroller = root.querySelector<HTMLElement>('.p-virtualscroller')
  if (!scroller) return 0

  const scrollerRect = scroller.getBoundingClientRect()
  const contentBottom = scrollerRect.top + scroller.clientHeight
  if (contentBottom <= scrollerRect.top) return 0

  let viewportTop = scrollerRect.top
  const header = root.querySelector<HTMLElement>('.p-datatable-thead')
  const headerRect = header?.getBoundingClientRect()

  if (
    headerRect &&
    headerRect.height > 0 &&
    headerRect.bottom > scrollerRect.top &&
    headerRect.top < contentBottom
  ) {
    viewportTop = Math.min(Math.max(headerRect.bottom, scrollerRect.top), contentBottom)
  }

  const viewportHeight = Math.max(0, contentBottom - viewportTop)
  return viewportHeight > 0 ? Math.max(1, Math.floor(viewportHeight / rowHeight)) : 0
}

function syncPreparedRowsFromRenderedTable(): void {
  const renderedCapacity = measureRenderedRowCapacity()
  const capacity = renderedCapacity || dataTableRef.value?.getVisibleRowCapacity()
  if (capacity) onVisibleRowCapacityChange(capacity)
}

let preparedRowSyncFrame: number | null = null
let preparedRowSyncTimer: ReturnType<typeof setTimeout> | null = null
let preparedRowSyncToken = 0
let isUnmounted = false

/** Cancel a superseded layout pass so repeated row updates do not stack animation-frame chains. */
function cancelPreparedRowSync(): void {
  if (preparedRowSyncFrame !== null) {
    cancelAnimationFrame(preparedRowSyncFrame)
    preparedRowSyncFrame = null
  }
  if (preparedRowSyncTimer !== null) {
    clearTimeout(preparedRowSyncTimer)
    preparedRowSyncTimer = null
  }
}

function schedulePreparedRowSync(): void {
  cancelPreparedRowSync()
  const token = ++preparedRowSyncToken
  void nextTick(() => {
    if (isUnmounted || token !== preparedRowSyncToken) return
    syncPreparedRowsFromRenderedTable()

    if (typeof requestAnimationFrame !== 'function') {
      preparedRowSyncTimer = setTimeout(() => {
        preparedRowSyncTimer = null
        if (!isUnmounted && token === preparedRowSyncToken) syncPreparedRowsFromRenderedTable()
      }, 0)
      return
    }

    let remainingFrames = 3
    const syncOnFrame = () => {
      preparedRowSyncFrame = null
      if (isUnmounted || token !== preparedRowSyncToken) return
      syncPreparedRowsFromRenderedTable()
      remainingFrames -= 1
      if (remainingFrames > 0) preparedRowSyncFrame = requestAnimationFrame(syncOnFrame)
    }
    preparedRowSyncFrame = requestAnimationFrame(syncOnFrame)
  })
}

// ── Copy table to clipboard ───────────────────────────────────
// TSV including a header row, so it round-trips with Excel and the Ctrl+V paste
// handler. Ports the legacy corner "copy" button (copyToClipboard('all')).
const isCopying = ref(false)

async function copyTargetTable(): Promise<void> {
  if (isCopying.value) return
  isCopying.value = true
  startLoading()

  try {
    const tsv = buildTargetTableTsv()
    if (await copyToClipboard(tsv)) {
      toast.brief(t('actions.copy.success'))
    } else {
      toast.warn(t('actions.copy.error'))
    }
  } finally {
    stopLoading()
    isCopying.value = false
  }
}

function cleanClipboardCell(value: unknown): string {
  return stringifyUnknown(value).replace(/[\t\r\n]+/g, ' ')
}

function targetCellValue(row: TargetRow, field: string): unknown {
  return Reflect.get(row, field)
}

function onBooleanCellChange(row: TargetRow, field: 'enabled' | 'priority', event: Event): void {
  // Native checkbox events are narrowed before updating the row
  if (event.currentTarget instanceof HTMLInputElement) {
    updateBooleanCell(row, field, event.currentTarget.checked)
  }
}

function inputCellText(row: TargetRow): string {
  if (!otModeRef.value) return row.input || ''

  const value = row.input?.trim()
  return value || t('grid.otInputPlaceholder')
}

function buildTargetTableTsv(): string {
  const lines: string[] = [
    PASTE_COLUMNS.map((column) => cleanClipboardCell(column.title)).join('\t'),
  ]

  for (const row of targetRows.value) {
    if (!row.input?.trim()) continue
    let line = ''
    for (let i = 0; i < PASTE_COLUMNS.length; i += 1) {
      if (i > 0) line += '\t'
      const column = PASTE_COLUMNS[i]
      if (column) line += cleanClipboardCell(targetCellValue(row, column.field))
    }
    lines.push(line)
  }

  return lines.join('\n')
}

// ── Delete row ───────────────────────────────────────────────

// Delete the backing row object, never the table's visible index. Sorting,
// filtering, and virtualization can all make visible indexes diverge.
function deleteRow(targetRow: TargetRow): void {
  if (props.readonly) return

  const rowIndex = targetRows.value.findIndex((row) => row === targetRow)
  if (rowIndex < 0) return

  const keepVirtualHeightStable = dataTableRef.value?.isScrolledToBottom() === true
  const previousLength = targetRows.value.length
  if (keepVirtualHeightStable) dataTableRef.value?.nudgeScrollAboveBottom()
  else dataTableRef.value?.preserveScrollPosition()

  const rowId = targetRow.rowId
  const nextRows = targetRows.value.slice()
  nextRows.splice(rowIndex, 1)
  maintainEmptyRows(nextRows)
  if (keepVirtualHeightStable) {
    while (nextRows.length < previousLength) nextRows.push(makeRow())
  }
  targetRows.value = nextRows

  if (invalidRowIds.value.has(rowId)) {
    const nextInvalid = new Set(invalidRowIds.value)
    nextInvalid.delete(rowId)
    invalidRowIds.value = nextInvalid
  }

  emit('edited')
}

function onCellEdit(edit: CellEditPayload<TargetRow>): void {
  let appendedEmptyRow = false
  if (edit.field === 'input') {
    validateRowInput(edit.row)
    const beforeLength = targetRows.value.length
    maintainEmptyRows(targetRows.value)
    appendedEmptyRow = targetRows.value.length > beforeLength
  }
  refreshRows()
  if (appendedEmptyRow) scrollTableToBottom()
  emit('edited')
}

function updateBooleanCell(row: TargetRow, field: 'enabled' | 'priority', checked: boolean): void {
  row[field] = checked
  refreshRows()
  emit('edited')
}

// ── Reset target ─────────────────────────────────────────────

onMounted(() => {
  schedulePreparedRowSync()
  document.addEventListener('keydown', onPasteShortcut)
})
onBeforeUnmount(() => {
  document.removeEventListener('keydown', onPasteShortcut)
  isUnmounted = true
  preparedRowSyncToken += 1
  cancelPreparedRowSync()
})

const resetConfirmVisible = ref(false)
const resetTargetInput = ref('')

// Hide the reset column entirely when viewing outside of scope-edit context (no scopeId).
const isResetActionVisible = computed(() => props.scopeId != null)

function openResetConfirm(input: string): void {
  resetTargetInput.value = input
  resetConfirmVisible.value = true
}

// Legacy parity: the reset button only appears once a target has actually been
// scanned — i.e. it carries a real `scan_started` timestamp (not empty or Go zero time).
// A freshly-added row that has never run cannot be reset.
function canResetTarget(row: TargetRow): boolean {
  return !isZeroDate(row.scan_started)
}

/** Called by the parent after a successful reset API call to clear the row's scan state. */
function clearTargetScanData(input: string): void {
  const row = targetRows.value.find((r) => r.input === input)
  if (!row) return
  row.scan_started = null
  row.scan_finished = null
  refreshRows()
}

// ── Editable metadata text columns ────────────────────────────
// The freeform text columns all render the same way (sortable header + filter box
// + click-to-edit InputText), so they're driven by one config list.
interface TextColumn {
  field: string
  header: string
  width?: string
  minWidth?: string
  maxWidth?: string
}

// Column definitions for the targets grid.
// minWidth to prevent column with from hopping too much during filtering.
const EDITABLE_TEXT_COLUMNS: TextColumn[] = [
  {
    field: 'timezone',
    header: t('grid.columns.timezone'),
    width: '7.5rem',
    minWidth: '7.5rem',
    maxWidth: '7.5rem',
  },
  {
    field: 'lat',
    header: t('grid.columns.lat'),
    width: '7.5rem',
    minWidth: '7.5rem',
    maxWidth: '10rem',
  },
  {
    field: 'lng',
    header: t('grid.columns.lng'),
    width: '7.5rem',
    minWidth: '7.5rem',
    maxWidth: '10rem',
  },
  {
    field: 'postal_address',
    header: t('grid.columns.postalAddress'),
    width: '10rem',
    minWidth: '10rem',
    maxWidth: '25rem',
  },
  {
    field: 'input_network',
    header: t('grid.columns.inputNetwork'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '25rem',
  },
  {
    field: 'input_country',
    header: t('grid.columns.inputCountry'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '15rem',
  },
  {
    field: 'input_location',
    header: t('grid.columns.inputLocation'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '15rem',
  },
  {
    field: 'input_routing_domain',
    header: t('common.labels.routingDomain'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_zone',
    header: t('common.labels.zone'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_purpose',
    header: t('common.labels.purpose'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_company',
    header: t('common.labels.company'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_department',
    header: t('common.labels.department'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_manager',
    header: t('common.labels.manager'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_contact',
    header: t('common.labels.contact'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '10rem',
  },
  {
    field: 'input_comment',
    header: t('common.labels.comment'),
    width: '8rem',
    minWidth: '8rem',
    maxWidth: '25rem',
  },
]

// ── DataTable column config ───────────────────────────────────
// Column definitions for the shared <DataTable> wrapper. Cells that need bespoke
// rendering (input span + editor, enabled/priority checkboxes, per-column filter
// boxes with their own test IDs) are provided by the matching #body:/#editor:/
// #filter: slots in the template; this config drives sort, filter match modes,
// widths, editability and the read-only status-column styling.
const targetColumns = computed<ColumnDef<TargetRow>[]>(() => [
  // Leading actions column (copy button in header, clear-filters in filter cell,
  // delete/reset per row — all supplied via slots).
  {
    field: 'actions',
    header: '',
    formatter: 'actions',
    sortable: false,
    width: '5rem',
    resizable: false,
  },
  {
    field: 'input',
    header: t('grid.columns.input'),
    sortable: true,
    filterable: true,
    filterType: 'text',
    editor: otModeRef.value ? undefined : 'input',
    width: '13rem',
    maxWidth: '25rem',
  },

  // Enabled / priority booleans: tri-state header filter, checkbox body (slots).
  {
    field: 'enabled',
    header: t('grid.columns.enabled'),
    sortable: true,
    filterable: true,
    filterType: 'tristate',
    matchMode: FilterMatchMode.EQUALS,
    width: '7rem',
  },
  {
    field: 'priority',
    header: t('grid.columns.priority'),
    sortable: true,
    filterable: true,
    filterType: 'tristate',
    matchMode: FilterMatchMode.EQUALS,
    width: '7rem',
  },
  ...EDITABLE_TEXT_COLUMNS.map((col): ColumnDef<TargetRow> => ({
    field: col.field,
    header: col.header,
    sortable: true,
    filterable: true,
    filterType: 'text',
    editor: 'input',
    width: col.width ?? '',
    minWidth: col.minWidth ?? '',
    maxWidth: col.maxWidth ?? '',
  })),

  // Read-only status columns: sortable + filterable but not editable, styled as a
  // dashed-off, muted block via columnClass.
  {
    field: 'scan_started',
    header: t('grid.columns.scanStarted'),
    sortable: true,
    filterable: true,
    filterType: 'dateText',
    columnClass: 'widget-targets__col-status widget-targets__col-status--first',
    width: '11.4rem',
    resizable: false,
  },
  {
    field: 'scan_finished',
    header: t('grid.columns.scanFinished'),
    sortable: true,
    filterable: true,
    filterType: 'dateText',
    columnClass: 'widget-targets__col-status',
    width: '11.4rem',
    resizable: false,
  },
])
</script>

<template>
  <div
    ref="tableFocusRef"
    class="widget-targets"
    data-testid="scope-targets-grid"
    tabindex="-1"
    @paste="onPaste"
  >
    <!-- Shared <DataTable> wrapper drives the table shell + skin; the exotic
         target-specific bits (copy button, checkboxes, status em-dash, paste) are
         supplied via slots. edit-mode drops to 'none' while synchronizing so the
         cells become read-only. -->
    <DataTable
      ref="dataTableRef"
      class="widget-targets__table"
      scroll-height="25rem"
      :rows="targetRows"
      :columns="targetColumns"
      data-key="rowId"
      :edit-mode="readonly ? 'none' : 'cell'"
      commit-on-blur
      :loading="isLoadingTargets"
      virtual-scroll
      :virtual-scroll-tolerated-items="8"
      row-number
      :row-number-header="t('grid.columns.rowNum')"
      actions-position="start"
      column-resize-mode="expand"
      :scroll-reset-key="tableScrollResetKey"
      @cell-edit="onCellEdit"
    >
      <template #actions-header>
        <Button
          v-tooltip.top="t('actions.copy.tooltip')"
          type="button"
          icon="pi pi-clipboard"
          text
          severity="contrast"
          class="data-table__button"
          :disabled="isCopying"
          :aria-label="t('actions.copy.tooltip')"
          @click="copyTargetTable"
        />
      </template>

      <template #actions-filter>
        <Button
          v-tooltip.top="t('actions.clearFilters')"
          type="button"
          icon="pi pi-filter-slash"
          text
          severity="contrast"
          class="data-table__button"
          :aria-label="t('actions.clearFilters')"
          @click="clearTableFilters"
        />
      </template>

      <!-- Per-row delete (+ reset in update mode once a target has scanned). -->
      <template #actions="{ data: targetRow }">
        <div class="data-table__actions">
          <Button
            v-tooltip.top="t('actions.deleteRow.tooltip')"
            type="button"
            icon="pi pi-trash"
            text
            severity="contrast"
            class="data-table__button"
            data-testid="targets-delete-row"
            :disabled="readonly"
            :aria-label="t('actions.deleteRow.tooltip')"
            @click="deleteRow(targetRow)"
          />
          <Button
            v-if="isResetActionVisible && canResetTarget(targetRow)"
            v-tooltip.top="t('actions.resetTarget.tooltip')"
            type="button"
            icon="pi pi-history"
            text
            severity="contrast"
            class="data-table__button"
            :disabled="readonly || !targetRow.input"
            :aria-label="t('actions.resetTarget.tooltip')"
            @click="openResetConfirm(targetRow.input ?? '')"
          />
        </div>
      </template>

      <!-- Scan Input: ellipsised body + editor (autocomplete off).
           Invalid values (non-empty but failing validateHostOrNetwork) are tinted red. -->
      <template #body:input="{ data: targetRow }">
        <span
          class="widget-targets__cell-text"
          :class="{
            'widget-targets__cell-text--invalid': !otModeRef && invalidRowIds.has(targetRow.rowId),
            'widget-targets__cell-text--ot': otModeRef,
          }"
          >{{ inputCellText(targetRow) }}</span
        >
      </template>
      <template #editor:input="{ data: targetRow }">
        <InputText v-model="targetRow.input" class="widget-targets__editor" autocomplete="off" />
      </template>

      <!-- Enabled / priority: native checkboxes to avoid mounting 2×N PrimeVue components. -->
      <template #body:enabled="{ data: targetRow }">
        <div class="widget-targets__cell-center">
          <input
            type="checkbox"
            class="widget-targets__checkbox"
            :checked="targetRow.enabled"
            :disabled="readonly"
            @change="onBooleanCellChange(targetRow, 'enabled', $event)"
          />
        </div>
      </template>
      <template #body:priority="{ data: targetRow }">
        <div class="widget-targets__cell-center">
          <input
            type="checkbox"
            class="widget-targets__checkbox"
            :checked="targetRow.priority"
            :disabled="readonly"
            @change="onBooleanCellChange(targetRow, 'priority', $event)"
          />
        </div>
      </template>

      <!-- Read-only status columns: format as datetime, hyphen when empty or Go zero time. -->
      <template #body:scan_started="{ data: targetRow }">{{
        isZeroDate(targetRow.scan_started) ? '-' : formatDateTime(targetRow.scan_started)
      }}</template>
      <template #body:scan_finished="{ data: targetRow }">{{
        isZeroDate(targetRow.scan_finished) ? '-' : formatDateTime(targetRow.scan_finished)
      }}</template>
    </DataTable>

    <FieldHint
      v-if="hasInvalidInputs"
      :error="t('invalidInputsHint')"
      class="widget-targets__validation-hint"
    />

    <DialogConfirm
      v-model:visible="resetConfirmVisible"
      :title="t('actions.resetTarget.title')"
      icon="pi pi-refresh"
      severity="primary"
      :message="t('actions.resetTarget.message')"
      :secondary-message="resetTargetInput"
      :text-confirm-label="t('common.actions.confirm')"
      @confirm="emit('resetTarget', resetTargetInput)"
    />
  </div>
</template>

<style scoped>
/* Root wrapper: relative so the confirmation overlay can be positioned against it. */
.widget-targets {
  position: relative;
  outline: none;
}

.widget-targets__editor {
  width: 100%;
}

/* Scan Input body cell: ellipsize long values so the column stays narrow. */
.widget-targets__cell-text {
  display: block;
  width: 100%;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.widget-targets__cell-text--invalid {
  color: var(--p-red-500);
}

.widget-targets__cell-text--ot {
  color: var(--p-text-muted-color, var(--p-surface-500));
}

.widget-targets__validation-hint {
  margin-top: 0.5rem;
}

/* The table shell + skin (outer frame, filter header, funnel-hide, header
   ellipsis, icon-button colors, active-filter border, per-column filter boxes)
   all come from the shared <DataTable> wrapper. Only target-specific rules for
   the slot content remain here. */

.widget-targets__cell-center {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

/* Status columns (last two): separated from editable columns with a dashed border and
   visually muted to signal read-only. */
.widget-targets__table :deep(.widget-targets__col-status--first) {
  border-left: 2px dashed var(--p-content-border-color);
}

.widget-targets__table :deep(.widget-targets__col-status) {
  background: var(--p-surface-50);
  color: var(--p-surface-800);
}

/* ── Styled native checkbox (replicates PrimeVue Checkbox appearance) ──────── */
/*
 * Uses the underlying field / primary / focus-ring CSS variables instead of
 * --p-checkbox-* tokens, which are only injected when a PrimeVue Checkbox is
 * rendered. These base variables are always present (InputText etc. guarantee it).
 */

.widget-targets__checkbox {
  appearance: none;
  -webkit-appearance: none;
  position: relative;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: var(--p-border-radius-sm);
  border: 1px solid var(--p-form-field-border-color);
  background: var(--p-form-field-background);
  box-shadow: var(--p-form-field-shadow);
  cursor: pointer;
  transition:
    background var(--p-form-field-transition-duration),
    border-color var(--p-form-field-transition-duration),
    box-shadow var(--p-form-field-transition-duration);
  vertical-align: middle;
}

.widget-targets__checkbox:hover {
  border-color: var(--p-form-field-hover-border-color);
}

.widget-targets__checkbox:checked {
  border-color: var(--p-primary-color);
  background: var(--p-primary-color);
  /* Checkmark: exact SVG path from @primevue/icons/check, white fill */
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 14 14' fill='none'%3E%3Cpath d='M4.86199 11.5948C4.78717 11.5923 4.71366 11.5745 4.64596 11.5426C4.57826 11.5107 4.51779 11.4652 4.46827 11.4091L0.753985 7.69483C0.683167 7.64891 0.623706 7.58751 0.580092 7.51525C0.536478 7.44299 0.509851 7.36177 0.502221 7.27771C0.49459 7.19366 0.506156 7.10897 0.536046 7.03004C0.565935 6.95111 0.613367 6.88 0.674759 6.82208C0.736151 6.76416 0.8099 6.72095 0.890436 6.69571C0.970973 6.67046 1.05619 6.66385 1.13966 6.67635C1.22313 6.68886 1.30266 6.72017 1.37226 6.76792C1.44186 6.81567 1.4997 6.8786 1.54141 6.95197L4.86199 10.2503L12.6397 2.49483C12.7444 2.42694 12.8689 2.39617 12.9932 2.40745C13.1174 2.41873 13.2343 2.47141 13.3251 2.55705C13.4159 2.64268 13.4753 2.75632 13.4938 2.87973C13.5123 3.00315 13.4888 3.1292 13.4271 3.23768L5.2557 11.4091C5.20618 11.4652 5.14571 11.5107 5.07801 11.5426C5.01031 11.5745 4.9368 11.5923 4.86199 11.5948Z' fill='white'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: center;
  background-size: 0.875rem;
}

.widget-targets__checkbox:focus-visible {
  outline: var(--p-focus-ring-width) var(--p-focus-ring-style) var(--p-focus-ring-color);
  outline-offset: var(--p-focus-ring-offset);
  box-shadow: var(--p-focus-ring-shadow);
  border-color: var(--p-form-field-border-color);
}

.widget-targets__checkbox:disabled {
  cursor: default;
  background: var(--p-form-field-disabled-background);
  border-color: var(--p-form-field-border-color);
}
</style>
