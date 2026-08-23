<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { CarePlan } from '@/types'
import { formatDate } from '@/utils/format'

const list = ref<CarePlan[]>([])
const total = ref(0)
const loading = ref(false)
// 分页状态
const page = ref(1)
const pageSize = 20
// 是否已加载完所有数据
const finished = ref(false)

// 计划类型文案
function planTypeText(type?: number) {
  const map: Record<number, string> = { 1: '生活照料', 2: '基础护理', 3: '康复训练', 4: '综合康养' }
  return type == null ? '' : (map[type] || '其他')
}

// 计划状态文案
function planStatusText(status?: number) {
  const map: Record<number, string> = { 0: '草稿', 1: '执行中', 2: '已暂停', 3: '已完成' }
  return status == null ? '' : (map[status] || '未知')
}

function planStatusClass(status?: number) {
  const map: Record<number, string> = { 0: 'status-draft', 1: 'status-running', 2: 'status-paused', 3: 'status-done' }
  return status == null ? '' : (map[status] || '')
}

// 计划周期文案（有起止日期则展示区间，否则省略）
function periodText(plan: CarePlan) {
  if (!plan.start_date && !plan.end_date) return ''
  const start = formatDate(plan.start_date || '', 'YYYY-MM-DD')
  const end = formatDate(plan.end_date || '', 'YYYY-MM-DD')
  return start || end ? `${start || '?'} ~ ${end || '长期'}` : ''
}

async function loadPlans(reset = false) {
  if (loading.value || (!reset && finished.value)) return
  loading.value = true
  const targetPage = reset ? 1 : page.value
  try {
    const res: any = await staffHealthApi.getMyCarePlans({ page: targetPage, page_size: pageSize })
    const items = (res?.list || []) as CarePlan[]
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
    console.error('[CarePlans] loadPlans error', e)
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/health/care-plan-detail?id=${id}` })
}

onLoad(() => {
  loadPlans(true)
})

// 从详情页返回时刷新，保证照护次数等数据最新
onShow(() => {
  if (list.value.length) loadPlans(true)
})

onPullDownRefresh(() => {
  loadPlans(true)
})

onReachBottom(() => {
  if (finished.value) return
  page.value += 1
  loadPlans(false)
})
</script>

<template>
  <view class="page">
    <view v-if="list.length === 0 && !loading" class="empty card">
      <text class="empty-title">暂无照护计划</text>
      <text class="empty-desc">您负责的照护计划将显示在这里，下拉可刷新</text>
    </view>

    <view
      v-for="plan in list"
      :key="plan.id"
      class="card plan-card"
      @tap="goDetail(plan.id)"
    >
      <view class="plan-header">
        <view class="plan-title-wrap">
          <text class="plan-name">{{ plan.name }}</text>
          <text v-if="plan.plan_type" class="plan-type">{{ planTypeText(plan.plan_type) }}</text>
        </view>
        <text v-if="plan.status != null" :class="['plan-status', planStatusClass(plan.status)]">
          {{ planStatusText(plan.status) }}
        </text>
      </view>

      <view class="plan-meta">
        <view v-if="periodText(plan)" class="meta-row">
          <text class="meta-label">周期</text>
          <text class="meta-value">{{ periodText(plan) }}</text>
        </view>
        <view v-if="plan.frequency" class="meta-row">
          <text class="meta-label">频次</text>
          <text class="meta-value">{{ plan.frequency }}</text>
        </view>
        <view v-if="plan.goals" class="meta-row">
          <text class="meta-label">目标</text>
          <text class="meta-value goal-text">{{ plan.goals }}</text>
        </view>
      </view>

      <view class="plan-footer">
        <text class="visit-count">照护次数：{{ plan.visit_count || 0 }}</text>
        <text v-if="plan.updated_at" class="plan-time">更新于 {{ formatDate(plan.updated_at, 'YYYY-MM-DD') }}</text>
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

.plan-card {
  padding: 28rpx 24rpx;
  margin-bottom: 20rpx;
}
.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.plan-title-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  .plan-name { font-size: 32rpx; font-weight: 600; color: #333; }
  .plan-type {
    font-size: 22rpx;
    color: var(--primary-color);
    background: rgba(81, 117, 40, 0.1);
    padding: 4rpx 16rpx;
    border-radius: 8rpx;
    flex-shrink: 0;
  }
}
.plan-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  flex-shrink: 0;
  &.status-draft { background: #f5f5f5; color: #999; }
  &.status-running { background: #e8f8ee; color: #07c160; }
  &.status-paused { background: #fff7e6; color: #fa8c16; }
  &.status-done { background: #e6f4ff; color: #1989fa; }
}

.plan-meta {
  padding: 8rpx 0;
}
.meta-row {
  display: flex;
  padding: 6rpx 0;
  .meta-label { width: 96rpx; font-size: 26rpx; color: #999; flex-shrink: 0; }
  .meta-value { flex: 1; font-size: 26rpx; color: #333; line-height: 1.5; }
  .goal-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.plan-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
  padding-top: 16rpx;
  border-top: 2rpx solid var(--border-color);
  .visit-count { font-size: 26rpx; color: var(--primary-color); font-weight: 500; }
  .plan-time { font-size: 24rpx; color: #999; }
}
</style>
