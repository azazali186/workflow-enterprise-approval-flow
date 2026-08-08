import { post } from './api/client'

export interface DateRangePayload {
  start_date?: string
  end_date?: string
}

/**
 * The analytics endpoints return free-form maps, so every consumer must
 * access fields defensively. These helpers keep the unknown-shape handling
 * in one place.
 */
export const analyticsService = {
  approvalStats(payload: DateRangePayload = {}) {
    return post<Record<string, unknown>>('/analytics/approvals', payload)
  },

  workflowPerformance(workflowId: string) {
    return post<Record<string, unknown>>('/analytics/workflows', { workflow_id: workflowId })
  },

  approverPerformance(approverId: string) {
    return post<Record<string, unknown>>('/analytics/approvers', { approver_id: approverId })
  },

  escalationMetrics(payload: DateRangePayload = {}) {
    return post<Record<string, unknown>>('/analytics/escalations', payload)
  },
}

/** Defensive field accessors for the free-form analytics payloads. */
export function pickNumber(source: Record<string, unknown> | undefined, key: string): number | null {
  if (!source) return null
  const value = source[key]
  const n = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(n) ? n : null
}

export function pickString(
  source: Record<string, unknown> | undefined,
  key: string,
): string | null {
  if (!source) return null
  const value = source[key]
  return typeof value === 'string' ? value : null
}

export function pickMap(
  source: Record<string, unknown> | undefined,
  key: string,
): Record<string, unknown> | null {
  if (!source) return null
  const value = source[key]
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null
}

export function pickArray(
  source: Record<string, unknown> | undefined,
  key: string,
): unknown[] | null {
  if (!source) return null
  const value = source[key]
  return Array.isArray(value) ? value : null
}
