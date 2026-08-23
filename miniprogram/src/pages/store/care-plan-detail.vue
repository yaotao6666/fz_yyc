<template>
  <view class="care-plan-detail-page">
    <view v-if="loading && !plan" class="loading">加载中...</view>

    <template v-else-if="plan">
      <!-- 计划信息 -->
      <view class="plan-header">
        <view class="header-top">
          <view class="plan-name">{{ plan.name || '未命名计划' }}</view>
          <view class="plan-status" :class="getStatusClass(plan.status)">{{ getStatusText(plan.status) }}</view>
        </view>
        <view class="header-chips">
          <text v-if="plan.plan_type !== undefined" class="type-chip">{{ getPlanTypeText(plan.plan_type) }}</text>
          <text v-if="plan.frequency" class="freq-chip">频次：{{ plan.frequency }}</text>
        </view>
        <view class="header-row" v-if="plan.start_date || plan.end_date">
          <text class="row-label">计划周期</text>
          <text class="row-value">{{ formatPeriod(plan.start_date, plan.end_date) }}</text>
        </view>
        <view class="header-row" v-if="plan.visit_count !== undefined">
          <text class="row-label">照护次数</text>
          <text class="row-value">已上门照护 {{ plan.visit_count }} 次</text>
        </view>
      </view>

      <!-- 照护目标 -->
      <view class="section-card" v-if="plan.goals">
        <view class="section-title">照护目标</view>
        <view class="section-content">{{ plan.goals }}</view>
      </view>

      <!-- 护理项清单 -->
      <view class="section-card" v-if="plan.items.length">
        <view class="section-title">护理项清单</view>
        <view v-for="(item, index) in plan.items" :key="index" class="care-item">
          <view class="care-item-name">{{ item.name }}</view>
          <view class="care-item-desc" v-if="item.desc">{{ item.desc }}</view>
        </view>
      </view>

      <!-- 上门照护记录 -->
      <view class="section-card visits-card">
        <view class="section-title">上门照护记录</view>
        <view v-if="visits.length" class="visit-list">
          <view v-for="visit in visits" :key="visit.id" class="visit-item">
            <view class="visit-header">
              <view class="visit-time">{{ formatTime(visit.visit_at || visit.created_at) }}</view>
            </view>

            <!-- 护理项完成情况 -->
            <view v-if="visit.nursing_items.length" class="visit-block">
              <view class="visit-block-title">护理项</view>
              <view v-for="(nursing, index) in visit.nursing_items" :key="index" class="nursing-item">
                <view class="nursing-status" :class="{ done: nursing.done }">
                  {{ nursing.done ? '✔' : '○' }}
                </view>
                <view class="nursing-info">
                  <view class="nursing-name">{{ nursing.name }}</view>
                  <view class="nursing-remark" v-if="nursing.remark">{{ nursing.remark }}</view>
                </view>
              </view>
            </view>

            <!-- 生命体征 -->
            <view v-if="hasVitals(visit.vitals)" class="visit-block">
              <view class="visit-block-title">生命体征</view>
              <view class="vitals-grid">
                <view class="vital-item" v-if="visit.vitals?.blood_pressure">
                  <view class="vital-label">血压</view>
                  <view class="vital-value">{{ visit.vitals.blood_pressure }}</view>
                </view>
                <view class="vital-item" v-if="visit.vitals?.blood_glucose">
                  <view class="vital-label">血糖</view>
                  <view class="vital-value">{{ visit.vitals.blood_glucose }}</view>
                </view>
                <view class="vital-item" v-if="visit.vitals?.heart_rate">
                  <view class="vital-label">心率</view>
                  <view class="vital-value">{{ visit.vitals.heart_rate }}</view>
                </view>
                <view class="vital-item" v-if="visit.vitals?.oxygen">
                  <view class="vital-label">血氧</view>
                  <view class="vital-value">{{ visit.vitals.oxygen }}</view>
                </view>
                <view class="vital-item" v-if="visit.vitals?.weight">
                  <view class="vital-label">体重</view>
                  <view class="vital-value">{{ visit.vitals.weight }}</view>
                </view>
              </view>
            </view>

            <!-- 照片 -->
            <view v-if="visit.photos?.length" class="visit-block">
              <view class="visit-block-title">现场照片</view>
              <view class="photos-row">
                <image
                  v-for="(photo, index) in visit.photos"
                  :key="index"
                  class="photo-thumb"
                  :src="photo"
                  mode="aspectFill"
                  @click="previewPhoto(visit.photos || [], index)"
                />
              </view>
            </view>

            <!-- 备注 -->
            <view class="visit-block" v-if="visit.remark">
              <view class="visit-block-title">记录备注</view>
              <view class="visit-remark">{{ visit.remark }}</view>
            </view>

            <!-- 下次随访建议 -->
            <view class="visit-block follow-up" v-if="visit.follow_up_advice">
              <view class="visit-block-title">下次随访建议</view>
              <view class="visit-remark">{{ visit.follow_up_advice }}</view>
            </view>
          </view>
        </view>
        <view v-else class="visit-empty">
          <view class="visit-empty-text">暂无上门照护记录</view>
        </view>
      </view>
    </template>

    <view v-else class="empty">
      <view class="empty-icon">📋</view>
      <view class="empty-text">计划不存在</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getUserCarePlan } from '../../api/health'
