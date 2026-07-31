<script setup lang="ts" generic="T extends object">
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

/** Provides the shared sortable, filterable, editable, resizable, and virtualized data table. */
import { FilterMatchMode } from '@primevue/core/api'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { DATA_TABLE_FILTER_MODE } from './dataTableFilterModes'
import DataTableTriStateFilter from './DataTableTriStateFilter.vue'
import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import { useToast } from '@/composables/useToast'
import { formatDate, formatDateTime } from '@/i18n/format'
import { stringifyUnknown } from '@/lib/collections'

import type {
  CellEditPayload,
  ColumnDef,
  DataTableSlots,
  InlineEditOptions,
  InlineEditToastInput,
} from './dataTableTypes'
import type {
  DataTableCellEditCompleteEvent,
  DataTableColumnResizeEndEvent,
} from 'primevue/datatable'
import type { VirtualScrollerProps } from 'primevue/virtualscroller'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    loading: 'Loading data...',
    empty: 'No data found.',
    error: 'Could not save "{field}"',
    compactHeight: 'Compact table height',
    expandHeight: 'Expand table height',
    clearFilters: 'Clear all filters',
    filterPlaceholder: 'Filter...',
    confirmEdit: 'Confirm',
    cancelEdit: 'Cancel',
  },
}

// Keep inline-edit feedback durations consistent across all consumers
const INLINE_EDIT_DEFAULT_LIFE = 3_000
const INLINE_EDIT_VALIDATION_LIFE = 4_000

type ExpandedRowsModel = Record<string, boolean> | T[]

const props = withDefaults(
  defineProps<{
    rows: T[]
    columns: ColumnDef<T>[]
    dataKey?: string
    /** Render a leading row-expander column and expose the `expansion` slot. */
    expandable?: boolean
    /** PrimeVue expanded-row model, keyed by `dataKey` or supplied as row objects. */
    expandedRows?: ExpandedRowsModel
    loading?: boolean
    paginator?: boolean
    rowsPerPage?: number
    rowsPerPageOptions?: number[]
    /** Group rows under subheaders by this field. */
    groupBy?: string | null
    reorderableColumns?: boolean
    resizableColumns?: boolean
    columnResizeMode?: 'fit' | 'expand'
    editMode?: 'none' | 'cell'
    /** Conditional class applied to complete rows. */
    rowClass?: (row: T) => string | undefined
    /** Configure validation, persistence, feedback, and rollback for inline edits. */
    inlineEdit?: InlineEditOptions<T>
    /**
     * When `true`, input/number cell edits commit immediately on blur (focus
     * leaving the cell) instead of requiring the user to click the tick (✓)
     * button. The tick/cross buttons are hidden entirely. Use this in contexts
     * where every keystroke is speculative anyway (e.g. TargetsEditor). Default
     * `false` — all other DataTable usages keep the explicit confirm/cancel flow.
     */
    commitOnBlur?: boolean
    /**
     * Message shown in the loading placeholder while the initial data fetch is
     * in progress. Defaults to the component's built-in loading string.
     */
    loadingMessage?: string
    /**
     * Message shown in the standard empty state (centered inbox icon + text)
     * when the table has no rows. Defaults to the component's built-in empty string.
     * Consumers can still fully replace it via the `#empty` slot.
     */
    emptyMessage?: string
    /** Replace the whole table with the shared empty placeholder when the source rows are empty. */
    hideWhenEmpty?: boolean
    /** Prepend a 1-based row-number column (after any actions column). */
    rowNumber?: boolean
    /** Header label for the row-number column. Default `'#'`. */
    rowNumberHeader?: string
    /** PrimeVue density. Shared admin/config tables default to `small`. */
    tableSize?: 'small' | 'normal' | 'large'
    /** Optional body-row height as a CSS length. Defaults to the shared preset value. */
    rowHeight?: string
    /**
     * Max height of the scroll body (all tables scroll rather than paginate).
     * Applied as a CSS `max-height` so short tables shrink to fit their rows and
     * only tables taller than this cap scroll. Default `'51.5rem'`.
     */
    scrollHeight?: string
    /**
     * Show the shared compact/expanded table-height control in the actions header.
     * The compact height defaults to `26rem`; `scrollHeight` remains the expanded
     * height. Consumers may persist the choice by supplying `heightStorageKey`.
     */
    heightToggle?: boolean
    compactScrollHeight?: string
    heightStorageKey?: string
    /** Show the clear-all-filters button when the table has filters. Default true. */
    clearFilterButton?: boolean
    /**
     * Show PrimeVue's per-column filter funnel button on text-like columns
     * (offering the match-mode popover). Default `true`. Action / checkbox / noop
     * columns never show it.
     */
    filterFunnel?: boolean
    /**
     * Side of the table where action column(s) appear. `'end'` (default) places
     * them as the rightmost column(s); `'start'` places them leftmost. The clear-
     * all-filters button always stays at the start so it is easy to reach regardless
     * of where the actions land.
     */
    actionsPosition?: 'start' | 'end'
    /**
     * Draw a 1px vertical border between every column (header + filter row + body),
     * matching DataGrid's per-cell border style. Default `false`.
     */
    columnBorders?: boolean
    /**
     * Disable horizontal scrolling and force all columns to fit within the table
     * width (Tabulator `layout: "fitColumns"`). `width` reserves a fixed allocation,
     * `minWidth` supplies the allocation basis for fixed-layout columns, and columns
     * without either value share the remaining space. Default `false` (scrollable).
     */
    fitColumns?: boolean
    /**
     * Optional caller-controlled token. When it changes, the wrapper resets the
     * table's scroll containers back to the top-left after the DOM has settled.
     * Existing tables opt out by leaving it undefined.
     */
    scrollResetKey?: unknown
    /**
     * Opt-in PrimeVue virtual scrolling for large grids. Off by default so
     * existing datatables keep their current DOM/rendering behavior.
     */
    virtualScroll?: boolean
    /** Extra rows rendered above/below the viewport to keep scrolling smooth. */
    virtualScrollToleratedItems?: number
  }>(),
  {
    dataKey: 'id',
    expandable: false,
    expandedRows: undefined,
    loading: false,

    // Tables scroll (fixed max-height body) rather than paginate, for a uniform
    // look with the scope-targets grid. Opt back into paging with :paginator.
    paginator: false,
    rowsPerPage: 10,

    // Keep the established page-size choices when callers do not provide their own
    rowsPerPageOptions: () => [5, 10, 20, 50, 100, 1000, 10000],
    groupBy: null,
    reorderableColumns: true,
    resizableColumns: true,
    columnResizeMode: 'fit',
    editMode: 'none',
    rowClass: undefined,
    inlineEdit: undefined,
    commitOnBlur: false,
    emptyMessage: undefined,
    hideWhenEmpty: false,
    rowNumber: false,
    rowNumberHeader: '#',
    tableSize: 'small',
    rowHeight: '2.5rem',

    // Applied as a CSS `max-height` (not PrimeVue's fixed `scroll-height`) so the
    // body shrinks to fit few rows but caps + scrolls beyond this. 51.5rem —
    // shared default so every table (admin + scope-targets) scrolls identically.
    scrollHeight: '51.5rem',
    heightToggle: false,
    compactScrollHeight: '26rem',
    heightStorageKey: undefined,
    clearFilterButton: true,
    filterFunnel: true,
    actionsPosition: 'end',
    columnBorders: false,
    fitColumns: false,
    scrollResetKey: undefined,
    virtualScroll: false,
    virtualScrollToleratedItems: 20,
  },
)

const emit = defineEmits<{
  cellEdit: [edit: CellEditPayload<T>]
  visibleRowCapacityChange: [capacity: number]
  'update:expandedRows': [value: ExpandedRowsModel]
  heightChange: [expanded: boolean]
}>()

// ── Composables ───────────────────────────────────────────────
const slots = defineSlots<DataTableSlots<T>>()
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()

function expansionKey(row: T): string {
  return String(Reflect.get(row, props.dataKey))
}

function expandedKeys(value: ExpandedRowsModel | undefined): Set<string> {
  if (!value) return new Set()
  if (Array.isArray(value)) return new Set(value.map(expansionKey))
  return new Set(Object.keys(value).filter((key) => value[key]))
}

/**
 * PrimeVue permits several expanded rows. SSS tables use accordion behavior:
 * retain the row added by the current interaction and close the old row in the
 * same model update.
 */
function keepSingleExpandedRow(
  value: ExpandedRowsModel | undefined,
  previous: ExpandedRowsModel | undefined,
): ExpandedRowsModel | undefined {
  if (!value) return value

  const previousKeys = expandedKeys(previous)
  if (Array.isArray(value)) {
    if (value.length <= 1) return value
    const addedRows = value.filter((row) => !previousKeys.has(expansionKey(row)))
    const rowToKeep = addedRows.at(-1) ?? value.at(-1)
    return rowToKeep ? [rowToKeep] : []
  }

  const activeKeys = Object.keys(value).filter((key) => value[key])
  if (activeKeys.length === 0) return {}
  const onlyKey = activeKeys.at(0)
  if (activeKeys.length === 1 && onlyKey !== undefined) return { [onlyKey]: true }

  const addedKeys = activeKeys.filter((key) => !previousKeys.has(key))
  const keyToKeep = addedKeys.at(-1) ?? activeKeys.at(-1)
  return keyToKeep ? { [keyToKeep]: true } : {}
}

const effectiveExpandedRows = computed(() =>
  keepSingleExpandedRow(props.expandedRows, props.expandedRows),
)

function updateExpandedRows(value: ExpandedRowsModel): void {
  emit('update:expandedRows', keepSingleExpandedRow(value, props.expandedRows) ?? value)
}

// ── State ─────────────────────────────────────────────────────
// Staged value for the active tick/cross editor. Only one cell is open at a
// time, so a single ref suffices. row[field] holds the committed value
// throughout — we do NOT mutate it until the user confirms with the tick.
// NOTE: we do NOT store a row reference here because the #editor slot provides
// editingRowData (a PrimeVue copy), not the real rowData that drives the body
// cell display. Commits always go through onCellEditComplete using event.data.
const stagedEdit = ref<{ field: string; value: unknown } | null>(null)
const firstRowIndex = ref(0)
const heightExpanded = ref(readStoredHeightState())
const virtualBodyHeight = ref<number | null>(null)
const cssVirtualRowHeight = ref<number | null>(null)
const measuredVirtualRowHeight = ref<number | null>(null)

type VirtualScrollerRef = {
  element?: HTMLElement | null
  calculateAutoSize?: () => void
  init?: () => void
  setSpacerSize?: () => void
  scrollTo?: (options: ScrollToOptions) => void
  scrollToIndex?: (index: number, behavior?: ScrollBehavior) => void
}

type PrimeDataTableRef = {
  getVirtualScrollerRef?: () => VirtualScrollerRef | null
  resizeTableCells?: (columnWidth: number, nextColumnWidth?: number) => void
}

type ColumnResizeSnapshot = {
  column: HTMLElement
  columnWidth: number
  columnMinWidth: number
  columnMaxWidth: number
  nextColumnWidth: number
  nextColumnMinWidth: number
  nextColumnMaxWidth: number
}

const primeTableRef = ref<PrimeDataTableRef | null>(null)
let columnResizeSnapshot: ColumnResizeSnapshot | null = null
let bottomScrollToken = 0
let scrollRestoreToken = 0
let virtualRowResizeObserver: ResizeObserver | null = null
let observedVirtualRow: HTMLElement | null = null
let virtualRowMeasurePending = false
let lastVisibleRowCapacity = 0
let virtualLayoutRefreshToken = 0
let isUnmounted = false

/**
 * PrimeVue renders row details in one cell spanning the complete row. Split
 * that structure into an actual leading table cell and a detail cell spanning
 * the remaining columns. Using real cells keeps the expansion geometry tied to
 * the table's own column layout instead of approximating it with a nested div.
 */
