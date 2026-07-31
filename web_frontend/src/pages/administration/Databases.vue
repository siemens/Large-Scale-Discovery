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
 * Admin / Databases page.
 *
 * Parity port of the legacy `pages/admin/databases`: a grid of database servers
 * with inline cell editing (auto-save), an add-database wizard, and a
 * confirm-gated delete. Renders through the A4 <DataTable> wrapper (no
 * bespoke grid, no Tabulator). The save itself is the legacy "connection test" —
 * the backend validates the connection on `database/update` and answers with a
 * success/failure message we surface as a toast plus a loading state.
 *
 * Security: the list endpoint omits the password field — the cell always shows
 * the static mask. Editing the password cell sends the typed value; leaving it
 * blank and confirming is a no-op (DataTable skips unchanged cells). Non-password
 * field edits always send an empty password so the backend keeps the existing
 * credential. Nothing here logs credentials.
 */

import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { errorDetail } from '@/api/errors'
import DataTable from '@/components/datatable/DataTable.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import FormDatabaseAdd from '@/components/forms/FormDatabaseAdd.vue'
import LayoutForm from '@/components/layout/LayoutForm.vue'
import LayoutReveal from '@/components/layout/LayoutReveal.vue'
import LayoutRightAligned from '@/components/layout/LayoutRightAligned.vue'
import { useConfirmDelete } from '@/composables/useConfirmAction'
import { useGuardedPanel } from '@/composables/useGuardedPanel'
import { usePanelDeeplink } from '@/composables/usePanelDeeplink'
import { useSectionScroll } from '@/composables/useSectionScroll'
import { TOAST_LIFE, useToast } from '@/composables/useToast'
import { validateHost, validatePort } from '@/lib/validators'
import { useDatabasesStore } from '@/stores/databases'

