export interface MerchantLoginRequest {
  username: string
  password: string
}

export interface MerchantStaffInfo {
  id: number
  username: string
  name?: string
  phone?: string
  openid?: string
  unionid?: string
  role?: string
  department_id?: number | null
  status?: number
  notify_enabled?: boolean
  browse_notify_enabled?: boolean
  created_at?: string
  last_login_at?: string
  department?: SysDepartment
  roles?: SysRole[]
}

export interface MerchantLoginResponse {
  token: string
  merchant_id: number
  staff: MerchantStaffInfo
  menus: SysMenu[]
  permissions: string[]
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
  name: string
  parent_id?: number
  level?: number
  sort: number
  status: number
  product_count?: number
  created_at?: string
  updated_at?: string
  children?: MerchantCategory[]
}

export interface MerchantCategoryPayload {
  name: string
  parent_id?: number
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
  product_types?: string
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
  cycle?: string // 套餐服务周期，如：一个月
  target_audience?: string // 适用人群
  services?: WellnessPackageItem[] // 服务项列表
  remark?: string // 套餐备注
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
  unit: string
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
  deposit_refund_amount?: number
  deposit_refunded_at?: string
  rental_returned_at?: string
  rental_return_remark?: string
  delivery_info?: SpOrderDeliveryInfo
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
  assigned_staff?: { id: number; name?: string; phone?: string }
  scheduled_at?: string
  actual_started_at?: string
  actual_ended_at?: string
  rental_end_at?: string
  parent_order_id?: number
  renew_flag?: number
  // 服务对象（健康档案）
  record_id?: number
  record_name?: string
  record_gender?: number
  record_birth_date?: string
  delivery_district?: string
  // 分账字段
  profit_sharing_status?: number
  profit_sharing_amount?: number
  profit_sharing_order_no?: string
  profit_sharing_at?: string
  profit_sharing_error?: string
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
  4: '陪诊服务'
}

// 工单业务状态文案（biz_status）
export const BizStatusText: Record<number, string> = {
  0: '无',
  1: '待接单',
  2: '待出发',
  3: '服务中',
  5: '已完成',
  6: '已取消'
}

