/**
 * 服务人员端全局类型定义（第0期占位，第2期按真实后端扩展）
 */

// 订单类型（与 orders.order_type 对齐）
export enum OrderType {
  /** 零售 */ Retail = 1,
  /** 租赁 */ Rental = 2,
  /** 康养上门 */ KangyangVisit = 3,
  /** 陪诊 */ Peizhen = 4,
  /** 科普体验 */ ScienceActivity = 5,
  /** 长护险服务 */ LongTermCare = 6
}

export const OrderTypeText: Record<OrderType, string> = {
  [OrderType.Retail]: '零售',
  [OrderType.Rental]: '租赁归还',
  [OrderType.KangyangVisit]: '康养上门',
  [OrderType.Peizhen]: '陪诊服务',
  [OrderType.ScienceActivity]: '科普体验',
  [OrderType.LongTermCare]: '长护险服务'
}

// 工单业务状态（biz_status，与后端 orders.biz_status 对齐）
export enum WorkorderBizStatus {
  /** 无（零售等无需派工的订单） */ None = 0,
  /** 待接单 */ Pending = 1,
  /** 已接单-待出发 */ Accepted = 2,
  /** 服务中-已签到 */ InService = 3,
  /** 待支付尾款 */ PendingPayment = 4,
  /** 已完成 */ Completed = 5,
  /** 已取消 */ Cancelled = 6
}

export interface StaffUser {
  id: number
  name: string
  phone: string
  avatar?: string
  role?: string
  status?: number
  openid?: string
  created_at?: string
}

export interface WorkorderItem {
  id: number
  order_id: number
  order_no: string
  order_type: OrderType
  biz_status: WorkorderBizStatus
  /** 预约时间 */
  scheduled_at?: string
  /** 实际签到/签退 */
  actual_started_at?: string
  actual_ended_at?: string
  /** 客户信息 */
  contact_name?: string
  contact_phone?: string
  delivery_address?: string
  lat?: number
  lng?: number
  /** 服务内容快照 */
  service_content?: any
  remark?: string
}

export interface ScheduleDay {
  date: string // YYYY-MM-DD
  type: 'work' | 'rest' | 'duty' | 'leave'
  shift?: 'morning' | 'afternoon' | 'night' | 'full'
  note?: string
}

// ============================================
// 居民健康档案 / 健康评估（阶段一：基层健康服务闭环）
// 注意：后端 JSON 列（past_history 等）可能返回字符串，取值时需做数组规范化
// ============================================

// 居民健康档案
export interface HealthRecord {
  id: number
  user_id: number
  real_name?: string
  gender?: number        // 1男2女
  birth_date?: string
  id_card?: string
  phone?: string
  emergency_contact?: string
  emergency_phone?: string
  address?: string
  height_cm?: number | null
  weight_kg?: number | null
  blood_type?: string
  past_history?: string[]
  allergy_history?: string[]
  family_history?: string[]
  surgery_history?: string[]
  medication_list?: string[]
  chronic_tags?: string[]   // 慢病标签
  smoking?: string
  drinking?: string
  assessment_level?: string
  remark?: string
  status?: number
  created_at?: string
  updated_at?: string
}

// 评估量表-单题
export interface AssessmentFormQuestion {
  key: string
  title: string
  options: { label: string; score: number }[]
}

// 评估量表
export interface AssessmentForm {
  id: number
  name: string
  dimension?: string
  description?: string
  questions: AssessmentFormQuestion[]
  score_rule?: { min: number; max: number; level: string; conclusion?: string }[]
  version?: number
  status?: number
}

// 健康评估记录
export interface HealthAssessment {
  id: number
  user_id: number
  form_id: number
  form_name?: string
  assessor_type?: number    // 1自助 2服务人员
  answers?: Record<string, string>
  total_score?: number
  level?: string
  conclusion?: string
  suggestions?: string[]
  symptom_desc?: string
  created_at?: string
}

// ============================================
// 康复辅具适配建议（阶段二：基层健康服务闭环）
// 注意：后端 recommended_products 可能返回 JSON 字符串，取值时需做数组规范化
// ============================================

