import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Workflow } from '@/types/models'

export interface WorkflowQuery extends ListQuery {
  category?: string
  is_active?: boolean
}

export interface WorkflowStepPayload {
  name: string
  step_order?: number
  approver_role?: string
  approver_id?: string
  action?: string
  timeout_hours?: number
  is_required?: boolean
}

export interface WorkflowPayload {
  name: string
  description?: string
  category: string
  is_active?: boolean
  steps?: WorkflowStepPayload[]
}

export interface UpdateWorkflowPayload extends Partial<WorkflowPayload> {
  workflow_id: string
}

export const workflowsService = {
  list(params: WorkflowQuery = {}) {
    return postList<Workflow>('/workflows', params)
  },

  get(workflowId: string): Promise<Workflow> {
    return post<Workflow>('/workflows/get', { workflow_id: workflowId })
  },

  create(payload: WorkflowPayload): Promise<Workflow> {
    return post<Workflow>('/workflows/create', payload)
  },

  update(payload: UpdateWorkflowPayload): Promise<Workflow> {
    return post<Workflow>('/workflows/update', payload)
  },

  remove(workflowId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/workflows/delete', { workflow_id: workflowId })
  },
}
