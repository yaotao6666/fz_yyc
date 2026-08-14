/**
 * 服务人员端 API 封装
 */
import { getApiBaseUrl } from '@/config/env'

const BASE_URL = getApiBaseUrl()

export interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  needAuth?: boolean
}

function getToken(): string {
  return uni.getStorageSync('staff_token') || ''
}

export function request<T = any>(options: RequestOptions): Promise<T> {
  return new Promise((resolve, reject) => {
    const header: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.header || {})
    }
    if (options.needAuth !== false) {
      const token = getToken()
      if (token) header['Authorization'] = `Bearer ${token}`
    }

    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header,
      success: (res) => {
        const statusCode = res.statusCode
        if (statusCode >= 200 && statusCode < 300) {
          resolve(res.data as T)
        } else if (statusCode === 401) {
          uni.removeStorageSync('staff_token')
          uni.showToast({ title: '请重新登录', icon: 'none' })
          setTimeout(() => {
            uni.reLaunch({ url: '/pages/login/index' })
          }, 1500)
          reject(res)
        } else {
          reject(res)
        }
      },
      fail: reject
    })
  })
}

/* ============ 认证接口 ============ */

export const staffAuthApi = {
  /** 账号密码登录 */
  login: (username: string, password: string) =>
    request({ url: '/api/v1/service-staff/login', method: 'POST', needAuth: false, data: { username, password } }),

  /** 微信快捷登录 */
  wechatLogin: (code: string) =>
    request({ url: '/api/v1/service-staff/wechat-login', method: 'POST', needAuth: false, data: { code } }),

  /** 注册申请 */
  register: (data: { username: string; password: string; name: string; phone: string }) =>
    request({ url: '/api/v1/service-staff/register', method: 'POST', needAuth: false, data })
}

/* ============ 工单/待办接口 ============ */

export const staffWorkorderApi = {
  /** 待办列表（待出发+服务中） */
  getTodoList: () =>
    request({ url: '/api/v1/service-staff/todo' }),

  /** 待接订单列表 */
  getPendingOrders: (params?: { page?: number; page_size?: number }) =>
    request({ url: '/api/v1/service-staff/orders/pending', data: params }),

  /** 已接订单列表 */
  getAcceptedOrders: (params?: { biz_status?: number; page?: number; page_size?: number }) =>
    request({ url: '/api/v1/service-staff/orders/accepted', data: params }),

  /** 订单详情 */
  getWorkorderDetail: (id: string | number) =>
    request({ url: `/api/v1/service-staff/orders/${id}` }),

  /** 接单 */
  acceptOrder: (id: string | number) =>
    request({ url: `/api/v1/service-staff/orders/${id}/accept`, method: 'POST' }),

  /** 签到 */
  checkIn: (id: string | number, lat?: number, lng?: number) =>
    request({ url: `/api/v1/service-staff/orders/${id}/check-in`, method: 'POST', data: { lat, lng } }),

  /** 签退 */
  checkOut: (id: string | number, remark?: string) =>
    request({ url: `/api/v1/service-staff/orders/${id}/check-out`, method: 'POST', data: { remark } })
}

/* ============ 个人信息接口 ============ */

export const staffProfileApi = {
  getProfile: () =>
    request({ url: '/api/v1/service-staff/profile' }),

  /** 接单统计 */
  getStatistics: () =>
    request({ url: '/api/v1/service-staff/statistics' })
}

/* ============ 排班接口（暂为占位，后端尚未实现） ============ */

export const staffScheduleApi = {
  getScheduleList: (_month: string) =>
    Promise.resolve({ code: 0, data: [] } as any)
}
