<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createCouponTemplate,
  deleteCouponTemplate,
  getCouponTemplates,
  grantCoupon,
  getUserCoupons,
  updateCouponTemplate,
  updateCouponTemplateStatus,
  type CouponTemplate,
  type CouponTemplatePayload,
  type UserCoupon
} from '@/api/coupon'

/* ============ 通用字典 ============ */
const typeOptions = [
  { label: '满减券', value: 1 },
  { label: '折扣券', value: 2 }
]
const scopeOptions = [
  { label: '全场通用', value: 1 },
  { label: '指定分类', value: 2 },
  { label: '指定商品', value: 3 }
]
const sourceOptions = [
  { label: '自主领取', value: 1 },
  { label: '系统发放', value: 2 },
  { label: '手动发放', value: 3 }
]
const ucStatusOptions = [
  { label: '未使用', value: 1 },
  { label: '已使用', value: 2 },
  { label: '已过期', value: 3 },
  { label: '已作废', value: 4 }
]

function typeText(t: number) {
  return typeOptions.find(o => o.value === t)?.label || t
}
function scopeText(s: number) {
  return scopeOptions.find(o => o.value === s)?.label || s
}
function sourceText(s: number) {
  return sourceOptions.find(o => o.value === s)?.label || s
}
function ucStatusText(s: number) {
  return ucStatusOptions.find(o => o.value === s)?.label || s
}
function ucStatusType(s: number) {
  switch (s) {
    case 1: return 'success'
    case 2: return 'info'
    case 3: return 'warning'
    default: return 'danger'
  }
}
function formatDateTime(t: string | null | undefined) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

/* ============ Tab 切换 ============ */
const activeTab = ref<'templates' | 'records'>('templates')

/* ============ 券模板列表 ============ */
const loading = ref(false)
const list = ref<CouponTemplate[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, name: '', status: '', type: '' })

async function loadTemplates() {
  loading.value = true
  try {
    const res = await getCouponTemplates({
      page: query.page,
      page_size: query.page_size,
      name: query.name || undefined,
      status: query.status || undefined,
      type: query.type || undefined
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
  loadTemplates()
}

function handlePageChange(p: number) {
  query.page = p
  loadTemplates()
}

/* ============ 新增/编辑弹窗 ============ */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)
const saving = ref(false)

const defaultForm = (): CouponTemplatePayload => ({
  name: '',
  type: 1,
  threshold_amount: 0,
  discount_amount: 0,
  discount_rate: 0.9,
  total_count: 0,
  per_user_limit: 1,
  valid_type: 2,
  valid_start_at: null,
  valid_end_at: null,
  valid_days: 7,
  apply_scope: 1,
  scope_ids: null,
  remark: ''
})

const form = ref<CouponTemplatePayload>(defaultForm())
const validRange = ref<[string, string] | null>(null)
const scopeIdsText = ref('')

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = null
  form.value = defaultForm()
  validRange.value = null
  scopeIdsText.value = ''
  dialogVisible.value = true
}

function openEdit(row: CouponTemplate) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.value = {
    name: row.name,
    type: row.type,
    threshold_amount: Number(row.threshold_amount) || 0,
    discount_amount: Number(row.discount_amount) || 0,
    discount_rate: Number(row.discount_rate) || 0,
    total_count: row.total_count ?? 0,
    per_user_limit: row.per_user_limit ?? 1,
    valid_type: row.valid_type,
    valid_start_at: row.valid_start_at,
    valid_end_at: row.valid_end_at,
    valid_days: row.valid_days ?? 7,
    apply_scope: row.apply_scope,
    scope_ids: Array.isArray(row.scope_ids) ? row.scope_ids : null,
    remark: row.remark || ''
  }
  validRange.value = row.valid_start_at && row.valid_end_at
    ? [formatDateTime(row.valid_start_at), formatDateTime(row.valid_end_at)]
    : null
  scopeIdsText.value = Array.isArray(row.scope_ids) ? row.scope_ids.join(',') : ''
  dialogVisible.value = true
}

function handleValidRangeChange(val: [string, string] | null) {
  if (val) {
    form.value.valid_start_at = val[0]
    form.value.valid_end_at = val[1]
  } else {
    form.value.valid_start_at = null
    form.value.valid_end_at = null
  }
}

