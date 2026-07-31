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

import { avataaars, bottts } from '@dicebear/collection'
import { createAvatar, type Style, type StyleOptions } from '@dicebear/core'

// Extract the avataaars style's option vocabulary so our legacy mapping is checked
// against the real v9 enum unions (top, clothing, eyes, …) at compile time.
type AvataaarsOptions = typeof avataaars extends Style<infer O> ? O : never

// ClothingItem is one deterministic avatar clothing option
type ClothingItem = NonNullable<AvataaarsOptions['clothing']>[number]

/**
 * Deterministic avatar generation, restoring the legacy DiceBear presentation.
 *
 * Legacy (`web_frontend/src/js/utils-app.js` + `agents.js`) rendered:
 *   - Users  → DiceBear `avataaars`, seeded by the user's email, drawn inside a
 *              brand-colored circle configured by the owning frontend.
 *   - Agents → DiceBear `bottts` (robot avatars), seeded by the agent name.
 *
 * Both are generated entirely offline (no network / CDN), so the strict CSP we
 * maintain is untouched. The SVG markup is encoded as an image data URI rather
 * than injected into the document, preserving the legacy appearance safely.
 *
 * Generation is deterministic: the same seed always yields the same avatar, so a
 * given user/agent keeps a stable identicon across renders and sessions.
 */

/**
 * Public avatar API. Return encoded image sources that are safe to bind to an
 * `<img src>`; raw SVG generation deliberately stays private so callers cannot
 * accidentally inject generated markup with `v-html`.
 */
