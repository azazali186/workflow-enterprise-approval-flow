import { useMemo } from 'react'
import {
  Activity,
  AlertOctagon,
  CheckCircle2,
  Clock,
  Gauge,
  Hourglass,
  TrendingUp,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { StatCard } from '@/features/dashboard/components/stat-card'
import { StatusDonut, statusMapToData } from '@/features/dashboard/components/status-donut'
import { useDashboardStats } from '@/features/dashboard/hooks/use-dashboard'
import { useAppSelector } from '@/store/hooks'
import { pickMap, pickNumber, pickString } from '@/services/analytics.service'
import { formatNumber } from '@/utils/format'

function greeting(): string {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 18) return 'Good afternoon'
  return 'Good evening'
}

export default function DashboardPage() {
  const user = useAppSelector((state) => state.auth.user)
  const { approvals, escalations } = useDashboardStats()
  const loading = approvals.isLoading || escalations.isLoading

  const approvalStats = approvals.data
  const escalationStats = escalations.data

  const statusData = useMemo(
    () => statusMapToData(pickMap(approvalStats, 'by_status')),
    [approvalStats],
  )
  const escalationByLevel = useMemo(
    () => statusMapToData(pickMap(escalationStats, 'by_level')),
    [escalationStats],
  )

  const approvalRate = pickNumber(approvalStats, 'approval_rate')
  const avgDecisionHours = pickNumber(approvalStats, 'avg_decision_time_hours')
  const avgResolutionHours = pickNumber(escalationStats, 'avg_resolution_time_hours')
  const recentSubmissions = pickNumber(approvalStats, 'recent_submissions')
  const lastSync = pickString(escalationStats, 'generated_at') ?? pickString(approvalStats, 'generated_at')

  const firstName = user?.name.split(' ')[0] ?? 'there'
  const today = new Date().toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
          {greeting()}, {firstName}
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          {today} · Here's what's happening across your workflows
          {lastSync ? ` · updated ${new Date(lastSync).toLocaleTimeString()}` : ''}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Total approvals"
          icon={CheckCircle2}
          tone="primary"
          loading={loading}
          value={formatNumber(pickNumber(approvalStats, 'total_approvals'))}
        />
        <StatCard
          label="Pending approvals"
          icon={Hourglass}
          tone="warning"
          loading={loading}
          value={formatNumber(pickNumber(approvalStats, 'pending_approvals'))}
        />
        <StatCard
          label="Approval rate"
          icon={TrendingUp}
          tone="success"
          loading={loading}
          value={approvalRate == null ? '—' : `${(approvalRate * 100).toFixed(0)}%`}
          hint={recentSubmissions != null ? `${recentSubmissions} submitted recently` : undefined}
        />
        <StatCard
          label="Avg decision time"
          icon={Clock}
          tone="info"
          loading={loading}
          value={avgDecisionHours == null ? '—' : `${avgDecisionHours.toFixed(1)}h`}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <div>
              <CardTitle>Approvals by status</CardTitle>
              <CardDescription>Distribution over the last 30 days</CardDescription>
            </div>
            <Activity className="h-4 w-4 text-slate-300" />
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex h-56 items-center justify-center">
                <div className="skeleton h-40 w-40 rounded-full" />
              </div>
            ) : (
              <StatusDonut data={statusData} totalLabel="approvals" />
            )}
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <div>
                <CardTitle>Escalations</CardTitle>
                <CardDescription>Last 30 days</CardDescription>
              </div>
              <AlertOctagon className="h-4 w-4 text-slate-300" />
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between rounded-lg border border-slate-100 bg-slate-50/60 px-3.5 py-3">
                <div>
                  <p className="text-xs font-medium text-slate-500">Total</p>
                  <p className="text-lg font-bold tabular-nums text-slate-900">
                    {loading ? '—' : formatNumber(pickNumber(escalationStats, 'total_escalations'))}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-xs font-medium text-slate-500">Pending</p>
                  <p className="text-lg font-bold tabular-nums text-amber-600">
                    {loading ? '—' : formatNumber(pickNumber(escalationStats, 'pending_escalations'))}
                  </p>
                </div>
              </div>
              <div className="flex items-center justify-between px-1">
                <span className="text-xs text-slate-400">Avg resolution</span>
                <span className="text-sm font-semibold tabular-nums text-slate-700">
                  {avgResolutionHours == null ? '—' : `${avgResolutionHours.toFixed(1)}h`}
                </span>
              </div>
              <div className="flex items-center justify-between px-1">
                <span className="text-xs text-slate-400">By level</span>
                <span className="text-sm font-semibold text-slate-700">
                  {escalationByLevel.length > 0 ? escalationByLevel.map((d) => `L${d.name}: ${d.value}`).join(' · ') : '—'}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quick actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-slate-500">
              <p className="flex items-center gap-2">
                <Gauge className="h-4 w-4 text-primary-500" />
                Use the Analytics page for deeper trends
              </p>
              <p className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                Approve pending items from the Approvals page
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
