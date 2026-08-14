import { get, post } from '../utils/request'
import type {
  CreateOrderRequest,
  CreateOrderResponse,
  MerchantBehaviorEventRequest,
  Product,
  ProductApiResponse,
  ProductApiSpec,
  ProductApiSpecOption,
  ProductListResponse,
  ProductType,
  SpecOption,
  StoreDeliveryRules,
  StoreHomeInfo,
  WellnessPackageContent
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
    ...normalizeDeliverySettings(data)
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

function normalizeServiceContent(value: unknown): WellnessPackageContent | undefined {
  if (value === null || value === undefined || value === '') return undefined
  if (typeof value === 'string') {
    try {
      return JSON.parse(value) as WellnessPackageContent
    } catch {
      return undefined
    }
  }
  return value as WellnessPackageContent
}

function normalizeProduct(product: ProductApiResponse | null | undefined): Product {
  const pt = normalizeNumberValue(product?.product_type)
  const productType: ProductType = (pt === 1 || pt === 2 || pt === 3 || pt === 4 || pt === 5) ? pt : (Number(product?.sale_type) === 2 ? 2 : 1)
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
    product_type: productType,
    service_content: normalizeServiceContent(product?.service_content),
    sale_type: normalizeRequiredNumber(product?.sale_type, 1),
    rental_unit: normalizeRequiredNumber(product?.rental_unit, 0),
    rental_price: normalizeRequiredNumber(product?.rental_price),
    deposit: normalizeRequiredNumber(product?.deposit),
    max_rental_duration: normalizeRequiredNumber(product?.max_rental_duration, 0),
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
          sales: Number(normalized.sales || 0),
          product_type: normalized.product_type,
          sale_type: normalized.sale_type,
          rental_unit: normalized.rental_unit,
          rental_price: normalized.rental_price,
          deposit: normalized.deposit
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

export function createOrder(merchantId: number, data: CreateOrderRequest) {
  return post<CreateOrderResponse>(`/api/v1/store/${merchantId}/orders`, data)
}

export function trackStoreBehaviorEvent(merchantId: number, data: MerchantBehaviorEventRequest) {
  return post<{ message: string }>(`/api/v1/store/${merchantId}/event`, data, {
    loading: false,
    showErrorToast: false
  })
}
