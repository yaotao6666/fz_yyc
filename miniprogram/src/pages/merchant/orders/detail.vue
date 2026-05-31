<template>
  <view class="detail-container" v-if="order">
    <!-- 订单状态 -->
    <view class="status-bar">
      <view class="status-icon">
        <text class="icon-text">{{ getStatusIcon() }}</text>
      </view>
      <view class="status-info">
        <view class="status-text">{{ getStatusText() }}</view>
        <view class="status-desc">{{ getStatusDesc() }}</view>
      </view>
    </view>

    <!-- 配送信息 -->
    <view class="section delivery-info" v-if="order.delivery_type">
      <view class="section-title">配送信息</view>
      <view class="info-row">
        <text class="info-label">配送方式</text>
        <text class="info-value">{{ getDeliveryTypeText() }}</text>
      </view>
      <view class="info-row" v-if="order.delivery_distance">
        <text class="info-label">配送距离</text>
        <text class="info-value">{{ order.delivery_distance }}公里</text>
      </view>
      <view class="info-row" v-if="deliveryAddressText">
        <text class="info-label">收货地址</text>
        <text class="info-value">{{ deliveryAddressText }}</text>
      </view>
      <view class="info-row" v-if="contactNameText">
        <text class="info-label">联系人</text>
        <text class="info-value">{{ contactNameText }}</text>
      </view>
      <view class="info-row" v-if="contactPhoneText">
        <text class="info-label">联系电话</text>
        <text class="info-value">{{ contactPhoneText }}</text>
      </view>
    </view>

    <!-- 用户信息 -->
    <view class="section user-info" v-if="order.user">
      <view class="section-title">用户信息</view>
      <view class="user-card">
        <image
          class="user-avatar"
          :src="order.user.avatar || '/static/default-avatar.png'"
          mode="aspectFill"
        />
        <view class="user-detail">
          <view class="user-name">{{ order.user.nickname || '微信用户' }}</view>
          <view class="user-phone" v-if="order.user.phone">{{ order.user.phone }}</view>
        </view>
      </view>
    </view>

    <!-- 商品信息 -->
    <view class="section goods-info">
      <view class="section-title">商品信息</view>
      <view class="goods-list">
        <view class="goods-item" v-for="(item, index) in order.items" :key="index">
          <image
            class="goods-image"
            :src="item.image || '/static/default-product.png'"
            mode="aspectFill"
          />
          <view class="goods-detail">
            <view class="goods-name">{{ item.product_name }}</view>
            <view class="goods-spec" v-if="item.specs">{{ item.specs }}</view>
          </view>
          <view class="goods-price">
            <text class="price">¥{{ item.price.toFixed(2) }}</text>
            <text class="quantity">x{{ item.quantity }}</text>
          </view>
        </view>
      </view>
      <view class="remark" v-if="order.remark">
        <text class="remark-label">备注:</text>
        <text class="remark-text">{{ order.remark }}</text>
      </view>
    </view>

    <!-- 订单信息 -->
    <view class="section order-info">
      <view class="section-title">订单信息</view>
      <view class="info-row">
        <text class="info-label">订单编号</text>
        <view class="info-value-group">
          <text class="info-value">{{ order.order_no }}</text>
          <text class="copy-btn" @click="copyOrderNo">复制</text>
        </view>
      </view>
      <view class="info-row">
        <text class="info-label">下单时间</text>
        <text class="info-value">{{ formatDateTime(order.created_at) }}</text>
      </view>
      <view class="info-row" v-if="order.paid_at">
        <text class="info-label">支付时间</text>
        <text class="info-value">{{ formatDateTime(order.paid_at) }}</text>
      </view>
      <view class="info-row" v-if="order.transaction_id">
        <text class="info-label">支付单号</text>
        <text class="info-value">{{ order.transaction_id }}</text>
      </view>
      <view class="info-row" v-if="order.verify_code">
        <text class="info-label">核销码</text>
        <view class="info-value-group">
          <text class="info-value verify-code">{{ order.verify_code }}</text>
          <text class="copy-btn" @click="copyVerifyCode">复制</text>
        </view>
      </view>
      <view class="info-row" v-if="order.completed_at">
        <text class="info-label">核销时间</text>
        <text class="info-value">{{ formatDateTime(order.completed_at) }}</text>
      </view>
      <view class="info-row" v-if="order.completed_by_name">
        <text class="info-label">核销人</text>
        <text class="info-value">{{ order.completed_by_name }}</text>
      </view>
    </view>

    <!-- 金额明细 -->
    <view class="section amount-info">
      <view class="section-title">金额明细</view>
      <view class="amount-row">
        <text class="amount-label">商品金额</text>
        <text class="amount-value">¥{{ order.total_amount.toFixed(2) }}</text>
      </view>
      <view class="amount-row" v-if="order.delivery_fee > 0">
        <text class="amount-label">配送费</text>
        <text class="amount-value">¥{{ order.delivery_fee.toFixed(2) }}</text>
      </view>
      <view class="amount-row discount" v-if="order.discount_amount > 0">
        <text class="amount-label">优惠</text>
        <text class="amount-value">-¥{{ order.discount_amount.toFixed(2) }}</text>
      </view>
      <view class="amount-row total">
        <text class="amount-label">实付金额</text>
        <text class="amount-value">¥{{ order.pay_amount.toFixed(2) }}</text>
      </view>
    </view>

    <!-- 操作按钮 -->
    <view class="bottom-bar">
      <template v-if="order.status === OrderStatus.PAID">
        <button class="btn-verify" :disabled="verifying" @click="confirmVerify">
          {{ verifying ? '核销中...' : '核销订单' }}
        </button>
      </template>
      <template v-if="canRefund">
        <button class="btn-refund" :disabled="refunding" @click="showRefundDialog">
          {{ refunding ? '提交中...' : '发起退款' }}
        </button>
      </template>
    </view>

    <!-- 退款弹窗 -->
    <view v-if="showRefund" class="dialog-mask" @click="closeRefundDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">发起退款</view>
        <view class="refund-summary">
          <text class="refund-summary-text">订单号：{{ order.order_no }}</text>
          <text class="refund-summary-text">退款金额：¥{{ order.pay_amount.toFixed(2) }}</text>
        </view>
        <textarea
          v-model="refundReason"
          class="refund-textarea"
          maxlength="120"
          placeholder="请输入退款原因"
        />
        <view class="dialog-actions">
          <button class="btn-cancel" @click="closeRefundDialog">取消</button>
          <button class="btn-confirm" :disabled="refunding" @click="submitRefund">
            {{ refunding ? '提交中...' : '确认退款' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getOrder, completeOrder, refundOrder } from '@api'
import { OrderStatus, OrderStatusText, DeliveryTypeText } from '@types'
import type { Order } from '@types'

const order = ref<Order | null>(null)
const verifying = ref(false)
const showRefund = ref(false)
const refundReason = ref('')
const refunding = ref(false)

const canRefund = computed(() => {
  return order.value?.status === OrderStatus.PAID || order.value?.status === OrderStatus.COMPLETED
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

onLoad((options: any) => {
  if (options.id) {
    loadOrder(Number(options.id))
  }
})

async function loadOrder(id: number) {
  try {
    order.value = await getOrder(id)
  } catch (error) {
    console.error('加载订单失败:', error)
    uni.showToast({ title: '加载失败', icon: 'none' })
  }
}

function getStatusText(): string {
  if (!order.value) return ''
  return OrderStatusText[order.value.status] || '未知状态'
}

function getStatusIcon(): string {
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

function getStatusDesc(): string {
  if (!order.value) return ''
  const descMap: Record<number, string> = {
    [OrderStatus.PENDING_PAYMENT]: '等待用户支付',
    [OrderStatus.PAID]: '用户已支付，请及时处理',
    [OrderStatus.COMPLETED]: '订单已完成',
    [OrderStatus.CANCELLED]: '订单已取消',
    [OrderStatus.REFUNDING]: '退款处理中',
    [OrderStatus.REFUNDED]: '已退款'
  }
  return descMap[order.value.status] || ''
}

function getDeliveryTypeText(): string {
  if (!order.value?.delivery_type) return ''
  return DeliveryTypeText[order.value.delivery_type] || ''
}

function formatDateTime(time: string): string {
  const date = new Date(time)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function copyOrderNo() {
  if (!order.value) return
  uni.setClipboardData({
    data: order.value.order_no,
    success: () => uni.showToast({ title: '已复制', icon: 'success' })
  })
}

function copyVerifyCode() {
  if (!order.value?.verify_code) return
  uni.setClipboardData({
    data: order.value.verify_code,
    success: () => uni.showToast({ title: '已复制', icon: 'success' })
  })
}

async function confirmVerify() {
  const code = order.value?.verify_code?.trim() || ''
  if (!code) {
    return uni.showToast({ title: '未获取到核销码', icon: 'none' })
  }

  if (!/^\d{6}$/.test(code)) {
    return uni.showToast({ title: '核销码应为6位数字', icon: 'none' })
  }

  if (!order.value) return

  verifying.value = true

  try {
    order.value = await completeOrder(order.value.id, code)
    uni.showToast({ title: '核销成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '核销失败', icon: 'none' })
  } finally {
    verifying.value = false
  }
}

function showRefundDialog() {
  refundReason.value = ''
  showRefund.value = true
}

function closeRefundDialog() {
  showRefund.value = false
  refundReason.value = ''
}

async function submitRefund() {
  if (!order.value) return
  refunding.value = true

  try {
    await refundOrder(order.value.id, {
      reason: refundReason.value.trim(),
      refund_amount: order.value.pay_amount
    })
    order.value.status = OrderStatus.REFUNDING
    order.value.refunded_at = new Date().toISOString()
    uni.showToast({ title: '退款已提交', icon: 'success' })
    closeRefundDialog()
  } catch (error: any) {
    uni.showToast({ title: error.message || '退款失败', icon: 'none' })
  } finally {
    refunding.value = false
  }
}
</script>

<style scoped>
.detail-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}

.status-bar {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  padding: 48rpx 32rpx;
  display: flex;
  align-items: center;
  color: #ffffff;
}

.status-icon {
  width: 100rpx;
  height: 100rpx;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 24rpx;
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
  margin-bottom: 8rpx;
}

.status-desc {
  font-size: 26rpx;
  opacity: 0.9;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 0;
}

.info-label {
  font-size: 28rpx;
  color: #999999;
}

.info-value {
  font-size: 28rpx;
  color: #1a1a1a;
}

.info-value-group {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.copy-btn {
  font-size: 24rpx;
  color: #007AFF;
  padding: 4rpx 12rpx;
  background: #e6f0ff;
  border-radius: 8rpx;
}

.verify-code {
  font-size: 32rpx;
  font-weight: 600;
  letter-spacing: 4rpx;
}

.user-card {
  display: flex;
  align-items: center;
}

.user-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.user-name {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.user-phone {
  font-size: 26rpx;
  color: #999999;
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
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.goods-detail {
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

.goods-price {
  text-align: right;
}

.goods-price .price {
  font-size: 28rpx;
  color: #1a1a1a;
  display: block;
}

.goods-price .quantity {
  font-size: 24rpx;
  color: #999999;
}

.remark {
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #f0f0f0;
  font-size: 28rpx;
}

.remark-label {
  color: #999999;
  margin-right: 8rpx;
}

.remark-text {
  color: #1a1a1a;
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

.amount-row.discount .amount-value {
  color: #ff4d4f;
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
  display: flex;
  justify-content: center;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-verify, .btn-refund {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-verify {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}

.btn-refund {
  background: #fff7e6;
  color: #fa8c16;
}

.dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.dialog-content {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  text-align: center;
  margin-bottom: 32rpx;
}

.verify-code-input {
  margin-bottom: 32rpx;
}

.code-input {
  width: 100%;
  height: 96rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  text-align: center;
  font-size: 36rpx;
  letter-spacing: 8rpx;
}

.dialog-actions {
  display: flex;
  gap: 24rpx;
}

.btn-cancel, .btn-confirm {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666666;
}

.btn-confirm {
  background: #007AFF;
  color: #ffffff;
}

.btn-confirm[disabled] {
  background: #cccccc;
}
</style>
