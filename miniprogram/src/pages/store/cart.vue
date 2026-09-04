<template>
  <view class="cart-container">
    <!-- 店铺信息 -->
    <view class="store-info" v-if="cartStore.merchantName">
      <text class="store-name">{{ cartStore.merchantName }}</text>
    </view>

    <!-- 购物车列表 -->
    <view class="cart-list" v-if="cartStore.items.length > 0">
      <view
        v-for="item in cartStore.items"
        :key="`${item.product_id}-${item.specs}-${item.rental_duration || 0}`"
        class="cart-item"
      >
        <view class="item-checkbox" @click="toggleSelect(item)">
          <checkbox :checked="selectedItems.has(`${item.product_id}-${item.specs}-${item.rental_duration || 0}`)" />
        </view>
        <image
          class="item-image"
          :src="item.image || '/static/default-product.png'"
          mode="aspectFill"
        />
        <view class="item-info">
          <view class="item-name">
            {{ item.product_name }}
            <text v-if="Number(item.product_type) === 2 || Number(item.sale_type) === 2" class="rental-badge">租赁</text>
            <text v-else-if="Number(item.product_type) === 3" class="wellness-badge">套餐</text>
            <text v-else-if="Number(item.product_type) === 4" class="escort-badge">陪诊</text>
          </view>
          <view class="item-spec" v-if="item.specs">{{ item.specs }}</view>
          <view class="item-rental" v-if="Number(item.sale_type) === 2">
            <text class="rental-text">租金 ¥{{ Number(item.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(item.rental_unit) }} × {{ item.rental_duration || 0 }}{{ getRentalUnitText(item.rental_unit) }}</text>
            <text class="rental-deposit">押金 ¥{{ Number(item.deposit || 0).toFixed(2) }}</text>
          </view>
          <view class="item-bottom">
            <text class="item-price">¥{{ getItemPayAmount(item).toFixed(2) }}</text>
            <view class="quantity-control">
              <view
                class="quantity-btn"
                @click="decreaseQuantity(item)"
              >-</view>
              <text class="quantity-value">{{ item.quantity }}</text>
              <view
                class="quantity-btn"
                @click="increaseQuantity(item)"
              >+</view>
            </view>
          </view>
        </view>
        <view class="delete-btn" @click="removeItem(item)">×</view>
      </view>
    </view>

    <!-- 空购物车 -->
    <template v-else>
      <view class="empty-cart">
        <text class="empty-icon">🛒</text>
        <text class="empty-text">购物车是空的</text>
        <button class="btn-shopping" @click="goShopping">去逛逛</button>
      </view>

      <!-- 购物车为空时填充首页推荐 -->
      <view class="recommend-module">
        <view class="recommend-header">✨ 首页推荐</view>

        <view v-if="recommendLoading" class="recommend-skeleton-list">
          <view v-for="item in 2" :key="item" class="recommend-skeleton-card">
            <view class="recommend-skeleton-image"></view>
            <view class="recommend-skeleton-line"></view>
            <view class="recommend-skeleton-line short"></view>
          </view>
        </view>

        <view v-else-if="!recommends.length" class="recommend-empty">暂无可推荐商品</view>

        <view v-else class="recommend-list">
          <view
            v-for="item in recommends"
            :key="item.id"
            class="recommend-card"
            @click="goRecommendItem(item)"
          >
            <image class="recommend-image" :src="getRecommendImage(item)" mode="aspectFill" />
            <view class="recommend-info">
              <view class="recommend-name">
                {{ getRecommendTitle(item) }}
                <text
                  v-if="getRecommendTag(item)"
                  :class="getRecommendTagClass(item)"
                >{{ getRecommendTag(item) }}</text>
              </view>
              <view class="recommend-price">
                <template v-if="Number(item.product?.sale_type) === 2">
                  ¥{{ Number(item.product?.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(item.product?.rental_unit) }}
                </template>
                <template v-else>
                  ¥{{ (item.product?.price || 0).toFixed(2) }}
                </template>
              </view>
            </view>
          </view>
        </view>
      </view>
    </template>

    <!-- 底部结算栏 -->
    <view class="bottom-bar" v-if="cartStore.items.length > 0">
      <view class="select-all" @click="toggleSelectAll">
        <checkbox :checked="isAllSelected" />
        <text class="select-text">全选</text>
      </view>
      <view class="total-info">
        <text class="total-label">合计:</text>
        <text class="total-amount">¥{{ selectedAmount.toFixed(2) }}</text>
        <text v-if="selectedDeposit > 0" class="total-deposit">(含押金 ¥{{ selectedDeposit.toFixed(2) }})</text>
      </view>
      <view class="checkout-btn" @click="goCheckout">
        去结算 ({{ selectedCount }})
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useCartStore, getItemPayAmount, getItemDeposit, getRentalUnitText } from '../../stores/cart'
import type { CartItem } from '../../stores/cart'
import { getStoreHomeRecommends } from '../../api/store'
import type { StoreRecommendItem } from '../../api/store'
import { BrandAsset } from '../../utils/constants'

const cartStore = useCartStore()
const selectedItems = ref<Set<string>>(new Set())

onShow(() => {
  cartStore.restoreFromStorage()

  // 默认全选
  selectAllItems()

  // 购物车为空时加载首页推荐
  if (!cartStore.items.length) {
    loadRecommends()
  }
})

function getItemKey(item: CartItem): string {
  return `${item.product_id}-${item.specs}-${item.rental_duration || 0}`
}

function toggleSelect(item: CartItem) {
  const key = getItemKey(item)
  if (selectedItems.value.has(key)) {
    selectedItems.value.delete(key)
  } else {
    selectedItems.value.add(key)
  }
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedItems.value.clear()
  } else {
    selectAllItems()
  }
}

