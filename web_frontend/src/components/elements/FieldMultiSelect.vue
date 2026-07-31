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
 * Shared chip-style multi-value field, aligned with the Large-Scale Discovery
 * grant-users control. PrimeVue AutoComplete supplies the removable chips and
 * can offer known values, accept free entries, validate both entries and the
 * complete list, and turn pasted delimited text into chips.
 */
import AutoComplete from 'primevue/autocomplete'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import Field from '@/components/elements/Field.vue'
import TagLabel from '@/components/elements/TagLabel.vue'

// Keep field-owned action wording next to the component that renders it
const localMessages = {
  'en-US': {
    removeValue: 'Remove {value}',
  },
}

// ── Types ─────────────────────────────────────────────────────
type TagVariant = 'info' | 'primary' | 'success' | 'warn' | 'orange' | 'error'

interface AutoCompleteInstance {
  $el: HTMLElement
  overlay?: HTMLElement | null
  overlayVisible: boolean
  focusedOptionIndex: number
  focusedMultipleOptionIndex: number
  visibleOptions: unknown[]
  show: (focusInput?: boolean) => void
}

// ── Props and models ──────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    label: string
    help?: string
    tooltip?: boolean
    required?: boolean
    options?: string[]
    seed?: string[]
    allowNew?: boolean
    typeahead?: boolean
    completeOnFocus?: boolean
    placeholder?: string
    /** Stable test hook forwarded to the editable text input. */
    testId?: string
    /** Maximum height of the wrapping chip/input area before it scrolls. */
    maxHeight?: string
    disabled?: boolean
    entryValidator?: (value: string) => string | undefined
    valuesValidator?: (values: readonly string[]) => string | undefined
    /** Color intent or value-specific color resolver for chips and valid suggestions. */
    tagVariant?: TagVariant | ((value: string) => TagVariant)
    /** Delimiters used when pasted text is converted to chips. */
    pasteSeparatorPattern?: RegExp
    validate?: () => void
  }>(),
  {
    help: undefined,
    tooltip: false,
    required: false,
    options: () => [],
    seed: () => [],
    allowNew: true,
    typeahead: true,
    completeOnFocus: false,
    placeholder: undefined,
    testId: undefined,
    maxHeight: '10rem',
    disabled: false,
    entryValidator: undefined,
    valuesValidator: undefined,
    tagVariant: 'warn',
    pasteSeparatorPattern: () => /[\s,;]+/u,
    validate: undefined,
  },
)

const model = defineModel<string[]>({ required: true })
const error = defineModel<string | undefined>('error', { default: undefined })
const slots = defineSlots<{
  suffix?: () => unknown
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})

// ── Reactive state ────────────────────────────────────────────
const filteredSuggestions = ref<string[]>([])
const hasRejectedDraft = ref(false)
const autocomplete = ref<AutoCompleteInstance | null>(null)
let reopenTimer: ReturnType<typeof setTimeout> | undefined
let shouldSkipNextSuggestionRefresh = false

// ── Computed ──────────────────────────────────────────────────
const candidateOptions = computed<string[]>(() => {
  const seen = new Set<string>()
  const options: string[] = []
  for (const option of [...props.seed, ...props.options]) {
    if (!seen.has(option)) {
      seen.add(option)
      options.push(option)
    }
  }
  return options
})

const hasSuggestions = computed(() => candidateOptions.value.length > 0)
const effectiveTypeahead = computed(() => hasSuggestions.value || props.allowNew || props.typeahead)
const effectiveCompleteOnFocus = computed(() => hasSuggestions.value || props.completeOnFocus)

const validatedModel = computed<string[]>({
  get: () => model.value,
  set: updateValues,
})

// ── Functions ─────────────────────────────────────────────────
function resolveTagVariant(value: unknown): TagVariant {
  const tagVariant = props.tagVariant
  return typeof tagVariant === 'function' ? tagVariant(String(value)) : tagVariant
}

function normalizeValues(values: readonly unknown[]): string[] {
  return [
    ...new Set(
      values
        .filter((value): value is string => typeof value === 'string')
        .map((value) => value.trim())
        .filter(Boolean),
    ),
  ]
}

