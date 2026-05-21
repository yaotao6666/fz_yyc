<template>
  <view class="login-container">
    <view class="login-header">
      <image class="logo" :src="BrandAsset.APP_LOGO" mode="aspectFit" />
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
          password
          placeholder="请输入密码"
          class="input"
          @blur="validatePassword"
        />
        <text v-if="errors.password" class="error-text">{{ errors.password }}</text>
      </view>

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
        <view class="agreement-tip">
          <text class="agreement-text">登录即表示同意</text>
          <text class="agreement-link" @click="showAgreement('service')">《商家服务协议》</text>
          <text class="agreement-text">和</text>
          <text class="agreement-link" @click="showAgreement('privacy')">《隐私政策》</text>
        </view>
      </view>

    <view v-if="agreementVisible" class="agreement-dialog-mask" @click="closeAgreement">
      <view class="agreement-dialog" @click.stop>
        <view class="agreement-dialog-header">
          <text class="agreement-dialog-title">{{ agreementTitle }}</text>
          <text class="agreement-dialog-close" @click="closeAgreement">×</text>
        </view>
        <scroll-view class="agreement-dialog-body" scroll-y>
          <rich-text :nodes="agreementContent"></rich-text>
        </scroll-view>
        <view class="agreement-dialog-footer">
          <view class="agreement-dialog-btn" @click="closeAgreement">我已知晓</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { BrandAsset } from '../../utils/constants'

const authStore = useAuthStore()

const formData = reactive({
  username: '',
  password: '123123'
})

const errors = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const wechatLoading = ref(false)
const agreementVisible = ref(false)
const agreementTitle = ref('')
const agreementContent = ref('')

const serviceAgreement = `
<span class="agreement-heading">一、服务条款</span>
<div>1.1 本协议是您与寻梦私域管家平台之间关于使用商家管理服务所订立的协议。请您仔细阅读本协议，在确认充分理解并同意后再开始使用。</div>
<div>1.2 商家账号由服务商统一创建并分配，您应妥善保管账号和密码，因您个人原因导致的账号泄露由您自行承担相关责任。</div>
<span class="agreement-heading">二、服务内容</span>
<div>2.1 平台为商家提供商品管理、订单管理、数据分析、配送管理等经营辅助服务。</div>
<div>2.2 平台有权根据业务发展需要调整、变更服务内容，并将及时通知商家。</div>
<span class="agreement-heading">三、商家义务</span>
<div>3.1 商家应确保所售商品符合国家法律法规要求，不得销售违禁、假冒伪劣商品。</div>
<div>3.2 商家应保证商品信息真实、准确，不得进行虚假宣传。</div>
<div>3.3 商家应及时处理订单，保障消费者合法权益。</div>
<span class="agreement-heading">四、费用与结算</span>
<div>4.1 平台服务费用及结算规则由服务商与商家另行约定。</div>
<div>4.2 商家应按照约定及时缴纳相关服务费用。</div>
<span class="agreement-heading">五、违约责任</span>
<div>5.1 任何一方违反本协议约定，应承担相应的违约责任。</div>
<div>5.2 如商家存在严重违规行为，平台有权暂停或终止服务。</div>
<span class="agreement-heading">六、协议变更</span>
<div>6.1 平台有权根据需要修改本协议条款，修改后的协议将在平台公示后生效。</div>
`

const privacyPolicy = `
<span class="agreement-heading">一、信息收集</span>
<div>1.1 我们将收集您在使用服务时主动提供的信息，包括但不限于：商家名称、联系方式、经营地址等。</div>
<div>1.2 我们会自动收集您的设备信息、操作日志等用于保障服务安全和提升用户体验。</div>
<span class="agreement-heading">二、信息使用</span>
<div>2.1 我们收集的信息将用于：为您提供商家管理服务、处理订单交易、改进产品功能、保障账户安全。</div>
<div>2.2 未经您的同意，我们不会将您的信息用于本协议约定以外的用途。</div>
<span class="agreement-heading">三、信息保护</span>
<div>3.1 我们采用业界通行的安全技术手段保护您的个人信息安全。</div>
<div>3.2 我们将对员工接触个人信息进行严格限制，并要求相关员工保密。</div>
<span class="agreement-heading">四、信息共享</span>
<div>4.1 未经您的同意，我们不会向第三方共享您的个人信息，法律法规另有规定的除外。</div>
<div>4.2 为完成交易需要，我们会在必要范围内向支付机构、物流服务商等共享订单相关信息。</div>
<span class="agreement-heading">五、您的权利</span>
<div>5.1 您有权访问、更正您的个人信息，也有权撤回授权同意或删除账号。</div>
<div>5.2 如您发现个人信息被违规使用，可联系我们进行处理。</div>
<span class="agreement-heading">六、未成年人保护</span>
<div>6.1 我们不会主动收集未成年人个人信息。如发现误收集，将及时删除。</div>
`

function showAgreement(type: string) {
  if (type === 'service') {
    agreementTitle.value = '商家服务协议'
    agreementContent.value = serviceAgreement
  } else {
    agreementTitle.value = '隐私政策'
    agreementContent.value = privacyPolicy
  }
  agreementVisible.value = true
}

function closeAgreement() {
  agreementVisible.value = false
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

async function handleWechatLogin() {
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
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  margin-top: 32rpx;
}

.agreement-text {
  font-size: 24rpx;
  color: #999999;
}

.agreement-link {
  font-size: 24rpx;
  color: #007AFF;
  padding: 12rpx 0 12rpx 4rpx;
}

.agreement-dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.agreement-dialog {
  width: 640rpx;
  max-height: 80vh;
  background: #ffffff;
  border-radius: 24rpx;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.agreement-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx 32rpx 20rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.agreement-dialog-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.agreement-dialog-close {
  font-size: 44rpx;
  color: #999999;
  padding: 0 8rpx;
  line-height: 1;
}

.agreement-dialog-body {
  padding: 24rpx 32rpx;
  max-height: 60vh;
  overflow-y: auto;
  font-size: 26rpx;
  color: #666666;
  line-height: 1.8;
  width: calc(90vw - 64rpx);
}

.agreement-dialog-body .agreement-heading {
  font-size: 28rpx;
  font-weight: 600;
  color: #333333;
  margin: 20rpx 0 12rpx;
}

.agreement-dialog-footer {
  padding: 20rpx 32rpx 32rpx;
}

.agreement-dialog-btn {
  height: 80rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 40rpx;
  font-size: 30rpx;
  font-weight: 500;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
