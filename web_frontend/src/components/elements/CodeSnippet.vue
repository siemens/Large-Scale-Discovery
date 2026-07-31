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
 * General-purpose monospace value box with a corner button that copies the value
 * to the clipboard. Used anywhere a value should be shown verbatim and copyable —
 * e.g. shell commands, generated passwords, or access tokens. The corner icon,
 * label, and copied-message are all configurable per use.
 */

import Button from 'primevue/button'
import { useI18n } from 'vue-i18n'

import { copyToClipboard } from '@/composables/useClipboardCopy'
import { useToast } from '@/composables/useToast'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    copyLabel: 'Copy to clipboard',
    copiedMessage: 'Copied to clipboard.',
    copyFailedMessage: 'Could not copy to clipboard.',
  },
}

// ── Props ─────────────────────────────────────────────────────
const props = withDefaults(
  defineProps<{
    /** The text / value shown in the box and copied to the clipboard. */
    text: string
    /** Accessible label and tooltip for the copy action. */
    textCopyLabel?: string
    /** Success message shown after copying. */
    textCopiedMessage?: string
    /** Error message shown when both clipboard strategies fail. */
    textCopyFailedMessage?: string
    /** Corner-button glyph (PrimeIcon class). Defaults to a clipboard icon. */
    copyIcon?: string
    /** Render as a multi-line block (`<pre>`, whitespace preserved) rather than a
     *  single-line value (`<code>`, break-all). Use for code/JSON snippets. */
    multiline?: boolean
    /** Wrap long multiline content or clip it at the snippet boundary. */
    overflow?: 'wrap' | 'hidden'
    /** Center the value text horizontally inside the box. */
    centered?: boolean
    /** Optional kind segment rendered as a gray left-side chip (e.g. `"#>"`). */
    kind?: string
    /** Optional PrimeIcon class rendered as an icon inside the kind chip instead of text. */
    kindIcon?: string
    /** CSS font-family stack used by the snippet content. */
    fontFamily?: string
    /** CSS font size used by the snippet content. */
    fontSize?: string
    /** CSS background color used by the snippet content. */
    backgroundColor?: string
    testId?: string
  }>(),
  {
    copyIcon: 'pi pi-clipboard',
    overflow: 'wrap',
    fontFamily: "'Courier New', serif",
    fontSize: 'var(--app-font-size-small)',
  },
)

// ── Slots ─────────────────────────────────────────────────────
const slots = defineSlots<{
  /** Optional caller-owned semantic content displayed beneath the snippet. */
  footer?: () => unknown
}>()

// ── Composables ───────────────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()

// ── Functions ─────────────────────────────────────────────────
// Report the real clipboard outcome instead of acknowledging a failed copy
async function onCopy(): Promise<void> {
  const copied = await copyToClipboard(props.text)
  if (copied) {
    toast.brief(props.textCopiedMessage ?? t('copiedMessage'))
  } else {
    toast.error(props.textCopyFailedMessage ?? t('copyFailedMessage'))
  }
}
</script>

<template>
  <div
    class="code-snippet"
    :data-testid="testId ? `code-snippet-${testId}` : undefined"
    :class="{
      'code-snippet--multiline': multiline,
      'code-snippet--overflow-hidden': overflow === 'hidden',
      'code-snippet--centered': centered,
    }"
  >
    <div class="code-snippet__body" :style="{ backgroundColor }">
      <span v-if="kind || kindIcon" class="code-snippet__kind" aria-hidden="true">
        <i v-if="kindIcon" :class="kindIcon" />
        <template v-else>{{ kind }}</template>
      </span>
      <pre
        v-if="multiline"
        class="code-snippet__text code-snippet__text--pre"
        :style="{ fontFamily, fontSize }"
        >{{ text }}</pre>
      <code v-else class="code-snippet__text" :style="{ fontFamily, fontSize }">
        <span class="code-snippet__text-inner">{{ text }}</span>
      </code>
    </div>
    <div v-if="slots.footer" class="code-snippet__footer">
      <slot name="footer" />
    </div>
    <!-- Legacy parity: teal "right corner label" badge with a clipboard glyph
         anchored in the top-right corner of the code block. -->
    <Button
      v-tooltip.top="textCopyLabel ?? t('copyLabel')"
      :data-testid="testId ? `code-snippet-copy-${testId}` : undefined"
      type="button"
      :icon="copyIcon"
      class="code-snippet__corner"
      :aria-label="textCopyLabel ?? t('copyLabel')"
      @click="onCopy"
    />
  </div>
