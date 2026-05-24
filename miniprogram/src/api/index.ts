/**
 * API接口封装 - 基于PRD文档
 * 统一管理所有API接口调用
 */

import { get, post, put, del } from '../utils/request'
import type { RequestOptions } from '../utils/request'
import type {
  MerchantLoginRequest,
  MerchantLoginResponse,
  ServiceProviderLoginRequest,
  ServiceProviderLoginResponse,
  SpSettings,
  MerchantListItem,
  MerchantDetail,
  AnnouncementStatus,
  MerchantDistributionData,
  OrderAnalyticsData,
  AmountAnalyticsData,
  TopMerchantRanking,
  MerchantInfo,
  MerchantWechatLoginRequest,
  MerchantSettings,
  MerchantStaffListResponse,
  CreateMerchantStaffRequest,
  UpdateMerchantStaffRequest,
  ChangePasswordRequest,
  UploadTokenResponse,
  DeliverySettings,
  MerchantDeliverySettings,
  StoreDeliveryRules,
  Category,
  Product,
  ProductListResponse,
  Order,
  OrderListResponse,
  OrderStatistics,
  StockAlert,
  ProductRanking,
  HourlyAnalysis,
  SalesOverview,
  SalesTrend,
  CustomerAnalysis,
  CustomerTrend,
  StoreHomeInfo,
  StoreProductGroup,
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantBehaviorEventRequest,
  Announcement,
  AnnouncementListResponse,
  SpMerchantFormData,
  UpdateSpMerchantFormData,
  MerchantPaymentConfigFormData,
  ProfitSharingRecordListResponse,
  ProfitSharingRecordQuery,
  UserAddress,
} from '../types'
import type { ApiResponse } from '../types'

export { get, post, put, del }
export * from '../types'

export const ResponseCode = {
  SUCCESS: 0,
  PARAM_ERROR: 1001,
  UNAUTHORIZED: 1002,
  FORBIDDEN: 1003,
  NOT_FOUND: 1004,
  SERVER_ERROR: 5000,
  PRODUCT_NOT_FOUND: 6001,
  ORDER_NOT_FOUND: 6002,
  CATEGORY_NOT_FOUND: 7001,
  CATEGORY_HAS_PRODUCT: 7002,
  QINIU_UPLOAD_FAILED: 8002,
} as const

function getToken(): string {
  return uni.getStorageSync('token') || ''
}

function getUserToken(): string {
  return uni.getStorageSync('user_token') || ''
}

function parseDistanceRules(value: unknown): { min_distance: number; max_distance: number; fee: number }[] {
  if (Array.isArray(value)) {
    return value.map((item: any) => ({
      min_distance: Number(item?.min_distance || 0),
      max_distance: Number(item?.max_distance || 0),
      fee: Number(item?.fee || 0)
    }))
  }

  if (typeof value === 'string' && value) {
    try {
      return parseDistanceRules(JSON.parse(value))
    } catch (error) {
      console.warn('解析配送规则失败:', error)
    }
  }

  return []
}

function normalizeDeliverySettings(data: Partial<DeliverySettings> | null | undefined): DeliverySettings {
  return {
    enabled: !!data?.enabled,
    base_fee: Number(data?.base_fee || 0),
    free_delivery_amount: Number(data?.free_delivery_amount || 0),
    max_distance: Number(data?.max_distance || 10),
    distance_rules: parseDistanceRules(data?.distance_rules)
  }
}

function normalizeStoreDeliveryRules(data: Partial<StoreDeliveryRules> | null | undefined): StoreDeliveryRules {
  return {
    ...normalizeDeliverySettings(data),
    takeout_enabled: !!data?.takeout_enabled,
    dine_in_enabled: !!data?.dine_in_enabled,
    pickup_enabled: !!data?.pickup_enabled
  }
}

function normalizeMerchantDeliverySettings(
  data: Partial<MerchantDeliverySettings> | null | undefined
): MerchantDeliverySettings {
  return {
    ...normalizeDeliverySettings(data),
    takeout_enabled: !!data?.takeout_enabled,
    dine_in_enabled: !!data?.dine_in_enabled,
    pickup_enabled: !!data?.pickup_enabled
  }
}

