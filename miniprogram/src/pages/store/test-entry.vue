<template>
  <view class="test-container">
    <view class="header">
      <text class="title">🧪 商家店铺测试入口</text>
      <text class="subtitle">在小程序发布前用于测试用户下单流程</text>
    </view>

    <view class="content">
      <view class="input-section">
        <text class="label">请输入商家ID：</text>
        <input
          v-model="merchantId"
          class="input"
          type="number"
          placeholder="例如：1"
          @confirm="enterStore"
        />
        <text class="hint">当前测试商家ID：1</text>
      </view>

      <button class="btn-primary" @click="enterStore">进入商家店铺</button>

      <view class="info-section">
        <text class="info-title">📝 使用说明</text>
        <text class="info-text">
          1. 在上方输入框中输入商家ID（默认：1）\n
          2. 点击"进入商家店铺"按钮\n
          3. 将进入该商家的店铺首页\n
          4. 可以浏览商品、加入购物车、下单支付\n
          \n\n⚠️ 注意事项：\n
          • 支付功能需要在微信开发者工具中开启\n
          • 或使用模拟支付进行测试
        </text>
      </view>

      <view class="merchant-list">
        <text class="list-title">📋 可用商家列表</text>
        <view
          v-for="merchant in merchants"
          :key="merchant.id"
          class="merchant-item"
          @click="selectMerchant(merchant.id)"
        >
          <text class="merchant-name">{{ merchant.name }}</text>
          <text class="merchant-id">ID: {{ merchant.id }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const merchantId = ref('1')

const merchants = ref([
  { id: 1, name: '美味餐厅' },
  { id: 2, name: '示例商家B' },
  { id: 3, name: '示例商家C' }
])

function selectMerchant(id: number) {
  merchantId.value = String(id)
}

function enterStore() {
  if (!merchantId.value) {
    uni.showToast({ title: '请输入商家ID', icon: 'none' })
    return
  }

  uni.navigateTo({
    url: `/pages/store/home?merchant_id=${merchantId.value}`
  })
}
</script>

<style scoped>
.test-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 60rpx 40rpx;
}

.header {
  text-align: center;
  margin-bottom: 60rpx;
}

.title {
  display: block;
  font-size: 48rpx;
  font-weight: 600;
  color: #ffffff;
  margin-bottom: 16rpx;
}

.subtitle {
  display: block;
  font-size: 28rpx;
  color: rgba(255, 255, 255, 0.8);
}

.content {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx;
  box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.1);
}

.input-section {
  margin-bottom: 40rpx;
}

.label {
  display: block;
  font-size: 32rpx;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.input {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 32rpx;
  margin-bottom: 12rpx;
}

.hint {
  display: block;
  font-size: 24rpx;
  color: #999999;
}

.btn-primary {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 48rpx;
}

.info-section {
  background: #f0f5ff;
  border-radius: 16rpx;
  padding: 32rpx;
  margin-bottom: 40rpx;
}

.info-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.info-text {
  display: block;
  font-size: 26rpx;
  color: #666666;
  line-height: 1.8;
  white-space: pre-line;
}

.merchant-list {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 40rpx;
}

.list-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.merchant-item {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.merchant-name {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 500;
}

.merchant-id {
  font-size: 24rpx;
  color: #999999;
}
</style>
