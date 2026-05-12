<template>
  <view class="sp-home-container">
    <!-- 页面标题 -->
    <view class="page-header">
      <view class="header-title">服务商管理中心</view>
      <view class="header-subtitle">{{ providerInfo?.name || '服务商' }}</view>
    </view>

    <!-- 数据加载状态 -->
    <view v-if="loading" class="loading-state">
      <view class="loading-spinner"></view>
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 错误状态 -->
    <view v-else-if="error" class="error-state">
      <text class="error-text">{{ error }}</text>
      <button class="retry-btn" @click="loadDashboard">重试</button>
    </view>

    <!-- 主内容 -->
    <template v-else>
      <!-- 核心数据指标卡片 -->
      <view class="stats-section">
        <view class="stats-grid">
          <view class="stat-card">
            <view class="stat-value">{{ dashboardData?.total_merchants || 0 }}</view>
            <view class="stat-label">商家总数</view>
          </view>
          <view class="stat-card">
            <view class="stat-value">{{ dashboardData?.today_orders || 0 }}</view>
            <view class="stat-label">今日订单</view>
          </view>
          <view class="stat-card">
            <view class="stat-value">¥{{ formatAmount(dashboardData?.today_revenue || 0) }}</view>
            <view class="stat-label">今日金额</view>
          </view>
          <view class="stat-card">
            <view class="stat-value">¥{{ formatAmount(avgOrderAmount) }}</view>
            <view class="stat-label">平均订单金额</view>
          </view>
        </view>
      </view>

      <!-- 待处理任务提示 -->
      <view class="todo-section" v-if="hasPendingTasks && dashboardData">
        <view class="section-title">待处理任务</view>
        <view class="todo-grid">
          <view class="todo-item" v-if="(dashboardData.pending_merchants || 0) > 0" @click="goMerchantAudit">
            <view class="todo-icon approval"></view>
            <view class="todo-content">
              <text class="todo-count">{{ dashboardData.pending_merchants }}</text>
              <text class="todo-label">待审核商家</text>
            </view>
            <text class="arrow">›</text>
          </view>
        </view>
      </view>

      <!-- 商家行业分布图表 -->
      <view class="chart-section">
        <view class="section-title">商家行业分布</view>
        <view class="chart-container">
          <view class="pie-chart" v-if="industryDistribution.length > 0">
            <view class="pie-legend">
              <view
                class="legend-item"
                v-for="(item, index) in industryDistribution"
                :key="index"
              >
                <view class="legend-color" :style="{ background: chartColors[index % chartColors.length] }"></view>
                <text class="legend-text">{{ item.name }}</text>
                <text class="legend-value">{{ item.value }} ({{ item.percent }}%)</text>
              </view>
            </view>
          </view>
          <view v-else class="empty-chart">
            <text class="empty-text">暂无数据</text>
          </view>
        </view>
      </view>

      <!-- 订单趋势图表 -->
      <view class="chart-section">
        <view class="section-title">订单趋势</view>
        <view class="chart-container">
          <view class="trend-chart" v-if="orderTrend.length > 0">
            <view class="trend-bars">
              <view
                class="trend-bar-item"
                v-for="(item, index) in orderTrend"
                :key="index"
              >
                <view
                  class="trend-bar"
                  :style="{ height: getBarHeight(item.orders) + 'rpx' }"
                ></view>
                <text class="trend-label">{{ item.date }}</text>
              </view>
            </view>
            <view class="trend-header">
              <text class="trend-title">近{{ orderTrend.length }}天订单趋势</text>
            </view>
          </view>
          <view v-else class="empty-chart">
            <text class="empty-text">暂无数据</text>
          </view>
        </view>
      </view>

      <!-- 快捷入口按钮 -->
      <view class="menu-section">
        <view class="section-title">快捷入口</view>
        <view class="menu-grid">
          <view class="menu-item" @click="goMerchantAudit">
            <view class="menu-icon" style="background: #e6f7ff;">
              <image src="/static/icons/product.png" />
            </view>
            <text class="menu-text">商家审核</text>
          </view>
          <view class="menu-item" @click="goMerchantList">
            <view class="menu-icon" style="background: #fff7e6;">
              <image src="/static/icons/category.png" />
            </view>
            <text class="menu-text">商家列表</text>
          </view>
          <view class="menu-item" @click="goAnalytics">
            <view class="menu-icon" style="background: #f6ffed;">
              <image src="/static/icons/analytics.png" />
            </view>
            <text class="menu-text">数据分析</text>
          </view>
          <view class="menu-item" @click="goAnnouncements">
            <view class="menu-icon" style="background: #fff1f0;">
              <image src="/static/icons/invite.png" />
            </view>
            <text class="menu-text">系统公告</text>
          </view>
        </view>
      </view>
    </template>

    <!-- 底部tabbar占位 -->
    <view class="tabbar-placeholder"></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { get } from '@api'

// 服务商信息
interface ServiceProviderInfo {
  id: number
  name: string
  logo?: string
}

// Dashboard数据
interface DashboardData {
  total_merchants: number
  pending_merchants: number
  today_orders: number
  today_revenue: number
  distribution: { category: string; count: number }[]
  trend: { date: string; orders: number }[]
}

// 图表颜色
const chartColors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2']

// 状态变量
const loading = ref(true)
const error = ref('')
const providerInfo = ref<ServiceProviderInfo | null>(null)
const dashboardData = ref<DashboardData | null>(null)

// 计算属性
const hasPendingTasks = computed(() => {
  return (dashboardData.value?.pending_merchants || 0) > 0
})

