import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { AnimatePresence, motion } from 'framer-motion'
import { Bell, BellOff, Inbox } from 'lucide-react'
import { notificationsService } from '@/services/notifications.service'
import { useAppSelector } from '@/store/hooks'
import { useToast } from '@/hooks/use-toast'
import { relativeTime, titleCase } from '@/utils/format'
import { cn } from '@/utils/cn'

export function NotificationBell() {
  const user = useAppSelector((state) => state.auth.user)
  const navigate = useNavigate()
  const toast = useToast()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', 'unread', user?.id],
    queryFn: () => notificationsService.unread({ user_id: user!.id, limit: 6 }),
    enabled: Boolean(user),
    refetchInterval: 60_000,
  })

  const unreadCount = data?.meta.total ?? 0
  const items = data?.rows ?? []

  const markRead = async (id: string) => {
    try {
      await notificationsService.markRead(id)
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    } catch {
      toast.error('Failed to update notification')
    }
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="relative rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
        aria-label={`Notifications${unreadCount ? `, ${unreadCount} unread` : ''}`}
      >
        <Bell className="h-[18px] w-[18px]" />
        {unreadCount > 0 && (
          <span className="absolute right-1 top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-rose-500 px-1 text-[9px] font-bold text-white shadow-sm">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      <AnimatePresence>
        {open && (
          <>
            <div className="fixed inset-0 z-30" onClick={() => setOpen(false)} aria-hidden />
            <motion.div
              initial={{ opacity: 0, y: 8, scale: 0.97 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: 6, scale: 0.97 }}
              transition={{ duration: 0.15 }}
              className="absolute right-0 top-full z-40 mt-2 w-80 overflow-hidden rounded-xl border border-slate-200/80 bg-white shadow-popover"
            >
              <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
                <p className="text-sm font-semibold text-slate-900">Notifications</p>
                {unreadCount > 0 && (
                  <span className="rounded-full bg-primary-50 px-2 py-0.5 text-[11px] font-medium text-primary-700">
                    {unreadCount} unread
                  </span>
                )}
              </div>

              <div className="scrollbar-thin max-h-80 overflow-y-auto">
                {isLoading ? (
                  <div className="space-y-2 p-4">
                    {[0, 1, 2].map((index) => (
                      <div key={index} className="skeleton h-12 w-full" />
                    ))}
                  </div>
                ) : items.length === 0 ? (
                  <div className="flex flex-col items-center px-4 py-8 text-center">
                    <BellOff className="h-6 w-6 text-slate-300" />
                    <p className="mt-2 text-sm font-medium text-slate-600">You're all caught up</p>
                    <p className="mt-0.5 text-xs text-slate-400">No unread notifications.</p>
                  </div>
                ) : (
                  items.map((item) => (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => markRead(item.id)}
                      className="flex w-full items-start gap-3 border-b border-slate-50 px-4 py-3 text-left transition-colors hover:bg-slate-50"
                    >
                      <span className={cn('mt-1.5 h-2 w-2 shrink-0 rounded-full', item.read_at ? 'bg-slate-200' : 'bg-primary-500')} />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium text-slate-800">
                          {item.title}
                        </span>
                        <span className="mt-0.5 block truncate text-xs text-slate-500">
                          {item.body}
                        </span>
                        <span className="mt-1 flex items-center gap-2 text-[11px] text-slate-400">
                          <span className="rounded bg-slate-100 px-1.5 py-px font-medium text-slate-500">
                            {titleCase(item.type)}
                          </span>
                          {relativeTime(item.sent_at ?? item.created_at)}
                        </span>
                      </span>
                    </button>
                  ))
                )}
              </div>

              <button
                type="button"
                onClick={() => {
                  setOpen(false)
                  navigate('/notifications')
                }}
                className="flex w-full items-center justify-center gap-1.5 border-t border-slate-100 px-4 py-2.5 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-50"
              >
                <Inbox className="h-3.5 w-3.5" />
                View all notifications
              </button>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  )
}
