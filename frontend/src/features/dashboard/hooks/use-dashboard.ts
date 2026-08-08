import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { analyticsService } from '@/services/analytics.service'
import { daysAgoIso } from '@/utils/format'

export function useDashboardStats() {
  // Stable query keys: the date range is computed once so ordinary re-renders
  // (hover states, layout changes) never trigger a new fetch per render.
  const { startDate, endDate } = useMemo(
    () => ({ startDate: daysAgoIso(30), endDate: new Date().toISOString() }),
    [],
  )

  const approvals = useQuery({
    queryKey: ['analytics', 'approvals', startDate, endDate],
    queryFn: () => analyticsService.approvalStats({ start_date: startDate, end_date: endDate }),
    staleTime: 5 * 60_000,
  })

  const escalations = useQuery({
    queryKey: ['analytics', 'escalations', startDate, endDate],
    queryFn: () => analyticsService.escalationMetrics({ start_date: startDate, end_date: endDate }),
    staleTime: 5 * 60_000,
  })

  return { approvals, escalations, startDate, endDate }
}
