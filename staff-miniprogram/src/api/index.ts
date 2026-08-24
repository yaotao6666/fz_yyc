/**
 * 服务人员端 API 封装
 */
import { getApiBaseUrl } from '@/config/env'
import type { FittingRecommendedProduct, FollowUpTaskResult } from '@/types'

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

/** 获取七牛上传凭证 */
async function getUploadToken() {
  const res: any = await request({ url: '/api/v1/upload/token' })
  if (res?.code !== 0) throw new Error(res?.message || '获取上传凭证失败')
  return res?.data
}

/** 资质材料上传 - 客户端直传七牛云 */
export async function uploadQualificationFile(filePath: string): Promise<{ url: string; key: string }> {
  uni.showLoading({ title: '上传中...', mask: true })
  try {
    const data = await getUploadToken()
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

  /** 资料变更申请（进入待审核） */
  requestProfileChange: (data: { name: string; phone: string; avatar?: string; qualifications?: { type: string; name: string; url: string }[] }) =>
    request({ url: '/api/v1/service-staff/profile', method: 'PUT', data }),

  /** 我的审核记录 */
  getMyAuditList: () =>
    request({ url: '/api/v1/service-staff/audits' }),

  /** 接单统计 */
  getStatistics: () =>
    request({ url: '/api/v1/service-staff/statistics' })
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

// 规范化为任意数组（兼容 JSON 字符串）
function toJsonArray(value: any): any[] {
  if (Array.isArray(value)) return value
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed
    } catch {
      // 非 JSON 字符串视为无数组内容
    }
  }
  return []
}

// 照护计划/记录中的对象字段（vitals 等）可能返回 JSON 字符串，规范化为对象
function toJsonObject(value: any): any {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return parsed
    } catch {
      // 解析失败视为空对象
    }
  }
  return {}
}

// 规范化护理项数组（items/nursing_items 可能返回 JSON 字符串，元素可能为字符串或对象）
function normalizeNursingItems(value: any): { name: string; desc?: string; done?: boolean; remark?: string }[] {
  return toJsonArray(value)
    .map((v) => {
      // 兼容元素为纯字符串的写法（如 ["翻身拍背"]）
      const obj = typeof v === 'string' ? { name: v } : (v && typeof v === 'object' ? v : {})
      return {
        name: obj.name || '',
        desc: obj.desc || '',
        done: obj.done != null ? !!obj.done : undefined,
        remark: obj.remark || ''
      }
    })
    .filter((v) => v.name)
}

// 规范化单条照护记录中的 JSON 字段
function normalizeCareVisit(item: any): any {
  if (!item || typeof item !== 'object') return item
  return {
    ...item,
    nursing_items: normalizeNursingItems(item.nursing_items),
    vitals: toJsonObject(item.vitals),
    photos: normalizeStringArray(item.photos)
  }
}

// 规范化照护计划中的 JSON 字段（items/visits 可能返回 JSON 字符串）
function normalizeCarePlan(item: any): any {
  if (!item || typeof item !== 'object') return item
  return {
    ...item,
    items: normalizeNursingItems(item.items),
    visits: toJsonArray(item.visits).map((v: any) => normalizeCareVisit(v))
  }
}

// 规范化随访任务中的 result JSON 字段（兼容字符串与对象）
function normalizeFollowUpTask(item: any): any {
  if (!item || typeof item !== 'object') return item
  return { ...item, result: toJsonObject(item.result) }
}

// 规范化宣教文章中的 tags JSON 字段（兼容字符串与数组）
function normalizeEducationArticle(item: any): any {
  if (!item || typeof item !== 'object') return item
  return { ...item, tags: normalizeStringArray(item.tags) }
}

