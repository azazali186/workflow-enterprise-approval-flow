import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { User } from '@/types/models'
import { usersService, type UpdateUserPayload, type UserQuery } from '@/services/users.service'
import { authService, type RegisterPayload } from '@/services/auth.service'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useUsersTable(): ServerTable<User> {
  return useServerTable<User>({
    queryKey: ['users'],
    fetcher: (params) => usersService.list(params as UserQuery),
    initialSortBy: 'created_at',
  })
}

export function useUserMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['users'] })

  const createUser = useMutation({
    mutationFn: (payload: RegisterPayload) => authService.register(payload),
    onSuccess: () => {
      invalidate()
      toast.success('User created', 'The user can now sign in.')
    },
    onError: (error: unknown) => toast.error('Could not create user', toErrorMessage(error)),
  })

  const updateUser = useMutation({
    mutationFn: (payload: UpdateUserPayload) => usersService.update(payload),
    onSuccess: () => {
      invalidate()
      toast.success('User updated')
    },
    onError: (error: unknown) => toast.error('Could not update user', toErrorMessage(error)),
  })

  const deleteUser = useMutation({
    mutationFn: (userId: string) => usersService.remove(userId),
    onSuccess: () => {
      invalidate()
      toast.success('User deleted')
    },
    onError: (error: unknown) => toast.error('Could not delete user', toErrorMessage(error)),
  })

  const assignRole = useMutation({
    mutationFn: ({ userId, roleId }: { userId: string; roleId: string }) =>
      usersService.assignRole(userId, roleId),
    onSuccess: () => {
      invalidate()
      toast.success('Role assigned')
    },
    onError: (error: unknown) => toast.error('Could not assign role', toErrorMessage(error)),
  })

  const removeRole = useMutation({
    mutationFn: ({ userId, roleId }: { userId: string; roleId: string }) =>
      usersService.removeRole(userId, roleId),
    onSuccess: () => {
      invalidate()
      toast.success('Role removed')
    },
    onError: (error: unknown) => toast.error('Could not remove role', toErrorMessage(error)),
  })

  return { createUser, updateUser, deleteUser, assignRole, removeRole }
}
