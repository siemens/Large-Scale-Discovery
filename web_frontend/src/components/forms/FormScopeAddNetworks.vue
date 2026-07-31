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
 * Creates or updates a network-based scan scope.
 * Fields: companies, departments, routing_domains, zones, purposes, countries,
 * locations, contacts, exclude_keywords.
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

import type { ScopeResponse } from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    scopeAddNetworksPanel: {
      fields: {
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
        routingDomains: {
          label: 'Routing Domains',
          hint: 'Routing domains. You can enter multiple values. Leave empty for any.',
          placeholder: '*',
        },
        zones: {
          label: 'Zones',
          hint: 'Zone names. You can enter multiple values. Leave empty for any.',
          placeholder: '-',
        },
        purposes: {
          label: 'Purposes',
          hint: 'Network purpose descriptions. You can enter multiple values. Leave empty for any.',
          placeholder: '-',
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
        excludeKeywords: {
          label: 'Exclude Keywords',
          hint: 'Keywords to skip. All fields are searched for those keywords! You can enter multiple values. Leave empty for any.',
          placeholder: '-',
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
const scopeName = ref('')
const scopeCycles = ref(false)
const scopeRetention = ref<number | 'All'>('All')
const scopeNetworkSize = ref(NETWORK_SIZE_MAX)
const scopeSync = ref(false)

const scopeCompanies = ref<string[]>([])
const scopeDepartments = ref<string[]>([])
const scopeRoutingDomains = ref<string[]>([])
const scopeZones = ref<string[]>([])
const scopePurposes = ref<string[]>([])
const scopeCountries = ref<string[]>([])
const scopeLocations = ref<string[]>([])
const scopeContacts = ref<string[]>([])
const scopeExcludeKeywords = ref<string[]>([])

// Routing-domain presets always offered first in the chip dropdown.
const routingDomainSeed = ['Global (Intranet + Internet)', 'Internet', 'Intranet']

const selectedGroupId = ref<number | null>(null)

const snapshot = useFormSnapshot(formSnapshot)

// ── Computed ──────────────────────────────────────────────────
const isEditMode = computed(() => !!props.scope)
const resolvedScopeType = computed(() => props.scope?.type ?? props.scopeType ?? 'networks')

const groupOptions = computed(() => groupsStore.groups.map((g) => ({ label: g.name, value: g.id })))

// Group is required only in create mode; schema shape changes based on mode.
const schema = computed(() =>
  z.object({
    name: z.string().trim().min(3, t('scopeAddNetworksPanel.fields.name.validation.min')),
    group: isEditMode.value
      ? z.number().nullable()
      : z.number({ message: t('scopeAddNetworksPanel.fields.group.validation.required') }),
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
    scopeCompanies: scopeCompanies.value,
    scopeDepartments: scopeDepartments.value,
    scopeRoutingDomains: scopeRoutingDomains.value,
    scopeZones: scopeZones.value,
    scopePurposes: scopePurposes.value,
    scopeCountries: scopeCountries.value,
    scopeLocations: scopeLocations.value,
    scopeContacts: scopeContacts.value,
    scopeExcludeKeywords: scopeExcludeKeywords.value,
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
  scopeCompanies.value = props.scope?.attributes?.asset_companies?.slice() ?? []
  scopeDepartments.value = props.scope?.attributes?.asset_departments?.slice() ?? []
  scopeRoutingDomains.value = props.scope?.attributes?.asset_routing_domains?.slice() ?? []
  scopeZones.value = props.scope?.attributes?.asset_zones?.slice() ?? []
  scopePurposes.value = props.scope?.attributes?.asset_purposes?.slice() ?? []
  scopeCountries.value = props.scope?.attributes?.asset_countries?.slice() ?? []
  scopeLocations.value = props.scope?.attributes?.asset_locations?.slice() ?? []
  scopeContacts.value = props.scope?.attributes?.asset_contacts?.slice() ?? []
  scopeExcludeKeywords.value = props.scope?.attributes?.asset_exclude_keywords?.slice() ?? []
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
    asset_companies: scopeCompanies.value,
    asset_departments: scopeDepartments.value,
    asset_routing_domains: scopeRoutingDomains.value,
    asset_zones: scopeZones.value,
    asset_purposes: scopePurposes.value,
    asset_countries: scopeCountries.value,
    asset_locations: scopeLocations.value,
    asset_contacts: scopeContacts.value,
    asset_exclude_keywords: scopeExcludeKeywords.value,
    ...scopeIdentifier,
  }

  const message = await scopesStore.submitNetworks(payload)
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
  <div class="form-scope-add-networks__grid form-scope-add-networks__grid--top">
    <Field
      v-if="!isEditMode"
      v-slot="{ id }"
      :label="t('common.labels.group')"
      required
      :help="t('scopeAddNetworksPanel.fields.group.hint')"
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
      :label="t('scopeAddNetworksPanel.fields.name.label')"
      required
      :help="t('scopeAddNetworksPanel.fields.name.hint')"
      tooltip
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="scopeName"
        :placeholder="t('scopeAddNetworksPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('scopeAddNetworksPanel.fields.name.label')"
      />
    </Field>

    <div class="form-scope-add-networks__stepper-pair">
      <FieldStepper
        v-model="scopeRetention"
        :label="t('common.labels.retention')"
        :help="t('scopeAddNetworksPanel.fields.retention.hint')"
        tooltip
      />
      <FieldStepper
        v-model="scopeNetworkSize"
        kind="splitSize"
        :label="t('scopeAddNetworksPanel.fields.splitSize.label')"
        :help="t('scopeAddNetworksPanel.fields.splitSize.hint')"
        tooltip
      />
    </div>
  </div>

  <SectionHeader
    level="h4"
    variant="horizontal"
    :help-text="t('scopeAddNetworksPanel.filters.hint')"
    test-id="san-filters-header"
    >{{ t('scopeAddNetworksPanel.filters.title') }}</SectionHeader
  >

  <div class="form-scope-add-networks__grid form-scope-add-networks__grid--thirds">
    <FieldMultiSelect
      v-model="scopeCompanies"
      :label="t('scopeAddNetworksPanel.fields.companies.label')"
      :help="t('scopeAddNetworksPanel.fields.companies.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.companies.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeDepartments"
      :label="t('scopeAddNetworksPanel.fields.departments.label')"
      :help="t('scopeAddNetworksPanel.fields.departments.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.departments.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeContacts"
      :label="t('scopeAddNetworksPanel.fields.contacts.label')"
      :help="t('scopeAddNetworksPanel.fields.contacts.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.contacts.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeRoutingDomains"
      :label="t('scopeAddNetworksPanel.fields.routingDomains.label')"
      :help="t('scopeAddNetworksPanel.fields.routingDomains.hint')"
      tooltip
      :seed="routingDomainSeed"
      :placeholder="t('scopeAddNetworksPanel.fields.routingDomains.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeZones"
      :label="t('scopeAddNetworksPanel.fields.zones.label')"
      :help="t('scopeAddNetworksPanel.fields.zones.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.zones.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopePurposes"
      :label="t('scopeAddNetworksPanel.fields.purposes.label')"
      :help="t('scopeAddNetworksPanel.fields.purposes.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.purposes.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeCountries"
      :label="t('scopeAddNetworksPanel.fields.countries.label')"
      :help="t('scopeAddNetworksPanel.fields.countries.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.countries.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeLocations"
      :label="t('scopeAddNetworksPanel.fields.locations.label')"
      :help="t('scopeAddNetworksPanel.fields.locations.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.locations.placeholder')"
    />
    <FieldMultiSelect
      v-model="scopeExcludeKeywords"
      :label="t('scopeAddNetworksPanel.fields.excludeKeywords.label')"
      :help="t('scopeAddNetworksPanel.fields.excludeKeywords.hint')"
      tooltip
      :placeholder="t('scopeAddNetworksPanel.fields.excludeKeywords.placeholder')"
    />
  </div>

  <AppMessage severity="warn" data-testid="san-responsibility">
    <p>{{ t('scopeAddNetworksPanel.warning') }}</p>
  </AppMessage>

  <!-- Bottom row: Cycles + Sync toggles each take a third, with the
       Cancel/Save actions filling the last third (matches the filter grid above). -->
  <div
    class="form-scope-add-networks__grid form-scope-add-networks__grid--thirds form-scope-add-networks__bottom-row"
  >
    <FieldToggle
      v-model="scopeCycles"
      :label="t('scopeAddNetworksPanel.fields.cycles.label')"
      :help="t('scopeAddNetworksPanel.fields.cycles.hint')"
      tooltip
    />
    <FieldToggle
      v-model="scopeSync"
      :label="t('scopeAddNetworksPanel.fields.sync.label')"
      :help="t('scopeAddNetworksPanel.fields.sync.hint')"
      tooltip
    />
    <div class="form-scope-add-networks__actions">
      <LayoutSubmitFooter
        cancel-test-id="san-cancel"
        :text-submit-label="t('common.actions.save')"
        @cancel="cancel"
      />
    </div>
  </div>
</template>

<style scoped>
/* Keep the field grid consistent with the scope settings form */
.form-scope-add-networks__grid {
  display: grid;
  gap: 0.75rem;
}

/* Top row Group|Name|Retention on one line; filter chips three per row.
   minmax(0,…) lets the dropdown/chip intrinsic widths shrink so columns stay
   even; controls fill their cell. */
.form-scope-add-networks__grid--top {
  grid-template-columns: 1fr 1fr 1fr;
}

.form-scope-add-networks__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

/* Retention and network segmentation share the final top-row cell */
.form-scope-add-networks__stepper-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.form-scope-add-networks__grid :deep(.p-select),
.form-scope-add-networks__grid :deep(.p-inputtext),
.form-scope-add-networks__grid :deep(.p-autocomplete) {
  width: 100%;
}

.form-scope-add-networks__bottom-row {
  align-items: center;
}

/* Remove field rhythm that would shift the visible toggles above the action row center */
.form-scope-add-networks__bottom-row :deep(.field-toggle) {
  margin-bottom: 0;
}

.form-scope-add-networks__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}
</style>
