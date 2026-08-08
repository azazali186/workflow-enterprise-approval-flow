import { useEffect, useState } from 'react'
import { History } from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { StatusBadge } from '@/components/ui/status-badge'
import { useLoginLogsTable } from '@/features/login-logs/hooks/use-login-logs'
import { formatDateTime, titleCase, truncate } from '@/utils/format'
import type { LoginLog } from '@/types/models'

export default function LoginLogsPage() {
  const table = useLoginLogsTable()
  const [statusFilter, setStatusFilter] = useState('')

  useEffect(() => {
    table.setFilters(statusFilter ? { status: statusFilter } : {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter])

  const columns: ColumnDef<LoginLog>[] = [
    {
      key: 'email',
      header: 'Account',
      render: (log) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-500">
            <History className="h-4 w-4" />
          </span>
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{log.email}</p>
            {log.user_id && <p className="truncate font-mono text-xs text-slate-400">{log.user_id}</p>}
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (log) => <StatusBadge value={log.status} />,
    },
    {
      key: 'failure_reason',
      header: 'Failure reason',
      hideBelow: 'md',
      render: (log) =>
        log.failure_reason ? (
          <span className="text-slate-500">{titleCase(log.failure_reason.replace(/_/g, ' '))}</span>
        ) : (
          <span className="text-slate-300">—</span>
        ),
    },
    {
      key: 'ip_address',
      header: 'IP address',
      hideBelow: 'sm',
      render: (log) => <span className="font-mono text-xs text-slate-500">{log.ip_address || '—'}</span>,
    },
    {
      key: 'user_agent',
      header: 'User agent',
      hideBelow: 'lg',
      render: (log) => (
        <span className="block max-w-[240px] truncate text-xs text-slate-500" title={log.user_agent}>
          {truncate(log.user_agent, 44)}
        </span>
      ),
    },
    {
      key: 'attempted_at',
      header: 'Attempted',
      sortable: true,
      render: (log) => <span className="text-slate-500">{formatDateTime(log.attempted_at)}</span>,
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Administration"
        title="Login logs"
        description="Authentication attempts across the platform."
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(log) => log.id}
        noun="login attempts"
        searchPlaceholder="Search by email…"
        toolbar={
          <Select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value)}
            placeholder="All statuses"
            className="w-32"
            aria-label="Filter by status"
          >
            <option value="success">Success</option>
            <option value="failed">Failed</option>
            <option value="locked">Locked</option>
          </Select>
        }
      />
    </div>
  )
}
