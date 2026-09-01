<template>
  <view class="assessment-page">
    <!-- 顶部 tab -->
    <view class="mode-tabs">
      <view class="tab-item" :class="{ active: mode === 'assess' }" @click="switchMode('assess')">自助评估</view>
      <view class="tab-item" :class="{ active: mode === 'records' }" @click="switchMode('records')">我的记录</view>
    </view>

    <!-- 自助评估 -->
    <template v-if="mode === 'assess'">
      <!-- 评估结果 -->
      <view v-if="result" class="result-view">
        <view class="result-icon">💚</view>
        <view class="result-score">
          {{ result.total_score ?? '--' }}<text class="result-score-unit">分</text>
        </view>
        <view class="result-level" :class="getLevelClass(result.level)">{{ result.level || '未分级' }}</view>
        <view class="result-form-name">{{ result.form_name || '健康评估' }}</view>
        <view class="result-conclusion" v-if="result.conclusion">{{ result.conclusion }}</view>
        <view class="result-suggestions" v-if="result.suggestions?.length">
          <view class="result-suggest-title">健康建议</view>
          <view v-for="(item, index) in result.suggestions" :key="index" class="suggestion-item">
            {{ index + 1 }}. {{ item }}
          </view>
        </view>
        <view class="result-actions">
          <view class="result-btn" @click="goHome">返回我的健康</view>
          <view class="result-btn secondary" @click="backToList">再测一次</view>
        </view>
      </view>

      <!-- 答题屏 -->
      <view v-else-if="currentForm" class="answer-view">
        <view class="answer-header">
          <view class="answer-title">{{ currentForm.name }}</view>
          <view class="answer-back" @click="backToList">返回列表</view>
        </view>
        <scroll-view class="answer-scroll" scroll-y>
          <view
            v-for="(question, index) in currentForm.questions"
            :key="question.key"
            class="question-block"
          >
            <view class="question-title">{{ index + 1 }}. {{ question.title }}</view>
            <view class="option-list">
              <view
                v-for="option in question.options"
                :key="option.label"
                class="option-item"
                :class="{ active: answers[question.key] === option.label }"
                @click="selectOption(question.key, option.label)"
              >{{ option.label }}</view>
            </view>
          </view>
          <view class="symptom-block">
            <view class="question-title">主诉 / 需求描述（选填）</view>
            <textarea
              v-model="symptomDesc"
              class="symptom-textarea"
              placeholder="请描述您当前的主要症状或健康需求"
              maxlength="200"
            />
          </view>
          <view class="answer-bottom-placeholder"></view>
        </scroll-view>
        <view class="submit-bar">
          <view class="submit-btn" :class="{ disabled: submitting }" @click="handleSubmit">
            {{ submitting ? '提交中...' : '提交评估' }}
          </view>
        </view>
      </view>

      <!-- 量表列表 -->
      <view v-else class="form-list-view">
        <view v-if="formsLoading" class="loading">加载中...</view>
        <view v-else-if="forms.length" class="form-list">
          <view
            v-for="form in forms"
            :key="form.id"
            class="form-card"
            @click="startForm(form)"
          >
            <view class="form-name">{{ form.name }}</view>
            <view class="form-desc" v-if="form.description">{{ form.description }}</view>
            <view class="form-meta">
              <text v-if="form.dimension" class="form-dimension">{{ form.dimension }}</text>
              <text class="form-count">{{ form.questions?.length || 0 }} 题</text>
            </view>
            <view class="form-start">开始评估 ›</view>
          </view>
        </view>
        <view v-else class="empty">暂无可用评估量表</view>
      </view>
    </template>

    <!-- 我的记录 -->
    <template v-else>
      <view class="records-view">
        <view v-if="recordsLoading && !records.length" class="loading">加载中...</view>
        <view v-else-if="records.length" class="records-list">
          <view v-for="item in records" :key="item.id" class="record-item">
            <view class="record-name">{{ item.form_name || '健康评估' }}</view>
            <view class="record-row">
              <text class="record-score">得分 {{ item.total_score ?? '--' }}</text>
              <text class="record-level" :class="getLevelClass(item.level)">{{ item.level || '未分级' }}</text>
            </view>
            <view class="record-desc" v-if="item.conclusion">{{ item.conclusion }}</view>
            <view class="record-time">{{ formatDateTime(item.created_at) }}</view>
          </view>
          <view v-if="recordsNoMore && records.length" class="no-more">没有更多了</view>
        </view>
        <view v-else class="empty">暂无评估记录</view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onLoad, onReachBottom } from '@dcloudio/uni-app'