function normalizeArrayResponse<T>(response: T[] | null | undefined): T[] {
  return Array.isArray(response) ? response : []
}

function normalizeListField<T, R extends { list?: T[] | null }>(response: R): R & { list: T[] } {
  return {
    ...response,
    list: normalizeArrayResponse(response?.list)
  }
}

function normalizeMerchantSettings(data: MerchantSettings): MerchantSettings {
  return {
    ...data,
    takeout_enabled: !!data?.takeout_enabled,
    dine_in_enabled: !!data?.dine_in_enabled,
    pickup_enabled: !!data?.pickup_enabled,
    delivery_settings: data?.delivery_settings
      ? normalizeDeliverySettings(data.delivery_settings)
      : undefined
  }
}

// ============ 认证相关 ============

/**
 * 商家登录
 */
export async function merchantLogin(data: MerchantLoginRequest): Promise<MerchantLoginResponse> {
  const res = await post<MerchantLoginResponse>('/api/v1/auth/merchant/login', data)
  if (res.token) {
    uni.setStorageSync('token', res.token)
    uni.setStorageSync('merchant_id', res.merchant_id)
    uni.setStorageSync('merchant_info', res.staff)
  }
  return res
}

export async function merchantWechatLogin(data: MerchantWechatLoginRequest): Promise<MerchantLoginResponse> {
  const res = await post<MerchantLoginResponse>('/api/v1/auth/merchant/wechat-login', data)
  if (res.token) {
    uni.setStorageSync('token', res.token)
    uni.setStorageSync('merchant_id', res.merchant_id)
    uni.setStorageSync('merchant_info', res.staff)
  }
  return res
}

export function spLogin(data: ServiceProviderLoginRequest) {
  return post<ServiceProviderLoginResponse>('/api/v1/sp/auth/login', data)
}

export function spLogout() {
  return post<{ message: string }>('/api/v1/sp/auth/logout', {})
}

export function getSpSettings() {
  return get<SpSettings>('/api/v1/sp/settings')
}

export function changeSpPassword(data: { old_password: string; new_password: string }) {
  return post<{ message: string }>('/api/v1/sp/account/change-password', data)
}

export function createSpMerchant(data: SpMerchantFormData) {
  return post<MerchantDetail>('/api/v1/sp/merchants', data)
}

export function updateSpMerchant(merchantId: number, data: UpdateSpMerchantFormData) {
  return put<MerchantDetail>(`/api/v1/sp/merchants/${merchantId}`, data)
}

export function updateSpMerchantPaymentConfig(merchantId: number, data: MerchantPaymentConfigFormData) {
  return put<MerchantDetail>(`/api/v1/sp/merchants/${merchantId}/payment-config`, data)
}

export async function getMerchantList(params?: any) {
  const res = await get<{ list: any[]; pagination: { total: number; page: number; page_size: number } }>(
    '/api/v1/sp/merchants/list',
    params
  )

  const list: MerchantListItem[] = (res.list || []).map((item: any) => ({
    ...item,
    total_users: Number(item?.total_users || 0),
    total_orders: Number(item?.total_orders || 0),
    total_amount: Number(item?.total_amount || 0)
  }))

  return { ...res, list }
}

export function getMerchantDetail(merchantId: number) {
  return get<MerchantDetail>(`/api/v1/sp/merchants/${merchantId}`).then(data => {
    if (data?.logo) data.logo = normalizeImageUrl(data.logo)
    if (data?.cover_image) data.cover_image = normalizeImageUrl(data.cover_image)
    return data
  })
}

export function getSpProfitSharingRecords(params?: ProfitSharingRecordQuery) {
  return get<ProfitSharingRecordListResponse>('/api/v1/sp/profit-sharing-records', params).then(normalizeListField)
}

export function getMerchantProfitSharingRecords(params?: Omit<ProfitSharingRecordQuery, 'merchant_id'>) {
  return get<ProfitSharingRecordListResponse>('/api/v1/merchant/profit-sharing-records', params).then(normalizeListField)
}

