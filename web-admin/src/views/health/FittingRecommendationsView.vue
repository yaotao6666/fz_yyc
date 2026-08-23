<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getFittingRecommendations,
  getFittingRecommendation,
  getMerchantProducts,
  updateFittingRecommendation,
  deleteFittingRecommendation
} from '@/api/sp'
import type { FittingRecommendation, MerchantProduct } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<FittingRecommendation[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  status: '' as number | ''
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 0 },
  { label: '已确认', value: 1 },
  { label: '已下单', value: 2 }
]

// 状态文案/标签映射（0草稿 1已确认 2已下单）
const statusTextMap: Record<number, string> = { 0: '草稿', 1: '已确认', 2: '已下单' }
const statusTypeMap: Record<number, string> = { 0: 'info', 1: 'success', 2: 'primary' }

function getUserName(row: FittingRecommendation) {
  return row.user?.nickname || row.user?.phone || '匿名用户'
}
function getUserPhone(row: FittingRecommendation) {
  return row.user?.phone || '-'
}
function statusText(s?: number) {
  return statusTextMap[Number(s ?? 0)] || '未知'
}
function statusType(s?: number) {
  return statusTypeMap[Number(s ?? 0)] || 'info'
}
function saleTypeText(t?: number) {
  return Number(t) === 2 ? '租赁' : '一口价'
}
function assessmentText(row: FittingRecommendation) {
  const a = row.assessment
  if (!a) return '-'
  const name = a.form_name || `评估#${a.id}`
  return a.level ? `${name}（${a.level}）` : name
}

