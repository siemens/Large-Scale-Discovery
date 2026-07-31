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
 * Group-owner form with a free-entry multi-select of user emails,
 * pre-filled from the group's existing ownerships and saved as a replacement
 * set. Rendered inline above the groups table, not
 * as a modal overlay.
 */

import { storeToRefs } from 'pinia'
import InputText from 'primevue/inputtext'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import Field from '@/components/elements/Field.vue'
import FieldMultiSelect from '@/components/elements/FieldMultiSelect.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { sameStringSet } from '@/lib/collections'
import { useGroupsStore } from '@/stores/groups'
import { useUsersStore } from '@/stores/users'

import type { GroupResponse } from '@/types/groups'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    groupOwnersPanel: {
      success: 'Owners updated.',
      fields: {
        group: {
          placeholder: 'Group Name',
        },
        owners: {
          help: "Owners are allowed to create and edit scans within the group's boundaries. If a user is not suggested, you can initialize it by entering it manually. Simply enter the user's e-mail address and confirm it with semicolon, tab or enter.",
          placeholder: '-',
        },
      },
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{ group: GroupResponse | null }>()
const emit = defineEmits<{ close: []; saved: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const groupsStore = useGroupsStore()
const usersStore = useUsersStore()
const { users } = storeToRefs(usersStore)

// ── Reactive state ────────────────────────────────────────────
const selectedOwners = ref<string[]>([])

// ── Computed ──────────────────────────────────────────────────
// Known user emails offered as suggestions (arbitrary emails still addable).
const userEmailOptions = computed(() => users.value.map((user) => user.email))

// ── Watchers and lifecycle ────────────────────────────────────
// Shared form lifecycle: reset-on-mount and on target change (pre-fill + user
// fetch), the unsaved-changes guard, cancel(), and the submit try/catch skeleton.
const { submit: submitForm, cancel } = useFormLifecycle({
  target: () => props.group,
  reset(): void {
    if (!props.group) return
    selectedOwners.value = (props.group.ownerships ?? []).map((ownership) => ownership.user.email)
    if (users.value.length === 0) void usersStore.fetchUsers()
  },

  // Dirty when the owner set differs (order-insensitively) from the group's
  // existing ownerships.
  dirty: () =>
    !sameStringSet(
      selectedOwners.value,
      (props.group?.ownerships ?? []).map((ownership) => ownership.user.email),
    ),
  async submit(): Promise<boolean> {
    if (!props.group) return false
    await groupsStore.assignOwners({
      id: props.group.id,
      owners: selectedOwners.value,
    })
    toast.success(t('groupOwnersPanel.success'))
    emit('saved')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-group-owners__row">
    <Field v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText
        :id="id"
        data-testid="od-group-name"
        :placeholder="t('groupOwnersPanel.fields.group.placeholder')"
        :model-value="group?.name ?? ''"
        disabled
        :aria-label="t('common.labels.group')"
      />
    </Field>

    <FieldMultiSelect
      v-model="selectedOwners"
      :label="t('common.labels.owners')"
      :help="t('groupOwnersPanel.fields.owners.help')"
      tooltip
      :options="userEmailOptions"
      allow-new
      :placeholder="t('groupOwnersPanel.fields.owners.placeholder')"
      class="form-group-owners__owners-field"
    />
  </div>
</template>

<style scoped>
/* Group (¼) beside Owners (¾) on one row (legacy four-wide / twelve-wide split).
   minmax(0,…) lets the chip field shrink so the split stays 1:3. */
.form-group-owners__row {
  display: grid;
  grid-template-columns: 1fr 3fr;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  align-items: start;
}

.form-group-owners__row :deep(.p-inputtext),
.form-group-owners__row :deep(.p-autocomplete) {
  width: 100%;
}
</style>
