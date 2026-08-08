import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Application } from '@/types/models'

export interface ApplicationQuery extends ListQuery {
  status?: string
  priority?: string
  applicant_id?: string
  workflow_id?: string
}

export interface SubmitApplicationPayload {
  applicant_id: string
  workflow_id: string
  template_id: string
  title: string
  description?: string
  priority: string
  data?: Record<string, unknown>
}

export interface UpdateApplicationPayload {
  application_id: string
  status?: string
  priority?: string
  data?: Record<string, unknown>
}

export const applicationsService = {
  list(params: ApplicationQuery) {
    return postList<Application>('/applications', params)
  },

  get(applicationId: string): Promise<Application> {
    return post<Application>('/applications/get', { application_id: applicationId })
  },

  submit(payload: SubmitApplicationPayload): Promise<Application> {
    return post<Application>('/applications/submit', payload)
  },

  update(payload: UpdateApplicationPayload): Promise<Application> {
    return post<Application>('/applications/update', payload)
  },

  remove(applicationId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/applications/delete', { application_id: applicationId })
  },
}
