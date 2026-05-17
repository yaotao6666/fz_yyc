<template>
  <view class="analytics-container">
    <view class="top-tabs">
      <view
        v-for="tab in topTabs"
        :key="tab.key"
        class="top-tab"
        :class="{ active: currentTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </view>
    </view>

    <scroll-view class="tab-content-scroll" scroll-y @scrolltolower="onScrollToLower">
      <view v-if="currentTab === 'overview'" class="tab-panel">
        <view class="section overview-section">
          <view class="section-title">商家经营总览</view>
          <view class="overview-grid">
            <view class="overview-card">
              <text class="overview-value">{{ merchantStats.totals.merchant_count }}</text>
              <text class="overview-label">商家数</text>
            </view>
            <view class="overview-card">
              <text class="overview-value">{{ merchantStats.totals.visit_users }}</text>
              <text class="overview-label">访问用户数</text>
            </view>
            <view class="overview-card">
              <text class="overview-value">{{ merchantStats.totals.order_users }}</text>
              <text class="overview-label">下单用户数</text>
            </view>
            <view class="overview-card">
              <text class="overview-value">¥{{ formatAmount(merchantStats.totals.order_amount) }}</text>
              <text class="overview-label">下单金额</text>
            </view>
          </view>
        </view>

        <view class="section">
          <view class="section-header">
            <text class="section-title">订单量分析</text>
            <view class="period-tabs">
              <view
                v-for="period in periodOptions"
                :key="period.value"
                class="period-tab"
                :class="{ active: currentPeriod === period.value }"
                @click="currentPeriod = period.value"
              >
                {{ period.label }}
              </view>
            </view>
          </view>
          <view v-if="loadingOrders" class="loading-state">加载中...</view>
          <view v-else-if="currentPeriodList.length === 0" class="empty-state">暂无订单趋势数据</view>
          <view v-else class="period-chart">
            <view v-for="item in currentPeriodList" :key="item.label" class="chart-item">
              <view class="chart-bar-bg">
                <view class="chart-bar" :style="{ height: `${getBarHeight(item.order_count)}%` }"></view>
              </view>
              <text class="chart-count">{{ item.order_count }}</text>
              <text class="chart-label">{{ item.label }}</text>
            </view>
          </view>
        </view>
      </view>

      <view v-if="currentTab === 'merchants'" class="tab-panel">
        <view class="section">
          <view class="section-header">
            <text class="section-title">商家转化分析</text>
            <view class="metric-tabs">
              <view
                v-for="metric in metricOptions"
                :key="metric.value"
                class="metric-tab"
                :class="{ active: currentMetric === metric.value }"
                @click="switchMetric(metric.value)"
              >
                {{ metric.label }}
              </view>
            </view>
          </view>

          <view v-if="loadingOverview" class="loading-state">加载中...</view>
          <view v-else-if="merchantStats.merchants.length === 0" class="empty-state">暂无商家分析数据</view>
          <view v-else class="merchant-list">
            <view v-for="merchant in merchantStats.merchants" :key="merchant.merchant_id" class="merchant-row">
              <view class="merchant-main">
                <text class="merchant-name">{{ merchant.merchant_name }}</text>
                <text class="merchant-sub">访问 {{ merchant.visit_users }} 人 · 下单 {{ merchant.order_users }} 人 · 支付 {{ merchant.paid_orders }} 单</text>
              </view>
              <view class="merchant-metric">
                <text class="metric-value">{{ formatMetricValue(merchant, currentMetric) }}</text>
                <text class="metric-detail">金额 ¥{{ formatAmount(merchant.order_amount) }}</text>
              </view>
            </view>
          </view>
          <view v-if="loadingMoreMerchants" class="list-bottom">加载中...</view>
          <view v-else-if="noMoreMerchants && merchantStats.merchants.length > 0" class="list-bottom">没有更多了</view>
        </view>
      </view>

      <view v-if="currentTab === 'ranking'" class="tab-panel">
        <view class="section">
          <view class="section-header">
            <text class="section-title">商家排行榜</text>
            <view class="metric-tabs">
              <view
                v-for="metric in metricOptions"
                :key="`rank-${metric.value}`"
                class="metric-tab"
                :class="{ active: rankingMetric === metric.value }"
                @click="changeRankingMetric(metric.value)"
              >
                {{ metric.label }}
              </view>
            </view>
          </view>

          <view v-if="loadingRanking" class="loading-state">加载中...</view>
          <view v-else-if="rankings.length === 0" class="empty-state">暂无排行榜数据</view>
          <view v-else class="ranking-list">
            <view v-for="item in rankings" :key="item.merchant_id" class="ranking-row">
              <view class="ranking-rank" :class="getRankClass(item.rank)">{{ item.rank }}</view>
              <view class="ranking-main">
                <text class="merchant-name">{{ item.merchant_name }}</text>
                <text class="merchant-sub">访问率 {{ formatPercent(item.visit_rate) }} · 下单率 {{ formatPercent(item.order_rate) }}</text>
              </view>
              <view class="merchant-metric">
                <text class="metric-value">{{ formatMetricValue(item, rankingMetric) }}</text>
                <text class="metric-detail">均价 ¥{{ formatAmount(item.avg_order_amount) }}</text>
              </view>
            </view>
          </view>
          <view v-if="loadingMoreRanking" class="list-bottom">加载中...</view>
          <view v-else-if="noMoreRanking && rankings.length > 0" class="list-bottom">没有更多了</view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getMerchantDistribution, getOrderAnalytics, getTopMerchants } from '@api'
