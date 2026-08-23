<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { staffWorkorderApi } from '@/api'
import { OrderTypeText } from '@/types'
import { formatDate, fromNow } from '@/utils/format'

const orderId = ref<string>('')
const loading = ref(false)
const detail = ref<any>(null)

// 签退备注弹窗
const checkoutRemark = ref('')
const showCheckoutModal = ref(false)

const orderTypeText = computed(() => {
  if (!detail.value) return ''
  return OrderTypeText[detail.value.order_type as keyof typeof OrderTypeText] || '未知'
})

function getBizStatusText(status: number) {
  const map: Record<number, string> = {
    0: '无', 1: '待接单', 2: '待出发', 3: '服务中', 4: '待支付尾款', 5: '已完成', 6: '已取消'
  }
  return map[status] || '未知'
}

function getBizStatusClass(status: number) {
  const map: Record<number, string> = {
    1: 'status-pending', 2: 'status-warn', 3: 'status-info', 4: 'status-pending', 5: 'status-success', 6: 'status-cancel'
  }
  return map[status] || ''
}

// 是否可接单
const canAccept = computed(() => detail.value?.biz_status === 1)
// 是否可签到
const canCheckIn = computed(() => detail.value?.biz_status === 2)
// 是否可签退
const canCheckOut = computed(() => detail.value?.biz_status === 3)
// 是否可录入照护记录（服务中/待支付/已完成且关联客户）
const canRecordVisit = computed(() =>
  !!detail.value?.user_id && [3, 4, 5].includes(detail.value?.biz_status)
)

async function loadDetail() {
  if (!orderId.value) return
  loading.value = true
  try {
    const res: any = await staffWorkorderApi.getWorkorderDetail(orderId.value)
    if (res.code === 0) {
      detail.value = res.data
    } else {
      uni.showToast({ title: res.message || '加载失败', icon: 'none' })
    }
  } catch (e: any) {
    const msg = e?.data?.message || '加载失败'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function handleAccept() {
  uni.showModal({
    title: '确认接单',
    content: '接单后将出现在您的待办中，请及时前往服务',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const result: any = await staffWorkorderApi.acceptOrder(orderId.value)
        if (result.code === 0) {
          uni.showToast({ title: '接单成功', icon: 'success' })
          loadDetail()
        } else {
          uni.showToast({ title: result.message || '接单失败', icon: 'none' })
        }
      } catch (e: any) {
        uni.showToast({ title: e?.data?.message || '接单失败', icon: 'none' })
      }
    }
  })
}

function handleCheckIn() {
  uni.showModal({
    title: '确认签到',
    content: '签到后将开始记录服务时间，确认已到达服务地点？',
    success: async (res) => {
      if (!res.confirm) return
      // 尝试获取定位（失败也不阻塞，后端不强制要求）
      let lat: number | undefined
      let lng: number | undefined
      try {
        const loc: any = await new Promise((resolve, reject) => {
          uni.getLocation({ type: 'gcj02', success: resolve, fail: reject })
        })
        lat = loc.latitude
        lng = loc.longitude
      } catch {
        // 定位失败不阻塞签到
      }
      try {
        const result: any = await staffWorkorderApi.checkIn(orderId.value, lat, lng)
        if (result.code === 0) {
          uni.showToast({ title: '签到成功', icon: 'success' })
          loadDetail()
        } else {
          uni.showToast({ title: result.message || '签到失败', icon: 'none' })
        }
      } catch (e: any) {
        uni.showToast({ title: e?.data?.message || '签到失败', icon: 'none' })
      }
    }
  })
}

function openCheckoutModal() {
  checkoutRemark.value = ''
  showCheckoutModal.value = true
}

async function confirmCheckOut() {
  showCheckoutModal.value = false
  try {
    const result: any = await staffWorkorderApi.checkOut(orderId.value, checkoutRemark.value)
    if (result.code === 0) {
      uni.showToast({ title: '签退成功', icon: 'success' })
      loadDetail()
      // 签退成功后询问是否录入本次上门照护记录
      uni.showModal({
        title: '录入照护记录',
        content: '是否录入本次上门照护记录？',
        success: (res) => {
          if (res.confirm) goRecordVisit()
        }
      })
    } else {
      uni.showToast({ title: result.message || '签退失败', icon: 'none' })
    }
  } catch (e: any) {
    uni.showToast({ title: e?.data?.message || '签退失败', icon: 'none' })
  }
}

