<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getFollowUpTasks,
  getFollowUpTask,
  createFollowUpTask,
  completeFollowUpTask,
  getHealthRecords,
  getServiceStaffList,
  getEducationArticles
} from '@/api/sp'
import type { EducationArticle, FollowUpTask, HealthRecord, ServiceStaffItem } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<FollowUpTask[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  task_type: '' as number | '',
  status: '' as number | ''
})

const taskTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '康复随访', value: 1 },
  { label: '租后回访', value: 2 },
  { label: '慢病随访', value: 3 },
  { label: '评估回访', value: 4 }
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '待执行', value: 0 },
  { label: '已完成', value: 1 },
  { label: '已跳过', value: 2 }
]

// 任务类型/来源/状态/随访方式文案与标签映射
const taskTypeTextMap: Record<number, string> = { 1: '康复随访', 2: '租后回访', 3: '慢病随访', 4: '评估回访' }
const taskTypeTagMap: Record<number, string> = { 1: 'success', 2: 'warning', 3: 'danger', 4: 'primary' }
const sourceTypeTextMap: Record<number, string> = { 1: '服务完成', 2: '租赁归还', 3: '评估完成', 4: '手动' }
const statusTextMap: Record<number, string> = { 0: '待执行', 1: '已完成', 2: '已跳过' }
const statusTagMap: Record<number, string> = { 0: 'warning', 1: 'success', 2: 'info' }
const contactMethodTextMap: Record<number, string> = { 1: '电话', 2: '上门', 3: '微信' }

