<template>
  <view class="detail-container" v-if="order">
    <view class="status-bar">
      <view class="status-icon">
        <text class="icon-text">{{ getStatusIcon() }}</text>
      </view>
      <view class="status-info">
        <view class="status-text">{{ getStatusText() }}</view>
        <view class="status-desc">{{ getStatusDesc() }}</view>
      </view>
    </view>

    <view class="section">
      <view class="section-title">商家信息</view>
      <view class="info-row">
        <text class="label">商家名称</text>
        <text class="value">{{ order.merchant?.name || '商家' }}</text>
      </view>
      <view class="info-row" v-if="order.merchant?.address">
        <text class="label">商家地址</text>
        <text class="value">{{ order.merchant?.address }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">配送信息</view>
      <view class="info-row">
        <text class="label">取餐方式</text>
        <text class="value">{{ getDeliveryTypeText() }}</text>
      </view>
      <view v-if="deliveryAddressText" class="info-row">
        <text class="label">收货地址</text>
        <text class="value">{{ deliveryAddressText }}</text>
      </view>
      <view v-if="contactNameText" class="info-row">
        <text class="label">联系人</text>
        <text class="value">{{ contactNameText }}</text>
      </view>
      <view v-if="contactPhoneText" class="info-row">
        <text class="label">联系电话</text>
        <text class="value">{{ contactPhoneText }}</text>
      </view>
      <view v-if="order.delivery_distance" class="info-row">
        <text class="label">配送距离</text>
        <text class="value">{{ Number(order.delivery_distance || 0).toFixed(2) }} 公里</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">订单信息</view>
      <view class="info-row">
        <text class="label">订单号</text>
        <view class="inline-value">
          <text class="value">{{ order.order_no }}</text>
          <text class="copy-btn" @click="copyOrderNo">复制</text>
        </view>
      </view>
      <view class="info-row">
        <text class="label">下单时间</text>
        <text class="value">{{ formatDateTime(order.created_at) }}</text>
      </view>
      <view class="info-row" v-if="order.paid_at">
        <text class="label">支付时间</text>
        <text class="value">{{ formatDateTime(order.paid_at) }}</text>
      </view>
      <view class="info-row" v-if="order.completed_at">
        <text class="label">完成时间</text>
        <text class="value">{{ formatDateTime(order.completed_at) }}</text>
      </view>
      <view class="info-row" v-if="order.refunded_at">
        <text class="label">退款时间</text>
        <text class="value">{{ formatDateTime(order.refunded_at) }}</text>
      </view>
      <view class="info-row" v-if="order.verify_code && order.status === OrderStatus.PAID">
        <text class="label">核销码</text>
        <view class="inline-value">
          <text class="value">{{ order.verify_code }}</text>
          <text class="copy-btn" @click="copyVerifyCode">复制</text>
        </view>
      </view>
      <view class="info-row" v-if="order.remark">
        <text class="label">备注</text>
        <text class="value">{{ order.remark }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">商品信息</view>
      <view
        v-for="(item, index) in order.items"
        :key="`${item.product_id}-${index}`"
        class="goods-item"
      >
        <image class="goods-image" :src="getOrderItemImage(item)" mode="aspectFill" />
        <view class="goods-info">
          <view class="goods-name">{{ item.product_name }}</view>
          <view class="goods-spec" v-if="item.specs">{{ item.specs }}</view>
        </view>
        <view class="goods-right">
          <view class="goods-price">¥{{ Number(item.price || 0).toFixed(2) }}</view>
          <view class="goods-quantity">x{{ item.quantity }}</view>
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section-title">金额明细</view>
      <view class="amount-row">
        <text class="amount-label">商品金额</text>
        <text class="amount-value">¥{{ Number(order.total_amount || 0).toFixed(2) }}</text>
      </view>
      <view class="amount-row">
        <text class="amount-label">配送费</text>
        <text class="amount-value">¥{{ Number(order.delivery_fee || 0).toFixed(2) }}</text>
      </view>
      <view v-if="Number(order.discount_amount || 0) > 0" class="amount-row discount-row">
        <text class="amount-label">优惠金额</text>
        <text class="amount-value discount-value">-¥{{ Number(order.discount_amount || 0).toFixed(2) }}</text>
      </view>
      <view class="amount-row total-row">
        <text class="amount-label total-label">支付金额</text>
        <text class="amount-value total-value">¥{{ Number(order.pay_amount || 0).toFixed(2) }}</text>
      </view>
    </view>

    <view class="bottom-bar">
      <button v-if="canCancel" class="btn secondary" :disabled="submitting" @click="cancelOrder">
        取消订单
      </button>
      <button v-if="canRefund" class="btn warning" :disabled="submitting" @click="contactMerchantForRefund">
        联系商家退款
      </button>
      <button v-if="order.verify_code && order.status === OrderStatus.PAID" class="btn primary" @click="showVerifyDialog">
        查看核销码
      </button>
    </view>

    <view v-if="showVerify" class="dialog-mask" @click="closeVerifyDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">核销码</view>
        <image
          v-if="order.verify_qrcode_url"
          class="verify-qrcode"
          :src="order.verify_qrcode_url"
          mode="aspectFit"
          show-menu-by-longpress
        />
        <view class="verify-code-display">
          <text class="code">{{ order.verify_code || '------' }}</text>
        </view>
        <view class="verify-hint">请将核销码出示给商家扫描</view>
        <view class="dialog-close" @click="closeVerifyDialog">关闭</view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { cancelMyOrder, getMyOrderDetail } from '@api'
import { DeliveryTypeText, OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'
import { BrandAsset } from '../../utils/constants'
import { useAuth } from '../../utils/useAuth'

const order = ref<Order | null>(null)
const showVerify = ref(false)
const submitting = ref(false)

const canCancel = computed(() => order.value?.status === OrderStatus.PENDING_PAYMENT)
const canRefund = computed(() => {
  return order.value?.status === OrderStatus.PAID
})
const deliveryAddressText = computed(() => {
  return order.value?.delivery_info?.address || order.value?.delivery_address || ''
})
const contactNameText = computed(() => {
  return order.value?.delivery_info?.contact_name || order.value?.contact_name || ''
})
const contactPhoneText = computed(() => {
  return order.value?.delivery_info?.contact_phone || order.value?.contact_phone || ''
})

onLoad(async (options: any) => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  const orderId = Number(options?.id || 0)
  if (orderId > 0) {
    await loadOrder(orderId)
  }
})

async function loadOrder(id: number) {
  try {
    order.value = await getMyOrderDetail(id)
  } catch (error: any) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  }
}

function getStatusText() {
  return order.value ? OrderStatusText[order.value.status] || '未知状态' : ''
}

function getStatusDesc() {
  if (!order.value) return ''
  const descMap: Record<number, string> = {
    [OrderStatus.PENDING_PAYMENT]: '订单待支付，可取消',
    [OrderStatus.PAID]: '订单已支付，请出示核销码',
    [OrderStatus.COMPLETED]: '订单已完成',
    [OrderStatus.CANCELLED]: '订单已取消',
    [OrderStatus.REFUNDING]: '退款处理中，请耐心等待',
    [OrderStatus.REFUNDED]: '订单已退款'
  }
  return descMap[order.value.status] || ''
}

function getStatusIcon() {
  if (!order.value) return ''
  const iconMap: Record<number, string> = {
    [OrderStatus.PENDING_PAYMENT]: '⏰',
    [OrderStatus.PAID]: '💰',
    [OrderStatus.COMPLETED]: '✅',
    [OrderStatus.CANCELLED]: '❌',
    [OrderStatus.REFUNDING]: '🔄',
    [OrderStatus.REFUNDED]: '💸'
  }
  return iconMap[order.value.status] || '❓'
}

function getDeliveryTypeText() {
  if (!order.value?.delivery_type) return ''
  return DeliveryTypeText[order.value.delivery_type] || ''
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function getOrderItemImage(item: any) {
  if (typeof item?.image === 'string' && item.image.trim()) {
    return item.image
  }

  if (Array.isArray(item?.images) && typeof item.images[0] === 'string' && item.images[0].trim()) {
    return item.images[0]
  }

  if (typeof item?.product_image === 'string' && item.product_image.trim()) {
    return item.product_image
  }

  return BrandAsset.DEFAULT_PRODUCT_IMAGE
}

function copyOrderNo() {
  if (!order.value?.order_no) return
  uni.setClipboardData({ data: order.value.order_no, success: () => uni.showToast({ title: '已复制', icon: 'success' }) })
}

function copyVerifyCode() {
  if (!order.value?.verify_code) return
  uni.setClipboardData({ data: order.value.verify_code, success: () => uni.showToast({ title: '已复制', icon: 'success' }) })
}

function showVerifyDialog() {
  showVerify.value = true
}

function closeVerifyDialog() {
  showVerify.value = false
}

async function cancelOrder() {
  if (!order.value) return
  uni.showModal({
    title: '确认取消',
    content: '确定要取消该订单吗？',
    success: async (res) => {
      if (!res.confirm || !order.value) return
      submitting.value = true
      try {
        await cancelMyOrder(order.value.id)
        order.value.status = OrderStatus.CANCELLED
        uni.showToast({ title: '订单已取消', icon: 'success' })
      } catch (error: any) {
        uni.showToast({ title: error.message || '取消失败', icon: 'none' })
      } finally {
        submitting.value = false
      }
    }
  })
}

function contactMerchantForRefund() {
  if (!order.value) return
  const phone = (order.value.merchant?.contact_phone || order.value.merchant?.phone || '').trim()
  const merchantName = order.value.merchant?.name || '商家'

  if (!phone) {
    uni.showModal({
      title: '联系商家退款',
      content: `请联系${merchantName}协助处理退款。`,
      showCancel: false,
      confirmText: '我知道了'
    })
    return
  }

  uni.makePhoneCall({
    phoneNumber: phone,
    fail: () => {
      uni.showToast({ title: `请联系${merchantName}退款`, icon: 'none' })
    }
  })
}
</script>

<style scoped>
.detail-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 160rpx;
}

.status-bar {
  display: flex;
  align-items: center;
  padding: 48rpx 32rpx;
  color: #ffffff;
  background: linear-gradient(135deg, #1677ff 0%, #0b57d0 100%);
}

.status-icon {
  width: 96rpx;
  height: 96rpx;
  margin-right: 24rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.18);
}

.icon-text {
  font-size: 48rpx;
}

.status-info {
  flex: 1;
}

.status-text {
  font-size: 36rpx;
  font-weight: 600;
}

.status-desc {
  margin-top: 8rpx;
  font-size: 26rpx;
  opacity: 0.9;
}

.section {
  margin: 24rpx;
  padding: 32rpx;
  border-radius: 20rpx;
  background: #ffffff;
}

.section-title {
  margin-bottom: 24rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
}

.info-row,
.amount-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24rpx;
  padding: 14rpx 0;
}

