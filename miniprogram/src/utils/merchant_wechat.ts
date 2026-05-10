export async function getMerchantWechatCode() {
  return new Promise<string>((resolve, reject) => {
    uni.login({
      provider: 'weixin',
      success: (result) => {
        if (result.code) {
          resolve(result.code)
          return
        }
        reject(new Error('未获取到微信登录凭证'))
      },
      fail: (error) => {
        console.error('获取微信登录凭证失败:', error)
        reject(new Error('获取微信登录凭证失败'))
      }
    })
  })
}