import type { MerchantDistributionData, OrderAnalyticsData, SpMerchantConversionItem, TopMerchantRanking } from '@types'

const PAGE_SIZE = 6

const topTabs = [
  { key: 'overview', label: '总览' },
  { key: 'merchants', label: '商家分析' },
  { key: 'ranking', label: '排行榜' }
] as const

type TabKey = typeof topTabs[number]['key']

const metricOptions = [
  { label: '访问率', value: 'visit_rate' },
  { label: '下单率', value: 'order_rate' },
  { label: '下单金额', value: 'order_amount' },
  { label: '下单均价', value: 'avg_order_amount' }
] as const

const periodOptions = [
  { label: '日', value: 'day' },
  { label: '周', value: 'week' },
  { label: '月', value: 'month' },
  { label: '年', value: 'year' }
] as const

type MetricValue = typeof metricOptions[number]['value']
type PeriodValue = typeof periodOptions[number]['value']

const currentTab = ref<TabKey>('overview')
const currentMetric = ref<MetricValue>('visit_rate')
const rankingMetric = ref<MetricValue>('order_amount')
const currentPeriod = ref<PeriodValue>('day')

const loadingOverview = ref(false)
const loadingOrders = ref(false)
const loadingRanking = ref(false)
const loadingMoreMerchants = ref(false)
const loadingMoreRanking = ref(false)

const merchantPage = ref(1)
const totalMerchants = ref(0)
const rankingPage = ref(1)
const totalRankings = ref(0)

const loadedTabs = ref<Set<TabKey>>(new Set(['overview']))

const merchantStats = ref<MerchantDistributionData>({
  merchants: [],
  totals: {
    merchant_count: 0,
    visit_users: 0,
    order_users: 0,
    paid_orders: 0,
    order_amount: 0
  }
})

const orderAnalytics = ref<OrderAnalyticsData>({
  day: [],
  week: [],
  month: [],
  year: []
})

const rankings = ref<TopMerchantRanking[]>([])

const noMoreMerchants = computed(() => {
  return merchantStats.value.merchants.length >= totalMerchants.value && totalMerchants.value > 0
})

const noMoreRanking = computed(() => {
  return rankings.value.length >= totalRankings.value && totalRankings.value > 0
})

const currentPeriodList = computed(() => orderAnalytics.value[currentPeriod.value] || [])

onMounted(() => {
  loadOverview()
  loadOrderAnalytics()
})

function switchTab(tab: TabKey) {
  if (currentTab.value === tab) return
  currentTab.value = tab
  if (!loadedTabs.value.has(tab)) {
    loadedTabs.value.add(tab)
    if (tab === 'merchants') loadOverview()
    if (tab === 'ranking') loadRankings()
  }
}

