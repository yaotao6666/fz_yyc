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

    <view class="section" v-if="order.assigned_staff_name">
      <view class="section-title">服务人员</view>
      <view class="info-row">
        <text class="label">指派服务人员</text>
        <text class="value">{{ order.assigned_staff_name }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">配送信息</view>
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
      <view class="info-row" v-if="isRentalOrder && order.rental_end_at">
        <text class="label">租赁到期</text>
        <text class="value" :class="{ overdue: isRentalOverdue }">
          {{ formatDateTime(order.rental_end_at) }}（{{ rentalDueText }}）
        </text>
      </view>
      <view class="info-row" v-if="isServiceBizOrder && order.biz_status">
        <text class="label">服务状态</text>
        <text class="value">{{ bizStatusText }}</text>
      </view>
      <view class="info-row" v-if="order.refunded_at">
        <text class="label">退款时间</text>
        <text class="value">{{ formatDateTime(order.refunded_at) }}</text>
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
          <view class="goods-name">
            {{ item.product_name }}
            <text v-if="Number(item.sale_type) === 2" class="goods-rental-tag">租赁</text>
          </view>
          <view class="goods-spec" v-if="item.specs">{{ item.specs }}</view>
          <view class="goods-rental-info" v-if="Number(item.sale_type) === 2">
            <text class="rental-info-text">¥{{ Number(item.unit_rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(item.rental_unit) }} × {{ item.rental_duration || 0 }}{{ getRentalUnitText(item.rental_unit) }}</text>
            <text class="rental-info-deposit">押金 ¥{{ Number(item.deposit || 0).toFixed(2) }} × {{ item.quantity }}</text>
          </view>
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
      <view v-if="Number(order.total_deposit || 0) > 0" class="amount-row deposit-row">
        <text class="amount-label">押金</text>
        <text class="amount-value deposit-value">¥{{ Number(order.total_deposit || 0).toFixed(2) }}</text>
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

      <view v-if="isRentalOrder" class="rental-deposit-block">
        <view class="rental-deposit-title">押金状态</view>
        <view class="amount-row">
          <text class="amount-label">状态</text>
          <text class="amount-value">{{ getDepositStatusText(order.deposit_status) }}</text>
        </view>
        <view v-if="Number(order.deposit_deduct_amount || 0) > 0" class="amount-row">
          <text class="amount-label">扣除金额</text>
          <text class="amount-value">¥{{ Number(order.deposit_deduct_amount || 0).toFixed(2) }}</text>
        </view>
        <view v-if="Number(order.deposit_refund_amount || 0) > 0" class="amount-row">
          <text class="amount-label">退还金额</text>
          <text class="amount-value">¥{{ Number(order.deposit_refund_amount || 0).toFixed(2) }}</text>
        </view>
        <view v-if="order.deposit_refunded_at" class="amount-row">
          <text class="amount-label">退还时间</text>
          <text class="amount-value">{{ order.deposit_refunded_at }}</text>
        </view>
        <view v-if="order.rental_returned_at" class="amount-row">
          <text class="amount-label">归还时间</text>
          <text class="amount-value">{{ order.rental_returned_at }}</text>
        </view>
        <view v-if="order.rental_return_remark" class="amount-row">
          <text class="amount-label">归还备注</text>
          <text class="amount-value">{{ order.rental_return_remark }}</text>
        </view>
      </view>
    </view>

    <view class="bottom-bar">
      <button v-if="canCancel" class="btn secondary" :disabled="submitting" @click="cancelOrder">
        取消订单
      </button>
      <button v-if="canRefund" class="btn warning" :disabled="submitting" @click="contactMerchantForRefund">
        联系商家退款
      </button>
      <button v-if="canRenew" class="btn primary" :disabled="submitting" @click="renewOrderNow">
        续租
      </button>
    </view>

  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { cancelMyOrder, getMyOrderDetail, renewOrder } from '@api'
import { OrderStatus, OrderStatusText } from '@types'
import type { Order } from '@types'
import { BrandAsset } from '../../utils/constants'
import { useAuth } from '../../utils/useAuth'

const order = ref<Order | null>(null)
const submitting = ref(false)

const canCancel = computed(() => order.value?.status === OrderStatus.PENDING_PAYMENT)
const canRefund = computed(() => {
  return order.value?.status === OrderStatus.PAID
})
const isRentalOrder = computed(() => Number(order.value?.total_deposit || 0) > 0)
const isServiceBizOrder = computed(() => Number(order.value?.biz_status || 0) > 0)
// 租赁订单：已支付且未归还时可续租
const canRenew = computed(() => {
  if (!order.value) return false
  if (!isRentalOrder.value) return false
  if (order.value.status !== OrderStatus.PAID) return false
  return !order.value.rental_returned_at
})
const isRentalOverdue = computed(() => {
  const end = order.value?.rental_end_at
  if (!end) return false
  return new Date(end).getTime() < Date.now()
})
const rentalDueText = computed(() => {
  const end = order.value?.rental_end_at
  if (!end) return ''
  const remain = new Date(end).getTime() - Date.now()
  if (remain <= 0) return '已逾期'
  const days = Math.ceil(remain / 86400000)
  if (days <= 1) return '今日到期'
  return `剩余${days}天`
})
const bizStatusText = computed(() => {
  return { 1: '待接单', 2: '已指派/待出发', 3: '服务中', 4: '待支付尾款', 5: '已完成', 6: '已取消' }[Number(order.value?.biz_status)] || ''
})

function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}

function getDepositStatusText(status?: number): string {
  return { 1: '待退还', 2: '已退还', 3: '已扣除' }[Number(status || 0)] || '未收取'
}

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
    [OrderStatus.PAID]: '订单已支付，等待配送',
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

async function renewOrderNow() {
  if (!order.value) return
  uni.showModal({
    title: '确认续租',
    content: '将按原租赁时长生成续租订单并跳转支付，是否继续？',
    success: async (res) => {
      if (!res.confirm || !order.value) return
      submitting.value = true
      uni.showLoading({ title: '生成续租订单...' })
      try {
        const renewRes = await renewOrder(order.value.id)
        const payParams = renewRes.pay_params
        if (payParams) {
          uni.requestPayment({
            provider: 'wxpay',
            timeStamp: payParams.timeStamp,
            nonceStr: payParams.nonceStr,
            package: payParams.package,
            signType: payParams.signType,
            paySign: payParams.paySign,
            success: () => {
              uni.hideLoading()
              uni.showToast({ title: '续租支付成功', icon: 'success' })
              setTimeout(() => {
                uni.redirectTo({ url: '/pages/store/my-orders?status=2' })
              }, 1500)
            },
            fail: (err) => {
              uni.hideLoading()
              if (err.errMsg?.includes('cancel')) {
                uni.showToast({ title: '支付已取消', icon: 'none' })
                setTimeout(() => {
                  uni.redirectTo({ url: '/pages/store/my-orders?status=1' })
                }, 1200)
              } else {
                uni.showToast({ title: '支付失败', icon: 'none' })
              }
            }
          })
        } else {
          uni.hideLoading()
          if (Number(renewRes.order?.pay_amount || 0) > 0) {
            uni.showModal({
              title: '续租订单已生成',
              content: renewRes.pay_hint || '商家支付配置未完成，请稍后在「待付款」中支付',
              showCancel: false,
              confirmText: '知道了',
              success: () => {
                uni.redirectTo({ url: '/pages/store/my-orders?status=1' })
              }
            })
          } else {
            uni.showToast({ title: '续租订单已生成', icon: 'success' })
            setTimeout(() => {
              uni.redirectTo({ url: '/pages/store/my-orders' })
            }, 1200)
          }
        }
      } catch (error: any) {
        uni.hideLoading()
        uni.showToast({ title: error.message || '续租失败', icon: 'none' })
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

.goods-rental-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.goods-rental-info {
  display: flex;
  flex-direction: column;
  margin-top: 8rpx;
  padding: 8rpx 12rpx;
  background: #fff7e6;
  border-radius: 8rpx;
  font-size: 22rpx;
}

.rental-info-text {
  color: #ff9500;
}

.rental-info-deposit {
  color: #86909c;
  margin-top: 4rpx;
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

.deposit-row .amount-label {
  color: #ff9500;
}

.deposit-value {
  color: #ff9500;
}

.overdue {
  color: #f53f3f;
}

.rental-deposit-block {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 1rpx dashed #ffd591;
}

.rental-deposit-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #d46b08;
  margin-bottom: 16rpx;
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

</style>
