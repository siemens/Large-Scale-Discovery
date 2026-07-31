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
 * Creates or updates a custom scan scope.
 *
 * Create mode: loads groups list, shows the targets grid pre-seeded with one blank row.
 * Update mode: loads existing targets from API, populates the grid.
 */
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import AppMessage from '@/components/elements/AppMessage.vue'
import Field from '@/components/elements/Field.vue'
import FieldStepper from '@/components/elements/FieldStepper.vue'
import FieldToggle from '@/components/elements/FieldToggle.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import WidgetTargets from '@/components/widgets/WidgetTargets.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useZodForm } from '@/composables/useZodForm'
import { useGroupsStore } from '@/stores/groups'
import { useScopesStore } from '@/stores/scopes'
import { NETWORK_SIZE_MAX } from '@/types/scopes'

import type { ScopeResponse, ScopeTarget } from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    scopeAddCustomPanel: {
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
        targets: {
          label: 'Scan Input',
          hint: 'Insert scan targets and optional metadata. You can also copy the columns to Excel or paste them from Excel.',
        },
        cycles: {
          label: 'Scan continuously',
          hint: 'Initiate new scan cycle after all targets were scanned.',
        },
        ot: {
          label: 'Enable OT discovery (L2)',
          hint: 'Enable OT device discovery mode on OSI layer 2, applying softer scan timings and utilizing protocols such as PROFINET DCP, EtherCAT, LLDP, IPv6 NDP and mDNS. IP port scanning techniques within the subnet of the L2 can be applied additionally but are not the main focus. The scan agent can only target the directly attached L2 network segments and will pick its scan target accordingly.',
        },
      },
      warning:
        'You are responsible for diligent scan target selection. Scans may impact target systems. Only scan systems you are responsible and have a scan approval for!',
      synchronizing: 'Synchronization in progress, target editing is disabled.',
      syncInfo: {
        intro:
          'You can update, reduce or extend scan targets at any time. The backend will compare the new list with the previous one, based on the Scan Input column:',
        changed:
          'Changed inputs get updated. Scan states (waiting, running, completed) continue. Fresh results are stored with updated details.',
        vanished: 'Vanished inputs get removed. Existing results are kept.',
        added: 'New inputs get added.',
        readded:
          'If you remove a scan target and add the same one again later, it will behave like a new and distinct entry.',
      },
    },
  },
}

