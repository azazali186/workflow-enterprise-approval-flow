import type { LucideIcon } from 'lucide-react'
import {
  AlertOctagon,
  BarChart3,
  Bell,
  CheckCircle2,
  FileText,
  FileType,
  History,
  KeyRound,
  LayoutDashboard,
  ShieldCheck,
  Users,
  Workflow,
} from 'lucide-react'

export interface NavItem {
  label: string
  path: string
  icon: LucideIcon
}

export interface NavGroup {
  label: string
  items: NavItem[]
  adminOnly?: boolean
}

export const navGroups: NavGroup[] = [
  {
    label: 'Overview',
    items: [{ label: 'Dashboard', path: '/', icon: LayoutDashboard }],
  },
  {
    label: 'Workflows',
    items: [
      { label: 'Applications', path: '/applications', icon: FileText },
      { label: 'Approvals', path: '/approvals', icon: CheckCircle2 },
      { label: 'Workflows', path: '/workflows', icon: Workflow },
      { label: 'Templates', path: '/templates', icon: FileType },
      { label: 'Escalations', path: '/escalations', icon: AlertOctagon },
    ],
  },
  {
    label: 'Admin',
    adminOnly: true,
    items: [
      { label: 'Users', path: '/users', icon: Users },
      { label: 'Roles', path: '/roles', icon: ShieldCheck },
      { label: 'Permissions', path: '/permissions', icon: KeyRound },
      { label: 'Login Logs', path: '/login-logs', icon: History },
    ],
  },
  {
    label: 'Insights',
    items: [
      { label: 'Notifications', path: '/notifications', icon: Bell },
      { label: 'Analytics', path: '/analytics', icon: BarChart3 },
    ],
  },
]

/** Resolve the current page title from the route for the header. */
export function pageTitleFor(pathname: string): string {
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) return 'Dashboard'
  const last = segments[segments.length - 1]
  return last
    .replace(/[-_]/g, ' ')
    .replace(/\w\S*/g, (word) => word.charAt(0).toUpperCase() + word.slice(1))
}
