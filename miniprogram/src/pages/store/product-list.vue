<template>
  <view class="pl-page">
    <!-- 顶部渐变 Hero（辅具零售/租赁差异化配色） -->
    <view class="pl-hero" :class="isRental ? 'hero-rental' : 'hero-retail'">
      <view class="hero-deco deco-1"></view>
      <view class="hero-deco deco-2"></view>
      <view class="hero-top">
        <text class="hero-badge">{{ isRental ? 'RENTAL' : 'RETAIL' }}</text>
      </view>
      <view class="hero-title">{{ heroTitle }}</view>
      <view class="hero-sub">{{ heroSub }}</view>
    </view>

    <!-- 商品列表 -->
    <view class="pl-body">
      <view class="section-head">
        <text class="section-title">{{ heroTitle }}</text>
        <text class="section-count">共 {{ list.length }} 件</text>
      </view>

      <view v-if="loading" class="pl-loading">
        <view v-for="n in 3" :key="n" class="pl-loading-card"></view>
      </view>

      <view v-else-if="list.length === 0" class="pl-empty">
        <text class="empty-icon">🛒</text>
        <text class="empty-text">暂无可选商品，敬请期待</text>
      </view>

      <view v-else class="pl-list">
        <view
          v-for="item in list"
          :key="item.id"
          class="pl-card"
          @click="goDetail(item)"
        >
          <view class="pl-card-photo">
            <image
              v-if="item.images && item.images.length"
              class="pl-photo"
              :src="item.images[0]"
              mode="aspectFill"
            />
            <view v-else class="pl-photo pl-photo-placeholder">
              <text>{{ isRental ? '🔑' : '🛍️' }}</text>
            </view>
          </view>
          <view class="pl-card-main">
            <view class="pl-name">{{ item.name }}</view>
            <text v-if="item.description" class="pl-desc">{{ item.description }}</text>
            <view class="pl-card-foot">
              <view class="pl-price-block">
                <template v-if="isRental">
                  <view class="pl-price">
                    <text class="price-symbol">¥</text>
                    <text class="price-value">{{ formatPrice(item.rental_price) }}</text>
                    <text class="price-unit">/{{ getRentalUnitText(item.rental_unit) }}</text>
                  </view>
                  <text class="rental-deposit-tip">押金 ¥{{ formatPrice(item.deposit) }}</text>
                </template>
                <template v-else>
                  <view class="pl-price">
                    <text class="price-symbol">¥</text>
                    <text class="price-value">{{ formatPrice(item.price) }}</text>
                  </view>
                  <text v-if="(Number(item.original_price) || 0) > 0" class="original-price">
                    ¥{{ formatPrice(item.original_price) }}
                  </text>
                </template>
              </view>
              <view class="pl-btn" :class="isRental ? 'btn-rental' : 'btn-retail'">查看</view>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getStoreProducts } from '../../api/store'
import type { Product } from '../../types'

// 1=辅具零售 2=辅具租赁
const type = ref<1 | 2>(1)
const list = ref<Product[]>([])
const loading = ref(false)

const isRental = computed(() => type.value === 2)
const heroTitle = computed(() => (type.value === 2 ? '辅具租赁' : '辅具零售'))
const heroSub = computed(() =>
  type.value === 2
    ? '按周/月灵活租赁，押金透明，康复辅具一站租用'
    : '专业康复辅助器具现货直购，正品保障，送货上门'
)

