import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_API_PROXY_TARGET || 'http://localhost:8080'

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        // The backend exposes every API endpoint via POST /api/v1/...
        '/api': { target: proxyTarget, changeOrigin: true },
        // Authenticated WebSocket endpoint (POST upgrade hijack)
        '/ws': { target: proxyTarget, ws: true },
      },
    },
    build: {
      sourcemap: false,
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        output: {
          // Vite 8 (Rolldown) requires manualChunks as a function.
          manualChunks(id: string) {
            if (id.includes('node_modules')) {
              if (id.includes('recharts') || id.includes('d3-')) return 'charts'
              if (id.includes('framer-motion') || id.includes('motion-dom')) return 'motion'
              if (id.includes('react') || id.includes('scheduler')) return 'react'
            }
          },
        },
      },
    },
  }
})
