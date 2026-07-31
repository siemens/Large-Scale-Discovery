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

/** Collects and submits authenticated user feedback. */

import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import http from '@/api/http'
import DialogForm from '@/components/elements/DialogForm.vue'
import Field from '@/components/elements/Field.vue'
import { useToast } from '@/composables/useToast'
import { useUnsavedChanges } from '@/composables/useUnsavedChanges'
import { useZodForm } from '@/composables/useZodForm'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    title: 'Contact & Feedback',
    fields: {
      subject: {
        label: 'Subject',
        placeholder: 'Regarding a page or a feature',
        validation: {
          required: 'Subject is required.',
        },
      },
      message: {
        label: 'Message',
        placeholder: 'I have an issue...\nI am missing something...',
        validation: {
          required: 'Message is required.',
        },
      },
    },
    actions: {
      submit: {
        success: 'Thank you for your feedback!',
        error: 'Could not submit feedback. Please try again.',
      },
    },
  },
}

/** FeedbackForm is the validated request body accepted by the feedback endpoint. */
interface FeedbackForm {
  subject: string
  message: string
}

// ── Model ─────────────────────────────────────────────────────
const visible = defineModel<boolean>('visible', { default: false })
defineProps<{
  userDisplayName: string
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({
  inheritLocale: true,
  messages: localMessages,
})
const toast = useToast()

// ── Reactive state ────────────────────────────────────────────
// Zod schema (uses t() so it cannot be module-level) plus its inferred type.
const feedbackSchema = z.object({
  subject: z.string().min(1, { message: t('fields.subject.validation.required') }),
  message: z.string().min(1, { message: t('fields.message.validation.required') }),
}) satisfies z.ZodType<FeedbackForm>

const form = ref<FeedbackForm>({ subject: '', message: '' })
const { errors, validateField, validateAll, clearErrors } = useZodForm(feedbackSchema, form)

// ── Computed ──────────────────────────────────────────────────
const isFormDirty = computed(() => form.value.subject.length > 0 || form.value.message.length > 0)

// ── Composable wiring ─────────────────────────────────────────
useUnsavedChanges(() => visible.value && isFormDirty.value)

// ── Functions ─────────────────────────────────────────────────
function resetForm(): void {
  form.value = { subject: '', message: '' }
  clearErrors()
}

function closeDialog(): void {
  resetForm()
  visible.value = false
}

async function submitFeedback(): Promise<void> {
  if (!validateAll()) return

  try {
    await http.post('/user/feedback', form.value, {
      errorToastMessage: t('actions.submit.error'),
    })
    toast.success(t('actions.submit.success'))
    closeDialog()
  } catch {
    // The shared HTTP client already displayed the contextual submission error
  }
}
</script>

<template>
  <DialogForm
    v-model:visible="visible"
    :title="t('title')"
    :text-submit-label="t('common.actions.submit')"
    icon="pi pi-comment"
    width="38rem"
    :dirty="isFormDirty"
    :submit-handler="submitFeedback"
    @cancel="closeDialog"
  >
    <Field v-slot="{ id }" :label="t('common.labels.name')">
      <InputText :id="id" disabled fluid :model-value="userDisplayName" />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('fields.subject.label')"
      :error="errors.subject"
      :validate="() => validateField('subject')"
      required
    >
      <InputText
        :id="id"
        v-model="form.subject"
        data-testid="feedback-subject"
        fluid
        :placeholder="t('fields.subject.placeholder')"
        :invalid="!!errors.subject"
        :aria-label="t('fields.subject.label')"
      />
    </Field>

    <div class="layout-feedback__last-field">
      <Field
        v-slot="{ id }"
        :label="t('fields.message.label')"
        :error="errors.message"
        :validate="() => validateField('message')"
        required
      >
        <Textarea
          :id="id"
          v-model="form.message"
          data-testid="feedback-message"
          fluid
          :placeholder="t('fields.message.placeholder')"
          :invalid="!!errors.message"
          rows="4"
          :aria-label="t('fields.message.label')"
        />
      </Field>
    </div>
  </DialogForm>
</template>

<style scoped>
.layout-feedback__last-field :deep(.field) {
  margin-bottom: 0;
}
</style>
