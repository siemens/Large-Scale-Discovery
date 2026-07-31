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
 * <LinuxIcon> — Tux the penguin SVG icon, sized to 1em × 1em so it scales with
 * surrounding text. Parity with the legacy Fomantic `linux` icon (fa-linux).
 */
const props = withDefaults(
  defineProps<{
    /** Accessible name. Omit when adjacent text already labels the icon. */
    label?: string
  }>(),
  { label: undefined },
)
</script>

<template>
  <!-- Tux the penguin (parity with the legacy Fomantic `linux` icon / fa-linux).
       Body is `currentColor` so it tracks text color; belly, eyes and the
       orange beak/feet are fixed so the penguin stays recognizable on any tint. -->
  <svg
    class="linux-icon"
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    :role="props.label ? 'img' : undefined"
    :aria-label="props.label"
    :aria-hidden="props.label ? undefined : 'true'"
    width="1em"
    height="1em"
    fill="currentColor"
  >
    <!-- Tux is a tall, narrow shape; scale it up about the center so it fills the
         viewBox as fully as the other (square) OS icons rather than reading tiny. -->
    <g transform="translate(12 12) scale(1.35) translate(-12 -12)">
      <!-- Feet -->
      <path
        d="M9.2 19.4c-.4 1.4-1.9 2.4-3.4 2.1-.9-.2-1-1-.4-1.7.5-.6 1.6-1.3 2.2-2.1l1.6 1.7Z"
        fill="#f2a007"
      />
      <path
        d="M14.8 19.4c.4 1.4 1.9 2.4 3.4 2.1.9-.2 1-1 .4-1.7-.5-.6-1.6-1.3-2.2-2.1l-1.6 1.7Z"
        fill="#f2a007"
      />
      <!-- Body -->
      <path
        d="M12 1.8c2.1 0 3.6 1.7 3.6 4 0 .8.1 1.4.6 2.2.4.7.9 1.3 1.4 2 .8 1.1 1.5 2.4 1.7 4.2.3 2.4-.3 4.6-1.6 6.1-1.2 1.4-3.2 2.1-5.7 2.1s-4.5-.7-5.7-2.1c-1.3-1.5-1.9-3.7-1.6-6.1.2-1.8.9-3.1 1.7-4.2.5-.7 1-1.3 1.4-2 .5-.8.6-1.4.6-2.2 0-2.3 1.5-4 3.6-4Z"
      />
      <!-- Belly -->
      <ellipse cx="12" cy="15.6" rx="3.3" ry="4.6" fill="#fff" />
      <!-- Eyes (white) -->
      <ellipse cx="10.5" cy="6" rx="1.25" ry="1.8" fill="#fff" />
      <ellipse cx="13.5" cy="6" rx="1.25" ry="1.8" fill="#fff" />
      <!-- Pupils -->
      <circle cx="10.9" cy="6.4" r=".7" />
      <circle cx="13.1" cy="6.4" r=".7" />
      <!-- Beak -->
      <path d="M10.7 7.7h2.6l-1.3 2Z" fill="#f2a007" />
    </g>
  </svg>
</template>

<style scoped>
/* Drop the icon 0.125em so it sits on the text baseline like a glyph */
.linux-icon {
  display: inline-block;
  vertical-align: -0.125em;
}
</style>
