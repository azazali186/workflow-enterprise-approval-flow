import { useCallback, useMemo } from 'react'
import { useAppDispatch } from '@/store/hooks'
import { pushToast, dismissToast } from '@/store/slices/toast.slice'
import type { ToastType } from '@/store/slices/toast.slice'

export interface ToastApi {
  success: (title: string, message?: string) => void
  error: (title: string, message?: string) => void
  info: (title: string, message?: string) => void
  warning: (title: string, message?: string) => void
  dismiss: (id: string) => void
}

export function useToast(): ToastApi {
  const dispatch = useAppDispatch()

  const show = useCallback(
    (type: ToastType, title: string, message?: string) => {
      dispatch(pushToast({ type, title, message }))
    },
    [dispatch],
  )

  return useMemo(
    () => ({
      success: (title: string, message?: string) => show('success', title, message),
      error: (title: string, message?: string) => show('error', title, message),
      info: (title: string, message?: string) => show('info', title, message),
      warning: (title: string, message?: string) => show('warning', title, message),
      dismiss: (id: string) => dispatch(dismissToast(id)),
    }),
    [dispatch, show],
  )
}
