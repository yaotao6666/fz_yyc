<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { getHealthAssessments, getAssessmentForms } from '@/api/sp'
import type { HealthAssessment, AssessmentForm } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<HealthAssessment[]>([])
const forms = ref<AssessmentForm[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const filters = reactive({
  keyword: '',
  form_id: '' as number | ''
})

function getUserName(row: HealthAssessment) {
  return row.user?.nickname || row.user?.phone || '匿名用户'
}

function assessorTypeText(t?: number) {
  return Number(t || 1) === 2 ? '服务人员' : '自助'
}
function assessorTypeTag(t?: number) {
  return Number(t || 1) === 2 ? 'warning' : 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getHealthAssessments({
      page: pagination.page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      form_id: filters.form_id === '' ? undefined : filters.form_id
    })
    list.value = res.list || []
    pagination.total = res.total || 0
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

async function loadForms() {
  try {
    forms.value = await getAssessmentForms()
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  filters.keyword = ''
  filters.form_id = ''
  pagination.page = 1
  loadData()
}

function handlePageChange(p: number) {
  pagination.page = p
  loadData()
}

/* ----- 详情弹窗 ----- */
const detailVisible = ref(false)
const current = ref<HealthAssessment | null>(null)

// 将答案 key 映射为题目文本（优先使用记录内嵌 form 的题目）
function answerList(record: HealthAssessment) {
  const answers = record.answers || {}
  const questions = record.form?.questions || []
  return Object.entries(answers).map(([key, value]) => {
    const q = questions.find((item) => item.key === key)
    return { key, title: q?.title || key, value }
  })
}

function openDetail(row: HealthAssessment) {
  current.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadData()
  loadForms()
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
        <el-select v-model="filters.form_id" placeholder="评估量表" style="width: 220px" clearable filterable @change="handleSearch">
          <el-option v-for="f in forms" :key="f.id" :label="f.name" :value="f.id" />
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
      <el-table-column label="量表" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.form_name || row.form?.name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="评估类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="assessorTypeTag(row.assessor_type)">{{ assessorTypeText(row.assessor_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="总分" width="80">
        <template #default="{ row }">
          {{ row.total_score ?? '-' }}
        </template>
      </el-table-column>
      <el-table-column label="等级" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.level" size="small" :type="row.level === '高风险' ? 'danger' : row.level === '中风险' ? 'warning' : 'primary'">{{ row.level }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="评估时间" width="170">
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

    <!-- 评估记录详情弹窗 -->
    <el-dialog v-model="detailVisible" title="评估记录详情" width="640px" destroy-on-close>
      <template v-if="current">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="用户">{{ getUserName(current) }}</el-descriptions-item>
          <el-descriptions-item label="量表">{{ current.form_name || current.form?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="评估类型">{{ assessorTypeText(current.assessor_type) }}</el-descriptions-item>
          <el-descriptions-item label="评估时间">{{ formatDateTime(current.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="总分">{{ current.total_score ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="等级">{{ current.level || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="current.symptom_desc" class="detail-section">
          <div class="section-title">症状自述</div>
          <div class="section-content">{{ current.symptom_desc }}</div>
        </div>

        <div v-if="answerList(current).length" class="detail-section">
          <div class="section-title">答题明细</div>
          <el-table :data="answerList(current)" size="small" stripe>
            <el-table-column label="题目" min-width="220">
              <template #default="{ row }">{{ row.title }}</template>
            </el-table-column>
            <el-table-column label="答案" min-width="160">
              <template #default="{ row }">{{ row.value }}</template>
            </el-table-column>
          </el-table>
        </div>

        <div v-if="current.conclusion" class="detail-section">
          <div class="section-title">评估结论</div>
          <div class="section-content">{{ current.conclusion }}</div>
        </div>

        <div v-if="current.suggestions && current.suggestions.length" class="detail-section">
          <div class="section-title">健康建议</div>
          <ul class="suggestion-list">
            <li v-for="(s, i) in current.suggestions" :key="i">{{ s }}</li>
          </ul>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 20px; }
.detail-section { margin-top: 18px; }
.section-title { font-weight: 600; margin-bottom: 8px; color: #303133; }
.section-content { color: #606266; line-height: 1.6; }
.suggestion-list { margin: 0; padding-left: 20px; color: #606266; line-height: 1.8; }
</style>
