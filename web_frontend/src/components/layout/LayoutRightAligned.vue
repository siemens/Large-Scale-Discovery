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
 * <LayoutRightAligned> — right-aligned row for table or list action controls, such as
 * add buttons. The alignment and item spacing are defined here so call sites stay clean.
 */
</script>

<template>
  <div class="layout-right-aligned">
    <slot />
  </div>
</template>

<style scoped>
.layout-right-aligned {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
