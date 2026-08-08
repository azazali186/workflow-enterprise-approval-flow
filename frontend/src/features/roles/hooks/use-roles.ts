import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Role } from '@/types/models'
import {
  rolesService,
  type RolePayload,
  type UpdateRolePayload,
} from '@/services/roles.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useRolesTable(): ServerTable<Role> {
  return useServerTable<Role>({
    queryKey: ['roles'],
    fetcher: (params) => rolesService.list(params),
    initialSortBy: 'created_at',
  })
}

export function useRoleMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['roles'] })

  const createRole = useMutation({
    mutationFn: (payload: RolePayload) => rolesService.create(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Role created')
    },
    onError: (error: unknown) => toast.error('Could not create role', toErrorMessage(error)),
  })

  const updateRole = useMutation({
    mutationFn: (payload: UpdateRolePayload) => rolesService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Role updated')
    },
    onError: (error: unknown) => toast.error('Could not update role', toErrorMessage(error)),
  })

  const deleteRole = useMutation({
    mutationFn: (roleId: string) => rolesService.remove(roleId),
    onSuccess: () => {
      invalidate()
      toast.success('Role deleted')
    },
    onError: (error: unknown) => toast.error('Could not delete role', toErrorMessage(error)),
  })

  const assignPermission = useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      rolesService.assignPermission(roleId, permissionId),
    onSuccess: () => {
      invalidate()
      toast.success('Permission assigned')
    },
    onError: (error: unknown) =>
      toast.error('Could not assign permission', toErrorMessage(error)),
  })

  const removePermission = useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      rolesService.removePermission(roleId, permissionId),
    onSuccess: () => {
      invalidate()
      toast.success('Permission removed')
    },
    onError: (error: unknown) =>
      toast.error('Could not remove permission', toErrorMessage(error)),
  })

  return { createRole, updateRole, deleteRole, assignPermission, removePermission }
}