function getUserName(row: FollowUpTask) {
  return row.user?.nickname || row.user?.phone || '匿名用户'
}
function getUserPhone(row: FollowUpTask) {
  return row.user?.phone || '-'
}
function taskTypeText(t?: number) {
  return taskTypeTextMap[Number(t)] || '-'
}
function taskTypeTag(t?: number) {
  return taskTypeTagMap[Number(t)] || 'info'
}
function sourceTypeText(t?: number) {
  return sourceTypeTextMap[Number(t)] || '-'
}
function statusText(s?: number) {
  return statusTextMap[Number(s)] || '未知'
}
function statusTag(s?: number) {
  return statusTagMap[Number(s)] || 'info'
}
function contactMethodText(m?: number) {
  return contactMethodTextMap[Number(m)] || '-'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getFollowUpTasks({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      task_type: filters.task_type === '' ? undefined : filters.task_type,
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
  filters.task_type = ''
  filters.status = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(p: number) {
  pagination.page = p
  loadData()
}

/* ----- 宣教文章选项（已发布，用于执行弹窗多选与结果标题展示） ----- */
const educationArticleOptions = ref<EducationArticle[]>([])
const educationArticleTitleMap = new Map<number, string>()

async function loadEducationArticles() {
  try {
    const res = await getEducationArticles({ status: 1, page: 1, page_size: 200 })
    educationArticleOptions.value = res.list || []
    educationArticleTitleMap.clear()
    educationArticleOptions.value.forEach((a) => educationArticleTitleMap.set(a.id, a.title))
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

function educationArticlesText(ids?: number[]) {
  if (!ids || !ids.length) return '-'
  return ids.map((id) => educationArticleTitleMap.get(id) || `宣教#${id}`).join('、')
}

/* ----- 详情抽屉 ----- */
const drawerVisible = ref(false)
const detailLoading = ref(false)
const current = ref<FollowUpTask | null>(null)

async function openDetail(row: FollowUpTask) {
  current.value = row
  drawerVisible.value = true
  detailLoading.value = true
  try {
    const detail = await getFollowUpTask(row.id)
    current.value = detail
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    detailLoading.value = false
  }
}

/* ----- 登记随访任务弹窗 ----- */
const createDialogVisible = ref(false)
const saving = ref(false)

const createForm = reactive({
  user_id: null as number | null,
  task_type: 1,
  plan_follow_time: '',
  staff_id: null as number | null,
  remark: ''
})

// 居民远程搜索选项（展示 real_name/phone）
const residentOptions = ref<HealthRecord[]>([])
const residentSearchLoading = ref(false)

async function searchResidents(keyword?: string) {
  residentSearchLoading.value = true
  try {
    const res = await getHealthRecords({ keyword: keyword?.trim() || undefined, page: 1, page_size: 50 })
    residentOptions.value = res.list || []
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    residentSearchLoading.value = false
  }
}

// 服务人员选项（仅启用中的）
const staffOptions = ref<ServiceStaffItem[]>([])

async function loadStaffOptions() {
  try {
    const res = await getServiceStaffList({ status: 1, page: 1, page_size: 100 })
    staffOptions.value = res.list || []
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

function residentLabel(record: HealthRecord) {
  const name = record.real_name || record.user?.nickname || `用户#${record.user_id}`
  return record.phone ? `${name}（${record.phone}）` : name
}

function staffLabel(staff: ServiceStaffItem) {
  const name = staff.name || staff.username
  return staff.phone ? `${name}（${staff.phone}）` : name
}

// 把当前任务关联的居民并入候选，保证下拉回显
function ensureResidentOption(record: HealthRecord) {
  if (!residentOptions.value.some((r) => r.id === record.id)) {
    residentOptions.value = [record, ...residentOptions.value]
  }
}

// 登记随访任务（从行内登记时预选居民与执行人员）
function openCreate(row?: FollowUpTask) {
  Object.assign(createForm, {
    user_id: row?.user_id ?? null,
    task_type: 1,
    plan_follow_time: '',
    staff_id: row?.staff_id ?? null,
    remark: ''
  })
  createDialogVisible.value = true
  searchResidents()
  loadStaffOptions()
  if (row?.user) {
    ensureResidentOption({
      id: row.user_id,
      user_id: row.user_id,
      real_name: row.user.nickname,
      phone: row.user.phone
    })
  }
}

async function handleCreateSave() {
  if (!createForm.user_id) {
    ElMessage.warning('请选择居民')
    return
  }
  saving.value = true
  try {
    await createFollowUpTask({
      user_id: createForm.user_id,
      task_type: createForm.task_type,
      plan_follow_time: createForm.plan_follow_time || undefined,
      staff_id: createForm.staff_id,
      remark: createForm.remark.trim() || undefined,
      source_type: 4 // 手动登记
    })
    ElMessage.success('登记成功')
    createDialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

/* ----- 执行随访弹窗 ----- */
const executeDialogVisible = ref(false)
const executingId = ref(0)
const executing = ref(false)

const executeForm = reactive({
  contact_method: 1,
  content: '',
  education_article_ids: [] as number[],
  satisfaction: 0,
  remark: ''
})

const contactMethodOptions = [
  { label: '电话', value: 1 },
  { label: '上门', value: 2 },
  { label: '微信', value: 3 }
]

// 打开执行弹窗：待执行清空表单，已完成回填已填结果便于调整
function openExecute(row: FollowUpTask) {
  executingId.value = row.id
  const result = row.result || {}
  Object.assign(executeForm, {
    contact_method: result.contact_method ?? row.contact_method ?? 1,
    content: result.content || '',
    education_article_ids: result.education_article_ids || [],
    satisfaction: result.satisfaction ?? 0,
    remark: result.remark || ''
  })
  executeDialogVisible.value = true
  loadEducationArticles()
}

async function handleExecute() {
  if (!executeForm.content.trim()) {
    ElMessage.warning('请填写随访内容')
    return
  }
  executing.value = true
  try {
    await completeFollowUpTask(executingId.value, {
      result: {
        contact_method: executeForm.contact_method,
        content: executeForm.content.trim(),
        education_article_ids: executeForm.education_article_ids,
        satisfaction: executeForm.satisfaction || undefined,
        remark: executeForm.remark.trim() || undefined
      }
    })
    ElMessage.success('执行成功')
    executeDialogVisible.value = false
    loadData()
    // 若详情抽屉正展示该任务，同步刷新
    if (current.value && current.value.id === executingId.value) {
      getFollowUpTask(executingId.value)
        .then((detail) => {
          current.value = detail
        })
        .catch(() => {
          // 错误已由拦截器提示
        })
    }
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    executing.value = false
  }
}

onMounted(() => {
  loadData()
  loadStaffOptions()
  loadEducationArticles()
})
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
        <el-select v-model="filters.task_type" placeholder="任务类型" style="width: 140px" clearable @change="handleSearch">
          <el-option v-for="o in taskTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" style="width: 130px" clearable @change="handleSearch">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
      <div>
        <el-button type="primary" v-permission="'followup:update'" @click="openCreate()">登记随访任务</el-button>
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
      <el-table-column label="任务类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="taskTypeTag(row.task_type)">{{ taskTypeText(row.task_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="来源" width="100">
        <template #default="{ row }">{{ sourceTypeText(row.source_type) }}</template>
      </el-table-column>
      <el-table-column label="计划随访时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.plan_follow_time) }}</template>
      </el-table-column>
      <el-table-column label="执行人员" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">{{ row.staff?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="完成时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.completed_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          <el-button
            v-if="row.status === 0 || row.status === 1"
            link
            type="primary"
            v-permission="'followup:update'"
            @click="openExecute(row)"
          >
            执行
          </el-button>
          <el-button link type="primary" v-permission="'followup:update'" @click="openCreate(row)">登记</el-button>
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

    <!-- 随访任务详情抽屉 -->
    <el-drawer v-model="drawerVisible" :title="`随访任务 #${current?.id ?? ''}`" size="640px" :close-on-click-modal="false">
      <div v-loading="detailLoading">
        <template v-if="current">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="用户">{{ getUserName(current) }}</el-descriptions-item>
            <el-descriptions-item label="手机">{{ getUserPhone(current) }}</el-descriptions-item>
            <el-descriptions-item label="任务类型">
              <el-tag size="small" :type="taskTypeTag(current.task_type)">{{ taskTypeText(current.task_type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="来源">{{ sourceTypeText(current.source_type) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag size="small" :type="statusTag(current.status)">{{ statusText(current.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="计划随访方式">{{ contactMethodText(current.contact_method) }}</el-descriptions-item>
            <el-descriptions-item label="计划随访时间">{{ formatDateTime(current.plan_follow_time) }}</el-descriptions-item>
            <el-descriptions-item label="执行人员">{{ current.staff?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(current.created_at) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="current.remark" class="detail-section">
            <div class="section-title">备注</div>
            <div class="section-content">{{ current.remark }}</div>
          </div>

          <div v-if="current.result" class="detail-section">
            <div class="section-title">执行结果</div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="随访方式">{{ contactMethodText(current.result.contact_method) }}</el-descriptions-item>
              <el-descriptions-item label="满意度">
                <template v-if="current.result.satisfaction">{{ current.result.satisfaction }} 分</template>
                <template v-else>-</template>
              </el-descriptions-item>
              <el-descriptions-item label="宣教内容" :span="2">
                {{ educationArticlesText(current.result.education_article_ids) }}
              </el-descriptions-item>
              <el-descriptions-item label="随访内容" :span="2">{{ current.result.content || '-' }}</el-descriptions-item>
              <el-descriptions-item label="结果备注" :span="2">{{ current.result.remark || '-' }}</el-descriptions-item>
              <el-descriptions-item label="完成时间" :span="2">{{ formatDateTime(current.completed_at) }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </template>
      </div>

      <template #footer>
        <template v-if="current">
          <el-button
            v-if="current.status === 0 || current.status === 1"
            type="primary"
            v-permission="'followup:update'"
            @click="openExecute(current)"
          >
            执行
          </el-button>
          <el-button @click="drawerVisible = false">关闭</el-button>
        </template>
      </template>
    </el-drawer>

    <!-- 登记随访任务弹窗 -->
    <el-dialog v-model="createDialogVisible" title="登记随访任务" width="560px" destroy-on-close :close-on-click-modal="false">
      <el-form label-width="100px">
        <el-form-item label="选择居民" required>
          <el-select
            v-model="createForm.user_id"
            filterable
            remote
            clearable
            :remote-method="searchResidents"
            :loading="residentSearchLoading"
            placeholder="输入姓名/手机号搜索居民"
            style="width: 100%"
          >
            <el-option v-for="r in residentOptions" :key="r.id" :label="residentLabel(r)" :value="r.user_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select v-model="createForm.task_type" style="width: 220px">
            <el-option v-for="o in taskTypeOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="计划随访时间">
          <el-date-picker
            v-model="createForm.plan_follow_time"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="计划随访时间（可空）"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item label="执行服务人员">
          <el-select v-model="createForm.staff_id" clearable placeholder="执行服务人员（可空）" style="width: 100%">
            <el-option v-for="s in staffOptions" :key="s.id" :label="staffLabel(s)" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" type="textarea" :rows="3" placeholder="任务备注（可空）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreateSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 执行随访弹窗 -->
    <el-dialog v-model="executeDialogVisible" title="执行随访" width="560px" destroy-on-close :close-on-click-modal="false">
      <el-form label-width="100px">
        <el-form-item label="随访方式">
          <el-select v-model="executeForm.contact_method" style="width: 200px">
            <el-option v-for="o in contactMethodOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="随访内容" required>
          <el-input v-model="executeForm.content" type="textarea" :rows="4" placeholder="随访沟通内容" />
        </el-form-item>
        <el-form-item label="宣教文章">
          <el-select
            v-model="executeForm.education_article_ids"
            multiple
            filterable
            placeholder="选择推送的健康宣教文章（可空）"
            style="width: 100%"
          >
            <el-option v-for="a in educationArticleOptions" :key="a.id" :label="a.title" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="满意度">
          <el-rate v-model="executeForm.satisfaction" :max="5" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="executeForm.remark" type="textarea" :rows="3" placeholder="随访结果备注（可空）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="executeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="executing" @click="handleExecute">提交</el-button>
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
</style>
