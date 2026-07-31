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
 * ViewsPage — Configuration / Views page.
 *
 * Lists scan-scope views grouped by owner group, their user grants and access
 * tokens, and gates create / edit / grant / revoke / delete behind dialogs.
 */
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppMessage from '@/components/elements/AppMessage.vue'
import DataGrid from '@/components/elements/DataGrid.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import SortableTh from '@/components/elements/SortableTh.vue'
import TagLabel from '@/components/elements/TagLabel.vue'
import FormViewAdd from '@/components/forms/FormViewAdd.vue'
import FormViewEdit from '@/components/forms/FormViewEdit.vue'
import FormViewGrant from '@/components/forms/FormViewGrant.vue'
import FormViewToken from '@/components/forms/FormViewToken.vue'
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
import { groupByKey } from '@/lib/collections'
import { formatDate } from '@/i18n/format'
import { useViewsStore } from '@/stores/views'

import type { TokenRow, ViewGrant, ViewResponse } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    revoke: 'Revoke',
    unsavedTitle: 'Discard unsaved changes?',
    unsavedMessage: 'You have unsaved changes. If you leave this page, they will be discarded.',
    configuration: {
      views: {
        help: 'This section lists views you are allowed to administer. A view provides access to the scan results of a specific scope. It can be filtered, and you can choose who may access it.',
        loading: 'Loading views...',
        empty: 'No views yet.',
        error: 'Could not load views.',
        panels: {
          add: 'Create View',
          edit: 'Configure View',
          grant: 'View Access Rights',
          token: 'Generate Access Token',
        },
        emptyStates: {
          filters: 'No Filters',
          users: 'No Users',
        },
        warnings: {
          inactiveUsers: {
            prefix: 'The following users have access rights but did not login for at least',
            duration: 'six months',
          },
        },
        columns: {
          filter: 'Filter',
        },
        tokens: {
          title: 'Access Tokens',
          help: 'This section lists long-lived access tokens that are not bound to a user. Create them only when necessary.',
          loading: 'Loading access tokens...',
          empty: 'No access tokens generated. Good, you are awesome!',
        },
        actions: {
          edit: {
            tooltip: 'Edit',
          },
          grant: {
            tooltip: 'Access Rights',
          },
          generateAccessToken: {
            tooltip: 'Generate Access Token',
          },
          delete: {
            title: 'Delete View',
            message: 'This will remove the view and all associated access rights.',
            question: 'Are you sure you want to delete the view',
          },
          revokeUser: {
            title: 'Revoke User',
            message: 'This will remove all access rights of this user from ',
            question: 'Are you sure you want to revoke the user',
          },
          revokeToken: {
            title: 'Revoke Access Token',
            message: 'This will remove all access rights of this token from ',
            question: 'Are you sure you want to revoke the token',
          },
        },
      },
    },
  },
}

type LabelVariant = 'info' | 'primary' | 'warn' | 'error'

const VIEW_PANEL_KINDS = ['add', 'edit', 'grant', 'token'] as const
type PanelKind = (typeof VIEW_PANEL_KINDS)[number]

// ── Composables and stores ────────────────────────────────────
// i18n, toast, views store, and auth store instances
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()
const viewsStore = useViewsStore()
const { views, tokens, isFetchingViews } = storeToRefs(viewsStore)

// The views and tokens lists each render as several sibling tables (one per
// group). Each group's table needs its own sort state, so a keyed sort hands out
// an independent GroupSort per group name — otherwise sorting one group's table
// would re-sort every table in the section.
// No column is sorted until the user clicks a header.
const viewsSort = useKeyedGroupSort()
const tokensSort = useKeyedGroupSort()

