<template>
  <view class="product-detail-container" v-if="product">
    <!-- 商品图片轮播 -->
    <swiper class="product-swiper" :indicator-dots="true" :autoplay="true" :interval="3000">
      <swiper-item v-for="(image, index) in product.images" :key="index">
        <image :src="image" mode="aspectFill" class="swiper-image" />
      </swiper-item>
      <swiper-item v-if="!product.images?.length">
        <image :src="BrandAsset.DEFAULT_PRODUCT_IMAGE" mode="aspectFill" class="swiper-image" />
      </swiper-item>
    </swiper>

    <!-- 商品信息 -->
    <view class="product-info">
      <view class="price-row" v-if="product.sale_type === 2">
        <text class="current-price">¥{{ product.rental_price.toFixed(2) }}/{{ getRentalUnitText(product.rental_unit) }}</text>
        <text class="deposit-tag">押金 ¥{{ product.deposit.toFixed(2) }}</text>
      </view>
      <view class="price-row" v-else>
        <text class="current-price">¥{{ selectedPrice.toFixed(2) }}</text>
        <text v-if="product.original_price > 0" class="original-price">
          ¥{{ product.original_price.toFixed(2) }}
        </text>
      </view>
      <view class="product-name">
        {{ product.name }}
        <text v-if="product.product_type === 2 || product.sale_type === 2" class="rental-badge">租赁</text>
        <text v-else-if="product.product_type === 3" class="wellness-badge">套餐</text>
        <text v-else-if="product.product_type === 4" class="escort-badge">陪诊</text>
      </view>
      <view class="product-meta">
        <text>库存: {{ selectedStock }}</text>
        <text>销量: {{ product.sales || 0 }}</text>
        <text>单位: {{ product.unit || '份' }}</text>
      </view>
      <view class="product-desc" v-if="product.description">
        {{ product.description }}
      </view>
    </view>

    <!-- 康养套餐服务内容 -->
    <view class="wellness-section" v-if="product.product_type === 3 && wellnessServices.length">
      <view class="section-title">🌿 套餐包含服务</view>
      <view class="wellness-meta" v-if="wellnessMeta.cycle || wellnessMeta.target_audience">
        <view class="wellness-meta-item" v-if="wellnessMeta.cycle">
          <text class="meta-label">套餐周期：</text>
          <text class="meta-value">{{ wellnessMeta.cycle }}</text>
        </view>
        <view class="wellness-meta-item" v-if="wellnessMeta.target_audience">
          <text class="meta-label">适用人群：</text>
          <text class="meta-value">{{ wellnessMeta.target_audience }}</text>
        </view>
      </view>
      <view class="wellness-service-list">
        <view
          v-for="(svc, idx) in wellnessServices"
          :key="idx"
          class="wellness-service-item"
        >
          <view class="svc-index">{{ idx + 1 }}</view>
          <view class="svc-content">
            <view class="svc-name">
              {{ svc.name }}
              <text class="svc-count" v-if="svc.count">
                {{ svc.count }}{{ svc.unit || '次' }}
              </text>
            </view>
            <view class="svc-desc" v-if="svc.description">{{ svc.description }}</view>
          </view>
        </view>
      </view>
      <view class="wellness-remark" v-if="wellnessMeta.remark">
        💡 {{ wellnessMeta.remark }}
      </view>
    </view>

    <!-- 陪诊服务说明 -->
    <view class="service-section" v-if="product.product_type === 4">
      <view class="section-title">🏥 服务说明</view>
      <view class="service-info-box">
        <view class="service-info-row">服务类型：专业陪诊服务</view>
        <view class="service-info-row">服务人员：经审核认证的专业陪诊师</view>
        <view class="service-info-row" v-if="product.service_content?.duration">
          服务时长：{{ product.service_content.duration }}
        </view>
        <view class="service-info-row" v-if="product.service_content?.includes">
          服务包含：{{ product.service_content.includes }}
        </view>
      </view>
    </view>

    <!-- 规格选择 -->
    <view class="spec-section" v-if="product.specs?.length">
      <view class="section-title">选择规格</view>
      <view class="spec-list">
        <view
          v-for="spec in product.specs"
          :key="spec.id"
          class="spec-group"
        >
          <view class="spec-name">{{ spec.name }}</view>
          <view class="spec-options">
            <view
              v-for="option in spec.options"
              :key="option.id"
              class="spec-option"
              :class="{
                selected: selectedSpecs[spec.name] === option.name,
                disabled: option.stock === 0
              }"
              @click="selectSpec(spec.name, option)"
            >
              <text class="option-name">{{ option.name }}</text>
              <text v-if="option.price > 0" class="option-price">+¥{{ option.price.toFixed(2) }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 租赁时长选择 -->
    <view class="rental-section" v-if="product.sale_type === 2">
      <view class="section-title">租赁时长</view>
      <view class="duration-options">
        <view
          v-for="d in availableDurations"
          :key="d"
          class="duration-option"
          :class="{ active: rentalDuration === d }"
          @click="rentalDuration = d"
        >
          {{ d }}{{ getRentalUnitText(product.rental_unit) }}
        </view>
      </view>
      <view class="rental-summary">
        <view class="rental-summary-row">
          <text>租金：¥{{ (product.rental_price * rentalDuration).toFixed(2) }}</text>
          <text>押金：¥{{ product.deposit.toFixed(2) }}</text>
        </view>
        <view class="rental-summary-total">
          合计：¥{{ (product.rental_price * rentalDuration + product.deposit).toFixed(2) }}
        </view>
      </view>
    </view>

    <!-- 购买数量 -->
    <view class="quantity-section">
      <view class="section-title">购买数量</view>
      <view class="quantity-selector">
        <view
          class="quantity-btn"
          :class="{ disabled: quantity <= 1 }"
          @click="decreaseQuantity"
        >-</view>
        <input
          v-model.number="quantity"
          type="number"
          class="quantity-input"
          @change="onQuantityChange"
        />
        <view
          class="quantity-btn"
          :class="{ disabled: quantity >= selectedStock }"
          @click="increaseQuantity"
        >+</view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <view class="action-icons">
        <view class="icon-item" @click="goHome">
          <text class="icon">🏠</text>
          <text class="icon-text">首页</text>
        </view>
        <view class="icon-item" @click="goCart">
          <text class="icon">🛒</text>
          <text class="icon-text">购物车</text>
          <view class="cart-badge" v-if="cartCount > 0">{{ cartCount }}</view>
        </view>
      </view>
      <view class="action-buttons">
        <view class="btn-add-cart" @click="addToCart">加入购物车</view>
        <view class="btn-buy-now" @click="buyNow">立即购买</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getStoreProduct } from '../../api/store'
