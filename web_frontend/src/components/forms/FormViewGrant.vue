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
 * View-access form rendered inline above the views table.
 * component: a free-entry multi-select of user e-mails, pre-filled from the
 * view's existing *user* grants. Saving replaces the grant set
 * (`POST /view/grant/users`). Admin/owner-only — gated by the parent page so
 * legacy permission semantics are preserved exactly.
 */
import InputText from 'primevue/inputtext'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import Field from '@/components/elements/Field.vue'
import FieldMultiSelect from '@/components/elements/FieldMultiSelect.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { sameStringSet } from '@/lib/collections'
import { useUsersStore } from '@/stores/users'
import { useViewsStore } from '@/stores/views'

import type { ViewResponse } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    viewGrantPanel: {
      fields: {
        users: {
          label: 'Access Rights',
          hint: "If a user is not suggested, you can initialize it by entering it manually. Simply enter the user's e-mail address and confirm it with semicolon, tab or enter.",
          placeholder: '-',
        },
      },
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{ view: ViewResponse | null }>()
const emit = defineEmits<{ close: []; saved: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const viewsStore = useViewsStore()
const usersStore = useUsersStore()

// ── Reactive state ────────────────────────────────────────────
const selectedUsers = ref<string[]>([])

// ── Computed ──────────────────────────────────────────────────
// Known user e-mails offered as suggestions (free entry still allowed on top).
const userEmailOptions = computed(() => usersStore.users.map((user) => user.email))

// ── Watchers and lifecycle ────────────────────────────────────
// Shared form lifecycle: reset-on-open (pre-fill + user fetch), the
// unsaved-changes guard, cancel(), and the submit try/catch skeleton.
const { submit: submitForm, cancel } = useFormLifecycle({
  target: () => props.view,
  async reset(): Promise<void> {
    selectedUsers.value = (props.view?.grants ?? [])
      .filter((grant) => grant.is_user === true)
      .map((grant) => grant.username)
    try {
      await usersStore.fetchUsers()
    } catch {
      // Manual e-mail entry remains available after the shared error reporter runs
    }
  },

  // Dirty when the grant set differs (order-insensitively) from the view's
  // existing user grants.
  dirty: () =>
    !sameStringSet(
      selectedUsers.value,
      (props.view?.grants ?? [])
        .filter((grant) => grant.is_user === true)
        .map((grant) => grant.username),
    ),
  async submit(): Promise<boolean> {
    if (!props.view) return false
    const message = await viewsStore.grantUsers(props.view.id, selectedUsers.value)
    toast.success(message)
    emit('saved')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-view-grant__grid form-view-grant__grid--thirds">
    <Field v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText :id="id" :model-value="view?.scan_scope.group_name" disabled />
    </Field>
    <Field v-slot="{ id }" :label="t('common.labels.scope')" tooltip>
      <InputText :id="id" :model-value="view?.scan_scope.name" disabled />
    </Field>
    <Field v-slot="{ id }" :label="t('common.labels.view')" tooltip>
      <InputText :id="id" :model-value="view?.name" disabled />
    </Field>
  </div>

  <FieldMultiSelect
    v-model="selectedUsers"
    :label="t('viewGrantPanel.fields.users.label')"
    :help="t('viewGrantPanel.fields.users.hint')"
    tooltip
    :options="userEmailOptions"
    :placeholder="t('viewGrantPanel.fields.users.placeholder')"
  />
</template>

<style scoped>
.form-view-grant__grid {
  display: grid;
  gap: 0.75rem;
}

.form-view-grant__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-view-grant__grid :deep(.p-inputtext) {
  width: 100%;
}
</style>
