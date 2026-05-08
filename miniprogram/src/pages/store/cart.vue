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
        :key="`${item.product_id}-${item.specs}`"
        class="cart-item"
      >
        <view class="item-checkbox" @click="toggleSelect(item)">
          <checkbox :checked="selectedItems.has(`${item.product_id}-${item.specs}`)" />
        </view>
        <image
          class="item-image"
          :src="item.image || '/static/default-product.png'"
          mode="aspectFill"
        />
        <view class="item-info">
          <view class="item-name">{{ item.product_name }}</view>
          <view class="item-spec" v-if="item.specs">{{ item.specs }}</view>
          <view class="item-bottom">
            <text class="item-price">¥{{ item.price.toFixed(2) }}</text>
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
    <view class="empty-cart" v-else>
      <text class="empty-icon">🛒</text>
      <text class="empty-text">购物车是空的</text>
      <button class="btn-shopping" @click="goShopping">去逛逛</button>
    </view>

    <!-- 底部结算栏 -->
    <view class="bottom-bar" v-if="cartStore.items.length > 0">
      <view class="select-all" @click="toggleSelectAll">
        <checkbox :checked="isAllSelected" />
        <text class="select-text">全选</text>
      </view>
      <view class="total-info">
        <text class="total-label">合计:</text>
        <text class="total-amount">¥{{ selectedAmount.toFixed(2) }}</text>
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
import { useCartStore } from '../../stores/cart'
import type { CartItem } from '../../stores/cart'

const cartStore = useCartStore()
const selectedItems = ref<Set<string>>(new Set())
const merchantId = ref(1)

onShow(() => {
  cartStore.restoreFromStorage()
  
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  merchantId.value = Number(currentPage?.options?.merchant_id) || 1
  
  // 默认全选
  selectAllItems()
})

function getItemKey(item: CartItem): string {
  return `${item.product_id}-${item.specs}`
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
    .reduce((sum, item) => sum + item.price * item.quantity, 0)
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
  uni.navigateTo({
    url: `/pages/store/home?merchant_id=${merchantId.value}`
  })
}

function goCheckout() {
  if (selectedCount.value === 0) {
    return uni.showToast({ title: '请选择商品', icon: 'none' })
  }
  
  uni.navigateTo({
    url: `/pages/store/confirm?merchant_id=${merchantId.value}`
  })
}
</script>

<style scoped>
.cart-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
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

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  padding: 0 32rpx;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
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

.checkout-btn {
  padding: 24rpx 48rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 44rpx;
  font-size: 30rpx;
  font-weight: 500;
  color: #ffffff;
}
</style>
