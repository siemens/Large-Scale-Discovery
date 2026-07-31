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

// ── Imports ───────────────────────────────────────────────────
import { definePreset } from '@primeuix/themes'
import Aura from '@primeuix/themes/aura'
import { createPinia } from 'pinia'
import 'primeicons/primeicons.css'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip'
import { createApp } from 'vue'

import '@/lib/zodConfig'
import { loadSession, onUnauthorized } from '@/api/session'
import App from '@/App.vue'
import { i18n } from '@/i18n'
import {
  APP_BLUE_100,
  APP_BLUE_200,
  APP_BLUE_300,
  APP_BLUE_400,
  APP_BLUE_50,
  APP_BLUE_500,
  APP_BLUE_600,
  APP_BLUE_700,
  APP_BLUE_800,
  APP_BLUE_900,
  APP_BLUE_950,
  APP_ERROR_100,
  APP_ERROR_200,
  APP_ERROR_300,
  APP_ERROR_50,
  APP_ERROR_700,
  APP_GREEN_100,
  APP_GREEN_200,
  APP_GREEN_300,
  APP_GREEN_400,
  APP_GREEN_50,
  APP_GREEN_500,
  APP_GREEN_600,
  APP_GREEN_700,
  APP_GREEN_800,
  APP_GREEN_900,
  APP_GREEN_950,
  APP_INFO_100,
  APP_INFO_300,
  APP_INFO_50,
  APP_INFO_700,
  APP_OFFWHITE,
  APP_PRIMARY_100,
  APP_PRIMARY_200,
  APP_PRIMARY_300,
  APP_PRIMARY_400,
  APP_PRIMARY_50,
  APP_PRIMARY_500,
  APP_PRIMARY_600,
  APP_PRIMARY_700,
  APP_PRIMARY_800,
  APP_PRIMARY_900,
  APP_PRIMARY_950,
  APP_RED_100,
  APP_RED_200,
  APP_RED_300,
  APP_RED_400,
  APP_RED_50,
  APP_RED_500,
  APP_RED_600,
  APP_RED_700,
  APP_RED_800,
  APP_RED_900,
  APP_RED_950,
  APP_SECONDARY_100,
  APP_SECONDARY_200,
  APP_SECONDARY_300,
  APP_SECONDARY_400,
  APP_SECONDARY_50,
  APP_SECONDARY_500,
  APP_SECONDARY_600,
  APP_SECONDARY_700,
  APP_SECONDARY_800,
  APP_SECONDARY_900,
  APP_SECONDARY_950,
  APP_SUCCESS_100,
  APP_SUCCESS_200,
  APP_SUCCESS_300,
  APP_SUCCESS_50,
  APP_SUCCESS_700,
  APP_WARN_100,
  APP_WARN_200,
  APP_WARN_300,
  APP_WARN_50,
  APP_WARN_600,
  APP_WARN_700,
  APP_YELLOW_100,
  APP_YELLOW_200,
  APP_YELLOW_300,
  APP_YELLOW_400,
  APP_YELLOW_50,
  APP_YELLOW_500,
  APP_YELLOW_600,
  APP_YELLOW_700,
  APP_YELLOW_800,
  APP_YELLOW_900,
  APP_YELLOW_950,
} from '@/lib/colors'
import { createAuthRedirectFlow } from '@/lib/redirect'
import { installStaleAssetReload } from '@/lib/staleAssets'
import { router } from '@/router'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { useUnsavedChangesStore } from '@/stores/unsavedChanges'

// ── CSS custom properties ─────────────────────────────────────
// Expose JS design tokens that are also needed in plain CSS as :root custom
// properties, so colors.ts stays the single source of truth (consumed both by
// the PrimeVue preset below and via var(--app-*) in component styles).
const rootStyle = document.documentElement.style
rootStyle.setProperty('--app-color-offwhite', APP_OFFWHITE)
rootStyle.setProperty('--app-font-size-small', '0.89rem') // Shared small-text size for buttons, tooltips, hints
rootStyle.setProperty('--app-font-size-tiny', '0.79rem') // Shared tiny-text size for labels, captions, and fine print
rootStyle.setProperty('--app-data-table-row-height', '2.5rem') // Shared DataTable body row height; virtual scrolling resolves this to pixels.