function buildPayload(): CouponTemplatePayload {
  const payload = { ...form.value }
  // 适用范围：全场时不传 ID 列表
  if (payload.apply_scope === 1) {
    payload.scope_ids = null
  } else {
    const ids = scopeIdsText.value
      .split(/[,，\s]+/)
      .map(s => parseInt(s, 10))
      .filter(n => Number.isInteger(n) && n > 0)
    payload.scope_ids = ids.length > 0 ? ids : null
  }
  // 满减券清空折扣率，折扣券清空面值
  if (payload.type === 1) {
    payload.discount_rate = 0
  } else {
    payload.discount_amount = 0
  }
  // 领取后N天模式清空固定期限
  if (payload.valid_type === 2) {
    payload.valid_start_at = null
    payload.valid_end_at = null
  } else {
    payload.valid_days = 0
  }
  return payload
}

function validateForm(): string {
  if (!form.value.name) return '请输入券名称'
  if (form.value.type === 1 && form.value.discount_amount <= 0) return '满减券必须指定满减面值'
  if (form.value.type === 2 && (form.value.discount_rate <= 0 || form.value.discount_rate >= 1)) return '折扣率必须为 0~1 之间的小数（如 0.90）'
  if (form.value.valid_type === 1 && (!form.value.valid_start_at || !form.value.valid_end_at)) return '固定期限券必须选择起止时间'
  if (form.value.valid_type === 2 && form.value.valid_days <= 0) return '领取后有效天数必须大于 0'
  if (form.value.apply_scope !== 1 && !scopeIdsText.value) return '指定适用范围时必须填写分类或商品 ID'
  return ''
}

async function handleSave() {
  const msg = validateForm()
  if (msg) {
    ElMessage.warning(msg)
    return
  }
  const payload = buildPayload()
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createCouponTemplate(payload)
      ElMessage.success('新增成功')
    } else if (editingId.value != null) {
      await updateCouponTemplate(editingId.value, payload)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadTemplates()
  } catch (_e) {
    // 拦截器提示
  } finally {
    saving.value = false
  }
}

/* ============ 启停 / 删除 / 发放 ============ */
async function handleStatusChange(row: CouponTemplate) {
  const targetStatus = row.status === 1 ? 0 : 1
  const action = targetStatus === 1 ? '启用' : '停用'
  try {
    await ElMessageBox.confirm(`确认${action}「${row.name}」？停用后用户不可再领取，已领取的券仍可使用。`, action, { type: 'warning' })
    await updateCouponTemplateStatus(row.id, targetStatus)
    ElMessage.success(`${action}成功`)
    loadTemplates()
  } catch (_e) {
    // 取消或失败
  }
}

async function handleDelete(row: CouponTemplate) {
  try {
    await ElMessageBox.confirm(`确认删除「${row.name}」？已有用户领取的券不可删除。`, '删除', { type: 'warning' })
    await deleteCouponTemplate(row.id)
    ElMessage.success('删除成功')
    loadTemplates()
  } catch (_e) {
    // 取消
  }
}

const grantVisible = ref(false)
const grantTemplate = ref<CouponTemplate | null>(null)
const grantUserId = ref<number | null>(null)
const granting = ref(false)

function openGrant(row: CouponTemplate) {
  grantTemplate.value = row
  grantUserId.value = null
  grantVisible.value = true
}

async function handleGrant() {
  if (!grantTemplate.value || !grantUserId.value) {
    ElMessage.warning('请输入用户 ID')
    return
  }
  granting.value = true
  try {
    await grantCoupon(grantTemplate.value.id, grantUserId.value)
    ElMessage.success('发放成功')
    grantVisible.value = false
  } catch (_e) {
    // 拦截器提示
  } finally {
    granting.value = false
  }
}

/* ============ 领取/使用记录 ============ */
const recordsLoading = ref(false)
const records = ref<UserCoupon[]>([])
const recordsTotal = ref(0)
const recordsQuery = reactive({ page: 1, page_size: 10, user_id: '', template_id: '', status: '', source: '' })

async function loadRecords() {
  recordsLoading.value = true
  try {
    const res = await getUserCoupons({
      page: recordsQuery.page,
      page_size: recordsQuery.page_size,
      user_id: recordsQuery.user_id || undefined,
      template_id: recordsQuery.template_id || undefined,
      status: recordsQuery.status || undefined,
      source: recordsQuery.source || undefined
    })
    records.value = res.list || []
    recordsTotal.value = res.total || 0
  } catch (_e) {
    // 拦截器提示
  } finally {
    recordsLoading.value = false
  }
}

function handleTabChange(tab: string | number) {
  if (tab === 'records' && records.value.length === 0) {
    loadRecords()
  }
}

function handleRecordsSearch() {
  recordsQuery.page = 1
  loadRecords()
}

function handleRecordsPageChange(p: number) {
  recordsQuery.page = p
  loadRecords()
}

