import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { Check, ChevronsUp, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Modal } from '@/components/ui/modal'
import { Textarea } from '@/components/ui/textarea'
import { useApprovalMutations } from '@/features/approvals/hooks/use-approvals'
import { cn } from '@/utils/cn'
import type { Approval } from '@/types/models'

type Decision = 'approved' | 'rejected' | 'escalated'

const decisionOptions: { value: Decision; label: string; description: string; className: string; icon: typeof Check }[] = [
  {
    value: 'approved',
    label: 'Approve',
    description: 'Accept and advance the application.',
    className: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    icon: Check,
  },
  {
    value: 'rejected',
    label: 'Reject',
    description: 'Decline and return the application.',
    className: 'border-rose-200 bg-rose-50 text-rose-700',
    icon: X,
  },
  {
    value: 'escalated',
    label: 'Escalate',
    description: 'Push to a higher approval level.',
    className: 'border-amber-200 bg-amber-50 text-amber-700',
    icon: ChevronsUp,
  },
]

export interface ApprovalDecideModalProps {
  approval: Approval | null
  onClose: () => void
}

export function ApprovalDecideModal({ approval, onClose }: ApprovalDecideModalProps) {
  const { decideApproval } = useApprovalMutations()
  const [decision, setDecision] = useState<Decision>('approved')
  const [comment, setComment] = useState('')

  useEffect(() => {
    if (approval) {
      setDecision((approval.decision as Decision) ?? 'approved')
      setComment(approval.comment ?? '')
    }
  }, [approval])

  const submit = () => {
    if (!approval) return
    decideApproval.mutate(
      { approval_id: approval.id, decision, comment: comment.trim() || undefined },
      { onSuccess: onClose },
    )
  }

  return (
    <Modal
      open={Boolean(approval)}
      onClose={onClose}
      title="Record decision"
      description={
        approval ? `Approval ${approval.id.slice(0, 8)} · currently ${approval.status}` : undefined
      }
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={decideApproval.isPending}>
            Cancel
          </Button>
          <Button
            onClick={submit}
            loading={decideApproval.isPending}
            className={cn(
              decision === 'rejected' && 'bg-rose-600 hover:bg-rose-700',
              decision === 'escalated' && 'bg-amber-600 hover:bg-amber-700',
            )}
          >
            Confirm {decision === 'approved' ? 'approval' : decision}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <div role="radiogroup" aria-label="Decision" className="grid gap-2.5 sm:grid-cols-3">
          {decisionOptions.map((option, index) => {
            const Icon = option.icon
            const active = decision === option.value
            return (
              <motion.button
                key={option.value}
                type="button"
                role="radio"
                aria-checked={active}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.05 }}
                onClick={() => setDecision(option.value)}
                className={cn(
                  'flex flex-col items-start gap-1.5 rounded-xl border p-3.5 text-left transition-all duration-150',
                  active
                    ? option.className
                    : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:bg-slate-50',
                )}
              >
                <Icon className="h-4 w-4" />
                <span className="text-sm font-semibold">{option.label}</span>
                <span className="text-xs opacity-70">{option.description}</span>
              </motion.button>
            )
          })}
        </div>
        <FormField label="Comment" htmlFor="decision-comment" hint="Visible to the applicant and approvers.">
          <Textarea
            id="decision-comment"
            rows={4}
            value={comment}
            onChange={(event) => setComment(event.target.value)}
            placeholder="Add a note explaining your decision…"
          />
        </FormField>
      </div>
    </Modal>
  )
}
