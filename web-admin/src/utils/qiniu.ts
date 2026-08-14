import { getUploadToken } from '@/api/sp'

function joinQiniuFileUrl(domain: string, key: string): string {
  if (!domain) {
    return key
  }
  const normalizedDomain = domain.startsWith('http') ? domain : `https://${domain}`
  return `${normalizedDomain.replace(/\/+$/, '')}/${key.replace(/^\/+/, '')}`
}

export async function uploadSpImage(file: File): Promise<{ url: string; key: string }> {
  const uploadData = await getUploadToken()
  const ext = file.name.split('.').pop() || 'jpg'
  const key = `${uploadData.prefix}/${Date.now()}.${ext}`
  const formData = new FormData()
  formData.append('token', uploadData.token)
  formData.append('key', key)
  formData.append('file', file)

  const response = await fetch(uploadData.upload_url || 'https://up.qiniup.com', {
    method: 'POST',
    body: formData
  })

  if (!response.ok) {
    throw new Error('上传图片失败')
  }

  const result = await response.json()
  if (!result?.key) {
    throw new Error('上传图片失败')
  }

  return {
    key: result.key,
    url: joinQiniuFileUrl(uploadData.domain, result.key)
  }
}
