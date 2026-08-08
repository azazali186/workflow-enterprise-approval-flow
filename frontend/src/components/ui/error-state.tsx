import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from './button'
import { toErrorMessage } from '@/services/api/errors'

export interface ErrorStateProps {
  title?: string
  description?: string
  error?: unknown
  onRetry?: () => void
  className?: string
}

export function ErrorState({
  title = 'Something went wrong',
  description,
  error,
  onRetry,
  className,
}: ErrorStateProps) {
  const message = description ?? toErrorMessage(error, 'Please try again in a moment.')

  return (
    <div className={`flex flex-col items-center justify-center px-6 py-14 text-center ${className ?? ''}`}>
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-rose-50 text-rose-500">
        <AlertCircle className="h-6 w-6" />
      </div>
      <h3 className="mt-3 text-sm font-semibold text-slate-900">{title}</h3>
      <p className="mt-1 max-w-sm text-sm text-slate-500">{message}</p>
      {onRetry && (
        <Button variant="outline" size="sm" className="mt-4" onClick={onRetry}>
          <RefreshCw className="h-3.5 w-3.5" />
          Try again
        </Button>
      )}
    </div>
  )
}
