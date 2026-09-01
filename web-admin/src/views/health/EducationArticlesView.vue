<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEducationArticles, createEducationArticle, updateEducationArticle, deleteEducationArticle, getHealthEducationCategories } from '@/api/sp'
import type { EducationArticle, HealthEducationCategory } from '@/types/sp'
import { ChronicTagOptions } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<EducationArticle[]>([])

// 宣教分类列表（两级，parent_id=0 为一级）
const categoryList = ref<HealthEducationCategory[]>([])
const categoryLoading = ref(false)

// 构造成一级+二级的级联选项
const categoryOptions = computed(() => {
  const tops = categoryList.value.filter((c) => c.parent_id === 0 && c.status === 1)
  return tops.map((top) => ({
    value: top.id,
    label: top.name,
    children: categoryList.value
      .filter((c) => c.parent_id === top.id && c.status === 1)
      .map((c) => ({ value: c.id, label: c.name }))
  }))
})

async function loadCategories() {
  categoryLoading.value = true
  try {
    categoryList.value = await getHealthEducationCategories()
  } catch (_e) {
    // 已由拦截器提示
  } finally {
    categoryLoading.value = false
  }
}

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  category: '',
  status: '' as number | ''
})

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 0 },
  { label: '发布', value: 1 }
]

// 状态文案与标签映射（0草稿 1发布）
const statusTextMap: Record<number, string> = { 0: '草稿', 1: '发布' }
const statusTagMap: Record<number, string> = { 0: 'info', 1: 'success' }

function statusText(s?: number) {
  return statusTextMap[Number(s)] || '未知'
}
function statusTag(s?: number) {
  return statusTagMap[Number(s)] || 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getEducationArticles({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      category: filters.category.trim() || undefined,
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
  filters.category = ''
  filters.status = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(p: number) {
  pagination.page = p
  loadData()
}

/* ----- 新增/编辑弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const editingId = ref(0)
const saving = ref(false)

const editForm = reactive({
  title: '',
  category_id: undefined as number | undefined,
  cover: '',
  content: '',
  tags: [] as string[],
  status: 0,
  publish_at: ''
})

function openCreate() {
  editingId.value = 0
  dialogMode.value = 'create'
  Object.assign(editForm, {
    title: '',
    category_id: undefined,
    cover: '',
    content: '',
    tags: [],
    status: 0,
    publish_at: ''
  })
  dialogVisible.value = true
}

function openEdit(row: EducationArticle) {
  editingId.value = row.id
  dialogMode.value = 'update'
  Object.assign(editForm, {
    title: row.title || '',
    category_id: row.category_id && row.category_id > 0 ? row.category_id : undefined,
    cover: row.cover || '',
    content: row.content || '',
    tags: row.tags || [],
    status: row.status ?? 0,
    publish_at: row.publish_at || ''
  })
  dialogVisible.value = true
}

async function handleSave() {
  if (!editForm.title.trim()) {
    ElMessage.warning('请填写文章标题')
    return
  }
  // 前端与后端保持一致：正文必填（后端 content binding:required）
  if (!editForm.content.trim()) {
    ElMessage.warning('请填写文章正文')
    return
  }
  // 发布（status=1）必须选择分类
  if (editForm.status === 1 && !editForm.category_id) {
    ElMessage.warning('发布前必须选择宣教分类')
    return
  }
  const payload = {
    title: editForm.title.trim(),
    category_id: editForm.category_id || 0,
    cover: editForm.cover.trim() || undefined,
    content: editForm.content,
    tags: editForm.tags,
    status: editForm.status,
    publish_at: editForm.publish_at || undefined
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createEducationArticle(payload)
      ElMessage.success('创建成功')
    } else {
      await updateEducationArticle(editingId.value, payload)
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

/* ----- 删除 ----- */
async function handleDelete(row: EducationArticle) {
  try {
    await ElMessageBox.confirm(`确认删除文章「${row.title}」？此操作不可恢复。`, '删除', { type: 'warning' })
    await deleteEducationArticle(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消或失败
  }
}

// 根据 category_id 或旧 category 字段获取分类展示名
function displayCategory(row: EducationArticle): string {
  if (row.category_id && row.category_id > 0) {
    const cat = categoryList.value.find((c) => c.id === row.category_id)
    if (cat) return cat.name
  }
  return row.category || '未分类'
}

onMounted(async () => {
  loadCategories()
  loadData()
})
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="文章标题"
          style="width: 220px"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-input
          v-model="filters.category"
          placeholder="分类"
          style="width: 140px"
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
      <div>
        <el-button type="primary" v-permission="'education:create'" @click="openCreate">新增文章</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="标题" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <span>{{ row.title || '-' }}</span>
          <span v-if="row.cover" class="cover-dot" title="有封面">🖼</span>
        </template>
      </el-table-column>
      <el-table-column label="分类" width="110">
        <template #default="{ row }">
          <el-tag size="small" type="primary" effect="plain">{{ displayCategory(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="定向标签" min-width="200">
        <template #default="{ row }">
          <template v-if="row.tags && row.tags.length">
            <el-tag v-for="tag in row.tags" :key="tag" size="small" effect="plain" class="tag-item">{{ tag }}</el-tag>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="浏览量" width="80">
        <template #default="{ row }">{{ row.views ?? 0 }}</template>
      </el-table-column>
      <el-table-column label="发布时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.publish_at) }}</template>
      </el-table-column>
      <el-table-column label="时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" v-permission="'education:create'" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" v-permission="'education:create'" @click="handleDelete(row)">删除</el-button>
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

    <!-- 新增/编辑宣教文章弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增文章' : '编辑文章'"
      width="680px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="标题" required>
          <el-input v-model="editForm.title" placeholder="文章标题" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="分类">
          <el-cascader
            v-model="editForm.category_id"
            :options="categoryOptions"
            :props="{ checkStrictly: true, emitPath: false, value: 'value', label: 'label', children: 'children' }"
            placeholder="请选择宣教分类"
            style="width: 100%"
            clearable
            :loading="categoryLoading"
          />
        </el-form-item>
        <el-form-item label="封面 URL">
          <el-input v-model="editForm.cover" placeholder="封面图片地址（可空）" />
        </el-form-item>
        <el-form-item label="正文" required>
          <el-input v-model="editForm.content" type="textarea" :rows="6" placeholder="文章正文内容（必填）" />
        </el-form-item>
        <el-form-item label="定向慢病标签">
          <el-select
            v-model="editForm.tags"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="选择或输入慢病标签"
            style="width: 100%"
          >
            <el-option v-for="tag in ChronicTagOptions" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 200px">
            <el-option v-for="o in statusOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布时间">
          <el-date-picker
            v-model="editForm.publish_at"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="发布时间（可空）"
            style="width: 260px"
          />
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
.cover-dot { margin-left: 6px; font-size: 12px; }
.tag-item { margin-right: 6px; }
</style>
