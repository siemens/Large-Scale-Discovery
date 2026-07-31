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
 * Edits scan settings for an existing scope.
 * Covers all scan module settings, timespans, and sensitive ports.
 */
import DatePicker from 'primevue/datepicker'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppMessage from '@/components/elements/AppMessage.vue'
import Field from '@/components/elements/Field.vue'
import FieldNumber from '@/components/elements/FieldNumber.vue'
import FieldSlider from '@/components/elements/FieldSlider.vue'
import FieldTimespans from '@/components/elements/FieldTimespans.vue'
import FieldToggle from '@/components/elements/FieldToggle.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useFormLifecycle, useFormSnapshot } from '@/composables/useFormLifecycle'
import { validatePort } from '@/lib/validators'
import { useScopesStore } from '@/stores/scopes'

import type { EditableTimespan, ScanSettings, ScanTimespan, ScopeResponse } from '@/types/scopes'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    scopeSettingsPanel: {
      fields: {
        sensitivePorts: {
          label: 'Sensitive Ports',
          hint: 'Ports to exclude from scanning, which are known to cause issues.',
          placeholder: '515, 631, 9100, 9101, 9102, 9103',
        },
        name: {
          label: 'Scope Name',
          placeholder: 'Scope Name',
        },
        maxInstancesDiscovery: {
          label: 'Discovery',
          hint: 'The discovery module is required to feed other modules.',
        },
        maxInstancesBanner: {
          label: 'Banner',
        },
        maxInstancesNfs: {
          label: 'NFS',
        },
        maxInstancesNuclei: {
          label: 'Nuclei',
        },
        maxInstancesSmb: {
          label: 'SMB',
        },
        maxInstancesSsh: {
          label: 'SSH',
        },
        maxInstancesSsl: {
          label: 'SSL',
        },
        maxInstancesWebcrawler: {
          label: 'Webcrawler',
        },
        maxInstancesWebenum: {
          label: 'Webenum',
        },
        nmapArgs: {
          label: 'Nmap Args',
          hint: 'Nmap arguments to use for the main port scan, version detection and script execution.',
          otPlaceholder:
            '-sT -p 21,22,23,25,80,102,135,139,161,443,445,502,623,993,995,2404,3389,4840,4843,5060,8080,8443,20000,34962,34963,34964,44818,47808,4840,4843 -T2 --randomize-hosts --host-timeout 6h --max-retries 1 --max-rate 30 --traceroute --resolve-all --script s7-info,modbus-discover,bacnet-info,enip-info',
          placeholder:
            '-sS -sU -PE -PP -Pn -O -p U:53,67,68,111,161,162,1900,2049,T:0-65535 -sV -T4 --randomize-hosts --host-timeout 6h --max-retries 2 --traceroute --resolve-all --script=default',
        },
        nmapArgsPrescan: {
          label: 'Nmap Args Pre-scan',
          hint: 'Nmap arguments to use for a first pre-scan, intended to collect some smaller results before a potential IDS might trigger. If left empty, no pre-scan will be conducted.',
          placeholder:
            '-sS -Pn -p 21,22,23,80,135,139,443,445,3389,5900,8080,8443 -T4 --randomize-hosts --host-timeout 2m --max-retries 2',
        },
        excludeDomains: {
          label: 'Discovery Exclude Domains',
          hint: 'Domains to be ignored when discovered. This allows to ignore common hostnames known to be unrelated to the scanned asset. Hostnames with these domains will not be stored nor used in sub modules, e.g., as vhosts.',
          placeholder:
            'cloudfront.net, wildcard.cloudfront.net, azurewebsites.net, scm.azure-mobile.net, scm.azurewebsites.net, sso.azurewebsites.net, wildcard.azure-mobile.net, wildcard.azurewebsites.net, wildcard.scm.azure-mobile.net, wildcard.scm.azurewebsites.net, wildcard.sso.azurewebsites.net',
        },
        httpUserAgent: {
          label: 'HTTP User Agent',
          hint: 'HTTP user agent used by the web crawler and web enumerator when sending requests.',
          placeholder:
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0',
        },
        networkTimeout: {
          label: 'Network Timeout',
          hint: 'Time to wait for socket connections before aborting.',
          placeholder: '8',
          unit: 'Seconds',
        },
        smbForcedShares: {
          label: 'SMB Force Shares',
          hint: 'Comma-separated list of shares you would always like to check, because they might exist but be invisible.',
          placeholder: 'C, D, E, F, G, H, C$, D$, E$, F$, G$, H$, ADMIN$, TEMP, TMP',
        },
        smbExcludeShares: {
          label: 'SMB Exclude Shares',
          hint: 'Comma-separated list of shares you would like to ignore.',
          placeholder:
            'print$, W7DP$, LSDP, LSDP_mosaic$, LSDP_test$, LSDP.WW005, lsdp-backup, lsdp_drivers_ww300$, LSDPW7$, custom_root$, gplmshare, SCCMContentLib$, SCCMContentLibD$, WsusContent',
        },
        smbExcludeFolders: {
          label: 'SMB Exclude Folders',
          hint: 'Comma-separated list of folders you would like to ignore.',
          placeholder: 'Comma, separated, list, of, folder, names',
        },
        smbExcludeExtensions: {
          label: 'SMB Exclude Extensions',
          hint: 'Comma-separated list of file extensions you would like to ignore.',
          placeholder: 'Comma, separated, list, of, file, extensions',
        },
        smbExcludeFileSizeBelow: {
          label: 'SMB Exclude Size',
          hint: 'File size below which files are skipped during crawling.',
          placeholder: '1024',
          unit: 'Bytes',
        },
        smbExcludeLastModified: {
          label: 'SMB Exclude Old Files',
          hint: 'File modification date below which files are skipped during crawling. Files with empty modification date will be collected, nevertheless.',
        },
        smbDepth: {
          label: 'SMB Depth',
          hint: "Depth of the share's folder structure to go down. -1 = no limit.",
          placeholder: '-1',
          unit: 'Folders',
        },
        smbThreads: {
          label: 'SMB Threads',
          hint: 'Amount of parallel requests per SMB target.',
          placeholder: '4',
          unit: 'Threads',
        },
        nucleiIncludeSeverities: {
          label: 'Nuclei Include Severities',
          hint: 'Comma-separated list of template finding severities you would like to include.',
          placeholder: '*',
        },
        nucleiExcludeSeverities: {
          label: 'Nuclei Exclude Severities',
          hint: 'Comma-separated list of template finding severities you would like to ignore.',
          placeholder: '-',
        },
        nucleiIncludeTags: {
          label: 'Nuclei Include Tags',
          hint: 'Comma-separated list of template tags you would like to include: cve, lfi, rce, xss.',
          placeholder: '*',
        },
        nucleiExcludeTags: {
          label: 'Nuclei Exclude Tags',
          hint: 'Comma-separated list of template tags you would like to ignore: cve, lfi, rce, xss.',
          placeholder: '-',
        },
        nucleiIncludeIds: {
          label: 'Nuclei Include IDs',
          hint: 'Comma-separated list of template ids you would like to include.',
          placeholder: '*',
        },
        nucleiExcludeIds: {
          label: 'Nuclei Exclude IDs',
          hint: 'Comma-separated list of template ids you would like to ignore.',
          placeholder: '-',
        },
        nucleiIncludeProtocols: {
          label: 'Nuclei Include Protocols',
          hint: 'Comma-separated list of template protocols you would like to include: tcp, http, dns, ssl.',
          placeholder: '*',
        },
        nucleiExcludeProtocols: {
          label: 'Nuclei Exclude Protocols',
          hint: 'Comma-separated list of template protocols you would like to ignore: tcp, http, dns, ssl.',
          placeholder: '-',
        },
        webcrawlerFollowTypes: {
          label: 'Web Crawler Follow Types',
          hint: 'Response content types to follow.',
          placeholder:
            'text/html, text/plain, text/javascript, application/javascript, application/json, application/atom+xml, application/rss+xml, application/xhtml+xml, application/x-latex, application/xml, application/xml-dtd, application/x-sh, application/x-tex, application/x-texinfo, text/cache-manifest, text/calendar, text/css, text/csv, text/csv-schema, text/directory, text/dns, text/ecmascript, text/encaprtp, text/example, text/fwdred, text/grammar-ref-list, text/jcr-cnd, text/markdown, text/mizar, text/n3, text/parameters, text/provenance-notation, text/prs.fallenstein.rst, text/prs.lines.tag, text/raptorfec, text/RED, text/rfc822-headers, text/rtf, text/rtp-enc-aescm128, text/rtploopback, text/rtx, text/SGML, text/t140, text/tab-separated-values, text/troff, text/turtle, text/ulpfec, text/uri-list, text/vcard, text/vnd.abc, text/vnd.debian.copyright, text/vnd.DMClientScript, text/vnd.dvb.subtitle, text/vnd.esmertec.theme-descriptor, text/vnd.fly, text/vnd.fmi.flexstor, text/vnd.graphviz, text/vnd.in3d.3dml, text/vnd.in3d.spot, text/vnd.IPTC.NewsML, text/vnd.IPTC.NITF, text/vnd.latex-z, text/vnd.motorola.reflex, text/vnd.ms-mediapackage, text/vnd.net2phone.commcenter.command, text/vnd.radisys.msml-basic-layout, text/vnd.si.uricatalogue, text/vnd.sun.j2me.app-descriptor, text/vnd.trolltech.linguist, text/vnd.wap.si, text/vnd.wap.sl, text/vnd.wap.wmlscript, text/vnd.wap-wml, text/vnd-a, text/vnd-curl, text/xml, text/xml-external-parsed-entity',
        },
        webcrawlerDepth: {
          label: 'Web Crawler Depth',
          hint: "Depth of the site's navigation tree to go down. -1 = no limit.",
          placeholder: '-1',
          unit: 'Links',
        },
        webcrawlerMaxThreads: {
          label: 'Web Crawler Threads',
          hint: 'Amount of parallel requests per HTTP target.',
          placeholder: '4',
          unit: 'Threads',
        },
        webcrawlerFollowQs: {
          label: 'Web Crawler QS',
          hint: 'Query strings are parameters in URLs that might lead to different page contents, but they might also cause the crawler to stall. Enable to include query strings in requests.',
        },
        webcrawlerStoreRoot: {
          label: 'Web Crawler Save Root',
          hint: 'Usually the web crawler only saves success responses. In case of the index page, other responses might be interesting too. Enable to save the index page independent of response code.',
        },
        webenumProbeRobots: {
          label: 'Web Enum robots.txt',
          hint: 'Sometimes robots.txt leaks information about hidden resources. Enable to add found entries to the enumeration URLs.',
        },
      },
      sections: {
        instances: {
          hint: "A scan agent executes multiple instances of each scan module in parallel. These are the default instance limits for agents, but they can be overridden in each agent's configuration.",
          title: 'Agent Processes',
        },
        scanSettings: {
          hint: 'Scan modules can be tailored to better suit your needs.',
          title: 'Scan Settings',
        },
      },
      agentLimitWarning:
        'At least one scan agent assigned to this scan scope has dedicated scan limits configured. Above scan scope limits only apply to scan agents without dedicated limits.',
      ot: {
        warning:
          'OT discovery scan scope intended to discover devices mainly on OSI layer 2! Customize scan arguments according to the resilience of the targeted networks.',
      },
    },
    formTimespans: {
      hint: "Day and time ranges in which new scans will be started. The ranges are relative to each target's local time. Ongoing scans may continue outside of these time ranges!",
      title: 'Scan Timespans',
      dayOptions: {
        mon: 'Monday',
        tue: 'Tuesday',
        wed: 'Wednesday',
        thu: 'Thursday',
        fri: 'Friday',
        sat: 'Saturday',
        sun: 'Sunday',
      },
    },
  },
}