export function updateSpMerchantAssets(merchantId: number, data: { logo?: string; cover_image?: string }) {
  return put<MerchantDetail>(`/api/v1/sp/merchants/${merchantId}/assets`, data)
}

export function getMerchantDistribution() {
  return get<MerchantDistributionData>('/api/v1/sp/merchants/analytics/distribution')
}

export function getOrderAnalytics(params?: { days?: number }) {
  return get<OrderAnalyticsData>('/api/v1/sp/orders/analytics', params)
}

export function getAmountAnalytics(params?: { days?: number }) {
  return get<AmountAnalyticsData>('/api/v1/sp/amount/analytics', params)
}

export function getTopMerchants(params?: { limit?: number; metric?: string }) {
  return get<TopMerchantRanking[] | null>('/api/v1/sp/amount/top-merchants', params).then(normalizeArrayResponse)
}

/**
 * 获取商家信息
 */
export async function getMerchantProfile() {
  const res = await get<any>('/api/v1/merchant/profile')
  const merchant = res?.merchant || res
  if (merchant?.logo) merchant.logo = normalizeImageUrl(merchant.logo)
  if (merchant?.cover_image) merchant.cover_image = normalizeImageUrl(merchant.cover_image)
  return merchant as MerchantInfo
}

// ============ 商家设置相关 ============

/**
 * 更新商家基本信息
 */
export function updateMerchantProfile(data: Partial<MerchantInfo>) {
  return put<null>('/api/v1/merchant/profile', data)
}

/**
 * 获取商家设置
 */
export function getMerchantSettings() {
  return get<MerchantSettings>('/api/v1/merchant/settings').then(normalizeMerchantSettings)
}

/**
 * 更新商家设置
 */
export function updateMerchantSettings(data: Partial<MerchantSettings>) {
  return put<null>('/api/v1/merchant/settings', data)
}

export function changeMerchantPassword(data: ChangePasswordRequest) {
  return post<{ message: string }>('/api/v1/merchant/account/change-password', data)
}

export function bindMerchantWechat(data: { code: string }) {
  return post<{ openid: string; unionid?: string; wechat_bound_at: string; message: string }>('/api/v1/merchant/account/wechat/bind', data)
}

export function unbindMerchantWechat() {
  return del<{ message: string }>('/api/v1/merchant/account/wechat/bind')
}

/**
 * 获取配送设置
 */
export function getDeliverySettings() {
  return get<MerchantDeliverySettings>('/api/v1/merchant/delivery-settings').then(normalizeMerchantDeliverySettings)
}

/**
 * 更新配送设置
 */
export function updateDeliverySettings(data: Partial<DeliverySettings>) {
  const payload = normalizeDeliverySettings(data)
  return put<MerchantDeliverySettings>('/api/v1/merchant/delivery-settings', payload).then(normalizeMerchantDeliverySettings)
}

/**
 * 更新商家状态
 */
export function updateMerchantStatus(status: number) {
  return post<null>('/api/v1/merchant/status', { status })
}

/**
 * 获取商家小程序码
 */
export function getMerchantQrcode(params?: { page?: string; width?: number }) {
  return get<{ qrcode_url: string; expire_time?: string }>('/api/v1/merchant/qrcode', params)
}

export function getMerchantAnnouncements(params?: { page?: number; page_size?: number }) {
  return get<AnnouncementListResponse>('/api/v1/merchant/announcements', params).then(normalizeListField)
}

export function getMerchantAnnouncementDetail(announcementId: number) {
  return get<Announcement>(`/api/v1/merchant/announcements/${announcementId}`)
}

// ============ 商品分类相关 ============

/**
 * 获取分类列表
 */
export function getCategories(options?: Partial<RequestOptions>) {
  const token = uni.getStorageSync('token') || ''
  const spToken = uni.getStorageSync('sp_token') || ''
  if (!token && spToken) {
    return Promise.resolve([])
  }
  return get<Category[] | null>('/api/v1/merchant/categories', undefined, options).then(normalizeArrayResponse)
}

/**
 * 创建分类
 */
