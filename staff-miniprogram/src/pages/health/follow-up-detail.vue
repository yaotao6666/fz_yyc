<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { FollowUpTask, EducationArticle } from '@/types'
import { formatDate } from '@/utils/format'

const taskId = ref<string>('')
const loading = ref(false)
const task = ref<FollowUpTask | null>(null)
const articles = ref<EducationArticle[]>([])

// 随访表单
const contactMethod = ref(1)
const content = ref('')
const selectedArticleIds = ref<number[]>([])
const satisfaction = ref(5)
const remark = ref('')
const submitting = ref(false)

// 仅待执行任务可执行随访
const isPending = computed(() => task.value?.status === 0)

const taskTypeText = (type?: number) => {
  const map: Record<number, string> = { 1: '康复随访', 2: '租后回访', 3: '慢病随访', 4: '评估回访' }
  return type == null ? '' : (map[type] || '其他')
}

const sourceTypeText = (type?: number) => {
  const map: Record<number, string> = { 1: '服务完成', 2: '租赁归还', 3: '评估完成', 4: '手动' }
  return type == null ? '' : (map[type] || '其他')
}

const statusText = (status?: number) => {
  const map: Record<number, string> = { 0: '待执行', 1: '已完成', 2: '已跳过' }
  return status == null ? '' : (map[status] || '未知')
}

function statusClass(status?: number) {
  const map: Record<number, string> = { 0: 'status-pending', 1: 'status-done', 2: 'status-skip' }
  return status == null ? '' : (map[status] || '')
}

// 随访方式选项
const contactMethodOptions = [
  { value: 1, label: '电话' },
  { value: 2, label: '上门' },
  { value: 3, label: '微信' }
]

const contactMethodText = (method?: number) => {
  const map: Record<number, string> = { 1: '电话', 2: '上门', 3: '微信' }
  return method == null ? '' : (map[method] || '未知')
}

async function loadDetail() {
  if (!taskId.value) return
  loading.value = true
  try {
    task.value = await staffHealthApi.getFollowUpTaskDetail(taskId.value)
    // 若为已完成任务，将历史结果回填到表单（便于查看）
    const r = task.value?.result
    if (r) {
      if (r.contact_method) contactMethod.value = r.contact_method
      if (r.content) content.value = r.content
      if (r.education_article_ids?.length) selectedArticleIds.value = r.education_article_ids
      if (r.satisfaction) satisfaction.value = r.satisfaction
      if (r.remark) remark.value = r.remark
    }
  } catch (e) {
    console.error('[FollowUpDetail] loadDetail error', e)
  } finally {
    loading.value = false
  }
}

async function loadArticles() {
  try {
    articles.value = await staffHealthApi.getEducationArticles()
  } catch (e) {
    console.error('[FollowUpDetail] loadArticles error', e)
  }
}

function toggleArticle(id: number) {
  const idx = selectedArticleIds.value.indexOf(id)
  if (idx >= 0) selectedArticleIds.value.splice(idx, 1)
  else selectedArticleIds.value.push(id)
}