// 适配建议-推荐商品
export interface FittingRecommendedProduct {
  product_id: number
  name: string
  reason?: string
  sale_type?: number // 1一口价 2租赁
}

// 适配建议
export interface FittingRecommendation {
  id: number
  user_id: number
  assessment_id?: number | null
  symptom_desc?: string
  fitting_result?: string
  recommended_products: FittingRecommendedProduct[]
  staff_id?: number | null
  status?: number // 0草稿 1已确认 2已下单
  order_id?: number | null
  created_at?: string
  updated_at?: string
}

// 商城商品（公开商品列表接口，服务端可访问）
export interface StoreProduct {
  id: number
  name: string
  price?: number
  rental_price?: number
  sale_type?: number // 1一口价 2租赁
  category_id?: number
  cover_url?: string
  unit?: string
  stock?: number
}

// ============================================
// 我的照护计划 / 上门照护记录（阶段三：基层健康服务闭环）
// 注意：后端 items/visits/nursing_items/vitals/photos 可能返回 JSON 字符串，取值时需做规范化
// ============================================

// 照护计划-护理项
export interface CarePlanItem {
  name: string
  desc?: string
}

// 照护计划
export interface CarePlan {
  id: number
  user_id: number
  name: string
  plan_type?: number // 1生活照料 2基础护理 3康复训练 4综合康养
  start_date?: string
  end_date?: string
  frequency?: string
  goals?: string
  items: CarePlanItem[]
  assigned_staff_id?: number | null
  order_id?: number | null
  status?: number // 0草稿 1执行中 2已暂停 3已完成
  visit_count?: number
  created_at?: string
  updated_at?: string
  visits?: CareVisit[]
}

// 照护记录-护理项完成情况
export interface CareVisitNursingItem {
  name: string
  done?: boolean
  remark?: string
}

// 照护记录-生命体征
export interface CareVisitVitals {
  blood_pressure?: string
  blood_glucose?: string
  heart_rate?: string
  oxygen?: string
  weight?: string
}

// 上门照护记录
export interface CareVisit {
  id: number
  plan_id?: number | null
  order_id?: number | null
  user_id: number
  staff_id?: number
  visit_at?: string
  nursing_items: CareVisitNursingItem[]
  vitals?: CareVisitVitals
  photos?: string[]
  remark?: string
  follow_up_advice?: string
  created_at?: string
}

// ============================================
// 随访任务 / 生命体征监测 / 健康宣教（阶段四：基层健康服务闭环）
// 注意：后端 result/tags/extra 可能返回 JSON 字符串，取值时需做规范化
// ============================================

// 随访结果内容（result JSON 字段）
export interface FollowUpTaskResult {
  contact_method?: number // 1电话 2上门 3微信
  content?: string
  education_article_ids?: number[]
  satisfaction?: number // 满意度 1-5
  remark?: string
}

// 随访任务
export interface FollowUpTask {
  id: number
  user_id: number
  task_type?: number // 1康复随访 2租后回访 3慢病随访 4评估回访
  source_type?: number // 1服务完成 2租赁归还 3评估完成 4手动
  source_id?: number | null
  plan_follow_time?: string
  staff_id?: number | null
  contact_method?: number // 1电话 2上门 3微信
  status?: number // 0待执行 1已完成 2已跳过
  result?: FollowUpTaskResult
  completed_at?: string
  remark?: string
  created_at?: string
  user?: { id: number; nickname?: string; phone?: string }
}

// 健康宣教文章（仅已发布，供随访选用）
export interface EducationArticle {
  id: number
  title: string
  category?: string
  cover?: string
  content?: string
  tags?: string[]
  status?: number
  publish_at?: string
}

// 生命体征监测记录
export interface HealthMonitoring {
  id: number
  user_id: number
  record_type?: number // 1血压 2血糖 3心率 4血氧 5体重
  value?: number
  unit?: string
  extra?: Record<string, unknown>
  recorded_by?: number
  recorded_at?: string
  remark?: string
  created_at?: string
}
