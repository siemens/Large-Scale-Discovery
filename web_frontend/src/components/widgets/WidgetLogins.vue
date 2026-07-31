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
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import LayoutLoading from '@/components/layout/LayoutLoading.vue'
import { formatDateTime } from '@/i18n/format'
import { userAvatarDataUri } from '@/lib/avatar'
import {
  APP_PRIMARY_500,
  APP_SECONDARY_200,
  APP_SECONDARY_500,
  APP_SECONDARY_700,
  APP_SECONDARY_800,
  APP_SECONDARY_900,
} from '@/lib/colors'

interface LoginUser {
  email: string
  name: string
  surname: string
  company: string | null
  department?: string | null
  gender?: string | null
  last_login?: string | null
}

// Local English defaults keep this module independent from the global language catalog
const localMessages = {
  'en-US': {
    recentLogins: {
      ago: {
        aMonth: 'a month ago',
        aWeek: 'a week ago',
        days: '{count} days ago',
        longAgo: 'long ago',
        today: 'today',
        weeks: 'weeks ago',
        yesterday: 'yesterday',
      },
      loading: 'Loading logins...',
      empty: 'No logins recorded yet.',
      ariaLabel: 'Logins',
      lastSeen: 'Last seen',
    },
  },
}

const props = withDefaults(
  defineProps<{
    users: LoginUser[]
    currentEmail?: string
    loading?: boolean
    avatarColor?: string
  }>(),
  { avatarColor: APP_PRIMARY_500 },
)
const { t } = useI18n({ inheritLocale: true, messages: localMessages })

const recent = computed(() =>
  [...props.users]
    .filter((user) => user.email !== props.currentEmail && validDate(user.last_login))
    .sort((a, b) => new Date(b.last_login ?? 0).getTime() - new Date(a.last_login ?? 0).getTime())
    .slice(0, 16),
)

function validDate(value?: string | null): value is string {
  return typeof value === 'string' && !Number.isNaN(new Date(value).getTime())
}

function userDescription(user: LoginUser): string {
  if (user.company && user.department) return `${user.company} (${user.department})`
  return user.company || user.department || ''
}

function userFullName(user: LoginUser): string {
  return [user.name, user.surname].filter(Boolean).join(' ')
}

function lastLoginStatus(lastLogin?: string | null): { color: string; text: string } {
  if (!lastLogin) return { color: APP_SECONDARY_200, text: t('recentLogins.ago.longAgo') }
  const midnight = new Date(lastLogin)
  midnight.setHours(0, 0, 0, 0)
  const days = (Date.now() - midnight.getTime()) / 86_400_000
  if (days >= 62) return { color: APP_SECONDARY_200, text: t('recentLogins.ago.longAgo') }
  if (days >= 31) return { color: APP_SECONDARY_500, text: t('recentLogins.ago.aMonth') }
  if (days >= 14) return { color: APP_SECONDARY_500, text: t('recentLogins.ago.weeks') }
  if (days >= 7) return { color: APP_SECONDARY_700, text: t('recentLogins.ago.aWeek') }
  if (days >= 2)
    return {
      color: APP_SECONDARY_800,
      text: t('recentLogins.ago.days', { count: Math.round(days) }),
    }
  if (days > 1) return { color: APP_SECONDARY_900, text: t('recentLogins.ago.yesterday') }
  return { color: 'inherit', text: t('recentLogins.ago.today') }
}
</script>

<template>
  <LayoutLoading
    :loading="loading"
    :loading-message="t('recentLogins.loading')"
    :empty="recent.length === 0"
    :empty-message="t('recentLogins.empty')"
    empty-test-id="logins-empty"
  >
    <div
      class="widget-logins__list"
      data-testid="logins-list"
      :style="{ '--widget-logins-avatar-color': avatarColor }"
      role="list"
      :aria-label="t('recentLogins.ariaLabel')"
    >
      <div
        v-for="user in recent"
        :key="user.email"
        class="widget-logins__item"
        :data-testid="`logins-user-${user.email}`"
        role="listitem"
      >
        <div
          v-tooltip.top="userDescription(user) || user.email"
          class="widget-logins__avatar"
          :aria-label="userFullName(user) || user.email"
          role="img"
          data-testid="logins-avatar"
        >
          <img
            class="widget-logins__avatar-image"
            :src="userAvatarDataUri(user.email, user.gender, avatarColor)"
            alt=""
            aria-hidden="true"
          />
        </div>

        <div class="widget-logins__info" :style="{ color: lastLoginStatus(user.last_login).color }">
          <div
            v-if="userFullName(user)"
            v-tooltip.top="userDescription(user)"
            class="widget-logins__name sensitive truncate"
            data-testid="logins-user-name"
          >
            <strong>{{ userFullName(user) }}</strong>
          </div>
          <div
            v-tooltip.top="user.email"
            class="widget-logins__email sensitive truncate"
            :class="{ 'widget-logins__email--standalone': !userFullName(user) }"
          >
            {{ user.email }}
          </div>
          <div
            v-tooltip.bottom="formatDateTime(user.last_login)"
            class="widget-logins__last-seen truncate"
          >
            {{ t('recentLogins.lastSeen') }}
            <strong>{{ lastLoginStatus(user.last_login).text }}</strong>
          </div>
        </div>
      </div>
    </div>
  </LayoutLoading>
</template>

<style scoped>
.widget-logins__list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(14.285714rem, 1fr));
  gap: 0.75rem 1rem;
}

.widget-logins__item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.widget-logins__avatar {
  position: relative;
  flex-shrink: 0;
  align-self: center;
  width: 3.214286rem;
  height: 3.214286rem;
  cursor: default;
}

.widget-logins__avatar::before {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: var(--widget-logins-avatar-color, var(--p-primary-color));
  content: '';
}

.widget-logins__avatar-image {
  position: absolute;
  top: 43%;
  left: 50%;
  display: block;
  width: 140%;
  height: 140%;
  transform: translate(-50%, -50%);
  clip-path: inset(-1.428571rem 0 0 0);
}

.widget-logins__info {
  min-width: 0;
  overflow: hidden;
}

.widget-logins__email--standalone {
  font-weight: 700;
}
</style>
