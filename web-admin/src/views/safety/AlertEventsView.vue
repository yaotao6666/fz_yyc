<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAlertEvents, getAlertSettings, handleAlertEvent, updateAlertSettings, type AlertEvent, type AlertSettings } from '@/api/safety'

/* ============ 字典 ============ */
const alertTypeOptions = [
  { label: 'SOS求助', value: 1 },
  { label: '服务超时未结束', value: 2 },
  { label: '实物超时未核销', value: 3 },
  { label: '服务超时未指派', value: 4 },
  { label: '陪诊超时未完成', value: 5 },
  { label: '指派超时未签到', value: 6 },
  { label: '租赁逾期未归还', value: 7 },
  { label: '退款卡在处理中', value: 8 }
]
const alertStatusOptions = [
  { label: '待处理', value: 1 },
  { label: '处理中', value: 2 },
  { label: '已处理', value: 3 }
]

function alertStatusType(s: number) {
  switch (s) {
    case 1: return 'danger'
    case 2: return 'warning'
    default: return 'success'
  }
}

function formatDateTime(t: string | null | undefined) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

/* ============ 列表 ============ */
const loading = ref(false)
const list = ref<AlertEvent[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, alert_type: '', status: '', keyword: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getAlertEvents({
      page: query.page,
      page_size: query.page_size,
      alert_type: query.alert_type || undefined,
      status: query.status || undefined,
      keyword: query.keyword || undefined
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

/* ============ 处理预警 ============ */
const handleVisible = ref(false)
const handling = ref(false)
const currentEvent = ref<AlertEvent | null>(null)
const handleForm = reactive({ status: 2, remark: '' })

function openHandle(row: AlertEvent) {
  currentEvent.value = row
  handleForm.status = row.status === 1 ? 2 : 3
  handleForm.remark = ''
  handleVisible.value = true
}

async function confirmHandle() {
  if (!currentEvent.value) return
  if (!handleForm.remark.trim()) {
    ElMessage.warning('请填写处理备注（留痕）')
    return
  }
  handling.value = true
  try {
    await handleAlertEvent(currentEvent.value.id, handleForm.status, handleForm.remark.trim())
    ElMessage.success('处理成功')
    handleVisible.value = false
    loadList()
  } catch (_e) {
    // 拦截器提示
  } finally {
    handling.value = false
  }
}

/* ============ 详情 ============ */
const detailVisible = ref(false)

async function openDetail(row: AlertEvent) {
  currentEvent.value = row
  detailVisible.value = true
}

/* ============ 快捷处理 ============ */
async function quickHandle(row: AlertEvent) {
  try {
    const { value } = await ElMessageBox.prompt('请填写处理备注（留痕）', '处理预警', {
      confirmButtonText: '标记已处理',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputValidator: (v: string) => (v && v.trim() ? true : '处理备注不能为空')
    })
    await handleAlertEvent(row.id, 3, value.trim())
    ElMessage.success('已处理')
    loadList()
  } catch (_e) {
    // 取消
  }
}

/* ============ 预警设置 ============ */
const settingsVisible = ref(false)
const settingsLoading = ref(false)
const settingsSaving = ref(false)
const settingsForm = reactive<AlertSettings>({
  enabled: false,
  goods_unverified_hours: 24,
  service_unassigned_hours: 24,
  escort_unfinished_minutes: 60,
  service_unstarted_minutes: 60,
  rental_overdue_hours: 24,
  refund_stuck_hours: 24
})

async function openSettings() {
  settingsVisible.value = true
  settingsLoading.value = true
  try {
    const res = await getAlertSettings()
    Object.assign(settingsForm, res)
  } catch (_e) {
    // 拦截器提示
  } finally {
    settingsLoading.value = false
  }
}

async function saveSettings() {
  settingsSaving.value = true
  try {
    await updateAlertSettings({ ...settingsForm })
    ElMessage.success('保存成功')
    settingsVisible.value = false
  } catch (_e) {
    // 拦截器提示
  } finally {
    settingsSaving.value = false
  }
}

onMounted(loadList)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">预警中心</h1>
        <p class="page-subtitle">服务人员 SOS 求助与服务超时预警事件的处理与留痕。</p>
      </div>
      <el-button v-permission="'alert-settings:update'" type="primary" plain @click="openSettings">预警设置</el-button>
    </div>

    <div class="page-card">
      <div class="filter-bar">
        <el-select v-model="query.alert_type" placeholder="预警类型" clearable style="width: 150px">
          <el-option v-for="o in alertTypeOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
        </el-select>
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 110px">
          <el-option v-for="o in alertStatusOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
        </el-select>
        <el-input v-model="query.keyword" placeholder="服务人员姓名/手机号" clearable style="width: 200px" @keyup.enter="handleSearch" />
        <el-button @click="handleSearch">查询</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="预警类型" width="130">
          <template #default="{ row }">
            <el-tag :type="row.alert_type === 1 ? 'danger' : 'warning'" effect="dark">
              {{ row.alert_type_cn }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="服务人员" min-width="130">
          <template #default="{ row }">
            <template v-if="row.staff">
              <div>{{ row.staff.name }}</div>
              <div class="sub-info">{{ row.staff.phone }}</div>
            </template>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column label="关联订单" min-width="150">
          <template #default="{ row }">
            <template v-if="row.order">
              <div>{{ row.order.order_no }}</div>
              <div class="sub-info">{{ row.order.address || '-' }}</div>
            </template>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="summary" label="摘要" min-width="180">
          <template #default="{ row }">
            <span v-if="row.summary">{{ row.summary }}</span>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column label="触发位置" min-width="150">
          <template #default="{ row }">
            <span v-if="row.address">{{ row.address }}</span>
            <span v-else-if="row.lat && row.lng" class="sub-info">{{ row.lat }}, {{ row.lng }}</span>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="alertStatusType(row.status)">{{ row.status_cn }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理人" width="110">
          <template #default="{ row }">{{ row.handler_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="触发时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
            <el-button size="small" type="primary" v-permission="'alert-events:update'" @click="openHandle(row)">处理</el-button>
            <el-button
              v-if="row.status !== 3"
              size="small"
              type="success"
              v-permission="'alert-events:update'"
              @click="quickHandle(row)"
            >
              完结
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

    <!-- 预警详情弹窗 -->
    <el-dialog v-model="detailVisible" title="预警事件详情" width="560px">
      <template v-if="currentEvent">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="预警类型">{{ currentEvent.alert_type_cn }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentEvent.status_cn }}</el-descriptions-item>
          <el-descriptions-item label="服务人员" :span="2">
            {{ currentEvent.staff ? `${currentEvent.staff.name}（${currentEvent.staff.phone}）` : `ID: ${currentEvent.staff_id}` }}
          </el-descriptions-item>
          <el-descriptions-item label="关联订单" :span="2">
            {{ currentEvent.order ? currentEvent.order.order_no : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="摘要" :span="2">{{ currentEvent.summary || '-' }}</el-descriptions-item>
          <el-descriptions-item label="触发位置" :span="2">
            {{ currentEvent.address || (currentEvent.lat && currentEvent.lng ? `${currentEvent.lat}, ${currentEvent.lng}` : '-') }}
          </el-descriptions-item>
          <el-descriptions-item label="处理人">{{ currentEvent.handler_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="处理时间">{{ formatDateTime(currentEvent.handled_at) }}</el-descriptions-item>
          <el-descriptions-item label="处理备注" :span="2">{{ currentEvent.handle_remark || '-' }}</el-descriptions-item>
          <el-descriptions-item label="触发时间" :span="2">{{ formatDateTime(currentEvent.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <!-- 处理预警弹窗 -->
    <el-dialog v-model="handleVisible" title="处理预警" width="520px" destroy-on-close>
      <template v-if="currentEvent">
        <el-alert
          :title="`${currentEvent.alert_type_cn} · ${currentEvent.staff ? currentEvent.staff.name : 'ID:' + currentEvent.staff_id}`"
          :type="currentEvent.alert_type === 1 ? 'error' : 'warning'"
          :closable="false"
          class="handle-alert"
        />
        <el-form :model="handleForm" label-width="90px">
          <el-form-item label="处理状态">
            <el-radio-group v-model="handleForm.status">
              <el-radio :value="2">处理中</el-radio>
              <el-radio :value="3">已处理</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="处理备注" required>
            <el-input
              v-model="handleForm.remark"
              type="textarea"
              :rows="4"
              maxlength="512"
              show-word-limit
              placeholder="填写处理措施与结果（必填留痕）"
            />
          </el-form-item>
        </el-form>
      </template>
      <template #footer>
        <el-button @click="handleVisible = false">取消</el-button>
        <el-button type="primary" :loading="handling" @click="confirmHandle">确认</el-button>
      </template>
    </el-dialog>

    <!-- 预警设置弹窗 -->
    <el-dialog v-model="settingsVisible" title="预警设置" width="520px" destroy-on-close>
      <el-form v-loading="settingsLoading" :model="settingsForm" label-width="200px">
        <el-form-item label="预警总开关">
          <el-switch v-model="settingsForm.enabled" active-text="开启" inactive-text="关闭" />
        </el-form-item>
        <el-form-item label="实物超时未核销（小时）">
          <el-input-number v-model="settingsForm.goods_unverified_hours" :min="1" :max="720" />
        </el-form-item>
        <el-form-item label="服务超时未指派（小时）">
          <el-input-number v-model="settingsForm.service_unassigned_hours" :min="1" :max="720" />
        </el-form-item>
        <el-form-item label="陪诊超时未完成（分钟）">
          <el-input-number v-model="settingsForm.escort_unfinished_minutes" :min="1" :max="1440" />
        </el-form-item>
        <el-form-item label="指派超时未签到（分钟）">
          <el-input-number v-model="settingsForm.service_unstarted_minutes" :min="1" :max="1440" />
        </el-form-item>
        <el-form-item label="租赁逾期未归还（小时）">
          <el-input-number v-model="settingsForm.rental_overdue_hours" :min="1" :max="720" />
        </el-form-item>
        <el-form-item label="退款卡在处理中（小时）">
          <el-input-number v-model="settingsForm.refund_stuck_hours" :min="1" :max="720" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settingsVisible = false">取消</el-button>
        <el-button type="primary" :loading="settingsSaving" @click="saveSettings">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.handle-alert {
  margin-bottom: 16px;
}
</style>
