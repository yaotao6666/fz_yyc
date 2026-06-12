<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  batchUpdateMerchantProductStatus,
  deleteMerchantProduct,
  getMerchantCategories,
  getMerchantProducts,
  merchantProductOffSale,
  merchantProductOnSale,
  updateMerchantProductStock
} from '@/api/sp'
import type { MerchantCategory, MerchantProduct } from '@/types/sp'
import { formatAmount, formatDateTime, getProductStatusText } from '@/utils/format'
import MerchantProductEditorDialog from './MerchantProductEditorDialog.vue'

const props = defineProps<{
  merchantId: number
}>()

const loading = ref(false)
const categories = ref<MerchantCategory[]>([])
const products = ref<MerchantProduct[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const editorVisible = ref(false)
const editingProductId = ref<number | null>(null)
const actionLoadingId = ref<number | null>(null)

const filters = reactive({
  category_id: undefined as number | undefined,
  status: '' as '' | '1' | '2',
  keyword: '',
  page: 1,
  page_size: 10
})

async function loadCategories() {
  if (!props.merchantId) return
  categories.value = await getMerchantCategories(props.merchantId)
}

async function loadProducts() {
  if (!props.merchantId) return
  loading.value = true
  try {
    const result = await getMerchantProducts(props.merchantId, {
      category_id: filters.category_id,
      status: filters.status,
      keyword: filters.keyword.trim() || undefined,
      page: filters.page,
      page_size: filters.page_size
    })
    products.value = Array.isArray(result.list) ? result.list : []
    total.value = Number(result.pagination?.total || 0)
  } finally {
    loading.value = false
  }
}

async function loadData() {
  if (!props.merchantId) return
  await Promise.all([loadCategories(), loadProducts()])
}

function handleSearch() {
  filters.page = 1
  void loadProducts()
}

function handleReset() {
  filters.category_id = undefined
  filters.status = ''
  filters.keyword = ''
  filters.page = 1
  void loadProducts()
}

function handleSelectionChange(selection: MerchantProduct[]) {
  selectedIds.value = selection.map((item) => item.id)
}

function openCreate() {
  editingProductId.value = null
  editorVisible.value = true
}

function openEdit(product: MerchantProduct) {
  editingProductId.value = product.id
  editorVisible.value = true
}

function handleEditorSuccess() {
  editorVisible.value = false
  editingProductId.value = null
  void loadData()
}

async function handleToggleStatus(product: MerchantProduct) {
  if (!props.merchantId) return
  actionLoadingId.value = product.id
  try {
    if (Number(product.status || 0) === 1) {
      await merchantProductOffSale(props.merchantId, product.id)
      ElMessage.success('商品已下架')
    } else {
      await merchantProductOnSale(props.merchantId, product.id)
      ElMessage.success('商品已上架')
    }
    await loadProducts()
  } finally {
    actionLoadingId.value = null
  }
}

async function handleDelete(product: MerchantProduct) {
  if (!props.merchantId) return
  try {
    await ElMessageBox.confirm(`确认删除商品「${product.name}」？`, '删除商品', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  actionLoadingId.value = product.id
  try {
    await deleteMerchantProduct(props.merchantId, product.id)
    ElMessage.success('删除成功')
    await loadProducts()
  } finally {
    actionLoadingId.value = null
  }
}

async function handleUpdateStock(product: MerchantProduct) {
  if (!props.merchantId) return

  try {
    const result = await ElMessageBox.prompt('请输入新的库存数量', '修改库存', {
      inputValue: String(product.stock || 0),
      inputPattern: /^\d+$/,
      inputErrorMessage: '库存必须为非负整数',
      confirmButtonText: '保存',
      cancelButtonText: '取消'
    })

    const stock = Number(result.value || 0)
    actionLoadingId.value = product.id
    await updateMerchantProductStock(props.merchantId, product.id, stock)
    ElMessage.success('库存更新成功')
    await loadProducts()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('更新库存失败:', error)
    }
  } finally {
    actionLoadingId.value = null
  }
}

async function handleBatchUpdate(status: number) {
  if (!props.merchantId) return
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择商品')
    return
  }

  const actionText = status === 1 ? '上架' : '下架'
  try {
    await ElMessageBox.confirm(`确认批量${actionText}选中的 ${selectedIds.value.length} 个商品？`, `批量${actionText}`, {
      type: 'warning',
      confirmButtonText: '确认',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  loading.value = true
  try {
    await batchUpdateMerchantProductStatus(props.merchantId, selectedIds.value, status)
    ElMessage.success(`批量${actionText}成功`)
    selectedIds.value = []
    await loadProducts()
  } finally {
    loading.value = false
  }
}

watch(
  () => props.merchantId,
  (merchantId) => {
    if (merchantId) {
      void loadData()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="tab-block">
    <div class="toolbar">
      <el-form inline>
        <el-form-item label="分类">
          <el-select v-model="filters.category_id" clearable placeholder="全部分类" style="width: 160px;">
            <el-option
              v-for="category in categories"
              :key="category.id"
              :label="category.name"
              :value="category.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 140px;">
            <el-option label="上架" value="1" />
            <el-option label="下架" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            clearable
            placeholder="商品名称"
            style="width: 220px;"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
      </el-form>

      <el-space wrap>
        <el-button @click="handleReset">重置</el-button>
        <el-button type="primary" plain @click="handleSearch">查询</el-button>
        <el-button type="success" plain @click="handleBatchUpdate(1)">批量上架</el-button>
        <el-button type="warning" plain @click="handleBatchUpdate(2)">批量下架</el-button>
        <el-button type="primary" @click="openCreate">新增商品</el-button>
      </el-space>
    </div>

    <el-card class="page-card" shadow="never">
      <el-table
        :data="products"
        v-loading="loading"
        style="width: 100%;"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column label="图片" width="92">
          <template #default="{ row }">
            <img
              v-if="row.images?.[0]"
              :src="row.images[0]"
              alt="商品图片"
              class="product-image"
            />
            <div v-else class="empty-image">无图</div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="商品名称" min-width="180" />
        <el-table-column prop="category_name" label="分类" min-width="120">
          <template #default="{ row }">
            {{ row.category_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="售价" width="110">
          <template #default="{ row }">
            ¥{{ formatAmount(row.price) }}
          </template>
        </el-table-column>
        <el-table-column label="库存" width="100">
          <template #default="{ row }">
            {{ row.stock }}
          </template>
        </el-table-column>
        <el-table-column label="销量" width="100">
          <template #default="{ row }">
            {{ row.sales }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="Number(row.status || 0) === 1 ? 'success' : 'info'">
              {{ getProductStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="168">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-space wrap>
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button
                link
                type="primary"
                :loading="actionLoadingId === row.id"
                @click="handleToggleStatus(row)"
              >
                {{ Number(row.status || 0) === 1 ? '下架' : '上架' }}
              </el-button>
              <el-button link type="primary" @click="handleUpdateStock(row)">改库存</el-button>
              <el-button
                link
                type="danger"
                :loading="actionLoadingId === row.id"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="filters.page"
          v-model:page-size="filters.page_size"
          background
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[10, 20, 50]"
          :total="total"
          @current-change="loadProducts"
          @size-change="handleSearch"
        />
      </div>
    </el-card>

    <MerchantProductEditorDialog
      v-model="editorVisible"
      :merchant-id="merchantId"
      :categories="categories"
      :product-id="editingProductId"
      @success="handleEditorSuccess"
    />
  </div>
</template>

<style scoped>
.tab-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

.product-image,
.empty-image {
  width: 56px;
  height: 56px;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
}

.product-image {
  object-fit: cover;
  display: block;
}

.empty-image {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 12px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