async function loadProducts() {
  loading.value = true
  try {
    const res = await getStoreProducts({ product_types: String(type.value), page_size: 50 })
    list.value = res?.list || []
  } catch (_e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

function formatPrice(price?: number | string): string {
  const num = Number(price || 0)
  return (Math.floor(num * 100) / 100).toString()
}

function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || '次'
}

function goDetail(item: Product) {
  uni.navigateTo({ url: `/pages/store/product?product_id=${item.id}` })
}

onLoad((options) => {
  const t = Number(options?.type || 1)
  type.value = t === 2 ? 2 : 1
})

onShow(() => {
  if (list.value.length === 0) loadProducts()
})
</script>

<style lang="scss" scoped>
.pl-page {
  min-height: 100vh;
  background: #f3f6f8;
  display: flex;
  flex-direction: column;
}

/* ===== Hero ===== */
.pl-hero {
  position: relative;
  overflow: hidden;
  padding: 40rpx 32rpx 72rpx;
  color: #fff;
}
.hero-retail {
  background: linear-gradient(135deg, #ff8a3d 0%, #f06a2a 55%, #d85a1e 100%);
}
.hero-rental {
  background: linear-gradient(135deg, #6a5cff 0%, #5242ee 55%, #3f31d6 100%);
}
.hero-deco {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
}
.deco-1 {
  width: 320rpx;
  height: 320rpx;
  right: -80rpx;
  top: -100rpx;
}
.deco-2 {
  width: 200rpx;
  height: 200rpx;
  left: -60rpx;
  bottom: -70rpx;
}
.hero-top {
  padding-bottom: 12rpx;
}
.hero-badge {
  font-size: 22rpx;
  letter-spacing: 6rpx;
  opacity: 0.85;
}
.hero-title {
  font-size: 52rpx;
  font-weight: 700;
}
.hero-sub {
  margin-top: 12rpx;
  font-size: 26rpx;
  opacity: 0.92;
  max-width: 560rpx;
  line-height: 1.5;
}

/* ===== body ===== */
.pl-body {
  flex: 1;
  padding: 28rpx 24rpx 40rpx;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 8rpx 20rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #111827;
}
.section-count {
  font-size: 24rpx;
  color: #9ca3af;
}

.pl-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.pl-card {
  display: flex;
  background: #fff;
  border-radius: 20rpx;
  padding: 20rpx;
  box-shadow: 0 6rpx 18rpx rgba(0, 0, 0, 0.05);
}
.pl-card-photo {
  width: 180rpx;
  height: 180rpx;
  border-radius: 16rpx;
  overflow: hidden;
  flex-shrink: 0;
  margin-right: 20rpx;
}
.pl-photo {
  width: 100%;
  height: 100%;
}
.pl-photo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64rpx;
  background: linear-gradient(135deg, #fff0e6, #ffe6d6);
}
.hero-rental ~ .pl-body .pl-photo-placeholder {
  background: linear-gradient(135deg, #eceaff, #dddbfc);
}
.pl-card-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.pl-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
  line-height: 1.3;
}
.pl-desc {
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #6b7280;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}
.pl-card-foot {
  margin-top: auto;
  padding-top: 14rpx;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
}
.pl-price {
  display: flex;
  align-items: baseline;
}
.pl-price .price-symbol {
  color: #e53935;
  font-size: 24rpx;
}
.pl-price .price-value {
  color: #e53935;
  font-size: 36rpx;
  font-weight: 700;
}
.pl-price .price-unit {
  font-size: 22rpx;
  color: #9ca3af;
}
.rental-deposit-tip {
  margin-left: 8rpx;
  font-size: 20rpx;
  color: #9ca3af;
}
.original-price {
  margin-left: 8rpx;
  font-size: 22rpx;
  color: #b0b0b0;
  text-decoration: line-through;
}
.pl-btn {
  font-size: 24rpx;
  color: #fff;
  padding: 12rpx 28rpx;
  border-radius: 999rpx;
}
.btn-retail {
  background: linear-gradient(135deg, #ff8a3d, #f06a2a);
}
.btn-rental {
  background: linear-gradient(135deg, #6a5cff, #5242ee);
}

/* loading / empty */
.pl-loading {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.pl-loading-card {
  height: 200rpx;
  border-radius: 20rpx;
  background: linear-gradient(90deg, #eef1f4 25%, #f7f9fb 50%, #eef1f4 75%);
  background-size: 400% 100%;
  animation: shimmer 1.4s infinite;
}
.pl-empty {
  padding: 120rpx 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20rpx;
  color: #9ca3af;
}
.empty-icon {
  font-size: 80rpx;
}
.empty-text {
  font-size: 26rpx;
}
@keyframes shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: 0 0;
  }
}
</style>