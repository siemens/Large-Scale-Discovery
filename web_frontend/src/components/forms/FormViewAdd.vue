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
 * View creation form rendered inline above the views table.
 * pick a source scope, optionally narrow with OR-combined filters, then
 * `POST /view/create`. Filter fields are free-entry token inputs; the scope's
 * own targets are offered as suggestions for the Targets field.
 */
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import AppMessage from '@/components/elements/AppMessage.vue'
import Field from '@/components/elements/Field.vue'
import FieldMultiSelect from '@/components/elements/FieldMultiSelect.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { useScopesStore } from '@/stores/scopes'
import { useViewsStore } from '@/stores/views'

import type { CreateViewRequest } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    viewAddPanel: {
      filters: {
        fields: {
          countries: {
            label: 'Country Code',
            placeholder: '-',
          },
          locations: {
            label: 'Location Code',
            placeholder: '-',
          },
          routingDomains: {
            placeholder: '-',
          },
          zones: {
            placeholder: '-',
          },
          purposes: {
            placeholder: '-',
          },
          companies: {
            placeholder: '-',
          },
          departments: {
            placeholder: '-',
          },
          managers: {
            placeholder: '-',
          },
          contacts: {
            placeholder: '-',
          },
          comments: {
            placeholder: '-',
          },
          targets: {
            label: 'Scope Input Targets',
            placeholder: '-',
          },
        },
        hint: 'Use the following columns to configure scan result visibility.',
        title: 'Filters (OR)',
        warning:
          'By default all results are accessible, unless at least one filter is set. All filters are OR-combined, loosening access one by one.',
      },
      fields: {
        name: {
          validation: {
            min: 'View name must be at least 3 characters.',
          },
          label: 'View Name',
          placeholder: 'View Name',
        },
        scope: {
          validation: {
            required: 'A source scope is required.',
          },
          label: 'Source Scope',
        },
      },
    },
  },
}

// The Targets filter is rendered separately (full-width, with scope-target
// suggestions); the rest flow three per row.
type FilterKey =
  | 'targets'
  | 'countries'
  | 'locations'
  | 'routingDomains'
  | 'zones'
  | 'purposes'
  | 'companies'
  | 'departments'
  | 'managers'
  | 'contacts'
  | 'comments'

const FILTER_FIELD_CONFIG: Array<{
  key: FilterKey
  label: string
  placeholder: string
}> = [
  {
    key: 'countries',
    label: 'viewAddPanel.filters.fields.countries.label',
    placeholder: 'viewAddPanel.filters.fields.countries.placeholder',
  },
  {
    key: 'locations',
    label: 'viewAddPanel.filters.fields.locations.label',
    placeholder: 'viewAddPanel.filters.fields.locations.placeholder',
  },
  {
    key: 'routingDomains',
    label: 'common.labels.routingDomain',
    placeholder: 'viewAddPanel.filters.fields.routingDomains.placeholder',
  },
  {
    key: 'zones',
    label: 'common.labels.zone',
    placeholder: 'viewAddPanel.filters.fields.zones.placeholder',
  },
  {
    key: 'purposes',
    label: 'common.labels.purpose',
    placeholder: 'viewAddPanel.filters.fields.purposes.placeholder',
  },
  {
    key: 'companies',
    label: 'common.labels.company',
    placeholder: 'viewAddPanel.filters.fields.companies.placeholder',
  },
  {
    key: 'departments',
    label: 'common.labels.department',
    placeholder: 'viewAddPanel.filters.fields.departments.placeholder',
  },
  {
    key: 'managers',
    label: 'common.labels.manager',
    placeholder: 'viewAddPanel.filters.fields.managers.placeholder',
  },
  {
    key: 'contacts',
    label: 'common.labels.contact',
    placeholder: 'viewAddPanel.filters.fields.contacts.placeholder',
  },
  {
    key: 'comments',
    label: 'common.labels.comment',
    placeholder: 'viewAddPanel.filters.fields.comments.placeholder',
  },
]

const emit = defineEmits<{ close: []; created: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const viewsStore = useViewsStore()
const scopesStore = useScopesStore()

// ── State ─────────────────────────────────────────────────────
const viewName = ref('')
const selectedScopeId = ref<number | null>(null)

const schema = z.object({
  name: z.string().trim().min(3, t('viewAddPanel.fields.name.validation.min')),
  scope: z.number({
    message: t('viewAddPanel.fields.scope.validation.required'),
  }),
})
const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, () => ({
  name: viewName.value,
  scope: selectedScopeId.value,
}))

const filters = ref<Record<FilterKey, string[]>>({
  targets: [],
  countries: [],
  locations: [],
  routingDomains: [],
  zones: [],
  purposes: [],
  companies: [],
  departments: [],
  managers: [],
  contacts: [],
  comments: [],
})

// ── Computed ──────────────────────────────────────────────────
const selectedScope = computed(() =>
  scopesStore.scopes.find((scope) => scope.id === selectedScopeId.value),
)

const targetFilterOptions = computed(() =>
  (selectedScope.value?.attributes?.targets ?? [])
    .map((target) => target.input)
    .filter((input): input is string => Boolean(input)),
)