function validationError(values: readonly string[]): string | undefined {
  for (const value of values) {
    const entryError = props.entryValidator?.(value)
    if (entryError) return entryError
  }
  return props.valuesValidator?.(values)
}

function updateValues(values: readonly string[]): void {
  const normalized = normalizeValues(values)
  const addedValues = normalized.filter((value) => !model.value.includes(value))
  const nextError = validationError(normalized)
  if (addedValues.length > 0 && nextError) {
    error.value = nextError
    return
  }
  model.value = normalized
  error.value = nextError
}

function suggestionsFor(query: string): string[] {
  const queryText = query.trim()
  const normalizedQuery = queryText.toLocaleLowerCase()
  const matches = candidateOptions.value.filter(
    (candidate) =>
      candidate.toLocaleLowerCase().includes(normalizedQuery) && !model.value.includes(candidate),
  )

  const queryMatchesKnownValue = matches.some(
    (candidate) => candidate.toLocaleLowerCase() === normalizedQuery,
  )
  const queryAlreadySelected = model.value.some(
    (value) => value.toLocaleLowerCase() === normalizedQuery,
  )

  return props.allowNew && queryText && !queryMatchesKnownValue && !queryAlreadySelected
    ? [...matches, queryText]
    : matches
}

function autocompleteInput(): HTMLInputElement | null {
  return autocomplete.value?.$el.querySelector<HTMLInputElement>('input') ?? null
}

function repairStaleOverlayState(): void {
  const instance = autocomplete.value
  if (!instance?.overlayVisible || instance.overlay) return

  // PrimeVue may keep overlayVisible=true after its overlay element was removed
  instance.overlayVisible = false
}

function showSuggestionsIfClosed(focusInput?: boolean): void {
  repairStaleOverlayState()

  const instance = autocomplete.value

  if (!instance || instance.overlayVisible) return
  instance.show(focusInput)
}

function updateSuggestionsForQuery(query: string): void {
  repairStaleOverlayState()
  filteredSuggestions.value = suggestionsFor(query)
}

function updateFilteredSuggestions(input: HTMLInputElement | null): void {
  updateSuggestionsForQuery(input?.value ?? '')
}

function refreshSuggestionsForModel(): void {
  // Keep filtered options and the focused overlay aligned with the selected values
  const input = autocompleteInput()
  updateFilteredSuggestions(input)
  if (input && document.activeElement === input && filteredSuggestions.value.length > 0) {
    showSuggestionsIfClosed()
  }
}

function completeSuggestions(event: { query: string }): void {
  updateSuggestionsForQuery(event.query)
}

function prepareForChipRemoval(): void {
  const instance = autocomplete.value
  if (instance) instance.overlayVisible = false

  // Reintroducing a removed value as a suggestion during the same PrimeVue update
  // can make AutoComplete align a destroyed overlay. Let the removal settle first.
  shouldSkipNextSuggestionRefresh = true
}

function onInputFocus(event: FocusEvent): void {
  const input = event.target
  if (!(input instanceof HTMLInputElement) || !input.value.trim()) return

  updateFilteredSuggestions(input)
  if (filteredSuggestions.value.length === 0) return

  void nextTick(() => {
    showSuggestionsIfClosed()
  })
}

function keepSuggestionsOpen(input: HTMLInputElement | null = autocompleteInput()): void {
  void nextTick(() => {
    if (reopenTimer !== undefined) clearTimeout(reopenTimer)

    // PrimeVue schedules its own hide after emitting option-select. Queueing
    // from nextTick places this timer after that hide so the refreshed list is
    // the final overlay state.
    reopenTimer = setTimeout(() => {
      reopenTimer = undefined
      if (!input || document.activeElement !== input) return

      updateFilteredSuggestions(input)
      if (filteredSuggestions.value.length > 0) showSuggestionsIfClosed()
    }, 0)
  })
}

function splitInputValues(value: string): string[] {
  return normalizeValues(value.split(props.pasteSeparatorPattern))
}

