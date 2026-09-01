import { del, get, post, put } from '../utils/request'
import type {
  AssessmentForm,
  EducationArticle,
  FittingRecommendation,
  FittingRecommendedProduct,
  HealthAssessment,
  HealthEducationCategory,
  HealthRecord,
  PaginationParams,
  PaginationResponse
} from '../types'

/**
 * 规范化字符串数组字段。
 * 后端部分数组字段可能以 JSON 字符串返回，统一规范化为 string[]。
 */
function normalizeStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === 'string')
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return parsed.filter((item): item is string => typeof item === 'string')
      }
    } catch (error) {
      console.warn('解析健康档案数组字段失败:', error)
    }
  }

  return []
}

/**
 * 规范化数值字段（身高/体重），空值返回 null。
 */
function normalizeNullableNumber(value: unknown): number | null {
  if (value === '' || value === null || value === undefined) {
    return null
  }
  const normalizedValue = Number(value)
  return Number.isFinite(normalizedValue) ? normalizedValue : null
}

function normalizeNullableScore(value: unknown): number | undefined {
  const normalizedValue = normalizeNullableNumber(value)
  return normalizedValue === null ? undefined : normalizedValue
}

/**
 * 规范化评估答案字段（可能为对象或 JSON 字符串）。
 */
function normalizeAnswers(value: unknown): Record<string, string> | undefined {
  if (value && typeof value === 'object') {
    return value as Record<string, string>
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object') {
        return parsed as Record<string, string>
      }
    } catch (error) {
      console.warn('解析评估答案失败:', error)
    }
  }

  return undefined
}

/**
 * 规范化健康档案，未建档（null/空）返回 null。
 */
function normalizeHealthRecord(data: any): HealthRecord | null {
  if (!data || typeof data !== 'object') {
    return null
  }

  return {
    ...data,
    id: Number(data.id || 0),
    user_id: Number(data.user_id || 0),
    height_cm: normalizeNullableNumber(data.height_cm),
    weight_kg: normalizeNullableNumber(data.weight_kg),
    blood_type: typeof data.blood_type === 'string' ? data.blood_type.replace(/型$/, '') : (data.blood_type || ''),
    past_history: normalizeStringArray(data.past_history),
    allergy_history: normalizeStringArray(data.allergy_history),
    family_history: normalizeStringArray(data.family_history),
    surgery_history: normalizeStringArray(data.surgery_history),
    medication_list: normalizeStringArray(data.medication_list),
    chronic_tags: normalizeStringArray(data.chronic_tags)
  }
}

/**
 * 规范化健康评估记录。
 */
function normalizeHealthAssessment(data: any): HealthAssessment {
  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    form_id: Number(data?.form_id || 0),
    total_score: normalizeNullableScore(data?.total_score),
    answers: normalizeAnswers(data?.answers),
    suggestions: normalizeStringArray(data?.suggestions),
    created_at: data?.created_at ? String(data.created_at) : undefined
  }
}

/**
 * 获取当前用户健康档案（未建档返回 null）。需要登录。
 */
export function getUserHealthRecord() {
  return get<HealthRecord | null>('/api/v1/user/health-record').then(data => normalizeHealthRecord(data))
}

/**
 * 保存健康档案（有则更新，无则创建；兼容单档案前端）。需要登录。
 */
export function saveUserHealthRecord(data: Partial<HealthRecord>) {
  return put<HealthRecord>('/api/v1/user/health-record', data).then(res => {
    return normalizeHealthRecord(res) as HealthRecord
  })
}

/**
 * 列出当前账号下的全部健康档案（多档案列表）。需要登录。
 */
export function listUserHealthRecords() {
  return get<HealthRecord[]>('/api/v1/user/health-records').then(list => {
    const arr = Array.isArray(list) ? list : []
    return arr.map(normalizeHealthRecord).filter((item): item is HealthRecord => !!item)
  })
}

/**
 * 新增一条健康档案（支持关系 relation）。需要登录。
 */
export function createUserHealthRecord(data: Partial<HealthRecord>) {
  return post<HealthRecord>('/api/v1/user/health-records', data).then(res => normalizeHealthRecord(res) as HealthRecord)
}

/**
 * 按档案 ID 更新健康档案。需要登录。
 */
export function updateUserHealthRecord(id: number, data: Partial<HealthRecord>) {
  return put<HealthRecord>(`/api/v1/user/health-records/${id}`, data).then(res => normalizeHealthRecord(res) as HealthRecord)
}

/**
 * 按档案 ID 删除健康档案；若被订单引用会被后端拒绝。需要登录。
 */
export function deleteUserHealthRecord(id: number) {
  return del<{ id: number }>(`/api/v1/user/health-records/${id}`)
}

/**
 * 获取启用的评估量表列表。
 */
export function getUserAssessmentForms() {
  return get<AssessmentForm[]>('/api/v1/user/assessment-forms').then(list => (Array.isArray(list) ? list : []))
}

