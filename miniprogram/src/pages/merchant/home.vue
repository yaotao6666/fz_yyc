<template>
  <view class="home-container">
    <!-- 顶部商家信息卡片 -->
    <view class="merchant-card">
      <view class="merchant-info">
        <image class="merchant-logo" :src="merchantInfo?.logo || '/static/default-logo.png'" mode="aspectFill" />
        <view class="merchant-detail">
          <view class="merchant-name">{{ merchantInfo?.name || '加载中...' }}</view>
          <view class="merchant-status">
            <text class="status-dot" :class="{ active: merchantInfo?.status === 1 }"></text>
            {{ merchantInfo?.status === 1 ? '营业中' : '休息中' }}
          </view>
        </view>
      </view>
      <view class="quick-actions">
        <view class="action-item" @click="toggleShopStatus">
          <image class="action-icon" src="/static/icons/store.png" />
          <text class="action-text">{{ merchantInfo?.status === 1 ? '暂停营业' : '开始营业' }}</text>
        </view>
        <view class="action-item" @click="showQrcode">
          <image class="action-icon" src="/static/icons/qrcode.png" />
          <text class="action-text">店铺二维码</text>
        </view>
      </view>
    </view>

    <!-- 今日数据概览 -->
    <view class="stats-section">
      <view class="section-title">今日概览</view>
      <view class="stats-grid">
        <view class="stat-card">
          <view class="stat-value">{{ statistics.today_orders || 0 }}</view>
          <view class="stat-label">今日订单</view>
        </view>
        <view class="stat-card">
          <view class="stat-value">¥{{ formatAmount(statistics.today_sales || 0) }}</view>
          <view class="stat-label">今日销售额</view>
        </view>
        <view class="stat-card">
          <view class="stat-value">{{ statistics.pending_orders || 0 }}</view>
          <view class="stat-label">待处理订单</view>
        </view>
        <view class="stat-card">
          <view class="stat-value">{{ statistics.total_products || 0 }}</view>
          <view class="stat-label">商品数量</view>
        </view>
      </view>
    </view>

    <!-- 快捷功能入口 -->
    <view class="menu-section">
      <view class="section-title">快捷功能</view>
      <view class="menu-grid">
        <view class="menu-item" @click="goProducts">
          <view class="menu-icon" style="background: #e6f7ff;">
            <image src="/static/icons/product.png" />
          </view>
          <text class="menu-text">商品管理</text>
        </view>
        <view class="menu-item" @click="goCategories">
          <view class="menu-icon" style="background: #fff7e6;">
            <image src="/static/icons/category.png" />
          </view>
          <text class="menu-text">分类管理</text>
        </view>
        <view class="menu-item" @click.stop="goOrders('paid')">
          <view class="menu-icon" style="background: #f6ffed;">
            <image src="/static/icons/order.png" />
          </view>
          <text class="menu-text">订单管理</text>
        </view>
        <view class="menu-item" @click="goAnalytics">
          <view class="menu-icon" style="background: #fff1f0;">
            <image src="/static/icons/analytics.png" />
          </view>
          <text class="menu-text">数据分析</text>
        </view>
        <view class="menu-item" @click="goInvite">
          <view class="menu-icon" style="background: #f9f0ff;">
            <image src="/static/icons/invite.png" />
          </view>
          <text class="menu-text">邀请入驻</text>
        </view>
        <view class="menu-item" @click="goSettings">
          <view class="menu-icon" style="background: #f0f5ff;">
            <image src="/static/icons/settings.png" />
          </view>
          <text class="menu-text">店铺设置</text>
        </view>
      </view>
    </view>

    <!-- 待处理事项 -->
    <view class="todo-section" v-if="hasPendingItems">
      <view class="section-title">待处理事项</view>
      <view class="todo-list">
        <view class="todo-item" v-if="statistics.pending_orders > 0" @click="goOrders('paid')">
          <view class="todo-left">
            <view class="todo-icon order"></view>
            <text class="todo-text">有待处理订单</text>
          </view>
          <view class="todo-right">
            <text class="todo-count">{{ statistics.pending_orders }}</text>
            <text class="arrow">›</text>
          </view>
        </view>
        <view class="todo-item" v-if="hasLowStock" @click="goProducts">
          <view class="todo-left">
            <view class="todo-icon stock"></view>
            <text class="todo-text">库存预警</text>
          </view>
          <view class="todo-right">
            <text class="todo-count">{{ lowStockCount }}</text>
            <text class="arrow">›</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部tabbar占位 -->
    <view class="tabbar-placeholder"></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '../../stores/auth'
import { getMerchantProfile, updateMerchantStatus, getOrderStatistics, getProducts, getCategories, getMerchantQrcode } from '../../api'

const authStore = useAuthStore()

const merchantInfo = ref(authStore.merchantInfo)
const merchantQrcode = ref<string>('')
const statistics = ref<any>({
  today_orders: 0,
  today_sales: 0,
  pending_orders: 0,
  total_products: 0
})
const lowStockCount = ref(0)

const hasPendingItems = computed(() => {
  return statistics.value.pending_orders > 0 || lowStockCount.value > 0
})

const hasLowStock = computed(() => lowStockCount.value > 0)

onShow(() => {
  loadData()
})

