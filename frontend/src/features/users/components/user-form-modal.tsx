import { useEffect, useMemo } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Select } from '@/components/ui/select'
import type { User } from '@/types/models'
import { useUserMutations } from '@/features/users/hooks/use-users'
import { useToast } from '@/hooks/use-toast'
import { toErrorMessage } from '@/services/api/errors'

const buildSchema = (isCreate: boolean) =>
  z
    .object({
      name: z.string().min(2, 'Name must be at least 2 characters'),
      email: z.string().email('Enter a valid email address'),
      password: z.string().optional(),
      status: z.enum(['active', 'inactive', 'locked', 'pending']),
    })
    .superRefine((values, ctx) => {
      if (!isCreate) return
      if (!values.password || values.password.length < 8) {
        ctx.addIssue({
          code: 'custom',
          path: ['password'],
          message: 'Password must be at least 8 characters',
        })
      }
    })

type FormValues = z.infer<ReturnType<typeof buildSchema>>

export interface UserFormModalProps {
  open: boolean
  onClose: () => void
  user?: User | null
}

export function UserFormModal({ open, onClose, user }: UserFormModalProps) {
  const isCreate = !user
  const { createUser, updateUser } = useUserMutations()
  const toast = useToast()
  const schema = useMemo(() => buildSchema(isCreate), [isCreate])

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', email: '', password: '', status: 'active' },
  })

  useEffect(() => {
    if (open) {
      reset(
        user
          ? { name: user.name, email: user.email, password: '', status: user.status }
          : { name: '', email: '', password: '', status: 'active' },
      )
    }
  }, [open, user, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      if (isCreate) {
        await createUser.mutateAsync({
          name: values.name,
          email: values.email,
          password: values.password ?? '',
        })
      } else {
        await updateUser.mutateAsync({
          user_id: user!.id,
          name: values.name,
          email: values.email,
          status: values.status,
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
      title={isCreate ? 'Create user' : 'Edit user'}
      description={
        isCreate
          ? 'Creates a sign-in account with the default user role.'
          : `Update details for ${user?.name ?? 'this user'}.`
      }
      size="sm"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="user-form" loading={isSubmitting}>
            {isCreate ? 'Create user' : 'Save changes'}
          </Button>
        </>
      }
    >
      <form id="user-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <FormField label="Full name" htmlFor="name" required error={errors.name?.message}>
          <Input
            id="name"
            autoComplete="name"
            placeholder="Ada Lovelace"
            invalid={Boolean(errors.name)}
            {...register('name')}
          />
        </FormField>
        <FormField label="Email address" htmlFor="email" required error={errors.email?.message}>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="ada@company.com"
            invalid={Boolean(errors.email)}
            {...register('email')}
          />
        </FormField>
        {isCreate && (
          <FormField
            label="Password"
            htmlFor="password"
            required
            error={errors.password?.message}
            hint="At least 8 characters."
          >
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              invalid={Boolean(errors.password)}
              {...register('password')}
            />
          </FormField>
        )}
        {!isCreate && (
          <FormField label="Status" htmlFor="status" required error={errors.status?.message}>
            <Select id="status" invalid={Boolean(errors.status)} {...register('status')}>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="locked">Locked</option>
              <option value="pending">Pending</option>
            </Select>
          </FormField>
        )}
      </form>
    </Modal>
  )
}