onMounted(loadTemplates)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">优惠券管理</h1>
        <p class="page-subtitle">创建满减/折扣券模板，控制发行量、限领与适用范围，并查看领取与核销记录。</p>
      </div>
      <el-button type="primary" v-permission="'coupon-templates:create'" @click="openCreate">+ 新增券模板</el-button>
    </div>

    <div class="page-card">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 券模板 -->
        <el-tab-pane label="券模板" name="templates">
          <div class="filter-bar">
            <el-input v-model="query.name" placeholder="券名称" clearable style="width: 180px" @keyup.enter="handleSearch" />
            <el-select v-model="query.type" placeholder="券类型" clearable style="width: 120px">
              <el-option v-for="o in typeOptions" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
            <el-select v-model="query.status" placeholder="状态" clearable style="width: 110px">
              <el-option label="启用" value="1" />
              <el-option label="停用" value="0" />
            </el-select>
            <el-button @click="handleSearch">查询</el-button>
          </div>

          <el-table v-loading="loading" :data="list" stripe>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="券名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="券类型" width="90">
              <template #default="{ row }">
                <el-tag :type="row.type === 1 ? 'danger' : 'warning'" effect="plain">{{ typeText(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="优惠内容" min-width="140">
              <template #default="{ row }">
                <span v-if="row.type === 1">减 {{ row.discount_amount }} 元</span>
                <span v-else>{{ (row.discount_rate * 10).toFixed(1) }} 折</span>
                <div class="sub-info">
                  {{ row.threshold_amount > 0 ? `满 ${row.threshold_amount} 元可用` : '无门槛' }}
                </div>
              </template>
            </el-table-column>
            <el-table-column label="领取 / 总量" width="110">
              <template #default="{ row }">
                {{ row.received_count }} / {{ row.total_count > 0 ? row.total_count : '不限' }}
              </template>
            </el-table-column>
            <el-table-column label="限领" width="70">
              <template #default="{ row }">每人 {{ row.per_user_limit }} 张</template>
            </el-table-column>
            <el-table-column label="有效期" min-width="150">
              <template #default="{ row }">
                <span v-if="row.valid_type === 1">{{ formatDateTime(row.valid_start_at) }} ~ {{ formatDateTime(row.valid_end_at) }}</span>
                <span v-else>领取后 {{ row.valid_days }} 天</span>
              </template>
            </el-table-column>
            <el-table-column label="适用范围" width="100">
              <template #default="{ row }">{{ scopeText(row.apply_scope) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="300" fixed="right">
              <template #default="{ row }">
                <el-button size="small" v-permission="'coupon-templates:update'" @click="openEdit(row)">编辑</el-button>
                <el-button
                  size="small"
                  :type="row.status === 1 ? 'warning' : 'primary'"
                  v-permission="'coupon-templates:update'"
                  @click="handleStatusChange(row)"
                >
                  {{ row.status === 1 ? '停用' : '启用' }}
                </el-button>
                <el-button size="small" type="success" v-permission="'coupon-templates:create'" @click="openGrant(row)">发放</el-button>
                <el-button size="small" type="danger" v-permission="'coupon-templates:delete'" @click="handleDelete(row)">删除</el-button>
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
        </el-tab-pane>

        <!-- 领取/使用记录 -->
        <el-tab-pane label="领取/使用记录" name="records">
          <div class="filter-bar">
            <el-input v-model="recordsQuery.user_id" placeholder="用户 ID" clearable style="width: 130px" @keyup.enter="handleRecordsSearch" />
            <el-input v-model="recordsQuery.template_id" placeholder="模板 ID" clearable style="width: 130px" @keyup.enter="handleRecordsSearch" />
            <el-select v-model="recordsQuery.status" placeholder="券状态" clearable style="width: 120px">
              <el-option v-for="o in ucStatusOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
            </el-select>
            <el-select v-model="recordsQuery.source" placeholder="来源" clearable style="width: 120px">
              <el-option v-for="o in sourceOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
            </el-select>
            <el-button @click="handleRecordsSearch">查询</el-button>
          </div>

          <el-table v-loading="recordsLoading" :data="records" stripe>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="用户" min-width="150">
              <template #default="{ row }">
                <div>{{ row.nickname || '-' }}</div>
                <div class="sub-info">ID: {{ row.user_id }}</div>
              </template>
            </el-table-column>
            <el-table-column label="券信息" min-width="160">
              <template #default="{ row }">
                <div>{{ row.template_name || `模板 ${row.template_id}` }}</div>
                <div class="sub-info">
                  <span v-if="row.template_type === 1">满减券 · 减 {{ row.discount_amount }} 元</span>
                  <span v-else-if="row.template_type === 2">折扣券 · {{ ((row.discount_rate || 0) * 10).toFixed(1) }} 折</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="来源" width="90">
              <template #default="{ row }">{{ sourceText(row.source) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="ucStatusType(row.status)">{{ ucStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="领取时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.received_at) }}</template>
            </el-table-column>
            <el-table-column label="过期时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.expired_at) }}</template>
            </el-table-column>
            <el-table-column label="核销订单" min-width="140">
              <template #default="{ row }">
                <span v-if="row.order_no">{{ row.order_no }}</span>
                <span v-else class="empty-tip">-</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination-wrap" v-if="recordsTotal > recordsQuery.page_size">
            <el-pagination
              background
              layout="total, prev, pager, next"
              :total="recordsTotal"
              :page-size="recordsQuery.page_size"
              :current-page="recordsQuery.page"
              @current-change="handleRecordsPageChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 新增/编辑券模板弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增券模板' : '编辑券模板'" width="680px" destroy-on-close>
      <el-form :model="form" label-width="110px">
        <el-form-item label="券名称" required>
          <el-input v-model="form.name" placeholder="如：满100减20券" maxlength="64" show-word-limit />
        </el-form-item>

        <el-form-item label="券类型" required>
          <el-radio-group v-model="form.type">
            <el-radio v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="使用门槛">
          <el-input-number v-model="form.threshold_amount" :min="0" :precision="2" :step="10" />
          <span class="form-tip">0 表示无门槛</span>
        </el-form-item>

        <el-form-item v-if="form.type === 1" label="满减面值" required>
          <el-input-number v-model="form.discount_amount" :min="0" :precision="2" :step="5" />
          <span class="form-tip">元</span>
        </el-form-item>

        <el-form-item v-else label="折扣率" required>
          <el-input-number v-model="form.discount_rate" :min="0" :max="0.99" :precision="2" :step="0.05" />
          <span class="form-tip">如 0.90 表示 9 折，最高优惠不超过商品总额</span>
        </el-form-item>

        <el-form-item label="发行总量">
          <el-input-number v-model="form.total_count" :min="0" :step="100" />
          <span class="form-tip">0 表示不限量</span>
        </el-form-item>

        <el-form-item label="每人限领">
          <el-input-number v-model="form.per_user_limit" :min="1" :max="99" />
          <span class="form-tip">张</span>
        </el-form-item>

        <el-form-item label="有效期类型" required>
          <el-radio-group v-model="form.valid_type">
            <el-radio :value="1">固定期限</el-radio>
            <el-radio :value="2">领取后 N 天</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.valid_type === 1" label="起止时间" required>
          <el-date-picker
            v-model="validRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            @change="handleValidRangeChange"
          />
        </el-form-item>

        <el-form-item v-else label="有效天数" required>
          <el-input-number v-model="form.valid_days" :min="1" :max="365" />
          <span class="form-tip">领取后 N 天内有效</span>
        </el-form-item>

        <el-form-item label="适用范围" required>
          <el-radio-group v-model="form.apply_scope">
            <el-radio v-for="o in scopeOptions" :key="o.value" :value="o.value">{{ o.label }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.apply_scope !== 1" :label="form.apply_scope === 2 ? '分类 ID' : '商品 ID'" required>
          <el-input v-model="scopeIdsText" placeholder="多个 ID 用英文逗号分隔，如 12,34,56" />
          <span class="form-tip">{{ form.apply_scope === 2 ? '仅指定分类下的商品可用' : '仅指定商品可用' }}</span>
        </el-form-item>

        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="512" show-word-limit placeholder="内部备注（可选）" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ dialogMode === 'create' ? '确认新增' : '保存修改' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 手动发放弹窗 -->
    <el-dialog v-model="grantVisible" :title="`手动发放：${grantTemplate?.name || ''}`" width="440px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="用户 ID" required>
          <el-input-number v-model="grantUserId" :min="1" :controls="false" style="width: 100%" placeholder="请输入 C 端用户 ID" />
        </el-form-item>
      </el-form>
      <p class="grant-tip">发放即时生效，受模板限领与总量规则约束；可在「领取/使用记录」中查看结果。</p>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" :loading="granting" @click="handleGrant">确认发放</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts">
export default { name: 'CouponTemplatesView' }
</script>

<style scoped>
.page-shell { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-end; }
.page-title { margin: 0 0 4px; font-size: 22px; }
.page-subtitle { margin: 0; font-size: 13px; color: #6b7280; }
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.sub-info { color: #6b7280; font-size: 12px; }
.empty-tip { color: #9ca3af; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
.form-tip { margin-left: 12px; font-size: 12px; color: #9ca3af; }
.grant-tip { margin: 8px 0 0; font-size: 12px; color: #9ca3af; }
</style>
