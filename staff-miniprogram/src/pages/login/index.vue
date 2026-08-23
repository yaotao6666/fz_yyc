<template>
  <view class="login-page">
    <view class="login-header">
      <view class="logo-circle">财</view>
      <text class="app-title">财旭服务端</text>
      <text class="app-subtitle">服务人员接单系统</text>
    </view>

    <view class="login-form" v-if="mode === 'login'">
      <view class="form-item">
        <input
          v-model="loginForm.username"
          placeholder="请输入用户名"
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <view class="form-item">
        <input
          v-model="loginForm.password"
          placeholder="请输入密码"
          password
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <button class="btn-primary" @click="handleLogin" :loading="loading">登 录</button>
      <view class="form-footer">
        <text class="link-text" @click="mode = 'register'">没有账号？去注册</text>
      </view>
    </view>

    <view class="login-form" v-else>
      <view class="form-item">
        <input
          v-model="registerForm.name"
          placeholder="姓名"
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <view class="form-item">
        <input
          v-model="registerForm.phone"
          placeholder="手机号"
          type="number"
          maxlength="11"
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <view class="form-item">
        <input
          v-model="registerForm.username"
          placeholder="登录用户名"
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <view class="form-item">
        <input
          v-model="registerForm.password"
          placeholder="密码（至少6位）"
          password
          class="form-input"
          :adjust-position="false"
        />
      </view>
      <view class="form-item">
        <view class="form-label">资质材料（选填，加快审核）</view>
        <view class="qual-list">
          <view v-for="(q, i) in registerForm.qualifications" :key="i" class="qual-item">
            <input v-model="q.name" placeholder="材料名称（如执业证书）" class="qual-input" />
            <image v-if="q.url" :src="q.url" class="qual-img" mode="aspectFill" @click="previewQual(i)" />
            <text class="qual-remove" @click="registerForm.qualifications.splice(i, 1)">移除</text>
          </view>
          <view class="qual-add" @click="pickQualification">+ 添加材料</view>
        </view>
      </view>
      <button class="btn-primary" @click="handleRegister" :loading="loading">提交注册申请</button>
      <view class="form-footer">
        <text class="link-text" @click="mode = 'login'">已有账号？去登录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { staffAuthApi, uploadQualificationFile } from '@/api'
import { useStaffAuthStore } from '@/stores/auth'

const authStore = useStaffAuthStore()
const mode = ref<'login' | 'register'>('login')
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const registerForm = reactive({
  name: '',
  phone: '',
  username: '',
  password: '',
  qualifications: [] as { type: string; name: string; url: string }[]
})

function previewQual(index: number) {
  const q = registerForm.qualifications[index]
  if (q?.url) uni.previewImage({ urls: [q.url] })
}

function pickQualification() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: (res) => {
      const filePath = res.tempFilePaths[0]
      uploadQualificationFile(filePath)
        .then(({ url }) => {
          registerForm.qualifications.push({ type: 'certificate', name: '', url })
        })
        .catch(() => {})
    }
  })
}

async function handleLogin() {
  if (!loginForm.username || !loginForm.password) {
    uni.showToast({ title: '请填写用户名和密码', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const res: any = await staffAuthApi.login(loginForm.username, loginForm.password)
    if (res.code === 0 && res.data?.token) {
      authStore.setToken(res.data.token)
      authStore.setUser(res.data.staff)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({ url: '/pages/todo/index' })
      }, 1000)
    } else {
      uni.showToast({ title: res.message || '登录失败', icon: 'none' })
    }
  } catch (err: any) {
    const msg = err?.data?.message || err?.errMsg || '登录失败'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!registerForm.name || !registerForm.phone || !registerForm.username || !registerForm.password) {
    uni.showToast({ title: '请填写完整信息', icon: 'none' })
    return
  }
  if (registerForm.password.length < 6) {
    uni.showToast({ title: '密码至少6位', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const res: any = await staffAuthApi.register({
      username: registerForm.username,
      password: registerForm.password,
      name: registerForm.name,
      phone: registerForm.phone,
      qualifications: registerForm.qualifications
    })
    if (res.code === 0) {
      uni.showToast({ title: '注册申请已提交，请等待审核', icon: 'none', duration: 2500 })
      setTimeout(() => {
        mode.value = 'login'
        loginForm.username = registerForm.username
        loginForm.password = ''
      }, 2500)
    } else {
      uni.showToast({ title: res.message || '注册失败', icon: 'none' })
    }
  } catch (err: any) {
    const msg = err?.data?.message || err?.errMsg || '注册失败'
    uni.showToast({ title: msg, icon: 'none' })
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #517528 0%, #3a551c 40%, #f5f5f5 40%);
  padding: 0 48rpx;
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 100rpx;
  padding-bottom: 60rpx;
}

.logo-circle {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  border: 4rpx solid rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 56rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.app-title {
  font-size: 40rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 8rpx;
}

.app-subtitle {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.8);
}

.login-form {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx 36rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.08);
}

.form-item {
  margin-bottom: 28rpx;
}

.form-input {
  width: 100%;
  height: 88rpx;
  background: #f8f8f8;
  border: 2rpx solid #eee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  box-sizing: border-box;
}

.btn-primary {
  width: 100%;
  height: 92rpx;
  line-height: 92rpx;
  background: linear-gradient(135deg, #517528, #7a9e4f);
  color: #ffffff;
  font-size: 32rpx;
  font-weight: 600;
  border-radius: 12rpx;
  border: none;
  margin-top: 12rpx;
}

.btn-primary::after {
  border: none;
}

.form-footer {
  text-align: center;
  margin-top: 32rpx;
}

.link-text {
  font-size: 28rpx;
  color: #517528;
}

.form-label {
  font-size: 26rpx;
  color: #666;
  margin-bottom: 16rpx;
}

.qual-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.qual-item {
  width: 200rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.qual-input {
  width: 100%;
  height: 60rpx;
  background: #f8f8f8;
  border: 2rpx solid #eee;
  border-radius: 10rpx;
  padding: 0 16rpx;
  font-size: 22rpx;
  box-sizing: border-box;
  margin-bottom: 10rpx;
}

.qual-img {
  width: 200rpx;
  height: 150rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
}

.qual-remove {
  font-size: 22rpx;
  color: #f56c6c;
  margin-top: 8rpx;
}

.qual-add {
  width: 200rpx;
  height: 150rpx;
  border: 2rpx dashed #ccc;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 26rpx;
  background: #fafafa;
}
</style>
