import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Textarea } from '@/components/ui/textarea'
import { useTemplateMutations } from '@/features/templates/hooks/use-templates'
import { safeJsonParse } from '@/utils/format'
import type { Template } from '@/types/models'

const schema = z.object({
  name: z.string().min(2, 'Name is required'),
  category: z.string().min(1, 'Category is required'),
  schema_json: z.string().optional(),
  ui_json: z.string().optional(),
})
  .superRefine((values, ctx) => {
    for (const key of ['schema_json', 'ui_json'] as const) {
      const raw = values[key]
      if (raw?.trim()) {
        const parsed = safeJsonParse(raw)
        if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
          ctx.addIssue({
            code: 'custom',
            path: [key],
            message: 'Must be a valid JSON object',
          })
        }
      }
    }
  })

type FormValues = z.infer<typeof schema>

export interface TemplateFormModalProps {
  open: boolean
  onClose: () => void
  template?: Template | null
}

export function TemplateFormModal({ open, onClose, template }: TemplateFormModalProps) {
  const isCreate = !template
  const { createTemplate, updateTemplate } = useTemplateMutations()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', category: '', schema_json: '', ui_json: '' },
  })

  useEffect(() => {
    if (open) {
      reset(
        template
          ? {
              name: template.name,
              category: template.category,
              schema_json: template.schema ? JSON.stringify(template.schema, null, 2) : '',
              ui_json: template.ui ? JSON.stringify(template.ui, null, 2) : '',
            }
          : { name: '', category: '', schema_json: '', ui_json: '' },
      )
    }
  }, [open, template, reset])

  const onSubmit = async (values: FormValues) => {
    const schemaJson = values.schema_json?.trim()
      ? (safeJsonParse(values.schema_json) as Record<string, unknown>)
      : undefined
    const uiJson = values.ui_json?.trim()
      ? (safeJsonParse(values.ui_json) as Record<string, unknown>)
      : undefined
    try {
      if (isCreate) {
        await createTemplate.mutateAsync({
          name: values.name,
          category: values.category,
          schema: schemaJson,
          ui: uiJson,
        })
      } else {
        await updateTemplate.mutateAsync({
          template_id: template!.id,
          name: values.name,
          category: values.category,
          schema: schemaJson,
          ui: uiJson,
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
      title={isCreate ? 'Create template' : 'Edit template'}
      description="Templates define the data schema for applications."
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="template-form" loading={isSubmitting}>
            {isCreate ? 'Create template' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="template-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Name" htmlFor="name" required error={errors.name?.message}>
            <Input id="name" placeholder="e.g. Expense claim" invalid={Boolean(errors.name)} {...register('name')} />
          </FormField>
          <FormField label="Category" htmlFor="category" required error={errors.category?.message}>
            <Input id="category" placeholder="e.g. Finance" invalid={Boolean(errors.category)} {...register('category')} />
          </FormField>
        </div>
        <FormField label="Schema (JSON)" htmlFor="schema_json" error={errors.schema_json?.message} hint="Field definitions stored with the template.">
          <Textarea
            id="schema_json"
            rows={4}
            className="font-mono text-xs"
            placeholder='{"amount": {"type": "number", "required": true}}'
            invalid={Boolean(errors.schema_json)}
            {...register('schema_json')}
          />
        </FormField>
        <FormField label="UI (JSON)" htmlFor="ui_json" error={errors.ui_json?.message} hint="Optional rendering hints for the form.">
          <Textarea
            id="ui_json"
            rows={3}
            className="font-mono text-xs"
            placeholder='{"amount": {"label": "Amount (USD)"}}'
            invalid={Boolean(errors.ui_json)}
            {...register('ui_json')}
          />
        </FormField>
      </form>
    </Modal>
  )
}
