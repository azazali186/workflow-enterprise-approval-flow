import { useEffect, useState } from 'react'
import { KeySquare, Pencil, Plus, Trash2, UserPlus } from 'lucide-react'
import { Avatar } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import { StatusBadge } from '@/components/ui/status-badge'
import { useUsersTable, useUserMutations } from '@/features/users/hooks/use-users'
import { UserFormModal } from '@/features/users/components/user-form-modal'
import { UserRolesModal } from '@/features/users/components/user-roles-modal'
import { formatDate, relativeTime, titleCase } from '@/utils/format'
import type { User } from '@/types/models'

export default function UsersPage() {
  const table = useUsersTable()
  const { deleteUser } = useUserMutations()
  const [statusFilter, setStatusFilter] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [rolesUser, setRolesUser] = useState<User | null>(null)
  const [deleting, setDeleting] = useState<User | null>(null)

  useEffect(() => {
    table.setFilters(statusFilter ? { status: statusFilter } : {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter])

  const columns: ColumnDef<User>[] = [
    {
      key: 'user',
      header: 'User',
      render: (user) => (
        <div className="flex items-center gap-3">
          <Avatar name={user.name} size="sm" />
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{user.name}</p>
            <p className="truncate text-xs text-slate-500">{user.email}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (user) => <StatusBadge value={user.status} />,
    },
    {
      key: 'roles',
      header: 'Roles',
      hideBelow: 'md',
      render: (user) =>
        user.roles?.length ? (
          <div className="flex max-w-[220px] flex-wrap gap-1">
            {user.roles.slice(0, 3).map((role) => (
              <Badge key={role.id} variant={role.name === 'admin' ? 'primary' : 'neutral'}>
                {titleCase(role.name)}
              </Badge>
            ))}
            {(user.roles?.length ?? 0) > 3 && (
              <Badge variant="neutral">+{user.roles.length - 3}</Badge>
            )}
          </div>
        ) : (
          <span className="text-slate-400">—</span>
        ),
    },
    {
      key: 'last_login_at',
      header: 'Last login',
      hideBelow: 'lg',
      sortable: true,
      render: (user) => <span className="text-slate-500">{relativeTime(user.last_login_at)}</span>,
    },
    {
      key: 'created_at',
      header: 'Created',
      hideBelow: 'xl',
      sortable: true,
      render: (user) => <span className="text-slate-500">{formatDate(user.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (user) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button variant="ghost" size="icon" title="Manage roles" onClick={() => setRolesUser(user)}>
            <KeySquare className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Edit user"
            onClick={() => {
              setEditing(user)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete user"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(user)}
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
        title="Users"
        description="Manage accounts, access status and role assignments."
        actions={
          <>
            <Button
              variant="outline"
              onClick={() => {
                setEditing(null)
                setFormOpen(true)
              }}
            >
              <UserPlus className="h-4 w-4" />
              <span className="hidden sm:inline">New user</span>
            </Button>
          </>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(user) => user.id}
        noun="users"
        searchPlaceholder="Search by name or email…"
        title="All users"
        description={table.meta.total > 0 ? `${table.meta.total} accounts in total` : undefined}
        toolbar={
          <Select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value)}
            placeholder="All statuses"
            className="w-36"
            aria-label="Filter by status"
          >
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="locked">Locked</option>
            <option value="pending">Pending</option>
          </Select>
        }
      />

      <UserFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        user={editing}
      />
      <UserRolesModal user={rolesUser} onClose={() => setRolesUser(null)} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete user?"
        description={
          deleting
            ? `${deleting.name} will lose access immediately and their record will be removed.`
            : undefined
        }
        confirmLabel="Delete user"
        loading={deleteUser.isPending}
        onConfirm={() => {
          if (deleting) deleteUser.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
