/**
 * API 请求封装
 * 基于 uni.request 封装统一请求方法
 */

import { API_BASE_URL } from '../config/env'

// 响应码
export const ResponseCode = {
  SUCCESS: 0,
  UNAUTHORIZED: 1002,
  FORBIDDEN: 1003,
  NOT_FOUND: 1004,
  SERVER_ERROR: 9001
} as const

// API 统一响应格式
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface RequestError extends Error {
  code?: number
  statusCode?: number
  response?: unknown
}

// 请求配置
export interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  loading?: boolean
  loadingText?: string
  showErrorToast?: boolean
}

/**
 * 获取 Token
 */
function getToken(url: string): string {
  if (url.startsWith('/api/v1/user/')) {
    return uni.getStorageSync('user_token') || ''
  }
  if (url.startsWith('/api/v1/store/')) {
    return uni.getStorageSync('user_token') || ''
  }
  return uni.getStorageSync('token') || ''
}

function getResponseMessage(response: unknown, fallbackMessage: string): string {
  if (typeof response === 'string' && response.trim()) {
    return response
  }

  if (response && typeof response === 'object' && 'message' in response) {
    const message = (response as { message?: unknown }).message
    if (typeof message === 'string' && message.trim()) {
      return message
    }
  }

  return fallbackMessage
}

function getResponseCode(response: unknown): number | undefined {
  if (response && typeof response === 'object' && 'code' in response) {
    const code = (response as { code?: unknown }).code
    if (typeof code === 'number') {
      return code
    }
  }

  return undefined
}

function createRequestError(params: {
  message: string
  code?: number
  statusCode?: number
  response?: unknown
}): RequestError {
  const { message, code, statusCode, response } = params
  const error = new Error(message) as RequestError
  error.code = code
  error.statusCode = statusCode
  error.response = response
  return error
}

let _loadingCount = 0

function showLoading(title: string) {
  _loadingCount++
  if (_loadingCount === 1) {
    uni.showLoading({ title, mask: true })
  }
}

function hideLoading() {
  if (_loadingCount > 0) {
    _loadingCount--
  }
  if (_loadingCount === 0) {
    uni.hideLoading()
  }
}

function sanitizeGetParams(data: unknown) {
  if (!data || typeof data !== 'object' || Array.isArray(data)) {
    return data
  }

  const sanitizedEntries = Object.entries(data).filter(([, value]) => {
    if (value === undefined || value === null) {
      return false
    }
    if (typeof value === 'string' && value.trim() === '') {
      return false
    }
    return true
  })

  return Object.fromEntries(sanitizedEntries)
}

/**
 * 请求核心方法
 */