function splitExpansionCell(target: unknown): void {
  if (typeof HTMLTableCellElement === 'undefined' || !(target instanceof HTMLTableCellElement)) {
    return
  }

  const row = target.parentElement
  if (!(row instanceof HTMLTableRowElement)) return

  let leadingCell = row.querySelector<HTMLTableCellElement>(
    ':scope > .data-table__expansion-leading-cell',
  )
  if (!leadingCell) {
    leadingCell = document.createElement('td')
    leadingCell.className = 'data-table__expansion-leading-cell'
    leadingCell.setAttribute('aria-hidden', 'true')
    row.insertBefore(leadingCell, target)
  }

  const sourceRow = row.previousElementSibling
  const sourceColumnCount =
    sourceRow instanceof HTMLTableRowElement ? sourceRow.cells.length : target.colSpan
  target.colSpan = Math.max(1, sourceColumnCount - 1)
}

const primeDataTablePt = {
  rowExpansionCell: {
    class: 'data-table__expansion-content-cell',
    ref: splitExpansionCell,
  },
}
const scheduledAnimationFrames = new Set<number>()
const scheduledTimeouts = new Set<ReturnType<typeof setTimeout>>()

function scheduleAnimationFrame(callback: () => void): number | null {
  if (isUnmounted) return null
  const frameId = requestAnimationFrame(() => {
    scheduledAnimationFrames.delete(frameId)
    if (!isUnmounted) callback()
  })
  scheduledAnimationFrames.add(frameId)
  return frameId
}

function scheduleTimeout(
  callback: () => void,
  delay: number,
): ReturnType<typeof setTimeout> | null {
  if (isUnmounted) return null
  const timeoutId = setTimeout(() => {
    scheduledTimeouts.delete(timeoutId)
    if (!isUnmounted) callback()
  }, delay)
  scheduledTimeouts.add(timeoutId)
  return timeoutId
}

// Set by saveInplace so onCellEditComplete can distinguish a tick-triggered
// close (pendingSave set → commit) from a genuine outside click (→ discard).
const pendingSave = ref<{ field: string; value: unknown } | null>(null)
const invalidEdit = ref<{ row: T; field: string } | null>(null)

function isInvalidEdit(row: T, field: string): boolean {
  return invalidEdit.value?.row === row && invalidEdit.value.field === field
}

function clearInvalidEdit(row: T, field: string): void {
  if (isInvalidEdit(row, field)) invalidEdit.value = null
}

function validateEditValue(row: T, field: string, value: unknown): boolean {
  const col = props.columns.find((column) => column.field === field)
  if (!col?.validator) {
    clearInvalidEdit(row, field)
    return true
  }

  const result = col.validator(value, row)
  if (result === true) {
    clearInvalidEdit(row, field)
    return true
  }

  invalidEdit.value = { row, field }
  return false
}

function updateStagedEdit(row: T, field: string, value: unknown): void {
  stagedEdit.value = { field, value }
  if (isInvalidEdit(row, field)) validateEditValue(row, field, value)
}

// PrimeVue's cell editor binds a `mousedown` listener on `document` and closes
// when the event target is NOT inside the editing TD (`!this.$el.contains(target)`).
// Dispatching on document.body with a target that is provably outside any TD
// sets selfClick=false → schedules completeEdit → switchCellToViewMode.
function closeInplaceEditor(): void {
  document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
}

// Called by the tick (✓) button. Promotes the staged value to pendingSave so
// onCellEditComplete (type 'outside') can commit it using event.data (real row).
function saveInplace(row: T, field: string): void {
  const value = stagedEdit.value?.field === field ? stagedEdit.value.value : cellValue(row, field)
  if (!validateEditValue(row, field, value)) return

  if (stagedEdit.value?.field === field) {
    pendingSave.value = { field, value: stagedEdit.value.value }
    stagedEdit.value = null
  }
  closeInplaceEditor()
}

// Called by the cross (✕) button. Discard staged value and close.
function cancelInplace(row: T, field: string): void {
  if (stagedEdit.value?.field === field) {
    stagedEdit.value = null
  }
  clearInvalidEdit(row, field)
  closeInplaceEditor()
}

// PrimeVue filters model. MUST be a `ref` (not `reactive`): the DataTable's
// `onFilterApply` does `$emit('update:filters', d_filters)`, and `v-model:filters`
// reassigns this binding — only a ref can be reassigned, so a `reactive` object
// would silently drop the sync-back and never actually filter. `null` value =
// "no filter" (PrimeVue skips it), which also encodes the tri-state "all" state.
const filters = ref<Record<string, { value: unknown; matchMode: string }>>({})

// Suppress the empty-state ("no entries") until the first load has actually
// settled. On mount `loading` is still false (the fetch fires in the page's
// onMounted right after), so without this guard PrimeVue briefly renders the
// empty message over the yet-to-load table. Flip to true once loading has gone
// true→false at least once, or when rows arrive.
const firstLoadSettled = ref(false)

// ---------------------------------------------------------------------------
// Leading and trailing action/control columns must never move, and data columns
// may not be dropped outside their band. PrimeVue only blocks DRAGGING a
// `reorderableColumn:false` column — it still allows drops around it, so illegal
// orders are rejected after the fact:
// bumping `reorderKey` remounts the inner table, which clears PrimeVue's internal
// column order back to our declarative (last-good) order.
const reorderKey = ref(0)

// ---------------------------------------------------------------------------
// The filter-menu funnel is tinted teal when its column uses a NON-DEFAULT match
// mode (e.g. "Starts with" instead of the default "Contains"). PrimeVue exposes no
// class for that state, so we toggle one ourselves — imperatively, NOT via a `:pt`
// on the column, because any per-render column prop re-creates the adjacent filter
// input and steals its focus on every keystroke. The watcher keys on match modes
// only (which change via the funnel menu, never while typing), so cost scales with
// column count (not rows) and it never fires per keystroke — cheap for large data.
const rootEl = ref<HTMLElement | null>(null)

// ── Computed ──────────────────────────────────────────────────
const resolvedLoadingMessage = computed(() => props.loadingMessage ?? t('loading'))
const resolvedEmptyMessage = computed(() => props.emptyMessage ?? t('empty'))

const hasFilters = computed(() => props.columns.some((c) => c.filterable))

// A leading control column carries the clear-all-filters button in its filter
// cell; only shown when the table both has filters and opts in.
const showClearButton = computed(() => props.clearFilterButton && hasFilters.value)

// Show PrimeVue's filter funnel only when the table opts in (`filterFunnel`)
// AND the column is a text-like filter, where the match-mode popover is useful.
// Boolean (tristate) / noop filters never get a funnel (they cycle / do nothing in
// place). Off by default: the admin tables use inline row filters without funnels.
// Precomputed as a per-field map so the template binds a cached value rather than
// calling a function per column per render (cheaper for wide tables).
const showMenuByField = computed<Record<string, boolean>>(() => {
  const map: Record<string, boolean> = {}
  for (const col of props.columns) {
    map[col.field] =
      col.filterFunnel !== false &&
      props.filterFunnel &&
      !!col.filterable &&
      col.filterType !== 'tristate' &&
      col.filterType !== 'noop'
  }
  return map
})

const actionColumns = computed(() => props.columns.filter((c) => c.formatter === 'actions'))
const dataColumns = computed(() => props.columns.filter((c) => c.formatter !== 'actions'))

const actionsAtStart = computed(() => props.actionsPosition === 'start')

// Count of the fixed leading columns, mirroring the template's render conditions.
// When actions are at the end they are trailing (not leading), so they are excluded.
// The standalone clear column only renders when there are no action columns at all.
const fixedLeadingCount = computed(
  () =>
    (props.expandable ? 1 : 0) +
    (actionsAtStart.value ? actionColumns.value.length : 0) +
    (props.rowNumber ? 1 : 0) +
    (showClearButton.value && actionColumns.value.length === 0 && actionsAtStart.value ? 1 : 0),
)

// PrimeVue prevents dragging a non-reorderable trailing control, but still
// permits another column to be dropped beyond it unless we reject that order.
const fixedTrailingCount = computed(() => {
  if (!actionsAtStart.value && actionColumns.value.length > 0) return actionColumns.value.length
  if (actionColumns.value.length === 0 && props.heightToggle) return 1
  if (actionColumns.value.length === 0 && showClearButton.value && !actionsAtStart.value) return 1
  return 0
})

const funnelColumns = computed(() =>
  dataColumns.value.filter((c) => showMenuByField.value[c.field]),
)

// Flips only when a funnel column's active (non-default match mode) state changes.
const funnelActiveSignature = computed(() =>
  funnelColumns.value
    .map((c) => {
      const f = filters.value[c.field]
      return f && f.matchMode !== matchModeFor(c) ? '1' : '0'
    })
    .join(''),
)

// Only show the empty placeholder after the first load settled and nothing is
// currently loading — otherwise the cell stays blank during the initial fetch.
// NOT gated on rows.length: PrimeVue invokes the #empty slot only when the
// DISPLAYED (post-filter) row count is 0, which also covers "a filter matched
// nothing" even when the underlying data is non-empty (e.g. the scope-targets
// grid always keeps a trailing blank row, so rows.length is never 0).
const showEmptyState = computed(() => firstLoadSettled.value && !props.loading)
const showExternalEmptyState = computed(
  () => props.hideWhenEmpty && firstLoadSettled.value && !props.loading && props.rows.length === 0,
)
const resolvedVirtualRowHeight = computed(
  () =>
    measuredVirtualRowHeight.value ??
    cssVirtualRowHeight.value ??
    resolveDataTableRowHeightPixels(),
)
const resolvedVirtualScrollerOptions = computed<VirtualScrollerProps | undefined>(() =>
  props.virtualScroll
    ? {
        itemSize: resolvedVirtualRowHeight.value,
        numToleratedItems: props.virtualScrollToleratedItems,
        delay: 0,
      }
    : undefined,
)
const effectiveScrollHeight = computed(() =>
  props.heightToggle && !heightExpanded.value ? props.compactScrollHeight : props.scrollHeight,
)
const virtualBodyHeightStyle = computed(() =>
  virtualBodyHeight.value && virtualBodyHeight.value > 0
    ? `${virtualBodyHeight.value}px`
    : effectiveScrollHeight.value,
)

// ── Functions and handlers ────────────────────────────────────
// Custom filter matchmodes (`dateText`, `noop`, `deepContains`) are registered in ./filterModes,
// shared with the scope-targets grid; imported above so the side-effect runs.

// Map a column's `filterType` to a PrimeVue match mode (explicit override wins).
function matchModeFor(col: ColumnDef<T>): string {
  if (col.matchMode) return col.matchMode
  if (!col.filterType && col.formatter === 'datetime') return DATA_TABLE_FILTER_MODE.dateText
  switch (col.filterType) {
    case 'select':
    case 'tristate':
      return FilterMatchMode.EQUALS
    case 'dateText':
      return DATA_TABLE_FILTER_MODE.dateText
    case 'noop':
      return DATA_TABLE_FILTER_MODE.noop
    case 'text':
    default:
      return FilterMatchMode.CONTAINS
  }
}

// Reconcile the filters model to the current filterable columns, PRESERVING any
// existing value/matchMode for fields that persist (so a consumer whose `columns`
// is a computed returning new arrays, or that tweaks lookupMap/editorOptions, does
// not wipe active filters). Only added fields are seeded and removed fields dropped.
function reconcileFilters(): void {
  const prev = filters.value
  const next: Record<string, { value: unknown; matchMode: string }> = {}
  for (const col of props.columns) {
    if (col.filterable) {
      next[col.field] = prev[col.field] ?? { value: null, matchMode: matchModeFor(col) }
    }
  }
  filters.value = next
}
reconcileFilters()