// One inline action panel is open at a time (add / edit / grant / token),
// rendered above the list like the Scopes page — not as modal overlays. The
// guarded panel owns which form + which view is shown, and gates any switch
// behind the discard confirm when the open form has unsaved edits.
const panel = useGuardedPanel<PanelKind, ViewResponse>()
const {
  activeKind: activePanelKind,
  target: selectedView,
  confirmVisible: discardVisible,
  openCount: panelOpenCount,
  keyFor: panelKey,
  open: openPanel,
  close: closePanel,
  onConfirmDiscard,
} = panel
// Mirror the open panel into the hash (e.g. `#edit=<id>`, `#add`) so a refresh or
// shared link re-opens it. Row panels carry the view id; add carries only the
// kind. Restore resolves the id against the loaded, access-filtered view list and
// opens the view directly (the row wrapper isn't needed — the panels take the raw
// ViewResponse as their target).
const actionPanelEl = ref<HTMLElement | null>(null)
interface ActionFormHandle {
  cancel: () => void
  submit: () => Promise<void> | void
}
const actionFormRef = ref<ActionFormHandle | null>(null)

const actionPanelTitle = computed(() =>
  activePanelKind.value ? t(`configuration.views.panels.${activePanelKind.value}`) : '',
)
const actionPanelHasFooter = computed(
  () => activePanelKind.value === 'edit' || activePanelKind.value === 'grant',
)
const actionPanelSubmitLabel = computed(() => {
  if (activePanelKind.value === 'add') return t('common.actions.create')
  if (activePanelKind.value === 'token') return t('common.actions.generate')
  return t('common.actions.save')
})

function submitActionForm(): Promise<void> | void {
  return actionFormRef.value?.submit()
}

function cancelActionForm(): void {
  actionFormRef.value?.cancel()
}

useSectionScroll(() => actionPanelEl.value, {
  activeSection: activePanelKind,
  target: selectedView,
  openCount: panelOpenCount,
})

function setActionPanelElement(element: Element | { $el: Element } | null): void {
  // Resolve Vue component refs and direct DOM refs to the section element used for scrolling
  const panelElement = element instanceof Element ? element : element?.$el
  actionPanelEl.value = panelElement instanceof HTMLElement ? panelElement : null
}

const deeplink = usePanelDeeplink(panel, {
  encode: (kind, view) => (view ? `${kind}=${view.id}` : kind),
  restore: (raw) => {
    const [kind, id] = raw.split('=')
    const restoredPanelKind = VIEW_PANEL_KINDS.find((panelKind) => panelKind === kind)
    if (!restoredPanelKind) return
    if (restoredPanelKind === 'add') {
      openAdd()
      return
    }
    const view = views.value.find((candidateView) => candidateView.id === id)
    if (!view) return
    openPanel(restoredPanelKind, view)
  },
})

// All three share one dialog and the same success and error handling. Only the
// title, confirm label, message and the action itself differ per invocation.
const {
  visible: actionConfirmVisible,
  ask: askConfirm,
  confirm: confirmQueuedAction,
} = useConfirmAction<void, string>()

// ── Reactive state ────────────────────────────────────────────
// Threshold constants used by the computed and handler sections
// A grant counts as "inactive" once its user has been idle for this long
const SIX_MONTHS_MS = 1000 * 60 * 60 * 24 * 30 * 6

// One day in milliseconds, used to bucket token expiry into urgency bands
const ONE_DAY_MS = 1000 * 60 * 60 * 24

// Shared dialog-confirm state; populated by each action before the dialog opens
const confirmTitle = ref('')
const confirmLabel = ref('')

// confirmMessage: optional impact line shown after the question
const confirmMessage = ref('')

// confirmMessageSuffix: optional highlighted name appended to confirmMessage (e.g. the view name in revoke dialogs)
const confirmMessageSuffix = ref('')

// confirmQuestion: the "Are you sure..." prefix before the highlighted name
const confirmQuestion = ref('')

// confirmName: the highlighted entity name
const confirmName = ref('')

// confirmNameSuffixPre: plain text between highlighted name and highlighted view (e.g. " from ")
const confirmNameSuffixPre = ref('')

// confirmViewName: highlighted view/scope name shown after the suffix pre-text
const confirmViewName = ref('')

// Dialog icon mirrors the trigger button so the confirm reads as the same action
const confirmIcon = ref('')

// ── Computed ──────────────────────────────────────────────────
// Shape for a single views-table row
interface ViewRow {
  id: string
  name: string
  scope_name: string
  group_name: string
  filtersList: { name: string; value: string }[]
  grants: ViewGrant[]
  _raw: ViewResponse
}

