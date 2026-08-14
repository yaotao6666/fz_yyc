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
  merchant_id?: number | string
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
  merchant_id?: number
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
