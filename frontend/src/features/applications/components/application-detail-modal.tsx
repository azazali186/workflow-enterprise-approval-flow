import { Code2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Modal } from '@/components/ui/modal'
import { PriorityBadge, StatusBadge } from '@/components/ui/status-badge'
import { formatDateTime, shortId } from '@/utils/format'
import type { Application } from '@/types/models'

interface DetailRowProps {
  label: string
  value: React.ReactNode
}

function DetailRow({ label, value }: DetailRowProps) {
  return (
    <div className="flex items-start justify-between gap-6 py-2.5">
      <dt className="w-32 shrink-0 text-xs font-medium uppercase tracking-wide text-slate-400">
        {label}
      </dt>
      <dd className="min-w-0 flex-1 text-right text-sm text-slate-700">{value}</dd>
    </div>
  )
}

export interface ApplicationDetailModalProps {
  application: Application | null
  onClose: () => void
}

export function ApplicationDetailModal({ application, onClose }: ApplicationDetailModalProps) {
  return (
    <Modal
      open={Boolean(application)}
      onClose={onClose}
      title="Application details"
      description={application ? `ID ${application.id}` : undefined}
      size="md"
      footer={
        <Button variant="outline" onClick={onClose}>
          Close
        </Button>
      }
    >
      {application && (
        <div>
          <dl className="divide-y divide-slate-100">
            <DetailRow label="Applicant" value={shortId(application.applicant_id, 12)} />
            <DetailRow label="Workflow" value={shortId(application.workflow_id, 12)} />
            <DetailRow label="Template" value={shortId(application.template_id, 12)} />
            <DetailRow label="Status" value={<StatusBadge value={application.status} />} />
            <DetailRow label="Priority" value={<PriorityBadge value={application.priority} />} />
            <DetailRow
              label="Submitted"
              value={formatDateTime(application.submitted_at ?? application.created_at)}
            />
            <DetailRow label="Updated" value={formatDateTime(application.updated_at)} />
          </dl>

          <div className="mt-4">
            <div className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-slate-400">
              <Code2 className="h-3.5 w-3.5" />
              Data payload
            </div>
            <pre className="scrollbar-thin mt-2 max-h-64 overflow-auto rounded-lg bg-slate-950 p-4 font-mono text-xs leading-relaxed text-slate-200">
              {application.data && Object.keys(application.data).length > 0
                ? JSON.stringify(application.data, null, 2)
                : '{}'}
            </pre>
          </div>
        </div>
      )}
    </Modal>
  )
}
