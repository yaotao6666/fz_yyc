import { defineConfig, loadEnv } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import path from 'path'

function trimTrailingSlash(url: string): string {
  return url.replace(/\/+$/, '')
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, '')
  const apiBaseUrl = trimTrailingSlash(env.VITE_API_BASE_URL || 'http://localhost:8080')
  const proxyTarget = trimTrailingSlash(env.VITE_DEV_PROXY_TARGET || apiBaseUrl)

  return {
    plugins: [
      uni()
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
        '@api': path.resolve(__dirname, 'src/api'),
        '@components': path.resolve(__dirname, 'src/components'),
        '@stores': path.resolve(__dirname, 'src/stores'),
        '@utils': path.resolve(__dirname, 'src/utils'),
        '@types': path.resolve(__dirname, 'src/types')
      }
    },
    server: {
      port: 3000,
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
          rewrite: (requestPath) => requestPath.replace(/^\/api/, '')
        }
      }
    },
    uni: {
      vueComponents: {
        // 自动导入组件
        // 不需要在这里配置，uni-app 会自动扫描
      }
    }
  }
})