// Flatten each view into a display row, truncating long filter value lists
const viewRows = computed<ViewRow[]>(() =>
  views.value.map((view) => ({
    id: view.id,
    name: view.name,
    scope_name: view.scan_scope.name,
    group_name: view.scan_scope.group_name,
    filtersList: Object.entries(view.filters ?? {}).map(([filterName, filterValues]) => {
      const joined = filterValues.join(' | ')
      return {
        name: filterName,
        value: joined.length > 60 ? joined.slice(0, 60) + '...' : joined,
      }
    }),
    grants: view.grants,
    _raw: view,
  })),
)

// Views rows grouped by group name for the accordion display
const viewsGrouped = computed(() =>
  groupByKey(
    viewRows.value,
    (row) => row.group_name || '-',
    (rows, key) => viewsSort.forGroup(key).sorted(rows),
  ),
)

// Access tokens grouped by group name for the tokens accordion
const tokensGrouped = computed(() =>
  groupByKey(
    tokens.value,
    (token) => token.group_name || '-',
    (rows, key) => tokensSort.forGroup(key).sorted(rows),
  ),
)

// Users whose last login is older than the six-month threshold
const inactiveUsernames = computed(() => {
  const inactive = new Set<string>()
  for (const view of views.value) {
    for (const grant of view.grants) {
      if (isInactiveGrant(grant)) inactive.add(grant.username)
    }
  }
  return inactive
})

const inactiveUserLabels = computed(() => {
  const seen = new Set<string>()
  const labels: string[] = []
  for (const view of views.value) {
    for (const grant of view.grants) {
      if (!isInactiveGrant(grant)) continue
      const label = grantInactiveLabel(grant)
      if (seen.has(label)) continue
      seen.add(label)
      labels.push(label)
    }
  }
  return labels
})

// ── Functions and handlers ────────────────────────────────────
// CSS class for a token expiry date based on how many days remain
function expiryClass(expiry: string | null | undefined): string {
  // Treat missing or unparseable dates as "no urgency"
  if (!expiry) return ''
  const ts = new Date(expiry).getTime()
  if (Number.isNaN(ts)) return ''

  // Bucket the remaining days into red (soon) / yellow (upcoming) / teal (far)
  const days = (ts - Date.now()) / ONE_DAY_MS
  if (days <= 7) return 'views-page__expiry--red'
  if (days <= 21) return 'views-page__expiry--yellow'
  return 'views-page__expiry--teal'
}

function userGrants(grants: ViewGrant[]): ViewGrant[] {
  return grants.filter((grant) => grant.is_user)
}

// Groups user grants by company, preserving first-seen order for stable rendering
function grantsByCompany(grants: ViewGrant[]): { company: string; members: ViewGrant[] }[] {
  // Build an insertion-ordered map keyed by company, collapsing self-named companies to ""
  const map = new Map<string, ViewGrant[]>()
  for (const grant of userGrants(grants)) {
    const key = grant.username === grant.user_company ? '' : (grant.user_company ?? '')
    const members = map.get(key)
    if (members) {
      members.push(grant)
    } else {
      map.set(key, [grant])
    }
  }

  // Map preserves first-seen company order without a parallel key collection
  return [...map].map(([company, members]) => ({ company, members }))
}

function isInactiveGrant(grant: ViewGrant): boolean {
  if (!grant.is_user || !grant.user_last_login) return false
  const lastLogin = new Date(grant.user_last_login).getTime()
  if (Number.isNaN(lastLogin)) return false
  return Date.now() - lastLogin > SIX_MONTHS_MS
}

function grantInactiveLabel(grant: ViewGrant): string {
  if (!grant.user_company || grant.username === grant.user_company) return grant.username
  return `${grant.username} (${grant.user_company})`
}

function grantLabelVariant(grant: ViewGrant): LabelVariant {
  if (inactiveUsernames.value.has(grant.username)) return 'error'
  if (grant.user_is_admin || grant.user_is_owner) return 'info'
  return 'primary'
}

function openAdd(): void {
  openPanel('add', null)
}

function openEdit(row: ViewRow): void {
  openPanel('edit', row._raw)
}
function openGrant(row: ViewRow): void {
  openPanel('grant', row._raw)
}
function openToken(row: ViewRow): void {
  openPanel('token', row._raw)
}