// Severity color trios (background / border / text) for plain-CSS consumers such
// as <FieldHint severity="...">. These mirror the PrimeVue message/toast theme
// mapping above exactly (50 = background, 200/300 = border, 600/700 = text) so a
// help label reads as the same visual language as a <Message> banner.
rootStyle.setProperty('--app-info-bg', APP_INFO_50)
rootStyle.setProperty('--app-info-border', APP_INFO_300)
rootStyle.setProperty('--app-info-text', APP_INFO_700)
rootStyle.setProperty('--app-success-bg', APP_SUCCESS_50)
rootStyle.setProperty('--app-success-border', APP_SUCCESS_200)
rootStyle.setProperty('--app-success-text', APP_SUCCESS_700)
rootStyle.setProperty('--app-warn-bg', APP_WARN_50)
rootStyle.setProperty('--app-warn-border', APP_WARN_200)
rootStyle.setProperty('--app-warn-text', APP_WARN_600)
rootStyle.setProperty('--app-error-bg', APP_ERROR_50)
rootStyle.setProperty('--app-error-border', APP_ERROR_200)
rootStyle.setProperty('--app-error-text', APP_ERROR_700)

// Global text colors — primary body text and the muted/secondary variant. Held
// as constants so the semantic text tokens (and any reuse) share one source.
const APP_TEXT_COLOR = 'black'

// Use the shared muted text color throughout the PrimeVue preset
const APP_TEXT_MUTED_COLOR = APP_SECONDARY_950

