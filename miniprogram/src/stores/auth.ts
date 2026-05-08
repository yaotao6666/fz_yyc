/**
 * Pinia Store - 认证状态管理
 */

import { defineStore } from 'pinia'
import { merchantLogin, getMerchantProfile } from '../api'
import type { MerchantStaff, MerchantInfo } from '../types/api'

interface AuthState {
  token: string
  merchantId: number | null
  merchantInfo: MerchantInfo | null
  staff: MerchantStaff | null
  isLoggedIn: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: uni.getStorageSync('token') || '',
    merchantId: uni.getStorageSync('merchantId') || null,
    merchantInfo: null,
    staff: null,
    isLoggedIn: !!uni.getStorageSync('token')
  }),

  getters: {
    isAuthenticated: (state) => state.isLoggedIn && !!state.token,
    merchantName: (state) => state.merchantInfo?.name || '',
    merchantStatus: (state) => state.merchantInfo?.status ?? 0
  },

  actions: {
    // 商家登录
    async login(username: string, password: string) {
      try {
        const res = await merchantLogin({ username, password })
        
        this.token = res.token
        this.merchantId = res.merchant_id
        this.staff = res.staff
        
        // 保存到本地存储
        uni.setStorageSync('token', res.token)
        uni.setStorageSync('merchantId', res.merchant_id)
        uni.setStorageSync('staff', JSON.stringify(res.staff))
        
        this.isLoggedIn = true
        
        // 获取商家信息
        await this.fetchMerchantInfo()
        
        return true
      } catch (error: any) {
        uni.showToast({ title: error.message || '登录失败', icon: 'none' })
        return false
      }
    },

    // 获取商家信息
    async fetchMerchantInfo() {
      try {
        const info = await getMerchantProfile()
        this.merchantInfo = info
        uni.setStorageSync('merchantInfo', JSON.stringify(info))
        return info
      } catch (error: any) {
        console.error('获取商家信息失败:', error)
        return null
      }
    },

    // 登出
    logout() {
      this.token = ''
      this.merchantId = null
      this.merchantInfo = null
      this.staff = null
      this.isLoggedIn = false
      
      uni.removeStorageSync('token')
      uni.removeStorageSync('merchantId')
      uni.removeStorageSync('staff')
      uni.removeStorageSync('merchantInfo')
      
      uni.reLaunch({ url: '/pages/auth/login' })
    },

    // 检查登录状态
    checkLogin() {
      const token = uni.getStorageSync('token')
      if (!token) {
        return false
      }
      
      this.token = token
      this.merchantId = uni.getStorageSync('merchantId')
      this.isLoggedIn = true
      
      const staffStr = uni.getStorageSync('staff')
      if (staffStr) {
        try {
          this.staff = JSON.parse(staffStr)
        } catch (e) {
          console.error('解析员工信息失败')
        }
      }
      
      const merchantStr = uni.getStorageSync('merchantInfo')
      if (merchantStr) {
        try {
          this.merchantInfo = JSON.parse(merchantStr)
        } catch (e) {
          console.error('解析商家信息失败')
        }
      }
      
      return true
    },

    // 更新商家信息
    updateMerchantInfo(info: MerchantInfo) {
      this.merchantInfo = info
      uni.setStorageSync('merchantInfo', JSON.stringify(info))
    }
  }
})
