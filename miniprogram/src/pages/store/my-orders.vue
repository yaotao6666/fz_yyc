<template>
  <view class="my-orders-container">
    <!-- 状态筛选 -->
    <view class="status-tabs">
      <view
        v-for="tab in statusTabs"
        :key="tab.value"
        class="tab-item"
        :class="{ active: currentStatus === tab.value }"
        @click="changeStatus(tab.value)"
      >
        {{ tab.label }}
      </view>
    </view>

    <!-- 订单列表 -->
    <scroll-view class="order-list" scroll-y @scrolltolower="loadMore">
      <view
        v-for="order in orders"
        :key="order.id"
        class="order-card"
      >
        <view class="order-header">
          <view class="merchant-info">
            <image
              class="merchant-logo"
              :src="order.merchant?.logo || '/static/default-logo.png'"
              mode="aspectFill"
            />
            <text class="merchant-name">{{ order.merchant?.name || '商家' }}</text>
          </view>
          <view class="order-status" :class="getStatusClass(order.status)">
            {{ getStatusText(order.status) }}
          </view>
        </view>

        <view class="order-items" @click="goDetail(order.id)">
          <view
            v-for="(item, index) in order.items.slice(0, 3)"
            :key="index"
            class="order-item"
          >
            <image
              class="item-image"
              :src="item.image || '/static/default-product.png'"
              mode="aspectFill"
            />
          </view>
          <view v-if="order.items.length > 3" class="more-items">
            +{{ order.items.length - 3 }}
          </view>
        </view>

        <view class="order-footer">
          <view class="order-info">
            <text class="order-no">{{ order.order_no }}</text>
            <text class="order-time">{{ formatTime(order.created_at) }}</text>
          </view>
          <view class="order-amount">
            <text class="amount-label">实付</text>
            <text class="amount-value">¥{{ order.pay_amount.toFixed(2) }}</text>
          </view>
        </view>

        <view class="order-actions">
          <template v-if="order.status === 1">
            <view class="action-btn cancel" @click="cancelOrder(order)">取消订单</view>
          </template>
          <template v-if="order.status === 2">
            <view class="action-btn verify" @click="showVerifyCode(order)">核销码</view>
          </template>
          <template v-if="order.status === 2 || order.status === 5">
            <view class="action-btn refund" @click="applyRefund(order)">申请退款</view>
          </template>
          <view class="action-btn primary" @click="goDetail(order.id)">查看详情</view>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && orders.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && orders.length === 0" class="empty">
        <text class="empty-icon">📋</text>
        <text class="empty-text">暂无订单</text>
        <button class="btn-shopping" @click="goShopping">去购物</button>
      </view>
    </scroll-view>

    <!-- 核销码弹窗 -->
    <view v-if="showVerify" class="dialog-mask" @click="closeVerifyDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">核销码</view>
        <view class="verify-code-display">
          <text class="code">{{ currentOrder?.verify_code || '------' }}</text>
        </view>
        <view class="verify-hint">请将核销码出示给商家扫描</view>
        <view class="dialog-close" @click="closeVerifyDialog">关闭</view>
      </view>
    </view>

    <!-- 退款原因弹窗 -->
    <view v-if="showRefund" class="dialog-mask" @click="closeRefundDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">申请退款</view>
        <view class="refund-reason">
          <textarea
            v-model="refundReason"
            class="reason-input"
            placeholder="请输入退款原因"
            maxlength="200"
          />
        </view>
        <view class="dialog-actions">
          <button class="btn-cancel" @click="closeRefundDialog">取消</button>
          <button class="btn-confirm" :disabled="submitting" @click="confirmRefund">
            {{ submitting ? '提交中...' : '确认提交' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onShow } from 'vue'
import { getMyOrders, cancelMyOrder, applyRefund } from '../../api'
import { OrderStatus, OrderStatusText } from '../../types/api'
import type { Order } from '../../types/api'

const statusTabs = [
  { label: '全部', value: 0 },
  { label: '待支付', value: OrderStatus.PENDING_PAYMENT },
  { label: '已支付', value: OrderStatus.PAID },
  { label: '已完成', value: OrderStatus.COMPLETED }
]

const currentStatus = ref(0)
const orders = ref<Order[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

const showVerify = ref(false)
const currentOrder = ref<Order | null>(null)

const showRefund = ref(false)
const refundReason = ref('')
const submitting = ref(false)
const refundOrderId = ref<number | null>(null)

onShow(() => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const status = currentPage?.options?.status
  
  if (status) {
    currentStatus.value = Number(status)
  }
  
  loadOrders(true)
})

async function loadOrders(reset = false) {
  if (reset) {
    page.value = 1
    noMore.value = false
    orders.value = []
  }

  if (noMore.value || loading.value) return

  loading.value = true

  try {
    const params: any = {
      page: page.value,
      page_size: pageSize
    }

    if (currentStatus.value !== 0) {
      params.status = currentStatus.value
    }

    const res = await getMyOrders(params)

    if (reset) {
      orders.value = res.list
    } else {
      orders.value.push(...res.list)
    }

    if (res.list.length < pageSize) {
      noMore.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载订单失败:', error)
  } finally {
    loading.value = false
  }
}

function loadMore() {
  loadOrders()
}

function changeStatus(status: number) {
  currentStatus.value = status
  loadOrders(true)
}

function getStatusText(status: number): string {
  return OrderStatusText[status] || '未知'
}

function getStatusClass(status: number): string {
  const classMap: Record<number, string> = {
    [OrderStatus.PENDING_PAYMENT]: 'pending',
    [OrderStatus.PAID]: 'paid',
    [OrderStatus.COMPLETED]: 'completed',
    [OrderStatus.CANCELLED]: 'cancelled',
    [OrderStatus.REFUNDING]: 'refunding',
    [OrderStatus.REFUNDED]: 'refunded'
  }
  return classMap[status] || ''
}

function formatTime(time: string): string {
  const date = new Date(time)
  return `${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

function goDetail(orderId: number) {
  uni.navigateTo({ url: `/pages/merchant/orders/detail?id=${orderId}` })
}

function goShopping() {
  uni.navigateTo({ url: '/pages/store/home' })
}

function cancelOrder(order: Order) {
  uni.showModal({
    title: '确认取消',
    content: '确定要取消该订单吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await cancelMyOrder(order.id)
          const index = orders.value.findIndex(o => o.id === order.id)
          if (index !== -1) {
            orders.value[index].status = OrderStatus.CANCELLED
          }
          uni.showToast({ title: '订单已取消', icon: 'success' })
        } catch (error: any) {
          uni.showToast({ title: error.message || '取消失败', icon: 'none' })
        }
      }
    }
  })
}

function showVerifyCode(order: Order) {
  currentOrder.value = order
  showVerify.value = true
}

function closeVerifyDialog() {
  showVerify.value = false
}

function applyRefund(order: Order) {
  refundOrderId.value = order.id
  refundReason.value = ''
  showRefund.value = true
}

function closeRefundDialog() {
  showRefund.value = false
  refundReason.value = ''
}

async function confirmRefund() {
  if (!refundReason.value.trim()) {
    return uni.showToast({ title: '请输入退款原因', icon: 'none' })
  }

  if (!refundOrderId.value) return

  submitting.value = true

  try {
    await applyRefund(refundOrderId.value, { refund_reason: refundReason.value })
    
    const index = orders.value.findIndex(o => o.id === refundOrderId.value)
    if (index !== -1) {
      orders.value[index].status = OrderStatus.REFUNDING
    }

    uni.showToast({ title: '退款申请已提交', icon: 'success' })
    closeRefundDialog()
  } catch (error: any) {
    uni.showToast({ title: error.message || '申请失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.my-orders-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.status-tabs {
  display: flex;
  background: #ffffff;
  padding: 24rpx 0;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tab-item {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #666666;
  padding: 16rpx 0;
  position: relative;
}

.tab-item.active {
  color: #007AFF;
  font-weight: 600;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 4rpx;
  background: #007AFF;
  border-radius: 2rpx;
}

.order-list {
  padding: 24rpx;
}

.order-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.merchant-info {
  display: flex;
  align-items: center;
}

.merchant-logo {
  width: 48rpx;
  height: 48rpx;
  border-radius: 8rpx;
  background: #f0f0f0;
  margin-right: 12rpx;
}

.merchant-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.order-status {
  font-size: 26rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
}

.order-status.pending { background: #fff7e6; color: #fa8c16; }
.order-status.paid { background: #e6f7ff; color: #007AFF; }
.order-status.completed { background: #f6ffed; color: #52c41a; }
.order-status.cancelled { background: #f5f5f5; color: #999999; }
.order-status.refunding { background: #fff7e6; color: #fa8c16; }
.order-status.refunded { background: #fff1f0; color: #ff4d4f; }

.order-items {
  display: flex;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.item-image {
  width: 140rpx;
  height: 140rpx;
  border-radius: 8rpx;
  background: #f0f0f0;
}

.more-items {
  width: 140rpx;
  height: 140rpx;
  border-radius: 8rpx;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #999999;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.order-no {
  font-size: 24rpx;
  color: #999999;
}

.order-time {
  font-size: 24rpx;
  color: #999999;
}

.order-amount {
  text-align: right;
}

.amount-label {
  font-size: 24rpx;
  color: #666666;
}

.amount-value {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
  margin-left: 8rpx;
}

.order-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  padding-top: 20rpx;
}

.action-btn {
  padding: 12rpx 28rpx;
  border-radius: 32rpx;
  font-size: 26rpx;
}

.action-btn.primary {
  background: #007AFF;
  color: #ffffff;
}

.action-btn.cancel {
  background: #f5f5f5;
  color: #666666;
}

.action-btn.verify {
  background: #e6f7ff;
  color: #007AFF;
}

.action-btn.refund {
  background: #fff1f0;
  color: #ff4d4f;
}

.loading, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 100rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999999;
  margin-bottom: 32rpx;
}

.btn-shopping {
  padding: 20rpx 48rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 40rpx;
  font-size: 28rpx;
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

.verify-code-display {
  text-align: center;
  margin-bottom: 24rpx;
}

.code {
  font-size: 64rpx;
  font-weight: 700;
  letter-spacing: 16rpx;
  color: #007AFF;
}

.verify-hint {
  text-align: center;
  font-size: 28rpx;
  color: #999999;
  margin-bottom: 32rpx;
}

.dialog-close {
  text-align: center;
  font-size: 30rpx;
  color: #666666;
  padding: 16rpx;
}

.refund-reason {
  margin-bottom: 32rpx;
}

.reason-input {
  width: 100%;
  height: 200rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 24rpx;
  font-size: 28rpx;
  box-sizing: border-box;
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
