<template>
  <view class="home-container">
    <!-- 顶部商家信息卡片 -->
    <view class="merchant-card">
      <view class="merchant-info">
        <image class="merchant-logo" :src="merchantInfo?.logo || BrandAsset.DEFAULT_MERCHANT_LOGO" mode="aspectFill" />
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

    <view v-if="announcementVisible" class="announcement-bar" @click="viewAnnouncement">
      <view class="announcement-left">
        <text class="announcement-icon">📢</text>
        <view class="announcement-marquee">
          <view class="announcement-text" :style="{ animationDuration: marqueeDuration + 's' }">
            {{ announcement?.title || '' }}
          </view>
        </view>
      </view>
      <view class="announcement-actions">
        <text class="announcement-view" @click.stop="viewAnnouncement">查看</text>
        <text class="announcement-close" @click.stop="closeAnnouncement">×</text>
      </view>
    </view>

    <view class="ws-status-section">
      <view class="section-title">商铺连接状态</view>
      <view class="ws-status-card">
        <view class="ws-status-row">
          <text class="ws-status-label">当前状态</text>
          <text class="ws-status-value" :class="socketStatusClass">{{ socketStatusText }}</text>
        </view>
        <view class="ws-status-row">
          <text class="ws-status-label">最近消息</text>
          <text class="ws-status-content">{{ lastSocketMessageText }}</text>
        </view>
        <view class="ws-status-row">
          <text class="ws-status-label">接收时间</text>
          <text class="ws-status-content">{{ authStore.lastSocketMessageAt || '暂无' }}</text>
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
        <view class="stat-card stat-card-clickable" @click="goOrders(OrderStatus.PAID)">
          <view class="stat-value">{{ statistics.pending_orders || 0 }}</view>
          <view class="stat-label">待核销订单</view>
          <view class="stat-subtitle">点击查看待核销订单</view>
        </view>
        <view class="stat-card stat-card-clickable" @click="goProducts('on_sale')">
          <view class="stat-value">{{ statistics.total_products || 0 }}</view>
          <view class="stat-label">上架商品数</view>
          <view class="stat-subtitle">点击查看已上架商品</view>
        </view>
      </view>
    </view>

    <!-- 快捷功能入口 -->
    <view class="menu-section">
      <view class="section-title">快捷功能</view>
      <view class="menu-grid">
        <view
          v-for="item in quickMenuItems"
          :key="item.title"
          class="menu-item"
          @click="item.action"
        >
          <view class="menu-icon" :style="{ background: item.background }">
            <image class="menu-icon-image" :src="item.icon" mode="aspectFit" />
          </view>
          <text class="menu-text">{{ item.title }}</text>
          <text class="menu-subtext">{{ item.subtitle }}</text>
        </view>
      </view>
    </view>

    <!-- 待处理事项 -->
    <view class="todo-section" v-if="hasPendingItems">
      <view class="section-title">待处理事项</view>
      <view class="todo-list">
        <view class="todo-item" v-if="statistics.pending_orders > 0" @click="goOrders(OrderStatus.PAID)">
          <view class="todo-left">
            <view class="todo-icon order"></view>
            <text class="todo-text">有待核销订单</text>
          </view>
          <view class="todo-right">
            <text class="todo-count">{{ statistics.pending_orders }}</text>
            <text class="arrow">›</text>
          </view>
        </view>
        <view class="todo-item" v-if="hasLowStock" @click="goProducts()">
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

    <view v-if="showQuickVerifyDialog" class="dialog-mask" @click="closeQuickVerifyDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">快速核销</view>
        <view class="dialog-subtitle">支持输入 6 位核销码，或扫一扫自动识别后直接核销</view>
        <view class="verify-input-section">
          <input
            v-model="quickVerifyCode"
            class="verify-input"
            type="text"
            maxlength="6"
            placeholder="请输入 6 位核销码"
          />
        </view>
        <view class="quick-verify-actions">
          <button class="scan-btn" :disabled="quickVerifying" @click="scanQuickVerifyCode">扫一扫</button>
          <button class="confirm-btn" :disabled="quickVerifying" @click="submitQuickVerify">
            {{ quickVerifying ? '核销中...' : '确认核销' }}
          </button>
        </view>
        <!-- <view class="dialog-cancel" @click="closeQuickVerifyDialog">取消</view> -->
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">

