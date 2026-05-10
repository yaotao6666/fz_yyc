/**
 * 微信授权登录工具
 * 提供自动登录和token管理功能
 * openid作为用户唯一标识
 */

import { ref } from 'vue'

const token = ref(uni.getStorageSync('token') || '')
const userInfo = ref(uni.getStorageSync('userInfo') || null)
const isLoggedIn = ref(!!token.value)
const openid = ref(uni.getStorageSync('openid') || '')

interface LoginResponse {
  token: string
  user: {
    id: number
    openid: string
    nickname: string
  }
}

interface LoginResult {
  success: boolean
  token?: string
  user?: any
  error?: string
}

export function useAuth() {
  const login = async (): Promise<LoginResult> => {
    try {
      // 已登录则直接返回
      if (token.value && openid.value) {
        console.log('useAuth: 已登录,token:', token.value)
        return { success: true, token: token.value, user: userInfo.value }
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
        url: 'http://localhost:8080/api/v1/auth/user/wechat-login',
        method: 'POST',
        data: { code },
        header: {
          'Content-Type': 'application/json'
        }
      }) as unknown as { data: LoginResponse }

      if (res.data?.token) {
        token.value = res.data.token
        userInfo.value = res.data.user
        openid.value = res.data.user.openid
        isLoggedIn.value = true

        // 保存到本地存储
        uni.setStorageSync('token', res.data.token)
        uni.setStorageSync('userInfo', res.data.user)
        uni.setStorageSync('openid', res.data.user.openid)

        console.log('useAuth: 登录成功,token:', res.data.token)
        console.log('useAuth: 用户信息:', res.data.user)

        return { success: true, token: res.data.token, user: res.data.user }
      }

      console.error('useAuth: 登录失败,无token')
      return { success: false, error: '登录失败' }
    } catch (error: any) {
      console.error('useAuth: 登录异常', error)

      // #ifndef MP-WEIXIN
      // 非微信环境联调时，后端不可用则退回到模拟登录，保证店铺页可继续验证链路。
      const mockToken = 'dev_token_' + Date.now()
      const mockUser = {
        id: 1,
        openid: 'mock_openid_dev',
        nickname: '测试用户'
      }

      token.value = mockToken
      userInfo.value = mockUser
      openid.value = mockUser.openid
      isLoggedIn.value = true

      uni.setStorageSync('token', mockToken)
      uni.setStorageSync('userInfo', mockUser)
      uni.setStorageSync('openid', mockUser.openid)

      console.log('useAuth: 开发环境模拟登录成功')

      return { success: true, token: mockToken, user: mockUser }
      // #endif

      return { success: false, error: error.message || '登录异常' }
    }
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    openid.value = ''
    isLoggedIn.value = false

    uni.removeStorageSync('token')
    uni.removeStorageSync('userInfo')
    uni.removeStorageSync('openid')

    console.log('useAuth: 退出登录')
  }

  const ensureAuth = async (): Promise<boolean> => {
    if (!isLoggedIn.value) {
      const result = await login()
      return result.success
    }
    return true
  }

  const getToken = (): string => {
    return token.value || uni.getStorageSync('token') || ''
  }

  const getUserInfo = () => {
    return userInfo.value || uni.getStorageSync('userInfo')
  }

  const getOpenid = (): string => {
    return openid.value || uni.getStorageSync('openid') || ''
  }

  const isAuthenticated = (): boolean => {
    return !!token.value && !!openid.value
  }

  const refreshLogin = async (): Promise<boolean> => {
    logout()
    const result = await login()
    return result.success
  }

  return {
    token,
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
