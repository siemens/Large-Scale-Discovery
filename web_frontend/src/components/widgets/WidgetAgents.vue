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
 * Home-page scan agents panel: scope-grouped agent cards with avatar, hardware
 * stats, last-seen color ladder, tasks hover popup, and admin delete action.
 */
import { storeToRefs } from 'pinia'
import Badge from 'primevue/badge'
import Button from 'primevue/button'
import Popover from 'primevue/popover'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import http from '@/api/http'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import LinuxIcon from '@/components/elements/LinuxIcon.vue'
import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import { useToast } from '@/composables/useToast'
import { formatBytes, formatDateTime, formatRelative } from '@/i18n/format'
import { agentAvatarDataUri } from '@/lib/avatar'
import { APP_PRIMARY_500, APP_RED_500, APP_SECONDARY_700, APP_YELLOW_500 } from '@/lib/colors'
import { useAgentsStore } from '@/stores/agents'
import { useAuthStore } from '@/stores/auth'

import type { AgentGroupResponse, AgentResponse } from '@/types/agents'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    lastSeen: 'Last seen',
    agentsPanel: {
      cpuCoresGhz: '{count} core × {ghz} GHz | {count} cores × {ghz} GHz',
      cpuCores: '{count} core | {count} cores',
      actions: {
        discard: {
          success: 'Scan agent stats removed.',
          error: 'Could not remove scan agent stats.',
          title: 'Discard Scan Agent',
          question: 'Are you sure you want to discard the scan agent',
          message: 'This will remove the cached statistics of this scan agent.',
        },
      },
      loading: 'Loading agents...',
      empty: 'No agents registered yet.',
      title: 'Scan Agents',
      shared: 'Shared',
      activeTasks: 'Active Tasks',
      taskTypes: {
        discovery: 'Discovery',
        banner: 'Banner',
        nfs: 'NFS',
        nuclei: 'Nuclei',
        smb: 'SMB',
        ssh: 'SSH',
        ssl: 'SSL',
        webcrawler: 'Webcrawler',
        webenum: 'Webenum',
      },
      commit: 'Commit',
      build: 'Build',
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const agentsStore = useAgentsStore()
const { agentGroups } = storeToRefs(agentsStore)
const authStore = useAuthStore()
const { isAdmin, isDemo } = storeToRefs(authStore)

// Demo accounts see the normal empty state without requesting or exposing agent data
const visibleAgentGroups = computed((): AgentGroupResponse[] | null =>
  isDemo.value ? [] : agentGroups.value,
)

// ── Reactive state ────────────────────────────────────────────
// Secondary details (task counts, build info, tool versions) are revealed by
// hovering the avatar image or focusing its accessible wrapper. Pointer events
// intentionally exclude the overlaid shared-status badge.
const tasksPopoverRef = ref<InstanceType<typeof Popover>>()
const taskPopoverAgent = ref<AgentResponse | null>(null)

// Removing an agent's stats is irreversible, so it routes through
// <DialogConfirm> like every other destructive action in the app.
const deleteConfirmVisible = ref(false)
const agentPendingDelete = ref<AgentResponse | null>(null)

// ── Computed ──────────────────────────────────────────────────
// Sort groups alphabetically; within each group move stale agents to rear.
const sortedAgentGroups = computed((): AgentGroupResponse[] => {
  if (!visibleAgentGroups.value) return []
  return [...visibleAgentGroups.value]
    .sort((a, b) => a.scope_name.localeCompare(b.scope_name, undefined, { numeric: true }))
    .map((g) => ({
      ...g,
      agents: [...(g.agents ?? [])].sort((x, y) => {
        const xStale = agentColor(x.last_seen) === APP_SECONDARY_700
        const yStale = agentColor(y.last_seen) === APP_SECONDARY_700
        return xStale === yStale ? 0 : xStale ? 1 : -1
      }),
    }))
})

// ── Functions ─────────────────────────────────────────────────
function agentColor(lastSeen: string): string {
  // Derive the elapsed time since last-seen at each granularity
  const ms = Date.now() - new Date(lastSeen).getTime()
  const minutes = ms / 60_000
  const hours = minutes / 60
  const days = hours / 24

  // Walk the freshness ladder from most-stale (grey) to most-recent (teal)
  if (days >= 14) return APP_SECONDARY_700
  if (days >= 7) return APP_RED_500
  if (days >= 2) return APP_RED_500
  if (days >= 1 && minutes > 30) return APP_YELLOW_500
  if (hours < 2 && hours >= 1) return APP_YELLOW_500
  if (hours >= 1) return APP_YELLOW_500
  if (minutes > 30) return APP_YELLOW_500
  return APP_PRIMARY_500
}

function canDelete(lastSeen: string): boolean {
  // The delete affordance appears once an agent has been silent over 30 minutes
  const minutes = (Date.now() - new Date(lastSeen).getTime()) / 60_000
  return minutes > 30
}

function hasToolVersionDetails(agent: AgentResponse): boolean {
  return agent.version_nmap !== undefined
}

// DiceBear bottts robot avatar, seeded by the agent name — restoring the
// legacy agents.js avatar. Generated offline and rendered as an image data URI.
function agentAvatar(agent: AgentResponse): string {
  return agentAvatarDataUri(agent.name || String(agent.id))
}

function platformType(platform: string): 'windows' | 'linux' | 'other' {
  const normalizedPlatform = platform.toLowerCase()
  if (normalizedPlatform.includes('windows')) return 'windows'
  if (normalizedPlatform.includes('ubuntu') || normalizedPlatform.includes('linux')) return 'linux'
  return 'other'
}

function cpuTooltip(agent: AgentResponse): string | undefined {
  if (!agent.cpu_cores) return undefined
  return agent.cpu_mhz
    ? t(
        'agentsPanel.cpuCoresGhz',
        { count: agent.cpu_cores, ghz: (agent.cpu_mhz / 1000).toFixed(1) },
        agent.cpu_cores,
      )
    : t('agentsPanel.cpuCores', { count: agent.cpu_cores }, agent.cpu_cores)
}

function memoryTooltip(agent: AgentResponse): string | undefined {
  const bytes = agent.memory_bytes ?? agent.memory_total
  return bytes ? formatBytes(bytes) : undefined
}

function osTooltip(agent: AgentResponse): string {
  return [agent.platform, agent.platform_family, agent.platform_version]
    .filter(Boolean)
    .join(' ')
    .trim()
}

function openTasksPopover(event: Event, agent: AgentResponse): void {
  if (!agent.tasks) return
  taskPopoverAgent.value = agent
  tasksPopoverRef.value?.show(event)
}

function closeTasksPopover(): void {
  tasksPopoverRef.value?.hide()
  taskPopoverAgent.value = null
}

function openDeleteConfirm(agent: AgentResponse, group: AgentGroupResponse): void {
  if (!group.scope_id || !group.scope_name) return
  agentPendingDelete.value = agent
  deleteConfirmVisible.value = true
}

async function deletePendingAgent(): Promise<void> {
  const agent = agentPendingDelete.value
  if (!agent) return
  try {
    await http.post(
      '/agent/delete',
      { id: agent.id },
      { errorToastMessage: t('agentsPanel.actions.discard.error') },
    )

    // Defer the success toast until the backend has acknowledged the delete AND
    // the agent list has been refetched. fetchAgents stores its own error
    // rather than throwing, so the toast still lands.
    await agentsStore.fetchAgents()
    toast.success(t('agentsPanel.actions.discard.success'))
  } catch {
    // The shared HTTP client already displayed the contextual deletion error
  } finally {
    agentPendingDelete.value = null
  }
}
</script>

<template>
  <LayoutLoading
    :loading="visibleAgentGroups === null"
    :loading-message="t('agentsPanel.loading')"
    :empty="visibleAgentGroups !== null && sortedAgentGroups.length === 0"
    :empty-message="t('agentsPanel.empty')"
  >
    <div
      class="widget-agents"
      data-testid="agents-list"
      role="list"
      :aria-label="t('agentsPanel.title')"
    >
      <template v-for="group in sortedAgentGroups" :key="group.scope_id">
        <div
          v-for="(agent, idx) in group.agents"
          :key="agent.id"
          class="widget-agents__card"
          :class="{ 'widget-agents__card--first': idx === 0 }"
          :data-testid="`agent-card-${agent.id}`"
          role="listitem"
        >
          <!-- Scope-name headline: shown ABOVE the agent body for the first agent
               in each group, blanked (but height-reserved) for siblings so every
               tile in a group is the same height. Tooltip carries the full DB id,
               positioned right-center to match legacy data-position. -->
          <div
            v-tooltip.top="idx === 0 ? group.scope_db : ''"
            class="widget-agents__scope sensitive"
            :class="{ 'widget-agents__scope--owned': group.scope_ownership }"
            :aria-label="idx === 0 ? group.scope_name : undefined"
            :aria-hidden="idx === 0 ? undefined : 'true'"
          >
            <template v-if="idx === 0">{{ group.scope_name }}</template>
            <template v-else>&nbsp;</template>
          </div>

          <div class="widget-agents__body">
            <div
              class="widget-agents__avatar"
              :class="{ 'widget-agents__avatar--has-popup': !!agent.tasks }"
              :style="{
                filter: agentColor(agent.last_seen) === APP_SECONDARY_700 ? 'grayscale(100%)' : '',
                opacity: agentColor(agent.last_seen) === APP_SECONDARY_700 ? '70%' : '',
              }"
              :aria-label="agent.name || agent.host"
              :aria-describedby="agent.tasks ? `agent-tasks-popup` : undefined"
              role="img"
              :tabindex="agent.tasks ? 0 : undefined"
              @focus="openTasksPopover($event, agent)"
              @blur="closeTasksPopover"
            >
              <img
                class="widget-agents__avatar-img"
                :src="agentAvatar(agent)"
                alt=""
                aria-hidden="true"
                @mouseenter="openTasksPopover($event, agent)"
                @mouseleave="closeTasksPopover"
              />
              <Badge
                v-if="agent.shared"
                v-tooltip.right="t('agentsPanel.shared')"
                class="widget-agents__shared-badge"
                severity="primary"
                aria-hidden="true"
              >
              </Badge>
            </div>

            <div
              class="widget-agents__info"
              :class="{
                'widget-agents__info--stale': agentColor(agent.last_seen) === APP_SECONDARY_700,
              }"
              :style="{
                color:
                  agentColor(agent.last_seen) === APP_SECONDARY_700
                    ? agentColor(agent.last_seen)
                    : '',
              }"
            >
              <div v-tooltip.top="`${agent.host} ${agent.ip}`" class="widget-agents__name">
                <strong>{{ agent.name || '&nbsp;' }}</strong>
              </div>

              <div class="widget-agents__stats">
                <span v-tooltip.top="cpuTooltip(agent)">
                  <!-- Filled CPU icon — solid silhouette to match the visual weight of
                     the Microsoft/Server OS glyphs sitting next to it. -->
                  <svg class="widget-agents__hw-icon" viewBox="0 0 24 24" aria-hidden="true">
                    <rect x="7.25" y="1.5" width="1.5" height="3.5" />
                    <rect x="11.25" y="1.5" width="1.5" height="3.5" />
                    <rect x="15.25" y="1.5" width="1.5" height="3.5" />
                    <rect x="7.25" y="19" width="1.5" height="3.5" />
                    <rect x="11.25" y="19" width="1.5" height="3.5" />
                    <rect x="15.25" y="19" width="1.5" height="3.5" />
                    <rect x="1.5" y="7.25" width="3.5" height="1.5" />
                    <rect x="1.5" y="11.25" width="3.5" height="1.5" />
                    <rect x="1.5" y="15.25" width="3.5" height="1.5" />
                    <rect x="19" y="7.25" width="3.5" height="1.5" />
                    <rect x="19" y="11.25" width="3.5" height="1.5" />
                    <rect x="19" y="15.25" width="3.5" height="1.5" />
                    <rect x="5" y="5" width="14" height="14" rx="1.5" />
                  </svg>
                  <span
                    class="widget-agents__stat-label"
                    :style="{
                      color:
                        agent.cpu_rate > 90 && agentColor(agent.last_seen) !== APP_SECONDARY_700
                          ? APP_RED_500
                          : '',
                    }"
                    >{{ Math.floor(agent.cpu_rate) }}%&nbsp;</span
                  >
                </span>
                <span v-tooltip.top="memoryTooltip(agent)">
                  <!-- Filled RAM icon — solid memory-stick silhouette matching the
                     CPU icon's weight. -->
                  <svg class="widget-agents__hw-icon" viewBox="0 0 24 24" aria-hidden="true">
                    <rect x="1.5" y="3" width="21" height="14" rx="1" />
                    <rect x="2.5" y="17.5" width="2.5" height="4" />
                    <rect x="6" y="17.5" width="2.5" height="4" />
                    <rect x="9.5" y="17.5" width="2.5" height="4" />
                    <rect x="13" y="17.5" width="2.5" height="4" />
                    <rect x="16.5" y="17.5" width="2.5" height="4" />
                    <rect x="20" y="17.5" width="2.5" height="4" />
                  </svg>
                  <span
                    class="widget-agents__stat-label"
                    :style="{
                      color:
                        agent.memory_rate > 90 && agentColor(agent.last_seen) !== APP_SECONDARY_700
                          ? APP_RED_500
                          : '',
                    }"
                    >{{ Math.floor(agent.memory_rate) }}%&nbsp;</span
                  >
                </span>
                <span
                  v-if="agent.platform"
                  v-tooltip.right="osTooltip(agent)"
                  class="widget-agents__hw-icon-wrap"
                >
                  <LinuxIcon v-if="platformType(agent.platform) === 'linux'" />
                  <svg v-else-if="platformType(agent.platform) === 'windows'" viewBox="0 0 24 24">
                    <rect x="1.5" y="1.5" width="9.5" height="9.5" rx=".5" />
                    <rect x="13" y="1.5" width="9.5" height="9.5" rx=".5" />
                    <rect x="1.5" y="13" width="9.5" height="9.5" rx=".5" />
                    <rect x="13" y="13" width="9.5" height="9.5" rx=".5" />
                  </svg>
                  <svg v-else viewBox="0 0 24 24">
                    <rect x="1.5" y="2" width="21" height="14" rx="1.5" />
                    <rect x="9" y="17" width="6" height="2.5" />
                    <rect x="6.5" y="20" width="11" height="2" rx="1" />
                  </svg>
                </span>
                <span class="widget-agents__delete-wrap">
                  <Button
                    v-if="(isAdmin || group.scope_ownership) && canDelete(agent.last_seen)"
                    v-tooltip.top="t('common.actions.discard')"
                    type="button"
                    severity="secondary"
                    icon="pi pi-eraser"
                    class="widget-agents__delete-btn"
                    data-testid="agent-delete-btn"
                    :aria-label="t('common.actions.discard')"
                    @click="openDeleteConfirm(agent, group)"
                  />
                </span>
              </div>

              <div class="widget-agents__last-seen">
                <span v-tooltip.bottom="formatDateTime(agent.last_seen)">
                  {{ t('lastSeen') }}
                  <strong :style="{ color: agentColor(agent.last_seen) }">
                    {{ formatRelative(agent.last_seen) }}
                  </strong>
                </span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <Popover id="agent-tasks-popup" ref="tasksPopoverRef">
      <div v-if="taskPopoverAgent?.tasks" class="widget-agents__tasks" data-testid="agent-tasks">
        <table class="widget-agents__tasks-table">
          <caption>
            {{
              t('agentsPanel.activeTasks')
            }}
          </caption>
          <tbody>
            <tr>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Discovery,
                }"
              >
                {{ taskPopoverAgent.tasks.Discovery || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Discovery,
                }"
              >
                {{ t('agentsPanel.taskTypes.discovery') }}
              </td>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Banner,
                }"
              >
                {{ taskPopoverAgent.tasks.Banner || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Banner,
                }"
              >
                {{ t('agentsPanel.taskTypes.banner') }}
              </td>
            </tr>
            <tr>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Nfs,
                }"
              >
                {{ taskPopoverAgent.tasks.Nfs || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Nfs,
                }"
              >
                {{ t('agentsPanel.taskTypes.nfs') }}
              </td>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Nuclei,
                }"
              >
                {{ taskPopoverAgent.tasks.Nuclei || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Nuclei,
                }"
              >
                {{ t('agentsPanel.taskTypes.nuclei') }}
              </td>
            </tr>
            <tr>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Smb,
                }"
              >
                {{ taskPopoverAgent.tasks.Smb || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Smb,
                }"
              >
                {{ t('agentsPanel.taskTypes.smb') }}
              </td>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Ssh,
                }"
              >
                {{ taskPopoverAgent.tasks.Ssh || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Ssh,
                }"
              >
                {{ t('agentsPanel.taskTypes.ssh') }}
              </td>
            </tr>
            <tr>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Ssl,
                }"
              >
                {{ taskPopoverAgent.tasks.Ssl || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Ssl,
                }"
              >
                {{ t('agentsPanel.taskTypes.ssl') }}
              </td>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Webcrawler,
                }"
              >
                {{ taskPopoverAgent.tasks.Webcrawler || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Webcrawler,
                }"
              >
                {{ t('agentsPanel.taskTypes.webcrawler') }}
              </td>
            </tr>
            <tr>
              <td
                class="widget-agents__tasks-count"
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Webenum,
                }"
              >
                {{ taskPopoverAgent.tasks.Webenum || 0 }}×
              </td>
              <td
                :class="{
                  'widget-agents__task--inactive': !taskPopoverAgent.tasks.Webenum,
                }"
              >
                {{ t('agentsPanel.taskTypes.webenum') }}
              </td>
              <td colspan="2"></td>
            </tr>
            <tr v-if="taskPopoverAgent.build_timestamp" class="widget-agents__tasks-build-info">
              <td colspan="4">
                <strong>{{ t('agentsPanel.commit') }}</strong>
                {{ taskPopoverAgent.build_commit.slice(0, 8) }}
                &nbsp;&nbsp;<strong>{{ t('common.labels.api') }}</strong>
                {{ taskPopoverAgent.api_version }}<br />
                <strong>{{ t('agentsPanel.build') }}</strong>
                {{ formatDateTime(taskPopoverAgent.build_timestamp) }}
              </td>
            </tr>
            <tr
              v-if="hasToolVersionDetails(taskPopoverAgent)"
              class="widget-agents__tasks-build-info"
            >
              <td colspan="4">
                <strong>Nmap</strong> {{ taskPopoverAgent.version_nmap || '–' }}
                <template
                  v-if="
                    taskPopoverAgent.version_npcap ||
                    taskPopoverAgent.platform.toLowerCase().includes('windows')
                  "
                >
                  &nbsp;&nbsp;<strong>Npcap</strong>
                  {{ taskPopoverAgent.version_npcap || '–' }}
                </template>
                <br />
                <strong>Sslyze</strong>
                {{ taskPopoverAgent.version_sslyze || '–' }}
                <template v-if="taskPopoverAgent.version_python">
                  &nbsp;&nbsp;<strong>Python</strong>
                  {{ taskPopoverAgent.version_python }}
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Popover>
  </LayoutLoading>

  <DialogConfirm
    v-model:visible="deleteConfirmVisible"
    :title="t('agentsPanel.actions.discard.title')"
    icon="pi pi-eraser"
    :message="
      agentPendingDelete
        ? `${t('agentsPanel.actions.discard.question')} '${agentPendingDelete.name || agentPendingDelete.host}'?`
        : undefined
    "
    :secondary-message="t('agentsPanel.actions.discard.message')"
    :text-confirm-label="t('common.actions.confirm')"
    @confirm="deletePendingAgent"
  />
