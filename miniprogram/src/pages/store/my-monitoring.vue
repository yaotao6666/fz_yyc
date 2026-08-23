<template>
  <view class="monitoring-page">
    <!-- 顶部分段切换 -->
    <view class="tab-bar">
      <view class="tab-item" :class="{ active: activeTab === 'list' }" @click="switchTab('list')">
        记录列表
      </view>
      <view class="tab-item" :class="{ active: activeTab === 'entry' }" @click="switchTab('entry')">
        录入记录
      </view>
    </view>

    <!-- 记录列表 -->
    <view v-if="activeTab === 'list'" class="record-tab">
      <view v-if="loading && !list.length" class="loading">加载中...</view>
      <view v-else-if="list.length" class="record-list">
        <view v-for="item in list" :key="item.id" class="record-card">
          <view class="record-header">
            <view class="record-type">{{ getRecordTypeText(item.record_type) }}</view>
            <view class="record-value">
              {{ item.value ?? '--' }}
              <text class="record-unit" v-if="item.unit">{{ item.unit }}</text>
            </view>
          </view>
          <view class="record-meta">
            <text class="meta-text">{{ formatTime(item.recorded_at || item.created_at) }}</text>
          </view>
          <view class="record-remark" v-if="item.remark">{{ item.remark }}</view>
        </view>

        <view v-if="noMore && list.length" class="no-more">没有更多了</view>
        <view v-if="loading && list.length" class="no-more">加载中...</view>
      </view>
      <view v-else class="empty">
        <view class="empty-icon">📊</view>
        <view class="empty-text">暂无体征记录</view>
        <view class="empty-desc">切换到"录入记录"记录您的每日生命体征</view>
      </view>
    </view>

    <!-- 录入表单 -->
    <view v-else class="entry-tab">
      <view class="entry-card">
        <view class="form-row">
          <view class="form-label">体征类型</view>
          <view class="type-chips">
            <view
              v-for="option in recordTypeOptions"
              :key="option.value"
              class="type-chip"
              :class="{ active: form.recordType === option.value }"
              @click="selectRecordType(option.value)"
            >
              {{ option.label }}
            </view>
          </view>
        </view>

        <view class="form-row">
          <view class="form-label">数值</view>
          <view class="form-control">
            <input
              class="form-input"
              type="digit"
              v-model="form.value"
              :placeholder="`请输入${getRecordTypeText(form.recordType)}数值`"
            />
            <text class="form-unit">{{ currentUnit }}</text>
          </view>
        </view>

        <view class="form-row">
          <view class="form-label">记录时间</view>
          <view class="form-control datetime-control">
            <picker mode="date" :value="form.recordedDate" @change="onDateChange">
              <view class="datetime-text">{{ form.recordedDate }}</view>
            </picker>
            <picker mode="time" :value="form.recordedTime" @change="onTimeChange">
              <view class="datetime-text">{{ form.recordedTime }}</view>
            </picker>
          </view>
        </view>

        <view class="form-row">
          <view class="form-label">备注</view>
          <view class="form-control">
            <textarea
              class="form-textarea"
              v-model="form.remark"
              placeholder="选填，可补充说明（如运动后、晨起空腹等）"
              :maxlength="100"
            />
          </view>
        </view>
      </view>

      <view class="submit-btn" @click="submitRecord">保存记录</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onReachBottom } from '@dcloudio/uni-app'
import { createUserMonitoring, getUserMonitoring } from '../../api/health'
import type { HealthMonitoring } from '../../types'
import { useAuth } from '../../utils/useAuth'

const list = ref<HealthMonitoring[]>([])
const loading = ref(false)
const page = ref(1)
const noMore = ref(false)
const pageSize = 10
const activeTab = ref<'list' | 'entry'>('list')

/** 体征类型选项：1血压 2血糖 3心率 4血氧 5体重 */
const recordTypeOptions = [
  { value: 1, label: '血压', unit: 'mmHg' },
  { value: 2, label: '血糖', unit: 'mmol/L' },
  { value: 3, label: '心率', unit: '次/分' },
  { value: 4, label: '血氧', unit: '%' },
  { value: 5, label: '体重', unit: 'kg' }
]

const form = ref({
  recordType: 1,
  value: '',
  recordedDate: '',
  recordedTime: '',
  remark: ''
})

/** 按所选类型预填单位 */
const currentUnit = computed(
  () => recordTypeOptions.find(option => option.value === form.value.recordType)?.unit || ''
)