const avgOrderAmount = computed(() => {
  const todayOrders = dashboardData.value?.today_orders || 0
  const todayRevenue = dashboardData.value?.today_revenue || 0
  if (todayOrders <= 0) {
    return 0
  }
  return todayRevenue / todayOrders
})

// 行业分布数据
const industryDistribution = computed(() => {
  if (!dashboardData.value?.distribution) return []
  const total = dashboardData.value.distribution.reduce((sum, item) => sum + item.count, 0)
  return dashboardData.value.distribution.map(item => ({
    name: item.category || '未分类',
    value: item.count,
    percent: total > 0 ? Math.round((item.count / total) * 100) : 0
  }))
})

// 订单趋势数据
const orderTrend = computed(() => {
  return dashboardData.value?.trend || []
})

// 最大订单数（用于计算柱状图高度）
const maxOrders = computed(() => {
  if (orderTrend.value.length === 0) return 0
  return Math.max(...orderTrend.value.map(item => item.orders))
})

// 获取柱状图高度
function getBarHeight(orders: number): number {
  if (maxOrders.value === 0) return 0
  const maxHeight = 160
  return (orders / maxOrders.value) * maxHeight
}

// 格式化金额
function formatAmount(amount: number): string {
  return amount.toFixed(2)
}

// 加载Dashboard数据
async function loadDashboard() {
  loading.value = true
  error.value = ''

  try {
    const response = await get<DashboardData>('/api/v1/sp/dashboard')
    dashboardData.value = response

    // 服务商信息在登录时以 JSON 字符串写入缓存，这里兼容字符串与对象两种情况。
    const cachedInfo = uni.getStorageSync('sp_info')
    if (cachedInfo) {
      if (typeof cachedInfo === 'string') {
        providerInfo.value = JSON.parse(cachedInfo)
      } else {
        providerInfo.value = cachedInfo
      }
    }
  } catch (err: any) {
    console.error('加载Dashboard失败:', err)
    error.value = err.message || '加载失败，请重试'
  } finally {
    loading.value = false
  }
}

// 跳转函数
function goMerchantAudit() {
  uni.navigateTo({ url: '/pages/sp/merchants/audit' })
}

function goMerchantList() {
  uni.navigateTo({ url: '/pages/sp/merchants/list' })
}

function goAnalytics() {
  uni.navigateTo({ url: '/pages/sp/analytics/merchant-stats' })
}

function goAnnouncements() {
  uni.navigateTo({ url: '/pages/sp/announcements/index' })
}

// 页面显示时加载数据
onShow(() => {
  loadDashboard()
})
</script>

<style scoped>
.sp-home-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.page-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 40rpx 32rpx 60rpx;
  color: #ffffff;
}

.header-title {
  font-size: 40rpx;
  font-weight: 600;
  margin-bottom: 8rpx;
}

.header-subtitle {
  font-size: 28rpx;
  opacity: 0.9;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-spinner {
  width: 64rpx;
  height: 64rpx;
  border: 4rpx solid #e5e5e5;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  margin-top: 24rpx;
  font-size: 28rpx;
  color: #999999;
}

.error-text {
  font-size: 28rpx;
  color: #ff4d4f;
  margin-bottom: 24rpx;
}

.retry-btn {
  padding: 16rpx 48rpx;
  background: #667eea;
  color: #ffffff;
  border-radius: 40rpx;
  font-size: 28rpx;
  border: none;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.stats-section {
  background: #ffffff;
  margin: -40rpx 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.06);
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

.todo-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.todo-grid {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.todo-item {
  display: flex;
  align-items: center;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
}

.todo-icon {
  width: 48rpx;
  height: 48rpx;
  border-radius: 12rpx;
  margin-right: 16rpx;
}

.todo-icon.approval {
  background: #e6f7ff;
}

.todo-icon.issue {
  background: #fff1f0;
}

.todo-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.todo-count {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.todo-label {
  font-size: 24rpx;
  color: #666666;
  margin-top: 4rpx;
}

.arrow {
  font-size: 32rpx;
  color: #cccccc;
}

.chart-section {
  background: #ffffff;
  margin: 0 24rpx 24rpx;
  padding: 32rpx;
  border-radius: 24rpx;
}

.chart-container {
  min-height: 200rpx;
}

.pie-chart {
  padding: 16rpx 0;
}

.pie-legend {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.legend-item {
  display: flex;
  align-items: center;
}

.legend-color {
  width: 24rpx;
  height: 24rpx;
  border-radius: 6rpx;
  margin-right: 12rpx;
}

.legend-text {
  flex: 1;
  font-size: 28rpx;
  color: #333333;
}

.legend-value {
  font-size: 28rpx;
  color: #666666;
}

.trend-chart {
  position: relative;
  padding: 16rpx 0;
}

.trend-header {
  text-align: center;
  margin-bottom: 16rpx;
}

.trend-title {
  font-size: 24rpx;
  color: #999999;
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  height: 200rpx;
  padding: 0 8rpx;
}

.trend-bar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.trend-bar {
  width: 32rpx;
  background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
  border-radius: 8rpx 8rpx 0 0;
  margin-bottom: 8rpx;
  min-height: 4rpx;
}

.trend-label {
  font-size: 20rpx;
  color: #999999;
  transform: rotate(-45deg);
  transform-origin: top left;
  white-space: nowrap;
}

.empty-chart {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200rpx;
}

.empty-text {
  font-size: 28rpx;
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
  grid-template-columns: repeat(2, 1fr);
  gap: 24rpx;
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: transparent;
  padding: 0;
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
  font-size: 28rpx;
  color: #666666;
}

.tabbar-placeholder {
  height: 120rpx;
}
</style>