// Reset every column filter: clear the value AND restore each column's default
// match mode, so "clear filters" also un-tints any funnel left on a non-default
// mode. Reassign as a fresh object so PrimeVue re-clones and re-applies.
function clearAll(): void {
  const next: Record<string, { value: unknown; matchMode: string }> = {}
  for (const col of props.columns) {
    if (col.filterable) next[col.field] = { value: null, matchMode: matchModeFor(col) }
  }
  filters.value = next
}

// Expose the filter reset so consumers that own surrounding logic (e.g. the
// scope-targets grid clearing filters after a paste) can trigger it via a ref.
defineExpose({
  clearAllFilters: clearAll,
  getVisibleRowCapacity,
  isScrolledToBottom,
  nudgeScrollAboveBottom,
  preserveScrollPosition,
  resetScrollPosition,
  scrollToBottomPosition,
})

// Delay text filtering until typing pauses while always applying the latest render callback.
const filterTimers = new Map<string, ReturnType<typeof setTimeout>>()

function scheduleFilter(field: string, callback: () => void): void {
  cancelScheduledFilter(field)

  const timer = setTimeout(() => {
    filterTimers.delete(field)
    callback()
  }, 500)
  filterTimers.set(field, timer)
}

function applyFilter(field: string, callback: () => void): void {
  cancelScheduledFilter(field)
  callback()
}

function cancelScheduledFilter(field: string): void {
  const timer = filterTimers.get(field)
  if (timer === undefined) return

  clearTimeout(timer)
  filterTimers.delete(field)
}

function onColumnReorder(event: { dragIndex: number; dropIndex: number }): void {
  const firstDataIndex = fixedLeadingCount.value
  const firstTrailingIndex = firstDataIndex + dataColumns.value.length
  const draggedFixedColumn =
    event.dragIndex < firstDataIndex || event.dragIndex >= firstTrailingIndex
  const droppedOutsideDataBand =
    event.dropIndex < firstDataIndex ||
    (fixedTrailingCount.value > 0 && event.dropIndex >= firstTrailingIndex)

  if (draggedFixedColumn || droppedOutsideDataBand) {
    reorderKey.value++
  }
}

function syncFunnelTint(): void {
  const el = rootEl.value
  if (!el) return
  const buttons = el.querySelectorAll<HTMLElement>('.p-datatable-column-filter-button')
  const cols = funnelColumns.value
  buttons.forEach((btn, i) => {
    const col = cols[i]
    const f = col ? filters.value[col.field] : undefined
    const active = !!col && !!f && f.matchMode !== matchModeFor(col)
    btn.classList.toggle('data-table__filter-btn--active', active)
  })
}

function maxScrollTop(element: HTMLElement): number {
  return Math.max(0, element.scrollHeight - element.clientHeight)
}

function resolveCssLengthPixels(value: string, host: HTMLElement): number {
  if (typeof document === 'undefined') return 0

  const probe = document.createElement('div')
  probe.style.position = 'absolute'
  probe.style.visibility = 'hidden'
  probe.style.pointerEvents = 'none'
  probe.style.height = value
  probe.style.width = '0'

  host.appendChild(probe)
  const height = probe.getBoundingClientRect().height
  probe.remove()

  return height > 0 ? height : 0
}

function resolveDataTableRowHeightPixels(): number {
  if (typeof document === 'undefined') return 0

  const host = rootEl.value ?? document.body ?? document.documentElement
  return (
    resolveCssLengthPixels('var(--app-data-table-row-height)', host) ||
    resolveCssLengthPixels('2.5rem', host)
  )
}

function syncCssVirtualRowHeight(): void {
  const height = resolveDataTableRowHeightPixels()
  if (height > 0 && Math.abs(height - (cssVirtualRowHeight.value ?? 0)) > 0.5) {
    cssVirtualRowHeight.value = height
  }
}

function disconnectVirtualRowHeightObserver(): void {
  virtualRowResizeObserver?.disconnect()
  virtualRowResizeObserver = null
  observedVirtualRow = null
  virtualRowMeasurePending = false
}

function syncVirtualRowHeightObserver(row: HTMLElement | null): void {
  if (!props.virtualScroll || typeof ResizeObserver === 'undefined') {
    disconnectVirtualRowHeightObserver()
    return
  }

  if (row === observedVirtualRow && virtualRowResizeObserver) return

  virtualRowResizeObserver?.disconnect()
  observedVirtualRow = row
  virtualRowResizeObserver = null

  const root = rootEl.value
  if (!row && !root) return

  virtualRowResizeObserver = new ResizeObserver(() => measureVirtualRowHeight())
  if (root) virtualRowResizeObserver.observe(root)
  if (row) virtualRowResizeObserver.observe(row)
}

function isEmptyMessageRow(row: HTMLElement): boolean {
  return (
    row.classList.contains('p-datatable-emptymessage') ||
    row.classList.contains('p-datatable-empty-message') ||
    row.querySelector('.data-table__empty') != null
  )
}

function firstRenderedDataRow(): HTMLElement | null {
  const rows = rootEl.value?.querySelectorAll<HTMLElement>('.p-datatable-tbody > tr')
  if (!rows) return null

  for (const row of rows) {
    if (!isEmptyMessageRow(row)) return row
  }

  return null
}

function applyVirtualRowHeightMeasurement(): void {
  if (!props.virtualScroll) return

  syncCssVirtualRowHeight()

  const row = firstRenderedDataRow()
  syncVirtualRowHeightObserver(row)
  if (!row) {
    syncVisibleRowCapacity()
    return
  }

  const height = row.getBoundingClientRect().height
  if (height > 0 && Math.abs(height - resolvedVirtualRowHeight.value) > 0.5) {
    measuredVirtualRowHeight.value = height
    scheduleVirtualScrollerLayoutRefresh()
  }
  syncVisibleRowCapacity()
}

function measureVirtualRowHeight(): void {
  if (!props.virtualScroll || virtualRowMeasurePending) return

  virtualRowMeasurePending = true

  const measure = () => {
    virtualRowMeasurePending = false
    applyVirtualRowHeightMeasurement()
  }

  void nextTick(() => {
    if (isUnmounted) return
    if (typeof requestAnimationFrame !== 'function') {
      scheduleTimeout(measure, 0)
      return
    }
    scheduleAnimationFrame(measure)
  })
}

function isAtScrollBottom(element: HTMLElement): boolean {
  return maxScrollTop(element) - element.scrollTop <= 2
}

function isScrolledToBottom(): boolean {
  const container = scrollContainer()
  return container ? isAtScrollBottom(container) : false
}

function nudgeScrollAboveBottom(): void {
  const container = scrollContainer()
  if (!container || !isAtScrollBottom(container)) return

  container.scrollTop = Math.max(0, maxScrollTop(container) - 2)
}

function nudgeVirtualScroller(container: HTMLElement): void {
  container.dispatchEvent(new Event('scroll'))
}

function refreshVirtualScrollerLayout(): void {
  if (!props.virtualScroll) return

  const virtualScroller = primeTableRef.value?.getVirtualScrollerRef?.()
  const container = virtualScroller?.element ?? scrollContainer()

  virtualScroller?.init?.()
  virtualScroller?.calculateAutoSize?.()
  virtualScroller?.setSpacerSize?.()

  if (container) {
    nudgeVirtualScroller(container)
  }

  syncVisibleRowCapacity()
}

function scheduleVirtualScrollerLayoutRefresh(): void {
  if (!props.virtualScroll) return

  const token = ++virtualLayoutRefreshToken

  void nextTick(() => {
    if (isUnmounted) return
    const refresh = () => {
      if (token !== virtualLayoutRefreshToken) return
      refreshVirtualScrollerLayout()
    }

    refresh()

    if (typeof requestAnimationFrame !== 'function') {
      scheduleTimeout(refresh, 0)
      return
    }

    let remainingFrames = 3
    const refreshOnFrame = () => {
      refresh()
      remainingFrames -= 1
      if (remainingFrames > 0) scheduleAnimationFrame(refreshOnFrame)
    }
    scheduleAnimationFrame(refreshOnFrame)
  })
}

function scrollContainer(): HTMLElement | null {
  const el = rootEl.value
  if (!el) return null

  if (props.virtualScroll) {
    return (
      primeTableRef.value?.getVirtualScrollerRef?.()?.element ??
      el.querySelector<HTMLElement>('.p-virtualscroller')
    )
  }

  return el.querySelector<HTMLElement>('.p-datatable-table-container')
}

function resolveScrollHeightPixels(): number {
  const el = rootEl.value
  if (!el || typeof document === 'undefined') return 0

  const probe = document.createElement('div')
  probe.style.position = 'absolute'
  probe.style.visibility = 'hidden'
  probe.style.pointerEvents = 'none'
  probe.style.height = effectiveScrollHeight.value
  probe.style.width = '0'

  el.appendChild(probe)
  const height = probe.getBoundingClientRect().height
  probe.remove()

  return height > 0 ? height : 0
}

function visibleVirtualBodyHeight(container: HTMLElement): number {
  const root = rootEl.value
  if (!root) return 0

  const containerRect = container.getBoundingClientRect()
  const contentBottom = containerRect.top + container.clientHeight
  if (contentBottom <= containerRect.top) return 0

  let viewportTop = containerRect.top
  const header = root.querySelector<HTMLElement>('.p-datatable-thead')
  const headerRect = header?.getBoundingClientRect()

  if (
    headerRect &&
    headerRect.height > 0 &&
    headerRect.bottom > containerRect.top &&
    headerRect.top < contentBottom
  ) {
    viewportTop = Math.min(Math.max(headerRect.bottom, containerRect.top), contentBottom)
  }

  return Math.max(0, contentBottom - viewportTop)
}

function renderedVirtualRowHeight(): number {
  const row = firstRenderedDataRow()
  return row?.getBoundingClientRect().height ?? 0
}

function getVisibleRowCapacity(): number {
  if (!props.virtualScroll) return Math.max(1, props.rows.length)

  const container = scrollContainer()
  const rowHeight = renderedVirtualRowHeight() || resolvedVirtualRowHeight.value
  if (rowHeight <= 0) return 1

  let viewportHeight = container ? visibleVirtualBodyHeight(container) : 0
  if (viewportHeight <= 0) viewportHeight = container?.clientHeight ?? 0
  if (viewportHeight <= 0) viewportHeight = resolveScrollHeightPixels()
  if (viewportHeight > 0) virtualBodyHeight.value = viewportHeight

  return Math.max(1, Math.floor(viewportHeight / rowHeight))
}

function syncVisibleRowCapacity(): void {
  if (!props.virtualScroll) return

  const capacity = getVisibleRowCapacity()
  if (capacity === lastVisibleRowCapacity) return

  lastVisibleRowCapacity = capacity
  emit('visibleRowCapacityChange', capacity)
  void nextTick(() => primeTableRef.value?.getVirtualScrollerRef?.()?.setSpacerSize?.())
}

function restoreScrollPosition(top: number, left: number): void {
  const container = scrollContainer()
  if (!container) return

  if (props.virtualScroll) {
    primeTableRef.value?.getVirtualScrollerRef?.()?.setSpacerSize?.()
  }

  container.scrollTop = Math.min(top, maxScrollTop(container))
  container.scrollLeft = left

  if (props.virtualScroll) nudgeVirtualScroller(container)
}

function preserveScrollPosition(): void {
  const container = scrollContainer()
  if (!container) return

  const top = container.scrollTop
  const left = container.scrollLeft
  const stickToBottom = isAtScrollBottom(container)
  const token = ++scrollRestoreToken

  void nextTick(() => {
    if (isUnmounted) return
    if (typeof requestAnimationFrame !== 'function') {
      scheduleTimeout(() => {
        if (token !== scrollRestoreToken) return
        if (stickToBottom) scrollToBottomPosition()
        else restoreScrollPosition(top, left)
      }, 0)
      return
    }

    let remainingFrames = props.virtualScroll ? 3 : 1
    const restoreOnFrame = () => {
      if (token !== scrollRestoreToken) return
      if (stickToBottom) {
        scrollToBottomPosition()
        return
      }

      restoreScrollPosition(top, left)
      remainingFrames -= 1
      if (remainingFrames > 0) scheduleAnimationFrame(restoreOnFrame)
    }
    scheduleAnimationFrame(restoreOnFrame)
  })
}

