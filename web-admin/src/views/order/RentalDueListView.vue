<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRentalDueOrders, renewOrder, returnRentalOrder } from '@/api/sp'
import type { SpOrder } from '@/types/sp'
import { formatAmount, formatDateTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const list = ref<any[]>([])
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const filters = reactive({ due_range: '' as string, keyword: '' })

async function loadData() {
  loading.value = true
  try {
    const res = await getRentalDueOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      due_range: filters.due_range,
      keyword: filters.keyword.trim() || undefined
    })
    list.value = res.list || []
    pagination.total = res.total || 0
  } finally {
    loading.value = false
  }
}

function getUserLabel(o: SpOrder) {
  return o.user?.nickname || o.user?.phone || '匿名用户'
}

function handleSearch() { pagination.page = 1; loadData() }
function handleReset() { filters.due_range = ''; filters.keyword = ''; pagination.page = 1; loadData() }
function handlePageChange(p: number) { pagination.page = p; loadData() }
function goDetail(id: number) { router.push(`/orders/${id}`) }

async function handleRenew(row: any) {
  const o: SpOrder = row.order
  try {
    await ElMessageBox.confirm(`确认为订单「${o.order_no}」发起续租？将生成待支付续租单。`, '续租', { type: 'warning' })
    const res = await renewOrder(o.id)
    ElMessage.success(`续租单已生成，新订单号：${res.order?.order_no}`)
    loadData()
  } catch (e) {
    if (e !== 'cancel' && (e as any)?.message) ElMessage.error((e as any).message)
  }
}

async function handleReturn(row: any) {
  const o: SpOrder = row.order
  if (!o.id) return
  try {
    const { value } = await ElMessageBox.prompt('请输入扣除金额（损坏赔偿，0=全额退还）', '归还退押金', {
      inputValue: '0',
      inputPattern: /^\d+(\.\d{1,2})?$/,
      inputErrorMessage: '请输入有效金额'
    })
    await returnRentalOrder(o.id, { deduct_amount: Number(value || 0) })
    ElMessage.success('归还成功，押金已退还')
    loadData()
  } catch (e) {
    if (e !== 'cancel' && (e as any)?.message) ElMessage.error((e as any).message)
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">租赁到期提醒</h1>
        <p class="page-subtitle">跟踪租赁订单到期情况，及时归还或续租。</p>
      </div>
    </div>

    <el-card class="page-card" shadow="never">
      <el-form class="toolbar-form" inline>
        <el-select v-model="filters.due_range" clearable placeholder="到期范围" style="width: 150px;">
          <el-option label="全部" value="" />
          <el-option label="将到期(7天内)" value="soon" />
          <el-option label="已逾期" value="overdue" />
        </el-select>
        <el-input v-model="filters.keyword" clearable placeholder="输入订单号/联系人" style="width: 220px;" />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form>

      <el-table :data="list" v-loading="loading" style="margin-top: 20px; width: 100%;">
        <el-table-column prop="order.order_no" label="订单号" min-width="180" />
        <el-table-column label="用户" min-width="130">
          <template #default="scope">{{ getUserLabel(scope.row.order) }}</template>
        </el-table-column>
        <el-table-column label="商品" min-width="160" show-overflow-tooltip>
          <template #default="scope">{{ scope.row.order?.items?.[0]?.product_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="实付金额" width="110">
          <template #default="scope">¥{{ formatAmount(scope.row.order?.pay_amount) }}</template>
        </el-table-column>
        <el-table-column label="押金状态" width="110">
          <template #default="scope">
            <el-tag :type="Number(scope.row.order?.deposit_status) === 1 ? 'warning' : 'info'" size="small">
              {{ ['无押金','已收','已退','部分扣除'][Number(scope.row.order?.deposit_status || 0)] || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="到期时间" min-width="170">
          <template #default="scope">{{ formatDateTime(scope.row.rental_end_at) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="scope">
            <el-tag v-if="scope.row.is_overdue" type="danger" size="small">已逾期</el-tag>
            <el-tag v-else-if="scope.row.days_left <= 7" type="warning" size="small">将到期({{ scope.row.days_left }}天)</el-tag>
            <el-tag v-else size="small">剩{{ scope.row.days_left }}天</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="goDetail(scope.row.order.id)">详情</el-button>
            <el-button link type="success" v-if="hasPerm('orderrental:renew')" @click="handleRenew(scope.row)">续租</el-button>
            <el-button link type="warning" v-if="hasPerm('orderrental:return')" @click="handleReturn(scope.row)">归还</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          background layout="total, sizes, prev, pager, next"
          :current-page="pagination.page" :page-size="pagination.page_size"
          :page-sizes="[10, 20, 50]" :total="pagination.total"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script lang="ts">
import { useAuthStore } from '@/stores/auth'
export default {
  methods: {
    hasPerm(code: string) { return useAuthStore().hasPermission(code) }
  }
}
</script>

<style scoped>
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 20px; }
</style>