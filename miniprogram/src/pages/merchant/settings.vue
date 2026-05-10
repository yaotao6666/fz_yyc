<template>
  <view class="settings-container">
    <!-- 商家信息 -->
    <view class="section">
      <view class="section-title">商家信息</view>
      
      <view class="setting-item" @click="goProfile">
        <view class="setting-left">
          <view class="setting-icon">
            <text>🏪</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">商家信息</view>
            <view class="setting-value">{{ merchantInfo?.name || '未设置' }}</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="setting-item" @click="goDeliverySettings">
        <view class="setting-left">
          <view class="setting-icon">
            <text>🚚</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">配送设置</view>
            <view class="setting-value">{{ deliveryEnabled ? '已开启' : '未开启' }}</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <!-- 账号安全 -->
    <view class="section">
      <view class="section-title">账号安全</view>
      
      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>🔑</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">修改密码</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>📱</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">绑定手机</view>
            <view class="setting-value">{{ merchantInfo?.contact_phone || '未绑定' }}</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <!-- 店铺运营 -->
    <view class="section">
      <view class="section-title">店铺运营</view>

      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>🔔</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">订单提示音</view>
            <view class="setting-value">{{ soundEnabled ? '已开启' : '未开启' }}</view>
          </view>
        </view>
        <switch :checked="soundEnabled" color="#007AFF" @change="handleSoundToggle" />
      </view>
      
      <view class="setting-item" @click="goQrcode">
        <view class="setting-left">
          <view class="setting-icon">
            <text>📱</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">店铺二维码</view>
            <view class="setting-value">推广店铺吸粉</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>📊</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">数据看板</view>
            <view class="setting-value">查看经营数据</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <!-- 其他 -->
    <view class="section">
      <view class="section-title">其他</view>
      
      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>📖</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">使用指南</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon">
            <text>ℹ️</text>
          </view>
          <view class="setting-info">
            <view class="setting-label">关于我们</view>
            <view class="setting-value">v1.0.0</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <!-- 退出登录 -->
    <view class="logout-area">
      <button class="btn-logout" @click="handleLogout">退出登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '../../stores/auth'
import { getMerchantSettings } from '@api'

const authStore = useAuthStore()

const merchantInfo = computed(() => authStore.merchantInfo)
const soundEnabled = computed(() => authStore.soundEnabled)
const deliveryEnabled = ref(false)

onShow(() => {
  loadDeliverySettings()
})

async function loadDeliverySettings() {
  try {
    const settings = await getMerchantSettings()
    deliveryEnabled.value = settings.delivery_settings?.enabled || false
  } catch (error) {
    console.error('加载配送设置失败:', error)
  }
}

function goProfile() {
  uni.navigateTo({ url: '/pages/merchant/settings' })
}

function goDeliverySettings() {
  uni.navigateTo({ url: '/pages/merchant/delivery-settings' })
}

function goQrcode() {
  uni.navigateTo({ url: '/pages/merchant/settings' })
}

function handleSoundToggle(e: any) {
  authStore.setSoundEnabled(!!e.detail.value)
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStore.logout()
      }
    }
  })
}
</script>

<style scoped>
.settings-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  padding-bottom: 200rpx;
}

.section {
  background: #ffffff;
  border-radius: 16rpx;
  margin-bottom: 24rpx;
  overflow: hidden;
}

.section-title {
  font-size: 26rpx;
  color: #999999;
  padding: 24rpx 32rpx 16rpx;
  background: #fafafa;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 32rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.setting-icon {
  width: 64rpx;
  height: 64rpx;
  background: #f0f5ff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
  font-size: 32rpx;
}

.setting-info {
  flex: 1;
}

.setting-label {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.setting-value {
  font-size: 26rpx;
  color: #999999;
}

.arrow {
  font-size: 32rpx;
  color: #cccccc;
}

.logout-area {
  margin-top: 48rpx;
  padding: 0 32rpx;
}

.btn-logout {
  width: 100%;
  height: 96rpx;
  background: #ffffff;
  border-radius: 48rpx;
  font-size: 32rpx;
  color: #ff4d4f;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