export function createCategory(data: { name: string; sort?: number }) {
  return post<Category>('/api/v1/merchant/categories', data)
}

/**
 * 更新分类
 */
export function updateCategory(categoryId: number, data: { name?: string; sort?: number }) {
  return put<Category>(`/api/v1/merchant/categories/${categoryId}`, data)
}

/**
 * 删除分类
 */
export function deleteCategory(categoryId: number) {
  return del<null>(`/api/v1/merchant/categories/${categoryId}`)
}

/**
 * 批量排序分类
 */
export function sortCategories(orders: { id: number; sort: number }[]) {
  return post<null>('/api/v1/merchant/categories/sort', { orders })
}

// ============ 商品管理相关 ============

function normalizeStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === 'string')
  }

  if (typeof value === 'string' && value) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return parsed.filter((item): item is string => typeof item === 'string')
      }
    } catch (error) {
      console.warn('解析商品图片失败:', error)
    }
  }

  return []
}

function joinQiniuFileUrl(domain: string, keyOrUrl: string): string {
  if (!keyOrUrl) {
    return ''
  }

  if (/^https?:\/\//i.test(keyOrUrl)) {
    return keyOrUrl
  }

  const normalizedDomain = domain.replace(/\/+$/, '')
  const normalizedKey = keyOrUrl.replace(/^\/+/, '')
  return `${normalizedDomain}/${normalizedKey}`
}

function normalizeImageUrl(keyOrUrl: string): string {
  const domain = uni.getStorageSync('qiniu_domain') || ''
  if (!domain) {
    return keyOrUrl
  }

  return joinQiniuFileUrl(domain, keyOrUrl)
}

function getPersistedImageUrl(keyOrUrl: string): string {
  if (!keyOrUrl) {
    return ''
  }

  const normalizedUrl = normalizeImageUrl(keyOrUrl)
  const queryIndex = normalizedUrl.indexOf('?')
  if (queryIndex === -1) {
    return normalizedUrl
  }

  return normalizedUrl.slice(0, queryIndex)
}

function normalizeSpecs(value: unknown): Product['specs'] {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((spec: any) => ({
    id: typeof spec?.id === 'number' ? spec.id : undefined,
    name: spec?.name || '',
    options: Array.isArray(spec?.options)
      ? spec.options.map((option: any) => ({
          id: typeof option?.id === 'number' ? option.id : undefined,
          name: option?.name || '',
          price: Number(option?.price || 0),
          stock: typeof option?.stock === 'number' ? option.stock : undefined
        }))
      : []
  }))
}

function normalizeProduct(product: any): Product {
  return {
    ...product,
    id: Number(product?.id || 0),
    category_id: Number(product?.category_id || 0),
    price: Number(product?.price || 0),
    original_price: product?.original_price !== undefined ? Number(product.original_price || 0) : undefined,
    stock: Number(product?.stock || 0),
    sales: product?.sales !== undefined ? Number(product.sales || 0) : undefined,
    sort: product?.sort !== undefined ? Number(product.sort || 0) : undefined,
    images: normalizeStringArray(product?.images).map(normalizeImageUrl),
    specs: normalizeSpecs(product?.specs)
  } as Product
}

function normalizeStockAlertItem(item: any): StockAlert {
  const normalizedImages = normalizeStringArray(item?.images).map(normalizeImageUrl)
  return {
    id: typeof item?.id === 'number' ? item.id : Number(item?.id || 0) || undefined,
    product_id: Number(item?.product_id || item?.id || 0),
    product_name: String(item?.product_name || item?.name || ''),
    image: normalizeImageUrl(item?.image || normalizedImages[0] || ''),
    stock: Number(item?.stock || 0),
    status: item?.status
  }
}

function normalizeOrderItem(item: any) {
  const normalizedImages = normalizeStringArray(item?.images).map(normalizeImageUrl)
  const primaryImage = normalizeImageUrl(item?.image || item?.product_image || normalizedImages[0] || '')

  return {
    ...item,
    product_id: Number(item?.product_id || 0),
    product_name: String(item?.product_name || ''),
    image: primaryImage,
    images: normalizedImages,
    price: Number(item?.price || 0),
    quantity: Number(item?.quantity || 0),
    specs: item?.specs || item?.spec_info || ''
  }
}

