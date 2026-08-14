import { defineStore } from 'pinia'
import type { MerchantStaffInfo } from '@/types/sp'
import { merchantLogin, merchantLogout } from '@/api/sp'

interface AuthState {
  token: string
  staff: MerchantStaffInfo | null
}

const STORAGE_TOKEN_KEY = 'merchant_token'
const STORAGE_INFO_KEY = 'merchant_info'

export const useAuthStore = defineStore('merchant-auth', {
  state: (): AuthState => ({
    token: window.localStorage.getItem(STORAGE_TOKEN_KEY) || '',
    staff: parseStaff(window.localStorage.getItem(STORAGE_INFO_KEY)),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token),
    staffName: (state) => state.staff?.name || state.staff?.username || '商家'
  },
  actions: {
    async login(username: string, password: string) {
      const result = await merchantLogin({ username, password })
      this.token = result.token
      this.staff = result.staff
      window.localStorage.setItem(STORAGE_TOKEN_KEY, result.token)
      window.localStorage.setItem(STORAGE_INFO_KEY, JSON.stringify(result.staff))
      return result
    },
    async logout() {
      try {
        await merchantLogout()
      } catch (_error) {
        // 忽略退出接口异常，优先清理本地登录态。
      }
      this.clearSession()
    },
    clearSession() {
      this.token = ''
      this.staff = null
      window.localStorage.removeItem(STORAGE_TOKEN_KEY)
      window.localStorage.removeItem(STORAGE_INFO_KEY)
    }
  }
})

function parseStaff(raw: string | null): MerchantStaffInfo | null {
  if (!raw) return null
  try {
    return JSON.parse(raw) as MerchantStaffInfo
  } catch (_error) {
    return null
  }
}
