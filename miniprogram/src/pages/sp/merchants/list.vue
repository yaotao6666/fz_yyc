<template>
  <view class="merchant-list-container">
    <view class="filter-card">
      <view class="filter-header">
        <view>
          <view class="page-title">商家列表</view>
          <view class="page-subtitle">统一查看商家资料、收款配置和分账比例</view>
        </view>
        <button class="create-btn" @click="goCreate">新增商家</button>
      </view>

      <view class="search-row">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商家名称"
          confirm-type="search"
          @confirm="handleSearch"
        />
        <button class="search-btn" @click="handleSearch">搜索</button>
      </view>

      <view class="status-tabs">
        <view
          v-for="item in statusOptions"
          :key="item.value"
          class="status-tab"
          :class="{ active: currentStatus === item.value }"
          @click="changeStatus(item.value)"
        >
          {{ item.label }}
        </view>
      </view>
    </view>

    <scroll-view class="merchant-scroll" scroll-y @scrolltolower="loadMore">
      <view
        v-for="merchant in merchants"
        :key="merchant.id"
        class="merchant-card"
        @click="goDetail(merchant.id)"
      >
        <view class="merchant-header">
          <view class="merchant-main">
            <text class="merchant-name">{{ merchant.name }}</text>
            <text class="merchant-meta">{{ merchant.contact_name || '未设置联系人' }} · {{ merchant.contact_phone || '未设置电话' }}</text>
          </view>
          <view class="merchant-status" :class="getStatusClass(merchant.status)">
            {{ getStatusText(merchant.status) }}
          </view>
        </view>

        <view class="summary-grid">
          <view class="summary-item">
            <text class="summary-label">收款商户号</text>
            <text class="summary-value">{{ merchant.sub_mch_id || '未配置' }}</text>
          </view>
          <view class="summary-item">
            <text class="summary-label">支付配置</text>
            <text class="summary-value" :class="getPaymentConfigClass(merchant.payment_config_status)">
              {{ getPaymentConfigText(merchant.payment_config_status) }}
            </text>
          </view>
          <view class="summary-item">
            <text class="summary-label">分账配置</text>
            <text class="summary-value">
              {{ merchant.profit_sharing_enabled ? `已开启 ${formatRatio(merchant.profit_sharing_ratio)}` : '未开启' }}
            </text>
          </view>
          <view class="summary-item">
            <text class="summary-label">行业分类</text>
            <text class="summary-value">{{ merchant.business_category || '未设置' }}</text>
          </view>
        </view>

        <view class="stats-row">
          <view class="stat-box">
            <text class="stat-value">{{ merchant.total_users || 0 }}</text>
            <text class="stat-label">用户数</text>
          </view>
          <view class="stat-box">
            <text class="stat-value">{{ merchant.total_orders || 0 }}</text>
            <text class="stat-label">订单数</text>
          </view>
          <view class="stat-box">
            <text class="stat-value">¥{{ formatAmount(merchant.total_amount) }}</text>
            <text class="stat-label">累计金额</text>
          </view>
        </view>

        <view class="created-at">创建时间：{{ formatDate(merchant.created_at) }}</view>
      </view>

      <view v-if="loading" class="list-state">加载中...</view>
      <view v-else-if="merchants.length === 0" class="empty-state">
        <text class="empty-title">暂无商家</text>
        <text class="empty-desc">可以直接新增商家并配置收款账户与分账比例</text>
      </view>
      <view v-else-if="noMore" class="list-state">没有更多了</view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMerchantList } from '@api'
import type { MerchantListItem, PaymentConfigStatus } from '@types'
import { PaymentConfigStatusText } from '@types'

const statusOptions = [
  { label: '全部', value: '' },
  { label: '营业中', value: '1' },
  { label: '休息中', value: '2' },
  { label: '已关闭', value: '3' }
]

