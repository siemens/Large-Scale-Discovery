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

// Each color is a 50-950 ladder (50 = lightest tint, 950 = darkest shade),
// mirroring the PrimeVue/Tailwind shade convention so these map directly onto
// the theme's --p-*-{50..950} tokens in main.ts. The 500 rung is the brand base
// value; the other rungs are derived from it (and any fixed anchors) in OKLCH —
// a perceptually-uniform space — with constant hue, an even lightness curve and
// gamut-fitted chroma, so every ladder is monotonic by lightness without
// perceptual reversals.

/** APP_OFFWHITE is the translucent white used on dark chart and overlay backgrounds. */
export const APP_OFFWHITE = 'rgba(255, 255, 255, 0.9)'

/** Shared chart series palette rendered directly by Canvas. */
export const APP_CHART_COLORS = [
  '#fecf00',
  '#00b5ad',
  '#9966ff',
  '#fee622',
  '#ff6384',
  '#009c95',
  '#36a2eb',
  '#ff9f40',
  '#c9cbcf',
  '#db2828',
  '#f0b407',
  '#dc8cc4',
  '#04646c',
  '#05b4ac',
  '#046564',
  '#e04868',
  '#fcd06a',
  '#2e999b',
  '#f4742c',
] as const

/** Translucent series color reserved for aggregate “Others” chart entries. */
export const APP_CHART_OTHERS_COLOR = 'rgba(201, 203, 207, 0.44)'

/** APP_CHART_MUTED_COLOR is the stronger neutral used for de-emphasized chart series. */
export const APP_CHART_MUTED_COLOR = 'rgba(201, 203, 207, 0.84)'

// Primary teal palette used by the discovery frontend
export const APP_PRIMARY_50 = '#e0f2f1'
export const APP_PRIMARY_100 = '#b2dfdb'
export const APP_PRIMARY_200 = '#80cbc4'
export const APP_PRIMARY_300 = '#4db6ac'
export const APP_PRIMARY_400 = '#26a69a'
export const APP_PRIMARY_500 = '#009590'
export const APP_PRIMARY_600 = '#008080'
export const APP_PRIMARY_700 = '#007070'
export const APP_PRIMARY_800 = '#005551'
export const APP_PRIMARY_900 = '#003d3a'
export const APP_PRIMARY_950 = '#002725'

/*
 *
 * Main colors used throughout the application
 *
 */

// Secondary — true-neutral grey (R=G=B, no blue tint). Drives the secondary
// button, and replaces Aura's slate surface ramp so every --p-surface-* token
// (backgrounds, borders, muted text/icons, hovers) is a consistent neutral grey.
export const APP_SECONDARY_50 = '#f6f6f6'
export const APP_SECONDARY_100 = '#ebebeb'
export const APP_SECONDARY_200 = '#dfdfdf'
export const APP_SECONDARY_300 = '#d2d2d2'
export const APP_SECONDARY_400 = '#c5c5c5'
export const APP_SECONDARY_500 = '#b8b8b8' // Base value
export const APP_SECONDARY_600 = '#aaaaaa'
export const APP_SECONDARY_700 = '#9a9a9a'
export const APP_SECONDARY_800 = '#818181'
export const APP_SECONDARY_900 = '#5d5d5d'
export const APP_SECONDARY_950 = '#3d3d3d'

// Yellow — bright brand yellow. Used for warn BUTTONS (the high-emphasis,
// saturated brand yellow), and exposed as the theme's --p-yellow-* ladder. For the
// softer warn message/toast banners use APP_WARN_* below instead.
export const APP_YELLOW_50 = '#fff6e5'
export const APP_YELLOW_100 = '#feecc7'
export const APP_YELLOW_200 = '#fee1a5'
export const APP_YELLOW_300 = '#fed582'
export const APP_YELLOW_400 = '#fcc64d'
export const APP_YELLOW_500 = '#f6b906' // Base value
export const APP_YELLOW_600 = '#e3aa00'
export const APP_YELLOW_700 = '#ca9700'
export const APP_YELLOW_800 = '#af8301'
export const APP_YELLOW_900 = '#936e03'
export const APP_YELLOW_950 = '#775801'

// Red — brand danger/error. Drives danger buttons, error message/toast banners,
// and status icons; exposed as the theme's --p-red-* ladder (replacing Aura's
// default Tailwind red so error states follow the brand hue).
export const APP_RED_50 = '#ffece9'
export const APP_RED_100 = '#ffcfca'
export const APP_RED_200 = '#f8aca4'
export const APP_RED_300 = '#ee897f'
export const APP_RED_400 = '#e46159'
export const APP_RED_500 = '#d73434' // Base value
export const APP_RED_600 = '#bc2729'
export const APP_RED_700 = '#a01e21'
export const APP_RED_800 = '#801618'
export const APP_RED_900 = '#5f1212'
export const APP_RED_950 = '#400c0b'

