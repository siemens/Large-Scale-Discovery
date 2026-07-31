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

import { storeToRefs } from 'pinia'
import Drawer from 'primevue/drawer'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import { LayoutFooter, LayoutNavSide, LayoutNavTop, appShell, useAuthStore } from '@/App'
import AppToast from '@/components/elements/AppToast.vue'
import LoadingBar from '@/components/elements/LoadingBar.vue'
import UnsavedDialog from '@/components/elements/UnsavedDialog.vue'
import LayoutFeedback from '@/components/layout/LayoutFeedback.vue'
import { useIsMobile } from '@/composables/useIsMobile'

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    a11y: {
      openFeedback: 'Open feedback form',
      skipToContent: 'Skip to main content',
    },
    nav: {
      side: {
        menu: 'Menu',
      },
    },
  },
}

// ── Composables and stores ────────────────────────────────────
// Route, i18n, and mobile-breakpoint composable instances
const route = useRoute()
const { t } = useI18n({ inheritLocale: true, messages: localMessages })
const isMobile = useIsMobile()
const authStore = useAuthStore()
const { isAuthenticated, profile } = storeToRefs(authStore)

// ── Reactive state ────────────────────────────────────────────
// Controls the feedback dialog visibility
const feedbackVisible = ref(false)

// At/below the md breakpoint (768px) the section nav collapses into an
// off-canvas drawer driven by the NavTop hamburger. The PrimeVue Drawer is
// modal, providing Esc-to-close and a focus trap natively.
const mobileNavOpen = ref(false)

// ── Computed ──────────────────────────────────────────────────
// Auth pages (login/register) render their own full-bleed teal shell; no top
// nav is shown because the user is, by definition, not signed in yet.
const isAuthRoute = computed(() => route.meta.auth === true)

// Let each application identify the routes that use its section sidebar
const hasSideNav = computed(() => appShell.hasSideNav(route.path))

// ── Watchers and lifecycle ────────────────────────────────────
// Returning to desktop width (or losing the sidebar) must close the drawer so a
// stale overlay never lingers.
watch([isMobile, hasSideNav], ([mobile, sidebar]) => {
  if (!mobile || !sidebar) mobileNavOpen.value = false
})

// Close the feedback form whenever authentication state is cleared
watch(isAuthenticated, (authenticated) => {
  if (!authenticated) feedbackVisible.value = false
})
</script>

<template>
  <div class="app-layout">
    <a class="skip-link" href="#main-content" data-testid="skip-link">
      {{ t('a11y.skipToContent') }}
    </a>

    <AppToast />

    <LayoutNavTop
      v-if="!isAuthRoute"
      :has-sidebar="hasSideNav"
      @open-feedback="feedbackVisible = true"
      @toggle-menu="mobileNavOpen = true"
    />

    <!-- Off-canvas section nav for mobile/tablet (<= md). On desktop the inline
         sidebar below is used instead; only one NavSide is in the DOM at a time
         so data-testid selectors stay unambiguous. -->
    <Drawer
      v-if="hasSideNav && isMobile"
      v-model:visible="mobileNavOpen"
      :header="t('nav.side.menu')"
      position="left"
      class="app-layout__drawer"
      data-testid="nav-side-drawer"
    >
      <LayoutNavSide @navigate="mobileNavOpen = false" />
    </Drawer>

    <main
      id="main-content"
      class="main-content"
      :class="{ 'main-content--auth': isAuthRoute }"
      data-testid="main-content"
    >
      <!-- Subtle fade between routes — parity with the legacy semantic-ui
             page transition. `mode="out-in"` lets the outgoing view fade out
             fully before the incoming one fades in, avoiding overlap jitter.
             Auth routes use a card-only variant so the teal backdrop stays
             fully opaque and doesn't flicker white during the transition. -->
      <RouterView v-slot="{ Component }">
        <Transition :name="isAuthRoute ? 'route-fade-card' : 'route-fade'" mode="out-in">
          <div
            :key="route.path"
            class="page-content"
            :data-testid="route.name ? `page-${String(route.name)}` : undefined"
          >
            <component :is="Component" />
          </div>
        </Transition>
      </RouterView>

      <!-- Section nav sits to the RIGHT of the content on desktop, matching the
           legacy `two wide column right aligned` placement (nav-side rendered
           after the content column). -->
      <LayoutNavSide v-if="hasSideNav && !isMobile" class="main-content__sidebar" />
    </main>

    <LoadingBar v-if="!isAuthRoute" :color="appShell.loadingBarColor" />

    <LayoutFooter v-if="!isAuthRoute" @open-feedback="feedbackVisible = true" />

    <LayoutFeedback
      v-if="isAuthenticated"
      v-model:visible="feedbackVisible"
      :user-display-name="profile?.fullName || profile?.email || ''"
    />
    <UnsavedDialog />

    <button
      v-if="appShell.showFeedbackFab && isAuthenticated && !isAuthRoute"
      type="button"
      class="feedback-fab"
      data-testid="open-feedback"
      :aria-label="t('a11y.openFeedback')"
      @click="feedbackVisible = true"
    >
      <i class="pi pi-address-book" aria-hidden="true" />
    </button>
  </div>