// Wraps a mutating action so a successful run reloads the list and toasts
function withReloadAndToast(run: () => Promise<string>): () => Promise<string> {
  return async () => {
    const message = await run()
    await reloadViewsPageData()
    toast.success(message)
    return message
  }
}

function openDeleteConfirm(row: ViewRow): void {
  confirmTitle.value = t('configuration.views.actions.delete.title')
  confirmLabel.value = t('common.actions.delete')
  confirmMessage.value = t('configuration.views.actions.delete.message')
  confirmMessageSuffix.value = ''
  confirmQuestion.value = t('configuration.views.actions.delete.question')
  confirmName.value = row.name
  confirmNameSuffixPre.value = ''
  confirmViewName.value = ''
  confirmIcon.value = 'pi pi-trash'
  askConfirm({
    item: undefined,
    run: withReloadAndToast(() => viewsStore.deleteView(row.id)),
  })
}

function openRevokeUserConfirm(viewRow: ViewRow, username: string): void {
  confirmTitle.value = t('configuration.views.actions.revokeUser.title')
  confirmLabel.value = t('revoke')
  confirmMessage.value = t('configuration.views.actions.revokeUser.message')
  confirmMessageSuffix.value = `${viewRow.scope_name} - ${viewRow.name}`
  confirmQuestion.value = t('configuration.views.actions.revokeUser.question')
  confirmName.value = username
  confirmNameSuffixPre.value = ''
  confirmViewName.value = ''
  confirmIcon.value = 'pi pi-trash'
  askConfirm({
    item: undefined,
    run: withReloadAndToast(() => viewsStore.revokeGrant(viewRow.id, username)),
  })
}

function openRevokeTokenConfirm(token: TokenRow): void {
  confirmTitle.value = t('configuration.views.actions.revokeToken.title')
  confirmLabel.value = t('revoke')
  confirmMessage.value = t('configuration.views.actions.revokeToken.message')
  confirmMessageSuffix.value = `${token.scope_name} - ${token.view_name}`
  confirmQuestion.value = t('configuration.views.actions.revokeToken.question')
  confirmName.value = token.username
  confirmNameSuffixPre.value = ''
  confirmViewName.value = ''
  confirmIcon.value = 'pi pi-trash'
  askConfirm({
    item: undefined,
    run: withReloadAndToast(() => viewsStore.revokeGrant(token.view_id, token.username)),
  })
}

async function reloadViewsPageData(): Promise<void> {
  try {
    await viewsStore.fetchViews()
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }
}

// A panel form reported success (created/saved): reload the list. Closing is
// driven by the child's own `close` emit (Add/Edit/Grant close themselves after
// a successful submit; ViewTokenPanel stays open to show one-time credentials),
// so success here only needs to refresh the lists.
function refreshAfterPanelSaved(): void {
  void reloadViewsPageData()
}

// ── Watchers and lifecycle ────────────────────────────────────

// Guards a one-time load so re-authentication does not refetch needlessly
let loaded = false

async function loadForRole(): Promise<void> {
  if (loaded) return
  loaded = true
  await reloadViewsPageData()

  // Re-open the panel encoded in the URL hash now that views/tokens are available
  deeplink.restoreFromHash()
}

onMounted(() => void loadForRole())

// ── Computed ──────────────────────────────────────────────────
const confirmDialogMessage = computed(() => {
  const parts = [`${confirmQuestion.value} '${confirmName.value}'`]
  if (confirmNameSuffixPre.value) parts.push(confirmNameSuffixPre.value)
  if (confirmViewName.value) parts.push(`'${confirmViewName.value}'`)
  return `${parts.join('')}?`
})

const confirmDialogSecondaryMessage = computed(() =>
  confirmMessage.value
    ? confirmMessageSuffix.value
      ? `${confirmMessage.value}'${confirmMessageSuffix.value}'`
      : confirmMessage.value
    : undefined,
)
</script>

