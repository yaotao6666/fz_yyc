/**
 * API接口封装 - 基于PRD文档
 * 统一管理所有API接口调用
 */

import { get, post, put, del } from '../utils/request'
import type { RequestOptions } from '../utils/request'
import type {
  MerchantLoginRequest,
  MerchantLoginResponse,
  MerchantInfo,
  MerchantWechatLoginRequest,
  MerchantSettings,
  MerchantFullReductionRule,
  MerchantFullReductionRulesResponse,
  MerchantPrinter,
  MerchantPrinterPayload,
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
  ProductApiResponse,
  ProductApiSpec,
  ProductApiSpecOption,
  ProductListResponse,
  SpecOption,
  ProductUpsertPayload,
  Order,
  OrderListResponse,
  OrderStatistics,
  StockAlert,
  ProductRanking,
  HourlyAnalysis,
  SalesOverview,
  SalesTrend,
  StoreHomeInfo,
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantBehaviorEventRequest,
  AnnouncementListResponse,
  ProfitSharingRecordListResponse,
  ProfitSharingRecordQuery,
  StoreFullReductionRulesResponse,
  UserAddress,
  UpdateMerchantFullReductionRulesRequest,
} from '../types'

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

function normalizeMerchantFullReductionRule(data: Partial<MerchantFullReductionRule> | null | undefined): MerchantFullReductionRule {
  return {
    id: normalizeNumberValue(data?.id),
    merchant_id: normalizeNumberValue(data?.merchant_id),
    threshold_amount: normalizeRequiredNumber(data?.threshold_amount),
    discount_amount: normalizeRequiredNumber(data?.discount_amount),
    sort: normalizeNumberValue(data?.sort),
    status: normalizeRequiredNumber(data?.status, 1),
    created_at: data?.created_at ? String(data.created_at) : undefined,
    updated_at: data?.updated_at ? String(data.updated_at) : undefined
  }
}

function normalizeMerchantFullReductionRulesResponse(
  data: MerchantFullReductionRulesResponse | null | undefined
): MerchantFullReductionRulesResponse {
  return {
    rules: normalizeArrayResponse(data?.rules).map(normalizeMerchantFullReductionRule),
    active_rules: normalizeArrayResponse(data?.active_rules).map(normalizeMerchantFullReductionRule)
  }
}

function parsePrinterPrintTypes(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === 'string')
  }

  if (typeof value === 'string' && value) {
    try {
      return parsePrinterPrintTypes(JSON.parse(value))
    } catch (error) {
      console.warn('解析打印类型失败:', error)
    }
  }

  return []
}

