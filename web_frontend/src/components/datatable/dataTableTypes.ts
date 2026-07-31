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

/** Filter controls supported by the shared data table. */
type FilterType = 'text' | 'select' | 'tristate' | 'dateText' | 'noop'

/** Inline editors supported by the shared data table. */
type EditorType = 'input' | 'number' | 'select' | 'checkbox'

/** Built-in cell renderers supported by the shared data table. */
type FormatterType = 'text' | 'datetime' | 'date' | 'rownum' | 'lookup' | 'password' | 'actions'

/** Option for `select` filters and editors. */
interface SelectOption {
  label: string
  value: unknown
}

/** DataTableSlotProps describes the row context exposed to body and action slots. */
interface DataTableSlotProps<T> {
  data: T
  index: number
}

/** DataTableFieldSlotProps describes the row context exposed to editor slots. */
interface DataTableFieldSlotProps<T> {
  data: T
  field: string
}

/** DataTableSlots declares the shared table's static and field-specific scoped slots. */
export type DataTableSlots<T> = {
  actions?: (props: DataTableSlotProps<T>) => unknown
  'actions-header'?: () => unknown
  'actions-filter'?: () => unknown
  empty?: () => unknown
  expansion?: (props: { data: T }) => unknown
  groupheader?: (props: { data: T }) => unknown
} & {
  [K in `body:${string}`]?: (props: DataTableSlotProps<T>) => unknown
} & {
  [K in `editor:${string}`]?: (props: DataTableFieldSlotProps<T>) => unknown
} & {
  [K in `filter:${string}`]?: (props: {
    filterModel: { value: unknown }
    filterCallback: () => void
  }) => unknown
} & Record<string, unknown>

/** Validate an inline edit before it is committed. */
type CellValidator<T> = (value: unknown, row: T) => true | string

/** Column definition consumed by `<DataTable>`. */
export interface ColumnDef<T = Record<string, unknown>> {
  /** Row data key this column reads/writes. */
  field: string
  /** Column header label. Callers pass an already-translated string. */
  header: string
  /** Enable header sort. Default `true` (false for action/password cols). */
  sortable?: boolean
  /** Optional API-backed value resolver used for sorting a complex field. */
  sortField?: string | ((row: T) => string)
  /** Render a header filter control. Default `false`. */
  filterable?: boolean
  /** Which filter control to render. Default `'text'`. */
  filterType?: FilterType
  /**
   * Override the PrimeVue match mode for this column's filter. Rarely needed —
   * defaults derive from `filterType`. Used e.g. for a numeric flag column
   * filtered by a tri-state control, where the boolean-EQUALS default won't
   * match the stored number.
   */
  matchMode?: string
  /** Options for a `select` filter. */
  filterOptions?: SelectOption[]
  /** Placeholder for a text filter box. Defaults to the component's built-in filter placeholder. */
  filterPlaceholder?: string
  /** Show the match-mode funnel button for this column. Defaults to the table-level `filterFunnel` prop. Set `false` to suppress the funnel on a specific column. */
  filterFunnel?: boolean
  /** Inline editor to render when the table `editMode` is `'cell'`. */
  editor?: EditorType
  /** HTML input type forwarded to the `input` editor's `<InputText>` (e.g. `'password'`). */
  editorInputType?: string
  /** Options for a `select` editor. */
  editorOptions?: SelectOption[]
  /** Validate an inline edit before commit. */
  validator?: CellValidator<T>
  /** Built-in cell renderer. Default `'text'`. */
  formatter?: FormatterType
  /** id→label map for the `lookup` formatter. */
  lookupMap?: Record<string | number, string>
  /** Explicit column width, e.g. `'12rem'`. */
  width?: string
  /** Minimum column width. In fit mode this becomes the fixed-layout allocation basis. */
  minWidth?: string
  /** Maximum column width, e.g. `'20rem'`. */
  maxWidth?: string
  /** Per-row body cell class. */
  bodyClass?: (row: T) => string
  /** Static class applied to every cell in this column (header + body), e.g. to
   *  style a whole column (dashed border, muted read-only look). */
  columnClass?: string
  /** Header help tooltip text. */
  headerTooltip?: string
  /**
   * Allow the user to resize this column. `false` keeps it fixed by disabling
   * both its own handle and a preceding fit-mode handle that would resize it.
   * Default `true` (inherits from the table).
   */
  resizable?: boolean
  /** Blur this column's cell content in presentation mode (adds the `sensitive` CSS class). */
  sensitive?: boolean
}

/** CellEditPayload describes the value change emitted and persisted by DataTable. */
export interface CellEditPayload<T> {
  row: T
  field: string
  /** The new (committed) value. */
  newValue: unknown
  /** The value before the edit began. */
  oldValue: unknown
}

/** InlineEditToast describes a toast emitted by a successful or failed inline edit. */
interface InlineEditToast {
  summary: string
  detail?: string
  /** Display duration in milliseconds. */
  life?: number
}

/** InlineEditToastInput allows an edit callback to suppress or configure its toast. */
export type InlineEditToastInput = string | InlineEditToast | null

/** InlineEditOptions configures validation, persistence, feedback, and rollback for inline edits. */
export interface InlineEditOptions<T> {
  /** Validate the pending edit and return a translated error message when it must be rejected. */
  validator?: (edit: CellEditPayload<T>) => string | null
  /** Persist the edit and optionally return the success toast to display. */
  save: (edit: CellEditPayload<T>) => Promise<InlineEditToastInput>
  /** Build the error toast after persistence fails. Omit to use the table's built-in error message. */
  errorToast?: (error: unknown, edit: CellEditPayload<T>) => InlineEditToastInput
  /** Display duration for validation warnings. Defaults to 4000 milliseconds. */
  validationLife?: number
  /** Display duration used when a returned toast does not provide one. Defaults to 3000 milliseconds. */
  defaultLife?: number
}
