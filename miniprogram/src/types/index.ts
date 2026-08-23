/**
 * API 类型定义 - 基于PRD文档
 */

// 通用分页参数
export interface PaginationParams {
  page?: number
  page_size?: number
}

// 分页响应
export interface PaginationResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// API统一响应格式
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// ============ 认证相关 ============

// C端用户登录请求
export interface WechatLoginRequest {
  code: string
  nickname?: string
  avatar?: string
}

// C端用户登录响应
export interface WechatLoginResponse {
  token: string
  app_mode: string
  app_id: string
  user: UserInfo
}

// ============ 配送设置 ============

export interface DeliverySettings {
  enabled: boolean
  base_fee: number
  free_delivery_amount: number
  distance_rules: DistanceRule[]
  max_distance: number
}

export interface StoreDeliveryRules extends DeliverySettings {}

// 配送距离规则
export interface DistanceRule {
  min_distance: number
  max_distance: number
  fee: number
}

// ============ 商品相关 ============

export type CompatibleAmountValue = number | string
export type OptionalCompatibleAmountValue = CompatibleAmountValue | null | undefined

export interface ProductApiSpecOption {
  id?: number | string
  name?: string
  price?: OptionalCompatibleAmountValue
  stock?: number | string | null
}

export interface ProductApiSpec {
  id?: number | string
  name?: string
  options?: ProductApiSpecOption[] | null
}

// 商品类型: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普资讯
export type ProductType = 1 | 2 | 3 | 4 | 5

// 康养套餐服务内容配置
export interface WellnessPackageItem {
  name: string
  description?: string
  count?: number
  unit?: string
}
export interface WellnessPackageContent {
  cycle?: string
  target_audience?: string
  services: WellnessPackageItem[]
  remark?: string
}

export interface ProductApiResponse {
  id?: number | string
  name?: string
  images?: string[] | string | null
  price?: OptionalCompatibleAmountValue
  original_price?: OptionalCompatibleAmountValue
  stock?: number | string | null
  sales?: number | string | null
  category_id?: number | string | null
  category_name?: string
  status?: number | string
  sort?: number | string | null
  unit?: string
  description?: string
  product_type?: number | string | null
  service_content?: any
  sale_type?: number | string | null
  rental_unit?: number | string | null
  rental_price?: OptionalCompatibleAmountValue
  deposit?: OptionalCompatibleAmountValue
  max_rental_duration?: number | string | null
  specs?: ProductApiSpec[] | null
  created_at?: string
  updated_at?: string
}

// 商品信息
export interface Product {
  id: number
  name: string
  images: string[]
  price: number
  original_price?: number
  stock: number
  sales?: number
  category_id: number
  category_name?: string
  status: number
  sort?: number
  unit?: string
  description?: string
  product_type: ProductType
  service_content?: WellnessPackageContent | any
  sale_type: number
  rental_unit: number
  rental_price: number
  deposit: number
  max_rental_duration: number
  specs?: ProductSpec[]
  created_at: string
  updated_at?: string
}

// 商品规格
export interface ProductSpec {
  id?: number
  name: string
  options: SpecOption[]
}

// 规格选项
export interface SpecOption {
  id?: number
  name: string
  price: number
  stock?: number
}

// 商品列表响应
export interface ProductListResponse {
  list: Product[]
  pagination?: {
    total: number
    page: number
    page_size: number
  }
}

// ============ 订单相关 ============

// 订单状态
export enum OrderStatus {
  PENDING_PAYMENT = 1,  // 待支付
  PAID = 2,              // 已支付
  COMPLETED = 3,         // 已完成
  CANCELLED = 4,         // 已取消
  REFUNDING = 5,         // 退款中
  REFUNDED = 6           // 已退款
}

// 订单状态文本
export const OrderStatusText: Record<number, string> = {
  [OrderStatus.PENDING_PAYMENT]: '待支付',
  [OrderStatus.PAID]: '已支付',
  [OrderStatus.COMPLETED]: '已完成',
  [OrderStatus.CANCELLED]: '已取消',
  [OrderStatus.REFUNDING]: '退款中',
  [OrderStatus.REFUNDED]: '已退款'
}

// 订单用户信息
export interface OrderUser {
  id: number
  nickname?: string
  avatar?: string
  phone?: string
}

// 订单商家信息
export interface OrderMerchant {
  id: number
  name: string
  logo?: string
  address?: string
  phone?: string
  contact_phone?: string
}

// 订单商品项
export interface OrderItem {
  product_id: number
  product_name: string
  image: string
  price: number
  quantity: number
  specs?: string
  subtotal?: number
  sale_type?: number
  rental_unit?: number
  rental_duration?: number
  unit_rental_price?: number
  rental_subtotal?: number
  deposit?: number
  deposit_deduct?: number
}

// 订单信息
export interface Order {
  id: number
  order_no: string
  user?: OrderUser
  merchant?: OrderMerchant
  items: OrderItem[]
  total_amount: number
  delivery_fee: number
  discount_amount: number
  pay_amount: number
  total_deposit?: number
  deposit_status?: number
  deposit_refund_amount?: number
  deposit_deduct_amount?: number
  deposit_refunded_at?: string
  rental_returned_at?: string
  rental_return_remark?: string
  delivery_address?: string
  contact_name?: string
  contact_phone?: string
  delivery_info?: DeliveryInfo
  status: number
  status_text?: string
  remark?: string
  transaction_id?: string
  created_at: string
  paid_at?: string
  completed_at?: string
  completed_by_name?: string
  cancelled_at?: string
  refunded_at?: string
}

// 配送信息
export interface DeliveryInfo {
  type: string
  address?: string
  contact_name?: string
  contact_phone?: string
  distance?: number
}

