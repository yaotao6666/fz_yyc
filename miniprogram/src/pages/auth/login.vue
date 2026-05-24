<template>
  <view class="login-container">
    <view class="login-header">
      <image class="logo" :src="BrandAsset.APP_LOGO" mode="aspectFit" />
      <text class="title">{{ APP_NAME }}</text>
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
          password
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
        <button
          class="btn-wechat-login"
          :disabled="wechatLoading"
          @click="handleWechatLogin"
        >
          {{ wechatLoading ? '登录中...' : '微信快捷登录' }}
        </button>
      </view>

      <view class="agreement-row" @click="toggleAgreement">
        <view class="agreement-checkbox" :class="{ checked: agreementChecked }">
          <text v-if="agreementChecked" class="agreement-checkbox-icon">✓</text>
        </view>
        <view class="agreement-tip">
          <text class="agreement-text">我已阅读并同意</text>
          <text class="agreement-link" @click="goToAgreementPage">《服务协议&隐私政策》</text>
        </view>
      </view>
    </view>
  </view>
</template>


<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { APP_NAME } from '../../config/env'
import { BrandAsset } from '../../utils/constants'

const authStore = useAuthStore()

const formData = reactive({
  username: '', // 默认账号
  password: ''
})

const errors = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const wechatLoading = ref(false)
const agreementChecked = ref(false)

function goToAgreementPage() {
  uni.navigateTo({ url: '/pages/auth/agreement' })
}

function toggleAgreement() {
  agreementChecked.value = !agreementChecked.value
}

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

function validateAgreement() {
  if (!agreementChecked.value) {
    uni.showToast({ title: '请先阅读并同意商家服务协议&隐私政策', icon: 'none' })
    return false
  }
  return true
}

// 登录
async function handleLogin() {
  if (!validateUsername() || !validatePassword() || !validateAgreement()) {
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

async function handleWechatLogin() {
  if (!validateAgreement()) {
    return
  }

  wechatLoading.value = true

  try {
    const success = await authStore.loginWithWechat()
    if (success) {
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({ url: '/pages/merchant/home' })
      }, 500)
    }
  } finally {
    wechatLoading.value = false
  }
}

</script>

<style>
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

.agreement-row {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-top: 28rpx;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #f8f9fa;
}

.agreement-checkbox {
  width: 34rpx;
  height: 34rpx;
  border-radius: 50%;
  border: 2rpx solid #cfd5dd;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.agreement-checkbox.checked {
  border-color: #007AFF;
  background: #007AFF;
}

.agreement-checkbox-icon {
  font-size: 20rpx;
  color: #ffffff;
  line-height: 1;
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

.btn-wechat-login {
  width: 100%;
  height: 96rpx;
  margin-top: 20rpx;
  background: #f6ffed;
  border: 2rpx solid #52c41a;
  border-radius: 48rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #389e0d;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-wechat-login[disabled] {
  opacity: 0.6;
}

.agreement-tip {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  flex: 1;
}

.agreement-text {
  font-size: 24rpx;
  color: #7a7f87;
}

.agreement-link {
  font-size: 24rpx;
  color: #007AFF;
  padding-left: 4rpx;
}
</style>