import type { CarePlan, CareVisit } from '../../types'
import { useAuth } from '../../utils/useAuth'

const plan = ref<CarePlan | null>(null)
const loading = ref(false)

/** 照护记录（后端已倒序返回，直接展示） */
const visits = computed<CareVisit[]>(() => plan.value?.visits || [])

onLoad(async (query) => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  const id = Number(query?.id)
  if (!id) {
    uni.showToast({ title: '计划参数错误', icon: 'none' })
    return
  }
  loadPlan(id)
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

function formatTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** 判断是否存在任意生命体征数据 */
function hasVitals(vitals?: CareVisit['vitals']): boolean {
  if (!vitals) return false
  return Object.values(vitals).some(value => !!value)
}

function previewPhoto(photos: string[], current: number) {
  uni.previewImage({ urls: photos, current })
}

async function loadPlan(id: number) {
  loading.value = true
  try {
    plan.value = await getUserCarePlan(id)
  } catch (error) {
    console.error('加载照护计划详情失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.care-plan-detail-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

.loading,
.empty {
  text-align: center;
  padding: 80rpx 0;
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
}

/* 计划头部 */
.plan-header {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.plan-name {
  flex: 1;
  min-width: 0;
  font-size: 34rpx;
  font-weight: 600;
  margin-right: 20rpx;
  word-break: break-all;
}

.plan-status {
  flex-shrink: 0;
  font-size: 22rpx;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
}

.plan-status.draft {
  background: rgba(255, 255, 255, 0.25);
  color: #ffffff;
}

.plan-status.running {
  background: #e6f0ff;
  color: #007AFF;
}

.plan-status.paused {
  background: #fff7e6;
  color: #fa8c16;
}

.plan-status.done {
  background: #f6ffed;
  color: #52c41a;
}

.header-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.type-chip {
  font-size: 22rpx;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.2);
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

.freq-chip {
  font-size: 22rpx;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.15);
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

.header-row {
  display: flex;
  margin-top: 12rpx;
  line-height: 1.5;
}

.header-row .row-label {
  width: 140rpx;
  flex-shrink: 0;
  font-size: 24rpx;
  opacity: 0.85;
}

.header-row .row-value {
  flex: 1;
  min-width: 0;
  font-size: 24rpx;
  word-break: break-all;
}

/* 通用卡片 */
.section-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20rpx;
}

.section-content {
  font-size: 26rpx;
  color: #333333;
  line-height: 1.7;
  word-break: break-all;
}

/* 护理项清单 */
.care-item {
  padding: 18rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.care-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.care-item-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.care-item-desc {
  font-size: 24rpx;
  color: #999999;
  line-height: 1.5;
  word-break: break-all;
}

/* 照护记录 */
.visit-item {
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.visit-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.visit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.visit-time {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.visit-block {
  background: #f8f9fa;
  border-radius: 14rpx;
  padding: 20rpx;
  margin-bottom: 16rpx;
}

.visit-block:last-child {
  margin-bottom: 0;
}

.visit-block-title {
  font-size: 24rpx;
  color: #999999;
  margin-bottom: 12rpx;
}

/* 护理项完成情况 */
.nursing-item {
  display: flex;
  align-items: flex-start;
  padding: 8rpx 0;
}

.nursing-status {
  width: 40rpx;
  font-size: 28rpx;
  color: #bbbbbb;
  flex-shrink: 0;
}

.nursing-status.done {
  color: #52c41a;
}

.nursing-info {
  flex: 1;
  min-width: 0;
}

.nursing-name {
  font-size: 26rpx;
  color: #1a1a1a;
}

.nursing-remark {
  font-size: 22rpx;
  color: #999999;
  line-height: 1.5;
  margin-top: 4rpx;
  word-break: break-all;
}

/* 生命体征 */
.vitals-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.vital-item {
  flex: 1;
  min-width: 180rpx;
  background: #ffffff;
  border-radius: 12rpx;
  padding: 16rpx;
  text-align: center;
}

.vital-label {
  font-size: 22rpx;
  color: #999999;
  margin-bottom: 8rpx;
}

.vital-value {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  word-break: break-all;
}

/* 照片 */
.photos-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.photo-thumb {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  background: #eeeeee;
}

/* 备注/建议 */
.visit-remark {
  font-size: 26rpx;
  color: #333333;
  line-height: 1.7;
  word-break: break-all;
}

.follow-up {
  border-left: 6rpx solid #007AFF;
}

.visit-empty {
  text-align: center;
  padding: 40rpx 0;
}

.visit-empty-text {
  font-size: 26rpx;
  color: #999999;
}
</style>
