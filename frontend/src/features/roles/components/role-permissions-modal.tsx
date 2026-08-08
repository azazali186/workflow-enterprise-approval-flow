import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { KeyRound, Loader2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/ui/empty-state'
import { Modal } from '@/components/ui/modal'
import { SearchInput } from '@/components/ui/search-input'
import { Switch } from '@/components/ui/switch'
import { rolesService } from '@/services/roles.service'
import { permissionsService } from '@/services/permissions.service'
import { useRoleMutations } from '@/features/roles/hooks/use-roles'
import { MethodBadge } from '@/components/ui/status-badge'
import type { Permission, Role } from '@/types/models'

export interface RolePermissionsModalProps {
  role: Role | null
  onClose: () => void
}

export function RolePermissionsModal({ role, onClose }: RolePermissionsModalProps) {
  const { assignPermission, removePermission } = useRoleMutations()
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [query, setQuery] = useState('')

  const { data: allPermissions, isLoading: loadingAll } = useQuery({
    queryKey: ['permissions', 'all'],
    queryFn: () => permissionsService.list({ limit: 100 }),
    enabled: Boolean(role),
  })

  const { data: rolePermissions, isLoading: loadingRole } = useQuery({
    queryKey: ['roles', role?.id, 'permissions'],
    queryFn: () => rolesService.permissions(role!.id),
    enabled: Boolean(role),
  })

  useEffect(() => {
    setSelected(new Set((rolePermissions ?? []).map((permission) => permission.id)))
    setQuery('')
  }, [rolePermissions, role])

  const toggle = (permission: Permission) => {
    if (!role) return
    const hasPermission = selected.has(permission.id)
    setSelected((previous) => {
      const next = new Set(previous)
      if (hasPermission) next.delete(permission.id)
      else next.add(permission.id)
      return next
    })
    const mutation = hasPermission ? removePermission : assignPermission
    mutation.mutate(
      { roleId: role.id, permissionId: permission.id },
      {
        onError: () => {
          setSelected((previous) => {
            const next = new Set(previous)
            if (hasPermission) next.add(permission.id)
            else next.delete(permission.id)
            return next
          })
        },
      },
    )
  }

  const permissions = useMemo(() => {
    const rows = allPermissions?.rows ?? []
    if (!query.trim()) return rows
    const needle = query.trim().toLowerCase()
    return rows.filter(
      (permission) =>
        permission.name.toLowerCase().includes(needle) ||
        permission.route.toLowerCase().includes(needle) ||
        permission.path.toLowerCase().includes(needle),
    )
  }, [allPermissions, query])

  const loading = loadingAll || loadingRole

  return (
    <Modal
      open={Boolean(role)}
      onClose={onClose}
      title="Manage permissions"
      description={role ? `Choose which API permissions ${role.name} can access.` : undefined}
      size="lg"
      footer={
        <Button variant="outline" onClick={onClose}>
          Done
        </Button>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-10">
          <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
        </div>
      ) : permissions.length === 0 ? (
        <EmptyState
          icon={KeyRound}
          title="No permissions found"
          description={query ? 'Try a different search.' : 'Create a permission first.'}
        />
      ) : (
        <div>
          <SearchInput
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onClear={() => setQuery('')}
            placeholder="Search permissions…"
            className="mb-3"
          />
          <div className="scrollbar-thin max-h-[24rem] overflow-y-auto rounded-lg border border-slate-200">
            <ul className="divide-y divide-slate-100">
              {permissions.map((permission, index) => {
                const checked = selected.has(permission.id)
                const busy = assignPermission.isPending || removePermission.isPending
                return (
                  <motion.li
                    key={permission.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: Math.min(index * 0.015, 0.3) }}
                    className="flex items-center justify-between gap-3 px-3.5 py-2.5"
                  >
                    <div className="flex min-w-0 items-center gap-2.5">
                      <MethodBadge method={permission.method} />
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-800">
                          {permission.name}
                        </p>
                        <p className="truncate font-mono text-xs text-slate-400">
                          {permission.route}
                        </p>
                      </div>
                    </div>
                    <Switch
                      checked={checked}
                      onChange={() => toggle(permission)}
                      disabled={busy}
                      aria-label={`${checked ? 'Remove' : 'Assign'} permission ${permission.name}`}
                    />
                  </motion.li>
                )
              })}
            </ul>
          </div>
          <p className="mt-3 text-xs text-slate-400">
            <Badge variant="neutral">{selected.size} selected</Badge> Permissions map 1:1 to backend
            API routes.
          </p>
        </div>
      )}
    </Modal>
  )
}
