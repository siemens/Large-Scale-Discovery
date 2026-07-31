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
 * Database creation form rendered inline above the databases table.
 *
 * Ports the legacy `components/databases/add` form to the A3 validation pattern
 * (zod schema + PrimeVue inputs + <Field>). The save itself is the legacy
 * "connection test": the backend validates the connection details and answers
 * with success / failure, which we surface as a toast.
 *
 * Security: the password lives only in this component's local
 * `form` ref for the lifetime of the form. It is forwarded to the store action
 * (and on to the HTTP layer) but never written to store state, never logged.
 * `resetForm()` clears it whenever the form closes.
 */
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { z } from 'zod'

import Field from '@/components/elements/Field.vue'
import { useFormLifecycle } from '@/composables/useFormLifecycle'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { validateHost, validatePasswordLength, validatePort } from '@/lib/validators'
import { useDatabasesStore } from '@/stores/databases'

import type { InputHTMLAttributes } from 'vue'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    admin: {
      databases: {
        fields: {
          name: {
            validation: {
              min: 'Name must be at least 3 characters.',
            },
            placeholder: 'Database Name',
          },
          dialect: {
            validation: {
              min: 'Dialect must be at least 3 characters.',
            },
            label: 'Dialect',
            placeholder: 'postgres',
          },
          admin: {
            validation: {
              required: 'Admin user is required.',
            },
            label: 'Admin User',
            placeholder: 'postgres',
          },
          password: {
            help: 'Leave empty if you do not want to update it.',
            placeholder: '**********',
          },
          host: {
            label: 'Host',
            placeholder: 'Database Host',
          },
          hostPublic: {
            label: 'Host Public',
            placeholder: 'Database Host Public',
          },
          port: {
            placeholder: 'Database Port',
          },
          args: {
            label: 'Connection Args',
            placeholder: 'sslmode=verify-ca',
          },
        },
        actions: {
          add: {
            success: 'Database saved.',
          },
        },
      },
    },
  },
}

// ── Props and emits ───────────────────────────────────────────
const emit = defineEmits<{ close: []; saved: [] }>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const toast = useToast()

const databasesStore = useDatabasesStore()

// ── Reactive state ────────────────────────────────────────────
const schema = z.object({
  name: z.string().min(3, { message: t('admin.databases.fields.name.validation.min') }),
  dialect: z.string().min(3, { message: t('admin.databases.fields.dialect.validation.min') }),
  admin: z.string().min(1, { message: t('admin.databases.fields.admin.validation.required') }),

  // Legacy parity: password is optional — the backend accepts an empty value
  // (no password set). The strong-password rule only applies when a value is provided.
  password: z.string().superRefine((value, context) => {
    if (value === '') return
    const error = validatePasswordLength(value, 10)
    if (error) context.addIssue({ code: z.ZodIssueCode.custom, message: error })
  }),
  host: z.string().superRefine((value, context) => {
    const error = validateHost(value)
    if (error) context.addIssue({ code: z.ZodIssueCode.custom, message: error })
  }),
  host_public: z.string().superRefine((value, context) => {
    const error = validateHost(value)
    if (error) context.addIssue({ code: z.ZodIssueCode.custom, message: error })
  }),

  // Legacy used a plain text input for the port; validate the numeric range here
  // and coerce to a number on submit.
  port: z.string().superRefine((value, context) => {
    const error = validatePort(value)
    if (error) context.addIssue({ code: z.ZodIssueCode.custom, message: error })
  }),
  args: z.string(),
})

type DatabaseForm = z.infer<typeof schema>

const form = ref<DatabaseForm>(createDefaultDatabaseForm())
const { errors, validateField, validateAll, clearErrors } = useZodForm(schema, form)

const pristineFormSnapshot = JSON.stringify(createDefaultDatabaseForm())

const passwordInputProps = {
  'aria-label': t('common.labels.password'),
  'data-testid': 'db-add-field-password',
  autocomplete: 'new-password',
} satisfies InputHTMLAttributes & { 'data-testid': string }

// ── Functions ─────────────────────────────────────────────────
// Defaults mirror the legacy add form placeholders/initial observables.
function createDefaultDatabaseForm(): DatabaseForm {
  return {
    name: '',
    dialect: 'postgres',
    admin: 'postgres',
    password: '',
    host: '',
    host_public: '',
    port: '5432',
    args: 'sslmode=verify-ca',
  }
}

function resetForm(): void {
  form.value = createDefaultDatabaseForm()
  clearErrors()
  databasesStore.saveError = null
}

