<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { approveStaffAudit, getStaffAuditDetail, listStaffAudits, rejectStaffAudit } from '@/api/sp'
import type { StaffAuditItem } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<StaffAuditItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const auditTypeFilter = ref<number | string>('')
const statusFilter = ref<number | string>('')
const keyword = ref('')

const auditTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '注册申请', value: 1 },
  { label: '信息变更', value: 2 },
  { label: '资质提交', value: 3 },
  { label: '状态变更', value: 4 }
]
const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待审核', value: 0 },
  { label: '已通过', value: 1 },
  { label: '已驳回', value: 2 }
]

function auditTypeText(t: number) {
  return { 1: '注册申请', 2: '信息变更', 3: '资质提交', 4: '状态变更' }[t] || '未知'
}
function auditTypeType(t: number) {
  return ({ 1: 'primary', 2: 'warning', 3: 'success', 4: 'info' } as const)[t] || 'info'
}
function statusText(s: number) {
  return { 0: '待审核', 1: '已通过', 2: '已驳回' }[s] || '未知'
}
function statusType(s: number) {
  return ({ 0: 'warning', 1: 'success', 2: 'danger' } as const)[s] || 'info'
}

// 资质材料展示
function normalizeQuals(quals: any): { type: string; name: string; url: string }[] {
  if (!quals) return []
  if (Array.isArray(quals)) return quals.map((q: any) => ({ type: q?.type || '', name: q?.name || '', url: q?.url || '' }))
  if (typeof quals === 'string' && quals.trim()) {
    try {
      const parsed = JSON.parse(quals)
      if (Array.isArray(parsed)) return parsed.map((q: any) => ({ type: q?.type || '', name: q?.name || '', url: q?.url || '' }))
    } catch { /* ignore */ }
  }
  return []
}

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (auditTypeFilter.value !== '') params.audit_type = auditTypeFilter.value
    if (statusFilter.value !== '') params.status = statusFilter.value
    if (keyword.value) params.keyword = keyword.value
    const res = await listStaffAudits(params as any)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref<any>(null)
const detailQuals = ref<any[]>([])

async function openDetail(row: StaffAuditItem) {
  try {
    const res = await getStaffAuditDetail(row.id)
    detail.value = res
    detailQuals.value = normalizeQuals(res.qualifications)
    detailVisible.value = true
  } catch (e) {
    // 错误已由拦截器提示
  }
}

async function handleApprove(row: StaffAuditItem) {
  try {
    await ElMessageBox.confirm(`确认通过该${auditTypeText(row.audit_type)}？`, '审核通过', { type: 'warning' })
    await approveStaffAudit(row.id)
    ElMessage.success('已通过')
    detailVisible.value = false
    loadData()
  } catch (e) { /* 取消或失败 */ }
}

async function handleReject(row: StaffAuditItem) {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回', {
      inputPattern: /^.{1,256}$/,
      inputErrorMessage: '请输入驳回原因'
    })
    await rejectStaffAudit(row.id, value)
    ElMessage.success('已驳回')
    detailVisible.value = false
    loadData()
  } catch (e) { /* 取消 */ }
}

function handlePageChange(p: number) {
  page.value = p
  loadData()
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="filter-bar">
      <el-select v-model="auditTypeFilter" style="width: 140px" @change="() => { page = 1; loadData() }">
        <el-option v-for="o in auditTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
      </el-select>
      <el-select v-model="statusFilter" style="width: 140px" @change="() => { page = 1; loadData() }">
        <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
      </el-select>
      <el-input v-model="keyword" placeholder="姓名/手机号/用户名" style="width: 220px" clearable @clear="() => { page = 1; loadData() }" @keyup.enter="() => { page = 1; loadData() }" />
      <el-button type="primary" @click="() => { page = 1; loadData() }">搜索</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="类型" width="110">
        <template #default="{ row }">
          <el-tag :type="auditTypeType(row.audit_type)">{{ auditTypeText(row.audit_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="staff_name" label="姓名" width="120" />
      <el-table-column prop="staff_phone" label="手机号" width="140" />
      <el-table-column prop="staff_username" label="用户名" width="130" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="review_remark" label="审核备注" min-width="150" show-overflow-tooltip />
      <el-table-column label="申请时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openDetail(row)">详情</el-button>
          <template v-if="row.status === 0">
            <el-button size="small" type="success" v-permission="'staffaudit:approve'" @click="handleApprove(row)">通过</el-button>
            <el-button size="small" type="danger" v-permission="'staffaudit:reject'" @click="handleReject(row)">驳回</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      @current-change="handlePageChange"
    />

    <!-- 审核详情弹窗 -->
    <el-dialog v-model="detailVisible" title="审核详情" width="640px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="审核类型">{{ auditTypeText(detail.record?.audit_type) }}</el-descriptions-item>
          <el-descriptions-item label="审核状态">{{ statusText(detail.record?.status) }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ detail.staff?.name }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ detail.staff?.phone }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ detail.staff?.username }}</el-descriptions-item>
          <el-descriptions-item label="当前状态">{{ detail.staff?.status === 1 ? '启用' : detail.staff?.status === 2 ? '禁用' : '待审核' }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="detailQuals.length" class="qual-section">
          <div class="qual-title">资质材料</div>
          <div class="qual-list">
            <div v-for="(q, i) in detailQuals" :key="i" class="qual-item">
              <span class="qual-name">{{ q.name || ('材料' + (i + 1)) }}</span>
              <el-image v-if="q.url" :src="q.url" fit="cover" class="qual-img" :preview-src-list="[q.url]" preview-teleported />
              <span v-else class="qual-type">{{ q.type }}</span>
            </div>
          </div>
        </div>

        <div v-if="detail.before_data || detail.after_data" class="change-section">
          <div class="qual-title">变更内容</div>
          <div class="change-grid">
            <div class="change-col">
              <div class="change-label">变更前</div>
              <pre class="change-pre">{{ JSON.stringify(detail.before_data || {}, null, 2) }}</pre>
            </div>
            <div class="change-col">
              <div class="change-label">变更后</div>
              <pre class="change-pre">{{ JSON.stringify(detail.after_data || {}, null, 2) }}</pre>
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <template v-if="detail?.record?.status === 0">
          <el-button type="success" v-if="hasPerm('staffaudit:approve')" @click="handleApprove(detail.record)">通过</el-button>
          <el-button type="danger" v-if="hasPerm('staffaudit:reject')" @click="handleReject(detail.record)">驳回</el-button>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts">
import { useAuthStore } from '@/stores/auth'
export default {
  methods: {
    hasPerm(code: string) {
      return useAuthStore().hasPermission(code)
    }
  }
}
</script>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
.qual-section, .change-section { margin-top: 20px; }
.qual-title { font-size: 14px; font-weight: 600; color: #111827; margin-bottom: 12px; }
.qual-list { display: flex; flex-wrap: wrap; gap: 16px; }
.qual-item { width: 120px; }
.qual-name { font-size: 13px; color: #333; display: block; margin-bottom: 6px; }
.qual-type { font-size: 13px; color: #666; }
.qual-img { width: 120px; height: 120px; border-radius: 8px; border: 1px solid #e5e7eb; }
.change-grid { display: flex; gap: 16px; }
.change-col { flex: 1; }
.change-label { font-size: 13px; color: #666; margin-bottom: 6px; }
.change-pre { background: #f8f8f8; border-radius: 8px; padding: 10px; font-size: 12px; min-height: 60px; max-height: 200px; overflow: auto; white-space: pre-wrap; }
</style>