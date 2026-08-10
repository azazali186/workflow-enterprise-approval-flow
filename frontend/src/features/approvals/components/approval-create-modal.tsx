import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Modal } from '@/components/ui/modal'
import { Combobox } from '@/components/ui/combobox'
import { useApplicationDropdown, useUserDropdown, useWorkflowStepDropdown } from '@/hooks/use-dropdowns'
import { useApprovalMutations } from '@/features/approvals/hooks/use-approvals'
import { applicationsService } from '@/services/applications.service'
import { useQuery } from '@tanstack/react-query'

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

  // Fetch dropdown options using the common id/name API.
  const { data: applicationOptions, isLoading: isLoadingApplications } = useApplicationDropdown(['submitted'], { enabled: open })
  const { data: userOptions, isLoading: isLoadingUsers } = useUserDropdown({ enabled: open })

  const {
    handleSubmit,
    reset,
    setValue,
    control,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { application_id: '', workflow_step_id: '', approver_id: '' },
  })

  const applicationId = watch('application_id')

  // Resolve the selected application's workflow so the step dropdown is scoped
  // to it — a manual workflow_step_id input would let an approval point at a
  // step belonging to a different workflow than the application.
  const { data: application, isLoading: isLoadingApplication } = useQuery({
    queryKey: ['applications', 'get', applicationId],
    queryFn: () => applicationsService.get(applicationId),
    enabled: open && Boolean(applicationId),
    staleTime: 5 * 60 * 1000,
  })

  const workflowId = application?.workflow_id
  const {
    data: stepOptions,
    isLoading: isLoadingSteps,
  } = useWorkflowStepDropdown(workflowId, { enabled: open })

  useEffect(() => {
    if (open) reset({ application_id: '', workflow_step_id: '', approver_id: '' })
  }, [open, reset])

  // When the application changes, the previously chosen step may belong to a
  // different workflow — clear it the moment the application changes (not when
  // its workflow resolves) so a stale step can never be submitted while the
  // new application's detail is still loading.
  useEffect(() => {
    setValue('workflow_step_id', '')
  }, [applicationId, setValue])

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
          label="Workflow step"
          htmlFor="workflow_step_id"
          required
          error={errors.workflow_step_id?.message}
          hint={workflowId ? undefined : 'Select an application first to load its workflow steps.'}
        >
          <Controller
            name="workflow_step_id"
            control={control}
            render={({ field }) => (
              <Combobox
                id="workflow_step_id"
                options={stepOptions || []}
                value={field.value}
                onChange={field.onChange}
                placeholder={workflowId ? 'Select workflow step...' : 'Select an application first'}
                searchPlaceholder="Search steps..."
                // Disabled while the selected application's workflow is being
                // resolved — otherwise the previous application's steps could
                // stay selectable and a cross-workflow step get submitted.
                disabled={isLoadingApplication || !workflowId || isLoadingSteps}
                invalid={Boolean(errors.workflow_step_id)}
                emptyText="No steps for this workflow"
              />
            )}
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
