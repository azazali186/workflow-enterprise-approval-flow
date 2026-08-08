import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Role, User } from '@/types/models'

export interface UserQuery extends ListQuery {
  status?: string
  role_id?: string
}

export interface UpdateUserPayload {
  user_id: string
  name?: string
  email?: string
  status?: string
}

export const usersService = {
  list(params: UserQuery) {
    return postList<User>('/admin/users', params)
  },

  get(userId: string): Promise<User> {
    return post<User>('/admin/users/get', { user_id: userId })
  },

  update(payload: UpdateUserPayload): Promise<User> {
    return post<User>('/admin/users/update', payload)
  },

  remove(userId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/users/delete', { user_id: userId })
  },

  roles(userId: string): Promise<Role[]> {
    return post<Role[]>('/admin/users/roles', { user_id: userId })
  },

  assignRole(userId: string, roleId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/users/roles/assign', {
      user_id: userId,
      role_id: roleId,
    })
  },

  removeRole(userId: string, roleId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/users/roles/remove', {
      user_id: userId,
      role_id: roleId,
    })
  },
}
