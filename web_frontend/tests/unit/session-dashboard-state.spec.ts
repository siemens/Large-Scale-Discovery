import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'

import { useListFetcher } from '@/composables/useAsyncState'
import { useAgentsStore } from '@/stores/agents'
import { useAuthStore } from '@/stores/auth'
import { useViewsStore } from '@/stores/views'

import type { AgentGroupResponse } from '@/types/agents'
import type { UserProfile } from '@/types/auth'
import type { GrantedViewResponse } from '@/types/views'

const firstUser: UserProfile = {
  id: '1',
  fullName: 'First User',
  email: 'first@example.test',
  role: 'user',
}

describe('session-scoped dashboard state', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('clears grants and agents when the authenticated account changes', () => {
    const authStore = useAuthStore()
    const viewsStore = useViewsStore()
    const agentsStore = useAgentsStore()

    authStore.setProfile(firstUser)
    viewsStore.grantedViews = [{} as GrantedViewResponse]
    agentsStore.agentGroups = [{} as AgentGroupResponse]

    authStore.setProfile({ ...firstUser, id: '2', email: 'second@example.test' })

    expect(viewsStore.grantedViews).toBeNull()
    expect(agentsStore.agentGroups).toBeNull()
  })

  it('clears grants and agents on logout', () => {
    const authStore = useAuthStore()
    const viewsStore = useViewsStore()
    const agentsStore = useAgentsStore()

    authStore.setProfile(firstUser)
    viewsStore.grantedViews = [{} as GrantedViewResponse]
    agentsStore.agentGroups = [{} as AgentGroupResponse]

    authStore.logout()

    expect(viewsStore.grantedViews).toBeNull()
    expect(agentsStore.agentGroups).toBeNull()
  })

  it('does not publish a response that finishes after a reset', async () => {
    let resolveRequest!: (value: string[]) => void
    const request = new Promise<string[]>((resolve) => {
      resolveRequest = resolve
    })
    const listFetcher = useListFetcher(() => request, 'grantedPanel.error')

    const fetchPromise = listFetcher.fetch()
    listFetcher.reset()
    resolveRequest(['previous user'])
    await fetchPromise

    expect(listFetcher.data.value).toBeNull()
  })
})
