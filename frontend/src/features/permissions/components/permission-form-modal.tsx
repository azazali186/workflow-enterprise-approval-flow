import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Select } from '@/components/ui/select'
import type { Permission } from '@/types/models'
import { usePermissionMutations } from '@/features/permissions/hooks/use-permissions'

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  route: z.string().min(1, 'Route is required'),
  path: z.string().min(1, 'Path is required'),
  method: z.enum(['POST', 'PATCH', 'DELETE']),
  service: z.string().min(1, 'Service is required'),
})

type FormValues = z.infer<typeof schema>

export interface PermissionFormModalProps {
  open: boolean
  onClose: () => void
  permission?: Permission | null
}

export function PermissionFormModal({ open, onClose, permission }: PermissionFormModalProps) {
  const isCreate = !permission
  const { createPermission, updatePermission } = usePermissionMutations()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', route: '', path: '', method: 'POST', service: 'approval-flow' },
  })

  useEffect(() => {
    if (open) {
      reset(
        permission
          ? {
              name: permission.name,
              route: permission.route,
              path: permission.path,
              method: permission.method as FormValues['method'],
              service: permission.service,
            }
          : { name: '', route: '', path: '', method: 'POST', service: 'approval-flow' },
      )
    }
  }, [open, permission, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      if (isCreate) {
        await createPermission.mutateAsync(values)
      } else {
        await updatePermission.mutateAsync({ permission_id: permission!.id, ...values })
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
      title={isCreate ? 'Create permission' : 'Edit permission'}
      description="Permissions map 1:1 to backend API routes."
      size="sm"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="permission-form" loading={isSubmitting}>
            {isCreate ? 'Create permission' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="permission-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <FormField label="Name" htmlFor="name" required error={errors.name?.message}>
          <Input id="name" placeholder="e.g. List Applications" invalid={Boolean(errors.name)} {...register('name')} />
        </FormField>
        <FormField label="Route" htmlFor="route" required error={errors.route?.message} hint="Format: METHOD /path, e.g. POST /api/v1/applications">
          <Input
            id="route"
            placeholder="POST /api/v1/applications"
            invalid={Boolean(errors.route)}
            {...register('route')}
          />
        </FormField>
        <FormField label="Path" htmlFor="path" required error={errors.path?.message}>
          <Input id="path" placeholder="/api/v1/applications" invalid={Boolean(errors.path)} {...register('path')} />
        </FormField>
        <div className="grid grid-cols-2 gap-3">
          <FormField label="Method" htmlFor="method" required error={errors.method?.message}>
            <Select id="method" invalid={Boolean(errors.method)} {...register('method')}>
              <option value="POST">POST</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </Select>
          </FormField>
          <FormField label="Service" htmlFor="service" required error={errors.service?.message}>
            <Input id="service" placeholder="approval-flow" invalid={Boolean(errors.service)} {...register('service')} />
          </FormField>
        </div>
      </form>
    </Modal>
  )
}
