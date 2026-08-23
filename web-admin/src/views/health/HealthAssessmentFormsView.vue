<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAssessmentForms,
  createAssessmentForm,
  updateAssessmentForm,
  setAssessmentFormStatus,
  deleteAssessmentForm
} from '@/api/sp'
import type { AssessmentForm, AssessmentFormQuestion } from '@/types/sp'
import { AssessmentDimensionOptions } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<AssessmentForm[]>([])

async function loadData() {
  loading.value = true
  try {
    list.value = await getAssessmentForms()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

function dimensionText(d?: string) {
  return AssessmentDimensionOptions.find((o) => o.value === d)?.label || d || '-'
}

function statusText(s?: number) {
  return Number(s || 0) === 1 ? '启用' : '草稿'
}
function statusType(s?: number) {
  return Number(s || 0) === 1 ? 'success' : 'info'
}

/* ----- 新增/编辑弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const saving = ref(false)
const formId = ref(0)
const form = reactive({
  name: '',
  dimension: '',
  description: '',
  questions: [] as AssessmentFormQuestion[],
  score_rule: [] as { min: number; max: number; level: string; conclusion?: string }[]
})

// 生成题目唯一 key
let keySeq = 0
function genQuestionKey(): string {
  keySeq += 1
  return `q${Date.now()}_${keySeq}`
}

function emptyQuestion(): AssessmentFormQuestion {
  return { key: genQuestionKey(), title: '', options: [{ label: '选项1', score: 0 }] }
}

function openCreate() {
  keySeq = 0
  formId.value = 0
  dialogMode.value = 'create'
  Object.assign(form, {
    name: '',
    dimension: '',
    description: '',
    questions: [emptyQuestion()],
    score_rule: []
  })
  dialogVisible.value = true
}

function openUpdate(row: AssessmentForm) {
  keySeq = 0
  formId.value = row.id
  dialogMode.value = 'update'
  Object.assign(form, {
    name: row.name || '',
    dimension: row.dimension || '',
    description: row.description || '',
    questions: (row.questions || []).map((q, i) => ({
      key: q.key || `q${i + 1}`,
      title: q.title || '',
      options: (q.options || []).map((o) => ({ label: o.label, score: o.score }))
    })),
    score_rule: (row.score_rule || []).map((r) => ({ min: r.min, max: r.max, level: r.level, conclusion: r.conclusion }))
  })
  dialogVisible.value = true
}

/* ----- 题目编辑 ----- */
function addQuestion() {
  form.questions.push(emptyQuestion())
}
function removeQuestion(index: number) {
  form.questions.splice(index, 1)
}
function addOption(question: AssessmentFormQuestion) {
  question.options.push({ label: `选项${question.options.length + 1}`, score: 0 })
}
function removeOption(question: AssessmentFormQuestion, index: number) {
  question.options.splice(index, 1)
}

/* ----- 评分规则编辑 ----- */
function addScoreRule() {
  form.score_rule.push({ min: 0, max: 0, level: '', conclusion: '' })
}
function removeScoreRule(index: number) {
  form.score_rule.splice(index, 1)
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写量表名称')
    return
  }
  const cleanedQuestions = form.questions
    .filter((q) => q.title.trim() && q.options.length)
    .map((q) => ({
      key: q.key,
      title: q.title.trim(),
      options: q.options.map((o) => ({ label: o.label.trim(), score: Number(o.score || 0) }))
    }))
  if (!cleanedQuestions.length) {
    ElMessage.warning('请至少添加一道有效题目')
    return
  }
  const cleanedRules = form.score_rule
    .filter((r) => r.level.trim())
    .map((r) => ({ min: Number(r.min || 0), max: Number(r.max || 0), level: r.level.trim(), conclusion: r.conclusion?.trim() || undefined }))

  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      dimension: form.dimension || undefined,
      description: form.description.trim() || undefined,
      questions: cleanedQuestions,
      score_rule: cleanedRules
    }
    if (dialogMode.value === 'create') {
      await createAssessmentForm(payload)
      ElMessage.success('创建成功')
    } else {
      await updateAssessmentForm(formId.value, payload)
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

/* ----- 状态切换 ----- */
async function handleToggleStatus(row: AssessmentForm) {
  const nextStatus = Number(row.status || 0) === 1 ? 0 : 1
  const action = nextStatus === 1 ? '启用' : '设为草稿'
  if (nextStatus === 1 && (!row.questions || !row.questions.length)) {
    ElMessage.warning('该量表暂无题目，无法启用')
    return
  }
  try {
    await ElMessageBox.confirm(`确认${action}量表「${row.name}」？`, action, { type: 'warning' })
    await setAssessmentFormStatus(row.id, nextStatus)
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (_e) {
    // 取消或失败
  }
}

/* ----- 删除 ----- */
async function handleDelete(row: AssessmentForm) {
  try {
    await ElMessageBox.confirm(`确认删除量表「${row.name}」？此操作不可恢复。`, '删除', { type: 'warning' })
    await deleteAssessmentForm(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <span class="desc-text">共 {{ list.length }} 个评估量表，可新增、编辑、启停用与删除。</span>
      </div>
      <div>
        <el-button type="primary" v-permission="'assessment:create'" @click="openCreate">新增量表</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
      <el-table-column label="维度" width="160">
        <template #default="{ row }">
          {{ dimensionText(row.dimension) }}
        </template>
      </el-table-column>
      <el-table-column label="版本" width="80">
        <template #default="{ row }">
          v{{ row.version ?? 1 }}
        </template>
      </el-table-column>
      <el-table-column label="题目数" width="90">
        <template #default="{ row }">
          {{ (row.questions || []).length }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="170">
        <template #default="{ row }">
          {{ formatDateTime(row.updated_at || row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-permission="'assessment:update'" @click="openUpdate(row)">编辑</el-button>
          <el-button
            size="small"
            :type="Number(row.status || 0) === 1 ? 'warning' : 'success'"
            v-permission="'assessment:update'"
            @click="handleToggleStatus(row)"
          >{{ Number(row.status || 0) === 1 ? '停用' : '启用' }}</el-button>
          <el-button size="small" type="danger" v-permission="'assessment:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑量表弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增评估量表' : '编辑评估量表'"
      width="860px"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form label-width="90px">
        <el-form-item label="量表名称" required>
          <el-input v-model="form.name" placeholder="量表名称，如：ADL 日常生活能力评估" />
        </el-form-item>
        <el-form-item label="评估维度">
          <el-select v-model="form.dimension" clearable placeholder="选择评估维度" style="width: 280px">
            <el-option v-for="o in AssessmentDimensionOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="量表描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="量表用途/适用人群说明" />
        </el-form-item>

        <el-divider content-position="left">题目编辑</el-divider>
        <div v-for="(question, qIndex) in form.questions" :key="question.key" class="question-block">
          <div class="question-head">
            <span class="question-title">题目 {{ qIndex + 1 }}</span>
            <el-button link type="danger" @click="removeQuestion(qIndex)">删除题目</el-button>
          </div>
          <el-input v-model="question.title" placeholder="题目内容" class="question-input" />
          <div class="option-row" v-for="(option, oIndex) in question.options" :key="oIndex">
            <el-input v-model="option.label" placeholder="选项文案" style="width: 240px" />
            <span class="option-score-label">分值</span>
            <el-input-number v-model="option.score" :min="0" :max="1000" :controls="false" style="width: 100px" />
            <el-button link type="danger" @click="removeOption(question, oIndex)">移除</el-button>
          </div>
          <el-button size="small" @click="addOption(question)">添加选项</el-button>
        </div>
        <el-button type="primary" plain @click="addQuestion">添加题目</el-button>

        <el-divider content-position="left">评分规则</el-divider>
        <div v-for="(rule, rIndex) in form.score_rule" :key="rIndex" class="option-row">
          <span class="option-score-label">区间</span>
          <el-input-number v-model="rule.min" :controls="false" style="width: 90px" />
          <span class="option-score-label">-</span>
          <el-input-number v-model="rule.max" :controls="false" style="width: 90px" />
          <el-input v-model="rule.level" placeholder="等级，如：低风险" style="width: 140px" />
          <el-input v-model="rule.conclusion" placeholder="结论建议（可选）" style="width: 200px" />
          <el-button link type="danger" @click="removeScoreRule(rIndex)">移除</el-button>
        </div>
        <el-button type="primary" plain @click="addScoreRule">添加评分规则</el-button>
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
.filter-bar { display: flex; gap: 12px; align-items: center; }
.desc-text { color: #909399; font-size: 13px; }
.question-block { border: 1px solid #ebeef5; border-radius: 8px; padding: 12px; margin-bottom: 12px; }
.question-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.question-title { font-weight: 600; color: #303133; }
.question-input { margin-bottom: 8px; }
.option-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.option-score-label { color: #909399; font-size: 13px; white-space: nowrap; }
</style>