/** Child state exposed by FieldTimespans for dirty tracking and submission. */
interface FieldTimespansHandle {
  timespans: EditableTimespan[]
  validTimespans: EditableTimespan[]
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{
  scope: ScopeResponse
}>()

const emit = defineEmits<{
  close: []
  saved: [message: string]
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const scopesStore = useScopesStore()

// This dense settings form has ~30 fields; presenting every hint as an inline
// pointing label makes it very tall, so each field sets `tooltip` to show its
// hint as a `?` icon beside the label. Validation errors still render inline below.

// ── Reactive state ────────────────────────────────────────────

// Live reactive reference to the scope's scan settings. Reset re-clones from
// props.scope on open and on scope change so a new scope re-seeds the form.
const form = reactive<ScanSettings>({ ...props.scope.scan_settings })

// Sensitive ports: stored as number[] in API, edited as comma-separated string.
const sensitivePortsStr = ref('')

// SMB last-modified-below: stored as ISO datetime string, edited as Date.
const smbLastModified = ref<Date | null>(null)

// Seed prop for <FieldTimespans>; reassigned in reset() to re-initialise the child.
const initialTimespans = ref(props.scope.scan_settings.discovery_timespans)

// Template ref to read back the child's timespans/validTimespans for dirty-tracking and submit.
const timespansRef = ref<FieldTimespansHandle | null>(null)

const sensitivePortsError = ref('')

const snapshot = useFormSnapshot(formSnapshot, '')

// ── Computed ──────────────────────────────────────────────────
const hasAgentLimits = computed(() => (props.scope.scan_agents ?? []).some((a) => a.limits))

// OT scopes carry a special risk profile; surface a warning banner and disable
// the (irrelevant) Nmap pre-scan field.
const isOt = computed(() => !!form.ot)

// Legacy greyed out a module's dependent fields when its agent-process instance
// limit was set to 0. Each computed mirrors one `maxInstancesX() > 0 ? '' :
// 'disabled'` binding from settings.html.
const isDiscoveryDisabled = computed(() => form.max_instances_discovery === 0)
const isSmbDisabled = computed(() => form.max_instances_smb === 0)
const isNucleiDisabled = computed(() => form.max_instances_nuclei === 0)
const isWebcrawlerDisabled = computed(() => form.max_instances_webcrawler === 0)
const isWebenumDisabled = computed(() => form.max_instances_webenum === 0)

// HTTP user agent serves both web modules — disabled only when both are off.
const isHttpDisabled = computed(
  () => form.max_instances_webcrawler === 0 && form.max_instances_webenum === 0,
)

// Network timeout applies to every module — disabled only when all are off.
const isNetworkTimeoutDisabled = computed(
  () =>
    form.max_instances_discovery === 0 &&
    form.max_instances_banner === 0 &&
    form.max_instances_nfs === 0 &&
    form.max_instances_smb === 0 &&
    form.max_instances_ssh === 0 &&
    form.max_instances_ssl === 0 &&
    form.max_instances_webcrawler === 0 &&
    form.max_instances_webenum === 0,
)

// ── Functions ─────────────────────────────────────────────────
// Baseline snapshot for dirty tracking. Captured at the end of reset() (after the
// dummy timespan is added) so only genuine user edits to this multi-section form flip it to dirty.
function formSnapshot(): string {
  return JSON.stringify({
    form,
    sensitivePortsStr: sensitivePortsStr.value,
    smbLastModified: smbLastModified.value,
    timespans: timespansRef.value?.timespans,
  })
}

// Deep-clone the scope's scan settings into the form; runs on mount and on scope
// change. Reassigning initialTimespans triggers the child's synchronous watcher
// so the timespan list is up to date before the baseline snapshot is captured.
function resetForm(): void {
  const scanSettings = props.scope.scan_settings

  sensitivePortsError.value = ''
  Object.assign(form, scanSettings)

  sensitivePortsStr.value = Array.isArray(scanSettings.sensitive_ports)
    ? scanSettings.sensitive_ports.join(',')
    : ''

  smbLastModified.value = scanSettings.smb_exclude_last_modified_below
    ? new Date(scanSettings.smb_exclude_last_modified_below)
    : null

  // Reassign triggers the child's flush:'sync' watcher, re-seeding its list before snapshot().
  initialTimespans.value = scanSettings.discovery_timespans

  // Capture the pristine baseline after init so subsequent edits register as dirty
  snapshot.capture()
}

/** Parse the sensitive-port input without changing when inline errors are shown. */
function evaluateSensitivePorts(): { ports: number[]; error: string } {
  const raw = sensitivePortsStr.value.trim()
  if (!raw) return { ports: [], error: '' }
  const parts = raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  const ports: number[] = []
  for (const part of parts) {
    const err = validatePort(part)
    if (err) return { ports: [], error: err }
    ports.push(parseInt(part, 10))
  }
  return { ports, error: '' }
}

// Validate the comma-separated port string and reveal its inline error.
function parseSensitivePorts(): number[] | null {
  const result = evaluateSensitivePorts()
  sensitivePortsError.value = result.error
  return result.error ? null : result.ports
}

// After this field has exposed an error, keep it live while the user corrects
// the value so the hint disappears and submit re-enables without an extra blur.
watch(sensitivePortsStr, () => {
  if (sensitivePortsError.value) parseSensitivePorts()
})

async function submitForm(): Promise<boolean> {
  const sensitivePorts = parseSensitivePorts()
  if (sensitivePorts === null) return false

  // Sort timespans by start day+time (Sun = 7 at end)
  const daySortIndex = (day: number) => (day === 0 ? 7 : day - 1)
  const sortedTimespans = [...(timespansRef.value?.validTimespans ?? [])].sort((left, right) => {
    const leftDay = daySortIndex(left.startDay)
    const rightDay = daySortIndex(right.startDay)
    if (leftDay !== rightDay) return leftDay - rightDay
    return left.startTime.localeCompare(right.startTime)
  })

  // Backend expects startDay/endDay as strings; strip the client-side id field.
  const apiTimespans: ScanTimespan[] = sortedTimespans.map(
    ({ startDay, startTime, endDay, endTime }) => ({
      startDay: String(startDay),
      startTime,
      endDay: String(endDay),
      endTime,
    }),
  )

  const payload: ScanSettings = {
    ...form,
    sensitive_ports: sensitivePorts,
    discovery_timespans: apiTimespans.length > 0 ? apiTimespans : null,
    smb_exclude_last_modified_below: smbLastModified.value
      ? smbLastModified.value.toISOString()
      : form.smb_exclude_last_modified_below,
  }

  const message = await scopesStore.submitSettings({ id: props.scope.id, scan_settings: payload })
  emit('saved', message)
  return true
}

// Instance field helpers — avoids TypeScript casts in templates.
type InstanceKey =
  | 'max_instances_discovery'
  | 'max_instances_banner'
  | 'max_instances_nfs'
  | 'max_instances_nuclei'
  | 'max_instances_smb'
  | 'max_instances_ssh'
  | 'max_instances_ssl'
  | 'max_instances_webcrawler'
  | 'max_instances_webenum'

function getInstanceLimit(key: InstanceKey): number {
  return form[key]
}
function setInstanceLimit(key: InstanceKey, value: number): void {
  form[key] = value
}

// ── Watchers and lifecycle ────────────────────────────────────
// Reset-on-open, dirty registration, and the submit/cancel handlers the template
// binds to. Submit closes the form on a truthy result; a store error is caught
// by useFormLifecycle and the form stays open (the store surfaces the message).
const { submit: submitWithLifecycle, cancel } = useFormLifecycle({
  target: () => props.scope,
  reset: resetForm,
  dirty: snapshot.isDirty,
  submit: submitForm,
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitWithLifecycle })
</script>

<template>
  <!-- Top row (legacy parity): read-only Group + Scope Name identify which
       scope is being edited, alongside the editable Sensitive Ports. -->
  <div class="form-scope-settings__grid form-scope-settings__grid--identity">
    <Field v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText :id="id" :model-value="scope.group_name" disabled />
    </Field>
    <Field v-slot="{ id }" :label="t('scopeSettingsPanel.fields.name.label')" tooltip>
      <InputText
        :id="id"
        :model-value="scope.name"
        :placeholder="t('scopeSettingsPanel.fields.name.placeholder')"
        disabled
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.sensitivePorts.label')"
      :help="t('scopeSettingsPanel.fields.sensitivePorts.hint')"
      tooltip
      :error="sensitivePortsError || undefined"
      :validate="() => parseSensitivePorts() !== null"
    >
      <InputText
        :id="id"
        v-model="sensitivePortsStr"
        :invalid="!!sensitivePortsError"
        :placeholder="t('scopeSettingsPanel.fields.sensitivePorts.placeholder')"
        data-testid="ss-sensitive-ports"
      />
    </Field>
  </div>

  <!-- Discovery timespans are a settings-form section; the child owns only
       the editable grid and its field-level validation. -->
  <SectionHeader level="h4" variant="horizontal" :help-text="t('formTimespans.hint')">
    {{ t('formTimespans.title') }}
  </SectionHeader>
  <FieldTimespans
    ref="timespansRef"
    class="form-scope-settings__timespans"
    :initial-timespans="initialTimespans"
    :text-day-options="[
      { label: t('formTimespans.dayOptions.mon'), value: 1 },
      { label: t('formTimespans.dayOptions.tue'), value: 2 },
      { label: t('formTimespans.dayOptions.wed'), value: 3 },
      { label: t('formTimespans.dayOptions.thu'), value: 4 },
      { label: t('formTimespans.dayOptions.fri'), value: 5 },
      { label: t('formTimespans.dayOptions.sat'), value: 6 },
      { label: t('formTimespans.dayOptions.sun'), value: 0 },
    ]"
  />

