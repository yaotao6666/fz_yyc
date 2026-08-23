<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { AssessmentForm, AssessmentFormQuestion, HealthAssessment } from '@/types'

const userId = ref<string>('')
const loading = ref(false)
const forms = ref<AssessmentForm[]>([])

// 当前作答的量表（null 表示量表列表页）
const currentForm = ref<AssessmentForm | null>(null)
// 各题答案 key -> 选中项 label
const answers = reactive<Record<string, string>>({})
const symptomDesc = ref('')
const submitting = ref(false)
// 提交成功后的评估结果
const result = ref<HealthAssessment | null>(null)

// 量表维度中文名
const dimensionTextMap: Record<string, string> = {
  adl: '日常生活能力',
  barthel: '巴氏指数',
  fall: '跌倒风险',
  nutrition: '营养评估',
  cognition: '认知评估',
  pressure: '压疮风险',
  weak: '衰弱评估',
  geriatric: '老年综合',
  self: '综合自评'
}

function dimensionText(dimension?: string) {
  if (!dimension) return ''
  return dimensionTextMap[dimension] || dimension
}

// 后端 JSON 列可能返回字符串，规范化为题目数组
function toQuestionArray(value: any): AssessmentFormQuestion[] {
  if (Array.isArray(value)) return value as AssessmentFormQuestion[]
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed as AssessmentFormQuestion[]
    } catch {
      // 解析失败视为无题目
    }
  }
  return []
}

// 是否全部题目已作答
const canSubmit = computed(() => {
  const questions = currentForm.value ? toQuestionArray(currentForm.value.questions) : []
  if (questions.length === 0) return false
  return questions.every(q => !!answers[q.key])
})

async function loadForms() {
  loading.value = true
  try {
    const data: any = await staffHealthApi.getAssessmentForms()
    forms.value = (data || []).map((f: any) => ({ ...f, questions: toQuestionArray(f.questions) }))
  } catch (e: any) {
    console.error('[HealthAssess] loadForms error', e)
  } finally {
    loading.value = false
  }
}

function startAssess(form: AssessmentForm) {
  Object.keys(answers).forEach(key => delete answers[key])
  symptomDesc.value = ''
  result.value = null
  currentForm.value = form
}

function selectOption(key: string, label: string) {
  answers[key] = label
}

function backToList() {
  currentForm.value = null
  result.value = null
}

