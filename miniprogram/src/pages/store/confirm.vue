<template>
  <view class="confirm-container">
    <!-- 收货信息 -->
    <view class="section delivery-section">
      <view class="section-title">收货信息</view>
      
      <view class="delivery-type-selector">
        <view
          v-for="type in availableDeliveryTypes"
          :key="type.value"
          class="type-item"
          :class="{ selected: deliveryType === type.value }"
          @click="selectDeliveryType(type.value)"
        >
          <text class="type-icon">{{ type.icon }}</text>
          <text class="type-name">{{ type.name }}</text>
        </view>
      </view>
      <view v-if="availableDeliveryTypes.length === 0" class="delivery-mode-empty">
        商家暂未开放下单方式
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
          <picker
            v-if="deliveryRules.length > 0"
            mode="selector"
            :range="deliveryRules"
            range-key="label"
            :value="deliveryDistanceIndex"
            @change="onDistanceChange"
          >
            <view class="picker-value">
              {{ deliveryRules[deliveryDistanceIndex]?.label || '请选择距离' }}
            </view>
          </picker>
          <view v-else class="picker-value disabled">
            当前暂无可选配送档位
          </view>
          <view class="distance-tip" v-if="deliveryRules.length > 0">
            当前距离对应费用由商家设置，您需自主选择距离，超出可能商家拒绝配送
          </view>
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
            :src="item.image || BrandAsset.DEFAULT_PRODUCT_IMAGE"
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
      <view class="submit-btn" :class="{ disabled: submitting }" @click="submitOrder">
        {{ submitting ? '处理中...' : '提交订单' }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { createOrder, getStoreDeliveryRules } from '../../api/store'
import { useCartStore } from '../../stores/cart'
import { useAnalytics } from '@utils/analytics'
import { useAuth } from '../../utils/useAuth'
import { parseStoreEntryOptions } from '@utils/storeEntry'
import type { CreateOrderRequest, StoreDeliveryRules } from '@types'
import { BrandAsset } from '../../utils/constants'

const cartStore = useCartStore()
const { trackPageView, trackPayment } = useAnalytics()

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
const deliveryDistanceIndex = ref(0)
const remark = ref('')
const merchantId = ref(1)
const entrySource = ref('scan')
const submitting = ref(false)
const deliveryConfig = ref<StoreDeliveryRules>({
  enabled: false,
  base_fee: 0,
  free_delivery_amount: 0,
  max_distance: 0,
  distance_rules: [] as { min_distance: number; max_distance: number; fee: number }[],
  takeout_enabled: false,
  dine_in_enabled: false,
  pickup_enabled: false
})

// 配送档位选项
const deliveryRules = ref<{ distance: number; fee: number; label: string }[]>([])

function applyEntryOptions(options?: Record<string, any>) {
  const { merchantId: nextMerchantId, source } = parseStoreEntryOptions(options, merchantId.value)
  merchantId.value = nextMerchantId
  entrySource.value = source
}

onLoad((options) => {
  applyEntryOptions(options as Record<string, any> | undefined)
})

onShow(async () => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  applyEntryOptions(currentPage?.options)

  // 配送规则属于公开店铺数据，不应被登录流程阻塞。
  await loadDeliveryRules()

  const { ensureAuth } = useAuth()
  void ensureAuth()
  void trackPageView('store_confirm', merchantId.value, entrySource.value)
})

async function loadDeliveryRules() {
  try {
    const rules = await getStoreDeliveryRules(merchantId.value)
    deliveryConfig.value = rules
    deliveryRules.value = (rules.distance_rules || []).map((item) => ({
      distance: item.max_distance,
      fee: item.fee,
      label: `${item.min_distance}-${item.max_distance}km · 配送费 ¥${item.fee.toFixed(2)}`
    }))

    if (deliveryRules.value.length === 0 && rules.max_distance > 0) {
      deliveryRules.value = [{
        distance: rules.max_distance,
        fee: rules.base_fee,
        label: `0-${rules.max_distance}km · 配送费 ¥${rules.base_fee.toFixed(2)}`
      }]
    }

    if (deliveryRules.value.length > 0) {
      deliveryDistanceIndex.value = 0
      deliveryDistance.value = deliveryRules.value[0].distance
    }

    if (availableDeliveryTypes.value.length > 0 && !availableDeliveryTypes.value.some(type => type.value === deliveryType.value)) {
      deliveryType.value = availableDeliveryTypes.value[0].value
    }
  } catch (error) {
    console.error('获取配送规则失败:', error)
    deliveryRules.value = []
  }
}

function selectDeliveryType(type: number) {
  if (!availableDeliveryTypes.value.some(item => item.value === type)) {
    return
  }
  deliveryType.value = type
}

function onDistanceChange(e: any) {
  deliveryDistanceIndex.value = e.detail.value
  deliveryDistance.value = deliveryRules.value[e.detail.value].distance
}

const selectedDeliveryRule = computed(() => deliveryRules.value[deliveryDistanceIndex.value] || null)
const availableDeliveryTypes = computed(() => {
  return deliveryTypes.filter((type) => {
    if (type.value === 1) {
      return deliveryConfig.value.takeout_enabled
    }
    if (type.value === 2) {
      return deliveryConfig.value.dine_in_enabled
    }
    if (type.value === 3) {
      return deliveryConfig.value.pickup_enabled
    }
    return false
  })
})

const deliveryFee = computed(() => {
  if (deliveryType.value !== 1) return 0

  if (cartStore.totalAmount >= deliveryConfig.value.free_delivery_amount && deliveryConfig.value.free_delivery_amount > 0) {
    return 0
  }

  if (!selectedDeliveryRule.value) {
    return deliveryConfig.value.base_fee || 0
  }

  return selectedDeliveryRule.value.fee
})

