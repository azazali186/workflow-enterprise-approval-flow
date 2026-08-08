import { Badge, type BadgeVariant } from './badge'
import { titleCase } from '@/utils/format'

const STATUS_META: Record<string, BadgeVariant> = {
  active: 'success',
  inactive: 'neutral',
  locked: 'danger',
  pending: 'warning',
  draft: 'neutral',
  submitted: 'info',
  in_review: 'info',
  inreview: 'info',
  approved: 'success',
  rejected: 'danger',
  escalated: 'warning',
  completed: 'success',
  cancelled: 'neutral',
  canceled: 'neutral',
  success: 'success',
  failed: 'danger',
  resolved: 'success',
  unresolved: 'warning',
  open: 'info',
  closed: 'neutral',
}

const PRIORITY_META: Record<string, BadgeVariant> = {
  low: 'neutral',
  medium: 'info',
  high: 'warning',
  urgent: 'danger',
}

export function StatusBadge({ value }: { value?: string | null }) {
  if (!value) return <Badge variant="neutral">—</Badge>
  const variant = STATUS_META[value.toLowerCase()] ?? 'neutral'
  return <Badge variant={variant}>{titleCase(value)}</Badge>
}

export function PriorityBadge({ value }: { value?: string | null }) {
  if (!value) return <Badge variant="neutral">—</Badge>
  const variant = PRIORITY_META[value.toLowerCase()] ?? 'neutral'
  return <Badge variant={variant}>{titleCase(value)}</Badge>
}

export function MethodBadge({ method }: { method?: string }) {
  if (!method) return <Badge variant="neutral">—</Badge>
  const variant: BadgeVariant =
    method.toUpperCase() === 'POST'
      ? 'primary'
      : method.toUpperCase() === 'PATCH'
        ? 'warning'
        : method.toUpperCase() === 'DELETE'
          ? 'danger'
          : 'neutral'
  return <Badge variant={variant}>{method.toUpperCase()}</Badge>
}
