<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { getMonitoring } from '@/api/sp'
import type { HealthMonitoring } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<HealthMonitoring[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  record_type: '' as number | ''
})

const recordTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '血压', value: 1 },
  { label: '血糖', value: 2 },
  { label: '心率', value: 3 },
  { label: '血氧', value: 4 },
  { label: '体重', value: 5 }
]

// 类型文案与标签映射
const recordTypeTextMap: Record<number, string> = { 1: '血压', 2: '血糖', 3: '心率', 4: '血氧', 5: '体重' }
const recordTypeTagMap: Record<number, string> = { 1: 'danger', 2: 'warning', 3: 'primary', 4: 'info', 5: 'success' }

function getUserName(row: HealthMonitoring) {
  return row.user?.nickname || row.user?.phone || '匿名用户'
}
function getUserPhone(row: HealthMonitoring) {
  return row.user?.phone || '-'
}
function recordTypeText(t?: number) {
  return recordTypeTextMap[Number(t)] || '-'
}
function recordTypeTag(t?: number) {
  return recordTypeTagMap[Number(t)] || 'info'
}
function recordValueText(row: HealthMonitoring) {
  const value = row.value ?? '-'
  return row.unit ? `${value} ${row.unit}` : String(value)
}
// 录入人：recorded_by=0 为用户本人，否则展示服务人员姓名
function recorderText(row: HealthMonitoring) {
  if (Number(row.recorded_by ?? 0) === 0) return '用户本人'
  const staff = (row as HealthMonitoring & { staff?: { name?: string } }).staff
  return staff?.name || `服务人员#${row.recorded_by}`
}

async function loadData() {
  loading.value = true
  try {
    const res = await getMonitoring({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      record_type: filters.record_type === '' ? undefined : filters.record_type
    })
    list.value = res.list || []
    pagination.total = res.total || 0
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  filters.keyword = ''
  filters.record_type = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(p: number) {
  pagination.page = p
  loadData()
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="用户昵称/手机号"
          style="width: 220px"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="filters.record_type" placeholder="体征类型" style="width: 140px" clearable @change="handleSearch">
          <el-option v-for="o in recordTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="用户" min-width="140">
        <template #default="{ row }">
          <div>{{ getUserName(row) }}</div>
          <div class="sub-text">{{ getUserPhone(row) }}</div>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="recordTypeTag(row.record_type)">{{ recordTypeText(row.record_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="数值" min-width="110">
        <template #default="{ row }">{{ recordValueText(row) }}</template>
      </el-table-column>
      <el-table-column label="测量时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.recorded_at) }}</template>
      </el-table-column>
      <el-table-column label="录入人" min-width="120">
        <template #default="{ row }">{{ recorderText(row) }}</template>
      </el-table-column>
      <el-table-column label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
      <el-table-column label="时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
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
        @size-change="(s: number) => { pagination.page_size = s; pagination.page = 1; loadData() }"
      />
    </div>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 20px; }
.sub-text { font-size: 12px; color: #909399; }
</style>
