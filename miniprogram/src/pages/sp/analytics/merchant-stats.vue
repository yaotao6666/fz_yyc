<template>
  <view class="merchant-stats-container">
    <!-- Tab 标签切换 -->
    <view class="tab-bar">
      <view
        v-for="tab in tabs"
        :key="tab.key"
        class="tab-item"
        :class="{ active: currentTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </view>
    </view>

    <!-- 数据加载状态 -->
    <view v-if="loading" class="loading-state">
      <view class="loading-spinner"></view>
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 错误状态 -->
    <view v-else-if="error" class="error-state">
      <text class="error-text">{{ error }}</text>
      <button class="retry-btn" @click="loadCurrentTabData">重试</button>
    </view>

    <!-- 主内容区域 -->
    <scroll-view v-else scroll-y class="content-scroll">
      <!-- 商家分布分析 -->
      <view v-show="currentTab === 'distribution'" class="tab-content">
        <!-- 行业分类统计 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">商家行业分布</view>
          </view>
          <view class="distribution-grid">
            <view
              v-for="(item, index) in distributionData.industry_distribution"
              :key="index"
              class="distribution-item"
            >
              <view class="item-value">{{ item.value }}</view>
              <view class="item-label">{{ item.name }}</view>
              <view class="item-percent">{{ item.percent || 0 }}%</view>
            </view>
          </view>
        </view>

        <!-- 商家状态统计 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">商家状态分布</view>
          </view>
          <view class="status-distribution">
            <view
              v-for="(item, index) in distributionData.status_distribution"
              :key="index"
              class="status-item"
            >
              <view class="status-indicator" :style="{ background: statusColors[index] }"></view>
              <view class="status-info">
                <text class="status-name">{{ item.name }}</text>
                <text class="status-value">{{ item.value }} ({{ item.percent || 0 }}%)</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 新增商家趋势 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">商家增长趋势</view>
          </view>
          <view class="trend-chart">
            <view class="chart-bars">
              <view
                v-for="(item, index) in distributionData.monthly_trend"
                :key="index"
                class="chart-bar-wrapper"
              >
                <view
                  class="chart-bar"
                  :style="{ height: getTrendBarHeight(item.count) + '%' }"
                ></view>
                <text class="bar-value">{{ item.count }}</text>
              </view>
            </view>
            <view class="chart-labels">
              <text v-for="(item, index) in distributionData.monthly_trend" :key="index">
                {{ item.month }}
              </text>
            </view>
          </view>
        </view>
      </view>

      <!-- 订单分析 -->
      <view v-show="currentTab === 'orders'" class="tab-content">
        <!-- 订单概览 -->
        <view class="section overview-card">
          <view class="overview-main">
            <view class="main-stat">
              <view class="stat-value">{{ orderData.total_orders }}</view>
              <view class="stat-label">总订单数</view>
            </view>
          </view>
          <view class="overview-sub">
            <view class="sub-stat">
              <view class="stat-value">{{ orderData.today_orders }}</view>
              <view class="stat-label">今日订单</view>
            </view>
            <view class="sub-stat">
              <view class="stat-value">{{ orderData.yesterday_orders }}</view>
              <view class="stat-label">昨日订单</view>
            </view>
            <view class="sub-stat">
              <view class="stat-value" :class="{ positive: orderData.growth_rate >= 0, negative: orderData.growth_rate < 0 }">
                {{ orderData.growth_rate >= 0 ? '↑' : '↓' }}{{ Math.abs(orderData.growth_rate) }}%
              </view>
              <view class="stat-label">环比增长</view>
            </view>
          </view>
        </view>

        <!-- 订单趋势图 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">订单趋势</view>
          </view>
          <view class="trend-chart">
            <view class="chart-bars">
              <view
                v-for="(item, index) in orderData.order_trend"
                :key="index"
                class="chart-bar-wrapper"
              >
                <view
                  class="chart-bar order-bar"
                  :style="{ height: getOrderBarHeight(item.orders) + '%' }"
                ></view>
                <text class="bar-value">{{ item.orders }}</text>
              </view>
            </view>
            <view class="chart-labels">
              <text v-for="(item, index) in orderData.order_trend" :key="index">
                {{ item.date }}
              </text>
            </view>
          </view>
        </view>

        <!-- 配送方式分布 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">配送方式分布</view>
          </view>
          <view class="delivery-distribution">
            <view
              v-for="(item, index) in orderData.delivery_distribution"
              :key="index"
              class="delivery-item"
            >
              <view class="delivery-header">
                <text class="delivery-name">{{ item.name }}</text>
                <text class="delivery-percent">{{ item.percent || 0 }}%</text>
              </view>
              <view class="delivery-bar-bg">
                <view
                  class="delivery-bar-fill"
                  :style="{ width: (item.percent || 0) + '%' }"
                ></view>
              </view>
              <text class="delivery-value">{{ item.value }} 单</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 金额分析 -->
      <view v-show="currentTab === 'amount'" class="tab-content">
        <!-- 金额概览 -->
        <view class="section overview-card amount-card">
          <view class="overview-main">
            <view class="main-stat">
              <view class="stat-value">¥{{ formatAmount(amountData.total_amount) }}</view>
              <view class="stat-label">累计金额</view>
            </view>
          </view>
          <view class="overview-sub">
            <view class="sub-stat">
              <view class="stat-value">¥{{ formatAmount(amountData.today_amount) }}</view>
              <view class="stat-label">今日金额</view>
            </view>
            <view class="sub-stat">
              <view class="stat-value">¥{{ formatAmount(amountData.week_amount) }}</view>
              <view class="stat-label">本周金额</view>
            </view>
            <view class="sub-stat">
              <view class="stat-value">¥{{ formatAmount(amountData.month_amount) }}</view>
              <view class="stat-label">本月金额</view>
            </view>
          </view>
        </view>

        <!-- 按商家分布 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">商家金额分布 TOP5</view>
          </view>
          <view class="merchant-distribution">
            <view
              v-for="(item, index) in amountData.by_merchant"
              :key="index"
              class="merchant-item"
            >
              <view class="merchant-rank" :class="getRankClass(index)">{{ index + 1 }}</view>
              <view class="merchant-info">
                <text class="merchant-name">{{ item.merchant_name }}</text>
                <view class="merchant-bar-bg">
                  <view
                    class="merchant-bar-fill"
                    :style="{ width: (item.percent || 0) + '%' }"
                  ></view>
                </view>
              </view>
              <view class="merchant-amount">
                <text class="amount-value">¥{{ formatAmount(item.amount) }}</text>
                <text class="amount-percent">{{ item.percent || 0 }}%</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 按行业分布 -->
        <view class="section">
          <view class="section-header">
            <view class="section-title">行业金额分布</view>
          </view>
          <view class="industry-distribution">
            <view
              v-for="(item, index) in amountData.by_industry"
              :key="index"
              class="industry-item"
            >
              <view class="industry-indicator" :style="{ background: industryColors[index % industryColors.length] }"></view>
              <view class="industry-info">
                <text class="industry-name">{{ item.name }}</text>
                <view class="industry-bar-bg">
                  <view
                    class="industry-bar-fill"
                    :style="{ width: (item.percent || 0) + '%', background: industryColors[index % industryColors.length] }"
                  ></view>
                </view>
              </view>
              <view class="industry-amount">
                <text class="amount-value">¥{{ formatShortAmount(item.amount) }}</text>
                <text class="amount-percent">{{ item.percent || 0 }}%</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- TOP商家排行 -->
      <view v-show="currentTab === 'top'" class="tab-content">
        <view class="section">
          <view class="section-header">
            <view class="section-title">商家金额排行 TOP10</view>
          </view>
          <view class="ranking-list">
            <view
              v-for="item in topMerchants"
              :key="item.merchant_id"
              class="ranking-item"
            >
              <view class="ranking-rank" :class="getRankClass(item.rank - 1)">
                {{ item.rank <= 3 ? '' : item.rank }}
              </view>
              <image
                v-if="item.merchant_logo"
                class="ranking-logo"
                :src="item.merchant_logo"
                mode="aspectFill"
              />
              <view v-else class="ranking-logo default-logo">
                <text>{{ item.merchant_name.charAt(0) }}</text>
              </view>
              <view class="ranking-info">
                <text class="ranking-name">{{ item.merchant_name }}</text>
                <text class="ranking-orders">{{ item.order_count }} 笔订单</text>
              </view>
              <view class="ranking-amount">
                <text class="amount-value">¥{{ formatAmount(item.amount) }}</text>
              </view>
            </view>

            <view v-if="topMerchants.length === 0" class="empty-ranking">
              <text>暂无数据</text>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  getMerchantDistribution,
  getOrderAnalytics,
  getAmountAnalytics,
  getTopMerchants
} from '@api'
import type {
  MerchantDistributionData,
  OrderAnalyticsData,
  AmountAnalyticsData,
  TopMerchantRanking
} from '@types'