// ── PrimeVue preset ───────────────────────────────────────────
// ColorPreset extends Aura with the brand color palette, typography, component
// sizes, and severity colors, so no component needs individual CSS overrides.
const ColorPreset = definePreset(Aura, {
  semantic: {
    // Self-hosted Lato — v1 typeface. Fallback stack mirrors the v1 font stack.
    font: {
      family: "'Lato', 'Helvetica Neue', Arial, Helvetica, system-ui, sans-serif",
    },

    // Aura defaults iconSize to 1rem which resolves to the browser root
    // (16px), not our 14px body. Pin it so icons match body text globally.
    iconSize: '1rem',

    // Square-ish corners to match the action buttons (which use borderRadius.sm =
    // 4px). These two semantic tokens are scheme-independent, so they sit at the
    // root: formField.borderRadius drives every input control (inputtext, textarea,
    // select, etc. all resolve to {form.field.border.radius}); content.borderRadius
    // drives content surfaces such as Message/Toast ({content.border.radius}).
    // Setting them here avoids per-component overrides and any CSS rules.
    formField: {
      borderRadius: '{borderRadius.sm}',
    },
    content: {
      borderRadius: '{borderRadius.sm}',
    },

    // Global text colors — body text uses rgba(0,0,0,0.87) for primary and
    // rgba(0,0,0,0.6) for muted/secondary text. Setting them here instead of
    // in :root CSS ensures PrimeVue components inherit the values.
    disabledOpacity: '0.7',
    colorScheme: {
      light: {
        text: {
          color: APP_TEXT_COLOR,
          mutedColor: APP_TEXT_MUTED_COLOR,
        },

        // Shared field colors. Every control (inputtext, textarea, select,
        // inputnumber, datepicker, etc.) resolves its color/placeholder/disabled
        // props to these form.field.* tokens, so setting them once here applies
        // app-wide instead of per component.
        formField: {
          color: APP_TEXT_MUTED_COLOR,
          placeholderColor: APP_SECONDARY_500,
          invalidPlaceholderColor: APP_SECONDARY_500,
          disabledBackground: APP_SECONDARY_50,
          disabledColor: APP_SECONDARY_900,

          // A resting FloatLabel acts as the field placeholder. Once focused or
          // populated it becomes a persistent label and uses the darker text color.
          floatLabelColor: APP_SECONDARY_500,
          floatLabelFocusColor: APP_TEXT_MUTED_COLOR,
          floatLabelActiveColor: APP_TEXT_MUTED_COLOR,
        },
        mask: {
          background: 'rgba(0, 0, 0, 0.5)',
        },

        // Full neutral gray surface ladder. Aura's default surface is Tailwind
        // slate (blue-tinted); we replace the whole ramp with the true-neutral
        // (R=G=B) APP_GRAY scale from colors.ts, so
        // every --p-surface-* (backgrounds, borders, muted text/icons, hovers)
        // is a consistent neutral grey. Only surface-0 stays pure white (#ffffff),
        // which has no APP_GRAY counterpart. The ladder is monotonic by luminance.
        surface: {
          0: '#ffffff',
          50: APP_SECONDARY_50,
          100: APP_SECONDARY_100,
          200: APP_SECONDARY_200,
          300: APP_SECONDARY_300,
          400: APP_SECONDARY_400,
          500: APP_SECONDARY_500, // Base color
          600: APP_SECONDARY_600,
          700: APP_SECONDARY_700,
          800: APP_SECONDARY_800,
          900: APP_SECONDARY_900,
          950: APP_SECONDARY_950,
        },
        primary: {
          contrastColor: APP_OFFWHITE,
        },
      },
    },

    // Custom teal palette anchored at the v1 teal (APP_PRIMARY_500) as the 500
    // shade. PrimeVue v4's Aura preset maps semantic.primary to Tailwind's
    // teal.* ladder, whose teal.500 (#14B8A6) is more cyan-shifted and
    // saturated than v1. These brand values restore the v1 hue
    // (≈hsl(177, 100%, 35%)) across every primary-tinted element (buttons,
    // links, focus rings, badges, active nav highlights, etc.). The 50-950
    // ladder is monotonic by luminance with no perceptual reversals.
    primary: {
      50: APP_PRIMARY_50,
      100: APP_PRIMARY_100,
      200: APP_PRIMARY_200,
      300: APP_PRIMARY_300,
      400: APP_PRIMARY_400,
      500: APP_PRIMARY_500, // Base color
      600: APP_PRIMARY_600,
      700: APP_PRIMARY_700,
      800: APP_PRIMARY_800,
      900: APP_PRIMARY_900,
      950: APP_PRIMARY_950,
    },
    red: {
      50: APP_RED_50,
      100: APP_RED_100,
      200: APP_RED_200,
      300: APP_RED_300,
      400: APP_RED_400,
      500: APP_RED_500, // Base color
      600: APP_RED_600,
      700: APP_RED_700,
      800: APP_RED_800,
      900: APP_RED_900,
      950: APP_RED_950,
    },
    yellow: {
      50: APP_YELLOW_50,
      100: APP_YELLOW_100,
      200: APP_YELLOW_200,
      300: APP_YELLOW_300,
      400: APP_YELLOW_400,
      500: APP_YELLOW_500, // Base color
      600: APP_YELLOW_600,
      700: APP_YELLOW_700,
      800: APP_YELLOW_800,
      900: APP_YELLOW_900,
      950: APP_YELLOW_950,
    },

    // Tailwind-compatible optional color ladders. Each family follows the App's
    // step rhythm; red/yellow keep the brand ladders, while blue and the
    // remaining families use the optional palettes in colors.ts.
    teal: {
      50: APP_PRIMARY_50,
      100: APP_PRIMARY_100,
      200: APP_PRIMARY_200,
      300: APP_PRIMARY_300,
      400: APP_PRIMARY_400,
      500: APP_PRIMARY_500, // Base color
      600: APP_PRIMARY_600,
      700: APP_PRIMARY_700,
      800: APP_PRIMARY_800,
      900: APP_PRIMARY_900,
      950: APP_PRIMARY_950,
    },
    blue: {
      50: APP_BLUE_50,
      100: APP_BLUE_100,
      200: APP_BLUE_200,
      300: APP_BLUE_300,
      400: APP_BLUE_400,
      500: APP_BLUE_500, // Base color
      600: APP_BLUE_600,
      700: APP_BLUE_700,
      800: APP_BLUE_800,
      900: APP_BLUE_900,
      950: APP_BLUE_950,
    },
    green: {
      50: APP_GREEN_50,
      100: APP_GREEN_100,
      200: APP_GREEN_200,
      300: APP_GREEN_300,
      400: APP_GREEN_400,
      500: APP_GREEN_500, // Base color
      600: APP_GREEN_600,
      700: APP_GREEN_700,
      800: APP_GREEN_800,
      900: APP_GREEN_900,
      950: APP_GREEN_950,
    },
  },
  components: {
    button: {
      root: {
        // v1 button base: 4px radius (Aura's borderRadius.sm token), bold weight,
        // and wider horizontal padding (was a `.p-button { padding-inline }`
        // !important override in App.vue — the token expresses it cleanly).
        borderRadius: 'var(--p-border-radius-sm)',
        paddingX: '1.5rem',
        label: {
          fontWeight: '600',
        },
      },
      colorScheme: {
        light: {
          root: {
            primary: {
              background: APP_PRIMARY_500,
              hoverBackground: APP_PRIMARY_600,
              activeBackground: APP_PRIMARY_700,
              borderColor: APP_PRIMARY_500,
              hoverBorderColor: APP_PRIMARY_600,
              activeBorderColor: APP_PRIMARY_700,
              color: APP_OFFWHITE,
              hoverColor: APP_OFFWHITE,
              activeColor: APP_OFFWHITE,
            },
            secondary: {
              background: APP_SECONDARY_300,
              hoverBackground: APP_SECONDARY_400,
              activeBackground: APP_SECONDARY_500,
              borderColor: APP_SECONDARY_300,
              hoverBorderColor: APP_SECONDARY_400,
              activeBorderColor: APP_SECONDARY_500,
              color: APP_TEXT_MUTED_COLOR, // Dark font on gray is a nicer contrast than white
              hoverColor: 'rgba(0, 0, 0, 0.7)', // Dark font on gray is a nicer contrast than white
              activeColor: 'rgba(0, 0, 0, 0.8)', // Dark font on gray is a nicer contrast than white
            },
            warn: {
              background: APP_YELLOW_500,
              hoverBackground: APP_YELLOW_600,
              activeBackground: APP_YELLOW_700,
              borderColor: APP_YELLOW_500,
              hoverBorderColor: APP_YELLOW_600,
              activeBorderColor: APP_YELLOW_700,
              color: APP_OFFWHITE,
              hoverColor: APP_OFFWHITE,
              activeColor: APP_OFFWHITE,
            },
            danger: {
              background: APP_RED_500,
              hoverBackground: APP_RED_600,
              activeBackground: APP_RED_700,
              borderColor: APP_RED_500,
              hoverBorderColor: APP_RED_600,
              activeBorderColor: APP_RED_700,
              color: APP_OFFWHITE,
              hoverColor: APP_OFFWHITE,
              activeColor: APP_OFFWHITE,
            },
          },
        },
      },
    },

    // Leading/trailing icons inside an IconField (e.g. the Login/Register fields).
    // Default is a muted field icon color; darken it to the primary body text
    // color so the glyphs read clearly against the input.
    iconfield: {
      icon: {
        color: APP_TEXT_MUTED_COLOR,
      },
    },

    // Teal slider — used by the scope Settings (agent-process counts) and the
    // token panel's retention picker. Set globally so the component CSS stays clean.
    // Restyled to the legacy Fomantic look: a taller light-grey rounded track with
    // a teal fill and a white circular thumb (border + subtle shadow) rather than
    // Aura's small solid-color knob.
    slider: {
      track: {
        background: APP_SECONDARY_200,
        borderRadius: '4px',
        size: '6px',
      },
      handle: {
        width: '18px',
        height: '18px',

        // Same fill at rest and on hover so the thumb keeps its color when hovered
        // (Aura's default hoverBackground would otherwise tint it).
        background: APP_SECONDARY_50,
        hoverBackground: APP_SECONDARY_100,

        // Inner content dot matches the thumb fill so the thumb reads as one solid
        // circle, mirroring the legacy gradient thumb. The thumb border is added in
        // App.vue (Aura exposes no handle border token).
        content: {
          width: '18px',
          height: '18px',
          background: APP_SECONDARY_50,
          hoverBackground: APP_SECONDARY_100,
          shadow: '0px 0.5px 0px 0px rgba(0, 0, 0, 0.08), 0px 1px 1px 0px rgba(0, 0, 0, 0.14)',
        },
      },
      colorScheme: {
        light: {
          // Aura's base preset pins handle.content.background to {surface.0} (white)
          // at the scheme level, which outranks the outer handle above. Re-assert
          // surface-50 (at rest and on hover) so the thumb's inner dot matches its fill.
          handle: {
            content: {
              background: APP_SECONDARY_50,
              hoverBackground: APP_SECONDARY_100,
            },
          },
        },
      },
    },

    // Message banner colors (was CSS overrides in App.vue). Each severity maps
    // onto its family's standard rungs — 50 = background, 200 = border,
    // 800 = text — so banners stay in lock-step with the brand ladders.
    message: {
      text: {
        fontSize: 'var(--app-font-size-small)',
      },
      colorScheme: {
        light: {
          info: {
            background: APP_INFO_50,
            borderColor: APP_INFO_300,
            color: APP_INFO_700,
          },
          success: {
            background: APP_SUCCESS_50,
            borderColor: APP_SUCCESS_300,
            color: APP_SUCCESS_700,
          },
          warn: {
            background: APP_WARN_50,
            borderColor: APP_WARN_300,
            color: APP_WARN_600,
          },
          error: {
            background: APP_ERROR_50,
            borderColor: APP_ERROR_300,
            color: APP_ERROR_700,
          },
        },
      },
    },

    // Tooltip style override: light surface background with muted text (Aura's
    // default is a dark surface — {surface.700} — for both light and dark modes).
    // The Aura tooltip theme defines background/color under colorScheme.light.root,
    // so overrides must go there to take effect. `border` and `fontSize` have no
    // tooltip preset token, so those two rules live in App.vue global CSS.
    tooltip: {
      root: {
        maxWidth: '360px',
      },
      colorScheme: {
        light: {
          root: {
            background: '{surface.0}',
            color: '{text.mutedColor}',
          },
        },
      },
    },

    // Popover (scope cycle-progress popup, agent scan-tasks popup) shares the
    // tooltip's surface look — white background, muted text, subtle border.
    // Setting the tokens (not CSS) also recolors the popover's arrow to match.
    popover: {
      root: {
        background: '{surface.0}',
        borderColor: '{surface.300}',
        color: '{text.mutedColor}',
      },
    },

    // Toast colors (match the message palette above; detailColor mirrors the
    // text color). Same rung mapping — 50 = background, 200 = border, 800 = text.
    toast: {
      colorScheme: {
        light: {
          info: {
            background: APP_INFO_100,
            borderColor: APP_INFO_300,
            color: APP_INFO_700,
            detailColor: APP_INFO_700,
          },
          success: {
            background: APP_SUCCESS_100,
            borderColor: APP_SUCCESS_300,
            color: APP_SUCCESS_700,
            detailColor: APP_SUCCESS_700,
          },
          warn: {
            background: APP_WARN_100,
            borderColor: APP_WARN_300,
            color: APP_WARN_700,
            detailColor: APP_WARN_700,
          },
          error: {
            background: APP_ERROR_100,
            borderColor: APP_ERROR_300,
            color: APP_ERROR_700,
            detailColor: APP_ERROR_700,
          },
        },
      },
    },

    // Tag colors mirror the action-button severities (solid brand fill + off-white
    // text) instead of Aura's default light-tint tags, so the API-docs method
    // badges (GET/POST → primary teal, PUT → warn yellow, DELETE/PATCH → danger
    // red), the endpoint count, and the status-code badges read as the same
    // visual language as the form buttons.
    tag: {
      colorScheme: {
        light: {
          primary: {
            background: APP_PRIMARY_500,
            color: APP_OFFWHITE,
          },
          secondary: {
            background: APP_SECONDARY_300,
            color: APP_TEXT_MUTED_COLOR, // Dark font on gray is a nicer contrast than white
          },
          warn: {
            background: APP_YELLOW_500,
            color: APP_OFFWHITE,
          },
          danger: {
            background: APP_RED_500,
            color: APP_OFFWHITE,
          },
        },
      },
    },

    // DataTable styled to match the hand-rolled DataGrid tables (scope/view
    // lists in components/elements/DataGrid.vue): a faint-grey header, bold column titles,
    // and compact cell padding. Consumers add `show-gridlines` for the
    // borders-all-around look; the border color comes from the content token.
    datatable: {
      headerCell: {
        background: APP_SECONDARY_50,
        hoverBackground: APP_SECONDARY_100,
        color: '{text.color}',
        hoverColor: '{text.color}',
        borderColor: '{content.border.color}',
        padding: '0.6rem 0.75rem',
      },
      columnTitle: {
        fontWeight: '700',
      },
      bodyCell: {
        borderColor: '{content.border.color}',
        padding: '0.5rem 0.75rem',
      },

      // Pin sort icon color at rest and on hover to the same text color so
      // PrimeVue never changes it. Opacity-based dimming in DataTable.vue
      // handles the visual effect without any specificity conflict.
      sortIcon: {
        color: '{text.color}',
        hoverColor: '{text.color}',
      },
    },
  },
})

