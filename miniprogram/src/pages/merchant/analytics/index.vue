<template>
  <view class="analytics-container">
    <view class="overview-card">
      <view class="period-tabs">
        <view
          v-for="tab in periodTabs"
          :key="tab.value"
          class="period-tab"
          :class="{ 'period-tab-active': tab.active }"
          @click="changePeriod(tab.value)"
        >
          {{ tab.label }}
        </view>
      </view>

      <view class="range-text">{{ currentRangeLabel }}</view>

      <view class="overview-main">
        <view class="main-stat">
          <view class="stat-value">¥{{ formatAmount(overview?.total_sales || 0) }}</view>
          <view class="stat-label">销售额</view>
          <view class="stat-growth" :class="{ positive: (overview?.sales_growth || 0) >= 0 }">
            {{ (overview?.sales_growth || 0) >= 0 ? '↑' : '↓' }}
            {{ Math.abs(overview?.sales_growth || 0).toFixed(1) }}%
          </view>
        </view>
      </view>

      <view class="overview-sub">
        <view class="sub-stat">
          <view class="stat-value">{{ overview?.total_orders || 0 }}</view>
          <view class="stat-label">支付订单</view>
          <view class="stat-growth" :class="{ positive: (overview?.orders_growth || 0) >= 0 }">
            {{ (overview?.orders_growth || 0) >= 0 ? '↑' : '↓' }}
            {{ Math.abs(overview?.orders_growth || 0).toFixed(1) }}%
          </view>
        </view>
        <view class="sub-stat">
          <view class="stat-value">{{ overview?.visit_users || 0 }}</view>
          <view class="stat-label">浏览人数</view>
        </view>
        <view class="sub-stat">
          <view class="stat-value">{{ overview?.pay_success_users || 0 }}</view>
          <view class="stat-label">支付人数</view>
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <view>
          <view class="section-title">经营趋势</view>
          <view class="section-subtitle">浏览人数与下单人数按所选时间维度联动</view>
        </view>
      </view>

      <view v-if="trendChartData.length > 0" class="line-chart">
        <view class="chart-legend">
          <view class="legend-item">
            <view class="legend-dot browse"></view>
            <text>浏览人数</text>
          </view>
          <view class="legend-item">
            <view class="legend-dot order"></view>
            <text>下单人数</text>
          </view>
        </view>

        <view class="chart-plot">
          <view v-for="line in 4" :key="line" class="chart-grid-line" :style="{ bottom: `${line * 25}%` }"></view>

          <view
            v-for="(item, index) in trendChartData"
            :key="item.date"
            class="chart-column"
          >
            <view
              v-if="index < trendChartData.length - 1"
              class="chart-line browse"
              :style="getTrendLineStyle(item.visit_users, trendChartData[index + 1].visit_users, 'browse')"
            ></view>
            <view
              v-if="index < trendChartData.length - 1"
              class="chart-line order"
              :style="getTrendLineStyle(item.submit_order_users, trendChartData[index + 1].submit_order_users, 'order')"
            ></view>

            <view
              class="chart-point browse"
              :style="{ bottom: `${getTrendBottom(item.visit_users)}%` }"
            ></view>
            <view
              class="chart-point order"
              :style="{ bottom: `${getTrendBottom(item.submit_order_users)}%` }"
            ></view>
            <text class="chart-label">{{ item.label }}</text>
          </view>
        </view>
      </view>

      <view v-else class="empty-ranking">
        <text>暂无趋势数据</text>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <view>
          <view class="section-title">商品销量排行</view>
          <view class="section-subtitle">{{ currentRangeLabel }}</view>
        </view>
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
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  getSalesOverview,
  getSalesTrend,
  getProductRanking,
  getStockAlert
} from '@api'
import type {
  SalesOverview,
  SalesTrend,
  ProductRanking,
  StockAlert
} from '@types'

type PeriodValue = 'today' | 'week' | 'month' | 'year'

const periodTabs = computed(() => [
  { label: '今日', value: 'today' as PeriodValue, active: currentPeriod.value === 'today' },
  { label: '本周', value: 'week' as PeriodValue, active: currentPeriod.value === 'week' },
  { label: '本月', value: 'month' as PeriodValue, active: currentPeriod.value === 'month' },
  { label: '本年', value: 'year' as PeriodValue, active: currentPeriod.value === 'year' }
])

const CHART_HEIGHT = 220
const CHART_COLUMN_WIDTH = 44

const currentPeriod = ref<PeriodValue>('today')
const overview = ref<SalesOverview | null>(null)
const salesTrend = ref<SalesTrend[]>([])
const productRanking = ref<ProductRanking[]>([])
const stockAlerts = ref<StockAlert[]>([])

const dateRange = reactive({
  start_date: '',
  end_date: '',
  granularity: 'day'
})

onShow(() => {
  applyDateRange(currentPeriod.value)
  loadData()
})

const currentRangeLabel = computed(() => `${dateRange.start_date} 至 ${dateRange.end_date}`)

const trendChartData = computed(() =>
  salesTrend.value.map((item) => ({
    date: item.date,
    label: formatChartLabel(item.date),
    visit_users: Number(item.visit_users ?? item.customers ?? 0),
    submit_order_users: Number(item.submit_order_users ?? item.orders ?? 0)
  }))
)