function normalizeOrder(order: any): Order {
  const merchant = order?.merchant
    ? {
        ...order.merchant,
        id: Number(order.merchant?.id || 0),
        logo: normalizeImageUrl(order.merchant?.logo || '')
      }
    : undefined

  return {
    ...order,
    id: Number(order?.id || 0),
    items: Array.isArray(order?.items) ? order.items.map(normalizeOrderItem) : [],
    total_amount: Number(order?.total_amount || 0),
    delivery_fee: Number(order?.delivery_fee || 0),
    discount_amount: Number(order?.discount_amount || 0),
    pay_amount: Number(order?.pay_amount || 0),
    status: Number(order?.status || 0),
    merchant
  } as Order
}

/**
 * 获取商品列表
 */
export async function getProducts(params?: {
  page?: number
  page_size?: number
  category_id?: number
  status?: string
  keyword?: string
}) {
  const res = await get<ProductListResponse>('/api/v1/merchant/products', params)
  return {
    ...res,
    list: Array.isArray(res?.list) ? res.list.map(normalizeProduct) : []
  }
}

/**
 * 获取商品详情
 */
export async function getProduct(productId: number, options?: Partial<RequestOptions>) {
  const res = await get<Product>(`/api/v1/merchant/products/${productId}`, undefined, options)
  return normalizeProduct(res)
}

/**
 * 创建商品
 */
export async function createProduct(data: {
  name: string
  description?: string
  images: string[]
  category_id: number
  price: number
  original_price?: number
  stock?: number
  unit?: string
  sort?: number
  specs?: { name: string; options: { name: string; price: number; stock?: number }[] }[]
}) {
  const payload = {
    ...data,
    images: Array.isArray(data.images) ? data.images.map(getPersistedImageUrl) : []
  }
  const res = await post<Product>('/api/v1/merchant/products', payload)
  return normalizeProduct(res)
}

/**
 * 更新商品
 */
export async function updateProduct(productId: number, data: Partial<Product>) {
  const payload = {
    ...data,
    images: Array.isArray(data.images) ? data.images.map(getPersistedImageUrl) : data.images
  }
  const res = await put<Product>(`/api/v1/merchant/products/${productId}`, payload)
  return normalizeProduct(res)
}

/**
 * 商品上架
 */
export function productOnSale(productId: number) {
  return post<null>(`/api/v1/merchant/products/${productId}/on-sale`)
}

/**
 * 商品下架
 */
export function productOffSale(productId: number) {
  return post<null>(`/api/v1/merchant/products/${productId}/off-sale`)
}

/**
 * 批量更新商品状态
 */
export function batchUpdateProductStatus(productIds: number[], status: number) {
  return post<null>('/api/v1/merchant/products/batch-status', { product_ids: productIds, status })
}

/**
 * 删除商品
 */
export function deleteProduct(productId: number) {
  return del<null>(`/api/v1/merchant/products/${productId}`)
}

/**
 * 更新商品库存
 */
export function updateProductStock(productId: number, stock: number) {
  return put<null>(`/api/v1/merchant/products/${productId}/stock`, { stock })
}

// ============ 订单管理相关 ============

/**
 * 获取订单列表
 */
export function getOrders(params?: {
  page?: number
  page_size?: number
  status?: number
  start_date?: string
  end_date?: string
  order_no?: string
}) {
  return get<OrderListResponse>('/api/v1/merchant/orders', params).then(normalizeListField)
}

/**
 * 获取订单详情
 */
export function getOrder(orderId: number) {
  return get<Order>(`/api/v1/merchant/orders/${orderId}`)
}

/**
 * 订单核销
 */
export function completeOrder(orderId: number, verifyCode: string) {
  return post<Order>(
    `/api/v1/merchant/orders/${orderId}/complete`,
    { verify_code: verifyCode }
  )
}

export function quickCompleteOrder(verifyCode: string) {
  return post<Order>('/api/v1/merchant/orders/quick-complete', {
    verify_code: verifyCode
  })
}