const totalAmount = computed(() => {
  return cartStore.totalAmount + deliveryFee.value
})

async function submitOrder() {
  if (submitting.value) {
    return
  }

  if (!availableDeliveryTypes.value.length) {
    return uni.showToast({ title: '商家暂未开放下单方式', icon: 'none' })
  }

  if (!availableDeliveryTypes.value.some(type => type.value === deliveryType.value)) {
    deliveryType.value = availableDeliveryTypes.value[0].value
    return uni.showToast({ title: '当前下单方式不可用', icon: 'none' })
  }

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    return uni.showToast({ title: '登录失败，请重试', icon: 'none' })
  }

  if (deliveryType.value === 1) {
    if (!deliveryConfig.value.takeout_enabled) {
      return uni.showToast({ title: '商家暂未开启配送', icon: 'none' })
    }
    if (!deliveryAddress.value) {
      return uni.showToast({ title: '请输入收货地址', icon: 'none' })
    }
    if (!contactName.value) {
      return uni.showToast({ title: '请输入联系人', icon: 'none' })
    }
    if (!contactPhone.value) {
      return uni.showToast({ title: '请输入联系电话', icon: 'none' })
    }
    if (!deliveryRules.value.length) {
      return uni.showToast({ title: '当前暂无可选配送档位', icon: 'none' })
    }
    if (!selectedDeliveryRule.value) {
      return uni.showToast({ title: '请选择配送距离档位', icon: 'none' })
    }
    if (deliveryDistance.value > deliveryConfig.value.max_distance) {
      return uni.showToast({ title: '已超出商家配送范围', icon: 'none' })
    }
    if (!/^1\d{10}$/.test(contactPhone.value)) {
      return uni.showToast({ title: '请输入正确的联系电话', icon: 'none' })
    }
  }

  if (cartStore.isEmpty) {
    return uni.showToast({ title: '购物车为空', icon: 'none' })
  }

  const orderData: CreateOrderRequest = {
    items: cartStore.items.map(item => ({
      product_id: item.product_id,
      spec_info: item.specs,
      quantity: item.quantity
    })),
    delivery_type: deliveryType.value,
    remark: remark.value
  }

  if (deliveryType.value === 1) {
    orderData.delivery_distance = deliveryDistance.value
    orderData.delivery_address = deliveryAddress.value
    orderData.contact_name = contactName.value
    orderData.contact_phone = contactPhone.value
  }

  try {
    submitting.value = true
    uni.showLoading({ title: '创建订单中...' })

    const res = await createOrder(merchantId.value, orderData)

    // 发起微信支付
    if (res.pay_params) {
      uni.requestPayment({
        provider: 'wxpay',
        timeStamp: res.pay_params.timeStamp,
        nonceStr: res.pay_params.nonceStr,
        package: res.pay_params.package,
        signType: res.pay_params.signType,
        paySign: res.pay_params.paySign,
        success: async () => {
          uni.hideLoading()
          cartStore.clearCart()
          uni.showToast({ title: '支付成功', icon: 'success' })
          await trackPayment(merchantId.value, res.order.id, res.order.pay_amount)
          
          setTimeout(() => {
            uni.redirectTo({
              url: `/pages/store/my-orders?merchant_id=${merchantId.value}&status=2`
            })
          }, 1500)
        },
        fail: (err) => {
          uni.hideLoading()
          if (err.errMsg?.includes('cancel')) {
            uni.showToast({ title: '支付已取消', icon: 'none' })
            
            setTimeout(() => {
              uni.redirectTo({
                url: `/pages/store/my-orders?merchant_id=${merchantId.value}&status=1`
              })
            }, 1500)
          } else {
            uni.showToast({ title: '支付失败', icon: 'none' })
          }
        },
        complete: () => {
          submitting.value = false
        }
      })
    } else {
      // 无需支付，直接跳转订单页
      uni.hideLoading()
      cartStore.clearCart()
      uni.showToast({ title: '订单创建成功', icon: 'success' })
      await trackPayment(merchantId.value, res.order.id, res.order.pay_amount)
      
      setTimeout(() => {
        uni.redirectTo({
          url: `/pages/store/my-orders?merchant_id=${merchantId.value}`
        })
      }, 1500)
      submitting.value = false
    }
  } catch (error: any) {
    uni.hideLoading()
    submitting.value = false
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

.delivery-mode-empty {
  padding: 24rpx;
  border-radius: 12rpx;
  background: #fff7e6;
  color: #d46b08;
  font-size: 28rpx;
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

.picker-value {
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

.picker-value.disabled {
  color: #999999;
}

.distance-tip {
  font-size: 24rpx;
  color: #ff4d4f;
  margin-top: 12rpx;
  padding: 8rpx 16rpx;
  background: #fff2f0;
  border-radius: 8rpx;
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
  align-items: flex-start;
  gap: 20rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.goods-item:last-child {
  border-bottom: none;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  background: #f0f0f0;
}

.goods-info {
  flex: 1;
  min-width: 0;
  padding-top: 4rpx;
}

.goods-name {
  font-size: 30rpx;
  line-height: 1.4;
  color: #1a1a1a;
  margin-bottom: 8rpx;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.goods-spec {
  font-size: 26rpx;
  line-height: 1.4;
  color: #999999;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.goods-right {
  width: 148rpx;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
  text-align: right;
  padding-top: 4rpx;
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

.submit-btn.disabled {
  opacity: 0.72;
}
</style>
