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
 * Register page. Collects account details and submits them through the auth
 * store, rendered inside the shared Auth layout.
 */

import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import FloatLabel from 'primevue/floatlabel'
import InputText from 'primevue/inputtext'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { z } from 'zod'

import Field from '@/components/elements/Field.vue'
import OrConnector from '@/components/elements/OrConnector.vue'
import LayoutAuth from '@/components/layout/LayoutAuth.vue'
import { useFormSubmitLock } from '@/composables/useFormSubmitLock'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    auth: {
      register: {
        fields: {
          name: {
            validation: {
              required: 'Name is required.',
            },
          },
          surname: {
            validation: {
              required: 'Surname is required.',
            },
          },
        },
        actions: {
          submit: {
            success: 'Registration successful. You can now log in.',
            error: 'Registration failed. Please check your details and try again.',
          },
        },
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const router = useRouter()
const toast = useToast()

const authStore = useAuthStore()
const { isSubmitting, run: runRegistrationSubmit } = useFormSubmitLock()

const settingsStore = useSettingsStore()
const { isDevelopmentLogin } = storeToRefs(settingsStore)

// ── Reactive state ────────────────────────────────────────────
type RegisterForm = {
  email: string
  password: string
  name: string
  surname: string
}

const form = ref<RegisterForm>({
  email: '',
  password: '',
  name: '',
  surname: '',
})

// ── Computed ──────────────────────────────────────────────────
// Password is only required outside dev mode (parity with legacy register.js),
// so the schema is computed against the live dev-mode flag
const schema = computed(() =>
  z.object({
    email: z
      .string()
      .min(1, { message: t('common.validation.email.required') })
      .email({ message: t('common.validation.email.invalid') }),
    password: isDevelopmentLogin.value
      ? z.string()
      : z.string().min(8, { message: t('common.validation.password.min') }),
    name: z.string().min(1, { message: t('auth.register.fields.name.validation.required') }),
    surname: z.string().min(1, {
      message: t('auth.register.fields.surname.validation.required'),
    }),
  }),
)

const { errors, validateField, validateAll } = useZodForm(schema, form)

// ── Functions and handlers ────────────────────────────────────
/** Validate all fields, submit the registration, and route back to login on success. */
async function submitRegistration(): Promise<void> {
  if (!validateAll()) return
  await runRegistrationSubmit(async () => {
    try {
      // Backend may return a status message we prefer over the generic success text
      const message = await authStore.register(form.value)
      toast.success(message || t('auth.register.actions.submit.success'))
      await router.push('/login')
    } catch {
      errors.value.email = t('auth.register.actions.submit.error')
      // The request and parsing layers already displayed the appropriate error
    }
  })
}

function goToLogin(): void {
  void router.push('/login')
}

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(async () => {
  try {
    // Settings are best-effort; the password stays required if they cannot load
    await settingsStore.fetchSettings('/backend/settings')
  } catch {
    // Settings failed to load; registration proceeds with the strict schema
  }
})
</script>

<template>
  <LayoutAuth>
    <!-- No required-asterisks on labels. Fields are still validated on submit. -->
    <form
      class="register-page__form"
      data-testid="register-form"
      novalidate
      @submit.prevent="submitRegistration"
    >
      <Field icon="pi pi-envelope" :validate="() => validateField('email')">
        <template #default="{ id }">
          <FloatLabel variant="on">
            <InputText
              :id="id"
              v-model="form.email"
              name="email"
              type="email"
              autocomplete="email"
              :invalid="!!errors.email"
              class="register-page__icon-input"
              data-testid="register-email-input"
            />
            <label :for="id">{{ t('common.labels.email') }}</label>
          </FloatLabel>
        </template>
      </Field>

      <Field icon="pi pi-lock" :validate="() => validateField('password')">
        <template #default="{ id }">
          <FloatLabel variant="on">
            <InputText
              :id="id"
              v-model="form.password"
              name="password"
              type="password"
              autocomplete="new-password"
              :invalid="!!errors.password"
              class="register-page__icon-input"
              data-testid="register-password-input"
            />
            <label :for="id">{{ t('common.labels.password') }}</label>
          </FloatLabel>
        </template>
      </Field>

      <Field icon="pi pi-user" :validate="() => validateField('name')">
        <template #default="{ id }">
          <FloatLabel variant="on">
            <InputText
              :id="id"
              v-model="form.name"
              name="name"
              type="text"
              autocomplete="given-name"
              :invalid="!!errors.name"
              class="register-page__icon-input"
              data-testid="register-name-input"
            />
            <label :for="id">{{ t('common.labels.name') }}</label>
          </FloatLabel>
        </template>
      </Field>

      <Field icon="pi pi-users" :validate="() => validateField('surname')">
        <template #default="{ id }">
          <FloatLabel variant="on">
            <InputText
              :id="id"
              v-model="form.surname"
              name="surname"
              type="text"
              autocomplete="family-name"
              :invalid="!!errors.surname"
              class="register-page__icon-input"
              data-testid="register-surname-input"
            />
            <label :for="id">{{ t('common.labels.surname') }}</label>
          </FloatLabel>
        </template>
      </Field>

      <div class="register-page__actions" data-testid="register-actions">
        <Button
          type="button"
          severity="secondary"
          :label="t('common.actions.cancel')"
          :disabled="isSubmitting"
          class="register-page__button"
          data-testid="register-cancel"
          @click="goToLogin"
        />
        <OrConnector />
        <Button
          type="submit"
          :label="t('common.actions.register')"
          :disabled="isSubmitting"
          class="register-page__button"
          data-testid="register-submit"
        />
      </div>
    </form>
  </LayoutAuth>
</template>

<style scoped>
.register-page__form {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* Stretch inputs to fill the field so they align with the full-width buttons. */
.register-page__icon-input {
  width: 100%;
}

.register-page__button {
  width: 100%;
  font-weight: 700;
}

/* Submission remains visibly solid while the disabled state prevents a
   duplicate request. Progress is shown by the app-level request bar. */
.register-page__button:disabled {
  opacity: 1;
}

.register-page__actions {
  display: flex;
  margin-top: auto;
}

/* The registration actions sit over the card artwork, where the connector's
 * default side gap exposes an uneven strip of the background. */
.register-page__actions :deep(.or-connector) {
  margin-inline: 0;
}

.register-page__actions .register-page__button {
  flex: 1;
  min-width: 0;
  width: auto;
}

@media (max-width: 360px) {
  .register-page__actions {
    flex-direction: column;
  }

  .register-page__actions .register-page__button {
    flex: none;
    width: 100%;
    border-radius: var(--p-content-border-radius);
  }

  .register-page__actions :deep(.or-connector) {
    flex-basis: 0.7rem;
    min-height: 0.7rem;
    margin: 0.2rem 0;
  }
}
</style>
