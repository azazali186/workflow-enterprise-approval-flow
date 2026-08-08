import { useState } from 'react'
import { KeyRound, Pencil, Plus, ShieldCheck, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { useRolesTable, useRoleMutations } from '@/features/roles/hooks/use-roles'
import { RoleFormModal } from '@/features/roles/components/role-form-modal'
import { RolePermissionsModal } from '@/features/roles/components/role-permissions-modal'
import { formatDate, titleCase } from '@/utils/format'
import type { Role } from '@/types/models'

export default function RolesPage() {
  const table = useRolesTable()
  const { deleteRole } = useRoleMutations()
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Role | null>(null)
  const [permissionRole, setPermissionRole] = useState<Role | null>(null)
  const [deleting, setDeleting] = useState<Role | null>(null)

  const columns: ColumnDef<Role>[] = [
    {
      key: 'name',
      header: 'Role',
      sortable: true,
      render: (role) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-50 text-primary-600">
            <ShieldCheck className="h-4 w-4" />
          </span>
          <div>
            <p className="font-medium text-slate-900">{titleCase(role.name)}</p>
            {role.description && (
              <p className="max-w-[280px] truncate text-xs text-slate-500">{role.description}</p>
            )}
          </div>
        </div>
      ),
    },
    {
      key: 'is_default',
      header: 'Type',
      hideBelow: 'md',
      render: (role) =>
        role.is_default ? <Badge variant="primary">Default</Badge> : <Badge variant="neutral">Custom</Badge>,
    },
    {
      key: 'created_at',
      header: 'Created',
      hideBelow: 'lg',
      sortable: true,
      render: (role) => <span className="text-slate-500">{formatDate(role.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (role) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button
            variant="ghost"
            size="icon"
            title="Manage permissions"
            onClick={() => setPermissionRole(role)}
          >
            <KeyRound className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Edit role"
            onClick={() => {
              setEditing(role)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete role"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(role)}
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
        title="Roles"
        description="Define roles and control which API permissions each one holds."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">New role</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(role) => role.id}
        noun="roles"
        searchPlaceholder="Search roles…"
        searchable={false}
      />

      <RoleFormModal open={formOpen} onClose={() => setFormOpen(false)} role={editing} />
      <RolePermissionsModal role={permissionRole} onClose={() => setPermissionRole(null)} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete role?"
        description={
          deleting
            ? `${titleCase(deleting.name)} will be removed and users will lose its permissions.`
            : undefined
        }
        confirmLabel="Delete role"
        loading={deleteRole.isPending}
        onConfirm={() => {
          if (deleting) deleteRole.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
