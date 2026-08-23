<template>
  <view class="follow-ups-page">
    <!-- 顶部说明 -->
    <view class="intro-banner">
      <view class="intro-title">🗓️ 我的康复随访</view>
      <view class="intro-desc">查看服务人员为您安排的随访任务与随访结果</view>
    </view>

    <!-- 列表 -->
    <view v-if="loading && !list.length" class="loading">加载中...</view>
    <view v-else-if="list.length" class="task-list">
      <view v-for="item in list" :key="item.id" class="task-card" @click="openDetail(item)">
        <view class="card-header">
          <view class="card-title">{{ getTaskTypeText(item.task_type) }}</view>
          <view class="card-status" :class="getStatusClass(item.status)">{{ getStatusText(item.status) }}</view>
        </view>

        <view class="card-row" v-if="item.plan_follow_time">
          <text class="row-label">随访时间</text>
          <text class="row-value">{{ formatTime(item.plan_follow_time) }}</text>
        </view>
        <view class="card-row" v-if="item.source_type !== undefined">
          <text class="row-label">任务来源</text>
          <text class="row-value">{{ getSourceTypeText(item.source_type) }}</text>
        </view>
        <view class="card-row" v-if="item.contact_method !== undefined">
          <text class="row-label">随访方式</text>
          <text class="row-value">{{ getContactMethodText(item.contact_method) }}</text>
        </view>

        <view class="card-footer">
          <view class="created-time" v-if="item.created_at">创建于 {{ formatDate(item.created_at) }}</view>
          <view class="go-detail">{{ item.status === 1 ? '查看结果' : '查看详情' }} ›</view>
        </view>
      </view>

      <view v-if="noMore && list.length" class="no-more">没有更多了</view>
      <view v-if="loading && list.length" class="no-more">加载中...</view>
    </view>
    <view v-else class="empty">
      <view class="empty-icon">🗓️</view>
      <view class="empty-text">暂无随访任务</view>
      <view class="empty-desc">完成服务或评估后，服务人员将为您安排康复随访</view>
    </view>

    <!-- 详情弹层 -->
    <view v-if="currentTask" class="popup-mask" @click="closeDetail">
      <view class="popup-panel" @click.stop>
        <view class="popup-header">
          <view class="popup-title">{{ getTaskTypeText(currentTask.task_type) }}</view>
          <view class="popup-close" @click="closeDetail">✕</view>
        </view>

        <view class="popup-row">
          <text class="popup-label">任务状态</text>
          <text class="popup-value">{{ getStatusText(currentTask.status) }}</text>
        </view>
        <view class="popup-row">
          <text class="popup-label">随访时间</text>
          <text class="popup-value">{{ formatTime(currentTask.plan_follow_time) }}</text>
        </view>
        <view class="popup-row" v-if="currentTask.source_type !== undefined">
          <text class="popup-label">任务来源</text>
          <text class="popup-value">{{ getSourceTypeText(currentTask.source_type) }}</text>
        </view>
        <view class="popup-row">
          <text class="popup-label">随访方式</text>
          <text class="popup-value">{{ getContactMethodText(currentTask.contact_method) }}</text>
        </view>

        <!-- 已完成任务的随访结果 -->
        <template v-if="currentTask.status === 1">
          <view class="popup-row">
            <text class="popup-label">完成时间</text>
            <text class="popup-value">{{ formatTime(currentTask.completed_at) }}</text>
          </view>
          <view class="popup-row" v-if="currentTask.result?.content">
            <text class="popup-label">随访内容</text>
            <text class="popup-value">{{ currentTask.result?.content }}</text>
          </view>
          <view class="popup-row" v-if="currentTask.result?.satisfaction !== undefined">
            <text class="popup-label">满意度</text>
            <text class="popup-value">{{ getSatisfactionText(currentTask.result?.satisfaction) }}</text>
          </view>
          <view class="popup-row" v-if="currentTask.result?.remark">
            <text class="popup-label">随访备注</text>
            <text class="popup-value">{{ currentTask.result?.remark }}</text>
          </view>
          <view class="popup-row" v-if="currentTask.result?.education_article_ids?.length">
            <text class="popup-label">宣教推送</text>
            <text class="popup-value">已推送 {{ currentTask.result?.education_article_ids?.length }} 篇健康宣教文章</text>
          </view>
        </template>

        <view class="popup-row" v-if="currentTask.remark">
          <text class="popup-label">任务备注</text>
          <text class="popup-value">{{ currentTask.remark }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getUserFollowUps } from '../../api/health'
import type { FollowUpTask } from '../../types'
import { useAuth } from '../../utils/useAuth'

const list = ref<FollowUpTask[]>([])
const loading = ref(false)
const page = ref(1)
const noMore = ref(false)
const pageSize = 10
const currentTask = ref<FollowUpTask | null>(null)

