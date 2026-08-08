import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { useToast } from '@/hooks/use-toast'
import { authService } from '@/services/auth.service'
import { toErrorMessage } from '@/services/api/errors'

const schema = z
  .object({
    old_password: z.string().min(1, 'Current password is required'),
    new_password: z.string().min(8, 'New password must be at least 8 characters'),
    confirm: z.string().min(1, 'Please confirm the new password'),
  })
  .refine((values) => values.new_password === values.confirm, {
    path: ['confirm'],
    message: 'Passwords do not match',
  })

type FormValues = z.infer<typeof schema>

export interface ChangePasswordModalProps {
  open: boolean
  onClose: () => void
}

export function ChangePasswordModal({ open, onClose }: ChangePasswordModalProps) {
  const toast = useToast()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { old_password: '', new_password: '', confirm: '' },
  })

  const onSubmit = async (values: FormValues) => {
    try {
      await authService.changePassword(values.old_password, values.new_password)
      toast.success('Password updated', 'Your password has been changed successfully.')
      reset()
      onClose()
    } catch (error) {
      toast.error('Could not update password', toErrorMessage(error))
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Change password"
      description="Use a strong password that you don't use on other sites."
      size="sm"
      footer={
        <Button type="submit" form="change-password-form" loading={isSubmitting}>
          Update password
        </Button>
      }
    >
      <form id="change-password-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <FormField label="Current password" htmlFor="old_password" required error={errors.old_password?.message}>
          <Input
            id="old_password"
            type="password"
            autoComplete="current-password"
            invalid={Boolean(errors.old_password)}
            {...register('old_password')}
          />
        </FormField>
        <FormField label="New password" htmlFor="new_password" required error={errors.new_password?.message} hint="At least 8 characters.">
          <Input
            id="new_password"
            type="password"
            autoComplete="new-password"
            invalid={Boolean(errors.new_password)}
            {...register('new_password')}
          />
        </FormField>
        <FormField label="Confirm new password" htmlFor="confirm" required error={errors.confirm?.message}>
          <Input
            id="confirm"
            type="password"
            autoComplete="new-password"
            invalid={Boolean(errors.confirm)}
            {...register('confirm')}
          />
        </FormField>
      </form>
    </Modal>
  )
}
