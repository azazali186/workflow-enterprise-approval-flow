import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Approval } from '@/types/models'
import {
  approvalsService,
  type ApprovalQuery,
  type CreateApprovalPayload,
  type DecideApprovalPayload,
} from '@/services/approvals.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useApprovalsTable(): ServerTable<Approval> {
  return useServerTable<Approval>({
    queryKey: ['approvals'],
    fetcher: (params) => approvalsService.list(params as ApprovalQuery),
    initialSortBy: 'created_at',
  })
}

export function useApprovalMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['approvals'] })

  const createApproval = useMutation({
    mutationFn: (payload: CreateApprovalPayload) => approvalsService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Approval created')
    },
    onError: (error: unknown) => toast.error('Could not create approval', toErrorMessage(error)),
  })

  const decideApproval = useMutation({
    mutationFn: (payload: DecideApprovalPayload) => approvalsService.decide(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Decision recorded')
    },
    onError: (error: unknown) => toast.error('Could not record decision', toErrorMessage(error)),
  })

  const deleteApproval = useMutation({
    mutationFn: (approvalId: string) => approvalsService.remove(approvalId),
    onSuccess: () => {
      invalidate()
      toast.success('Approval deleted')
    },
    onError: (error: unknown) => toast.error('Could not delete approval', toErrorMessage(error)),
  })

  return { createApproval, decideApproval, deleteApproval }
}
