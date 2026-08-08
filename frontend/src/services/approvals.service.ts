import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Approval } from '@/types/models'

export interface ApprovalQuery extends ListQuery {
  status?: string
  decision?: string
  approver_id?: string
  application_id?: string
}

export interface CreateApprovalPayload {
  application_id: string
  workflow_step_id: string
  approver_id: string
}

export interface DecideApprovalPayload {
  approval_id: string
  decision: 'approved' | 'rejected' | 'escalated'
  comment?: string
}

export const approvalsService = {
  list(params: ApprovalQuery) {
    return postList<Approval>('/approvals', params)
  },

  get(approvalId: string): Promise<Approval> {
    return post<Approval>('/approvals/get', { approval_id: approvalId })
  },

  create(payload: CreateApprovalPayload): Promise<Approval> {
    return post<Approval>('/approvals/create', payload)
  },

  decide(payload: DecideApprovalPayload): Promise<Approval> {
    return post<Approval>('/approvals/decide', payload)
  },

  update(payload: { approval_id: string; status?: string; comment?: string }): Promise<Approval> {
    return post<Approval>('/approvals/update', payload)
  },

  remove(approvalId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/approvals/delete', { approval_id: approvalId })
  },

  pending(params: { approver_id: string } & ListQuery) {
    return postList<Approval>('/approvals/pending', params)
  },
}
