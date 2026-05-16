const CACHE_KEY_PREFIX = 'img_cache_'

function getCacheKey(url: string): string {
  let hash = 0
  const base = url.split('?')[0]
  for (let i = 0; i < base.length; i++) {
    hash = ((hash << 5) - hash + base.charCodeAt(i)) & 0x7fffffff
  }
  return CACHE_KEY_PREFIX + hash.toString(36)
}

export function getCachedImagePath(url: string): string {
  if (!url) return ''
  const key = getCacheKey(url)
  return uni.getStorageSync(key) || ''
}

export async function cacheImage(url: string): Promise<string> {
  if (!url) return ''

  const cached = getCachedImagePath(url)
  if (cached) return cached

  return new Promise((resolve) => {
    uni.getImageInfo({
      src: url,
      success: (res) => {
        if (res.path) {
          const key = getCacheKey(url)
          uni.setStorageSync(key, res.path)
          resolve(res.path)
        } else {
          resolve(url)
        }
      },
      fail: () => {
        resolve(url)
      }
    })
  })
}

export function clearImageCache() {
  const res = uni.getStorageInfoSync()
  const keys = res.keys || []
  for (const key of keys) {
    if (key.startsWith(CACHE_KEY_PREFIX)) {
      uni.removeStorageSync(key)
    }
  }
}