// 规范化生命体征记录中的 extra JSON 字段（兼容字符串与对象）
function normalizeMonitoring(item: any): any {
  if (!item || typeof item !== 'object') return item
  return { ...item, extra: toJsonObject(item.extra) }
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

  /** 获取客户评估记录（分页倒序） */
  async getResidentAssessments(userId: number | string, params?: { page?: number; page_size?: number }) {
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

  /** 获取我的照护计划列表（分页，仅返回当前服务人员负责的计划） */
  async getMyCarePlans(params?: { page?: number; page_size?: number }) {
    const res: any = await request({ url: '/api/v1/service-staff/care-plans', data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    const data = res?.data || {}
    const list = Array.isArray(data.list) ? data.list.map((item: any) => normalizeCarePlan(item)) : []
    return { list, total: data.total || 0 }
  },

  /** 获取照护计划详情（含历史照护记录 visits，倒序） */
  async getCarePlanDetail(id: number | string) {
    const res: any = await request({ url: `/api/v1/service-staff/care-plans/${id}` })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeCarePlan(res?.data)
  },

  /** 录入上门照护记录（plan_id 与 order_id 至少提供一个） */
  async createCareVisit(data: {
    plan_id?: number | null
    order_id?: number | null
    visit_at?: string
    nursing_items?: { name: string; done?: boolean; remark?: string }[]
    vitals?: { blood_pressure?: string; blood_glucose?: string; heart_rate?: string; oxygen?: string; weight?: string }
    photos?: string[]
    remark?: string
    follow_up_advice?: string
  }) {
    const res: any = await request({ url: '/api/v1/service-staff/care-visits', method: 'POST', data })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeCareVisit(res?.data)
  },

  /** 获取我的随访任务列表（分页倒序，可按状态筛选；status 传 0/1/2） */
  async getMyFollowUpTasks(params?: { status?: number; page?: number; page_size?: number }) {
    const res: any = await request({ url: '/api/v1/service-staff/follow-up-tasks', data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    const data = res?.data || {}
    const list = Array.isArray(data.list) ? data.list.map((item: any) => normalizeFollowUpTask(item)) : []
    return { list, total: data.total || 0 }
  },

  /** 获取随访任务详情 */
  async getFollowUpTaskDetail(id: number | string) {
    const res: any = await request({ url: `/api/v1/service-staff/follow-up-tasks/${id}` })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeFollowUpTask(res?.data)
  },

  /** 完成随访任务（提交随访结果） */
  async completeFollowUpTask(id: number | string, result: FollowUpTaskResult) {
    const res: any = await request({ url: `/api/v1/service-staff/follow-up-tasks/${id}/complete`, method: 'POST', data: { result } })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeFollowUpTask(res?.data)
  },

  /** 跳过随访任务 */
  async skipFollowUpTask(id: number | string) {
    const res: any = await request({ url: `/api/v1/service-staff/follow-up-tasks/${id}/skip`, method: 'POST' })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeFollowUpTask(res?.data)
  },

  /** 获取已发布的健康宣教文章（供随访选用） */
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
  },

  /** 获取居民生命体征监测记录（分页倒序） */
  async getResidentMonitoring(userId: number | string, params?: { page?: number; page_size?: number }) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/monitoring`, data: params })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    const data = res?.data || {}
    const list = Array.isArray(data.list) ? data.list.map((item: any) => normalizeMonitoring(item)) : []
    return { list, total: data.total || 0 }
  },

  /** 录入居民生命体征（record_type: 1血压 2血糖 3心率 4血氧 5体重） */
  async createResidentMonitoring(
    userId: number | string,
    data: { record_type: number; value: number; unit?: string; extra?: Record<string, unknown>; recorded_at?: string; remark?: string }
  ) {
    const res: any = await request({ url: `/api/v1/service-staff/residents/${userId}/monitoring`, method: 'POST', data })
    if (res?.code !== 0) {
      uni.showToast({ title: res?.message || '请求失败', icon: 'none' })
      throw new Error(res?.message)
    }
    return normalizeMonitoring(res?.data)
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
