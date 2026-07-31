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
 * Creates or updates an asset-based scan scope.
 * Covers asset_type, asset_companies, asset_departments, asset_countries,
 * asset_locations, asset_contacts, asset_critical fields.
 */
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import AppMessage from '@/components/elements/AppMessage.vue'
import Field from '@/components/elements/Field.vue'
import FieldMultiSelect from '@/components/elements/FieldMultiSelect.vue'
import FieldStepper from '@/components/elements/FieldStepper.vue'
import FieldToggle from '@/components/elements/FieldToggle.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormLifecycle, useFormSnapshot } from '@/composables/useFormLifecycle'
import { useZodForm } from '@/composables/useZodForm'
import { useGroupsStore } from '@/stores/groups'
import { useScopesStore } from '@/stores/scopes'
import { NETWORK_SIZE_MAX } from '@/types/scopes'

import type { AssetCritical, ScopeResponse } from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    any: 'Any',
    scopeAddAssetsPanel: {
      fields: {
        assetType: {
          options: {
            server: 'Server',
            network: 'Network',
          },
          label: 'Device Type',
          hint: 'Restrict the scope to a specific device type. Leave as Any for all types.',
        },
        critical: {
          options: {
            yes: 'Yes',
            no: 'No',
          },
          label: 'Is Critical',
          hint: 'Filter by asset criticality. Any does not filter on this attribute.',
        },
        name: {
          validation: {
            min: 'Name must be at least 3 characters.',
          },
          label: 'Scope Name',
          hint: 'A descriptive name for this scope (at least 3 characters).',
          placeholder: 'Scope Name',
        },
        group: {
          validation: {
            required: 'Please select a group.',
          },
          hint: 'The group this scope belongs to.',
        },
        retention: {
          hint: 'Set the amount of (full) scan cycles you want to retain. Scan results of older scan cycles will be cleaned up.',
        },
        splitSize: {
          label: 'Split Size',
          hint: 'Set the maximum size of network input target. Larger ones will be split automatically. A smaller split size will lead to more and smaller network input targets. A larger split size will allow Nmap to optimize performance. Nmap supports at most 2,048 addresses.',
        },
        companies: {
          label: 'Companies',
          hint: 'Company names. You can enter multiple values. Leave empty for any.',
          placeholder: '*',
        },
        departments: {
          label: 'Departments, Inclusive Subdivision',
          hint: 'Department codes. You can enter multiple values. Leave empty for any. Sub-departments will automatically be included!',
          placeholder: '*',
        },
        contacts: {
          label: 'Contact GIDs',
          hint: 'Contact GIDs. You can enter multiple values. Leave empty for any.',
          placeholder: '*',
        },
        countries: {
          label: 'Country Codes',
          hint: 'Country codes. You can enter multiple values. Leave empty for any.',
          placeholder: '*',
        },
        locations: {
          label: 'Location Codes',
          hint: 'Location codes. You can enter multiple values. Leave empty for any.',
          placeholder: '*',
        },
        cycles: {
          label: 'Scan continuously',
          hint: 'Initiate new scan cycle after all targets were scanned.',
        },
        sync: {
          label: 'Sync scan targets',
          hint: 'Regularly update scan input targets.',
        },
      },
      filters: {
        hint: 'Use the following columns to configure scan result visibility. By default all assets are retrieved, unless at least one filter is set.',
        title: 'Filters (AND)',
      },
      warning:
        'You are responsible for diligent scan target selection. Scans may impact target systems. Only scan systems you are responsible and have a scan approval for!',
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{
  scope?: ScopeResponse
  scopeType?: string
}>()

const emit = defineEmits<{
  close: []
  saved: [message: string]
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const scopesStore = useScopesStore()
const groupsStore = useGroupsStore()

// ── Reactive state ────────────────────────────────────────────
const assetTypeOptions = [
  { label: t('any'), value: 'Any' },
  { label: t('scopeAddAssetsPanel.fields.assetType.options.server'), value: 'Server' },
  { label: t('scopeAddAssetsPanel.fields.assetType.options.network'), value: 'Network' },
  { label: t('common.labels.client'), value: 'Client' },
]

// Is Critical tri-state options (backend stores 'Any' | 'Yes' | 'No').
const criticalOptions = [
  { label: t('any'), value: 'Any' },
  { label: t('scopeAddAssetsPanel.fields.critical.options.yes'), value: 'Yes' },
  { label: t('scopeAddAssetsPanel.fields.critical.options.no'), value: 'No' },
]

const scopeName = ref('')
const scopeCycles = ref(false)
const scopeRetention = ref<number | 'All'>('All')
const scopeNetworkSize = ref(NETWORK_SIZE_MAX)
const scopeSync = ref(false)
const scopeAssetType = ref('Any')
const scopeAssetCritical = ref<AssetCritical>('Any')

const scopeCompanies = ref<string[]>([])
const scopeDepartments = ref<string[]>([])
const scopeCountries = ref<string[]>([])
const scopeLocations = ref<string[]>([])
const scopeContacts = ref<string[]>([])

const selectedGroupId = ref<number | null>(null)

const snapshot = useFormSnapshot(formSnapshot)

// ── Computed ──────────────────────────────────────────────────
const isEditMode = computed(() => !!props.scope)
const resolvedScopeType = computed(() => props.scope?.type ?? props.scopeType ?? 'assets')

const groupOptions = computed(() => groupsStore.groups.map((g) => ({ label: g.name, value: g.id })))

// Group is required only in create mode; schema shape changes based on mode.
const schema = computed(() =>
  z.object({
    name: z.string().trim().min(3, t('scopeAddAssetsPanel.fields.name.validation.min')),
    group: isEditMode.value
      ? z.number().nullable()
      : z.number({ message: t('scopeAddAssetsPanel.fields.group.validation.required') }),
  }),
)

const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, () => ({
  name: scopeName.value,
  group: selectedGroupId.value,
}))

// ── Functions ─────────────────────────────────────────────────
// Baseline snapshot for dirty tracking. Group is excluded since it auto-selects
// when only one group exists and that is not meaningful user input.
function formSnapshot(): string {
  return JSON.stringify({
    scopeName: scopeName.value,
    scopeCycles: scopeCycles.value,
    scopeRetention: scopeRetention.value,
    scopeNetworkSize: scopeNetworkSize.value,
    scopeSync: scopeSync.value,
    scopeAssetType: scopeAssetType.value,
    scopeAssetCritical: scopeAssetCritical.value,
    scopeCompanies: scopeCompanies.value,
    scopeDepartments: scopeDepartments.value,
    scopeCountries: scopeCountries.value,
    scopeLocations: scopeLocations.value,
    scopeContacts: scopeContacts.value,
  })
}

// Initialize the form from props.scope; runs on mount and on scope change. In
// create mode it loads the group list (auto-selecting a lone group). The pristine
// baseline snapshot is captured after init so the dirty guard starts clean.
async function resetForm(): Promise<void> {
  clearErrors()
  scopeName.value = props.scope?.name ?? ''
  scopeCycles.value = props.scope?.cycles ?? false
  scopeRetention.value =
    props.scope && props.scope.cycles_retention > 0 ? props.scope.cycles_retention : 'All'
  scopeNetworkSize.value = props.scope?.split_size ?? NETWORK_SIZE_MAX
  scopeSync.value = props.scope?.attributes?.sync ?? false
  scopeAssetType.value = props.scope?.attributes?.asset_type ?? 'Any'
  scopeAssetCritical.value = props.scope?.attributes?.asset_critical ?? 'Any'
  scopeCompanies.value = props.scope?.attributes?.asset_companies?.slice() ?? []
  scopeDepartments.value = props.scope?.attributes?.asset_departments?.slice() ?? []
  scopeCountries.value = props.scope?.attributes?.asset_countries?.slice() ?? []
  scopeLocations.value = props.scope?.attributes?.asset_locations?.slice() ?? []
  scopeContacts.value = props.scope?.attributes?.asset_contacts?.slice() ?? []
  selectedGroupId.value = null

  if (!isEditMode.value) {
    await groupsStore.fetchGroups()

    // Auto-select the only group; narrow via the element itself so index access is defined
    const onlyGroup = groupsStore.groups.length === 1 ? groupsStore.groups[0] : undefined
    if (onlyGroup) {
      selectedGroupId.value = onlyGroup.id
    }
  }

  // Capture the pristine baseline after init so subsequent edits register as dirty
  snapshot.capture()
}

async function submitForm(): Promise<boolean> {
  if (!validateAll()) return false

  // Resolve either the existing scope or the validated owner group without assertions
  let scopeIdentifier: { scope_id: number } | { group_id: number }
  if (props.scope) {
    scopeIdentifier = { scope_id: props.scope.id }
  } else {
    const groupId = selectedGroupId.value
    if (groupId === null) return false
    scopeIdentifier = { group_id: groupId }
  }

  const retention =
    typeof scopeRetention.value === 'number' && scopeRetention.value > 0 ? scopeRetention.value : -1

  const payload = {
    type: resolvedScopeType.value,
    name: scopeName.value.trim(),
    cycles: scopeCycles.value,
    cycles_retention: retention,
    split_size: scopeNetworkSize.value,
    sync: scopeSync.value,
    asset_type: scopeAssetType.value,
    asset_companies: scopeCompanies.value,
    asset_departments: scopeDepartments.value,
    asset_countries: scopeCountries.value,
    asset_locations: scopeLocations.value,
    asset_contacts: scopeContacts.value,
    asset_critical: scopeAssetCritical.value,
    ...scopeIdentifier,
  }

  const message = await scopesStore.submitAssets(payload)
  emit('saved', message)
  return true
}

// ── Watchers and lifecycle ────────────────────────────────────
// Reset-on-open, dirty registration, and the submit/cancel handlers the template
// binds to. A store error is caught by useFormLifecycle and the form stays open.
const { submit: submitWithLifecycle, cancel } = useFormLifecycle({
  target: () => props.scope,
  reset: resetForm,
  dirty: snapshot.isDirty,
  submit: submitForm,
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitWithLifecycle })
</script>

<template>
  <div class="form-scope-add-assets__grid form-scope-add-assets__grid--top">
    <Field
      v-if="!isEditMode"
      v-slot="{ id }"
      :label="t('common.labels.group')"
      required
      :help="t('scopeAddAssetsPanel.fields.group.hint')"
      tooltip
      :error="errors.group"
      :validate="() => validateField('group')"
    >
      <Select
        :id="id"
        v-model="selectedGroupId"
        :options="groupOptions"
        option-label="label"
        option-value="value"
        :placeholder="t('common.placeholders.select')"
        :invalid="!!errors.group"
        :aria-label="t('common.labels.group')"
      />
    </Field>
    <Field v-else v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText :id="id" :model-value="scope?.group_name" disabled />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('scopeAddAssetsPanel.fields.name.label')"
      required
      :help="t('scopeAddAssetsPanel.fields.name.hint')"
      tooltip
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="scopeName"
        :placeholder="t('scopeAddAssetsPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('scopeAddAssetsPanel.fields.name.label')"
      />
    </Field>

    <div class="form-scope-add-assets__stepper-pair">
      <FieldStepper
        v-model="scopeRetention"
        :label="t('common.labels.retention')"
        :help="t('scopeAddAssetsPanel.fields.retention.hint')"
        tooltip
      />
      <FieldStepper
        v-model="scopeNetworkSize"
        kind="splitSize"
        :label="t('scopeAddAssetsPanel.fields.splitSize.label')"
        :help="t('scopeAddAssetsPanel.fields.splitSize.hint')"
        tooltip
      />
    </div>
  </div>

  <SectionHeader
    level="h4"
    variant="horizontal"
    test-id="saa-filters-header"
    :help-text="t('scopeAddAssetsPanel.filters.hint')"
    >{{ t('scopeAddAssetsPanel.filters.title') }}</SectionHeader
  >

  <div class="form-scope-add-assets__grid form-scope-add-assets__grid--thirds">
    <FieldMultiSelect
      v-model="scopeCompanies"
      :label="t('scopeAddAssetsPanel.fields.companies.label')"
      :help="t('scopeAddAssetsPanel.fields.companies.hint')"
      tooltip
      :placeholder="t('scopeAddAssetsPanel.fields.companies.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeDepartments"
      :label="t('scopeAddAssetsPanel.fields.departments.label')"
      :help="t('scopeAddAssetsPanel.fields.departments.hint')"
      tooltip
      :placeholder="t('scopeAddAssetsPanel.fields.departments.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeContacts"
      :label="t('scopeAddAssetsPanel.fields.contacts.label')"
      :help="t('scopeAddAssetsPanel.fields.contacts.hint')"
      tooltip
      :placeholder="t('scopeAddAssetsPanel.fields.contacts.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeCountries"
      :label="t('scopeAddAssetsPanel.fields.countries.label')"
      :help="t('scopeAddAssetsPanel.fields.countries.hint')"
      tooltip
      :placeholder="t('scopeAddAssetsPanel.fields.countries.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeLocations"
      :label="t('scopeAddAssetsPanel.fields.locations.label')"
      :help="t('scopeAddAssetsPanel.fields.locations.hint')"
      tooltip
      :placeholder="t('scopeAddAssetsPanel.fields.locations.placeholder')"
    />

    <div class="form-scope-add-assets__pair">
      <Field
        v-slot="{ id }"
        :label="t('scopeAddAssetsPanel.fields.assetType.label')"
        :help="t('scopeAddAssetsPanel.fields.assetType.hint')"
        tooltip
      >
        <Select
          :id="id"
          v-model="scopeAssetType"
          :options="assetTypeOptions"
          option-label="label"
          option-value="value"
          :aria-label="t('scopeAddAssetsPanel.fields.assetType.label')"
        />
      </Field>

      <Field
        v-slot="{ id }"
        :label="t('scopeAddAssetsPanel.fields.critical.label')"
        :help="t('scopeAddAssetsPanel.fields.critical.hint')"
        tooltip
      >
        <Select
          :id="id"
          v-model="scopeAssetCritical"
          data-testid="saa-critical"
          :options="criticalOptions"
          option-label="label"
          option-value="value"
          :aria-label="t('scopeAddAssetsPanel.fields.critical.label')"
        />
      </Field>
    </div>
  </div>

  <AppMessage severity="warn" data-testid="saa-responsibility">
    <p>{{ t('scopeAddAssetsPanel.warning') }}</p>
  </AppMessage>

  <!-- Bottom row: Cycles + Sync toggles each take a third, with the
       Cancel/Save actions filling the last third (matches the filter grid above). -->
  <div
    class="form-scope-add-assets__grid form-scope-add-assets__grid--thirds form-scope-add-assets__bottom-row"
  >
    <FieldToggle
      v-model="scopeCycles"
      :label="t('scopeAddAssetsPanel.fields.cycles.label')"
      :help="t('scopeAddAssetsPanel.fields.cycles.hint')"
      tooltip
    />
    <FieldToggle
      v-model="scopeSync"
      :label="t('scopeAddAssetsPanel.fields.sync.label')"
      :help="t('scopeAddAssetsPanel.fields.sync.hint')"
      tooltip
    />
    <div class="form-scope-add-assets__actions">
      <LayoutSubmitFooter
        cancel-test-id="saa-cancel"
        :text-submit-label="t('common.actions.save')"
        @cancel="cancel"
      />
    </div>
  </div>
</template>

<style scoped>
/* Keep the field grid consistent with the scope settings form */
.form-scope-add-assets__grid {
  display: grid;
  gap: 0.75rem;
}

/* Top row Group|Name|Retention on one line; filter chips three per row.
   minmax(0,…) lets the dropdown/chip intrinsic widths shrink so columns stay
   even; controls fill their cell. */
.form-scope-add-assets__grid--top {
  grid-template-columns: 1fr 1fr 1fr;
}

.form-scope-add-assets__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

/* Retention and network segmentation share the final top-row cell */
.form-scope-add-assets__stepper-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

/* Two fields side-by-side sharing one grid cell */
.form-scope-add-assets__pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.form-scope-add-assets__grid :deep(.p-select),
.form-scope-add-assets__grid :deep(.p-inputtext),
.form-scope-add-assets__grid :deep(.p-autocomplete) {
  width: 100%;
}

.form-scope-add-assets__bottom-row {
  align-items: center;
}

/* Remove field rhythm that would shift the visible toggles above the action row center */
.form-scope-add-assets__bottom-row :deep(.field-toggle) {
  margin-bottom: 0;
}

.form-scope-add-assets__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}
</style>
