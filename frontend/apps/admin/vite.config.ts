import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
export default defineConfig({
  plugins: [vue()],
  resolve: { alias: { '@': resolve(__dirname, 'src') } },
  server: {
    port: 5174,
    proxy: {
      '/api': { target: 'http://localhost:8082', changeOrigin: true },
      '/ws':  { target: 'ws://localhost:8082', ws: true },
    },
  },
  define: {
    'import.meta.env.VITE_STORE_NAME': JSON.stringify(process.env.VITE_STORE_NAME ?? 'Mi Tienda'),
  },
})
