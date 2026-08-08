import { configureStore } from '@reduxjs/toolkit'
import authReducer from './slices/auth.slice'
import toastReducer from './slices/toast.slice'
import uiReducer from './slices/ui.slice'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    toast: toastReducer,
    ui: uiReducer,
  },
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
