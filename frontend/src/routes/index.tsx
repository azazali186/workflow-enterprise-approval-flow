import { lazy, Suspense } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { DashboardLayout } from '@/layouts/dashboard-layout'
import { AdminRoute, ProtectedRoute } from './protected-route'
import { Spinner } from '@/components/ui/spinner'
import { AppErrorFallback } from '@/components/ui/app-error-fallback'

// Renders when a route element (or a lazy chunk) throws. React Router's data
// router catches these internally and renders the nearest errorElement — the
// error never reaches outer React error boundaries, so this is the only place
// a page crash can get a styled, recoverable fallback.
function RouteErrorElement() {
  return <AppErrorFallback />
}

const LoginPage = lazy(() => import('@/features/auth/pages/login-page'))
const DashboardPage = lazy(() => import('@/features/dashboard/pages/dashboard-page'))
const ApplicationsPage = lazy(() => import('@/features/applications/pages/applications-page'))
const ApprovalsPage = lazy(() => import('@/features/approvals/pages/approvals-page'))
const WorkflowsPage = lazy(() => import('@/features/workflows/pages/workflows-page'))
const TemplatesPage = lazy(() => import('@/features/templates/pages/templates-page'))
const EscalationsPage = lazy(() => import('@/features/escalations/pages/escalations-page'))
const UsersPage = lazy(() => import('@/features/users/pages/users-page'))
const RolesPage = lazy(() => import('@/features/roles/pages/roles-page'))
const PermissionsPage = lazy(() => import('@/features/permissions/pages/permissions-page'))
const LoginLogsPage = lazy(() => import('@/features/login-logs/pages/login-logs-page'))
const NotificationsPage = lazy(() => import('@/features/notifications/pages/notifications-page'))
const AnalyticsPage = lazy(() => import('@/features/analytics/pages/analytics-page'))

function LoginLoader() {
  return (
    <div className="flex min-h-dvh items-center justify-center">
      <Spinner size="lg" />
    </div>
  )
}

export const router = createBrowserRouter([
  {
    path: '/login',
    errorElement: <RouteErrorElement />,
    element: (
      <Suspense fallback={<LoginLoader />}>
        <LoginPage />
      </Suspense>
    ),
  },
  {
    path: '/',
    errorElement: <RouteErrorElement />,
    element: (
      <ProtectedRoute>
        <DashboardLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'applications', element: <ApplicationsPage /> },
      { path: 'approvals', element: <ApprovalsPage /> },
      { path: 'workflows', element: <WorkflowsPage /> },
      { path: 'templates', element: <TemplatesPage /> },
      { path: 'escalations', element: <EscalationsPage /> },
      { path: 'notifications', element: <NotificationsPage /> },
      { path: 'analytics', element: <AnalyticsPage /> },
      { path: 'users', element: <AdminRoute><UsersPage /></AdminRoute> },
      { path: 'roles', element: <AdminRoute><RolesPage /></AdminRoute> },
      { path: 'permissions', element: <AdminRoute><PermissionsPage /></AdminRoute> },
      { path: 'login-logs', element: <AdminRoute><LoginLogsPage /></AdminRoute> },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
])
