<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCarePlans,
  getCarePlan,
  createCarePlan,
  updateCarePlan,
  deleteCarePlan,
  getHealthRecords,
  getServiceStaffList
} from '@/api/sp'
import type { CarePlan, CarePlanItem, CareVisit, HealthRecord, ServiceStaffItem } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<CarePlan[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  plan_type: '' as number | '',
  status: '' as number | ''
})

const planTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '生活照料', value: 1 },
  { label: '基础护理', value: 2 },
  { label: '康复训练', value: 3 },
  { label: '综合康养', value: 4 }
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 0 },
  { label: '执行中', value: 1 },
  { label: '已暂停', value: 2 },
  { label: '已完成', value: 3 }
]

// 类型/状态文案与标签映射
const planTypeTextMap: Record<number, string> = { 1: '生活照料', 2: '基础护理', 3: '康复训练', 4: '综合康养' }
const planTypeTagMap: Record<number, string> = { 1: 'success', 2: 'primary', 3: 'warning', 4: 'danger' }
const statusTextMap: Record<number, string> = { 0: '草稿', 1: '执行中', 2: '已暂停', 3: '已完成' }
const statusTagMap: Record<number, string> = { 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }

function getUserName(row: CarePlan) {
  return row.user?.nickname || row.user?.phone || '匿名用户'
}
function getUserPhone(row: CarePlan) {
  return row.user?.phone || '-'
}
function planTypeText(t?: number) {
  return planTypeTextMap[Number(t)] || '-'
}
function planTypeTag(t?: number) {
  return planTypeTagMap[Number(t)] || 'info'
}
function statusText(s?: number) {
  return statusTextMap[Number(s)] || '未知'
}
function statusTag(s?: number) {
  return statusTagMap[Number(s)] || 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getCarePlans({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      plan_type: filters.plan_type === '' ? undefined : filters.plan_type,
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
  filters.plan_type = ''
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
const current = ref<CarePlan | null>(null)

async function openDetail(row: CarePlan) {
  current.value = row
  drawerVisible.value = true
  detailLoading.value = true
  try {
    const detail = await getCarePlan(row.id)
    current.value = detail
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    detailLoading.value = false
  }
}

/* ----- 新增/编辑弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const editingId = ref(0)
const saving = ref(false)

const editForm = reactive({
  user_id: null as number | null,
  name: '',
  plan_type: 1,
  start_date: '',
  end_date: '',
  frequency: '',
  goals: '',
  items: [] as CarePlanItem[],
  assigned_staff_id: null as number | null,
  order_id: null as number | null,
  status: 0
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

// 把当前计划关联的居民并入候选，保证远程下拉回显
function ensureResidentOption(record: HealthRecord) {
  if (!residentOptions.value.some((r) => r.id === record.id)) {
    residentOptions.value = [record, ...residentOptions.value]
  }
}

// 把当前计划指派的服务人员并入候选，保证下拉回显（可能已停用）
function ensureStaffOption(staff: ServiceStaffItem) {
  if (!staffOptions.value.some((s) => s.id === staff.id)) {
    staffOptions.value = [staff, ...staffOptions.value]
  }
}

function addItem() {
  editForm.items.push({ name: '', desc: '' })
}
function removeItem(index: number) {
  editForm.items.splice(index, 1)
}

function openCreate() {
  editingId.value = 0
  dialogMode.value = 'create'
  Object.assign(editForm, {
    user_id: null,
    name: '',
    plan_type: 1,
    start_date: '',
    end_date: '',
    frequency: '',
    goals: '',
    items: [{ name: '', desc: '' }],
    assigned_staff_id: null,
    order_id: null,
    status: 0
  })
  dialogVisible.value = true
  searchResidents()
}

async function openEdit(row: CarePlan) {
  editingId.value = row.id
  dialogMode.value = 'update'
  Object.assign(editForm, {
    user_id: row.user_id,
    name: row.name || '',
    plan_type: row.plan_type ?? 1,
    start_date: row.start_date || '',
    end_date: row.end_date || '',
    frequency: row.frequency || '',
    goals: row.goals || '',
    items: (row.items || []).length
      ? (row.items || []).map((i) => ({ name: i.name || '', desc: i.desc || '' }))
      : [{ name: '', desc: '' }],
    assigned_staff_id: row.assigned_staff_id ?? null,
    order_id: row.order_id ?? null,
    status: row.status ?? 0
  })
  dialogVisible.value = true
  // 先加载候选，再把当前关联的居民/服务人员并入，保证已选值可正常回显
  await Promise.all([searchResidents(), loadStaffOptions()])
  if (row.user) {
    ensureResidentOption({
      id: row.user_id,
      user_id: row.user_id,
      real_name: row.user.nickname,
      phone: row.user.phone
    })
  }
  if (row.assigned_staff && row.assigned_staff_id) {
    ensureStaffOption({
      id: row.assigned_staff_id,
      username: '',
      name: row.assigned_staff.name || '',
      phone: row.assigned_staff.phone || '',
      status: 1
    })
  }
}

async function handleSave() {
  if (!editForm.user_id) {
    ElMessage.warning('请选择居民')
    return
  }
  if (!editForm.name.trim()) {
    ElMessage.warning('请填写计划名称')
    return
  }
  const cleanedItems = editForm.items
    .filter((i) => i.name.trim())
    .map((i) => ({ name: i.name.trim(), desc: i.desc?.trim() || undefined }))
  const payload = {
    user_id: editForm.user_id,
    name: editForm.name.trim(),
    plan_type: editForm.plan_type,
    start_date: editForm.start_date || undefined,
    end_date: editForm.end_date || undefined,
    frequency: editForm.frequency.trim() || undefined,
    goals: editForm.goals.trim() || undefined,
    items: cleanedItems,
    assigned_staff_id: editForm.assigned_staff_id,
    order_id: editForm.order_id,
    status: editForm.status
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createCarePlan(payload)
      ElMessage.success('创建成功')
    } else {
      await updateCarePlan(editingId.value, payload)
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
async function handleDelete(row: CarePlan) {
  try {
    await ElMessageBox.confirm(`确认删除照护计划「${row.name}」？此操作不可恢复。`, '删除', { type: 'warning' })
    await deleteCarePlan(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消或失败
  }
}

/* ----- 照护记录展示辅助 ----- */
function nursingItemsText(visit: CareVisit) {
  const items = visit.nursing_items || []
  if (!items.length) return '-'
  return items
    .map((i) => i.name || '')
    .filter(Boolean)
    .join('、')
}

function vitalsText(visit: CareVisit) {
  const v = visit.vitals || {}
  const parts: string[] = []
  if (v.blood_pressure) parts.push(`血压 ${v.blood_pressure}`)
  if (v.blood_glucose) parts.push(`血糖 ${v.blood_glucose}`)
  if (v.heart_rate) parts.push(`心率 ${v.heart_rate}`)
  if (v.oxygen) parts.push(`血氧 ${v.oxygen}`)
  if (v.weight) parts.push(`体重 ${v.weight}`)
  return parts.join('；') || '-'
}

onMounted(() => {
  loadData()
  loadStaffOptions()
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
        <el-select v-model="filters.plan_type" placeholder="计划类型" style="width: 140px" clearable @change="handleSearch">
          <el-option v-for="o in planTypeOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" style="width: 130px" clearable @change="handleSearch">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
      <div>
        <el-button type="primary" v-permission="'care:create'" @click="openCreate">新增照护计划</el-button>
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
      <el-table-column prop="name" label="计划名称" min-width="160" show-overflow-tooltip />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="planTypeTag(row.plan_type)">{{ planTypeText(row.plan_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="照护周期" min-width="180">
        <template #default="{ row }">
          <span v-if="row.start_date || row.end_date">{{ row.start_date || '…' }} ~ {{ row.end_date || '…' }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="频次" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">{{ row.frequency || '-' }}</template>
      </el-table-column>
      <el-table-column label="指派服务人员" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.assigned_staff?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="照护次数" width="90">
        <template #default="{ row }">
          <el-tag size="small" type="primary">{{ row.visit_count ?? 0 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          <el-button link type="primary" v-permission="'care:update'" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" v-permission="'care:delete'" @click="handleDelete(row)">删除</el-button>
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

    <!-- 照护计划详情抽屉 -->
    <el-drawer v-model="drawerVisible" :title="`照护计划 #${current?.id ?? ''}`" size="760px" :close-on-click-modal="false">
      <div v-loading="detailLoading">
        <template v-if="current">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="用户">{{ getUserName(current) }}</el-descriptions-item>
            <el-descriptions-item label="手机">{{ getUserPhone(current) }}</el-descriptions-item>
            <el-descriptions-item label="计划名称" :span="2">{{ current.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="类型">
              <el-tag size="small" :type="planTypeTag(current.plan_type)">{{ planTypeText(current.plan_type) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag size="small" :type="statusTag(current.status)">{{ statusText(current.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="照护周期" :span="2">
              {{ current.start_date || '-' }} ~ {{ current.end_date || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="频次">{{ current.frequency || '-' }}</el-descriptions-item>
            <el-descriptions-item label="照护次数">{{ current.visit_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="指派服务人员">
              {{ current.assigned_staff?.name || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="关联订单">{{ current.order_id ? `#${current.order_id}` : '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(current.created_at) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="current.goals" class="detail-section">
            <div class="section-title">护理目标</div>
            <div class="section-content">{{ current.goals }}</div>
          </div>

          <div class="detail-section">
            <div class="section-title">护理项（{{ (current.items || []).length }}）</div>
            <el-table :data="current.items || []" size="small" stripe>
              <el-table-column prop="name" label="护理项" min-width="160" show-overflow-tooltip />
              <el-table-column prop="desc" label="说明" min-width="220" show-overflow-tooltip>
                <template #default="{ row }">{{ row.desc || '-' }}</template>
              </el-table-column>
            </el-table>
          </div>

          <div class="detail-section">
            <div class="section-title">照护记录（{{ (current.visits || []).length }}）</div>
            <el-table v-if="current.visits && current.visits.length" :data="current.visits" size="small" stripe>
              <el-table-column label="到访时间" width="150">
                <template #default="{ row }">{{ formatDateTime(row.visit_at) }}</template>
              </el-table-column>
              <el-table-column label="护理项摘要" min-width="160" show-overflow-tooltip>
                <template #default="{ row }">{{ nursingItemsText(row) }}</template>
              </el-table-column>
              <el-table-column label="生命体征" min-width="150" show-overflow-tooltip>
                <template #default="{ row }">{{ vitalsText(row) }}</template>
              </el-table-column>
              <el-table-column label="照片" width="110">
                <template #default="{ row }">
                  <el-image
                    v-for="(p, i) in row.photos || []"
                    :key="i"
                    :src="p"
                    :preview-src-list="row.photos || []"
                    :initial-index="i"
                    fit="cover"
                    preview-teleported
                    style="width: 40px; height: 40px; margin-right: 4px; border-radius: 4px"
                  />
                  <span v-if="!row.photos || !row.photos.length">-</span>
                </template>
              </el-table-column>
              <el-table-column label="备注" min-width="120" show-overflow-tooltip>
                <template #default="{ row }">{{ row.remark || '-' }}</template>
              </el-table-column>
              <el-table-column label="下次随访建议" min-width="150" show-overflow-tooltip>
                <template #default="{ row }">{{ row.follow_up_advice || '-' }}</template>
              </el-table-column>
              <el-table-column label="记录人" width="90">
                <template #default="{ row }">{{ row.staff?.name || '-' }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="暂无照护记录" :image-size="60" />
          </div>
        </template>
      </div>

      <template #footer>
        <template v-if="current">
          <el-button type="primary" v-permission="'care:update'" @click="openEdit(current)">编辑</el-button>
          <el-button @click="drawerVisible = false">关闭</el-button>
        </template>
      </template>
    </el-drawer>

    <!-- 新增/编辑照护计划弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增照护计划' : '编辑照护计划'"
      width="680px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form label-width="100px">
        <el-form-item label="选择居民" required>
          <el-select
            v-model="editForm.user_id"
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
        <el-form-item label="计划名称" required>
          <el-input v-model="editForm.name" placeholder="计划名称，如：居家生活照料计划" />
        </el-form-item>
        <el-form-item label="计划类型">
          <el-select v-model="editForm.plan_type" style="width: 200px">
            <el-option v-for="o in planTypeOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="照护周期">
          <el-date-picker
            v-model="editForm.start_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="开始日期"
            style="width: 170px"
          />
          <span class="date-separator">至</span>
          <el-date-picker
            v-model="editForm.end_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="结束日期"
            style="width: 170px"
          />
        </el-form-item>
        <el-form-item label="频次">
          <el-input v-model="editForm.frequency" placeholder="如：每周上门 2 次" style="width: 100%" />
        </el-form-item>
        <el-form-item label="护理目标">
          <el-input v-model="editForm.goals" type="textarea" :rows="3" placeholder="照护目标描述（可空）" />
        </el-form-item>
        <el-form-item label="护理项">
          <div style="width: 100%">
            <div v-for="(item, index) in editForm.items" :key="index" class="item-row">
              <el-input v-model="item.name" placeholder="护理项名称" style="width: 220px" />
              <el-input v-model="item.desc" placeholder="说明（可空）" style="flex: 1" />
              <el-button link type="danger" @click="removeItem(index)">移除</el-button>
            </div>
            <el-button type="primary" plain size="small" @click="addItem">添加护理项</el-button>
          </div>
        </el-form-item>
        <el-form-item label="指派服务人员">
          <el-select v-model="editForm.assigned_staff_id" clearable placeholder="指派服务人员（可空）" style="width: 100%">
            <el-option v-for="s in staffOptions" :key="s.id" :label="staffLabel(s)" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联订单">
          <el-input-number v-model="editForm.order_id" :min="1" :controls="false" clearable placeholder="订单号（可空）" style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 200px">
            <el-option v-for="o in statusOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
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
.item-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.date-separator { margin: 0 8px; color: #909399; }
</style>