/** Child methods and state exposed by WidgetTargets to this form. */
interface WidgetTargetsHandle {
  setItems(items: ScopeTarget[]): void
  getItems(): ScopeTarget[]
  clearTargetScanData(input: string): void
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{
  scope?: ScopeResponse
}>()

const emit = defineEmits<{
  close: []
  saved: [message: string, warnings?: string[]]
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const scopesStore = useScopesStore()
const groupsStore = useGroupsStore()

// ── Reactive state ────────────────────────────────────────────
// Scalar form fields — initialized in reset().
const scopeName = ref(props.scope?.name ?? '')
const scopeCycles = ref(props.scope?.cycles ?? false)
const scopeOt = ref(props.scope?.scan_settings?.ot ?? false)

// Retention: stored as -1 = All, positive int = count.
const scopeRetention = ref<number | 'All'>(
  props.scope && props.scope.cycles_retention > 0 ? props.scope.cycles_retention : 'All',
)
const scopeNetworkSize = ref(initialSplitSize())

const selectedGroupId = ref<number | null>(null)

// Targets grid: ref to the child grid component; synchronizing hides the grid in
// update mode when the scope is being synced from an external source.
const targetsWidgetRef = ref<WidgetTargetsHandle | null>(null)
const isSynchronizing = ref(false)
const hasEditedTargets = ref(false)

// ── Computed ──────────────────────────────────────────────────
const isEditMode = computed(() => !!props.scope)

// Group dropdown options; populated by fetchGroups() inside reset() (create mode only).
const groupOptions = computed(() =>
  groupsStore.groups.map((group) => ({ label: group.name, value: group.id })),
)

// Group is required only in create mode; schema shape changes based on mode.
const schema = computed(() =>
  z.object({
    name: z.string().trim().min(3, t('scopeAddCustomPanel.fields.name.validation.min')),
    group: isEditMode.value
      ? z.number().nullable()
      : z.number({ message: t('scopeAddCustomPanel.fields.group.validation.required') }),
  }),
)

const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, () => ({
  name: scopeName.value,
  group: selectedGroupId.value,
}))

// ── Functions ─────────────────────────────────────────────────
function markTargetsEdited(): void {
  hasEditedTargets.value = true
}

// OT discovery always uses Nmap's maximum supported split size
function initialSplitSize(): number {
  return props.scope?.scan_settings?.ot
    ? NETWORK_SIZE_MAX
    : (props.scope?.split_size ?? NETWORK_SIZE_MAX)
}

// Dirty when the grid has been edited or any scalar field differs from its
// initial value (create defaults, or the scope being updated). Group is excluded
// because it auto-selects when only one group exists, which isn't user-meaningful.
function isFormDirty(): boolean {
  return (
    hasEditedTargets.value ||
    scopeName.value.trim() !== (props.scope?.name ?? '') ||
    scopeCycles.value !== (props.scope?.cycles ?? false) ||
    scopeOt.value !== (props.scope?.scan_settings?.ot ?? false) ||
    scopeRetention.value !==
      (props.scope && props.scope.cycles_retention > 0 ? props.scope.cycles_retention : 'All') ||
    scopeNetworkSize.value !== initialSplitSize()
  )
}

// Initialize the form from props.scope; runs on mount and on scope change. In
// create mode it loads the group list (auto-selecting a lone group); in update
// mode it loads the scope's existing targets into the grid.
async function resetForm(): Promise<void> {
  clearErrors()
  scopeName.value = props.scope?.name ?? ''
  scopeCycles.value = props.scope?.cycles ?? false
  scopeOt.value = props.scope?.scan_settings?.ot ?? false
  scopeRetention.value =
    props.scope && props.scope.cycles_retention > 0 ? props.scope.cycles_retention : 'All'
  scopeNetworkSize.value = initialSplitSize()
  selectedGroupId.value = null
  isSynchronizing.value = false
  hasEditedTargets.value = false

  if (!isEditMode.value) {
    await groupsStore.fetchGroups()

    // Auto-select the only group; narrow via the element itself so index access is defined
    const onlyGroup = groupsStore.groups.length === 1 ? groupsStore.groups[0] : undefined
    if (onlyGroup) {
      selectedGroupId.value = onlyGroup.id
    }
  } else if (props.scope) {
    const result = await scopesStore.fetchTargets(props.scope.id)
    isSynchronizing.value = result.synchronization
    if (!result.synchronization) {
      targetsWidgetRef.value?.setItems(result.targets)
    }
  }
}

async function resetTarget(input: string): Promise<void> {
  if (!props.scope) return
  try {
    await scopesStore.resetTarget(props.scope.id, input)
    targetsWidgetRef.value?.clearTargetScanData(input)
  } catch {
    // error already in store
  }
}

// Required by WidgetTargets to suppress paste-on-focus when a dialog is open.
// This form never opens a modal, so always returns false.
function isModalOpen(): boolean {
  return false
}

function sanitizeTargets(targets: ScopeTarget[]): ScopeTarget[] {
  return targets.filter((target) => target.input?.trim()).map(stripScanMetadata)
}

function stripScanMetadata(target: ScopeTarget): ScopeTarget {
  const sanitized = { ...target }
  delete sanitized.scan_started
  delete sanitized.scan_finished
  return sanitized
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

  let targets: ScopeTarget[] | undefined

  if (scopeOt.value) {
    const all: ScopeTarget[] = targetsWidgetRef.value?.getItems() ?? []
    const first = all[0] ?? { enabled: true }
    targets = [stripScanMetadata({ ...first, input: first.input ?? '' })]
  } else if (!isEditMode.value || hasEditedTargets.value) {
    // Create mode: always send targets. Edit mode: only send when changed.
    targets = sanitizeTargets(targetsWidgetRef.value?.getItems() ?? [])
  }

  // else: edit mode, no edits — omit `targets` so the backend leaves them unchanged.

  const payload = {
    name: scopeName.value.trim(),
    cycles: scopeCycles.value,
    ot: scopeOt.value,
    cycles_retention: retention,
    split_size: scopeNetworkSize.value,
    targets,
    ...scopeIdentifier,
  }

  const result = await scopesStore.submitCustom(payload)
  emit('saved', result.message, result.warnings)
  return true
}

// ── Watchers and lifecycle ────────────────────────────────────
// Reset split size immediately when OT discovery becomes active
watch(scopeOt, (isOt) => {
  if (isOt) scopeNetworkSize.value = NETWORK_SIZE_MAX
})

// Reset-on-open, dirty registration, and the submit/cancel handlers the template
// binds to. A store error is caught by useFormLifecycle and the form stays open.
const { submit: submitWithLifecycle, cancel } = useFormLifecycle({
  target: () => props.scope,
  reset: resetForm,
  dirty: isFormDirty,
  submit: submitForm,
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitWithLifecycle })
</script>

<template>
  <div class="form-scope-add-custom__grid form-scope-add-custom__grid--top">
    <Field
      v-if="!isEditMode"
      v-slot="{ id }"
      :label="t('common.labels.group')"
      required
      :help="t('scopeAddCustomPanel.fields.group.hint')"
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
        data-testid="sac-group"
      />
    </Field>
    <Field v-else v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText :id="id" :model-value="scope?.group_name" disabled />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('scopeAddCustomPanel.fields.name.label')"
      required
      :help="t('scopeAddCustomPanel.fields.name.hint')"
      tooltip
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="scopeName"
        :placeholder="t('scopeAddCustomPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('scopeAddCustomPanel.fields.name.label')"
        data-testid="sac-name"
      />
    </Field>

