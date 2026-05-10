import { defineStore } from 'pinia'
import { spLogin } from '../api'

interface SpState {
  token: string
  spId: number | null
  spInfo: any
  isLoggedIn: boolean
}

export const useSpStore = defineStore('sp', {
  state: (): SpState => ({
    token: uni.getStorageSync('sp_token') || '',
    spId: uni.getStorageSync('sp_id') || null,
    spInfo: null,
    isLoggedIn: !!uni.getStorageSync('sp_token')
  }),

  getters: {
    isAuthenticated: (state) => state.isLoggedIn && !!state.token,
    spName: (state) => state.spInfo?.name || ''
  },

  actions: {
    async login(username: string, password: string) {
      try {
        const res = await spLogin({ username, password })

        this.token = res.token
        this.spId = res.service_provider?.id
        this.spInfo = res.service_provider

        uni.setStorageSync('sp_token', res.token)
        uni.setStorageSync('sp_id', res.service_provider?.id)
        uni.setStorageSync('sp_info', JSON.stringify(res.service_provider))

        this.isLoggedIn = true

        return true
      } catch (error: any) {
        uni.showToast({ title: error.message || '登录失败', icon: 'none' })
        return false
      }
    },

    logout() {
      this.token = ''
      this.spId = null
      this.spInfo = null
      this.isLoggedIn = false

      uni.removeStorageSync('sp_token')
      uni.removeStorageSync('sp_id')
      uni.removeStorageSync('sp_info')

      uni.reLaunch({ url: '/pages/sp/login' })
    },

    checkLogin() {
      const token = uni.getStorageSync('sp_token')
      if (!token) {
        return false
      }

      this.token = token
      this.spId = uni.getStorageSync('sp_id')
      this.isLoggedIn = true

      const spInfoStr = uni.getStorageSync('sp_info')
      if (spInfoStr) {
        try {
          this.spInfo = JSON.parse(spInfoStr)
        } catch (e) {
          console.error('解析服务商信息失败')
        }
      }

      return true
    }
  }
})
