import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Escalation } from '@/types/models'

export interface EscalationQuery extends ListQuery {
  approval_id?: string
}

export interface CreateEscalationPayload {
  approval_id: string
  level: number
  escalated_to: string
  reason: string
}

export const escalationsService = {
  list(params: EscalationQuery = {}) {
    return postList<Escalation>('/escalations', params)
  },

  get(escalationId: string): Promise<Escalation> {
    return post<Escalation>('/escalations/get', { escalation_id: escalationId })
  },

  create(payload: CreateEscalationPayload): Promise<Escalation> {
    return post<Escalation>('/escalations/create', payload)
  },

  resolve(escalationId: string): Promise<Escalation> {
    return post<Escalation>('/escalations/resolve', { escalation_id: escalationId })
  },
}
