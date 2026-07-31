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
 * <LayoutAuth> — shared chrome for the auth pages: a full-bleed dark-teal backdrop
 * with a vertically-centered white card and the app title above the slotted form.
 */
</script>

<template>
  <div class="layout-auth__screen">
    <section class="layout-auth__card" data-testid="layout-auth__card">
      <div class="layout-auth__card-body">
        <slot />
      </div>
    </section>
  </div>
</template>

<style scoped>
/* Auth chrome: the login/register screens sit on a dark-teal dimmer (#008080),
 * vertically centered. Brand teal is scoped to a local token here rather than
 * the global Aura primary (which uses Aura's teal palette for the rest of the
 * app). */
.layout-auth__screen {
  display: flex;
  flex: 1;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  background-color: var(--p-primary-700);
}

.layout-auth__card {
  width: 50rem;
  height: 28rem;
  padding: 3.5rem 26rem 3.5rem 5rem;
  border-radius: var(--p-content-border-radius);
  background-color: #114948;
  background-image: url(/img/login.png);
  background-size: contain;
  background-position: center;
  background-repeat: no-repeat;
  box-shadow:
    2px 3px 3px #0003,
    3px 3px 15px 2px #0003;
}

/* App title: app name centered above the slotted form */
.layout-auth__card-title {
  margin: 0 0 1rem;
  text-align: center;
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--p-text-color);
}

/* Stack slotted form controls vertically, centered within the card height */
.layout-auth__card-body {
  display: flex;
  flex-direction: column;
  height: 100%;
  justify-content: space-between;
}

/* Button parity for both auth pages (DRY): centered, bold. Colors, padding and
 * font-size (including the default primary hover) are inherited from the global
 * preset / App.vue overrides — same as every other button in the app. */
.layout-auth__card :deep(.p-button) {
  justify-content: center;
  font-weight: 700;
}
</style>
