<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard } from '@/api/sp'
import type { DashboardData } from '@/types/sp'
import { formatAmount } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const dashboard = ref<DashboardData>({
  total_orders: 0,
  total_amount: 0,
  today_orders: 0,
  today_amount: 0,
  pending_orders: 0,
  completed_orders: 0,
  refunded_amount: 0
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

onMounted(loadDashboard)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">商家工作台</h1>
        <p class="page-subtitle">查看经营总览，快速进入订单、商品与服务人员管理。</p>
      </div>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <div class="metric-grid">
        <div class="metric-card">
          <div class="metric-label">今日订单</div>
          <div class="metric-value">{{ dashboard.today_orders }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">今日成交</div>
          <div class="metric-value">¥{{ formatAmount(dashboard.today_amount) }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">待处理订单</div>
          <div class="metric-value">{{ dashboard.pending_orders }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">累计订单</div>
          <div class="metric-value">{{ dashboard.total_orders }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">累计成交</div>
          <div class="metric-value">¥{{ formatAmount(dashboard.total_amount) }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">已完成订单</div>
          <div class="metric-value">{{ dashboard.completed_orders }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">累计退款</div>
          <div class="metric-value">¥{{ formatAmount(dashboard.refunded_amount) }}</div>
        </div>
      </div>

      <el-card class="page-card" shadow="never" style="margin-top: 20px;">
        <template #header>
          <span>快捷入口</span>
        </template>
        <div class="quick-grid">
          <div class="quick-item" @click="goTo('/orders')">
            <div class="quick-title">订单管理</div>
            <small>查看与核销订单、工单状态筛选</small>
          </div>
          <div class="quick-item" @click="goTo('/products')">
            <div class="quick-title">商品管理</div>
            <small>维护辅具、租赁、康养套餐等商品</small>
          </div>
          <div class="quick-item" @click="goTo('/staff')">
            <div class="quick-title">服务人员</div>
            <small>审核与管理接单小程序账号</small>
          </div>
          <div class="quick-item" @click="goTo('/profile')">
            <div class="quick-title">商家资料</div>
            <small>维护商家信息与支付配置</small>
          </div>
          <div class="quick-item" @click="goTo('/analytics')">
            <div class="quick-title">数据分析</div>
            <small>订单趋势与商品排行</small>
          </div>
        </div>
      </el-card>
    </el-skeleton>
  </div>
</template>

<style scoped>
.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.quick-item {
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-item:hover {
  border-color: #3b82f6;
  background: #f0f7ff;
}

.quick-title {
  font-weight: 600;
  color: #111827;
  margin-bottom: 6px;
}

.quick-item small {
  color: #6b7280;
  font-size: 13px;
}
</style>
