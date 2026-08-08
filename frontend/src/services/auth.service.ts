import { post } from './api/client'
import type { AuthResult, User } from '@/types/models'

export interface RegisterPayload {
  name: string
  email: string
  password: string
}

export const authService = {
  login(email: string, password: string): Promise<AuthResult> {
    return post<AuthResult>('/auth/login', { email, password })
  },

  register(payload: RegisterPayload): Promise<User> {
    return post<User>('/auth/register', payload)
  },

  refresh(refreshToken: string): Promise<AuthResult> {
    return post<AuthResult>('/auth/refresh', { refresh_token: refreshToken }, { skipAuth: true })
  },

  profile(userId: string): Promise<User> {
    return post<User>('/profile', { user_id: userId })
  },

  logout(userId: string): Promise<{ message: string }> {
    return post<{ message: string }>('/logout', { user_id: userId })
  },

  changePassword(oldPassword: string, newPassword: string): Promise<{ message: string }> {
    return post<{ message: string }>('/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
