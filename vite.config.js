import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: 'pixelcraft-studio.store',
    allowedHosts: ['pixelcraft-studio.store']
  },
  esbuild: {
    drop: ['console', 'debugger']
  }
})
