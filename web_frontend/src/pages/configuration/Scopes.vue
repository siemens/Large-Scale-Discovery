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
 * ScopesPage — B4 Configuration / Scopes page.
 *
 * Routes: /configuration/scopes.
 * Layout: scopes list + action panel below (create/edit/settings form).
 */
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Popover from 'primevue/popover'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import DataGrid from '@/components/elements/DataGrid.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import ProgressBar3 from '@/components/elements/ProgressBar3.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import SortableTh from '@/components/elements/SortableTh.vue'
import FormScopeAddAssets from '@/components/forms/FormScopeAddAssets.vue'
import FormScopeAddCustom from '@/components/forms/FormScopeAddCustom.vue'
import FormScopeAddNetworks from '@/components/forms/FormScopeAddNetworks.vue'
import FormScopeSettings from '@/components/forms/FormScopeSettings.vue'
import LayoutForm from '@/components/layout/LayoutForm.vue'
import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import LayoutReveal from '@/components/layout/LayoutReveal.vue'
import LayoutRightAligned from '@/components/layout/LayoutRightAligned.vue'
import { useConfirmAction } from '@/composables/useConfirmAction'
import { useKeyedGroupSort } from '@/composables/useGroupSort'
import { useGuardedPanel } from '@/composables/useGuardedPanel'
import { usePanelDeeplink } from '@/composables/usePanelDeeplink'
import { useSectionScroll } from '@/composables/useSectionScroll'
import { useToast } from '@/composables/useToast'
import { daysAgo, formatDate, formatDateTime, isZeroDate } from '@/i18n/format'
import { groupByKey, toCycleProgress } from '@/lib/collections'
import { useScopesStore } from '@/stores/scopes'

import type { CycleProgressTuple } from '@/lib/collections'
import type { ScopeResponse } from '@/types/scopes'
import type { MenuItem } from 'primevue/menuitem'

// Define every scope creation option and the generic form family that handles it
const scopeTypes: Record<string, { type: 'custom' | 'assets' | 'networks'; label: string }> = {
  custom: { type: 'custom', label: 'Custom' },
  assets: { type: 'assets', label: 'Assets' },
  networks: { type: 'networks', label: 'Networks' },
}