import { useCartStore } from '../../stores/cart'
import { useAnalytics } from '@utils/analytics'
import { parseStoreProductEntryOptions } from '@utils/storeEntry'
import type { Product, SpecOption, WellnessPackageItem } from '@types'
import { BrandAsset } from '../../utils/constants'

const cartStore = useCartStore()
const { trackProductView } = useAnalytics()

const product = ref<Product | null>(null)
const quantity = ref(1)
const rentalDuration = ref(1)
const selectedSpecs = reactive<Record<string, string>>({})
const productId = ref(1)
const entrySource = ref('scan')

const cartCount = computed(() => cartStore.totalCount)

// 康养套餐：服务项列表
const wellnessServices = computed<WellnessPackageItem[]>(() => {
  if (!product.value || Number(product.value.product_type) !== 3) return []
  const sc = product.value.service_content
  if (sc && Array.isArray(sc.services)) return sc.services as WellnessPackageItem[]
  return []
})
// 康养套餐：其他元数据
const wellnessMeta = computed<{ cycle?: string; target_audience?: string; remark?: string }>(() => {
  if (!product.value || Number(product.value.product_type) !== 3) return {}
  const sc = product.value.service_content
  if (!sc) return {}
  return {
    cycle: sc.cycle || '',
    target_audience: sc.target_audience || '',
    remark: sc.remark || ''
  }
})

function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}

const availableDurations = computed(() => {
  if (!product.value || product.value.sale_type !== 2) return []
  const max = Number(product.value.max_rental_duration || 0)
  if (max > 0) {
    return Array.from({ length: max }, (_, i) => i + 1)
  }
  // 默认提供 1-12 的时长选项
  return [1, 2, 3, 5, 7, 10, 15, 30]
})

const selectedPrice = computed(() => {
  if (!product.value) return 0
  
  let price = product.value.price

  if (product.value.specs?.length) {
    for (const spec of product.value.specs) {
      const selectedOption = spec.options.find(o => o.name === selectedSpecs[spec.name])
      if (selectedOption) {
        price += selectedOption.price
      }
    }
  }

  return price
})

const selectedStock = computed(() => {
  if (!product.value) return 0
  
  if (product.value.specs?.length) {
    for (const spec of product.value.specs) {
      const selectedOption = spec.options.find(o => o.name === selectedSpecs[spec.name])
      if (selectedOption) {
        return selectedOption.stock ?? product.value.stock
      }
    }
  }
  
  return product.value.stock
})