function commitValues(input: HTMLInputElement, values: readonly string[]): boolean {
  const normalizedValues = normalizeValues(values)
  if (normalizedValues.length === 0) return false

  const nextValues = normalizeValues([...model.value, ...normalizedValues])
  const nextError = validationError(nextValues)
  if (nextError) {
    hasRejectedDraft.value = true
    error.value = nextError
    return false
  }

  model.value = nextValues
  hasRejectedDraft.value = false
  error.value = undefined
  input.value = ''
  updateFilteredSuggestions(input)
  keepSuggestionsOpen(input)
  return true
}

function commitValue(input: HTMLInputElement, value: string): boolean {
  return commitValues(input, [value])
}

function commitInputValue(input: HTMLInputElement): boolean {
  return commitValues(input, splitInputValues(input.value))
}

function isInputSeparator(key: string): boolean {
  if (key.length !== 1) return false
  const separatorPattern = new RegExp(
    props.pasteSeparatorPattern.source,
    props.pasteSeparatorPattern.flags,
  )
  return separatorPattern.test(key)
}

function focusedSuggestion(): string | undefined {
  const instance = autocomplete.value
  if (!instance?.overlayVisible || instance.focusedOptionIndex < 0) return undefined

  const suggestion = instance.visibleOptions[instance.focusedOptionIndex]
  return typeof suggestion === 'string' ? suggestion : undefined
}

function shouldPrepareBackspaceRemoval(event: KeyboardEvent): boolean {
  if (event.key !== 'Backspace' || event.isComposing || model.value.length === 0) return false

  const target = event.target
  if (target instanceof HTMLInputElement) return !target.value

  // PrimeVue also removes a focused chip from the multi-chip container itself
  return (autocomplete.value?.focusedMultipleOptionIndex ?? -1) >= 0
}

function prepareBackspaceRemoval(event: KeyboardEvent): void {
  if (shouldPrepareBackspaceRemoval(event)) prepareForChipRemoval()
}

function onCommitKeydown(event: KeyboardEvent): void {
  prepareBackspaceRemoval(event)

  const isEnterKey = event.key === 'Enter' || event.key === 'NumpadEnter'
  const isSelectionKey = event.key === ' ' || event.key === 'Tab' || isEnterKey
  const isFreeEntrySeparator = props.allowNew && isInputSeparator(event.key)

  if (
    event.isComposing ||
    event.ctrlKey ||
    event.altKey ||
    event.metaKey ||
    (!isSelectionKey && !isFreeEntrySeparator)
  ) {
    return
  }

  const input = event.target
  if (!(input instanceof HTMLInputElement)) return

  const suggestion = focusedSuggestion() ?? suggestionsFor(input.value)[0]
  const canAddInput = props.allowNew && Boolean(input.value.trim())

  if (!suggestion && !canAddInput) {
    const pendingValue = input.value.trim()
    const nextError = pendingValue ? props.entryValidator?.(pendingValue) : undefined
    if (nextError) {
      event.preventDefault()
      event.stopPropagation()
      hasRejectedDraft.value = true
      error.value = nextError
      return
    }

    // Keep selection keys away from PrimeVue's potentially stale option index.
    // Tab retains its native focus movement while Enter keeps feedback visible.
    if (event.key === 'Tab') event.stopPropagation()
    if (isEnterKey) {
      event.preventDefault()
      event.stopPropagation()
    }
    return
  }

  // Resolve before PrimeVue's handler so all selection keys honor its highlighted option
  event.preventDefault()
  event.stopPropagation()
  if (suggestion && candidateOptions.value.includes(suggestion)) {
    commitValue(input, suggestion)
    return
  }
  commitInputValue(input)
}

function containEnterKey(event: KeyboardEvent): void {
  if (event.key !== 'Enter' && event.key !== 'NumpadEnter') return

  // This bubble-phase handler runs after PrimeVue's input handler. Keep Enter
  // entirely inside the multiselect so form-level key handlers cannot submit.
  if (!event.isComposing) event.preventDefault()
  event.stopPropagation()
}

