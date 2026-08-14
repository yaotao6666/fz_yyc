<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'

onLaunch(() => {
  console.log('[Staff] App Launch')
  // 登录拦截：无 token 时跳转登录页
  const token = uni.getStorageSync('staff_token')
  if (!token) {
    uni.reLaunch({ url: '/pages/login/index' })
  }
})

onShow(() => {
  console.log('[Staff] App Show')
})

onHide(() => {
  console.log('[Staff] App Hide')
})
</script>

<style lang="scss">
/* 主题色: 517528 财旭绿 */
page {
  --primary-color: #517528;
  --primary-color-light: #7a9e4f;
  --primary-color-dark: #3a551c;
  --success-color: #07c160;
  --warning-color: #ff9500;
  --danger-color: #ff4d4f;
  --info-color: #1989fa;
  --text-color: #333333;
  --text-color-secondary: #666666;
  --text-color-placeholder: #999999;
  --border-color: #eeeeee;
  --bg-color: #f5f5f5;
  --white: #ffffff;
  --card-radius: 16rpx;
  --shadow-sm: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4rpx 20rpx rgba(0, 0, 0, 0.08);
}

page {
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 28rpx;
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  line-height: 1.5;
}

.container {
  padding: 24rpx;
  box-sizing: border-box;
}

.card {
  background: var(--white);
  border-radius: var(--card-radius);
  padding: 32rpx;
  box-shadow: var(--shadow-sm);
  margin-bottom: 24rpx;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  color: #fff;
  border-radius: 48rpx;
  padding: 20rpx 48rpx;
  font-size: 30rpx;
  font-weight: 500;
  border: none;
  line-height: 1.4;
}

.btn-primary[disabled] {
  opacity: 0.5;
}

.btn-outline {
  background: #fff;
  color: var(--primary-color);
  border: 2rpx solid var(--primary-color);
  border-radius: 48rpx;
  padding: 18rpx 46rpx;
  font-size: 30rpx;
}
</style>
