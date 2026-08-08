import { useMemo } from 'react'
import { useAppSelector } from '@/store/hooks'

/**
 * Role-based access helpers. The backend derives UI-relevant roles from the
 * authenticated user's `roles[]` (returned by /auth/login and /auth/refresh).
 * Admin-only surfaces are additionally guarded by <AdminRoute>.
 */
export function usePermission() {
  const user = useAppSelector((state) => state.auth.user)

  const roles = useMemo(
    () => (user?.roles ?? []).map((role) => role.name).filter(Boolean),
    [user],
  )

  return useMemo(
    () => ({
      roles,
      isAdmin: roles.includes('admin'),
      hasRole: (role: string) => roles.includes(role),
    }),
    [roles],
  )
}
