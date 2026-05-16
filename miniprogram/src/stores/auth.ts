/**
 * Pinia Store - 认证状态管理
 */

import { defineStore } from 'pinia'
import { merchantLogin, merchantWechatLogin, getMerchantProfile } from '../api'
import type { MerchantStaff, MerchantInfo } from '../types/index'
import { MERCHANT_SOCKET_URL } from '../config/env'
import { getMerchantWechatCode } from '../utils/merchant_wechat'

let socketTask: any = null
let socketListenersBound = false
let activeStore: any = null
let reconnectTimer: any = null
let orderAudio: any = null
let browseAudio: any = null
let isManualSocketDisconnect = false

export type MerchantSocketStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

export interface MerchantSocketMessageSummary {
  type: string
  merchantId?: number
  orderNo?: string
  visitorOpenId?: string
  source?: string
}

function getStoredBoolean(key: string, defaultValue = true): boolean {
  const value = uni.getStorageSync(key)
  if (value === '') {
    return defaultValue
  }
  return !!value
}

function getCurrentDateTime(): string {
  const now = new Date()
  const datePart = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
  const timePart = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  return `${datePart} ${timePart}`
}

function buildSocketMessageSummary(message: any): MerchantSocketMessageSummary {
  return {
    type: message?.type || 'unknown',
    merchantId: Number(message?.payload?.merchant_id || 0) || undefined,
    orderNo: message?.payload?.order_no || undefined,
    visitorOpenId: message?.payload?.visitor_openid || undefined,
    source: message?.payload?.source || undefined
  }
}

function createAudio(src: string) {
  const audio = uni.createInnerAudioContext()
  audio.src = src
  audio.obeyMuteSwitch = false
  return audio
}

function playAudio(audio: any) {
  try {
    audio.stop()
  } catch (error) {
  }
  audio.play()
}