function selectAllItems() {
  selectedItems.value.clear()
  cartStore.items.forEach(item => {
    selectedItems.value.add(getItemKey(item))
  })
}

const isAllSelected = computed(() => {
  return cartStore.items.length > 0 && selectedItems.value.size === cartStore.items.length
})

const selectedCount = computed(() => {
  return cartStore.items.filter(item => selectedItems.value.has(getItemKey(item))).length
})

const selectedAmount = computed(() => {
  return cartStore.items
    .filter(item => selectedItems.value.has(getItemKey(item)))
    .reduce((sum, item) => sum + getItemPayAmount(item), 0)
})

const selectedDeposit = computed(() => {
  return cartStore.items
    .filter(item => selectedItems.value.has(getItemKey(item)))
    .reduce((sum, item) => sum + getItemDeposit(item), 0)
})

function decreaseQuantity(item: CartItem) {
  if (item.quantity > 1) {
    cartStore.updateQuantity(item.product_id, item.specs, item.quantity - 1)
  }
}

function increaseQuantity(item: CartItem) {
  if (item.max_stock && item.quantity < item.max_stock) {
    cartStore.updateQuantity(item.product_id, item.specs, item.quantity + 1)
  } else if (!item.max_stock) {
    cartStore.updateQuantity(item.product_id, item.specs, item.quantity + 1)
  }
}

function removeItem(item: CartItem) {
  uni.showModal({
    title: '确认删除',
    content: '确定要从购物车中移除该商品吗？',
    success: (res) => {
      if (res.confirm) {
        cartStore.removeItem(item.product_id, item.specs)
        selectedItems.value.delete(getItemKey(item))
      }
    }
  })
}

function goShopping() {
  uni.switchTab({
    url: `/pages/store/home`
  })
}

/* ============ 购物车为空时展示首页推荐 ============ */
const recommends = ref<StoreRecommendItem[]>([])
const recommendLoading = ref(false)

async function loadRecommends() {
  recommendLoading.value = true
  try {
    const res = await getStoreHomeRecommends()
    recommends.value = res.list || []
  } catch (_e) {
    recommends.value = []
  } finally {
    recommendLoading.value = false
  }
}

function getRecommendImage(item: StoreRecommendItem): string {
  return item.product?.images?.[0] || item.product?.image || BrandAsset.DEFAULT_PRODUCT_IMAGE
}

function getRecommendTitle(item: StoreRecommendItem): string {
  return item.title || item.product?.name || ''
}

function getRecommendTag(item: StoreRecommendItem): string {
  const targetType = Number(item.target_type || 0)
  const productType = Number(item.product?.product_type || 0)
  const saleType = Number(item.product?.sale_type || 0)
  if (targetType === 2 || productType === 3) return '套餐'
  if (productType === 4) return '陪诊'
  if (saleType === 2) return '租赁'
  return ''
}

function getRecommendTagClass(item: StoreRecommendItem): string {
  const tag = getRecommendTag(item)
  if (tag === '租赁') return 'product-rental-tag'
  if (tag === '套餐') return 'product-wellness-tag'
  if (tag === '陪诊') return 'product-escort-tag'
  return ''
}

function goProductDetail(productId: number) {
  uni.navigateTo({ url: `/pages/store/product?product_id=${productId}` })
}

function goRecommendItem(item: StoreRecommendItem) {
  const productId = item.product_id || item.product?.id
  if (!productId) {
    uni.showToast({ title: '该推荐商品暂不可用', icon: 'none' })
    return
  }
  goProductDetail(productId)
}

function goCheckout() {
  if (selectedCount.value === 0) {
    return uni.showToast({ title: '请选择商品', icon: 'none' })
  }
  
  uni.navigateTo({
    url: `/pages/store/confirm`
  })
}
</script>

<style scoped>
.cart-container {
  min-height: 100vh;
  background: #f5f5f5;
  /* 为底部固定的全选结算栏（100rpx）预留空间；tabBar 页面无需额外叠加安全区 */
  padding-bottom: calc(100rpx + 32rpx);
  box-sizing: border-box;
}

.store-info {
  background: #ffffff;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.store-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.cart-list {
  padding: 24rpx;
}

