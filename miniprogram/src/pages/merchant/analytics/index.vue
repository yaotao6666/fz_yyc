<template>
  <view class="analytics-container">
    <!-- 概览卡片 -->
    <view class="overview-card">
      <view class="overview-header">
        <view class="period-tabs">
          <view
            v-for="tab in periodTabs"
            :key="tab.value"
            class="period-tab"
            :class="{ active: currentPeriod === tab.value }"
            @click="changePeriod(tab.value)"
          >
            {{ tab.label }}
          </view>
        </view>
      </view>

      <view class="overview-main">
        <view class="main-stat">
          <view class="stat-value">¥{{ formatAmount(overview?.total_sales || 0) }}</view>
          <view class="stat-label">销售额</view>
          <view class="stat-growth" :class="{ positive: (overview?.sales_growth || 0) >= 0 }">
            {{ (overview?.sales_growth || 0) >= 0 ? '↑' : '↓' }}
            {{ Math.abs(overview?.sales_growth || 0) }}%
          </view>
        </view>
      </view>

      <view class="overview-sub">
        <view class="sub-stat">
          <view class="stat-value">{{ overview?.total_orders || 0 }}</view>
          <view class="stat-label">订单数</view>
          <view class="stat-growth" :class="{ positive: (overview?.orders_growth || 0) >= 0 }">
            {{ (overview?.orders_growth || 0) >= 0 ? '↑' : '↓' }}
            {{ Math.abs(overview?.orders_growth || 0) }}%
          </view>
        </view>
        <view class="sub-stat">
          <view class="stat-value">{{ overview?.total_customers || 0 }}</view>
          <view class="stat-label">客户数</view>
          <view class="stat-growth" :class="{ positive: (overview?.customers_growth || 0) >= 0 }">
            {{ (overview?.customers_growth || 0) >= 0 ? '↑' : '↓' }}
            {{ Math.abs(overview?.customers_growth || 0) }}%
          </view>
        </view>
        <view class="sub-stat">
          <view class="stat-value">¥{{ formatAmount(overview?.avg_order_amount || 0) }}</view>
          <view class="stat-label">客单价</view>
        </view>
      </view>
    </view>

    <!-- 销售趋势 -->
    <view class="section">
      <view class="section-header">
        <view class="section-title">销售趋势</view>
        <picker
          mode="date"
          :value="trendParams.start_date"
          @change="onTrendDateChange"
          fields="day"
        >
          <view class="date-picker">
            <text>{{ trendParams.start_date }}</text>
            <text class="arrow">›</text>
          </view>
        </picker>
      </view>
      <view class="chart-placeholder">
        <view class="chart-bars">
          <view
            v-for="(item, index) in salesTrend.slice(-7)"
            :key="index"
            class="chart-bar"
            :style="{ height: getBarHeight(item.sales) + '%' }"
          >
            <text class="bar-value">¥{{ formatShortAmount(item.sales) }}</text>
          </view>
        </view>
        <view class="chart-labels">
          <text v-for="(item, index) in salesTrend.slice(-7)" :key="index">
            {{ item.date.slice(5) }}
          </text>
        </view>
      </view>
    </view>

    <!-- 商品排行 -->
    <view class="section">
      <view class="section-header">
        <view class="section-title">商品销量排行</view>
        <text class="more-link" @click="viewProductRanking">查看全部 ›</text>
      </view>
      <view class="product-ranking">
        <view
          v-for="(item, index) in productRanking"
          :key="item.product_id"
          class="ranking-item"
        >
          <view class="rank-badge" :class="getRankClass(index)">{{ index + 1 }}</view>
          <image
            class="product-image"
            :src="item.image || '/static/default-product.png'"
            mode="aspectFill"
          />
          <view class="product-info">
            <view class="product-name">{{ item.product_name }}</view>
            <view class="product-sales">销量 {{ item.sales_count }}</view>
          </view>
          <view class="product-amount">¥{{ formatAmount(item.sales_amount) }}</view>
        </view>

        <view v-if="productRanking.length === 0" class="empty-ranking">
          <text>暂无数据</text>
        </view>
      </view>
    </view>

    <!-- 时段分析 -->
    <view class="section">
      <view class="section-header">
        <view class="section-title">今日时段分析</view>
      </view>
      <view class="hourly-chart">
        <view
          v-for="hour in hourlyData"
          :key="hour.hour"
          class="hour-bar"
        >
          <view
            class="hour-value"
            :style="{ height: getHourBarHeight(hour.orders) + '%' }"
          ></view>
          <text class="hour-label">{{ hour.hour }}时</text>
        </view>
      </view>
    </view>

    <!-- 库存预警 -->
    <view class="section" v-if="stockAlerts.length > 0">
      <view class="section-header">
        <view class="section-title">库存预警</view>
        <text class="warning-badge">{{ stockAlerts.length }}个商品</text>
      </view>
      <view class="stock-alerts">
        <view
          v-for="item in stockAlerts"
          :key="item.product_id"
          class="alert-item"
        >
          <view class="alert-name">{{ item.product_name }}</view>
          <view class="alert-stock">剩余 {{ item.stock }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getSalesOverview,
  getSalesTrend,
  getProductRanking,
  getHourlyAnalysis,
  getStockAlert
} from '../../../api'
import type {
  SalesOverview,
  SalesTrend,
  ProductRanking,
  HourlyAnalysis,
  StockAlert
} from '../../../types/index'