onLoad(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  loadList(true)
})

/**
 * 任务类型文案：1康复随访 2租后回访 3慢病随访 4评估回访
 */
function getTaskTypeText(taskType?: number): string {
  return { 1: '康复随访', 2: '租后回访', 3: '慢病随访', 4: '评估回访' }[Number(taskType ?? 0)] || '随访任务'
}

/**
 * 任务来源文案：1服务完成 2租赁归还 3评估完成 4手动
 */
function getSourceTypeText(sourceType?: number): string {
  return { 1: '服务完成', 2: '租赁归还', 3: '评估完成', 4: '手动创建' }[Number(sourceType ?? 0)] || '系统创建'
}

/**
 * 任务状态文案：0待执行 1已完成 2已跳过
 */
function getStatusText(status?: number): string {
  return { 0: '待执行', 1: '已完成', 2: '已跳过' }[Number(status ?? 0)] || '待执行'
}

function getStatusClass(status?: number): string {
  const map: Record<number, string> = { 0: 'pending', 1: 'done', 2: 'skipped' }
  return map[Number(status ?? 0)] || 'pending'
}

/**
 * 随访方式文案：1电话 2上门 3微信
 */
function getContactMethodText(contactMethod?: number): string {
  return { 1: '电话', 2: '上门', 3: '微信' }[Number(contactMethod ?? 0)] || '--'
}

/**
 * 满意度文案：1-5 档展示星星，其他值展示原始分数。
 */
function getSatisfactionText(satisfaction?: number): string {
  if (satisfaction === undefined || satisfaction === null) return '--'
  if (satisfaction >= 1 && satisfaction <= 5) {
    return '★'.repeat(satisfaction) + '☆'.repeat(5 - satisfaction)
  }
  return `${satisfaction} 分`
}

function formatDate(date?: string): string {
  if (!date) return ''
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}`
}

function formatTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function openDetail(item: FollowUpTask) {
  currentTask.value = item
}

function closeDetail() {
  currentTask.value = null
}

async function loadList(reset = false) {
  if (loading.value) return
  if (!reset && noMore.value) return

  loading.value = true
  try {
    const res = await getUserFollowUps({ page: page.value, page_size: pageSize })
    const data = res?.list || []
    if (reset) {
      list.value = data
    } else {
      list.value.push(...data)
    }
    if (data.length < pageSize) {
      noMore.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载随访任务失败:', error)
  } finally {
    loading.value = false
  }
}

onReachBottom(() => {
  loadList(false)
})

onPullDownRefresh(async () => {
  page.value = 1
  noMore.value = false
  await loadList(true)
  uni.stopPullDownRefresh()
})
</script>

<style scoped>
.follow-ups-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

/* 顶部说明 */
.intro-banner {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.intro-title {
  font-size: 34rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.intro-desc {
  font-size: 24rpx;
  opacity: 0.85;
}

/* 列表 */
.task-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.task-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.card-title {
  flex: 1;
  min-width: 0;
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-right: 20rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-status {
  flex-shrink: 0;
  font-size: 22rpx;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
}

.card-status.pending {
  background: #fff7e6;
  color: #fa8c16;
}

.card-status.done {
  background: #f6ffed;
  color: #52c41a;
}

.card-status.skipped {
  background: #f0f0f0;
  color: #999999;
}

.card-row {
  display: flex;
  margin-bottom: 16rpx;
  line-height: 1.5;
}

.row-label {
  width: 140rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  color: #999999;
}

.row-value {
  flex: 1;
  min-width: 0;
  font-size: 26rpx;
  color: #1a1a1a;
  word-break: break-all;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8rpx;
}

.created-time {
  font-size: 22rpx;
  color: #999999;
}

.go-detail {
  font-size: 24rpx;
  color: #007AFF;
}

.loading,
.no-more,
.empty {
  text-align: center;
  padding: 40rpx 0;
  font-size: 26rpx;
  color: #999999;
}

.empty-icon {
  font-size: 80rpx;
  margin-bottom: 16rpx;
}

.empty-text {
  font-size: 30rpx;
  color: #666666;
  margin-bottom: 10rpx;
}

.empty-desc {
  font-size: 24rpx;
  color: #999999;
}

/* 详情弹层 */
.popup-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}

.popup-panel {
  width: 620rpx;
  max-height: 80vh;
  overflow-y: auto;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  box-sizing: border-box;
}

.popup-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.popup-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.popup-close {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #999999;
  background: #f5f5f5;
  border-radius: 50%;
}

.popup-row {
  display: flex;
  margin-bottom: 20rpx;
  line-height: 1.6;
}

.popup-label {
  width: 160rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  color: #999999;
}

.popup-value {
  flex: 1;
  min-width: 0;
  font-size: 26rpx;
  color: #1a1a1a;
  word-break: break-all;
}
</style>
