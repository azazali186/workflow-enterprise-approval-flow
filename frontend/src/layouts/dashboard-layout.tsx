import { Suspense } from 'react'
import { Outlet } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { cn } from '@/utils/cn'
import { useAppSelector, useAppDispatch } from '@/store/hooks'
import { setMobileSidebarOpen } from '@/store/slices/ui.slice'
import { useMediaQuery } from '@/hooks/use-media-query'
import { useRealtime } from '@/hooks/use-realtime'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { PageLoader } from '@/components/ui/page-loader'

export function DashboardLayout() {
  // Keep the authenticated realtime socket (notifications/approvals) alive
  // for the lifetime of the console.
  useRealtime()

  const collapsed = useAppSelector((state) => state.ui.sidebarCollapsed)
  const mobileOpen = useAppSelector((state) => state.ui.mobileSidebarOpen)
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const dispatch = useAppDispatch()

  const closeMobile = () => dispatch(setMobileSidebarOpen(false))

  return (
    <div className="min-h-dvh bg-slate-50">
      {/* Desktop sidebar */}
      {isDesktop && (
        <aside
          className={cn(
            'fixed inset-y-0 left-0 z-30 hidden border-r border-slate-200/70 bg-white/90 backdrop-blur-xl transition-[width] duration-200 lg:flex lg:flex-col',
            collapsed ? 'w-[68px]' : 'w-60',
          )}
        >
          <Sidebar collapsed={collapsed} />
        </aside>
      )}

      {/* Mobile drawer */}
      <AnimatePresence>
        {mobileOpen && !isDesktop && (
          <>
            <motion.div
              className="fixed inset-0 z-40 bg-slate-950/40 backdrop-blur-[2px] lg:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              onClick={closeMobile}
              aria-hidden
            />
            <motion.aside
              className="fixed inset-y-0 left-0 z-50 w-72 border-r border-slate-200/70 bg-white shadow-float lg:hidden"
              initial={{ x: -288 }}
              animate={{ x: 0 }}
              exit={{ x: -288 }}
              transition={{ type: 'spring', stiffness: 380, damping: 36 }}
            >
              <Sidebar collapsed={false} onNavigate={closeMobile} />
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Content */}
      <div
        className={cn(
          'transition-[padding] duration-200',
          isDesktop && (collapsed ? 'lg:pl-[68px]' : 'lg:pl-60'),
        )}
      >
        <Header />
        <main className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
          <Suspense fallback={<PageLoader />}>
            <Outlet />
          </Suspense>
        </main>
      </div>
    </div>
  )
}
