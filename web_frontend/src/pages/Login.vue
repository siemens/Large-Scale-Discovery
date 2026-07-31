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
 * Login page. Two-step credentials flow with SSO handoff and optional dev-mode
 * shortcut, rendered inside the shared Auth layout.
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
import SegmentedButton from '@/components/elements/SegmentedButton.vue'
import LayoutAuth from '@/components/layout/LayoutAuth.vue'
import { useFormSubmitLock } from '@/composables/useFormSubmitLock'
import { useToast } from '@/composables/useToast'
import { useZodForm } from '@/composables/useZodForm'
import { clearPendingRedirect, consumePendingRedirect } from '@/lib/redirect'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    auth: {
      login: {
        actions: {
          submit: {
            button: 'Login',
            error: {
              invalidCredentialsGeneric: 'Invalid e-mail or password.',
            },
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
const { isSubmitting, run: runAuthSubmit } = useFormSubmitLock()

const settingsStore = useSettingsStore()
const { isDevelopmentLogin, isCredentialsRegistration } = storeToRefs(settingsStore)

// ── Reactive state ────────────────────────────────────────────
// Two-step login: enter email, then (unless dev mode / SSO) enter the password
type Step = 'email' | 'password'
const step = ref<Step>('email')

// Single schema; each step validates only its own field (validateField)
const schema = z.object({
  email: z
    .string()
    .min(1, { message: t('common.validation.email.required') })
    .email({ message: t('common.validation.email.invalid') }),
  password: z.string().min(8, { message: t('common.validation.password.min') }),
})

const form = ref({ email: '', password: '' })
const { errors, validateField } = useZodForm(schema, form)
const loginFailure = ref<string>()

// ── Computed ──────────────────────────────────────────────────
const showRegister = computed(() => isCredentialsRegistration.value || isDevelopmentLogin.value)

// ── Functions and handlers ────────────────────────────────────
/** Redirect to the originally requested URL (if any) or the user's home. */
async function redirectAfterLogin(): Promise<void> {
  const target = consumePendingRedirect()
  if (target) {
    await router.push(target)
  } else {
    await router.push('/home')
  }
}

async function onLogin(password?: string): Promise<void> {
  loginFailure.value = undefined
  try {
    await authStore.login(form.value.email, password)
    await redirectAfterLogin()
  } catch {
    clearPendingRedirect()
    loginFailure.value =
      authStore.loginError ?? t('auth.login.actions.submit.error.invalidCredentialsGeneric')
    toast.error(loginFailure.value)
  }
}

/** Step 1 — resolve how this email authenticates. */
async function onEmailSubmit(): Promise<void> {
  if (!validateField('email')) return
  await runAuthSubmit(async () => {
    try {
      const entryUrl = await authStore.resolveAuthenticator(form.value.email)
      if (entryUrl) {
        // External SSO entry point — hand off to the browser.
        window.location.href = entryUrl
        return
      }
      if (isDevelopmentLogin.value) {
        // Dev mode: no password required.
        await onLogin(undefined)
        return
      }
      step.value = 'password'
    } catch {
      clearPendingRedirect()
      toast.error(t('auth.login.actions.submit.error.invalidCredentialsGeneric'))
    }
  })
}

async function onPasswordSubmit(): Promise<void> {
  if (!validateField('password')) return
  await runAuthSubmit(() => onLogin(form.value.password))
}

async function onSubmit(): Promise<void> {
  // Dispatch the persistent form to the action represented by its visible step
  if (step.value === 'email') await onEmailSubmit()
  else await onPasswordSubmit()
}

function cancelPassword(): void {
  step.value = 'email'
  form.value.password = ''
  errors.value.password = undefined
  loginFailure.value = undefined
}

function goToRegister(): void {
  void router.push('/register')
}

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(async () => {
  try {
    // Settings are best-effort here; credentials login still works without them
    await settingsStore.fetchSettings('/backend/settings')
  } catch {
    // Settings failed to load; the credentials flow degrades gracefully
  }
})
</script>

<template>
  <LayoutAuth>
    <form data-testid="login-form" novalidate @submit.prevent="onSubmit">
      <Field
        v-show="step === 'email'"
        icon="pi pi-user"
        :validate="() => validateField('email')"
        class="login-page__field"
      >
        <template #default="{ id }">
          <FloatLabel variant="on">
            <InputText
              :id="id"
              v-model="form.email"
              name="email"
              type="email"
              autocomplete="username"
              :invalid="!!errors.email"
              class="login-page__icon-input"
              data-testid="login-email-input"
            />
            <label :for="id">{{ t('common.labels.email') }}</label>
          </FloatLabel>
        </template>
      </Field>

      <SegmentedButton
        v-show="step === 'email'"
        type="submit"
        icon="pi pi-sign-in"
        :label="t('auth.login.actions.submit.button')"
        :disabled="isSubmitting"
        class="login-page__button"
        data-testid="login-submit"
      />

      <template v-if="step === 'password'">
        <Field icon="pi pi-lock" class="login-page__field">
          <template #default="{ id }">
            <FloatLabel variant="on">
              <InputText
                :id="id"
                v-model="form.password"
                name="password"
                type="password"
                autocomplete="current-password"
                :invalid="!!errors.password || !!loginFailure"
                class="login-page__icon-input"
                data-testid="login-password-input"
              />
              <label :for="id">{{ t('common.labels.password') }}</label>
            </FloatLabel>
          </template>
        </Field>

        <div class="login-page__actions" data-testid="login-password-actions">
          <Button
            type="button"
            severity="secondary"
            :label="t('common.actions.cancel')"
            :disabled="isSubmitting"
            class="login-page__button"
            data-testid="login-cancel"
            @click="cancelPassword"
          />
          <OrConnector />
          <Button
            type="submit"
            :label="t('auth.login.actions.submit.button')"
            :disabled="isSubmitting"
            class="login-page__button"
            data-testid="login-password-submit"
          />
        </div>
      </template>
    </form>

    <template v-if="showRegister">
      <SegmentedButton
        type="button"
        severity="secondary"
        icon="pi pi-plus"
        :label="t('common.actions.register')"
        :disabled="isSubmitting"
        class="login-page__button"
        data-testid="login-register-link"
        @click="goToRegister"
      />
    </template>
  </LayoutAuth>
</template>

<style scoped>
.login-page__field {
  padding-top: 3.38rem;
}

/* Stretch icon-prefixed inputs to fill the auth card width */
.login-page__icon-input {
  width: 100%;
}

/* Full-width primary action button (Login, Register) */
.login-page__button {
  width: 100%;
}

/* Submission remains visibly solid while the disabled state prevents a
   duplicate request. Progress is shown by the app-level request bar. */
.login-page__button:disabled {
  opacity: 1;
}

.login-page__actions {
  display: flex;
  margin-top: 1rem;
}

.login-page__actions .login-page__button {
  flex: 1;
  min-width: 0;
  width: auto;
}

@media (max-width: 360px) {
  .login-page__actions {
    flex-direction: column;
  }

  .login-page__actions .login-page__button {
    flex: none;
    width: 100%;
    border-radius: var(--p-content-border-radius);
  }

  .login-page__actions :deep(.or-connector) {
    flex-basis: 0.7rem;
    min-height: 0.7rem;
    margin: 0.2rem 0;
  }
}
</style>
