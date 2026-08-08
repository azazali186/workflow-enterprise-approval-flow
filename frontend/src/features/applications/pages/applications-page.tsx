import { useEffect, useState } from 'react'
import { Eye, FilePlus2, Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { PriorityBadge, StatusBadge } from '@/components/ui/status-badge'
import { useApplicationsTable, useApplicationMutations } from '@/features/applications/hooks/use-applications'
import { ApplicationFormModal } from '@/features/applications/components/application-form-modal'
import { ApplicationDetailModal } from '@/features/applications/components/application-detail-modal'
import { formatDateTime, shortId } from '@/utils/format'
import type { Application } from '@/types/models'

export default function ApplicationsPage() {
  const table = useApplicationsTable()
  const { deleteApplication } = useApplicationMutations()
  const [statusFilter, setStatusFilter] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Application | null>(null)
  const [viewing, setViewing] = useState<Application | null>(null)
  const [deleting, setDeleting] = useState<Application | null>(null)

  useEffect(() => {
    const filters: Record<string, unknown> = {}
    if (statusFilter) filters.status = statusFilter
    if (priorityFilter) filters.priority = priorityFilter
    table.setFilters(filters)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter, priorityFilter])

  const columns: ColumnDef<Application>[] = [
    {
      key: 'title',
      header: 'Title',
      render: (application) => (
        <span className="max-w-[220px] truncate font-medium text-slate-700">
          {application.title || 'Untitled'}
        </span>
      ),
    },
    {
      key: 'id',
      header: 'ID',
      render: (application) => (
        <span className="font-mono text-xs text-slate-500">{shortId(application.id)}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (application) => <StatusBadge value={application.status} />,
    },
    {
      key: 'priority',
      header: 'Priority',
      render: (application) => <PriorityBadge value={application.priority} />,
    },
    {
      key: 'workflow_id',
      header: 'Workflow',
      hideBelow: 'md',
      render: (application) => (
        <span className="font-mono text-xs text-slate-500">{shortId(application.workflow_id)}</span>
      ),
    },
    {
      key: 'submitted_at',
      header: 'Submitted',
      hideBelow: 'lg',
      sortable: true,
      render: (application) => (
        <span className="text-slate-500">
          {formatDateTime(application.submitted_at ?? application.created_at)}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (application) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button variant="ghost" size="icon" title="View details" onClick={() => setViewing(application)}>
            <Eye className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Update application"
            onClick={() => {
              setEditing(application)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete application"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(application)}
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
        title="Applications"
        description="Track every submission as it flows through its approval workflow."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Submit application</span>
            <FilePlus2 className="h-4 w-4 sm:hidden" />
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(application) => application.id}
        noun="applications"
        searchPlaceholder="Search applications…"
        onRowClick={(application) => setViewing(application)}
        toolbar={
          <div className="flex items-center gap-2">
            <Select
              value={priorityFilter}
              onChange={(event) => setPriorityFilter(event.target.value)}
              placeholder="All priorities"
              className="w-32"
              aria-label="Filter by priority"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </Select>
            <Select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
              placeholder="All statuses"
              className="w-32"
              aria-label="Filter by status"
            >
              <option value="draft">Draft</option>
              <option value="submitted">Submitted</option>
              <option value="in_review">In review</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
              <option value="completed">Completed</option>
            </Select>
          </div>
        }
      />

      <ApplicationFormModal open={formOpen} onClose={() => setFormOpen(false)} application={editing} />
      <ApplicationDetailModal application={viewing} onClose={() => setViewing(null)} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete application?"
        description={
          deleting
            ? `Application ${deleting.id.slice(0, 8)} and its approvals will be removed.`
            : undefined
        }
        confirmLabel="Delete application"
        loading={deleteApplication.isPending}
        onConfirm={() => {
          if (deleting) deleteApplication.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