const tabs = [
  { key: 'distribution', label: '商家分布' },
  { key: 'orders', label: '订单分析' },
  { key: 'amount', label: '金额分析' },
  { key: 'top', label: 'TOP排行' }
]

const currentTab = ref('distribution')
const loading = ref(false)
const error = ref('')

const distributionData = ref<MerchantDistributionData>({
  industry_distribution: [],
  status_distribution: [],
  monthly_trend: []
})

const orderData = ref<OrderAnalyticsData>({
  total_orders: 0,
  today_orders: 0,
  yesterday_orders: 0,
  growth_rate: 0,
  order_trend: [],
  delivery_distribution: []
})

const amountData = ref<AmountAnalyticsData>({
  total_amount: 0,
  today_amount: 0,
  week_amount: 0,
  month_amount: 0,
  growth_rate: 0,
  by_merchant: [],
  by_industry: []
})

const topMerchants = ref<TopMerchantRanking[]>([])

const statusColors = ['#52c41a', '#faad14', '#f5222d']
const industryColors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2']

onMounted(() => {
  loadCurrentTabData()
})

function switchTab(tabKey: string) {
  currentTab.value = tabKey
  loadCurrentTabData()
}

async function loadCurrentTabData() {
  loading.value = true
  error.value = ''

  try {
    switch (currentTab.value) {
      case 'distribution':
        await loadDistributionData()
        break
      case 'orders':
        await loadOrderData()
        break
      case 'amount':
        await loadAmountData()
        break
      case 'top':
        await loadTopMerchants()
        break
    }
  } catch (err: any) {
    console.error('加载数据失败:', err)
    error.value = err.message || '加载失败，请重试'
  } finally {
    loading.value = false
  }
}