async function loadData() {
  loading.value = true
  try {
    const res = await getFittingRecommendations({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      status: filters.status === '' ? undefined : filters.status
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
  filters.status = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(p: number) {
  pagination.page = p
  loadData()
}

/* ----- 详情抽屉 ----- */
const drawerVisible = ref(false)
const detailLoading = ref(false)
const current = ref<FittingRecommendation | null>(null)

async function openDetail(row: FittingRecommendation) {
  current.value = row
  drawerVisible.value = true
  detailLoading.value = true
  try {
    const detail = await getFittingRecommendation(row.id)
    current.value = detail
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    detailLoading.value = false
  }
}

/* ----- 编辑弹窗 ----- */
const dialogVisible = ref(false)
const editingId = ref(0)
const saving = ref(false)

// 编辑表单中的推荐商品（含推荐理由）
interface EditProductItem {
  product_id: number
  name: string
  sale_type?: number
  reason: string
}

const editForm = reactive({
  fitting_result: '',
  status: 0,
  order_id: null as number | null,
  products: [] as EditProductItem[]
})

const selectedProductIds = ref<number[]>([])
const productOptions = ref<MerchantProduct[]>([])

// 关键字搜索商品（仅上架商品）
async function searchProducts(keyword?: string) {
  try {
    const res = await getMerchantProducts({ keyword: keyword?.trim() || undefined, status: 1, page: 1, page_size: 50 })
    productOptions.value = res.list || []
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

// 商品多选变化时，同步编辑表单的推荐商品明细
function onProductSelectChange(ids: number[]) {
  const existing = new Map(editForm.products.map((p) => [p.product_id, p]))
  const next: EditProductItem[] = []
  for (const id of ids) {
    if (existing.has(id)) {
      next.push(existing.get(id)!)
    } else {
      const product = productOptions.value.find((p) => p.id === id)
      next.push({ product_id: id, name: product?.name || `商品#${id}`, sale_type: product?.sale_type, reason: '' })
    }
  }
  editForm.products = next
}

function openEdit(row?: FittingRecommendation) {
  const target = row || current.value
  if (!target) return
  editingId.value = target.id
  Object.assign(editForm, {
    fitting_result: target.fitting_result || '',
    status: target.status ?? 0,
    order_id: target.order_id ?? null,
    products: (target.recommended_products || []).map((p) => ({
      product_id: p.product_id,
      name: p.name,
      sale_type: p.sale_type,
      reason: p.reason || ''
    }))
  })
  selectedProductIds.value = editForm.products.map((p) => p.product_id)
  dialogVisible.value = true
  searchProducts()
}

async function handleSave() {
  if (!editingId.value) return
  if (!editForm.products.length) {
    ElMessage.warning('请至少选择一个推荐商品')
    return
  }
  saving.value = true
  try {
    await updateFittingRecommendation(editingId.value, {
      fitting_result: editForm.fitting_result.trim() || undefined,
      status: editForm.status,
      order_id: editForm.order_id,
      recommended_products: editForm.products.map((p) => ({ product_id: p.product_id, name: p.name, reason: p.reason.trim() || undefined }))
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

/* ----- 删除 ----- */
async function handleDelete(row: FittingRecommendation) {
  try {
    await ElMessageBox.confirm(`确认删除适配建议 #${row.id}？此操作不可恢复。`, '删除', { type: 'warning' })
    await deleteFittingRecommendation(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消或失败
  }
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
        <el-select v-model="filters.status" placeholder="状态" style="width: 130px" clearable @change="handleSearch">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
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
      <el-table-column label="症状需求" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.symptom_desc || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="适配结论" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.fitting_result || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="推荐商品数" width="110">
        <template #default="{ row }">
          <el-tag size="small" type="primary">{{ (row.recommended_products || []).length }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="关联评估" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          {{ assessmentText(row) }}
        </template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          <el-button link type="primary" v-permission="'fitting:update'" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" v-permission="'fitting:update'" @click="handleDelete(row)">删除</el-button>
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
        @size-change="(s: number) => { pagination.page_size = s; pagination.page = 1; loadData() }"
      />
    </div>

    <!-- 适配建议详情抽屉 -->
    <el-drawer v-model="drawerVisible" :title="`适配建议 #${current?.id ?? ''}`" size="640px" :close-on-click-modal="false">
      <div v-loading="detailLoading">
        <template v-if="current">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="用户">{{ getUserName(current) }}</el-descriptions-item>
            <el-descriptions-item label="手机">{{ getUserPhone(current) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag size="small" :type="statusType(current.status)">{{ statusText(current.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="关联订单">{{ current.order_id ? `#${current.order_id}` : '-' }}</el-descriptions-item>
            <el-descriptions-item label="关联评估" :span="2">{{ assessmentText(current) }}</el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(current.created_at) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="current.symptom_desc" class="detail-section">
            <div class="section-title">症状需求</div>
            <div class="section-content">{{ current.symptom_desc }}</div>
          </div>

          <div v-if="current.fitting_result" class="detail-section">
            <div class="section-title">适配结论</div>
            <div class="section-content">{{ current.fitting_result }}</div>
          </div>

          <div class="detail-section">
            <div class="section-title">推荐商品（{{ (current.recommended_products || []).length }}）</div>
            <el-table :data="current.recommended_products || []" size="small" stripe>
              <el-table-column prop="name" label="名称" min-width="160" show-overflow-tooltip />
              <el-table-column label="销售类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small" :type="Number(row.sale_type) === 2 ? 'warning' : 'success'">{{ saleTypeText(row.sale_type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="推荐理由" min-width="160" show-overflow-tooltip>
                <template #default="{ row }">{{ row.reason || '-' }}</template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </div>

      <template #footer>
        <template v-if="current">
          <el-button type="primary" v-permission="'fitting:update'" @click="openEdit()">编辑</el-button>
          <el-button @click="drawerVisible = false">关闭</el-button>
        </template>
      </template>
    </el-drawer>

    <!-- 编辑适配建议弹窗 -->
    <el-dialog v-model="dialogVisible" title="编辑适配建议" width="680px" destroy-on-close :close-on-click-modal="false">
      <el-form label-width="90px">
        <el-form-item label="适配结论">
          <el-input v-model="editForm.fitting_result" type="textarea" :rows="3" placeholder="适配结论，如：建议使用四脚拐杖辅助行走" />
        </el-form-item>
        <el-form-item label="推荐商品" required>
          <el-select
            v-model="selectedProductIds"
            multiple
            filterable
            remote
            :remote-method="searchProducts"
            placeholder="输入关键词搜索商品（仅上架商品）"
            style="width: 100%"
            @change="onProductSelectChange"
          >
            <el-option v-for="p in productOptions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <div v-if="editForm.products.length" class="product-list">
            <div v-for="p in editForm.products" :key="p.product_id" class="product-row">
              <span class="product-name">{{ p.name }}</span>
              <el-tag size="small" :type="Number(p.sale_type) === 2 ? 'warning' : 'success'">{{ saleTypeText(p.sale_type) }}</el-tag>
              <el-input v-model="p.reason" size="small" placeholder="推荐理由（可空）" style="flex: 1" />
            </div>
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 200px">
            <el-option v-for="o in statusOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联订单">
          <el-input-number v-model="editForm.order_id" :min="1" :controls="false" clearable placeholder="订单号（可空）" style="width: 200px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 20px; }
.sub-text { font-size: 12px; color: #909399; }
.detail-section { margin-top: 18px; }
.section-title { font-weight: 600; margin-bottom: 8px; color: #303133; }
.section-content { color: #606266; line-height: 1.6; }
.product-list { margin-top: 8px; }
.product-row { display: flex; align-items: center; gap: 8px; margin-top: 6px; }
.product-name { min-width: 140px; color: #303133; }
</style>
