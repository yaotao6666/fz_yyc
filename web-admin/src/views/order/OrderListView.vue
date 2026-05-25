<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMerchantList, getSpOrders } from '@/api/sp'
import type { MerchantListItem, SpOrder } from '@/types/sp'
import { SpDeliveryTypeText, SpOrderStatusText } from '@/types/sp'
import { formatAmount, formatDateTime } from '@/utils/format'

type SelectOptionValue = number | ''
type MerchantSelectOption = { label: string; value: SelectOptionValue }

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const orders = ref<SpOrder[]>([])
const merchantOptions = ref<MerchantSelectOption[]>([{ label: '全部商家', value: '' }])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  merchant_id: route.query.merchant_id ? Number(route.query.merchant_id) : '' as SelectOptionValue,
  status: '' as SelectOptionValue,
  keyword: typeof route.query.keyword === 'string' ? route.query.keyword : '',
  dateRange: [] as string[]
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待支付', value: 1 },
  { label: '已支付', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
  { label: '退款中', value: 5 },
  { label: '已退款', value: 6 }
]

async function loadMerchantOptions() {
  const response = await getMerchantList({ page: 1, page_size: 100 })
  merchantOptions.value = [
    { label: '全部商家', value: '' },
    ...response.list.map((item: MerchantListItem): MerchantSelectOption => ({
      label: item.name,
      value: item.id
    }))
  ]
}

function validateDateRange() {
  if (filters.dateRange.length !== 2) {
    return true
  }
  if (filters.dateRange[0] > filters.dateRange[1]) {
    ElMessage.warning('开始日期不能晚于结束日期')
    return false
  }
  return true
}

async function loadOrders() {
  if (!validateDateRange()) {
    return
  }

  loading.value = true
  try {
    const response = await getSpOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      merchant_id: filters.merchant_id === '' ? undefined : filters.merchant_id,
      status: filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword.trim() || undefined,
      start_date: filters.dateRange[0],
      end_date: filters.dateRange[1]
    })
    orders.value = response.list
    pagination.total = response.pagination.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadOrders()
}

function handleReset() {
  filters.merchant_id = ''
  filters.status = ''
  filters.keyword = ''
  filters.dateRange = []
  pagination.page = 1
  loadOrders()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadOrders()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadOrders()
}

function getStatusType(status: number) {
  return {
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'warning',
    5: 'danger',
    6: 'danger'
  }[status] || 'info'
}

function getUserLabel(order: SpOrder) {
  return order.user?.nickname || order.user?.phone || '匿名用户'
}

function getGoodsSummary(order: SpOrder) {
  if (!order.items?.length) {
    return '-'
  }
  const firstItem = order.items[0]
  return order.items.length > 1
    ? `${firstItem.product_name} 等${order.items.length}件`
    : firstItem.product_name
}

function goDetail(orderId: number) {
  router.push(`/orders/${orderId}`)
}

onMounted(async () => {
  await loadMerchantOptions()
  await loadOrders()
})
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">订单管理</h1>
        <p class="page-subtitle">支持按商家、状态、日期和订单号查看服务商名下商家的订单记录。</p>
      </div>
    </div>

    <el-card class="page-card" shadow="never">
      <el-form class="toolbar-form" inline>
        <el-select v-model="filters.merchant_id" clearable placeholder="选择商家">
          <el-option
            v-for="item in merchantOptions"
            :key="String(item.value ?? 'all')"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="选择订单状态">
          <el-option
            v-for="item in statusOptions"
            :key="String(item.value ?? 'all')"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-date-picker
          v-model="filters.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
        />
        <el-input v-model="filters.keyword" clearable placeholder="输入订单号" style="width: 220px;" />
        <el-button type="primary" @click="handleSearch">查询订单</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form>

      <el-table :data="orders" v-loading="loading" style="margin-top: 20px; width: 100%;">
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column label="商家" min-width="150">
          <template #default="scope">
            {{ scope.row.merchant?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="140">
          <template #default="scope">
            {{ getUserLabel(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column label="商品" min-width="180" show-overflow-tooltip>
          <template #default="scope">
            {{ getGoodsSummary(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column label="配送方式" width="100">
          <template #default="scope">
            {{ SpDeliveryTypeText[scope.row.delivery_type || 0] || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="订单金额" width="120">
          <template #default="scope">
            ¥{{ formatAmount(scope.row.pay_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ SpOrderStatusText[scope.row.status] || '未知状态' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="下单时间" min-width="170">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="goDetail(scope.row.id)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
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

<style scoped>
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
