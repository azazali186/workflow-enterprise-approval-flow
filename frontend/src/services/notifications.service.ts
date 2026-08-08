import { post, postList } from './api/client'
import type { ListQuery } from '@/types/api'
import type { Notification } from '@/types/models'

export interface NotificationQuery extends ListQuery {
  user_id: string
  type?: string
  channel?: string
  is_read?: boolean
}

export interface SendNotificationPayload {
  user_id: string
  type: string
  channel: string
  title: string
  body: string
  data?: Record<string, unknown>
}

export const notificationsService = {
  list(params: NotificationQuery) {
    return postList<Notification>('/notifications', params)
  },

  unread(params: { user_id: string } & ListQuery) {
    return postList<Notification>('/notifications/unread', params)
  },

  send(payload: SendNotificationPayload): Promise<Notification> {
    return post<Notification>('/notifications/send', payload)
  },

  markRead(notificationId: string): Promise<Notification> {
    return post<Notification>('/notifications/read', { notification_id: notificationId })
  },

  stats(userId: string): Promise<Record<string, unknown>> {
    return post<Record<string, unknown>>('/notifications/stats', { user_id: userId })
  },
}
