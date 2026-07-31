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
 * <LayoutNavTop> — top navigation bar with the brand hero, primary links, and
 * password/logout actions. Sticks to the viewport once the hero has scrolled away.
 */
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import http from '@/api/http'
import CodeSnippet from '@/components/elements/CodeSnippet.vue'
import DialogConfirm from '@/components/elements/DialogConfirm.vue'
import { usePresentationMode } from '@/composables/usePresentationMode'
import { TOAST_LIFE, useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { isSafeFrontendHref } from '@/types/settings'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    applicationName: 'Large-Scale Discovery',
    nav: {
      top: {
        actions: {
          generateDatabasePassword: {
            success: 'Password has been sent to your e-mail address.',
            error: 'Could not generate database password.',
            tooltip: 'Generate Database Password',
            title: 'Reset Database Password',
            question: 'Are you sure you want to generate a new database password?',
            message:
              'A temporary database password will be generated. Active sessions are not affected and can be continued.',
          },
        },
        openMenu: 'Open section menu',
        home: 'Home',
        configuration: 'Configuration',
        admin: 'Administration',
        togglePresentation: 'Toggle presentation mode',
        installation: 'Scan Agent',
        logout: 'Logout',
        navigationFailed: 'Could not open the login page.',
      },
      demo: {
        notice:
          'Unfortunately, your {unit} is {status}. You can only use this web interface to get an impression about what you are missing out on.',
        unitLabel: 'business unit',
        statusLabel: 'not a signed-up customer yet',
      },
    },
    a11y: {
      mainNav: 'Main navigation',
    },
    dialog: {
      generateDatabasePassword: {
        title: 'Reset Database Password',
        hint: 'Please copy the database password; it will disappear shortly.',
      },
    },
  },
}

// Scroll distance at which the bar pin triggers (equals the bar's own height)
const BAR_HEIGHT_PX = 40

// ── Props and emits ──────────────────────────────────────────
defineProps<{
  // When true the route has a section sidebar, so the mobile hamburger toggle
  // is rendered (it is additionally hidden above the `md` breakpoint via CSS).
  hasSidebar?: boolean
}>()

const emit = defineEmits<{
  openFeedback: []
  toggleMenu: []
}>()

// ── Composables and stores ────────────────────────────────────
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()
const { isAuthenticated, isAdmin, isOwner, isDemo } = storeToRefs(authStore)
const settingsStore = useSettingsStore()
const { frontendLinks } = storeToRefs(settingsStore)

// Presentation mode toggle masks sensitive values via a body class
const presentationMode = usePresentationMode()
const presentationActive = presentationMode.isPresentationMode

// ── Reactive state ────────────────────────────────────────────
const showPasswordConfirm = ref(false)
const generatedPassword = ref<string | null>(null)

const heroRef = ref<HTMLElement | null>(null)
const navScrolled = ref(false)

// ── Computed ──────────────────────────────────────────────────
const menuLinks = computed(() =>
  (frontendLinks.value.menu ?? []).filter((link) => isSafeFrontendHref(link.href)),
)

// ── Functions ─────────────────────────────────────────────────
function isRouteSectionActive(currentPath: string, targetPath: string): boolean {
  // Top navigation groups every descendant beneath the first path segment
  const currentSection = currentPath.split('/').find(Boolean)
  const targetSection = targetPath.split('/').find(Boolean)

  // The root route is a section only when both paths point to the root
  if (targetSection === undefined) return currentSection === undefined
  return currentSection === targetSection
}

function onTogglePresentation(): void {
  presentationMode.toggle()
}

async function generateDatabasePassword(): Promise<void> {
  try {
    const res = await http.post<{
      body: { password: string }
      message?: string
    }>('/user/password', undefined, {
      errorToastMessage: t('nav.top.actions.generateDatabasePassword.error'),
    })
    const password = res.data.body?.password ?? ''
    if (password === '') {
      toast.add({
        severity: 'success',
        summary: res.data.message ?? t('nav.top.actions.generateDatabasePassword.success'),
        life: TOAST_LIFE.success,
      })
    } else {
      generatedPassword.value = password
    }
  } catch {
    // The shared HTTP client already displayed the contextual generation error
  }
}

function clearGeneratedPassword(): void {
  generatedPassword.value = null
}

function onScroll(): void {
  if (!heroRef.value) return
  const heroBottom = heroRef.value.getBoundingClientRect().bottom
  navScrolled.value = heroBottom <= BAR_HEIGHT_PX
}

async function onLogout(): Promise<void> {
  authStore.logout()

  try {
    await router.push('/login')
  } catch {
    toast.error(t('nav.top.navigationFailed'))
  }
}

// ── Watchers and lifecycle ────────────────────────────────────
onMounted(() => {
  // Browser scroll restoration can mount the shell below the hero
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
})

onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>

