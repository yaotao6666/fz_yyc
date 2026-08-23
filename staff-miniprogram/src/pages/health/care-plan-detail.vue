<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { CarePlan, CareVisit, CareVisitVitals } from '@/types'
import { formatDate } from '@/utils/format'

const planId = ref<string>('')
const loading = ref(false)
const plan = ref<CarePlan | null>(null)

const planTypeText = (type?: number) => {
  const map: Record<number, string> = { 1: '生活照料', 2: '基础护理', 3: '康复训练', 4: '综合康养' }
  return type == null ? '' : (map[type] || '其他')
}

const planStatusText = (status?: number) => {
  const map: Record<number, string> = { 0: '草稿', 1: '执行中', 2: '已暂停', 3: '已完成' }
  return status == null ? '' : (map[status] || '未知')
}

function planStatusClass(status?: number) {
  const map: Record<number, string> = { 0: 'status-draft', 1: 'status-running', 2: 'status-paused', 3: 'status-done' }
  return status == null ? '' : (map[status] || '')
}

// 计划周期文案
const periodText = computed(() => {
  if (!plan.value?.start_date && !plan.value?.end_date) return ''
  const start = formatDate(plan.value?.start_date || '', 'YYYY-MM-DD')
  const end = formatDate(plan.value?.end_date || '', 'YYYY-MM-DD')
  return start || end ? `${start || '?'} ~ ${end || '长期'}` : ''
})

// 生命体征可读字段，仅展示已填写的项
function vitalsRows(vitals?: CareVisitVitals) {
  if (!vitals || typeof vitals !== 'object') return []
  const fields: { label: string; key: keyof CareVisitVitals }[] = [
    { label: '血压', key: 'blood_pressure' },
    { label: '血糖', key: 'blood_glucose' },
    { label: '心率', key: 'heart_rate' },
    { label: '血氧', key: 'oxygen' },
    { label: '体重', key: 'weight' }
  ]
  return fields
    .filter(f => vitals[f.key] != null && String(vitals[f.key]).trim() !== '')
    .map(f => ({ label: f.label, value: String(vitals[f.key]).trim() }))
}

// 照片预览
function previewPhoto(photos: string[] | undefined, current: string) {
  if (!photos?.length) return
  uni.previewImage({ urls: photos, current })
}

async function loadDetail() {
  if (!planId.value) return
  loading.value = true
  try {
    const data: any = await staffHealthApi.getCarePlanDetail(planId.value)
    plan.value = data as CarePlan
  } catch (e) {
    console.error('[CarePlanDetail] loadDetail error', e)
  } finally {
    loading.value = false
  }
}

function goRecordVisit() {
  if (!plan.value) return
  uni.navigateTo({ url: `/pages/health/care-visit-form?planId=${plan.value.id}` })
}

