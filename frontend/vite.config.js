import {defineConfig, loadEnv} from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react()],
    server: {
      host: true,
      port: 5173,
      proxy: {
        '/api/v1': {
          target: env.VITE_API_TARGET || 'http://gateway:8080',
          changeOrigin: true,
          // Если ваш gateway ждет /api/v1/users, а фронт шлет /api/users
          // rewrite: (path) => path.replace(/^\/api/, ''),
        }
      }
    }
  }
})