.amount-row {
  align-items: center;
}

.label {
  color: #86909c;
  font-size: 26rpx;
}

.value {
  flex: 1;
  text-align: right;
  color: #1f2329;
  font-size: 26rpx;
  word-break: break-all;
}

.inline-value {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16rpx;
  flex: 1;
}

.copy-btn {
  color: #1677ff;
  font-size: 24rpx;
}

.goods-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.goods-item:last-child {
  border-bottom: none;
}

.goods-image {
  width: 120rpx;
  height: 120rpx;
  margin-right: 20rpx;
  border-radius: 16rpx;
  background: #f2f3f5;
}

.goods-info {
  flex: 1;
  min-width: 0;
}

.goods-name {
  font-size: 28rpx;
  color: #1f2329;
}

.goods-spec {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #86909c;
}

.goods-right {
  text-align: right;
}

.goods-price {
  font-size: 28rpx;
  color: #1f2329;
}

.goods-quantity {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #86909c;
}

.amount-label {
  font-size: 26rpx;
  color: #4e5969;
}

.amount-value {
  font-size: 26rpx;
  color: #1f2329;
  font-weight: 500;
}

.total-row {
  padding-top: 24rpx;
  border-top: 1rpx solid #f2f3f5;
}

.discount-value {
  color: #ff4d4f;
}

