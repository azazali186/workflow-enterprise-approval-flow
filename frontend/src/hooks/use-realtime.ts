import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useAppSelector } from '@/store/hooks'
import { realtime, type RealtimeEvent } from '@/services/realtime'

/**
 * Keep the single shared socket connected while the user is authenticated and
 * invalidate the React Query caches that backend events affect, so tables and
 * the notification bell update in real time instead of on their poll interval.
 *
 * Event → cache mapping mirrors the backend saga broadcasts:
 *   - approval_needed / approval_created → pending approvals, notifications
 *   - decision_made                  → approvals, applications, notifications
 *   - application_submitted          → applications, approvals
 *   - escalation_trigger             → escalations, approvals, notifications
 *   - notification                   → notifications
 */
export function useRealtime(): void {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated)
  // queryClient is a stable reference from the provider, safe to capture.
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!isAuthenticated) {
      realtime.disconnect()
      return
    }

    realtime.connect()

    const unsubscribe = realtime.subscribe((event: RealtimeEvent) => {
      switch (event.event) {
        case 'approval_needed':
        case 'approval_created':
          queryClient.invalidateQueries({ queryKey: ['approvals', 'pending'] })
          queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
          break
        case 'decision_made':
          queryClient.invalidateQueries({ queryKey: ['approvals'] })
          queryClient.invalidateQueries({ queryKey: ['applications'] })
          queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
          break
        case 'application_submitted':
          queryClient.invalidateQueries({ queryKey: ['applications'] })
          queryClient.invalidateQueries({ queryKey: ['approvals'] })
          break
        case 'escalation_trigger':
          queryClient.invalidateQueries({ queryKey: ['escalations'] })
          queryClient.invalidateQueries({ queryKey: ['approvals'] })
          queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
          break
        case 'notification':
          queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
          break
        default:
          break
      }
    })

    // A background tab's socket can be killed by the browser or a proxy idle
    // timeout; retry immediately when the tab becomes visible again.
    const onVisibility = () => {
      if (document.visibilityState === 'visible') realtime.reconnectNow()
    }
    document.addEventListener('visibilitychange', onVisibility)

    return () => {
      document.removeEventListener('visibilitychange', onVisibility)
      unsubscribe()
      // On logout the next effect run (isAuthenticated=false) disconnects;
      // on unmount while still authenticated (unlikely), close the socket.
      if (!isAuthenticated) realtime.disconnect()
    }
  }, [isAuthenticated])
}
