import type { ComponentType } from 'react'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/utils/cn'

export interface StatCardProps {
  label: string
  value: React.ReactNode
  hint?: string
  icon: ComponentType<{ className?: string }>
  tone?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  loading?: boolean
}

const tones = {
  primary: 'bg-primary-50 text-primary-600',
  success: 'bg-emerald-50 text-emerald-600',
  warning: 'bg-amber-50 text-amber-600',
  danger: 'bg-rose-50 text-rose-600',
  info: 'bg-sky-50 text-sky-600',
}

export function StatCard({ label, value, hint, icon: Icon, tone = 'primary', loading }: StatCardProps) {
  return (
    <Card className="p-5" hover>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wide text-slate-400">{label}</p>
          {loading ? (
            <Skeleton className="mt-2.5 h-8 w-16" />
          ) : (
            <p className="mt-2 text-2xl font-bold tracking-tight text-slate-900 tabular-nums">
              {value}
            </p>
          )}
          {hint && <p className="mt-1 text-xs text-slate-500">{hint}</p>}
        </div>
        <span className={cn('flex h-9 w-9 shrink-0 items-center justify-center rounded-lg', tones[tone])}>
          <Icon className="h-[18px] w-[18px]" />
        </span>
      </div>
    </Card>
  )
}