onLoad((options: any) => {
  planId.value = options?.id || ''
  if (!planId.value) {
    uni.showToast({ title: '缺少计划ID', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  loadDetail()
})
</script>

<template>
  <view class="page">
    <view v-if="!loading && !plan" class="empty card">
      <text class="empty-title">计划不存在</text>
      <text class="empty-desc">请返回重试</text>
    </view>

    <template v-else-if="plan">
      <view class="container">
        <!-- 计划信息 -->
        <view class="card section">
          <view class="plan-header">
            <view class="plan-title-wrap">
              <text class="plan-name">{{ plan.name }}</text>
              <text v-if="plan.plan_type" class="plan-type">{{ planTypeText(plan.plan_type) }}</text>
            </view>
            <text v-if="plan.status != null" :class="['plan-status', planStatusClass(plan.status)]">
              {{ planStatusText(plan.status) }}
            </text>
          </view>

          <view class="info-row" v-if="periodText">
            <text class="info-label">计划周期</text>
            <text class="info-value">{{ periodText }}</text>
          </view>
          <view class="info-row" v-if="plan.frequency">
            <text class="info-label">服务频次</text>
            <text class="info-value">{{ plan.frequency }}</text>
          </view>
          <view class="info-row" v-if="plan.goals">
            <text class="info-label">照护目标</text>
            <text class="info-value goal-text">{{ plan.goals }}</text>
          </view>
          <view class="info-row">
            <text class="info-label">照护次数</text>
            <text class="info-value">{{ plan.visit_count || 0 }}</text>
          </view>
        </view>

        <!-- 护理项 -->
        <view class="card section">
          <view class="section-title">护理项</view>
          <view v-if="!plan.items || plan.items.length === 0" class="no-data">暂无护理项</view>
          <view v-else>
            <view v-for="(item, idx) in plan.items" :key="idx" class="care-item">
              <view class="care-item-dot"></view>
              <view class="care-item-main">
                <text class="care-item-name">{{ item.name }}</text>
                <text v-if="item.desc" class="care-item-desc">{{ item.desc }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 历史照护记录 -->
        <view class="card section">
          <view class="section-title">
            <text>历史照护记录</text>
            <text v-if="plan.visits?.length" class="record-total">共 {{ plan.visits.length }} 条</text>
          </view>
          <view v-if="!plan.visits || plan.visits.length === 0" class="no-data">暂无照护记录</view>
          <view v-else>
            <view v-for="visit in plan.visits" :key="visit.id" class="visit-item">
              <view class="visit-header">
                <text class="visit-time">{{ formatDate(visit.visit_at || visit.created_at || '', 'YYYY-MM-DD HH:mm') }}</text>
                <text class="visit-id">#{{ visit.id }}</text>
              </view>

              <!-- 护理项完成情况 -->
              <view v-if="visit.nursing_items.length" class="visit-block">
                <view
                  v-for="(item, idx) in visit.nursing_items"
                  :key="idx"
                  class="visit-nursing-row"
                >
                  <text :class="['visit-nursing-dot', { done: item.done }]">
                    {{ item.done ? '✓' : '' }}
                  </text>
                  <text :class="['visit-nursing-name', { undone: !item.done }]">{{ item.name }}</text>
                  <text v-if="item.remark" class="visit-nursing-remark">备注：{{ item.remark }}</text>
                </view>
              </view>

              <!-- 生命体征 -->
              <view v-if="vitalsRows(visit.vitals).length" class="visit-block">
                <view class="vitals-wrap">
                  <text v-for="v in vitalsRows(visit.vitals)" :key="v.label" class="vital-chip">
                    {{ v.label }} {{ v.value }}
                  </text>
                </view>
              </view>

              <!-- 照片 -->
              <view v-if="visit.photos?.length" class="visit-block">
                <view class="photos-wrap">
                  <image
                    v-for="(url, idx) in visit.photos"
                    :key="idx"
                    class="photo-thumb"
                    :src="url"
                    mode="aspectFill"
                    @tap="previewPhoto(visit.photos, url)"
                  />
                </view>
              </view>

              <!-- 备注与随访建议 -->
              <view v-if="visit.remark" class="visit-block">
                <text class="visit-block-label">服务备注</text>
                <text class="visit-block-text">{{ visit.remark }}</text>
              </view>
              <view v-if="visit.follow_up_advice" class="visit-block">
                <text class="visit-block-label">下次随访建议</text>
                <text class="visit-block-text">{{ visit.follow_up_advice }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 底部操作栏 -->
      <view class="footer-bar">
        <button class="btn-record" @tap="goRecordVisit">录入照护记录</button>
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

.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.plan-title-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
  .plan-name { font-size: 34rpx; font-weight: 600; color: #333; }
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

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 10rpx 0;
  .info-label { width: 160rpx; font-size: 26rpx; color: #999; flex-shrink: 0; }
  .info-value { flex: 1; font-size: 26rpx; color: #333; text-align: right; line-height: 1.5; }
  .goal-text {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
    overflow: hidden;
  }
}

/* 护理项 */
.care-item {
  display: flex;
  align-items: flex-start;
  padding: 14rpx 0;
  &:last-child { padding-bottom: 0; }
  .care-item-dot {
    width: 16rpx; height: 16rpx;
    border-radius: 50%;
    background: var(--primary-color);
    margin-top: 12rpx;
    margin-right: 16rpx;
    flex-shrink: 0;
  }
  .care-item-main { flex: 1; min-width: 0; }
  .care-item-name { font-size: 28rpx; color: #333; }
  .care-item-desc { display: block; font-size: 24rpx; color: #999; margin-top: 4rpx; }
}

/* 历史照护记录 */
.no-data { font-size: 26rpx; color: #999; padding: 12rpx 0; }
.visit-item {
  padding: 24rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; padding-bottom: 0; }
}
.visit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
  .visit-time { font-size: 28rpx; font-weight: 600; color: #333; }
  .visit-id { font-size: 24rpx; color: #999; }
}
.visit-block {
  padding: 10rpx 0;
}
.visit-nursing-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  padding: 4rpx 0;
  .visit-nursing-dot {
    width: 30rpx; height: 30rpx;
    border-radius: 50%;
    border: 2rpx solid #ccc;
    margin-right: 12rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22rpx;
    color: #fff;
    flex-shrink: 0;
    &.done { background: var(--primary-color); border-color: var(--primary-color); }
  }
  .visit-nursing-name {
    font-size: 26rpx; color: #333;
    &.undone { color: #999; text-decoration: line-through; }
  }
  .visit-nursing-remark {
    font-size: 24rpx; color: #999;
    margin-left: 12rpx;
    flex-basis: 100%;
    padding-left: 42rpx;
    box-sizing: border-box;
  }
}
.vitals-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.vital-chip {
  font-size: 24rpx;
  color: var(--info-color);
  background: #e6f4ff;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}
.photos-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.photo-thumb {
  width: 140rpx; height: 140rpx;
  border-radius: 8rpx;
  background: var(--bg-color);
}
.visit-block-label {
  display: block;
  font-size: 24rpx;
  color: #999;
  margin-bottom: 6rpx;
}
.visit-block-text {
  display: block;
  font-size: 26rpx;
  color: #555;
  line-height: 1.6;
}

/* 底部操作栏 */
.footer-bar {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}
.btn-record {
  width: 100%;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
}
</style>
