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

import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import Button from 'primevue/button'
import Steps from 'primevue/steps'
import Tooltip from 'primevue/tooltip'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppMessage from '@/components/elements/AppMessage.vue'
import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import LinuxIcon from '@/components/elements/LinuxIcon.vue'
import SectionHeader from '@/components/elements/SectionHeader.vue'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    installation: {
      title: 'Installation Guide',
      help: 'Running one or more scan agent(s) is required if you want to operate your own scan scope. Please check the support matrix to decide which operating system to use and follow the other steps to successfully deploy your scan agent.',
      warning: {
        line1:
          'Please make sure you understand the implications of executing network scans, as done by this scan agent. Deploy and run this scan agent on your own risk. Ensure your scan scope is limited to systems you are responsible for!',
        line2:
          'The scan agent does not execute vulnerability scans, hence no dangerous payloads are sent against target systems. The scan agent merely uses standard protocols and enumeration techniques in order to discover systems and data. Nevertheless, an impact on your production networks may arise. In some edge cases, targets might show reduced response times or crash due to unexpected request parameters.',
      },
      step: {
        supportMatrix: 'Support Matrix',
        nmap: 'Prepare Nmap',
        sslyze: 'Prepare Sslyze',
        agent: 'Install Scan Agent',
        launch: 'Launch Scan Agent',
      },
      platform: {
        windows: 'Windows',
        linux: 'Linux',
        domain: 'Windows (Domain)',
      },
      matrix: {
        info: 'The support matrix below shows which scan agent features will work, depending on the operating environment the scan agent is running on.',
        support: {
          full: 'Fully Supported',
          partial: 'Partially Supported',
          none: 'Not Supported',
          na: 'Not Applicable',
        },
        row: {
          discoveryModule: {
            title: 'Discovery Module',
            comment:
              'The discovery module executes devices discovery, port scanning and service detection and is required to run. Furthermore, it expands scan results with information from various repositories.',
          },
          deviceDiscovery: {
            title: 'Device Discovery',
          },
          otDiscovery: {
            title: 'OT Discovery',
            platformComment:
              'Only network-based protocols (NDP, mDNS/SSDP) are supported. L2 protocols (PROFINET DCP, EtherCAT, LLDP) require raw Ethernet sockets which are not available on Windows.',
          },
          portEnumeration: {
            title: 'Port Enumeration',
          },
          serviceDetection: {
            title: 'Service Detection',
          },
          hostnameDiscovery: {
            title: 'Hostname Discovery',
          },
          remoteInterfaces: {
            title: 'Extraction Of Remote Interfaces',
            domainComment: 'Via implicit authentication.',
          },
          adminRdpUsers: {
            title: 'Extraction Of Admin/Rdp Users',
            domainComment: 'Via implicit authentication.',
          },
          adLookup: {
            title: 'Active Directory Lookup',
            linuxComment:
              'You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target’s domain is queried, but you can define a static LDAP server.',
            windowsComment:
              'You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target’s domain is queried, but you can define a static LDAP server.',
            domainComment:
              'Implicit authentication without credentials can be used. You can configure explicit credentials. Cross-domain authentication is possible via GSSAPI. By default the discovered target’s domain is queried, but you can define a static LDAP server.',
          },
          assetInventory: {
            title: 'Asset Inventory Lookup',
          },
          bannerGrabbing: {
            title: 'Banner Grabbing',
            comment:
              'The banner grabbing module connects to discovered ports and extracts data returned by the remote service.',
          },
          smbCrawling: {
            title: 'SMB Crawling',
            comment:
              'The SMB crawling module iterates shares and folders to discover accessible and/or writable files.',
          },
          smbMimeType: {
            title: 'Mime Type Detection',
          },
          smbMip: {
            title: 'Microsoft Information Protection',
          },
          nfsCrawling: {
            title: 'NFS Crawling',
            comment:
              'The NFS crawling module mounts NFS shares and iterates folders to discover accessible and/or writable files.',
          },
          nfsv3: {
            title: 'NFSv3',
          },
          nfsv4: {
            title: 'NFSv4',
          },
          unixAcl: {
            title: 'Unix ACL Flags',
          },
          nfsMimeType: {
            title: 'Mime Type Detection',
          },
          nfsMip: {
            title: 'Microsoft Information Protection',
          },
          nuclei: {
            title: 'Nuclei Scanning',
            comment:
              'The Nuclei module leverages the large list of community-curated templates at ProjectDiscovery to find security vulnerabilities.',
          },
          webCrawling: {
            title: 'Web Crawling',
            comment:
              'The web crawling module crawls web services to extract links, response headers and HTML contents.',
          },
          webEnumeration: {
            title: 'Web Enumeration',
            comment:
              'The web enumeration module guesses common URLs in order to discover common sensitive hidden resources.',
          },
          sslEnumeration: {
            title: 'SSL Enumeration',
            comment:
              'The SSL module enumerates existing SSL protocols and ciphers in order to detect deployed configurations and vulnerabilities.',
          },
          sshEnumeration: {
            title: 'SSH Enumeration',
            comment:
              'The SSH module enumerates existing SSH protocols and ciphers in order to detect deployed configurations and vulnerabilities.',
          },
        },
      },
      nmap: {
        windows: {
          download: {
            title: 'Download Nmap',
            linkLabel: 'Nmap',
          },
          install: {
            title: 'Install Nmap',
            body: 'Run the installer or extract the standalone version into your desired destination folder.',
          },
          npcap: {
            title: 'Check Npcap',
            body1: 'On Windows, Nmap requires either WinPcap or the newer Npcap installed.',
            body2:
              'If you did *not* use the Nmap installer, manually install Npcap shipped with the standalone version.',
          },
        },
        linux: {
          install: {
            title: 'Install Nmap',
            command: 'sudo apt-get install nmap',
          },
        },
      },
      sslyze: {
        info: 'SSLyze is only required if the SSL module, collecting data about SSL endpoints, is enabled in your scan scope.',
        windows: {
          download: {
            title: 'Download SSLyze',
            linkLabel: 'SSLyze',
            warning: 'Please choose the executable version!',
          },
          extract: {
            title: 'Extract SSLyze',
            body: 'Extract SSLyze into your desired destination folder.',
          },
        },
        linux: {
          install: {
            title: 'Install SSLyze',
            command: 'pip install --upgrade sslyze',
            warning: 'SSLyze requires Python >=3.7!',
          },
        },
      },
      agent: {
        download: {
          title: 'Download Scan Agent',
          linkLabel: 'Scan Agent',
        },
        setup: {
          title: 'Setup Environment',
          intro:
            'Some changes need to be made to your system in order to run the scan agent. The scan agent will take care of this for you:',
          windowsCommand: './agent.exe --setup',
          linuxCommand: './agent.bin --setup',
          windowsRightsIntro: 'This setup routine requires admin rights to:',
          windowsRights1: 'Apply Nmap performance registry patch',
          windowsRights2: 'Whitelist Nmap in Firewall settings',
          windowsRights3: 'Enable NFS control commands',
          linuxRightsIntro: 'This setup routine requires admin rights to:',
          linuxRights1: 'Add the executing user to sudoers file to allow NFS commands',
        },
        configure: {
          title: 'Configure Scan Agent',
          intro: 'Open agent.conf and adapt your settings as required:',
          nmap: 'Set path to Nmap executable',
          python: 'Set path to Python executable',
          nuclei: 'Set path to Nuclei templates',
          sslyze: 'Set path to SSLyze executable',
          secret: 'Set scan scope secret to associate scan agent',
          info: 'If no config file exists, run the agent to generate a default one.',
        },
      },
      launch: {
        windows: {
          title: 'Launch Scan Agent',
          command: './agent.exe',
          warning:
            'The scan agent does *not* require admin privileges to execute! Please run it in user mode.',
        },
        linux: {
          title: 'Launch Scan Agent',
          command: './agent',
          warning:
            'The scan agent does *not* require admin privileges to execute! Please run it in user mode.',
        },
        info: {
          secret:
            'The scan agent authenticates using the configured scope secret and will be assigned to the according scan scope by the broker.',
          processIntro: 'Scan process:',
          process1: 'The scan agent will connect to the broker and request scan tasks to execute',
          process2:
            'The broker manages outstanding scan tasks and assigns them to the scan agent accordingly',
          process3: 'After successful execution, the scan agent sends results to the broker',
          multiple:
            'Multiple scan agents can work on the same scan scope in order to boost execution time.',
        },
      },
    },
  },
}

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })

// Local directive registration (Tooltip is not registered globally in main.ts)
const vTooltip = Tooltip

// ── Reactive state ────────────────────────────────────────────
const Support = { NA: -1, NONE: 0, FULL: 1, PARTIAL: 2 } as const

type SupportLevel = (typeof Support)[keyof typeof Support]

interface MatrixRow {
  key: string
  sub: boolean
  linux: SupportLevel
  windows: SupportLevel
  domain: SupportLevel
  hasComment?: boolean
  linuxComment?: string
  windowsComment?: string
  domainComment?: string
}

// Shared i18n key prefix for per-row comment lookups
const ROW = 'installation.matrix.row'

const SUPPORT_MATRIX: readonly MatrixRow[] = [
  { key: 'discoveryModule', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'deviceDiscovery', sub: true, linux: 1, windows: 1, domain: 1 },
  {
    key: 'otDiscovery',
    sub: true,
    linux: 1,
    windows: 2,
    domain: 2,
    windowsComment: `${ROW}.otDiscovery.platformComment`,
    domainComment: `${ROW}.otDiscovery.platformComment`,
  },
  { key: 'portEnumeration', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'serviceDetection', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'hostnameDiscovery', sub: true, linux: 1, windows: 1, domain: 1 },
  {
    key: 'remoteInterfaces',
    sub: true,
    linux: 1,
    windows: 1,
    domain: 1,
    domainComment: `${ROW}.remoteInterfaces.domainComment`,
  },
  {
    key: 'adminRdpUsers',
    sub: true,
    linux: 0,
    windows: 0,
    domain: 1,
    domainComment: `${ROW}.adminRdpUsers.domainComment`,
  },
  {
    key: 'adLookup',
    sub: true,
    linux: 1,
    windows: 1,
    domain: 1,
    linuxComment: `${ROW}.adLookup.linuxComment`,
    windowsComment: `${ROW}.adLookup.windowsComment`,
    domainComment: `${ROW}.adLookup.domainComment`,
  },
  { key: 'assetInventory', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'bannerGrabbing', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'smbCrawling', sub: false, linux: 0, windows: 1, domain: 1, hasComment: true },
  { key: 'smbMimeType', sub: true, linux: -1, windows: 1, domain: 1 },
  { key: 'smbMip', sub: true, linux: -1, windows: 1, domain: 1 },
  { key: 'nfsCrawling', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'nfsv3', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'nfsv4', sub: true, linux: 1, windows: 0, domain: 0 },
  { key: 'unixAcl', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'nfsMimeType', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'nfsMip', sub: true, linux: 1, windows: 1, domain: 1 },
  { key: 'nuclei', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'webCrawling', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'webEnumeration', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'sslEnumeration', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
  { key: 'sshEnumeration', sub: false, linux: 1, windows: 1, domain: 1, hasComment: true },
] as const

// Authorization (admin/owner only) is enforced by the router guard, which logs
// out and redirects unauthorized users to /login. No in-component gate needed.

const activeStep = ref(0)

interface StepDef {
  key: string
  labelKey: string
  icon: string
}

const STEPS: readonly StepDef[] = [
  // Icon parity with legacy `ui vertical steps` (installation.html):
  //   table → wrench → screwdriver → hammer → rocket.
  // PrimeIcons 7 has no screwdriver (sslyze keeps pi-cog) and no rocket
  // (launch keeps pi-send, the closest available). See MR "CAN'T MATCH" notes.
  { key: 'supportMatrix', labelKey: 'installation.step.supportMatrix', icon: 'pi pi-table' },
  { key: 'nmap', labelKey: 'installation.step.nmap', icon: 'pi pi-wrench' },
  { key: 'sslyze', labelKey: 'installation.step.sslyze', icon: 'pi pi-cog' },
  { key: 'agent', labelKey: 'installation.step.agent', icon: 'pi pi-hammer' },
  { key: 'launch', labelKey: 'installation.step.launch', icon: 'pi pi-send' },
]

const warningVisible = ref(true)

// ── Computed ──────────────────────────────────────────────────
const stepItems = computed(() =>
  STEPS.map((s, index) => ({
    key: s.key,
    label: t(s.labelKey),
    icon: s.icon,
    command: () => {
      activeStep.value = index
    },
  })),
)

// ── Functions and handlers ────────────────────────────────────
// Resolve each platform cell once so the template can narrow nullable cells naturally
const supportMatrixRows = computed(() =>
  SUPPORT_MATRIX.map((row) => ({
    ...row,
    linuxCell: cell(row.linux, row.linuxComment),
    windowsCell: cell(row.windows, row.windowsComment),
    domainCell: cell(row.domain, row.domainComment),
  })),
)

function goToStep(index: number): void {
  activeStep.value = index
}

interface CellInfo {
  /** PrimeVue icon class, or null for "not applicable" (renders nothing). */
  icon: string | null
  /** CSS modifier mapping to the brand color. */
  tone: 'full' | 'partial' | 'none'
  tooltip: string
}

// Localized fallback tooltip for a support level when no per-cell override exists
function defaultLabel(level: SupportLevel): string {
  switch (level) {
    case Support.FULL:
      return t('installation.matrix.support.full')
    case Support.PARTIAL:
      return t('installation.matrix.support.partial')
    case Support.NONE:
      return t('installation.matrix.support.none')
    default:
      return t('installation.matrix.support.na')
  }
}

function cell(level: SupportLevel, overrideKey?: string): CellInfo | null {
  // Not-applicable cells render nothing
  if (level === Support.NA) return null

  // Prefer a per-cell override tooltip, otherwise fall back to the level label
  const tooltip = overrideKey ? t(overrideKey) : defaultLabel(level)
  if (level === Support.FULL) return { icon: 'pi pi-check', tone: 'full', tooltip }
  if (level === Support.PARTIAL) return { icon: 'pi pi-wave-pulse', tone: 'partial', tooltip }
  return { icon: 'pi pi-times', tone: 'none', tooltip }
}
</script>

<template>
  <section class="page-section installation-page">
    <SectionHeader
      class="installation-page__title"
      test-id="installation-help"
      :title="t('installation.title')"
      :help-text="t('installation.help')"
    />

    <AppMessage
      v-if="warningVisible"
      severity="warn"
      closable
      class="installation-page__warning"
      @close="warningVisible = false"
    >
      <p>{{ t('installation.warning.line1') }}</p>
      <p>{{ t('installation.warning.line2') }}</p>
    </AppMessage>

    <!-- Legacy parity: steps render as a vertical left-column sidebar
         (`ui vertical steps`), with the content panel to its right. -->
    <div class="installation-page__body">
      <Steps
        v-model:active-step="activeStep"
        :model="stepItems"
        :readonly="false"
        class="installation-page__steps"
      >
        <template #item="{ item, active, props }">
          <a
            :class="['nav-item', { 'nav-item--active': active }]"
            :data-testid="`installation-step-${item.key}`"
            v-bind="props.action"
            :aria-current="active ? 'step' : undefined"
          >
            <span :class="['nav-item__icon', item.icon]" aria-hidden="true"></span>
            <span class="nav-item__label">{{ item.label }}</span>
          </a>
        </template>
      </Steps>

      <div class="installation-page__panel">
        <!-- Directional slide between steps (legacy switched instantly). -->
        <Transition name="step-reveal" mode="out-in">
          <!-- Step 0: Support Matrix -->
          <div
            v-if="activeStep === 0"
            key="step-0"
            data-testid="installation-content-supportMatrix"
          >
            <AppMessage severity="info" :message="t('installation.matrix.info')" />
            <div class="installation-page__table-wrap">
              <table class="installation-page__support-matrix" data-testid="installation-matrix">
                <thead>
                  <tr>
                    <th scope="col"></th>
                    <th scope="col" class="installation-page__support-matrix-col">
                      <LinuxIcon class="installation-page__support-matrix-os-icon" />
                      &nbsp;{{ t('installation.platform.linux') }}
                    </th>
                    <th scope="col" class="installation-page__support-matrix-col">
                      <i
                        class="pi pi-microsoft installation-page__support-matrix-os-icon"
                        aria-hidden="true"
                      ></i>
                      &nbsp;&nbsp;{{ t('installation.platform.windows') }}
                    </th>
                    <th scope="col" class="installation-page__support-matrix-col">
                      <i
                        class="pi pi-microsoft installation-page__support-matrix-os-icon"
                        aria-hidden="true"
                      ></i>
                      &nbsp;&nbsp;{{ t('installation.platform.domain') }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in supportMatrixRows" :key="row.key">
                    <td :class="{ 'installation-page__support-matrix-sub': row.sub }">
                      <i
                        v-if="row.sub"
                        class="pi pi-angle-right installation-page__support-matrix-sub-arrow"
                        aria-hidden="true"
                      ></i>
                      <span>{{ t(`installation.matrix.row.${row.key}.title`) }}</span>
                      <i
                        v-if="row.hasComment"
                        v-tooltip.right="t(`installation.matrix.row.${row.key}.comment`)"
                        class="pi pi-question-circle installation-page__support-matrix-hint"
                        :aria-label="t(`installation.matrix.row.${row.key}.comment`)"
                      ></i>
                    </td>
                    <td class="installation-page__support-matrix-cell">
                      <i
                        v-if="row.linuxCell"
                        v-tooltip.top="row.linuxCell.tooltip"
                        :class="[
                          row.linuxCell.icon,
                          `installation-page__support-matrix-icon--${row.linuxCell.tone}`,
                        ]"
                        :aria-label="row.linuxCell.tooltip"
                      ></i>
                    </td>
                    <td class="installation-page__support-matrix-cell">
                      <i
                        v-if="row.windowsCell"
                        v-tooltip.top="row.windowsCell.tooltip"
                        :class="[
                          row.windowsCell.icon,
                          `installation-page__support-matrix-icon--${row.windowsCell.tone}`,
                        ]"
                        :aria-label="row.windowsCell.tooltip"
                      ></i>
                    </td>
                    <td class="installation-page__support-matrix-cell">
                      <i
                        v-if="row.domainCell"
                        v-tooltip.top="row.domainCell.tooltip"
                        :class="[
                          row.domainCell.icon,
                          `installation-page__support-matrix-icon--${row.domainCell.tone}`,
                        ]"
                        :aria-label="row.domainCell.tooltip"
                      ></i>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Step 1: Prepare Nmap -->
          <div
            v-else-if="activeStep === 1"
            key="step-1"
            class="installation-page__two-col"
            data-testid="installation-content-nmap"
          >
            <div class="installation-page__col">
              <SectionHeader
                level="h4"
                icon="pi pi-microsoft"
                :title="t('installation.platform.windows')"
              />
              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>{{
                    t('installation.nmap.windows.download.title')
                  }}</AccordionHeader>
                  <AccordionContent>
                    <div class="installation-page__link-box">
                      <i class="pi pi-link" aria-hidden="true"></i
                      ><a href="https://nmap.org/download.html" target="_blank" rel="noopener">
                        {{ t('installation.nmap.windows.download.linkLabel') }}
                      </a>
                    </div>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="1">
                  <AccordionHeader>{{
                    t('installation.nmap.windows.install.title')
                  }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.nmap.windows.install.body') }}</p>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="2">
                  <AccordionHeader>{{
                    t('installation.nmap.windows.npcap.title')
                  }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.nmap.windows.npcap.body1') }}</p>
                    <p>{{ t('installation.nmap.windows.npcap.body2') }}</p>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </div>
            <div class="installation-page__col">
              <SectionHeader level="h4">
                <LinuxIcon />
                {{ t('installation.platform.linux') }}
              </SectionHeader>
              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>{{
                    t('installation.nmap.linux.install.title')
                  }}</AccordionHeader>
                  <AccordionContent>
                    <CodeSnippet
                      :text="t('installation.nmap.linux.install.command')"
                      kind="#>"
                      test-id="nmap-linux"
                    />
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </div>
          </div>

          <!-- Step 2: Prepare Sslyze -->
          <div v-else-if="activeStep === 2" key="step-2" data-testid="installation-content-sslyze">
            <div class="installation-page__two-col">
              <div class="installation-page__col">
                <SectionHeader
                  level="h4"
                  icon="pi pi-microsoft"
                  :title="t('installation.platform.windows')"
                />
                <Accordion value="0">
                  <AccordionPanel value="0">
                    <AccordionHeader>{{
                      t('installation.sslyze.windows.download.title')
                    }}</AccordionHeader>
                    <AccordionContent>
                      <div class="installation-page__link-box">
                        <i class="pi pi-link" aria-hidden="true"></i
                        ><a
                          href="https://github.com/nabla-c0d3/sslyze/releases"
                          target="_blank"
                          rel="noopener"
                        >
                          {{ t('installation.sslyze.windows.download.linkLabel') }}
                        </a>
                      </div>

                      <AppMessage
                        severity="warn"
                        class="installation-page__inline-msg"
                        :message="t('installation.sslyze.windows.download.warning')"
                      />
                    </AccordionContent>
                  </AccordionPanel>
                  <AccordionPanel value="1">
                    <AccordionHeader>{{
                      t('installation.sslyze.windows.extract.title')
                    }}</AccordionHeader>
                    <AccordionContent>
                      <p>{{ t('installation.sslyze.windows.extract.body') }}</p>
                    </AccordionContent>
                  </AccordionPanel>
                </Accordion>
              </div>
              <div class="installation-page__col">
                <SectionHeader level="h4">
                  <LinuxIcon />
                  {{ t('installation.platform.linux') }}
                </SectionHeader>
                <Accordion value="0">
                  <AccordionPanel value="0">
                    <AccordionHeader>{{
                      t('installation.sslyze.linux.install.title')
                    }}</AccordionHeader>
                    <AccordionContent>
                      <CodeSnippet
                        :text="t('installation.sslyze.linux.install.command')"
                        kind="#>"
                      />
                      <AppMessage
                        severity="warn"
                        class="installation-page__inline-msg"
                        :message="t('installation.sslyze.linux.install.warning')"
                      />
                    </AccordionContent>
                  </AccordionPanel>
                </Accordion>
              </div>
            </div>
            <AppMessage
              severity="info"
              class="installation-page__wide-message"
              :message="t('installation.sslyze.info')"
            />
          </div>

          <!-- Step 3: Install Scan Agent -->
          <div
            v-else-if="activeStep === 3"
            key="step-3"
            class="installation-page__two-col"
            data-testid="installation-content-agent"
          >
            <div class="installation-page__col">
              <SectionHeader
                level="h4"
                icon="pi pi-microsoft"
                :title="t('installation.platform.windows')"
              />
              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>{{ t('installation.agent.download.title') }}</AccordionHeader>
                  <AccordionContent>
                    <div class="installation-page__link-box">
                      <i class="pi pi-download" aria-hidden="true"></i
                      ><a href="/app/downloads/agent.zip">
                        {{ t('installation.agent.download.linkLabel') }}
                      </a>
                    </div>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="1">
                  <AccordionHeader>{{ t('installation.agent.setup.title') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.agent.setup.intro') }}</p>
                    <CodeSnippet
                      :text="t('installation.agent.setup.windowsCommand')"
                      class="installation-page__setup-command"
                      kind="#>"
                    />
                    <p>
                      {{ t('installation.agent.setup.windowsRightsIntro') }}
                    </p>
                    <ul>
                      <li>{{ t('installation.agent.setup.windowsRights1') }}</li>
                      <li>{{ t('installation.agent.setup.windowsRights2') }}</li>
                      <li>{{ t('installation.agent.setup.windowsRights3') }}</li>
                    </ul>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="2">
                  <AccordionHeader>{{ t('installation.agent.configure.title') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.agent.configure.intro') }}</p>
                    <ul>
                      <li>{{ t('installation.agent.configure.nmap') }}</li>
                      <li>{{ t('installation.agent.configure.sslyze') }}</li>
                      <li>{{ t('installation.agent.configure.nuclei') }}</li>
                      <li>{{ t('installation.agent.configure.secret') }}</li>
                    </ul>
                    <AppMessage
                      severity="info"
                      class="installation-page__inline-msg"
                      :message="t('installation.agent.configure.info')"
                    />
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </div>
            <div class="installation-page__col">
              <SectionHeader level="h4">
                <LinuxIcon />
                {{ t('installation.platform.linux') }}
              </SectionHeader>
              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>{{ t('installation.agent.download.title') }}</AccordionHeader>
                  <AccordionContent>
                    <div class="installation-page__link-box">
                      <i class="pi pi-download" aria-hidden="true"></i
                      ><a href="/app/downloads/agent.zip">
                        {{ t('installation.agent.download.linkLabel') }}
                      </a>
                    </div>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="1">
                  <AccordionHeader>{{ t('installation.agent.setup.title') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.agent.setup.intro') }}</p>
                    <CodeSnippet
                      :text="t('installation.agent.setup.linuxCommand')"
                      class="installation-page__setup-command"
                      kind="#>"
                    />
                    <p>{{ t('installation.agent.setup.linuxRightsIntro') }}</p>
                    <ul>
                      <li>{{ t('installation.agent.setup.linuxRights1') }}</li>
                    </ul>
                  </AccordionContent>
                </AccordionPanel>
                <AccordionPanel value="2">
                  <AccordionHeader>{{ t('installation.agent.configure.title') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ t('installation.agent.configure.intro') }}</p>
                    <ul>
                      <li>{{ t('installation.agent.configure.nmap') }}</li>
                      <li>{{ t('installation.agent.configure.python') }}</li>
                      <li>{{ t('installation.agent.configure.nuclei') }}</li>
                      <li>{{ t('installation.agent.configure.secret') }}</li>
                    </ul>
                    <AppMessage
                      severity="info"
                      class="installation-page__inline-msg"
                      :message="t('installation.agent.configure.info')"
                    />
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </div>
          </div>

          <!-- Step 4: Launch Scan Agent -->
          <div v-else key="step-4" data-testid="installation-content-launch">
            <div class="installation-page__two-col">
              <div class="installation-page__col">
                <SectionHeader
                  level="h4"
                  icon="pi pi-microsoft"
                  :title="t('installation.platform.windows')"
                />
                <Accordion value="0">
                  <AccordionPanel value="0">
                    <AccordionHeader>{{ t('installation.launch.windows.title') }}</AccordionHeader>
                    <AccordionContent>
                      <CodeSnippet :text="t('installation.launch.windows.command')" kind="#>" />
                      <AppMessage
                        severity="warn"
                        class="installation-page__inline-msg"
                        :message="t('installation.launch.windows.warning')"
                      />
                    </AccordionContent>
                  </AccordionPanel>
                </Accordion>
              </div>
              <div class="installation-page__col">
                <SectionHeader level="h4">
                  <LinuxIcon />
                  {{ t('installation.platform.linux') }}
                </SectionHeader>
                <Accordion value="0">
                  <AccordionPanel value="0">
                    <AccordionHeader>{{ t('installation.launch.linux.title') }}</AccordionHeader>
                    <AccordionContent>
                      <CodeSnippet :text="t('installation.launch.linux.command')" kind="#>" />
                      <AppMessage
                        severity="warn"
                        class="installation-page__inline-msg"
                        :message="t('installation.launch.linux.warning')"
                      />
                    </AccordionContent>
                  </AccordionPanel>
                </Accordion>
              </div>
            </div>
            <AppMessage severity="info" class="installation-page__wide-message">
              <p>{{ t('installation.launch.info.secret') }}</p>
              <p class="installation-page__info-strong">
                {{ t('installation.launch.info.processIntro') }}
              </p>
              <ul>
                <li>{{ t('installation.launch.info.process1') }}</li>
                <li>{{ t('installation.launch.info.process2') }}</li>
                <li>{{ t('installation.launch.info.process3') }}</li>
              </ul>
              <p>{{ t('installation.launch.info.multiple') }}</p>
            </AppMessage>
          </div>
        </Transition>

        <!-- Linear convenience navigation (parity: legacy also allowed jumping
           between steps; Next/Back keeps keyboard + small-screen flow easy). -->
        <div class="installation-page__nav">
          <Button
            type="button"
            severity="secondary"
            icon="pi pi-chevron-left"
            class="installation-page__nav-btn"
            data-testid="installation-prev"
            :disabled="activeStep === 0"
            @click="goToStep(activeStep - 1)"
          />
          <Button
            type="button"
            severity="secondary"
            icon="pi pi-chevron-right"
            class="installation-page__nav-btn"
            data-testid="installation-next"
            :disabled="activeStep === STEPS.length - 1"
            @click="goToStep(activeStep + 1)"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* Add horizontal breathing room to full-width installation notices */
.installation-page__wide-message {
  padding-right: 1.25rem;
  padding-left: 1.25rem;
}

/* Separate setup commands from the surrounding explanatory text */
.installation-page__setup-command {
  margin-top: 0.5rem;
  margin-bottom: 0.75rem;
}

/* Page shell */
.installation-page {
  padding: 1rem 0;
}

/* Let surrounding content control vertical rhythm around unordered lists */
.installation-page ul {
  margin: 0;
}

/* Top warning message: justify body text. */
.installation-page__warning :deep(p) {
  text-align: justify;
}

/* Two-column layout: vertical step sidebar (left) + content panel (right) */
.installation-page__body {
  display: flex;
  align-items: flex-start;
  gap: 1.5rem;
}

/* Fixed-width step sidebar column. Gray left-edge line; each active item
   paints a teal segment over it via box-shadow. */
.installation-page__steps {
  flex: 0 0 13rem;
  border-left: 3px solid var(--p-surface-300);
}

/* Force the PrimeVue <Steps> list vertical (it is horizontal-only by default)
   and remove its horizontal connector line (legacy `ui vertical steps`). */
.installation-page__steps :deep(.p-steps-list) {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.15rem;
}

/* Each step item fills its row and removes the default horizontal separator line */
.installation-page__steps :deep(.p-steps-item) {
  flex: 1 1 auto;
}

.installation-page__steps :deep(.p-steps-item::before) {
  content: none;
}

/* Nav item overrides matching LayoutNavSide: no rounding (so the active shadow
   lands flush on the border-right), no left-border marker. */
.installation-page__steps :deep(.nav-item) {
  border-radius: 0;
  border-left: none;
}

/* Active step: teal left-edge shadow over the gray border-left. */
.installation-page__steps :deep(.nav-item--active) {
  box-shadow: -3px 0 0 0 var(--p-primary-600);
}

/* Right-hand content panel: grows to fill the remaining width with a card border */
.installation-page__panel {
  flex: 1 1 auto;
  min-width: 0;
  border: 1px solid var(--p-content-border-color);
  border-radius: var(--p-content-border-radius);
  padding: 1.25rem;
}

/* "step-reveal" transition: directional slide between installation steps. */
.step-reveal-enter-active,
.step-reveal-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

/* New panel slides in from the right; old panel slides out to the left */
.step-reveal-enter-from {
  opacity: 0;
  transform: translateX(14px);
}

.step-reveal-leave-to {
  opacity: 0;
  transform: translateX(-14px);
}

/* Narrow viewport: stack body vertically, expand the step list horizontally */
@media (max-width: 768px) {
  /* Body collapses from side-by-side to top-and-bottom */
  .installation-page__body {
    flex-direction: column;
  }

  /* Steps span the full width at the top of the stacked layout */
  .installation-page__steps {
    flex: 1 1 auto;
    width: 100%;
  }

  /* Step items wrap horizontally so they act as a tab bar */
  .installation-page__steps :deep(.p-steps-list) {
    flex-direction: row;
    flex-wrap: wrap;
  }
}

/* Remove the default accordion panel border so the steps card border is the only frame */
.installation-page :deep(.p-accordionpanel) {
  border: none;
}

/* Column platform headers use the shared SectionHeader (dividing variant, whose
   underline replaces the former <hr>). Drop its default 2em/1em section margin so
   the header hugs the top of its column and stays compact above the accordion. */
.installation-page__col .section-header {
  margin: 0 0 0.55rem;
}

/* Platform icons in the column headers: slightly below cap-height and baseline-
   aligned so they don't stand taller than the heading text next to them. The
   `!important` on font-size overrides the global `h3 .pi { font-size: inherit
   !important }` base rule (App.vue); without it the Windows .pi glyph would jump
   to full heading size while the Linux SVG (not a .pi) stayed at 0.85em. */
.installation-page__col .section-header :deep(.linux-icon),
.installation-page__col .section-header .pi {
  font-size: 0.85em !important;
  width: 0.85em;
  height: 0.85em;
  vertical-align: -0.05em;
}

/* Separate an inline message from the command or content directly above it. */
.installation-page__inline-msg {
  padding-top: 0.75rem;
}

.installation-page__link-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 0.75rem;
}

.installation-page__info-strong {
  font-weight: 600;
  padding-bottom: 0.25rem;
}

/* Two-column grid for the Windows / Linux download options; position:relative anchors the pseudo-element divider */
.installation-page__two-col {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4rem;
}

/* Vertical "or" divider between the Windows / Linux columns: a centred 1px
   line with an "or" badge on top. */
.installation-page__two-col::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0.5rem;
  left: 50%;
  width: 1px;
  background: var(--p-content-border-color);
  transform: translateX(-50%);
}

/* "or" badge centered on the vertical divider line */
.installation-page__two-col::after {
  position: absolute;
  top: 50%;
  left: 50%;
  padding: 0.1rem 0.45rem;
  border: 1px solid var(--p-content-border-color);
  border-radius: 999px;
  background: var(--p-surface-0);
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
  transform: translate(-50%, -50%);
  content: 'or';
}

@media (max-width: 768px) {
  .installation-page__two-col {
    grid-template-columns: 1fr;
  }

  /* No vertical divider when the columns stack. */
  .installation-page__two-col::before,
  .installation-page__two-col::after {
    content: none;
  }
}

/* Prev / Next navigation row below the content panel */
.installation-page__nav {
  display: flex;
  justify-content: space-between;
  margin-top: 1.25rem;
}

/* Square icon-only buttons using surface tokens; !important overrides PrimeVue Button defaults */
.installation-page__nav-btn {
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  background-color: var(--p-surface-50) !important;
  color: var(--p-text-color) !important;
}

.installation-page__nav-btn:hover {
  background-color: var(--p-surface-100) !important;
}

.installation-page__nav-btn:disabled {
  opacity: 0.4 !important;
}

/* Support matrix table */
.installation-page__table-wrap {
  overflow-x: auto;
}

/* Partial-border table: open top-left corner, top edge on data column headers
   only, left edge on body rows only. border-collapse: separate so border-radius
   can be applied to the corner cells where borders actually meet. */
.installation-page__support-matrix {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

/* Cell padding and bottom row separator */
.installation-page__support-matrix th,
.installation-page__support-matrix td {
  padding: 0.45rem 0.6rem;
  border-bottom: 1px solid var(--p-content-border-color);
  text-align: left;
}

/* Right outer edge on every row */
.installation-page__support-matrix th:last-child,
.installation-page__support-matrix td:last-child {
  border-right: 1px solid var(--p-content-border-color);
}

/* Left outer edge: body rows only (header row intentionally open) */
.installation-page__support-matrix tbody td:first-child {
  border-left: 1px solid var(--p-content-border-color);
}

/* Top outer edge: data column headers only, not the empty corner cell */
.installation-page__support-matrix thead th:not(:first-child) {
  border-top: 1px solid var(--p-content-border-color);
}

/* Column separator between row-label column and data columns */
.installation-page__support-matrix th:first-child,
.installation-page__support-matrix td:first-child {
  border-right: 1px solid var(--p-content-border-color);
}

/* Rounded corners at the four cells where outer edges meet */
.installation-page__support-matrix thead th:last-child {
  border-top-right-radius: var(--p-content-border-radius);
}
.installation-page__support-matrix tr:last-child td:first-child {
  border-bottom-left-radius: var(--p-content-border-radius);
}
.installation-page__support-matrix tr:last-child td:last-child {
  border-bottom-right-radius: var(--p-content-border-radius);
}

/* Header row: light surface tint and medium weight; row-label column matches the tint */
.installation-page__support-matrix thead th {
  background-color: var(--p-surface-50);
  font-weight: 500;
}

.installation-page__support-matrix td:first-child {
  background-color: var(--p-surface-50);
  white-space: nowrap;
}

/* Empty top-left corner cell: transparent so only the data column headers are tinted */
.installation-page__support-matrix thead th:first-child {
  background-color: var(--p-surface-0);
  white-space: nowrap;
}

/* Data column: fixed share of the table width, centered text */
.installation-page__support-matrix-col {
  width: 22%;
  text-align: center !important;
  white-space: nowrap;
}

/* OS icons in column headers: sized to match surrounding text */
.installation-page__support-matrix-os-icon {
  width: 0.85em;
  height: 0.85em;
  vertical-align: -0.05em;
}

/* Data cell: centered support-level icon */
.installation-page__support-matrix-cell {
  text-align: center !important;
}

/* Indented sub-feature row below a main feature */
.installation-page__support-matrix-sub {
  padding-left: 1.5rem;
  font-weight: 400;
  color: var(--p-text-muted-color);
}

/* Rotated arrow glyph prefixing a sub-row (legacy "level up alternate"). */
.installation-page__support-matrix-sub-arrow {
  display: inline-block;
  margin-right: 0.35rem;
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
  transform: rotate(45deg);
}

/* Inline hint icon: small, muted, and shows the help cursor on hover */
.installation-page__support-matrix-hint {
  margin-left: 0.4rem;
  font-size: var(--app-font-size-small);
  color: var(--p-text-muted-color);
  cursor: help;
}

/* Support-level tone colors: full = green (positive), partial = amber, none = red. */
.installation-page__support-matrix-icon--full {
  color: var(--p-green-500);
}

.installation-page__support-matrix-icon--partial {
  color: var(--p-yellow-500);
}

.installation-page__support-matrix-icon--none {
  color: var(--p-red-500);
}
</style>
