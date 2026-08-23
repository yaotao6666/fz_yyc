export interface StoreEntryOptions {
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
  options?: Record<string, any>
): StoreEntryOptions {
  const normalizedOptions = options || {}
  const scene = typeof normalizedOptions.scene === 'string'
    ? decodeURIComponent(normalizedOptions.scene)
    : ''
  const sceneParams = parseStoreSceneParams(scene)
  const source = String(normalizedOptions.source || sceneParams.source || (scene ? 'scene' : 'scan'))

  return { source }
}

export function parseStoreProductEntryOptions(
  options?: Record<string, any>,
  fallbackProductId = 1
): StoreProductEntryOptions {
  const normalizedOptions = options || {}
  const { source } = parseStoreEntryOptions(normalizedOptions)
  const scene = typeof normalizedOptions.scene === 'string'
    ? decodeURIComponent(normalizedOptions.scene)
    : ''
  const sceneParams = parseStoreSceneParams(scene)
  const rawProductId = normalizedOptions.product_id || sceneParams.product_id || fallbackProductId || 1
  const productId = Number(rawProductId) || 1

  return { productId, source }
}
