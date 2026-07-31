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

/** Sticky bottom progress bar that animates while any HTTP request is in flight. */
import { isLoading } from '@/api/http'

withDefaults(defineProps<{ color?: string }>(), { color: 'var(--p-primary-500)' })
</script>

<template>
  <div class="loading-bar">
    <div
      v-show="isLoading"
      class="loading-bar__track"
      data-testid="loading-bar-track"
      aria-hidden="true"
    >
      <div
        class="loading-bar__indicator"
        :style="{ backgroundColor: color }"
        data-testid="loading-bar-indicator"
      />
    </div>
  </div>
</template>

<style scoped>
/* Sits in normal flow directly above the footer. The wrapper is 0px tall (the
   track is v-show-hidden when idle and only 6px when active), and it carries no
   margin, so it adds no layout height — this avoids the earlier overflow while
   keeping the bar at the footer's top edge rather than the viewport bottom.
   `sticky` keeps it visible at the bottom of the scroll area during requests. */
.loading-bar {
  position: sticky;
  bottom: 0;
  z-index: 100;
}

.loading-bar__track {
  width: 100%;
  height: 0.428571rem;
  background: var(--p-surface-300);
  overflow: hidden;
}

.loading-bar__indicator {
  height: 100%;
  width: 40%;
  border-radius: 0.214286rem;
  animation: loading-slide 1.8s ease-in-out infinite;
}

@keyframes loading-slide {
  0% {
    transform: translateX(-100%);
  }
  50% {
    transform: translateX(200%);
  }
  100% {
    transform: translateX(-100%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-bar__indicator {
    animation: none;
    width: 100%;
    opacity: 0.5;
  }
}
</style>
