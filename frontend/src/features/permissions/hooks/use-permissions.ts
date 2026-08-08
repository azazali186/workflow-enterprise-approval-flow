import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Permission } from '@/types/models'
import {
  permissionsService,
  type PermissionPayload,
  type UpdatePermissionPayload,
} from '@/services/permissions.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function usePermissionsTable(): ServerTable<Permission> {
  return useServerTable<Permission>({
    queryKey: ['permissions'],
    fetcher: (params) => permissionsService.list(params),
    initialSortBy: 'created_at',
  })
}

export function usePermissionMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['permissions'] })

  const createPermission = useMutation({
    mutationFn: (payload: PermissionPayload) => permissionsService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Permission created')
    },
    onError: (error: unknown) =>
      toast.error('Could not create permission', toErrorMessage(error)),
  })

  const updatePermission = useMutation({
    mutationFn: (payload: UpdatePermissionPayload) => permissionsService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Permission updated')
    },
    onError: (error: unknown) =>
      toast.error('Could not update permission', toErrorMessage(error)),
  })

  const deletePermission = useMutation({
    mutationFn: (permissionId: string) => permissionsService.remove(permissionId),
    onSuccess: () => {
      invalidate()
      toast.success('Permission deleted')
    },
    onError: (error: unknown) =>
      toast.error('Could not delete permission', toErrorMessage(error)),
  })

  return { createPermission, updatePermission, deletePermission }
}