// ── App bootstrap ─────────────────────────────────────────────
loadSession()

// Create the root Vue application
const app = createApp(App)

// Create the shared Pinia state container
const pinia = createPinia()

// Coordinate authentication redirects with the router
const authRedirect = createAuthRedirectFlow(router)

// Old hashed chunks disappear during deploys, so reload once when a stale bundle asks for them
installStaleAssetReload(router, {
  beforeReload: () => useUnsavedChangesStore(pinia).discardAll(),
})

// PrimeVue injects theme styles at runtime. Forward the server-provided nonce
// so those style elements comply with the nonce-based CSP.
const cspNonce = document.querySelector<HTMLMetaElement>('meta[name="csp-nonce"]')?.content ?? ''

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(PrimeVue, {
  theme: {
    preset: ColorPreset,
    options: { darkModeSelector: false },
  },
  csp: { nonce: cspNonce },
})
app.use(ToastService)

// Register the styled tooltip once for every `v-tooltip` consumer.
app.directive('tooltip', Tooltip)

// Resolve authentication state after Pinia is installed
const authStore = useAuthStore(pinia)
authRedirect.observeAuthentication(() => authStore.isAuthenticated)

// A centralized 401 handler clears local authentication and returns protected
// routes to login without duplicating the behavior in every API consumer.
onUnauthorized(() => {
  // Session loss cannot preserve open form edits, so bypass the user-driven
  // discard dialog before redirecting to the login page.
  useUnsavedChangesStore(pinia).discardAll()
  authStore.logout()
  if (router.currentRoute.value.name !== 'login') {
    void router.push('/login')
  }
})

// Restore persisted authentication before fetching authenticated settings.
void authStore.restore().then(async () => {
  if (!authStore.isAuthenticated) {
    authRedirect.handleUnauthenticatedRestore()
    return
  }

  void useSettingsStore(pinia).fetchSettings('/backend/settings')
  await authRedirect.restorePendingRedirect()
})

app.mount('#app')
authRedirect.showFailure()
