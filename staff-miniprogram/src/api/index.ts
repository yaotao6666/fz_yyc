/**
 * 服务人员端 API 封装
 */
import { getApiBaseUrl } from '@/config/env'
import type { FittingRecommendedProduct } from '@/types'

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

/** 获取七牛上传凭证（mime 限定：image/* 默认 / audio/* 录音） */
async function getUploadToken(mime?: string) {
  const url = mime === 'audio/*' ? '/api/v1/upload/token?mime=audio/*' : '/api/v1/upload/token'
  const res: any = await request({ url })
  if (res?.code !== 0) throw new Error(res?.message || '获取上传凭证失败')
  return res?.data
}

/** 资质材料上传 - 客户端直传七牛云 */
export async function uploadQualificationFile(filePath: string): Promise<{ url: string; key: string }> {
  return uploadToQiniu(filePath)
}

/** 服务录音上传 - 客户端直传七牛云（限定 audio/* 凭证） */
export async function uploadServiceAudio(filePath: string): Promise<{ url: string; key: string }> {
  return uploadToQiniu(filePath, 'audio/*')
}

/** 七牛客户端直传（mime 限定 image/* 或 audio/*） */
async function uploadToQiniu(filePath: string, mime?: string): Promise<{ url: string; key: string }> {
  uni.showLoading({ title: '上传中...', mask: true })
  try {
    const data = await getUploadToken(mime)
    const ext = (filePath.split('.').pop() || 'jpg')
    const key = `${data.prefix}/${Date.now()}.${ext}`
    const domain = (data.domain || '').replace(/\/+$/, '')
    return new Promise((resolve, reject) => {
      uni.uploadFile({
        url: data.upload_url || 'https://up.qiniup.com',
        method: 'POST',
        filePath,
        name: 'file',
        formData: { token: data.token, key },
        success: (res) => {
          uni.hideLoading()
          if (res.statusCode === 200) {
            const parsed = JSON.parse(res.data)
            if (parsed.key) {
              resolve({ url: `${domain}/${String(parsed.key).replace(/^\/+/, '')}`, key: parsed.key })
            } else {
              uni.showToast({ title: '上传失败', icon: 'none' })
              reject(new Error('上传失败'))
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
  } catch (error) {
    uni.hideLoading()
    uni.showToast({ title: '获取上传凭证失败', icon: 'none' })
    throw error
  }
}

export const staffAuthApi = {
  /** 账号密码登录 */
  login: (username: string, password: string) =>
    request({ url: '/api/v1/service-staff/login', method: 'POST', needAuth: false, data: { username, password } }),

  /** 微信快捷登录 */
  wechatLogin: (code: string) =>
    request({ url: '/api/v1/service-staff/wechat-login', method: 'POST', needAuth: false, data: { code } }),

  /** 注册申请（可携带资质材料） */
  register: (data: { username: string; password: string; name: string; phone: string; qualifications?: { type: string; name: string; url: string }[] }) =>
    request({ url: '/api/v1/service-staff/register', method: 'POST', needAuth: false, data })
}

/* ============ 工单/待办接口 ============ */

export const staffWorkorderApi = {
  /** 待办列表（待出发+服务中） */
  getTodoList: () =>
    request({ url: '/api/v1/service-staff/todo' }),

  /** 待接订单列表 */
  getPendingOrders: (params?: { category?: number; page?: number; page_size?: number }) =>
    request({ url: '/api/v1/service-staff/orders/pending', data: params }),

  /** 已接订单列表 */
  getAcceptedOrders: (params?: { category?: number; biz_status?: number; page?: number; page_size?: number }) =>
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
    request({ url: `/api/v1/service-staff/orders/${id}/check-out`, method: 'POST', data: { remark } }),

  /** 放弃工单（待出发阶段，biz_status === 2） */
  giveUpOrder: (id: string | number) =>
    request({ url: `/api/v1/service-staff/orders/${id}/give-up`, method: 'POST' })
}

/* ============ 服务过程安全接口（阶段三） ============ */

export const staffSafetyApi = {
  /** 服务中上报定位（60s/次） */
  reportLocation: (id: string | number, lat: number, lng: number) =>
    request({ url: `/api/v1/service-staff/orders/${id}/location`, method: 'POST', data: { lat, lng } }),

  /** 提交服务录音 URL（七牛直传后） */
  submitAudio: (id: string | number, audioUrl: string) =>
    request({ url: `/api/v1/service-staff/orders/${id}/audio`, method: 'POST', data: { audio_url: audioUrl } }),

  /** 一键SOS */
  sos: (data: { order_id?: number; lat: number; lng: number; address?: string }) =>
    request({ url: '/api/v1/service-staff/sos', method: 'POST', data }),

  /** 我的服务区域 */
  getMyRegion: () =>
    request({ url: '/api/v1/service-staff/my-region' }),

  /** 获取当前生效协议（默认 type=3 录音/定位授权） */
  getActiveAgreement: (type: number = 3) =>
    request({ url: '/api/v1/service-staff/agreements', data: { type } }),

  /** 同意协议留痕 */
  consentAgreement: (id: string | number) =>
    request({ url: `/api/v1/service-staff/agreements/${id}/consent`, method: 'POST' })
}

/* ============ 个人信息接口 ============ */

export const staffProfileApi = {
  getProfile: () =>
    request({ url: '/api/v1/service-staff/profile' }),

  /** 资料变更申请（进入待审核） */
  requestProfileChange: (data: { name: string; phone: string; avatar?: string; qualifications?: { type: string; name: string; url: string }[] }) =>
    request({ url: '/api/v1/service-staff/profile', method: 'PUT', data }),

  /** 我的审核记录 */
  getMyAuditList: () =>
    request({ url: '/api/v1/service-staff/audits' }),

  /** 接单统计 */
  getStatistics: () =>
    request({ url: '/api/v1/service-staff/statistics' }),

  /** 我的质量分与近期评价（阶段四） */
  getMyQualityScore: () =>
    request({ url: '/api/v1/service-staff/my-quality-score' })
}

/* ============ 健康档案/评估接口 ============ */

// 后端 JSON 列可能返回字符串或数组，统一规范化为字符串数组
function normalizeStringArray(value: any): string[] {
  if (Array.isArray(value)) {
    return value.filter((v): v is string => typeof v === 'string')
  }
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return parsed.filter((v): v is string => typeof v === 'string')
      }
    } catch {
      // 非 JSON 字符串，按单元素处理
    }
    return [value]
  }
  return []
}

// 规范化健康档案中的 JSON 数组字段（兼容字符串与数组）
function normalizeHealthRecord(record: any): any {
  if (!record || typeof record !== 'object') return record
  const jsonListKeys = ['past_history', 'allergy_history', 'family_history', 'surgery_history', 'medication_list', 'chronic_tags']
  const normalized: any = { ...record }
  jsonListKeys.forEach((key) => {
    normalized[key] = normalizeStringArray(record[key])
  })
  return normalized
}

// 推荐商品可能为 JSON 字符串或对象数组，统一规范化为对象数组
function normalizeRecommendedProducts(value: any): FittingRecommendedProduct[] {
  let list: any[] = []
  if (Array.isArray(value)) {
    list = value
  } else if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) list = parsed
    } catch {
      // 非 JSON 字符串视为无推荐商品
    }
  }
  return list
    .filter((v): v is Record<string, any> => v && typeof v === 'object' && v.product_id != null)
    .map((v) => ({
      product_id: Number(v.product_id),
      name: v.name || '',
      reason: v.reason || '',
      sale_type: v.sale_type != null ? Number(v.sale_type) : undefined
    }))
}

// 规范化适配建议中的 JSON 数组字段（recommended_products 可能返回字符串）
function normalizeFittingRecommendation(item: any): any {
  if (!item || typeof item !== 'object') return item
  return { ...item, recommended_products: normalizeRecommendedProducts(item.recommended_products) }
}

// 规范化宣教文章中的 tags JSON 字段（兼容字符串与数组）
function normalizeEducationArticle(item: any): any {
  if (!item || typeof item !== 'object') return item
  return { ...item, tags: normalizeStringArray(item.tags) }
}

export const staffHealthApi = {
  /** 获取启用的评估量表列表 */
  async getAssessmentForms() {
    const res: any = await request({ url: '/api/v1/service-staff/assessment-forms' })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return res?.data
  },

  /** 获取客户健康档案（未建档返回 null；无权限 code=1003） */
  async getResidentHealthRecord(userId: number | string) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/health-record` })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeHealthRecord(res?.data)
  },

  /** 获取客户评估记录（分页倒序；传入 record_id 时仅返回该档案的评估） */
  async getResidentAssessments(userId: number | string, params?: { page?: number; page_size?: number; record_id?: number | string }) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/assessments`, data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return res?.data
  },

  /** 为客户登记评估 */
  async createResidentAssessment(userId: number | string, data: { form_id: number; answers: Record<string, string>; symptom_desc?: string }) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/assessments`, method: 'POST', data })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return res?.data
  },

  /** 获取客户康复辅具适配建议列表（分页倒序） */
  async getResidentFittingRecommendations(userId: number | string, params?: { page?: number; page_size?: number }) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/fitting-recommendations`, data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    const data = res?.data || {}
    const list = Array.isArray(data.list) ? data.list.map((item: any) => normalizeFittingRecommendation(item)) : []
    return { list, total: data.total || 0 }
  },

  /** 为客户生成康复辅具适配建议 */
  async createResidentFittingRecommendation(
    userId: number | string,
    data: {
      assessment_id?: number | null
      symptom_desc?: string
      fitting_result?: string
      recommended_products: { product_id: number; reason?: string }[]
    }
  ) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/fitting-recommendations`, method: 'POST', data })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeFittingRecommendation(res?.data)
  },

  /** 获取已发布的健康宣教文章（供参考选用） */
  async getEducationArticles() {
    const res: any = await request({ url: '/api/v1/service-staff/health-education' })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    // 兼容直接返回数组与 {list} 两种结构
    const data = res?.data
    const raw = Array.isArray(data) ? data : (data?.list || [])
    return Array.isArray(raw) ? raw.map((item: any) => normalizeEducationArticle(item)) : []
  }
}

/* ============ 商城商品接口（复用公开商品列表，OptionalJWTAuth） ============ */

export const staffStoreApi = {
  /** 商品列表 */
  async getProducts(params?: { keyword?: string; category_id?: number; page?: number; page_size?: number }) {
    const res: any = await request({ url: '/api/v1/store/products', data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return res?.data || { list: [], total: 0 }
  }
}