</template>

<style scoped>
/* Auto-fill grid: cards stretch to fill the container width (1fr per track)
   instead of leaving ragged right-margin space. minmax keeps each tile at
   least the legacy 208px before a new column wraps in. */
.widget-agents {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(208px, 1fr));
  align-items: start;
}

.widget-agents__card {
  display: flex;
  flex-direction: column;
  padding: 0.1875rem 0.5rem;
  margin-top: 0.3125rem;
  margin-bottom: 0.3125rem;
}

/* First agent in each scope group gets the legacy OPENING bracket only:
   left 2px border + top-left and bottom-left corner ticks. The next group's
   opening bracket implicitly terminates this one, so there is no right border
   and no closing-side ticks. */
.widget-agents__card--first {
  background-image:
    linear-gradient(var(--p-surface-300), var(--p-surface-300)),
    linear-gradient(var(--p-surface-300), var(--p-surface-300));
  background-repeat: no-repeat;
  background-size: 8px 2px;
  background-position:
    top left,
    bottom left;
  border-left: 2px solid var(--p-surface-300);
  border-radius: var(--p-border-radius-sm);
}

.widget-agents__scope {
  width: 100%;
  height: 1.4375rem;
  margin-bottom: 0.3125rem;
  overflow: hidden;
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--p-surface-500);
  line-height: 1.4375rem;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Emphasize scopes administered by the current user */
