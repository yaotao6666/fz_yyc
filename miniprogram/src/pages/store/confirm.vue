<template>
  <view class="confirm-container">
    <!-- 收货信息 -->
    <view class="section delivery-section">
      <view class="section-title">收货信息</view>
      
      <view class="delivery-type-selector">
        <view
          v-for="type in deliveryTypes"
          :key="type.value"
          class="type-item"
          :class="{ selected: deliveryType === type.value }"
          @click="selectDeliveryType(type.value)"
        >
          <text class="type-icon">{{ type.icon }}</text>
          <text class="type-name">{{ type.name }}</text>
        </view>
      </view>

      <view class="delivery-form" v-if="deliveryType === 1">
        <view class="form-item">
          <view class="form-label">收货地址</view>
          <input
            v-model="deliveryAddress"
            class="form-input"
            placeholder="请输入详细收货地址"
          />
        </view>
        <view class="form-item">
          <view class="form-label">联系人</view>
          <input
            v-model="contactName"
            class="form-input"
            placeholder="请输入联系人姓名"
          />
        </view>
        <view class="form-item">
          <view class="form-label">联系电话</view>
          <input
            v-model="contactPhone"
            class="form-input"
            type="number"
            placeholder="请输入联系电话"
          />
        </view>
        <view class="form-item">
          <view class="form-label">配送距离</view>
          <input
            v-model.number="deliveryDistance"
            class="form-input"
            type="digit"
            placeholder="请输入配送距离（公里）"
          />
        </view>
      </view>

      <view class="remark-form">
        <view class="form-label">备注</view>
        <input
          v-model="remark"
          class="form-input"
          placeholder="口味、偏好等要求（选填）"
        />
      </view>
    </view>

    <!-- 商品信息 -->
    <view class="section goods-section">
      <view class="section-title">商品信息</view>
      <view class="goods-list">
        <view
          v-for="item in cartStore.items"
          :key="`${item.product_id}-${item.specs}`"
          class="goods-item"
        >
          <image
            class="goods-image"
            :src="item.image || '/static/default-product.png'"
            mode="aspectFill"
          />
          <view class="goods-info">
            <view class="goods-name">{{ item.product_name }}</view>
            <view class="goods-spec" v-if="item.specs">{{ item.specs }}</view>
          </view>
          <view class="goods-right">
            <view class="goods-price">¥{{ item.price.toFixed(2) }}</view>
            <view class="goods-quantity">x{{ item.quantity }}</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 金额明细 -->
    <view class="section amount-section">
      <view class="amount-row">
        <text class="amount-label">商品金额</text>
        <text class="amount-value">¥{{ cartStore.totalAmount.toFixed(2) }}</text>
      </view>
      <view class="amount-row">
        <text class="amount-label">配送费</text>
        <text class="amount-value">¥{{ deliveryFee.toFixed(2) }}</text>
      </view>
      <view class="amount-row total">
        <text class="amount-label">合计</text>
        <text class="amount-value">¥{{ totalAmount.toFixed(2) }}</text>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <view class="total-info">
        <text class="total-label">实付金额</text>
        <text class="total-amount">¥{{ totalAmount.toFixed(2) }}</text>
      </view>
      <view class="submit-btn" @click="submitOrder">
        提交订单
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onShow } from 'vue'
import { createOrder } from '../../api'
import { useCartStore } from '../../stores/cart'
import type { CreateOrderRequest } from '../../types/api'

const cartStore = useCartStore()

const deliveryTypes = [
  { value: 1, name: '配送', icon: '🚚' },
  { value: 2, name: '堂食', icon: '🍽️' },
  { value: 3, name: '自提', icon: '📦' }
]

const deliveryType = ref(1)
const deliveryAddress = ref('')
const contactName = ref('')
const contactPhone = ref('')
const deliveryDistance = ref(0)
const remark = ref('')
const merchantId = ref(1)

onShow(() => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  merchantId.value = Number(currentPage?.options?.merchant_id) || 1
})

function selectDeliveryType(type: number) {
  deliveryType.value = type
}

const deliveryFee = computed(() => {
  if (deliveryType.value !== 1) return 0
  
  if (deliveryDistance.value <= 2) return 0
  if (deliveryDistance.value <= 5) return 3
  if (deliveryDistance.value <= 10) return 6
  return 10
})