async function loadDistributionData() {
  try {
    distributionData.value = await getMerchantDistribution()
  } catch (err: any) {
    console.error('加载商家分布数据失败:', err)
    throw err
  }
}

async function loadOrderData() {
  try {
    orderData.value = await getOrderAnalytics()
  } catch (err: any) {
    console.error('加载订单分析数据失败:', err)
    throw err
  }
}

async function loadAmountData() {
  try {
    amountData.value = await getAmountAnalytics()
  } catch (err: any) {
    console.error('加载金额分析数据失败:', err)
    throw err
  }
}

async function loadTopMerchants() {
  try {
    topMerchants.value = await getTopMerchants({ limit: 10 })
  } catch (err: any) {
    console.error('加载TOP商家数据失败:', err)
    throw err
  }
}

function formatAmount(amount: number): string {
  return amount.toFixed(2)
}

function formatShortAmount(amount: number): string {
  if (amount >= 10000) {
    return (amount / 10000).toFixed(1) + 'w'
  }
  return amount.toFixed(0)
}

function getTrendBarHeight(count: number): number {
  if (distributionData.value.monthly_trend.length === 0) return 0
  const maxCount = Math.max(...distributionData.value.monthly_trend.map(item => item.count))
  if (maxCount === 0) return 0
  return (count / maxCount) * 100
}

function getOrderBarHeight(orders: number): number {
  if (orderData.value.order_trend.length === 0) return 0
  const maxOrders = Math.max(...orderData.value.order_trend.map(item => item.orders))
  if (maxOrders === 0) return 0
  return (orders / maxOrders) * 100
}

function getRankClass(index: number): string {
  const classMap = ['gold', 'silver', 'bronze']
  return index < 3 ? classMap[index] : ''
}
</script>