function resetScrollPosition(): void {
  firstRowIndex.value = 0

  const reset = () => {
    firstRowIndex.value = 0
    const virtualScroller = primeTableRef.value?.getVirtualScrollerRef?.()
    if (virtualScroller?.element) virtualScroller.scrollToIndex?.(0)
    virtualScroller?.scrollTo?.({ top: 0, left: 0 })

    const el = rootEl.value
    if (!el) return
    const containers = el.querySelectorAll<HTMLElement>(
      '.p-datatable-table-container, .p-virtualscroller',
    )
    containers.forEach((container) => {
      container.scrollTop = 0
      container.scrollLeft = 0
    })
  }

  void nextTick(() => {
    if (isUnmounted) return
    reset()
    if (typeof requestAnimationFrame !== 'function') {
      scheduleTimeout(reset, 0)
      return
    }

    let remainingFrames = 3
    const resetOnFrame = () => {
      reset()
      remainingFrames -= 1
      if (remainingFrames > 0) scheduleAnimationFrame(resetOnFrame)
    }
    scheduleAnimationFrame(resetOnFrame)
  })
}

function scrollToBottomPosition(): void {
  const token = ++bottomScrollToken

  const scroll = () => {
    const el = rootEl.value
    if (!el) return

    if (props.virtualScroll) {
      const virtualScroller = primeTableRef.value?.getVirtualScrollerRef?.()
      const container = virtualScroller?.element ?? scrollContainer()
      if (!container) return

      virtualScroller?.setSpacerSize?.()
      const top = maxScrollTop(container)
      virtualScroller?.scrollTo?.({ top, left: container.scrollLeft })
      container.scrollTop = top
      nudgeVirtualScroller(container)
      return
    }

    const container = el.querySelector<HTMLElement>('.p-datatable-table-container')
    if (container) container.scrollTop = maxScrollTop(container)
  }

  void nextTick(() => {
    if (isUnmounted) return
    if (typeof requestAnimationFrame !== 'function') {
      scheduleTimeout(scroll, 0)
      return
    }

    // Let Vue/PrimeVue settle, then retry for a few frames. setSpacerSize() is a
    // lightweight spacer-height update; unlike init(), it does not blank the
    // rendered rows. The repeated calls cover the case where the new item count
    // reaches the VirtualScroller one frame before the DOM height does.
    let remainingFrames = props.virtualScroll ? 5 : 1
    const waitForVirtualLayout = () => {
      if (token !== bottomScrollToken) return
      remainingFrames -= 1
      scroll()
      if (remainingFrames > 0) scheduleAnimationFrame(waitForVirtualLayout)
    }
    scheduleAnimationFrame(waitForVirtualLayout)
  })
}

// ---------------------------------------------------------------------------
function cellValue(row: T, field: string): unknown {
  return Reflect.get(row, field)
}

function setCellValue(row: T, field: string, value: unknown): void {
  Reflect.set(row, field, value)
}

function lookupLabel(col: ColumnDef<T>, value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value !== 'string' && typeof value !== 'number') return stringifyUnknown(value)

  return col.lookupMap?.[value] ?? stringifyUnknown(value)
}

function dateValue(value: unknown): Date | string | number | null | undefined {
  if (
    value === null ||
    value === undefined ||
    value instanceof Date ||
    typeof value === 'string' ||
    typeof value === 'number'
  ) {
    return value
  }

  return undefined
}

function triStateValue(value: unknown): boolean | null {
  return typeof value === 'boolean' ? value : null
}

function numberValue(value: unknown): number | null {
  return typeof value === 'number' ? value : null
}

function isTruthyBool(value: unknown): boolean {
  return value === true || value === 1 || value === '1' || value === 'true'
}

function isEditable(col: ColumnDef<T>): boolean {
  return props.editMode === 'cell' && !!col.editor
}

// Fixed table layout ignores min-width during allocation, so use it as the fit-mode basis
function columnStyle(col: ColumnDef<T>): Record<string, string | undefined> | undefined {
  if (!col.width && !col.minWidth && !col.maxWidth) return undefined

  return {
    width: col.width ?? (props.fitColumns ? col.minWidth : undefined),
    minWidth: col.minWidth,
    maxWidth: col.maxWidth,
  }
}

// Resolve configured CSS lengths without reading PrimeVue's generated resize rules
function columnPixelLimit(column: HTMLElement, property: 'minWidth' | 'maxWidth'): number {
  const configuredLength = column.style[property]
  const host = rootEl.value
  if (!configuredLength || !host) {
    return property === 'minWidth' ? 15 : Number.POSITIVE_INFINITY
  }

  const probe = document.createElement('div')
  probe.style.position = 'absolute'
  probe.style.visibility = 'hidden'
  probe.style.pointerEvents = 'none'
  probe.style.width = configuredLength
  host.appendChild(probe)

  const value = probe.getBoundingClientRect().width
  probe.remove()
  return Number.isFinite(value) ? value : property === 'minWidth' ? 15 : Number.POSITIVE_INFINITY
}

// Capture both columns because PrimeVue fit-mode resizing always changes the neighboring column
function onColumnResizePointerDown(event: MouseEvent): void {
  if (!props.fitColumns || props.columnResizeMode !== 'fit') return

  const target = event.target
  if (!(target instanceof Element)) return

  const resizer = target.closest('.p-datatable-column-resizer')
  const column = resizer?.parentElement
  const nextColumn = column?.nextElementSibling
  if (!(column instanceof HTMLElement) || !(nextColumn instanceof HTMLElement)) return

  const nextColumnIsFixed = nextColumn.classList.contains('data-table__col-no-resize')
  columnResizeSnapshot = {
    column,
    columnWidth: column.offsetWidth,
    columnMinWidth: Math.max(15, columnPixelLimit(column, 'minWidth')),
    columnMaxWidth: columnPixelLimit(column, 'maxWidth'),
    nextColumnWidth: nextColumn.offsetWidth,
    nextColumnMinWidth: nextColumnIsFixed
      ? nextColumn.offsetWidth
      : Math.max(15, columnPixelLimit(nextColumn, 'minWidth')),
    nextColumnMaxWidth: nextColumnIsFixed
      ? nextColumn.offsetWidth
      : columnPixelLimit(nextColumn, 'maxWidth'),
  }
}

// Replace an invalid PrimeVue result synchronously before the browser can paint the overflow
function onColumnResizeEnd(event: DataTableColumnResizeEndEvent): void {
  const snapshot = columnResizeSnapshot
  columnResizeSnapshot = null
  if (!snapshot || snapshot.column !== event.element) return

  const requestedDelta = Number(event.delta)
  if (!Number.isFinite(requestedDelta)) return

  const minimumDelta = Math.max(
    snapshot.columnMinWidth - snapshot.columnWidth,
    snapshot.nextColumnWidth - snapshot.nextColumnMaxWidth,
  )
  const maximumDelta = Math.min(
    snapshot.columnMaxWidth - snapshot.columnWidth,
    snapshot.nextColumnWidth - snapshot.nextColumnMinWidth,
  )
  const clampedDelta =
    minimumDelta <= maximumDelta
      ? Math.min(Math.max(requestedDelta, minimumDelta), maximumDelta)
      : 0

  if (clampedDelta === requestedDelta) return
  primeTableRef.value?.resizeTableCells?.(
    snapshot.columnWidth + clampedDelta,
    snapshot.nextColumnWidth - clampedDelta,
  )
}

function bodyCellClass(col: ColumnDef<T>, row: T): string {
  const parts: string[] = []
  if (isInvalidEdit(row, col.field)) parts.push('data-table__cell--invalid')
  if (col.sensitive) parts.push('sensitive')
  if (col.bodyClass) {
    const extra = col.bodyClass(row)
    if (extra) parts.push(extra)
  }
  return parts.join(' ')
}

// Expose clipped cell content without adding redundant titles to values that fit
function onCellMouseOver(event: MouseEvent): void {
  const target = event.target
  if (!(target instanceof Element)) return

  const cell = target.closest('td')
  if (!(cell instanceof HTMLTableCellElement)) return

  const isOverflowHidden = window.getComputedStyle(cell).overflowX === 'hidden'
  const content = cell.innerText.trim()
  if (isOverflowHidden && cell.scrollWidth > cell.clientWidth && content !== '') {
    cell.title = content
  } else {
    cell.removeAttribute('title')
  }
}

function rejectionReason(err: unknown): string | undefined {
  if (err instanceof Error) return err.message
  if (typeof err === 'string' && err.trim() !== '') return err
  return undefined
}

// Surface consumer-provided edit feedback through the table's shared toast service
function showInlineEditToast(
  severity: 'success' | 'error',
  input: InlineEditToastInput,
  defaultLife: number,
): void {
  if (input === null) return

  const message = typeof input === 'string' ? { summary: input } : input
  toast.add({
    severity,
    summary: message.summary,
    detail: message.detail,
    life: message.life ?? defaultLife,
  })
}

// ---------------------------------------------------------------------------
// Validate, persist, and emit edits through one path for every editor type.
async function commitEdit(
  row: T,
  field: string,
  newValue: unknown,
  oldValue: unknown,
): Promise<void> {
  const col = props.columns.find((column) => column.field === field)

  if (newValue === oldValue) {
    clearInvalidEdit(row, field)
    return
  }

  if (!validateEditValue(row, field, newValue)) {
    setCellValue(row, field, oldValue)
    return
  }

  const editPayload: CellEditPayload<T> = { row, field, newValue, oldValue }
  const inlineEdit = props.inlineEdit
  const validationError = inlineEdit?.validator?.(editPayload) ?? null
  if (validationError !== null) {
    toast.add({
      severity: 'warn',
      summary: validationError,
      life: inlineEdit?.validationLife ?? INLINE_EDIT_VALIDATION_LIFE,
    })
    setCellValue(row, field, oldValue)
    return
  }

  setCellValue(row, field, newValue)
  if (!inlineEdit) {
    emit('cellEdit', editPayload)
    return
  }

  try {
    const successToast = await inlineEdit.save(editPayload)
    showInlineEditToast('success', successToast, inlineEdit.defaultLife ?? INLINE_EDIT_DEFAULT_LIFE)
    emit('cellEdit', editPayload)
  } catch (error: unknown) {
    if (inlineEdit.errorToast) {
      showInlineEditToast(
        'error',
        inlineEdit.errorToast(error, editPayload),
        inlineEdit.defaultLife ?? INLINE_EDIT_DEFAULT_LIFE,
      )
    } else {
      const fieldLabel = col?.header ?? field
      toast.error(t('error', { field: fieldLabel }), rejectionReason(error))
    }

    setCellValue(row, field, oldValue)
  }
}

async function onCellEditComplete(event: DataTableCellEditCompleteEvent<T>): Promise<void> {
  const { data: row, value: oldValue, newValue: completedValue, field } = event
  const col = props.columns.find((c) => c.field === field)
  if (col?.editor === 'select') {
    await commitEdit(row, field, completedValue, oldValue)
  } else if (col?.editor === 'input' || col?.editor === 'number') {
    if (pendingSave.value?.field === field) {
      // Tick button or Enter key set pendingSave and triggered this close.
      // Commit via event.data (real rowData), not the editingRowData copy from the slot.
      const newValue = pendingSave.value.value
      pendingSave.value = null
      stagedEdit.value = null
      await commitEdit(row, field, newValue, oldValue)
    } else if (props.commitOnBlur) {
      // Blur-commit mode: no tick button needed. Use staged value from the built-in
      // inplace editor, or fall back to event.newValue for consumer-slot editors
      // (e.g. #editor:input with v-model that writes directly to editingRowData).
      const newValue = stagedEdit.value?.field === field ? stagedEdit.value.value : completedValue
      stagedEdit.value = null
      await commitEdit(row, field, newValue, oldValue)
    } else {
      // Outside click / Escape without save intent — discard staged value (cancel).
      stagedEdit.value = null
      clearInvalidEdit(row, field)
    }
  }
}

