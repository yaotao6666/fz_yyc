<template>
  <view class="fitting-page">
    <!-- 推荐说明 -->
    <view class="intro-banner">
      <view class="intro-title">🛒 我的适配建议</view>
      <view class="intro-desc">根据您的健康评估结果，为您推荐合适的医疗辅具</view>
    </view>

    <!-- 列表 -->
    <view v-if="loading && !list.length" class="loading">加载中...</view>
    <view v-else-if="list.length" class="fitting-list">
      <view v-for="item in list" :key="item.id" class="fitting-card">
        <view class="card-header">
          <view class="card-status" :class="getStatusClass(item.status)">{{ getStatusText(item.status) }}</view>
          <view class="card-time">{{ formatTime(item.created_at || item.updated_at) }}</view>
        </view>

        <view class="card-row" v-if="item.symptom_desc">
          <text class="row-label">症状需求</text>
          <text class="row-value">{{ item.symptom_desc }}</text>
        </view>

        <view class="card-row" v-if="item.fitting_result">
          <text class="row-label">适配结论</text>
          <text class="row-value">{{ item.fitting_result }}</text>
        </view>

        <view class="product-chips" v-if="item.recommended_products?.length">
          <text v-for="product in item.recommended_products" :key="product.product_id" class="product-chip">
            {{ product.name }}
          </text>
        </view>
        <view v-else class="card-row">
          <text class="row-label">推荐商品</text>
          <text class="row-value empty-text">暂无推荐商品</text>
        </view>

        <!-- 展开详情：推荐商品明细 -->
        <view class="detail-box" v-if="expandedIds.has(item.id) && item.recommended_products?.length">
          <view
            v-for="product in item.recommended_products"
            :key="product.product_id"
            class="product-item"
          >
            <view class="product-info">
              <view class="product-name">
                {{ product.name }}
                <text class="sale-tag" :class="{ rental: product.sale_type === 2 }">
                  {{ product.sale_type === 2 ? '租赁' : '一口价' }}
                </text>
              </view>
              <view class="product-reason" v-if="product.reason">推荐理由：{{ product.reason }}</view>
            </view>
            <view class="go-buy-btn" @click="goProduct(product.product_id)">去选购</view>
          </view>
        </view>

        <!-- 操作区 -->
        <view class="card-footer">
          <view class="toggle-btn" @click="toggleExpand(item.id)">
            {{ expandedIds.has(item.id) ? '收起明细' : '查看明细' }}
          </view>
          <view
            v-if="item.status === 0"
            class="confirm-btn"
            @click="confirmItem(item)"
          >确认建议</view>
        </view>
      </view>

      <view v-if="noMore && list.length" class="no-more">没有更多了</view>
      <view v-if="loading && list.length" class="no-more">加载中...</view>
    </view>
    <view v-else class="empty">
      <view class="empty-icon">🛒</view>
      <view class="empty-text">暂无适配建议</view>
      <view class="empty-desc">完成健康评估后，可获取个性化的辅具适配建议</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import {
  confirmUserFittingRecommendation,
  getUserFittingRecommendations
} from '../../api/health'
import type { FittingRecommendation } from '../../types'
import { useAuth } from '../../utils/useAuth'

const list = ref<FittingRecommendation[]>([])
const loading = ref(false)
const page = ref(1)
const noMore = ref(false)
const pageSize = 10
const expandedIds = ref<Set<number>>(new Set())

onLoad(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  loadList(true)
})

function getStatusText(status?: number): string {
  return { 0: '草稿', 1: '已确认', 2: '已下单' }[Number(status ?? 0)] || '草稿'
}

function getStatusClass(status?: number): string {
  const map: Record<number, string> = { 0: 'draft', 1: 'confirmed', 2: 'ordered' }
  return map[Number(status ?? 0)] || 'draft'
}

function formatTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function toggleExpand(id: number) {
  const next = new Set(expandedIds.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  expandedIds.value = next
}

function goProduct(productId: number) {
  if (!productId) {
    uni.showToast({ title: '商品不存在', icon: 'none' })
    return
  }
  uni.navigateTo({ url: `/pages/store/product?product_id=${productId}` })
}

async function loadList(reset = false) {
  if (loading.value) return
  if (!reset && noMore.value) return

  loading.value = true
  try {
    const res = await getUserFittingRecommendations({ page: page.value, page_size: pageSize })
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
    console.error('加载适配建议失败:', error)
  } finally {
    loading.value = false
  }
}

function confirmItem(item: FittingRecommendation) {
  uni.showModal({
    title: '确认适配建议',
    content: '确认后将视为已采纳该适配建议，确定确认吗？',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const updated = await confirmUserFittingRecommendation(item.id)
        const index = list.value.findIndex(value => value.id === item.id)
        if (index !== -1) {
          list.value[index] = updated
        }
        uni.showToast({ title: '已确认', icon: 'success' })
      } catch (error: any) {
        uni.showToast({ title: error.message || '确认失败', icon: 'none' })
      }
    }
  })
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
.fitting-page {
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
.fitting-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.fitting-card {
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

.card-status {
  font-size: 22rpx;
  padding: 6rpx 18rpx;
  border-radius: 999rpx;
}

.card-status.draft {
  background: #f0f0f0;
  color: #999999;
}

.card-status.confirmed {
  background: #f6ffed;
  color: #52c41a;
}

.card-status.ordered {
  background: #e6f0ff;
  color: #007AFF;
}

.card-time {
  font-size: 22rpx;
  color: #999999;
}

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

.empty-text {
  color: #bbbbbb;
}

/* 推荐商品 chips */
.product-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.product-chip {
  font-size: 22rpx;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

/* 展开明细 */
.detail-box {
  background: #f8f9fa;
  border-radius: 14rpx;
  padding: 8rpx 20rpx;
  margin-bottom: 20rpx;
}

.product-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #eeeeee;
}

.product-item:last-child {
  border-bottom: none;
}

.product-info {
  flex: 1;
  min-width: 0;
  margin-right: 20rpx;
}

.product-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
  word-break: break-all;
}

.sale-tag {
  display: inline-block;
  margin-left: 10rpx;
  font-size: 20rpx;
  font-weight: 400;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 2rpx 12rpx;
  border-radius: 999rpx;
  vertical-align: middle;
}

.sale-tag.rental {
  color: #ff9500;
  background: rgba(255, 149, 0, 0.1);
}

.product-reason {
  font-size: 24rpx;
  color: #666666;
  line-height: 1.5;
  word-break: break-all;
}

.go-buy-btn {
  flex-shrink: 0;
  height: 56rpx;
  line-height: 56rpx;
  padding: 0 24rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 24rpx;
  font-weight: 600;
}

/* 操作区 */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 20rpx;
}

.toggle-btn {
  font-size: 26rpx;
  color: #007AFF;
}

.confirm-btn {
  height: 64rpx;
  line-height: 64rpx;
  padding: 0 36rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 26rpx;
  font-weight: 600;
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
