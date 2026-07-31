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
 * Group creation form with its database assignment and per-resource
 * limits, and allowed scan types. Limits accept `*` or empty as "unlimited".
 */

import { storeToRefs } from 'pinia'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import Field from '@/components/elements/Field.vue'
import FieldToggle from '@/components/elements/FieldToggle.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { useDatabasesStore } from '@/stores/databases'
import { useGroupsStore } from '@/stores/groups'

import type { DatabaseResponse } from '@/types/databases'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    groupAddPanel: {
      fields: {
        name: {
          validation: {
            required: 'Group name is required.',
            min: 'Group name must be at least 3 characters.',
          },
          placeholder: 'Group Name',
        },
        database: {
          validation: {
            required: 'Please select a database.',
          },
        },
        maxScopes: {
          validation: {
            invalid: 'Enter a number ≥ -1 or * for unlimited.',
          },
          label: 'Max Scopes',
          placeholder: '*',
        },
        maxViews: {
          validation: {
            invalid: 'Enter a number ≥ -1 or * for unlimited.',
          },
          label: 'Max Views',
          placeholder: '*',
        },
        maxTargets: {
          validation: {
            invalid: 'Enter a number ≥ -1 or * for unlimited.',
          },
          label: 'Max Targets',
          placeholder: '*',
        },
        maxOwners: {
          validation: {
            invalid: 'Enter a number ≥ -1 or * for unlimited.',
          },
          label: 'Max Owners',
          placeholder: '*',
        },
        allowCustom: {
          label: 'Custom Input',
          help: 'Allow users to create scan scopes with custom input lists.',
        },
        allowNetwork: {
          label: 'Network Imports',
          help: 'Allow users to create scan scopes with inputs imported from the network inventory.',
        },
        allowAsset: {
          label: 'Asset Imports',
          help: 'Allow users to create scan scopes with inputs imported from the asset inventory.',
        },
      },
      success: 'Group created.',
    },
  },
}

type GroupForm = {
  name: string
  db_server_id: number | null
  max_scopes: string
  max_views: string
  max_targets: string
  max_owners: string
  allow_custom: boolean
  allow_network: boolean
  allow_asset: boolean
}