  <!-- Agent Processes is a section grouping, not an input, so its help stays
         a heading tooltip rather than an inline field hint. -->
  <SectionHeader
    level="h4"
    variant="horizontal"
    :help-text="t('scopeSettingsPanel.sections.instances.hint')"
  >
    {{ t('scopeSettingsPanel.sections.instances.title') }}
  </SectionHeader>

  <!-- Dedicated-agent-limits notice: sits at the end of the agent-process
           section (legacy parity), where the per-module instance limits set
           above may be overridden by an agent's own dedicated limits. -->
  <AppMessage v-if="hasAgentLimits" severity="warn">
    <p>{{ t('scopeSettingsPanel.agentLimitWarning') }}</p>
  </AppMessage>

  <div class="form-scope-settings__grid form-scope-settings__grid--sliders">
    <div
      v-for="instanceControl in [
        {
          key: 'max_instances_discovery' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesDiscovery.label'),
        },
        {
          key: 'max_instances_banner' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesBanner.label'),
        },
        {
          key: 'max_instances_nfs' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesNfs.label'),
        },
        {
          key: 'max_instances_nuclei' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesNuclei.label'),
        },
        {
          key: 'max_instances_smb' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesSmb.label'),
        },
        {
          key: 'max_instances_ssh' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesSsh.label'),
        },
        {
          key: 'max_instances_ssl' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesSsl.label'),
        },
        {
          key: 'max_instances_webcrawler' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesWebcrawler.label'),
        },
        {
          key: 'max_instances_webenum' as const,
          label: t('scopeSettingsPanel.fields.maxInstancesWebenum.label'),
        },
      ]"
      :key="instanceControl.key"
    >
      <!-- Agent-process count: teal drag slider with a live value readout.
             Only the discovery module carries a help hint (legacy parity). -->
      <FieldSlider
        :model-value="getInstanceLimit(instanceControl.key)"
        :label="instanceControl.label"
        :min="0"
        :max="150"
        :step="1"
        :help="
          instanceControl.key === 'max_instances_discovery'
            ? t('scopeSettingsPanel.fields.maxInstancesDiscovery.hint')
            : undefined
        "
        tooltip
        show-value
        inline-label
        :data-testid="`ss-${instanceControl.key}`"
        @update:model-value="(value: number) => setInstanceLimit(instanceControl.key, value)"
      />
    </div>
  </div>

  <!-- Scan Settings — all module scan-argument tuning under one divider
       (legacy parity: settings.html grouped every module setting here). -->
  <SectionHeader
    level="h4"
    variant="horizontal"
    :help-text="t('scopeSettingsPanel.sections.scanSettings.hint')"
  >
    {{ t('scopeSettingsPanel.sections.scanSettings.title') }}
  </SectionHeader>

  <!-- OT-scope risk warning: OT scopes tune the Nmap arguments below for the
         softer timing OSI-layer-2 discovery needs. -->
  <AppMessage v-if="isOt" severity="warn" data-testid="ss-ot-warning">
    <p>{{ t('scopeSettingsPanel.ot.warning') }}</p>
  </AppMessage>

  <!-- Nmap arguments (main + pre-scan). Pre-scan is not applicable to OT.
         Two equal halves so each argument box gets half the line. -->
  <div
    class="form-scope-settings__grid form-scope-settings__grid--halves"
    :class="{ 'form-scope-settings__field--disabled': isDiscoveryDisabled }"
  >
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nmapArgs.label')"
      :help="t('scopeSettingsPanel.fields.nmapArgs.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.discovery_nmap_args"
        data-testid="ss-nmap-args"
        rows="2"
        class="form-scope-settings__textarea"
        :placeholder="
          isOt
            ? t('scopeSettingsPanel.fields.nmapArgs.otPlaceholder')
            : t('scopeSettingsPanel.fields.nmapArgs.placeholder')
        "
        :disabled="isDiscoveryDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nmapArgsPrescan.label')"
      :help="t('scopeSettingsPanel.fields.nmapArgsPrescan.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.discovery_nmap_args_prescan"
        data-testid="ss-nmap-args-prescan"
        rows="2"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nmapArgsPrescan.placeholder')"
        :disabled="isDiscoveryDisabled || isOt"
      />
    </Field>
  </div>
  <Field
    v-slot="{ id }"
    :label="t('scopeSettingsPanel.fields.excludeDomains.label')"
    :help="t('scopeSettingsPanel.fields.excludeDomains.hint')"
    tooltip
  >
    <Textarea
      :id="id"
      v-model="form.discovery_exclude_domains"
      rows="2"
      class="form-scope-settings__textarea"
      :placeholder="t('scopeSettingsPanel.fields.excludeDomains.placeholder')"
      :disabled="isDiscoveryDisabled"
    />
  </Field>

  <!-- HTTP user agent (both web modules, 3/4 line) + shared network timeout (1/4). -->
  <div class="form-scope-settings__grid form-scope-settings__grid--3-1">
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.httpUserAgent.label')"
      :help="t('scopeSettingsPanel.fields.httpUserAgent.hint')"
      tooltip
      icon="pi pi-user"
    >
      <InputText
        :id="id"
        v-model="form.http_user_agent"
        :placeholder="t('scopeSettingsPanel.fields.httpUserAgent.placeholder')"
        :disabled="isHttpDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.networkTimeout.label')"
      :help="t('scopeSettingsPanel.fields.networkTimeout.hint')"
      tooltip
    >
      <FieldNumber
        :id="id"
        v-model="form.network_timeout_seconds"
        :min="1"
        :placeholder="t('scopeSettingsPanel.fields.networkTimeout.placeholder')"
        :disabled="isNetworkTimeoutDisabled"
        :unit="t('scopeSettingsPanel.fields.networkTimeout.unit')"
      />
    </Field>
  </div>

  <div
    class="form-scope-settings__grid"
    :class="{ 'form-scope-settings__field--disabled': isSmbDisabled }"
  >
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbForcedShares.label')"
      :help="t('scopeSettingsPanel.fields.smbForcedShares.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.smb_forced_shares"
        rows="3"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.smbForcedShares.placeholder')"
        :disabled="isSmbDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbExcludeShares.label')"
      :help="t('scopeSettingsPanel.fields.smbExcludeShares.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.smb_exclude_shares"
        rows="3"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.smbExcludeShares.placeholder')"
        :disabled="isSmbDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbExcludeFolders.label')"
      :help="t('scopeSettingsPanel.fields.smbExcludeFolders.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.smb_exclude_folders"
        rows="3"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.smbExcludeFolders.placeholder')"
        :disabled="isSmbDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbExcludeExtensions.label')"
      :help="t('scopeSettingsPanel.fields.smbExcludeExtensions.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.smb_exclude_extensions"
        rows="3"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.smbExcludeExtensions.placeholder')"
        :disabled="isSmbDisabled"
      />
    </Field>
  </div>
  <div
    class="form-scope-settings__grid"
    :class="{ 'form-scope-settings__field--disabled': isSmbDisabled }"
  >
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbExcludeFileSizeBelow.label')"
      :help="t('scopeSettingsPanel.fields.smbExcludeFileSizeBelow.hint')"
      tooltip
    >
      <FieldNumber
        :id="id"
        v-model="form.smb_exclude_file_size_below"
        :min="-1"
        :placeholder="t('scopeSettingsPanel.fields.smbExcludeFileSizeBelow.placeholder')"
        :disabled="isSmbDisabled"
        :unit="t('scopeSettingsPanel.fields.smbExcludeFileSizeBelow.unit')"
        data-testid="ss-smb-filesize"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbExcludeLastModified.label')"
      :help="t('scopeSettingsPanel.fields.smbExcludeLastModified.hint')"
      tooltip
    >
      <DatePicker
        v-model="smbLastModified"
        :input-id="id"
        show-time
        hour-format="24"
        date-format="dd.mm.yy"
        show-icon
        icon-display="input"
        icon="pi pi-calendar"
        :disabled="isSmbDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbDepth.label')"
      :help="t('scopeSettingsPanel.fields.smbDepth.hint')"
      tooltip
    >
      <FieldNumber
        :id="id"
        v-model="form.smb_depth"
        data-testid="ss-smb-depth"
        :min="-1"
        :placeholder="t('scopeSettingsPanel.fields.smbDepth.placeholder')"
        :disabled="isSmbDisabled"
        :unit="t('scopeSettingsPanel.fields.smbDepth.unit')"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.smbThreads.label')"
      :help="t('scopeSettingsPanel.fields.smbThreads.hint')"
      tooltip
    >
      <FieldNumber
        :id="id"
        v-model="form.smb_threads"
        :min="1"
        :placeholder="t('scopeSettingsPanel.fields.smbThreads.placeholder')"
        :disabled="isSmbDisabled"
        :unit="t('scopeSettingsPanel.fields.smbThreads.unit')"
      />
    </Field>
  </div>

  <div
    class="form-scope-settings__grid"
    :class="{ 'form-scope-settings__field--disabled': isNucleiDisabled }"
  >
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiIncludeSeverities.label')"
      :help="t('scopeSettingsPanel.fields.nucleiIncludeSeverities.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_include_severities"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiIncludeSeverities.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiExcludeSeverities.label')"
      :help="t('scopeSettingsPanel.fields.nucleiExcludeSeverities.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_exclude_severities"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiExcludeSeverities.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiIncludeTags.label')"
      :help="t('scopeSettingsPanel.fields.nucleiIncludeTags.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_include_tags"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiIncludeTags.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiExcludeTags.label')"
      :help="t('scopeSettingsPanel.fields.nucleiExcludeTags.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_exclude_tags"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiExcludeTags.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiIncludeIds.label')"
      :help="t('scopeSettingsPanel.fields.nucleiIncludeIds.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_include_ids"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiIncludeIds.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiExcludeIds.label')"
      :help="t('scopeSettingsPanel.fields.nucleiExcludeIds.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_exclude_ids"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiExcludeIds.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiIncludeProtocols.label')"
      :help="t('scopeSettingsPanel.fields.nucleiIncludeProtocols.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_include_protocols"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiIncludeProtocols.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.nucleiExcludeProtocols.label')"
      :help="t('scopeSettingsPanel.fields.nucleiExcludeProtocols.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.nuclei_exclude_protocols"
        rows="1"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.nucleiExcludeProtocols.placeholder')"
        :disabled="isNucleiDisabled"
      />
    </Field>
  </div>

  <!-- Web crawler tuning: follow-types textarea takes 3/4 of the line; depth
         and threads stack vertically inside the remaining 1/4. -->
  <div
    class="form-scope-settings__grid form-scope-settings__grid--3-1"
    :class="{ 'form-scope-settings__field--disabled': isWebcrawlerDisabled }"
  >
    <Field
      v-slot="{ id }"
      :label="t('scopeSettingsPanel.fields.webcrawlerFollowTypes.label')"
      :help="t('scopeSettingsPanel.fields.webcrawlerFollowTypes.hint')"
      tooltip
    >
      <Textarea
        :id="id"
        v-model="form.webcrawler_follow_types"
        rows="5"
        class="form-scope-settings__textarea"
        :placeholder="t('scopeSettingsPanel.fields.webcrawlerFollowTypes.placeholder')"
        :disabled="isWebcrawlerDisabled"
      />
    </Field>
    <div class="form-scope-settings__stack">
      <Field
        v-slot="{ id }"
        :label="t('scopeSettingsPanel.fields.webcrawlerDepth.label')"
        :help="t('scopeSettingsPanel.fields.webcrawlerDepth.hint')"
        tooltip
      >
        <FieldNumber
          :id="id"
          v-model="form.webcrawler_depth"
          :min="-1"
          :placeholder="t('scopeSettingsPanel.fields.webcrawlerDepth.placeholder')"
          :disabled="isWebcrawlerDisabled"
          :unit="t('scopeSettingsPanel.fields.webcrawlerDepth.unit')"
        />
      </Field>
      <Field
        v-slot="{ id }"
        :label="t('scopeSettingsPanel.fields.webcrawlerMaxThreads.label')"
        :help="t('scopeSettingsPanel.fields.webcrawlerMaxThreads.hint')"
        tooltip
      >
        <FieldNumber
          :id="id"
          v-model="form.webcrawler_max_threads"
          :min="1"
          :placeholder="t('scopeSettingsPanel.fields.webcrawlerMaxThreads.placeholder')"
          :disabled="isWebcrawlerDisabled"
          :unit="t('scopeSettingsPanel.fields.webcrawlerMaxThreads.unit')"
        />
      </Field>
    </div>
  </div>

  <!-- Bottom row: the three web toggles each take a quarter of the line, with
         the Cancel/Save actions filling the fourth quarter (legacy parity). -->
  <div
    class="form-scope-settings__grid form-scope-settings__grid--quarters form-scope-settings__bottom-row"
  >
    <FieldToggle
      v-model="form.webcrawler_follow_query_strings"
      :label="t('scopeSettingsPanel.fields.webcrawlerFollowQs.label')"
      :help="t('scopeSettingsPanel.fields.webcrawlerFollowQs.hint')"
      tooltip
      :disabled="isWebcrawlerDisabled"
    />
    <FieldToggle
      v-model="form.webcrawler_always_store_root"
      :label="t('scopeSettingsPanel.fields.webcrawlerStoreRoot.label')"
      :help="t('scopeSettingsPanel.fields.webcrawlerStoreRoot.hint')"
      tooltip
      :disabled="isWebcrawlerDisabled"
    />
    <FieldToggle
      v-model="form.webenum_probe_robots"
      :label="t('scopeSettingsPanel.fields.webenumProbeRobots.label')"
      :help="t('scopeSettingsPanel.fields.webenumProbeRobots.hint')"
      tooltip
      :disabled="isWebenumDisabled"
    />

    <div class="form-scope-settings__actions">
      <LayoutSubmitFooter
        cancel-test-id="ss-cancel"
        submit-test-id="ss-submit"
        :text-submit-label="t('common.actions.save')"
        @cancel="cancel"
      />
    </div>
  </div>
