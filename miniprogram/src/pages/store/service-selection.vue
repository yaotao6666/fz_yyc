<template>
  <view class="svc-page">
    <!-- ===== 顶部渐变 Hero（随模式切换配色，非统一风格） ===== -->
    <view class="svc-hero" :class="mode === 'escort' ? 'hero-escort' : 'hero-wellness'">
      <view class="hero-deco deco-1"></view>
      <view class="hero-deco deco-2"></view>
      <view class="hero-top">
        <text class="hero-badge">{{ mode === 'escort' ? 'PEIZHEN' : 'KANGYANG' }}</text>
      </view>
      <view class="hero-title">{{ heroTitle }}</view>
      <view class="hero-sub">{{ heroSub }}</view>
    </view>

    <!-- ===== 服务列表 ===== -->
    <view class="svc-body">
      <view class="section-head">
        <text class="section-title">{{ mode === 'escort' ? '陪同就医服务' : '居家照护套餐' }}</text>
        <text class="section-count">共 {{ list.length }} 项</text>
      </view>

      <view v-if="loading" class="svc-loading">
        <view v-for="n in 3" :key="n" class="svc-loading-card"></view>
      </view>

      <view v-else-if="list.length === 0" class="svc-empty">
        <text class="empty-icon">🍃</text>
        <text class="empty-text">暂无可选服务，敬请期待</text>
      </view>

      <view v-else class="svc-list">
        <view
          v-for="item in list"
          :key="item.id"
          class="svc-card"
          @click="goDetail(item)"
        >
          <view class="svc-card-photo">
            <image
              v-if="item.images && item.images.length"
              class="svc-photo"
              :src="item.images[0]"
              mode="aspectFill"
            />
            <view v-else class="svc-photo svc-photo-placeholder">
              <text>{{ mode === 'escort' ? '🏥' : '🌿' }}</text>
            </view>
          </view>
          <view class="svc-card-main">
            <view class="svc-name">{{ item.name }}</view>
            <text class="svc-desc">{{ cardDesc(item) }}</text>
            <view v-if="tags(item).length" class="svc-tags">
              <text v-for="(t, i) in tags(item)" :key="i" class="svc-tag">{{ t }}</text>
            </view>
            <view class="svc-card-foot">
              <view class="svc-price">
                <text class="price-symbol">¥</text>
                <text class="price-value">{{ formatPrice(item.price) }}</text>
                <text class="price-unit">/{{ item.unit || '次' }}</text>
              </view>
              <view class="svc-btn" :class="mode === 'escort' ? 'btn-escort' : 'btn-wellness'">
                立即预约
              </view>
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

type SvcMode = 'wellness' | 'escort'

const mode = ref<SvcMode>('wellness')
const list = ref<Product[]>([])
const loading = ref(false)

const heroTitle = computed(() => (mode.value === 'escort' ? '陪诊服务' : '康养套餐'))
const heroSub = computed(() =>
  mode.value === 'escort'
    ? '全程陪同挂号、问诊、检查、取药，子女安心'
    : '专业护理员上门，生活照料到康复辅助一站配齐'
)

async function loadServices() {
  loading.value = true
  try {
    const productTypes = mode.value === 'escort' ? '4' : '3'
    const res = await getStoreProducts({ product_types: productTypes, page_size: 50 })
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

function serviceItems(item: Product): Array<{ name?: string; description?: string }> {
  const sc = (item.service_content || {}) as Record<string, unknown>
  if (Array.isArray(sc.services)) return sc.services as Array<{ name?: string; description?: string }>
  return []
}

function cardDesc(item: Product): string {
  if (mode.value === 'escort') {
    const includes = (item.service_content as any)?.includes || ''
    return includes ? (String(includes).slice(0, 40)) : (item.description || '')
  }
  const items = serviceItems(item)
  if (items.length) {
    const first = items[0]?.name || ''
    return first ? `含${items.length}项护理服务，如${first}` : (item.description || '')
  }
  return item.description || ''
}

function tags(item: Product): string[] {
  const sc = (item.service_content || {}) as Record<string, unknown>
  if (mode.value === 'escort') {
    const duration = sc.duration
    const list: string[] = []
    if (duration) list.push(String(duration))
    if (item.product_type === 4) list.push('医院陪同')
    return list.slice(0, 3)
  }
  return serviceItems(item).slice(0, 3).map((s) => s.name || '').filter(Boolean).slice(0, 3)
}

function goDetail(item: Product) {
  uni.navigateTo({ url: `/pages/store/product?product_id=${item.id}` })
}

onLoad((options) => {
  const m = (options?.mode || 'wellness') as SvcMode
  mode.value = m === 'escort' ? 'escort' : 'wellness'
})

onShow(() => {
  if (list.value.length === 0) loadServices()
})
</script>

<style lang="scss" scoped>
.svc-page {
  min-height: 100vh;
  background: #f3f6f8;
  display: flex;
  flex-direction: column;
}

/* ===== Hero ===== */
.svc-hero {
  position: relative;
  overflow: hidden;
  padding: 40rpx 32rpx 72rpx;
  color: #fff;
}
.hero-wellness {
  background: linear-gradient(135deg, #2bb3a3 0%, #12806f 55%, #0f6b5e 100%);
}
.hero-escort {
  background: linear-gradient(135deg, #3d7ef7 0%, #2963d6 55%, #1f4fae 100%);
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
  max-width: 520rpx;
  line-height: 1.5;
}

/* ===== body ===== */
.svc-body {
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

.svc-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.svc-card {
  display: flex;
  background: #fff;
  border-radius: 20rpx;
  padding: 20rpx;
  box-shadow: 0 6rpx 18rpx rgba(0, 0, 0, 0.05);
}
.svc-card-photo {
  width: 180rpx;
  height: 180rpx;
  border-radius: 16rpx;
  overflow: hidden;
  flex-shrink: 0;
  margin-right: 20rpx;
}
.svc-photo {
  width: 100%;
  height: 100%;
}
.svc-photo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64rpx;
  background: linear-gradient(135deg, #e6f7f4, #d3eee9);
}
.svc-card-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.svc-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
  line-height: 1.3;
}
.svc-desc {
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #6b7280;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}
.svc-tags {
  margin-top: 12rpx;
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}
.svc-tag {
  font-size: 20rpx;
  color: #12806f;
  background: #e6f7f4;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}
.svc-card-foot {
  margin-top: auto;
  padding-top: 14rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.svc-price .price-symbol {
  color: #e53935;
  font-size: 24rpx;
}
.svc-price .price-value {
  color: #e53935;
  font-size: 36rpx;
  font-weight: 700;
}
.svc-price .price-unit {
  font-size: 22rpx;
  color: #9ca3af;
}
.svc-btn {
  font-size: 24rpx;
  color: #fff;
  padding: 12rpx 26rpx;
  border-radius: 999rpx;
}
.btn-wellness {
  background: linear-gradient(135deg, #2bb3a3, #12806f);
}
.btn-escort {
  background: linear-gradient(135deg, #3d7ef7, #2963d6);
}

/* loading / empty */
.svc-loading {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.svc-loading-card {
  height: 200rpx;
  border-radius: 20rpx;
  background: linear-gradient(90deg, #eef1f4 25%, #f7f9fb 50%, #eef1f4 75%);
  background-size: 400% 100%;
  animation: shimmer 1.4s infinite;
}
.svc-empty {
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