import {
  createUserAssessment,
  getUserAssessmentForms,
  getUserAssessments
} from '../../api/health'
import type { AssessmentForm, HealthAssessment } from '../../types'
import { useAuth } from '../../utils/useAuth'

type Mode = 'assess' | 'records'

const mode = ref<Mode>('assess')

// 当前成员档案 ID（可选，用于按档案隔离评估记录）
const recordId = ref<number | undefined>(undefined)

// 量表列表
const forms = ref<AssessmentForm[]>([])
const formsLoading = ref(false)

// 答题状态
const currentForm = ref<AssessmentForm | null>(null)
const answers = reactive<Record<string, string>>({})
const symptomDesc = ref('')
const submitting = ref(false)

// 评估结果
const result = ref<HealthAssessment | null>(null)

// 评估记录
const records = ref<HealthAssessment[]>([])
const recordsLoading = ref(false)
const recordsPage = ref(1)
const recordsNoMore = ref(false)
const pageSize = 10

onLoad(async (options: any) => {
  mode.value = options?.mode === 'records' ? 'records' : 'assess'
  recordId.value = Number(options?.record_id) || undefined

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  loadForms()
  if (mode.value === 'records') {
    loadRecords(true)
  }
})

function switchMode(target: Mode) {
  if (mode.value === target) return
  mode.value = target
  if (target === 'records') {
    loadRecords(true)
  } else {
    loadForms()
  }
}

async function loadForms() {
  if (formsLoading.value) return
  formsLoading.value = true
  try {
    forms.value = await getUserAssessmentForms()
  } catch (error) {
    console.error('加载评估量表失败:', error)
  } finally {
    formsLoading.value = false
  }
}

function startForm(form: AssessmentForm) {
  currentForm.value = form
  Object.keys(answers).forEach(key => delete answers[key])
  symptomDesc.value = ''
  result.value = null
}

function backToList() {
  currentForm.value = null
  result.value = null
  Object.keys(answers).forEach(key => delete answers[key])
  symptomDesc.value = ''
}

function selectOption(questionKey: string, label: string) {
  answers[questionKey] = label
}

async function handleSubmit() {
  if (submitting.value || !currentForm.value) return

  const questions = currentForm.value.questions || []
  const unanswered = questions.filter(question => !answers[question.key])
  if (unanswered.length) {
    uni.showToast({ title: `请完成第 ${questions.indexOf(unanswered[0]) + 1} 题`, icon: 'none' })
    return
  }

  uni.showModal({
    title: '确认提交',
    content: '提交后将根据本次回答生成评估结果，确定提交吗？',
    success: async (res) => {
      if (!res.confirm || !currentForm.value) return
      try {
        submitting.value = true
        const assessment = await createUserAssessment({
          form_id: currentForm.value.id,
          record_id: recordId.value,
          answers: { ...answers },
          symptom_desc: symptomDesc.value.trim() || undefined
        })
        result.value = assessment
        currentForm.value = null
      } catch (error: any) {
        uni.showToast({ title: error.message || '提交失败', icon: 'none' })
      } finally {
        submitting.value = false
      }
    }
  })
}

async function loadRecords(reset = false) {
  if (recordsLoading.value) return
  if (reset) {
    recordsPage.value = 1
    recordsNoMore.value = false
    records.value = []
  } else if (recordsNoMore.value) {
    return
  }

  recordsLoading.value = true
  try {
    const res = await getUserAssessments({ page: recordsPage.value, page_size: pageSize, record_id: recordId.value })
    const list = res?.list || []
    if (reset) {
      records.value = list
    } else {
      records.value.push(...list)
    }
    if (list.length < pageSize) {
      recordsNoMore.value = true
    } else {
      recordsPage.value++
    }
  } catch (error) {
    console.error('加载评估记录失败:', error)
  } finally {
    recordsLoading.value = false
  }
}

onReachBottom(() => {
  if (mode.value === 'records') {
    loadRecords(false)
  }
})

function goHome() {
  uni.reLaunch({ url: '/pages/store/my-health' })
}

function getLevelClass(level?: string): string {
  if (!level) return ''
  if (level.includes('高')) return 'high'
  if (level.includes('中')) return 'medium'
  return 'low'
}

function formatDateTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
</script>

<style scoped>
.assessment-page {
  min-height: 100vh;
  background: #f5f5f5;
}

