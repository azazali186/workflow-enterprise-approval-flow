import { useEffect, useState } from 'react'
import { Gavel, Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { StatusBadge } from '@/components/ui/status-badge'
import { useApprovalsTable, useApprovalMutations } from '@/features/approvals/hooks/use-approvals'
import { ApprovalDecideModal } from '@/features/approvals/components/approval-decide-modal'
import { ApprovalCreateModal } from '@/features/approvals/components/approval-create-modal'
import { formatDateTime, shortId, titleCase, truncate } from '@/utils/format'
import type { Approval } from '@/types/models'

export default function ApprovalsPage() {
  const table = useApprovalsTable()
  const { deleteApproval } = useApprovalMutations()
  const [statusFilter, setStatusFilter] = useState('')
  const [deciding, setDeciding] = useState<Approval | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [deleting, setDeleting] = useState<Approval | null>(null)

  useEffect(() => {
    table.setFilters(statusFilter ? { status: statusFilter } : {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter])

  const columns: ColumnDef<Approval>[] = [
    {
      key: 'id',
      header: 'Approval',
      render: (approval) => (
        <div>
          <p className="font-mono text-xs text-slate-500">{shortId(approval.id)}</p>
          <p className="mt-0.5 text-xs text-slate-400">app {shortId(approval.application_id)}</p>
        </div>
      ),
    },
    {
      key: 'approver_id',
      header: 'Approver',
      hideBelow: 'sm',
      render: (approval) => <span className="font-mono text-xs text-slate-500">{shortId(approval.approver_id)}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (approval) => <StatusBadge value={approval.status} />,
    },
    {
      key: 'decision',
      header: 'Decision',
      render: (approval) =>
        approval.decision ? (
          <span className="text-sm capitalize text-slate-700">{titleCase(approval.decision)}</span>
        ) : (
          <span className="text-slate-300">—</span>
        ),
    },
    {
      key: 'comment',
      header: 'Comment',
      hideBelow: 'lg',
      render: (approval) =>
        approval.comment ? (
          <span className="block max-w-[240px] truncate text-slate-500" title={approval.comment}>
            {truncate(approval.comment, 40)}
          </span>
        ) : (
          <span className="text-slate-300">—</span>
        ),
    },
    {
      key: 'decided_at',
      header: 'Decided',
      hideBelow: 'xl',
      sortable: true,
      render: (approval) => (
        <span className="text-slate-500">{approval.decided_at ? formatDateTime(approval.decided_at) : '—'}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (approval) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button
            variant="ghost"
            size="icon"
            title="Record decision"
            className="text-primary-600 hover:bg-primary-50"
            onClick={() => setDeciding(approval)}
          >
            <Gavel className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete approval"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(approval)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Workflows"
        title="Approvals"
        description="Review, decide and escalate pending approvals."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Create approval</span>
            <Pencil className="h-4 w-4 sm:hidden" />
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(approval) => approval.id}
        noun="approvals"
        searchPlaceholder="Search approvals…"
        toolbar={
          <Select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value)}
            placeholder="All statuses"
            className="w-32"
            aria-label="Filter by status"
          >
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
            <option value="escalated">Escalated</option>
            <option value="completed">Completed</option>
          </Select>
        }
      />

      <ApprovalDecideModal approval={deciding} onClose={() => setDeciding(null)} />
      <ApprovalCreateModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete approval?"
        description={
          deleting ? `Approval ${deleting.id.slice(0, 8)} and its escalations will be removed.` : undefined
        }
        confirmLabel="Delete approval"
        loading={deleteApproval.isPending}
        onConfirm={() => {
          if (deleting) deleteApproval.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