<template>
  <!-- Action panel — opens as its own section ABOVE the page header, so an
         open form reads as the top-most section and the page scrolls up to
         reveal its header. -->
  <LayoutReveal>
    <LayoutForm
      v-if="activePanelKind"
      :ref="setActionPanelElement"
      data-testid="views-action-panel"
      class="page-section"
      :title="actionPanelTitle"
      :text-submit-label="actionPanelSubmitLabel"
      :submit-handler="submitActionForm"
      :footer="actionPanelHasFooter"
      :submit-test-id="
        activePanelKind === 'edit'
          ? 'view-edit-submit'
          : activePanelKind === 'grant'
            ? 'view-grant-submit'
            : undefined
      "
      @cancel="cancelActionForm"
    >
      <FormViewAdd
        v-if="activePanelKind === 'add'"
        :key="panelKey('add')"
        ref="actionFormRef"
        @created="refreshAfterPanelSaved"
        @close="closePanel"
      />
      <FormViewEdit
        v-else-if="activePanelKind === 'edit'"
        :key="panelKey('edit', selectedView?.id)"
        ref="actionFormRef"
        :view="selectedView"
        @saved="refreshAfterPanelSaved"
        @close="closePanel"
      />
      <FormViewGrant
        v-else-if="activePanelKind === 'grant'"
        :key="panelKey('grant', selectedView?.id)"
        ref="actionFormRef"
        :view="selectedView"
        @saved="refreshAfterPanelSaved"
        @close="closePanel"
      />
      <FormViewToken
        v-else-if="activePanelKind === 'token'"
        :key="panelKey('token', selectedView?.id)"
        ref="actionFormRef"
        :view="selectedView"
        @created="refreshAfterPanelSaved"
        @close="closePanel"
      />
    </LayoutForm>
  </LayoutReveal>

  <section class="page-section views-page">
    <SectionHeader :title="t('common.labels.views')" :help-text="t('configuration.views.help')" />
    <AppMessage
      v-if="inactiveUserLabels.length > 0"
      severity="warn"
      data-testid="views-inactive-warning"
    >
      {{ t('configuration.views.warnings.inactiveUsers.prefix') }}
      <strong>{{ t('configuration.views.warnings.inactiveUsers.duration') }}</strong
      >:
      <ul class="views-page__inactive-list">
        <li v-for="inactiveUserLabel in inactiveUserLabels" :key="inactiveUserLabel">
          <strong class="sensitive">{{ inactiveUserLabel }}</strong>
        </li>
      </ul>
    </AppMessage>

    <LayoutLoading
      :loading="isFetchingViews && views.length === 0"
      :loading-message="t('configuration.views.loading')"
      :empty="viewsGrouped.length === 0"
      :empty-message="t('configuration.views.empty')"
    >
      <DataGrid
        test-id="views-table"
        :groups="viewsGrouped"
        :sort="viewsSort"
        :row-key="(row) => row.id"
        table-class="views-table"
        cell-align="top"
      >
        <template #header="{ groupSort }">
          <SortableTh
            field="name"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__column-name"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.name') }}
          </SortableTh>
          <SortableTh
            field="scope_name"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__column-name"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.scope') }}
          </SortableTh>
          <th>{{ t('configuration.views.columns.filter') }}</th>
          <th>{{ t('common.labels.users') }}</th>
          <th class="data-grid__th-actions"></th>
        </template>

        <template #row="{ row: view }">
          <td class="data-grid__cell--ellipsis views-page__column-name">
            <span class="sensitive views-page__label-aligned" :title="view.name">{{
              view.name
            }}</span>
          </td>
          <td class="data-grid__cell--ellipsis views-page__column-name">
            <span class="sensitive views-page__label-aligned" :title="view.scope_name">{{
              view.scope_name
            }}</span>
          </td>
          <td>
            <div v-if="view.filtersList.length > 0" class="views-page__filters">
              <TagLabel
                v-for="(filter, filterIndex) in view.filtersList"
                :key="filterIndex"
                :text="filter.name"
                :detail="filter.value"
                text-max-width="100%"
                detail-max-width="3rem"
                variant="primary"
              />
            </div>
            <div v-else>
              <TagLabel :text="t('configuration.views.emptyStates.filters')" />
            </div>
          </td>
          <td>
            <div v-if="grantsByCompany(view.grants).length > 0" class="views-page__users-grouped">
              <template v-for="group in grantsByCompany(view.grants)" :key="group.company">
                <div v-if="group.company" class="views-page__company-header sensitive">
                  {{ group.company }}
                </div>
                <div class="views-page__users-row">
                  <TagLabel
                    v-for="grant in group.members"
                    :key="grant.username"
                    :text="grant.username"
                    :detail="grant.user_department ? `${grant.user_department}` : undefined"
                    text-max-width="5.5rem"
                    detail-max-width="3.5rem"
                    :variant="grantLabelVariant(grant)"
                    :button-label="t('revoke')"
                    button-icon="pi pi-times"
                    sensitive
                    @action="openRevokeUserConfirm(view, grant.username)"
                  />
                </div>
              </template>
            </div>
            <div v-else>
              <TagLabel :text="t('configuration.views.emptyStates.users')" />
            </div>
          </td>
          <td class="data-grid__cell--actions">
            <div class="data-grid__actions">
              <Button
                v-tooltip.top="t('configuration.views.actions.edit.tooltip')"
                type="button"
                icon="pi pi-pencil"
                class="data-grid__btn"
                :data-testid="`views-edit-${view.id}`"
                :aria-label="t('configuration.views.actions.edit.tooltip')"
                @click="openEdit(view)"
              />
              <Button
                v-tooltip.top="t('configuration.views.actions.grant.tooltip')"
                type="button"
                icon="pi pi-users"
                class="data-grid__btn"
                :data-testid="`views-grant-${view.id}`"
                :aria-label="t('configuration.views.actions.grant.tooltip')"
                @click="openGrant(view)"
              />
              <Button
                v-tooltip.top="t('configuration.views.actions.generateAccessToken.tooltip')"
                type="button"
                severity="warn"
                icon="pi pi-key"
                class="data-grid__btn"
                :data-testid="`views-token-${view.id}`"
                :aria-label="t('configuration.views.actions.generateAccessToken.tooltip')"
                @click="openToken(view)"
              />
              <Button
                v-tooltip.top="t('common.actions.delete')"
                type="button"
                severity="danger"
                icon="pi pi-trash"
                class="data-grid__btn"
                :data-testid="`views-delete-${view.id}`"
                :aria-label="t('common.actions.delete')"
                @click="openDeleteConfirm(view)"
              />
            </div>
          </td>
        </template>
      </DataGrid>
    </LayoutLoading>

    <LayoutRightAligned>
      <SegmentedButton
        data-testid="views-add-button"
        icon="pi pi-plus"
        :label="t('common.labels.view')"
        @click="openAdd"
      />
    </LayoutRightAligned>
  </section>

  <!-- ══════════════════════════════════════════════════════════════════════
           Access Tokens
           ══════════════════════════════════════════════════════════════════════ -->
  <section class="page-section views-page">
    <SectionHeader
      :title="t('configuration.views.tokens.title')"
      :help-text="t('configuration.views.tokens.help')"
    />
    <LayoutLoading
      :loading="isFetchingViews && tokens.length === 0"
      :loading-message="t('configuration.views.tokens.loading')"
      :empty="tokensGrouped.length === 0"
      :empty-message="t('configuration.views.tokens.empty')"
    >
      <DataGrid
        test-id="tokens-table"
        :groups="tokensGrouped"
        :sort="tokensSort"
        :row-key="(token) => `${token.view_id}-${token.username}`"
        table-class="tokens-table"
      >
        <template #header="{ groupSort }">
          <SortableTh
            field="view_name"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-view"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.view') }}
          </SortableTh>
          <SortableTh
            field="scope_name"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-scope"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.scope') }}
          </SortableTh>
          <SortableTh
            field="username"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-user"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.username') }}
          </SortableTh>
          <SortableTh
            field="description"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.description') }}
          </SortableTh>
          <SortableTh
            field="created_by"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-created-by"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.createdBy') }}
          </SortableTh>
          <SortableTh
            field="created"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-date"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.created') }}
          </SortableTh>
          <SortableTh
            field="expiry"
            :active="groupSort.field.value"
            :direction="groupSort.direction.value"
            class="views-page__token-column-date"
            @toggle="groupSort.toggle"
          >
            {{ t('common.labels.expiry') }}
          </SortableTh>
          <th class="data-grid__th-actions views-page__token-column-actions"></th>
        </template>

        <template #row="{ row: token }">
          <td class="data-grid__cell--ellipsis">
            <span class="sensitive" :title="token.view_name">{{ token.view_name }}</span>
          </td>
          <td class="data-grid__cell--ellipsis">
            <span class="sensitive" :title="token.scope_name">{{ token.scope_name }}</span>
          </td>
          <td class="data-grid__cell--ellipsis">
            <span :title="token.username">{{ token.username }}</span>
          </td>
          <td class="data-grid__cell--ellipsis">
            <span class="sensitive" :title="token.description">{{ token.description }}</span>
          </td>
          <td class="data-grid__cell--nowrap">
            {{ formatDate(token.created) }}
          </td>
          <td class="data-grid__cell--nowrap">
            <span
              :class="expiryClass(token.expiry)"
              class="views-page__expiry"
              :data-testid="`tokens-expiry-${token.username}`"
            >
              {{ formatDate(token.expiry) }}
            </span>
          </td>
          <td class="data-grid__cell--actions">
            <div class="data-grid__actions">
              <Button
                v-tooltip.top="t('revoke')"
                type="button"
                severity="danger"
                icon="pi pi-trash"
                class="data-grid__btn"
                :data-testid="`tokens-revoke-${token.username}`"
                :aria-label="t('revoke')"
                @click="openRevokeTokenConfirm(token)"
              />
            </div>
          </td>
        </template>
      </DataGrid>
    </LayoutLoading>
  </section>

  <DialogConfirm
    v-model:visible="actionConfirmVisible"
    :title="confirmTitle"
    :icon="confirmIcon"
    :text-confirm-label="confirmLabel"
    :message="confirmDialogMessage"
    :secondary-message="confirmDialogSecondaryMessage"
    @confirm="confirmQueuedAction"
  />

  <!-- Discard-changes gate shown when switching panels while a form is dirty -->
  <DialogConfirm
    v-model:visible="discardVisible"
    :title="t('unsavedTitle')"
    :text-confirm-label="t('common.actions.discard')"
    :message="t('unsavedMessage')"
    @confirm="onConfirmDiscard"
  />
