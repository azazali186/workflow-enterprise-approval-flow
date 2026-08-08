import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Template } from '@/types/models'

export interface TemplateQuery extends ListQuery {
  category?: string
}

export interface TemplatePayload {
  name: string
  category: string
  schema?: Record<string, unknown>
  ui?: Record<string, unknown>
}

export interface UpdateTemplatePayload extends Partial<TemplatePayload> {
  template_id: string
}

export const templatesService = {
  list(params: TemplateQuery = {}) {
    return postList<Template>('/templates', params)
  },

  get(templateId: string): Promise<Template> {
    return post<Template>('/templates/get', { template_id: templateId })
  },

  create(payload: TemplatePayload): Promise<Template> {
    return post<Template>('/templates/create', payload)
  },

  update(payload: UpdateTemplatePayload): Promise<Template> {
    return post<Template>('/templates/update', payload)
  },

  remove(templateId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/templates/delete', { template_id: templateId })
  },
}