function hasSlot(name: string): boolean {
  return !!slots[name]
}

function readStoredHeightState(): boolean {
  if (!props.heightStorageKey) return false
  try {
    return globalThis.sessionStorage?.getItem(props.heightStorageKey) === 'true'
  } catch {
    return false
  }
}

function toggleTableHeight(): void {
  heightExpanded.value = !heightExpanded.value
  emit('heightChange', heightExpanded.value)
  if (!props.heightStorageKey) return

  try {
    globalThis.sessionStorage?.setItem(props.heightStorageKey, String(heightExpanded.value))
  } catch {
    // Storage can be unavailable in locked-down browsers; the control still works.
  }
}

// ── Watchers and lifecycle ────────────────────────────────────
// Watch a cheap signature (filterable field + default match mode) instead of a deep
// traversal of the whole columns array — re-runs only when the filterable field set
// or a match mode actually changes.
watch(
  () =>
    props.columns
      .filter((c) => c.filterable)
      .map((c) => `${c.field}:${matchModeFor(c)}`)
      .join('|'),
  reconcileFilters,
)

watch(
  () => props.loading,
  (isLoading, wasLoading) => {
    if (wasLoading && !isLoading) firstLoadSettled.value = true
  },
)
watch(
  () => props.rows.length,
  (count) => {
    if (count > 0) firstLoadSettled.value = true
    measureVirtualRowHeight()
    scheduleVirtualScrollerLayoutRefresh()
  },
  { immediate: true },
)

// Re-sync after the DOM settles when the active signature, the funnel set, or a
// reorder-remount changes. `flush: 'post'` so PrimeVue has rendered the buttons.
watch(
  [funnelActiveSignature, () => funnelColumns.value.length, reorderKey],
  () => nextTick(syncFunnelTint),
  { flush: 'post' },
)
watch(() => props.scrollResetKey, resetScrollPosition, { flush: 'post' })
watch(effectiveScrollHeight, () => scheduleVirtualScrollerLayoutRefresh(), { flush: 'post' })
watch(
  () => props.virtualScroll,
  (enabled) => {
    if (enabled) {
      measureVirtualRowHeight()
      scheduleVirtualScrollerLayoutRefresh()
    } else disconnectVirtualRowHeightObserver()
  },
  { flush: 'post' },
)
onMounted(async () => {
  void nextTick(syncFunnelTint)
  measureVirtualRowHeight()
  scheduleVirtualScrollerLayoutRefresh()

  // Let the parent start any mount-time request before treating an empty source as settled.
  await nextTick()
  if (!props.loading) firstLoadSettled.value = true
})
onBeforeUnmount(() => {
  isUnmounted = true
  bottomScrollToken += 1
  scrollRestoreToken += 1
  virtualLayoutRefreshToken += 1
  disconnectVirtualRowHeightObserver()
  filterTimers.forEach((timer) => clearTimeout(timer))
  filterTimers.clear()
  if (typeof cancelAnimationFrame === 'function') {
    scheduledAnimationFrames.forEach((frameId) => cancelAnimationFrame(frameId))
  }
  scheduledAnimationFrames.clear()
  scheduledTimeouts.forEach((timeoutId) => clearTimeout(timeoutId))
  scheduledTimeouts.clear()
})
</script>

