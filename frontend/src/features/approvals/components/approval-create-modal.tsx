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
import { applicationsService } from '@/services/applications.service'
import { usersService } from '@/services/users.service'
import { useApprovalMutations } from '@/features/approvals/hooks/use-approvals'

const schema = z.object({
  application_id: z.string().min(1, 'Application is required'),
  workflow_step_id: z.string().min(1, 'Workflow step is required'),
  approver_id: z.string().min(1, 'Approver is required'),
})

type FormValues = z.infer<typeof schema>

export interface ApprovalCreateModalProps {
  open: boolean
  onClose: () => void
}

export function ApprovalCreateModal({ open, onClose }: ApprovalCreateModalProps) {
  const { createApproval } = useApprovalMutations()

  const { data: applications } = useQuery({
    queryKey: ['applications', 'options'],
    queryFn: () => applicationsService.list({ limit: 100, status: 'submitted' }),
    enabled: open,
  })

  const { data: approvers } = useQuery({
    queryKey: ['users', 'approver-options'],
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
    defaultValues: { application_id: '', workflow_step_id: '', approver_id: '' },
  })

  useEffect(() => {
    if (open) reset({ application_id: '', workflow_step_id: '', approver_id: '' })
  }, [open, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      await createApproval.mutateAsync(values)
      onClose()
    } catch {
      // Errors are surfaced by the mutation toast.
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create approval"
      description="Assign a workflow step to an approver for a submitted application."
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="approval-create-form" loading={isSubmitting}>
            Create approval
          </Button>
        </>
      }
    >
      <form id="approval-create-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <FormField label="Application" htmlFor="application_id" required error={errors.application_id?.message}>
          <Select id="application_id" invalid={Boolean(errors.application_id)} {...register('application_id')}>
            <option value="" disabled hidden>
              Select application…
            </option>
            {(applications?.rows ?? []).map((application) => (
              <option key={application.id} value={application.id}>
                {application.id.slice(0, 8)} · {application.status}
              </option>
            ))}
          </Select>
        </FormField>
        <FormField
          label="Workflow step ID"
          htmlFor="workflow_step_id"
          required
          error={errors.workflow_step_id?.message}
          hint="The UUID of the workflow step this approval applies to."
        >
          <Input
            id="workflow_step_id"
            placeholder="00000000-0000-0000-0000-000000000000"
            className="font-mono text-xs"
            invalid={Boolean(errors.workflow_step_id)}
            {...register('workflow_step_id')}
          />
        </FormField>
        <FormField label="Approver" htmlFor="approver_id" required error={errors.approver_id?.message}>
          <Select id="approver_id" invalid={Boolean(errors.approver_id)} {...register('approver_id')}>
            <option value="" disabled hidden>
              Select approver…
            </option>
            {(approvers?.rows ?? []).map((user) => (
              <option key={user.id} value={user.id}>
                {user.name} · {user.email}
              </option>
            ))}
          </Select>
        </FormField>
      </form>
    </Modal>
  )
}
