import { get, post, put } from '../utils/request'
import type {
  AssessmentForm,
  CarePlan,
  CarePlanItem,
  CareVisit,
  EducationArticle,
  FittingRecommendation,
  FittingRecommendedProduct,
  FollowUpTask,
  FollowUpTaskResult,
  HealthAssessment,
  HealthMonitoring,
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
 * 保存健康档案（有则更新，无则创建）。需要登录。
 */
export function saveUserHealthRecord(data: Partial<HealthRecord>) {
  return put<HealthRecord>('/api/v1/user/health-record', data).then(res => {
    return normalizeHealthRecord(res) as HealthRecord
  })
}

/**
 * 获取启用的评估量表列表。
 */
export function getUserAssessmentForms() {
  return get<AssessmentForm[]>('/api/v1/user/assessment-forms').then(list => (Array.isArray(list) ? list : []))
}

/**
 * 获取我的评估记录列表。需要登录。
 */
export function getUserAssessments(params?: PaginationParams) {
  return get<PaginationResponse<HealthAssessment>>('/api/v1/user/assessments', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeHealthAssessment)
    }
  })
}

/**
 * 提交自助评估。需要登录。
 */
export function createUserAssessment(data: {
  form_id: number
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

// ============ 我的照护计划 ============

/**
 * 规范化护理项数组字段。
 * 后端 items/nursing_items 可能以 JSON 字符串返回，统一规范化为数组。
 */
function normalizeCarePlanItems(value: unknown): CarePlanItem[] {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is CarePlanItem => !!item && typeof item === 'object')
      .map(item => ({
        name: String(item.name || ''),
        desc: item.desc !== undefined && item.desc !== null ? String(item.desc) : undefined
      }))
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return normalizeCarePlanItems(parsed)
      }
    } catch (error) {
      console.warn('解析护理项列表失败:', error)
    }
  }

  return []
}

/**
 * 规范化护理记录中的护理项（含完成情况），可能为 JSON 字符串。
 */
function normalizeNursingItems(value: unknown): { name: string; done?: boolean; remark?: string }[] {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is Record<string, any> => !!item && typeof item === 'object')
      .map(item => ({
        name: String(item.name || ''),
        done: typeof item.done === 'boolean' ? item.done : undefined,
        remark: item.remark !== undefined && item.remark !== null ? String(item.remark) : undefined
      }))
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return normalizeNursingItems(parsed)
      }
    } catch (error) {
      console.warn('解析护理项完成情况失败:', error)
    }
  }

  return []
}

/**
 * 规范化生命体征字段（可能为 JSON 字符串）。
 */
function normalizeVitals(value: unknown): CareVisit['vitals'] {
  if (value && typeof value === 'object') {
    return value as CareVisit['vitals']
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object') {
        return parsed as CareVisit['vitals']
      }
    } catch (error) {
      console.warn('解析生命体征失败:', error)
    }
  }

  return undefined
}

/**
 * 规范化上门照护记录。
 */
function normalizeCareVisit(data: any): CareVisit {
  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    plan_id: data?.plan_id !== undefined && data?.plan_id !== null ? Number(data.plan_id) : null,
    order_id: data?.order_id !== undefined && data?.order_id !== null ? Number(data.order_id) : null,
    staff_id: data?.staff_id !== undefined && data?.staff_id !== null ? Number(data.staff_id) : undefined,
    visit_at: data?.visit_at ? String(data.visit_at) : undefined,
    nursing_items: normalizeNursingItems(data?.nursing_items),
    vitals: normalizeVitals(data?.vitals),
    photos: normalizeStringArray(data?.photos),
    remark: data?.remark !== undefined && data?.remark !== null ? String(data.remark) : undefined,
    follow_up_advice:
      data?.follow_up_advice !== undefined && data?.follow_up_advice !== null
        ? String(data.follow_up_advice)
        : undefined,
    created_at: data?.created_at ? String(data.created_at) : undefined
  }
}

/**
 * 规范化照护计划。
 */
function normalizeCarePlan(data: any): CarePlan {
  const rawVisits = data?.visits
  const visits = Array.isArray(rawVisits)
    ? rawVisits.filter((item: any) => !!item && typeof item === 'object').map(normalizeCareVisit)
    : typeof rawVisits === 'string' && rawVisits.trim()
      ? (() => {
          try {
            const parsed = JSON.parse(rawVisits)
            return Array.isArray(parsed) ? parsed.map(normalizeCareVisit) : undefined
          } catch (error) {
            console.warn('解析照护记录列表失败:', error)
            return undefined
          }
        })()
      : undefined

  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    name: String(data?.name || ''),
    plan_type: data?.plan_type !== undefined && data?.plan_type !== null ? Number(data.plan_type) : undefined,
    start_date: data?.start_date ? String(data.start_date) : undefined,
    end_date: data?.end_date ? String(data.end_date) : undefined,
    frequency: data?.frequency ? String(data.frequency) : undefined,
    goals: data?.goals ? String(data.goals) : undefined,
    items: normalizeCarePlanItems(data?.items),
    assigned_staff_id:
      data?.assigned_staff_id !== undefined && data?.assigned_staff_id !== null
        ? Number(data.assigned_staff_id)
        : null,
    order_id: data?.order_id !== undefined && data?.order_id !== null ? Number(data.order_id) : null,
    status: data?.status !== undefined && data?.status !== null ? Number(data.status) : undefined,
    visit_count:
      data?.visit_count !== undefined && data?.visit_count !== null ? Number(data.visit_count) : undefined,
    created_at: data?.created_at ? String(data.created_at) : undefined,
    updated_at: data?.updated_at ? String(data.updated_at) : undefined,
    visits
  }
}

