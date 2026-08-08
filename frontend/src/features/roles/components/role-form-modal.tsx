import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import type { Role } from '@/types/models'
import { useRoleMutations } from '@/features/roles/hooks/use-roles'

const schema = z.object({
  name: z
    .string()
    .min(2, 'Role name must be at least 2 characters')
    .max(50, 'Role name must be at most 50 characters'),
  description: z.string().max(500, 'Description must be at most 500 characters').optional(),
  is_default: z.boolean(),
})

type FormValues = z.infer<typeof schema>

export interface RoleFormModalProps {
  open: boolean
  onClose: () => void
  role?: Role | null
}

export function RoleFormModal({ open, onClose, role }: RoleFormModalProps) {
  const isCreate = !role
  const { createRole, updateRole } = useRoleMutations()
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', description: '', is_default: false },
  })

  useEffect(() => {
    if (open) {
      reset(
        role
          ? { name: role.name, description: role.description ?? '', is_default: role.is_default }
          : { name: '', description: '', is_default: false },
      )
    }
  }, [open, role, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      if (isCreate) {
        await createRole.mutateAsync({
          name: values.name,
          description: values.description || undefined,
          is_default: values.is_default,
        })
      } else {
        await updateRole.mutateAsync({ role_id: role!.id, ...values })
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
      title={isCreate ? 'Create role' : 'Edit role'}
      description={
        isCreate
          ? 'Roles group permissions that you can assign to users.'
          : `Update ${role?.name ?? 'this role'}.`
      }
      size="sm"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="role-form" loading={isSubmitting}>
            {isCreate ? 'Create role' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="role-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <FormField label="Role name" htmlFor="name" required error={errors.name?.message}>
          <Input id="name" placeholder="e.g. finance-manager" invalid={Boolean(errors.name)} {...register('name')} />
        </FormField>
        <FormField
          label="Description"
          htmlFor="description"
          error={errors.description?.message}
          hint="A short summary of what this role can do."
        >
          <Textarea
            id="description"
            rows={3}
            invalid={Boolean(errors.description)}
            {...register('description')}
          />
        </FormField>
        <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-slate-50/60 px-3.5 py-3">
          <div>
            <p className="text-sm font-medium text-slate-800">Default role</p>
            <p className="text-xs text-slate-500">Automatically assigned to new registrations.</p>
          </div>
          <Switch
            checked={watch('is_default')}
            onChange={(checked) => setValue('is_default', checked, { shouldValidate: true })}
            aria-label="Default role"
          />
        </div>
      </form>
    </Modal>
  )
}