function normalizeMerchantPrinter(printer: Partial<MerchantPrinter> | null | undefined): MerchantPrinter {
  return {
    id: normalizeRequiredNumber(printer?.id),
    merchant_id: normalizeRequiredNumber(printer?.merchant_id),
    name: printer?.name || '',
    type: String(printer?.type || ''),
    device_no: printer?.device_no || '',
    api_url: printer?.api_url || '',
    feie_user: printer?.feie_user || '',
    feie_sn: printer?.feie_sn || '',
    print_types: parsePrinterPrintTypes(printer?.print_types),
    status: normalizeRequiredNumber(printer?.status, 1),
    auto_print: !!printer?.auto_print,
    is_default: !!printer?.is_default,
    print_count: normalizeRequiredNumber(printer?.print_count),
    last_print_at: printer?.last_print_at ? String(printer.last_print_at) : undefined,
    has_api_key: !!printer?.has_api_key,
    has_feie_ukey: !!printer?.has_feie_ukey,
    created_at: String(printer?.created_at || ''),
    updated_at: String(printer?.updated_at || '')
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

export function getMerchantProfitSharingRecords(params?: Omit<ProfitSharingRecordQuery, 'merchant_id'>) {
  return get<ProfitSharingRecordListResponse>('/api/v1/merchant/profit-sharing-records', params).then(normalizeListField)
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
export interface MerchantQrcodeResponse {
  qrcode_url: string
  scene: string
  page: string
  placeholder: boolean
  message: string
  expire_time?: string
}

export function getMerchantQrcode() {
  return get<MerchantQrcodeResponse>('/api/v1/merchant/qrcode')
}

export function getMerchantAnnouncements(params?: { page?: number; page_size?: number }) {
  return get<AnnouncementListResponse>('/api/v1/merchant/announcements', params).then(normalizeListField)
}

export function getMerchantFullReductionRules() {
  return get<MerchantFullReductionRulesResponse>('/api/v1/merchant/full-reduction-rules')
    .then(normalizeMerchantFullReductionRulesResponse)
}

export function updateMerchantFullReductionRules(data: UpdateMerchantFullReductionRulesRequest) {
  return put<MerchantFullReductionRulesResponse>('/api/v1/merchant/full-reduction-rules', data)
    .then(normalizeMerchantFullReductionRulesResponse)
}

export function getStoreFullReductionRules(merchantId: number) {
  return get<StoreFullReductionRulesResponse>(`/api/v1/store/${merchantId}/full-reduction-rules`)
    .then((response) => ({
      rules: normalizeArrayResponse(response?.rules).map(normalizeMerchantFullReductionRule)
    }))
}

// ============ 商品分类相关 ============

/**
 * 获取分类列表
 */
export function getCategories(options?: Partial<RequestOptions>) {
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

function normalizeNumberValue(value: number | string | null | undefined): number | undefined {
  if (value === '' || value === null || value === undefined) {
    return undefined
  }

  const normalizedValue = Number(value)
  if (!Number.isFinite(normalizedValue)) {
    return undefined
  }

  return normalizedValue
}

function normalizeRequiredNumber(value: number | string | null | undefined, fallback = 0): number {
  return normalizeNumberValue(value) ?? fallback
}

function normalizeSpecOption(option: ProductApiSpecOption): SpecOption {
  return {
    name: option?.name || '',
    price: normalizeRequiredNumber(option?.price),
    stock: normalizeNumberValue(option?.stock)
  }
}

function normalizeSpecs(value: ProductApiResponse['specs']): Product['specs'] {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((spec: ProductApiSpec) => ({
    id: normalizeNumberValue(spec?.id),
    name: spec?.name || '',
    options: Array.isArray(spec?.options)
      ? spec.options.map(normalizeSpecOption)
      : []
  }))
}

function normalizeProduct(product: ProductApiResponse | null | undefined): Product {
  return {
    ...product,
    id: normalizeRequiredNumber(product?.id),
    merchant_id: normalizeNumberValue(product?.merchant_id),
    category_id: normalizeRequiredNumber(product?.category_id),
    price: normalizeRequiredNumber(product?.price),
    original_price: normalizeNumberValue(product?.original_price),
    stock: normalizeRequiredNumber(product?.stock),
    sales: normalizeNumberValue(product?.sales),
    sort: normalizeNumberValue(product?.sort),
    images: normalizeStringArray(product?.images).map(normalizeImageUrl),
    specs: normalizeSpecs(product?.specs),
    created_at: String(product?.created_at || ''),
    updated_at: product?.updated_at ? String(product.updated_at) : undefined
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

  const deliveryInfo = order?.delivery_info
    ? {
        ...order.delivery_info,
        address: String(order.delivery_info?.address || order?.delivery_address || ''),
        contact_name: String(order.delivery_info?.contact_name || order?.contact_name || ''),
        contact_phone: String(order.delivery_info?.contact_phone || order?.contact_phone || ''),
        distance: Number(order.delivery_info?.distance || order?.delivery_distance || 0)
      }
    : {
        type: '',
        address: String(order?.delivery_address || ''),
        contact_name: String(order?.contact_name || ''),
        contact_phone: String(order?.contact_phone || ''),
        distance: Number(order?.delivery_distance || 0)
      }

  return {
    ...order,
    id: Number(order?.id || 0),
    items: Array.isArray(order?.items) ? order.items.map(normalizeOrderItem) : [],
    total_amount: Number(order?.total_amount || 0),
    delivery_fee: Number(order?.delivery_fee || 0),
    discount_amount: Number(order?.discount_amount || 0),
    pay_amount: Number(order?.pay_amount || 0),
    status: Number(order?.status || 0),
    delivery_address: String(order?.delivery_address || ''),
    contact_name: String(order?.contact_name || ''),
    contact_phone: String(order?.contact_phone || ''),
    verify_qrcode_url: String(order?.verify_qrcode_url || ''),
    delivery_info: deliveryInfo,
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
  const res = normalizeListField(await get<ProductListResponse>('/api/v1/merchant/products', params))
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
export async function createProduct(data: ProductUpsertPayload) {
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
export async function updateProduct(productId: number, data: ProductUpsertPayload) {
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
  return get<Order>(`/api/v1/merchant/orders/${orderId}`).then(normalizeOrder)
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

export function getMerchantPrinters() {
  return get<MerchantPrinter[] | null>('/api/v1/merchant/printers').then((response) => {
    return normalizeArrayResponse(response).map(normalizeMerchantPrinter)
  })
}

export function createMerchantPrinter(data: MerchantPrinterPayload) {
  return post<MerchantPrinter>('/api/v1/merchant/printers', data).then(normalizeMerchantPrinter)
}

export function updateMerchantPrinter(printerId: number, data: Partial<MerchantPrinterPayload>) {
  return put<MerchantPrinter>(`/api/v1/merchant/printers/${printerId}`, data).then(normalizeMerchantPrinter)
}

export function deleteMerchantPrinter(printerId: number) {
  return del<null>(`/api/v1/merchant/printers/${printerId}`)
}

export function testMerchantPrinter(printerId: number) {
  return post<{ success: boolean; message: string }>(`/api/v1/merchant/printers/${printerId}/test`)
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


export default {
  // 认证
  merchantLogin,
  merchantWechatLogin,
  getMerchantProfile,
  getMerchantProfitSharingRecords,
  // 商家设置
  getMerchantSettings,
  updateMerchantSettings,
  changeMerchantPassword,
  bindMerchantWechat,
  unbindMerchantWechat,
  getDeliverySettings,
  updateDeliverySettings,
  updateMerchantStatus,
  getMerchantQrcode,
  getMerchantFullReductionRules,
  updateMerchantFullReductionRules,
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
  // 文件上传
  getUploadToken,
  uploadImage,
  // C端店铺
  getStoreHome,
  getStoreProducts,
  getStoreProduct,
  getStoreDeliveryRules,
  getStoreFullReductionRules,
  trackStoreBehaviorEvent,
  // C端订单
  createOrder,
  getMyOrders,
  cancelMyOrder,
  applyRefund,
  // 云打印
  getPrintLogs,
  getMerchantPrinters,
  createMerchantPrinter,
  updateMerchantPrinter,
  deleteMerchantPrinter,
  testMerchantPrinter,
  // C端订单详情
  getMyOrderDetail,
  // C端地址管理
  getUserAddresses,
  createUserAddress,
  updateUserAddress,
  deleteUserAddress,
}