</template>

<style scoped>
/* Separate discovery timespans from the following settings section */
.form-scope-settings__timespans {
  padding-bottom: 1rem;
}

.form-scope-settings__field--disabled {
  opacity: 0.55;
}

/* ── Field grid layout ─────────────────────────────────────────── */

.form-scope-settings__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
  gap: 0.75rem;
}

/* Agent-process sliders sit in exactly two columns (legacy parity: the old form
   laid the sliders out two-per-row). Collapses to one column on narrow screens. */
.form-scope-settings__grid--sliders {
  grid-template-columns: repeat(2, 1fr);
  column-gap: 1rem;
  row-gap: 0.35rem;
  padding-bottom: 0.5rem;
}

.form-scope-settings__grid--sliders :deep(.field-slider--inline-label) {
  gap: 0.75rem;
}

.form-scope-settings__grid--sliders :deep(.field-slider__value .p-inputtext) {
  height: 1.75rem;
  padding-block: 0.15rem;
}

@media (max-width: 768px) {
  .form-scope-settings__grid--sliders {
    grid-template-columns: 1fr;
  }
}

/* Identity/top row: read-only Group + Scope Name (a quarter each) beside the
   editable Sensitive Ports (the remaining half). Four equal tracks with the
   third child spanning two keeps the center gap exactly at 50%. Inputs fill
   their cell. */