.widget-agents__scope--owned {
  color: var(--p-surface-800);
}

.widget-agents__body {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.widget-agents__avatar {
  position: relative;
  flex-shrink: 0;
  width: 45px;
  height: 45px;
  border-radius: 50%;
}

.widget-agents__avatar--has-popup {
  cursor: pointer;
}

.widget-agents__avatar-img {
  display: block;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  overflow: hidden;
}

.widget-agents__shared-badge {
  position: absolute;
  bottom: -0.1rem;
  right: -0rem;
  width: 0.7rem;
  height: 0.7rem;
  box-shadow: 0 0 0 0.2rem var(--p-surface-0);
}

.widget-agents__info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  width: 100%;
  min-width: 0;
  padding: 0 0 0 0.3em;
}

.widget-agents__name {
  max-width: 100%;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.widget-agents__stats {
  display: flex;
  align-items: center;
  gap: 0.1rem;
  color: var(--p-surface-800);
}

.widget-agents__info--stale .widget-agents__stats {
  color: inherit;
}

.widget-agents__stats > span {
  display: inline-flex;
  align-items: center;
}

.widget-agents__stat-label {
  margin-left: 0.125rem;
}

.widget-agents__last-seen {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Wrapper that pushes the delete button to the right end of the stats row
   without contributing to row height: margin-left:auto claims the leftover
   space; overflow:visible + height:0 means the button overflows its own line
   without stretching the flex row. */
.widget-agents__delete-wrap {
  display: flex;
  align-items: center;
  margin-left: auto;
  height: 0;
  overflow: visible;
}

/* Eraser button: inline ghost icon button inside the stats row.
   !important overrides the PrimeVue Button preset's fixed size, padding,
   background, border, color, and font-size tokens — the preset applies these
   via scoped selectors that outweigh our component-scoped rule without it. */
.widget-agents__delete-btn {
  width: 1.5rem !important;
  height: 1.5rem !important;
  padding: 0 !important;
  background: none !important;
  border: none !important;
  color: var(--p-text-muted-color) !important;
  font-size: 1rem !important;
}

/* !important matches the specificity reason above — overrides the preset icon rule. */
.widget-agents__delete-btn :deep(.p-button-icon) {
  font-size: 1rem !important;
}

/* !important matches the specificity reason above for hover state. */
.widget-agents__delete-btn:hover {
  background: none !important;
  color: var(--p-red-500) !important;
}

.widget-agents__tasks {
  font-size: var(--app-font-size-small);
}

/* ── Hardware icons (inside .widget-agents__stats) ─────────────────── */

/* Filled CPU / RAM glyphs: inline SVG matching the visual weight of the OS
   icon (LinuxIcon / custom Windows four-squares / custom monitor SVG) sitting next to them. Sized to an explicit
   1rem square so all three hardware-row icons render at identical dimensions
   — the previous `1em` resolved against the 0.8rem `.widget-agents__stats` font
   (≈12.8px) and read visibly smaller than the 1rem (16px) PrimeIcon OS
   glyph. `fill: currentColor` inherits the gray/red color from the parent
   span's inline style. */
.widget-agents__hw-icon {
  display: inline-block;
  width: 1rem;
  height: 1rem;
  fill: currentColor;
  vertical-align: middle;
  pointer-events: none;
}

/* OS icon wrapper: sized to match the 1rem CPU/RAM hw-icons so the OS glyph
   reads at the same scale as its neighbours. */
.widget-agents__hw-icon-wrap {
  display: inline-block;
  width: 1rem;
  height: 1rem;
  vertical-align: middle;
}

.widget-agents__hw-icon-wrap :deep(svg) {
  display: block;
  width: 100%;
  height: 100%;
  fill: currentColor;
}

/* ── Popover internals ──────────────────────────────────────── */

.widget-agents__tasks-table {
  width: 100%;
  border-collapse: collapse;
  border: none;
  font-size: var(--app-font-size-small);
}

.widget-agents__tasks-table caption {
  font-weight: 600;
  text-align: center;
  padding: 0.4em 0.6em;
  padding-bottom: 0.6em;
  border-bottom: 1px solid var(--p-surface-200);
}

.widget-agents__tasks-table td {
  padding: 0.4em 0.6em;
}

.widget-agents__tasks-table tbody tr:first-child td {
  padding-top: 0.6em;
}

.widget-agents__tasks-table tbody tr + tr {
  border-top: 1px solid var(--p-surface-200);
}

/* Count cell: narrow gutter so the number reads flush against the task name.
   !important overrides the general td padding rule above that applies via the
   same specificity — both target td inside this table, so !important wins. */
.widget-agents__tasks-count {
  text-align: right;
  padding-right: 0.25rem !important;
  padding-left: 0.25rem !important;
}

/* Task name cells: remove left padding so the name reads as a direct label
   for the count beside it. !important same reason as .widget-agents__tasks-count. */
.widget-agents__tasks-table td:nth-child(2),
.widget-agents__tasks-table td:nth-child(4) {
  padding-left: 0 !important;
}

.widget-agents__tasks-build-info td {
  opacity: 0.35;
  font-size: var(--app-font-size-small);
  text-align: center;
  padding: 0.4em 0;
  line-height: 1.6;
}

.widget-agents__task--inactive {
  opacity: 0.35;
}
</style>
