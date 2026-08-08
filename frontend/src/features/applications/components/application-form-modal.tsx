import { useEffect, useMemo } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Select } from '@/components/ui/select'
import { Combobox } from '@/components/ui/combobox'
import { Textarea } from '@/components/ui/textarea'
import { useUserDropdown, useWorkflowDropdown, useTemplateDropdown } from '@/hooks/use-dropdowns'
import { usePermission } from '@/hooks/use-permission'
import { useAppSelector } from '@/store/hooks'
import { useApplicationMutations } from '@/features/applications/hooks/use-applications'
import { safeJsonParse } from '@/utils/format'
import type { Application } from '@/types/models'

const buildSchema = (isSubmit: boolean) =>
  z
    .object({
      applicant_id: z.string().min(1, 'Applicant is required'),
      workflow_id: z.string().min(1, 'Workflow is required'),
      template_id: z.string().min(1, 'Template is required'),
      title: z.string().min(2, 'Title is required'),
      description: z.string().optional(),
      priority: z.enum(['low', 'medium', 'high', 'urgent']),
      status: z.string().optional(),
      data_json: z.string().optional(),
    })
    .superRefine((values, ctx) => {
      if (values.data_json?.trim()) {
        const parsed = safeJsonParse(values.data_json)
        if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
          ctx.addIssue({ code: 'custom', path: ['data_json'], message: 'Must be a valid JSON object' })
        }
      }
    })

type FormValues = z.infer<ReturnType<typeof buildSchema>>

export interface ApplicationFormModalProps {
  open: boolean
  onClose: () => void
  application?: Application | null
}