/** 提交时的时间串：本地时间转 RFC3339（带时区偏移，后端按 *time.Time 解析） */
const recordedAt = computed(() => {
  const dateStr = `${form.value.recordedDate}T${form.value.recordedTime}:00`
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) {
    return dateStr
  }
  const offsetMinutes = -date.getTimezoneOffset()
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const absMinutes = Math.abs(offsetMinutes)
  const offsetHours = String(Math.floor(absMinutes / 60)).padStart(2, '0')
  const offsetMins = String(absMinutes % 60).padStart(2, '0')
  const localIso = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}T${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}`
  return `${localIso}${sign}${offsetHours}:${offsetMins}`
})

onLoad(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  initFormTime()
  loadList(true)
})

/** 初始化记录时间默认为当前时间 */
function initFormTime() {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  form.value.recordedDate = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
  form.value.recordedTime = `${pad(now.getHours())}:${pad(now.getMinutes())}`
}

function switchTab(tab: 'list' | 'entry') {
  activeTab.value = tab
}

function selectRecordType(type: number) {
  form.value.recordType = type
}

function onDateChange(event: any) {
  form.value.recordedDate = event.detail.value
}

function onTimeChange(event: any) {
  form.value.recordedTime = event.detail.value
}

function getRecordTypeText(recordType?: number): string {
  return recordTypeOptions.find(option => option.value === Number(recordType ?? 0))?.label || '体征'
}

function formatTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function loadList(reset = false) {
  if (loading.value) return
  if (!reset && noMore.value) return

  loading.value = true
  try {
    const res = await getUserMonitoring({ page: page.value, page_size: pageSize })
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
    console.error('加载体征记录失败:', error)
  } finally {
    loading.value = false
  }
}

async function submitRecord() {
  const value = Number(form.value.value)
  if (!form.value.value || !Number.isFinite(value)) {
    uni.showToast({ title: '请输入有效的数值', icon: 'none' })
    return
  }

  const remark = form.value.remark.trim()
  try {
    await createUserMonitoring({
      record_type: form.value.recordType,
      value,
      unit: currentUnit.value,
      recorded_at: recordedAt.value,
      remark: remark || undefined
    })
    uni.showToast({ title: '保存成功', icon: 'success' })
    resetForm()
    // 刷新列表并切回列表页
    page.value = 1
    noMore.value = false
    await loadList(true)
    activeTab.value = 'list'
  } catch (error: any) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  }
}

function resetForm() {
  form.value.value = ''
  form.value.remark = ''
  initFormTime()
}

onReachBottom(() => {
  if (activeTab.value === 'list') {
    loadList(false)
  }
})
</script>

<style scoped>
.monitoring-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

/* 分段切换 */
.tab-bar {
  display: flex;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 8rpx;
  margin-bottom: 24rpx;
}

.tab-item {
  flex: 1;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #666666;
  border-radius: 14rpx;
}

.tab-item.active {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-weight: 600;
}

/* 记录列表 */
.record-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.record-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.record-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.record-type {
  font-size: 28rpx;
  color: #666666;
}

.record-value {
  font-size: 40rpx;
  font-weight: 700;
  color: #1a1a1a;
}

.record-unit {
  font-size: 24rpx;
  font-weight: 400;
  color: #999999;
  margin-left: 8rpx;
}

.record-meta {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.meta-text {
  font-size: 22rpx;
  color: #999999;
}

.record-remark {
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #666666;
  line-height: 1.5;
  word-break: break-all;
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

/* 录入表单 */
.entry-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.form-row {
  display: flex;
  margin-bottom: 32rpx;
}

.form-row:last-child {
  margin-bottom: 0;
}

.form-label {
  width: 150rpx;
  flex-shrink: 0;
  font-size: 28rpx;
  color: #1a1a1a;
  padding-top: 12rpx;
}

.form-control {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
}

.form-input {
  flex: 1;
  min-width: 0;
  height: 72rpx;
  font-size: 30rpx;
  color: #1a1a1a;
  border-bottom: 1rpx solid #eeeeee;
}

.form-unit {
  font-size: 26rpx;
  color: #999999;
  margin-left: 16rpx;
}

/* 类型选择 */
.type-chips {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.type-chip {
  font-size: 26rpx;
  color: #666666;
  background: #f5f5f5;
  padding: 12rpx 28rpx;
  border-radius: 999rpx;
}

.type-chip.active {
  color: #ffffff;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  font-weight: 600;
}

/* 时间选择 */
.datetime-control {
  justify-content: space-between;
}

.datetime-text {
  min-width: 200rpx;
  height: 72rpx;
  line-height: 72rpx;
  font-size: 28rpx;
  color: #1a1a1a;
  background: #f5f5f5;
  border-radius: 12rpx;
  text-align: center;
  padding: 0 20rpx;
}

/* 备注输入 */
.form-textarea {
  width: 100%;
  min-height: 140rpx;
  font-size: 28rpx;
  color: #1a1a1a;
  background: #f5f5f5;
  border-radius: 12rpx;
  padding: 20rpx;
  box-sizing: border-box;
}

/* 提交按钮 */
.submit-btn {
  margin-top: 32rpx;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
}
</style>
