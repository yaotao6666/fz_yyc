<template>
  <view class="order-list-container">
    <view class="filter-header-card">
      <view class="filter-header-top">
        <view>
          <view class="filter-title">订单筛选</view>
          <view class="filter-subtitle">{{ currentFilterSummary }}</view>
        </view>
        <view class="filter-highlight">{{ currentStatusLabel }}</view>
      </view>

      <view class="filter-section">
        <view class="filter-section-title">订单状态</view>
        <view class="status-tabs">
          <view
            v-for="tab in statusTabs"
            :key="tab.value"
            class="tab-item"
            :class="{ active: currentStatus === tab.value }"
            @click="changeStatus(tab.value)"
          >
            <text>{{ tab.label }}</text>
            <text v-if="tab.count" class="tab-count">{{ tab.count }}</text>
          </view>
        </view>
      </view>

      <view class="filter-section filter-section-date">
        <view class="filter-section-row">
          <view class="filter-section-title">时间范围</view>
          <view v-if="hasDateFilter" class="date-filter-tip">已按日期筛选</view>
        </view>

        <view class="date-range-tabs">
        <view
          v-for="tab in dateRangeTabs"
          :key="tab.value"
          class="date-range-tab"
          :class="{ active: currentDateRange === tab.value }"
          @click="changeDateRange(tab.value)"
        >
          {{ tab.label }}
        </view>
      </view>

        <view class="date-picker-row">
          <picker mode="date" :value="startDate" :end="maxDate" @change="handleStartDateChange">
            <view class="date-picker-field">
              <text class="date-picker-label">开始日期</text>
              <text class="date-picker-value">{{ startDate || '请选择' }}</text>
            </view>
          </picker>
          <view class="date-separator-wrap">
            <text class="date-separator">至</text>
          </view>
          <picker mode="date" :value="endDate" :end="maxDate" @change="handleEndDateChange">
            <view class="date-picker-field">
              <text class="date-picker-label">结束日期</text>
              <text class="date-picker-value">{{ endDate || '请选择' }}</text>
            </view>
          </picker>
        </view>

        <view class="date-filter-actions">
          <view class="date-filter-summary">{{ dateFilterSummary }}</view>
          <view class="date-reset-btn" @click="resetDateFilter">重置筛选</view>
        </view>
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
          <template v-if="order.status === OrderStatus.PAID">
            <view class="action-btn primary" @click="handleVerify(order)">核销</view>
          </template>
          <template v-if="order.status === OrderStatus.PAID || order.status === OrderStatus.COMPLETED">
            <view class="action-btn warning" @click="openRefundDialog(order)">发起退款</view>
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

    <!-- 退款弹窗 -->
    <view v-if="showRefund" class="dialog-mask" @click="closeRefundDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">发起退款</view>
        <view class="verify-order-info">
          <view class="order-no">订单号: {{ currentOrder?.order_no }}</view>
          <view class="order-amount">退款金额: ¥{{ currentOrder?.pay_amount?.toFixed(2) }}</view>
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
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getOrder, getOrders, completeOrder, refundOrder } from '@api'
import { OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'

type DateRangeValue = 'all' | 'today' | 'last7' | 'last30' | 'custom'
const ORDER_LIST_ROUTE_STATE_KEY = 'merchant_order_list_route_state'

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
const currentDateRange = ref<DateRangeValue>('all')
const startDate = ref('')
const endDate = ref('')
const maxDate = computed(() => formatDate(new Date()))
const dateRangeTabs: Array<{ label: string; value: DateRangeValue }> = [
  { label: '全部', value: 'all' },
  { label: '今日', value: 'today' },
  { label: '近7天', value: 'last7' },
  { label: '近30天', value: 'last30' },
  { label: '自定义', value: 'custom' }
]

const currentOrder = ref<Order | null>(null)
const verifying = ref(false)
const showRefund = ref(false)
const refundReason = ref('')
const refunding = ref(false)
const currentStatusLabel = computed(() => {
  return statusTabs.find(tab => tab.value === currentStatus.value)?.label || '全部'
})
const hasDateFilter = computed(() => !!startDate.value || !!endDate.value || currentDateRange.value !== 'all')
const dateFilterSummary = computed(() => {
  if (!hasDateFilter.value) {
    return '当前展示全部时间范围'
  }
  if (startDate.value || endDate.value) {
    return `${startDate.value || '不限'} 至 ${endDate.value || '不限'}`
  }
  return dateRangeTabs.find(tab => tab.value === currentDateRange.value)?.label || '自定义'
})
const currentFilterSummary = computed(() => `${currentStatusLabel.value} · ${dateFilterSummary.value}`)

onLoad((options: any) => {
  const status = Number(options?.status || 0)
  if (!Number.isNaN(status) && status > 0) {
    currentStatus.value = status
  }

  const initialStartDate = String(options?.start_date || '')
  const initialEndDate = String(options?.end_date || '')
  if (initialStartDate || initialEndDate) {
    startDate.value = initialStartDate
    endDate.value = initialEndDate
    currentDateRange.value = 'custom'
  }
})

onShow(() => {
  applyPendingRouteState()
  loadOrders(true)
  loadStatistics()
})

function applyPendingRouteState() {
  const rawState = uni.getStorageSync(ORDER_LIST_ROUTE_STATE_KEY)
  if (!rawState) {
    return
  }

  uni.removeStorageSync(ORDER_LIST_ROUTE_STATE_KEY)

  try {
    const routeState = typeof rawState === 'string' ? JSON.parse(rawState) : rawState
    const nextStatus = Number(routeState?.status || 0)
    if (!Number.isNaN(nextStatus) && nextStatus >= 0) {
      currentStatus.value = nextStatus
    }

    const nextStartDate = String(routeState?.start_date || '')
    const nextEndDate = String(routeState?.end_date || '')
    if (nextStartDate || nextEndDate) {
      startDate.value = nextStartDate
      endDate.value = nextEndDate
      currentDateRange.value = 'custom'
      return
    }

    currentDateRange.value = 'all'
    startDate.value = ''
    endDate.value = ''
  } catch (error) {
    console.error('解析订单列表跳转状态失败:', error)
  }
}

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

    if (startDate.value) {
      params.start_date = startDate.value
    }

    if (endDate.value) {
      params.end_date = endDate.value
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

function changeDateRange(range: DateRangeValue) {
  currentDateRange.value = range

  if (range === 'all') {
    startDate.value = ''
    endDate.value = ''
    loadOrders(true)
    return
  }

  if (range === 'custom') {
    if (!startDate.value && !endDate.value) {
      const today = formatDate(new Date())
      startDate.value = today
      endDate.value = today
    }
    loadOrders(true)
    return
  }

  const { start, end } = getPresetDateRange(range)
  startDate.value = start
  endDate.value = end
  loadOrders(true)
}

function handleStartDateChange(event: any) {
  startDate.value = event?.detail?.value || ''
  currentDateRange.value = 'custom'
  applyCustomDateFilter()
}

function handleEndDateChange(event: any) {
  endDate.value = event?.detail?.value || ''
  currentDateRange.value = 'custom'
  applyCustomDateFilter()
}

function resetDateFilter() {
  currentDateRange.value = 'all'
  startDate.value = ''
  endDate.value = ''
  loadOrders(true)
}

function applyCustomDateFilter() {
  if (!isDateRangeValid()) {
    return
  }
  loadOrders(true)
}

function isDateRangeValid() {
  if (startDate.value && endDate.value && startDate.value > endDate.value) {
    uni.showToast({ title: '开始日期不能晚于结束日期', icon: 'none' })
    return false
  }
  return true
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

function formatDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function getPresetDateRange(range: Exclude<DateRangeValue, 'all' | 'custom'>) {
  const end = new Date()
  const start = new Date(end)

  if (range === 'last7') {
    start.setDate(end.getDate() - 6)
  } else if (range === 'last30') {
    start.setDate(end.getDate() - 29)
  }

  return {
    start: formatDate(start),
    end: formatDate(end)
  }
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/merchant/orders/detail?id=${id}` })
}

async function handleVerify(order: Order) {
  currentOrder.value = order
  const code = order.verify_code?.trim() || ''
  if (!code) {
    try {
      const detail = await getOrder(order.id)
      if (!detail.verify_code) {
        return uni.showToast({ title: '未获取到核销码', icon: 'none' })
      }
      await submitVerify(detail.id, detail.verify_code)
    } catch (error: any) {
      uni.showToast({ title: error?.message || '核销失败', icon: 'none' })
    }
    return
  }
  await submitVerify(order.id, code)
}

async function submitVerify(orderId: number, verifyCode: string) {
  const code = verifyCode.trim()
  if (!/^\d{6}$/.test(code)) {
    return uni.showToast({ title: '核销码应为6位数字', icon: 'none' })
  }

  verifying.value = true
  try {
    const updatedOrder = await completeOrder(orderId, code)
    const index = orders.value.findIndex(item => item.id === orderId)
    if (index !== -1) {
      orders.value[index] = updatedOrder
    }
    uni.showToast({ title: '核销成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '核销失败', icon: 'none' })
  } finally {
    verifying.value = false
  }
}

function openRefundDialog(order: Order) {
  currentOrder.value = order
  refundReason.value = ''
  showRefund.value = true
}

function closeRefundDialog() {
  showRefund.value = false
  refundReason.value = ''
}

async function submitRefund() {
  if (!currentOrder.value) return
  refunding.value = true
  try {
    await refundOrder(currentOrder.value.id, {
      reason: refundReason.value.trim(),
      refund_amount: currentOrder.value.pay_amount
    })
    const index = orders.value.findIndex(item => item.id === currentOrder.value?.id)
    if (index !== -1) {
      orders.value[index].status = OrderStatus.REFUNDING
      orders.value[index].refunded_at = new Date().toISOString()
    }
    uni.showToast({ title: '退款已提交', icon: 'success' })
    closeRefundDialog()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '退款失败', icon: 'none' })
  } finally {
    refunding.value = false
  }
}
</script>

<style scoped>
.order-list-container {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  height: 140vh;
  box-sizing: border-box;
}

.filter-header-card {
  margin: 24rpx 24rpx 0;
  padding: 28rpx 24rpx 24rpx;
  border-radius: 24rpx;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 12rpx 32rpx rgba(0, 86, 204, 0.06);
  flex-shrink: 0;
}

.filter-header-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.filter-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.filter-subtitle {
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.6;
  color: #7a8699;
}

.filter-highlight {
  flex-shrink: 0;
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: #eaf3ff;
  color: #0056cc;
  font-size: 24rpx;
  font-weight: 600;
}

.filter-section {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 1rpx solid #eef3f9;
}

.filter-section-date {
  margin-top: 20rpx;
}

.filter-section-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #333333;
}

.filter-section-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.date-filter-tip {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #fff7e6;
  color: #b26a00;
  font-size: 22rpx;
}

.status-tabs {
  display: flex;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-top: 16rpx;
}

.tab-item {
  min-width: 128rpx;
  text-align: center;
  font-size: 26rpx;
  color: #666666;
  padding: 18rpx 20rpx;
  position: relative;
  border-radius: 18rpx;
  background: #f4f7fb;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.tab-item.active {
  color: #0056cc;
  font-weight: 600;
  background: #eaf3ff;
  box-shadow: inset 0 0 0 2rpx rgba(0, 122, 255, 0.12);
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

.date-filter-panel {
  padding: 0;
}

.date-range-tabs {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
  margin-top: 16rpx;
}

.date-range-tab {
  padding: 14rpx 24rpx;
  border-radius: 999rpx;
  background: #f4f7fb;
  color: #666666;
  font-size: 24rpx;
}

.date-range-tab.active {
  background: #eaf3ff;
  color: #0056cc;
  font-weight: 600;
}

.date-picker-row {
  margin-top: 20rpx;
  display: flex;
  align-items: stretch;
  gap: 12rpx;
}

.date-picker-field {
  min-width: 0;
  flex: 1;
  padding: 20rpx;
  border-radius: 18rpx;
  background: #f7f9fc;
  border: 1rpx solid #edf1f5;
}

.date-picker-label {
  display: block;
  font-size: 22rpx;
  color: #999999;
}

.date-picker-value {
  display: block;
  margin-top: 8rpx;
  font-size: 26rpx;
  color: #2f3a4a;
  font-weight: 500;
}

.date-separator-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
}

.date-separator {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #9aa4b2;
}

.date-filter-actions {
  margin-top: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.date-filter-summary {
  flex: 1;
  font-size: 24rpx;
  line-height: 1.6;
  color: #7a8699;
}

.date-reset-btn {
  flex-shrink: 0;
  padding: 0 28rpx;
  height: 68rpx;
  border-radius: 34rpx;
  background: #eef5ff;
  color: #0056cc;
  font-size: 24rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.order-list {
  width: calc(100% - 48rpx);
  padding: 24rpx 24rpx 0;
  flex: 1;
  overflow: hidden;
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
