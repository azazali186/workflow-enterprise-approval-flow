import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Permission } from '@/types/models'

export interface PermissionPayload {
  name: string
  route: string
  path: string
  method: string
  service: string
}

export interface UpdatePermissionPayload extends Partial<PermissionPayload> {
  permission_id: string
}

export const permissionsService = {
  list(params: ListQuery = {}) {
    return postList<Permission>('/admin/permissions', params)
  },

  get(permissionId: string): Promise<Permission> {
    return post<Permission>('/admin/permissions/get', { permission_id: permissionId })
  },

  create(payload: PermissionPayload): Promise<Permission> {
    return post<Permission>('/admin/permissions/create', payload)
  },

  update(payload: UpdatePermissionPayload): Promise<Permission> {
    return post<Permission>('/admin/permissions/update', payload)
  },

  remove(permissionId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/admin/permissions/delete', { permission_id: permissionId })
  },
}