// Blue - optional utility blue.
export const APP_BLUE_50 = '#e1f0fd'
export const APP_BLUE_100 = '#b5d9f9'
export const APP_BLUE_200 = '#87c2f4'
export const APP_BLUE_300 = '#5aa9e9'
export const APP_BLUE_400 = '#3d98dd'
export const APP_BLUE_500 = '#2e88ca' // Base value
export const APP_BLUE_600 = '#2875ae'
export const APP_BLUE_700 = '#226699'
export const APP_BLUE_800 = '#164d75'
export const APP_BLUE_900 = '#0d3755'
export const APP_BLUE_950 = '#062338'

// Green - optional utility green.
export const APP_GREEN_50 = '#e3f2e5'
export const APP_GREEN_100 = '#bae0c0'
export const APP_GREEN_200 = '#8fcc9b'
export const APP_GREEN_300 = '#65b879'
export const APP_GREEN_400 = '#4ba764'
export const APP_GREEN_500 = '#3c9756' // Base value
export const APP_GREEN_600 = '#34824a'
export const APP_GREEN_700 = '#2d7240'
export const APP_GREEN_800 = '#1f562e'
export const APP_GREEN_900 = '#143e20'
export const APP_GREEN_950 = '#0a2812'

/*
 *
 * Less intense color variations for messages and toasts
 *
 */

// Info — calm blue for low-emphasis surfaces (info message/toast banners).
// There is no brand blue elsewhere, so this exists only for the info severity.
// Anchored on the legacy Fomantic info palette and mapped onto the banner as
// 50/100 = backgrounds, 300 = border, 700 = text.
export const APP_INFO_50 = '#f8ffff' // Fomantic info anchor (background)
export const APP_INFO_100 = '#d6e8f0'
export const APP_INFO_200 = '#a9d5de' // Fomantic info anchor (border)
export const APP_INFO_300 = '#8dbed2'
export const APP_INFO_400 = '#6fa9c0'
export const APP_INFO_500 = '#5594ac' // Base value
export const APP_INFO_600 = '#407f96'
export const APP_INFO_700 = '#31697e'
export const APP_INFO_800 = '#0e566c' // Fomantic info anchor (text)
export const APP_INFO_900 = '#123642'
export const APP_INFO_950 = '#07212a'

// Success — soft teal for positive surfaces (success message/toast banners) and
// positive status glyphs (tick icons, "full support" markers). Derived from the
// Primary brand hue at reduced chroma to match the tone of APP_INFO_*, APP_WARN_*
// and APP_ERROR_*. Mapped onto the banner as 50 = background, 200 = border, 700 = text.
// Monotonic by luminance.
export const APP_SUCCESS_50 = '#fcfff5' // Fomantic positive anchor (background)
export const APP_SUCCESS_100 = '#e0f1d2'
export const APP_SUCCESS_200 = '#a3c293' // Fomantic positive anchor (border)
export const APP_SUCCESS_300 = '#8bb078'
export const APP_SUCCESS_400 = '#73a05f'
export const APP_SUCCESS_500 = '#5d9150' // Base value
export const APP_SUCCESS_600 = '#4a8241'
export const APP_SUCCESS_700 = '#3a7134'
export const APP_SUCCESS_800 = '#2c662d' // Fomantic positive anchor (text)
export const APP_SUCCESS_900 = '#1c4a1e'
export const APP_SUCCESS_950 = '#123212'

// Warn — soft amber for low-emphasis surfaces (warn message/toast banners),
// distinct from the bright brand yellow above. Mapped onto the banner as
// 50 = background, 200 = border, 600 = text. Keeps the legacy Fomantic warning
// lightness for parity but is hue-shifted toward yellow (less brown).
export const APP_WARN_50 = '#fffaf1'
export const APP_WARN_100 = '#f8d692'
export const APP_WARN_200 = '#e2b450'
export const APP_WARN_300 = '#cd9e2c'
export const APP_WARN_400 = '#b68800'
export const APP_WARN_500 = '#9c7400' // Base value
export const APP_WARN_600 = '#836100'
export const APP_WARN_700 = '#6b4e00'
export const APP_WARN_800 = '#543c00'
export const APP_WARN_900 = '#352500'
export const APP_WARN_950 = '#211500'

// Error — soft red for low-emphasis surfaces (error message/toast banners),
// distinct from the bright brand red above. Anchored on the legacy Fomantic
// error palette and mapped onto the banner as 50 = background, 200 = border,
// 700 = text.
export const APP_ERROR_50 = '#fff6f6' // Fomantic error anchor (background)
export const APP_ERROR_100 = '#f8d2ce'
export const APP_ERROR_200 = '#e0b4b4' // Fomantic error anchor (border)
export const APP_ERROR_300 = '#eb958e'
export const APP_ERROR_400 = '#e27f78'
export const APP_ERROR_500 = '#d56b65' // Base value
export const APP_ERROR_600 = '#c25b57'
export const APP_ERROR_700 = '#ab4f4b'
export const APP_ERROR_800 = '#9f3a38' // Fomantic error anchor (text)
export const APP_ERROR_900 = '#632927'
export const APP_ERROR_950 = '#411716'
