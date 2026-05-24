<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getMerchantList } from '@/api/sp'
import type { MerchantListItem } from '@/types/sp'
import { formatAmount, formatDate, formatPercent, getMerchantStatusText, getPaymentConfigText } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const merchants = ref<MerchantListItem[]>([])
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})
const filters = reactive({
  keyword: '',
  status: ''
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '营业中', value: '1' },
  { label: '休息中', value: '2' },
  { label: '已关闭', value: '3' }
]

async function loadMerchants() {
  loading.value = true
  try {
    const response = await getMerchantList({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      status: filters.status || undefined
    })
    merchants.value = response.list
    pagination.total = response.pagination.total
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadMerchants()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadMerchants()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadMerchants()
}

onMounted(loadMerchants)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">商家列表</h1>
        <p class="page-subtitle">统一查看商家资料、支付配置、分账状态与经营概览。</p>
      </div>
      <el-button type="primary" @click="router.push('/merchants/new')">新增商家</el-button>
    </div>

    <el-card class="page-card" shadow="never">
      <el-form class="toolbar-form" inline @submit.prevent>
        <el-input v-model="filters.keyword" clearable placeholder="搜索商家名称" @keyup.enter="handleSearch" />
        <el-select v-model="filters.status" placeholder="选择状态" clearable>
          <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">查询</el-button>
      </el-form>

      <el-table :data="merchants" v-loading="loading" style="margin-top: 20px; width: 100%;">
        <el-table-column prop="name" label="商家名称" min-width="180" />
        <el-table-column label="联系人 / 电话" min-width="180">
          <template #default="scope">
            {{ scope.row.contact_name || '未设置' }} / {{ scope.row.contact_phone || '未设置' }}
          </template>
        </el-table-column>
        <el-table-column prop="business_category" label="经营分类" min-width="140" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : scope.row.status === 2 ? 'warning' : 'info'">
              {{ getMerchantStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付配置" width="120">
          <template #default="scope">
            <el-tag :type="Number(scope.row.payment_config_status || 0) === 1 ? 'success' : 'warning'">
              {{ getPaymentConfigText(scope.row.payment_config_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分账配置" min-width="150">
          <template #default="scope">
            {{ scope.row.profit_sharing_enabled ? `已开启 ${formatPercent(scope.row.profit_sharing_ratio)}` : '未开启' }}
          </template>
        </el-table-column>
        <el-table-column label="经营数据" min-width="180">
          <template #default="scope">
            用户 {{ scope.row.total_users }} / 订单 {{ scope.row.total_orders }} / 金额 ¥{{ formatAmount(scope.row.total_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-space>
              <el-button link type="primary" @click="router.push(`/merchants/${scope.row.id}`)">详情</el-button>
              <el-button link type="primary" @click="router.push(`/merchants/${scope.row.id}/edit`)">编辑</el-button>
            </el-space>
          </template>
        </el-table-column>
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