import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '../../stores/auth'
import { getMerchantProfile, updateMerchantStatus, getOrderStatistics, getProducts, getMerchantQrcode, getMerchantAnnouncements, getSalesOverview, quickCompleteOrder } from '@api'
import { OrderStatus } from '../../types'
import type { Announcement, Order, OrderStatistics } from '../../types'
import { BrandAsset } from '../../utils/constants'

const authStore = useAuthStore()
const ORDER_LIST_ROUTE_STATE_KEY = 'merchant_order_list_route_state'

const merchantInfo = ref(authStore.merchantInfo)
const merchantQrcode = ref<string>('')
const statistics = ref({
  today_orders: 0,
  today_sales: 0,
  pending_orders: 0,
  total_products: 0
})
const lowStockCount = ref(0)
const showQuickVerifyDialog = ref(false)
const quickVerifyCode = ref('')
const quickVerifying = ref(false)

const announcement = ref<Announcement | null>(null)
const announcementVisible = ref(false)
const marqueeDuration = computed(() => {
  const titleLength = announcement.value?.title?.length || 0
  return Math.max(8, Math.min(20, Math.ceil(titleLength / 6) * 4))
})

const hasPendingItems = computed(() => {
  return statistics.value.pending_orders > 0 || lowStockCount.value > 0
})

const hasLowStock = computed(() => lowStockCount.value > 0)
const socketStatusTextMap: Record<string, string> = {
  disconnected: '未连接',
  connecting: '连接中',
  connected: '已连接',
  reconnecting: '重连中'
}
const socketStatusText = computed(() => socketStatusTextMap[authStore.socketStatus] || '未连接')
const socketStatusClass = computed(() => `socket-${authStore.socketStatus}`)
const lastSocketMessageText = computed(() => {
  const message = authStore.lastSocketMessage
  if (!message) {
    return '暂无'
  }

  if (message.type === 'order_notify') {
    return `订单提醒，订单号：${message.orderNo || '未提供'}`
  }

  if (message.type === 'store_visit_notify') {
    return `顾客进店，访客：${message.visitorOpenId || '未提供'}，来源：${message.source || '未提供'}`
  }

  return message.type || '未知消息'
})
const quickMenuItems = computed(() => [
  {
    title: '分类管理',
    subtitle: '维护分类结构',
    icon: '/static/icons/category.png',
    background: '#fff7e6',
    action: () => goCategories()
  },
  {
    title: '商品管理',
    subtitle: '查看上架商品',
    icon: '/static/icons/product.png',
    background: '#e6f7ff',
    action: () => goProducts()
  },
  {
    title: '快速核销',
    subtitle: '扫码或输码核销',
    icon: '/static/icons/order.png',
    background: '#f6ffed',
    action: () => openQuickVerifyDialog()
  }
])

onShow(() => {
  loadData()
})

async function loadData() {
  await Promise.all([
    loadMerchantInfo(),
    loadStatistics(),
    loadLowStockCount(),
    loadMerchantQrcode(),
    loadAnnouncement()
  ])
}

async function loadAnnouncement() {
  try {
    const res = await getMerchantAnnouncements({ page: 1, page_size: 1 })
    const first = res?.list?.[0]
    if (!first) {
      announcement.value = null
      announcementVisible.value = false
      return
    }
    const merchantId = authStore.merchantId || 0
    const dismissKey = `merchant_home_announcement_dismissed_${merchantId}`
    const dismissedId = Number(uni.getStorageSync(dismissKey) || 0)

    announcement.value = first
    announcementVisible.value = Number(first.id) !== dismissedId
  } catch (error) {
    announcement.value = null
    announcementVisible.value = false
  }
}

