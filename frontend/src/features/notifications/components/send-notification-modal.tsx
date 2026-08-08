import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Modal } from '@/components/ui/modal'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Combobox } from '@/components/ui/combobox'
import { useUserDropdown } from '@/hooks/use-dropdowns'
import { useNotificationMutations } from '@/features/notifications/hooks/use-notifications'

const schema = z.object({
  user_id: z.string().min(1, 'Recipient is required'),
  type: z.string().min(1, 'Type is required'),
  channel: z.string().min(1, 'Channel is required'),
  title: z.string().min(1, 'Title is required'),
  body: z.string().min(1, 'Message is required'),
})

type FormValues = z.infer<typeof schema>

export interface SendNotificationModalProps {
  open: boolean
  onClose: () => void
}

export function SendNotificationModal({ open, onClose }: SendNotificationModalProps) {
  const { sendNotification } = useNotificationMutations()
  const { data: userOptions, isLoading: isLoadingUsers } = useUserDropdown({ enabled: open })

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { user_id: '', type: 'info', channel: 'in_app', title: '', body: '' },
  })

  useEffect(() => {
    if (open) {
      reset({ user_id: '', type: 'info', channel: 'in_app', title: '', body: '' })
    }
  }, [open, reset])

  const onSubmit = async (values: FormValues) => {
    try {
      await sendNotification.mutateAsync(values)
      onClose()
    } catch {
      // Errors are surfaced by the mutation toast.
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Send notification"
      description="Deliver a message to a user through the selected channel."
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="send-notification-form" loading={isSubmitting}>
            Send notification
          </Button>
        </>
      }
    >
      <form id="send-notification-form" onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <FormField label="Recipient" htmlFor="user_id" required error={errors.user_id?.message}>
          <Controller
            name="user_id"
            control={control}
            render={({ field }) => (
              <Combobox
                id="user_id"
                options={userOptions || []}
                value={field.value}
                onChange={field.onChange}
                placeholder="Select recipient..."
                searchPlaceholder="Search users..."
                disabled={isLoadingUsers}
                invalid={Boolean(errors.user_id)}
              />
            )}
          />
        </FormField>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Type" htmlFor="type" required error={errors.type?.message}>
            <Select id="type" invalid={Boolean(errors.type)} {...register('type')}>
              <option value="info">Info</option>
              <option value="approval">Approval</option>
              <option value="escalation">Escalation</option>
              <option value="system">System</option>
              <option value="alert">Alert</option>
            </Select>
          </FormField>
          <FormField label="Channel" htmlFor="channel" required error={errors.channel?.message}>
            <Select id="channel" invalid={Boolean(errors.channel)} {...register('channel')}>
              <option value="in_app">In-app</option>
              <option value="email">Email</option>
              <option value="sms">SMS</option>
              <option value="push">Push</option>
            </Select>
          </FormField>
        </div>
        <FormField label="Title" htmlFor="title" required error={errors.title?.message}>
          <Input id="title" placeholder="e.g. Approval pending" invalid={Boolean(errors.title)} {...register('title')} />
        </FormField>
        <FormField label="Message" htmlFor="body" required error={errors.body?.message}>
          <Textarea id="body" rows={3} invalid={Boolean(errors.body)} {...register('body')} />
        </FormField>
      </form>
    </Modal>
  )
}