function validatePendingInput(event: Event): void {
  if (!hasRejectedDraft.value) return

  const input = event.target
  if (!(input instanceof HTMLInputElement)) return

  const pendingValues = splitInputValues(input.value)
  if (pendingValues.length === 0) hasRejectedDraft.value = false
  error.value = validationError(normalizeValues([...model.value, ...pendingValues]))
}

function onInput(event: Event): void {
  validatePendingInput(event)

  const input = event.target
  if (!(input instanceof HTMLInputElement) || input.value.length > 0) return

  // PrimeVue does not complete an empty query because its default minimum length is one
  keepSuggestionsOpen(input)
}

function suggestionValidationError(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined
  return validationError(normalizeValues([...model.value, value]))
}

function isSuggestionInvalid(value: unknown): boolean {
  return Boolean(suggestionValidationError(value))
}

function guardInvalidSuggestionClick(event: MouseEvent, value: unknown): void {
  const nextError = suggestionValidationError(value)
  if (!nextError) return

  event.preventDefault()
  event.stopImmediatePropagation()
  hasRejectedDraft.value = true
  error.value = nextError
}

function onOptionSelect(): void {
  const input = autocompleteInput()
  if (input) input.value = ''
  hasRejectedDraft.value = false
  keepSuggestionsOpen(input)
}

function removeChip(removeCallback: (event: Event) => void, event: MouseEvent): void {
  prepareForChipRemoval()
  removeCallback(event)
}

function onPaste(event: ClipboardEvent): void {
  const pastedText = event.clipboardData?.getData('text') ?? ''
  if (!pastedText) return

  const pastedValues = splitInputValues(pastedText)
  if (pastedValues.length === 0) return

  event.preventDefault()
  if (props.allowNew) {
    const input = event.target
    if (!(input instanceof HTMLInputElement)) return

    input.value = pastedText
    commitValues(input, pastedValues)
    return
  }

  const candidatesByLowerCase = new Map(
    candidateOptions.value.map((candidate) => [candidate.toLocaleLowerCase(), candidate]),
  )
  const knownValues = pastedValues
    .map((value) => candidatesByLowerCase.get(value.toLocaleLowerCase()))
    .filter((value): value is string => value !== undefined)
  updateValues([...model.value, ...knownValues])
}

function validateCurrent(): void {
  const pendingValues = splitInputValues(autocompleteInput()?.value ?? '')
  if (pendingValues.length > 0) {
    if (hasRejectedDraft.value) {
      error.value = validationError(normalizeValues([...model.value, ...pendingValues]))
    }
    return
  }

  hasRejectedDraft.value = false
  props.validate?.()
}

function commitPendingInput(): boolean {
  const input = autocompleteInput()
  if (!input || !input.value.trim()) return true
  return commitInputValue(input)
}

defineExpose({ commitPendingInput })

// ── Watchers and lifecycle ────────────────────────────────────
watch(
  model,
  () => {
    if (shouldSkipNextSuggestionRefresh) {
      shouldSkipNextSuggestionRefresh = false

      // Wait until PrimeVue removes the chip before reintroducing its value as an option
      void nextTick(refreshSuggestionsForModel)
      return
    }

    // Apply ordinary model changes within the current post-render watcher cycle
    refreshSuggestionsForModel()
  },
  { deep: true, flush: 'post' },
)

onBeforeUnmount(() => {
  if (reopenTimer !== undefined) clearTimeout(reopenTimer)
})
</script>

