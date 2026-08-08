import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Application } from '@/types/models'
import {
  applicationsService,
  type ApplicationQuery,
  type SubmitApplicationPayload,
  type UpdateApplicationPayload,
} from '@/services/applications.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useApplicationsTable(): ServerTable<Application> {
  return useServerTable<Application>({
    queryKey: ['applications'],
    fetcher: (params) => applicationsService.list(params as ApplicationQuery),
    initialSortBy: 'created_at',
  })
}

export function useApplicationMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['applications'] })

  const submitApplication = useMutation({
    mutationFn: (payload: SubmitApplicationPayload) => applicationsService.submit(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Application submitted', 'It has been routed into the workflow.')
    },
    onError: (error: unknown) =>
      toast.error('Could not submit application', toErrorMessage(error)),
  })

  const updateApplication = useMutation({
    mutationFn: (payload: UpdateApplicationPayload) => applicationsService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Application updated')
    },
    onError: (error: unknown) =>
      toast.error('Could not update application', toErrorMessage(error)),
  })

  const deleteApplication = useMutation({
    mutationFn: (applicationId: string) => applicationsService.remove(applicationId),
    onSuccess: () => {
      invalidate()
      toast.success('Application deleted')
    },
    onError: (error: unknown) =>
      toast.error('Could not delete application', toErrorMessage(error)),
  })

  return { submitApplication, updateApplication, deleteApplication }
}
