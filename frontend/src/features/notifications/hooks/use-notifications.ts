import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useServerTable } from '@/components/data-table/use-server-table'
import type { ServerTable } from '@/components/data-table/types'
import type { Notification } from '@/types/models'
import {
  notificationsService,
  type NotificationQuery,
  type SendNotificationPayload,
} from '@/services/notifications.service'
import { useAppSelector } from '@/store/hooks'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

export function useNotificationsTable(): ServerTable<Notification> {
  const user = useAppSelector((state) => state.auth.user)
  return useServerTable<Notification>({
    queryKey: ['notifications', user?.id],
    fetcher: (params) =>
      notificationsService.list({
        ...(params as unknown as NotificationQuery),
        user_id: user?.id ?? '',
      }),
    initialSortBy: 'created_at',
    enabled: Boolean(user),
  })
}

export function useNotificationMutations() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['notifications'] })

  const sendNotification = useMutation({
    mutationFn: (payload: SendNotificationPayload) => notificationsService.send(payload),
    onSuccess: () => {
      invalidate()
      toast.success('Notification sent')
    },
    onError: (error: unknown) =>
      toast.error('Could not send notification', toErrorMessage(error)),
  })

  const markRead = useMutation({
    mutationFn: (notificationId: string) => notificationsService.markRead(notificationId),
    onSuccess: () => {
      invalidate()
    },
    onError: (error: unknown) =>
      toast.error('Could not update notification', toErrorMessage(error)),
  })

  return { sendNotification, markRead }
}