</template>

<style scoped>
.code-snippet {
  --code-snippet-frame-shadow: inset 0 0 0 1px var(--p-content-border-color);

  position: relative;
  min-width: 0;
  /* Draw the 1px frame as an inset shadow rather than a layout border, so the
     padding edge (the containing block for the absolutely-positioned corner
     badge) coincides with the visual edge. A real border pushes `right:0`/`top:0`
     to the border's INNER edge, leaving a 1px content-border sliver between the
     teal badge and the box edge — a seam that becomes visible whenever the box
     lands on a fractional device-pixel width (e.g. the narrow password dialog). */
  box-shadow: var(--code-snippet-frame-shadow);
  border-radius: var(--p-content-border-radius);
  overflow: hidden;
  background-color: var(--p-surface-50);
}

/* Keep the shell frame above opaque caller-supplied content backgrounds. */
.code-snippet::after {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  box-shadow: var(--code-snippet-frame-shadow);
  content: '';
  pointer-events: none;
}

.code-snippet__body {
  display: flex;
  align-items: stretch;
  min-width: 0;
  /* Reserve room on the right so the corner badge never overlaps the command */
  padding: 0 2.4rem 0 0;
}

.code-snippet__kind {
  display: flex;
  align-items: center;
  padding: 0.55rem 0.6rem;
  background-color: var(--p-surface-50);
  color: var(--p-text-muted-color);
  border: 1px solid var(--p-content-border-color);
  border-right: none;
  white-space: nowrap;
  user-select: none;
  font-weight: 700;
}

.code-snippet__kind .pi {
  font-size: inherit;
  line-height: 1;
  vertical-align: middle;
}

.code-snippet__text {
  flex: 1;
  display: grid;
  align-items: center;
  min-width: 0;
  padding: 0.55rem 0.75rem;
}

.code-snippet__text-inner {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.code-snippet--centered .code-snippet__text {
  text-align: center;
}

/* Multi-line variant: top-align the value + corner badge and the kind chip. */
.code-snippet--multiline .code-snippet__body {
  align-items: flex-start;
}

/* Preserve the snippet's own whitespace/newlines while still wrapping long
   lines, and strip the <pre> default margin so the box padding stays even. */
.code-snippet__text--pre {
  margin: 0;
  width: 100%;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
}

/* Preserve authored line breaks without wrapping or exposing horizontal overflow */
.code-snippet--overflow-hidden .code-snippet__text--pre {
  overflow-x: hidden;
  white-space: pre;
  word-break: normal;
}

.code-snippet__footer {
  display: flex;
  align-items: center;
  min-width: 0;
  padding: 0.4rem 0.75rem;
  border-top: 1px solid var(--p-content-border-color);
  color: var(--p-text-muted-color);
  font-size: var(--app-font-size-tiny);
}

/* Teal corner badge (legacy "ui mini teal right corner label").
   !important overrides the PrimeVue Button preset's size, padding, border, and
   background tokens — all of which cannot be targeted from a scoped rule here. */
.code-snippet__corner {
  position: absolute;
  top: 0;
  right: 0;
  width: 2.4rem !important;
  height: 2.4rem !important;
  padding: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: var(--p-primary-500) !important;
  color: var(--p-primary-contrast-color) !important;
  /* Right-triangle filling the top-right corner. */
  clip-path: polygon(100% 0, 0 0, 100% 100%);
}

.code-snippet__corner:hover {
  background: var(--p-primary-600) !important;
}

/* Nudge the glyph into the visible part of the triangle */
.code-snippet__corner :deep(.p-button-icon) {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  font-size: var(--app-font-size-small);
}
</style>