async function loadData() {
  await Promise.all([
    loadMerchantInfo(),
    loadStatistics(),
    loadLowStockCount(),
    loadMerchantQrcode()
  ])
}

async function loadMerchantQrcode() {
  try {
    const res = await getMerchantQrcode()
    merchantQrcode.value = res.qrcode_url
  } catch (error) {
    console.error('加载二维码失败:', error)
  }
}

async function loadMerchantInfo() {
  try {
    const info = await getMerchantProfile()
    merchantInfo.value = info
    authStore.updateMerchantInfo(info)
  } catch (error) {
    console.error('加载商家信息失败:', error)
  }
}

async function loadStatistics() {
  try {
    // 获取订单统计
    const orderStats = await getOrderStatistics()
    
    statistics.value = {
      today_orders: orderStats.pending_payment + orderStats.completed,
      today_sales: 0, // 需要从今日统计中获取
      pending_orders: orderStats.pending_payment + orderStats.pending_complete,
      total_products: 0
    }

    // 获取商品数量
    const productRes = await getProducts({ page: 1, page_size: 1 })
    statistics.value.total_products = productRes.total
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

async function loadLowStockCount() {
  try {
    const res = await getProducts({ status: 'on_sale', page: 1, page_size: 100 })
    const lowStockProducts = res.list.filter((p: any) => p.stock <= 10)
    lowStockCount.value = lowStockProducts.length
  } catch (error) {
    console.error('加载库存预警失败:', error)
  }
}

async function toggleShopStatus() {
  const newStatus = merchantInfo.value?.status === 1 ? 0 : 1
  const actionText = newStatus === 1 ? '营业' : '暂停'
  
  uni.showModal({
    title: '提示',
    content: `确定要${actionText}吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await updateMerchantStatus(newStatus)
          merchantInfo.value!.status = newStatus
          uni.showToast({ title: `${actionText}成功`, icon: 'success' })
        } catch (error: any) {
          uni.showToast({ title: error.message || '操作失败', icon: 'none' })
        }
      }
    }
  })
}

function showQrcode() {
  if (!merchantQrcode.value) {
    uni.showToast({ title: '二维码加载中，请稍后', icon: 'none' })
    return
  }
  uni.previewImage({
    urls: [merchantQrcode.value],
    current: 0
  })
}

function formatAmount(amount: number): string {
  return amount.toFixed(2)
}

function goProducts() {
  uni.navigateTo({ url: '/pages/merchant/products/list' })
}

function goCategories() {
  uni.navigateTo({ url: '/pages/merchant/categories' })
}

function goOrders(status?: string) {
  const url = status ? `/pages/merchant/orders/list?status=${status}` : '/pages/merchant/orders/list'
  uni.navigateTo({ url })
}

function goAnalytics() {
  uni.navigateTo({ url: '/pages/merchant/analytics/index' })
}

function goInvite() {
  uni.navigateTo({ url: '/pages/merchant/invite' })
}

function goSettings() {
  uni.switchTab({ url: '/pages/merchant/settings' })
}
</script>

<style scoped>
.home-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.merchant-card {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  padding: 32rpx;
  margin: 24rpx;
  border-radius: 24rpx;
  color: #ffffff;
}

.merchant-info {
  display: flex;
  align-items: center;
  margin-bottom: 32rpx;
}

.merchant-logo {
  width: 100rpx;
  height: 100rpx;
  border-radius: 20rpx;
  background: #ffffff;
  margin-right: 24rpx;
}

.merchant-detail {
  flex: 1;
}

.merchant-name {
  font-size: 36rpx;
  font-weight: 600;
  margin-bottom: 8rpx;
}

.merchant-status {
  display: flex;
  align-items: center;
  font-size: 26rpx;
  opacity: 0.9;
}

.status-dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: #999999;
  margin-right: 8rpx;
}

.status-dot.active {
  background: #52c41a;
}

.quick-actions {
  display: flex;
  gap: 24rpx;
}

.action-item {
  flex: 1;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.action-icon {
  width: 48rpx;
  height: 48rpx;
  margin-bottom: 12rpx;
}

.action-text {
  font-size: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.stats-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24rpx;
}

.stat-card {
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 24rpx;
  text-align: center;
}

.stat-value {
  font-size: 40rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
}

.menu-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 32rpx;
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: transparent;
  padding: 0;
  margin: 0;
  width: 100%;
  min-height: 160rpx;
}

.menu-item:active {
  opacity: 0.7;
}

.menu-icon {
  width: 96rpx;
  height: 96rpx;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16rpx;
}

.menu-icon image {
  width: 48rpx;
  height: 48rpx;
}

.menu-text {
  font-size: 26rpx;
  color: #666666;
}

.todo-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.todo-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.todo-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
}

.todo-left {
  display: flex;
  align-items: center;
}

.todo-icon {
  width: 48rpx;
  height: 48rpx;
  border-radius: 12rpx;
  margin-right: 16rpx;
}

.todo-icon.order {
  background: #fff7e6;
}

.todo-icon.stock {
  background: #fff1f0;
}

.todo-text {
  font-size: 28rpx;
  color: #333333;
}

.todo-right {
  display: flex;
  align-items: center;
}

.todo-count {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 600;
  margin-right: 8rpx;
}

.arrow {
  font-size: 32rpx;
  color: #cccccc;
}

.tabbar-placeholder {
  height: 120rpx;
}
</style>