/* 顶部 tab */
.mode-tabs {
  display: flex;
  background: #ffffff;
  padding: 0 24rpx;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tab-item {
  flex: 1;
  text-align: center;
  font-size: 30rpx;
  color: #666666;
  padding: 24rpx 0;
  position: relative;
}

.tab-item.active {
  color: #007AFF;
  font-weight: 600;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 4rpx;
  background: #007AFF;
  border-radius: 2rpx;
}

/* 量表列表 */
.form-list-view,
.records-view {
  padding: 24rpx;
}

.form-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.form-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.form-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 10rpx;
}

.form-desc {
  font-size: 24rpx;
  color: #666666;
  line-height: 1.5;
  margin-bottom: 16rpx;
}

.form-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.form-dimension {
  font-size: 22rpx;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
}

.form-count {
  font-size: 22rpx;
  color: #999999;
}

.form-start {
  height: 72rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 答题屏 */
.answer-view {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 96rpx - env(safe-area-inset-bottom));
}

.answer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 24rpx;
  background: #ffffff;
}

.answer-title {
  flex: 1;
  min-width: 0;
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-right: 20rpx;
}

.answer-back {
  font-size: 26rpx;
  color: #007AFF;
  flex-shrink: 0;
}

.answer-scroll {
  flex: 1;
  padding: 24rpx;
  box-sizing: border-box;
}

.question-block {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 20rpx;
}

.question-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.5;
  margin-bottom: 20rpx;
}

.option-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.option-item {
  padding: 24rpx 28rpx;
  border-radius: 14rpx;
  background: #f8f9fa;
  color: #333333;
  font-size: 28rpx;
  border: 2rpx solid transparent;
}

.option-item.active {
  background: rgba(0, 122, 255, 0.1);
  border-color: #007AFF;
  color: #007AFF;
  font-weight: 600;
}

.symptom-block {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.symptom-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 180rpx;
  background: #f8f9fa;
  border-radius: 14rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  color: #1a1a1a;
  line-height: 1.5;
}

.answer-bottom-placeholder {
  height: 40rpx;
}

/* 提交栏 */
.submit-bar {
  background: #ffffff;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.submit-btn {
  height: 88rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-btn.disabled {
  opacity: 0.72;
}

/* 结果展示 */
.result-view {
  padding: 48rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.result-icon {
  font-size: 96rpx;
  margin-bottom: 20rpx;
}

.result-score {
  font-size: 88rpx;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1;
}

.result-score-unit {
  font-size: 32rpx;
  font-weight: 400;
  color: #999999;
  margin-left: 8rpx;
}

.result-level {
  margin-top: 20rpx;
  font-size: 30rpx;
  font-weight: 600;
  padding: 8rpx 28rpx;
  border-radius: 999rpx;
}

.result-level.high {
  background: #fff1f0;
  color: #ff4d4f;
}

.result-level.medium {
  background: #fff7e6;
  color: #fa8c16;
}

.result-level.low {
  background: #f6ffed;
  color: #52c41a;
}

.result-form-name {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #999999;
}

.result-conclusion {
  margin-top: 28rpx;
  width: 100%;
  box-sizing: border-box;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  font-size: 28rpx;
  color: #333333;
  line-height: 1.6;
  text-align: left;
}

.result-suggestions {
  margin-top: 24rpx;
  width: 100%;
  box-sizing: border-box;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.result-suggest-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.suggestion-item {
  font-size: 26rpx;
  color: #666666;
  line-height: 1.6;
  margin-bottom: 10rpx;
}

.result-actions {
  margin-top: 40rpx;
  width: 100%;
  box-sizing: border-box;
  display: flex;
  gap: 20rpx;
}

.result-btn {
  flex: 1;
  height: 84rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.result-btn.secondary {
  background: #ffffff;
  color: #007AFF;
  border: 2rpx solid #007AFF;
}

/* 我的记录 */
.records-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.record-item {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.record-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12rpx;
}

.record-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 10rpx;
}

.record-score {
  font-size: 26rpx;
  color: #666666;
}

.record-level {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
}

.record-level.high {
  background: #fff1f0;
  color: #ff4d4f;
}

.record-level.medium {
  background: #fff7e6;
  color: #fa8c16;
}

.record-level.low {
  background: #f6ffed;
  color: #52c41a;
}

.record-desc {
  font-size: 24rpx;
  color: #666666;
  line-height: 1.5;
  margin-bottom: 10rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.record-time {
  font-size: 22rpx;
  color: #999999;
}

.loading,
.no-more,
.empty {
  text-align: center;
  padding: 40rpx 0;
  font-size: 26rpx;
  color: #999999;
}
</style>
