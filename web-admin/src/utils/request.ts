import axios from 'axios'
import type { AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { API_BASE_URL } from '@/config/env'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 20000
})
const loginPath = `${import.meta.env.BASE_URL.replace(/\/$/, '')}/login`

request.interceptors.request.use((config) => {
  const token = window.localStorage.getItem('sp_token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    const status = error?.response?.status
    const message = error?.response?.data?.message || error?.message || '网络请求失败'

    if (status === 401) {
      window.localStorage.removeItem('sp_token')
      window.localStorage.removeItem('sp_info')
      if (window.location.pathname !== loginPath) {
        window.location.href = loginPath
      }
    }

    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export function unwrapApiResponse<T>(response: AxiosResponse<ApiResponse<T>>): T {
  const payload = response.data
  if (payload.code === 0) {
    return payload.data
  }

  const message = payload.message || '请求失败'
  ElMessage.error(message)
  throw new Error(message)
}

export default request
