import { useEffect, useState } from 'react'
import { KeyRound, Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { MethodBadge } from '@/components/ui/status-badge'
import {
  usePermissionsTable,
  usePermissionMutations,
} from '@/features/permissions/hooks/use-permissions'
import { PermissionFormModal } from '@/features/permissions/components/permission-form-modal'
import { formatDate } from '@/utils/format'
import type { Permission } from '@/types/models'

export default function PermissionsPage() {
  const table = usePermissionsTable()
  const { deletePermission } = usePermissionMutations()
  const [methodFilter, setMethodFilter] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Permission | null>(null)
  const [deleting, setDeleting] = useState<Permission | null>(null)

  useEffect(() => {
    table.setFilters(methodFilter ? { method: methodFilter } : {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [methodFilter])

  const columns: ColumnDef<Permission>[] = [
    {
      key: 'name',
      header: 'Permission',
      sortable: true,
      render: (permission) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-50 text-violet-600">
            <KeyRound className="h-4 w-4" />
          </span>
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{permission.name}</p>
            <p className="truncate font-mono text-xs text-slate-400">{permission.route}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'method',
      header: 'Method',
      hideBelow: 'sm',
      render: (permission) => <MethodBadge method={permission.method} />,
    },
    {
      key: 'path',
      header: 'Path',
      hideBelow: 'md',
      render: (permission) => (
        <span className="block max-w-[300px] truncate font-mono text-xs text-slate-500">
          {permission.path}
        </span>
      ),
    },
    {
      key: 'service',
      header: 'Service',
      hideBelow: 'lg',
      render: (permission) => <span className="text-slate-500">{permission.service}</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      hideBelow: 'xl',
      render: (permission) => <span className="text-slate-500">{formatDate(permission.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (permission) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button
            variant="ghost"
            size="icon"
            title="Edit permission"
            onClick={() => {
              setEditing(permission)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete permission"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(permission)}
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
        eyebrow="Administration"
        title="Permissions"
        description="API route permissions that roles can be granted."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">New permission</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(permission) => permission.id}
        noun="permissions"
        searchPlaceholder="Search by name or route…"
        toolbar={
          <Select
            value={methodFilter}
            onChange={(event) => setMethodFilter(event.target.value)}
            placeholder="All methods"
            className="w-32"
            aria-label="Filter by method"
          >
            <option value="POST">POST</option>
            <option value="PATCH">PATCH</option>
            <option value="DELETE">DELETE</option>
          </Select>
        }
      />

      <PermissionFormModal open={formOpen} onClose={() => setFormOpen(false)} permission={editing} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete permission?"
        description={
          deleting ? `“${deleting.name}” will be removed from every role that uses it.` : undefined
        }
        confirmLabel="Delete permission"
        loading={deletePermission.isPending}
        onConfirm={() => {
          if (deleting) deletePermission.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
