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
 * Accessible Views panel on the home page.
 *
 * Renders the granted-view cards: 6 metadata rows, 3-color progress bar
 * inside the table, paused dimmer over the description area only, and a CSS
 * 3D flip for connection details. The "Connection Details" button lives
 * outside the flip area (card footer) and stays visible during the flip.
 */
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import ProgressBar3 from '@/components/elements/ProgressBar3.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import TagLabel from '@/components/elements/TagLabel.vue'
import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import { daysAgo, formatDate, formatDateTime, isZeroDate } from '@/i18n/format'
import { useAuthStore } from '@/stores/auth'
import { useViewsStore } from '@/stores/views'

import type { GrantedViewResponse } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    title: 'Accessible Views',
    loading: 'Loading accessible views...',
    empty: 'No access granted yet.',
    copy: 'Copy',
    notSynced: 'Targets not synchronized',
    lastSync: 'Last synchronized: {date}',
    retention: {
      all: 'Retaining all cycles',
      lastN: 'Retaining last {count} cycles',
    },
    summary: {
      type: {
        label: 'Scope Type',
        value: '{type} Input',
      },
      targets: {
        value: '{count} IPs',
      },
      scopeCreated: {
        label: 'Scope Created',
      },
      cycle: {
        label: 'Current Cycle',
        value: 'Cycle {cycle}',
      },
      cycleStarted: {
        label: 'Cycle Started',
      },
      retention: {
        label: 'Cycles',
      },
    },
    connection: {
      title: 'Connection',
      fields: {
        server: {
          label: 'Database Server',
        },
        accessibleViews: {
          label: 'Accessible Views',
        },
      },
    },
    status: {
      paused: 'Paused',
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const viewsStore = useViewsStore()
const { grantedViews } = storeToRefs(viewsStore)

const authStore = useAuthStore()

// Connection details flip independently for each granted view card
const isFlipped = ref<Set<number>>(new Set())

// ── Computed ──────────────────────────────────────────────────
const currentUserEmail = computed(() => authStore.profile?.email ?? '')

// Demo accounts see the normal empty state without requesting or exposing service data
const visibleGrantedViews = computed(() => (authStore.isDemo ? [] : grantedViews.value))

function isSyncStale(view: GrantedViewResponse): boolean {
  const lastSync = view.scan_scope.last_sync
  if (isZeroDate(lastSync)) return false
  return Date.now() - new Date(lastSync).getTime() > 8 * 24 * 60 * 60 * 1000
}

function syncText(view: GrantedViewResponse): string {
  const lastSync = view.scan_scope.last_sync
  if (isZeroDate(lastSync)) return t('notSynced')
  return t('lastSync', { date: formatDateTime(lastSync) })
}

function retentionText(view: GrantedViewResponse): string {
  const retention = view.scan_scope.cycles_retention
  return retention === -1 ? t('retention.all') : t('retention.lastN', { count: retention })
}

function viewNames(view: GrantedViewResponse): string[] {
  return view.view_names
    .split(',')
    .map((name) => name.trim())
    .filter(Boolean)
    .sort((first, second) => first.localeCompare(second))
}

function flip(index: number): void {
  const next = new Set(isFlipped.value)
  if (next.has(index)) next.delete(index)
  else next.add(index)
  isFlipped.value = next
}

function toggleConnectionDetails(index: number, event: Event): void {
  flip(index)
  const target = event.currentTarget
  if (target instanceof HTMLElement) {
    target.blur()
  }
}
</script>

<template>
  <LayoutLoading
    :loading="visibleGrantedViews === null"
    :loading-message="t('loading')"
    :empty="visibleGrantedViews !== null && visibleGrantedViews.length === 0"
    :empty-message="t('empty')"
  >
    <div class="widget-granted__grid" data-testid="views-list" role="list" :aria-label="t('title')">
      <article
        v-for="(view, index) in visibleGrantedViews"
        :key="`${view.scan_scope?.id ?? index}-${index}`"
        class="widget-granted__card"
        :data-testid="`view-card-${view.id}`"
        role="listitem"
      >
        <!-- Card content area -->
        <div class="widget-granted__content">
          <!-- Header: view name + scope badge — stays fixed during flip -->
          <div class="widget-granted__header">
            <div class="sensitive widget-granted__header-inner">
              <span class="widget-granted__name">{{ view.name }}</span>
              <span class="widget-granted__db">
                <span>{{ view.scan_scope.name }}</span>
                <i class="pi pi-database" aria-hidden="true" />
              </span>
            </div>
          </div>

          <div class="widget-granted__meta" />

          <!-- Description area: flips between front/back -->
          <div class="widget-granted__description">
            <div class="widget-granted__flip-viewport">
              <div
                class="widget-granted__flip-rotor"
                :class="{ 'widget-granted__flip-rotor--flipped': isFlipped.has(index) }"
              >
                <!-- Front face: metadata table + progress -->
                <div
                  class="widget-granted__face widget-granted__face--front"
                  :aria-hidden="isFlipped.has(index)"
                >
                  <table class="widget-granted__table">
                    <tbody>
                      <tr v-tooltip.left="t('summary.type.label')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-box" aria-hidden="true" />
                        </td>
                        <td>
                          {{
                            view.scan_scope.type
                              ? t('summary.type.value', {
                                  type:
                                    view.scan_scope.type.charAt(0).toUpperCase() +
                                    view.scan_scope.type.slice(1),
                                })
                              : '—'
                          }}
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('common.labels.targets')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-bullseye" aria-hidden="true" />
                        </td>
                        <td>
                          <span class="widget-granted__targets-text">{{
                            t('summary.targets.value', {
                              count: new Intl.NumberFormat().format(view.scan_scope.size),
                            })
                          }}</span>
                          <template v-if="view.scan_scope.attributes?.sync">
                            &nbsp;
                            <span
                              v-tooltip.top="syncText(view)"
                              class="widget-granted__sync-trigger"
                              :title="syncText(view)"
                            >
                              <i
                                class="pi pi-sync widget-granted__sync-icon"
                                :class="
                                  isZeroDate(view.scan_scope.last_sync)
                                    ? 'widget-granted__sync-default'
                                    : isSyncStale(view)
                                      ? 'widget-granted__sync-red'
                                      : 'widget-granted__sync-teal'
                                "
                                aria-hidden="true"
                              />
                            </span>
                          </template>
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('summary.scopeCreated.label')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-calendar-plus" aria-hidden="true" />
                        </td>
                        <td>{{ formatDate(view.scan_scope.created) }}</td>
                      </tr>
                      <tr v-tooltip.left="t('summary.cycle.label')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-refresh" aria-hidden="true" />
                        </td>
                        <td>
                          {{
                            t('summary.cycle.value', {
                              cycle: new Intl.NumberFormat().format(view.scan_scope.cycle),
                            })
                          }}
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('summary.cycleStarted.label')">
                        <td class="widget-granted__icon-cell">
                          <span class="widget-granted__stack-icon">
                            <i class="pi pi-refresh" aria-hidden="true" />
                            <i
                              class="pi pi-calendar widget-granted__corner-icon"
                              aria-hidden="true"
                            />
                          </span>
                        </td>
                        <td>
                          {{ formatDate(view.scan_scope.cycle_started) }} |
                          {{ daysAgo(view.scan_scope.cycle_started) }}
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('summary.retention.label')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-eraser" aria-hidden="true" />
                        </td>
                        <td>{{ retentionText(view) }}</td>
                      </tr>
                      <tr>
                        <td class="widget-granted__progress-cell" colspan="2">
                          <ProgressBar3
                            :progress="view.scan_scope.cycle_progress ?? [0, 0, 0]"
                            :total="100"
                            size="sm"
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>

                  <!-- Paused dimmer — only over the description area -->
                  <div
                    v-if="!view.scan_scope.enabled && !isFlipped.has(index)"
                    class="widget-granted__paused-overlay"
                    :aria-label="t('status.paused')"
                  >
                    <i class="pi pi-pause widget-granted__paused-icon" aria-hidden="true" />
                    <br />
                    <span class="widget-granted__paused-label">{{ t('status.paused') }}</span>
                  </div>
                </div>

                <!-- Back face: connection details -->
                <div
                  class="widget-granted__face widget-granted__face--back"
                  :aria-hidden="!isFlipped.has(index)"
                >
                  <table class="widget-granted__table">
                    <tbody>
                      <tr v-tooltip.left="t('connection.fields.server.label')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-server" aria-hidden="true" />
                        </td>
                        <td>
                          <CodeSnippet
                            :text="`${view.scan_scope.connection.host}:${view.scan_scope.connection.port}`"
                            :text-copy-label="t('copy')"
                            class="sensitive"
                          />
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('common.labels.database')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-database" aria-hidden="true" />
                        </td>
                        <td>
                          <CodeSnippet
                            :text="view.scan_scope.connection.database"
                            :text-copy-label="t('copy')"
                            class="sensitive"
                          />
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('common.labels.username')">
                        <td class="widget-granted__icon-cell">
                          <i class="pi pi-user" aria-hidden="true" />
                        </td>
                        <td>
                          <CodeSnippet
                            :text="currentUserEmail"
                            :text-copy-label="t('copy')"
                            class="sensitive"
                          />
                        </td>
                      </tr>
                      <tr v-tooltip.left="t('connection.fields.accessibleViews.label')">
                        <td class="widget-granted__icon-cell widget-granted__icon-cell--top">
                          <i class="pi pi-table" aria-hidden="true" />
                        </td>
                        <td>
                          <TagLabel
                            v-for="viewName in viewNames(view)"
                            :key="viewName"
                            :text="viewName"
                            variant="primary"
                            sensitive
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Card footer: Connection Details button — stays outside the flip area -->
        <div class="widget-granted__extra-spacer">
          <div class="widget-granted__extra">
            <SegmentedButton
              v-tooltip.bottom="view.scan_scope.name"
              type="button"
              :severity="isFlipped.has(index) ? 'primary' : 'secondary'"
              size="small"
              class="widget-granted__connection-button"
              :data-testid="`view-flip-${index}`"
              icon="pi pi-database"
              :label="t('connection.title')"
              :aria-pressed="isFlipped.has(index)"
              @click="toggleConnectionDetails(index, $event)"
            />
          </div>
        </div>
      </article>
    </div>
  </LayoutLoading>
</template>

<style scoped>
.widget-granted__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(13.75rem, 1fr));
  gap: 1rem;
  width: 100%;
}