async function loadOverview() {
  loadingOverview.value = true
  merchantPage.value = 1
  try {
    const data = await getMerchantDistribution({
      page: 1,
      page_size: PAGE_SIZE,
      sort_by: currentMetric.value,
      sort_order: 'desc'
    })
    merchantStats.value = data
    totalMerchants.value = data.pagination?.total || data.merchants.length
  } catch (error: any) {
    uni.showToast({ title: error?.message || '加载商家分析失败', icon: 'none' })
    merchantStats.value = {
      merchants: [],
      totals: { merchant_count: 0, visit_users: 0, order_users: 0, paid_orders: 0, order_amount: 0 }
    }
    totalMerchants.value = 0
  } finally {
    loadingOverview.value = false
  }
}

async function loadMoreMerchants() {
  if (loadingMoreMerchants.value || noMoreMerchants.value) return
  loadingMoreMerchants.value = true
  merchantPage.value++
  try {
    const data = await getMerchantDistribution({
      page: merchantPage.value,
      page_size: PAGE_SIZE,
      sort_by: currentMetric.value,
      sort_order: 'desc'
    })
    merchantStats.value = {
      ...data,
      merchants: [...merchantStats.value.merchants, ...data.merchants]
    }
    totalMerchants.value = data.pagination?.total || totalMerchants.value
  } catch (error: any) {
    merchantPage.value--
    uni.showToast({ title: error?.message || '加载更多失败', icon: 'none' })
  } finally {
    loadingMoreMerchants.value = false
  }
}

async function loadOrderAnalytics() {
  loadingOrders.value = true
  try {
    orderAnalytics.value = await getOrderAnalytics()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '加载订单趋势失败', icon: 'none' })
    orderAnalytics.value = { day: [], week: [], month: [], year: [] }
  } finally {
    loadingOrders.value = false
  }
}

async function loadRankings() {
  loadingRanking.value = true
  rankingPage.value = 1
  try {
    const data = await getTopMerchants({ limit: PAGE_SIZE, metric: rankingMetric.value })
    const list = Array.isArray(data) ? data : []
    rankings.value = list
    totalRankings.value = list.length < PAGE_SIZE ? list.length : list.length + PAGE_SIZE
  } catch (error: any) {
    uni.showToast({ title: error?.message || '加载排行榜失败', icon: 'none' })
    rankings.value = []
    totalRankings.value = 0
  } finally {
    loadingRanking.value = false
  }
}

async function loadMoreRankings() {
  if (loadingMoreRanking.value || noMoreRanking.value) return
  loadingMoreRanking.value = true
  rankingPage.value++
  try {
    const data = await getTopMerchants({ limit: PAGE_SIZE, metric: rankingMetric.value, page: rankingPage.value })
    const list = Array.isArray(data) ? data : []
    rankings.value = [...rankings.value, ...list]
    if (list.length < PAGE_SIZE) {
      totalRankings.value = rankings.value.length
    } else {
      totalRankings.value = rankings.value.length + PAGE_SIZE
    }
  } catch (error: any) {
    rankingPage.value--
    uni.showToast({ title: error?.message || '加载更多失败', icon: 'none' })
  } finally {
    loadingMoreRanking.value = false
  }
}

function onScrollToLower() {
  if (currentTab.value === 'merchants') {
    loadMoreMerchants()
  } else if (currentTab.value === 'ranking') {
    loadMoreRankings()
  }
}

function switchMetric(metric: MetricValue) {
  currentMetric.value = metric
  loadOverview()
}

function changeRankingMetric(metric: MetricValue) {
  rankingMetric.value = metric
  loadRankings()
}

function getMetricNumber(item: Pick<SpMerchantConversionItem, MetricValue>, metric: MetricValue): number {
  return Number(item[metric] || 0)
}

function formatAmount(value: number): string {
  return Number(value || 0).toFixed(2)
}

function formatPercent(value: number): string {
  return `${Number(value || 0).toFixed(1)}%`
}

function formatMetricValue(item: Pick<SpMerchantConversionItem, MetricValue>, metric: MetricValue): string {
  const value = getMetricNumber(item, metric)
  if (metric === 'order_amount' || metric === 'avg_order_amount') {
    return `¥${formatAmount(value)}`
  }
  return formatPercent(value)
}