async function handleSubmit() {
  if (!currentForm.value || !canSubmit.value) {
    uni.showToast({ title: '请完成全部题目', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const data: any = await staffHealthApi.createResidentAssessment(userId.value, {
      form_id: currentForm.value.id,
      answers: { ...answers },
      symptom_desc: symptomDesc.value
    })
    result.value = data
  } catch (e: any) {
    console.error('[HealthAssess] handleSubmit error', e)
  } finally {
    submitting.value = false
  }
}

function goBack() {
  uni.navigateBack()
}

onLoad((options: any) => {
  userId.value = options?.userId || ''
  if (!userId.value) {
    uni.showToast({ title: '缺少客户ID', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  loadForms()
})
</script>

<template>
  <view class="container">
    <!-- 量表列表 -->
    <template v-if="!currentForm">
      <view v-if="forms.length === 0 && !loading" class="empty card">
        <text class="empty-title">暂无可用评估量表</text>
        <text class="empty-desc">量表由管理端配置并启用后展示</text>
      </view>

      <view
        v-for="form in forms"
        :key="form.id"
        class="card form-card"
        @tap="startAssess(form)"
      >
        <view class="form-header">
          <text class="form-name">{{ form.name }}</text>
          <text v-if="form.dimension" class="form-dim">{{ dimensionText(form.dimension) }}</text>
        </view>
        <text v-if="form.description" class="form-desc">{{ form.description }}</text>
        <view class="form-footer">
          <text class="form-count">共 {{ form.questions.length }} 题</text>
          <text class="form-start">开始评估 ›</text>
        </view>
      </view>
    </template>

    <!-- 答题 -->
    <template v-else-if="!result">
      <view
        v-for="(q, idx) in currentForm.questions"
        :key="q.key"
        class="card question-card"
      >
        <view class="q-title">
          <text class="q-index">{{ idx + 1 }}</text>
          <text class="q-text">{{ q.title }}</text>
        </view>
        <view class="q-options">
          <view
            v-for="opt in q.options"
            :key="opt.label"
            class="q-option"
            :class="{ selected: answers[q.key] === opt.label }"
            @tap="selectOption(q.key, opt.label)"
          >
            <view class="q-radio" :class="{ checked: answers[q.key] === opt.label }"></view>
            <text class="q-option-label">{{ opt.label }}</text>
          </view>
        </view>
      </view>

      <!-- 主诉/需求描述 -->
      <view class="card question-card">
        <view class="q-title">
          <text class="q-index plus">＋</text>
          <text class="q-text">主诉/需求描述</text>
        </view>
        <textarea
          v-model="symptomDesc"
          class="symptom-input"
          placeholder="请描述客户的主诉、健康需求或服务背景（选填）"
          maxlength="500"
        />
      </view>

      <view class="footer-bar">
        <button class="btn-outline-bar" @tap="backToList">返回</button>
        <button class="btn-submit" :disabled="submitting" @tap="handleSubmit">
          {{ submitting ? '提交中...' : '提交评估' }}
        </button>
      </view>
    </template>

    <!-- 评估结果 -->
    <template v-else>
      <view class="result-card card">
        <view class="result-icon">✓</view>
        <view class="result-title">评估已完成</view>
        <view class="result-form">{{ result.form_name || (currentForm && currentForm.name) }}</view>
        <view class="result-row">
          <text class="result-label">总分</text>
          <text class="result-score">{{ result.total_score }}</text>
        </view>
        <view v-if="result.level" class="result-row">
          <text class="result-label">评估等级</text>
          <text class="result-level">{{ result.level }}</text>
        </view>
        <view v-if="result.conclusion" class="result-conclusion">
          <text class="result-label">评估结论</text>
          <text class="result-text">{{ result.conclusion }}</text>
        </view>
      </view>

      <view class="footer-bar">
        <button class="btn-outline-bar" @tap="backToList">继续评估</button>
        <button class="btn-submit" @tap="goBack">完成</button>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped>
.container {
  padding-bottom: 140rpx;
}
.empty {
  text-align: center;
  padding: 120rpx 32rpx !important;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}

/* 量表列表 */
.form-card {
  padding: 28rpx 24rpx;
}
.form-header {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
  .form-name { font-size: 32rpx; font-weight: 600; color: #333; }
  .form-dim {
    margin-left: 16rpx;
    font-size: 22rpx;
    color: var(--primary-color);
    background: rgba(81, 117, 40, 0.1);
    padding: 4rpx 12rpx;
    border-radius: 8rpx;
  }
}
.form-desc {
  display: block;
  font-size: 26rpx;
  color: #666;
  line-height: 1.5;
  margin-bottom: 16rpx;
}
.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  .form-count { font-size: 24rpx; color: #999; }
  .form-start { font-size: 28rpx; color: var(--primary-color); font-weight: 500; }
}

/* 答题 */
.question-card {
  padding: 28rpx 24rpx;
  margin-bottom: 20rpx;
}
.q-title {
  display: flex;
  align-items: flex-start;
  margin-bottom: 20rpx;
  .q-index {
    width: 40rpx; height: 40rpx;
    line-height: 40rpx;
    text-align: center;
    background: var(--primary-color);
    color: #fff;
    font-size: 24rpx;
    border-radius: 50%;
    flex-shrink: 0;
    margin-top: 4rpx;
    &.plus { background: var(--primary-color-light); font-size: 26rpx; }
  }
  .q-text { flex: 1; font-size: 30rpx; color: #333; line-height: 1.5; margin-left: 16rpx; font-weight: 500; }
}
.q-options { padding-left: 56rpx; }
.q-option {
  display: flex;
  align-items: center;
  padding: 20rpx 24rpx;
  border: 2rpx solid var(--border-color);
  border-radius: 12rpx;
  margin-bottom: 16rpx;
  background: #fafafa;
  transition: all 0.15s;
  &.selected {
    border-color: var(--primary-color);
    background: rgba(81, 117, 40, 0.06);
  }
}
.q-radio {
  width: 32rpx; height: 32rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  margin-right: 16rpx;
  flex-shrink: 0;
  box-sizing: border-box;
  &.checked {
    border-color: var(--primary-color);
    border-width: 8rpx;
  }
}
.q-option-label { font-size: 28rpx; color: #333; }

.symptom-input {
  width: 100%;
  height: 200rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
  margin-top: 8rpx;
}

/* 底部操作栏 */
.footer-bar {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  display: flex;
  gap: 20rpx;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}
.btn-submit {
  flex: 1.4;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &[disabled] { opacity: 0.5; }
}
.btn-outline-bar {
  flex: 1;
  font-size: 30rpx;
  color: var(--primary-color);
  background: #fff;
  border: 2rpx solid var(--primary-color);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
}

/* 评估结果 */
.result-card {
  margin-top: 24rpx;
  text-align: center;
  padding: 64rpx 40rpx;
}
.result-icon {
  width: 96rpx; height: 96rpx;
  line-height: 96rpx;
  margin: 0 auto 24rpx;
  background: var(--success-color);
  color: #fff;
  font-size: 52rpx;
  border-radius: 50%;
}
.result-title { font-size: 36rpx; font-weight: 600; color: #333; margin-bottom: 8rpx; }
.result-form { font-size: 26rpx; color: #999; margin-bottom: 32rpx; }
.result-row {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12rpx 0;
  .result-label { font-size: 28rpx; color: #666; margin-right: 16rpx; }
  .result-score { font-size: 44rpx; font-weight: 700; color: var(--primary-color); }
  .result-level {
    font-size: 30rpx; font-weight: 600; color: #fff;
    background: var(--warning-color);
    padding: 6rpx 24rpx;
    border-radius: 32rpx;
  }
}
.result-conclusion {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 2rpx solid var(--border-color);
  text-align: left;
  .result-label { display: block; font-size: 26rpx; color: #666; margin-bottom: 8rpx; }
  .result-text { font-size: 28rpx; color: #333; line-height: 1.6; }
}
</style>
