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
 * View editing form rendered inline above the views table.
 * group + scope are read-only, only the view name is editable
 * (`POST /view/update`).
 */
import InputText from 'primevue/inputtext'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import Field from '@/components/elements/Field.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { useViewsStore } from '@/stores/views'

import type { ViewResponse } from '@/types/views'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    viewEditPanel: {
      fields: {
        name: {
          validation: {
            min: 'View name must be at least 3 characters.',
          },
          label: 'View Name',
          placeholder: 'Scope Name',
        },
      },
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const props = defineProps<{ view: ViewResponse | null }>()
const emit = defineEmits<{ close: []; saved: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()
const viewsStore = useViewsStore()

// ── Reactive state ────────────────────────────────────────────
const viewName = ref('')

// Only the name is editable, so validation covers just that field
const schema = z.object({
  name: z.string().trim().min(3, t('viewEditPanel.fields.name.validation.min')),
})
const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, () => ({
  name: viewName.value,
}))

// ── Watchers and lifecycle ────────────────────────────────────
const { submit: submitForm, cancel } = useFormLifecycle({
  target: () => props.view,
  reset(): void {
    viewName.value = props.view?.name ?? ''
    clearErrors()
  },
  dirty: () => viewName.value.trim() !== (props.view?.name ?? ''),
  async submit(): Promise<boolean> {
    if (!props.view || !validateAll()) return false
    const message = await viewsStore.updateView(props.view.id, viewName.value.trim())
    toast.success(message)
    emit('saved')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-view-edit__grid form-view-edit__grid--thirds">
    <Field v-slot="{ id }" :label="t('common.labels.group')" tooltip>
      <InputText :id="id" :model-value="view?.scan_scope.group_name" disabled />
    </Field>
    <Field v-slot="{ id }" :label="t('common.labels.scope')" tooltip>
      <InputText :id="id" :model-value="view?.scan_scope.name" disabled />
    </Field>
    <Field
      v-slot="{ id }"
      :label="t('viewEditPanel.fields.name.label')"
      required
      tooltip
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="viewName"
        data-testid="view-edit-name"
        :placeholder="t('viewEditPanel.fields.name.placeholder')"
        :invalid="!!errors.name"
        :aria-label="t('viewEditPanel.fields.name.label')"
      />
    </Field>
  </div>
</template>

<style scoped>
.form-view-edit__grid {
  display: grid;
  gap: 0.75rem;
}

.form-view-edit__grid--thirds {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-view-edit__grid :deep(.p-inputtext) {
  width: 100%;
}
</style>