<template>
  <Field
    :label="label"
    :help="help"
    :tooltip="tooltip"
    :required="required"
    :error="error"
    :validate="validateCurrent"
  >
    <template #default="{ id }">
      <div
        :class="[
          'field-multi-select__wrapper',
          { 'field-multi-select__wrapper--with-suffix': slots.suffix },
        ]"
      >
        <AutoComplete
          ref="autocomplete"
          v-model="validatedModel"
          :input-id="id"
          multiple
          :typeahead="effectiveTypeahead"
          :complete-on-focus="effectiveCompleteOnFocus"
          auto-option-focus
          :show-empty-message="!allowNew"
          :suggestions="filteredSuggestions"
          :option-disabled="isSuggestionInvalid"
          :placeholder="placeholder"
          :invalid="Boolean(error)"
          :disabled="disabled"
          :aria-label="label"
          :style="{ '--field-multi-select-max-height': maxHeight }"
          :pt="{
            inputMultiple: {
              onKeydownCapture: prepareBackspaceRemoval,
            },
            input: {
              'data-testid': testId,
              onFocus: onInputFocus,
              onKeydownCapture: onCommitKeydown,
              onInput: onInput,
            },
            option: ({ context }: { context: { option: unknown } }) => ({
              class: 'field-multi-select__option',
              onClickCapture: (event: MouseEvent) =>
                guardInvalidSuggestionClick(event, context.option),
            }),
          }"
          :class="[
            'field-multi-select__input',
            { 'field-multi-select__input--invalid': Boolean(error) },
          ]"
          @complete="completeSuggestions"
          @option-select="onOptionSelect"
          @keydown="containEnterKey"
          @paste="onPaste"
        >
          <template #chip="{ value, removeCallback }">
            <TagLabel
              class="p-chip field-multi-select__tag"
              :text="String(value)"
              :variant="resolveTagVariant(value)"
              button-icon="pi pi-times p-chip-remove-icon"
              :button-label="t('removeValue', { value: String(value) })"
              @action="removeChip(removeCallback, $event)"
            />
          </template>
          <template #option="{ option }">
            <span v-if="isSuggestionInvalid(option)" class="field-multi-select__invalid-suggestion">
              {{ String(option) }}
            </span>
            <TagLabel
              v-else
              class="field-multi-select__suggestion"
              :text="String(option)"
              :variant="resolveTagVariant(option)"
            />
          </template>
        </AutoComplete>
        <div v-if="slots.suffix" class="field-multi-select__suffix">
          <slot name="suffix" />
        </div>
      </div>
    </template>
  </Field>
</template>

<style scoped>
.field-multi-select__wrapper {
  display: flex;
}

.field-multi-select__input {
  width: 100%;
}

.field-multi-select__wrapper--with-suffix .field-multi-select__input {
  flex: 1;
  min-width: 0;
}

.field-multi-select__wrapper--with-suffix
  .field-multi-select__input
  :deep(.p-autocomplete-input-multiple) {
  border-radius: var(--p-form-field-border-radius) 0 0 var(--p-form-field-border-radius);
}

.field-multi-select__suffix {
  display: flex;
  align-items: stretch;
  border: 1px solid var(--p-form-field-border-color);
  border-left: 0;
  border-radius: 0 var(--p-form-field-border-radius) var(--p-form-field-border-radius) 0;
  background: var(--p-surface-100);
  white-space: nowrap;
}

.field-multi-select__input :deep(.p-autocomplete-input-multiple) {
  max-height: var(--field-multi-select-max-height);
  overflow-x: hidden;
  overflow-y: auto;
  align-content: flex-start;
  overscroll-behavior: contain;
}

.field-multi-select__input :deep(.field-multi-select__tag.p-chip) {
  gap: 0;
  padding: 0;
  border: 0;
  background: transparent;
}

.field-multi-select__input :deep(.field-multi-select__tag .tag-label__button) {
  padding-left: 0.25em;
}

.field-multi-select__input :deep(.field-multi-select__option) {
  min-height: 0;
  padding: 0.15rem 0.25rem;
  opacity: 1;
}

@media (max-width: 40rem) {
  .field-multi-select__wrapper--with-suffix {
    flex-direction: column;
  }

  .field-multi-select__wrapper--with-suffix
    .field-multi-select__input
    :deep(.p-autocomplete-input-multiple) {
    border-radius: var(--p-form-field-border-radius) var(--p-form-field-border-radius) 0 0;
  }

  .field-multi-select__suffix {
    border-left: 1px solid var(--p-form-field-border-color);
    border-top: 0;
    border-radius: 0 0 var(--p-form-field-border-radius) var(--p-form-field-border-radius);
  }
}
</style>
