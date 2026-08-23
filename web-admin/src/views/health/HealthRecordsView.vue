<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getHealthRecords, getHealthRecord, updateHealthRecord } from '@/api/sp'
import type { HealthRecord } from '@/types/sp'
import { ChronicTagOptions } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<HealthRecord[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  assessment_level: '' as string,
  status: '' as number | ''
})

const assessmentLevelOptions = [
  { label: '全部等级', value: '' },
  { label: '低风险', value: '低风险' },
  { label: '中风险', value: '中风险' },
  { label: '高风险', value: '高风险' },
  { label: '需关注', value: '需关注' }
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '正常', value: 1 },
  { label: '待完善', value: 0 }
]

function genderText(g?: number) {
  if (g === 1) return '男'
  if (g === 2) return '女'
  return '-'
}

function getUserName(row: HealthRecord) {
  return row.real_name || row.user?.nickname || row.user?.phone || '匿名用户'
}

function statusText(s?: number) {
  return Number(s || 0) === 1 ? '正常' : '待完善'
}
function statusType(s?: number) {
  return Number(s || 0) === 1 ? 'success' : 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getHealthRecords({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      assessment_level: filters.assessment_level || undefined,
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
  filters.assessment_level = ''
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
const current = ref<HealthRecord | null>(null)
const editMode = ref(false)

async function openDetail(row: HealthRecord) {
  current.value = row
  editMode.value = false
  drawerVisible.value = true
  detailLoading.value = true
  try {
    const detail = await getHealthRecord(row.id)
    current.value = detail
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    detailLoading.value = false
  }
}

/* ----- 编辑 ----- */
const saving = ref(false)
const editForm = reactive({
  real_name: '',
  gender: null as number | null,
  height_cm: null as number | null,
  weight_kg: null as number | null,
  blood_type: '',
  chronic_tags: [] as string[],
  smoking: '',
  drinking: '',
  assessment_level: '',
  remark: ''
})

function openEdit() {
  if (!current.value) return
  editMode.value = true
  Object.assign(editForm, {
    real_name: current.value.real_name || '',
    gender: current.value.gender ?? null,
    height_cm: current.value.height_cm ?? null,
    weight_kg: current.value.weight_kg ?? null,
    blood_type: current.value.blood_type || '',
    chronic_tags: [...(current.value.chronic_tags || [])],
    smoking: current.value.smoking || '',
    drinking: current.value.drinking || '',
    assessment_level: current.value.assessment_level || '',
    remark: current.value.remark || ''
  })
}

async function handleSave() {
  if (!current.value) return
  saving.value = true
  try {
    const updated = await updateHealthRecord(current.value.id, {
      real_name: editForm.real_name.trim() || undefined,
      gender: editForm.gender ?? undefined,
      height_cm: editForm.height_cm,
      weight_kg: editForm.weight_kg,
      blood_type: editForm.blood_type.trim() || undefined,
      chronic_tags: editForm.chronic_tags,
      smoking: editForm.smoking.trim() || undefined,
      drinking: editForm.drinking.trim() || undefined,
      assessment_level: editForm.assessment_level || undefined,
      remark: editForm.remark.trim() || undefined
    })
    current.value = updated
    editMode.value = false
    ElMessage.success('保存成功')
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <el-input
          v-model="filters.keyword"
          placeholder="姓名/手机号"
          style="width: 220px"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="filters.assessment_level" placeholder="评估等级" style="width: 140px" clearable @change="handleSearch">
          <el-option v-for="o in assessmentLevelOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" style="width: 120px" clearable @change="handleSearch">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="用户" min-width="140">
        <template #default="{ row }">
          {{ getUserName(row) }}
        </template>
      </el-table-column>
      <el-table-column label="性别" width="80">
        <template #default="{ row }">
          {{ genderText(row.gender) }}
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="手机" width="130" />
      <el-table-column label="慢病标签" min-width="200">
        <template #default="{ row }">
          <el-tag v-for="tag in (row.chronic_tags || [])" :key="tag" size="small" type="warning" style="margin-right: 4px; margin-bottom: 2px">{{ tag }}</el-tag>
          <span v-if="!row.chronic_tags || !row.chronic_tags.length">-</span>
        </template>
      </el-table-column>
      <el-table-column label="评估等级" width="110">
        <template #default="{ row }">
          <el-tag v-if="row.assessment_level" size="small" :type="row.assessment_level === '高风险' ? 'danger' : row.assessment_level === '中风险' ? 'warning' : 'primary'">{{ row.assessment_level }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="建档时间" width="170">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
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

    <!-- 档案详情抽屉 -->
    <el-drawer v-model="drawerVisible" :title="`健康档案 #${current?.id ?? ''}`" size="640px" :close-on-click-modal="false">
      <div v-loading="detailLoading">
        <template v-if="current">
          <!-- 编辑模式 -->
          <el-form v-if="editMode" label-width="100px">
            <el-form-item label="真实姓名">
              <el-input v-model="editForm.real_name" placeholder="真实姓名" />
            </el-form-item>
            <el-form-item label="性别">
              <el-radio-group v-model="editForm.gender">
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="身高(cm)">
              <el-input-number v-model="editForm.height_cm" :min="0" :max="300" :controls="false" style="width: 160px" />
            </el-form-item>
            <el-form-item label="体重(kg)">
              <el-input-number v-model="editForm.weight_kg" :min="0" :max="500" :controls="false" style="width: 160px" />
            </el-form-item>
            <el-form-item label="血型">
              <el-select v-model="editForm.blood_type" clearable placeholder="选择血型" style="width: 160px">
                <el-option v-for="b in ['A', 'B', 'AB', 'O']" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
            <el-form-item label="慢病标签">
              <el-select
                v-model="editForm.chronic_tags"
                multiple
                allow-create
                filterable
                default-first-option
                placeholder="选择或输入慢病标签"
                style="width: 100%"
              >
                <el-option v-for="tag in ChronicTagOptions" :key="tag" :label="tag" :value="tag" />
              </el-select>
            </el-form-item>
            <el-form-item label="吸烟">
              <el-select v-model="editForm.smoking" clearable placeholder="吸烟情况" style="width: 160px">
                <el-option v-for="s in ['不吸烟', '偶尔吸烟', '经常吸烟', '已戒烟']" :key="s" :label="s" :value="s" />
              </el-select>
            </el-form-item>
            <el-form-item label="饮酒">
              <el-select v-model="editForm.drinking" clearable placeholder="饮酒情况" style="width: 160px">
                <el-option v-for="d in ['不饮酒', '偶尔饮酒', '经常饮酒', '已戒酒']" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
            <el-form-item label="评估等级">
              <el-select v-model="editForm.assessment_level" clearable placeholder="评估等级" style="width: 160px">
                <el-option v-for="o in assessmentLevelOptions.filter((x) => x.value !== '')" :key="o.value" :label="o.label" :value="o.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="editForm.remark" type="textarea" :rows="3" placeholder="备注" />
            </el-form-item>
          </el-form>

          <!-- 查看模式 -->
          <template v-else>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="真实姓名">{{ current.real_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ genderText(current.gender) }}</el-descriptions-item>
              <el-descriptions-item label="出生日期">{{ current.birth_date || '-' }}</el-descriptions-item>
              <el-descriptions-item label="身份证号">{{ current.id_card || '-' }}</el-descriptions-item>
              <el-descriptions-item label="手机">{{ current.phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="紧急联系人">{{ current.emergency_contact || '-' }} {{ current.emergency_phone ? `（${current.emergency_phone}）` : '' }}</el-descriptions-item>
              <el-descriptions-item label="身高/体重">{{ current.height_cm ? `${current.height_cm}cm` : '-' }} / {{ current.weight_kg ? `${current.weight_kg}kg` : '-' }}</el-descriptions-item>
              <el-descriptions-item label="血型">{{ current.blood_type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="吸烟">{{ current.smoking || '-' }}</el-descriptions-item>
              <el-descriptions-item label="饮酒">{{ current.drinking || '-' }}</el-descriptions-item>
              <el-descriptions-item label="评估等级">{{ current.assessment_level || '-' }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ statusText(current.status) }}</el-descriptions-item>
              <el-descriptions-item label="地址" :span="2">{{ current.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="备注" :span="2">{{ current.remark || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div class="section-title">病史信息</div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="既往病史">
                <el-tag v-for="t in (current.past_history || [])" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.past_history || !current.past_history.length">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="过敏史">
                <el-tag v-for="t in (current.allergy_history || [])" :key="t" size="small" type="danger" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.allergy_history || !current.allergy_history.length">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="家族史">
                <el-tag v-for="t in (current.family_history || [])" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.family_history || !current.family_history.length">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="手术史">
                <el-tag v-for="t in (current.surgery_history || [])" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.surgery_history || !current.surgery_history.length">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="用药清单">
                <el-tag v-for="t in (current.medication_list || [])" :key="t" size="small" type="success" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.medication_list || !current.medication_list.length">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="慢病标签">
                <el-tag v-for="t in (current.chronic_tags || [])" :key="t" size="small" type="warning" style="margin-right: 4px">{{ t }}</el-tag>
                <span v-if="!current.chronic_tags || !current.chronic_tags.length">-</span>
              </el-descriptions-item>
            </el-descriptions>

            <div class="section-title">评估记录（{{ (current.assessments || []).length }}）</div>
            <el-table :data="current.assessments || []" size="small" stripe>
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="form_name" label="量表" min-width="140" />
              <el-table-column label="总分" width="80">
                <template #default="{ row }">{{ row.total_score ?? '-' }}</template>
              </el-table-column>
              <el-table-column prop="level" label="等级" width="100" />
              <el-table-column label="结论" min-width="160" show-overflow-tooltip>
                <template #default="{ row }">{{ row.conclusion || '-' }}</template>
              </el-table-column>
              <el-table-column label="时间" width="160">
                <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
              </el-table-column>
            </el-table>
          </template>
        </template>
      </div>

      <template #footer>
        <template v-if="current">
          <template v-if="editMode">
            <el-button @click="editMode = false">取消</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
          </template>
          <template v-else>
            <el-button type="primary" v-permission="'health:update'" @click="openEdit">编辑</el-button>
            <el-button @click="drawerVisible = false">关闭</el-button>
          </template>
        </template>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 20px; }
.section-title { font-weight: 600; margin: 18px 0 10px; color: #303133; }
</style>
