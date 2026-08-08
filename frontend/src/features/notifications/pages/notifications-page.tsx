import { useEffect, useState } from 'react'
import { Bell, CheckCheck, Send } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { Select } from '@/components/ui/select'
import {
  useNotificationsTable,
  useNotificationMutations,
} from '@/features/notifications/hooks/use-notifications'
import { SendNotificationModal } from '@/features/notifications/components/send-notification-modal'
import { formatDateTime, titleCase, truncate } from '@/utils/format'
import type { Notification } from '@/types/models'

export default function NotificationsPage() {
  const table = useNotificationsTable()
  const { markRead } = useNotificationMutations()
  const [typeFilter, setTypeFilter] = useState('')
  const [sendOpen, setSendOpen] = useState(false)

  useEffect(() => {
    table.setFilters(typeFilter ? { type: typeFilter } : {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [typeFilter])

  const columns: ColumnDef<Notification>[] = [
    {
      key: 'title',
      header: 'Notification',
      render: (notification) => (
        <div className="flex items-start gap-2.5">
          <span
            className={
              notification.read_at
                ? 'mt-1.5 h-2 w-2 shrink-0 rounded-full bg-slate-200'
                : 'mt-1.5 h-2 w-2 shrink-0 rounded-full bg-primary-500'
            }
          />
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{notification.title}</p>
            <p className="mt-0.5 truncate text-xs text-slate-500">
              {truncate(notification.body, 72)}
            </p>
          </div>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      render: (notification) => <Badge variant="primary">{titleCase(notification.type)}</Badge>,
    },
    {
      key: 'channel',
      header: 'Channel',
      hideBelow: 'sm',
      render: (notification) => <Badge variant="neutral">{notification.channel}</Badge>,
    },
    {
      key: 'read_at',
      header: 'Read',
      hideBelow: 'md',
      render: (notification) =>
        notification.read_at ? <Badge variant="success">Read</Badge> : <Badge variant="warning">Unread</Badge>,
    },
    {
      key: 'sent_at',
      header: 'Sent',
      hideBelow: 'lg',
      sortable: true,
      render: (notification) => (
        <span className="text-slate-500">
          {formatDateTime(notification.sent_at ?? notification.created_at)}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (notification) =>
        notification.read_at ? (
          <span className="text-xs text-slate-300">Read</span>
        ) : (
          <Button
            variant="outline"
            size="sm"
            onClick={(event) => {
              event.stopPropagation()
              markRead.mutate(notification.id)
            }}
            loading={markRead.isPending}
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Mark read
          </Button>
        ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Insights"
        title="Notifications"
        description="Messages delivered to users across all channels."
        actions={
          <Button onClick={() => setSendOpen(true)}>
            <Send className="h-4 w-4" />
            <span className="hidden sm:inline">Send notification</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(notification) => notification.id}
        noun="notifications"
        searchPlaceholder="Search notifications…"
        emptyTitle="No notifications yet"
        emptyDescription="Notifications appear here when they are sent to users."
        toolbar={
          <Select
            value={typeFilter}
            onChange={(event) => setTypeFilter(event.target.value)}
            placeholder="All types"
            className="w-32"
            aria-label="Filter by type"
          >
            <option value="info">Info</option>
            <option value="approval">Approval</option>
            <option value="escalation">Escalation</option>
            <option value="system">System</option>
            <option value="alert">Alert</option>
          </Select>
        }
      />

      <SendNotificationModal open={sendOpen} onClose={() => setSendOpen(false)} />
    </div>
  )
}