.cart-item {
  display: flex;
  align-items: center;
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  position: relative;
}

.item-checkbox {
  margin-right: 16rpx;
}

.item-image {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.item-info {
  flex: 1;
}

.item-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  margin-bottom: 8rpx;
}

.item-spec {
  font-size: 26rpx;
  color: #999999;
  margin-bottom: 16rpx;
}

.rental-badge {
  font-size: 22rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.wellness-badge {
  font-size: 22rpx;
  color: #ffffff;
  background: #22c55e;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.escort-badge {
  font-size: 22rpx;
  color: #ffffff;
  background: #6366f1;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.info-badge {
  font-size: 22rpx;
  color: #ffffff;
  background: #64748b;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 12rpx;
  vertical-align: middle;
}

.item-rental {
  display: flex;
  flex-direction: column;
  font-size: 24rpx;
  color: #ff9500;
  margin-bottom: 12rpx;
  background: #fff7e6;
  padding: 8rpx 12rpx;
  border-radius: 8rpx;
}

.rental-text {
  color: #ff9500;
}

.rental-deposit {
  color: #666666;
  margin-top: 4rpx;
}

.item-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-price {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.quantity-control {
  display: flex;
  align-items: center;
}

.quantity-btn {
  width: 48rpx;
  height: 48rpx;
  background: #f5f5f5;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  color: #666666;
}

.quantity-value {
  width: 64rpx;
  text-align: center;
  font-size: 28rpx;
  color: #1a1a1a;
}

.delete-btn {
  position: absolute;
  top: 16rpx;
  right: 16rpx;
  width: 40rpx;
  height: 40rpx;
  background: #f5f5f5;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #999999;
}

.empty-cart {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 200rpx 0;
}

.empty-icon {
  font-size: 160rpx;
  margin-bottom: 32rpx;
}

.empty-text {
  font-size: 30rpx;
  color: #999999;
  margin-bottom: 48rpx;
}

.btn-shopping {
  padding: 24rpx 64rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 30rpx;
}

/* 购物车为空时展示首页推荐 */
.recommend-module {
  margin: 0 24rpx 20rpx;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 24rpx 32rpx;
}

.recommend-header {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20rpx;
}

.recommend-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
}

.recommend-card {
  background: #ffffff;
  border: 1rpx solid #f0f0f0;
  border-radius: 16rpx;
  overflow: hidden;
  box-shadow: 0 4rpx 12rpx rgba(15, 23, 42, 0.05);
}

.recommend-image {
  width: 100%;
  height: 300rpx;
  background: #e0e0e0;
}

.recommend-info {
  padding: 16rpx;
}

.recommend-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 12rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.recommend-price {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.recommend-empty {
  padding: 48rpx 24rpx;
  text-align: center;
  font-size: 26rpx;
  color: #999999;
}

.recommend-skeleton-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
}

.recommend-skeleton-card {
  background: #ffffff;
  border: 1rpx solid #f0f0f0;
  border-radius: 16rpx;
  overflow: hidden;
}

.recommend-skeleton-image {
  width: 100%;
  height: 300rpx;
  background: linear-gradient(90deg, #f2f3f5 0%, #e9ecef 50%, #f2f3f5 100%);
  background-size: 200% 100%;
  animation: loadingShimmer 1.2s linear infinite;
}

.recommend-skeleton-line {
  height: 24rpx;
  border-radius: 12rpx;
  background: linear-gradient(90deg, #f2f3f5 0%, #e9ecef 50%, #f2f3f5 100%);
  background-size: 200% 100%;
  animation: loadingShimmer 1.2s linear infinite;
  margin: 16rpx 16rpx 0;
}

.recommend-skeleton-line.short {
  width: 40%;
  margin-bottom: 24rpx;
}

@keyframes loadingShimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.product-rental-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.product-wellness-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #22c55e;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.product-escort-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #6366f1;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.bottom-bar {
  position: fixed;
  /* 购物车是 tabBar 页面：fixed 定位参照的视口底部就是 tabBar 顶部，
     bottom: 0 即贴住 tabBar 上方；再叠加偏移反而会悬空漂浮 */
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  padding: 0 32rpx;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
  z-index: 99;
}

.select-all {
  display: flex;
  align-items: center;
  margin-right: 24rpx;
}

.select-text {
  font-size: 28rpx;
  color: #666666;
  margin-left: 8rpx;
}

.total-info {
  flex: 1;
  display: flex;
  align-items: baseline;
}

.total-label {
  font-size: 28rpx;
  color: #666666;
}

.total-amount {
  font-size: 40rpx;
  font-weight: 600;
  color: #ff4d4f;
  margin-left: 8rpx;
}

.total-deposit {
  font-size: 22rpx;
  color: #ff9500;
  margin-left: 8rpx;
}

.checkout-btn {
  padding: 24rpx 48rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 44rpx;
  font-size: 30rpx;
  font-weight: 500;
  color: #ffffff;
}
</style>
