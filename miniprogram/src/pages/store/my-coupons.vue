<template>
  <view class="coupon-page">
    <!-- 状态 Tab -->
    <view class="tab-bar">
      <view
        v-for="tab in tabs"
        :key="tab.value"
        class="tab-item"
        :class="{ active: activeStatus === tab.value }"
        @click="switchTab(tab.value)"
      >
        {{ tab.label }}
      </view>
    </view>

    <!-- 券列表 -->
    <view v-if="list.length" class="coupon-list">
      <view v-for="coupon in list" :key="coupon.id" class="coupon-card" :class="{ disabled: activeStatus !== 1 }">
        <view class="coupon-left">
          <template v-if="couponType(coupon) === 1">
            <view class="coupon-value">¥{{ Number(coupon.discount_amount || 0).toFixed(0) }}</view>
            <view class="coupon-threshold">
              {{ Number(coupon.threshold_amount) > 0 ? `满${Number(coupon.threshold_amount).toFixed(0)}可用` : '无门槛' }}
            </view>
          </template>
          <template v-else>
            <view class="coupon-value">{{ ((Number(coupon.discount_rate) || 0) * 10).toFixed(1) }}<text class="coupon-unit">折</text></view>
            <view class="coupon-threshold">
              {{ Number(coupon.threshold_amount) > 0 ? `满${Number(coupon.threshold_amount).toFixed(0)}可用` : '无门槛' }}
            </view>
          </template>
        </view>
        <view class="coupon-right">
          <view class="coupon-name">{{ coupon.name || '优惠券' }}</view>
          <view class="coupon-meta">
            <text v-if="couponType(coupon) === 2" class="coupon-type-tag">折扣券</text>
            <text v-else class="coupon-type-tag">满减券</text>
            <text class="coupon-validity">{{ validityText(coupon) }}</text>
          </view>
          <view class="coupon-footer">
            <text v-if="activeStatus === 2 && coupon.used_at" class="coupon-order">核销订单：{{ coupon.order_no || '-' }}</text>
            <text v-if="activeStatus === 3" class="coupon-order">已过期，无法使用</text>
            <view v-if="activeStatus === 1" class="use-btn" @click="goUse">去使用</view>
          </view>
        </view>
      </view>

      <view v-if="noMore && list.length" class="no-more">没有更多了</view>
      <view v-if="loading && list.length" class="no-more">加载中...</view>
    </view>

    <!-- 空状态 -->
    <view v-else-if="!loading" class="empty">
      <view class="empty-icon">🎟️</view>
      <view class="empty-text">{{ emptyText }}</view>
      <view class="empty-desc">下单时可选择优惠券抵扣</view>
      <view class="empty-btn" @click="goHome">去逛逛</view>
    </view>

    <!-- 首次加载中 -->
    <view v-else class="loading">加载中...</view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { getMyCoupons } from '../../api/coupon'
import type { MyCoupon } from '../../types/coupon'
import { useAuth } from '../../utils/useAuth'

const tabs = [
  { label: '未使用', value: 1 },
  { label: '已使用', value: 2 },
  { label: '已过期', value: 3 }
]

const activeStatus = ref(1)
const list = ref<MyCoupon[]>([])
const loading = ref(false)
const page = ref(1)
const noMore = ref(false)
const pageSize = 10

const emptyText = ref('暂无可用优惠券')

async function loadList(reset = false) {
  if (loading.value) {
    return
  }
  if (reset) {
    page.value = 1
    noMore.value = false
    list.value = []
  }
  if (noMore.value) {
    return
  }

  loading.value = true
  try {
    const res = await getMyCoupons(activeStatus.value, page.value, pageSize)
    const items = res.list || []
    if (page.value === 1) {
      list.value = items
    } else {
      list.value = [...list.value, ...items]
    }
    if (list.value.length >= res.total || items.length < pageSize) {
      noMore.value = true
    }
  } catch (_e) {
    // 拦截器提示
  } finally {
    loading.value = false
  }
}

