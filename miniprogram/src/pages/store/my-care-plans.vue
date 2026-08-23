<template>
  <view class="care-plans-page">
    <!-- 顶部说明 -->
    <view class="intro-banner">
      <view class="intro-title">📋 我的照护计划</view>
      <view class="intro-desc">查看为您定制的照护方案与上门照护记录</view>
    </view>

    <!-- 列表 -->
    <view v-if="loading && !list.length" class="loading">加载中...</view>
    <view v-else-if="list.length" class="plan-list">
      <view
        v-for="item in list"
        :key="item.id"
        class="plan-card"
        @click="goDetail(item.id)"
      >
        <view class="card-header">
          <view class="card-name">{{ item.name || '未命名计划' }}</view>
          <view class="card-status" :class="getStatusClass(item.status)">{{ getStatusText(item.status) }}</view>
        </view>

        <view class="card-chips">
          <text v-if="item.plan_type !== undefined" class="type-chip">{{ getPlanTypeText(item.plan_type) }}</text>
          <text v-if="item.frequency" class="freq-chip">频次：{{ item.frequency }}</text>
        </view>

        <view class="card-row" v-if="item.start_date || item.end_date">
          <text class="row-label">计划周期</text>
          <text class="row-value">{{ formatPeriod(item.start_date, item.end_date) }}</text>
        </view>

        <view class="card-row" v-if="item.goals">
          <text class="row-label">照护目标</text>
          <text class="row-value goal-text">{{ item.goals }}</text>
        </view>

        <view class="card-footer">
          <view class="visit-count" v-if="item.visit_count !== undefined">
            已上门照护 {{ item.visit_count }} 次
          </view>
          <view class="go-detail">查看详情 ›</view>
        </view>
      </view>

      <view v-if="noMore && list.length" class="no-more">没有更多了</view>
      <view v-if="loading && list.length" class="no-more">加载中...</view>
    </view>
    <view v-else class="empty">
      <view class="empty-icon">📋</view>
      <view class="empty-text">暂无照护计划</view>
      <view class="empty-desc">购买康养服务后，服务人员将为您制定专属照护计划</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getUserCarePlans } from '../../api/health'
import type { CarePlan } from '../../types'
import { useAuth } from '../../utils/useAuth'

const list = ref<CarePlan[]>([])
const loading = ref(false)
const page = ref(1)
const noMore = ref(false)
const pageSize = 10

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
 * 计划类型文案：1生活照料 2基础护理 3康复训练 4综合康养
 */
function getPlanTypeText(planType?: number): string {
  return { 1: '生活照料', 2: '基础护理', 3: '康复训练', 4: '综合康养' }[Number(planType ?? 0)] || '照护计划'
}

/**
 * 计划状态文案：0草稿 1执行中 2已暂停 3已完成
 */
function getStatusText(status?: number): string {
  return { 0: '草稿', 1: '执行中', 2: '已暂停', 3: '已完成' }[Number(status ?? 0)] || '草稿'
}

function getStatusClass(status?: number): string {
  const map: Record<number, string> = { 0: 'draft', 1: 'running', 2: 'paused', 3: 'done' }
  return map[Number(status ?? 0)] || 'draft'
}

function formatDate(date?: string): string {
  if (!date) return ''
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}`
}

function formatPeriod(startDate?: string, endDate?: string): string {
  const start = formatDate(startDate)
  const end = formatDate(endDate)
  if (!start && !end) return ''
  if (start && end) return `${start} 至 ${end}`
  return start || end
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/store/care-plan-detail?id=${id}` })
}

async function loadList(reset = false) {
  if (loading.value) return
  if (!reset && noMore.value) return

  loading.value = true
  try {
    const res = await getUserCarePlans({ page: page.value, page_size: pageSize })
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
    console.error('加载照护计划失败:', error)
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
.care-plans-page {
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
.plan-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.plan-card {
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

.card-name {
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

.card-status.draft {
  background: #f0f0f0;
  color: #999999;
}

.card-status.running {
  background: #e6f0ff;
  color: #007AFF;
}

.card-status.paused {
  background: #fff7e6;
  color: #fa8c16;
}

.card-status.done {
  background: #f6ffed;
  color: #52c41a;
}

/* 类型/频次标签 */
.card-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.type-chip {
  font-size: 22rpx;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

.freq-chip {
  font-size: 22rpx;
  color: #666666;
  background: #f5f5f5;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

/* 信息行 */
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

.goal-text {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

/* 底部信息 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8rpx;
}

.visit-count {
  font-size: 24rpx;
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
</style>
