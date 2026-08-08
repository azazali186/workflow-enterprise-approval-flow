import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Workflow } from '@/types/models'
import {
  workflowsService,
  type UpdateWorkflowPayload,
  type WorkflowPayload,
  type WorkflowQuery,
} from '@/services/workflows.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useWorkflowsTable(): ServerTable<Workflow> {
  return useServerTable<Workflow>({
    queryKey: ['workflows'],
    fetcher: (params) => workflowsService.list(params as WorkflowQuery),
    initialSortBy: 'created_at',
  })
}

export function useWorkflowMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['workflows'] })

  const createWorkflow = useMutation({
    mutationFn: (payload: WorkflowPayload) => workflowsService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Workflow created')
    },
    onError: (error: unknown) => toast.error('Could not create workflow', toErrorMessage(error)),
  })

  const updateWorkflow = useMutation({
    mutationFn: (payload: UpdateWorkflowPayload) => workflowsService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Workflow updated')
    },
    onError: (error: unknown) => toast.error('Could not update workflow', toErrorMessage(error)),
  })

  const deleteWorkflow = useMutation({
    mutationFn: (workflowId: string) => workflowsService.remove(workflowId),
    onSuccess: () => {
      invalidate()
      toast.success('Workflow deleted')
    },
    onError: (error: unknown) => toast.error('Could not delete workflow', toErrorMessage(error)),
  })

  return { createWorkflow, updateWorkflow, deleteWorkflow }
}
