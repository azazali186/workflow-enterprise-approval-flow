import { useQuery } from '@tanstack/react-query'
import { analyticsService } from '@/services/analytics.service'

export interface AnalyticsRange {
  startDate: string
  endDate: string
}

export function useApprovalStats(range: AnalyticsRange) {
  return useQuery({
    queryKey: ['analytics', 'approvals', range.startDate, range.endDate],
    queryFn: () =>
      analyticsService.approvalStats({
        start_date: range.startDate,
        end_date: range.endDate,
      }),
    enabled: Boolean(range.startDate && range.endDate),
  })
}

export function useEscalationMetrics(range: AnalyticsRange) {
  return useQuery({
    queryKey: ['analytics', 'escalations', range.startDate, range.endDate],
    queryFn: () =>
      analyticsService.escalationMetrics({
        start_date: range.startDate,
        end_date: range.endDate,
      }),
    enabled: Boolean(range.startDate && range.endDate),
  })
}

export function useWorkflowPerformance(workflowId: string | null) {
  return useQuery({
    queryKey: ['analytics', 'workflows', workflowId],
    queryFn: () => analyticsService.workflowPerformance(workflowId!),
    enabled: Boolean(workflowId),
  })
}

export function useApproverPerformance(approverId: string | null) {
  return useQuery({
    queryKey: ['analytics', 'approvers', approverId],
    queryFn: () => analyticsService.approverPerformance(approverId!),
    enabled: Boolean(approverId),
  })
}
