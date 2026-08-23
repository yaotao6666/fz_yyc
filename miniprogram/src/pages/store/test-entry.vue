<template>
  <view class="test-container">
    <view class="header">
      <text class="title">🧪 商家店铺测试入口</text>
      <text class="subtitle">在小程序发布前用于测试用户下单流程</text>
    </view>

    <view class="content">
      <view class="input-section">
        <button class="btn-primary" @click="enterStore">进入商家店铺</button>
      </view>

      <view class="info-section">
        <text class="info-title">📝 使用说明</text>
        <text class="info-text">
          1. 点击"进入商家店铺"按钮\n
          2. 将进入店铺首页\n
          3. 可以浏览商品、加入购物车、下单支付\n
          \n\n⚠️ 注意事项：\n
          • 支付功能需要在微信开发者工具中开启\n
          • 或使用模拟支付进行测试
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
function enterStore() {
  uni.navigateTo({
    url: `/pages/store/home`
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
</style>