// 订单列表响应
export interface OrderListResponse {
  list: Order[]
  total: number
  page: number
  page_size: number
}

// 创建订单请求
export interface CreateOrderRequest {
  items: {
    product_id: number
    spec_info?: string
    quantity: number
    rental_duration?: number
  }[]
  delivery_distance?: number
  delivery_address?: string
  contact_name?: string
  contact_phone?: string
  remark?: string
}

// 创建订单响应
export interface CreateOrderResponse {
  order: Order
  pay_params?: WechatPayParams
  pay_hint?: string
}

// 微信支付参数
export interface WechatPayParams {
  appId?: string
  timeStamp: string
  nonceStr: string
  package: string
  signType: string
  paySign: string
  prepay_id?: string
}

// ============ C端用户相关 ============

// C端用户
export interface UserInfo {
  id: number
  openid?: string
  union_id?: string
  nickname: string
  avatar: string
  phone?: string
  status: number
}

export interface UserAddress {
  id?: number
  user_id?: number
  name: string
  phone: string
  province?: string
  city?: string
  district?: string
  address: string
  lat?: number
  lng?: number
  is_default?: boolean
}

// ============ 店铺相关 ============

// 店铺首页信息
export interface StoreHomeInfo {
  merchant: {
    id: number
    name: string
    logo: string
    cover_image: string
    address: string
    contact_phone?: string
    business_hours: string
    announcement: string
    status: number
    rating: number
    sales_count: number
  }
  categories: {
    id: number
    name: string
    parent_id?: number
    level?: number
    sort: number
    product_count: number
  }[]
  hot_products: {
    id: number
    name: string
    images: string[]
    price: number
    original_price?: number
    sales: number
    product_type?: ProductType
    sale_type?: number
    rental_unit?: number
    rental_price?: number
    deposit?: number
  }[]
}

// 店铺商品列表（按分类分组）
export interface StoreProductGroup {
  category: {
    id: number
    name: string
  }
  products: Product[]
}

// ============ 通用 ============

export interface UploadTokenResponse {
  token: string
  domain: string
  prefix: string
  upload_url: string
}

export interface MerchantBehaviorEventRequest {
  openid?: string
  event_type: 'page_view' | 'product_view' | 'submit_order' | 'pay_success'
  page?: string
  product_id?: number
  order_id?: number
  source?: string
  payload?: Record<string, any>
}

// ============ 基层健康服务 ============

// 健康档案（字段与后端一致）
export interface HealthRecord {
  id: number
  user_id: number
  real_name?: string
  gender?: number        // 1男 2女
  birth_date?: string    // YYYY-MM-DD
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
  chronic_tags?: string[]
  smoking?: string
  drinking?: string
  assessment_level?: string
  remark?: string
  status?: number
  created_at?: string
  updated_at?: string
}

// 评估量表题目
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
  assessor_type?: number
  answers?: Record<string, string>
  total_score?: number
  level?: string
  conclusion?: string
  suggestions?: string[]
  symptom_desc?: string
  created_at?: string
}

// ============ 适配建议 ============

/** 适配建议-推荐商品 */
export interface FittingRecommendedProduct {
  product_id: number
  name: string
  reason?: string
  sale_type?: number // 1一口价 2租赁
}

/** 适配建议 */
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

// ============ 我的照护计划 ============

/** 照护计划-护理项 */
export interface CarePlanItem {
  name: string
  desc?: string
}

/** 照护计划（字段与后端一致） */
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

/** 照护计划-上门照护记录 */
export interface CareVisit {
  id: number
  plan_id?: number | null
  order_id?: number | null
  user_id: number
  staff_id?: number
  visit_at?: string
  nursing_items: { name: string; done?: boolean; remark?: string }[]
  vitals?: {
    blood_pressure?: string
    blood_glucose?: string
    heart_rate?: string
    oxygen?: string
    weight?: string
  }
  photos?: string[]
  remark?: string
  follow_up_advice?: string
  created_at?: string
}

// ============ 我的康复随访 ============

/** 随访任务结果（可能以 JSON 字符串返回，接口封装时规范化） */
export interface FollowUpTaskResult {
  contact_method?: number // 实际随访方式 1电话 2上门 3微信
  content?: string
  education_article_ids?: number[]
  satisfaction?: number
  remark?: string
}

/** 随访任务（字段与后端一致） */
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
}

// ============ 健康宣教 ============

/** 宣教文章（字段与后端一致） */
export interface EducationArticle {
  id: number
  title: string
  category?: string
  cover?: string
  content?: string
  tags?: string[]
  status?: number
  publish_at?: string
  views?: number
}

// ============ 生命体征记录 ============

/** 生命体征记录（字段与后端一致） */
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

// ============ 错误码 ============

export const ErrorCode = {
  SUCCESS: 0,
  PARAM_ERROR: 1001,
  UNAUTHORIZED: 1002,
  FORBIDDEN: 1003,
  NOT_FOUND: 1004,
  MERCHANT_NOT_EXIST: 3001,
  MERCHANT_NOT_APPROVED: 3002,
  MERCHANT_DISABLED: 3003,
  PRODUCT_NOT_EXIST: 4001,
  PRODUCT_OFF_SALE: 4002,
  STOCK_INSUFFICIENT: 4003,
  ORDER_NOT_EXIST: 5001,
  ORDER_STATUS_ERROR: 5002,
  ORDER_PAID: 5003,
  ORDER_CANCELLED: 5004,
  PAYMENT_FAILED: 6001,
  REFUND_FAILED: 6002,
  SERVER_ERROR: 9001
} as const
