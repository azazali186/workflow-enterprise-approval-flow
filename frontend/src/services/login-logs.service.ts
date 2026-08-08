import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { LoginLog } from '@/types/models'

export interface LoginLogQuery extends ListQuery {
  email?: string
  status?: string
  user_id?: string
  start_date?: string
  end_date?: string
}

export const loginLogsService = {
  list(params: LoginLogQuery = {}) {
    return postList<LoginLog>('/admin/login-logs', params)
  },

  history(userId: string, limit = 50) {
    return postList<LoginLog>('/auth/login-history', { user_id: userId, limit })
  },

  historyByEmail(email: string, limit = 50) {
    return postList<LoginLog>('/auth/login-history/email', { email, limit })
  },

  stats(userId: string): Promise<Record<string, unknown>> {
    return post<Record<string, unknown>>('/auth/login-stats', { user_id: userId })
  },
}
