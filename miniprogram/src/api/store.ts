import { get, post } from '../utils/request'
import type {
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantBehaviorEventRequest,
  Product,
  ProductListResponse,
  StoreDeliveryRules,
  StoreHomeInfo
} from '../types'

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

function normalizeDeliverySettings(data: any) {
  return {
    enabled: !!data?.enabled,
    base_fee: Number(data?.base_fee || 0),
    free_delivery_amount: Number(data?.free_delivery_amount || 0),
    max_distance: Number(data?.max_distance || 10),
    distance_rules: parseDistanceRules(data?.distance_rules)
  }
}

function normalizeStoreDeliveryRules(data: any): StoreDeliveryRules {
  return {
    ...normalizeDeliverySettings(data),
    takeout_enabled: !!data?.takeout_enabled,
    dine_in_enabled: !!data?.dine_in_enabled,
    pickup_enabled: !!data?.pickup_enabled
  }
}

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

function normalizeListField<T, R extends { list?: T[] | null }>(response: R): R & { list: T[] } {
  return {
    ...response,
    list: Array.isArray(response?.list) ? response.list : []
  }
}

export function getStoreHome(merchantId: number) {
  return get<StoreHomeInfo>(`/api/v1/store/${merchantId}/home`).then(data => {
    if (data?.hot_products) {
      data.hot_products = data.hot_products.map((item: any) => normalizeProduct(item))
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

export function getStoreProducts(merchantId: number, params?: { category_id?: number }) {
  return get<ProductListResponse>(`/api/v1/store/${merchantId}/products`, params).then(response => {
    const normalized = normalizeListField(response)
    return { ...normalized, list: normalized.list.map(normalizeProduct) }
  })
}

export function getStoreProduct(merchantId: number, productId: number) {
  return get<Product>(`/api/v1/store/${merchantId}/products/${productId}`).then(normalizeProduct)
}

export function getStoreDeliveryRules(merchantId: number) {
  return get<StoreDeliveryRules>(`/api/v1/store/${merchantId}/delivery-rules`).then(normalizeStoreDeliveryRules)
}

export function createOrder(merchantId: number, data: CreateOrderRequest) {
  return post<CreateOrderResponse>(`/api/v1/store/${merchantId}/orders`, data)
}

export function trackStoreBehaviorEvent(merchantId: number, data: MerchantBehaviorEventRequest) {
  return post<{ message: string }>(`/api/v1/store/${merchantId}/event`, data, {
    loading: false,
    showErrorToast: false
  })
}
