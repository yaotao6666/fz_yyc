<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getAmountAnalytics, getMerchantDistribution, getOrderAnalytics, getTopMerchants } from '@/api/sp'
import type { AmountTrendData, MerchantDistributionData, OrderAnalyticsData, TopMerchantRanking } from '@/types/sp'
import { formatAmount, formatPercent } from '@/utils/format'

const currentMetric = ref<'visit_rate' | 'order_rate' | 'order_amount' | 'avg_order_amount'>('visit_rate')
const rankingMetric = ref<'visit_rate' | 'order_rate' | 'order_amount' | 'avg_order_amount'>('order_amount')
const currentPeriod = ref<'day' | 'week' | 'month' | 'year'>('day')
const loading = ref(false)

const merchantStats = ref<MerchantDistributionData>({
  merchants: [],
  totals: {
    merchant_count: 0,
    visit_users: 0,
    order_users: 0,
    paid_orders: 0,
    order_amount: 0,
  },
  pagination: {
    total: 0,
    page: 1,
    page_size: 6,
  }
})

const orderAnalytics = ref<OrderAnalyticsData>({
  day: [],
  week: [],
  month: [],
  year: [],
})

const amountTrends = ref<AmountTrendData>({ trends: [] })
const rankings = ref<TopMerchantRanking[]>([])

const currentPeriodList = computed(() => orderAnalytics.value[currentPeriod.value] || [])

function metricValue(item: MerchantDistributionData['merchants'][number] | TopMerchantRanking, metric: typeof currentMetric.value) {
  const value = Number(item[metric] || 0)
  if (metric === 'order_amount' || metric === 'avg_order_amount') {
    return `¥${formatAmount(value)}`
  }
  return formatPercent(value)
}

function getBarWidth(value: number, max: number) {
  if (max <= 0) return '8%'
  return `${Math.max((value / max) * 100, 8)}%`
}

async function loadData() {
  loading.value = true
  try {
    const [distribution, orderData, amountData, rankingData] = await Promise.all([
      getMerchantDistribution({ page: 1, page_size: 10, sort_by: currentMetric.value, sort_order: 'desc' }),
      getOrderAnalytics(),
      getAmountAnalytics({ days: 14 }),
      getTopMerchants({ limit: 10, metric: rankingMetric.value })
    ])
    merchantStats.value = distribution
    orderAnalytics.value = orderData
    amountTrends.value = amountData
    rankings.value = rankingData
  } finally {
    loading.value = false
  }
}

function reloadDistribution(metric: typeof currentMetric.value) {
  currentMetric.value = metric
  loadData()
}

function reloadRanking(metric: typeof rankingMetric.value) {
  rankingMetric.value = metric
  loadData()
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">数据分析</h1>
        <p class="page-subtitle">查看订单趋势、商家转化排行和金额走势。</p>
      </div>
    </div>

    <el-skeleton :rows="10" animated :loading="loading">
      <div class="metric-grid">
        <div class="metric-card">
          <div class="metric-label">商家数</div>
          <div class="metric-value">{{ merchantStats.totals.merchant_count }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">访问用户数</div>
          <div class="metric-value">{{ merchantStats.totals.visit_users }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">下单用户数</div>
          <div class="metric-value">{{ merchantStats.totals.order_users }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">下单金额</div>
          <div class="metric-value">¥{{ formatAmount(merchantStats.totals.order_amount) }}</div>
        </div>
      </div>

      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>
            <span>订单趋势</span>
          </template>
          <el-radio-group v-model="currentPeriod" size="small" style="margin-bottom: 16px;">
            <el-radio-button value="day">日</el-radio-button>
            <el-radio-button value="week">周</el-radio-button>
            <el-radio-button value="month">月</el-radio-button>
            <el-radio-button value="year">年</el-radio-button>
          </el-radio-group>
          <div v-if="currentPeriodList.length" class="simple-list">
            <div v-for="item in currentPeriodList" :key="item.label" class="simple-list-item">
              <div style="min-width: 80px;">{{ item.label }}</div>
              <div class="progress-row" style="flex: 1;">
                <div class="progress-track">
                  <div class="progress-bar" :style="{ width: getBarWidth(item.order_count, Math.max(...currentPeriodList.map((row) => row.order_count), 1)) }"></div>
                </div>
                <strong>{{ item.order_count }}</strong>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无订单趋势数据" />
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <span>金额趋势（近 14 天）</span>
          </template>
          <div v-if="amountTrends.trends.length" class="simple-list">
            <div v-for="item in amountTrends.trends" :key="item.date" class="simple-list-item">
              <span>{{ item.date }}</span>
              <strong>¥{{ formatAmount(item.amount) }}</strong>
            </div>
          </div>
          <el-empty v-else description="暂无金额趋势数据" />
        </el-card>
      </div>

      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between; gap: 12px;">
              <span>商家分析</span>
              <el-select :model-value="currentMetric" style="width: 160px;" @change="reloadDistribution">
                <el-option label="访问率" value="visit_rate" />
                <el-option label="下单率" value="order_rate" />
                <el-option label="下单金额" value="order_amount" />
                <el-option label="下单均价" value="avg_order_amount" />
              </el-select>
            </div>
          </template>
          <el-table :data="merchantStats.merchants" size="small">
            <el-table-column prop="merchant_name" label="商家" min-width="140" />
            <el-table-column label="访问/下单/支付" min-width="170">
              <template #default="scope">
                {{ scope.row.visit_users }} / {{ scope.row.order_users }} / {{ scope.row.paid_orders }}
              </template>
            </el-table-column>
            <el-table-column label="当前指标" min-width="120">
              <template #default="scope">
                {{ metricValue(scope.row, currentMetric) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between; gap: 12px;">
              <span>排行榜</span>
              <el-select :model-value="rankingMetric" style="width: 160px;" @change="reloadRanking">
                <el-option label="访问率" value="visit_rate" />
                <el-option label="下单率" value="order_rate" />
                <el-option label="下单金额" value="order_amount" />
                <el-option label="下单均价" value="avg_order_amount" />
              </el-select>
            </div>
          </template>
          <div v-if="rankings.length" class="simple-list">
            <div v-for="item in rankings" :key="item.merchant_id" class="simple-list-item">
              <div style="display: flex; align-items: center; gap: 12px;">
                <span class="ranking-badge" :class="{ 'top-1': item.rank === 1, 'top-2': item.rank === 2, 'top-3': item.rank === 3 }">{{ item.rank }}</span>
                <div>
                  <div>{{ item.merchant_name }}</div>
                  <small style="color: #6b7280;">访问率 {{ formatPercent(item.visit_rate) }} · 下单率 {{ formatPercent(item.order_rate) }}</small>
                </div>
              </div>
              <strong>{{ metricValue(item, rankingMetric) }}</strong>
            </div>
          </div>
          <el-empty v-else description="暂无排行榜数据" />
        </el-card>
      </div>
    </el-skeleton>
  </div>
</template>