<template>
  <header class="nav-top" data-testid="nav-top">
    <button v-if="isDemo" type="button" class="nav-top__demo-banner" @click="emit('openFeedback')">
      <div class="nav-top__demo-banner-inner">
        <i class="pi pi-address-book" aria-hidden="true"></i>
        {{
          t('nav.demo.notice', {
            unit: t('nav.demo.unitLabel'),
            status: t('nav.demo.statusLabel'),
          })
        }}
      </div>
    </button>

    <div ref="heroRef" class="nav-top__hero" role="img" :aria-label="t('applicationName')" />

    <nav
      class="nav-top__bar"
      data-testid="nav-top-bar"
      :class="{ 'nav-top__bar--scrolled': navScrolled }"
      :aria-label="t('a11y.mainNav')"
    >
      <div class="nav-top__bar-inner">
        <template v-if="isAuthenticated">
          <button
            v-if="hasSidebar"
            type="button"
            data-testid="nav-menu-toggle"
            class="nav-top__link nav-top__action nav-top__menu-toggle"
            :aria-label="t('nav.top.openMenu')"
            @click="emit('toggleMenu')"
          >
            <i class="pi pi-bars" aria-hidden="true" />
          </button>

          <router-link
            to="/home"
            class="nav-top__link"
            :class="{
              'router-link-active': isRouteSectionActive(route.path, '/home'),
            }"
            :aria-current="isRouteSectionActive(route.path, '/home') ? 'page' : undefined"
            data-testid="nav-link-home"
          >
            <i class="pi pi-home" aria-hidden="true"></i>
            {{ t('nav.top.home') }}
          </router-link>

          <router-link
            v-if="isAdmin || isOwner"
            to="/configuration/scopes"
            class="nav-top__link"
            :class="{
              'router-link-active': isRouteSectionActive(route.path, '/configuration/scopes'),
            }"
            :aria-current="
              isRouteSectionActive(route.path, '/configuration/scopes') ? 'page' : undefined
            "
            data-testid="nav-link-configuration"
          >
            <i class="pi pi-sliders-h" aria-hidden="true"></i>
            {{ t('nav.top.configuration') }}
          </router-link>

          <a
            v-for="link in menuLinks"
            :key="link.href"
            :href="link.href"
            :target="link.target"
            class="nav-top__link"
            rel="noopener noreferrer"
          >
            <i class="pi pi-external-link" aria-hidden="true"></i>
            <span>{{ link.text }}</span>
          </a>

          <router-link
            v-if="isAdmin"
            to="/admin/users"
            class="nav-top__link"
            :class="{
              'router-link-active': isRouteSectionActive(route.path, '/admin/users'),
            }"
            :aria-current="isRouteSectionActive(route.path, '/admin/users') ? 'page' : undefined"
            data-testid="nav-link-admin"
          >
            <i class="pi pi-user" aria-hidden="true"></i>
            {{ t('nav.top.admin') }}
          </router-link>

          <div class="nav-top__spacer" aria-hidden="true"></div>

          <button
            type="button"
            class="nav-top__link nav-top__action"
            :aria-label="t('nav.top.togglePresentation')"
            :aria-pressed="presentationActive"
            @click="onTogglePresentation"
          >
            <i :class="presentationActive ? 'pi pi-eye-slash' : 'pi pi-eye'" aria-hidden="true" />
          </button>

          <router-link
            v-if="isAdmin || isOwner"
            to="/api"
            class="nav-top__link"
            :class="{
              'router-link-active': isRouteSectionActive(route.path, '/api'),
            }"
            :aria-current="isRouteSectionActive(route.path, '/api') ? 'page' : undefined"
            data-testid="nav-link-api"
          >
            <i class="pi pi-code" aria-hidden="true"></i>
            {{ t('common.labels.api') }}
          </router-link>

          <router-link
            v-if="isAdmin || isOwner"
            to="/installation"
            class="nav-top__link"
            :class="{
              'router-link-active': isRouteSectionActive(route.path, '/installation'),
            }"
            :aria-current="isRouteSectionActive(route.path, '/installation') ? 'page' : undefined"
            data-testid="nav-link-installation"
          >
            <i class="pi pi-download" aria-hidden="true"></i>
            {{ t('nav.top.installation') }}
          </router-link>

          <button
            type="button"
            class="nav-top__link nav-top__action"
            data-testid="nav-link-logout"
            :aria-label="t('nav.top.logout')"
            @click="onLogout"
          >
            <i class="pi pi-sign-out" aria-hidden="true" />
            {{ t('nav.top.logout') }}
          </button>

          <button
            v-if="!isDemo"
            type="button"
            class="nav-top__link nav-top__action nav-top__link--danger"
            data-testid="nav-action-generate-database-password"
            :aria-label="t('nav.top.actions.generateDatabasePassword.tooltip')"
            @click="showPasswordConfirm = true"
          >
            <i class="pi pi-key" aria-hidden="true" />
          </button>
        </template>
      </div>
    </nav>

    <DialogConfirm
      v-if="!isDemo"
      v-model:visible="showPasswordConfirm"
      :title="t('nav.top.actions.generateDatabasePassword.title')"
      icon="pi pi-key"
      :text-confirm-label="t('common.actions.generate')"
      :message="t('nav.top.actions.generateDatabasePassword.question')"
      :secondary-message="t('nav.top.actions.generateDatabasePassword.message')"
      @confirm="generateDatabasePassword"
    />

    <Dialog
      v-if="!isDemo"
      :visible="generatedPassword !== null"
      modal
      dismissable-mask
      closable
      :draggable="false"
      class="app-dialog nav-top__password-dialog"
      @update:visible="
        (v: boolean) => {
          if (!v) clearGeneratedPassword()
        }
      "
    >
      <template #header>
        <div class="app-dialog__header">
          <i class="pi pi-database app-dialog__header-icon" aria-hidden="true" />
          <span class="app-dialog__header-title">{{
            t('dialog.generateDatabasePassword.title')
          }}</span>
        </div>
      </template>
      <div class="nav-top__dialog-body">
        <p>{{ t('dialog.generateDatabasePassword.hint') }}</p>
        <CodeSnippet
          v-if="generatedPassword"
          :text="generatedPassword"
          :text-copy-label="t('common.actions.copy')"
          centered
        />
      </div>
      <template #footer>
        <Button
          type="button"
          severity="secondary"
          :label="t('common.actions.close')"
          @click="clearGeneratedPassword"
        />
      </template>
    </Dialog>
  </header>
