import { useEffect } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckCircle2, AlertCircle, Info, AlertTriangle, X } from 'lucide-react'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { dismissToast } from '@/store/slices/toast.slice'
import type { ToastType } from '@/store/slices/toast.slice'
import { cn } from '@/utils/cn'

const typeStyles: Record<ToastType, { icon: typeof Info; accent: string }> = {
  success: { icon: CheckCircle2, accent: 'text-emerald-500' },
  error: { icon: AlertCircle, accent: 'text-rose-500' },
  info: { icon: Info, accent: 'text-sky-500' },
  warning: { icon: AlertTriangle, accent: 'text-amber-500' },
}

const durations: Record<ToastType, number> = {
  success: 4000,
  error: 7000,
  info: 4500,
  warning: 5500,
}

function ToastCard({ id, type, title, message }: { id: string; type: ToastType; title: string; message?: string }) {
  const dispatch = useAppDispatch()
  const { icon: Icon, accent } = typeStyles[type]

  useEffect(() => {
    const timer = setTimeout(() => dispatch(dismissToast(id)), durations[type])
    return () => clearTimeout(timer)
  }, [dispatch, id, type])

  return (
    <motion.div
      layout
      initial={{ opacity: 0, x: 32, scale: 0.96 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 24, scale: 0.96 }}
      transition={{ type: 'spring', stiffness: 380, damping: 30 }}
      className="pointer-events-auto flex w-80 max-w-[calc(100vw-2rem)] items-start gap-3 rounded-xl border border-slate-200/80 bg-white p-3.5 shadow-popover"
      role="status"
    >
      <Icon className={cn('mt-0.5 h-5 w-5 shrink-0', accent)} />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold text-slate-900">{title}</p>
        {message && <p className="mt-0.5 text-xs leading-relaxed text-slate-500">{message}</p>}
      </div>
      <button
        type="button"
        onClick={() => dispatch(dismissToast(id))}
        className="rounded-md p-1 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
        aria-label="Dismiss notification"
      >
        <X className="h-3.5 w-3.5" />
      </button>
    </motion.div>
  )
}

export function ToastViewport() {
  const toasts = useAppSelector((state) => state.toast.toasts)

  return (
    <div className="pointer-events-none fixed right-4 top-4 z-[60] flex flex-col gap-2">
      <AnimatePresence>
        {toasts.map((toast) => (
          <ToastCard key={toast.id} {...toast} />
        ))}
      </AnimatePresence>
    </div>
  )
}
