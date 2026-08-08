import { useEffect, type ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAppSelector } from '@/store/hooks'
import { usePermission } from '@/hooks/use-permission'
import { useToast } from '@/hooks/use-toast'

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated)
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }
  return <>{children}</>
}

export function AdminRoute({ children }: { children: ReactNode }) {
  const { isAdmin } = usePermission()
  const toast = useToast()
  const location = useLocation()

  useEffect(() => {
    if (!isAdmin) {
      toast.error('Admin access required', 'You do not have permission to view this page.')
    }
  }, [isAdmin, toast])

  if (!isAdmin) return <Navigate to="/" replace />
  return <>{children}</>
}
