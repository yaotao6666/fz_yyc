<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { FollowUpTask } from '@/types'
import { formatDate } from '@/utils/format'

const list = ref<FollowUpTask[]>([])
const total = ref(0)
const loading = ref(false)
// 分页状态
const page = ref(1)
const pageSize = 20
// 是否已加载完所有数据
const finished = ref(false)
// 当前状态筛选（'' 全部，0 待执行，1 已完成，2 已跳过）
const activeStatus = ref<number | ''>('')

const tabs: { label: string; value: number | '' }[] = [
  { label: '全部', value: '' },
  { label: '待执行', value: 0 },
  { label: '已完成', value: 1 },
  { label: '已跳过', value: 2 }
]

// 任务类型文案
function taskTypeText(type?: number) {
  const map: Record<number, string> = { 1: '康复随访', 2: '租后回访', 3: '慢病随访', 4: '评估回访' }
  return type == null ? '' : (map[type] || '其他')
}

// 来源文案
function sourceTypeText(type?: number) {
  const map: Record<number, string> = { 1: '服务完成', 2: '租赁归还', 3: '评估完成', 4: '手动' }
  return type == null ? '' : (map[type] || '其他')
}

// 状态文案与样式
function statusText(status?: number) {
  const map: Record<number, string> = { 0: '待执行', 1: '已完成', 2: '已跳过' }
  return status == null ? '' : (map[status] || '未知')
}

function statusClass(status?: number) {
  const map: Record<number, string> = { 0: 'status-pending', 1: 'status-done', 2: 'status-skip' }
  return status == null ? '' : (map[status] || '')
}

async function loadTasks(reset = false) {
  if (loading.value || (!reset && finished.value)) return
  loading.value = true
  const targetPage = reset ? 1 : page.value
  try {
    const res: any = await staffHealthApi.getMyFollowUpTasks({
      status: activeStatus.value === '' ? undefined : activeStatus.value,
      page: targetPage,
      page_size: pageSize
    })
    const items = (res?.list || []) as FollowUpTask[]
    total.value = res?.total || 0
    if (reset) {
      list.value = items
      page.value = 1
    } else {
      // 追加下一页数据，避免重复
      list.value = list.value.concat(items)
    }
    finished.value = list.value.length >= total.value
  } catch (e) {
    console.error('[FollowUps] loadTasks error', e)
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function switchTab(value: number | '') {
  if (activeStatus.value === value) return
  activeStatus.value = value
  loadTasks(true)
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/health/follow-up-detail?id=${id}` })
}

onLoad((options: any) => {
  // URL query 可选 status 预筛
  const s = Number(options?.status)
  if ([0, 1, 2].includes(s)) activeStatus.value = s as number
  loadTasks(true)
})

// 从详情页返回时刷新，保证任务状态最新
onShow(() => {
  if (list.value.length) loadTasks(true)
})

onPullDownRefresh(() => {
  loadTasks(true)
})

onReachBottom(() => {
  if (finished.value) return
  page.value += 1
  loadTasks(false)
})
</script>

<template>
  <view class="page">
    <!-- 状态筛选 tab -->
    <view class="tabs">
      <view
        v-for="tab in tabs"
        :key="tab.label"
        :class="['tab-item', { active: activeStatus === tab.value }]"
        @tap="switchTab(tab.value)"
      >
        {{ tab.label }}
      </view>
    </view>

    <view v-if="list.length === 0 && !loading" class="empty card">
      <text class="empty-title">暂无随访任务</text>
      <text class="empty-desc">指派给您的随访任务将显示在这里，下拉可刷新</text>
    </view>

    <view
      v-for="task in list"
      :key="task.id"
      class="card task-card"
      @tap="goDetail(task.id)"
    >
      <view class="task-header">
        <view class="task-title-wrap">
          <text class="task-name">{{ taskTypeText(task.task_type) || '随访任务' }}</text>
          <text v-if="task.source_type" class="task-source">{{ sourceTypeText(task.source_type) }}</text>
        </view>
        <text v-if="task.status != null" :class="['task-status', statusClass(task.status)]">
          {{ statusText(task.status) }}
        </text>
      </view>

      <view v-if="task.user" class="task-customer">
        <text class="customer-name">{{ task.user.nickname || '未填写姓名' }}</text>
        <text v-if="task.user.phone" class="customer-phone">{{ task.user.phone }}</text>
      </view>
      <view v-else class="task-customer">
        <text class="customer-name">客户信息未知</text>
      </view>

      <view class="task-meta">
        <view class="meta-row">
          <text class="meta-label">计划随访</text>
          <text class="meta-value">{{ task.plan_follow_time ? formatDate(task.plan_follow_time, 'YYYY-MM-DD HH:mm') : '未设置' }}</text>
        </view>
        <view class="meta-row" v-if="task.remark">
          <text class="meta-label">备注</text>
          <text class="meta-value">{{ task.remark }}</text>
        </view>
      </view>

      <view class="task-footer">
        <text class="task-time">创建于 {{ formatDate(task.created_at, 'YYYY-MM-DD') }}</text>
        <text v-if="task.completed_at" class="task-time">{{ task.status === 1 ? '完成于' : '处理于' }} {{ formatDate(task.completed_at, 'YYYY-MM-DD') }}</text>
      </view>
    </view>

    <view v-if="loading" class="loading-wrap">
      <text>加载中...</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  padding: 20rpx 0;
}
.empty {
  text-align: center;
  padding: 120rpx 32rpx !important;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}
.loading-wrap {
  text-align: center;
  padding: 24rpx 0;
  color: #999;
  font-size: 26rpx;
}

/* 状态筛选 tab */
.tabs {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  background: #fff;
  padding: 0 24rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}
.tab-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666;
  padding: 24rpx 0;
  position: relative;
  &.active {
    color: var(--primary-color);
    font-weight: 600;
    &::after {
      content: '';
      position: absolute;
      left: 50%;
      bottom: 0;
      transform: translateX(-50%);
      width: 48rpx;
      height: 6rpx;
      border-radius: 3rpx;
      background: var(--primary-color);
    }
  }
}

/* 任务卡片 */
.task-card {
  padding: 28rpx 24rpx;
  margin-bottom: 20rpx;
}
.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.task-title-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  .task-name { font-size: 32rpx; font-weight: 600; color: #333; }
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

.task-customer {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 12rpx;
  .customer-name { font-size: 28rpx; font-weight: 500; color: #333; }
  .customer-phone { font-size: 26rpx; color: #666; }
}

.task-meta {
  padding: 8rpx 0;
}
.meta-row {
  display: flex;
  padding: 6rpx 0;
  .meta-label { width: 128rpx; font-size: 26rpx; color: #999; flex-shrink: 0; }
  .meta-value { flex: 1; font-size: 26rpx; color: #333; line-height: 1.5; }
}

.task-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
  padding-top: 16rpx;
  border-top: 2rpx solid var(--border-color);
  .task-time { font-size: 24rpx; color: #999; }
}
</style>
