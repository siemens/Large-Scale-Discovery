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
 * Admin / Users page. Migrates the legacy
 * `web_frontend/src/pages/admin/users/*` (Knockout + Tabulator) to Vue 3 +
 * PrimeVue, rendered exclusively through the A4 `<DataTable>` wrapper and
 * the A4 `<DialogConfirm>`.
 *
 * Parity contract with the legacy page:
 *   - same column set + order (e-mail, active, admin, demo, name, surname,
 *     company, last-login, delete);
 *   - inline cell editing committed on blur, reverted when the API rejects
 *     (Tabulator `restoreOldValue`); `last_login` is read-only;
 *   - newest users first (legacy reverses the list);
 *   - delete is gated behind a confirmation dialog and toasts on success.
 *
 * Auth/role parity: the legacy ViewModel redirected anonymous users to /login
 * and non-admins to /home before rendering; we do the same on mount.
 */

import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

import { errorDetail } from '@/api/errors'
import DataTable from '@/components/datatable/DataTable.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import { useConfirmDelete } from '@/composables/useConfirmAction'
import { TOAST_LIFE, useToast } from '@/composables/useToast'
import { useUsersStore } from '@/stores/users'

import type { ColumnDef, InlineEditOptions } from '@/components/datatable/dataTableTypes'
import type { UserResponse } from '@/types/users'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    admin: {
      users: {
        help: 'This section lists users that have previously logged in or were granted access to scan scopes. The system creates required users automatically. However, they can be removed or disabled by administrators.',
        error: 'Could not load users.',
        columns: {
          active: 'Active',
          demo: 'Demo',
          lastLogin: 'Last Login',
        },
        actions: {
          save: {
            success: 'User updated.',
          },
          delete: {
            tooltip: 'Delete user {email}',
            title: 'Delete User',
            message: 'This will remove all access rights and privileges.',
            question: 'Are you sure you want to delete the user',
            success: 'User deleted.',
            error: 'Could not delete the user.',
          },
        },
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()

const usersStore = useUsersStore()
const { users, isFetching } = storeToRefs(usersStore)
const { fetchUsers, updateUser, deleteUser } = usersStore

// Inline edits use the shared persist/toast/revert flow. E-mail is deliberately
// absent from the editable fields because it is the user's stable identifier.
const inlineEdit: InlineEditOptions<UserResponse> = {
  save: async (edit) => {
    await updateUser(edit.row)
    return {
      summary: t('admin.users.actions.save.success'),
      life: TOAST_LIFE.success,
    }
  },
  errorToast: (err) => ({
    summary: errorDetail(err) ?? '',
    life: TOAST_LIFE.error,
  }),
}

const {
  deleteConfirmVisible,
  pendingDeleteTarget: userToDelete,
  openDeleteConfirm,
  confirmDelete: confirmUserDelete,
} = useConfirmDelete<UserResponse, string>({
  remove: (user) => deleteUser(user.id),
  onSuccess: (message) => {
    toast.success(message || t('admin.users.actions.delete.success'))
  },
})

// ── Computed ──────────────────────────────────────────────────
// Column definitions for the users grid; computed so i18n keys resolve reactively.
const columns = computed<ColumnDef<UserResponse>[]>(() => [
  // The update endpoint identifies users by id and does not accept e-mail changes.
  {
    field: 'email',
    header: t('common.labels.email'),
    filterable: true,
    filterType: 'text',
    width: '12rem',
    sensitive: true,
  },
  {
    field: 'active',
    header: t('admin.users.columns.active'),
    editor: 'checkbox',
    filterable: true,
    filterType: 'tristate',
    width: '5.5rem',
  },
  {
    field: 'admin',
    header: t('common.labels.admin'),
    editor: 'checkbox',
    filterable: true,
    filterType: 'tristate',
    width: '5.5rem',
    sensitive: true,
  },
  {
    field: 'demo',
    header: t('admin.users.columns.demo'),
    editor: 'checkbox',
    filterable: true,
    filterType: 'tristate',
    width: '5.5rem',
  },
  {
    field: 'name',
    header: t('common.labels.name'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    sensitive: true,
  },
  {
    field: 'surname',
    header: t('common.labels.surname'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    sensitive: true,
  },
  {
    field: 'company',
    header: t('common.labels.company'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    sensitive: true,
  },
  {
    field: 'last_login',
    header: t('admin.users.columns.lastLogin'),
    formatter: 'datetime',
    filterable: true,
    width: '11.4rem',
  },
  {
    field: 'actions',
    header: t('common.labels.actions'),
    formatter: 'actions',
    sortable: false,
    width: '3.5rem',
    resizable: false,
  },
])

// ── Watchers and lifecycle ────────────────────────────────────
// Auth + admin gating is enforced centrally by the router (meta.admin guard).
onMounted(async () => {
  try {
    await fetchUsers()
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }
})
</script>

<template>
  <section class="page-section users-page">
    <SectionHeader
      class="users-page__title"
      :title="t('common.labels.users')"
      :help-text="t('admin.users.help')"
      test-id="admin-users-help"
    />

    <DataTable
      :rows="users"
      :columns="columns"
      :loading="isFetching"
      column-borders
      fit-columns
      edit-mode="cell"
      :inline-edit="inlineEdit"
    >
      <template #actions="{ data: user }">
        <div class="data-table__actions">
          <Button
            v-tooltip.top="t('admin.users.actions.delete.tooltip', { email: user.email })"
            severity="danger"
            icon="pi pi-trash"
            class="data-table__button"
            :data-testid="`adt-action-delete-${user.id}`"
            :aria-label="t('admin.users.actions.delete.tooltip', { email: user.email })"
            @click="openDeleteConfirm(user)"
          />
        </div>
      </template>
    </DataTable>

    <DialogConfirm
      v-model:visible="deleteConfirmVisible"
      :title="t('admin.users.actions.delete.title')"
      icon="pi pi-trash"
      :message="`${t('admin.users.actions.delete.question')} '${userToDelete?.email}'?`"
      :secondary-message="t('admin.users.actions.delete.message')"
      :text-confirm-label="t('common.actions.confirm')"
      @confirm="confirmUserDelete"
    />
  </section>
</template>

<style scoped></style>