<style scoped>
.merchant-stats-container {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.tab-bar {
  display: flex;
  background: #ffffff;
  padding: 0 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  font-size: 28rpx;
  color: #666666;
  position: relative;
  transition: all 0.3s;
}

.tab-item.active {
  color: #667eea;
  font-weight: 600;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 6rpx;
  background: #667eea;
  border-radius: 3rpx;
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

.content-scroll {
  flex: 1;
  padding: 24rpx;
  padding-bottom: 48rpx;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.section-header {
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.overview-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
}

.overview-card .stat-label {
  color: rgba(255, 255, 255, 0.9);
}

.overview-card .amount-value {
  color: #ffffff;
}

.overview-main {
  text-align: center;
  padding: 24rpx 0;
}

.main-stat .stat-value {
  font-size: 56rpx;
  font-weight: 700;
  margin-bottom: 8rpx;
}

.main-stat .stat-label {
  font-size: 28rpx;
  opacity: 0.9;
}

.overview-sub {
  display: flex;
  justify-content: space-around;
  padding-top: 24rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.2);
}

.sub-stat .stat-value {
  font-size: 36rpx;
  font-weight: 600;
  margin-bottom: 8rpx;
}

.sub-stat .stat-label {
  font-size: 24rpx;
  opacity: 0.8;
}

.stat-value.positive {
  color: #52c41a;
}

.stat-value.negative {
  color: #ff4d4f;
}

.distribution-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24rpx;
}

.distribution-item {
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 24rpx;
  text-align: center;
}

.item-value {
  font-size: 40rpx;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 8rpx;
}

.item-label {
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 4rpx;
}

.item-percent {
  font-size: 24rpx;
  color: #999999;
}

.status-distribution {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 20rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
}

.status-indicator {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  margin-right: 16rpx;
}

.status-info {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-name {
  font-size: 28rpx;
  color: #333333;
}

.status-value {
  font-size: 26rpx;
  color: #666666;
}

.trend-chart {
  height: 240rpx;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  height: 180rpx;
  padding-bottom: 20rpx;
}

.chart-bar-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.chart-bar {
  width: 48rpx;
  background: linear-gradient(180deg, #667eea 0%, #e6e6f0 100%);
  border-radius: 8rpx 8rpx 0 0;
  margin-bottom: 8rpx;
  min-height: 8rpx;
  transition: height 0.3s ease;
}

.chart-bar.order-bar {
  background: linear-gradient(180deg, #667eea 0%, #e6e0ff 100%);
}

.bar-value {
  font-size: 20rpx;
  color: #666666;
  margin-top: 4rpx;
}

.chart-labels {
  display: flex;
  justify-content: space-around;
  font-size: 22rpx;
  color: #999999;
  margin-top: 8rpx;
}

.delivery-distribution {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.delivery-item {
  padding: 16rpx 0;
}

.delivery-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.delivery-name {
  font-size: 28rpx;
  color: #333333;
}

.delivery-percent {
  font-size: 26rpx;
  color: #667eea;
  font-weight: 600;
}

.delivery-bar-bg {
  height: 16rpx;
  background: #f0f0f0;
  border-radius: 8rpx;
  overflow: hidden;
  margin-bottom: 8rpx;
}

.delivery-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  border-radius: 8rpx;
  transition: width 0.3s ease;
}

.delivery-value {
  font-size: 24rpx;
  color: #999999;
}

.merchant-distribution,
.industry-distribution {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.merchant-item,
.industry-item {
  display: flex;
  align-items: center;
  padding: 16rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
}

.merchant-rank,
.industry-indicator {
  width: 40rpx;
  height: 40rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  font-weight: 600;
  margin-right: 16rpx;
  background: #e5e5e5;
  color: #666666;
}

.merchant-rank.gold {
  background: #ffd700;
  color: #ffffff;
}

.merchant-rank.silver {
  background: #c0c0c0;
  color: #ffffff;
}

.merchant-rank.bronze {
  background: #cd7f32;
  color: #ffffff;
}

.merchant-info,
.industry-info {
  flex: 1;
  margin-right: 16rpx;
}

.merchant-name,
.industry-name {
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 8rpx;
  display: block;
}

.merchant-bar-bg,
.industry-bar-bg {
  height: 12rpx;
  background: #e5e5e5;
  border-radius: 6rpx;
  overflow: hidden;
}

.merchant-bar-fill,
.industry-bar-fill {
  height: 100%;
  border-radius: 6rpx;
  transition: width 0.3s ease;
}

.merchant-amount,
.industry-amount {
  text-align: right;
  min-width: 140rpx;
}

.amount-value {
  font-size: 28rpx;
  font-weight: 600;
  color: #ff4d4f;
  display: block;
}

.amount-percent {
  font-size: 22rpx;
  color: #999999;
}

.ranking-list {
  display: flex;
  flex-direction: column;
}

.ranking-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.ranking-item:last-child {
  border-bottom: none;
}

.ranking-rank {
  width: 48rpx;
  height: 48rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
  background: #f0f0f0;
  color: #666666;
  margin-right: 16rpx;
}

.ranking-rank.gold {
  background: linear-gradient(135deg, #ffd700 0%, #ffb700 100%);
  color: #ffffff;
}

.ranking-rank.silver {
  background: linear-gradient(135deg, #c0c0c0 0%, #a8a8a8 100%);
  color: #ffffff;
}

.ranking-rank.bronze {
  background: linear-gradient(135deg, #cd7f32 0%, #b86c2a 100%);
  color: #ffffff;
}

.ranking-logo {
  width: 80rpx;
  height: 80rpx;
  border-radius: 16rpx;
  background: #f0f0f0;
  margin-right: 16rpx;
}

.default-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  font-size: 32rpx;
  font-weight: 600;
}

.ranking-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.ranking-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.ranking-orders {
  font-size: 24rpx;
  color: #999999;
}

.ranking-amount {
  text-align: right;
}

.empty-ranking {
  text-align: center;
  padding: 48rpx 0;
  font-size: 28rpx;
  color: #999999;
}
</style>