.form-scope-settings__grid--identity {
  grid-template-columns: repeat(4, 1fr);
}

.form-scope-settings__grid--identity > :last-child {
  grid-column: span 2;
}

.form-scope-settings__grid--identity :deep(.p-inputtext) {
  width: 100%;
}

/* Fixed proportional splits (override the default auto-fill grid): two equal
   halves, a 3/4 + 1/4 split, and four equal quarters. Fields fill their cell. */
.form-scope-settings__grid--halves {
  grid-template-columns: 1fr 1fr;
}

/* 3/4 + 1/4 split. Use the SAME four equal columns as the plain grids above so
   the gaps line up: the first child spans three columns (matching the width of
   three fields + their two gaps), the second spans the last column. minmax(0,1fr)
   lets cells shrink below their content's intrinsic width. */
.form-scope-settings__grid--3-1 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.form-scope-settings__grid--3-1 > *:first-child {
  grid-column: span 3;
}

.form-scope-settings__grid--3-1 > *:last-child {
  grid-column: span 1;
}

.form-scope-settings__grid--quarters {
  grid-template-columns: repeat(4, 1fr);
}

.form-scope-settings__grid--3-1 :deep(.p-inputtext) {
  width: 100%;
}

/* Stacked column: the depth/threads pair sits vertically inside the last quarter
   of the web-crawler row, each filling the column width. */
.form-scope-settings__stack {
  display: flex;
  flex-direction: column;
}

/* Multi-row argument textareas stretch to the field width and wrap long
   argument strings. */
.form-scope-settings__textarea {
  width: 100%;
  font-family: var(--p-font-family, inherit);
  resize: vertical;
}

/* Bottom row: keep the toggles and the actions cell vertically centered on one
   another so the buttons line up with the toggle switches. */
.form-scope-settings__bottom-row {
  align-items: center;
}

.form-scope-settings__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0;
}
</style>
