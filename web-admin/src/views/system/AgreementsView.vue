<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createAgreement,
  getAgreementDetail,
  getAgreements,
  publishAgreement,
  updateAgreement,
  type Agreement
} from '@/api/safety'

/* ============ 字典 ============ */
const typeOptions = [
  { label: '用户协议', value: 1 },
  { label: '隐私政策', value: 2 },
  { label: '录音/定位授权协议', value: 3 }
]

function formatDateTime(t: string | null | undefined) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

/* ============ 列表 ============ */
const loading = ref(false)
const list = ref<Agreement[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, type: '', status: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getAgreements({
      page: query.page,
      page_size: query.page_size,
      type: query.type || undefined,
      status: query.status || undefined
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch (_e) {
    // 拦截器提示
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  loadList()
}

function handlePageChange(p: number) {
  query.page = p
  loadList()
}

/* ============ 新建/编辑 ============ */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)
const saving = ref(false)
const form = reactive({
  type: 1,
  title: '',
  content: '',
  version: ''
})

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = null
  form.type = 1
  form.title = ''
  form.content = ''
  form.version = 'v1.0'
  dialogVisible.value = true
}

async function openEdit(row: Agreement) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  try {
    const detail = await getAgreementDetail(row.id)
    form.type = detail.type
    form.title = detail.title
    form.content = detail.content || ''
    form.version = detail.version
  } catch (_e) {
    return
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.title.trim()) {
    ElMessage.warning('请输入协议标题')
    return
  }
  if (!form.version.trim()) {
    ElMessage.warning('请输入版本号')
    return
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createAgreement({
        type: form.type,
        title: form.title.trim(),
        content: form.content,
        version: form.version.trim()
      })
      ElMessage.success('创建成功（草稿）')
    } else if (editingId.value != null) {
      await updateAgreement(editingId.value, {
        title: form.title.trim(),
        content: form.content
      })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadList()
  } catch (_e) {
    // 拦截器提示
  } finally {
    saving.value = false
  }
}

/* ============ 发布 ============ */
async function handlePublish(row: Agreement) {
  try {
    await ElMessageBox.confirm(
      `确认发布「${row.title} ${row.version}」？发布后将成为该类型当前生效版本，旧版本将自动下线。`,
      '发布协议',
      { type: 'warning' }
    )
    await publishAgreement(row.id)
    ElMessage.success('发布成功')
    loadList()
  } catch (_e) {
    // 取消或失败
  }
}

/* ============ 预览 ============ */
const previewVisible = ref(false)
const previewRow = ref<Agreement | null>(null)

async function openPreview(row: Agreement) {
  try {
    const detail = await getAgreementDetail(row.id)
    previewRow.value = detail
    previewVisible.value = true
  } catch (_e) {
    // 拦截器提示
  }
}

onMounted(loadList)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">协议管理</h1>
        <p class="page-subtitle">维护用户协议、隐私政策与录音/定位授权协议，同类型仅一个版本发布生效。</p>
      </div>
      <el-button type="primary" v-permission="'agreements:create'" @click="openCreate">+ 新建协议</el-button>
    </div>

    <div class="page-card">
      <div class="filter-bar">
        <el-select v-model="query.type" placeholder="协议类型" clearable style="width: 180px">
          <el-option v-for="o in typeOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
        </el-select>
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="已发布" value="1" />
          <el-option label="草稿" value="0" />
        </el-select>
        <el-button @click="handleSearch">查询</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="协议类型" width="160">
          <template #default="{ row }">
            <el-tag :type="row.type === 3 ? 'danger' : row.type === 2 ? 'warning' : 'info'" effect="plain">
              {{ row.type_cn }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="version" label="版本号" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '已发布' : '草稿' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.published_at) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openPreview(row)">预览</el-button>
            <el-button
              size="small"
              type="primary"
              :disabled="row.status === 1"
              v-permission="'agreements:update'"
              @click="openEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="success"
              :disabled="row.status === 1"
              v-permission="'agreements:update'"
              @click="handlePublish(row)"
            >
              发布
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap" v-if="total > query.page_size">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="query.page_size"
          :current-page="query.page"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 新建/编辑协议弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建协议' : '编辑协议'"
      width="720px"
      destroy-on-close
    >
      <el-form :model="form" label-width="90px">
        <el-form-item label="协议类型">
          <el-select v-model="form.type" :disabled="dialogMode === 'edit'" style="width: 220px">
            <el-option v-for="o in typeOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本号">
          <el-input v-model="form.version" :disabled="dialogMode === 'edit'" placeholder="如 v1.0（同类型唯一）" style="width: 220px" />
        </el-form-item>
        <el-form-item label="协议标题" required>
          <el-input v-model="form.title" placeholder="协议标题" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="协议正文" required>
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="14"
            placeholder="协议正文（支持 HTML 富文本，小程序端以富文本渲染）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存草稿</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog v-model="previewVisible" title="协议预览" width="680px">
      <template v-if="previewRow">
        <div class="preview-header">
          <h3 class="preview-title">{{ previewRow.title }}</h3>
          <div class="preview-meta">{{ previewRow.type_cn }} · {{ previewRow.version }}</div>
        </div>
        <div class="preview-content" v-html="previewRow.content"></div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.preview-header {
  text-align: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.preview-title {
  margin: 0 0 8px;
  font-size: 18px;
}
.preview-meta {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.preview-content {
  max-height: 55vh;
  overflow-y: auto;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