</template>

<style scoped>
/* Match shared confirmation dialogs while remaining responsive */
.nav-top__password-dialog {
  width: 38rem;
  max-width: calc(100vw - 2rem);
}

.nav-top {
  width: 100%;
}

/* Demo-mode notice banner: clicking it opens the feedback dialog */
.nav-top__demo-banner {
  display: block;
  width: 100%;
  padding: 0.7rem 1rem;
  border: 0;
  background: var(--p-red-600);
  color: var(--p-surface-0);
  font: inherit;
  text-align: center;
  cursor: pointer;
}

.nav-top__demo-banner .pi {
  margin-right: 0.5rem;
}

.nav-top__demo-banner-inner {
  max-width: 1200px;
  margin: 0 auto;
}

/* Brand image strip: responsive height, teal fallback, wordmark variant */
.nav-top__hero {
  height: min(15vw, 160px);
  /* Legacy teal fallback for when the image hasn't loaded. */
  background-color: var(--p-primary-600);
  /* header2.png has the "Large Scale Discovery" wordmark baked in, matching
     the legacy look exactly. header.png is the text-free variant (dropped). */
  background-image: url('/img/header2.png');
  background-position: center bottom;
  background-size: cover;
}

.nav-top__bar {
  z-index: 10;
  margin-top: -40px;
  background-color: transparent;
}

.nav-top__bar--scrolled {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  margin-top: 0;
  background-color: var(--p-primary-600);
}

.nav-top__bar-inner {
  display: flex;
  align-items: stretch;
  height: 40px;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.nav-top__link {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.625rem 1rem;
  color: var(--app-color-offwhite);
  text-decoration: none;
  border-bottom: 2px solid transparent;
  transition:
    background-color 0.15s,
    border-color 0.15s;
}

/* Native buttons share the same visual treatment as navigation links */
.nav-top__action {
  border: 0;
  background: transparent;
  font: inherit;
  cursor: pointer;
}

.nav-top__link:hover {
  background-color: rgba(255, 255, 255, 0.15);
}

/* Active link: solid white tile with teal text (pointing-menu active item),
   not the 10%-white + underline treatment. Negative bottom margin lets the tile
   overlap the 2px border so no gap appears between bar and tile. */
.nav-top__link.router-link-active {
  background-color: var(--p-surface-0);
  color: var(--p-primary-600);
  border-bottom-color: var(--p-surface-0);
  margin-bottom: -2px;
  padding-bottom: calc(0.5rem + 2px);
}

.nav-top__spacer {
  flex: 1;
}

/* Danger action (generate password): red tile using Aura red severity tokens */
.nav-top__link--danger {
  background-color: var(--p-red-600);
  color: var(--p-red-50);
}

.nav-top__link--danger:hover {
  background-color: var(--p-red-700);
}

/* Hamburger: only visible at/below the md (768px) breakpoint, where NavSide
   collapses into the off-canvas drawer (see docs/responsive.md). */
.nav-top__menu-toggle {
  display: none;
}

@media (max-width: 768px) {
  .nav-top__menu-toggle {
    display: flex;
  }

  /* The top nav has more links than fit a phone width; let them wrap instead of
     pushing horizontal page scroll. The flex spacer is meaningless once wrapped,
     so collapse it and pack links to the left. */
  .nav-top__bar-inner {
    flex-wrap: wrap;
  }

  /* Flex spacer is meaningless once links wrap; collapse it to pack left */
  .nav-top__spacer {
    display: none;
  }

  /* Tighter horizontal padding on small screens */
  .nav-top__link {
    padding: 0.5rem 0.75rem;
  }
}

.nav-top__dialog-body {
  text-align: center;
}
</style>
