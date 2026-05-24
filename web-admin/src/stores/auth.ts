import { defineStore } from 'pinia'
import type { ServiceProviderInfo } from '@/types/sp'
import { spLogin, spLogout } from '@/api/sp'

interface AuthState {
  token: string
  serviceProvider: ServiceProviderInfo | null
}

const STORAGE_TOKEN_KEY = 'sp_token'
const STORAGE_INFO_KEY = 'sp_info'

export const useAuthStore = defineStore('sp-auth', {
  state: (): AuthState => ({
    token: window.localStorage.getItem(STORAGE_TOKEN_KEY) || '',
    serviceProvider: parseServiceProvider(window.localStorage.getItem(STORAGE_INFO_KEY)),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token),
    serviceProviderName: (state) => state.serviceProvider?.name || '服务商',
    operatorName: (state) => state.serviceProvider?.sp_name || '服务商账号'
  },
  actions: {
    async login(username: string, password: string) {
      const result = await spLogin({ username, password })
      this.token = result.token
      this.serviceProvider = result.service_provider
      window.localStorage.setItem(STORAGE_TOKEN_KEY, result.token)
      window.localStorage.setItem(STORAGE_INFO_KEY, JSON.stringify(result.service_provider))
      return result
    },
    async logout() {
      try {
        await spLogout()
      } catch (_error) {
        // 忽略退出接口异常，优先清理本地登录态。
      }
      this.clearSession()
    },
    clearSession() {
      this.token = ''
      this.serviceProvider = null
      window.localStorage.removeItem(STORAGE_TOKEN_KEY)
      window.localStorage.removeItem(STORAGE_INFO_KEY)
    }
  }
})

function parseServiceProvider(raw: string | null): ServiceProviderInfo | null {
  if (!raw) return null
  try {
    return JSON.parse(raw) as ServiceProviderInfo
  } catch (_error) {
    return null
  }
}
