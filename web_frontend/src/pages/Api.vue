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
 * API page — port of web_frontend/src/pages/api/*.
 *
 * Two concerns, matching the legacy page:
 *   1. API documentation browser — fetches the swagger spec, groups endpoints
 *      by tag in a disclosure accordion, and per endpoint shows the parameter
 *      schema, an editable request-example form, copy-ready usage snippets
 *      (curl / Python / JavaScript / HTTPie), and the response schema/example.
 *   2. API access-token management — "Access Token" generates (and implicitly
 *      rotates/revokes) the account's long-lived token. The secret is shown
 *      ONCE in a transient dialog and dropped on dismissal; it is never stored
 *      in Pinia (page AC), so it cannot leak into the devtools state tree.
 *
 * CSP-clean: no v-html (token + snippets render via text interpolation), no
 * eval/new Function. `:style` is used only for dynamic depth indentation,
 * matching the A4 <DataTable> precedent. All user-facing strings route
 * through t().
 */
import { storeToRefs } from 'pinia'
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import SelectButton from 'primevue/selectbutton'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import http from '@/api/http'
import { useAuthStore } from '@/App'
import AppMessage from '@/components/elements/AppMessage.vue'
import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import FieldHint from '@/components/elements/FieldHint.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import { useToast } from '@/composables/useToast'
import { type Endpoint, renderUsage } from '@/lib/swagger'
import { useSwaggerStore } from '@/stores/swagger'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    api: {
      title: 'API Documentation',
      tooltip:
        'Browse every available endpoint, fill example values into the parameters and copy the ready-to-use snippet for curl, Python, JavaScript or HTTPie.',
      loading: 'Loading API documentation...',
      empty: 'No API endpoints are documented.',
      error: 'Could not load API documentation.',
      section: {
        requestParameters: 'Request Parameters',
        requestExample: 'Request Example',
        responseCodes: 'Response Codes',
        responseParameters: 'Response Parameters',
        responseExample: 'Response Example',
      },
      table: {
        in: 'In',
        required: 'Required',
        status: 'Status',
        field: 'Field',
      },
      hint: {
        commaList: 'Comma-separated list - will be sent as a JSON array.',
        jsonArray: 'JSON array of objects - edit the template above.',
        datetime: 'Datetime in RFC3339 format.',
      },
      usage: {
        selectLanguage: 'Select snippet language',
        actions: {
          copy: {
            success: 'Snippet copied to the clipboard.',
          },
        },
      },
      actions: {
        generateApiToken: {
          button: 'API Token',
          title: 'Generate API Access Token',
          message: {
            line1: 'A new long-lived API access token will be generated, valid for 1 year.',
            line2: 'Any previously issued API token for your account will be invalidated.',
          },
          question: 'Are you sure you want to generate a new API access token',
          error: 'Could not generate the API access token.',
        },
      },
    },
    dialog: {
      generateApiToken: {
        title: 'Generated API Access Token',
        hint: 'Please copy the API token; it will disappear shortly.',
        actions: {
          copy: {
            button: 'Copy token',
            success: 'Copied to clipboard.',
          },
        },
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()

const swaggerStore = useSwaggerStore()
const { groups, isFetching, isLoaded, fetchError } = storeToRefs(swaggerStore)
const { fetchDocs } = swaggerStore

const authStore = useAuthStore()
const { isDemo } = storeToRefs(authStore)

// ── Reactive state ────────────────────────────────────────────
// Origin for absolute URLs in the snippets (parity with legacy window.location.origin)
const origin = window.location.origin

const usageLanguages = [
  { label: 'curl', value: 'curl' },
  { label: 'python', value: 'python' },
  { label: 'javascript', value: 'javascript' },
  { label: 'httpie', value: 'httpie' },
]

const booleanOptions = [
  { label: 'false', value: 'false' },
  { label: 'true', value: 'true' },
]

const tokenConfirmVisible = ref(false)
const generatedAccessTokenVisible = ref(false)
/** Transient one-time secret. NEVER assigned into a store — cleared on dismiss. */
const generatedAccessToken = ref('')
const generatedAccessTokenLifetimeMs = 30_000
let generatedAccessTokenTimer: ReturnType<typeof setTimeout> | undefined

// Groups use a <Accordion multiple>: several tags may be open at once. The open
// panels are tracked by tag in `openGroups`. Legacy parity — groups start
// expanded (`collapsed: false`) — so once the docs load we seed the array with
// every tag (see the watch below).
const openGroups = ref<string[]>([])

// Endpoints keep the legacy "only one open across the entire page" rule. Each
// group's endpoint <Accordion> is single-select and binds the SAME source of
// truth `openEndpoint` (a page-wide endpoint key). Opening an endpoint in one
// group rewrites the key, so every other group's Accordion — seeing a value
// that matches none of its own panels — collapses. Endpoints start collapsed.
const openEndpoint = ref<string | null>(null)

// ── Functions and handlers ────────────────────────────────────
function openGenerateApiTokenConfirm(): void {
  tokenConfirmVisible.value = true
}

async function generateApiToken(): Promise<void> {
  closeGeneratedAccessTokenDialog()
  try {
    const res = await http.post<{
      message?: string
      body?: { token: string | null; expires_at: string }
    }>('/user/api-token', undefined, {
      errorToastMessage: t('api.actions.generateApiToken.error'),
    })
    const token = res.data.body?.token ?? null
    if (token) {
      generatedAccessToken.value = token
      generatedAccessTokenVisible.value = true
      generatedAccessTokenTimer = setTimeout(
        closeGeneratedAccessTokenDialog,
        generatedAccessTokenLifetimeMs,
      )
    } else {
      // No secret returned, for example in demo mode, so surface the backend message
      toast.success(res.data.message ?? '')
    }
  } catch {
    // The shared HTTP client already displayed the contextual generation error
  }
}

/** Drop the secret the moment the dialog closes (show-once guarantee). */
function clearGeneratedAccessToken(): void {
  if (generatedAccessTokenTimer !== undefined) {
    clearTimeout(generatedAccessTokenTimer)
    generatedAccessTokenTimer = undefined
  }
  generatedAccessToken.value = ''
}

function closeGeneratedAccessTokenDialog(): void {
  generatedAccessTokenVisible.value = false
  clearGeneratedAccessToken()
}

function onGeneratedAccessTokenDialogVisibilityChange(visible: boolean): void {
  if (!visible) closeGeneratedAccessTokenDialog()
}

function methodSeverity(method: string): 'warn' | 'danger' | undefined {
  if (method === 'PUT') return 'warn'
  if (method === 'DELETE' || method === 'PATCH') return 'danger'
  return undefined // GET and POST use the primary teal color
}

function statusSeverity(code: string): 'warn' | 'danger' | undefined {
  if (code.charAt(0) === '4') return 'warn'
  if (code.charAt(0) === '5') return 'danger'
  return undefined // Successful responses use the primary teal color
}

/** Indent nested schema rows / fields by depth (rem so it stays CSP-object-style). */
function indentByDepth(depth: number, unit: number): Record<string, string> {
  return { paddingLeft: `${depth * unit}rem` }
}

function updateFileInput(event: Event, endpoint: Endpoint, fieldName: string): void {
  const input = event.currentTarget
  if (!(input instanceof HTMLInputElement)) return
  const file = input.files?.[0]
  if (file) {
    endpoint.fileInputs[fieldName] = file
  } else {
    delete endpoint.fileInputs[fieldName]
  }
}

function snippetFor(endpoint: Endpoint): string {
  return renderUsage(endpoint, origin)
}

/** Stable page-wide key for an endpoint (method + path is unique across groups). */
function endpointKey(endpoint: Endpoint): string {
  return `${endpoint.method} ${endpoint.path}`
}

// ── Watchers and lifecycle ────────────────────────────────────
// Seed all groups open the first time the docs finish loading (groups default to
// expanded), matching the previous `collapsed: false` initial state.
watch(groups, (list) => {
  openGroups.value = list.map((group) => group.tag)
})

onMounted(async () => {
  try {
    await fetchDocs()
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }
})

onBeforeUnmount(clearGeneratedAccessToken)
</script>

<template>
  <section class="page-section api-page">
    <SectionHeader test-id="api-help" :title="t('api.title')" :help-text="t('api.tooltip')">
      <template v-if="!isDemo" #suffix>
        <SegmentedButton
          data-testid="api-generate-token"
          type="button"
          severity="warn"
          icon="pi pi-lock"
          :label="t('api.actions.generateApiToken.button')"
          @click="openGenerateApiTokenConfirm"
        />
      </template>
    </SectionHeader>

    <AppMessage
      v-if="fetchError"
      severity="error"
      data-testid="api-error"
      :message="t('api.error')"
    />

    <LayoutLoading
      :loading="isFetching && !isLoaded"
      :loading-message="t('api.loading')"
      :empty="isLoaded && groups.length === 0"
      :empty-message="t('api.empty')"
    >
      <!-- Group level: independent multi-open accordion (groups default open). -->
      <Accordion v-model:value="openGroups" multiple class="api-page__groups">
        <AccordionPanel
          v-for="group in groups"
          :key="group.tag"
          :value="group.tag"
          class="api-page__group"
          :data-testid="`api-group-${group.tag.toLowerCase()}`"
        >
          <AccordionHeader class="api-page__group-header">
            <span class="api-page__group-label">
              <span class="api-page__group-tag">{{ group.tag }}</span>
              <Tag :value="String(group.endpoints.length)" rounded class="api-page__group-count" />
            </span>
          </AccordionHeader>

          <AccordionContent>
            <!-- Endpoint level: single-open accordion, but bound to the page-wide
                 `openEndpoint` key so at most one endpoint is open across ALL
                 groups (opening one collapses any endpoint in other groups). -->
            <Accordion v-model:value="openEndpoint" class="api-page__endpoints">
              <AccordionPanel
                v-for="endpoint in group.endpoints"
                :key="endpointKey(endpoint)"
                :value="endpointKey(endpoint)"
                class="api-page__endpoint"
              >
                <AccordionHeader class="api-page__endpoint-header">
                  <Tag
                    :value="endpoint.method"
                    :severity="methodSeverity(endpoint.method)"
                    class="api-page__method"
                  />
                  <span class="api-page__path">{{ endpoint.fullPath }}</span>
                  <span class="api-page__summary">{{ endpoint.summary }}</span>
                </AccordionHeader>

                <AccordionContent>
                  <div class="api-page__detail" data-testid="api-endpoint-detail">
                    <template v-if="endpoint.description">
                      <SectionHeader level="h4" :title="t('common.labels.description')" />
                      <p class="api-page__text">{{ endpoint.description }}</p>
                    </template>

                    <template v-if="endpoint.inputFields.length > 0">
                      <SectionHeader level="h4" :title="t('api.section.requestParameters')" />
                      <table class="api-page__table">
                        <thead>
                          <tr>
                            <th>{{ t('common.labels.name') }}</th>
                            <th>{{ t('api.table.in') }}</th>
                            <th>{{ t('common.labels.type') }}</th>
                            <th class="api-page__th-center">
                              {{ t('api.table.required') }}
                            </th>
                            <th>{{ t('common.labels.description') }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="field in endpoint.inputFields" :key="`p-${field.path}`">
                            <td class="api-page__nowrap">
                              <span :style="indentByDepth(field.depth, 1.2)">
                                <i
                                  v-if="field.depth > 0"
                                  class="pi pi-angle-right api-page__angle"
                                  aria-hidden="true"
                                />
                                {{ field.label }}
                              </span>
                            </td>
                            <td>
                              <Tag :value="field.location" severity="secondary" />
                            </td>
                            <td class="api-page__type">
                              {{ field.typeLabel }}
                            </td>
                            <td class="api-page__th-center">
                              <i
                                v-if="field.required"
                                class="pi pi-check api-page__check"
                                :aria-label="t('api.table.required')"
                              />
                              <span v-else class="api-page__dash">-</span>
                            </td>
                            <td class="api-page__muted">
                              {{ field.description }}
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </template>

                    <template v-if="endpoint.inputFields.length > 0">
                      <SectionHeader level="h4" :title="t('api.section.requestExample')" />
                      <form class="api-page__form" @submit.prevent>
                        <template v-for="field in endpoint.inputFields" :key="`f-${field.path}`">
                          <!-- Object group header (nested struct) -->
                          <div
                            v-if="field.isObject"
                            class="api-page__field"
                            :style="indentByDepth(field.depth, 1.5)"
                          >
                            <div class="api-page__object-label">
                              <b>{{ field.label }}</b>
                              <span class="api-page__muted-sm"> (object) </span>
                            </div>
                            <FieldHint v-if="field.description">{{ field.description }}</FieldHint>
                          </div>

                          <!-- Editable value -->
                          <div
                            v-else
                            class="api-page__field"
                            :style="indentByDepth(field.depth, 1.5)"
                          >
                            <label class="api-page__label" :for="`field-${field.path}`">
                              <span>{{ field.label }}</span>
                              <span
                                v-if="field.required"
                                v-tooltip.top="t('api.table.required')"
                                class="api-page__required"
                                >&nbsp;*</span
                              >
                              <span class="api-page__muted-sm">&nbsp;({{ field.typeLabel }})</span>
                            </label>

                            <input
                              v-if="field.paramType === 'file'"
                              :id="`field-${field.path}`"
                              type="file"
                              class="api-page__file"
                              :aria-label="field.label"
                              @change="updateFileInput($event, endpoint, field.name)"
                            />

                            <Select
                              v-else-if="field.paramType === 'boolean'"
                              :id="`field-${field.path}`"
                              v-model="field.value"
                              :options="booleanOptions"
                              option-label="label"
                              option-value="value"
                              class="api-page__input"
                              :aria-label="field.label"
                            />

                            <Textarea
                              v-else-if="field.isArrayOfObjects"
                              :id="`field-${field.path}`"
                              v-model="field.value"
                              :rows="10"
                              :placeholder="field.placeholder"
                              class="api-page__input api-page__mono"
                              :aria-label="field.label"
                            />

                            <InputText
                              v-else
                              :id="`field-${field.path}`"
                              v-model="field.value"
                              :data-testid="`api-field-${field.name}`"
                              :type="field.paramType === 'integer' ? 'number' : 'text'"
                              :placeholder="field.placeholder"
                              class="api-page__input"
                              :aria-label="field.label"
                            />

                            <FieldHint
                              v-if="field.isArrayOfObjects"
                              :text="t('api.hint.jsonArray')"
                              severity="info"
                            />
                            <FieldHint
                              v-else-if="field.paramType === 'array'"
                              :text="t('api.hint.commaList')"
                              severity="info"
                            />
                            <FieldHint
                              v-else-if="field.typeLabel === 'string (date-time)'"
                              :text="t('api.hint.datetime')"
                              severity="info"
                            />
                          </div>
                        </template>
                      </form>
                    </template>

                    <!-- Usage snippets: language selector, then the copyable code block
                   (CodeSnippet supplies its own corner copy button). -->
                    <div class="api-page__usage-bar">
                      <SelectButton
                        v-model="endpoint.usageTab"
                        data-testid="api-usage-lang"
                        :options="usageLanguages"
                        option-label="label"
                        option-value="value"
                        :allow-empty="false"
                        :aria-label="t('api.usage.selectLanguage')"
                      />
                    </div>
                    <CodeSnippet
                      test-id="api-usage-snippet"
                      class="api-page__usage-snippet"
                      :text="snippetFor(endpoint)"
                      :text-copy-label="t('common.actions.copy')"
                      :text-copied-message="t('api.usage.actions.copy.success')"
                      multiline
                    />

                    <template v-if="endpoint.responses.length > 0">
                      <SectionHeader level="h4" :title="t('api.section.responseCodes')" />
                      <table class="api-page__table">
                        <thead>
                          <tr>
                            <th class="api-page__status-col">
                              {{ t('api.table.status') }}
                            </th>
                            <th>{{ t('common.labels.description') }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="response in endpoint.responses" :key="`r-${response.code}`">
                            <td class="api-page__th-center">
                              <Tag
                                :value="response.code"
                                :severity="statusSeverity(response.code)"
                              />
                            </td>
                            <td class="api-page__muted">
                              {{ response.description }}
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </template>

                    <template v-for="response in endpoint.responses" :key="`rb-${response.code}`">
                      <template v-if="response.code === '200' && response.fields.length > 0">
                        <SectionHeader level="h4" :title="t('api.section.responseParameters')" />
                        <table class="api-page__table">
                          <thead>
                            <tr>
                              <th>{{ t('api.table.field') }}</th>
                              <th>{{ t('common.labels.type') }}</th>
                              <th>{{ t('common.labels.description') }}</th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr
                              v-for="responseField in response.fields"
                              :key="`response-field-${responseField.name}-${responseField.depth}`"
                            >
                              <td class="api-page__nowrap">
                                <span :style="indentByDepth(responseField.depth, 1.2)">
                                  <i
                                    v-if="responseField.depth > 0"
                                    class="pi pi-angle-right api-page__angle"
                                    aria-hidden="true"
                                  />
                                  {{ responseField.name }}
                                </span>
                              </td>
                              <td class="api-page__type">
                                {{ responseField.typeLabel }}
                              </td>
                              <td class="api-page__muted">
                                {{ responseField.description }}
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </template>
                      <template v-if="response.code === '200' && response.example">
                        <SectionHeader level="h4" :title="t('api.section.responseExample')" />
                        <CodeSnippet
                          test-id="api-response-example"
                          multiline
                          :text="response.example"
                          :text-copy-label="t('common.actions.copy')"
                          :text-copied-message="t('api.usage.actions.copy.success')"
                        />
                      </template>
                    </template>
                  </div>
                </AccordionContent>
              </AccordionPanel>
            </Accordion>
          </AccordionContent>
        </AccordionPanel>
      </Accordion>
    </LayoutLoading>

    <!-- Confirm generation (rotates/revokes the previous token) -->
    <DialogConfirm
      v-model:visible="tokenConfirmVisible"
      :title="t('api.actions.generateApiToken.title')"
      icon="pi pi-lock"
      severity="warn"
      :text-confirm-label="t('common.actions.generate')"
      @confirm="generateApiToken"
    >
      <p>{{ t('api.actions.generateApiToken.question') }}?</p>
      <p class="api-page__text api-page__confirm-message">
        {{ t('api.actions.generateApiToken.message.line1') }}<br />
        {{ t('api.actions.generateApiToken.message.line2') }}
      </p>
    </DialogConfirm>

    <Dialog
      data-testid="api-token-dialog"
      :visible="generatedAccessTokenVisible"
      modal
      dismissable-mask
      :draggable="false"
      class="app-dialog api-page__token-dialog"
      @update:visible="onGeneratedAccessTokenDialogVisibilityChange"
    >
      <template #header>
        <div class="app-dialog__header">
          <i class="pi pi-lock app-dialog__header-icon" aria-hidden="true" />
          <span class="app-dialog__header-title">{{ t('dialog.generateApiToken.title') }}</span>
        </div>
      </template>
      <div class="api-page__dialog-body">
        <p class="api-page__text">{{ t('dialog.generateApiToken.hint') }}</p>
        <CodeSnippet
          v-if="generatedAccessToken"
          test-id="api-token-secret"
          :text="generatedAccessToken"
          centered
          :text-copy-label="t('dialog.generateApiToken.actions.copy.button')"
          :text-copied-message="t('dialog.generateApiToken.actions.copy.success')"
        />
      </div>
      <template #footer>
        <Button
          data-testid="api-token-close"
          type="button"
          severity="secondary"
          :label="t('common.actions.close')"
          @click="closeGeneratedAccessTokenDialog"
        />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
/* Match the shared confirmation-dialog width without relying on template styles */
.api-page__token-dialog {
  width: 38rem;
  max-width: calc(100vw - 2rem);
}

/* ── Endpoint groups (PrimeVue Accordion, grouped by tag) ──────────────────── */
/* Each group is its own bordered card; drop the default panel bottom border so
   panels don't double up, then frame each one individually. */
.api-page__groups :deep(.p-accordionpanel) {
  margin-bottom: 0.75rem;
  overflow: hidden;
  border: none;
}

.api-page__dialog-body {
  text-align: center;
}

/* Group header bar: surface grey.
   !important is required — PrimeVue injects theme CSS via [data-p-active] attribute
   selectors at runtime, which have higher specificity than our :deep() class rules. */
.api-page__group > :deep(.p-accordionheader) {
  gap: 0.5rem;
  padding: 0.75rem 1rem !important;
  margin-bottom: 0.5rem;
  background: var(--p-surface-0) !important;
  color: var(--p-text-color) !important;
}

.api-page__group > :deep(.p-accordionheader:hover),
.api-page__group > :deep(.p-accordionheader[aria-expanded='true']),
.api-page__group > :deep(.p-accordionheader[aria-expanded='true']:hover) {
  background: var(--p-surface-50) !important;
  color: var(--p-text-color) !important;
}

/* Wrapper keeps tag text + count badge as one flex unit on the left */
.api-page__group-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

/* Tag name in the group header: bold, sentence-cased */
.api-page__group-tag {
  font-weight: 600;
  text-transform: capitalize;
}

/* Group body padding around the nested endpoint accordion. */
.api-page__group :deep(.p-accordioncontent-content) {
  padding: 0;
}

/* ── Endpoints (nested Accordion, one collapsible card per route) ──────────── */
.api-page__endpoints {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.api-page__endpoints :deep(.p-accordionpanel) {
  border: 1px solid var(--p-content-border-color);
  border-radius: var(--p-content-border-radius);
  margin-bottom: 0;
  /* Subtle shadow lift for the segment card. */
  box-shadow: 0 1px 2px 0 var(--p-content-border-color);
}

/* Endpoint header row: method Tag + path + summary + the built-in chevron. */
.api-page__endpoint > :deep(.p-accordionheader) {
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: transparent;
}

/* The endpoint body already carries its own padding via .api-page__detail. */
.api-page__endpoint :deep(.p-accordioncontent-content) {
  padding: 0;
}

/* HTTP method Tag (GET/POST): centered with a fixed min-width */
.api-page__method {
  justify-content: center;
  min-width: 4rem;
}

/* Endpoint path: grows to fill available width */
.api-page__path {
  flex: 1;
  /* Legacy rendered endpoint paths in the body sans-serif, not monospace. */
  font-family: var(--p-font-family);
  word-break: break-all;
}

/* One-line summary beside the path: secondary text color */
.api-page__summary {
  color: var(--p-text-muted-color);
}

/* Expanded endpoint body: padded content area below the header divider */
.api-page__detail {
  padding: 0.4rem 1rem 1rem;
  border-top: 1px solid var(--p-content-border-color);
}

/* Description paragraph inside the detail area */
.api-page__text {
  margin: 0 0 0.5rem;
}

/* ── Schema tables ─────────────────────────────────────────────────────────── */

/* Rounded-corner bordered table. border-collapse: separate + border-spacing: 0
   lets border-radius render; overflow: hidden clips the corners. */
.api-page__table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid var(--p-content-border-color);
  border-radius: var(--p-content-border-radius);
  overflow: hidden;
  margin-bottom: 1rem;
}

.api-page__table th,
.api-page__table td {
  border-bottom: 1px solid var(--p-content-border-color);
  border-right: 1px solid var(--p-content-border-color);
  padding: 0.35rem 0.5rem;
  text-align: left;
  vertical-align: top;
}

.api-page__table th:last-child,
.api-page__table td:last-child {
  border-right: none;
}

.api-page__table tr:last-child th,
.api-page__table tr:last-child td {
  border-bottom: none;
}

.api-page__table th {
  background: var(--p-surface-50);
  font-weight: 600;
}

.api-page__table th:hover {
  background: var(--p-surface-100);
}

/* Center-aligned header cell (e.g. Required, Status columns) */
.api-page__th-center {
  text-align: center;
}

/* Fixed-width column for HTTP status codes */
.api-page__status-col {
  width: 5rem;
}

/* Prevent a cell from wrapping (e.g. field names, type labels) */
.api-page__nowrap {
  white-space: nowrap;
}

/* Schema field type displayed in monospace, muted */
.api-page__type {
  font-family: var(--p-font-family-mono, monospace);
  color: var(--p-text-muted-color);
}

/* Secondary/muted text in table cells */
.api-page__muted {
  color: var(--p-text-muted-color);
}

/* Nesting indicator before a nested schema field. */
.api-page__angle {
  margin-right: 0.2rem;
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
}

/* Checkmark icon in the Required column: teal */
.api-page__check {
  color: var(--p-primary-500);
}

/* Dash icon in the Required column when field is optional */
.api-page__dash {
  color: var(--p-text-muted-color);
}

/* Small muted label (e.g. nested sub-schema headings) */
.api-page__muted-sm {
  color: var(--p-text-muted-color);
  font-weight: 400;
}

/* ── Editable request-example form ─────────────────────────────────────────── */
.api-page__form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.api-page__field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

/* Required-field asterisk: teal to match the checkmark in the schema table */
.api-page__required {
  color: var(--p-primary-500);
}

/* Text and select inputs span the full field width */
.api-page__input {
  width: 100%;
}

/* Monospace textarea and its wrapper (JSON body fields) */
.api-page__mono :deep(textarea),
.api-page__mono {
  font-family: var(--p-font-family-mono, monospace);
}

/* File input: inherit the surrounding font so it blends with other inputs */
.api-page__file {
  font: inherit;
}

/* ── Usage snippets ────────────────────────────────────────────────────────── */
.api-page__usage-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.5rem;
  /* Space above (from the form) and a gap below so the language tab bar doesn't
     sit flush against the code snippet. */
  margin-top: 0.5rem;
}

/* Separate the request snippet from the response details that follow */
.api-page__usage-snippet {
  margin-bottom: 1rem;
}

/* Flatten the SelectButton into a tab bar matching the side-nav style:
   a continuous 3px gray bottom track with a teal 3px active segment. */
.api-page__usage-bar :deep(.p-selectbutton) {
  display: flex;
  background: transparent;
  border: none;
  border-bottom: 3px solid var(--p-surface-300);
  border-radius: 0;
}

.api-page__usage-bar :deep(.p-togglebutton) {
  margin-bottom: -3px;
  padding: 0.45rem 0.9rem;
  border: none;
  border-bottom: 3px solid transparent;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  color: var(--p-text-color);
}

.api-page__usage-bar :deep(.p-togglebutton:first-child),
.api-page__usage-bar :deep(.p-togglebutton:last-child) {
  border-radius: 0;
}

/* Remove the v4 ToggleButton inner "card" highlight and content shadow. */
.api-page__usage-bar :deep(.p-togglebutton::before) {
  display: none;
}

.api-page__usage-bar :deep(.p-togglebutton-content) {
  box-shadow: none !important;
}

.api-page__usage-bar :deep(.p-togglebutton:hover),
.api-page__usage-bar :deep(.p-togglebutton:hover .p-togglebutton-label) {
  background: transparent !important;
  color: var(--p-text-color) !important;
}

.api-page__usage-bar :deep(.p-togglebutton.p-togglebutton-checked) {
  background: transparent;
  border-bottom-color: var(--p-primary-600);
  color: var(--p-text-color);
  font-weight: 600;
}

.api-page__usage-bar :deep(.p-togglebutton-checked .p-togglebutton-label) {
  font-weight: 600;
}
</style>
