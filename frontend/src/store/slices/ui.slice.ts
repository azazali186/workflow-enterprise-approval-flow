import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

export interface UIState {
  sidebarCollapsed: boolean
  mobileSidebarOpen: boolean
}

const initialState: UIState = {
  sidebarCollapsed: false,
  mobileSidebarOpen: false,
}

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed
    },
    setSidebarCollapsed(state, action: PayloadAction<boolean>) {
      state.sidebarCollapsed = action.payload
    },
    setMobileSidebarOpen(state, action: PayloadAction<boolean>) {
      state.mobileSidebarOpen = action.payload
    },
  },
})

export const { toggleSidebar, setSidebarCollapsed, setMobileSidebarOpen } = uiSlice.actions
export default uiSlice.reducer