interface AuthState {
  token: string
  merchantId: number | null
  merchantInfo: MerchantInfo | null
  staff: MerchantStaff | null
  isLoggedIn: boolean
  orderSoundEnabled: boolean
  browseSoundEnabled: boolean
  socketStatus: MerchantSocketStatus
  lastSocketMessage: MerchantSocketMessageSummary | null
  lastSocketMessageAt: string
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: uni.getStorageSync('token') || '',
    merchantId: uni.getStorageSync('merchantId') || null,
    merchantInfo: null,
    staff: null,
    isLoggedIn: !!uni.getStorageSync('token'),
    orderSoundEnabled: getStoredBoolean('merchant_sound_enabled'),
    browseSoundEnabled: getStoredBoolean('merchant_browse_sound_enabled'),
    socketStatus: 'disconnected',
    lastSocketMessage: null,
    lastSocketMessageAt: ''
  }),

  getters: {
    isAuthenticated: (state) => state.isLoggedIn && !!state.token,
    merchantName: (state) => state.merchantInfo?.name || '',
    merchantStatus: (state) => state.merchantInfo?.status ?? 0,
    soundEnabled: (state) => state.orderSoundEnabled
  },

  actions: {
    clearAuthState(shouldRedirect = false) {
      this.isLoggedIn = false
      this.disconnectOrderSocket(true)

      this.token = ''
      this.merchantId = null
      this.merchantInfo = null
      this.staff = null
      this.lastSocketMessage = null
      this.lastSocketMessageAt = ''

      uni.removeStorageSync('token')
      uni.removeStorageSync('merchantId')
      uni.removeStorageSync('staff')
      uni.removeStorageSync('merchantInfo')

      if (shouldRedirect) {
        uni.reLaunch({ url: '/pages/auth/login' })
      }
    },

    persistAuthState() {
      uni.setStorageSync('token', this.token)
      uni.setStorageSync('merchantId', this.merchantId)
      uni.setStorageSync('staff', JSON.stringify(this.staff))
    },

    hydrateSoundSettingsFromStaffSettings() {
      if (!this.staff) {
        return
      }
      if (typeof this.staff.notify_enabled === 'boolean') {
        this.setOrderSoundEnabled(this.staff.notify_enabled)
      }
      if (typeof this.staff.browse_notify_enabled === 'boolean') {
        this.setBrowseSoundEnabled(this.staff.browse_notify_enabled)
      }
    },

    // 商家登录
    async login(username: string, password: string) {
      try {
        const res = await merchantLogin({ username, password })

        this.token = res.token
        this.merchantId = res.merchant_id
        this.staff = res.staff
        this.isLoggedIn = true
        this.persistAuthState()
        this.hydrateSoundSettingsFromStaffSettings()
        await this.fetchMerchantInfo()
        this.connectOrderSocket()
        return true
      } catch (error: any) {
        uni.showToast({ title: error.message || '登录失败', icon: 'none' })
        return false
      }
    },

    async loginWithWechat() {
      try {
        const code = await getMerchantWechatCode()
        const res = await merchantWechatLogin({ code })
        this.token = res.token
        this.merchantId = res.merchant_id
        this.staff = res.staff
        this.isLoggedIn = true
        this.persistAuthState()
        this.hydrateSoundSettingsFromStaffSettings()
        await this.fetchMerchantInfo()
        this.connectOrderSocket()
        return true
      } catch (error: any) {
        uni.showToast({ title: error.message || '快捷登录失败', icon: 'none' })
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
      this.clearAuthState(true)
    },

    handleSessionExpired() {
      this.clearAuthState(false)
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

      this.orderSoundEnabled = getStoredBoolean('merchant_sound_enabled')
      this.browseSoundEnabled = getStoredBoolean('merchant_browse_sound_enabled')
      
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

    setOrderSoundEnabled(enabled: boolean) {
      this.orderSoundEnabled = enabled
      uni.setStorageSync('merchant_sound_enabled', enabled)
    },

    setBrowseSoundEnabled(enabled: boolean) {
      this.browseSoundEnabled = enabled
      uni.setStorageSync('merchant_browse_sound_enabled', enabled)
    },

    setSocketStatus(status: MerchantSocketStatus) {
      this.socketStatus = status
    },

    setLastSocketMessage(message: MerchantSocketMessageSummary) {
      this.lastSocketMessage = message
      this.lastSocketMessageAt = getCurrentDateTime()
    },

    connectOrderSocket() {
      if (socketTask) return
      const token = this.token || uni.getStorageSync('token')
      if (!token) {
        this.setSocketStatus('disconnected')
        return
      }

      isManualSocketDisconnect = false
      activeStore = this
      this.setSocketStatus('connecting')
      socketTask = uni.connectSocket({
        url: MERCHANT_SOCKET_URL,
        header: {
          Authorization: `Bearer ${token}`
        }
      })

      if (!socketListenersBound) {
        socketListenersBound = true

        uni.onSocketOpen(() => {
          activeStore?.setSocketStatus('connected')
          if (reconnectTimer) {
            clearTimeout(reconnectTimer)
            reconnectTimer = null
          }
        })

        uni.onSocketMessage((res) => {
          try {
            const dataStr = typeof res.data === 'string' ? res.data : ''
            const msg = dataStr ? JSON.parse(dataStr) : null
            if (!msg) return

            activeStore?.setLastSocketMessage(buildSocketMessageSummary(msg))

            if (msg.type === 'order_notify') {
              const orderNo = msg.payload?.order_no || ''
              if (activeStore?.orderSoundEnabled) {
                if (!orderAudio) {
                  orderAudio = createAudio('/static/sounds/order.mp3')
                }
                playAudio(orderAudio)
              }
              uni.showToast({ title: orderNo ? `新订单 ${orderNo}` : '新订单提醒', icon: 'none' })
              return
            }

            if (msg.type === 'store_visit_notify') {
              if (activeStore?.browseSoundEnabled) {
                if (!browseAudio) {
                  browseAudio = createAudio('/static/sounds/browse.mp3')
                }
                playAudio(browseAudio)
              }
              //uni.showToast({ title: '有顾客正在浏览店铺', icon: 'none' })
            }
          } catch (e) {
          }
        })

        uni.onSocketError(() => {
          activeStore?.setSocketStatus('disconnected')
          socketTask = null
          if (isManualSocketDisconnect) {
            activeStore = null
          }
        })

        uni.onSocketClose(() => {
          socketTask = null
          if (isManualSocketDisconnect) {
            activeStore?.setSocketStatus('disconnected')
            activeStore = null
            return
          }

          if (activeStore?.isLoggedIn) {
            activeStore?.setSocketStatus('reconnecting')
            if (reconnectTimer) {
              clearTimeout(reconnectTimer)
            }
            reconnectTimer = setTimeout(() => {
              activeStore?.connectOrderSocket()
            }, 2000)
          } else {
            activeStore?.setSocketStatus('disconnected')
          }
        })
      }
    },

    disconnectOrderSocket(isManual = false) {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
      isManualSocketDisconnect = isManual
      if (socketTask) {
        try {
          socketTask.close()
        } catch (e) {
        }
      }
      this.setSocketStatus('disconnected')
      socketTask = null
      if (!isManual) {
        activeStore = null
      }
    },

    // 更新商家信息
    updateMerchantInfo(info: MerchantInfo) {
      this.merchantInfo = info
      uni.setStorageSync('merchantInfo', JSON.stringify(info))
    },

    updateStaffInfo(staff: MerchantStaff) {
      this.staff = staff
      uni.setStorageSync('staff', JSON.stringify(staff))
      this.hydrateSoundSettingsFromStaffSettings()
    }
  }
})