const periodTabs = [
  { label: '今日', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '本年', value: 'year' }
]

const currentPeriod = ref('today')
const overview = ref<SalesOverview | null>(null)
const salesTrend = ref<SalesTrend[]>([])
const productRanking = ref<ProductRanking[]>([])
const hourlyData = ref<HourlyAnalysis[]>([])
const stockAlerts = ref<StockAlert[]>([])

const trendParams = reactive({
  start_date: getTodayDate(),
  end_date: getTodayDate(),
  granularity: 'day' as const
})

onShow(() => {
  loadData()
})

async function loadData() {
  await Promise.all([
    loadOverview(),
    loadSalesTrend(),
    loadProductRanking(),
    loadHourlyData(),
    loadStockAlerts()
  ])
}

async function loadOverview() {
  try {
    overview.value = await getSalesOverview({ period: currentPeriod.value as any })
  } catch (error) {
    console.error('加载销售概览失败:', error)
  }
}

async function loadSalesTrend() {
  try {
    salesTrend.value = await getSalesTrend(trendParams)
  } catch (error) {
    console.error('加载销售趋势失败:', error)
  }
}

async function loadProductRanking() {
  try {
    productRanking.value = await getProductRanking({ limit: 5 })
  } catch (error) {
    console.error('加载商品排行失败:', error)
  }
}

async function loadHourlyData() {
  try {
    hourlyData.value = await getHourlyAnalysis({ date: getTodayDate() })
  } catch (error) {
    console.error('加载时段分析失败:', error)
  }
}

async function loadStockAlerts() {
  try {
    stockAlerts.value = await getStockAlert({ threshold: 10 })
  } catch (error) {
    console.error('加载库存预警失败:', error)
  }
}

function changePeriod(period: string) {
  currentPeriod.value = period
  loadOverview()
}

function onTrendDateChange(e: any) {
  trendParams.start_date = e.detail.value
  loadSalesTrend()
}

function getTodayDate(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
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

function getBarHeight(sales: number): number {
  if (salesTrend.value.length === 0) return 0
  const maxSales = Math.max(...salesTrend.value.map(item => item.sales))
  if (maxSales === 0) return 0
  return (sales / maxSales) * 100
}

function getHourBarHeight(orders: number): number {
  if (hourlyData.value.length === 0) return 0
  const maxOrders = Math.max(...hourlyData.value.map(item => item.orders))
  if (maxOrders === 0) return 0
  return (orders / maxOrders) * 100
}

function getRankClass(index: number): string {
  const classMap = ['gold', 'silver', 'bronze']
  return index < 3 ? classMap[index] : ''
}

function viewProductRanking() {
  uni.navigateTo({ url: '/pages/merchant/analytics/products' })
}
</script>

<style scoped>
.analytics-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  padding-bottom: 48rpx;
}

