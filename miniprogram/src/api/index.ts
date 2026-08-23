/**
 * API接口封装 - 基于PRD文档
 * 统一管理所有API接口调用
 */

import { get, post, put, del } from '../utils/request'
import type { RequestOptions } from '../utils/request'
import type {
  UploadTokenResponse,
  DeliverySettings,
  StoreDeliveryRules,
  Product,
  ProductApiResponse,
  ProductApiSpec,
  ProductApiSpecOption,
  ProductListResponse,
  SpecOption,
  Order,
  OrderListResponse,
  StoreHomeInfo,
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantBehaviorEventRequest,
  UserAddress,
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
    ...normalizeDeliverySettings(data)
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

// ============ 商品相关通用辅助函数 ============

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
    delivery_info: deliveryInfo,
    merchant
  } as Order
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
export function getStoreHome() {
  return get<StoreHomeInfo>('/api/v1/store/home').then(data => {
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
export function getStoreProducts(params?: { category_id?: number }) {
  return get<ProductListResponse>('/api/v1/store/products', params).then(res => {
    const normalized = normalizeListField(res)
    return { ...normalized, list: normalized.list.map(normalizeProduct) }
  })
}

/**
 * 获取店铺商品详情
 */
export function getStoreProduct(productId: number) {
  return get<Product>(`/api/v1/store/products/${productId}`).then(normalizeProduct)
}

/**
 * 获取配送费规则
 */
export function getStoreDeliveryRules() {
  return get<StoreDeliveryRules>('/api/v1/store/delivery-rules').then(normalizeStoreDeliveryRules)
}

// ============ C端订单相关 ============

/**
 * 创建订单
 */
export function createOrder(data: CreateOrderRequest) {
  return post<CreateOrderResponse>('/api/v1/store/orders', data)
}

/**
 * 上报店铺行为事件
 */
export function trackStoreBehaviorEvent(data: MerchantBehaviorEventRequest) {
  return post<{ message: string }>('/api/v1/store/event', data, { loading: false, showErrorToast: false })
}

/**
 * 获取我的订单列表
 * @param params.page 页码
 * @param params.page_size 每页数量
 * @param params.status 订单状态
 */
export function getMyOrders(params?: {
  page?: number
  page_size?: number
  status?: number
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
  // C端订单详情
  getMyOrderDetail,
  // C端地址管理
  getUserAddresses,
  createUserAddress,
  updateUserAddress,
  deleteUserAddress,
}
