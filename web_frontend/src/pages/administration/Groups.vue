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
 * Admin / Groups page. Migrates the legacy Knockout + Tabulator groups page to
 * Vue 3 + PrimeVue, with inline cell editing, an inline add-group panel, an
 * inline owners panel, and a confirm-gated delete rendered through the A4
 * <DataTable>.
 */

import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable from '@/components/datatable/DataTable.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import FormGroupAdd from '@/components/forms/FormGroupAdd.vue'
import FormGroupOwners from '@/components/forms/FormGroupOwners.vue'
import LayoutForm from '@/components/layout/LayoutForm.vue'
import LayoutReveal from '@/components/layout/LayoutReveal.vue'
import LayoutRightAligned from '@/components/layout/LayoutRightAligned.vue'
import { useConfirmDelete } from '@/composables/useConfirmAction'
import { useGuardedPanel } from '@/composables/useGuardedPanel'
import { usePanelDeeplink } from '@/composables/usePanelDeeplink'
import { useSectionScroll } from '@/composables/useSectionScroll'
import { TOAST_LIFE, useToast } from '@/composables/useToast'
import { useDatabasesStore } from '@/stores/databases'
import { useGroupsStore } from '@/stores/groups'

import type { ColumnDef, InlineEditOptions } from '@/components/datatable/dataTableTypes'
import type { GroupResponse } from '@/types/groups'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    unsavedTitle: 'Discard unsaved changes?',
    unsavedMessage: 'You have unsaved changes. If you leave this page, they will be discarded.',
    admin: {
      groups: {
        panels: {
          create: 'Create Group',
          owners: 'Group Owners',
        },
        help: 'This section lists existing groups and allows to create new ones. Each group can create its own scan scopes and views according to configured limits.',
        error: 'Could not load groups.',
        columns: {
          allowCustom: 'Custom Input',
          allowNetwork: 'Network Imports',
          allowAsset: 'Asset Imports',
        },
        fields: {
          name: {
            validation: {
              min: 'Group name must be at least 3 characters.',
            },
          },
          maxScopes: {
            validation: {
              invalid: 'Enter a number ≥ -1 or * for unlimited.',
            },
          },
          maxViews: {
            validation: {
              invalid: 'Enter a number ≥ -1 or * for unlimited.',
            },
          },
          maxTargets: {
            validation: {
              invalid: 'Enter a number ≥ -1 or * for unlimited.',
            },
          },
          maxOwners: {
            validation: {
              invalid: 'Enter a number ≥ -1 or * for unlimited.',
            },
          },
          allowCustom: {
            help: 'Allow users to create scan scopes with custom input lists.',
          },
          allowNetwork: {
            help: 'Allow users to create scan scopes with inputs imported from the network inventory.',
          },
          allowAsset: {
            help: 'Allow users to create scan scopes with inputs imported from the asset inventory.',
          },
        },
        actions: {
          save: {
            success: 'Group updated.',
            error: 'Could not update group.',
          },
          owners: {
            tooltip: 'Manage group owners',
          },
          delete: {
            tooltip: 'Delete group',
            title: 'Delete Group',
            message: 'This will remove all associated SCAN SCOPES, VIEWS and ACCESS RIGHTS.',
            question: 'Are you sure you want to delete the group',
            success: 'Group deleted.',
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
const groupsStore = useGroupsStore()
const databasesStore = useDatabasesStore()

const { groups, isFetching } = storeToRefs(groupsStore)
const { databases } = storeToRefs(databasesStore)

// One inline action panel is open at a time (add / owners), rendered above the
// list. The guarded panel gates any switch behind the discard confirm when the
// open form has unsaved edits.
type GroupPanelKind = 'group-add' | 'group-owners'
const panel = useGuardedPanel<GroupPanelKind, GroupResponse>()
const {
  activeKind: activePanelKind,
  target: selectedGroup,
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
  target: selectedGroup,
  openCount: panelOpenCount,
})

function setActionPanelElement(element: Element | { $el: Element } | null): void {
  // Resolve Vue component refs and direct DOM refs to the section element used for scrolling
  const panelElement = element instanceof Element ? element : element?.$el
  actionPanelEl.value = panelElement instanceof HTMLElement ? panelElement : null
}

// Keep the group-specific update payload local while DataTable owns the shared edit workflow
const inlineEdit: InlineEditOptions<GroupResponse> = {
  save: async ({ row: group }) => {
    await groupsStore.updateGroup({
      id: group.id,
      name: group.name,
      db_server_id: group.db_server_id,
      max_scopes: group.max_scopes,
      max_views: group.max_views,
      max_targets: group.max_targets,
      max_owners: group.max_owners,
      allow_custom: group.allow_custom,
      allow_network: group.allow_network,
      allow_asset: group.allow_asset,
    })
    return {
      summary: t('admin.groups.actions.save.success'),
      life: TOAST_LIFE.success,
    }
  },
  errorToast: () => ({
    summary: t('admin.groups.actions.save.error'),
    life: TOAST_LIFE.error,
  }),
}

// Mirror the open panel into the hash (`#add`, `#owners=<id>`) so a refresh or
// shared link re-opens it. The owners panel carries the group id; add carries
// only the kind. Restore resolves the id against the loaded group list.
const deeplink = usePanelDeeplink(panel, {
  encode: (kind, group) => (kind === 'group-add' ? 'add' : group ? `owners=${group.id}` : 'owners'),
  restore: (raw) => {
    const [token, id] = raw.split('=')
    if (token === 'add') {
      openAdd()
      return
    }
    if (token === 'owners' && id) {
      const group = groups.value.find((candidateGroup) => String(candidateGroup.id) === id)
      if (group) openOwners(group)
    }
  },
})

const {
  deleteConfirmVisible,
  pendingDeleteTarget: groupToDelete,
  openDeleteConfirm,
  confirmDelete: confirmGroupDelete,
} = useConfirmDelete<GroupResponse>({
  remove: (group) => groupsStore.deleteGroup(group.id),
  onSuccess: () => {
    toast.success(t('admin.groups.actions.delete.success'))
  },
})

// ── Computed ──────────────────────────────────────────────────
// Resolve database names without repeatedly searching the full collection
const databaseNameById = computed<Record<number, string>>(() => {
  const map: Record<number, string> = {}
  for (const database of databases.value) map[database.id] = database.name
  return map
})

// Label/value pairs for the database select editor
const databaseEditorOptions = computed(() =>
  databases.value.map((database) => ({
    label: database.name,
    value: database.id,
  })),
)

// Column definitions (legacy order + semantics).
const columns = computed<ColumnDef<GroupResponse>[]>(() => [
  {
    field: 'name',
    header: t('common.labels.name'),
    filterable: true,
    filterType: 'text',
    sortable: true,
    editor: 'input',
    validator: (value: unknown) => {
      if (typeof value !== 'string' || value.trim().length < 3) {
        return t('admin.groups.fields.name.validation.min')
      }
      return true
    },
    width: '12rem',
    sensitive: true,
  },
  {
    field: 'created',
    header: t('common.labels.created'),
    formatter: 'datetime',
    filterable: true,
    sortable: true,
    width: '11.4rem',
  },
  {
    field: 'max_scopes',
    header: t('common.labels.scopes'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    sortable: true,
    editor: 'number',
    validator: (value: unknown) => {
      const parsedLimit = Number(value)
      if (!Number.isInteger(parsedLimit) || parsedLimit < -1)
        return t('admin.groups.fields.maxScopes.validation.invalid')
      return true
    },
    width: '4rem',
  },
  {
    field: 'max_views',
    header: t('common.labels.views'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    sortable: true,
    editor: 'number',
    validator: (value: unknown) => {
      const parsedLimit = Number(value)
      if (!Number.isInteger(parsedLimit) || parsedLimit < -1)
        return t('admin.groups.fields.maxViews.validation.invalid')
      return true
    },
    width: '3.7rem',
  },
  {
    field: 'max_targets',
    header: t('common.labels.targets'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    sortable: true,
    editor: 'number',
    validator: (value: unknown) => {
      const parsedLimit = Number(value)
      if (!Number.isInteger(parsedLimit) || parsedLimit < -1)
        return t('admin.groups.fields.maxTargets.validation.invalid')
      return true
    },
    width: '4rem',
  },
  {
    field: 'max_owners',
    header: t('common.labels.owners'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    sortable: true,
    editor: 'number',
    validator: (value: unknown) => {
      const parsedLimit = Number(value)
      if (!Number.isInteger(parsedLimit) || parsedLimit < -1)
        return t('admin.groups.fields.maxOwners.validation.invalid')
      return true
    },
    width: '4rem',
  },
  {
    field: 'allow_custom',
    header: t('admin.groups.columns.allowCustom'),
    filterable: true,
    filterType: 'tristate',
    sortable: true,
    editor: 'checkbox',
    width: '4rem',
    headerTooltip: t('admin.groups.fields.allowCustom.help'),
  },
  {
    field: 'allow_network',
    header: t('admin.groups.columns.allowNetwork'),
    filterable: true,
    filterType: 'tristate',
    sortable: true,
    editor: 'checkbox',
    width: '4rem',
    headerTooltip: t('admin.groups.fields.allowNetwork.help'),
  },
  {
    field: 'allow_asset',
    header: t('admin.groups.columns.allowAsset'),
    filterable: true,
    filterType: 'tristate',
    sortable: true,
    editor: 'checkbox',
    width: '4rem',
    headerTooltip: t('admin.groups.fields.allowAsset.help'),
  },
  {
    field: 'db_server_id',
    header: t('common.labels.database'),
    filterable: true,
    filterType: 'text',
    sortable: true,
    formatter: 'lookup',
    lookupMap: databaseNameById.value,
    editor: 'select',
    editorOptions: databaseEditorOptions.value,
    width: '10rem',
  },
  {
    field: 'actions',
    header: t('common.labels.actions'),
    formatter: 'actions',
    sortable: false,
    width: '5.5rem',
    resizable: false,
  },
])

// ── Functions and handlers ────────────────────────────────────
function openAdd(): void {
  openPanel('group-add')
}

function openOwners(group: GroupResponse): void {
  openPanel('group-owners', group)
}

function submitForm(): Promise<void> | void {
  return formRef.value?.submit()
}

function cancelForm(): void {
  formRef.value?.cancel()
}

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(async () => {
  try {
    // Groups and databases load together; the database list feeds the lookup column
    await Promise.all([groupsStore.fetchGroups(), databasesStore.fetchDatabases()])
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }

  // Re-open the panel encoded in the URL hash now that groups are available
  deeplink.restoreFromHash()
})
</script>

<template>
  <!-- Action panels — open as their own section ABOVE the page header (one at a
         time), so an open form reads as the top-most section and the page scrolls
         up to reveal its header. -->
  <LayoutReveal>
    <LayoutForm
      v-if="activePanelKind === 'group-add'"
      :key="panelKey('group-add')"
      :ref="setActionPanelElement"
      data-testid="groups-action-panel"
      class="page-section"
      :title="t('admin.groups.panels.create')"
      :footer="false"
      :text-submit-label="t('common.actions.create')"
      :submit-handler="submitForm"
    >
      <FormGroupAdd ref="formRef" @created="groupsStore.fetchGroups()" @close="closePanel" />
    </LayoutForm>
    <LayoutForm
      v-else-if="activePanelKind === 'group-owners'"
      :key="panelKey('group-owners', selectedGroup?.id)"
      :ref="setActionPanelElement"
      data-testid="groups-action-panel"
      class="page-section"
      :title="t('admin.groups.panels.owners')"
      :text-submit-label="t('common.actions.submit')"
      :submit-handler="submitForm"
      submit-test-id="od-submit"
      @cancel="cancelForm"
    >
      <FormGroupOwners
        ref="formRef"
        :group="selectedGroup"
        @saved="groupsStore.fetchGroups()"
        @close="closePanel"
      />
    </LayoutForm>
  </LayoutReveal>

  <section class="page-section groups-page">
    <SectionHeader
      :title="t('common.labels.groups')"
      :help-text="t('admin.groups.help')"
      test-id="agp-title"
    />

    <DataTable
      class="groups-page__data-table-container"
      :rows="groups"
      :columns="columns"
      column-borders
      fit-columns
      :loading="isFetching"
      edit-mode="cell"
      :inline-edit="inlineEdit"
    >
      <template #actions="{ data: group }">
        <div class="data-table__actions">
          <Button
            icon="pi pi-users"
            class="data-table__button"
            :data-testid="`agp-owners-btn-${group.id}`"
            :aria-label="t('admin.groups.actions.owners.tooltip')"
            @click="openOwners(group)"
          />
          <Button
            icon="pi pi-trash"
            severity="danger"
            class="data-table__button"
            :data-testid="`agp-delete-btn-${group.id}`"
            :aria-label="t('admin.groups.actions.delete.tooltip')"
            @click="openDeleteConfirm(group)"
          />
        </div>
      </template>
    </DataTable>

    <LayoutRightAligned>
      <SegmentedButton
        data-testid="agp-add-btn"
        :label="t('common.labels.group')"
        icon="pi pi-plus"
        :aria-label="t('common.labels.group')"
        @click="openAdd"
      />
    </LayoutRightAligned>

    <!-- Prompts before navigating away while the Add-group / Owners panel
         reports unsaved changes. Inline cell edits commit immediately, so only
         the panels feed the guard. -->
    <DialogConfirm
      v-model:visible="deleteConfirmVisible"
      :title="t('admin.groups.actions.delete.title')"
      :confirm-word="groupToDelete?.name"
      :text-confirm-label="t('common.actions.confirm')"
      icon="pi pi-trash"
      :message="`${t('admin.groups.actions.delete.question')} '${groupToDelete?.name}'?`"
      :secondary-message="t('admin.groups.actions.delete.message')"
      @confirm="confirmGroupDelete"
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
.groups-page__data-table-container {
  margin-bottom: 1.5rem;
}
</style>