@media (max-width: 1100px) {
  .widget-granted__grid {
    grid-template-columns: repeat(3, minmax(13.75rem, 1fr));
  }
}
@media (max-width: 820px) {
  .widget-granted__grid {
    grid-template-columns: repeat(2, minmax(13.75rem, 1fr));
  }
}
@media (max-width: 560px) {
  .widget-granted__grid {
    grid-template-columns: 1fr;
  }
}

/* Card: Semantic UI .card parity — shadow, rounded, flex column */
.widget-granted__card {
  display: flex;
  flex-direction: column;
  padding: 0.3em;
  border: 0;
  border-radius: 0.28571429rem;
  background: var(--p-surface-0);
  box-shadow:
    0 1px 3px 0 var(--p-content-border-color),
    0 0 0 1px var(--p-content-border-color);
}

.widget-granted__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 1rem 1rem;
}

/* ── Header ────────────────────────────────────────────────────────────── */
.widget-granted__header {
  padding: 0.5em 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 600;
  line-height: 1.28571429em;
  color: var(--p-surface-700);
}

.widget-granted__header-inner {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.25rem;
}

.widget-granted__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  line-height: 1;
  color: var(--p-text-color);
  font-size: 1.3rem;
}

.widget-granted__db {
  display: inline-flex;
  align-items: center;
  gap: 0.25em;
  flex-shrink: 0;
  padding-right: 2px;
  white-space: nowrap;
  line-height: 1;
  color: var(--p-surface-700);
  font-size: 1rem;
}