function applyEntryOptions(options?: Record<string, any>) {
  const { productId: nextProductId, source } = parseStoreProductEntryOptions(
    options,
    productId.value
  )
  productId.value = nextProductId
  entrySource.value = source
}

onLoad((options) => {
  applyEntryOptions(options as Record<string, any> | undefined)
})

onShow(() => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  applyEntryOptions(currentPage?.options)
  loadProduct()
})

async function loadProduct() {
  try {
    product.value = await getStoreProduct(productId.value)
    await trackProductView(productId.value)
    
    // 默认选中第一个规格
    if (product.value.specs?.length) {
      for (const spec of product.value.specs) {
        if (spec.options.length > 0) {
          selectedSpecs[spec.name] = spec.options[0].name
        }
      }
    }
  } catch (error) {
    console.error('加载商品详情失败:', error)
    uni.showToast({ title: '加载失败', icon: 'none' })
  }
}

function selectSpec(specName: string, option: SpecOption) {
  if (option.stock === 0) return
  selectedSpecs[specName] = option.name
}

function decreaseQuantity() {
  if (quantity.value > 1) {
    quantity.value--
  }
}

function increaseQuantity() {
  if (quantity.value < selectedStock.value) {
    quantity.value++
  }
}

function onQuantityChange() {
  if (quantity.value < 1) quantity.value = 1
  if (quantity.value > selectedStock.value) quantity.value = selectedStock.value
}

function getSpecString(): string {
  const specs: string[] = []
  for (const spec of product.value?.specs || []) {
    if (selectedSpecs[spec.name]) {
      specs.push(selectedSpecs[spec.name])
    }
  }
  return specs.join('/')
}

function addToCart() {
  if (!product.value) return

  if (product.value.sale_type === 2 && rentalDuration.value <= 0) {
    uni.showToast({ title: '请选择租赁时长', icon: 'none' })
    return
  }

  cartStore.addItem({
    merchant_name: cartStore.merchantName || '',
    product_id: product.value.id,
    product_name: product.value.name,
    image: product.value.images?.[0] || '',
    price: selectedPrice.value,
    quantity: quantity.value,
    specs: getSpecString(),
    max_stock: selectedStock.value,
    product_type: product.value.product_type,
    sale_type: product.value.sale_type,
    rental_unit: product.value.rental_unit,
    rental_price: product.value.rental_price,
    deposit: product.value.deposit,
    rental_duration: product.value.sale_type === 2 ? rentalDuration.value : 0
  })

  uni.showToast({
    title: '已加入购物车',
    icon: 'success'
  })
}

function buyNow() {
  if (!product.value) return

  if (product.value.sale_type === 2 && rentalDuration.value <= 0) {
    uni.showToast({ title: '请选择租赁时长', icon: 'none' })
    return
  }

  const item = {
    merchant_name: cartStore.merchantName || '',
    product_id: product.value.id,
    product_name: product.value.name,
    image: product.value.images?.[0] || '',
    price: selectedPrice.value,
    quantity: quantity.value,
    specs: getSpecString(),
    max_stock: selectedStock.value,
    product_type: product.value.product_type,
    sale_type: product.value.sale_type,
    rental_unit: product.value.rental_unit,
    rental_price: product.value.rental_price,
    deposit: product.value.deposit,
    rental_duration: product.value.sale_type === 2 ? rentalDuration.value : 0
  }
  cartStore.setBuyNowItem(item)

  uni.navigateTo({
    url: `/pages/store/confirm?buy_now=1`
  })
}

function goHome() {
  uni.switchTab({ url: `/pages/store/home` })
}

function goCart() {
  uni.switchTab({
    url: `/pages/store/cart`
  })
}
</script>

<style scoped>
.product-detail-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 160rpx;
}

.product-swiper {
  height: 750rpx;
  background: #ffffff;
}

.swiper-image {
  width: 100%;
  height: 100%;
}

.product-info {
  background: #ffffff;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.price-row {
  display: flex;
  align-items: baseline;
  margin-bottom: 16rpx;
}

.current-price {
  font-size: 48rpx;
  font-weight: 700;
  color: #ff4d4f;
}

.original-price {
  font-size: 28rpx;
  color: #999999;
  text-decoration: line-through;
  margin-left: 16rpx;
}