function getBarHeight(value: number): number {
  if (currentPeriodList.value.length === 0) return 0
  const maxValue = Math.max(...currentPeriodList.value.map(item => item.order_count), 0)
  if (maxValue <= 0) return 8
  return Math.max((value / maxValue) * 100, 8)
}

function getRankClass(rank: number): string {
  if (rank === 1) return 'gold'
  if (rank === 2) return 'silver'
  if (rank === 3) return 'bronze'
  return ''
}
</script>

<style scoped>
.analytics-container {
  min-height: 100vh;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.top-tabs {
  display: flex;
  background: #ffffff;
  padding: 0 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
  flex-shrink: 0;
}

.top-tab {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  font-size: 28rpx;
  color: #666666;
  position: relative;
}

.top-tab.active {
  color: #2f54eb;
  font-weight: 600;
}

.top-tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 6rpx;
  background: #2f54eb;
  border-radius: 3rpx;
}

.tab-content-scroll {
  flex: 1;
  min-height: 0;
}

.tab-panel {
  padding: 24rpx;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
}

.section {
  margin-bottom: 24rpx;
  padding: 28rpx;
  background: #ffffff;
  border-radius: 20rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f1f1f;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.overview-card {
  padding: 24rpx;
  background: linear-gradient(135deg, #f6f8ff 0%, #eef3ff 100%);
  border-radius: 16rpx;
}

.overview-value {
  display: block;
  font-size: 36rpx;
  font-weight: 600;
  color: #2f54eb;
  margin-bottom: 8rpx;
}

.overview-label {
  display: block;
  font-size: 24rpx;
  color: #666666;
}

.metric-tabs,
.period-tabs {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.metric-tab,
.period-tab {
  padding: 10rpx 20rpx;
  border-radius: 999rpx;
  background: #f4f4f5;
  font-size: 24rpx;
  color: #666666;
}

.metric-tab.active,
.period-tab.active {
  background: #2f54eb;
  color: #ffffff;
}

.merchant-list,
.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.merchant-row,
.ranking-row {
  display: flex;
  align-items: center;
  gap: 18rpx;
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.merchant-row:last-child,
.ranking-row:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.merchant-main,
.ranking-main {
  flex: 1;
  min-width: 0;
}

.merchant-name {
  display: block;
  font-size: 28rpx;
  color: #1f1f1f;
  font-weight: 500;
  margin-bottom: 8rpx;
}

.merchant-sub {
  display: block;
  font-size: 24rpx;
  color: #8c8c8c;
  line-height: 1.6;
}

.merchant-metric {
  text-align: right;
}

.metric-value {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #2f54eb;
  margin-bottom: 8rpx;
}

.metric-detail {
  display: block;
  font-size: 22rpx;
  color: #8c8c8c;
}

.period-chart {
  display: flex;
  align-items: flex-end;
  gap: 16rpx;
  min-height: 360rpx;
}

.chart-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 12rpx;
}

.chart-bar-bg {
  width: 100%;
  height: 220rpx;
  background: #f0f5ff;
  border-radius: 16rpx 16rpx 0 0;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
}

.chart-bar {
  width: 100%;
  background: linear-gradient(180deg, #85a5ff 0%, #2f54eb 100%);
  border-radius: 16rpx 16rpx 0 0;
}

.chart-count,
.chart-label {
  font-size: 22rpx;
  color: #666666;
}

.ranking-rank {
  width: 52rpx;
  height: 52rpx;
  border-radius: 50%;
  background: #f0f0f0;
  color: #666666;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  font-weight: 600;
}

.ranking-rank.gold {
  background: #fff7e6;
  color: #d48806;
}

.ranking-rank.silver {
  background: #f5f5f5;
  color: #595959;
}

.ranking-rank.bronze {
  background: #fff2e8;
  color: #d46b08;
}

.loading-state,
.empty-state {
  padding: 80rpx 0;
  text-align: center;
  font-size: 26rpx;
  color: #8c8c8c;
}

.list-bottom {
  padding: 24rpx 0;
  text-align: center;
  font-size: 26rpx;
  color: #8c8c8c;
}
</style>
