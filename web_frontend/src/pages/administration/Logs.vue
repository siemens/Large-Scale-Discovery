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
 * Admin / Query Logs page.
 *
 * Parity migration of legacy `web_frontend/src/pages/admin/logs`. Renders the
 * SQL query log through the A4 `<DataTable>` (no bespoke grid), with the
 * largest column set in the app. Datetimes go through the A2 `formatDate`
 * formatter; the timestamp column filters on the formatted string (`dateText`),
 * every other column is a plain text filter, matching the legacy Tabulator
 * `headerFilter: "input"` behavior.
 *
 * Severity coloring: the legacy Query Logs page has NO severity/level field and
 * applied no conditional row classes, so none are used here. The matrix #6
 * conditional-row-class capability is retained by `<DataTable>` but is
 * intentionally not exercised on this page (confirmed against the legacy grid).
 */

import { storeToRefs } from 'pinia'
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable from '@/components/datatable/DataTable.vue'
import AppMessage from '@/components/elements/AppMessage.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import { useSqlLogsStore } from '@/stores/sqlLogs'

import type { ColumnDef } from '@/components/datatable/dataTableTypes'
import type { SqlResponse } from '@/types/sql'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    admin: {
      logs: {
        title: 'Query Logs',
        help: 'This section lists SQL queries executed by users across the different scan scopes.',
        error: 'Could not load query logs.',
        info: 'Showing the most recent 2,000 entries. Use the database directly for older logs.',
        columns: {
          query: 'Query',
          table: 'Table',
          user: 'User',
          timestamp: 'Timestamp',
          execTime: 'Exec Time',
          totalTime: 'Total Time',
          results: 'Results',
        },
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })

const sqlLogsStore = useSqlLogsStore()
const { sqlLogs, isFetching } = storeToRefs(sqlLogsStore)

// ── Reactive state ────────────────────────────────────────────
// Column definitions for the query-logs grid
const columns = computed<ColumnDef<SqlResponse>[]>(() => [
  {
    field: 'query',
    header: t('admin.logs.columns.query'),
    filterable: true,
    filterType: 'text',
    minWidth: '10rem',
    maxWidth: '30rem',
    sensitive: true,
  },
  {
    field: 'db_table',
    header: t('admin.logs.columns.table'),
    filterable: true,
    filterType: 'text',
    maxWidth: '10rem',
  },
  {
    field: 'db_user',
    header: t('admin.logs.columns.user'),
    filterable: true,
    filterType: 'text',
    maxWidth: '10rem',
    sensitive: true,
  },
  {
    field: 'db_name',
    header: t('common.labels.database'),
    filterable: true,
    filterType: 'text',
    maxWidth: '10rem',
  },
  {
    field: 'query_timestamp',
    header: t('admin.logs.columns.timestamp'),
    formatter: 'datetime',
    filterable: true,
    maxWidth: '11.4rem',
    minWidth: '11.4rem',
  },
  {
    field: 'query_duration_string',
    header: t('admin.logs.columns.execTime'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    minWidth: '4rem',
    maxWidth: '10rem',
  },
  {
    field: 'total_duration_string',
    header: t('admin.logs.columns.totalTime'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    minWidth: '4rem',
    maxWidth: '10rem',
  },
  {
    field: 'query_results',
    header: t('admin.logs.columns.results'),
    filterable: true,
    filterType: 'text',
    filterFunnel: false,
    minWidth: '4rem',
    maxWidth: '10rem',
  },
  {
    field: 'client_name',
    header: t('common.labels.client'),
    filterable: true,
    filterType: 'text',
    maxWidth: '15rem',
  },
])

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(async () => {
  try {
    await sqlLogsStore.fetchLogs()
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }
})
</script>

<template>
  <section class="page-section logs-page">
    <SectionHeader
      class="logs-page__title"
      :title="t('admin.logs.title')"
      :help-text="t('admin.logs.help')"
    />

    <AppMessage severity="info" :message="t('admin.logs.info')" />

    <DataTable :rows="sqlLogs" :columns="columns" :loading="isFetching" column-borders />
  </section>
</template>

<style scoped>
/* Ensure the section fills the available content width */
.logs-page {
  width: 100%;
}
</style>