/**
 * 退款订单
 */
export function refundOrder(orderId: number, data: { reason?: string; refund_amount?: number }) {
  return post<null>(`/api/v1/merchant/orders/${orderId}/refund`, data)
}

/**
 * 获取订单统计
 */
export function getOrderStatistics() {
  return get<OrderStatistics>('/api/v1/merchant/orders/statistics')
}

// ============ 数据分析相关 ============

/**
 * 获取销售概览
 */
export function getSalesOverview(params?: { period?: string }) {
  return get<SalesOverview>('/api/v1/merchant/analytics/overview', params)
}

/**
 * 获取销售趋势
 */
export function getSalesTrend(params: { start_date: string; end_date: string; granularity?: string }) {
  return get<SalesTrend[] | null>('/api/v1/merchant/analytics/sales-trend', params).then(normalizeArrayResponse)
}

/**
 * 获取商品排行
 */
export function getProductRanking(params?: { start_date?: string; end_date?: string; limit?: number; sort_by?: string }) {
  return get<ProductRanking[] | null>('/api/v1/merchant/analytics/product-ranking', params).then(normalizeArrayResponse)
}

/**
 * 获取时段分析
 */
export function getHourlyAnalysis(params: { date: string }) {
  return get<HourlyAnalysis[] | null>('/api/v1/merchant/analytics/hourly', params).then(normalizeArrayResponse)
}

/**
 * 获取库存预警
 */
export function getStockAlert(params?: { threshold?: number }) {
  // 后端无预警数据时可能返回 null，这里统一兜底为空数组，避免页面直接读取 length 报错。
  return get<StockAlert[] | null>('/api/v1/merchant/analytics/stock-alert', params)
    .then(normalizeArrayResponse)
    .then(list => list.map(normalizeStockAlertItem))
}

/**
 * 获取客户分析
 */
export function getCustomerAnalysis() {
  return get<CustomerAnalysis>('/api/v1/merchant/analytics/customers')
}

/**
 * 获取客户趋势
 */
export function getCustomerTrend(params: { start_date: string; end_date: string }) {
  return get<CustomerTrend[] | null>('/api/v1/merchant/analytics/customer-trend', params).then(normalizeArrayResponse)
}

export function getMerchantStaffList(params?: { page?: number; page_size?: number }) {
  return get<MerchantStaffListResponse>('/api/v1/merchant/staff', params).then(normalizeListField)
}

export function createMerchantStaff(data: CreateMerchantStaffRequest) {
  return post<{ id: number; message: string }>('/api/v1/merchant/staff', data)
}

export function updateMerchantStaff(staffId: number, data: UpdateMerchantStaffRequest) {
  return put(`/api/v1/merchant/staff/${staffId}`, data)
}

export function deleteMerchantStaff(staffId: number) {
  return del<{ message: string }>(`/api/v1/merchant/staff/${staffId}`)
}

export function resetMerchantStaffPassword(staffId: number, newPassword: string) {
  return post<{ message: string }>(`/api/v1/merchant/staff/${staffId}/reset-password`, {
    new_password: newPassword
  })
}

// ============ 文件上传相关 ============

/**
 * 获取上传凭证
 */
export async function getUploadToken() {
  const res = await get<UploadTokenResponse>('/api/v1/upload/token')
  if (res?.domain) {
    uni.setStorageSync('qiniu_domain', res.domain)
  }
  return res
}

/**
 * 上传图片 - 客户端直传七牛云
 */