</template>

<style scoped>
/* Constrain view and scope labels so flexible tag columns retain enough room */
.views-page__column-name {
  width: 10rem;
  max-width: 10rem;
}

/* Keep token metadata columns compact and predictable */
.views-page__token-column-view {
  width: 6.428571rem;
}

.views-page__token-column-scope,
.views-page__token-column-created-by {
  width: 10.714286rem;
}

.views-page__token-column-user {
  width: 11.785714rem;
}

.views-page__token-column-date {
  width: 4.571429rem;
}

/* Let the action column shrink to the width of its controls */
.views-page__token-column-actions {
  width: 1%;
}

/* Aligns plain text cells with the top of TagLabel elements in the same row.
   Matches TagLabel's margin-top (1px) + segment padding-top (0.3em at 0.79rem). */
.views-page__label-aligned {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-top: 0.15rem;
}

/* Inline label layout wrappers. Segment styling lives in TagLabel.vue. */

.views-page__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
}

/* User grant grouping. Segment styling lives in TagLabel.vue. */

.views-page__users-grouped {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.views-page__company-header {
  margin-top: 0.15rem;
  font-size: var(--app-font-size-small);
  font-weight: 600;
  color: var(--p-text-muted-color);
}

.views-page__users-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
}

/* ── Token-expiry urgency color-coding (set by expiryClass()) ─────────────── */

.views-page__expiry--red {
  color: var(--p-red-500);
}

.views-page__expiry--yellow {
  color: var(--p-yellow-600);
}

.views-page__expiry--teal {
  color: var(--p-primary-500);
}

/* ── Inactive users warning ───────────────────────────────────────────────── */

.views-page__inactive-list {
  margin: 0;
  padding-top: 0.25rem;
  padding-left: 2.25rem;
}

:deep(.views-table .data-grid__cell--actions) {
  width: 9.3rem;
  min-width: 5.5rem;
  overflow: hidden;
}

:deep(.tokens-table .data-grid__cell--actions) {
  width: 0rem;
  min-width: 0rem;
  overflow: hidden;
}
</style>
