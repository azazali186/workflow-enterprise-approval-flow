import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Template } from '@/types/models'
import {
  templatesService,
  type TemplatePayload,
  type TemplateQuery,
  type UpdateTemplatePayload,
} from '@/services/templates.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useTemplatesTable(): ServerTable<Template> {
  return useServerTable<Template>({
    queryKey: ['templates'],
    fetcher: (params) => templatesService.list(params as TemplateQuery),
    initialSortBy: 'created_at',
  })
}

export function useTemplateMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['templates'] })

  const createTemplate = useMutation({
    mutationFn: (payload: TemplatePayload) => templatesService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Template created')
    },
    onError: (error: unknown) => toast.error('Could not create template', toErrorMessage(error)),
  })

  const updateTemplate = useMutation({
    mutationFn: (payload: UpdateTemplatePayload) => templatesService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Template updated')
    },
    onError: (error: unknown) => toast.error('Could not update template', toErrorMessage(error)),
  })

  const deleteTemplate = useMutation({
    mutationFn: (templateId: string) => templatesService.remove(templateId),
    onSuccess: () => {
      invalidate()
      toast.success('Template deleted')
    },
    onError: (error: unknown) => toast.error('Could not delete template', toErrorMessage(error)),
  })

  return { createTemplate, updateTemplate, deleteTemplate }
}