<template>
  <LayoutLoading
    v-if="loading && !firstLoadSettled"
    loading
    :loading-message="resolvedLoadingMessage"
    :empty="false"
    empty-message=""
  />
  <LayoutLoading
    v-else-if="showExternalEmptyState"
    :loading="false"
    empty
    :empty-message="resolvedEmptyMessage"
    empty-test-id="adt-empty"
  />
  <div
    v-else
    ref="rootEl"
    class="data-table"
    data-testid="data-table"
    :class="{
      'data-table--column-borders': columnBorders,
      'data-table--fit-columns': fitColumns,
      'data-table--commit-on-blur': commitOnBlur,
      'data-table--virtual-scroll': virtualScroll,
    }"
    :style="{
      '--data-table-max-height': effectiveScrollHeight,
      '--data-table-virtual-row-height': `${resolvedVirtualRowHeight}px`,
      '--data-table-virtual-body-height': virtualBodyHeightStyle,
      '--app-data-table-row-height': rowHeight,
    }"
    @mousedown.capture="onColumnResizePointerDown"
    @mouseover="onCellMouseOver"
  >
    <DataTable
      :key="reorderKey"
      ref="primeTableRef"
      v-model:filters="filters"
      v-model:first="firstRowIndex"
      :value="rows"
      :data-key="dataKey"
      :size="tableSize === 'normal' ? undefined : tableSize"
      :expanded-rows="effectiveExpandedRows"
      :paginator="paginator"
      :rows="rowsPerPage"
      :rows-per-page-options="rowsPerPageOptions"
      :scroll-height="virtualScroll ? effectiveScrollHeight : undefined"
      :virtual-scroller-options="resolvedVirtualScrollerOptions"
      :filter-display="hasFilters ? 'row' : undefined"
      :reorderable-columns="reorderableColumns"
      :resizable-columns="resizableColumns"
      :column-resize-mode="resizableColumns ? columnResizeMode : undefined"
      :row-group-mode="groupBy ? 'subheader' : undefined"
      :group-rows-by="groupBy ?? undefined"
      :sort-field="groupBy ?? undefined"
      :sort-order="groupBy ? 1 : undefined"
      :edit-mode="editMode === 'cell' ? 'cell' : undefined"
      :row-class="rowClass"
      :pt="primeDataTablePt"
      paginator-template="RowsPerPageDropdown FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink"
      removable-sort
      scrollable
      @cell-edit-complete="onCellEditComplete"
      @column-resize-end="onColumnResizeEnd"
      @column-reorder="onColumnReorder"
      @update:expanded-rows="updateExpandedRows"
    >
      <!-- Optional row expander. It is deliberately first so disclosure remains
           visually separate from destructive/action controls at either edge. -->
      <Column
        v-if="expandable"
        expander
        column-key="__expander"
        class="data-table__col-expander"
        :pt="{ root: { class: 'data-table__col-no-resize' } }"
        :reorderable-column="false"
        :show-filter-menu="false"
      />

      <!-- Action column(s) at the START (left) of the table — only when
           actionsPosition === 'start'. The filter cell of the first action column
           hosts the clear-all-filters button (or a consumer override slot). -->
      <Column
        v-for="(col, i) in actionsAtStart ? actionColumns : []"
        :key="`action-start-${col.field}`"
        :column-key="`__actions_${col.field}`"
        class="data-table__col-actions data-table__col-actions--start"
        :pt="col.resizable === false ? { root: { class: 'data-table__col-no-resize' } } : undefined"
        :reorderable-column="false"
        :style="columnStyle(col)"
        :show-filter-menu="false"
      >
        <template v-if="i === 0 && (hasSlot('actions-header') || heightToggle)" #header>
          <slot v-if="hasSlot('actions-header')" name="actions-header" />
          <div v-else class="data-table__header-end">
            <Button
              v-tooltip.top="heightExpanded ? t('compactHeight') : t('expandHeight')"
              data-testid="adt-toggle-height"
              type="button"
              text
              severity="secondary"
              class="data-table__button data-table__height-button"
              :aria-label="heightExpanded ? t('compactHeight') : t('expandHeight')"
              :aria-pressed="heightExpanded"
              @click="toggleTableHeight"
            >
              <template #icon="{ class: iconClass }">
                <i
                  v-if="heightExpanded"
                  :class="[
                    iconClass,
                    'data-table__height-icon',
                    'pi',
                    'pi-arrow-down-left-and-arrow-up-right-to-center',
                  ]"
                  data-icon="pi-arrow-down-left-and-arrow-up-right-to-center"
                  aria-hidden="true"
                />
                <i
                  v-else
                  :class="[iconClass, 'data-table__height-icon', 'pi', 'pi-expand']"
                  data-icon="pi-expand"
                  aria-hidden="true"
                />
              </template>
            </Button>
          </div>
        </template>
        <!-- Filter cell of the first action column. A consumer slot overrides the
             built-in clear-all-filters button when supplied. -->
        <template v-if="i === 0 && hasSlot('actions-filter')" #filter>
          <slot name="actions-filter" />
        </template>
        <template v-else-if="showClearButton && i === 0" #filter>
          <div
            :class="
              actionsPosition === 'start' ? 'data-table__filter-start' : 'data-table__filter-end'
            "
          >
            <Button
              v-tooltip.top="t('clearFilters')"
              type="button"
              icon="pi pi-filter-slash"
              text
              severity="secondary"
              class="data-table__button"
              :aria-label="t('clearFilters')"
              @click="clearAll"
            />
          </div>
        </template>
        <template #body="{ data: row, index }">
          <slot name="actions" :data="row" :index="index" />
        </template>
      </Column>

      <!-- Standalone clear-filters column at the START — only when actionsPosition
           === 'start' and there are no action columns (tables without actions but
           with filters whose clear button should align left). -->
      <Column
        v-if="showClearButton && actionColumns.length === 0 && actionsAtStart"
        column-key="__clear"
        class="data-table__col-actions data-table__col--utility"
        :pt="{ root: { class: 'data-table__col-no-resize' } }"
        :reorderable-column="false"
        :show-filter-menu="false"
      >
        <template #filter>
          <div class="data-table__filter-start">
            <Button
              v-tooltip.top="t('clearFilters')"
              type="button"
              icon="pi pi-filter-slash"
              text
              severity="secondary"
              class="data-table__button"
              :aria-label="t('clearFilters')"
              @click="clearAll"
            />
          </div>
        </template>
      </Column>

      <Column
        v-if="rowNumber"
        column-key="__rownum"
        :header="rowNumberHeader"
        class="data-table__col-rownum data-table__col--utility"
        :pt="{ root: { class: 'data-table__col-no-resize' } }"
        :reorderable-column="false"
      >
        <template #body="{ index }">{{ index + 1 }}</template>
      </Column>

      <Column
        v-for="col in dataColumns"
        :key="col.field"
        :column-key="col.field"
        :field="col.field"
        :sort-field="col.sortField"
        :sortable="col.sortable !== false && !groupBy"
        :body-class="isEditable(col) ? 'data-table__cell--editable' : undefined"
        :class="[
          col.editor === 'checkbox' ? 'data-table__col-checkbox' : undefined,
          col.columnClass,
        ]"
        :style="columnStyle(col)"
        :show-filter-menu="showMenuByField[col.field]"
        :pt="col.resizable === false ? { root: { class: 'data-table__col-no-resize' } } : undefined"
      >
        <!-- Header (with optional help tooltip — header tooltips only, #16) -->
        <template #header>
          <span v-tooltip.top="col.headerTooltip" class="data-table__header">{{ col.header }}</span>
        </template>

        <template #body="{ data: row, index }">
          <slot
            v-if="hasSlot(`body:${col.field}`)"
            :name="`body:${col.field}`"
            :data="row"
            :index="index"
          />
          <!-- checkbox editor (#9) — rendered in the body (always interactive),
               committing through the shared edit path. Kept out of the #editor
               block below because it's a single-click body control, not
               click-to-edit like the other editors. -->
          <div
            v-else-if="isEditable(col) && col.editor === 'checkbox'"
            :class="['data-table__checkbox-cell', { sensitive: col.sensitive }]"
          >
            <Checkbox
              :model-value="isTruthyBool(cellValue(row, col.field))"
              binary
              @update:model-value="
                (value: boolean) => commitEdit(row, col.field, value, cellValue(row, col.field))
              "
            />
          </div>
          <span v-else :class="bodyCellClass(col, row)">
            <!-- datetime (#14) — `DD.MM.YYYY | HH:mm:ss` parity. -->
            <template v-if="col.formatter === 'datetime'">
              {{ formatDateTime(dateValue(cellValue(row, col.field))) }}
            </template>
            <!-- date — `DD.MM.YYYY` parity. -->
            <template v-else-if="col.formatter === 'date'">
              {{ formatDate(dateValue(cellValue(row, col.field))) }}
            </template>
            <!-- rownum (#14) -->
            <template v-else-if="col.formatter === 'rownum'">
              {{ index + 1 }}
            </template>
            <!-- lookup id→label (#14) -->
            <template v-else-if="col.formatter === 'lookup'">
              {{ lookupLabel(col, cellValue(row, col.field)) }}
            </template>
            <!-- password mask (#14) — 10 asterisks matching legacy `**********` -->
            <template v-else-if="col.formatter === 'password'">**********</template>
            <template v-else>{{ cellValue(row, col.field) }}</template>
          </span>
        </template>

        <!-- Inline editor (#9). Checkbox is excluded — it renders in the body
             above as an always-interactive control, not via click-to-edit. A
             consumer may fully replace an editor via `#editor:<field>`. Select
             commits on item pick (cell-edit-complete). Input/number show a
             tick/cross button row so the user explicitly confirms or cancels. -->
        <template
          v-if="isEditable(col) && col.editor !== 'checkbox'"
          #editor="{ data: row, field }"
        >
          <div
            v-if="hasSlot(`editor:${col.field}`)"
            class="data-table__editor data-table__editor--custom"
            :data-testid="`adt-editor-${field}`"
            :class="{ 'data-table__editor--inplace': !commitOnBlur }"
          >
            <slot :name="`editor:${col.field}`" :data="row" :field="field" />
          </div>
          <Select
            v-else-if="col.editor === 'select'"
            :model-value="cellValue(row, field)"
            :options="col.editorOptions"
            option-label="label"
            option-value="value"
            :invalid="isInvalidEdit(row, field)"
            fluid
            class="data-table__editor"
            :data-testid="`adt-editor-${field}`"
            @update:model-value="(value: unknown) => setCellValue(row, field, value)"
          />
          <!-- Tick/cross editor. The wrapper fills the active TD absolutely via
               data-table__editor; overflow:visible + z-index on the TD let the
               button group overlap adjacent cells. @mousedown.prevent on each
               button stops the click from blurring the input before the handler
               runs; the handler itself blurs programmatically after acting.
               In commitOnBlur mode the tick/cross buttons are hidden and the
               editor commits automatically on focus leave. -->
          <div
            v-else
            class="data-table__editor"
            :class="{
              'data-table__editor--inplace': !commitOnBlur,
              'data-table__editor--invalid': isInvalidEdit(row, field),
            }"
          >
            <InputNumber
              v-if="col.editor === 'number'"
              :model-value="
                stagedEdit?.field === field
                  ? numberValue(stagedEdit.value)
                  : numberValue(cellValue(row, field))
              "
              :class="commitOnBlur ? undefined : 'data-table__inplace-input'"
              :invalid="isInvalidEdit(row, field)"
              @update:model-value="(v) => updateStagedEdit(row, field, v)"
              @input="
                (e: { value: string | number | undefined }) => updateStagedEdit(row, field, e.value)
              "
              @keydown.enter.stop.prevent="
                commitOnBlur ? closeInplaceEditor() : saveInplace(row, field)
              "
              @keydown.esc.stop.prevent="cancelInplace(row, field)"
            />
            <InputText
              v-else
              :data-testid="`adt-editor-${field}`"
              :model-value="
                stagedEdit?.field === field
                  ? String(stagedEdit.value ?? '')
                  : String(cellValue(row, field) ?? '')
              "
              :class="commitOnBlur ? undefined : 'data-table__inplace-input'"
              :type="col.editorInputType ?? 'text'"
              :invalid="isInvalidEdit(row, field)"
              @value-change="(v: string | undefined) => updateStagedEdit(row, field, v ?? '')"
              @keydown.enter.stop.prevent="
                commitOnBlur ? closeInplaceEditor() : saveInplace(row, field)
              "
              @keydown.esc.stop.prevent="cancelInplace(row, field)"
            />
            <template v-if="!commitOnBlur">
              <Button
                icon="pi pi-check"
                severity="secondary"
                size="small"
                :aria-label="t('confirmEdit')"
                class="data-table__inplace-btn data-table__inplace-btn--confirm"
                @mousedown.prevent
                @click="saveInplace(row, field)"
              />
              <Button
                icon="pi pi-times"
                severity="danger"
                size="small"
                :aria-label="t('cancelEdit')"
                class="data-table__inplace-btn data-table__inplace-btn--cancel"
                @mousedown.prevent
                @click="cancelInplace(row, field)"
              />
            </template>
          </div>
        </template>

        <!-- Header filter control (#5 / #5b / #5c / #5d) -->
        <template v-if="col.filterable" #filter="{ filterModel, filterCallback }">
          <slot
            v-if="hasSlot(`filter:${col.field}`)"
            :name="`filter:${col.field}`"
            :filter-model="filterModel"
            :filter-callback="filterCallback"
          />
          <!-- select / dropdown (#5b) -->
          <Select
            v-else-if="col.filterType === 'select'"
            v-model="filterModel.value"
            :data-testid="`adt-filter-${col.field}`"
            :options="col.filterOptions"
            option-label="label"
            option-value="value"
            show-clear
            class="data-table__filter"
            @change="filterCallback()"
          />
          <!-- tri-state boolean (#5c) — polyfill via <DataTableTriStateFilter>
               (PrimeVue's TriStateCheckbox was deprecated in v4). Centered in the
               filter cell to match the centered boolean body cells.
               update:modelValue fires after the internal cycle(), so filterCallback
               always sees the new value. -->
          <div
            v-else-if="col.filterType === 'tristate'"
            class="data-table__filter-center"
            :data-testid="`adt-filter-${col.field}`"
          >
            <DataTableTriStateFilter
              :model-value="triStateValue(filterModel.value)"
              :label="col.header"
              class="data-table__tristate"
              @update:model-value="
                (value: boolean | null) => {
                  filterModel.value = value
                  filterCallback()
                }
              "
            />
          </div>
          <!-- text / dateText (#5 / #5d) -->
          <InputText
            v-else
            v-model="filterModel.value"
            :data-testid="`adt-filter-${col.field}`"
            type="search"
            class="data-table__filter"
            :placeholder="col.filterPlaceholder ?? t('filterPlaceholder')"
            :aria-label="`filter-${col.field}`"
            @input="scheduleFilter(col.field, filterCallback)"
            @keydown.enter.prevent="applyFilter(col.field, filterCallback)"
          />
        </template>
      </Column>

      <!-- Action column(s) at the END (right) of the table — only when
           actionsPosition === 'end' (the default). No clear button here; it lives
           in the standalone control column at the start instead. -->
      <!-- Standalone clear-filters column at the END — only when actionsPosition
           !== 'start' and there are no action columns (e.g. Logs). -->
      <Column
        v-if="showClearButton && actionColumns.length === 0 && !actionsAtStart && !heightToggle"
        column-key="__clear"
        class="data-table__col-actions data-table__col--utility"
        :pt="{ root: { class: 'data-table__col-no-resize' } }"
        :reorderable-column="false"
        :show-filter-menu="false"
      >
        <template #filter>
          <div class="data-table__filter-end">
            <Button
              v-tooltip.top="t('clearFilters')"
              type="button"
              icon="pi pi-filter-slash"
              text
              severity="secondary"
              class="data-table__button"
              :aria-label="t('clearFilters')"
              @click="clearAll"
            />
          </div>
        </template>
      </Column>

      <!-- Tables without an actions column still receive the shared height control
           at the right edge. When actions are also at the end, this column carries
           the clear-filter control in its filter cell to avoid a duplicate column. -->
      <Column
        v-if="heightToggle && actionColumns.length === 0"
        column-key="__height"
        class="data-table__col-actions data-table__col--utility"
        :pt="{ root: { class: 'data-table__col-no-resize' } }"
        :reorderable-column="false"
        :show-filter-menu="false"
      >
        <template #header>
          <div class="data-table__header-end">
            <Button
              v-tooltip.top="heightExpanded ? t('compactHeight') : t('expandHeight')"
              data-testid="adt-toggle-height"
              type="button"
              text
              severity="secondary"
              class="data-table__button data-table__height-button"
              :aria-label="heightExpanded ? t('compactHeight') : t('expandHeight')"
              :aria-pressed="heightExpanded"
              @click="toggleTableHeight"
            >
              <template #icon="{ class: iconClass }">
                <i
                  v-if="heightExpanded"
                  :class="[
                    iconClass,
                    'data-table__height-icon',
                    'pi',
                    'pi-arrow-down-left-and-arrow-up-right-to-center',
                  ]"
                  data-icon="pi-arrow-down-left-and-arrow-up-right-to-center"
                  aria-hidden="true"
                />
                <i
                  v-else
                  :class="[iconClass, 'data-table__height-icon', 'pi', 'pi-expand']"
                  data-icon="pi-expand"
                  aria-hidden="true"
                />
              </template>
            </Button>
          </div>
        </template>
        <template v-if="showClearButton && !actionsAtStart" #filter>
          <div class="data-table__filter-end">
            <Button
              v-tooltip.top="t('clearFilters')"
              type="button"
              icon="pi pi-filter-slash"
              text
              severity="secondary"
              class="data-table__button"
              :aria-label="t('clearFilters')"
              @click="clearAll"
            />
          </div>
        </template>
      </Column>

      <Column
        v-for="(col, i) in actionsAtStart ? [] : actionColumns"
        :key="`action-end-${col.field}`"
        :column-key="`__actions_${col.field}`"
        class="data-table__col-actions"
        :pt="col.resizable === false ? { root: { class: 'data-table__col-no-resize' } } : undefined"
        :reorderable-column="false"
        :style="columnStyle(col)"
        :show-filter-menu="false"
      >
        <template v-if="i === 0 && (hasSlot('actions-header') || heightToggle)" #header>
          <slot v-if="hasSlot('actions-header')" name="actions-header" />
          <div v-else class="data-table__header-end">
            <Button
              v-tooltip.top="heightExpanded ? t('compactHeight') : t('expandHeight')"
              data-testid="adt-toggle-height"
              type="button"
              text
              severity="secondary"
              class="data-table__button data-table__height-button"
              :aria-label="heightExpanded ? t('compactHeight') : t('expandHeight')"
              :aria-pressed="heightExpanded"
              @click="toggleTableHeight"
            >
              <template #icon="{ class: iconClass }">
                <i
                  v-if="heightExpanded"
                  :class="[
                    iconClass,
                    'data-table__height-icon',
                    'pi',
                    'pi-arrow-down-left-and-arrow-up-right-to-center',
                  ]"
                  data-icon="pi-arrow-down-left-and-arrow-up-right-to-center"
                  aria-hidden="true"
                />
                <i
                  v-else
                  :class="[iconClass, 'data-table__height-icon', 'pi', 'pi-expand']"
                  data-icon="pi-expand"
                  aria-hidden="true"
                />
              </template>
            </Button>
          </div>
        </template>
        <!-- Filter cell of the first action column. A consumer slot overrides the
             built-in clear-all-filters button when supplied. -->
        <template v-if="i === 0 && hasSlot('actions-filter')" #filter>
          <slot name="actions-filter" />
        </template>
        <template v-else-if="showClearButton && i === 0" #filter>
          <div
            :class="
              actionsPosition === 'start' ? 'data-table__filter-start' : 'data-table__filter-end'
            "
          >
            <Button
              v-tooltip.top="t('clearFilters')"
              type="button"
              icon="pi pi-filter-slash"
              text
              severity="secondary"
              class="data-table__button"
              :aria-label="t('clearFilters')"
              @click="clearAll"
            />
          </div>
        </template>
        <template #body="{ data: row, index }">
          <slot name="actions" :data="row" :index="index" />
        </template>
      </Column>

      <!-- Group subheader (#1) — a DataTable-level slot in PrimeVue v4 -->
      <template v-if="groupBy" #groupheader="{ data: groupRow }">
        <slot name="groupheader" :data="groupRow">
          <span class="data-table__group-header">{{ cellValue(groupRow, groupBy) }}</span>
        </slot>
      </template>

      <template v-if="expandable" #expansion="{ data: expandedRow }">
        <div class="data-table__expansion-content">
          <slot name="expansion" :data="expandedRow" />
        </div>
      </template>

      <!-- Request activity is represented by the app-wide loading bar. Keep an
           initially empty table body quiet while loading, then show the empty
           state only after the first load has settled. -->
      <template #empty>
        <slot v-if="showEmptyState" name="empty">
          <div class="data-table__empty" data-testid="adt-empty">
            <i class="pi pi-inbox" aria-hidden="true" />
            <span>{{ resolvedEmptyMessage }}</span>
          </div>
        </slot>
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
/* Centralize the widths and colors shared by the fixed utility columns */
.data-table {
  --data-table-expander-column-width: 3rem;
  --data-table-utility-column-width: 3rem;
  --data-table-expander-column-background: var(--p-surface-50);
}
/* ── PrimeVue DataTable skin ────────────────────────────────────────────────
   The look shared by every table rendered through this wrapper (the admin grids
   and the scope-targets spreadsheet): an outer frame, a seamless filter header,
   no funnel buttons, ellipsised header titles, consistent icon-button colors. The
   grey header, cell padding and horizontal row lines come from the PrimeVue
   `datatable` preset (main.ts); these add what the preset doesn't. `:deep()` on
   the wrapper root reaches the inline PrimeVue table's `.p-datatable-*` DOM. */

/* Outer border + rounded corners around the whole table (Aura has none). The
   container-query context pairs with the empty-state rule at the end. */
.data-table :deep(.p-datatable-table-container) {
  border: 1px solid var(--p-content-border-color);
  border-radius: var(--p-border-radius-sm);
  /* Cap the scroll-body height from the `scrollHeight` prop: short tables shrink
     to fit their rows; taller ones scroll past this. `overflow: auto` so the cap
     actually scrolls (not clips). */
  max-height: var(--data-table-max-height);
  overflow: auto;
  container-type: inline-size;
}

.data-table--virtual-scroll :deep(.p-datatable-table-container) {
  max-height: none;
  overflow: visible !important;
}

.data-table--virtual-scroll :deep(.p-virtualscroller) {
  height: var(--data-table-max-height) !important;
  min-height: var(--data-table-max-height);
  max-height: var(--data-table-max-height);
  overflow: auto !important;
  overflow-anchor: none;
  overscroll-behavior: contain;
  border-radius: inherit;
}

/* fitColumns mode: columns with explicit widths keep them; columns without width
   share the remaining space. The initial layout has no scrollbar; one appears only
   if the user manually resizes columns past the container width. */
.data-table--fit-columns :deep(.p-datatable-table) {
  width: 100%;
  table-layout: fixed;
}

/* Merge the column-title row and the filter row into one seamless header block. */
.data-table :deep(.p-datatable-header-cell) {
  border-bottom: none;
}

.data-table :deep(.p-datatable-filter) {
  border-top: none;
}

/* Keep the funnel button narrow so the adjacent filter input gains space without
   shrinking the icon. PrimeVue marks it as rounded, so restore the normal button
   radius for its hover and active backgrounds. */
.data-table :deep(.p-datatable-column-filter-button) {
  width: 1.2rem;
  border-radius: var(--p-button-border-radius);
}

/* The per-column filter-menu (funnel) button is shown only where useful: the
   `show-filter-menu` binding (see showMenuByField) enables it for text-like columns
   and disables it for action, checkbox (tristate) and noop columns. No blanket
   CSS hide here — that would override that per-column decision. */

/* Active-filter funnel: tint teal when a non-default match mode is selected.
   PrimeVue gives the funnel no active class of its own, so the wrapper toggles
   `--active` on the button imperatively (see syncFunnelTint). */
.data-table :deep(.data-table__filter-btn--active) {
  color: var(--p-primary-500);
}

.data-table :deep(.p-datatable-column-title) {
  display: block;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Weight 700 matches DataGrid's thead th rule. Set on the header cell rather
   than the .p-datatable-column-title span because the Aura theme injects
   --p-datatable-column-title-font-weight: 600 on that span via its token system,
   which overrides a static :deep() rule. The <th> has no such override. */
.data-table :deep(.p-datatable-header-cell) {
  font-weight: 700;
}

/* Sort icon opacity matches SortableTh (DataGrid): dim at rest → more opaque on hover.
   Color is pinned via the preset (sortIcon.color/hoverColor); only opacity needs CSS
   since PrimeVue exposes no opacity token for sort icons. */
.data-table :deep(.p-datatable-sort-icon) {
  opacity: 0.5;
}

.data-table :deep(.p-datatable-header-cell:hover .p-datatable-sort-icon) {
  opacity: 0.8;
}

.data-table
  :deep(
    .p-datatable-header-cell.p-datatable-sortable-column.p-datatable-column-sorted
      .p-datatable-sort-icon
  ) {
  opacity: 1;
  color: inherit;
}

/* TEXT (borderless) icon buttons in the grid use the standard text color, muting
   on hover. Scoped to `.p-button-text` so SOLID action buttons keep their preset
   white-on-color icons. */
.data-table :deep(.data-table__button.p-button-text) {
  color: var(--p-text-color);
}

.data-table :deep(.data-table__button.p-button-text:hover) {
  color: var(--p-text-muted-color);
}

/* The height control keeps the muted, borderless text-button treatment. Its
   fill follows the header cell exactly. */
.data-table :deep(.data-table__height-button.p-button-text.p-button-secondary),
.data-table :deep(.data-table__height-button.p-button-text.p-button-secondary:hover),
.data-table :deep(.data-table__height-button.p-button-text.p-button-secondary:active) {
  border: 0;
  background: var(--p-datatable-header-cell-background, var(--p-surface-50));
}

.data-table :deep(.data-table__height-icon) {
  width: 1em;
  height: 1em;
  font-size: 1em;
  line-height: 1;
}

/* Active filter box: a text filter with content keeps the primary border so it's
   obvious which columns are being filtered even when the box isn't focused. */
.data-table :deep(.p-datatable-filter input[type='search']:not(:placeholder-shown)) {
  border-color: var(--p-primary-500);
}

/* Empty-state cell: the message must center within the VISIBLE width, not the
   full (often much wider) table width. The emptymessage <td> is colspanned, so a
   `width` on it is ignored by the table layout — instead we zero its padding and
   size + stick the inner `.data-table__empty` block to the visible width (see its
   rule below). The container-query context is `.p-datatable-table-container`. */
.data-table :deep(.p-datatable-emptymessage > td) {
  padding: 0;
}

.data-table--virtual-scroll :deep(.p-datatable-emptymessage) {
  height: var(--data-table-virtual-body-height);
}

.data-table--virtual-scroll :deep(.p-datatable-emptymessage > td) {
  height: var(--data-table-virtual-body-height);
  border-bottom: none;
  vertical-align: middle;
}

/* ── Wrapper-specific ───────────────────────────────────────────────────────── */

.data-table :deep(.data-table__cell--editable) {
  cursor: pointer;
}

.data-table :deep(.data-table__cell--invalid) {
  border-radius: var(--p-form-field-border-radius, 4px);
  box-shadow: inset 0 0 0 1px var(--p-form-field-invalid-border-color, var(--p-red-500));
}

.data-table :deep(.data-table__editor--invalid .p-inputtext),
.data-table :deep(.data-table__editor--invalid .p-inputnumber-input) {
  border-color: var(--p-form-field-invalid-border-color, var(--p-red-500));
}

/* ── Header filter controls ────────────────────────────────────────────────── */

/* Compact the filter row: tighter cell padding and smaller input height.
   The filter row is always the last <tr> in <thead> (PrimeVue has no dedicated
   class on it), so thead tr:last-child targets it without ambiguity. */
.data-table :deep(thead tr:last-child .p-datatable-header-cell) {
  padding-block: 0.25rem;
}

.data-table :deep(thead tr:last-child input) {
  padding-block: 0.25rem;
}

.data-table__filter {
  width: 100%;
}

.data-table__tristate {
  padding: 0.25rem;
}

/* Alignment wrappers for the clear-filters button and boolean/tristate filter
   controls. Must fill the full cell width so justify-content has room to push. */
.data-table__filter-center,
.data-table__filter-start,
.data-table__filter-end {
  display: flex;
  width: 100%;
}

.data-table__filter-center {
  justify-content: center;
}

.data-table__filter-start {
  justify-content: flex-start;
}

.data-table__filter-end {
  justify-content: flex-end;
}

.data-table__header-end {
  display: flex;
  width: 100%;
  justify-content: flex-end;
}

.data-table :deep(.data-table__col-actions .p-datatable-column-header-content) {
  width: 100%;
}

/* ── Inline cell editor ────────────────────────────────────────────────────── */

.data-table__editor {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.data-table__checkbox-cell {
  display: flex;
  justify-content: center;
}

/* Center the header title of checkbox columns too, so the header label sits over
   the centered body checkboxes rather than left-aligned above them. The class is
   set on the <Column> and lands on its header + body cells. */
.data-table :deep(.data-table__col-checkbox .data-table__header) {
  display: block;
  text-align: center;
}

/* Actions column shrinks to content width; exact width is set via the actions
   ColumnDef's `width` / `minWidth`. `white-space: nowrap` on the cell keeps the
   button row on a single line; `flex-wrap: nowrap` on the flex container below
   enforces the same at the button level. */
.data-table :deep(.data-table__col-actions),
.data-table :deep(.data-table__col-rownum) {
  width: 1%;
  white-space: nowrap;
  background-color: var(--p-surface-50);
}

/* Keep standalone controls and row numbers stable without template-level styles */
.data-table :deep(.data-table__col--utility) {
  width: var(--data-table-utility-column-width);
  min-width: var(--data-table-utility-column-width);
  max-width: var(--data-table-utility-column-width);
}

/* Fix every expander cell to one box width. Without matching min/max constraints,
   automatic table layout recalculates the first column when an expansion row is
   inserted and its continuation cell participates in column sizing. */
.data-table :deep(.data-table__col-expander) {
  box-sizing: border-box;
  width: var(--data-table-expander-column-width);
  min-width: var(--data-table-expander-column-width);
  max-width: var(--data-table-expander-column-width);
  white-space: nowrap;
  background-color: var(--data-table-expander-column-background) !important;
}

/* PrimeVue's expansion cell is split at mount into two real table cells: the
   actual expander-column continuation and the detail cell spanning all other
   columns. Real cells inherit the table's precise column geometry. */
.data-table :deep(.p-datatable-row-expansion > .data-table__expansion-leading-cell) {
  box-sizing: border-box;
  width: var(--data-table-expander-column-width);
  min-width: var(--data-table-expander-column-width);
  max-width: var(--data-table-expander-column-width);
  padding: 0;
  background: var(--data-table-expander-column-background);
}

.data-table--column-borders
  :deep(.p-datatable-row-expansion > .data-table__expansion-leading-cell) {
  border-right: 1px solid var(--p-content-border-color);
}

.data-table :deep(.p-datatable-row-expansion > .data-table__expansion-content-cell) {
  min-width: 0;
  padding: 0.5rem 0.75rem;
}

/* Actions column header cell (column-name row only): zero vertical padding so
   an icon button in the slot doesn't inflate the row beyond adjacent text cells.
   The filter row already gets its own padding via thead tr:last-child. */
.data-table :deep(thead tr:first-child th.data-table__col-actions) {
  padding-block: 0;
}

.data-table :deep(.data-table__col-rownum > .p-datatable-column-resizer),
.data-table :deep(.data-table__col-expander > .p-datatable-column-resizer),
.data-table :deep(.data-table__col-no-resize > .p-datatable-column-resizer) {
  display: none;
}

/* PrimeVue fit-mode handles resize their own and the following column. Hide a
   handle when its neighbor is fixed so the fixed column cannot move indirectly. */
.data-table
  :deep(.p-datatable-header-cell:has(+ .data-table__col-no-resize) > .p-datatable-column-resizer) {
  display: none;
}

.data-table :deep(.data-table__col-expander .p-datatable-row-toggle-button) {
  display: inline-flex;
  flex: 0 0 1.8rem;
  align-items: center;
  justify-content: center;
  width: 1.8rem;
  height: 1.8rem;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: var(--p-text-muted-color);
  --p-icon-size: 0.7rem;
}

.data-table :deep(.data-table__col-expander .p-datatable-row-toggle-button:enabled:hover),
.data-table
  :deep(
    .p-datatable-tbody
      > tr.p-datatable-row-selected
      .data-table__col-expander
      .p-datatable-row-toggle-button:hover
  ) {
  background: transparent;
  color: var(--p-text-color);
}

/* Mini square icon buttons in action cells. Must fill the full cell width so
   justify-content can push the buttons right (at end) or left (at start). */
.data-table :deep(.data-table__actions) {
  display: flex;
  width: 100%;
  flex-wrap: nowrap;
  gap: 2px;
  align-items: center;
  justify-content: flex-end;
}

.data-table :deep(.data-table__col-actions--start .data-table__actions) {
  justify-content: flex-start;
}

.data-table :deep(.p-datatable-tbody > tr > td.data-table__col-actions) {
  padding-block: 0.25rem;
  vertical-align: middle;
}

.data-table :deep(.data-table__button) {
  width: 1.8rem;
  height: 1.8rem;
  padding: 0;
  margin: 0 0.15em 0 0;
  font-size: 0.75rem;
}

/* Anti-reflow: pin every body cell to a fixed height so a PrimeVue edit-mode
   cell swap (span → InputText) doesn't cause rows below to jump. overflow:hidden
   clips any editor that PrimeVue sizes larger than the column — prevents the
   cell from bleeding into its neighbour. All tables using editMode="cell" are
   affected (admin tables + TargetsEditor). The SSS-only expansion-row exclusion
   stays inside :where() so it does not outrank the LSD active-editor override. */
.data-table
  :deep(
    .p-datatable-tbody
      > tr:not(.p-datatable-emptymessage):not(.p-datatable-empty-message):not(
        :where(.p-datatable-row-expansion)
      ):not(:has(.data-table__empty))
      > td
  ) {
  height: var(--app-data-table-row-height);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Under table-layout:fixed, max-width:0 on td surrenders intrinsic content width
   to the column's assigned th width, which makes text-overflow:ellipsis fire at
   the padding boundary rather than the border edge. Not applied in auto-layout
   (non-fitColumns) tables where it would interfere with column resizing. */
.data-table--fit-columns
  :deep(.p-datatable-tbody > tr:not(:where(.p-datatable-row-expansion)) > td) {
  max-width: 0;
}

/* Inline cell editors: all editor types share these rules. PrimeVue marks the
   active TD with data-p-cell-editing="true" (a DOM data attribute — no CSS class
   equivalent exists). Making it position:relative turns it into the containing
   block for the absolutely-positioned editor below. */
.data-table :deep([data-p-cell-editing='true']) {
  position: relative;
}

/* Editors never participate in table layout. Their mode-specific rules below
   either fit the existing cell or allow the control to overlap on the right. */
.data-table :deep([data-p-cell-editing='true'] .data-table__editor) {
  box-sizing: border-box;
  max-width: none;
}

/* InputText: slim vertical padding to fit inside the pinned row height. */
.data-table :deep(.data-table__editor.p-inputtext) {
  padding-block: 0.15rem;
}

.data-table :deep([data-p-cell-editing='true'] .data-table__editor .p-inputnumber-input) {
  width: 100%;
  padding-block: 0.15rem;
}

/* Editors without accept/reject buttons fit the unchanged cell content box.
   Selects follow the same geometry and commit when an option is picked. */
.data-table--commit-on-blur :deep([data-p-cell-editing='true'] .data-table__editor),
.data-table :deep([data-p-cell-editing='true'] .data-table__editor.p-select) {
  position: absolute;
  display: flex;
  align-items: center;
  inset-block-start: 50%;
  inset-inline: 0.75rem;
  transform: translateY(-50%);
  height: fit-content;
  width: auto;
  min-width: 0;
}

.data-table--commit-on-blur :deep([data-p-cell-editing='true'] .data-table__editor > .p-inputtext),
.data-table--commit-on-blur
  :deep([data-p-cell-editing='true'] .data-table__editor > .p-inputnumber) {
  flex: 1 1 0;
  min-width: 0;
  height: auto;
  padding-block: 0.15rem;
}

/* Tick/cross editors have one shared minimum width and may extend over cells to
   the right without changing the active TD's padding, height, or width. */
.data-table:not(.data-table--commit-on-blur)
  :deep(
    .p-datatable-tbody > tr > td[data-p-cell-editing='true']:has(.data-table__editor--inplace)
  ) {
  overflow: visible;
  z-index: 1;
}

/* Blur-commit editors: z-index keeps the editor on top of adjacent cells while
   the column width stays locked (editor is position:absolute). The long selector
   beats the (0,2,2) specificity of the overflow:hidden rule on the same TD — we
   keep overflow:hidden here so the input clips at the column edge instead of
   overflowing into the next column. */
.data-table--commit-on-blur :deep(.p-datatable-tbody > tr > td[data-p-cell-editing='true']) {
  z-index: 1;
}

.data-table :deep(.data-table__editor--inplace) {
  position: absolute;
  display: flex;
  align-items: stretch;
  inset-block-start: 50%;
  inset-inline-start: 0.75rem;
  transform: translateY(-50%);
  width: max(calc(100% - 1.5rem), 12rem);
  min-width: 12rem;
  height: 1.8rem;
  min-height: 0;
}

.data-table :deep(.data-table__inplace-input) {
  flex: 1 1 0;
  height: 1.8rem;
  min-height: 0;
  min-width: 0;
  box-sizing: border-box;
  padding-block: 0.1rem;
}

.data-table :deep(.data-table__inplace-input.p-inputnumber) {
  height: 1.8rem;
  min-height: 0;
  padding-block: 0;
}

.data-table :deep(.data-table__inplace-input .p-inputnumber-input) {
  width: 100%;
  height: 1.8rem;
  min-height: 0;
  box-sizing: border-box;
  padding-block: 0.1rem;
}

.data-table :deep(.data-table__inplace-input.p-inputtext),
.data-table :deep(.data-table__inplace-input .p-inputnumber-input) {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.data-table :deep(.data-table__inplace-btn) {
  flex: 0 0 1.8rem;
  width: 1.8rem;
  min-width: 1.8rem;
  height: 1.8rem;
  min-height: 0;
  padding: 0;
  font-size: 0.75rem;
}

.data-table :deep(.data-table__inplace-btn--confirm) {
  border-radius: 0;
  border-left: none;
  border-right: none;
  background-color: var(--p-teal-500);
  border-color: var(--p-teal-500);
  color: var(--p-surface-0);
}

.data-table :deep(.data-table__inplace-btn--confirm:hover) {
  background-color: var(--p-teal-600);
  border-color: var(--p-teal-600);
}

.data-table :deep(.data-table__inplace-btn--cancel) {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  border-left: none;
}

/* Select: zero the built-in min-width token so the component does not push the
   absolutely-positioned element wider than its inset:0 bounds. The focus outline
   is suppressed because it renders outside the box model and cannot be clipped —
   PrimeVue's border-color and box-shadow already signal focus. */
.data-table :deep(.data-table__editor.p-select) {
  --p-select-min-width: 0;
  height: 1.8rem;
  min-height: 0;
  min-width: 0;
  align-items: center;
}

.data-table :deep(.data-table__editor.p-select.p-focus) {
  outline: none;
}

.data-table :deep(.data-table__editor.p-select .p-select-label) {
  padding-block: 0.1rem;
}

.data-table :deep(.data-table__editor.p-select .p-select-dropdown) {
  align-self: stretch;
  width: 1.8rem;
}

/* ── Optional column borders ───────────────────────────────────────────────── */

/* Vertical borders between columns: appear on the filter-row header cells and
   body cells only. The column-title row (with column names) intentionally gets
   no border — borders start at the filter row and extend down through the body.
   Both header rows use .p-datatable-header-cell on their <th>; :has() narrows
   to only the filter row (whose <th> contains .p-datatable-filter). Last child
   in each row has its border cleared so no doubled edge sits against the outer
   frame. */
.data-table--column-borders :deep(.p-datatable-header-cell:has(.p-datatable-filter)) {
  border-right: 1px solid var(--p-content-border-color);
}

.data-table--column-borders :deep(.p-datatable-header-cell:has(.p-datatable-filter):last-child) {
  border-right: none;
}

.data-table--column-borders :deep(.p-datatable-tbody > tr > td) {
  border-right: 1px solid var(--p-content-border-color);
}

.data-table--column-borders :deep(.p-datatable-tbody > tr > td:last-child) {
  border-right: none;
}

/* ── Group subheader + empty state ─────────────────────────────────────────── */

.data-table__group-header {
  font-weight: 600;
}

/* Centered inbox icon + muted italic message, matching the shared LayoutLoading /
   scope-targets empty state so every table's "no rows" look is identical.
   Sized + stuck to the VISIBLE scroll width (100cqw against the table container's
   inline-size container query) and pinned left, so the message stays centered in
   view even when the table is wider than the viewport and scrolled horizontally. */
.data-table__empty {
  position: sticky;
  left: 0;
  width: 100cqw;
  min-height: 4.5rem;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem;
  color: var(--p-text-muted-color);
  font-style: italic;
}

.data-table--virtual-scroll :deep(.p-virtualscroller:has(.p-datatable-emptymessage)),
.data-table--virtual-scroll :deep(.p-virtualscroller:has(.p-datatable-empty-message)),
.data-table--virtual-scroll :deep(.p-virtualscroller:has(.data-table__empty)) {
  overflow-y: hidden !important;
}

.data-table--virtual-scroll :deep(.p-datatable-tbody > tr.p-datatable-emptymessage),
.data-table--virtual-scroll :deep(.p-datatable-tbody > tr.p-datatable-empty-message),
.data-table--virtual-scroll :deep(.p-datatable-tbody > tr:has(.data-table__empty)) {
  height: calc(var(--data-table-virtual-body-height) - 1px);
}

.data-table--virtual-scroll :deep(.p-datatable-tbody > tr.p-datatable-emptymessage > td),
.data-table--virtual-scroll :deep(.p-datatable-tbody > tr.p-datatable-empty-message > td),
.data-table--virtual-scroll :deep(.p-datatable-tbody > tr:has(.data-table__empty) > td) {
  height: calc(var(--data-table-virtual-body-height) - 1px);
  border-bottom: none;
  overflow: hidden;
  vertical-align: middle;
}

.data-table--virtual-scroll .data-table__empty {
  min-height: 0;
  height: calc(var(--data-table-virtual-body-height) - 1px);
  padding-block: 0;
  overflow: hidden;
}
</style>
