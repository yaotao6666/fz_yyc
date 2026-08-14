<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createMerchantCategory,
  deleteMerchantCategory,
  getMerchantCategories,
  sortMerchantCategories,
  updateMerchantCategory
} from '@/api/sp'
import type { MerchantCategory, MerchantCategoryPayload } from '@/types/sp'
import { formatDateTime, getCategoryStatusText } from '@/utils/format'

const loading = ref(false)
const savingSort = ref(false)
const dialogVisible = ref(false)
const dialogSubmitting = ref(false)
const editingCategory = ref<MerchantCategory | null>(null)
const categories = ref<MerchantCategory[]>([])

const form = reactive<MerchantCategoryPayload>({
  name: '',
  sort: 0,
  status: 1
})

function resetForm() {
  form.name = ''
  form.sort = 0
  form.status = 1
}

function openCreate() {
  editingCategory.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(category: MerchantCategory) {
  editingCategory.value = category
  form.name = category.name
  form.sort = category.sort
  form.status = category.status
  dialogVisible.value = true
}

async function loadData() {
  loading.value = true
  try {
    categories.value = await getMerchantCategories()
  } finally {
    loading.value = false
  }
}

function buildPayload(status?: number): MerchantCategoryPayload {
  return {
    name: form.name.trim(),
    sort: Number(form.sort || 0),
    status: Number(status ?? form.status ?? 1)
  }
}

async function submitForm() {
  if (dialogSubmitting.value) return

  const payload = buildPayload()
  if (!payload.name) {
    ElMessage.warning('请输入分类名称')
    return
  }

  dialogSubmitting.value = true
  try {
    if (editingCategory.value) {
      await updateMerchantCategory(editingCategory.value.id, payload)
      ElMessage.success('分类更新成功')
    } else {
      await createMerchantCategory(payload)
      ElMessage.success('分类创建成功')
    }
    dialogVisible.value = false
    await loadData()
  } finally {
    dialogSubmitting.value = false
  }
}

async function handleDelete(category: MerchantCategory) {
  try {
    await ElMessageBox.confirm(`确认删除分类「${category.name}」？`, '删除分类', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  await deleteMerchantCategory(category.id)
  ElMessage.success('删除成功')
  await loadData()
}

async function toggleStatus(category: MerchantCategory) {
  const nextStatus = Number(category.status || 0) === 1 ? 0 : 1
  await updateMerchantCategory(category.id, {
    name: category.name,
    sort: category.sort,
    status: nextStatus
  })
  ElMessage.success(nextStatus === 1 ? '已启用' : '已停用')
  await loadData()
}

async function saveSort() {
  if (savingSort.value) return

  savingSort.value = true
  try {
    await sortMerchantCategories(categories.value.map((item) => ({
      id: item.id,
      sort: Number(item.sort || 0)
    })))
    ElMessage.success('排序保存成功')
    await loadData()
  } finally {
    savingSort.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="tab-block">
    <div class="toolbar">
      <el-space wrap>
        <el-button type="primary" @click="openCreate">新增分类</el-button>
        <el-button plain :loading="savingSort" @click="saveSort">保存排序</el-button>
      </el-space>
    </div>

    <el-card class="page-card" shadow="never">
      <el-table :data="categories" v-loading="loading" style="width: 100%;">
        <el-table-column prop="name" label="分类名称" min-width="180" />
        <el-table-column label="商品数" width="100">
          <template #default="{ row }">
            {{ row.product_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="排序" width="160">
          <template #default="{ row }">
            <el-input-number v-model="row.sort" :min="0" :precision="0" :step="1" style="width: 120px;" />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="Number(row.status || 0) === 1 ? 'success' : 'info'">
              {{ getCategoryStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="168">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-space wrap>
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="primary" @click="toggleStatus(row)">
                {{ Number(row.status || 0) === 1 ? '停用' : '启用' }}
              </el-button>
              <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingCategory ? '编辑分类' : '新增分类'" width="480px">
      <el-form label-width="92px">
        <el-form-item label="分类名称" required>
          <el-input v-model="form.name" maxlength="30" show-word-limit placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :precision="0" :step="1" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="dialogSubmitting" @click="submitForm">保存</el-button>
        </el-space>
      </template>
    </el-dialog>
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
  justify-content: flex-end;
}
</style>
