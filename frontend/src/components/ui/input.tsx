import { forwardRef } from 'react'
import { cn } from '@/utils/cn'

const base =
  'h-9 w-full rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 shadow-sm transition-colors placeholder:text-slate-400 hover:border-slate-300 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/25 disabled:cursor-not-allowed disabled:opacity-50'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { className, invalid, ...props },
  ref,
) {
  return (
    <input
      ref={ref}
      className={cn(
        base,
        invalid && 'border-rose-300 focus:border-rose-500 focus:ring-rose-500/25',
        className,
      )}
      aria-invalid={invalid || undefined}
      {...props}
    />
  )
})
