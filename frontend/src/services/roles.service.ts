import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Permission, Role } from '@/types/models'

export interface RolePayload {
  name: string
  description?: string
  is_default?: boolean
}

export interface UpdateRolePayload extends RolePayload {
  role_id: string
}

export const rolesService = {
  list(params: ListQuery = {}) {
    return postList<Role>('/admin/roles', params)
  },

  get(roleId: string): Promise<Role> {
    return post<Role>('/admin/roles/get', { role_id: roleId })
  },

  create(payload: RolePayload): Promise<Role> {
    return post<Role>('/admin/roles/create', payload)
  },

  update(payload: UpdateRolePayload): Promise<Role> {
    return post<Role>('/admin/roles/update', payload)
  },

  remove(roleId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/roles/delete', { role_id: roleId })
  },

  permissions(roleId: string): Promise<Permission[]> {
    return post<Permission[]>('/admin/roles/permissions', { role_id: roleId })
  },

  assignPermission(roleId: string, permissionId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/roles/permissions/assign', {
      role_id: roleId,
      permission_id: permissionId,
    })
  },

  removePermission(roleId: string, permissionId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/roles/permissions/remove', {
      role_id: roleId,
      permission_id: permissionId,
    })
  },
}
