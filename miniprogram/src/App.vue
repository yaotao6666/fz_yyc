<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useAuthStore } from '@/stores/auth'

onLaunch(() => {
  console.log('App Launch')
  const authStore = useAuthStore()

  // 统一收口商家登录失效事件，避免 401 跳转时残留 WebSocket 连接。
  uni.$off('merchant-session-expired')
  uni.$on('merchant-session-expired', () => {
    authStore.handleSessionExpired()
  })

  // 检查登录状态
  authStore.checkLogin()
})

onShow(() => {
  console.log('App Show')
})

onHide(() => {
  console.log('App Hide')
})
</script>

<style lang="scss">
@import '@/uni.scss';

/* 全局样式 */

/* CSS变量 */
page {
  --primary-color: #007aff;
  --success-color: #07c160;
  --warning-color: #ff9500;
  --danger-color: #ff4d4f;
  --text-color: #333333;
  --text-color-secondary: #666666;
  --text-color-placeholder: #999999;
  --border-color: #eeeeee;
  --bg-color: #f5f5f5;
  --white: #ffffff;
}
</style>
