import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Loader2, ShieldCheck } from 'lucide-react'
import { Avatar } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/ui/empty-state'
import { Modal } from '@/components/ui/modal'
import { Switch } from '@/components/ui/switch'
import { rolesService } from '@/services/roles.service'
import { useUserMutations } from '@/features/users/hooks/use-users'
import type { Role, User } from '@/types/models'

export interface UserRolesModalProps {
  user: User | null
  onClose: () => void
}

export function UserRolesModal({ user, onClose }: UserRolesModalProps) {
  const { assignRole, removeRole } = useUserMutations()
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const { data: allRoles, isLoading } = useQuery({
    queryKey: ['roles', 'all'],
    queryFn: () => rolesService.list({ limit: 100 }),
    enabled: Boolean(user),
  })

  useEffect(() => {
    setSelected(new Set((user?.roles ?? []).map((role) => role.id)))
  }, [user])

  const toggle = (role: Role) => {
    if (!user) return
    const hasRole = selected.has(role.id)
    setSelected((previous) => {
      const next = new Set(previous)
      if (hasRole) next.delete(role.id)
      else next.add(role.id)
      return next
    })
    const mutation = hasRole ? removeRole : assignRole
    mutation.mutate(
      { userId: user.id, roleId: role.id },
      {
        onError: () => {
          // Roll back the optimistic toggle.
          setSelected((previous) => {
            const next = new Set(previous)
            if (hasRole) next.add(role.id)
            else next.delete(role.id)
            return next
          })
        },
      },
    )
  }

  const roles = allRoles?.rows ?? []

  return (
    <Modal
      open={Boolean(user)}
      onClose={onClose}
      title="Manage roles"
      description={
        user ? (
          <span className="flex items-center gap-2">
            <Avatar name={user.name} size="sm" />
            {user.name} · {user.email}
          </span>
        ) : undefined
      }
      size="md"
      footer={
        <Button variant="outline" onClick={onClose}>
          Done
        </Button>
      }
    >
      {isLoading ? (
        <div className="flex items-center justify-center py-10">
          <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
        </div>
      ) : roles.length === 0 ? (
        <EmptyState
          icon={ShieldCheck}
          title="No roles available"
          description="Create a role before assigning it to users."
        />
      ) : (
        <ul className="divide-y divide-slate-100">
          {roles.map((role, index) => {
            const checked = selected.has(role.id)
            const busy = assignRole.isPending || removeRole.isPending
            return (
              <motion.li
                key={role.id}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.03 }}
                className="flex items-center justify-between gap-3 py-3"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-medium text-slate-900">{role.name}</p>
                    {role.is_default && <Badge variant="neutral">default</Badge>}
                  </div>
                  <p className="mt-0.5 truncate text-xs text-slate-500">
                    {role.description || 'No description'}
                  </p>
                </div>
                <Switch
                  checked={checked}
                  onChange={() => toggle(role)}
                  disabled={busy}
                  aria-label={`${checked ? 'Remove' : 'Assign'} role ${role.name}`}
                />
              </motion.li>
            )
          })}
        </ul>
      )}
    </Modal>
  )
}
