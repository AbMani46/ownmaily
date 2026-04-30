import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:4400',
      '/track': 'http://localhost:4400',
      '/unsubscribe': 'http://localhost:4400',
      '/confirm': 'http://localhost:4400',
      '/webhooks': 'http://localhost:4400',
    },
  },
})
