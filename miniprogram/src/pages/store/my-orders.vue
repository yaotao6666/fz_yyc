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
              :src="order.merchant?.logo || BrandAsset.DEFAULT_MERCHANT_LOGO"
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
              :src="getOrderItemImage(item)"
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
            <view class="amount-summary-row">
              <text class="amount-label">商品金额</text>
              <text class="amount-summary-value">¥{{ order.total_amount.toFixed(2) }}</text>
            </view>
            <view class="amount-summary-row">
              <text class="amount-label">配送费</text>
              <text class="amount-summary-value">¥{{ order.delivery_fee.toFixed(2) }}</text>
            </view>
            <view class="amount-summary-row">
              <text class="amount-label">优惠</text>
              <text class="amount-summary-value discount">
                {{ order.discount_amount > 0 ? `-¥${order.discount_amount.toFixed(2)}` : '¥0.00' }}
              </text>
            </view>
            <view class="amount-summary-row total">
              <text class="amount-label total-label">实付</text>
              <text class="amount-value">¥{{ order.pay_amount.toFixed(2) }}</text>
            </view>
          </view>
        </view>

        <view class="order-actions">
          <template v-if="order.status === 1">
            <view class="action-btn cancel" @click="cancelOrder(order)">取消订单</view>
          </template>
          <template v-if="order.status === 2">
            <view class="action-btn verify" @click="showVerifyCode(order)">核销码</view>
          </template>
          <template v-if="order.status === 2">
            <view class="action-btn refund" @click="contactMerchantForRefund(order)">联系商家退款</view>
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
        <image
          v-if="currentOrder?.verify_qrcode_url"
          class="verify-qrcode"
          :src="currentOrder.verify_qrcode_url"
          mode="aspectFit"
          show-menu-by-longpress
        />
        <view class="verify-code-display">
          <text class="code">{{ currentOrder?.verify_code || '------' }}</text>
        </view>
        <view class="verify-hint">请将核销码出示给商家扫描</view>
        <view class="dialog-close" @click="closeVerifyDialog">关闭</view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMyOrders, cancelMyOrder } from '@api'
import { OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'
import { BrandAsset } from '../../utils/constants'
import { useAuth } from '../../utils/useAuth'

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
const merchantId = ref<number>(0) // 当前商家ID
const merchantName = ref<string>('') // 当前商家名称

const showVerify = ref(false)
const currentOrder = ref<Order | null>(null)

onShow(async () => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const mid = currentPage?.options?.merchant_id
  const status = currentPage?.options?.status

  const { ensureAuth } = useAuth()
  await ensureAuth()

  if (status) {
    const parsed = Number(status)
    if (Number.isFinite(parsed)) {
      currentStatus.value = parsed
    } else if (status === 'paid') {
      currentStatus.value = OrderStatus.PAID
    } else if (status === 'pending') {
      currentStatus.value = OrderStatus.PENDING_PAYMENT
    } else {
      currentStatus.value = 0
    }
  }

  if (mid) {
    merchantId.value = Number(mid)
    loadOrders(true)
  } else {
    merchantId.value = 0
    loadOrders(true)
  }
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

    if (merchantId.value > 0) {
      params.merchant_id = merchantId.value
    }

    const res = await getMyOrders(params)

    if (merchantId.value > 0) {
      const filteredOrders = res.list.filter(order => order.merchant?.id === merchantId.value || (order as any).merchant_id === merchantId.value)
      if (reset) {
        orders.value = filteredOrders
      } else {
        orders.value.push(...filteredOrders)
      }
      if (filteredOrders.length > 0 && filteredOrders[0].merchant) {
        merchantName.value = filteredOrders[0].merchant.name || ''
      }
    } else {
      if (reset) {
        orders.value = res.list
      } else {
        orders.value.push(...res.list)
      }
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

function goDetail(orderId: number) {
  uni.navigateTo({ url: `/pages/store/order-detail?id=${orderId}` })
}

function goShopping() {
  const targetMerchantId = merchantId.value || currentOrder.value?.merchant?.id || orders.value[0]?.merchant?.id || 1
  uni.navigateTo({ url: `/pages/store/home?merchant_id=${targetMerchantId}` })
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

function contactMerchantForRefund(order: Order) {
  const phone = order.merchant?.phone?.trim()
  const merchantName = order.merchant?.name || '商家'

  if (!phone) {
    uni.showModal({
      title: '联系商家退款',
      content: `请联系${merchantName}协助处理退款。`,
      showCancel: false,
      confirmText: '我知道了'
    })
    return
  }

  uni.showModal({
    title: '联系商家退款',
    content: `请联系${merchantName}退款\n联系电话：${phone}`,
    confirmText: '拨打电话',
    cancelText: '取消',
    success: (res) => {
      if (res.confirm) {
        uni.makePhoneCall({ phoneNumber: phone })
      }
    }
  })
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
  width: calc(100% - 48rpx);
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
  min-width: 220rpx;
  text-align: right;
}

.amount-label {
  font-size: 24rpx;
  color: #666666;
}

.amount-summary-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12rpx;
  margin-bottom: 8rpx;
}

.amount-summary-row:last-child {
  margin-bottom: 0;
}

.amount-summary-row.total {
  margin-top: 4rpx;
}

.amount-summary-value {
  font-size: 24rpx;
  color: #1a1a1a;
}

.amount-summary-value.discount {
  color: #ff4d4f;
}

.total-label {
  color: #1a1a1a;
  font-weight: 600;
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

.verify-qrcode {
  width: 320rpx;
  height: 320rpx;
  margin: 0 auto 24rpx;
  display: block;
  border-radius: 20rpx;
  background: #f7f8fa;
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
