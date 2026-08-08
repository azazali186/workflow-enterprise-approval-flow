import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { BarChart3, Clock, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { KvTable } from '@/features/analytics/components/kv-table'
import { StatusDonut, statusMapToData } from '@/features/dashboard/components/status-donut'
import {
  useApprovalStats,
  useApproverPerformance,
  useEscalationMetrics,
  useWorkflowPerformance,
} from '@/features/analytics/hooks/use-analytics'
import { pickMap, pickNumber } from '@/services/analytics.service'
import { workflowsService } from '@/services/workflows.service'
import { usersService } from '@/services/users.service'
import { cn } from '@/utils/cn'
import { daysAgoIso, formatNumber, toIsoDate } from '@/utils/format'

type Tab = 'approvals' | 'escalations' | 'workflows' | 'approvers'

const tabs: { id: Tab; label: string }[] = [
  { id: 'approvals', label: 'Approval stats' },
  { id: 'escalations', label: 'Escalation metrics' },
  { id: 'workflows', label: 'Workflow performance' },
  { id: 'approvers', label: 'Approver performance' },
]

function DateRangePicker({
  value,
  onChange,
}: {
  value: { startDate: string; endDate: string }
  onChange: (next: { startDate: string; endDate: string }) => void
}) {
  return (
    <div className="flex items-center gap-2">
      <input
        type="date"
        value={toIsoDate(new Date(value.startDate))}
        max={toIsoDate(new Date(value.endDate))}
        onChange={(event) => event.target.value && onChange({ ...value, startDate: new Date(event.target.value).toISOString() })}
        className="h-9 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/25"
        aria-label="Start date"
      />
      <span className="text-slate-400">→</span>
      <input
        type="date"
        value={toIsoDate(new Date(value.endDate))}
        min={toIsoDate(new Date(value.startDate))}
        onChange={(event) => event.target.value && onChange({ ...value, endDate: new Date(event.target.value).toISOString() })}
        className="h-9 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/25"
        aria-label="End date"
      />
    </div>
  )
}

function LoadingBlock() {
  return (
    <div className="space-y-3">
      <Skeleton className="h-6 w-40" />
      <Skeleton className="h-48 w-full" />
    </div>
  )
}

function MetricCard({
  label,
  value,
  icon: Icon,
}: {
  label: string
  value: React.ReactNode
  icon: React.ComponentType<{ className?: string }>
}) {
  return (
    <div className="rounded-xl border border-slate-200/80 bg-white p-4 shadow-card">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-slate-400">
        <Icon className="h-3.5 w-3.5" />
        {label}
      </div>
      <p className="mt-2 text-xl font-bold tabular-nums text-slate-900">{value}</p>
    </div>
  )
}

export default function AnalyticsPage() {
  const [tab, setTab] = useState<Tab>('approvals')
  const [range, setRange] = useState({
    startDate: daysAgoIso(30),
    endDate: new Date().toISOString(),
  })
  const [workflowId, setWorkflowId] = useState('')
  const [approverId, setApproverId] = useState('')

  const approvalStats = useApprovalStats(range)
  const escalationMetrics = useEscalationMetrics(range)
  const workflowPerf = useWorkflowPerformance(workflowId || null)
  const approverPerf = useApproverPerformance(approverId || null)

  const { data: workflows } = useQuery({
    queryKey: ['workflows', 'options'],
    queryFn: () => workflowsService.list({ limit: 100 }),
  })
  const { data: approvers } = useQuery({
    queryKey: ['users', 'approver-options'],
    queryFn: () => usersService.list({ limit: 100 }),
  })

  const byStatus = statusMapToData(pickMap(approvalStats.data, 'by_status'))
  const byLevel = statusMapToData(pickMap(escalationMetrics.data, 'by_level'))
  const approvalRate = pickNumber(approvalStats.data, 'approval_rate')

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Insights"
        title="Analytics"
        description="Dig into approval and escalation trends across the platform."
      />

      <div className="flex flex-wrap gap-1.5 rounded-xl border border-slate-200/80 bg-white p-1 shadow-card">
        {tabs.map(({ id, label }) => (
          <button
            key={id}
            type="button"
            onClick={() => setTab(id)}
            className={cn(
              'rounded-lg px-3.5 py-2 text-sm font-medium transition-colors duration-150',
              tab === id
                ? 'bg-primary-600 text-white shadow-sm'
                : 'text-slate-600 hover:bg-slate-100',
            )}
            aria-pressed={tab === id}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'approvals' && (
        <Card>
          <CardHeader>
            <div>
              <CardTitle>Approval statistics</CardTitle>
              <CardDescription>Aggregates over the selected period</CardDescription>
            </div>
            <DateRangePicker value={range} onChange={setRange} />
          </CardHeader>
          <CardContent>
            {approvalStats.isLoading ? (
              <LoadingBlock />
            ) : (
              <div className="grid gap-6 lg:grid-cols-2">
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-3">
                    <MetricCard label="Total" icon={BarChart3} value={formatNumber(pickNumber(approvalStats.data, 'total_approvals'))} />
                    <MetricCard label="Pending" icon={Clock} value={formatNumber(pickNumber(approvalStats.data, 'pending_approvals'))} />
                    <MetricCard label="Approval rate" icon={TrendingUp} value={approvalRate == null ? '—' : `${(approvalRate * 100).toFixed(0)}%`} />
                    <MetricCard label="Avg decision" icon={Clock} value={pickNumber(approvalStats.data, 'avg_decision_time_hours') == null ? '—' : `${pickNumber(approvalStats.data, 'avg_decision_time_hours')!.toFixed(1)}h`} />
                  </div>
                  <KvTable data={approvalStats.data} />
                </div>
                <StatusDonut data={byStatus} totalLabel="approvals" />
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {tab === 'escalations' && (
        <Card>
          <CardHeader>
            <div>
              <CardTitle>Escalation metrics</CardTitle>
              <CardDescription>Aggregates over the selected period</CardDescription>
            </div>
            <DateRangePicker value={range} onChange={setRange} />
          </CardHeader>
          <CardContent>
            {escalationMetrics.isLoading ? (
              <LoadingBlock />
            ) : (
              <div className="grid gap-6 lg:grid-cols-2">
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-3">
                    <MetricCard label="Total" icon={BarChart3} value={formatNumber(pickNumber(escalationMetrics.data, 'total_escalations'))} />
                    <MetricCard label="Resolved" icon={Clock} value={formatNumber(pickNumber(escalationMetrics.data, 'resolved_escalations'))} />
                    <MetricCard label="Pending" icon={Clock} value={formatNumber(pickNumber(escalationMetrics.data, 'pending_escalations'))} />
                    <MetricCard label="Avg resolution" icon={TrendingUp} value={pickNumber(escalationMetrics.data, 'avg_resolution_time_hours') == null ? '—' : `${pickNumber(escalationMetrics.data, 'avg_resolution_time_hours')!.toFixed(1)}h`} />
                  </div>
                  <KvTable data={escalationMetrics.data} />
                </div>
                <div>
                  <p className="mb-3 text-xs font-medium uppercase tracking-wide text-slate-400">
                    By level
                  </p>
                  {byLevel.length === 0 ? (
                    <div className="flex h-48 items-center justify-center rounded-lg border border-dashed border-slate-200 text-sm text-slate-400">
                      No escalation data
                    </div>
                  ) : (
                    <ResponsiveContainer width="100%" height={220}>
                      <BarChart data={byLevel}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
                        <XAxis dataKey="name" tick={{ fontSize: 12, fill: '#64748b' }} />
                        <YAxis tick={{ fontSize: 12, fill: '#64748b' }} allowDecimals={false} width={32} />
                        <Tooltip
                          cursor={{ fill: 'rgb(99 102 241 / 0.06)' }}
                          contentStyle={{
                            borderRadius: 12,
                            border: '1px solid rgb(226 232 240)',
                            fontSize: 12,
                          }}
                        />
                        <Bar dataKey="value" name="Count" radius={[6, 6, 0, 0]}>
                          {byLevel.map((item) => (
                            <Cell key={item.name} fill={item.color} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {tab === 'workflows' && (
        <Card>
          <CardHeader>
            <div>
              <CardTitle>Workflow performance</CardTitle>
              <CardDescription>Per-workflow approval behaviour</CardDescription>
            </div>
            <Select
              value={workflowId}
              onChange={(event) => setWorkflowId(event.target.value)}
              placeholder="Select workflow…"
              className="w-56"
              aria-label="Select workflow"
            >
              {(workflows?.rows ?? []).map((workflow) => (
                <option key={workflow.id} value={workflow.id}>
                  {workflow.name}
                </option>
              ))}
            </Select>
          </CardHeader>
          <CardContent>
            {workflowPerf.isLoading ? (
              <LoadingBlock />
            ) : (
              <KvTable data={workflowPerf.data} />
            )}
          </CardContent>
        </Card>
      )}

      {tab === 'approvers' && (
        <Card>
          <CardHeader>
            <div>
              <CardTitle>Approver performance</CardTitle>
              <CardDescription>Individual approval activity</CardDescription>
            </div>
            <Select
              value={approverId}
              onChange={(event) => setApproverId(event.target.value)}
              placeholder="Select approver…"
              className="w-56"
              aria-label="Select approver"
            >
              {(approvers?.rows ?? []).map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name}
                </option>
              ))}
            </Select>
          </CardHeader>
          <CardContent>
            {approverPerf.isLoading ? (
              <LoadingBlock />
            ) : (
              <KvTable data={approverPerf.data} />
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