.deposit-tag {
  font-size: 26rpx;
  color: #ff9500;
  margin-left: 16rpx;
  background: #fff7e6;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.rental-badge {
  font-size: 24rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.wellness-badge {
  font-size: 24rpx;
  color: #ffffff;
  background: #22c55e;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.escort-badge {
  font-size: 24rpx;
  color: #ffffff;
  background: #6366f1;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.info-badge {
  font-size: 24rpx;
  color: #ffffff;
  background: #64748b;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

/* 康养套餐服务内容 */
.wellness-section {
  background: #ffffff;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.wellness-meta {
  background: #f0fdf4;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  margin-bottom: 24rpx;
  border-left: 6rpx solid #22c55e;
}

.wellness-meta-item {
  font-size: 26rpx;
  color: #166534;
  margin-bottom: 8rpx;
}
.wellness-meta-item:last-child { margin-bottom: 0; }

.meta-label {
  color: #166534;
  font-weight: 500;
}
.meta-value {
  color: #14532d;
}

.wellness-service-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.wellness-service-item {
  display: flex;
  align-items: flex-start;
  padding: 20rpx;
  background: #f9fafb;
  border-radius: 12rpx;
}

.svc-index {
  width: 48rpx;
  height: 48rpx;
  border-radius: 50%;
  background: #22c55e;
  color: #ffffff;
  font-size: 24rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}

.svc-content { flex: 1; min-width: 0; }

.svc-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.svc-count {
  display: inline-block;
  margin-left: 10rpx;
  font-size: 22rpx;
  color: #16a34a;
  background: #dcfce7;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
}

.svc-desc {
  font-size: 24rpx;
  color: #666666;
  line-height: 1.5;
}

.wellness-remark {
  margin-top: 24rpx;
  padding: 16rpx 20rpx;
  background: #fffbeb;
  border-radius: 10rpx;
  font-size: 24rpx;
  color: #92400e;
  line-height: 1.6;
}

/* 陪诊服务说明 */
.service-section {
  background: #ffffff;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.service-info-box {
  background: #eef2ff;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  border-left: 6rpx solid #6366f1;
}

.service-info-row {
  font-size: 26rpx;
  color: #3730a3;
  margin-bottom: 10rpx;
  line-height: 1.6;
}
.service-info-row { margin-bottom: 0; }

.rental-section {
  background: #ffffff;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.duration-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-bottom: 24rpx;
}

.duration-option {
  padding: 16rpx 32rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  font-size: 28rpx;
  color: #1a1a1a;
}

.duration-option.active {
  background: #e6f0ff;
  border-color: #007AFF;
  color: #007AFF;
}

.rental-summary {
  background: #f9fafb;
  border-radius: 12rpx;
  padding: 24rpx;
}

.rental-summary-row {
  display: flex;
  justify-content: space-between;
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.rental-summary-total {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.product-name {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.product-meta {
  display: flex;
  gap: 32rpx;
  font-size: 26rpx;
  color: #999999;
  margin-bottom: 16rpx;
}

.product-desc {
  font-size: 28rpx;
  color: #666666;
  line-height: 1.6;
}

.spec-section, .quantity-section {
  background: #ffffff;
  padding: 32rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.spec-list {
  display: flex;
  flex-direction: column;
}

.spec-group {
  margin-bottom: 24rpx;
}

.spec-group:last-child {
  margin-bottom: 0;
}

.spec-name {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 16rpx;
}

.spec-options {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.spec-option {
  padding: 16rpx 32rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.spec-option.selected {
  background: #e6f0ff;
  border-color: #007AFF;
}

.spec-option.disabled {
  opacity: 0.5;
}

.option-name {
  font-size: 28rpx;
  color: #1a1a1a;
}

.option-price {
  font-size: 24rpx;
  color: #666666;
  margin-top: 4rpx;
}

.quantity-selector {
  display: flex;
  align-items: center;
}

.quantity-btn {
  width: 64rpx;
  height: 64rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  color: #666666;
}

.quantity-btn.disabled {
  opacity: 0.5;
}

.quantity-input {
  width: 120rpx;
  height: 64rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
  text-align: center;
  font-size: 32rpx;
  margin: 0 16rpx;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 120rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  padding: 0 32rpx;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
  z-index: 100;
}

.action-icons {
  display: flex;
  gap: 32rpx;
  margin-right: 24rpx;
}

.icon-item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.icon {
  font-size: 40rpx;
}

.icon-text {
  font-size: 22rpx;
  color: #666666;
  margin-top: 4rpx;
}

.cart-badge {
  position: absolute;
  top: -10rpx;
  right: -16rpx;
  min-width: 32rpx;
  height: 32rpx;
  background: #ff4d4f;
  color: #ffffff;
  border-radius: 16rpx;
  font-size: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 6rpx;
}

.action-buttons {
  flex: 1;
  display: flex;
  gap: 16rpx;
}

.btn-add-cart, .btn-buy-now {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 500;
}

.btn-add-cart {
  background: linear-gradient(135deg, #ff9500 0%, #ff5e3a 100%);
  color: #ffffff;
}

.btn-buy-now {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}
</style>