</template>

<style>
/* Global fonts and application-wide rules live with the root application shell. */
@font-face {
  font-family: 'Lato';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url('/fonts/lato-400-normal.woff2') format('woff2');
}

@font-face {
  font-family: 'Lato';
  font-style: italic;
  font-weight: 400;
  font-display: swap;
  src: url('/fonts/lato-400-italic.woff2') format('woff2');
}

@font-face {
  font-family: 'Lato';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/lato-700-normal.woff2') format('woff2');
}

@font-face {
  font-family: 'Lato';
  font-style: italic;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/lato-700-italic.woff2') format('woff2');
}

*,
*::before,
*::after {
  box-sizing: border-box;
}

textarea {
  margin: 0;
}

html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: var(--p-font-family);
  background-color: var(--p-surface-0);
  color: var(--p-text-color);
  font-size: 14px;
}

/* A permanent vertical scrollbar prevents the docked top navigation from shifting. */
body {
  overflow-y: scroll;
}

#app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

::selection {
  background: var(--p-blue-100);
  color: var(--p-text-color);
}

/* Collapse motion to a single frame when the operating system requests it. */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

h1,
h2,
h3,
h4,
h5 {
  font-family: var(--p-font-family);
  line-height: 1.28571429em;
}

h1 {
  font-size: 1.8rem;
  font-weight: 700;
  margin: 2em 0 1rem;
}

h2 {
  font-size: 1.8rem;
  font-weight: 400;
  margin: 1.7em 0 0.85rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--p-content-border-color);
}

h3 {
  font-size: 1.28rem;
  font-weight: 700;
  margin: 0 0 0.7rem;
}

h4 {
  font-size: 1.071rem;
  font-weight: 700;
  margin: 0 0 0.55rem;
}

h5 {
  font-size: 1rem;
  font-weight: 700;
  margin: 0 0 0.4rem;
}

h1:first-child,
h2:first-child,
h3:first-child,
h4:first-child,
h5:first-child {
  margin-top: 0;
}

h1 .pi,
h2 .pi,
h3 .pi,
h4 .pi,
h5 .pi {
  font-size: inherit !important;
}

p {
  margin: 0;
  padding-bottom: 0.5rem;
}

p:last-child {
  padding-bottom: 0;
}

/* PrimeVue has no theme token for the button font size. */
.p-button,
.p-button .p-button-icon {
  font-size: var(--app-font-size-small);
}

:focus-visible {
  outline: 2px solid var(--p-primary-500);
  outline-offset: 2px;
}

.truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.page-section {
  width: 100%;
  min-width: 0;
  padding-top: 3rem;
}

body.presentation-mode .sensitive {
  filter: blur(4px);
}

/* Global PrimeVue overrides for behavior not represented by preset tokens. */
.p-dialog {
  max-width: 95vw;
}

/* Dialogs still close through Escape, the backdrop, or their explicit action. */
.p-dialog-close-button {
  display: none !important;
}

.p-drawer-left,
.p-drawer-right {
  max-width: 85vw;
}

.p-toast {
  max-width: 95vw;
}

.p-tooltip .p-tooltip-text {
  border: 1px solid var(--p-surface-300);
  font-size: var(--app-font-size-small);
}

/* Aura does not expose disabled-border and disabled-shadow theme tokens. */
.p-inputtext:disabled,
.p-textarea:disabled,
.p-disabled > .p-inputtext,
.p-select.p-disabled,
.p-inputnumber.p-disabled .p-inputtext,
.p-datepicker.p-disabled .p-inputtext,
.p-multiselect.p-disabled,
.p-togglebutton.p-disabled {
  border-color: var(--p-surface-200);
  box-shadow: none;
}

/* Dialogs teleport to body, so their shared layout cannot use scoped styles. */
.app-dialog .p-dialog-header {
  padding: 2rem 2rem 1.5rem;
}