.widget-granted__db span {
  max-width: 7rem;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* PrimeIcons pin `.pi` to font-size 1rem, so the database glyph renders smaller
   than the 1.3rem badge text beside it — force it to track the badge size.
   (Vertical centering is handled by the flex container above.) */
.widget-granted__db .pi {
  font-size: inherit;
}

.widget-granted__meta {
  min-height: 0;
}

/* ── Description area (flippable) ──────────────────────────────────────── */
.widget-granted__description {
  flex: 1;
  position: relative;
  margin-top: 0.5em;
}

.widget-granted__flip-viewport {
  width: 100%;
  perspective: 1000px;
}

.widget-granted__flip-rotor {
  display: grid; /* both faces share one cell → height = max(front, back) */
  width: 100%;
  transform-style: preserve-3d;
  transition: transform 0.2s ease;
}

.widget-granted__flip-rotor--flipped {
  transform: rotateY(180deg);
}

/* Both faces sit in the same grid cell so the rotor never shrinks below the
   taller face. The card can grow when the back is taller, but never shrink. */
.widget-granted__face {
  grid-area: 1 / 1;
  width: 100%;
  min-width: 0;
  background: var(--p-surface-0);
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

.widget-granted__face,
.widget-granted__face * {
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

.widget-granted__face--back {
  transform: rotateY(180deg);
  pointer-events: none;
}

.widget-granted__flip-rotor--flipped .widget-granted__face--front {
  pointer-events: none;
}

.widget-granted__flip-rotor--flipped .widget-granted__face--back {
  pointer-events: auto;
}

/* ── Table: minimal borderless collapsing layout ──────────────────────── */
.widget-granted__table {
  width: 100%;
  max-width: 100%;
  table-layout: fixed;
  border-collapse: separate;
  border-spacing: 0;
  background: var(--p-surface-0);
  margin-bottom: 0;
}

.widget-granted__table td {
  padding: 0.78571429em;
  text-align: inherit;
  vertical-align: inherit;
  border-top: 1px solid var(--p-content-border-color);
}

.widget-granted__table tr:first-child td {
  border-top: none;
}

.widget-granted__table td + td {
  border-left: 1px solid var(--p-content-border-color);
}

/* !important overrides the table cell's default padding so the icon column
   stays narrowly left-aligned regardless of the browser's UA stylesheet. */
.widget-granted__icon-cell {
  width: 28px;
  padding-left: 10px !important;
}

.widget-granted__stack-icon {
  position: relative;
  display: inline-block;
  line-height: 1;
}

.widget-granted__corner-icon {
  position: absolute;
  bottom: -0.2rem;
  right: -0.45rem;
  font-size: 0.55rem;
}

.widget-granted__targets-text {
  vertical-align: middle;
}

.widget-granted__sync-trigger {
  display: inline-block;
  line-height: 1rem;
  vertical-align: middle;
  cursor: default;
}

.widget-granted__sync-icon {
  display: inline-block;
  font-size: 0.75em;
  line-height: 1;
  vertical-align: middle;
}

.widget-granted__sync-red {
  color: var(--p-red-500);
}
.widget-granted__sync-teal {
  color: var(--p-primary-500);
}

.widget-granted__progress-cell {
  text-align: center;
  border-top: 1px solid var(--p-content-border-color) !important;
  padding-top: 2rem !important;
}

.widget-granted__progress-cell:last-of-type {
  padding: 1.8rem 0 1rem;
}

/* ── Paused dimmer (over description area only) ───────────────────────── */
.widget-granted__paused-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--app-color-offwhite);
  text-align: center;
}

.widget-granted__paused-icon {
  font-size: 2rem;
  font-weight: 900;
  color: var(--p-surface-700);
}

.widget-granted__paused-label {
  margin-top: 0.4rem;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--p-surface-700);
}

/* ── Connection details (back face) ───────────────────────────────────── */

.widget-granted__face--back .widget-granted__table tr:not(:last-child) td {
  padding-top: 0.35em;
  padding-bottom: 0.35em;
}

.widget-granted__icon-cell--top {
  vertical-align: top;
  padding-top: 4px !important;
}

/* ── Card footer: Connection Details button ───────────────────────────── */
.widget-granted__extra-spacer {
  padding: 1rem;
  padding-top: 0;
}

.widget-granted__extra {
  padding-top: 1rem;
  border-top: 1px solid var(--p-content-border-color);
  text-align: center;
}
</style>
