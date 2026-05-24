<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getMerchantList, getSpProfitSharingRecords } from '@/api/sp'
import type { ProfitSharingRecord, MerchantListItem } from '@/types/sp'
import { formatAmount, formatDateTime, formatPercent } from '@/utils/format'

type SelectOptionValue = number | ''
type MerchantSelectOption = { label: string; value: SelectOptionValue }

const route = useRoute()
const loading = ref(false)
const records = ref<ProfitSharingRecord[]>([])
const merchantOptions = ref<MerchantSelectOption[]>([{ label: '全部商家', value: '' }])
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})
const filters = reactive({
  merchant_id: route.query.merchant_id ? Number(route.query.merchant_id) : '' as SelectOptionValue,
  status: '' as SelectOptionValue,
  dateRange: [] as string[]
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待处理', value: 0 },
  { label: '已成功', value: 1 },
  { label: '已失败', value: 2 },
  { label: '已跳过', value: 3 }
]

async function loadMerchantOptions() {
  const response = await getMerchantList({ page: 1, page_size: 100 })
  const options: MerchantSelectOption[] = [
    { label: '全部商家', value: '' },
    ...response.list.map((item: MerchantListItem): MerchantSelectOption => ({
      label: item.name,
      value: item.id
    }))
  ]
  merchantOptions.value = options
}

async function loadRecords() {
  loading.value = true
  try {
    const response = await getSpProfitSharingRecords({
      page: pagination.page,
      page_size: pagination.page_size,
      merchant_id: filters.merchant_id === '' ? undefined : filters.merchant_id,
      status: filters.status === '' ? undefined : filters.status,
      start_date: filters.dateRange[0],
      end_date: filters.dateRange[1],
    })
    records.value = response.list
    pagination.total = response.pagination.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadRecords()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadRecords()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadRecords()
}

function getStatusType(status: number) {
  return {
    0: 'info',
    1: 'success',
    2: 'danger',
    3: 'warning'
  }[status] || 'info'
}

function getStatusText(status: number) {
  return {
    0: '待处理',
    1: '已成功',
    2: '已失败',
    3: '已跳过'
  }[status] || '未知状态'
}

onMounted(async () => {
  await loadMerchantOptions()
  await loadRecords()
})
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">分账历史</h1>
        <p class="page-subtitle">按商家、状态和日期筛选服务商抽佣记录。</p>
      </div>
    </div>

    <el-card class="page-card" shadow="never">
      <el-form class="toolbar-form" inline>
        <el-select v-model="filters.merchant_id" clearable placeholder="选择商家">
          <el-option v-for="item in merchantOptions" :key="String(item.value ?? 'all')" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="选择状态">
          <el-option v-for="item in statusOptions" :key="String(item.value ?? 'all')" :label="item.label" :value="item.value" />
        </el-select>
        <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" />
        <el-button type="primary" @click="handleSearch">查询记录</el-button>
      </el-form>

      <el-table :data="records" v-loading="loading" style="margin-top: 20px; width: 100%;">
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column label="分账时间" min-width="160">
          <template #default="scope">
            {{ formatDateTime(scope.row.profit_sharing_date || scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="支付金额" width="120">
          <template #default="scope">¥{{ formatAmount(scope.row.pay_amount) }}</template>
        </el-table-column>
        <el-table-column label="抽佣比例" width="120">
          <template #default="scope">{{ formatPercent(scope.row.profit_sharing_ratio) }}</template>
        </el-table-column>
        <el-table-column label="抽佣金额" width="120">
          <template #default="scope">¥{{ formatAmount(scope.row.profit_sharing_amount) }}</template>
        </el-table-column>
        <el-table-column label="商家实收" width="120">
          <template #default="scope">¥{{ formatAmount(scope.row.merchant_received_amount) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error_message" label="失败原因" min-width="200" show-overflow-tooltip />
      </el-table>

      <div style="display: flex; justify-content: flex-end; margin-top: 20px;">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.page_size"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>
