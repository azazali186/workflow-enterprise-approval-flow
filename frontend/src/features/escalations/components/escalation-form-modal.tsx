import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { approvalsService } from '@/services/approvals.service'
import { usersService } from '@/services/users.service'
import { useEscalationMutations } from '@/features/escalations/hooks/use-escalations'

const schema = z.object({
  approval_id: z.string().min(1, 'Approval is required'),
  level: z
    .number('Level is required')
    .int('Level must be a whole number')
    .min(0, 'Minimum level is 0')
    .max(10, 'Maximum level is 10'),
  escalated_to: z.string().min(1, 'Assignee is required'),
  reason: z.string().min(3, 'Reason is required'),
})

type FormValues = z.infer<typeof schema>

export interface EscalationFormModalProps {
  open: boolean
  onClose: () => void
}

export function EscalationFormModal({ open, onClose }: EscalationFormModalProps) {
  const { createEscalation } = useEscalationMutations()

  const { data: approvals } = useQuery({
    queryKey: ['approvals', 'options'],
    queryFn: () => approvalsService.list({ limit: 100 }),
    enabled: open,
  })

  const { data: users } = useQuery({
    queryKey: ['users', 'assignee-options'],
    queryFn: () => usersService.list({ limit: 100 }),
    enabled: open,
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { approval_id: '', level: 1, escalated_to: '', reason: '' },
  })

  useEffect(() => {
    if (open) reset({ approval_id: '', level: 1, escalated_to: '', reason: '' })
  }, [open, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      await createEscalation.mutateAsync(values)
      onClose()
    } catch {
      // Errors are surfaced by the mutation toast.
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create escalation"
      description="Escalate an approval to a higher-level assignee."
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="escalation-form" loading={isSubmitting}>
            Create escalation
          </Button>
        </>
      }
    >
      <form id="escalation-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Approval" htmlFor="approval_id" required error={errors.approval_id?.message}>
            <Select id="approval_id" invalid={Boolean(errors.approval_id)} {...register('approval_id')}>
              <option value="" disabled hidden>
                Select approval…
              </option>
              {(approvals?.rows ?? []).map((approval) => (
                <option key={approval.id} value={approval.id}>
                  {approval.id.slice(0, 8)} · {approval.status}
                </option>
              ))}
            </Select>
          </FormField>
          <FormField label="Level" htmlFor="level" required error={errors.level?.message} hint="0–10">
            <Input
              id="level"
              type="number"
              min={0}
              max={10}
              invalid={Boolean(errors.level)}
              {...register('level', { valueAsNumber: true })}
            />
          </FormField>
        </div>
        <FormField label="Escalate to" htmlFor="escalated_to" required error={errors.escalated_to?.message}>
          <Select id="escalated_to" invalid={Boolean(errors.escalated_to)} {...register('escalated_to')}>
            <option value="" disabled hidden>
              Select assignee…
            </option>
            {(users?.rows ?? []).map((user) => (
              <option key={user.id} value={user.id}>
                {user.name} · {user.email}
              </option>
            ))}
          </Select>
        </FormField>
        <FormField label="Reason" htmlFor="reason" required error={errors.reason?.message}>
          <Textarea
            id="reason"
            rows={3}
            invalid={Boolean(errors.reason)}
            placeholder="Explain why this needs escalation…"
            {...register('reason')}
          />
        </FormField>
      </form>
    </Modal>
  )
}
