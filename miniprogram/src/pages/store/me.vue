<template>
  <view class="me-container">
    <!-- 用户信息卡片 -->
    <view class="profile-card">
      <image class="avatar" :src="user?.avatar || BrandAsset.DEFAULT_MERCHANT_LOGO" mode="aspectFill" />
      <view class="profile-info">
        <view class="nickname">{{ user?.nickname || '微信用户' }}</view>
        <view class="sub">{{ user?.phone ? maskPhone(user.phone) : '点击个人服务查看订单与健康记录' }}</view>
      </view>
    </view>

    <!-- 核心入口 -->
    <view class="feature-grid">
      <view class="feature-item" v-for="feature in features" :key="feature.title" @click="onFeature(feature.key)">
        <view class="feature-icon" :class="feature.theme">{{ feature.icon }}</view>
        <view class="feature-title">{{ feature.title }}</view>
        <view class="feature-desc">{{ feature.desc }}</view>
      </view>
    </view>

    <!-- 功能列表 -->
    <view class="menu-card">
      <view class="menu-row" v-for="item in menuItems" :key="item.title" @click="onMenu(item.key)">
        <view class="menu-left">
          <text class="menu-icon">{{ item.icon }}</text>
          <text class="menu-title">{{ item.title }}</text>
        </view>
        <text class="menu-arrow">›</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuth } from '../../utils/useAuth'
import { getStoreHome } from '../../api/store'
import { BrandAsset } from '../../utils/constants'

const user = ref<any>(null)
const merchantPhone = ref('')

const features = [
  { key: 'orders', icon: '📋', theme: 'blue', title: '我的订单', desc: '订单与租赁/服务进度' },
  { key: 'health', icon: '🩺', theme: 'green', title: '我的健康', desc: '档案与健康服务' }
]

const menuItems = [
  { key: 'orders', icon: '📦', title: '我的订单' },
  { key: 'health', icon: '❤️', title: '我的健康' },
  { key: 'address', icon: '📍', title: '收货地址' },
  { key: 'contact', icon: '📞', title: '联系商家' }
]

onShow(() => {
  const { getUserInfo, ensureAuth } = useAuth()
  user.value = getUserInfo()
  void ensureAuth().then((authed) => {
    if (authed) {
      user.value = useAuth().getUserInfo()
    }
  })
  loadMerchantPhone()
})

async function loadMerchantPhone() {
  try {
    const home = await getStoreHome()
    merchantPhone.value = String((home.merchant?.contact_phone || '') || '').trim()
  } catch (error) {
    // 商家电话获取失败不阻塞页面
  }
}

function maskPhone(phone: string): string {
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2')
}

function onFeature(key: string) {
  onMenu(key)
}

function onMenu(key: string) {
  if (key === 'orders') {
    uni.navigateTo({ url: '/pages/store/my-orders' })
  } else if (key === 'health') {
    uni.navigateTo({ url: '/pages/store/my-health' })
  } else if (key === 'address') {
    uni.navigateTo({ url: '/pages/store/address-list' })
  } else if (key === 'contact') {
    callMerchant()
  }
}

function callMerchant() {
  if (!merchantPhone.value) {
    uni.showToast({ title: '暂无商家联系电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({
    phoneNumber: merchantPhone.value,
    fail: () => uni.showToast({ title: '拨打失败', icon: 'none' })
  })
}
</script>

<style scoped>
.me-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

.profile-card {
  display: flex;
  align-items: center;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 36rpx 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: #ffffff;
  border: 4rpx solid rgba(255, 255, 255, 0.4);
  margin-right: 24rpx;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.nickname {
  font-size: 38rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.sub {
  font-size: 24rpx;
  opacity: 0.9;
}

.feature-grid {
  display: flex;
  gap: 20rpx;
  margin-bottom: 24rpx;
}

.feature-item {
  flex: 1;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx 16rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.feature-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 42rpx;
  margin-bottom: 14rpx;
}

.feature-icon.blue {
  background: rgba(0, 122, 255, 0.1);
}

.feature-icon.green {
  background: rgba(22, 163, 74, 0.1);
}

.feature-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.feature-desc {
  font-size: 20rpx;
  color: #999999;
}

.menu-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 0 32rpx;
}

.menu-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 30rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.menu-row:last-child {
  border-bottom: none;
}

.menu-left {
  display: flex;
  align-items: center;
}

.menu-icon {
  font-size: 32rpx;
  margin-right: 20rpx;
}

.menu-title {
  font-size: 28rpx;
  color: #1a1a1a;
}

.menu-arrow {
  font-size: 40rpx;
  color: #c0c4cc;
}
</style>