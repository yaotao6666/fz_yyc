<template>
  <view class="store-tabbar">
    <view
      class="tab-item"
      :class="{ active: activeTab === 'cart' }"
      @click.stop="switchTab('cart')"
    >
      <text class="tab-icon">🛒</text>
      <text class="tab-text">购物车</text>
      <view class="badge" v-if="cartCount > 0">{{ cartCount > 99 ? '99+' : cartCount }}</view>
    </view>

    <view class="divider"></view>

    <view
      class="tab-item"
      :class="{ active: activeTab === 'orders' }"
      @click.stop="switchTab('orders')"
    >
      <text class="tab-icon">📋</text>
      <text class="tab-text">我的订单</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCartStore } from '@/stores/cart'

const props = defineProps<{
  activeTab?: 'cart' | 'orders'
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', value: 'cart' | 'orders'): void
}>()

const cartStore = useCartStore()
const cartCount = computed(() => cartStore.totalCount)

function switchTab(tab: 'cart' | 'orders') {
  emit('update:activeTab', tab)
}
</script>

<style scoped>
.store-tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-around;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 100;
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  padding: 10rpx 0;
  transition: all 0.3s ease;
}

.tab-item.active .tab-icon {
  transform: scale(1.1);
}

.tab-item.active .tab-text {
  color: #007AFF;
  font-weight: 600;
}

.tab-icon {
  font-size: 44rpx;
  margin-bottom: 4rpx;
  transition: transform 0.3s ease;
}

.tab-text {
  font-size: 24rpx;
  color: #666666;
  transition: all 0.3s ease;
}

.badge {
  position: absolute;
  top: 0;
  right: 25%;
  min-width: 32rpx;
  height: 32rpx;
  background: #ff4d4f;
  color: #ffffff;
  border-radius: 16rpx;
  font-size: 20rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8rpx;
  box-shadow: 0 2rpx 4rpx rgba(255, 77, 79, 0.3);
}

.divider {
  width: 1rpx;
  height: 60rpx;
  background: linear-gradient(to bottom, transparent, #e5e5e5, transparent);
}
</style>
