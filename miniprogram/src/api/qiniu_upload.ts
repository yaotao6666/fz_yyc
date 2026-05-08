export function uploadImage(filePath: string) {
  return new Promise(async (resolve, reject) => {
    uni.showLoading({ title: '上传中...', mask: true })
    
    try {
      const uploadData = await getUploadToken()
      
      const ext = filePath.split('.').pop() || 'jpg'
      const key = `${uploadData.prefix}/${Date.now()}.${ext}`
      
      uni.uploadFile({
        url: `https://upload.qiniup.com`,
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
                url: `${uploadData.domain}/${data.key}`,
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
    } catch (error) {
      uni.hideLoading()
      uni.showToast({ title: '获取上传凭证失败', icon: 'none' })
      reject(error)
    }
  })
}