import type { ColumnDef, InlineEditOptions } from '@/components/datatable/dataTableTypes'
import type { DatabaseResponse } from '@/types/databases'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    unsavedTitle: 'Discard unsaved changes?',
    unsavedMessage: 'You have unsaved changes. If you leave this page, they will be discarded.',
    admin: {
      databases: {
        help: 'This section lists database servers that can be used to store scan scopes on. The servers might belong to different customers. They can only be removed again if there are no remaining scan scopes stored on them.',
        empty: 'No database servers configured yet.',
        error: 'Could not load the databases.',
        columns: {
          name: 'DB Name',
          dialect: 'Dialect',
          host: 'Host',
          hostPublic: 'Host Public',
          args: 'Args',
          passwordHelp: 'Leave the password empty if you do not want to update it.',
        },
        fields: {
          name: {
            placeholder: 'Database Name',
            validation: {
              min: 'Name must be at least 3 characters.',
            },
          },
          dialect: {
            label: 'Dialect',
            placeholder: 'postgres',
            validation: {
              min: 'Dialect must be at least 3 characters.',
            },
          },
          admin: {
            label: 'Admin User',
            placeholder: 'postgres',
            validation: {
              required: 'Admin user is required.',
            },
          },
          password: {
            placeholder: '**********',
            help: 'Leave empty if you do not want to update it.',
            validation: {
              complex:
                'Password must be at least 10 characters and mix upper, lower, digits or symbols.',
            },
          },
          host: {
            label: 'Host',
            placeholder: 'Database Host',
          },
          hostPublic: {
            label: 'Host Public',
            placeholder: 'Database Host Public',
          },
          port: {
            placeholder: 'Database Port',
          },
          args: {
            label: 'Connection Args',
            placeholder: 'sslmode=verify-ca',
          },
        },
        actions: {
          add: {
            title: 'Add Database',
            success: 'Database saved.',
            error: 'Could not save the database.',
          },
          save: {
            success: 'Database saved.',
          },
          delete: {
            tooltip: 'Delete database',
            title: 'Delete Database',
            message:
              'This will remove a database server if no scan scopes are currently stored on it.',
            question: 'Are you sure you want to delete the database server',
            success: 'Database removed.',
            error: 'Could not remove the database.',
          },
        },
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()

const databasesStore = useDatabasesStore()
const { databases, isFetching } = storeToRefs(databasesStore)

// Keep the database-specific update payload local while DataTable owns the shared edit workflow
const inlineEdit: InlineEditOptions<DatabaseResponse> = {
  save: async ({ row: database, field, newValue }) => {
    // Password field: send only when the user actually typed something.
    // The list endpoint never returns the password, so the cell is always empty
    // before editing; the DataTable no-op guard (newValue === oldValue) already
    // prevents saves when the user opens and closes without typing. Any other
    // field edit sends an empty password so the backend keeps the existing credential.
    const password =
      field === 'password' && typeof newValue === 'string' && newValue.length > 0 ? newValue : ''
    await databasesStore.saveDatabase({
      id: database.id,
      name: database.name,
      dialect: database.dialect,
      host: database.host,
      host_public: database.host_public,
      port: Number(database.port),
      admin: database.admin,
      password,
      args: database.args ?? '',
    })
    return {
      summary: t('admin.databases.actions.save.success'),
      life: TOAST_LIFE.short,
    }
  },
  errorToast: (err) => ({
    summary: errorDetail(err) ?? '',
    life: TOAST_LIFE.error,
  }),
}

// One inline action panel is open at a time, rendered above the list. The
// guarded panel gates any switch behind the discard confirm when the open form
// has unsaved edits.
const panel = useGuardedPanel<'db-add'>()
const {
  activeKind: activePanelKind,
  confirmVisible: discardVisible,
  openCount: panelOpenCount,
  keyFor: panelKey,
  open: openPanel,
  close: closePanel,
  onConfirmDiscard,
} = panel

// Keep the revealed action section visible below the sticky navigation
const actionPanelEl = ref<HTMLElement | null>(null)
const formRef = ref<{
  cancel: () => void
  submit: () => Promise<void> | void
} | null>(null)
useSectionScroll(() => actionPanelEl.value, {
  activeSection: activePanelKind,
  openCount: panelOpenCount,
})

function setActionPanelElement(element: Element | { $el: Element } | null): void {
  // Resolve Vue component refs and direct DOM refs to the section element used for scrolling
  const panelElement = element instanceof Element ? element : element?.$el
  actionPanelEl.value = panelElement instanceof HTMLElement ? panelElement : null
}

// The add-database panel has no target, so it encodes as just `#add`; a refresh
// or shared link with that hash re-opens the blank form.
const deeplink = usePanelDeeplink(panel, {
  encode: () => 'add',
  restore: (raw) => {
    if (raw === 'add') openAdd()
  },
})

const {
  deleteConfirmVisible,
  pendingDeleteTarget: databaseToDelete,
  openDeleteConfirm,
  confirmDelete: confirmDatabaseDelete,
} = useConfirmDelete<DatabaseResponse>({
  remove: (row) => databasesStore.deleteDatabase(row.id),
  onSuccess: () => {
    toast.success(t('admin.databases.actions.delete.success'))
  },
})

// ── Functions and handlers ────────────────────────────────────
function nonEmpty(value: unknown, message: string): true | string {
  return String(value).trim().length > 0 || message
}

// Column definitions for the databases grid.
const columns = computed<ColumnDef<DatabaseResponse>[]>(() => [
  {
    field: 'name',
    header: t('common.labels.name'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    validator: (v) => nonEmpty(v, t('admin.databases.fields.name.validation.min')),
    width: '12rem',
    sensitive: true,
  },
  {
    field: 'dialect',
    header: t('admin.databases.columns.dialect'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    validator: (v) => nonEmpty(v, t('admin.databases.fields.dialect.validation.min')),
  },
  {
    field: 'host',
    header: t('admin.databases.columns.host'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    validator: (v) => validateHost(String(v)) ?? true,
    sensitive: true,
  },
  {
    field: 'host_public',
    header: t('admin.databases.columns.hostPublic'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    validator: (v) => validateHost(String(v)) ?? true,
    sensitive: true,
  },
  {
    field: 'port',
    header: t('common.labels.port'),
    editor: 'number',
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    width: '5rem',
    validator: (v) => validatePort(String(v)) ?? true,
  },
  {
    field: 'args',
    header: t('admin.databases.columns.args'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    sensitive: true,
  },
  {
    field: 'admin',
    header: t('common.labels.admin'),
    editor: 'input',
    filterable: true,
    filterType: 'text',
    width: '7rem',
    validator: (v) => nonEmpty(v, t('admin.databases.fields.admin.validation.required')),
    sensitive: true,
  },
  {
    field: 'password',
    header: t('common.labels.password'),
    formatter: 'password',
    editor: 'input',
    editorInputType: 'password',
    sortable: false,

    // Header filter: decorative only — the list endpoint omits the credential,
    // so the cell holds only a mask and there is nothing to filter against.
    // `noop` keeps the input visible without zeroing the grid.
    filterable: true,
    filterType: 'noop',
    width: '6rem',
    headerTooltip: t('admin.databases.columns.passwordHelp'),
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

function openAdd(): void {
  openPanel('db-add')
}

function submitForm(): Promise<void> | void {
  return formRef.value?.submit()
}

function cancelForm(): void {
  formRef.value?.cancel()
}

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(() => {
  // Initial load; a fetch failure surfaces as a toast and leaves the grid empty.
  // Restore the hash-encoded panel afterwards (add has no target to resolve, but
  // this keeps the pattern uniform across pages).
  databasesStore
    .fetchDatabases()
    .catch(() => {
      // The request and parsing layers already displayed the appropriate error
    })
    .finally(() => {
      deeplink.restoreFromHash()
    })
})
</script>

<template>
  <!-- Add-database form opens as its own section ABOVE the page header, so it
         reads as the top-most section and the page scrolls up to reveal it. -->
  <LayoutReveal>
    <LayoutForm
      v-if="activePanelKind === 'db-add'"
      :key="panelKey('db-add')"
      :ref="setActionPanelElement"
      data-testid="databases-action-panel"
      class="page-section"
      :title="t('admin.databases.actions.add.title')"
      :text-submit-label="t('common.actions.create')"
      :submit-handler="submitForm"
      submit-test-id="db-add-btn-submit"
      @cancel="cancelForm"
    >
      <FormDatabaseAdd ref="formRef" @saved="databasesStore.fetchDatabases()" @close="closePanel" />
    </LayoutForm>
  </LayoutReveal>

  <section class="page-section databases-page">
    <SectionHeader :title="t('common.labels.databases')" :help-text="t('admin.databases.help')" />

    <DataTable
      class="databases-page__data-table-container"
      :rows="databases"
      :columns="columns"
      :loading="isFetching"
      column-borders
      fit-columns
      edit-mode="cell"
      :inline-edit="inlineEdit"
      :empty-message="t('admin.databases.empty')"
    >
      <template #actions="{ data: database }">
        <div class="data-table__actions">
          <Button
            severity="danger"
            icon="pi pi-trash"
            class="data-table__button"
            :data-testid="`db-action-delete-${database.id}`"
            :aria-label="t('admin.databases.actions.delete.tooltip')"
            @click="openDeleteConfirm(database)"
          />
        </div>
      </template>
    </DataTable>

    <LayoutRightAligned>
      <SegmentedButton
        data-testid="db-add-open"
        icon="pi pi-plus"
        :label="t('common.labels.database')"
        @click="openAdd"
      />
    </LayoutRightAligned>

    <!-- Prompts before navigating away while the Add-database panel reports
         unsaved changes. Inline cell edits commit immediately, so only the
         panel feeds the guard. -->
    <DialogConfirm
      v-model:visible="deleteConfirmVisible"
      :title="t('admin.databases.actions.delete.title')"
      icon="pi pi-trash"
      :text-confirm-label="t('common.actions.delete')"
      :confirm-word="databaseToDelete?.name"
      :message="`${t('admin.databases.actions.delete.question')} '${databaseToDelete?.name}'?`"
      :secondary-message="t('admin.databases.actions.delete.message')"
      @confirm="confirmDatabaseDelete"
    />

    <!-- Discard-changes gate shown when switching panels while a form is dirty -->
    <DialogConfirm
      v-model:visible="discardVisible"
      :title="t('unsavedTitle')"
      :text-confirm-label="t('common.actions.discard')"
      :message="t('unsavedMessage')"
      @confirm="onConfirmDiscard"
    />
  </section>
</template>

<style scoped>
.databases-page__data-table-container {
  margin-bottom: 1.5rem;
}
</style>