const maxTrendValue = computed(() => {
  const values = trendChartData.value.flatMap((item) => [item.visit_users, item.submit_order_users])
  return Math.max(...values, 1)
})

async function loadData() {
  await Promise.all([
    loadOverview(),
    loadSalesTrend(),
    loadProductRanking(),
    loadStockAlerts()
  ])
}

async function loadOverview() {
  try {
    overview.value = await getSalesOverview({ period: currentPeriod.value })
  } catch (error) {
    console.error('加载销售概览失败:', error)
  }
}

async function loadSalesTrend() {
  try {
    salesTrend.value = await getSalesTrend({
      start_date: dateRange.start_date,
      end_date: dateRange.end_date,
      granularity: 'day'
    })
  } catch (error) {
    console.error('加载销售趋势失败:', error)
  }
}

async function loadProductRanking() {
  try {
    productRanking.value = await getProductRanking({
      start_date: dateRange.start_date,
      end_date: dateRange.end_date,
      limit: 5
    })
  } catch (error) {
    console.error('加载商品排行失败:', error)
  }
}

async function loadStockAlerts() {
  try {
    stockAlerts.value = await getStockAlert({ threshold: 10 })
  } catch (error) {
    console.error('加载库存预警失败:', error)
  }
}

function changePeriod(period: PeriodValue) {
  currentPeriod.value = period
  applyDateRange(period)
  loadData()
}

function getTodayDate(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
}

function formatDate(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function applyDateRange(period: PeriodValue) {
  const now = new Date()
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  let start = new Date(end)

  if (period === 'week') {
    start.setDate(end.getDate() - 6)
  } else if (period === 'month') {
    start = new Date(end.getFullYear(), end.getMonth(), 1)
  } else if (period === 'year') {
    start = new Date(end.getFullYear(), 0, 1)
  }

  dateRange.start_date = formatDate(start)
  dateRange.end_date = formatDate(end)
}

function formatAmount(amount: number): string {
  return amount.toFixed(2)
}

function formatChartLabel(date: string) {
  if (currentPeriod.value === 'year') {
    return date.slice(5, 7)
  }
  return date.slice(5)
}

function getTrendBottom(value: number) {
  if (!value) {
    return 0
  }
  return (value / maxTrendValue.value) * 100
}

function getTrendLineStyle(currentValue: number, nextValue: number, type: 'browse' | 'order') {
  const currentBottom = getTrendBottom(currentValue)
  const nextBottom = getTrendBottom(nextValue)
  const deltaY = ((currentBottom - nextBottom) / 100) * CHART_HEIGHT
  const angle = Math.atan2(deltaY, CHART_COLUMN_WIDTH) * 180 / Math.PI
  const length = Math.sqrt(CHART_COLUMN_WIDTH * CHART_COLUMN_WIDTH + deltaY * deltaY)

  return {
    bottom: `${currentBottom}%`,
    width: `${length}px`,
    transform: `rotate(${angle}deg)`,
    background: type === 'browse' ? '#5b8cff' : '#52c41a'
  }
}

function getRankClass(index: number): string {
  const classMap = ['gold', 'silver', 'bronze']
  return index < 3 ? classMap[index] : ''
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
  transition: all 0.2s ease;
}

.period-tab-active {
  background: #ffffff !important;
  color: #007AFF !important;
  font-weight: 600 !important;
}

.range-text {
  font-size: 24rpx;
  opacity: 0.9;
  margin-bottom: 20rpx;
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

.overview-extra {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.extra-item {
  flex: 1;
  background: rgba(255, 255, 255, 0.12);
  border-radius: 16rpx;
  padding: 16rpx 12rpx;
  text-align: center;
}

.extra-label {
  display: block;
  font-size: 22rpx;
  opacity: 0.85;
  margin-bottom: 8rpx;
}

.extra-value {
  font-size: 30rpx;
  font-weight: 600;
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

.section-subtitle {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.line-chart {
  margin-top: 8rpx;
}

.chart-legend {
  display: flex;
  gap: 24rpx;
  margin-bottom: 20rpx;
}

.legend-item {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  color: #666666;
}

.legend-dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  margin-right: 10rpx;
}

.legend-dot.browse {
  background: #5b8cff;
}

.legend-dot.order {
  background: #52c41a;
}

.chart-plot {
  position: relative;
  display: flex;
  align-items: flex-end;
  height: 320rpx;
  padding: 24rpx 0 56rpx;
}

.chart-grid-line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1rpx dashed #eaeaea;
}

.chart-column {
  position: relative;
  flex: 1;
  height: 100%;
}

.chart-line {
  position: absolute;
  left: 50%;
  height: 4rpx;
  border-radius: 999rpx;
  transform-origin: left center;
  opacity: 0.95;
}

.chart-point {
  position: absolute;
  left: 50%;
  width: 18rpx;
  height: 18rpx;
  margin-left: -9rpx;
  border-radius: 50%;
  border: 4rpx solid #ffffff;
  box-sizing: border-box;
}

.chart-point.browse {
  background: #5b8cff;
}

.chart-point.order {
  background: #52c41a;
}

.chart-label {
  position: absolute;
  bottom: -40rpx;
  left: 50%;
  transform: translateX(-50%);
  font-size: 22rpx;
  color: #999999;
  white-space: nowrap;
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
