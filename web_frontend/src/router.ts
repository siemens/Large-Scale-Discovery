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

import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

import { hasValidSession } from '@/api/session'
import { savePendingRedirect } from '@/lib/redirect'
import { useAuthStore } from '@/stores/auth'

// Redirect-only parent records intentionally have neither names nor components.
// Authentication is required by default and elevated routes declare role metadata
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/home',
  },
  {
    path: '/home',
    name: 'home',
    component: () => import('@/pages/Home.vue'),
  },
  {
    path: '/installation',
    name: 'installation',
    component: () => import('@/pages/Installation.vue'),
    meta: {
      owner: true,
    },
  },
  {
    path: '/api',
    name: 'api',
    component: () => import('@/pages/Api.vue'),
    meta: {
      owner: true,
    },
  },
  {
    path: '/configuration',
    redirect: '/configuration/scopes',
  },
  {
    path: '/configuration/scopes',
    name: 'configuration-scopes',
    component: () => import('@/pages/configuration/Scopes.vue'),
    meta: {
      owner: true,
    },
  },
  {
    path: '/configuration/views',
    name: 'configuration-views',
    component: () => import('@/pages/configuration/Views.vue'),
    meta: {
      owner: true,
    },
  },
  {
    path: '/admin',
    redirect: '/admin/users',
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: () => import('@/pages/administration/Users.vue'),
    meta: {
      admin: true,
    },
  },
  {
    path: '/admin/groups',
    name: 'admin-groups',
    component: () => import('@/pages/administration/Groups.vue'),
    meta: {
      admin: true,
    },
  },
  {
    path: '/admin/databases',
    name: 'admin-databases',
    component: () => import('@/pages/administration/Databases.vue'),
    meta: {
      admin: true,
    },
  },
  {
    path: '/admin/logs',
    name: 'admin-logs',
    component: () => import('@/pages/administration/Logs.vue'),
    meta: {
      admin: true,
    },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/Login.vue'),
    meta: {
      public: true,
      auth: true,
    },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/pages/Register.vue'),
    meta: {
      public: true,
      auth: true,
    },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: {
      render: () => null,
    },
  },
]

/** router is the app-wide vue-router instance using HTML5 history mode. */
export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth-by-default: every route requires a session unless explicitly marked
// `meta.public`. Auth pages (login/register) additionally bounce authenticated
// users back to /home.
router.beforeEach(async (to) => {
  // Resolve the auth store fresh per navigation to read current session state
  const authStore = useAuthStore()

  // Unknown URLs are navigation errors, not authentication failures. Preserve
  // a valid session and send the user to the safest useful destination.
  if (to.name === 'not-found') {
    return authStore.isAuthenticated || hasValidSession() ? '/home' : '/login'
  }

  // Auth pages bounce already-authenticated users back to home
  if (to.meta.auth === true && authStore.isAuthenticated) {
    return '/home'
  }

  // Public routes need no session
  if (to.meta.public) return true

  // Without a session, remember the target and redirect to login
  if (!authStore.isAuthenticated && !hasValidSession()) {
    savePendingRedirect(to.fullPath)
    return '/login'
  }

  // Authorization is enforced here, not just in the navigation UI: hidden menu
  // links are not a security control, so any user can type a protected URL.
  // Role gates require the resolved profile — on a hard refresh the session may
  // be valid while the profile is still being fetched. Await the (idempotent)
  // session restore so a valid admin/owner is never bounced mid-restore.
  if (to.meta.admin || to.meta.owner) {
    if (!authStore.isAuthenticated) {
      await authStore.restore()
    }

    // Admin-only routes (/admin/*). Owners are NOT admins and must be blocked.
    if (to.meta.admin && !authStore.isAdmin) {
      return '/home'
    }

    // Admin-or-owner routes (API, configuration, installation).
    if (to.meta.owner && !authStore.isAdmin && !authStore.isOwner) {
      return '/home'
    }
  }
  return true
})
