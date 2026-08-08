import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Escalation } from '@/types/models'
import {
  escalationsService,
  type CreateEscalationPayload,
  type EscalationQuery,
} from '@/services/escalations.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useEscalationsTable(): ServerTable<Escalation> {
  return useServerTable<Escalation>({
    queryKey: ['escalations'],
    fetcher: (params) => escalationsService.list(params as EscalationQuery),
    initialSortBy: 'escalated_at',
  })
}

export function useEscalationMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['escalations'] })

  const createEscalation = useMutation({
    mutationFn: (payload: CreateEscalationPayload) => escalationsService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Escalation created', 'The approval has been escalated.')
    },
    onError: (error: unknown) =>
      toast.error('Could not create escalation', toErrorMessage(error)),
  })

  const resolveEscalation = useMutation({
    mutationFn: (escalationId: string) => escalationsService.resolve(escalationId),
    onSuccess: () => {
      invalidate()
      toast.success('Escalation resolved')
    },
    onError: (error: unknown) =>
      toast.error('Could not resolve escalation', toErrorMessage(error)),
  })

  return { createEscalation, resolveEscalation }
}
