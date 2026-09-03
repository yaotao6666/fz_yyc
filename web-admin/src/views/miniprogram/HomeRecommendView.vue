<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, ref } from 'vue'
import {
  createHomeRecommend,
  deleteHomeRecommend,
  getHomeRecommends,
  getMerchantProducts,
  updateHomeRecommend,
  updateHomeRecommendStatus
} from '@/api/sp'
import type { HomeRecommend } from '@/types/sp'

const loading = ref(false)
const list = ref<HomeRecommend[]>([])

const targetTypeOptions = [
  { label: '实物商品（零售/租赁）', value: 1 },
  { label: '服务（康养套餐/陪诊）', value: 2 }
]
function targetTypeText(t: number) {
  return targetTypeOptions.find(o => o.value === t)?.label || '未知'
}
function statusText(s: number) {
  return s === 1 ? '启用' : '禁用'
}
function statusType(s: number) {
  return s === 1 ? 'success' : 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getHomeRecommends()
    list.value = res.list || []
  } catch (_e) {
    // 拦截器提示
  } finally {
    loading.value = false
  }
}

/* ------ 新增/编辑弹窗 ------ */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)
const saving = ref(false)
const productLoading = ref(false)
const productOptions = ref<{ id: number; name: string; price: number }[]>([])

const form = ref({
  product_id: 0 as number,
  target_type: 1,
  title: '',
  sort: 0,
  status: 1
})

// product_types 拉取池：1,2=实物 3,4=服务
async function loadProductOptions(targetType: number, selectedProductId = 0) {
  productLoading.value = true
  productOptions.value = []
  form.value.product_id = selectedProductId
  try {
    const res = await getMerchantProducts({
      page: 1,
      page_size: 200,
      status: 1,
      product_types: targetType === 1 ? '1,2' : '3,4'
    })
    productOptions.value = (res.list || []).map(p => ({
      id: p.id,
      name: p.name,
      price: p.price,
      product_type: p.product_type
    }))
  } catch (_e) {
    // 拦截器提示
  } finally {
    productLoading.value = false
  }
}

const selectedProductText = computed(() => {
  const p = productOptions.value.find(o => o.id === form.value.product_id)
  return p ? `${p.name}（¥${p.price ?? 0}）` : ''
})

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = null
  form.value = { product_id: 0, target_type: 1, title: '', sort: 0, status: 1 }
  dialogVisible.value = true
  loadProductOptions(1, 0)
}

function openEdit(row: HomeRecommend) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.value = {
    product_id: row.product_id,
    target_type: row.target_type,
    title: row.title || '',
    sort: row.sort ?? 0,
    status: row.status ?? 1
  }
  dialogVisible.value = true
  // 编辑时需保留已绑定的目标商品/服务，否则选项拉取后选择会被重置
  loadProductOptions(row.target_type, row.product_id)
}

function handleTargetTypeChange() {
  loadProductOptions(form.value.target_type, 0)
}

async function handleSave() {
  if (!form.value.product_id) {
    ElMessage.warning('请选择推荐的商品或服务')
    return
  }
  const payload = {
    product_id: form.value.product_id,
    target_type: form.value.target_type,
    title: form.value.title || undefined,
    sort: form.value.sort,
    status: form.value.status
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createHomeRecommend(payload)
      ElMessage.success('新增成功')
    } else if (editingId.value != null) {
      await updateHomeRecommend(editingId.value, payload)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 拦截器提示
  } finally {
    saving.value = false
  }
}

/* ------ 操作 ------ */
async function handleStatusChange(row: HomeRecommend) {
  const targetStatus = row.status === 1 ? 0 : 1
  const action = targetStatus === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确认${action}该推荐项？`, action, { type: 'warning' })
    await updateHomeRecommendStatus(row.id, targetStatus)
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (_e) {
    // 取消或失败
  }
}

async function handleDelete(row: HomeRecommend) {
  try {
    await ElMessageBox.confirm('确认删除该推荐项？删除后不可恢复。', '删除', { type: 'warning' })
    await deleteHomeRecommend(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">首页推荐</h1>
        <p class="page-subtitle">配置 C 端小程序首页展示的推荐商品/服务，生成「推荐」模块。</p>
      </div>
      <el-button type="primary" v-permission="'home-recommend:create'" @click="openCreate">+ 新增推荐</el-button>
    </div>

    <div class="page-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="封面" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.product_image"
              :src="row.product_image"
              fit="cover"
              style="width: 64px; height: 64px; border-radius: 8px"
              :preview-src-list="[row.product_image]"
            />
            <span v-else class="empty-tip">无图</span>
          </template>
        </el-table-column>
        <el-table-column label="推荐对象" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div>{{ row.title || row.product_name || '-' }}</div>
            <div class="sub-info" v-if="row.product_name && row.title && row.title !== row.product_name">
              {{ row.product_name }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="200">
          <template #default="{ row }">
            <el-tag :type="row.target_type === 2 ? 'success' : 'primary'" size="small">
              {{ targetTypeText(row.target_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="110">
          <template #default="{ row }">
            <span v-if="row.product_price !== undefined && row.product_price !== null">¥{{ row.product_price }}</span>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="80">
          <template #default="{ row }">{{ row.sort ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" v-permission="'home-recommend:update'" @click="openEdit(row)">编辑</el-button>
            <el-button
              size="small"
              :type="row.status === 1 ? 'warning' : 'primary'"
              v-permission="'home-recommend:status'"
              @click="handleStatusChange(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button size="small" type="danger" v-permission="'home-recommend:delete'" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!loading && list.length === 0" class="empty-state">
        <p>还未配置任何推荐项，点击右上角「新增推荐」开始配置。</p>
      </div>
    </div>

    <!-- 新增/编辑 弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增推荐' : '编辑推荐'" width="620px" destroy-on-close>
      <el-form :model="form" label-width="110px">
        <el-form-item label="推荐类型" required>
          <el-radio-group v-model="form.target_type" @change="handleTargetTypeChange">
            <el-radio v-for="o in targetTypeOptions" :key="o.value" :value="o.value">
              {{ o.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="目标商品/服务" required>
          <el-select
            v-model="form.product_id"
            :loading="productLoading"
            filterable
            placeholder="选择已上架的实物商品或服务"
            style="width: 100%"
          >
            <el-option v-for="p in productOptions" :key="p.id" :label="`${p.name}（¥${p.price ?? 0}）`" :value="p.id" />
          </el-select>
          <p v-if="selectedProductText" class="form-tip">已选择：{{ selectedProductText }}</p>
        </el-form-item>

        <el-form-item label="展示标题">
          <el-input v-model="form.title" placeholder="留空默认使用商品/服务名称（可选）" maxlength="50" show-word-limit />
        </el-form-item>

        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
          <span class="form-tip">值越小越靠前，留空默认为 0。</span>
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ dialogMode === 'create' ? '确认新增' : '保存修改' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-shell { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-end; }
.page-title { margin: 0 0 4px; font-size: 22px; }
.page-subtitle { margin: 0; font-size: 13px; color: #6b7280; }
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.empty-tip { color: #9ca3af; }
.sub-info { color: #6b7280; font-size: 12px; margin-top: 2px; }
.empty-state { padding: 60px 20px; text-align: center; color: #9ca3af; }
.empty-state p { margin: 0; }
.form-tip { margin: 4px 0 0; font-size: 12px; color: #6b7280; }
</style>