// ── Watchers and lifecycle ────────────────────────────────────
const { submit: submitForm, cancel } = useFormLifecycle({
  // Load scopes and reset whenever the form opens.
  async reset(): Promise<void> {
    try {
      await scopesStore.fetchScopes()
    } catch {
      // Keep the form usable after the shared error reporter handled the failure
    }
    viewName.value = ''

    // Auto-select the only scope; narrow via the element itself so index access is defined
    selectedScopeId.value =
      scopesStore.scopes.length === 1 ? (scopesStore.scopes[0]?.id ?? null) : null
    clearErrors()
    for (const values of Object.values(filters.value)) values.length = 0
  },

  // `selectedScopeId` is excluded: it auto-selects when a single scope exists, which
  // isn't user-meaningful unsaved input (the name + filters are).
  dirty: () =>
    viewName.value.trim() !== '' ||
    Object.values(filters.value).some((values) => values.length > 0),
  async submit(): Promise<boolean> {
    if (!validateAll()) return false
    const scopeId = selectedScopeId.value
    if (scopeId === null) return false

    const payload: CreateViewRequest = {
      scope_id: scopeId,
      view_name: viewName.value.trim(),
      filter_input_targets: filters.value.targets,
      filter_input_countries: filters.value.countries,
      filter_input_locations: filters.value.locations,
      filter_input_routing_domains: filters.value.routingDomains,
      filter_input_zones: filters.value.zones,
      filter_input_purposes: filters.value.purposes,
      filter_input_companies: filters.value.companies,
      filter_input_departments: filters.value.departments,
      filter_input_managers: filters.value.managers,
      filter_input_contacts: filters.value.contacts,
      filter_input_comments: filters.value.comments,
    }
    const message = await viewsStore.createView(payload)
    toast.success(message)
    emit('created')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-view-add__grid form-view-add__grid--top">
    <Field
      v-slot="{ id }"
      :label="t('viewAddPanel.fields.name.label')"
      required
      tooltip
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="viewName"
        data-testid="view-add-name"
        class="form-view-add__input"
        :placeholder="t('viewAddPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('viewAddPanel.fields.name.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('viewAddPanel.fields.scope.label')"
      required
      tooltip
      :error="errors.scope"
      :validate="() => validateField('scope')"
    >
      <Select
        :id="id"
        v-model="selectedScopeId"
        data-testid="view-add-scope"
        :options="scopesStore.scopes"
        option-label="name"
        option-value="id"
        class="form-view-add__input"
        :placeholder="t('common.placeholders.select')"
        :invalid="!!errors.scope"
        :loading="scopesStore.isFetching"
        :aria-label="t('viewAddPanel.fields.scope.label')"
      />
    </Field>
  </div>

  <SectionHeader level="h4" variant="horizontal" :help-text="t('viewAddPanel.filters.hint')">{{
    t('viewAddPanel.filters.title')
  }}</SectionHeader>

  <AppMessage severity="warn">
    <p>{{ t('viewAddPanel.filters.warning') }}</p>
  </AppMessage>

  <!-- Scope Input Targets spans the full width (legacy layout), then the
         remaining filter fields sit three per row. The last row shares space
         with the Cancel/Save actions (col 2 empty, col 3 = buttons). -->
  <FieldMultiSelect
    v-model="filters.targets"
    :label="t('viewAddPanel.filters.fields.targets.label')"
    tooltip
    :options="targetFilterOptions"
    :placeholder="t('viewAddPanel.filters.fields.targets.placeholder')"
  />

  <div class="form-view-add__grid form-view-add__grid--thirds">
    <FieldMultiSelect
      v-for="filterField in FILTER_FIELD_CONFIG"
      :key="filterField.key"
      v-model="filters[filterField.key]"
      :label="t(filterField.label)"
      tooltip
      :placeholder="t(filterField.placeholder)"
      :class="{
        'form-view-add__field--no-margin': filterField.key === 'comments',
      }"
    />
    <div></div>
    <div class="form-view-add__actions">
      <div class="form-view-add__actions-spacer" aria-hidden="true">&nbsp;</div>
      <div class="form-view-add__actions-buttons">
        <LayoutSubmitFooter
          submit-test-id="view-add-submit"
          :text-submit-label="t('common.actions.create')"
          @cancel="cancel"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.form-view-add__input {
  width: 100%;
}

/* Multi-column rows (legacy layout parity): Name|Scope on one line, filter
   chips three per row. minmax(0,…) keeps columns even against chip intrinsic
   widths; controls fill their cell. */
.form-view-add__grid {
  display: grid;
  gap: 0 0.75rem;
}

.form-view-add__grid--top {
  grid-template-columns: 1fr 1fr;
}

.form-view-add__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-view-add__grid :deep(.p-select),
.form-view-add__grid :deep(.p-inputtext),
.form-view-add__grid :deep(.p-autocomplete) {
  width: 100%;
}

.form-view-add__field--no-margin {
  margin-bottom: 0;
}

.form-view-add__actions {
  display: flex;
  flex-direction: column;
}

.form-view-add__actions-spacer {
  min-height: 1.5rem;
}

.form-view-add__actions-buttons {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}
</style>
