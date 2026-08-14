<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getAnalyticsOverview, getProductRanking, getSalesTrend } from '@/api/sp'
import { formatAmount } from '@/utils/format'

type Period = 'today' | 'week' | 'month' | 'year'

const period = ref<Period>('week')
const loading = ref(false)

interface OverviewData {
  total_sales: number
  total_orders: number
  total_customers: number
  avg_order_amount: number
  sales_growth: number
  orders_growth: number
  customers_growth: number
  visit_count: number
  visit_users: number
  pay_success_users: number
}

interface TrendItem {
  date: string
  orders: number
  sales: number
  customers: number
  visit_users: number
  submit_order_users: number
}

interface RankingItem {
  product_id: number
  product_name: string
  image: string
  sales_count: number
  sales_amount: number
}

const overview = ref<OverviewData>({
  total_sales: 0,
  total_orders: 0,
  total_customers: 0,
  avg_order_amount: 0,
  sales_growth: 0,
  orders_growth: 0,
  customers_growth: 0,
  visit_count: 0,
  visit_users: 0,
  pay_success_users: 0
})

const trends = ref<TrendItem[]>([])
const rankings = ref<RankingItem[]>([])

const periodOptions = [
  { label: '今日', value: 'today' },
  { label: '近7天', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '本年', value: 'year' }
]

const maxTrendOrders = computed(() => Math.max(...trends.value.map((t) => Number(t.orders || 0)), 1))
const maxTrendSales = computed(() => Math.max(...trends.value.map((t) => Number(t.sales || 0)), 1))

function growthText(value: number) {
  if (!Number.isFinite(value)) return '-'
  const sign = value >= 0 ? '+' : ''
  return `${sign}${value.toFixed(1)}%`
}

function growthType(value: number) {
  return value >= 0 ? 'success' : 'danger'
}

async function loadData() {
  loading.value = true
  try {
    const days = period.value === 'today' ? 1 : period.value === 'week' ? 7 : period.value === 'month' ? 30 : 365
    const [overviewData, trendData, rankingData] = await Promise.all([
      getAnalyticsOverview({ period: period.value }),
      getSalesTrend({ days }),
      getProductRanking({ limit: 10 })
    ])
    overview.value = overviewData as unknown as OverviewData
    trends.value = (trendData as unknown as TrendItem[]) || []
    rankings.value = (rankingData as unknown as RankingItem[]) || []
  } finally {
    loading.value = false
  }
}

function getOrdersBarWidth(orders: number) {
  return `${Math.max((Number(orders || 0) / maxTrendOrders.value) * 100, 6)}%`
}

function getSalesBarWidth(sales: number) {
  return `${Math.max((Number(sales || 0) / maxTrendSales.value) * 100, 6)}%`
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">数据分析</h1>
        <p class="page-subtitle">查看经营概览、销售趋势与商品排行。</p>
      </div>
      <el-radio-group v-model="period" size="small" @change="loadData">
        <el-radio-button v-for="opt in periodOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</el-radio-button>
      </el-radio-group>
    </div>

    <el-skeleton :rows="10" animated :loading="loading">
      <div class="metric-grid">
        <div class="metric-card">
          <div class="metric-label">成交金额</div>
          <div class="metric-value">¥{{ formatAmount(overview.total_sales) }}</div>
          <div class="metric-growth">
            <el-tag :type="growthType(overview.sales_growth)" size="small">{{ growthText(overview.sales_growth) }}</el-tag>
            <span class="growth-hint">环比</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">订单数</div>
          <div class="metric-value">{{ overview.total_orders }}</div>
          <div class="metric-growth">
            <el-tag :type="growthType(overview.orders_growth)" size="small">{{ growthText(overview.orders_growth) }}</el-tag>
            <span class="growth-hint">环比</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">下单用户</div>
          <div class="metric-value">{{ overview.total_customers }}</div>
          <div class="metric-growth">
            <el-tag :type="growthType(overview.customers_growth)" size="small">{{ growthText(overview.customers_growth) }}</el-tag>
            <span class="growth-hint">环比</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">客单价</div>
          <div class="metric-value">¥{{ formatAmount(overview.avg_order_amount) }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">到访次数</div>
          <div class="metric-value">{{ overview.visit_count }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">到访用户</div>
          <div class="metric-value">{{ overview.visit_users }}</div>
        </div>
      </div>

      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>订单趋势</template>
          <div v-if="trends.length" class="trend-list">
            <div v-for="item in trends" :key="item.date" class="trend-row">
              <div class="trend-date">{{ item.date }}</div>
              <div class="trend-bar-wrap">
                <div class="progress-track">
                  <div class="progress-bar orders-bar" :style="{ width: getOrdersBarWidth(item.orders) }"></div>
                </div>
                <strong>{{ item.orders }} 单</strong>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无订单趋势数据" />
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>销售额趋势</template>
          <div v-if="trends.length" class="trend-list">
            <div v-for="item in trends" :key="item.date" class="trend-row">
              <div class="trend-date">{{ item.date }}</div>
              <div class="trend-bar-wrap">
                <div class="progress-track">
                  <div class="progress-bar sales-bar" :style="{ width: getSalesBarWidth(item.sales) }"></div>
                </div>
                <strong>¥{{ formatAmount(item.sales) }}</strong>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无销售额趋势数据" />
        </el-card>
      </div>

      <el-card class="page-card" shadow="never" style="margin-top: 20px;">
        <template #header>商品销量排行（Top 10）</template>
        <el-table v-if="rankings.length" :data="rankings" size="small">
          <el-table-column label="排名" width="70" type="index" />
          <el-table-column label="商品" min-width="220">
            <template #default="scope">
              <div class="rank-product">
                <img v-if="scope.row.image" :src="scope.row.image" class="rank-image" alt="" />
                <span>{{ scope.row.product_name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="销量" width="120">
            <template #default="scope">
              {{ scope.row.sales_count }}
            </template>
          </el-table-column>
          <el-table-column label="销售额" width="140">
            <template #default="scope">
              ¥{{ formatAmount(scope.row.sales_amount) }}
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无商品排行数据" />
      </el-card>
    </el-skeleton>
  </div>
</template>

<style scoped>
.metric-growth {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.growth-hint {
  font-size: 12px;
  color: #9ca3af;
}

.section-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.trend-list {
  display: grid;
  gap: 12px;
  max-height: 360px;
  overflow-y: auto;
}

.trend-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.trend-date {
  min-width: 96px;
  color: #6b7280;
  font-size: 13px;
}

.trend-bar-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-track {
  flex: 1;
  height: 8px;
  background: #f1f5f9;
  border-radius: 999px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s;
}

.orders-bar {
  background: #3b82f6;
}

.sales-bar {
  background: #10b981;
}

.rank-product {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rank-image {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  object-fit: cover;
}

@media (max-width: 1200px) {
  .section-grid {
    grid-template-columns: 1fr;
  }
}
</style>
