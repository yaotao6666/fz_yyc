<template>
  <view class="container">
      <view class="card">
      <view class="title">商家 WebSocket 联调测试</view>
      <view class="subtitle">输入目标商家 ID，手动下发进店与订单通知，验证商家端 WebSocket 是否已连通</view>

      <view class="field">
        <view class="label">商家ID</view>
        <input v-model="merchantId" class="input" type="number" placeholder="请输入 merchant_id" />
      </view>

      <view class="field">
        <view class="label">订单号</view>
        <input v-model="orderNo" class="input" type="text" placeholder="订单提醒时使用，可选" />
      </view>

      <view class="field">
        <view class="label">访客 OpenID</view>
        <input v-model="visitorOpenId" class="input" type="text" placeholder="进店提醒时使用，可选" />
      </view>

      <view class="field">
        <view class="label">来源 source</view>
        <input v-model="visitSource" class="input" type="text" placeholder="默认 dev，可选" />
      </view>

      <button class="btn" :disabled="submitting" @click="sendOrderNotify">
        {{ submitting && currentAction === 'order' ? '发送中...' : '发送订单成功提醒' }}
      </button>

      <button class="btn secondary" :disabled="submitting" @click="sendVisitNotify">
        {{ submitting && currentAction === 'visit' ? '发送中...' : '发送顾客进店提醒' }}
      </button>

      <view v-if="lastResult" class="result-card" :class="{ success: lastResult.delivered > 0, warning: lastResult.delivered === 0 }">
        <view class="result-title">最近发送结果</view>
        <view class="result-row">类型：{{ lastResult.label }}</view>
        <view class="result-row">目标商家：{{ lastResult.merchantId }}</view>
        <view class="result-row">在线连接数：{{ lastResult.delivered }}</view>
        <view class="result-row">发送时间：{{ lastResult.sentAt }}</view>
        <view class="result-hint">
          {{ lastResult.delivered > 0 ? '目标商家已有在线 WebSocket 连接。' : '未发现在线连接，请确认目标商家已登录且已连接 WebSocket。' }}
        </view>
      </view>

      <view class="tips">
        <view class="tip">1. 先用目标商家账号登录并进入商家首页</view>
        <view class="tip">2. 确认商家首页已显示 WebSocket 连接状态</view>
        <view class="tip">3. 订单提醒会触发“新订单”Toast，进店提醒会触发“顾客浏览店铺”Toast</view>
        <view class="tip">4. 若声音开关开启，还会同步播放对应提示音</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { post } from '@api'

type NotifyAction = 'order' | 'visit'

interface NotifyResponse {
  delivered: number
}

interface SendResult {
  label: string
  merchantId: number
  delivered: number
  sentAt: string
}

const merchantId = ref('1')
const orderNo = ref('TEST-ORDER')
const visitorOpenId = ref('')
const visitSource = ref('dev')
const submitting = ref(false)
const currentAction = ref<NotifyAction | ''>('')
const lastResult = ref<SendResult | null>(null)

function getMerchantId(): number | null {
  const id = Number(merchantId.value)
  if (!id) {
    uni.showToast({ title: '请输入商家ID', icon: 'none' })
    return null
  }

  return id
}

function formatCurrentTime(): string {
  const now = new Date()
  const datePart = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
  const timePart = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  return `${datePart} ${timePart}`
}

function updateLastResult(label: string, id: number, delivered: number) {
  lastResult.value = {
    label,
    merchantId: id,
    delivered,
    sentAt: formatCurrentTime()
  }
}

async function sendOrderNotify() {
  const id = getMerchantId()
  if (!id) {
    return
  }

  currentAction.value = 'order'
  submitting.value = true
  try {
    const result = await post<NotifyResponse>('/api/v1/dev/order-notify', {
      merchant_id: id,
      order_no: orderNo.value.trim()
    })
    updateLastResult('订单成功提醒', id, result.delivered)
    uni.showToast({ title: result.delivered > 0 ? '发送成功' : '已发送但无在线连接', icon: 'none' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '发送失败', icon: 'none' })
  } finally {
    submitting.value = false
    currentAction.value = ''
  }
}

async function sendVisitNotify() {
  const id = getMerchantId()
  if (!id) {
    return
  }

  currentAction.value = 'visit'
  submitting.value = true
  try {
    const result = await post<NotifyResponse>('/api/v1/dev/store-visit-notify', {
      merchant_id: id,
      visitor_openid: visitorOpenId.value.trim(),
      source: visitSource.value.trim()
    })
    updateLastResult('顾客进店提醒', id, result.delivered)
    uni.showToast({ title: result.delivered > 0 ? '发送成功' : '已发送但无在线连接', icon: 'none' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '发送失败', icon: 'none' })
  } finally {
    submitting.value = false
    currentAction.value = ''
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}

.card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 32rpx;
}

.title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.subtitle {
  margin-top: 12rpx;
  margin-bottom: 28rpx;
  font-size: 24rpx;
  line-height: 1.7;
  color: #666666;
}

.field {
  margin-bottom: 22rpx;
}

.label {
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.input {
  height: 88rpx;
  border-radius: 14rpx;
  background: #f8f9fa;
  padding: 0 20rpx;
  font-size: 30rpx;
}

.btn {
  height: 96rpx;
  border-radius: 48rpx;
  background: #007AFF;
  color: #ffffff;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 10rpx;
}

.btn.secondary {
  background: #f0f5ff;
  color: #0056CC;
}

.btn[disabled] {
  background: #cccccc;
  color: #ffffff;
}

.result-card {
  margin-top: 28rpx;
  border-radius: 12rpx;
  padding: 24rpx;
  background: #f8f9fa;
}

.result-card.success {
  background: #f6ffed;
}

.result-card.warning {
  background: #fffbe6;
}

.result-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12rpx;
}

.result-row {
  font-size: 24rpx;
  color: #333333;
  line-height: 1.8;
}

.result-hint {
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.7;
  color: #666666;
}

.tips {
  margin-top: 28rpx;
  background: #f0f5ff;
  border-radius: 12rpx;
  padding: 20rpx;
}

.tip {
  font-size: 24rpx;
  color: #333333;
  margin-bottom: 10rpx;
}

.tip:last-child {
  margin-bottom: 0;
}
</style>
