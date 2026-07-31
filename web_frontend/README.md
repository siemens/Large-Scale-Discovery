# web_frontend

Vue 3 + TypeScript + Vite + PrimeVue rewrite of the legacy KnockoutJS `web_frontend_old/`. It provides
the same authenticated SPA (dashboard, admin pages, configuration views) with a modern stack:
component-scoped CSS, Pinia stores, Zod-validated API contracts, and a strict Content Security
Policy. `web_backend` serves `dist/` in both dev and production mode.

## Quick start

```bash
npm ci
npm run dev
```

The Vite development server runs at `http://127.0.0.1:5173` and proxies only
`/api/v1` to the backend.

## Quality gates

```bash
npm run format:check  # Prettier consistency
npm run lint          # typed ESLint rules for Vue, TypeScript, and tests
npm run typecheck     # application, configuration, and test projects
npm run dead-code     # unused files, exports, and dependencies
npm run test:unit     # Vitest unit and component tests
npm run test:e2e      # Playwright tests; starts Vite automatically
npm run build         # production bundle
```

`npm run check` runs the deterministic formatting, lint, type, dead-code, unit-test,
and build gates in sequence. Playwright remains a separate gate because it is
browser- and platform-dependent. Unit tests live in `tests/unit`; browser
tests live in `tests/e2e`.

## Architecture

- `src/main.ts` is the composition root. It defines the PrimeVue theme,
  installs framework plugins, restores the session, and mounts the app.
- `src/pages` contains route entry points and page orchestration.
- `src/components/forms` and `src/components/widgets` contain feature-level
  presentation; `src/components/elements` contains reusable UI primitives.
- `src/stores` owns client and server state orchestration. Stores may depend on
  the API and domain modules, never on pages or components.
- `src/api` owns transport, normalized errors, retries, authentication headers,
  and session persistence. It must not depend on presentation or Pinia stores.
- `src/types` contains backend/domain contracts; `src/lib` contains pure
  domain helpers; `src/composables` contains reusable Vue behavior.
- `src/App.vue` owns the global reset, typography, accessibility utilities,
  and unavoidable teleported PrimeVue overrides alongside its scoped shell
  styles. Other component styling remains scoped and colocated.

These dependency directions are enforced by `eslint.config.mjs`.

## Running with web_backend

The backend serves `dist/` in both dev and production mode. In dev mode, if `dist/` does not exist
the backend still starts (API-only) and logs a warning.

### Development (auto-rebuild on save)

```bash
# Terminal 1 — rebuild on every file change
cd web_frontend
npm install
npm run build:watch

# Terminal 2 — start the backend (serves dist/)
cd _bin
./backend          # or: go run ../web_backend
```

Save a `.vue` / `.ts` file, refresh the browser — done. No second server, no proxy.

### Development (Vite dev server — alternative)

If you prefer hot-module replacement during UI work, use the Vite dev server instead:

```bash
cd web_frontend
npm run dev        # http://localhost:5173, proxies /api/v1 to the backend
```

### Production

```bash
# 1. Build the frontend
cd web_frontend && npm install && npm run build && cd ..

# 2. Build the Go backend with the prod tag
go build -tags prod -o web_backend/web_backend ./web_backend

# 3. Run — serves dist/
cd _bin && ./backend
```

Backend listens on `listen_address_http` / `listen_address_https` from `backend.conf`.

## Content Security Policy

`web_backend` emits a strict CSP on every response:

```
default-src 'self'; script-src 'self'; style-src 'self' 'nonce-…';
style-src-attr 'unsafe-inline'; connect-src 'self'; img-src 'self' data:;
font-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'
```

No `'unsafe-eval'` is needed — the Vite config aliases `vue` to the runtime-only build
(no runtime template compiler) and sets `__INTLIFY_JIT_COMPILATION__` so vue-i18n uses its
AST interpreter instead of `new Function()`. The `style-src` nonce is generated per request
by the backend for PrimeVue's runtime `<style>` injection. `style-src-attr 'unsafe-inline'`
allows the inline `style` attributes that PrimeVue sets on elements (e.g. DataTable column
widths, overlay positioning) — nonces cannot cover style attributes, so this is scoped to
`style-src-attr` only, keeping `<style>` elements nonce-gated.

By default the header is `Content-Security-Policy-Report-Only` (violations logged, not blocked).
To enforce it, set `csp_report_only = false` in `web_backend/backend.conf`.