.app-dialog .p-dialog-content {
  padding: 0 2rem;
}

.app-dialog .p-dialog-footer {
  padding: 2rem;
}

.app-dialog .app-dialog__header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 1rem 0 0;
  text-align: center;
}

.app-dialog .app-dialog__header-icon {
  font-size: 3rem;
  color: var(--p-text-color);
}

.app-dialog .app-dialog__header-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--p-text-color);
}

/* These sibling selectors shape the buttons around an OrConnector as one control. */
.or-connector + .p-button {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}

.p-button:has(+ .or-connector) {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

/* Shared vertical navigation entry used by section navigation and step lists. */
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.625rem 1rem;
  color: var(--p-text-color);
  text-decoration: none;
  border-radius: var(--p-content-border-radius);
  border-left: 3px solid transparent;
  cursor: pointer;
  transition:
    color 0.15s,
    border-color 0.15s;
}

.nav-item:hover {
  color: var(--p-text-muted-color);
}

.nav-item__icon {
  color: var(--p-text-muted-color);
}

.nav-item--active,
.nav-item.router-link-active,
.nav-item--active .nav-item__label,
.nav-item.router-link-active .nav-item__label {
  font-weight: 600;
}
</style>

<style scoped>
/* ── App shell ──────────────────────────────────────────────── */

/* Reveal the keyboard skip link only while it has focus */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0.5rem;
  z-index: 9999;
  padding: 0.5rem 1rem;
  border-radius: 0 0 4px 4px;
  background: var(--p-primary-500);
  color: var(--p-primary-contrast-color);
  text-decoration: none;
}

.skip-link:focus {
  top: 0;
}

.app-layout {
  display: flex;
  flex-direction: column;
  /* Fill #app (which already carries min-height: 100vh) rather than declaring a
     second 100vh here — a nested 100vh plus any inner margin overflows the
     viewport by those extra pixels. */
  flex: 1;
}

.main-content {
  display: flex;
  flex: 1;
  align-items: flex-start;
  gap: 1.5rem;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto 3rem;
  /* Top gap below the nav bar (legacy `#divPage { margin-top: 20px }`) so neither
     the content nor the section sidebar sits flush against the top navigation. */
  padding: 0 2rem 1.5rem;
}

/* Auth routes: full-bleed, no max-width / padding so the teal shell reaches
   the viewport edges and the card sits flush above the footer. */
.main-content--auth {
  max-width: none;
  margin: 0;
  padding: 0;
  gap: 0;
  align-items: stretch;
}

.main-content--auth .page-content {
  display: flex;
  min-height: 0;
}

.page-content {
  flex: 1;
  min-width: 0;
  /* No fixed min-height (was 70vh, which overflowed once nav + footer + 70vh
     exceeded the viewport). The main landmark is flex:1 in the app column, so it already
     occupies exactly viewport − nav − footer; stretching main to fill the body
     gives the content that leftover height without ever exceeding it — even
     during the route-transition gap the landmark keeps the region from collapsing. */
  align-self: stretch;
}

/* Route fade-in — short (150ms) opacity ease so page changes feel smooth
   rather than popping in instantly. Kept minimal to avoid hurting perceived
   performance. Users with reduced-motion preference get no animation. */
.route-fade-enter-active,
.route-fade-leave-active {
  transition: opacity 0.15s ease;
}

.route-fade-enter-from,
.route-fade-leave-to {
  opacity: 0;
}

/* Auth-page variant: only the card fades, not the full-bleed teal backdrop.
   The host element stays fully opaque so the background never reveals white. */
.route-fade-card-enter-active :deep(.layout-auth__card),
.route-fade-card-leave-active :deep(.layout-auth__card) {
  transition: opacity 0.15s ease;
}

.route-fade-card-enter-from :deep(.layout-auth__card),
.route-fade-card-leave-to :deep(.layout-auth__card) {
  opacity: 0;
}

.main-content__sidebar {
  flex-shrink: 0;
}

/* ── Responsive layout ──────────────────────────────────────── */

/* Fluid container with a 1200px (xl) cap on wide screens; padding steps down on
   narrow viewports so content fits 320px without horizontal page scroll. The
   inline sidebar gap is irrelevant below md (sidebar becomes the drawer). */
@media (max-width: 768px) {
  .main-content {
    gap: 0.75rem;
    padding: 1rem 0.75rem 0.75rem;
  }
}

@media (max-width: 480px) {
  .main-content {
    padding: 1rem 0.5rem 0.5rem;
  }
}
</style>
