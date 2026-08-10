import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { useWorkflowMutations } from '@/features/workflows/hooks/use-workflows'
import { safeJsonParse } from '@/utils/format'
import type { Workflow } from '@/types/models'
import type { WorkflowStepPayload } from '@/services/workflows.service'

const schema = z
  .object({
    name: z.string().min(2, 'Name is required'),
    description: z.string().optional(),
    category: z.string().min(1, 'Category is required'),
    is_active: z.boolean(),
    steps_json: z.string().optional(),
  })
  .superRefine((values, ctx) => {
    if (values.steps_json?.trim()) {
      const parsed = safeJsonParse(values.steps_json)
      const valid =
        Array.isArray(parsed) &&
        parsed.every(
          (step) =>
            typeof step === 'object' &&
            step !== null &&
            typeof (step as { name?: unknown }).name === 'string' &&
            ((step as { name?: string }).name ?? '').length > 0,
        )
      if (!valid) {
        ctx.addIssue({
          code: 'custom',
          path: ['steps_json'],
          message: 'Must be a JSON array of steps, e.g. [{"name":"Manager review","approver_role":"manager"}]',
        })
      }
    }
  })

type FormValues = z.infer<typeof schema>

export interface WorkflowFormModalProps {
  open: boolean
  onClose: () => void
  workflow?: Workflow | null
}

export function WorkflowFormModal({ open, onClose, workflow }: WorkflowFormModalProps) {
  const isCreate = !workflow
  const { createWorkflow, updateWorkflow } = useWorkflowMutations()
  const resolver = useMemo(() => zodResolver(schema), [])

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver,
    defaultValues: { name: '', description: '', category: '', is_active: true, steps_json: '' },
  })

  useEffect(() => {
    if (open) {
      reset(
        workflow
          ? {
              name: workflow.name,
              description: workflow.description ?? '',
              category: workflow.category,
              is_active: workflow.is_active,
              steps_json: workflow.steps ? JSON.stringify(workflow.steps, null, 2) : '',
            }
          : { name: '', description: '', category: '', is_active: true, steps_json: '' },
      )
    }
  }, [open, workflow, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      const steps = values.steps_json?.trim()
        ? (safeJsonParse(values.steps_json) as WorkflowStepPayload[])
        : undefined
      if (isCreate) {
        await createWorkflow.mutateAsync({
          name: values.name,
          description: values.description || undefined,
          category: values.category,
          is_active: values.is_active,
          steps,
        })
      } else {
        await updateWorkflow.mutateAsync({
          workflow_id: workflow!.id,
          name: values.name,
          description: values.description || undefined,
          category: values.category,
          is_active: values.is_active,
          steps,
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
      title={isCreate ? 'Create workflow' : 'Edit workflow'}
      description="Workflows route applications through approval steps."
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="workflow-form" loading={isSubmitting}>
            {isCreate ? 'Create workflow' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="workflow-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Name" htmlFor="name" required error={errors.name?.message}>
            <Input id="name" placeholder="e.g. Expense approval" invalid={Boolean(errors.name)} {...register('name')} />
          </FormField>
          <FormField label="Category" htmlFor="category" required error={errors.category?.message}>
            <Input id="category" placeholder="e.g. Finance" invalid={Boolean(errors.category)} {...register('category')} />
          </FormField>
        </div>
        <FormField label="Description" htmlFor="description" error={errors.description?.message}>
          <Textarea id="description" rows={2} invalid={Boolean(errors.description)} {...register('description')} />
        </FormField>
        <FormField
          label="Steps (JSON)"
          htmlFor="steps_json"
          error={errors.steps_json?.message}
          hint="Advanced: define workflow steps as a JSON array. Each step needs a name; route it with approver_role (e.g. manager)."
        >
          <Textarea
            id="steps_json"
            rows={5}
            className="font-mono text-xs"
            placeholder='[{"name": "Manager review", "step_order": 1, "approver_role": "manager", "timeout_hours": 48}]'
            invalid={Boolean(errors.steps_json)}
            {...register('steps_json')}
          />
        </FormField>
        <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-slate-50/60 px-3.5 py-3">
          <div>
            <p className="text-sm font-medium text-slate-800">Active</p>
            <p className="text-xs text-slate-500">Only active workflows accept new submissions.</p>
          </div>
          <Switch
            checked={watch('is_active')}
            onChange={(checked) => setValue('is_active', checked, { shouldValidate: true })}
            aria-label="Active workflow"
          />
        </div>
      </form>
    </Modal>
  )
}