function switchTab(status: number) {
  if (activeStatus.value === status) {
    return
  }
  activeStatus.value = status
  emptyText.value = status === 1 ? '暂无可用优惠券' : status === 2 ? '暂无已使用优惠券' : '暂无已过期优惠券'
  loadList(true)
}

function couponType(coupon: MyCoupon): number {
  return Number(coupon.type || 1)
}

function validityText(coupon: MyCoupon): string {
  const expired = (coupon.expired_at || '').replace('T', ' ').slice(0, 10)
  return `有效期至 ${expired}`
}

function goUse() {
  uni.switchTab({ url: '/pages/store/home' })
}

function goHome() {
  uni.switchTab({ url: '/pages/store/home' })
}

onLoad(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (authed) {
    loadList(true)
  }
})

onPullDownRefresh(() => {
  loadList(true).finally(() => uni.stopPullDownRefresh())
})

onReachBottom(() => {
  if (!noMore.value) {
    page.value += 1
    loadList()
  }
})
</script>

<style scoped>
.coupon-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  box-sizing: border-box;
}

/* Tab 栏 */
.tab-bar {
  display: flex;
  background: #ffffff;
  border-radius: 16rpx;
  padding: 8rpx;
  margin-bottom: 24rpx;
}
.tab-item {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666666;
  border-radius: 12rpx;
}
.tab-item.active {
  background: #007AFF;
  color: #ffffff;
  font-weight: 600;
}

/* 券卡片 */
.coupon-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.coupon-card {
  display: flex;
  background: #ffffff;
  border-radius: 16rpx;
  overflow: hidden;
}
.coupon-card.disabled {
  opacity: 0.55;
}
.coupon-left {
  width: 200rpx;
  background: linear-gradient(135deg, #ff6b3b 0%, #ff3b30 100%);
  color: #ffffff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32rpx 16rpx;
  box-sizing: border-box;
}
.coupon-card.disabled .coupon-left {
  background: linear-gradient(135deg, #b8b8b8 0%, #999999 100%);
}
.coupon-value {
  font-size: 52rpx;
  font-weight: 700;
  line-height: 1.1;
}
.coupon-unit {
  font-size: 28rpx;
  font-weight: 600;
}
.coupon-threshold {
  font-size: 22rpx;
  margin-top: 8rpx;
  opacity: 0.9;
}
.coupon-right {
  flex: 1;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
}
.coupon-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #222222;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.coupon-meta {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 8rpx;
}
.coupon-type-tag {
  font-size: 20rpx;
  color: #ff3b30;
  background: rgba(255, 59, 48, 0.08);
  border-radius: 6rpx;
  padding: 2rpx 8rpx;
}
.coupon-validity {
  font-size: 22rpx;
  color: #999999;
}
.coupon-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12rpx;
}
.coupon-order {
  font-size: 22rpx;
  color: #999999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.use-btn {
  font-size: 24rpx;
  color: #ffffff;
  background: #007AFF;
  border-radius: 999rpx;
  padding: 8rpx 28rpx;
}

/* 空状态与加载 */
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 40rpx;
}
.empty-icon {
  font-size: 96rpx;
}
.empty-text {
  font-size: 30rpx;
  color: #333333;
  margin-top: 24rpx;
  font-weight: 600;
}
.empty-desc {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}
.empty-btn {
  margin-top: 32rpx;
  font-size: 26rpx;
  color: #007AFF;
  border: 1rpx solid #007AFF;
  border-radius: 999rpx;
  padding: 12rpx 48rpx;
}
.loading {
  text-align: center;
  color: #999999;
  padding: 120rpx 0;
  font-size: 26rpx;
}
.no-more {
  text-align: center;
  color: #999999;
  font-size: 24rpx;
  padding: 24rpx 0;
}
</style>
