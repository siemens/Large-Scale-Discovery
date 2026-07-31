<!--
  Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.

  Copyright (c) Siemens AG, 2025-2026.

  This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
  directory or visit <https://opensource.org/licenses/MIT>.

  SPDX-License-Identifier: MIT
-->

<!--
  Thin wrapper around <Transition name="layout-reveal"> that co-locates the
  CSS with the transition name so pages don't need to copy the reveal rules.

  The active phase clips the form while its height collapses.
-->
<template>
  <Transition name="layout-reveal">
    <slot />
  </Transition>
</template>

<style scoped>
.layout-reveal-enter-active,
.layout-reveal-leave-active {
  overflow: hidden;
  transition:
    max-height 0.2s ease,
    opacity 0.2s ease;
  max-height: 2000px;
}

.layout-reveal-enter-from,
.layout-reveal-leave-to {
  max-height: 0;
  opacity: 0;
}
</style>
