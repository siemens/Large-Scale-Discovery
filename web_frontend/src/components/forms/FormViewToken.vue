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
 * Access-token form rendered inline above the views table.
 * `views-token-add` component: a description + an expiry slider (1 day to the
 * configured maximum, default 7), then `POST /view/grant/token`. When the backend cannot e-mail the
 * credentials it returns them once — we switch to a credentials result instead
 * of the legacy 20-second auto-closing overlay. Admin/owner-only (gated by the
 * parent page) so legacy permission semantics are preserved exactly.
 */
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import AppMessage from '@/components/elements/AppMessage.vue'
import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import Field from '@/components/elements/Field.vue'
import FieldSlider from '@/components/elements/FieldSlider.vue'
import LayoutSubmitFooter from '@/components/layout/LayoutSubmitFooter.vue'
import { useToast } from '@/composables/useToast'
import { useUnsavedChanges } from '@/composables/useUnsavedChanges'
import { useZodForm } from '@/composables/useZodForm'
import { useSettingsStore } from '@/stores/settings'
import { useViewsStore } from '@/stores/views'

import type { ViewGrantTokenResponse, ViewResponse } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    viewTokenPanel: {
      fields: {
        description: {
          validation: {
            min: 'Description must be at least 5 characters.',
          },
          placeholder: 'Access token for XYZ',
        },
        view: {
          label: 'View Name',
        },
        expiry: {
          unit: 'Days',
        },
      },
      result: {
        hint: 'Please copy the access token, it will disappear shorty.',
      },
      warning:
        'Increased security risk! Access tokens are valid for a prolonged time frame and not bound to a single user. The risk of lost or stolen access tokens is thereby increased. Any risk originating from generated access tokens must be accepted and taken over by you. Please rather grant access to users directly instead!',
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{ view: ViewResponse | null }>()
const emit = defineEmits<{ close: []; created: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const viewsStore = useViewsStore()
const settingsStore = useSettingsStore()

// ── Reactive state ────────────────────────────────────────────
const tokenDescription = ref('')
const tokenExpiryDays = ref(7)

// One-time-reveal credentials kept only in local state for the display lifetime
const generatedCredentials = ref<ViewGrantTokenResponse | null>(null)

const schema = z.object({
  description: z.string().trim().min(5, t('viewTokenPanel.fields.description.validation.min')),
})
const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, () => ({
  description: tokenDescription.value,
}))

// Unsaved-changes guard — dirty before the token is generated, while the
// description/expiry differ from their defaults. Once credentials are shown the
// token already exists, so it's no longer dirty.
useUnsavedChanges(
  () =>
    generatedCredentials.value === null &&
    (tokenDescription.value.trim() !== '' || tokenExpiryDays.value !== 7),
)

// ── Computed ──────────────────────────────────────────────────
// ── Functions ─────────────────────────────────────────────────
async function submitTokenForm(): Promise<void> {
  if (!props.view || !validateAll()) return
  try {
    const { message, credentials: tokenCredentials } = await viewsStore.generateAccessToken(
      props.view.id,
      tokenDescription.value.trim(),
      tokenExpiryDays.value,
    )
    emit('created')

    if (tokenCredentials) {
      generatedCredentials.value = tokenCredentials
    } else {
      toast.success(message)
      emit('close')
    }
  } catch {
    // The request and parsing layers already displayed the appropriate error
  }
}

function closeForm(): void {
  emit('close')
}

// ── Watchers and lifecycle ────────────────────────────────────
// The form mounts fresh on open (the page keys it per view), so initialize on mount
onMounted(() => {
  tokenDescription.value = ''
  tokenExpiryDays.value = 7
  clearErrors()
  generatedCredentials.value = null
})

defineExpose({ cancel: closeForm, submit: submitTokenForm })
</script>

<template>
  <!-- Show credentials only when e-mail delivery fails -->
  <div
    v-if="generatedCredentials"
    class="form-view-token__creds"
    data-testid="view-token-credentials"
  >
    <p class="form-view-token__hint">
      {{ t('viewTokenPanel.result.hint') }}
    </p>
    <Field :label="t('common.labels.username')" tooltip>
      <CodeSnippet
        test-id="view-token-username"
        :text="generatedCredentials.username"
        :text-copy-label="t('common.actions.copy')"
      />
    </Field>
    <Field :label="t('common.labels.password')" tooltip>
      <CodeSnippet
        test-id="view-token-password"
        :text="generatedCredentials.password"
        :text-copy-label="t('common.actions.copy')"
      />
    </Field>
  </div>

  <!-- Token form fields (LayoutForm supplies the surrounding <form>) -->
  <div v-else class="form-view-token__fields">
    <div class="form-view-token__grid form-view-token__grid--thirds">
      <Field v-slot="{ id }" :label="t('common.labels.group')" tooltip>
        <InputText :id="id" :model-value="view?.scan_scope.group_name" disabled />
      </Field>
      <Field v-slot="{ id }" :label="t('common.labels.scope')" tooltip>
        <InputText :id="id" :model-value="view?.scan_scope.name" disabled />
      </Field>
      <Field v-slot="{ id }" :label="t('viewTokenPanel.fields.view.label')" tooltip>
        <InputText :id="id" :model-value="view?.name" disabled />
      </Field>
    </div>

    <!-- Description (wide) beside the expiry slider (legacy description + slider
           row). The current day count is shown inline in the slider's label. -->
    <div class="form-view-token__grid form-view-token__grid--desc">
      <Field
        v-slot="{ id }"
        :label="t('common.labels.description')"
        required
        tooltip
        :error="errors.description"
        :validate="() => validateField('description')"
      >
        <InputText
          :id="id"
          v-model="tokenDescription"
          data-testid="view-token-description"
          class="form-view-token__input"
          :placeholder="t('viewTokenPanel.fields.description.placeholder')"
          :invalid="!!errors.description"
          :aria-label="t('common.labels.description')"
        />
      </Field>

      <FieldSlider
        v-model="tokenExpiryDays"
        :label="t('common.labels.expiry')"
        :min="1"
        :max="settingsStore.tokenExpiryMaxDays"
        tooltip
        show-value
        :unit="t('viewTokenPanel.fields.expiry.unit')"
      />
    </div>

    <!-- Risk warning below the form fields. -->
    <AppMessage severity="warn" class="form-view-token__warning" data-testid="view-token-warning">
      <p>{{ t('viewTokenPanel.warning') }}</p>
    </AppMessage>
  </div>

  <!-- Footer: a single Close after credentials are shown, else Cancel · or · Generate. -->
  <div v-if="generatedCredentials" class="form-view-token__done">
    <Button type="button" :label="t('common.actions.close')" @click="closeForm" />
  </div>
  <LayoutSubmitFooter
    v-else
    submit-test-id="view-token-submit"
    submit-type="submit"
    submit-severity="warn"
    :text-submit-label="t('common.actions.generate')"
    @cancel="closeForm"
  />
</template>

<style scoped>
.form-view-token__done {
  display: flex;
  justify-content: flex-end;
}

.form-view-token__input {
  width: 100%;
}

/* Multi-column rows (legacy layout parity): Group|Scope|View three per row
   (--thirds, minmax(0,…) prevents chip intrinsic widths from distorting columns);
   Description beside the expiry slider (--desc, 1fr/2fr intentionally uneven). */
.form-view-token__grid {
  display: grid;
  gap: 0.75rem;
}

.form-view-token__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-view-token__grid--desc {
  grid-template-columns: 1fr 2fr;
  align-items: start;
}

.form-view-token__grid :deep(.p-inputtext) {
  width: 100%;
}

/* The footer follows outside the fields wrapper, so preserve terminal spacing. */
.form-view-token__warning {
  --app-message-terminal-padding-bottom: 1rem;
}

/* Muted helper text under the credentials heading */
.form-view-token__hint {
  margin: 0 0 1rem;
}
</style>
