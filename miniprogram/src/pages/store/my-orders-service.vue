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
          <text v-if="order.can_review" class="review-badge">待评价</text>
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
            <text class="order-summary">服务对象：{{ order.record_name || '未派单' }}</text>
            <text class="order-summary">服务地址：{{ getDeliveryAddressText(order) || '未填写' }}</text>
          </view>
          <view class="order-amount">
            <view class="amount-summary-row">
              <text class="amount-label">商品金额</text>
              <text class="amount-summary-value">¥{{ order.total_amount.toFixed(2) }}</text>
            </view>
            <view class="amount-summary-row">
              <text class="amount-label">优惠</text>
              <text class="amount-summary-value discount">
                {{ order.discount_amount > 0 ? `-${order.discount_amount.toFixed(2)}` : '¥0.00' }}
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
            <view class="action-btn refund" @click="contactMerchantForRefund(order)">联系商家退款</view>
          </template>
          <view v-if="order.can_review" class="action-btn review" @click="goReview(order)">去评价</view>
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
  </view>
</template>

<script setup lang="ts">
import { BrandAsset } from '../../utils/constants'
import { useOrderList } from './composables/useOrderList'

const {
  statusTabs,
  currentStatus,
  orders,
  loading,
  noMore,
  loadMore,
  changeStatus,
  getStatusText,
  getStatusClass,
  formatTime,
  getOrderItemImage,
  getDeliveryAddressText,
  goDetail,
  goReview,
  goShopping,
  cancelOrder,
  contactMerchantForRefund
} = useOrderList(2) // 固定服务订单
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

.order-summary {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #666;
  line-height: 1.5;
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

.action-btn.refund {
  background: #fff1f0;
  color: #ff4d4f;
}

.action-btn.review {
  background: #e6f7ff;
  color: #007AFF;
}

.review-badge {
  font-size: 22rpx;
  color: #ff4d4f;
  background: #fff1f0;
  padding: 2rpx 12rpx;
  border-radius: 8rpx;
  margin-left: 12rpx;
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
</style>