const keyword = ref('')
const currentStatus = ref('')
const merchants = ref<MerchantListItem[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

onShow(() => {
  loadMerchants(true)
})

async function loadMerchants(reset = false) {
  if (loading.value || (!reset && noMore.value)) {
    return
  }

  if (reset) {
    page.value = 1
    noMore.value = false
    merchants.value = []
  }

  loading.value = true
  try {
    const response = await getMerchantList({
      page: page.value,
      page_size: pageSize,
      keyword: keyword.value.trim() || undefined,
      status: currentStatus.value || undefined
    })

    merchants.value = reset ? response.list : merchants.value.concat(response.list)
    if (response.list.length < pageSize) {
      noMore.value = true
    } else {
      page.value += 1
    }
  } catch (requestError) {
    console.error('加载商家列表失败:', requestError)
    uni.showToast({ title: '加载商家列表失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  loadMerchants(true)
}

function changeStatus(status: string) {
  if (currentStatus.value === status) {
    return
  }
  currentStatus.value = status
  loadMerchants(true)
}

function loadMore() {
  loadMerchants()
}

function getStatusText(status: number) {
  const statusMap: Record<number, string> = {
    1: '营业中',
    2: '休息中',
    3: '已关闭'
  }
  return statusMap[status] || '未知状态'
}

function getStatusClass(status: number) {
  const classMap: Record<number, string> = {
    1: 'open',
    2: 'rest',
    3: 'closed'
  }
  return classMap[status] || ''
}

function getPaymentConfigText(status?: number) {
  return PaymentConfigStatusText[(status ?? 0) as PaymentConfigStatus] || '待完善'
}

function getPaymentConfigClass(status?: number) {
  return Number(status || 0) === 1 ? 'success' : 'warning'
}

function formatAmount(amount = 0) {
  return Number(amount || 0).toFixed(2)
}

function formatRatio(ratio = 0) {
  return `${Number(ratio || 0).toFixed(2)}%`
}

function formatDate(value?: string) {
  if (!value) {
    return '-'
  }
  return value.slice(0, 10)
}

function goCreate() {
  uni.navigateTo({ url: '/pages/sp/merchants/edit' })
}

function goDetail(merchantId: number) {
  uni.navigateTo({ url: `/pages/sp/merchants/detail?id=${merchantId}` })
}
</script>

<style scoped>
.merchant-list-container {
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

.filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.page-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2329;
}

.page-subtitle {
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.6;
  color: #86909c;
}

.create-btn {
  margin: 0;
  min-width: 180rpx;
  height: 72rpx;
  line-height: 72rpx;
  border: none;
  border-radius: 999rpx;
  background: #1677ff;
  color: #ffffff;
  font-size: 26rpx;
}

.search-row {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.search-input {
  flex: 1;
  height: 76rpx;
  padding: 0 28rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  font-size: 28rpx;
}

.search-btn {
  margin: 0;
  width: 150rpx;
  height: 76rpx;
  line-height: 76rpx;
  border: none;
  border-radius: 18rpx;
  background: #eef3ff;
  color: #1677ff;
  font-size: 26rpx;
}

.status-tabs {
  display: flex;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-top: 24rpx;
}

.status-tab {
  padding: 14rpx 24rpx;
  border-radius: 999rpx;
  background: #f7f8fa;
  font-size: 24rpx;
  color: #4e5969;
}

.status-tab.active {
  background: #e8f3ff;
  color: #1677ff;
}

.merchant-scroll {
  height: calc(100vh - 292rpx);
}

.merchant-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.merchant-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.merchant-main {
  flex: 1;
}

.merchant-name {
  display: block;
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2329;
}

.merchant-meta {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #86909c;
}

.merchant-status {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
}

.merchant-status.open {
  color: #1677ff;
  background: #e8f3ff;
}

.merchant-status.rest {
  color: #d48806;
  background: #fff7e6;
}

.merchant-status.closed {
  color: #cf1322;
  background: #fff1f0;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
  margin-top: 24rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f7f8fa;
}

.summary-item {
  min-width: 0;
}

.summary-label {
  display: block;
  font-size: 22rpx;
  color: #86909c;
}

.summary-value {
  display: block;
  margin-top: 10rpx;
  font-size: 25rpx;
  line-height: 1.5;
  color: #1f2329;
  word-break: break-all;
}

.summary-value.success {
  color: #389e0d;
}

.summary-value.warning {
  color: #d48806;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 24rpx;
}

.stat-box {
  padding: 24rpx 18rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
}

.stat-label {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: #86909c;
}

.created-at {
  margin-top: 20rpx;
  font-size: 23rpx;
  color: #86909c;
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
  justify-content: center;
  font-size: 28rpx;
}

.category-picker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 72rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;
}

.arrow {
  font-size: 20rpx;
  color: #999999;
  margin-left: 12rpx;
}

.merchant-list {
  padding: 24rpx;
}

.merchant-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.merchant-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.merchant-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  flex: 1;
}

.merchant-status {
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  font-weight: 500;
}

.merchant-status.open {
  background: #f6ffed;
  color: #52c41a;
}

.merchant-status.rest {
  background: #fff7e6;
  color: #fa8c16;
}

.merchant-status.closed {
  background: #f5f5f5;
  color: #999999;
}

.merchant-info {
  margin-bottom: 20rpx;
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.info-item {
  display: flex;
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-item .label {
  color: #999999;
}

.info-item .value {
  color: #333333;
}

.merchant-stats {
  display: flex;
  gap: 32rpx;
}

.stat-item {
  flex: 1;
  text-align: center;
}

.stat-value {
  font-size: 36rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
}

.loading, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 32rpx;
  color: #333333;
  font-weight: 500;
  margin-bottom: 12rpx;
}

.empty-hint {
  font-size: 26rpx;
  color: #999999;
}
</style>
