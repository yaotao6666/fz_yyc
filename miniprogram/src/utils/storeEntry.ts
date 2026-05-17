export interface StoreEntryOptions {
  merchantId: number
  source: string
}

export interface StoreProductEntryOptions extends StoreEntryOptions {
  productId: number
}

export function parseStoreSceneParams(scene: string) {
  const params: Record<string, string> = {}
  scene.split('&').forEach((segment) => {
    if (!segment) return
    const [rawKey, rawValue = ''] = segment.split('=')
    if (!rawKey) return
    params[rawKey] = rawValue
  })
  return params
}

export function parseStoreEntryOptions(
  options?: Record<string, any>,
  fallbackMerchantId = 1
): StoreEntryOptions {
  const normalizedOptions = options || {}
  const scene = typeof normalizedOptions.scene === 'string'
    ? decodeURIComponent(normalizedOptions.scene)
    : ''
  const sceneParams = parseStoreSceneParams(scene)
  const rawMerchantId = normalizedOptions.merchant_id || sceneParams.merchant_id || fallbackMerchantId || 1
  const merchantId = Number(rawMerchantId) || 1
  const source = String(normalizedOptions.source || sceneParams.source || (scene ? 'scene' : 'scan'))

  return { merchantId, source }
}

export function parseStoreProductEntryOptions(
  options?: Record<string, any>,
  fallbackMerchantId = 1,
  fallbackProductId = 1
): StoreProductEntryOptions {
  const normalizedOptions = options || {}
  const { merchantId, source } = parseStoreEntryOptions(normalizedOptions, fallbackMerchantId)
  const scene = typeof normalizedOptions.scene === 'string'
    ? decodeURIComponent(normalizedOptions.scene)
    : ''
  const sceneParams = parseStoreSceneParams(scene)
  const rawProductId = normalizedOptions.product_id || sceneParams.product_id || fallbackProductId || 1
  const productId = Number(rawProductId) || 1

  return { merchantId, productId, source }
}