async function handleComplete() {
  if (!task.value) return
  if (!content.value.trim()) {
    uni.showToast({ title: '请填写随访内容', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    await staffHealthApi.completeFollowUpTask(task.value.id, {
      contact_method: contactMethod.value,
      content: content.value.trim(),
      education_article_ids: selectedArticleIds.value,
      satisfaction: satisfaction.value,
      remark: remark.value.trim() || undefined
    })
    uni.showToast({ title: '随访已完成', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (e) {
    console.error('[FollowUpDetail] handleComplete error', e)
  } finally {
    submitting.value = false
  }
}

function handleSkip() {
  if (!task.value) return
  uni.showModal({
    title: '确认跳过',
    content: '跳过后将不再执行该随访任务，确定跳过吗？',
    success: async (res) => {
      if (!res.confirm || !task.value) return
      submitting.value = true
      try {
        await staffHealthApi.skipFollowUpTask(task.value.id)
        uni.showToast({ title: '已跳过', icon: 'success' })
        setTimeout(() => uni.navigateBack(), 800)
      } catch (e) {
        console.error('[FollowUpDetail] handleSkip error', e)
      } finally {
        submitting.value = false
      }
    }
  })
}

onLoad((options: any) => {
  taskId.value = options?.id || ''
  if (!taskId.value) {
    uni.showToast({ title: '缺少任务ID', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  loadDetail()
  loadArticles()
})
</script>

<template>
  <view class="page">
    <view v-if="!loading && !task" class="empty card">
      <text class="empty-title">任务不存在</text>
      <text class="empty-desc">请返回重试</text>
    </view>

    <template v-else-if="task">
      <view class="container">
        <!-- 任务信息 -->
        <view class="card section">
          <view class="task-header">
            <view class="task-title-wrap">
              <text class="task-name">{{ taskTypeText(task.task_type) || '随访任务' }}</text>
              <text v-if="task.source_type" class="task-source">{{ sourceTypeText(task.source_type) }}</text>
            </view>
            <text v-if="task.status != null" :class="['task-status', statusClass(task.status)]">
              {{ statusText(task.status) }}
            </text>
          </view>
          <view class="info-row" v-if="task.plan_follow_time">
            <text class="info-label">计划随访</text>
            <text class="info-value">{{ formatDate(task.plan_follow_time, 'YYYY-MM-DD HH:mm') }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">创建时间</text>
            <text class="info-value">{{ formatDate(task.created_at, 'YYYY-MM-DD HH:mm') }}</text>
          </view>
          <view class="info-row" v-if="task.completed_at">
            <text class="info-label">处理时间</text>
            <text class="info-value">{{ formatDate(task.completed_at, 'YYYY-MM-DD HH:mm') }}</text>
          </view>
          <view class="info-row" v-if="task.remark">
            <text class="info-label">任务备注</text>
            <text class="info-value">{{ task.remark }}</text>
          </view>
        </view>

        <!-- 客户信息 -->
        <view class="card section">
          <view class="section-title">客户信息</view>
          <view v-if="task.user" class="customer-row">
            <text class="customer-name">{{ task.user.nickname || '未填写姓名' }}</text>
            <text v-if="task.user.phone" class="customer-phone">{{ task.user.phone }}</text>
          </view>
          <view v-else class="no-data">暂无客户信息</view>
        </view>

        <!-- 待执行：开始随访表单 -->
        <template v-if="isPending">
          <view class="card section">
            <view class="section-title">随访方式</view>
            <view class="method-list">
              <view
                v-for="opt in contactMethodOptions"
                :key="opt.value"
                :class="['method-item', { active: contactMethod === opt.value }]"
                @tap="contactMethod = opt.value"
              >
                <view :class="['method-radio', { checked: contactMethod === opt.value }]">
                  <text v-if="contactMethod === opt.value" class="radio-dot"></text>
                </view>
                <text class="method-label">{{ opt.label }}</text>
              </view>
            </view>
          </view>

          <view class="card section">
            <view class="section-title">随访内容</view>
            <textarea
              v-model="content"
              class="desc-input"
              placeholder="请填写本次随访的沟通内容、客户反馈等情况"
              maxlength="500"
            />
          </view>

          <view class="card section">
            <view class="section-title">
              <text>宣教文章</text>
              <text v-if="articles.length" class="record-total">已选 {{ selectedArticleIds.length }} 篇</text>
            </view>
            <view v-if="articles.length === 0" class="no-data">暂无可选的宣教文章</view>
            <view v-else>
              <view
                v-for="article in articles"
                :key="article.id"
                class="article-item"
                @tap="toggleArticle(article.id)"
              >
                <view :class="['article-check', { checked: selectedArticleIds.includes(article.id) }]">
                  <text v-if="selectedArticleIds.includes(article.id)" class="check-mark">✓</text>
                </view>
                <view class="article-info">
                  <text class="article-title">{{ article.title }}</text>
                  <text v-if="article.category" class="article-category">{{ article.category }}</text>
                </view>
              </view>
            </view>
          </view>

          <view class="card section">
            <view class="section-title">
              <text>满意度</text>
              <text class="satisfaction-num">{{ satisfaction }} 分</text>
            </view>
            <view class="stars">
              <text
                v-for="star in 5"
                :key="star"
                :class="['star', { active: star <= satisfaction }]"
                @tap="satisfaction = star"
              >★</text>
            </view>
          </view>

          <view class="card section">
            <view class="section-title">备注（选填）</view>
            <textarea
              v-model="remark"
              class="desc-input"
              placeholder="其他需要记录的信息"
              maxlength="500"
            />
          </view>
        </template>

        <!-- 已处理：展示随访结果 -->
        <view v-else class="card section">
          <view class="section-title">
            <text>随访结果</text>
            <text v-if="task.status === 2" class="skip-tip">该任务已被跳过</text>
          </view>
          <template v-if="task.status === 1 && task.result">
            <view class="info-row">
              <text class="info-label">随访方式</text>
              <text class="info-value">{{ contactMethodText(task.result.contact_method) || '未记录' }}</text>
            </view>
            <view class="info-row" v-if="task.result.content">
              <text class="info-label">随访内容</text>
              <text class="info-value">{{ task.result.content }}</text>
            </view>
            <view class="info-row" v-if="task.result.education_article_ids?.length">
              <text class="info-label">宣教文章</text>
              <text class="info-value">{{ task.result.education_article_ids.length }} 篇</text>
            </view>
            <view class="info-row" v-if="task.result.satisfaction">
              <text class="info-label">满意度</text>
              <text class="info-value">{{ task.result.satisfaction }} 分</text>
            </view>
            <view class="info-row" v-if="task.result.remark">
              <text class="info-label">备注</text>
              <text class="info-value">{{ task.result.remark }}</text>
            </view>
          </template>
          <view v-else-if="task.status === 1" class="no-data">暂无随访结果</view>
        </view>
      </view>

      <!-- 底部操作栏：仅待执行任务可操作 -->
      <view class="footer-bar" v-if="isPending">
        <button class="btn-skip" :disabled="submitting" @tap="handleSkip">跳过</button>
        <button class="btn-complete" :disabled="submitting" @tap="handleComplete">
          {{ submitting ? '提交中...' : '完成随访' }}
        </button>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped>
.page { min-height: 100vh; }
.container { padding-bottom: 140rpx; }
.empty {
  text-align: center;
  padding: 200rpx 32rpx !important;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}

.section {
  margin-bottom: 20rpx;
  padding: 28rpx 24rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 20rpx;
  padding-bottom: 16rpx;
  border-bottom: 2rpx solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  .record-total { font-size: 24rpx; color: #999; font-weight: normal; }
}
.no-data { font-size: 26rpx; color: #999; padding: 12rpx 0; }

/* 任务信息 */
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.task-title-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  .task-name { font-size: 34rpx; font-weight: 600; color: #333; }
  .task-source {
    font-size: 22rpx;
    color: var(--info-color);
    background: #e6f4ff;
    padding: 4rpx 16rpx;
    border-radius: 8rpx;
    flex-shrink: 0;
  }
}
.task-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  flex-shrink: 0;
  &.status-pending { background: #fff7e6; color: #fa8c16; }
  &.status-done { background: #e8f8ee; color: #07c160; }
  &.status-skip { background: #f5f5f5; color: #999; }
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 10rpx 0;
  .info-label { width: 160rpx; font-size: 26rpx; color: #999; flex-shrink: 0; }
  .info-value { flex: 1; font-size: 26rpx; color: #333; text-align: right; line-height: 1.5; }
}

/* 客户信息 */
.customer-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 8rpx 0;
  .customer-name { font-size: 30rpx; font-weight: 500; color: #333; }
  .customer-phone { font-size: 26rpx; color: #666; }
}

/* 随访方式 */
.method-list {
  display: flex;
  gap: 20rpx;
}
.method-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 20rpx 0;
  background: var(--bg-color);
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  &.active {
    border-color: var(--primary-color);
    background: rgba(81, 117, 40, 0.06);
  }
  .method-radio {
    width: 36rpx; height: 36rpx;
    border-radius: 50%;
    border: 2rpx solid #ccc;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    &.checked { border-color: var(--primary-color); }
    .radio-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: var(--primary-color); }
  }
  .method-label { font-size: 28rpx; color: #333; }
}

/* 描述输入 */
.desc-input {
  width: 100%;
  height: 180rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}

/* 宣教文章 */
.article-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
  .article-check {
    width: 36rpx; height: 36rpx;
    border-radius: 8rpx;
    border: 2rpx solid #ccc;
    margin-right: 16rpx;
    flex-shrink: 0;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    &.checked { border-color: var(--primary-color); background: var(--primary-color); }
    .check-mark { font-size: 24rpx; color: #fff; }
  }
  .article-info {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 16rpx;
    .article-title { font-size: 28rpx; color: #333; }
    .article-category {
      font-size: 22rpx;
      color: var(--warning-color);
      background: #fff7e6;
      padding: 2rpx 12rpx;
      border-radius: 8rpx;
      flex-shrink: 0;
    }
  }
}

/* 满意度 */
.satisfaction-num { font-size: 24rpx; color: var(--warning-color); font-weight: normal; }
.stars {
  display: flex;
  gap: 20rpx;
  justify-content: center;
  padding: 16rpx 0;
  .star {
    font-size: 64rpx;
    color: #e0e0e0;
    line-height: 1;
    &.active { color: var(--warning-color); }
  }
}

.skip-tip {
  font-size: 24rpx;
  color: #999;
  font-weight: normal;
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
.btn-skip {
  flex: 1;
  font-size: 30rpx;
  color: #999;
  background: #fff;
  border: 2rpx solid #ddd;
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
}
.btn-complete {
  flex: 1.5;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &[disabled] { opacity: 0.5; }
}
</style>
