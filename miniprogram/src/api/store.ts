import { get, post } from '../utils/request'
import type {
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantFullReductionRule,
  MerchantBehaviorEventRequest,
  Product,
  ProductApiResponse,
  ProductApiSpec,
  ProductApiSpecOption,
  ProductListResponse,
  PickupPoint,
  SpecOption,
  StoreDeliveryRules,
  StoreFullReductionRulesResponse,
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
    id: normalizeNumberValue(option?.id),
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

function normalizeListField<T, R extends { list?: T[] | null }>(response: R): R & { list: T[] } {
  return {
    ...response,
    list: Array.isArray(response?.list) ? response.list : []
  }
}

export function getStoreHome(merchantId: number) {
  return get<StoreHomeInfo>(`/api/v1/store/${merchantId}/home`).then(data => {
    if (data?.hot_products) {
      data.hot_products = data.hot_products.map((item: any) => {
        const normalized = normalizeProduct(item)
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

export function getStorePickupPoints(merchantId: number) {
  return get<PickupPoint[]>(`/api/v1/store/${merchantId}/pickup-points`).then((list) => (
    Array.isArray(list) ? list : []
  ))
}

export function getStoreFullReductionRules(merchantId: number) {
  return get<StoreFullReductionRulesResponse>(`/api/v1/store/${merchantId}/full-reduction-rules`).then((response) => ({
    rules: Array.isArray(response?.rules) ? response.rules.map(normalizeMerchantFullReductionRule) : []
  }))
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
