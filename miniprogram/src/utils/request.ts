/**
 * API 请求封装
 * 基于 uni.request 封装统一请求方法
 */

// API 基础配置
const BASE_URL = process.env.NODE_ENV === 'development' 
  ? 'http://localhost:8080' 
  : 'https://api.example.com'

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

// 请求配置
interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  loading?: boolean
  loadingText?: string
}

/**
 * 获取 Token
 */
function getToken(): string {
  return uni.getStorageSync('token') || ''
}

/**
 * 请求核心方法
 */
function request<T = any>(options: RequestOptions): Promise<T> {
  const { url, method = 'GET', data, header = {}, loading = true, loadingText = '加载中...' } = options

  // 显示加载中
  if (loading) {
    uni.showLoading({ title: loadingText, mask: true })
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getToken()}`,
        ...header
      },
      success: (res) => {
        if (loading) {
          uni.hideLoading()
        }

        const { statusCode, data: response } = res
        
        if (statusCode === 200) {
          const apiResponse = response as ApiResponse<T>
          
          if (apiResponse.code === ResponseCode.SUCCESS) {
            resolve(apiResponse.data)
          } else if (apiResponse.code === ResponseCode.UNAUTHORIZED) {
            // Token 过期，跳转登录
            uni.removeStorageSync('token')
            uni.removeStorageSync('userInfo')
            uni.showToast({ title: '请先登录', icon: 'none' })
            uni.reLaunch({ url: '/pages/auth/login' })
            reject(new Error(apiResponse.message || '未授权'))
          } else {
            uni.showToast({ title: apiResponse.message || '请求失败', icon: 'none' })
            reject(new Error(apiResponse.message))
          }
        } else {
          uni.showToast({ title: `请求失败(${statusCode})`, icon: 'none' })
          reject(new Error(`请求失败: ${statusCode}`))
        }
      },
      fail: (err) => {
        if (loading) {
          uni.hideLoading()
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
      url: `${BASE_URL}${url}`,
      filePath,
      name,
      formData,
      header: {
        'Authorization': `Bearer ${getToken()}`
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