export async function uploadImage(filePath: string): Promise<{ url: string; key: string }> {
  uni.showLoading({ title: '上传中...', mask: true })
  
  try {
    const uploadData = await getUploadToken()
    
    const ext = filePath.split('.').pop() || 'jpg'
    const key = `${uploadData.prefix}/${Date.now()}.${ext}`
    
    return new Promise((resolve, reject) => {
      uni.uploadFile({
        url: uploadData.upload_url || 'https://up.qiniup.com',
        method: 'POST',
        filePath,
        name: 'file',
        formData: {
          token: uploadData.token,
          key: key
        },
        success: (res) => {
          uni.hideLoading()
          if (res.statusCode === 200) {
            const data = JSON.parse(res.data)
            if (data.key) {
              resolve({
                url: joinQiniuFileUrl(uploadData.domain, data.key),
                key: data.key
              })
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

// ============ C端店铺相关 ============

/**
 * 获取店铺首页信息
 */
export function getStoreHome(merchantId: number) {
  return get<StoreHomeInfo>(`/api/v1/store/${merchantId}/home`).then(data => {
    if (data?.hot_products) {
      data.hot_products = data.hot_products.map((p: any) => {
        const normalized = normalizeProduct(p)
        return {
          id: normalized.id,
          name: normalized.name,
          images: normalized.images || [],
          price: normalized.price,
          original_price: normalized.original_price,
          sales: Number(normalized.sales || 0)
        }
      })
    }
    if (data?.merchant?.logo) {
      data.merchant.logo = normalizeImageUrl(data.merchant.logo)
    }
    if (data?.merchant?.cover_image) {
      data.merchant.cover_image = normalizeImageUrl(data.merchant.cover_image)
    }
    return data
  })
}

/**
 * 获取店铺商品列表
 */
export function getStoreProducts(merchantId: number, params?: { category_id?: number }) {
  return get<ProductListResponse>(`/api/v1/store/${merchantId}/products`, params).then(res => {
    const normalized = normalizeListField(res)
    return { ...normalized, list: normalized.list.map(normalizeProduct) }
  })
}

/**
 * 获取店铺商品详情
 */
export function getStoreProduct(merchantId: number, productId: number) {
  return get<Product>(`/api/v1/store/${merchantId}/products/${productId}`).then(normalizeProduct)
}

/**
 * 获取配送费规则
 */
export function getStoreDeliveryRules(merchantId: number) {
  return get<StoreDeliveryRules>(`/api/v1/store/${merchantId}/delivery-rules`).then(normalizeStoreDeliveryRules)
}

// ============ C端订单相关 ============

/**
 * 创建订单
 */
export function createOrder(merchantId: number, data: CreateOrderRequest) {
  return post<CreateOrderResponse>(`/api/v1/store/${merchantId}/orders`, data)
}

export function trackStoreBehaviorEvent(merchantId: number, data: MerchantBehaviorEventRequest) {
  return post<{ message: string }>(`/api/v1/store/${merchantId}/event`, data, { loading: false, showErrorToast: false })
}

/**
 * 获取我的订单列表
 * @param params.page 页码
 * @param params.page_size 每页数量
 * @param params.status 订单状态
 * @param params.merchant_id 商家ID（可选，用于筛选特定商家的订单）
 */
export function getMyOrders(params?: {
  page?: number
  page_size?: number
  status?: number
  merchant_id?: number
}) {
  return get<OrderListResponse>('/api/v1/user/orders', params).then((response) => {
    const normalized = normalizeListField(response)
    return {
      ...normalized,
      list: normalized.list.map(normalizeOrder)
    }
  })
}

/**
 * 取消订单
 */
export function cancelMyOrder(orderId: number) {
  return post<null>(`/api/v1/user/orders/${orderId}/cancel`)
}

/**
 * 申请退款
 */
export function applyRefund(orderId: number, data: { reason: string }) {
  return post<any>(`/api/v1/user/orders/${orderId}/refund`, data)
}

// ============ 云打印相关 ============

/**
 * 获取打印记录
 */
export function getPrintLogs(params?: { page?: number; page_size?: number; start_date?: string; end_date?: string }) {
  return get<any>('/api/v1/merchant/print-logs', params)
}

// ============ C端订单详情 ============

export function getMyOrderDetail(orderId: number) {
  return get<Order>(`/api/v1/user/orders/${orderId}`).then(normalizeOrder)
}

// ============ C端地址管理 ============

export function getUserAddresses() {
  return get<UserAddress[]>('/api/v1/user/addresses')
}

export function createUserAddress(data: Partial<UserAddress>) {
  return post<UserAddress>('/api/v1/user/addresses', data)
}

export function updateUserAddress(addressId: number, data: Partial<UserAddress>) {
  return put<UserAddress>(`/api/v1/user/addresses/${addressId}`, data)
}

export function deleteUserAddress(addressId: number) {
  return del<null>(`/api/v1/user/addresses/${addressId}`)
}

// ============ SP端缺失API ============

export function updateSpSettings(data: Partial<SpSettings>) {
  return put<SpSettings>('/api/v1/sp/settings', data)
}

export function getMerchantFee(merchantId: number) {
  return get<any>(`/api/v1/sp/merchants/${merchantId}/fee`)
}

export function getMerchantRate(merchantId: number) {
  return get<any>(`/api/v1/sp/merchants/${merchantId}/rate`)
}

export function setMerchantRate(merchantId: number, data: any) {
  return post<any>(`/api/v1/sp/merchants/${merchantId}/rate`, data)
}

export function getSpMerchantQrcode(merchantId: number) {
  return get<{ qrcode_url: string }>(`/api/v1/sp/merchants/${merchantId}/qrcode`)
}

export function getSpRefunds(params?: any) {
  return get<any>('/api/v1/sp/orders/refunds', params)
}

export function getSpActivities(params?: any) {
  return get<any>('/api/v1/sp/activities', params)
}

export function createSpActivity(data: any) {
  return post<any>('/api/v1/sp/activities', data)
}

export function updateSpActivity(activityId: number, data: any) {
  return put<any>(`/api/v1/sp/activities/${activityId}`, data)
}

export function deleteSpActivity(activityId: number) {
  return del<any>(`/api/v1/sp/activities/${activityId}`)
}

export function getSpWechatConfig() {
  return get<any>('/api/v1/sp/wechat-config')
}

export function updateSpWechatConfig(data: any) {
  return put<any>('/api/v1/sp/wechat-config', data)
}

export default {
  // 认证
  merchantLogin,
  merchantWechatLogin,
  getMerchantProfile,
  createSpMerchant,
  updateSpMerchant,
  updateSpMerchantPaymentConfig,
  getSpProfitSharingRecords,
  getMerchantProfitSharingRecords,
  // 商家设置
  updateMerchantProfile,
  getMerchantSettings,
  updateMerchantSettings,
  changeMerchantPassword,
  bindMerchantWechat,
  unbindMerchantWechat,
  getDeliverySettings,
  updateDeliverySettings,
  updateMerchantStatus,
  getMerchantQrcode,
  getMerchantStaffList,
  createMerchantStaff,
  updateMerchantStaff,
  deleteMerchantStaff,
  resetMerchantStaffPassword,
  // 商品分类
  getCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  sortCategories,
  // 商品管理
  getProducts,
  getProduct,
  createProduct,
  updateProduct,
  productOnSale,
  productOffSale,
  batchUpdateProductStatus,
  deleteProduct,
  updateProductStock,
  // 订单管理
  getOrders,
  getOrder,
  completeOrder,
  quickCompleteOrder,
  refundOrder,
  getOrderStatistics,
  // 数据分析
  getSalesOverview,
  getSalesTrend,
  getProductRanking,
  getHourlyAnalysis,
  getStockAlert,
  getCustomerAnalysis,
  getCustomerTrend,
  // 文件上传
  getUploadToken,
  uploadImage,
  // C端店铺
  getStoreHome,
  getStoreProducts,
  getStoreProduct,
  getStoreDeliveryRules,
  trackStoreBehaviorEvent,
  // C端订单
  createOrder,
  getMyOrders,
  cancelMyOrder,
  applyRefund,
  // 云打印
  getPrintLogs,
  // C端订单详情
  getMyOrderDetail,
  // C端地址管理
  getUserAddresses,
  createUserAddress,
  updateUserAddress,
  deleteUserAddress,
  // SP端缺失API
  updateSpSettings,
  getMerchantFee,
  getMerchantRate,
  setMerchantRate,
  getSpMerchantQrcode,
  getSpRefunds,
  getSpActivities,
  createSpActivity,
  updateSpActivity,
  deleteSpActivity,
  getSpWechatConfig,
  updateSpWechatConfig,
}