// ── Props and emits ───────────────────────────────────────────
const emit = defineEmits<{ close: []; created: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const groupsStore = useGroupsStore()
const databasesStore = useDatabasesStore()
const { databases } = storeToRefs(databasesStore)

// ── Reactive state ────────────────────────────────────────────
const form = ref<GroupForm>(createDefaultGroupForm())

// ── Computed ──────────────────────────────────────────────────
// Validation schema. Limits are validated with isLimitValueValid below.
const addGroupSchema = computed(() =>
  z.object({
    name: z
      .string()
      .min(1, t('groupAddPanel.fields.name.validation.required'))
      .min(3, t('groupAddPanel.fields.name.validation.min')),
    db_server_id: z
      .number({ error: t('groupAddPanel.fields.database.validation.required') })
      .min(1, t('groupAddPanel.fields.database.validation.required')),
    max_scopes: z
      .string()
      .refine(isLimitValueValid, t('groupAddPanel.fields.maxScopes.validation.invalid')),
    max_views: z
      .string()
      .refine(isLimitValueValid, t('groupAddPanel.fields.maxViews.validation.invalid')),
    max_targets: z
      .string()
      .refine(isLimitValueValid, t('groupAddPanel.fields.maxTargets.validation.invalid')),
    max_owners: z
      .string()
      .refine(isLimitValueValid, t('groupAddPanel.fields.maxOwners.validation.invalid')),
    allow_custom: z.boolean(),
    allow_network: z.boolean(),
    allow_asset: z.boolean(),
  }),
)

const databaseOptions = computed<DatabaseResponse[]>(() => databases.value)

const { errors, validateField, validateAll, clearErrors } = useZodForm(addGroupSchema, form)

// ── Functions ─────────────────────────────────────────────────
// Translate a limit field into its numeric form, where `*` or empty means unlimited (-1).
function parseLimitValue(raw: string): number {
  const trimmed = raw.trim()
  if (trimmed === '' || trimmed === '*') return -1
  return parseInt(trimmed, 10)
}

// Accept `*`, empty, or any integer at or above -1 as a valid limit.
function isLimitValueValid(raw: string): boolean {
  const trimmed = raw.trim()
  if (trimmed === '' || trimmed === '*') return true
  const parsedLimit = parseInt(trimmed, 10)
  return !isNaN(parsedLimit) && parsedLimit >= -1
}

// Produce a blank form. Hoisted so form = ref(createDefaultGroupForm()) above can call it.
function createDefaultGroupForm(): GroupForm {
  return {
    name: '',
    db_server_id: null,
    max_scopes: '*',
    max_views: '*',
    max_targets: '*',
    max_owners: '*',
    allow_custom: true,
    allow_network: false,
    allow_asset: false,
  }
}

// ── Watchers and lifecycle ────────────────────────────────────
// Auto-select the only database once the list loads asynchronously. The fetch is
// kicked off in reset() when the form opens with an empty list.
watch(databases, (dbs) => {
  const onlyDatabase = dbs.length === 1 ? dbs[0] : undefined
  if (onlyDatabase && form.value.db_server_id === null) {
    form.value.db_server_id = onlyDatabase.id
  }
})

// Shared form lifecycle: reset-on-mount, the unsaved-changes guard, cancel(),
// and the submit try/catch skeleton.
const { submit: submitForm, cancel } = useFormLifecycle({
  reset(): void {
    form.value = createDefaultGroupForm()
    clearErrors()
    if (databases.value.length === 0) void databasesStore.fetchDatabases()

    // Auto-select the only database; narrow via the element itself so index access is defined
    else if (databases.value.length === 1) form.value.db_server_id = databases.value[0]?.id ?? null
  },

  // `db_server_id` is excluded: it auto-selects when a single database exists,
  // which isn't user-meaningful unsaved input.
  dirty: () =>
    form.value.name !== '' ||
    form.value.max_scopes !== '*' ||
    form.value.max_views !== '*' ||
    form.value.max_targets !== '*' ||
    form.value.max_owners !== '*' ||
    form.value.allow_custom !== true ||
    form.value.allow_network !== false ||
    form.value.allow_asset !== false,
  async submit(): Promise<boolean> {
    if (!validateAll()) return false

    // Narrow the validated database before constructing the backend payload
    const databaseId = form.value.db_server_id
    if (databaseId === null) return false

    await groupsStore.createGroup({
      name: form.value.name,
      db_server_id: databaseId,
      max_scopes: parseLimitValue(form.value.max_scopes),
      max_views: parseLimitValue(form.value.max_views),
      max_targets: parseLimitValue(form.value.max_targets),
      max_owners: parseLimitValue(form.value.max_owners),
      allow_custom: form.value.allow_custom,
      allow_network: form.value.allow_network,
      allow_asset: form.value.allow_asset,
    })
    toast.success(t('groupAddPanel.success'))
    emit('created')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-group-add__row-2">
    <Field
      v-slot="{ id }"
      :label="t('common.labels.name')"
      required
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="form.name"
        data-testid="agd-name"
        :placeholder="t('groupAddPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('common.labels.name')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('common.labels.database')"
      required
      :error="errors.db_server_id"
      :validate="() => validateField('db_server_id')"
    >
      <Select
        :id="id"
        v-model="form.db_server_id"
        data-testid="agd-database"
        :options="databaseOptions"
        option-label="name"
        option-value="id"
        :placeholder="t('common.placeholders.select')"
        :invalid="!!errors.db_server_id"
        :aria-label="t('common.labels.database')"
      />
    </Field>
  </div>

  <div class="form-group-add__row-4">
    <Field
      v-slot="{ id }"
      :label="t('groupAddPanel.fields.maxScopes.label')"
      required
      :error="errors.max_scopes"
      :validate="() => validateField('max_scopes')"
    >
      <InputText
        :id="id"
        v-model="form.max_scopes"
        :placeholder="t('groupAddPanel.fields.maxScopes.placeholder')"
        :invalid="!!errors.max_scopes"
        :aria-label="t('groupAddPanel.fields.maxScopes.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('groupAddPanel.fields.maxViews.label')"
      required
      :error="errors.max_views"
      :validate="() => validateField('max_views')"
    >
      <InputText
        :id="id"
        v-model="form.max_views"
        :placeholder="t('groupAddPanel.fields.maxViews.placeholder')"
        :invalid="!!errors.max_views"
        :aria-label="t('groupAddPanel.fields.maxViews.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('groupAddPanel.fields.maxTargets.label')"
      required
      :error="errors.max_targets"
      :validate="() => validateField('max_targets')"
    >
      <InputText
        :id="id"
        v-model="form.max_targets"
        :placeholder="t('groupAddPanel.fields.maxTargets.placeholder')"
        :invalid="!!errors.max_targets"
        :aria-label="t('groupAddPanel.fields.maxTargets.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('groupAddPanel.fields.maxOwners.label')"
      required
      :error="errors.max_owners"
      :validate="() => validateField('max_owners')"
    >
      <InputText
        :id="id"
        v-model="form.max_owners"
        :placeholder="t('groupAddPanel.fields.maxOwners.placeholder')"
        :invalid="!!errors.max_owners"
        :aria-label="t('groupAddPanel.fields.maxOwners.label')"
      />
    </Field>
  </div>

  <div class="form-group-add__bottom-row">
    <FieldToggle
      v-model="form.allow_custom"
      :label="t('groupAddPanel.fields.allowCustom.label')"
      :help="t('groupAddPanel.fields.allowCustom.help')"
      tooltip
    />
    <FieldToggle
      v-model="form.allow_network"
      :label="t('groupAddPanel.fields.allowNetwork.label')"
      :help="t('groupAddPanel.fields.allowNetwork.help')"
      tooltip
    />
    <FieldToggle
      v-model="form.allow_asset"
      :label="t('groupAddPanel.fields.allowAsset.label')"
      :help="t('groupAddPanel.fields.allowAsset.help')"
      tooltip
    />

    <div class="form-group-add__actions">
      <LayoutSubmitFooter
        submit-test-id="agd-submit"
        cancel-test-id="agd-cancel"
        :text-submit-label="t('common.actions.create')"
        @cancel="cancel"
      />
    </div>
  </div>
</template>

<style scoped>
.form-group-add__row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group-add__row-4 {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
}

.form-group-add__row-4 :deep(.p-inputtext) {
  width: 100%;
}

/* ── Bottom row: toggles + actions ─────────────────────────── */
.form-group-add__bottom-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  align-items: center;
  gap: 1rem;
}

.form-group-add__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}

/* Narrow viewports: collapse multi-column rows to a single column. */
@media (max-width: 768px) {
  .form-group-add__row-2,
  .form-group-add__row-4,
  .form-group-add__bottom-row {
    grid-template-columns: 1fr;
  }
}
</style>
