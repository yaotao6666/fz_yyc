<template>
  <scroll-view class="sp-home-container" scroll-y>
    <view class="page-header">
      <view>
        <view class="header-title">服务商管理中心</view>
        <view class="header-subtitle">{{ providerInfo?.name || '服务商' }}</view>
      </view>
      <view class="header-tag">支付分账模式</view>
    </view>

    <view v-if="loading" class="state-card">加载中...</view>
    <view v-else-if="error" class="state-card error-state">
      <text>{{ error }}</text>
      <button class="retry-btn" @click="loadDashboard">重新加载</button>
    </view>

    <template v-else>
      <view class="section">
        <view class="section-title">经营总览</view>
        <view class="stats-grid">
          <view class="stat-card">
            <text class="stat-value">{{ dashboardData.total_merchants }}</text>
            <text class="stat-label">商家总数</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">{{ dashboardData.today_orders }}</text>
            <text class="stat-label">今日订单</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">¥{{ formatAmount(dashboardData.today_revenue) }}</text>
            <text class="stat-label">今日成交</text>
          </view>
          <view class="stat-card">
            <text class="stat-value">¥{{ formatAmount(avgOrderAmount) }}</text>
            <text class="stat-label">平均客单</text>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-title">快捷入口</view>
        <view class="menu-grid">
          <view class="menu-card primary" @click="goMerchantCreate">
            <text class="menu-title">新增商家</text>
            <text class="menu-desc">直接创建商家并配置账号与收款信息</text>
          </view>
          <view class="menu-card" @click="goMerchantList">
            <text class="menu-title">商家列表</text>
            <text class="menu-desc">查看支付配置状态、分账比例和经营概览</text>
          </view>
          <view class="menu-card" @click="goProfitSharingHistory">
            <text class="menu-title">分账历史</text>
            <text class="menu-desc">按日期、商家、状态筛选抽佣记录</text>
          </view>
          <view class="menu-card" @click="goAnalytics">
            <text class="menu-title">数据分析</text>
            <text class="menu-desc">查看访问率、下单率、订单量和排行榜</text>
          </view>
          <view class="menu-card" @click="goAnnouncements">
            <text class="menu-title">系统公告</text>
            <text class="menu-desc">维护面对商家的统一公告与通知内容</text>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-title">当前说明</view>
        <view class="tip-card">
          <text class="tip-line">商家由服务商直接创建，并在商家管理中维护资料、收款账户和分账配置。</text>
          <text class="tip-line">每个商家单独配置 `sub_mch_id`、分账开关和抽佣比例，支付成功后按配置自动分账。</text>
        </view>
      </view>
    </template>
  </scroll-view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { get } from '@api'

interface ServiceProviderInfo {
  id: number
  name: string
  logo?: string
}

interface DashboardData {
  total_merchants: number
  pending_merchants?: number
  today_orders: number
  today_revenue: number
}

const loading = ref(true)
const error = ref('')
const providerInfo = ref<ServiceProviderInfo | null>(null)
const dashboardData = ref<DashboardData>({
  total_merchants: 0,
  today_orders: 0,
  today_revenue: 0
})

const avgOrderAmount = computed(() => {
  if (!dashboardData.value.today_orders) {
    return 0
  }
  return dashboardData.value.today_revenue / dashboardData.value.today_orders
})

function formatAmount(amount: number) {
  return Number(amount || 0).toFixed(2)
}

function parseProviderInfo() {
  const cachedInfo = uni.getStorageSync('sp_info')
  if (!cachedInfo) {
    providerInfo.value = null
    return
  }

  if (typeof cachedInfo === 'string') {
    try {
      providerInfo.value = JSON.parse(cachedInfo)
      return
    } catch (parseError) {
      console.warn('解析服务商缓存信息失败:', parseError)
    }
  }

  providerInfo.value = cachedInfo
}

async function loadDashboard() {
  loading.value = true
  error.value = ''
  parseProviderInfo()

  try {
    dashboardData.value = await get<DashboardData>('/api/v1/sp/dashboard')
  } catch (requestError: any) {
    console.error('加载服务商工作台失败:', requestError)
    error.value = requestError?.message || '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goMerchantCreate() {
  uni.navigateTo({ url: '/pages/sp/merchants/edit' })
}

function goMerchantList() {
  uni.navigateTo({ url: '/pages/sp/merchants/list' })
}

function goProfitSharingHistory() {
  uni.navigateTo({ url: '/pages/sp/settlements/history' })
}

function goAnalytics() {
  uni.navigateTo({ url: '/pages/sp/analytics/merchant-stats' })
}

function goAnnouncements() {
  uni.navigateTo({ url: '/pages/sp/announcements/index' })
}

onShow(() => {
  loadDashboard()
})
</script>

<style scoped>
.sp-home-container {
  min-height: 100vh;
  width: calc(100% - 48rpx);
  background: #f5f5f5;
  padding: 24rpx;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 40rpx 32rpx;
  border-radius: 28rpx;
  background: linear-gradient(135deg, #3f7cff 0%, #635bff 100%);
  color: #ffffff;
  margin-bottom: 24rpx;
}

.header-title {
  font-size: 40rpx;
  font-weight: 600;
}

.header-subtitle {
  margin-top: 12rpx;
  font-size: 26rpx;
  opacity: 0.9;
}

.header-tag {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  font-size: 22rpx;
}

.section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 24rpx;
}

.state-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx 32rpx;
  text-align: center;
  color: #666666;
}

.error-state {
  color: #cf1322;
}

.retry-btn {
  margin-top: 24rpx;
  width: 240rpx;
  height: 76rpx;
  line-height: 76rpx;
  border: none;
  border-radius: 999rpx;
  background: #1677ff;
  color: #ffffff;
  font-size: 28rpx;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.stat-card {
  padding: 28rpx 24rpx;
  border-radius: 20rpx;
  background: #f7f8fa;
}

.stat-value {
  display: block;
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2329;
}

.stat-label {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #86909c;
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.menu-card {
  padding: 28rpx 24rpx;
  border-radius: 22rpx;
  background: #f7f8fa;
}

.menu-card.primary {
  background: #edf5ff;
}

.menu-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
}

.menu-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.6;
  color: #86909c;
}

.tip-card {
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f7f8fa;
}

.tip-line {
  display: block;
  font-size: 25rpx;
  line-height: 1.7;
  color: #4e5969;
}

.tip-line + .tip-line {
  margin-top: 12rpx;
}
</style>