// 服务人员（PC 后台管理）
export interface ServiceStaffItem {
  id: number
  username: string
  name: string
  phone: string
  openid?: string
  avatar?: string
  status: number // 0=待审核 1=启用 2=禁用
  service_region?: string // 服务区域(区县,逗号分隔,空=不限)
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

// ===================== 服务人员审核 =====================

export interface StaffAuditItem {
  id: number
  staff_id: number
  audit_type: number // 1=注册申请 2=信息变更 3=资质提交 4=状态变更
  apply_type: number // 1=自注册 2=PC添加 3=端上变更
  before_data?: any
  after_data?: any
  qualifications?: any
  status: number // 0=待审 1=通过 2=驳回
  reviewer_id?: number
  review_remark?: string
  review_at?: string
  created_at?: string
  staff_name?: string
  staff_phone?: string
  staff_username?: string
}

export interface StaffAuditListResponse {
  list: StaffAuditItem[]
  total: number
  pagination: { page: number; page_size: number }
}

export interface StaffAuditDetailResponse {
  record: StaffAuditItem
  staff: ServiceStaffItem
  qualifications?: any[]
  before_data?: any
  after_data?: any
}

// ===================== RBAC 类型 =====================

// 系统菜单
export interface SysMenu {
  id: number
  parent_id: number
  menu_type: number // 1=菜单/目录 2=按钮
  name: string
  path?: string
  icon?: string
  sort: number
  status: number
  visible: number
  permission?: string
  created_at?: string
  updated_at?: string
  children?: SysMenu[]
}

// 系统角色
export interface SysRole {
  id: number
  name: string
  code: string
  remark?: string
  status: number
  created_at?: string
  updated_at?: string
  menu_count?: number
}

// 系统部门
export interface SysDepartment {
  id: number
  parent_id: number
  name: string
  leader?: string
  phone?: string
  sort: number
  status: number
  created_at?: string
  updated_at?: string
  children?: SysDepartment[]
}

// 员工列表项（扩展 RBAC）
export interface MerchantStaffListItem extends MerchantStaffInfo {
  department?: SysDepartment
  roles?: SysRole[]
}

// 员工列表响应
export interface MerchantStaffListResponse {
  list: MerchantStaffListItem[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

// RBAC 权限信息（菜单树 + 权限码）
export interface RbacPermissions {
  menus: SysMenu[]
  permissions: string[]
}

// 角色分配菜单请求
export interface AssignRoleMenusRequest {
  menu_ids: number[]
}

// ===================== 健康服务 =====================

// 健康档案
export interface HealthRecord {
  id: number
  user_id: number
  relation?: number // 1本人 2父母 3其他亲属
  real_name?: string
  gender?: number // 1男2女
  birth_date?: string
  id_card?: string
  phone?: string
  emergency_contact?: string
  emergency_phone?: string
  address?: string
  height_cm?: number | null
  weight_kg?: number | null
  blood_type?: string
  past_history?: string[] // JSON 数组
  allergy_history?: string[]
  family_history?: string[]
  surgery_history?: string[]
  medication_list?: string[]
  chronic_tags?: string[] // 慢病标签数组
  smoking?: string
  drinking?: string
  assessment_level?: string
  remark?: string
  status?: number
  created_at?: string
  updated_at?: string
  user?: { id: number; nickname?: string; phone?: string; openid?: string }
  assessments?: HealthAssessment[]
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
  status?: number // 0草稿 1启用
  created_at?: string
  updated_at?: string
}

// 评估记录
export interface HealthAssessment {
  id: number
  user_id: number
  form_id: number
  form_name?: string
  assessor_type?: number // 1自助 2服务人员
  staff_id?: number | null
  answers?: Record<string, string>
  total_score?: number
  level?: string
  conclusion?: string
  suggestions?: string[]
  symptom_desc?: string
  created_at?: string
  user?: { id: number; nickname?: string; phone?: string }
  form?: AssessmentForm
}

// 健康档案列表响应
export interface HealthRecordListResponse {
  list: HealthRecord[]
  total: number
}

// 评估记录列表响应
export interface HealthAssessmentListResponse {
  list: HealthAssessment[]
  total: number
}

// 常见慢病标签（编辑档案时的预置选项）
export const ChronicTagOptions = [
  '高血压',
  '糖尿病',
  '冠心病',
  '脑卒中',
  '慢阻肺',
  '骨质疏松',
  '帕金森',
  '阿尔茨海默',
  '关节炎',
  '其他'
]

// 评估量表维度选项
export const AssessmentDimensionOptions = [
  { label: 'ADL 日常生活能力', value: 'adl' },
  { label: '巴氏指数', value: 'barthel' },
  { label: '跌倒风险', value: 'fall' },
  { label: '营养评估', value: 'nutrition' },
  { label: '认知评估', value: 'cognition' },
  { label: '压疮风险', value: 'pressure' },
  { label: '衰弱筛查', value: 'weak' },
  { label: '老年综合', value: 'geriatric' },
  { label: '通用自评', value: 'self' }
]

// ===================== 适配建议 =====================

// 适配建议推荐商品
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
  user?: { id: number; nickname?: string; phone?: string }
  assessment?: { id: number; form_name?: string; total_score?: number; level?: string }
}

// 适配建议列表响应
export interface FittingRecommendationListResponse {
  list: FittingRecommendation[]
  total: number
}

// ===================== 健康宣教 =====================

// 健康宣教分类
export interface HealthEducationCategory {
  id: number
  parent_id: number // 0=一级
  name: string
  sort: number
  status: number // 1启用 0停用
  created_at?: string
  updated_at?: string
}

// 健康宣教文章
export interface EducationArticle {
  id: number
  title: string
  category_id?: number
  category?: string // 兼容保留旧字符串分类
  cover?: string
  content?: string
  tags?: string[]
  status?: number // 0草稿 1发布
  publish_at?: string
  views?: number
  created_at?: string
  updated_at?: string
}

// 健康宣教文章列表响应
export interface EducationArticleListResponse {
  list: EducationArticle[]
  total: number
}

// ===================== 服务商分账 =====================

// 分账接收方
export interface ProfitSharingReceiver {
  id: number
  merchant_id: number
  receiver_type: number // 1=商户号 2=个人微信
  name: string
  account: string
  personal_name?: string
  relation_type?: string
  default_ratio: number // 自动分账默认比例(%)
  wechat_bound: number // 0=未建立 1=已建立
  wechat_error?: string
  status: number // 1=启用 0=停用
  sort: number
  remark?: string
  created_at?: string
  updated_at?: string
}

// 分账接收方创建/编辑载荷
export interface ProfitSharingReceiverPayload {
  receiver_type: number
  name: string
  account: string
  personal_name?: string
  relation_type?: string
  default_ratio: number
  status: number
  sort?: number
  remark?: string
}

// 分账单明细（各方）
export interface ProfitSharingRecordReceiver {
  id: number
  record_id: number
  receiver_id?: number
  receiver_type: number
  receiver_name: string
  account: string
  amount: number
  result_status?: string
  detail_id?: string
  fail_reason?: string
  finish_time?: string
}

// 分账单
export interface ProfitSharingRecord {
  id: number
  merchant_id: number
  order_id: number
  order_no: string
  sp_mchid?: string
  sub_mchid?: string
  appid?: string
  transaction_id?: string
  out_order_no: string
  total_amount: number
  total_share_amount: number
  status: number // 0=待分账 1=分账中 2=成功 3=失败 4=跳过
  share_time?: string
  error_message?: string
  created_at?: string
  order?: SpOrder
  receivers?: ProfitSharingRecordReceiver[]
}

// 分账记录列表响应
export interface ProfitSharingRecordListResponse {
  list: ProfitSharingRecord[]
  pagination: {
    total: number
    page: number
    page_size: number
  }
}

// 分账配置
export interface ProfitSharingConfig {
  profit_sharing_enabled: boolean
  sub_mch_id: string
  max_ratio_pct: string
  max_ratio_raw?: string
}

// ---------- 小程序轮播图 ----------

export interface MiniProgramBanner {
  id: number
  merchant_id: number
  title: string
  image: string
  link_type: 'none' | 'product' | 'category' | 'url'
  link_value: string
  sort: number
  status: number // 1=启用 0=禁用
  created_at: string
  updated_at: string
}

export interface MiniProgramBannerListResponse {
  list: MiniProgramBanner[]
}

export interface MiniProgramBannerPayload {
  title?: string
  image: string
  link_type: 'none' | 'product' | 'category' | 'url'
  link_value?: string
  sort?: number
  status?: number
}

// ===================== 小程序首页推荐配置 =====================

export interface HomeRecommend {
  id: number
  merchant_id: number
  product_id: number
  target_type: number // 1=实物商品(产品类型1/2) 2=服务(产品类型3/4)
  title: string
  sort: number
  status: number // 1=启用 0=禁用
  product_name?: string
  product_price?: number
  product_image?: string
  created_at: string
  updated_at: string
}

export interface HomeRecommendListResponse {
  list: HomeRecommend[]
}

export interface HomeRecommendPayload {
  product_id: number
  target_type: number
  title?: string
  sort?: number
  status?: number
}

// ===================== 商家运营设置 =====================

// 商家营业设置：营业状态 + 下单方式 + 公告/营业时间等
export interface MerchantSettings {
  takeout_enabled?: boolean // 是否开启外卖配送
  dine_in_enabled?: boolean // 是否开启堂食
  pickup_enabled?: boolean // 是否开启自提
  delivery_settings?: Record<string, unknown> // 配送相关配置
  announcement?: string // 商家公告
  business_hours?: string // 营业时间文本
}

// ===================== 打印机管理 =====================

// 打印机
export interface Printer {
  id: number
  merchant_id?: number
  name: string
  type: number // 1=飞鹅 2=通用
  feie_user?: string // 飞鹅云平台用户名
  feie_ukey?: string // 飞鹅云平台密钥（仅新增时返回，编辑不回传）
  feie_sn?: string // 飞鹅打印机编号
  status: number // 1=启用 0=禁用
  auto_print: number // 1=自动打印 0=手动
  is_default: number // 1=默认打印机 0=否
  print_count?: number // 已打印次数
  last_print_at?: string // 最近打印时间
  remark?: string
  has_feie_ukey?: boolean // 编辑时后端是否已保存密钥（存在则不回显 ukey）
  created_at?: string
  updated_at?: string
}

// 打印机新增/编辑载荷
export interface PrinterPayload {
  name: string
  type: number
  feie_user?: string
  feie_ukey?: string
  feie_sn?: string
  status: number
  auto_print: number
  remark?: string
}

// 打印机列表响应
export interface PrinterListResponse {
  list: Printer[]
}

// 打印机类型文案（1=飞鹅 2=通用）
export const PrinterTypeText: Record<number, string> = {
  1: '飞鹅',
  2: '通用'
}
