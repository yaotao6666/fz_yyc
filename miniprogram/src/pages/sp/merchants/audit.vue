<template>
  <view class="audit-container">
    <!-- 搜索筛选栏 -->
    <view class="filter-bar">
      <view class="search-box">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商家名称/联系人/电话"
          @confirm="handleSearch"
        />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
    </view>

    <!-- 状态标签切换 -->
    <view class="status-tabs">
      <view
        v-for="tab in statusTabs"
        :key="tab.value"
        class="tab-item"
        :class="{ active: currentStatus === tab.value }"
        @click="changeStatus(tab.value)"
      >
        {{ tab.label }}
        <text v-if="tab.count" class="tab-count">{{ tab.count }}</text>
      </view>
    </view>

    <!-- 商家列表 -->
    <scroll-view class="merchant-list" scroll-y @scrolltolower="loadMore">
      <view
        v-for="merchant in merchants"
        :key="merchant.id"
        class="merchant-card"
        @click="goDetail(merchant.id)"
      >
        <view class="merchant-header">
          <view class="merchant-name">{{ merchant.name }}</view>
          <view class="merchant-status" :class="getStatusClass(merchant.status)">
            {{ getStatusText(merchant.status) }}
          </view>
        </view>

        <view class="merchant-info">
          <view class="info-item">
            <text class="label">联系人：</text>
            <text class="value">{{ merchant.contact_name }}</text>
          </view>
          <view class="info-item">
            <text class="label">联系电话：</text>
            <text class="value">{{ merchant.contact_phone }}</text>
          </view>
          <view class="info-item">
            <text class="label">行业分类：</text>
            <text class="value">{{ merchant.business_category }}</text>
          </view>
          <view class="info-item">
            <text class="label">申请时间：</text>
            <text class="value">{{ formatTime(merchant.applied_at) }}</text>
          </view>
        </view>

        <view v-if="merchant.status === 2 && merchant.reject_reason" class="reject-reason">
          <text class="label">拒绝原因：</text>
          <text class="value">{{ merchant.reject_reason }}</text>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && merchants.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && merchants.length === 0" class="empty">
        <text class="empty-icon">📋</text>
        <text class="empty-text">暂无商家入驻申请</text>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getSpMerchantsPending } from '@api'
import type { MerchantApplication } from '@types'
import { MerchantAuditStatus, MerchantAuditStatusText } from '@types'

type AuditTabValue = 'all' | MerchantAuditStatus

// 状态标签配置
const statusTabs: Array<{ label: string; value: AuditTabValue; count: number }> = [
  { label: '全部', value: 'all', count: 0 },
  { label: '待审核', value: MerchantAuditStatus.PENDING, count: 0 },
  { label: '已通过', value: MerchantAuditStatus.APPROVED, count: 0 },
  { label: '已拒绝', value: MerchantAuditStatus.REJECTED, count: 0 }
]

const currentStatus = ref<AuditTabValue>('all')
const keyword = ref('')
const merchants = ref<MerchantApplication[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

onShow(() => {
  loadMerchants(true)
})

async function loadMerchants(reset = false) {
  if (reset) {
    page.value = 1
    noMore.value = false
    merchants.value = []
  }

  if (noMore.value || loading.value) return

  loading.value = true

  try {
    const params: any = {
      page: page.value,
      page_size: pageSize
    }

    if (currentStatus.value !== 'all') {
      params.status = currentStatus.value
    }

    if (keyword.value) {
      params.keyword = keyword.value
    }

    const res = await getSpMerchantsPending(params)

    if (reset) {
      merchants.value = res.list
    } else {
      merchants.value.push(...res.list)
    }

    if (res.list.length < pageSize) {
      noMore.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载商家列表失败:', error)
  } finally {
    loading.value = false
  }
}

function loadMore() {
  loadMerchants()
}

function handleSearch() {
  loadMerchants(true)
}

function changeStatus(status: AuditTabValue) {
  currentStatus.value = status
  loadMerchants(true)
}

function getStatusText(status: number): string {
  return MerchantAuditStatusText[status] || '未知'
}

function getStatusClass(status: number): string {
  const classMap: Record<number, string> = {
    [MerchantAuditStatus.PENDING]: 'pending',
    [MerchantAuditStatus.APPROVED]: 'approved',
    [MerchantAuditStatus.REJECTED]: 'rejected'
  }
  return classMap[status] || ''
}

function formatTime(time: string): string {
  if (!time) return '-'
  const date = new Date(time)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/sp/merchants/audit-detail?id=${id}` })
}
</script>

<style scoped>
.audit-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.filter-bar {
  background: #ffffff;
  padding: 24rpx;
}

.search-box {
  display: flex;
  gap: 16rpx;
}

.search-input {
  flex: 1;
  height: 72rpx;
  background: #f8f9fa;
  border-radius: 36rpx;
  padding: 0 32rpx;
  font-size: 28rpx;
}

.search-btn {
  width: 120rpx;
  height: 72rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
}

.status-tabs {
  display: flex;
  background: #ffffff;
  padding: 24rpx 0;
  margin-bottom: 16rpx;
}

.tab-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666666;
  padding: 16rpx 0;
  position: relative;
}

.tab-item.active {
  color: #007AFF;
  font-weight: 600;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 4rpx;
  background: #007AFF;
  border-radius: 2rpx;
}

.tab-count {
  display: inline-block;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  background: #ff4d4f;
  color: #ffffff;
  border-radius: 16rpx;
  font-size: 22rpx;
  padding: 0 8rpx;
  margin-left: 8rpx;
}

.merchant-list {
  padding: 0 24rpx;
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
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.merchant-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.merchant-status {
  font-size: 26rpx;
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
}

.merchant-status.pending {
  background: #fff7e6;
  color: #fa8c16;
}

.merchant-status.approved {
  background: #f6ffed;
  color: #52c41a;
}

.merchant-status.rejected {
  background: #fff1f0;
  color: #ff4d4f;
}

.merchant-info {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.info-item {
  display: flex;
  font-size: 28rpx;
  color: #666666;
}

.info-item .label {
  width: 140rpx;
  color: #999999;
}

.info-item .value {
  flex: 1;
  color: #1a1a1a;
}

.reject-reason {
  margin-top: 16rpx;
  padding: 16rpx;
  background: #fff1f0;
  border-radius: 8rpx;
  font-size: 26rpx;
  color: #ff4d4f;
  display: flex;
}

.reject-reason .label {
  width: 140rpx;
}

.reject-reason .value {
  flex: 1;
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
  padding: 100rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999999;
}
</style>