/**
 * 获取我的评估记录列表（可按档案 record_id 过滤）。需要登录。
 */
export function getUserAssessments(params?: PaginationParams & { record_id?: number }) {
  return get<PaginationResponse<HealthAssessment>>('/api/v1/user/assessments', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeHealthAssessment)
    }
  })
}

/**
 * 提交自助评估。需要登录。record_id 可选：指定档案维度，缺省回写最近档案。
 */
export function createUserAssessment(data: {
  form_id: number
  record_id?: number
  answers: Record<string, string>
  symptom_desc?: string
}) {
  return post<HealthAssessment>('/api/v1/user/assessments', data).then(normalizeHealthAssessment)
}

/**
 * 规范化推荐商品列表字段。
 * 后端 recommended_products 可能以 JSON 字符串返回，统一规范化为数组。
 */
function normalizeRecommendedProducts(value: unknown): FittingRecommendedProduct[] {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is FittingRecommendedProduct => !!item && typeof item === 'object')
      .map(item => ({
        ...item,
        product_id: Number(item.product_id || 0),
        name: String(item.name || '')
      }))
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return normalizeRecommendedProducts(parsed)
      }
    } catch (error) {
      console.warn('解析推荐商品列表失败:', error)
    }
  }

  return []
}

/**
 * 规范化适配建议。
 */
function normalizeFittingRecommendation(data: any): FittingRecommendation {
  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    recommended_products: normalizeRecommendedProducts(data?.recommended_products),
    created_at: data?.created_at ? String(data.created_at) : undefined,
    updated_at: data?.updated_at ? String(data.updated_at) : undefined
  }
}

/**
 * 获取我的适配建议列表。需要登录。
 */
export function getUserFittingRecommendations(params?: PaginationParams) {
  return get<PaginationResponse<FittingRecommendation>>('/api/v1/user/fitting-recommendations', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeFittingRecommendation)
    }
  })
}

/**
 * 获取适配建议详情。需要登录。
 */
export function getUserFittingRecommendation(id: number) {
  return get<FittingRecommendation>(`/api/v1/user/fitting-recommendations/${id}`).then(normalizeFittingRecommendation)
}

/**
 * 确认适配建议（草稿→已确认）。需要登录。
 */
export function confirmUserFittingRecommendation(id: number) {
  return post<FittingRecommendation>(`/api/v1/user/fitting-recommendations/${id}/confirm`).then(normalizeFittingRecommendation)
}

// ============ 健康宣教 ============

/**
 * 规范化宣教文章。
 * 后端 tags 可能以 JSON 字符串返回，统一规范化为 string[]。
 */
function normalizeEducationArticle(data: any): EducationArticle {
  return {
    ...data,
    id: Number(data?.id || 0),
    title: String(data?.title || ''),
    category_id: data?.category_id !== undefined && data?.category_id !== null ? Number(data.category_id) || undefined : undefined,
    category: data?.category !== undefined && data?.category !== null ? String(data.category) : undefined,
    cover: data?.cover !== undefined && data?.cover !== null ? String(data.cover) : undefined,
    content: data?.content !== undefined && data?.content !== null ? String(data.content) : undefined,
    tags: normalizeStringArray(data?.tags),
    status: data?.status !== undefined && data?.status !== null ? Number(data.status) : undefined,
    publish_at: data?.publish_at ? String(data.publish_at) : undefined,
    views: data?.views !== undefined && data?.views !== null ? Number(data.views) : undefined
  }
}

/**
 * 获取健康宣教文章列表（已发布，按本人慢病标签匹配度排序）。需要登录。
 */
export function getUserEducationArticles(params?: { category?: string }) {
  return get<EducationArticle[]>('/api/v1/user/health-education', params).then(list =>
    (Array.isArray(list) ? list : []).map(normalizeEducationArticle)
  )
}

/**
 * 获取宣教文章详情（浏览量+1）。需要登录。
 */
export function getUserEducationArticle(id: number) {
  return get<EducationArticle>(`/api/v1/user/health-education/${id}`).then(normalizeEducationArticle)
}

/**
 * 获取首页健康宣教文章（公开，无需登录；分页返回 {list,total}）。
 */
export function getStoreEducationArticles(params?: { category_id?: number; page?: number; page_size?: number }) {
  return get<{ list: EducationArticle[]; total: number }>('/api/v1/store/education/articles', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return { ...res, list: list.map(normalizeEducationArticle) }
  })
}

/**
 * 获取启用中的健康宣教分类（两级，公开无需登录）。
 */
export function getStoreEducationCategories() {
  return get<HealthEducationCategory[]>('/api/v1/store/education/categories').then(list =>
    (Array.isArray(list) ? list : []).map(item => ({
      ...item,
      id: Number(item.id || 0),
      parent_id: Number(item.parent_id || 0),
      sort: Number(item.sort || 0),
      status: Number(item.status ?? 1)
    }))
  )
}