function closeAnnouncement() {
  if (!announcement.value) return
  const merchantId = authStore.merchantId || 0
  const dismissKey = `merchant_home_announcement_dismissed_${merchantId}`
  uni.setStorageSync(dismissKey, announcement.value.id)
  announcementVisible.value = false
}

function viewAnnouncement() {
  if (!announcement.value) return
  uni.showModal({
    title: announcement.value.title,
    content: announcement.value.content || '',
    showCancel: false,
    confirmText: '知道了'
  })
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
    const [orderStats, overview, productRes] = await Promise.all([
      getOrderStatistics(),
      getSalesOverview({ period: 'today' }),
      getProducts({ status: '1', page: 1, page_size: 1 })
    ])

    const normalizedOrderStats = orderStats as unknown as Partial<OrderStatistics> & {
      today_orders?: number
      pending_orders?: number
      completed_orders?: number
    }

    statistics.value = {
      today_orders: overview.total_orders || normalizedOrderStats.today_orders || 0,
      today_sales: overview.total_sales || 0,
      // 首页待处理口径统一按待核销订单展示，避免与待支付订单混用。
      pending_orders: normalizedOrderStats.pending_orders || 0,
      total_products: productRes.total || 0
    }
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
  const confirmContent = newStatus === 1
    ? '开始营业后用户可继续下单，当前在线连接会继续保持。'
    : '休息后将停止接收新订单，但仍保持在线并继续接收顾客浏览提醒。'
  
  uni.showModal({
    title: '提示',
    content: confirmContent,
    success: async (res) => {
      if (res.confirm) {
        try {
          await updateMerchantStatus(newStatus)
          merchantInfo.value!.status = newStatus
          authStore.updateMerchantInfo({
            ...(authStore.merchantInfo || {}),
            ...merchantInfo.value!
          })
          const successMessage = newStatus === 1 ? '已开始营业' : '已进入休息，仍会接收浏览提醒'
          uni.showToast({ title: successMessage, icon: 'success' })
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

function goProducts(filterStatus?: 'on_sale' | 'off_sale') {
  const url = filterStatus
    ? `/pages/merchant/products/list?status=${filterStatus}`
    : '/pages/merchant/products/list'
  uni.navigateTo({ url })
}

function goCategories() {
  uni.navigateTo({ url: '/pages/merchant/categories' })
}

function goOrders(status?: number) {
  if (typeof status === 'number' && status > 0) {
    uni.setStorageSync(ORDER_LIST_ROUTE_STATE_KEY, JSON.stringify({ status }))
  } else {
    uni.removeStorageSync(ORDER_LIST_ROUTE_STATE_KEY)
  }
  uni.switchTab({ url: '/pages/merchant/orders/list' })
}

function openQuickVerifyDialog() {
  quickVerifyCode.value = ''
  showQuickVerifyDialog.value = true
}

function closeQuickVerifyDialog() {
  showQuickVerifyDialog.value = false
  quickVerifyCode.value = ''
}

function extractVerifyCode(rawValue: string): string {
  const content = String(rawValue || '').trim()
  if (/^\d{6}$/.test(content)) {
    return content
  }

  const queryMatch = content.match(/verify_code=(\d{6})/)
  if (queryMatch?.[1]) {
    return queryMatch[1]
  }

  const digitMatch = content.match(/(\d{6})/)
  return digitMatch?.[1] || ''
}

async function handleQuickVerifySuccess(order: Order) {
  await loadData()
  uni.showModal({
    title: '核销成功',
    content: `订单号：${order.order_no}\n核销人：${order.completed_by_name || '未知'}\n核销时间：${order.completed_at ? formatDateTime(order.completed_at) : '未知'}`,
    confirmText: '查看订单',
    cancelText: '关闭',
    success: (result) => {
      if (result.confirm) {
        uni.navigateTo({ url: `/pages/merchant/orders/detail?id=${order.id}` })
      }
    }
  })
}

async function submitQuickVerify() {
  const code = quickVerifyCode.value.trim()
  if (!/^\d{6}$/.test(code)) {
    uni.showToast({ title: '核销码应为6位数字', icon: 'none' })
    return
  }

  quickVerifying.value = true
  try {
    const order = await quickCompleteOrder(code)
    closeQuickVerifyDialog()
    await handleQuickVerifySuccess(order)
  } catch (error: any) {
    uni.showToast({ title: error?.message || '核销失败', icon: 'none' })
  } finally {
    quickVerifying.value = false
  }
}

function scanQuickVerifyCode() {
  uni.scanCode({
    success: async (result) => {
      const code = extractVerifyCode(result.result)
      if (!code) {
        uni.showToast({ title: '未识别到核销码', icon: 'none' })
        return
      }

      quickVerifyCode.value = code
      await submitQuickVerify()
    },
    fail: (error) => {
      const message = String((error as any)?.errMsg || '')
      if (message.includes('cancel')) {
        return
      }
      uni.showToast({ title: '扫码失败', icon: 'none' })
    }
  })
}

function formatDateTime(time: string): string {
  const date = new Date(time)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

</script>

<style scoped>
.home-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.announcement-bar {
  margin: 0 24rpx 24rpx;
  padding: 18rpx 20rpx;
  background: #ffffff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 4rpx 24rpx rgba(0, 0, 0, 0.04);
}

.announcement-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.announcement-icon {
  margin-right: 12rpx;
  font-size: 28rpx;
}

.announcement-marquee {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
}

.announcement-text {
  display: inline-block;
  padding-left: 100%;
  animation-name: marquee;
  animation-timing-function: linear;
  animation-iteration-count: infinite;
  font-size: 26rpx;
  color: #333333;
}

.announcement-actions {
  display: flex;
  align-items: center;
  margin-left: 16rpx;
}

.announcement-view {
  font-size: 26rpx;
  color: #007AFF;
  padding: 8rpx 12rpx;
}

.announcement-close {
  font-size: 34rpx;
  color: #999999;
  padding: 0 8rpx;
  line-height: 1;
}

@keyframes marquee {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-100%);
  }
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

.ws-status-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.ws-status-card {
  background: #f8faff;
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
}

.ws-status-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  padding: 12rpx 0;
}

.ws-status-row:not(:last-child) {
  border-bottom: 1rpx solid #eef2ff;
}

.ws-status-label {
  flex-shrink: 0;
  font-size: 26rpx;
  color: #666666;
}

.ws-status-value,
.ws-status-content {
  flex: 1;
  text-align: right;
  font-size: 26rpx;
  line-height: 1.6;
  color: #1a1a1a;
}

.socket-connected {
  color: #16a34a;
}

.socket-connecting,
.socket-reconnecting {
  color: #2563eb;
}

.socket-disconnected {
  color: #dc2626;
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

.stat-card-clickable {
  position: relative;
}

.stat-card-clickable:active {
  opacity: 0.75;
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

.stat-subtitle {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #b0b0b0;
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

.menu-icon-image {
  width: 48rpx;
  height: 48rpx;
}

.menu-text {
  font-size: 26rpx;
  color: #333333;
  font-weight: 600;
}

.menu-subtext {
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #999999;
  line-height: 1.5;
  text-align: center;
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

.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32rpx;
  z-index: 1000;
}

.dialog-content {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.dialog-subtitle {
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.7;
  color: #666666;
}

.verify-input-section {
  margin-top: 28rpx;
}

.verify-input {
  height: 88rpx;
  border-radius: 16rpx;
  background: #f8f9fa;
  padding: 0 24rpx;
  font-size: 30rpx;
}

.quick-verify-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}

.scan-btn,
.confirm-btn {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.scan-btn {
  background: #f0f5ff;
  color: #0056CC;
}

.confirm-btn {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}

.dialog-cancel {
  margin-top: 24rpx;
  text-align: center;
  font-size: 28rpx;
  color: #999999;
}
</style>