function request<T = any>(options: RequestOptions): Promise<T> {
  const {
    url,
    method = 'GET',
    data,
    header = {},
    loading = true,
    loadingText = '加载中...',
    showErrorToast = true
  } = options

  // 显示加载中
  if (loading) {
    showLoading(loadingText)
  }

  const isMerchantRequest = url.startsWith('/api/v1/merchant/')
  const isUserRequest = url.startsWith('/api/v1/user/')
  const isStoreRequest = url.startsWith('/api/v1/store/')
  const requestData = method === 'GET' ? sanitizeGetParams(data) : data

  const requestToken = getToken(url)
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...header
  }
  if (requestToken) {
    headers['Authorization'] = `Bearer ${requestToken}`
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${API_BASE_URL}${url}`,
      method,
      data: requestData,
      header: headers,
      success: (res) => {
        if (loading) {
          hideLoading()
        }

        const { statusCode, data: response } = res
        
        if (statusCode === 200) {
          const apiResponse = response as ApiResponse<T>
          
          if (apiResponse.code === ResponseCode.SUCCESS) {
            resolve(apiResponse.data)
          } else if (apiResponse.code === ResponseCode.UNAUTHORIZED) {
            if (isMerchantRequest) {
              uni.removeStorageSync('token')
              uni.removeStorageSync('merchantId')
              uni.removeStorageSync('staff')
              uni.removeStorageSync('merchantInfo')
              uni.$emit('merchant-session-expired')
              uni.showToast({ title: '登录已失效，请重新登录', icon: 'none' })
              uni.reLaunch({ url: '/pages/auth/login' })
            } else if (isUserRequest || isStoreRequest) {
              uni.removeStorageSync('user_token')
              uni.removeStorageSync('userInfo')
              uni.removeStorageSync('openid')
              uni.showToast({ title: '登录已过期，请重新进入', icon: 'none' })
            } else {
              uni.removeStorageSync('token')
              uni.removeStorageSync('userInfo')
              uni.removeStorageSync('openid')
              uni.showToast({ title: '请先登录', icon: 'none' })
              uni.reLaunch({ url: '/pages/auth/login' })
            }
            reject(createRequestError({
              message: apiResponse.message || '未授权',
              code: apiResponse.code,
              statusCode
            }))
          } else {
            const errorMessage = apiResponse.message || '请求失败'
            if (showErrorToast) {
              uni.showToast({ title: errorMessage, icon: 'none' })
            }
            reject(createRequestError({
              message: errorMessage,
              code: apiResponse.code,
              statusCode,
              response
            }))
          }
        } else {
          // 非 200 时优先透传后端给出的错误码与错误信息，避免丢失排查线索。
          const errorMessage = getResponseMessage(response, `请求失败(${statusCode})`)
          const errorCode = getResponseCode(response)
          if (statusCode === 401 && errorCode === ResponseCode.UNAUTHORIZED) {
            if (isMerchantRequest) {
              uni.removeStorageSync('token')
              uni.removeStorageSync('merchantId')
              uni.removeStorageSync('staff')
              uni.removeStorageSync('merchantInfo')
              uni.$emit('merchant-session-expired')
              uni.showToast({ title: '登录已失效，请重新登录', icon: 'none' })
              uni.reLaunch({ url: '/pages/auth/login' })
            } else if (isUserRequest || isStoreRequest) {
              uni.removeStorageSync('user_token')
              uni.removeStorageSync('userInfo')
              uni.removeStorageSync('openid')
              uni.showToast({ title: '登录已过期，请重新进入', icon: 'none' })
            } else {
              uni.removeStorageSync('token')
              uni.removeStorageSync('userInfo')
              uni.removeStorageSync('openid')
              uni.showToast({ title: '请先登录', icon: 'none' })
              uni.reLaunch({ url: '/pages/auth/login' })
            }
          }
          if (showErrorToast) {
            uni.showToast({ title: errorMessage, icon: 'none' })
          }
          reject(createRequestError({
            message: errorMessage,
            code: errorCode,
            statusCode,
            response
          }))
        }
      },
      fail: (err) => {
        if (loading) {
          hideLoading()
        }
        uni.showToast({ title: '网络请求失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

/**
 * GET 请求
 */
export function get<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'GET', data, ...options })
}

/**
 * POST 请求
 */
export function post<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'POST', data, ...options })
}

/**
 * PUT 请求
 */
export function put<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'PUT', data, ...options })
}

/**
 * DELETE 请求
 */
export function del<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'DELETE', data, ...options })
}

/**
 * 上传文件
 */
export function upload<T = any>(
  url: string,
  filePath: string,
  name: string = 'file',
  formData?: Record<string, string>
): Promise<T> {
  return new Promise((resolve, reject) => {
    uni.showLoading({ title: '上传中...', mask: true })
    
    uni.uploadFile({
      url: `${API_BASE_URL}${url}`,
      filePath,
      name,
      formData,
      header: {
        'Authorization': `Bearer ${getToken(url)}`
      },
      success: (res) => {
        uni.hideLoading()
        if (res.statusCode === 200) {
          const data = JSON.parse(res.data) as ApiResponse<T>
          if (data.code === ResponseCode.SUCCESS) {
            resolve(data.data)
          } else {
            uni.showToast({ title: data.message || '上传失败', icon: 'none' })
            reject(new Error(data.message))
          }
        } else {
          uni.showToast({ title: '上传失败', icon: 'none' })
          reject(new Error(`上传失败: ${res.statusCode}`))
        }
      },
      fail: (err) => {
        uni.hideLoading()
        uni.showToast({ title: '上传失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

export const uploadFile = upload

export default {
  get,
  post,
  put,
  del,
  upload
}
