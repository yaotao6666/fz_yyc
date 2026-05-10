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

export interface Announcement {
  id: number
  service_provider_id: number
  title: string
  content: string
  status: number
  created_at: string
  updated_at: string
}

export interface AnnouncementListResponse {
  list: Announcement[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

// ============ 认证相关 ============

// 商家登录请求
export interface MerchantLoginRequest {
  username: string
  password: string
}

// 商家登录响应
export interface MerchantLoginResponse {
  token: string
  merchant_id: number
  staff: MerchantStaff
}

export interface ServiceProviderLoginRequest {
  username: string
  password: string
}

export interface ServiceProviderLoginResponse {
  token: string
  service_provider: {
    id: number
    name: string
  }
}

// C端用户登录请求
export interface WechatLoginRequest {
  code: string
  nickname?: string
  avatar?: string
}

// C端用户登录响应
export interface WechatLoginResponse {
  token: string
  user: UserInfo
}

// ============ 服务商相关 ============

// 服务商管理员
export interface ServiceProviderAdmin {
  id: number
  service_provider_id: number
  username: string
  name: string
  phone: string
  role: string
  status: number
}

// ============ 商家相关 ============

// 商家员工
export interface MerchantStaff {
  id: number
  merchant_id: number
  username: string
  name: string
  phone: string
  openid?: string
  unionid?: string
  wechat_bound_at?: string
  role: string
  notify_enabled?: boolean
  browse_notify_enabled?: boolean
  status: number
  last_login_at?: string
  last_wechat_login_at?: string
}

// 商家信息
export interface MerchantInfo {
  id: number
  name: string
  logo: string
  contact_name: string
  contact_phone: string
  address: string
  business_category: string
  status: number
  license?: MerchantLicense
  settings?: MerchantSettings
  wechat_pay?: WechatPayInfo
  qrcode_url?: string
  created_at: string
}

// 商家营业执照
export interface MerchantLicense {
  license_no: string
  license_name: string
  license_image: string
  legal_person: string
  valid_from: string
  valid_to: string
}

// 商家设置
export interface MerchantSettings {
  announcement: string
  business_hours: string
  min_order_amount: number
  takeout_enabled: boolean
  dine_in_enabled: boolean
  notify_enabled?: boolean
  browse_notify_enabled?: boolean
  wechat_bound?: boolean
  unionid?: string
  wechat_bound_at?: string
  delivery_settings?: DeliverySettings
}

// 配送设置
export interface DeliverySettings {
  enabled: boolean
  base_fee: number
  free_delivery_amount: number
  distance_rules: DistanceRule[]
  max_distance: number
}

// 配送距离规则
export interface DistanceRule {
  min_distance: number
  max_distance: number
  fee: number
}

// 微信支付信息
export interface WechatPayInfo {
  sub_mch_id: string
  status: string
  applyment_status: number
}

// 商家入驻请求
export interface MerchantRegisterRequest {
  name: string
  contact_name: string
  contact_phone: string
  contact_email?: string
  address: string
  business_category: string
  invite_code?: string
  license?: {
    license_no: string
    license_name: string
    license_image: string
    legal_person: string
    legal_person_id: string
    legal_person_id_front: string
    legal_person_id_back: string
    valid_from: string
    valid_to: string
  }
  bank_account?: {
    bank_name: string
    bank_branch: string
    account_no: string
    account_name: string
    account_type: number
  }
  store_info?: {
    store_name: string
    store_images: string[]
  }
}

// 商家入驻响应
export interface MerchantRegisterResponse {
  merchant_id: number
  application_id: number
  status: string
  invited_by?: {
    name: string
    reward_eligible: boolean
  }
}

// 商家进件状态
export interface MerchantApplicationStatus {
  application_id: number
  status: string
  applyment_id?: string
  sub_mch_id?: string
  submit_time?: string
  audit_time?: string
  audit_detail?: {
    legal_person_validation?: string
    license_validation?: string
    bank_account_validation?: string
  }
}

// ============ 商品分类相关 ============

// 商品分类
export interface Category {
  id: number
  name: string
  sort: number
  product_count?: number
  status: number
  created_at: string
}

// ============ 商品相关 ============

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
  specs?: ProductSpec[]
  created_at: string
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
  total: number
  page: number
  page_size: number
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

// 配送类型
export enum DeliveryType {
  DELIVERY = 1,  // 配送
  DINE_IN = 2,   // 堂食
  PICKUP = 3     // 自提
}

// 配送类型文本
export const DeliveryTypeText: Record<number, string> = {
  [DeliveryType.DELIVERY]: '配送',
  [DeliveryType.DINE_IN]: '堂食',
  [DeliveryType.PICKUP]: '自提'
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
  delivery_type?: number
  delivery_distance?: number
  delivery_info?: DeliveryInfo
  status: number
  status_text?: string
  remark?: string
  verify_code?: string
  transaction_id?: string
  created_at: string
  paid_at?: string
  completed_at?: string
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

// 订单统计
export interface OrderStatistics {
  total_orders: number
  total_sales: number
  pending_payment: number
  pending_complete: number
  completed: number
  refunded: number
  cancelled: number
}

// 创建订单请求
export interface CreateOrderRequest {
  items: {
    product_id: number
    spec_info?: string
    quantity: number
  }[]
  delivery_type: number
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
}

// 微信支付参数
export interface WechatPayParams {
  timeStamp: string
  nonceStr: string
  package: string
  signType: string
  paySign: string
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

// ============ 店铺相关 ============

// 店铺首页信息
export interface StoreHomeInfo {
  merchant: {
    id: number
    name: string
    logo: string
    images: string[]
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

// ============ 数据分析相关 ============

// 销售概览
export interface SalesOverview {
  total_sales: number
  total_orders: number
  total_customers: number
  avg_order_amount: number
  sales_growth: number
  orders_growth: number
  customers_growth: number
  visit_count?: number
  visit_users?: number
  pay_success_users?: number
}

// 销售趋势
export interface SalesTrend {
  date: string
  sales: number
  orders: number
  customers?: number
  visit_users?: number
  submit_order_users?: number
}

// 商品排行
export interface ProductRanking {
  product_id: number
  product_name: string
  image: string
  sales_count: number
  sales_amount: number
}

// 时段分析
export interface HourlyAnalysis {
  hour: number
  orders: number
  sales: number
}

// 库存预警
export interface StockAlert {
  id?: number
  product_id: number
  product_name: string
  stock: number
  status?: string | number
}

export interface CustomerAnalysis {
  total_customers: number
  new_customers: number
  repeat_rate: number
  visit_users?: number
  visit_count?: number
  submit_order_users?: number
  pay_success_users?: number
}

export interface CustomerTrend {
  date: string
  total_users: number
  new_users: number
  order_count: number
}

export interface MerchantStaffListResponse {
  list: MerchantStaff[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface CreateMerchantStaffRequest {
  name: string
  phone: string
  username: string
  password: string
  role?: string
}

export interface UpdateMerchantStaffRequest {
  name?: string
  phone?: string
  role?: string
  notify_enabled?: boolean
  browse_notify_enabled?: boolean
  status?: number
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UploadTokenResponse {
  token: string
  domain: string
  prefix: string
  upload_url: string
}

export interface MerchantWechatLoginRequest {
  code: string
}

export interface MerchantBehaviorEventRequest {
  openid: string
  event_type: 'page_view' | 'product_view' | 'submit_order' | 'pay_success'
  page?: string
  product_id?: number
  order_id?: number
  source?: string
  payload?: Record<string, any>
}

// ============ 邀请入驻相关 ============

// 邀请信息
export interface InviteInfo {
  invite_code: string
  qrcode_url?: string
  share_link?: string
  poster_url?: string
  expire_time?: string
}

// 我的邀请信息
export interface MyInviteInfo {
  invite_code: string
  total_invites: number
  completed_invites: number
  pending_invites: number
  rewards: {
    free_year_count: number
    lowest_rate_qualified: boolean
  }
}

// 邀请记录
export interface InviteRecord {
  id: number
  invitee_name: string
  invitee_phone: string
  invite_code: string
  status: number
  reward_type?: string
  created_at: string
  completed_at?: string
}

// 邀请记录列表响应
export interface InviteRecordListResponse {
  list: InviteRecord[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
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