.overview-card {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.period-tabs {
  display: flex;
  gap: 16rpx;
  margin-bottom: 32rpx;
}

.period-tab {
  flex: 1;
  text-align: center;
  padding: 16rpx;
  border-radius: 12rpx;
  font-size: 28rpx;
  background: rgba(255, 255, 255, 0.2);
}

.period-tab.active {
  background: #ffffff;
  color: #007AFF;
}

.overview-main {
  text-align: center;
  padding: 24rpx 0;
}

.main-stat .stat-value {
  font-size: 64rpx;
  font-weight: 700;
  margin-bottom: 8rpx;
}

.main-stat .stat-label {
  font-size: 28rpx;
  opacity: 0.9;
}

.main-stat .stat-growth {
  font-size: 24rpx;
  margin-top: 8rpx;
  opacity: 0.8;
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
  margin-bottom: 8rpx;
}

.sub-stat .stat-growth {
  font-size: 22rpx;
  opacity: 0.8;
}

.stat-growth.positive {
  color: #52c41a;
}

.section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-bottom: 24rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.date-picker {
  display: flex;
  align-items: center;
  font-size: 26rpx;
  color: #007AFF;
}

.arrow {
  margin-left: 4rpx;
}

.more-link {
  font-size: 26rpx;
  color: #007AFF;
}

.chart-placeholder {
  height: 300rpx;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  height: 240rpx;
  padding-bottom: 20rpx;
}

.chart-bar {
  width: 60rpx;
  background: linear-gradient(180deg, #007AFF 0%, #e6f0ff 100%);
  border-radius: 8rpx 8rpx 0 0;
  position: relative;
  min-height: 20rpx;
}

.bar-value {
  position: absolute;
  top: -40rpx;
  left: 50%;
  transform: translateX(-50%);
  font-size: 20rpx;
  color: #666666;
  white-space: nowrap;
}

.chart-labels {
  display: flex;
  justify-content: space-around;
  font-size: 22rpx;
  color: #999999;
}

.product-ranking {
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

.rank-badge {
  width: 40rpx;
  height: 40rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  font-weight: 600;
  background: #f0f0f0;
  color: #666666;
  margin-right: 16rpx;
}

.rank-badge.gold {
  background: #ffd700;
  color: #ffffff;
}

.rank-badge.silver {
  background: #c0c0c0;
  color: #ffffff;
}

.rank-badge.bronze {
  background: #cd7f32;
  color: #ffffff;
}

.product-image {
  width: 80rpx;
  height: 80rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 16rpx;
}

.product-info {
  flex: 1;
}

.product-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.product-sales {
  font-size: 24rpx;
  color: #999999;
}

.product-amount {
  font-size: 28rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.empty-ranking {
  text-align: center;
  padding: 48rpx 0;
  font-size: 28rpx;
  color: #999999;
}

.hourly-chart {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  height: 200rpx;
  padding-bottom: 40rpx;
}

.hour-bar {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  justify-content: flex-end;
}

.hour-value {
  width: 32rpx;
  background: #007AFF;
  border-radius: 4rpx 4rpx 0 0;
  min-height: 4rpx;
}

.hour-label {
  font-size: 20rpx;
  color: #999999;
  margin-top: 8rpx;
}

.warning-badge {
  font-size: 24rpx;
  color: #ff4d4f;
  padding: 4rpx 12rpx;
  background: #fff1f0;
  border-radius: 8rpx;
}

.stock-alerts {
  display: flex;
  flex-direction: column;
}

.alert-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.alert-item:last-child {
  border-bottom: none;
}

.alert-name {
  font-size: 28rpx;
  color: #1a1a1a;
}

.alert-stock {
  font-size: 26rpx;
  color: #ff4d4f;
}
</style>