export function userAvatarDataUri(
  seed: string,
  gender?: string | null,
  backgroundColor?: string,
): string {
  return svgDataUri(userAvatarSvg(seed, gender, backgroundColor?.replace(/^#/, '')))
}

/** agentAvatarDataUri returns a deterministic avatar image for an agent name. */
export function agentAvatarDataUri(seed: string): string {
  return svgDataUri(agentAvatarSvg(seed))
}

// Character scale within the avatar viewBox. The v9 avataaars artwork renders
// larger inside the circle than the legacy v3 generator, so the portrait is
// scaled to 85% to match the smaller legacy proportions (THEA-660). The theme's
// primary-colored CSS backing remains full-size behind the transparent SVG.
const USER_AVATAR_SCALE = 85

// Memoise rendered SVG per (style, seed) so re-renders of the home panels don't
// re-run the generator for avatars we have already produced.
const userCache = new Map<string, string>()

// Cache generated agent avatars by their stable seed
const agentCache = new Map<string, string>()

// Material Design colors matching legacy agents.js `colors` option, as v9 hex values.
const BOTTTS_COLORS = [
  'ffb300',
  '1e88e5',
  '546e7a',
  '6d4c41',
  '00acc1',
  'f4511e',
  '5e35b1',
  '43a047',
  '3949ab',
  '039be5',
  '7cb342',
  'c0ca33',
  '8e24aa',
  '00897b',
  'fdd835',
]

// ── Legacy avataaars option mapping (utils-app.js initAvatar) ───────────────
// The legacy code ran an old DiceBear (v3-era) avataaars generator whose option
// names and value vocabularies differ from @dicebear v9. The maps below port the
// legacy human-characteristic config — gender-seeded traits included — onto the
// v9 `avataaars` schema so the rendered portraits match the legacy look:
//   - face spills OVER the top of the brand-colored circle (native v9 `style: 'circle'`),
//   - hair/clothing/eyes/mouth/skin variety enabled,
//   - male vs female vs unknown seeded with distinct trait pools.

// Legacy named skin tones → avataaars skin hex.
// Tuned to the warm golden-tan skin tone of the auth background illustration.
const SKIN = {
  light: 'dfa070', // warm golden tan, lighter variation
  brown: 'ddab73', // main tone matching the auth image character
}

// Legacy named hair colours → avataaars hair hex.
// Palette tuned to the warm comic-illustration tone of the app's auth background.
const HAIR = {
  black: '2c1b18',
  blond: 'a87838', // richer amber-gold
  brown: '6a3820', // warm deep brown matching illustration style
  platinum: 'c4a870', // warm champagne rather than washed-out pale
  gray: '9a8878', // warm taupe-gray instead of cool neutral
  red: 'c83010', // orange-red matching image's accent lightning bolts
}

// Legacy named clothing colours → avataaars clothing hex.
// Palette tuned to harmonize with the brand-colored, amber and warm-red tones of the auth background.
const CLOTHING = {
  black: '101820', // near-black warm charcoal
  gray: 'c8c0b0', // bright warm gray
  heather: '253545', // rich dark slate
  amber: 'e09808', // vivid gold
  rust: 'd42010', // strong rust-red
  red: 'f04010', // intense orange-red
  white: 'f8f2e8', // crisp warm white
  navy: '1c3050', // deep warm navy — contrasts the brand-colored circle
  olive: '5a6535', // warm earthy olive
  burgundy: '6a1c28', // deep warm burgundy
  rose: 'd4607a', // warm rose
  blush: 'e8a0aa', // soft blush pink
  mauve: '9c5070', // dusty mauve
}

// Legacy "longHair"/"shortHair" categories expand to concrete v9 `top` styles.
const LONG_HAIR = [
  'bob',
  'bun',
  'curly',
  'curvy',
  'frida',
  'fro',
  'longButNotTooLong',
  'miaWallace',
  'straight01',
  'straight02',
  'straightAndStrand',
  'shaggy',
  'bigHair',
  'dreads01',
] as const

// Keep short hairstyles together for deterministic avatar selection
const SHORT_HAIR = [
  'shortCurly',
  'shortFlat',
  'shortRound',
  'shortWaved',
  'sides',
  'theCaesar',
  'theCaesarAndSidePart',
  'shavedSides',
  'shaggyMullet',
] as const

// Legacy "hat" category (eyepatch became an accessory in v9, so it drops out of `top`).
const HATS = ['hat', 'winterHat02', 'winterHat03', 'winterHat04', 'turban'] as const

// Legacy facialHair list → v9 facialHair vocabulary.
const FACIAL_HAIR = [
  'beardMedium',
  'beardLight',
  'beardMajestic',
  'moustacheFancy',
  'moustacheMagnum',
] as const

// Legacy eyes list → v9 eyes vocabulary (close→closed, dizzy→xDizzy, roll→eyeRoll).
const EYES = [
  'closed',
  'default',
  'xDizzy',
  'eyeRoll',
  'happy',
  'side',
  'squint',
  'surprised',
  'wink',
  'winkWacky',
] as const

// Legacy mouth list (names unchanged in v9).
const MOUTH = [
  'default',
  'disbelief',
  'eating',
  'grimace',
  'screamOpen',
  'serious',
  'smile',
  'tongue',
  'twinkle',
] as const

// Keep graphic clothing options together for deterministic avatar selection
const CLOTHING_GRAPHIC = [
  'bat',
  'bear',
  'deer',
  'diamond',
  'hola',
  'pizza',
  'skull',
  'skullOutline',
] as const

/**
 * Build the gender-seeded avataaars option set, mirroring legacy initAvatar().
 * Typed as StyleOptions<AvataaarsOptions> so TypeScript catches unknown keys and
 * invalid enum values at the return statement rather than silently passing them
 * through a cast.
 */
function userOptions(
  gender: string | null | undefined,
  backgroundColor?: string,
): StyleOptions<AvataaarsOptions> {
  // Normalize the gender code once for the branch comparisons below
  const normalizedGender = (gender ?? '').toUpperCase()

  // Defaults (unknown gender) — legacy default branch
  let top = [...LONG_HAIR, ...SHORT_HAIR, ...HATS]

  // Derive the chance of rendering top hair from the avatar seed
  let topProbability = 95

  // Pick a deterministic hair color from the avatar seed
  let hairColor = Object.values(HAIR)

  // Pick deterministic clothing from the avatar seed
  let clothing: ClothingItem[] = [
    'collarAndSweater',
    'blazerAndShirt',
    'shirtCrewNeck',
    'graphicShirt',
    'hoodie',
    'overall',
  ]

  // Pick a deterministic clothing color from the avatar seed
  let clothesColor = Object.values(CLOTHING)

  // Derive the chance of facial hair from the avatar seed
  let facialHairProbability = 20

  // Derive the chance of accessories from the avatar seed
  let accessoriesProbability = 10

  // Pick a deterministic skin tone from the avatar seed
  let skinColor = [SKIN.light, SKIN.brown]

  if (normalizedGender === 'M') {
    top = [...SHORT_HAIR, ...HATS]
    hairColor = [HAIR.black, HAIR.blond, HAIR.brown]
    clothing = ['collarAndSweater', 'shirtCrewNeck', 'graphicShirt', 'hoodie']
    clothesColor = [
      CLOTHING.black,
      CLOTHING.gray,
      CLOTHING.heather,
      CLOTHING.amber,
      CLOTHING.rust,
      CLOTHING.red,
      CLOTHING.white,
      CLOTHING.navy,
      CLOTHING.olive,
      CLOTHING.burgundy,
    ]
    facialHairProbability = 50
    skinColor = [SKIN.light, SKIN.brown]
  } else if (normalizedGender === 'W' || normalizedGender === 'F') {
    top = [...LONG_HAIR]
    topProbability = 100
    clothing = ['collarAndSweater', 'blazerAndShirt', 'overall']
    facialHairProbability = 0
    accessoriesProbability = 30
  }

  return {
    style: ['circle'],
    backgroundColor: [backgroundColor ?? 'transparent'],
    scale: USER_AVATAR_SCALE,
    top,
    topProbability,
    hairColor,
    clothing,
    clothesColor,
    facialHair: [...FACIAL_HAIR],
    facialHairColor: [HAIR.black, HAIR.brown, HAIR.blond, HAIR.platinum, HAIR.red, HAIR.gray],
    facialHairProbability,
    accessoriesProbability,
    clothingGraphic: [...CLOTHING_GRAPHIC],
    eyes: [...EYES],
    mouth: [...MOUTH],
    skinColor,
  }
}

/**
 * Generate a user avatar (avataaars) as an inline SVG string.
 *
 * Restores the legacy widget-logins presentation (initAvatar): the portrait is
 * drawn with the native avataaars `circle` style inside a theme-colored CSS backing,
 * scaled to 85% so the character matches the smaller legacy v3 proportions
 * (see USER_AVATAR_SCALE). Human characteristics (hair, clothing,
 * eyes, mouth, skin) are enabled, and traits are seeded per gender. Seeded by
 * email + "1" (legacy appended "1") so each user keeps a stable portrait.
 */
function userAvatarSvg(
  seed: string,
  gender: string | null | undefined,
  backgroundColor?: string,
): string {
  // Serve a memoized SVG when this (gender, seed) pair was already rendered
  const key = `${backgroundColor ?? 'transparent'}:${(gender ?? '').toUpperCase()}:${seed || 'unknown'}`

  // Reuse the previously generated value when available
  const cached = userCache.get(key)
  if (cached) return cached

  // Generate the portrait
  const svg = createAvatar(avataaars, {
    seed: (seed || 'unknown') + 'n273m',
    ...userOptions(gender, backgroundColor),
  }).toString() // cast to the strict v9 vocabulary validated by AvataaarsOptions

  // Cache and return the rendered markup
  userCache.set(key, svg)
  return svg
}

/**
 * Generate an agent avatar (bottts) as an inline SVG string.
 * Seeded by the agent name, matching the legacy agents panel robot avatars.
 * No radius (legacy had none); baseColor restricted to the same Material Design
 * palette as the legacy `colors` option.
 */
function agentAvatarSvg(seed: string): string {
  // Serve a memoized SVG when this agent name was already rendered
  const key = seed || 'unknown'

  // Reuse the previously generated value when available
  const cached = agentCache.get(key)
  if (cached) return cached

  // Generate the robot avatar restricted to the legacy Material Design palette
  const svg = createAvatar(bottts, {
    seed: key,
    baseColor: BOTTTS_COLORS,
  }).toString()

  // Cache and return the rendered markup
  agentCache.set(key, svg)
  return svg
}

// Encode SVG markup as a browser-safe data URI
function svgDataUri(svg: string): string {
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
}
