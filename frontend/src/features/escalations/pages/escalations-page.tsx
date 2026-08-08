import { useState } from 'react'
import { AlertOctagon, CheckCircle2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { StatusBadge } from '@/components/ui/status-badge'
import {
  useEscalationsTable,
  useEscalationMutations,
} from '@/features/escalations/hooks/use-escalations'
import { EscalationFormModal } from '@/features/escalations/components/escalation-form-modal'
import { formatDateTime, shortId, truncate } from '@/utils/format'
import type { Escalation } from '@/types/models'

export default function EscalationsPage() {
  const table = useEscalationsTable()
  const { resolveEscalation } = useEscalationMutations()
  const [createOpen, setCreateOpen] = useState(false)

  const columns: ColumnDef<Escalation>[] = [
    {
      key: 'id',
      header: 'Escalation',
      render: (escalation) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
            <AlertOctagon className="h-4 w-4" />
          </span>
          <div>
            <p className="font-mono text-xs text-slate-500">{shortId(escalation.id)}</p>
            <p className="text-xs text-slate-400">approval {shortId(escalation.approval_id)}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'level',
      header: 'Level',
      render: (escalation) => (
        <span className="rounded-md bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700">
          L{escalation.level}
        </span>
      ),
    },
    {
      key: 'escalated_to',
      header: 'Escalated to',
      hideBelow: 'sm',
      render: (escalation) => <span className="font-mono text-xs text-slate-500">{shortId(escalation.escalated_to)}</span>,
    },
    {
      key: 'reason',
      header: 'Reason',
      render: (escalation) => (
        <span className="block max-w-[260px] truncate text-slate-600" title={escalation.reason}>
          {truncate(escalation.reason, 48)}
        </span>
      ),
    },
    {
      key: 'escalated_at',
      header: 'Escalated',
      hideBelow: 'lg',
      sortable: true,
      render: (escalation) => <span className="text-slate-500">{formatDateTime(escalation.escalated_at)}</span>,
    },
    {
      key: 'resolved_at',
      header: 'Status',
      hideBelow: 'md',
      render: (escalation) =>
        escalation.resolved_at ? <StatusBadge value="resolved" /> : <StatusBadge value="unresolved" />,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (escalation) =>
        escalation.resolved_at ? (
          <span className="text-xs text-slate-400">Resolved</span>
        ) : (
          <Button
            variant="outline"
            size="sm"
            className="text-emerald-700 hover:border-emerald-300 hover:bg-emerald-50"
            onClick={(event) => {
              event.stopPropagation()
              resolveEscalation.mutate(escalation.id)
            }}
            loading={resolveEscalation.isPending}
          >
            <CheckCircle2 className="h-3.5 w-3.5" />
            Resolve
          </Button>
        ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Workflows"
        title="Escalations"
        description="Approvals pushed to higher levels for review."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Create escalation</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(escalation) => escalation.id}
        noun="escalations"
        searchPlaceholder="Search escalations…"
      />

      <EscalationFormModal open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  )
}
