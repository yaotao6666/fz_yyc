<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getOrders, dispatchOrder, getDispatchableStaffList, renewOrder } from '@/api/sp'
import type { SpOrder } from '@/types/sp'
import { SpOrderStatusText, OrderTypeText, BizStatusText } from '@/types/sp'
import { formatAmount, formatDateTime } from '@/utils/format'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const orders = ref<SpOrder[]>([])

const renewingId = ref<number | null>(null)
// 派单
const dispatchingId = ref<number | null>(null)
const dispatchDialogVisible = ref(false)
const staffLoading = ref(false)
const staffOptions = ref<{ id: number; name: string; phone?: string }[]>([])
const selectedStaffId = ref<number | null>(null)

async function openDispatchDialog(row: SpOrder) {
  dispatchTarget.value = row
  selectedStaffId.value = row.assigned_staff_id || null
  dispatchDialogVisible.value = true
  staffLoading.value = true
  try {
    staffOptions.value = await getDispatchableStaffList()
  } catch (error: any) {
    ElMessage.error(error?.message || '获取服务人员失败')
  } finally {
    staffLoading.value = false
  }
}

async function confirmDispatch() {
  const target = dispatchTarget.value
  if (!target || !target.id || !selectedStaffId.value) {
    ElMessage.warning('请选择服务人员')
    return
  }
  dispatchingId.value = target.id
  try {
    await dispatchOrder(target.id, selectedStaffId.value)
    ElMessage.success('派单成功')
    dispatchDialogVisible.value = false
    loadOrders()
  } catch (error: any) {
    ElMessage.error(error?.message || '派单失败')
  } finally {
    dispatchingId.value = null
  }
}

function canDispatchOrder(row: SpOrder) {
  return row.status === 2 && Number(row.order_type) !== 1 && Number(row.order_type) !== 5
}

// 派单目标订单（用于弹窗确认）
const dispatchTarget = ref<SpOrder | null>(null)
async function handleRenewFromList(row: SpOrder) {
  if (!row.id) return
  try {
    await ElMessageBox.confirm(`确认对订单「${row.order_no}」发起续租？`, '续租', { type: 'warning' })
  } catch {
    return
  }
  renewingId.value = row.id
  try {
    const res: any = await renewOrder(row.id)
    ElMessage.success(`续租单已生成，新订单号：${res?.order?.order_no}`)
    loadOrders()
  } catch (error: any) {
    ElMessage.error(error?.message || '续租失败')
  } finally {
    renewingId.value = null
  }
}

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

type SelectOptionValue = number | ''

const filters = reactive({
  category: 0, // 0=全部 1=实物订单 2=服务订单
  status: '' as SelectOptionValue,
  order_type: '' as SelectOptionValue,
  biz_status: '' as SelectOptionValue,
  keyword: '',
  dateRange: [] as string[]
})

const categoryTabs = [
  { label: '全部订单', value: 0 },
  { label: '实物订单', value: 1 },
  { label: '服务订单', value: 2 }
]

function handleCategoryChange(category: number) {
  filters.category = category
  pagination.page = 1
  loadOrders()
}

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待支付', value: 1 },
  { label: '已支付', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
  { label: '退款中', value: 5 },
  { label: '已退款', value: 6 }
]

const orderTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '零售', value: 1 },
  { label: '租赁', value: 2 },
  { label: '康养上门', value: 3 },
  { label: '陪诊服务', value: 4 }
]

const bizStatusOptions = [
  { label: '全部工单', value: '' },
  { label: '待接单', value: 1 },
  { label: '待出发', value: 2 },
  { label: '服务中', value: 3 },
  { label: '已完成', value: 5 },
  { label: '已取消', value: 6 }
]

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
    const response = await getOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      category: filters.category === 0 ? undefined : filters.category,
      status: filters.status === '' ? undefined : filters.status,
      order_type: filters.order_type === '' ? undefined : filters.order_type,
      biz_status: filters.biz_status === '' ? undefined : filters.biz_status,
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
  filters.category = 0
  filters.status = ''
  filters.order_type = ''
  filters.biz_status = ''
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