// Publish the registered type names through the local i18n catalog
const scopeTypeLabels = Object.fromEntries(
  Object.entries(scopeTypes).map(([scopeType, definition]) => [scopeType, definition.label]),
)

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    unsavedTitle: 'Discard unsaved changes?',
    unsavedMessage: 'You have unsaved changes. If you leave this page, they will be discarded.',
    configuration: {
      scopes: {
        help: 'This section lists scan scopes you are allowed to administer. A scan scope is a dedicated database that can contain arbitrary scan targets. You must run one or more scan agents to process your scan scope.',
        loading: 'Loading scopes...',
        empty: 'No scopes configured.',
        error: 'Could not load scopes.',
        panels: {
          create: 'Create {type} Scope',
          update: 'Update {type} Scope',
          settings: 'Scan Settings',
        },
        columns: {
          config: 'Config',
          cycleProgress: 'Cycle Progress',
        },
        configIndicators: {
          continuous: 'Continuous scan',
          retentionAll: 'Retain all cycles',
          retentionLastN: 'Retain last {count} cycle | Retain last {count} cycles',
          ot: 'OT network',
          sync: 'Synchronized',
        },
        cyclePopup: {
          ips: '{count} IPs',
          notSynced: 'Targets not synchronized',
          cycle: 'Cycle {n}',
          queued: 'Queued:',
        },
        actions: {
          createScope: {
            button: '{type} Scope',
          },
          createMore: {
            button: 'More scope types',
          },
          targets: {
            tooltip: 'Scan Targets',
          },
          settings: {
            tooltip: 'Scan Settings',
          },
          rescanFailed: {
            tooltip: 'Rescan Failed',
            title: 'Rescan Failed Targets',
            message:
              'This will reset all failed scan targets within the current scan cycle. They will be scanned again starting with a full discovery port scan.',
            question: 'Are you sure you want to rescan all failed scan targets of',
          },
          newCycle: {
            tooltip: 'New Cycle',
            title: 'New Scan Cycle',
            message:
              'This will initialize a new scan cycle. Results of running agent tasks will be dropped, but already stored results will be kept!',
            question: 'Are you sure you want to initialize a new scan cycle for',
          },
          pause: {
            tooltip: 'Pause',
          },
          resume: {
            tooltip: 'Resume',
          },
          generateScopeSecret: {
            tooltip: 'Reset Secret',
            title: 'Reset Scope Secret',
            message: 'This will lock out all scan agents using the old secret.',
            question: 'Are you sure you want to reset the scope secret of',
            success: 'The new scope secret has been sent to your e-mail address.',
          },
          delete: {
            title: 'Delete Scope',
            message: 'This will delete all scan progress, associated views and access rights.',
            question: 'Are you sure you want to delete the scan scope',
          },
        },
        types: scopeTypeLabels,
      },
    },
    dialog: {
      generateScopeSecret: {
        title: 'Reset Scope Secret',
        hint: 'Please copy the scope secret, it will disappear shorty.',
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
// i18n, toast, scopes store, and keyed-sort helper instances
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()
const scopesStore = useScopesStore()
const { allowCustom, allowNetwork, allowAsset } = storeToRefs(scopesStore)
const scopesSort = useKeyedGroupSort()

// One inline action panel is open at a time (custom / assets / networks /
// settings), rendered above the list. The guarded panel owns which form + which
// scope is shown, and gates any switch behind the discard confirm when the open
// form has unsaved edits.
type ScopePanelKind = 'custom' | 'assets' | 'networks' | 'settings'
type ScopeTargetsPanelKind = Exclude<ScopePanelKind, 'settings'>

const panel = useGuardedPanel<ScopePanelKind, ScopeResponse>()
const {
  activeKind: activePanelKind,
  target: editingScope,
  confirmVisible: discardVisible,
  openCount: panelOpenCount,
  keyFor: panelKey,
  open: openPanel,
  close: closePanel,
  onConfirmDiscard,
} = panel

// Hash tokens name edit actions or the exact type selected for a create action.
const actionPanelEl = ref<HTMLElement | null>(null)
interface ActionFormHandle {
  submit: () => Promise<void> | void
}
const actionFormRef = ref<ActionFormHandle | null>(null)

const actionPanelTitle = computed(() => {
  const kind = activePanelKind.value
  if (!kind) return ''
  if (kind === 'settings') return t('configuration.scopes.panels.settings')

  // Name edit panels after the concrete scope type instead of its generic form family
  const scope = editingScope.value
  if (scope) {
    return t('configuration.scopes.panels.update', { type: scopeTypeLabel(scope.type) })
  }

  // Name create panels after the exact configured scope type selected by the action
  let scopeType = 'custom'
  if (kind === 'assets') scopeType = panelAssetScopeType.value
  if (kind === 'networks') scopeType = panelNetworkScopeType.value

  return t('configuration.scopes.panels.create', { type: scopeTypeLabel(scopeType) })
})

function submitActionForm(): Promise<void> | void {
  return actionFormRef.value?.submit()
}

useSectionScroll(() => actionPanelEl.value, {
  activeSection: activePanelKind,
  target: editingScope,
  openCount: panelOpenCount,
})

function setActionPanelElement(element: Element | { $el: Element } | null): void {
  // Resolve Vue component refs and direct DOM refs to the section element used for scrolling
  const panelElement = element instanceof Element ? element : element?.$el
  actionPanelEl.value = panelElement instanceof HTMLElement ? panelElement : null
}

const deeplink = usePanelDeeplink(panel, {
  encode: (kind, scope) => {
    // Edit: the three type-specific edit forms all deeplink as `edit=<id>`;
    // Settings keeps its own action token. Both carry the scope id.
    if (scope) return `${kind === 'settings' ? 'settings' : 'edit'}=${scope.id}`

    // Preserve the selected inventory type so multiple variants can share one form family
    if (kind === 'assets') return `add_${panelAssetScopeType.value}`
    if (kind === 'networks') return `add_${panelNetworkScopeType.value}`
    return 'add_custom'
  },
  restore: (raw) => {
    const [action = '', idStr] = raw.split('=')

    // Create actions carry no id and reopen the exact generic or inventory-specific type
    if (action.startsWith('add_')) {
      let scopeType = action.slice(4)
      if (scopeType === 'asset') scopeType = 'assets'
      if (scopeType === 'network') scopeType = 'networks'

      const definition = scopeTypes[scopeType]
      if (definition) openCreatePanel(definition.type, scopeType)
      return
    }

    // Edit actions resolve the scope from the loaded list, then open as a click
    // would (openTargetsWidget picks the assets/networks/custom form by scope type)
    if (action !== 'edit' && action !== 'settings') return
    const scope = scopesStore.scopes.find((candidateScope) => candidateScope.id === Number(idStr))
    if (!scope) return
    if (action === 'settings') openSettingsPanel(scope)
    else openTargetsWidget(scope)
  },
})

// One dialog gates two distinct actions (delete + reset-secret); the title and
// message are set per-ask, and the action itself is queued on the confirm flow.
const {
  visible: actionConfirmVisible,
  ask: askConfirm,
  confirm: confirmQueuedScopeAction,
} = useConfirmAction<ScopeResponse, string>()

// ── Reactive state ────────────────────────────────────────────
// The scope subtype passed to the networks/assets forms when adding a new scope;
// set before opening the panel, mirroring the legacy add flow.
const panelNetworkScopeType = ref('networks')
const panelAssetScopeType = ref('assets')

// Shared dialog-confirm state; populated by each action before the dialog opens
const confirmTitle = ref('')
const confirmBody = ref<{ message: string; question: string; name: string }>({
  message: '',
  question: '',
  name: '',
})

// Per-action confirm button color: new-cycle/rescan teal (primary), reset-secret
// yellow (warn), delete red (danger).
const confirmSeverity = ref<'danger' | 'warn' | 'primary'>('danger')

// Per-action confirm icon mirrors the trigger button: reset-secret a ticket, delete a trash bin.
const confirmIcon = ref('pi pi-exclamation-triangle')

// When set, the dialog requires typing this word before confirm enables.
// Only used for highly disruptive actions (new cycle, delete). Not rescan or reset-secret.
const confirmWord = ref<string | undefined>(undefined)

// One-time-reveal secret, populated only when the backend could not e-mail it
// (the user has no certificate). Kept in local state for its display lifetime.
const revealedSecret = ref<string | null>(null)

// Cycle-progress popover template ref and the scope currently under the pointer
const cyclePopoverRef = ref<InstanceType<typeof Popover>>()
const cyclePopoverScope = ref<ScopeResponse | null>(null)

// ── Computed ──────────────────────────────────────────────────
// Scopes render as one table per group. Each group's table gets its own sort
// state (keyed by group name), so sorting one group never re-sorts the others.
// No column is sorted until the user clicks a header.
const groupedScopes = computed(() =>
  groupByKey(
    scopesStore.scopes,
    (scope) => scope.group_name || '-',
    (rows, key) => scopesSort.forGroup(key).sorted(rows),
  ),
)

// The available "add scope" actions depend on the current user's per-group
// creation rights (allow_custom / allow_network / allow_asset from GET /scopes).
// Hidden actions are legacy parity: a user without a right never sees the entry,
// and the whole add bar disappears when they have none. The backend re-checks
// the right on submit regardless — this is UX, not the security boundary.
interface AddAction {
  label: string
  command: () => void
}

const allowedAddActions = computed<AddAction[]>(() =>
  Object.entries(scopeTypes)
    .filter(([, definition]) => canCreateScopeType(definition.type))
    .map(([scopeType, definition]) => ({
      label: t('configuration.scopes.actions.createScope.button', {
        type: scopeTypeLabel(scopeType),
      }),
      command: () => openCreatePanel(definition.type, scopeType),
    })),
)

// First allowed action drives the split-button's primary label/command
const primaryAddAction = computed<AddAction | undefined>(() => allowedAddActions.value[0])

// Remaining allowed actions become the split-button's dropdown menu items
const addMenuItems = computed<MenuItem[]>(() =>
  allowedAddActions.value.slice(1).map((addAction) => ({
    label: addAction.label,
    icon: 'pi pi-plus',
    command: addAction.command,
  })),
)

// Attribute bags forwarded to the split-button's main and dropdown parts
const addButtonProps = computed<Record<string, unknown>>(() => ({
  'aria-label': primaryAddAction.value?.label ?? '',
  'data-testid': 'sl-add-custom',
}))
const addMenuButtonProps: Record<string, unknown> = {
  'aria-label': t('configuration.scopes.actions.createMore.button'),
  'data-testid': 'sl-add-menu',
}

// ── Functions and handlers ────────────────────────────────────
// Route the edit action to the form shared by the scope type's generic family.
function openTargetsWidget(scope: ScopeResponse): void {
  const panelKind = scopeTypes[scope.type]?.type ?? 'custom'

  if (panelKind === 'assets') {
    panelAssetScopeType.value = scope.type
  } else if (panelKind === 'networks') {
    panelNetworkScopeType.value = scope.type
  }

  openPanel(panelKind, scope)
}

function openSettingsPanel(scope: ScopeResponse): void {
  openPanel('settings', scope)
}

function openCreatePanel(panelKind: ScopeTargetsPanelKind, scopeType: string): void {
  // Pass the selected inventory type through the generic assets or networks form
  if (panelKind === 'assets') {
    panelAssetScopeType.value = scopeType
  } else if (panelKind === 'networks') {
    panelNetworkScopeType.value = scopeType
  }

  openPanel(panelKind, null)
}

function canCreateScopeType(scopeType: ScopeTargetsPanelKind): boolean {
  // Apply the existing family permissions without altering the configured registry
  if (scopeType === 'custom') return allowCustom.value
  if (scopeType === 'assets') return allowAsset.value
  return allowNetwork.value
}

// Maps a scope type to its localized label, falling back to the raw type.
function scopeTypeLabel(type: string): string {
  const key = `configuration.scopes.types.${type}`
  const label = t(key)
  return label === key ? type : label
}

function isCustomLike(scope: ScopeResponse): boolean {
  return scope.type === 'custom'
}

// Cycle-progress popup
function cycleProgressTuple(scope: ScopeResponse): CycleProgressTuple {
  return toCycleProgress(scope.cycle_done, scope.cycle_failed, scope.cycle_active)
}

function openCyclePopover(event: MouseEvent, scope: ScopeResponse): void {
  cyclePopoverScope.value = scope
  cyclePopoverRef.value?.show(event)
}

function closeCyclePopover(): void {
  cyclePopoverRef.value?.hide()
  cyclePopoverScope.value = null
}

function formatNumber(n: number): string {
  return new Intl.NumberFormat().format(n)
}

function cycleQueueEntries(scope: ScopeResponse): [string, number][] {
  // The optional queue map is present while the backend is scheduling agent work
  const queueByAgent = scope.cycle_queue
  if (!queueByAgent || Object.keys(queueByAgent).length === 0) return []
  return Object.entries(queueByAgent)
}

// Scope actions (pause, new cycle, rescan, reset secret, delete)
async function toggleScopePause(scope: ScopeResponse): Promise<void> {
  try {
    await scopesStore.pauseScope(scope.id)
  } catch {
    // The API client already reported the failed scope update
  }
}

function openNewCycleConfirm(scope: ScopeResponse): void {
  confirmTitle.value = t('configuration.scopes.actions.newCycle.title')
  confirmBody.value = {
    message: t('configuration.scopes.actions.newCycle.message'),
    question: t('configuration.scopes.actions.newCycle.question'),
    name: scope.name,
  }
  confirmSeverity.value = 'primary'
  confirmIcon.value = 'pi pi-forward'

  // Starting a fresh cycle is disruptive, so gate it behind typing the scope name
  // (legacy parity: the old UI required the scope name before starting a cycle).
  confirmWord.value = scope.name
  askConfirm({
    item: scope,
    run: (confirmedScope) => scopesStore.startCycle(confirmedScope.id),
  })
}

function openRescanFailedConfirm(scope: ScopeResponse): void {
  confirmTitle.value = t('configuration.scopes.actions.rescanFailed.title')
  confirmBody.value = {
    message: t('configuration.scopes.actions.rescanFailed.message'),
    question: t('configuration.scopes.actions.rescanFailed.question'),
    name: scope.name,
  }
  confirmSeverity.value = 'primary'
  confirmIcon.value = 'pi pi-replay'
  confirmWord.value = undefined
  askConfirm({
    item: scope,
    run: (confirmedScope) => scopesStore.rescanFailedTargets(confirmedScope.id),
  })
}

function openGenerateScopeSecretConfirm(scope: ScopeResponse): void {
  confirmTitle.value = t('configuration.scopes.actions.generateScopeSecret.title')
  confirmBody.value = {
    message: t('configuration.scopes.actions.generateScopeSecret.message'),
    question: t('configuration.scopes.actions.generateScopeSecret.question'),
    name: scope.name,
  }
  confirmSeverity.value = 'warn'
  confirmIcon.value = 'pi pi-ticket'
  confirmWord.value = undefined
  askConfirm({
    item: scope,
    run: (confirmedScope) => scopesStore.generateScopeSecret(confirmedScope.id),

    // A non-empty secret means e-mail delivery was not possible, so reveal it
    // once; an empty secret means it was e-mailed, so just confirm with a toast
    onSuccess: (secret) => {
      if (secret) {
        revealedSecret.value = secret
      } else {
        toast.success(t('configuration.scopes.actions.generateScopeSecret.success'))
      }
    },
  })
}

// Dismiss the revealed-secret dialog
function clearRevealedSecret(): void {
  revealedSecret.value = null
}

function onSecretDialogVisibilityChange(visible: boolean): void {
  if (!visible) clearRevealedSecret()
}

function openDeleteConfirm(scope: ScopeResponse): void {
  confirmTitle.value = t('configuration.scopes.actions.delete.title')
  confirmBody.value = {
    message: t('configuration.scopes.actions.delete.message'),
    question: t('configuration.scopes.actions.delete.question'),
    name: scope.name,
  }
  confirmSeverity.value = 'danger'
  confirmIcon.value = 'pi pi-trash'

  // Permanently deletes the scope and all its results, so gate it behind typing
  // the scope name (same guard as starting a new cycle).
  confirmWord.value = scope.name
  askConfirm({
    item: scope,
    run: (confirmedScope) => scopesStore.deleteScope(confirmedScope.id),
  })
}

// The store action already reloaded the scopes, so nothing needs to happen here
function onSaved(): void {}

// ── Watchers and lifecycle ────────────────────────────────────

onMounted(async () => {
  // Load the scope list on first render; surface a fetch failure as an error
  // toast (site-wide convention — the list itself only shows loading / empty).
  try {
    await scopesStore.fetchScopes()
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }

  // Re-open the panel encoded in the URL hash now that scopes are available
  deeplink.restoreFromHash()
})
</script>

<template>
  <!-- Action panel — opens as its own section ABOVE the page header, so an
         open create/edit form reads as the top-most section and the page scrolls
         up to reveal its header. -->
  <LayoutReveal>
    <LayoutForm
      v-if="activePanelKind"
      :ref="setActionPanelElement"
      class="page-section"
      data-testid="sp-action-panel"
      :title="actionPanelTitle"
      :text-submit-label="t('common.actions.save')"
      :submit-handler="submitActionForm"
      :footer="false"
    >
      <FormScopeAddCustom
        v-if="activePanelKind === 'custom'"
        :key="panelKey('custom', editingScope?.id)"
        ref="actionFormRef"
        :scope="editingScope ?? undefined"
        @close="closePanel"
        @saved="onSaved"
      />
      <FormScopeAddAssets
        v-else-if="activePanelKind === 'assets'"
        :key="panelKey('assets', editingScope?.id)"
        ref="actionFormRef"
        :scope="editingScope ?? undefined"
        :scope-type="panelAssetScopeType"
        @close="closePanel"
        @saved="onSaved"
      />
      <FormScopeAddNetworks
        v-else-if="activePanelKind === 'networks'"
        :key="panelKey('networks', editingScope?.id)"
        ref="actionFormRef"
        :scope="editingScope ?? undefined"
        :scope-type="panelNetworkScopeType"
        @close="closePanel"
        @saved="onSaved"
      />
      <FormScopeSettings
        v-else-if="activePanelKind === 'settings' && editingScope"
        :key="panelKey('settings', editingScope.id)"
        ref="actionFormRef"
        :scope="editingScope"
        @close="closePanel"
        @saved="onSaved"
      />
    </LayoutForm>
  </LayoutReveal>

  <section class="page-section scopes-page">
    <SectionHeader :title="t('common.labels.scopes')" :help-text="t('configuration.scopes.help')" />
    <div data-testid="scope-list">
      <LayoutLoading
        class="scopes-page__layout-loading-container"
        :loading="scopesStore.isFetching && scopesStore.scopes.length === 0"
        :loading-message="t('configuration.scopes.loading')"
        :empty="groupedScopes.length === 0"
        :empty-message="t('configuration.scopes.empty')"
      >
        <DataGrid :groups="groupedScopes" :sort="scopesSort" :row-key="(scope) => scope.id">
          <template #group-header="{ groupName }">
            <h4 data-testid="sl-group-header">
              <span class="sensitive">{{ groupName }}</span>
            </h4>
          </template>
          <template #header="{ groupSort }">
            <SortableTh
              field="name"
              :active="groupSort.field.value"
              :direction="groupSort.direction.value"
              @toggle="groupSort.toggle"
            >
              {{ t('common.labels.name') }}
            </SortableTh>
            <SortableTh
              field="created_by"
              :active="groupSort.field.value"
              :direction="groupSort.direction.value"
              class="scopes-page__column-created-by"
              @toggle="groupSort.toggle"
            >
              {{ t('common.labels.createdBy') }}
            </SortableTh>
            <SortableTh
              field="created"
              :active="groupSort.field.value"
              :direction="groupSort.direction.value"
              class="scopes-page__column-created"
              @toggle="groupSort.toggle"
            >
              {{ t('common.labels.created') }}
            </SortableTh>
            <SortableTh
              field="type"
              :active="groupSort.field.value"
              :direction="groupSort.direction.value"
              class="scopes-page__column-type"
              @toggle="groupSort.toggle"
            >
              {{ t('common.labels.type') }}
            </SortableTh>
            <th class="scopes-page__column-config">
              {{ t('configuration.scopes.columns.config') }}
            </th>
            <th class="scopes-page__column-progress">
              {{ t('configuration.scopes.columns.cycleProgress') }}
            </th>
            <th class="data-grid__th-actions"></th>
          </template>

          <template #row="{ row: scope }">
            <td class="data-grid__cell--ellipsis">
              <span
                v-tooltip.left="scope.connection?.database ?? ''"
                class="sensitive"
                data-testid="sl-scope-name"
                >{{ scope.name }}</span
              >
            </td>
            <td class="data-grid__cell--ellipsis">
              <span class="sensitive" :title="scope.created_by">{{ scope.created_by }}</span>
            </td>
            <td class="data-grid__cell--nowrap">
              {{ formatDateTime(scope.created) }}
            </td>
            <td class="data-grid__cell--nowrap scopes-page__cell-type">
              {{ scopeTypeLabel(scope.type) }}
            </td>
            <td class="data-grid__cell--nowrap">
              <div class="scopes-page__config-icons">
                <i
                  v-tooltip.top="t('configuration.scopes.configIndicators.continuous')"
                  class="pi pi-refresh"
                  :class="
                    scope.cycles
                      ? 'scopes-page__config-icon--active'
                      : 'scopes-page__config-icon--inactive'
                  "
                />
                <i
                  v-tooltip.top="
                    scope.cycles_retention === -1
                      ? t('configuration.scopes.configIndicators.retentionAll')
                      : scope.cycles_retention > 0
                        ? t(
                            'configuration.scopes.configIndicators.retentionLastN',
                            { count: scope.cycles_retention },
                            scope.cycles_retention,
                          )
                        : t('common.labels.retention')
                  "
                  class="pi pi-eraser"
                  :class="
                    scope.cycles_retention > 0
                      ? 'scopes-page__config-icon--active'
                      : 'scopes-page__config-icon--inactive'
                  "
                />
                <i
                  v-if="isCustomLike(scope)"
                  v-tooltip.top="t('configuration.scopes.configIndicators.ot')"
                  class="pi pi-sitemap"
                  :class="
                    scope.scan_settings?.ot
                      ? 'scopes-page__config-icon--active'
                      : 'scopes-page__config-icon--inactive'
                  "
                />
                <i
                  v-if="scope.attributes?.sync !== undefined"
                  v-tooltip.top="t('configuration.scopes.configIndicators.sync')"
                  class="pi pi-sync"
                  :class="
                    scope.attributes.sync
                      ? 'scopes-page__config-icon--active'
                      : 'scopes-page__config-icon--inactive'
                  "
                />
              </div>
            </td>
            <td class="data-grid__cell--nowrap">
              <div
                class="scopes-page__progress"
                @mouseenter="openCyclePopover($event, scope)"
                @mouseleave="closeCyclePopover"
              >
                <ProgressBar3
                  class="scopes-page__progress-bar"
                  :progress="cycleProgressTuple(scope)"
                  :total="100"
                  size="sm"
                />
              </div>
            </td>
            <td class="data-grid__cell--actions">
              <div class="data-grid__actions">
                <Button
                  v-tooltip.top="t('configuration.scopes.actions.targets.tooltip')"
                  type="button"
                  icon="pi pi-bullseye"
                  class="data-grid__btn"
                  data-testid="sl-action-targets"
                  :aria-label="t('configuration.scopes.actions.targets.tooltip')"
                  @click="openTargetsWidget(scope)"
                />
                <Button
                  v-tooltip.top="t('configuration.scopes.actions.settings.tooltip')"
                  type="button"
                  icon="pi pi-sliders-h"
                  class="data-grid__btn"
                  data-testid="sl-action-settings"
                  :aria-label="t('configuration.scopes.actions.settings.tooltip')"
                  @click="openSettingsPanel(scope)"
                />
                <Button
                  v-tooltip.top="t('configuration.scopes.actions.rescanFailed.tooltip')"
                  type="button"
                  icon="pi pi-replay"
                  class="data-grid__btn"
                  data-testid="sl-action-rescan"
                  :aria-label="t('configuration.scopes.actions.rescanFailed.tooltip')"
                  @click="openRescanFailedConfirm(scope)"
                />
                <Button
                  v-tooltip.top="t('configuration.scopes.actions.newCycle.tooltip')"
                  type="button"
                  icon="pi pi-forward"
                  class="data-grid__btn"
                  data-testid="sl-action-new-cycle"
                  :aria-label="t('configuration.scopes.actions.newCycle.tooltip')"
                  @click="openNewCycleConfirm(scope)"
                />
                <Button
                  v-tooltip.top="
                    !scope.enabled
                      ? t('configuration.scopes.actions.resume.tooltip')
                      : t('configuration.scopes.actions.pause.tooltip')
                  "
                  type="button"
                  :severity="!scope.enabled ? 'secondary' : undefined"
                  :icon="!scope.enabled ? 'pi pi-play' : 'pi pi-pause'"
                  :pt="
                    !scope.enabled
                      ? {
                          icon: {
                            style: { color: 'var(--p-text-muted-color)' },
                          },
                        }
                      : undefined
                  "
                  class="data-grid__btn"
                  data-testid="sl-action-pause"
                  :aria-label="
                    !scope.enabled
                      ? t('configuration.scopes.actions.resume.tooltip')
                      : t('configuration.scopes.actions.pause.tooltip')
                  "
                  @click="toggleScopePause(scope)"
                />
                <Button
                  v-tooltip.top="t('configuration.scopes.actions.generateScopeSecret.tooltip')"
                  type="button"
                  severity="warn"
                  icon="pi pi-ticket"
                  class="data-grid__btn"
                  data-testid="sl-action-secret"
                  :aria-label="t('configuration.scopes.actions.generateScopeSecret.tooltip')"
                  @click="openGenerateScopeSecretConfirm(scope)"
                />
                <Button
                  v-tooltip.top="t('common.actions.delete')"
                  type="button"
                  severity="danger"
                  icon="pi pi-trash"
                  class="data-grid__btn"
                  data-testid="sl-action-delete"
                  :aria-label="t('common.actions.delete')"
                  @click="openDeleteConfirm(scope)"
                />
              </div>
            </td>
          </template>
        </DataGrid>
      </LayoutLoading>

      <!-- Add-scope action(s): visible only for the scope types the current
           user is allowed to create. Hidden entirely when they have no rights;
           a plain button when only one type is allowed; a split button (primary
           + drop-down) when several are. -->
      <LayoutRightAligned v-if="primaryAddAction">
        <SegmentedButton
          v-if="addMenuItems.length"
          split
          data-testid="sl-add-split"
          :label="primaryAddAction.label"
          icon="pi pi-plus"
          :model="addMenuItems"
          :button-props="addButtonProps"
          :menu-button-props="addMenuButtonProps"
          @click="primaryAddAction.command()"
        />
        <SegmentedButton
          v-else
          :label="primaryAddAction.label"
          icon="pi pi-plus"
          :aria-label="primaryAddAction.label"
          @click="primaryAddAction.command()"
        />
      </LayoutRightAligned>

      <Popover ref="cyclePopoverRef">
        <div v-if="cyclePopoverScope" class="cycle-progress">
          <table class="cycle-progress__table">
            <tbody>
              <tr>
                <td class="cycle-progress__icon-cell">
                  <i class="pi pi-bullseye cycle-progress__icon" aria-hidden="true" />
                </td>
                <td>
                  {{
                    t('configuration.scopes.cyclePopup.ips', {
                      count: formatNumber(cyclePopoverScope.size),
                    })
                  }}
                </td>
              </tr>
              <tr>
                <td class="cycle-progress__icon-cell">
                  <span class="cycle-progress__icon-stack">
                    <i class="pi pi-bullseye cycle-progress__icon" aria-hidden="true" />
                    <i class="pi pi-calendar cycle-progress__corner-icon" aria-hidden="true" />
                  </span>
                </td>
                <td>
                  <span
                    v-if="isZeroDate(cyclePopoverScope.last_sync)"
                    class="cycle-progress__italic"
                  >
                    {{ t('configuration.scopes.cyclePopup.notSynced') }}
                  </span>
                  <span v-else
                    >{{ formatDate(cyclePopoverScope.last_sync)
                    }}<template v-if="daysAgo(cyclePopoverScope.last_sync)">
                      | {{ daysAgo(cyclePopoverScope.last_sync) }}</template
                    ></span
                  >
                </td>
              </tr>
              <tr>
                <td class="cycle-progress__icon-cell">
                  <i class="pi pi-refresh cycle-progress__icon" aria-hidden="true" />
                </td>
                <td>
                  {{
                    t('configuration.scopes.cyclePopup.cycle', {
                      n: formatNumber(cyclePopoverScope.cycle),
                    })
                  }}
                </td>
              </tr>
              <tr>
                <td class="cycle-progress__icon-cell">
                  <span class="cycle-progress__icon-stack">
                    <i class="pi pi-refresh cycle-progress__icon" aria-hidden="true" />
                    <i class="pi pi-calendar cycle-progress__corner-icon" aria-hidden="true" />
                  </span>
                </td>
                <td>
                  {{ formatDate(cyclePopoverScope.cycle_started)
                  }}<template v-if="daysAgo(cyclePopoverScope.cycle_started)">
                    | {{ daysAgo(cyclePopoverScope.cycle_started) }}</template
                  >
                </td>
              </tr>
              <template v-if="cycleQueueEntries(cyclePopoverScope).length > 0">
                <tr class="cycle-progress__queue-header">
                  <td class="cycle-progress__icon-cell">
                    <i class="pi pi-list-check cycle-progress__icon" aria-hidden="true" />
                  </td>
                  <td>{{ t('configuration.scopes.cyclePopup.queued') }}</td>
                </tr>
                <!-- Queue entries in a nested table so their count/agent column
                     widths don't affect the outer stats table's layout. -->
                <tr class="cycle-progress__queue-row">
                  <td colspan="2">
                    <table class="cycle-progress__queue-table">
                      <tbody>
                        <tr
                          v-for="[agent, count] in cycleQueueEntries(cyclePopoverScope)"
                          :key="agent"
                        >
                          <td class="cycle-progress__queue-count">{{ count }}×</td>
                          <td class="cycle-progress__queue-agent">
                            {{ agent }}
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </Popover>

      <DialogConfirm
        v-model:visible="actionConfirmVisible"
        :title="confirmTitle"
        :icon="confirmIcon"
        :severity="confirmSeverity"
        :confirm-word="confirmWord"
        :text-confirm-label="t('common.actions.confirm')"
        :message="`${confirmBody.question} '${confirmBody.name}'?`"
        :secondary-message="confirmBody.message"
        @confirm="confirmQueuedScopeAction"
      />

      <Dialog
        :visible="revealedSecret !== null"
        modal
        dismissable-mask
        :draggable="false"
        class="app-dialog scopes-page__secret-dialog"
        @update:visible="onSecretDialogVisibilityChange"
      >
        <template #header>
          <div class="app-dialog__header">
            <i class="pi pi-key app-dialog__header-icon" aria-hidden="true" />
            <span class="app-dialog__header-title">{{
              t('dialog.generateScopeSecret.title')
            }}</span>
          </div>
        </template>
        <div class="scopes-page__dialog-body">
          <p>{{ t('dialog.generateScopeSecret.hint') }}</p>
          <CodeSnippet
            :text="revealedSecret ?? ''"
            :text-copy-label="t('common.actions.copy')"
            centered
          />
        </div>
        <template #footer>
          <Button type="button" :label="t('common.actions.close')" @click="clearRevealedSecret" />
        </template>
      </Dialog>
    </div>

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
/* Keep compact metadata columns stable while the name column absorbs free space */
.scopes-page__column-created-by,
.scopes-page__column-progress {
  width: 10.714286rem;
}

/* Reserve the minimum width needed for formatted creation dates */
.scopes-page__column-created {
  width: 4.571429rem;
}

/* Keep the scope type label on one compact line */
.scopes-page__column-type {
  width: 6rem;
}

/* Reserve space for the scope configuration icons */
.scopes-page__column-config {
  width: 5.285714rem;
}

/* Match shared confirmation dialogs while remaining responsive */
.scopes-page__secret-dialog {
  width: 38rem;
  max-width: calc(100vw - 2rem);
}

/* ── Table cell overrides ───────────────────────────────────── */

.scopes-page__cell-type {
  text-transform: capitalize;
}

/* ── Per-scope config icons (continuous / retention / OT / sync) ───────────── */

.scopes-page__config-icons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.scopes-page__config-icon--active {
  color: var(--p-primary-500);
}

.scopes-page__config-icon--inactive {
  opacity: 0.35;
}

/* ── Cycle progress bar cell ───────────────────────────────────────────────── */

.scopes-page__progress {
  display: flex;
  align-items: center;
}

.scopes-page__progress-bar {
  flex: 1;
  min-width: 5rem;
}

/* ── Cycle-progress hover popup (IPs, sync, cycle, queue) ──────────────────── */
/* Laid out as a table styled to match the scan-agent tasks popup
   (WidgetAgents.vue .widget-agents__tasks-table): 0.4em 0.6em cell padding with surface-200
   horizontal separators between rows. */

.cycle-progress {
  font-size: var(--app-font-size-small);
  /* Muted text to match the tooltip/popover surface (set on .p-popover). */
  color: var(--p-text-muted-color);
  min-width: 11.428571rem;
}

.cycle-progress__table {
  width: 100%;
  border-collapse: collapse;
  border: none;
}

/* Direct-child selectors so the rules don't leak into the nested queue table. */
.cycle-progress__table > tbody > tr > td {
  padding: 0.4em 0.6em;
  /* Match the taller row rhythm of the agent active-tasks tooltip (its text
     cells use the default ~1.5 line box); the icon cells stay centred via
     their own line-height:1. */
  line-height: 1.5;
}

.cycle-progress__table > tbody > tr + tr {
  border-top: 1px solid var(--p-surface-200);
}

/* Icon column: narrow, centred, shrink-to-fit. */
.cycle-progress__icon-cell {
  width: 1rem;
  text-align: center;
  white-space: nowrap;
}

.cycle-progress__icon {
  /* Match the icon glyph to the row text size; PrimeIcons default to 1rem (16px)
     which is taller than the ~14px popup text. line-height:1 keeps it centred. */
  font-size: var(--app-font-size-small);
  line-height: 1;
}

/* Stacks a small calendar badge over a base icon (e.g. "last synced" date). */
.cycle-progress__icon-stack {
  position: relative;
  display: inline-block;
}

.cycle-progress__corner-icon {
  position: absolute;
  bottom: 0;
  right: -4px;
  font-size: 0.45rem;
}

.cycle-progress__italic {
  font-style: italic;
}

/* Extra top padding to set the "Queued:" section apart from the stats rows.
   Selector mirrors the base `.cycle-progress__table > tbody > tr > td` rule's
   specificity (which sets `padding` shorthand) so this override actually wins. */
.cycle-progress__table > tbody > tr.cycle-progress__queue-header > td {
  padding-top: 0.6em;
  padding-bottom: 0;
}

.cycle-progress__queue-row {
  border-top: none !important;
}

/* The colspan wrapper cell holds the nested table flush — no extra padding. */
.cycle-progress__queue-row > td {
  padding: 0 !important;
}

/* Nested queue table: count (right) / agent (left), no inner borders. */
.cycle-progress__queue-table {
  width: 100%;
  border-collapse: collapse;
}

.cycle-progress__queue-table td {
  padding: 0.4em 0.6em;
  /* Tight rows for the count/agent list; the outer table's 1.5 line-height
     would otherwise inherit in and make these rows look padded. */
  line-height: 1;
}

.cycle-progress__queue-count {
  text-align: right;
  width: 50%;
}

.cycle-progress__queue-agent {
  width: 50%;
}

.scopes-page__dialog-body {
  text-align: center;
}

:deep(.data-grid__cell--actions) {
  width: 15.2rem;
  min-width: 5.5rem;
  overflow: hidden;
}
</style>