.total-label {
  color: #1f2329;
  font-weight: 600;
}

.total-value {
  color: #f53f3f;
  font-weight: 600;
  font-size: 30rpx;
}

.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  gap: 20rpx;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -8rpx 24rpx rgba(0, 0, 0, 0.06);
}

.btn {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 999rpx;
  font-size: 28rpx;
  border: none;
}

.btn.primary {
  color: #ffffff;
  background: #1677ff;
}

.btn.secondary {
  color: #1677ff;
  background: #eef3ff;
}

.btn.warning {
  color: #ffffff;
  background: #fa8c16;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48rpx;
  background: rgba(0, 0, 0, 0.45);
}

.dialog-content {
  width: 100%;
  padding: 36rpx 32rpx;
  border-radius: 24rpx;
  background: #ffffff;
}

.dialog-title {
  text-align: center;
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2329;
}

.verify-code-display {
  margin: 24rpx 0 20rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f7f8fa;
  text-align: center;
}

.verify-qrcode {
  width: 320rpx;
  height: 320rpx;
  margin: 32rpx auto 0;
  display: block;
  border-radius: 20rpx;
  background: #f7f8fa;
}

.code {
  font-size: 48rpx;
  font-weight: 700;
  letter-spacing: 8rpx;
  color: #1677ff;
}

.verify-hint {
  text-align: center;
  font-size: 24rpx;
  color: #86909c;
}

.dialog-close {
  margin-top: 32rpx;
  text-align: center;
  color: #1677ff;
  font-size: 28rpx;
}

</style>
