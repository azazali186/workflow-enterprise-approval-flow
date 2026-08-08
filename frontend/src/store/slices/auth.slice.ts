import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type { User } from '@/types/models'
import { storage } from '@/utils/storage'

export interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: string | null
  isAuthenticated: boolean
}

const initialState: AuthState = {
  user: null,
  accessToken: storage.getAccessToken(),
  refreshToken: storage.getRefreshToken(),
  expiresAt: null,
  isAuthenticated: Boolean(storage.getAccessToken()),
}

interface CredentialsPayload {
  user: User
  accessToken: string
  refreshToken: string
  expiresAt: string
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials(state, action: PayloadAction<CredentialsPayload>) {
      state.user = action.payload.user
      state.accessToken = action.payload.accessToken
      state.refreshToken = action.payload.refreshToken
      state.expiresAt = action.payload.expiresAt
      state.isAuthenticated = true
      storage.setAccessToken(action.payload.accessToken)
      storage.setRefreshToken(action.payload.refreshToken)
    },
    setUser(state, action: PayloadAction<User>) {
      state.user = action.payload
    },
    clearCredentials(state) {
      state.user = null
      state.accessToken = null
      state.refreshToken = null
      state.expiresAt = null
      state.isAuthenticated = false
      storage.clearAuth()
    },
  },
})

export const { setCredentials, setUser, clearCredentials } = authSlice.actions
export default authSlice.reducer
