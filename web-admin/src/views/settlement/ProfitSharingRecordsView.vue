<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProfitSharingRecords,
  getProfitSharingRecord,
  retryProfitSharingRecord
} from '@/api/sp'
import type { ProfitSharingRecord } from '@/types/sp'

const loading = ref(false)
const list = ref<ProfitSharingRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const statusFilter = ref<number | string>('')
const orderNo = ref('')

const status = ref<ProfitSharingRecord | null>(null)
const detailVisible = ref(false)
const detailLoading = ref(false)
const retrying = ref(false)

const statusOptions = [
  { label: '全部', value: '' },
  { label: '待分账', value: 0 },
  { label: '分账中', value: 1 },
  { label: '分账成功', value: 2 },
  { label: '分账失败', value: 3 },
  { label: '已跳过', value: 4 }
]

function statusText(s: number) {
  return { 0: '待分账', 1: '分账中', 2: '分账成功', 3: '分账失败', 4: '已跳过' }[s] || '未知'
}

function statusType(s: number) {
  return ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger', 4: 'info' } as const)[s] || 'info'
}

function receiverTypeText(t: number) {
  return t === 2 ? '个人微信' : '商户号'
}

function resultText(s?: string) {
  return { PENDING: '待分账', SUCCESS: '分账成功', CLOSED: '已关闭', FAILED: '分账失败', FINISHED: '已完成' }[s || ''] || (s || '-')
}

function clearFilters() {
  statusFilter.value = ''
  orderNo.value = ''
  page.value = 1
  loadData()
}

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value !== '') params.status = statusFilter.value
    if (orderNo.value.trim()) params.order_no = orderNo.value.trim()
    const res = await getProfitSharingRecords(params)
    list.value = res.list || []
    total.value = res.pagination?.total || 0
  } catch (e) {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

async function openDetail(row: ProfitSharingRecord) {
  detailLoading.value = true
  detailVisible.value = true
  status.value = null
  try {
    status.value = await getProfitSharingRecord(row.id)
  } catch (e) {
    // 拦截器已提示
  } finally {
    detailLoading.value = false
  }
}

async function handleRetry(row: ProfitSharingRecord) {
  try {
    await ElMessageBox.confirm(`确认重试订单「${row.order_no}」的分账？`, '重试分账', { type: 'warning' })
    retrying.value = true
    await retryProfitSharingRecord(row.id)
    ElMessage.success('已发起重试，请稍后刷新查看结果')
    loadData()
  } catch (e) {
    // 取消或失败
  } finally {
    retrying.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="profit-records-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header">
          <span class="title">分账记录</span>
          <div class="filters">
            <el-select v-model="statusFilter" placeholder="状态" style="width: 130px" @change="loadData">
              <el-option v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
            <el-input
              v-model="orderNo"
              placeholder="订单号"
              clearable
              style="width: 200px"
              @keyup.enter="loadData"
              @clear="loadData"
            />
            <el-button type="primary" @click="loadData">查询</el-button>
            <el-button @click="clearFilters">重置</el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_no" label="订单号" min-width="160" />
        <el-table-column label="订单金额" width="110">
          <template #default="{ row }">￥{{ row.total_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="分账总额" width="110">
          <template #default="{ row }">￥{{ row.total_share_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="out_order_no" label="微信分账单号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="share_time" label="分账时间" width="170">
          <template #default="{ row }">{{ row.share_time || '-' }}</template>
        </el-table-column>
        <el-table-column prop="error_message" label="失败原因" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.error_message || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">明细</el-button>
            <el-button
              v-permission="'profit:share'"
              link
              type="warning"
              :disabled="row.status === 2"
              @click="handleRetry(row)"
            >
              重试
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        background
        style="margin-top: 16px; justify-content: flex-end"
        @current-change="loadData"
        @size-change="loadData"
      />
    </el-card>

    <el-drawer v-model="detailVisible" title="分账记录明细" size="620px" :loading="detailLoading">
      <template v-if="status">
        <el-descriptions :column="2" border class="desc">
          <el-descriptions-item label="订单号">{{ status.order_no }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(status.status)">{{ statusText(status.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="订单金额">￥{{ status.total_amount?.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="分账总额">￥{{ status.total_share_amount?.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="分账时间">{{ status.share_time || '-' }}</el-descriptions-item>
          <el-descriptions-item label="分账单号">{{ status.out_order_no }}</el-descriptions-item>
          <el-descriptions-item label="交易单号" :span="2">{{ status.transaction_id || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="status.error_message" label="失败原因" :span="2">
            {{ status.error_message }}
          </el-descriptions-item>
        </el-descriptions>

        <h4 class="recv-title">分账各方</h4>
        <el-table :data="status.receivers || []" border stripe>
          <el-table-column label="类型" width="90">
            <template #default="{ row }">{{ receiverTypeText(row.receiver_type) }}</template>
          </el-table-column>
          <el-table-column prop="receiver_name" label="名称" min-width="110" />
          <el-table-column prop="account" label="账号" min-width="150" show-overflow-tooltip />
          <el-table-column label="金额" width="100">
            <template #default="{ row }">￥{{ row.amount?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="row.result_status === 'SUCCESS' ? 'success' : row.result_status === 'CLOSED' ? 'danger' : 'info'">
                {{ resultText(row.result_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="fail_reason" label="失败原因" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">{{ row.fail_reason || '-' }}</template>
          </el-table-column>
        </el-table>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.title {
  font-weight: 600;
}
.filters {
  display: flex;
  gap: 8px;
  align-items: center;
}
.desc {
  margin-bottom: 8px;
}
.recv-title {
  margin: 16px 0 8px;
}
</style>