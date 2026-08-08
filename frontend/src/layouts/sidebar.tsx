import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronDown, Zap } from 'lucide-react'
import { cn } from '@/utils/cn'
import { navGroups } from './nav'
import { usePermission } from '@/hooks/use-permission'

export interface SidebarProps {
  collapsed: boolean
  onNavigate?: () => void
}

export function Sidebar({ collapsed, onNavigate }: SidebarProps) {
  const { isAdmin } = usePermission()
  const location = useLocation()
  const groups = navGroups.filter((group) => !group.adminOnly || isAdmin)
  const [expanded, setExpanded] = useState<Record<string, boolean>>(() =>
    Object.fromEntries(groups.map((group) => [group.label, true])),
  )

  const toggle = (label: string) =>
    setExpanded((previous) => ({ ...previous, [label]: !previous[label] }))

  return (
    <div className="flex h-full flex-col">
      <div className={cn('flex h-16 shrink-0 items-center gap-2.5 px-5', collapsed && 'justify-center px-0')}>
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-primary-500 to-violet-600 shadow-sm">
          <Zap className="h-4 w-4 text-white" fill="currentColor" />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <p className="truncate text-sm font-bold tracking-tight text-slate-900">
              Approval Flow
            </p>
            <p className="truncate text-[10px] font-medium uppercase tracking-widest text-slate-400">
              Admin Console
            </p>
          </div>
        )}
      </div>

      <nav className="scrollbar-thin flex-1 space-y-1 overflow-y-auto px-3 pb-4">
        {groups.map((group) => {
          const isExpanded = expanded[group.label] ?? true
          const showItems = collapsed || isExpanded

          return (
            <div key={group.label}>
              {!collapsed && (
                <button
                  type="button"
                  onClick={() => toggle(group.label)}
                  className="flex w-full items-center justify-between rounded-md px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-slate-400 transition-colors hover:text-slate-600"
                  aria-expanded={isExpanded}
                >
                  {group.label}
                  <ChevronDown
                    className={cn(
                      'h-3.5 w-3.5 transition-transform duration-200',
                      isExpanded ? 'rotate-0' : '-rotate-90',
                    )}
                  />
                </button>
              )}

              <AnimatePresence initial={false}>
                {showItems && (
                  <motion.ul
                    key={group.label}
                    initial={collapsed ? false : { height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2, ease: 'easeInOut' }}
                    className="overflow-hidden"
                  >
                    {group.items.map((item) => {
                      const isActive =
                        item.path === '/' ? location.pathname === '/' : location.pathname.startsWith(item.path)
                      return (
                        <li key={item.path} className="mt-0.5">
                          <NavLink
                            to={item.path}
                            end={item.path === '/'}
                            onClick={onNavigate}
                            title={collapsed ? item.label : undefined}
                            className={cn(
                              'relative flex items-center rounded-lg py-2 text-sm font-medium transition-colors duration-150',
                              collapsed ? 'justify-center px-0' : 'gap-3 px-3',
                              isActive
                                ? 'bg-primary-50 text-primary-700'
                                : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
                            )}
                          >
                            {isActive && (
                              <motion.span
                                layoutId="sidebar-active"
                                className="absolute left-0.5 top-1/2 h-4 w-1 -translate-y-1/2 rounded-full bg-primary-600"
                                transition={{ type: 'spring', stiffness: 500, damping: 35 }}
                              />
                            )}
                            <item.icon className="h-[18px] w-[18px] shrink-0" />
                            {!collapsed && <span className="truncate">{item.label}</span>}
                          </NavLink>
                        </li>
                      )
                    })}
                  </motion.ul>
                )}
              </AnimatePresence>
            </div>
          )
        })}
      </nav>

      {!collapsed && (
        <div className="border-t border-slate-100 px-3 py-3">
          <p className="px-3 text-[11px] leading-relaxed text-slate-400">
            Approval Flow Enterprise
            <br />
            Workflow Management System
          </p>
        </div>
      )}
    </div>
  )
}
