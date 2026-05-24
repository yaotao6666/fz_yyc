<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard } from '@/api/sp'
import type { DashboardData } from '@/types/sp'
import { formatAmount } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const dashboard = ref<DashboardData>({
  total_merchants: 0,
  today_orders: 0,
  today_revenue: 0,
  distribution: [],
  trend: []
})

async function loadDashboard() {
  loading.value = true
  try {
    dashboard.value = await getDashboard()
  } finally {
    loading.value = false
  }
}

function goTo(path: string) {
  router.push(path)
}

function getTrendWidth(orders: number) {
  const max = Math.max(...(dashboard.value.trend || []).map((item) => Number(item.orders || 0)), 1)
  return `${Math.max((Number(orders || 0) / max) * 100, 8)}%`
}

onMounted(loadDashboard)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">服务商工作台</h1>
        <p class="page-subtitle">统一查看经营总览，并快速进入商家、分账和数据分析能力。</p>
      </div>
      <el-button type="primary" @click="goTo('/merchants/new')">新增商家</el-button>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <div class="metric-grid">
        <div class="metric-card">
          <div class="metric-label">商家总数</div>
          <div class="metric-value">{{ dashboard.total_merchants }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">今日订单</div>
          <div class="metric-value">{{ dashboard.today_orders }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">今日成交</div>
          <div class="metric-value">¥{{ formatAmount(dashboard.today_revenue) }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">待补配置商家</div>
          <div class="metric-value">{{ dashboard.pending_merchants || 0 }}</div>
        </div>
      </div>

      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>
            <span>快捷入口</span>
          </template>
          <div class="simple-list">
            <div class="simple-list-item">
              <div>
                <div>商家列表</div>
                <small>统一查看资料、支付配置与经营概览</small>
              </div>
              <el-button text type="primary" @click="goTo('/merchants')">进入</el-button>
            </div>
            <div class="simple-list-item">
              <div>
                <div>分账历史</div>
                <small>按商家、状态和日期筛选抽佣记录</small>
              </div>
              <el-button text type="primary" @click="goTo('/profit-sharing')">进入</el-button>
            </div>
            <div class="simple-list-item">
              <div>
                <div>数据分析</div>
                <small>查看订单趋势、转化排行和金额走势</small>
              </div>
              <el-button text type="primary" @click="goTo('/analytics')">进入</el-button>
            </div>
            <div class="simple-list-item">
              <div>
                <div>服务商设置</div>
                <small>维护联系方式并修改登录密码</small>
              </div>
              <el-button text type="primary" @click="goTo('/settings')">进入</el-button>
            </div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <span>商家行业分布</span>
          </template>
          <div v-if="dashboard.distribution?.length" class="simple-list">
            <div v-for="item in dashboard.distribution" :key="item.category || 'unknown'" class="simple-list-item">
              <span>{{ item.category || '未分类' }}</span>
              <strong>{{ item.count }}</strong>
            </div>
          </div>
          <el-empty v-else description="暂无行业分布数据" />
        </el-card>
      </div>

      <el-card class="page-card" shadow="never" style="margin-top: 20px;">
        <template #header>
          <span>近 7 日订单趋势</span>
        </template>
        <div v-if="dashboard.trend?.length" class="simple-list">
          <div v-for="item in dashboard.trend" :key="item.date" class="simple-list-item">
            <div style="min-width: 92px;">{{ item.date }}</div>
            <div class="progress-row" style="flex: 1;">
              <div class="progress-track">
                <div class="progress-bar" :style="{ width: getTrendWidth(item.orders) }"></div>
              </div>
              <strong>{{ item.orders }}</strong>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无订单趋势数据" />
      </el-card>
    </el-skeleton>
  </div>
</template>
