export interface MerchantLoginRequest {
  username: string
  password: string
}

export interface MerchantStaffInfo {
  id: number
  merchant_id: number
  username: string
  name?: string
  phone?: string
  avatar?: string
  status?: number
}

export interface MerchantLoginResponse {
  token: string
  merchant_id: number
  staff: MerchantStaffInfo
}

export interface DashboardData {
  total_orders: number
  total_amount: number
  today_orders: number
  today_amount: number
  pending_orders: number
  completed_orders: number
  refunded_amount: number
}

export interface MerchantListItem {
  id: number
  name: string
  logo?: string
  cover_image?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  business_category?: string
  business_hours?: string
  announcement?: string
  sub_mch_id?: string
  payment_config_status?: number
  status: number
  created_at?: string
}

export interface MerchantDetail extends MerchantListItem {
  qrcode_url?: string
  admin_staff_id?: number
  admin_username?: string
  admin_name?: string
  admin_phone?: string
  admin_status?: number
}

export interface MerchantCategory {
  id: number
  merchant_id: number
  name: string
  sort: number
  status: number
  product_count?: number
  created_at?: string
  updated_at?: string
}

export interface MerchantCategoryPayload {
  name: string
  sort?: number
  status?: number
}

export interface MerchantCategorySortItem {
  id: number
  sort: number
}

export interface MerchantProductQuery {
  page?: number
  page_size?: number
  category_id?: number
  status?: number | string
  sale_type?: number | string
  product_type?: number | string
  keyword?: string
}

export interface MerchantProductSpecOption {
  id?: number
  name: string
  price: number
  stock?: number
}

export interface MerchantProductSpec {
  id?: number
  name: string
  options: MerchantProductSpecOption[]
}

export interface MerchantProduct {
  id: number
  merchant_id: number
  category_id: number
  name: string
  description: string
  images: string[]
  price: number
  original_price: number
  stock: number
  unit: string
  product_type: number
  service_content?: WellnessPackageContent | Record<string, unknown> | null
  sale_type: number
  rental_unit: number
  rental_price: number
  deposit: number
  max_rental_duration: number
  sales: number
  sort: number
  status: number
  category_name?: string
  specs: MerchantProductSpec[]
  created_at?: string
  updated_at?: string
}

// 康养套餐内容：服务项列表
export interface WellnessPackageItem {
  id?: string
  name: string // 服务项名称，如：血压测量
  description?: string // 服务项描述
  count?: number // 服务次数
  unit?: string // 计量单位，如：次、小时
}

export interface WellnessPackageContent {
  duration?: string // 套餐服务周期，如：一个月
  items?: WellnessPackageItem[] // 服务项列表
  notes?: string // 套餐备注
  applicable_groups?: string // 适用人群
}

export interface MerchantProductListResponse {
  list: MerchantProduct[]
  pagination?: {
    total: number
    page: number
    page_size: number
  }
}

export interface MerchantProductUpsertPayload {
  category_id?: number
  name: string
  description?: string
  images: string[]
  price: number
  original_price?: number
  stock?: number
  unit?: string
  product_type?: number
  service_content?: WellnessPackageContent | Record<string, unknown> | null
  sale_type?: number
  rental_unit?: number
  rental_price?: number
  deposit?: number
  max_rental_duration?: number
  sort?: number
  sales?: number
  specs?: MerchantProductSpec[]
}

export interface MerchantProductEditableSpec {
  id?: number
  name: string
  values: string[]
}

export interface MerchantProductSpecsResponse {
  specs: MerchantProductEditableSpec[]
  skus: unknown[]
}

export interface MerchantProductSpecsPayload {
  specs: MerchantProductEditableSpec[]
  skus: unknown[]
}

export interface UpdateSpMerchantFormData {
  name?: string
  logo?: string
  cover_image?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  business_category?: string
  business_hours?: string
  announcement?: string
  status?: number
}

export interface MerchantPaymentConfigFormData {
  sub_mch_id: string
}

export const SpOrderStatusText: Record<number, string> = {
  1: '待支付',
  2: '已支付',
  3: '已完成',
  4: '已取消',
  5: '退款中',
  6: '已退款'
}

export interface SpOrderUser {
  id: number
  nickname?: string
  avatar?: string
  phone?: string
}

export interface SpOrderMerchant {
  id: number
  name: string
  logo?: string
  address?: string
  phone?: string
  contact_phone?: string
}

export interface SpOrderItem {
  id?: number
  product_id: number
  product_name: string
  image: string
  price: number
  quantity: number
  specs?: string
  spec_info?: string
  subtotal?: number
  sale_type?: number
  rental_unit?: number
  rental_duration?: number
  unit_rental_price?: number
  rental_subtotal?: number
  deposit?: number
  deposit_deduct?: number
}

export interface SpOrderDeliveryInfo {
  type?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  distance?: number
}

export interface SpOrder {
  id: number
  order_no: string
  user?: SpOrderUser
  merchant?: SpOrderMerchant
  items: SpOrderItem[]
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
  delivery_info?: SpOrderDeliveryInfo
  delivery_address?: string
  contact_name?: string
  contact_phone?: string
  status: number
  remark?: string
  transaction_id?: string
  created_at: string
  paid_at?: string
  completed_at?: string
  completed_by_name?: string
  cancelled_at?: string
  refunded_at?: string
  // 工单字段
  order_type?: number
  biz_status?: number
  assigned_staff_id?: number
  scheduled_at?: string
  actual_started_at?: string
  actual_ended_at?: string
}

export interface SpOrderListResponse {
  list: SpOrder[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

export interface OrderAnalyticsBucket {
  label: string
  order_count: number
}

export interface OrderAnalyticsData {
  day: OrderAnalyticsBucket[]
  week: OrderAnalyticsBucket[]
  month: OrderAnalyticsBucket[]
  year: OrderAnalyticsBucket[]
}

export interface AmountTrendData {
  trends: Array<{
    date: string
    amount: number
  }>
}

export interface UploadTokenResponse {
  token: string
  domain: string
  prefix: string
  upload_url?: string
}

// 订单类型文案（order_type）
export const OrderTypeText: Record<number, string> = {
  1: '零售',
  2: '租赁',
  3: '康养上门',
  4: '陪诊服务',
  5: '科普体验',
  6: '长护险服务'
}

// 工单业务状态文案（biz_status）
export const BizStatusText: Record<number, string> = {
  0: '无',
  1: '待接单',
  2: '待出发',
  3: '服务中',
  4: '待支付尾款',
  5: '已完成',
  6: '已取消'
}

// 服务人员（PC 后台管理）
export interface ServiceStaffItem {
  id: number
  merchant_id: number
  username: string
  name: string
  phone: string
  openid?: string
  avatar?: string
  status: number // 0=待审核 1=启用 2=禁用
  last_login_at?: string
  created_at?: string
  updated_at?: string
}

export interface ServiceStaffListResponse {
  list: ServiceStaffItem[]
  total: number
  pagination: {
    page: number
    page_size: number
  }
}