export function ApplicationFormModal({ open, onClose, application }: ApplicationFormModalProps) {
  const isSubmit = !application
  const { submitApplication, updateApplication } = useApplicationMutations()
  const { isAdmin } = usePermission()
  const currentUser = useAppSelector((state) => state.auth.user)
  const schema = useMemo(() => buildSchema(isSubmit), [isSubmit])

  // Fetch dropdown options using the common API
  const { data: userOptions, isLoading: isLoadingUsers } = useUserDropdown({ enabled: open && isSubmit && isAdmin })
  const { data: workflowOptions, isLoading: isLoadingWorkflows } = useWorkflowDropdown(false, { enabled: open && isSubmit })
  const { data: templateOptions, isLoading: isLoadingTemplates } = useTemplateDropdown({ enabled: open && isSubmit })

  const {
    register,
    handleSubmit,
    reset,
    watch,
    control,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      applicant_id: currentUser?.id ?? '',
      workflow_id: '',
      template_id: '',
      title: '',
      description: '',
      priority: 'medium',
      status: '',
      data_json: '',
    },
  })

  useEffect(() => {
    if (open) {
      reset(
        application
          ? {
              applicant_id: application.applicant_id,
              workflow_id: application.workflow_id,
              template_id: application.template_id,
              title: '',
              description: '',
              priority: application.priority as FormValues['priority'],
              status: application.status,
              data_json: application.data ? JSON.stringify(application.data, null, 2) : '',
            }
          : {
              applicant_id: currentUser?.id ?? '',
              workflow_id: '',
              template_id: '',
              title: '',
              description: '',
              priority: 'medium',
              status: '',
              data_json: '',
            },
      )
    }
  }, [open, application, currentUser?.id, reset])

  const onSubmit = async (values: FormValues) => {
    const data = values.data_json?.trim() ? (safeJsonParse(values.data_json) as Record<string, unknown>) : undefined
    try {
      if (isSubmit) {
        await submitApplication.mutateAsync({
          applicant_id: values.applicant_id,
          workflow_id: values.workflow_id,
          template_id: values.template_id,
          title: values.title,
          description: values.description || undefined,
          priority: values.priority,
          data,
        })
      } else {
        await updateApplication.mutateAsync({
          application_id: application!.id,
          status: values.status || undefined,
          priority: values.priority,
          data,
        })
      }
      onClose()
    } catch {
      // Errors are surfaced by the mutation toast.
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={isSubmit ? 'Submit application' : 'Update application'}
      description={
        isSubmit
          ? 'Routed through the selected workflow for approval.'
          : `Update ${application ? `application ${application.id.slice(0, 8)}` : 'this application'}.`
      }
      size="lg"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="application-form" loading={isSubmitting}>
            {isSubmit ? 'Submit application' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="application-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {isSubmit ? (
          <>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField label="Applicant" htmlFor="applicant_id" required error={errors.applicant_id?.message}>
                <Controller
                  name="applicant_id"
                  control={control}
                  render={({ field }) => (
                    <Combobox
                      id="applicant_id"
                      options={userOptions || []}
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="Select applicant..."
                      searchPlaceholder="Search users..."
                      disabled={isLoadingUsers}
                      invalid={Boolean(errors.applicant_id)}
                    />
                  )}
                />
              </FormField>
              <FormField label="Priority" htmlFor="priority" required error={errors.priority?.message}>
                <Select id="priority" invalid={Boolean(errors.priority)} {...register('priority')}>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="urgent">Urgent</option>
                </Select>
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField label="Workflow" htmlFor="workflow_id" required error={errors.workflow_id?.message}>
                <Controller
                  name="workflow_id"
                  control={control}
                  render={({ field }) => (
                    <Combobox
                      id="workflow_id"
                      options={workflowOptions || []}
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="Select workflow..."
                      searchPlaceholder="Search workflows..."
                      disabled={isLoadingWorkflows}
                      invalid={Boolean(errors.workflow_id)}
                    />
                  )}
                />
              </FormField>
              <FormField label="Template" htmlFor="template_id" required error={errors.template_id?.message}>
                <Controller
                  name="template_id"
                  control={control}
                  render={({ field }) => (
                    <Combobox
                      id="template_id"
                      options={templateOptions || []}
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="Select template..."
                      searchPlaceholder="Search templates..."
                      disabled={isLoadingTemplates}
                      invalid={Boolean(errors.template_id)}
                    />
                  )}
                />
              </FormField>
            </div>
            <FormField label="Title" htmlFor="title" required error={errors.title?.message}>
              <Input id="title" placeholder="e.g. Travel expense reimbursement" invalid={Boolean(errors.title)} {...register('title')} />
            </FormField>
            <FormField label="Description" htmlFor="description" error={errors.description?.message}>
              <Textarea id="description" rows={2} invalid={Boolean(errors.description)} {...register('description')} />
            </FormField>
          </>
        ) : (
          <>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField label="Status" htmlFor="status" error={errors.status?.message}>
                <Select id="status" invalid={Boolean(errors.status)} {...register('status')}>
                  <option value="" disabled hidden>
                    Keep current
                  </option>
                  <option value="draft">Draft</option>
                  <option value="submitted">Submitted</option>
                  <option value="in_review">In review</option>
                  <option value="approved">Approved</option>
                  <option value="rejected">Rejected</option>
                  <option value="completed">Completed</option>
                  <option value="cancelled">Cancelled</option>
                </Select>
              </FormField>
              <FormField label="Priority" htmlFor="priority" required error={errors.priority?.message}>
                <Select id="priority" invalid={Boolean(errors.priority)} {...register('priority')}>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="urgent">Urgent</option>
                </Select>
              </FormField>
            </div>
            <div className="rounded-lg border border-slate-200 bg-slate-50/60 px-3.5 py-3 text-xs text-slate-500">
              Application <span className="font-mono">{application?.id}</span> · status{' '}
              <span className="font-medium">{application?.status}</span> (submitted{' '}
              {application?.submitted_at ? 'yes' : 'no'})
            </div>
          </>
        )}
        <FormField
          label="Data (JSON)"
          htmlFor="data_json"
          error={errors.data_json?.message}
          hint="Optional structured payload stored with the application."
        >
          <Textarea
            id="data_json"
            rows={watch('data_json') ? 6 : 3}
            className="font-mono text-xs"
            placeholder='{"department": "Finance"}'
            invalid={Boolean(errors.data_json)}
            {...register('data_json')}
          />
        </FormField>
      </form>
    </Modal>
  )
}
