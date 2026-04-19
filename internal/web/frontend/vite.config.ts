import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/dashboard/',
  clearScreen: false,
  server: {
    proxy: {
      '/admin': 'http://localhost:8080',
      '/v1': 'http://localhost:8080',
    },
  },
})
