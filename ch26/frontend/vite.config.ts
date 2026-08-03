import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// 开发/预览时把 /api 代理到 Go 后端。
// 默认 localhost:8080;docker compose 里通过 VITE_API_PROXY=http://backend:8080 覆盖。
const apiProxy = process.env.VITE_API_PROXY || 'http://localhost:8080'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': apiProxy,
    },
  },
  preview: {
    proxy: {
      '/api': apiProxy,
    },
  },
})
