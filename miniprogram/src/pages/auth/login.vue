<template>
  <view class="login-container">
    <view class="login-header">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <text class="title">寻梦私域管家</text>
      <text class="subtitle">商家管理平台</text>
    </view>

    <view class="login-form">
      <view class="form-item">
        <view class="label">账号</view>
        <input
          v-model="formData.username"
          type="text"
          placeholder="请输入商家账号"
          class="input"
          @blur="validateUsername"
        />
        <text v-if="errors.username" class="error-text">{{ errors.username }}</text>
      </view>

      <view class="form-item">
        <view class="label">密码</view>
        <input
          v-model="formData.password"
          type="password"
          placeholder="请输入密码"
          class="input"
          @blur="validatePassword"
        />
        <text v-if="errors.password" class="error-text">{{ errors.password }}</text>
      </view>

      <view class="actions">
        <button
          class="btn-login"
          :disabled="loading"
          @click="handleLogin"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <view class="link-group">
          <text class="link" @click="goRegister">商家入驻</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from from 'vue'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()

const formData = reactive({
  username: '',
  password: ''
})

const errors = reactive({
  username: '',
  password: ''
})

const loading = ref(false)

// 表单验证
function validateUsername() {
  if (!formData.username) {
    errors.username = '请输入账号'
    return false
  }
  errors.username = ''
  return true
}

function validatePassword() {
  if (!formData.password) {
    errors.password = '请输入密码'
    return false
  }
  if (formData.password.length < 6) {
    errors.password = '密码至少6位'
    return false
  }
  errors.password = ''
  return true
}

// 登录
async function handleLogin() {
  if (!validateUsername() || !validatePassword()) {
    return
  }

  loading.value = true

  try {
    const success = await authStore.login(formData.username, formData.password)

    if (success) {
      uni.showToast({ title: '登录成功', icon: 'success' })
      
      // 跳转到商户首页
      setTimeout(() => {
        uni.switchTab({ url: '/pages/merchant/home' })
      }, 500)
    }
  } finally {
    loading.value = false
  }
}

// 跳转注册
function goRegister() {
  uni.navigateTo({ url: '/pages/auth/register' })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  background: linear-gradient(180deg, #f8f9fa 0%, #ffffff 100%);
  padding: 120rpx 60rpx 60rpx;
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 80rpx;
}

.logo {
  width: 160rpx;
  height: 160rpx;
  margin-bottom: 30rpx;
  background: #f0f0f0;
  border-radius: 32rpx;
}

.title {
  font-size: 48rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.subtitle {
  font-size: 28rpx;
  color: #999999;
}

.login-form {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx 40rpx;
  box-shadow: 0 4rpx 24rpx rgba(0, 0, 0, 0.06);
}

.form-item {
  margin-bottom: 32rpx;
}

.label {
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 16rpx;
  font-weight: 500;
}

.input {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.input::placeholder {
  color: #cccccc;
}

.error-text {
  font-size: 24rpx;
  color: #ff4d4f;
  margin-top: 8rpx;
  display: block;
}

.actions {
  margin-top: 48rpx;
}

.btn-login {
  width: 100%;
  height: 96rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 48rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-login[disabled] {
  background: #cccccc;
  color: #ffffff;
}

.link-group {
  display: flex;
  justify-content: center;
  margin-top: 32rpx;
}

.link {
  font-size: 28rpx;
  color: #007AFF;
  padding: 16rpx;
}
</style>
