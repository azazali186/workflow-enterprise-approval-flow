import { useMutation } from '@tanstack/react-query'
import { useAppDispatch } from '@/store/hooks'
import { setCredentials } from '@/store/slices/auth.slice'
import { useToast } from '@/hooks/use-toast'
import { authService } from '@/services/auth.service'

export interface LoginInput {
  email: string
  password: string
}

export function useLogin() {
  const dispatch = useAppDispatch()
  const toast = useToast()

  return useMutation({
    mutationFn: ({ email, password }: LoginInput) => authService.login(email, password),
    onSuccess: (result) => {
      dispatch(
        setCredentials({
          user: result.user,
          accessToken: result.access_token,
          refreshToken: result.refresh_token,
          expiresAt: result.expires_at,
        }),
      )
      const firstName = result.user.name.split(' ')[0] || 'there'
      toast.success(`Welcome back, ${firstName}`)
    },
  })
}
