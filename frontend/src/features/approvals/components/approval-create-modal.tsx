import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Combobox } from '@/components/ui/combobox'
import { useApplicationDropdown, useUserDropdown } from '@/hooks/use-dropdowns'
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

  // Fetch dropdown options using the common API
  const { data: applicationOptions, isLoading: isLoadingApplications } = useApplicationDropdown(['submitted'], { enabled: open })
  const { data: userOptions, isLoading: isLoadingUsers } = useUserDropdown({ enabled: open })

  const {
    register,
    handleSubmit,
    reset,
    control,
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
          <Controller
            name="application_id"
            control={control}
            render={({ field }) => (
              <Combobox
                id="application_id"
                options={applicationOptions || []}
                value={field.value}
                onChange={field.onChange}
                placeholder="Select application..."
                searchPlaceholder="Search applications..."
                disabled={isLoadingApplications}
                invalid={Boolean(errors.application_id)}
              />
            )}
          />
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
          <Controller
            name="approver_id"
            control={control}
            render={({ field }) => (
              <Combobox
                id="approver_id"
                options={userOptions || []}
                value={field.value}
                onChange={field.onChange}
                placeholder="Select approver..."
                searchPlaceholder="Search approvers..."
                disabled={isLoadingUsers}
                invalid={Boolean(errors.approver_id)}
              />
            )}
          />
        </FormField>
      </form>
    </Modal>
  )
}
