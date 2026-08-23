import { defineStore } from 'pinia'
import type { MerchantStaffInfo, SysMenu } from '@/types/sp'
import { merchantLogin, merchantLogout, getRbacPermissions } from '@/api/sp'

interface AuthState {
  token: string
  staff: MerchantStaffInfo | null
  menus: SysMenu[]
  permissions: string[]
}

const STORAGE_TOKEN_KEY = 'merchant_token'
const STORAGE_INFO_KEY = 'merchant_info'
const STORAGE_MENUS_KEY = 'merchant_menus'
const STORAGE_PERMS_KEY = 'merchant_permissions'

function readArray(key: string): SysMenu[] {
  try {
    const raw = window.localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as SysMenu[]) : []
  } catch (_error) {
    return []
  }
}

function readStringArray(key: string): string[] {
  try {
    const raw = window.localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as string[]) : []
  } catch (_error) {
    return []
  }
}

export const useAuthStore = defineStore('merchant-auth', {
  state: (): AuthState => ({
    token: window.localStorage.getItem(STORAGE_TOKEN_KEY) || '',
    staff: parseStaff(window.localStorage.getItem(STORAGE_INFO_KEY)),
    menus: readArray(STORAGE_MENUS_KEY),
    permissions: readStringArray(STORAGE_PERMS_KEY),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token),
    staffName: (state) => state.staff?.name || state.staff?.username || '商家',
    isOwner: (state) => state.staff?.role === 'owner',
    hasPermission: (state) => (code: string) => {
      if (!code) return true
      if (state.staff?.role === 'owner') return true
      return state.permissions.includes(code)
    },
  },
  actions: {
    async login(username: string, password: string) {
      const result = await merchantLogin({ username, password })
      this.applySession(result.token, result.staff, result.menus || [], result.permissions || [])
      return result
    },
    // 登录态持久化辅助（登录接口与微信登录响应一致时均可调用）
    applySession(token: string, staff: MerchantStaffInfo, menus: SysMenu[], permissions: string[]) {
      this.token = token
      this.staff = staff
      this.menus = menus
      this.permissions = permissions
      window.localStorage.setItem(STORAGE_TOKEN_KEY, token)
      window.localStorage.setItem(STORAGE_INFO_KEY, JSON.stringify(staff))
      window.localStorage.setItem(STORAGE_MENUS_KEY, JSON.stringify(menus))
      window.localStorage.setItem(STORAGE_PERMS_KEY, JSON.stringify(permissions))
    },
    // 登录后刷新权限（菜单/角色变更后调用）
    async fetchPermissions() {
      const result = await getRbacPermissions()
      this.menus = result.menus || []
      this.permissions = result.permissions || []
      window.localStorage.setItem(STORAGE_MENUS_KEY, JSON.stringify(this.menus))
      window.localStorage.setItem(STORAGE_PERMS_KEY, JSON.stringify(this.permissions))
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
      this.menus = []
      this.permissions = []
      window.localStorage.removeItem(STORAGE_TOKEN_KEY)
      window.localStorage.removeItem(STORAGE_INFO_KEY)
      window.localStorage.removeItem(STORAGE_MENUS_KEY)
      window.localStorage.removeItem(STORAGE_PERMS_KEY)
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