/**
 * 获取我的照护计划列表。需要登录。
 */
export function getUserCarePlans(params?: PaginationParams) {
  return get<PaginationResponse<CarePlan>>('/api/v1/user/care-plans', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeCarePlan)
    }
  })
}

/**
 * 获取照护计划详情（含上门照护记录）。需要登录。
 */
export function getUserCarePlan(id: number) {
  return get<CarePlan>(`/api/v1/user/care-plans/${id}`).then(normalizeCarePlan)
}

// ============ 我的康复随访 ============

/**
 * 规范化数值数组字段（如随访结果中的文章 ID 列表）。
 */
function normalizeNumberArray(value: unknown): number[] {
  if (Array.isArray(value)) {
    return value
      .map(item => Number(item))
      .filter(item => Number.isFinite(item))
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return normalizeNumberArray(parsed)
      }
    } catch (error) {
      console.warn('解析数值数组字段失败:', error)
    }
  }

  return []
}

/**
 * 规范化随访任务结果字段。
 * 后端 result 可能以 JSON 字符串返回，统一规范化为对象。
 */
function normalizeFollowUpTaskResult(value: unknown): FollowUpTaskResult | undefined {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    const raw = value as Record<string, any>
    return {
      contact_method:
        raw?.contact_method !== undefined && raw?.contact_method !== null ? Number(raw.contact_method) : undefined,
      content: raw?.content !== undefined && raw?.content !== null ? String(raw.content) : undefined,
      education_article_ids: normalizeNumberArray(raw?.education_article_ids),
      satisfaction:
        raw?.satisfaction !== undefined && raw?.satisfaction !== null ? Number(raw.satisfaction) : undefined,
      remark: raw?.remark !== undefined && raw?.remark !== null ? String(raw.remark) : undefined
    }
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object') {
        return normalizeFollowUpTaskResult(parsed)
      }
    } catch (error) {
      console.warn('解析随访任务结果失败:', error)
    }
  }

  return undefined
}

/**
 * 规范化随访任务。
 */
function normalizeFollowUpTask(data: any): FollowUpTask {
  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    task_type: data?.task_type !== undefined && data?.task_type !== null ? Number(data.task_type) : undefined,
    source_type:
      data?.source_type !== undefined && data?.source_type !== null ? Number(data.source_type) : undefined,
    source_id: data?.source_id !== undefined && data?.source_id !== null ? Number(data.source_id) : null,
    plan_follow_time: data?.plan_follow_time ? String(data.plan_follow_time) : undefined,
    staff_id: data?.staff_id !== undefined && data?.staff_id !== null ? Number(data.staff_id) : undefined,
    contact_method:
      data?.contact_method !== undefined && data?.contact_method !== null ? Number(data.contact_method) : undefined,
    status: data?.status !== undefined && data?.status !== null ? Number(data.status) : undefined,
    result: normalizeFollowUpTaskResult(data?.result),
    completed_at: data?.completed_at ? String(data.completed_at) : undefined,
    remark: data?.remark !== undefined && data?.remark !== null ? String(data.remark) : undefined,
    created_at: data?.created_at ? String(data.created_at) : undefined
  }
}

/**
 * 获取我的随访任务列表。需要登录。
 */
export function getUserFollowUps(params?: PaginationParams) {
  return get<PaginationResponse<FollowUpTask>>('/api/v1/user/follow-ups', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeFollowUpTask)
    }
  })
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

// ============ 生命体征记录 ============

/**
 * 规范化体征扩展字段。
 * 后端 extra 可能以 JSON 字符串返回，统一规范化为对象。
 */
function normalizeExtra(value: unknown): Record<string, unknown> | undefined {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }

  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return parsed as Record<string, unknown>
      }
    } catch (error) {
      console.warn('解析体征扩展字段失败:', error)
    }
  }

  return undefined
}

/**
 * 规范化生命体征记录。
 */
function normalizeHealthMonitoring(data: any): HealthMonitoring {
  return {
    ...data,
    id: Number(data?.id || 0),
    user_id: Number(data?.user_id || 0),
    record_type:
      data?.record_type !== undefined && data?.record_type !== null ? Number(data.record_type) : undefined,
    value: data?.value !== undefined && data?.value !== null ? Number(data.value) : undefined,
    unit: data?.unit !== undefined && data?.unit !== null ? String(data.unit) : undefined,
    extra: normalizeExtra(data?.extra),
    recorded_by:
      data?.recorded_by !== undefined && data?.recorded_by !== null ? Number(data.recorded_by) : undefined,
    recorded_at: data?.recorded_at ? String(data.recorded_at) : undefined,
    remark: data?.remark !== undefined && data?.remark !== null ? String(data.remark) : undefined,
    created_at: data?.created_at ? String(data.created_at) : undefined
  }
}

/**
 * 获取我的生命体征记录列表。需要登录。
 */
export function getUserMonitoring(params?: PaginationParams) {
  return get<PaginationResponse<HealthMonitoring>>('/api/v1/user/monitoring', params).then(res => {
    const list = Array.isArray(res?.list) ? res.list : []
    return {
      ...res,
      list: list.map(normalizeHealthMonitoring)
    }
  })
}

/**
 * 新增生命体征记录。需要登录。
 */
export function createUserMonitoring(data: {
  record_type: number
  value: number
  unit?: string
  extra?: Record<string, unknown>
  recorded_at?: string
  remark?: string
}) {
  return post<HealthMonitoring>('/api/v1/user/monitoring', data).then(normalizeHealthMonitoring)
}