function callPhone(phone: string) {
  if (!phone) return
  uni.makePhoneCall({ phoneNumber: phone })
}

// 跳转录入照护记录（携带订单ID，无计划时走自由输入护理项）
function goRecordVisit() {
  if (!orderId.value) return
  uni.navigateTo({ url: `/pages/health/care-visit-form?orderId=${orderId.value}` })
}

// 跳转客户健康档案（居民健康档案/健康评估）
function goResident() {
  if (!detail.value?.user_id) return
  uni.navigateTo({ url: `/pages/health/resident?userId=${detail.value.user_id}` })
}

onLoad((options: any) => {
  orderId.value = options?.id || ''
  loadDetail()
})

onShow(() => {
  // 从其他页面返回时刷新
  if (orderId.value && detail.value) {
    loadDetail()
  }
})
</script>

<template>
  <view class="container" v-if="detail">
    <!-- 状态头部 -->
    <view class="status-header">
      <view class="status-row">
        <text class="order-type">{{ orderTypeText }}</text>
        <text :class="['status-badge', getBizStatusClass(detail.biz_status)]">
          {{ getBizStatusText(detail.biz_status) }}
        </text>
      </view>
      <view class="order-no">单号：{{ detail.order_no }}</view>
    </view>

    <!-- 客户信息 -->
    <view class="card section" v-if="detail.contact_name || detail.contact_phone || detail.delivery_address">
      <view class="section-title">客户信息</view>
      <view class="info-row" v-if="detail.contact_name">
        <text class="info-label">联系人</text>
        <text class="info-value">{{ detail.contact_name }}</text>
      </view>
      <view class="info-row" v-if="detail.contact_phone">
        <text class="info-label">联系电话</text>
        <view class="info-value-row">
          <text class="info-value">{{ detail.contact_phone }}</text>
          <text class="call-btn" @tap="callPhone(detail.contact_phone)">拨打</text>
        </view>
      </view>
      <view class="info-row" v-if="detail.delivery_address">
        <text class="info-label">服务地址</text>
        <text class="info-value">{{ detail.delivery_address }}</text>
      </view>
      <view class="info-row" v-if="detail.scheduled_at">
        <text class="info-label">预约时间</text>
        <text class="info-value">{{ formatDate(detail.scheduled_at) }}</text>
      </view>
    </view>

    <!-- 客户健康档案入口 -->
    <view class="card section" v-if="detail.user_id">
      <view class="record-entry" @tap="goResident">
        <view class="record-entry-info">
          <text class="record-entry-title">客户健康档案</text>
          <text class="record-entry-desc">查看健康档案与评估记录，可为客户开展健康评估</text>
        </view>
        <text class="record-entry-arrow">›</text>
      </view>
    </view>

    <!-- 商品/服务明细 -->
    <view class="card section" v-if="detail.items && detail.items.length">
      <view class="section-title">服务明细</view>
      <view
        v-for="item in detail.items"
        :key="item.id"
        class="goods-item"
      >
        <image v-if="item.image" class="goods-img" :src="item.image" mode="aspectFill" />
        <view class="goods-info">
          <text class="goods-name">{{ item.product_name }}</text>
          <view class="goods-meta">
            <text class="goods-price">¥{{ Number(item.price || 0).toFixed(2) }}</text>
            <text class="goods-qty">x{{ item.quantity }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 金额信息 -->
    <view class="card section">
      <view class="section-title">费用信息</view>
      <view class="info-row">
        <text class="info-label">商品金额</text>
        <text class="info-value">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</text>
      </view>
      <view class="info-row" v-if="detail.total_deposit > 0">
        <text class="info-label">押金</text>
        <text class="info-value">¥{{ Number(detail.total_deposit || 0).toFixed(2) }}</text>
      </view>
      <view class="info-row" v-if="detail.delivery_fee > 0">
        <text class="info-label">配送费</text>
        <text class="info-value">¥{{ Number(detail.delivery_fee || 0).toFixed(2) }}</text>
      </view>
      <view class="info-row" v-if="detail.discount_amount > 0">
        <text class="info-label">优惠</text>
        <text class="info-value discount">-¥{{ Number(detail.discount_amount || 0).toFixed(2) }}</text>
      </view>
      <view class="info-row total-row">
        <text class="info-label">实付金额</text>
        <text class="total-amount">¥{{ Number(detail.pay_amount || 0).toFixed(2) }}</text>
      </view>
    </view>

    <!-- 时间线 -->
    <view class="card section">
      <view class="section-title">服务进度</view>
      <view class="timeline">
        <view class="timeline-item" v-if="detail.paid_at">
          <view class="dot done"></view>
          <view class="timeline-content">
            <text class="timeline-label">用户支付</text>
            <text class="timeline-time">{{ formatDate(detail.paid_at) }}</text>
          </view>
        </view>
        <view class="timeline-item" v-if="detail.actual_started_at">
          <view class="dot done"></view>
          <view class="timeline-content">
            <text class="timeline-label">服务签到</text>
            <text class="timeline-time">{{ formatDate(detail.actual_started_at) }}</text>
          </view>
        </view>
        <view class="timeline-item" v-if="detail.actual_ended_at">
          <view class="dot done"></view>
          <view class="timeline-content">
            <text class="timeline-label">服务签退</text>
            <text class="timeline-time">{{ formatDate(detail.actual_ended_at) }}</text>
          </view>
        </view>
        <view class="timeline-item" v-if="detail.rental_returned_at">
          <view class="dot done"></view>
          <view class="timeline-content">
            <text class="timeline-label">租赁归还</text>
            <text class="timeline-time">{{ formatDate(detail.rental_returned_at) }}</text>
          </view>
        </view>
        <view class="timeline-item" v-if="detail.completed_at">
          <view class="dot done"></view>
          <view class="timeline-content">
            <text class="timeline-label">订单完成</text>
            <text class="timeline-time">{{ formatDate(detail.completed_at) }}</text>
          </view>
        </view>
        <view class="timeline-item" v-if="!detail.actual_started_at && !detail.actual_ended_at">
          <view class="dot pending"></view>
          <view class="timeline-content">
            <text class="timeline-label muted">等待服务开始</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 备注 -->
    <view class="card section" v-if="detail.remark || detail.rental_return_remark">
      <view class="section-title">备注信息</view>
      <view class="remark-text" v-if="detail.remark">
        <text class="info-label">用户备注</text>
        <text class="remark-content">{{ detail.remark }}</text>
      </view>
      <view class="remark-text" v-if="detail.rental_return_remark">
        <text class="info-label">归还/验机备注</text>
        <text class="remark-content">{{ detail.rental_return_remark }}</text>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="footer-bar" v-if="canAccept || canCheckIn || canCheckOut || canRecordVisit">
      <button v-if="canAccept" class="action-btn primary" @tap="handleAccept">立即接单</button>
      <button v-if="canCheckIn" class="action-btn primary" @tap="handleCheckIn">签到开始服务</button>
      <button v-if="canCheckOut" class="action-btn warn" @tap="openCheckoutModal">签退结束服务</button>
      <button v-if="canRecordVisit" class="action-btn outline" @tap="goRecordVisit">录入照护记录</button>
    </view>

    <!-- 签退备注弹窗 -->
    <view v-if="showCheckoutModal" class="modal-mask" @tap="showCheckoutModal = false">
      <view class="modal-content" @tap.stop>
        <view class="modal-title">签退确认</view>
        <view class="modal-desc">签退后服务将标记为已完成，请确认服务已结束</view>
        <textarea
          v-model="checkoutRemark"
          class="modal-textarea"
          placeholder="可填写服务备注或验机情况（选填）"
          maxlength="200"
        />
        <view class="modal-actions">
          <button class="modal-btn cancel" @tap="showCheckoutModal = false">取消</button>
          <button class="modal-btn confirm" @tap="confirmCheckOut">确认签退</button>
        </view>
      </view>
    </view>
  </view>

  <view v-else-if="!loading" class="empty">
    <text class="empty-title">工单不存在</text>
    <text class="empty-desc">请返回重试</text>
  </view>
</template>

<style lang="scss" scoped>
.status-header {
  background: var(--white);
  border-radius: var(--card-radius);
  padding: 32rpx;
  margin-bottom: 24rpx;
  box-shadow: var(--shadow-sm);
}
.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8rpx;
}
.order-type { font-size: 34rpx; font-weight: 600; color: #333; }
.status-badge {
  font-size: 24rpx;
  padding: 6rpx 20rpx;
  border-radius: 8rpx;
  &.status-pending { background: #fff7e6; color: #fa8c16; }
  &.status-warn { background: #fff2e8; color: #ff9500; }
  &.status-info { background: #e6f4ff; color: #1989fa; }
  &.status-success { background: #e8f8ee; color: #07c160; }
  &.status-cancel { background: #f5f5f5; color: #999; }
}
.order-no { font-size: 24rpx; color: #999; }

.section {
  .section-title {
    font-size: 28rpx;
    font-weight: 600;
    color: #333;
    margin-bottom: 20rpx;
    padding-bottom: 16rpx;
    border-bottom: 2rpx solid var(--border-color);
  }
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12rpx 0;
}
.info-label { font-size: 28rpx; color: #666; flex-shrink: 0; }
.info-value { font-size: 28rpx; color: #333; text-align: right; }
.info-value-row { display: flex; align-items: center; gap: 16rpx; }
.call-btn {
  font-size: 24rpx;
  color: var(--primary-color);
  padding: 4rpx 16rpx;
  border: 2rpx solid var(--primary-color);
  border-radius: 8rpx;
}
.record-entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8rpx 0;
}
.record-entry-info {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.record-entry-title {
  font-size: 30rpx;
  font-weight: 600;
  color: var(--primary-color);
}
.record-entry-desc {
  font-size: 24rpx;
  color: #999;
}
.record-entry-arrow {
  font-size: 36rpx;
  color: #ccc;
}
.discount { color: #ff4d4f; }
.total-row { padding-top: 20rpx; border-top: 2rpx solid var(--border-color); margin-top: 8rpx; }
.total-amount { font-size: 34rpx; font-weight: 600; color: #ff4d4f; }

.goods-item {
  display: flex;
  gap: 20rpx;
  padding: 16rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
}
.goods-img { width: 120rpx; height: 120rpx; border-radius: 8rpx; flex-shrink: 0; }
.goods-info { flex: 1; display: flex; flex-direction: column; justify-content: space-between; }
.goods-name { font-size: 28rpx; color: #333; line-height: 1.4; }
.goods-meta { display: flex; justify-content: space-between; }
.goods-price { font-size: 28rpx; color: #ff4d4f; }
.goods-qty { font-size: 26rpx; color: #999; }

.timeline { padding-left: 8rpx; }
.timeline-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  padding-bottom: 28rpx;
  position: relative;
  &:not(:last-child)::before {
    content: '';
    position: absolute;
    left: 11rpx;
    top: 28rpx;
    bottom: 0;
    width: 2rpx;
    background: var(--border-color);
  }
}
.dot {
  width: 24rpx; height: 24rpx;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 4rpx;
  &.done { background: var(--primary-color); }
  &.pending { background: #ddd; }
}
.timeline-content { display: flex; flex-direction: column; gap: 4rpx; }
.timeline-label { font-size: 28rpx; color: #333; &.muted { color: #999; } }
.timeline-time { font-size: 24rpx; color: #999; }

.remark-text { padding: 12rpx 0; }
.remark-content { display: block; font-size: 28rpx; color: #333; margin-top: 8rpx; line-height: 1.5; }

.footer-bar {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  display: flex;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}
.action-btn {
  flex: 1;
  font-size: 30rpx;
  color: #fff;
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &.primary { background: var(--primary-color); }
  &.warn { background: var(--warning-color); }
  &.outline {
    background: #fff;
    color: var(--primary-color);
    border: 2rpx solid var(--primary-color);
  }
}

.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal-content {
  width: 600rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx;
}
.modal-title { font-size: 32rpx; font-weight: 600; color: #333; text-align: center; margin-bottom: 16rpx; }
.modal-desc { font-size: 26rpx; color: #666; text-align: center; margin-bottom: 24rpx; }
.modal-textarea {
  width: 100%;
  height: 160rpx;
  background: var(--bg-color);
  border-radius: 8rpx;
  padding: 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
  margin-bottom: 32rpx;
}
.modal-actions { display: flex; gap: 24rpx; }
.modal-btn {
  flex: 1;
  font-size: 30rpx;
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &.cancel { background: var(--bg-color); color: #666; }
  &.confirm { background: var(--primary-color); color: #fff; }
}

.empty {
  text-align: center;
  padding: 200rpx 32rpx;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}
</style>