const totalAmount = computed(() => {
  return cartStore.totalAmount + deliveryFee.value
})

async function submitOrder() {
  if (deliveryType.value === 1) {
    if (!deliveryAddress.value) {
      return uni.showToast({ title: '请输入收货地址', icon: 'none' })
    }
    if (!contactName.value) {
      return uni.showToast({ title: '请输入联系人', icon: 'none' })
    }
    if (!contactPhone.value) {
      return uni.showToast({ title: '请输入联系电话', icon: 'none' })
    }
  }

  if (cartStore.isEmpty) {
    return uni.showToast({ title: '购物车为空', icon: 'none' })
  }

  const orderData: CreateOrderRequest = {
    items: cartStore.items.map(item => ({
      product_id: item.product_id,
      spec_option: item.specs,
      quantity: item.quantity
    })),
    delivery_type: deliveryType.value,
    delivery_distance: deliveryDistance.value,
    remark: remark.value
  }

  try {
    uni.showLoading({ title: '创建订单中...' })

    const res = await createOrder(merchantId.value, orderData)

    // 清除购物车
    cartStore.clearCart()

    // 发起微信支付
    if (res.pay_params) {
      uni.requestPayment({
        provider: 'wxpay',
        timeStamp: res.pay_params.timeStamp,
        nonceStr: res.pay_params.nonceStr,
        package: res.pay_params.package,
        signType: res.pay_params.signType,
        paySign: res.pay_params.paySign,
        success: () => {
          uni.hideLoading()
          uni.showToast({ title: '支付成功', icon: 'success' })
          
          setTimeout(() => {
            uni.redirectTo({
              url: `/pages/store/my-orders?status=paid`
            })
          }, 1500)
        },
        fail: (err) => {
          uni.hideLoading()
          if (err.errMsg?.includes('cancel')) {
            uni.showToast({ title: '支付已取消', icon: 'none' })
            
            setTimeout(() => {
              uni.redirectTo({
                url: `/pages/store/my-orders?status=pending`
              })
            }, 1500)
          } else {
            uni.showToast({ title: '支付失败', icon: 'none' })
          }
        }
      })
    } else {
      // 无需支付，直接跳转订单页
      uni.hideLoading()
      uni.showToast({ title: '订单创建成功', icon: 'success' })
      
      setTimeout(() => {
        uni.redirectTo({
          url: `/pages/store/my-orders`
        })
      }, 1500)
    }
  } catch (error: any) {
    uni.hideLoading()
    uni.showToast({ title: error.message || '创建订单失败', icon: 'none' })
  }
}
</script>

<style scoped>
.confirm-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.delivery-type-selector {
  display: flex;
  gap: 24rpx;
  margin-bottom: 24rpx;
}

.type-item {
  flex: 1;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 2rpx solid transparent;
}

.type-item.selected {
  background: #e6f0ff;
  border-color: #007AFF;
}

.type-icon {
  font-size: 48rpx;
  margin-bottom: 8rpx;
}

.type-name {
  font-size: 28rpx;
  color: #1a1a1a;
}

.delivery-form {
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 20rpx;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.form-input {
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
}

.remark-form {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
}

.goods-list {
  display: flex;
  flex-direction: column;
}

.goods-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.goods-item:last-child {
  border-bottom: none;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.goods-info {
  flex: 1;
}

.goods-name {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.goods-spec {
  font-size: 26rpx;
  color: #999999;
}

.goods-right {
  text-align: right;
}

.goods-price {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.goods-quantity {
  font-size: 26rpx;
  color: #999999;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  padding: 16rpx 0;
  font-size: 28rpx;
}

.amount-label {
  color: #666666;
}

.amount-value {
  color: #1a1a1a;
}

.amount-row.total {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
  margin-top: 8rpx;
}

.amount-row.total .amount-label {
  font-weight: 600;
  color: #1a1a1a;
}

.amount-row.total .amount-value {
  font-size: 36rpx;
  font-weight: 600;
  color: #ff4d4f;
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
}

.total-info {
  flex: 1;
}

.total-label {
  font-size: 26rpx;
  color: #666666;
  margin-right: 8rpx;
}

.total-amount {
  font-size: 40rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.submit-btn {
  padding: 24rpx 64rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #ffffff;
}
</style>
