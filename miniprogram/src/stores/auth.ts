/**
 * Pinia Store - 认证状态管理
 */

import { defineStore } from 'pinia'
import { merchantLogin, getMerchantProfile } from '../api'
import type { MerchantStaff, MerchantInfo } from '../types/index'

let socketTask: any = null
let socketListenersBound = false
let activeStore: any = null
let reconnectTimer: any = null
let orderAudio: any = null

const WS_URL = process.env.NODE_ENV === 'development'
  ? 'ws://localhost:8080/api/v1/ws/merchant'
  : 'wss://api.example.com/api/v1/ws/merchant'

interface AuthState {
  token: string
  merchantId: number | null
  merchantInfo: MerchantInfo | null
  staff: MerchantStaff | null
  isLoggedIn: boolean
  soundEnabled: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: uni.getStorageSync('token') || '',
    merchantId: uni.getStorageSync('merchantId') || null,
    merchantInfo: null,
    staff: null,
    isLoggedIn: !!uni.getStorageSync('token'),
    soundEnabled: uni.getStorageSync('merchant_sound_enabled') !== '' ? !!uni.getStorageSync('merchant_sound_enabled') : true
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

        this.connectOrderSocket()
        
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
      this.disconnectOrderSocket()

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

      const soundEnabled = uni.getStorageSync('merchant_sound_enabled')
      if (soundEnabled !== '') {
        this.soundEnabled = !!soundEnabled
      }
      
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

      this.connectOrderSocket()
      
      return true
    },

    setSoundEnabled(enabled: boolean) {
      this.soundEnabled = enabled
      uni.setStorageSync('merchant_sound_enabled', enabled)
    },

    connectOrderSocket() {
      if (socketTask) return
      const token = this.token || uni.getStorageSync('token')
      if (!token) return

      activeStore = this
      socketTask = uni.connectSocket({
        url: WS_URL,
        header: {
          Authorization: `Bearer ${token}`
        }
      })

      if (!socketListenersBound) {
        socketListenersBound = true

        uni.onSocketOpen(() => {
          if (reconnectTimer) {
            clearTimeout(reconnectTimer)
            reconnectTimer = null
          }
        })

        uni.onSocketMessage((res) => {
          try {
            const dataStr = typeof res.data === 'string' ? res.data : ''
            const msg = dataStr ? JSON.parse(dataStr) : null
            if (!msg || msg.type !== 'order_notify') return

            const orderNo = msg.payload?.order_no || ''

            if (activeStore?.soundEnabled) {
              if (!orderAudio) {
                orderAudio = uni.createInnerAudioContext()
                orderAudio.src = '/static/sounds/order.wav'
                orderAudio.obeyMuteSwitch = false
              }
              try {
                orderAudio.stop()
              } catch (e) {
              }
              orderAudio.play()
            }

            uni.showToast({ title: orderNo ? `新订单 ${orderNo}` : '新订单提醒', icon: 'none' })
          } catch (e) {
          }
        })

        uni.onSocketError(() => {
          socketTask = null
        })

        uni.onSocketClose(() => {
          socketTask = null
          if (activeStore?.isLoggedIn) {
            if (reconnectTimer) {
              clearTimeout(reconnectTimer)
            }
            reconnectTimer = setTimeout(() => {
              activeStore?.connectOrderSocket()
            }, 2000)
          }
        })
      }
    },

    disconnectOrderSocket() {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
      if (socketTask) {
        try {
          socketTask.close()
        } catch (e) {
        }
      }
      socketTask = null
      activeStore = null
    },

    // 更新商家信息
    updateMerchantInfo(info: MerchantInfo) {
      this.merchantInfo = info
      uni.setStorageSync('merchantInfo', JSON.stringify(info))
    }
  }
})