    <div class="form-scope-add-custom__stepper-pair">
      <FieldStepper
        v-model="scopeRetention"
        :label="t('common.labels.retention')"
        :help="t('scopeAddCustomPanel.fields.retention.hint')"
        tooltip
        test-id="sac-retention"
      />
      <FieldStepper
        v-model="scopeNetworkSize"
        kind="splitSize"
        :label="t('scopeAddCustomPanel.fields.splitSize.label')"
        :help="t('scopeAddCustomPanel.fields.splitSize.hint')"
        :disabled="scopeOt"
        tooltip
        test-id="sac-network-size"
      />
    </div>
  </div>

  <Field
    :label="t('scopeAddCustomPanel.fields.targets.label')"
    required
    class="form-scope-add-custom__targets-field"
    :help="t('scopeAddCustomPanel.fields.targets.hint')"
    tooltip
  >
    <AppMessage
      v-if="isSynchronizing"
      severity="warn"
      class="form-scope-add-custom__synchronizing"
      :message="t('scopeAddCustomPanel.synchronizing')"
    />
    <WidgetTargets
      ref="targetsWidgetRef"
      class="form-scope-add-custom__targets"
      :ot-mode="scopeOt"
      :readonly="isSynchronizing"
      :scope-id="isEditMode ? scope?.id : undefined"
      :is-modal-open="isModalOpen"
      @edited="markTargetsEdited"
      @reset-target="resetTarget"
    />
  </Field>

  <AppMessage severity="info" class="form-scope-add-custom__sync-info" data-testid="sac-sync-info">
    <p>
      {{ t('scopeAddCustomPanel.syncInfo.intro') }}
    </p>
    <ul>
      <li>{{ t('scopeAddCustomPanel.syncInfo.changed') }}</li>
      <li>{{ t('scopeAddCustomPanel.syncInfo.vanished') }}</li>
      <li>{{ t('scopeAddCustomPanel.syncInfo.added') }}</li>
      <li>{{ t('scopeAddCustomPanel.syncInfo.readded') }}</li>
    </ul>
  </AppMessage>
  <AppMessage severity="warn" data-testid="sac-responsibility">
    <p>{{ t('scopeAddCustomPanel.warning') }}</p>
  </AppMessage>

  <!-- Bottom row: Cycles + OT toggles each take a third, with the Cancel/Save
       actions filling the last third (matches the filter grid above). -->
  <div
    class="form-scope-add-custom__grid form-scope-add-custom__grid--thirds form-scope-add-custom__bottom-row"
  >
    <FieldToggle
      v-model="scopeCycles"
      :label="t('scopeAddCustomPanel.fields.cycles.label')"
      :help="t('scopeAddCustomPanel.fields.cycles.hint')"
      tooltip
    />
    <FieldToggle
      v-model="scopeOt"
      :label="t('scopeAddCustomPanel.fields.ot.label')"
      :help="t('scopeAddCustomPanel.fields.ot.hint')"
      :disabled="isEditMode"
      tooltip
      test-id="sac-ot"
    />
    <div class="form-scope-add-custom__actions">
      <LayoutSubmitFooter
        cancel-test-id="sac-cancel"
        submit-test-id="sac-submit"
        :text-submit-label="t('common.actions.save')"
        @cancel="cancel"
      />
    </div>
  </div>
</template>

<style scoped>
/* Separate the target editor from the preceding controls */
.form-scope-add-custom__targets-field {
  margin-bottom: 0.5rem;
}

/* Keep target-related notices visually attached to the editor */
.form-scope-add-custom__synchronizing {
  padding-top: 0.25rem;
  padding-bottom: 0;
}

/* Leave a small gap between the synchronization notice and target editor */
.form-scope-add-custom__targets {
  padding-top: 0.25rem;
}

/* Preserve space below the synchronization explanation */
.form-scope-add-custom__sync-info {
  padding-bottom: 0.5rem;
}

/* Keep the field grid consistent with the scope settings form */
.form-scope-add-custom__grid {
  display: grid;
  gap: 0.75rem;
}

/* Top row Group|Name|Retention on one line (legacy layout parity); controls
   fill their cell. */
.form-scope-add-custom__grid--top {
  grid-template-columns: 1fr 1fr 1fr;
}

.form-scope-add-custom__grid--thirds {
  grid-template-columns: repeat(3, 1fr);
}

/* Retention and network segmentation share the final top-row cell */
.form-scope-add-custom__stepper-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.form-scope-add-custom__grid :deep(.p-select),
.form-scope-add-custom__grid :deep(.p-inputtext) {
  width: 100%;
}

/* ── Info / warning messages ─────────────────────────────────── */
/* Separate the synchronization state from the targets widget above it. */
.form-scope-add-custom__synchronizing {
  padding-top: 0.75rem;
}

/* Align the final toggles and actions in equal-width columns */
.form-scope-add-custom__bottom-row {
  align-items: center;
}

/* Remove field rhythm that would shift the visible toggles above the action row center */
.form-scope-add-custom__bottom-row :deep(.field-toggle) {
  margin-bottom: 0;
}

.form-scope-add-custom__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}
</style>
