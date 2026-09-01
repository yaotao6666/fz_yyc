<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getHealthEducationCategories,
  createHealthEducationCategory,
  updateHealthEducationCategory,
  deleteHealthEducationCategory
} from '@/api/sp'
import type { HealthEducationCategory } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<HealthEducationCategory[]>([])

// 一级分类（parent_id=0），用于新建二级分类时选择父级
const topLevel = computed(() => list.value.filter((c) => c.parent_id === 0))

// 根据 category_id 取一级分类展示名
function parentName(c: HealthEducationCategory): string {
  if (c.parent_id === 0) return '—'
  const parent = list.value.find((p) => p.id === c.parent_id)
  return parent ? parent.name : '未知'
}

async function loadData() {
  loading.value = true
  try {
    list.value = await getHealthEducationCategories()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

/* ----- 新增/编辑弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const editingId = ref(0)
const saving = ref(false)

const editForm = reactive({
  parent_id: 0,
  name: '',
  sort: 0,
  status: 1
})

function openCreate() {
  editingId.value = 0
  dialogMode.value = 'create'
  Object.assign(editForm, { parent_id: 0, name: '', sort: 0, status: 1 })
  dialogVisible.value = true
}

function openEdit(row: HealthEducationCategory) {
  editingId.value = row.id
  dialogMode.value = 'update'
  Object.assign(editForm, {
    parent_id: row.parent_id || 0,
    name: row.name || '',
    sort: row.sort || 0,
    status: row.status ?? 1
  })
  dialogVisible.value = true
}

async function handleSave() {
  if (!editForm.name.trim()) {
    ElMessage.warning('请填写分类名称')
    return
  }
  const payload = {
    parent_id: editForm.parent_id,
    name: editForm.name.trim(),
    sort: editForm.sort,
    status: editForm.status
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createHealthEducationCategory(payload)
      ElMessage.success('创建成功')
    } else {
      await updateHealthEducationCategory(editingId.value, payload)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

/* ----- 启停切换 ----- */
async function toggleStatus(row: HealthEducationCategory) {
  const next = row.status === 1 ? 0 : 1
  try {
    await updateHealthEducationCategory(row.id, {
      parent_id: row.parent_id,
      name: row.name,
      sort: row.sort,
      status: next
    })
    row.status = next
    ElMessage.success(next === 1 ? '已启用' : '已停用')
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

/* ----- 删除 ----- */
async function handleDelete(row: HealthEducationCategory) {
  try {
    await ElMessageBox.confirm(`确认删除分类「${row.name}」？存在子分类或关联文章时将无法删除。`, '删除', { type: 'warning' })
    await deleteHealthEducationCategory(row.id)
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
      <div class="toolbar-title">健康宣教分类（支持两级）</div>
      <div>
        <el-button type="primary" v-permission="'education-categories:create'" @click="openCreate">新建分类</el-button>
        <el-button @click="loadData">刷新</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="ID" width="80" prop="id" />
      <el-table-column label="分类名称" min-width="200">
        <template #default="{ row }">
          <span :class="{ 'sub-category': row.parent_id !== 0 }">{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="层级" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.parent_id === 0 ? 'primary' : 'info'" effect="plain">
            {{ row.parent_id === 0 ? '一级' : '二级' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="所属父分类" width="160">
        <template #default="{ row }">{{ parentName(row) }}</template>
      </el-table-column>
      <el-table-column label="排序" width="80" prop="sort" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="190" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" v-permission="'education-categories:update'" @click="openEdit(row)">编辑</el-button>
          <el-button link :type="row.status === 1 ? 'warning' : 'success'" v-permission="'education-categories:update'" @click="toggleStatus(row)">
            {{ row.status === 1 ? '停用' : '启用' }}
          </el-button>
          <el-button link type="danger" v-permission="'education-categories:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && list.length === 0" description="暂无分类，请先新建" />

    <!-- 新建/编辑分类弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建分类' : '编辑分类'"
      width="480px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form label-width="90px">
        <el-form-item label="父分类">
          <el-select v-model="editForm.parent_id" style="width: 100%">
            <el-option :value="0" label="一级分类（作为顶级）" />
            <el-option v-for="c in topLevel" :key="c.id" :value="c.id" :label="c.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类名称" required>
          <el-input v-model="editForm.name" placeholder="分类名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editForm.sort" :min="0" :step="1" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
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
.toolbar-title { font-size: 15px; font-weight: 600; color: #111827; }
.sub-category { padding-left: 22px; color: #6b7280; }
</style>