import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronDown, KeyRound, LogOut, Menu, PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { Avatar } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { clearCredentials } from '@/store/slices/auth.slice'
import { setMobileSidebarOpen, toggleSidebar } from '@/store/slices/ui.slice'
import { usePermission } from '@/hooks/use-permission'
import { useToast } from '@/hooks/use-toast'
import { authService } from '@/services/auth.service'
import { titleCase } from '@/utils/format'
import { pageTitleFor } from './nav'
import { NotificationBell } from '@/features/notifications/components/notification-bell'
import { ChangePasswordModal } from '@/features/auth/components/change-password-modal'

export function Header() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const toast = useToast()
  const collapsed = useAppSelector((state) => state.ui.sidebarCollapsed)
  const user = useAppSelector((state) => state.auth.user)
  const { isAdmin, roles } = usePermission()
  const [menuOpen, setMenuOpen] = useState(false)
  const [passwordOpen, setPasswordOpen] = useState(false)

  const roleLabel = isAdmin ? 'Administrator' : roles[0] ? titleCase(roles[0]) : 'User'

  const handleLogout = async () => {
    setMenuOpen(false)
    try {
      if (user) await authService.logout(user.id)
    } catch {
      // Logout is best-effort; always clear the local session.
    }
    dispatch(clearCredentials())
    toast.info('Signed out', 'You have been signed out successfully.')
    navigate('/login', { replace: true })
  }

  return (
    <header className="sticky top-0 z-20 border-b border-slate-200/70 bg-white/80 backdrop-blur-xl">
      <div className="flex h-16 items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <div className="flex min-w-0 items-center gap-2">
          <button
            type="button"
            onClick={() => dispatch(setMobileSidebarOpen(true))}
            className="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 lg:hidden"
            aria-label="Open navigation"
          >
            <Menu className="h-5 w-5" />
          </button>
          <button
            type="button"
            onClick={() => dispatch(toggleSidebar())}
            className="hidden rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 lg:inline-flex"
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? <PanelLeftOpen className="h-[18px] w-[18px]" /> : <PanelLeftClose className="h-[18px] w-[18px]" />}
          </button>
          <div className="min-w-0">
            <h1 className="truncate text-sm font-semibold text-slate-900 sm:text-base">
              {pageTitleFor(location.pathname)}
            </h1>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <NotificationBell />

          <div className="mx-1 hidden h-6 w-px bg-slate-200 sm:block" />

          <div className="relative">
            <button
              type="button"
              onClick={() => setMenuOpen((value) => !value)}
              className="flex items-center gap-2.5 rounded-lg p-1.5 transition-colors hover:bg-slate-100"
              aria-haspopup="menu"
              aria-expanded={menuOpen}
            >
              <Avatar name={user?.name ?? '?'} size="sm" />
              <div className="hidden text-left sm:block">
                <p className="max-w-[140px] truncate text-sm font-medium leading-tight text-slate-900">
                  {user?.name}
                </p>
                <p className="text-xs leading-tight text-slate-500">{roleLabel}</p>
              </div>
              <ChevronDown className="hidden h-4 w-4 text-slate-400 sm:block" />
            </button>

            <AnimatePresence>
              {menuOpen && (
                <>
                  <div className="fixed inset-0 z-30" onClick={() => setMenuOpen(false)} aria-hidden />
                  <motion.div
                    role="menu"
                    initial={{ opacity: 0, y: 8, scale: 0.97 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: 6, scale: 0.97 }}
                    transition={{ duration: 0.15 }}
                    className="absolute right-0 top-full z-40 mt-2 w-60 overflow-hidden rounded-xl border border-slate-200/80 bg-white shadow-popover"
                  >
                    <div className="border-b border-slate-100 px-4 py-3">
                      <p className="truncate text-sm font-semibold text-slate-900">{user?.name}</p>
                      <p className="truncate text-xs text-slate-500">{user?.email}</p>
                      <div className="mt-2 flex flex-wrap gap-1">
                        {roles.map((role) => (
                          <Badge key={role} variant={role === 'admin' ? 'primary' : 'neutral'}>
                            {titleCase(role)}
                          </Badge>
                        ))}
                      </div>
                    </div>
                    <div className="p-1.5">
                      <button
                        type="button"
                        role="menuitem"
                        onClick={() => {
                          setMenuOpen(false)
                          setPasswordOpen(true)
                        }}
                        className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900"
                      >
                        <KeyRound className="h-4 w-4" />
                        Change password
                      </button>
                      <button
                        type="button"
                        role="menuitem"
                        onClick={handleLogout}
                        className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-rose-600 transition-colors hover:bg-rose-50"
                      >
                        <LogOut className="h-4 w-4" />
                        Sign out
                      </button>
                    </div>
                  </motion.div>
                </>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <ChangePasswordModal open={passwordOpen} onClose={() => setPasswordOpen(false)} />
    </header>
  )
}