onMounted(loadOrders)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">订单管理</h1>
        <p class="page-subtitle">按订单状态、类型与工单状态筛选，并查看订单记录。</p>
      </div>
    </div>

    <el-card class="page-card" shadow="never">
      <div class="category-tabs">
        <div
          v-for="tab in categoryTabs"
          :key="tab.value"
          class="category-tab-item"
          :class="{ active: filters.category === tab.value }"
          @click="handleCategoryChange(tab.value)"
        >
          {{ tab.label }}
        </div>
      </div>

      <el-form class="toolbar-form" inline>
        <el-select v-model="filters.status" clearable placeholder="订单状态" style="width: 130px;">
          <el-option
            v-for="item in statusOptions"
            :key="String(item.value ?? 'all')"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select v-model="filters.order_type" clearable placeholder="订单类型" style="width: 140px;">
          <el-option
            v-for="item in orderTypeOptions"
            :key="String(item.value ?? 'all')"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select v-model="filters.biz_status" clearable placeholder="工单状态" style="width: 140px;">
          <el-option
            v-for="item in bizStatusOptions"
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
        <el-table-column label="订单类型" width="110">
          <template #default="scope">
            <el-tag v-if="scope.row.order_type" size="small">{{ OrderTypeText[scope.row.order_type] || '-' }}</el-tag>
            <span v-else>-</span>
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
        <el-table-column label="服务对象" width="110">
          <template #default="scope">
            <el-tag v-if="scope.row.record_name" type="success" size="small">{{ scope.row.record_name }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="订单金额" width="120">
          <template #default="scope">
            ¥{{ formatAmount(scope.row.pay_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="订单状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ SpOrderStatusText[scope.row.status] || '未知状态' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="工单状态" width="110">
          <template #default="scope">
            <el-tag v-if="scope.row.biz_status && scope.row.biz_status > 0" type="warning" size="small">
              {{ BizStatusText[scope.row.biz_status] || '-' }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="租赁到期" width="170">
          <template #default="scope">
            <span v-if="scope.row.rental_end_at">{{ formatDateTime(scope.row.rental_end_at) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="服务人员" width="130">
          <template #default="scope">
            <span v-if="scope.row.assigned_staff?.name">{{ scope.row.assigned_staff.name }}</span>
            <span v-else-if="scope.row.assigned_staff_id">服务人员#{{ scope.row.assigned_staff_id }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="下单时间" min-width="170">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="goDetail(scope.row.id)">查看详情</el-button>
            <el-button v-if="canDispatchOrder(scope.row)" link type="primary" v-permission="'order:dispatch'" :loading="dispatchingId === scope.row.id" @click="openDispatchDialog(scope.row)">派单</el-button>
            <el-button v-if="Number(scope.row.order_type) === 2 && scope.row.status !== 1" link type="warning" v-permission="'order:renew'" :loading="renewingId === scope.row.id" @click="handleRenewFromList(scope.row)">续租</el-button>
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

    <el-dialog v-model="dispatchDialogVisible" title="派单给服务人员" width="440px">
      <p v-if="dispatchTarget" class="dispatch-tip">订单号：{{ dispatchTarget.order_no }}，请选择已审核的服务人员。</p>
      <el-select v-model="selectedStaffId" v-loading="staffLoading" placeholder="选择已审核服务人员" style="width: 100%;" filterable>
        <el-option v-for="s in staffOptions" :key="s.id" :label="`${s.name} (${s.phone || '无手机号'})`" :value="s.id" />
      </el-select>
      <template #footer>
        <el-button @click="dispatchDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dispatchingId !== null" @click="confirmDispatch">确认派单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.category-tabs {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.category-tab-item {
  padding: 8px 20px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #4b5563;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.category-tab-item.active {
  background: #409eff;
  color: #ffffff;
  font-weight: 600;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.dispatch-tip {
  margin: 0 0 12px;
  font-size: 13px;
  color: #6b7280;
}
</style>
