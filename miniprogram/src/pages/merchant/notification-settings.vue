<template>
  <view class="notification-container">
    <view class="section">
      <view class="section-title">声音提醒开关</view>

      <view class="setting-item">
        <view class="setting-info">
          <view class="setting-label">用户下单提醒</view>
          <view class="setting-value">收到新订单时播放提醒音</view>
        </view>
        <switch :checked="orderEnabled" :disabled="saving" color="#007AFF" @change="toggleOrderNotify" />
      </view>

      <view class="setting-item">
        <view class="setting-info">
          <view class="setting-label">用户浏览提醒</view>
          <view class="setting-value">顾客进入店铺首页时播放提醒音</view>
        </view>
        <switch :checked="browseEnabled" :disabled="saving" color="#007AFF" @change="toggleBrowseNotify" />
      </view>
    </view>

    <view class="section">
      <view class="section-title">测试播放</view>
      <button class="test-btn" @click="handleTestOrderSound">测试下单提醒</button>
      <button class="test-btn secondary" @click="handleTestBrowseSound">测试浏览提醒</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMerchantSettings, updateMerchantSettings } from '@api'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()

const orderEnabled = ref(true)
const browseEnabled = ref(true)
const saving = ref(false)

onShow(() => {
  loadSettings()
})

async function loadSettings() {
  try {
    const settings = await getMerchantSettings()
    orderEnabled.value = settings.notify_enabled ?? true
    browseEnabled.value = settings.browse_notify_enabled ?? true
    authStore.setOrderSoundEnabled(orderEnabled.value)
    authStore.setBrowseSoundEnabled(browseEnabled.value)
  } catch (error) {
    console.error('加载声音提醒设置失败:', error)
  }
}

async function saveSettings(payload: { notify_enabled?: boolean; browse_notify_enabled?: boolean }) {
  if (saving.value) {
    return
  }

  saving.value = true
  try {
    await updateMerchantSettings(payload)
    if (typeof payload.notify_enabled === 'boolean') {
      authStore.setOrderSoundEnabled(payload.notify_enabled)
      if (authStore.staff) {
        authStore.updateStaffInfo({ ...authStore.staff, notify_enabled: payload.notify_enabled })
      }
    }
    if (typeof payload.browse_notify_enabled === 'boolean') {
      authStore.setBrowseSoundEnabled(payload.browse_notify_enabled)
      if (authStore.staff) {
        authStore.updateStaffInfo({ ...authStore.staff, browse_notify_enabled: payload.browse_notify_enabled })
      }
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error?.message || '保存失败', icon: 'none' })
    await loadSettings()
  } finally {
    saving.value = false
  }
}

function toggleOrderNotify(event: any) {
  orderEnabled.value = !!event.detail.value
  saveSettings({ notify_enabled: orderEnabled.value })
}

function toggleBrowseNotify(event: any) {
  browseEnabled.value = !!event.detail.value
  saveSettings({ browse_notify_enabled: browseEnabled.value })
}

function handleTestOrderSound() {
  authStore.testPlayOrderSound()
}

function handleTestBrowseSound() {
  authStore.testPlayBrowseSound()
}
</script>

<style scoped>
.notification-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}

.section {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 32rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20rpx;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
  padding-right: 24rpx;
}

.setting-label {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.setting-value {
  font-size: 24rpx;
  color: #999999;
}

.test-btn {
  width: 100%;
  height: 88rpx;
  border-radius: 44rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 20rpx;
}

.test-btn.secondary {
  background: #f0f5ff;
  color: #0056CC;
}
</style>
