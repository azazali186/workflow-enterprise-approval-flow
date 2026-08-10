import React from 'react'
import ReactDOM from 'react-dom/client'
import { Provider } from 'react-redux'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { store } from '@/store'
import { router } from '@/routes'
import { ErrorBoundary } from '@/components/ui/error-boundary'
import { AppErrorFallback } from '@/components/ui/app-error-fallback'
import { ToastViewport } from '@/components/ui/toast'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
    mutations: {
      retry: 0,
    },
  },
})

// Root safety net for provider/render-level errors outside the router
// (route-element crashes are handled by the router's own errorElement — see
// routes/index.tsx — because React Router never propagates those upward).
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <ErrorBoundary fallback={<AppErrorFallback />}>
          <RouterProvider router={router} />
        </ErrorBoundary>
        <ToastViewport />
      </QueryClientProvider>
    </Provider>
  </React.StrictMode>,
)
