import { post } from './api/client'
import type { ComboboxOption } from '@/components/ui/combobox'

// Re-export for convenience
export type DropdownOption = ComboboxOption

export interface DropdownRequest {
  entities: string[]
  include_inactive?: boolean
  statuses?: string[]
  /** Scopes the workflow_steps entity to a single workflow. */
  workflow_id?: string
}

export interface DropdownResponse {
  [key: string]: DropdownOption[]
}

export const dropdownsService = {
  /**
   * Fetch dropdown options for one or more entity types.
   * Returns a map of entity type to [{id, name}] options.
   */
  list(params: DropdownRequest): Promise<DropdownResponse> {
    return post<DropdownResponse>('/dropdowns', params)
  },

  /**
   * Convenience methods for common entity types.
   */
  users(): Promise<DropdownOption[]> {
    return this.list({ entities: ['users'] }).then((res) => res.users || [])
  },

  workflows(includeInactive = false): Promise<DropdownOption[]> {
    return this.list({ entities: ['workflows'], include_inactive: includeInactive }).then(
      (res) => res.workflows || []
    )
  },

  templates(): Promise<DropdownOption[]> {
    return this.list({ entities: ['templates'] }).then((res) => res.templates || [])
  },

  roles(): Promise<DropdownOption[]> {
    return this.list({ entities: ['roles'] }).then((res) => res.roles || [])
  },

  applications(statuses: string[] = ['submitted']): Promise<DropdownOption[]> {
    return this.list({ entities: ['applications'], statuses }).then(
      (res) => res.applications || []
    )
  },

  approvals(): Promise<DropdownOption[]> {
    return this.list({ entities: ['approvals'] }).then((res) => res.approvals || [])
  },

  /**
   * Steps of a single workflow — used to keep a created approval consistent
   * with the selected application's workflow.
   */
  workflowSteps(workflowId: string): Promise<DropdownOption[]> {
    return this.list({ entities: ['workflow_steps'], workflow_id: workflowId }).then(
      (res) => res.workflow_steps || []
    )
  },

  /**
   * Fetch multiple entity types in a single request.
   */
  multiple(entities: string[], options?: { include_inactive?: boolean; statuses?: string[] }): Promise<DropdownResponse> {
    return this.list({ entities, ...options })
  },
}
