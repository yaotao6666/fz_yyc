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
    server: {
      host: '0.0.0.0',
      // 内网穿透（cpolar 等）域名访问放行，避免 Dev Server 校验 Host 被拦截
      allowedHosts: ['250c63c0.r15.cpolar.top', '.cpolar.top'],
      ...(proxyTarget
        ? {
            proxy: {
              '/api': {
                target: proxyTarget,
                changeOrigin: true
              }
            }
          }
        : {})
    }
  }
})
