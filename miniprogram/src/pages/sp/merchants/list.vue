<template>
  <view class="merchant-list-container">
    <!-- 搜索和筛选栏 -->
    <view class="filter-bar">
      <view class="search-box">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商家名称"
          @confirm="handleSearch"
        />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>
      <picker
        mode="selector"
        :range="categoryOptions"
        range-key="name"
        :value="selectedCategoryIndex"
        @change="onCategoryChange"
      >
        <view class="category-picker">
          <text>{{ selectedCategoryLabel }}</text>
          <text class="arrow">▼</text>
        </view>
      </picker>
    </view>

    <!-- 商家列表 -->
    <scroll-view
      class="merchant-list"
      scroll-y
      @scrolltolower="loadMore"
    >
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
            <text class="label">行业分类：</text>
            <text class="value">{{ merchant.business_category }}</text>
          </view>
          <view class="info-item">
            <text class="label">入驻时间：</text>
            <text class="value">{{ formatDate(merchant.created_at) }}</text>
          </view>
        </view>

        <view class="merchant-stats">
          <view class="stat-item">
            <view class="stat-value">{{ merchant.total_users || 0 }}</view>
            <view class="stat-label">用户数</view>
          </view>
          <view class="stat-item">
            <view class="stat-value">{{ merchant.total_orders || 0 }}</view>
            <view class="stat-label">累计订单</view>
          </view>
          <view class="stat-item">
            <view class="stat-value">¥{{ formatAmount(merchant.total_amount) }}</view>
            <view class="stat-label">累计金额</view>
          </view>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && merchants.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && merchants.length === 0" class="empty">
        <text class="empty-icon">🏪</text>
        <text class="empty-text">暂无商家</text>
        <text class="empty-hint">试试调整搜索条件</text>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMerchantList, getCategories } from '@api'
import type { MerchantListItem, Category } from '@types'

// 搜索和筛选
const keyword = ref('')
const selectedCategoryId = ref<number | null>(null)
const categoryOptions = ref<{ id: number | null; name: string }[]>([])
const selectedCategoryIndex = ref(0)

// 列表数据
const merchants = ref<MerchantListItem[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

// 选中的分类标签
const selectedCategoryLabel = computed(() => {
  if (selectedCategoryIndex.value === 0) {
    return '全部分类'
  }
  return categoryOptions.value[selectedCategoryIndex.value]?.name || '全部分类'
})

onShow(() => {
  loadCategories()
  loadMerchants(true)
})

async function loadCategories() {
  try {
    const categories = await getCategories()
    categoryOptions.value = [
      { id: null, name: '全部分类' },
      ...categories.map((c: Category) => ({ id: c.id, name: c.name }))
    ]
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

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

    if (keyword.value) {
      params.keyword = keyword.value
    }

    if (selectedCategoryId.value !== null) {
      params.category_id = selectedCategoryId.value
    }

    const res = await getMerchantList(params)

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

function onCategoryChange(e: any) {
  const index = e.detail.value
  selectedCategoryIndex.value = index
  selectedCategoryId.value = categoryOptions.value[index]?.id || null
  loadMerchants(true)
}

function getStatusText(status: number): string {
  const statusMap: Record<number, string> = {
    1: '营业中',
    2: '休息中',
    3: '已关闭'
  }
  return statusMap[status] || '未知'
}

function getStatusClass(status: number): string {
  const classMap: Record<number, string> = {
    1: 'open',
    2: 'rest',
    3: 'closed'
  }
  return classMap[status] || ''
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function formatAmount(amount: number): string {
  if (amount >= 10000) {
    return (amount / 10000).toFixed(1) + '万'
  }
  return amount.toFixed(2)
}

function goDetail(merchantId: number) {
  uni.navigateTo({ url: `/pages/sp/merchants/detail?id=${merchantId}` })
}
</script>

<style scoped>
.merchant-list-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.filter-bar {
  background: #ffffff;
  padding: 24rpx;
  position: sticky;
  top: 0;
  z-index: 10;
}

.search-box {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
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
