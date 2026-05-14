<template>
  <view class="history-container">
    <view class="filter-card">
      <view class="section-title">分账历史</view>
      <view class="filter-row">
        <picker mode="selector" :range="statusOptions" range-key="label" :value="statusIndex" @change="onStatusChange">
          <view class="picker-field">{{ statusOptions[statusIndex]?.label || '全部状态' }}</view>
        </picker>
      </view>
      <view class="filter-row">
        <picker mode="date" :value="filters.start_date" @change="onStartDateChange">
          <view class="picker-field">{{ filters.start_date || '开始日期' }}</view>
        </picker>
        <picker mode="date" :value="filters.end_date" @change="onEndDateChange">
          <view class="picker-field">{{ filters.end_date || '结束日期' }}</view>
        </picker>
      </view>
      <button class="search-btn" @click="reloadRecords">查询记录</button>
    </view>

    <scroll-view class="history-scroll" scroll-y @scrolltolower="loadMore">
      <view v-for="record in records" :key="record.id" class="record-card">
        <view class="record-header">
          <view>
            <text class="record-order">{{ record.order_no }}</text>
            <text class="record-date">{{ formatDateTime(record.profit_sharing_date || record.created_at) }}</text>
          </view>
          <view class="status-tag" :class="getStatusClass(record.status)">
            {{ getStatusText(record.status) }}
          </view>
        </view>

        <view class="record-grid">
          <view class="record-item">
            <text class="item-label">支付金额</text>
            <text class="item-value">¥{{ formatAmount(record.pay_amount) }}</text>
          </view>
          <view class="record-item">
            <text class="item-label">抽佣比例</text>
            <text class="item-value">{{ formatRatio(record.profit_sharing_ratio) }}</text>
          </view>
          <view class="record-item">
            <text class="item-label">抽佣金额</text>
            <text class="item-value highlight">¥{{ formatAmount(record.profit_sharing_amount) }}</text>
          </view>
          <view class="record-item">
            <text class="item-label">商家实收</text>
            <text class="item-value">¥{{ formatAmount(record.merchant_received_amount) }}</text>
          </view>
        </view>

        <view v-if="record.error_message" class="error-text">处理说明：{{ record.error_message }}</view>
      </view>

      <view v-if="loading" class="list-state">加载中...</view>
      <view v-else-if="records.length === 0" class="empty-state">
        <text class="empty-title">暂无分账记录</text>
        <text class="empty-desc">支付完成后，系统会在这里沉淀每笔抽佣处理结果。</text>
      </view>
      <view v-else-if="noMore" class="list-state">没有更多了</view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMerchantProfitSharingRecords } from '@api'
import type { ProfitSharingRecord } from '@types'

const statusIndex = ref(0)
const statusOptions = [
  { label: '全部状态' },
  { label: '待处理', value: 0 },
  { label: '已成功', value: 1 },
  { label: '已失败', value: 2 },
  { label: '已跳过', value: 3 }
]

const filters = reactive({
  status: undefined as number | undefined,
  start_date: '',
  end_date: ''
})

const records = ref<ProfitSharingRecord[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 20

onShow(() => {
  reloadRecords()
})

async function loadRecords(reset = false) {
  if (loading.value || (!reset && noMore.value)) {
    return
  }

  if (reset) {
    page.value = 1
    noMore.value = false
    records.value = []
  }

  loading.value = true
  try {
    const response = await getMerchantProfitSharingRecords({
      page: page.value,
      page_size: pageSize,
      status: filters.status,
      start_date: filters.start_date || undefined,
      end_date: filters.end_date || undefined
    })
    records.value = reset ? response.list : records.value.concat(response.list)
    if (response.list.length < pageSize) {
      noMore.value = true
    } else {
      page.value += 1
    }
  } catch (requestError) {
    console.error('加载分账记录失败:', requestError)
    uni.showToast({ title: '加载分账记录失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function reloadRecords() {
  if (filters.start_date && filters.end_date && filters.start_date > filters.end_date) {
    uni.showToast({ title: '开始日期不能晚于结束日期', icon: 'none' })
    return
  }
  loadRecords(true)
}

function loadMore() {
  loadRecords()
}

function onStatusChange(event: { detail: { value: string } }) {
  statusIndex.value = Number(event.detail.value || 0)
  filters.status = statusOptions[statusIndex.value]?.value
  reloadRecords()
}

function onStartDateChange(event: { detail: { value: string } }) {
  filters.start_date = event.detail.value
}

function onEndDateChange(event: { detail: { value: string } }) {
  filters.end_date = event.detail.value
}

function getStatusText(status: number) {
  const textMap: Record<number, string> = {
    0: '待处理',
    1: '已成功',
    2: '已失败',
    3: '已跳过'
  }
  return textMap[status] || '未知状态'
}

function getStatusClass(status: number) {
  const classMap: Record<number, string> = {
    0: 'pending',
    1: 'success',
    2: 'failed',
    3: 'skipped'
  }
  return classMap[status] || ''
}

function formatAmount(amount = 0) {
  return Number(amount || 0).toFixed(2)
}

function formatRatio(ratio = 0) {
  return `${Number(ratio || 0).toFixed(2)}%`
}

function formatDateTime(value?: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').slice(0, 19)
}
</script>

<style scoped>
.history-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  box-sizing: border-box;
}

.filter-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 24rpx;
}

.filter-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-bottom: 16rpx;
}

.picker-field {
  padding: 20rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  font-size: 26rpx;
  color: #1f2329;
}

.search-btn {
  margin-top: 12rpx;
  width: 100%;
  height: 80rpx;
  line-height: 80rpx;
  border: none;
  border-radius: 18rpx;
  background: #1677ff;
  color: #ffffff;
  font-size: 28rpx;
}

.history-scroll {
  height: calc(100vh - 280rpx);
}

.record-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.record-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.record-order {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
}

.record-date {
  display: block;
  margin-top: 10rpx;
  font-size: 23rpx;
  color: #86909c;
}

.status-tag {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
}

.status-tag.pending {
  color: #1677ff;
  background: #e8f3ff;
}

.status-tag.success {
  color: #389e0d;
  background: #f6ffed;
}

.status-tag.failed {
  color: #cf1322;
  background: #fff1f0;
}

.status-tag.skipped {
  color: #d48806;
  background: #fff7e6;
}

.record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
  margin-top: 24rpx;
}

.record-item {
  padding: 22rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
}

.item-label {
  display: block;
  font-size: 22rpx;
  color: #86909c;
}

.item-value {
  display: block;
  margin-top: 10rpx;
  font-size: 26rpx;
  font-weight: 600;
  color: #1f2329;
}

.item-value.highlight {
  color: #1677ff;
}

.error-text {
  margin-top: 20rpx;
  font-size: 24rpx;
  line-height: 1.6;
  color: #cf1322;
}

.list-state,
.empty-state {
  padding: 40rpx 24rpx;
  text-align: center;
  color: #86909c;
}

.empty-title {
  display: block;
  font-size: 30rpx;
  color: #1f2329;
}

.empty-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.6;
}
</style>
