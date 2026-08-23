import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, '')
  const proxyTarget = env.VITE_API_PROXY_TARGET
  const appBase = '/sm/'

  return {
    plugins: [vue()],
    base: appBase,
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src')
      }
    },
    build: {
      outDir: 'sm'
    },
    server: proxyTarget
      ? {
          host: '0.0.0.0',
          proxy: {
            '/api': {
              target: proxyTarget,
              changeOrigin: true
            }
          }
        }
      : undefined
  }
})
