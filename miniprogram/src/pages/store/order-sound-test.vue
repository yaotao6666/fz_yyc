<template>
  <view class="container">
    <view class="card">
      <view class="title">订单提示音测试</view>

      <view class="field">
        <view class="label">商家ID</view>
        <input v-model="merchantId" class="input" type="number" placeholder="请输入 merchant_id" />
      </view>

      <view class="field">
        <view class="label">订单号</view>
        <input v-model="orderNo" class="input" type="text" placeholder="可选" />
      </view>

      <button class="btn" :disabled="submitting" @click="sendNotify">
        {{ submitting ? '发送中...' : '发送提醒' }}
      </button>

      <view class="tips">
        <view class="tip">1. 先用商家账号登录并保持小程序在前台</view>
        <view class="tip">2. 确保“设置-订单提示音”已开启</view>
        <view class="tip">3. 点击发送后，商家端会收到提示音与Toast</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { post } from '@api'

const merchantId = ref('1')
const orderNo = ref('TEST-ORDER')
const submitting = ref(false)

async function sendNotify() {
  const id = Number(merchantId.value)
  if (!id) {
    uni.showToast({ title: '请输入商家ID', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    await post('/api/v1/dev/order-notify', {
      merchant_id: id,
      order_no: orderNo.value.trim()
    })
    uni.showToast({ title: '已发送', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e?.message || '发送失败', icon: 'none' })
  } finally {
    submitting.value = false
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
  margin-bottom: 28rpx;
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

.btn[disabled] {
  background: #cccccc;
  color: #ffffff;
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

