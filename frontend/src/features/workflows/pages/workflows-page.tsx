import { useState } from 'react'
import { GitBranch, Pencil, Plus, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Switch } from '@/components/ui/switch'
import { useWorkflowsTable, useWorkflowMutations } from '@/features/workflows/hooks/use-workflows'
import { WorkflowFormModal } from '@/features/workflows/components/workflow-form-modal'
import { formatDate } from '@/utils/format'
import type { Workflow } from '@/types/models'

export default function WorkflowsPage() {
  const table = useWorkflowsTable()
  const { updateWorkflow, deleteWorkflow } = useWorkflowMutations()
  const [optimistic, setOptimistic] = useState<Record<string, boolean>>({})
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Workflow | null>(null)
  const [deleting, setDeleting] = useState<Workflow | null>(null)

  const isActive = (workflow: Workflow) => optimistic[workflow.id] ?? workflow.is_active

  const toggleActive = (workflow: Workflow) => {
    const next = !isActive(workflow)
    setOptimistic((previous) => ({ ...previous, [workflow.id]: next }))
    updateWorkflow.mutate(
      { workflow_id: workflow.id, is_active: next },
      {
        onError: () => setOptimistic((previous) => ({ ...previous, [workflow.id]: !next })),
      },
    )
  }

  const columns: ColumnDef<Workflow>[] = [
    {
      key: 'name',
      header: 'Workflow',
      sortable: true,
      render: (workflow) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600">
            <GitBranch className="h-4 w-4" />
          </span>
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{workflow.name}</p>
            {workflow.description && (
              <p className="max-w-[260px] truncate text-xs text-slate-500">{workflow.description}</p>
            )}
          </div>
        </div>
      ),
    },
    {
      key: 'category',
      header: 'Category',
      render: (workflow) => <Badge variant="neutral">{workflow.category}</Badge>,
    },
    {
      key: 'version',
      header: 'Version',
      hideBelow: 'sm',
      render: (workflow) => (
        <span className="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">
          v{workflow.version}
        </span>
      ),
    },
    {
      key: 'is_active',
      header: 'Active',
      hideBelow: 'md',
      render: (workflow) => (
        <Switch
          checked={isActive(workflow)}
          onChange={() => toggleActive(workflow)}
          disabled={updateWorkflow.isPending}
          aria-label={`Toggle workflow ${workflow.name}`}
        />
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      hideBelow: 'lg',
      sortable: true,
      render: (workflow) => <span className="text-slate-500">{formatDate(workflow.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (workflow) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button
            variant="ghost"
            size="icon"
            title="Edit workflow"
            onClick={() => {
              setEditing(workflow)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete workflow"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(workflow)}
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
        title="Workflows"
        description="Define the approval paths applications travel through."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">New workflow</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(workflow) => workflow.id}
        noun="workflows"
        searchPlaceholder="Search workflows…"
      />

      <WorkflowFormModal open={formOpen} onClose={() => setFormOpen(false)} workflow={editing} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete workflow?"
        description={
          deleting ? `“${deleting.name}” will be removed. Active applications may be affected.` : undefined
        }
        confirmLabel="Delete workflow"
        loading={deleteWorkflow.isPending}
        onConfirm={() => {
          if (deleting) deleteWorkflow.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
