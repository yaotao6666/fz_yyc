<template>
  <view class="order-list-container">
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
        <text v-if="tab.count" class="tab-count">{{ tab.count }}</text>
      </view>
    </view>

    <!-- 订单列表 -->
    <scroll-view class="order-list" scroll-y @scrolltolower="loadMore">
      <view
        v-for="order in orders"
        :key="order.id"
        class="order-card"
        @click="goDetail(order.id)"
      >
        <view class="order-header">
          <view class="order-no">订单号: {{ order.order_no }}</view>
          <view class="order-status" :class="getStatusClass(order.status)">
            {{ getStatusText(order.status) }}
          </view>
        </view>

        <view class="order-items">
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
            <view class="item-info">
              <view class="item-name">{{ item.product_name }}</view>
              <view class="item-spec" v-if="item.specs">{{ item.specs }}</view>
            </view>
            <view class="item-price">
              <text class="price">¥{{ item.price.toFixed(2) }}</text>
              <text class="quantity">x{{ item.quantity }}</text>
            </view>
          </view>
          <view v-if="order.items.length > 3" class="more-items">
            还有{{ order.items.length - 3 }}件商品
          </view>
        </view>

        <view class="order-footer">
          <view class="order-time">{{ formatTime(order.created_at) }}</view>
          <view class="order-amount">
            <text>共{{ order.items.length }}件</text>
            <text class="amount">¥{{ order.pay_amount.toFixed(2) }}</text>
          </view>
        </view>

        <view class="order-actions" @click.stop>
          <template v-if="order.status === 2">
            <view class="action-btn primary" @click="showVerifyDialog(order)">核销</view>
          </template>
          <template v-if="order.status === 5">
            <view class="action-btn warning">处理退款</view>
          </template>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && orders.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && orders.length === 0" class="empty">
        <text class="empty-icon">📋</text>
        <text class="empty-text">暂无订单</text>
      </view>
    </scroll-view>

    <!-- 核销弹窗 -->
    <view v-if="showVerify" class="dialog-mask" @click="closeVerifyDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">订单核销</view>
        <view class="verify-order-info">
          <view class="order-no">订单号: {{ currentOrder?.order_no }}</view>
          <view class="order-amount">金额: ¥{{ currentOrder?.pay_amount?.toFixed(2) }}</view>
        </view>
        <view class="verify-code">
          <input
            v-model="verifyCode"
            type="text"
            class="code-input"
            placeholder="请输入核销码"
            maxlength="6"
          />
        </view>
        <view class="dialog-actions">
          <button class="btn-cancel" @click="closeVerifyDialog">取消</button>
          <button class="btn-confirm" :disabled="verifying" @click="confirmVerify">
            {{ verifying ? '核销中...' : '确认核销' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getOrders, completeOrder } from '@api'
import { OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'

const statusTabs = [
  { label: '全部', value: 0, count: 0 },
  { label: '待支付', value: OrderStatus.PENDING_PAYMENT, count: 0 },
  { label: '已支付', value: OrderStatus.PAID, count: 0 },
  { label: '已完成', value: OrderStatus.COMPLETED, count: 0 },
  { label: '退款', value: OrderStatus.REFUNDING, count: 0 }
]

const currentStatus = ref(0)
const orders = ref<Order[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

const showVerify = ref(false)
const currentOrder = ref<Order | null>(null)
const verifyCode = ref('')
const verifying = ref(false)

onShow(() => {
  loadOrders(true)
  loadStatistics()
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

    const res = await getOrders(params)

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

async function loadStatistics() {
  try {
    // 简化：直接从订单列表获取统计
    // 实际应该从专门的统计接口获取
  } catch (error) {
    console.error('加载统计失败:', error)
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

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/merchant/orders/detail?id=${id}` })
}

function showVerifyDialog(order: Order) {
  currentOrder.value = order
  verifyCode.value = ''
  showVerify.value = true
}

function closeVerifyDialog() {
  showVerify.value = false
  verifyCode.value = ''
}

async function confirmVerify() {
  const code = verifyCode.value.trim()
  if (!code) {
    return uni.showToast({ title: '请输入核销码', icon: 'none' })
  }

  if (!/^\d{6}$/.test(code)) {
    return uni.showToast({ title: '核销码应为6位数字', icon: 'none' })
  }

  if (!currentOrder.value) return

  verifying.value = true

  try {
    const updatedOrder = await completeOrder(currentOrder.value.id, code)
    
    // 更新订单状态
    const index = orders.value.findIndex(o => o.id === currentOrder.value!.id)
    if (index !== -1) {
      orders.value[index] = updatedOrder
    }

    uni.showToast({ title: '核销成功', icon: 'success' })
    closeVerifyDialog()
  } catch (error: any) {
    uni.showToast({ title: error.message || '核销失败', icon: 'none' })
  } finally {
    verifying.value = false
  }
}
</script>

<style scoped>
.order-list-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.status-tabs {
  display: flex;
  background: #ffffff;
  padding: 24rpx 24rpx;
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

.tab-count {
  display: inline-block;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  background: #ff4d4f;
  color: #ffffff;
  border-radius: 16rpx;
  font-size: 22rpx;
  padding: 0 8rpx;
  margin-left: 8rpx;
}

.order-list {
  padding: 24rpx 0rpx;
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
  padding-bottom: 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.order-no {
  font-size: 26rpx;
  color: #999999;
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
  margin-bottom: 20rpx;
}

.order-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
}

.item-image {
  width: 100rpx;
  height: 100rpx;
  border-radius: 8rpx;
  background: #f0f0f0;
  margin-right: 16rpx;
}

.item-info {
  flex: 1;
}

.item-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.item-spec {
  font-size: 24rpx;
  color: #999999;
}

.item-price {
  text-align: right;
}

.price {
  font-size: 28rpx;
  color: #1a1a1a;
  display: block;
}

.quantity {
  font-size: 24rpx;
  color: #999999;
}

.more-items {
  font-size: 24rpx;
  color: #999999;
  text-align: center;
  padding: 12rpx 0;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16rpx;
  border-top: 1rpx solid #f0f0f0;
}

.order-time {
  font-size: 24rpx;
  color: #999999;
}

.order-amount {
  font-size: 26rpx;
  color: #666666;
}

.amount {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
  margin-left: 8rpx;
}

.order-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #f0f0f0;
}

.action-btn {
  padding: 12rpx 32rpx;
  border-radius: 32rpx;
  font-size: 26rpx;
}

.action-btn.primary {
  background: #007AFF;
  color: #ffffff;
}

.action-btn.warning {
  background: #fff7e6;
  color: #fa8c16;
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

.verify-order-info {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
}

.order-no {
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 8rpx;
}

.order-amount {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 600;
}

.verify-code {
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
