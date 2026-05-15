/**
 * 微信授权登录工具
 * 提供自动登录和token管理功能
 * openid作为用户唯一标识
 */

import { ref } from 'vue'
import { API_BASE_URL } from '../config/env'
import type { ApiResponse, WechatLoginResponse } from '../types'

const userToken = ref(uni.getStorageSync('user_token') || '')
const userInfo = ref(uni.getStorageSync('userInfo') || null)
const openid = ref(uni.getStorageSync('openid') || '')
const isLoggedIn = ref(!!userToken.value && !!openid.value)

let loginPromise: Promise<LoginResult> | null = null

interface LoginResult {
  success: boolean
  token?: string
  user?: any
  error?: string
}

export function useAuth() {
  const syncStateFromStorage = () => {
    const storedToken = uni.getStorageSync('user_token') || ''
    const storedUserInfo = uni.getStorageSync('userInfo') || null
    const storedOpenid = uni.getStorageSync('openid') || ''

    if (storedToken !== userToken.value) {
      userToken.value = storedToken
    }

    if (storedUserInfo !== userInfo.value) {
      userInfo.value = storedUserInfo
    }

    if (storedOpenid !== openid.value) {
      openid.value = storedOpenid
    }

    isLoggedIn.value = !!userToken.value && !!openid.value
  }

  const login = async (): Promise<LoginResult> => {
    if (loginPromise) {
      return loginPromise
    }

    loginPromise = (async () => {
    try {
      syncStateFromStorage()

      // 已登录则直接返回
      if (userToken.value && openid.value) {
        console.log('useAuth: 已登录,token:', userToken.value)
        return { success: true, token: userToken.value, user: userInfo.value }
      }

      // 获取微信授权码
      let code = ''

      // #ifdef MP-WEIXIN
      const loginRes = await uni.login({ provider: 'weixin' })
      if (loginRes.errMsg !== 'login:ok') {
        console.error('useAuth: 获取微信授权码失败', loginRes.errMsg)
        return { success: false, error: '获取授权码失败' }
      }
      code = loginRes.code
      // #endif

      // #ifndef MP-WEIXIN
      // 开发环境使用模拟code
      code = 'dev_' + Date.now()
      console.log('useAuth: 开发环境模拟code:', code)
      // #endif

      if (!code) {
        console.error('useAuth: code为空')
        return { success: false, error: '授权码为空' }
      }

      // 调用后端微信登录接口
      const res = await uni.request({
        url: `${API_BASE_URL}/api/v1/auth/user/wechat-login`,
        method: 'POST',
        data: { code },
        header: {
          'Content-Type': 'application/json'
        }
      }) as unknown as { data: ApiResponse<WechatLoginResponse> | WechatLoginResponse | unknown }

      const responseData = res.data as any

      if (typeof responseData?.code === 'number' && responseData.code !== 0) {
        const errorMessage = typeof responseData?.message === 'string' && responseData.message.trim()
          ? responseData.message
          : '登录失败'
        console.error('useAuth: 登录失败', responseData)
        return { success: false, error: errorMessage }
      }

      const payload: WechatLoginResponse | undefined = responseData?.data ?? responseData
      const token = payload?.token
      const user = payload?.user

      if (token && user?.openid) {
        userToken.value = token
        userInfo.value = user
        openid.value = user.openid
        isLoggedIn.value = true

        uni.setStorageSync('user_token', token)
        uni.setStorageSync('userInfo', user)
        uni.setStorageSync('openid', user.openid)

        console.log('useAuth: 登录成功,token:', token)
        console.log('useAuth: 用户信息:', user)

        return { success: true, token, user }
      }

      const errorMessage = typeof responseData?.message === 'string' && responseData.message.trim()
        ? responseData.message
        : '登录失败'
      console.error('useAuth: 登录失败', responseData)
      return { success: false, error: errorMessage }
    } catch (error: any) {
      console.error('useAuth: 登录异常', error)

      return { success: false, error: error.message || '登录异常' }
    }
    })().finally(() => {
      loginPromise = null
    })

    return loginPromise
  }

  const logout = () => {
    userToken.value = ''
    userInfo.value = null
    openid.value = ''
    isLoggedIn.value = false

    uni.removeStorageSync('user_token')
    uni.removeStorageSync('userInfo')
    uni.removeStorageSync('openid')

    console.log('useAuth: 退出登录')
  }

  const ensureAuth = async (): Promise<boolean> => {
    syncStateFromStorage()

    if (!isAuthenticated()) {
      const result = await login()
      return result.success
    }
    return true
  }

  const getToken = (): string => {
    syncStateFromStorage()
    return userToken.value || ''
  }

  const getUserInfo = () => {
    syncStateFromStorage()
    return userInfo.value
  }

  const getOpenid = (): string => {
    syncStateFromStorage()
    return openid.value || ''
  }

  const isAuthenticated = (): boolean => {
    syncStateFromStorage()
    return !!userToken.value && !!openid.value
  }

  const refreshLogin = async (): Promise<boolean> => {
    logout()
    const result = await login()
    return result.success
  }

  return {
    token: userToken,
    userInfo,
    openid,
    isLoggedIn,
    login,
    logout,
    ensureAuth,
    getToken,
    getUserInfo,
    getOpenid,
    isAuthenticated,
    refreshLogin
  }
}

export default useAuth