// ── Watchers and lifecycle ────────────────────────────────────
// Shared form lifecycle: reset-on-mount, the unsaved-changes guard, cancel(),
// and the submit try/catch skeleton. reset() runs whenever the form opens.
const { submit: submitForm, cancel } = useFormLifecycle({
  reset: resetForm,
  dirty: () => JSON.stringify(form.value) !== pristineFormSnapshot,
  async submit(): Promise<boolean> {
    if (!validateAll()) return false
    const { port, ...rest } = form.value
    await databasesStore.saveDatabase({ ...rest, port: Number(port) })
    toast.success(t('admin.databases.actions.add.success'))
    emit('saved')
    return true
  },
  close: () => emit('close'),
})

defineExpose({ cancel, submit: submitForm })
</script>

<template>
  <div class="form-database-add__grid" data-testid="db-add-form">
    <Field
      v-slot="{ id }"
      :label="t('common.labels.name')"
      required
      :error="errors.name"
      :validate="() => validateField('name')"
    >
      <InputText
        :id="id"
        v-model="form.name"
        data-testid="db-add-field-name"
        :placeholder="t('admin.databases.fields.name.placeholder')"
        :invalid="!!errors.name"
        class="form-database-add__full-width"
        :aria-label="t('common.labels.name')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('admin.databases.fields.dialect.label')"
      required
      :error="errors.dialect"
      :validate="() => validateField('dialect')"
    >
      <InputText
        :id="id"
        v-model="form.dialect"
        data-testid="db-add-field-dialect"
        :placeholder="t('admin.databases.fields.dialect.placeholder')"
        :invalid="!!errors.dialect"
        class="form-database-add__full-width"
        :aria-label="t('admin.databases.fields.dialect.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('admin.databases.fields.admin.label')"
      required
      :error="errors.admin"
      :validate="() => validateField('admin')"
    >
      <InputText
        :id="id"
        v-model="form.admin"
        data-testid="db-add-field-admin"
        :placeholder="t('admin.databases.fields.admin.placeholder')"
        :invalid="!!errors.admin"
        class="form-database-add__full-width"
        :aria-label="t('admin.databases.fields.admin.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('common.labels.password')"
      :help="t('admin.databases.fields.password.help')"
      :error="errors.password"
      :validate="() => validateField('password')"
    >
      <Password
        v-model="form.password"
        :input-id="id"
        :placeholder="t('admin.databases.fields.password.placeholder')"
        :invalid="!!errors.password"
        :feedback="false"
        toggle-mask
        fluid
        input-class="form-database-add__full-width"
        :input-props="passwordInputProps"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('admin.databases.fields.host.label')"
      required
      :error="errors.host"
      :validate="() => validateField('host')"
    >
      <InputText
        :id="id"
        v-model="form.host"
        data-testid="db-add-field-host"
        :placeholder="t('admin.databases.fields.host.placeholder')"
        :invalid="!!errors.host"
        class="form-database-add__full-width"
        :aria-label="t('admin.databases.fields.host.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('admin.databases.fields.hostPublic.label')"
      required
      :error="errors.host_public"
      :validate="() => validateField('host_public')"
    >
      <InputText
        :id="id"
        v-model="form.host_public"
        data-testid="db-add-field-hostPublic"
        :placeholder="t('admin.databases.fields.hostPublic.placeholder')"
        :invalid="!!errors.host_public"
        class="form-database-add__full-width"
        :aria-label="t('admin.databases.fields.hostPublic.label')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('common.labels.port')"
      required
      :error="errors.port"
      :validate="() => validateField('port')"
    >
      <InputText
        :id="id"
        v-model="form.port"
        data-testid="db-add-field-port"
        inputmode="numeric"
        :placeholder="t('admin.databases.fields.port.placeholder')"
        :invalid="!!errors.port"
        class="form-database-add__full-width"
        :aria-label="t('common.labels.port')"
      />
    </Field>

    <Field
      v-slot="{ id }"
      :label="t('admin.databases.fields.args.label')"
      :error="errors.args"
      :validate="() => validateField('args')"
    >
      <InputText
        :id="id"
        v-model="form.args"
        data-testid="db-add-field-args"
        :placeholder="t('admin.databases.fields.args.placeholder')"
        :invalid="!!errors.args"
        class="form-database-add__full-width"
        :aria-label="t('admin.databases.fields.args.label')"
      />
    </Field>
  </div>
</template>

<style scoped>
.form-database-add__grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0 1rem;
}

/* Full-width rule — also passed to the Password component's inner input via
   input-class, so it must be a plain (scoped) selector, not :deep. */
.form-database-add__full-width {
  width: 100%;
}

@media (max-width: 768px) {
  .form-database-add__grid {
    grid-template-columns: 1fr;
  }
}
</style>
