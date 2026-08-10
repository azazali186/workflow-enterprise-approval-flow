import { AlertTriangle } from 'lucide-react'
import { Button } from './button'

/**
 * Full-screen recoverable error fallback. Used in two places:
 *  - the root <ErrorBoundary> in main.tsx (provider/render-level errors), and
 *  - React Router's route errorElement (page crashes — the data router
 *    renders errorElement internally and never propagates to outer React
 *    error boundaries, so a page throwing must be caught here).
 *
 * The message is intentionally generic: internal details stay in the console
 * and server logs, never on screen.
 */
export function AppErrorFallback() {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-4 bg-slate-50 p-6 text-center">
      <AlertTriangle className="h-10 w-10 text-rose-500" />
      <div>
        <h1 className="text-lg font-semibold text-slate-900">Something went wrong</h1>
        <p className="mt-1 text-sm text-slate-600">
          An unexpected error occurred while rendering the app.
        </p>
      </div>
      <Button variant="outline" size="sm" onClick={() => window.location.reload()}>
        Reload page
      </Button>
    </div>
  )
}
