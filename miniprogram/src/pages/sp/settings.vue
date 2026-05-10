<template>
  <view class="settings-container">
    <!-- 页面标题 -->
    <view class="page-header">
      <text class="header-title">服务商设置</text>
    </view>

    <!-- 加载状态 -->
    <view v-if="loading" class="loading-container">
      <text class="loading-text">加载中...</text>
    </view>

    <template v-else>
      <!-- 服务商信息 -->
      <view class="section">
        <view class="section-title">服务商信息</view>

        <view class="info-item">
          <view class="info-label">服务商名称</view>
          <view class="info-value">{{ spInfo.name || '未设置' }}</view>
        </view>

        <view class="info-item">
          <view class="info-label">管理员姓名</view>
          <view class="info-value">{{ spInfo.admin_name || '未设置' }}</view>
        </view>

        <view class="info-item">
          <view class="info-label">联系方式</view>
          <view class="info-value">{{ spInfo.contact_phone || '未设置' }}</view>
        </view>
      </view>

      <!-- 功能设置 -->
      <view class="section">
        <view class="section-title">功能设置</view>

        <view class="setting-item" @click="handleChangePassword">
          <view class="setting-left">
            <view class="setting-icon">
              <text>🔑</text>
            </view>
            <view class="setting-info">
              <view class="setting-label">修改密码</view>
            </view>
          </view>
          <text class="arrow">›</text>
        </view>

        <view class="setting-item" @click="handleClearCache">
          <view class="setting-left">
            <view class="setting-icon">
              <text>🗑️</text>
            </view>
            <view class="setting-info">
              <view class="setting-label">清除缓存</view>
              <view class="setting-value">{{ cacheSize }}</view>
            </view>
          </view>
          <text class="arrow">›</text>
        </view>

        <view class="setting-item">
          <view class="setting-left">
            <view class="setting-icon">
              <text>ℹ️</text>
            </view>
            <view class="setting-info">
              <view class="setting-label">版本信息</view>
              <view class="setting-value">{{ version }}</view>
            </view>
          </view>
        </view>
      </view>

      <!-- 退出登录 -->
      <view class="logout-area">
        <button class="btn-logout" :disabled="logoutLoading" @click="handleLogout">
          {{ logoutLoading ? '退出中...' : '退出登录' }}
        </button>
      </view>
    </template>

    <!-- 修改密码弹窗 -->
    <view v-if="showPasswordModal" class="modal-overlay" @click="closePasswordModal">
      <view class="modal-content" @click.stop>
        <view class="modal-header">
          <text class="modal-title">修改密码</text>
          <text class="modal-close" @click="closePasswordModal">×</text>
        </view>

        <view class="modal-body">
          <view class="form-item">
            <view class="form-label">旧密码</view>
            <input
              v-model="passwordForm.old_password"
              type="password"
              password
              placeholder="请输入旧密码"
              class="form-input"
            />
          </view>

          <view class="form-item">
            <view class="form-label">新密码</view>
            <input
              v-model="passwordForm.new_password"
              type="password"
              password
              placeholder="请输入新密码"
              class="form-input"
            />
          </view>

          <view class="form-item">
            <view class="form-label">确认密码</view>
            <input
              v-model="passwordForm.confirm_password"
              type="password"
              password
              placeholder="请再次输入新密码"
              class="form-input"
            />
          </view>
        </view>

        <view class="modal-footer">
          <button class="btn-cancel" @click="closePasswordModal">取消</button>
          <button class="btn-confirm" :disabled="passwordLoading" @click="handleConfirmPassword">
            {{ passwordLoading ? '提交中...' : '确认' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useSpStore } from '../../stores/sp'
import { getSpSettings, changeSpPassword, spLogout } from '@api'
import type { SpSettings } from '@types'

const spStore = useSpStore()

const loading = ref(false)
const logoutLoading = ref(false)
const passwordLoading = ref(false)
const showPasswordModal = ref(false)

const spInfo = ref<SpSettings>({
  name: '',
  admin_name: '',
  contact_phone: '',
  contact_email: '',
  created_at: ''
})

const cacheSize = ref('0 KB')
const version = ref('v1.0.0')

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

onMounted(() => {
  loadSpSettings()
  calculateCacheSize()
})

async function loadSpSettings() {
  loading.value = true
  try {
    const res = await getSpSettings()
    spInfo.value = res
  } catch (error: any) {
    console.error('加载服务商设置失败:', error)
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function calculateCacheSize() {
  try {
    const info = uni.getSystemInfoSync()
    const sdcardPath = info.SD_CARD_PATH || ''
    const cacheDir = info.cacheDirectory || ''

    let totalSize = 0
    try {
      const storedData = uni.getStorageInfoSync()
      const dataSize = JSON.stringify(storedData).length

      if (dataSize > 1024 * 1024) {
        cacheSize.value = `${(dataSize / (1024 * 1024)).toFixed(2)} MB`
      } else if (dataSize > 1024) {
        cacheSize.value = `${(dataSize / 1024).toFixed(2)} KB`
      } else {
        cacheSize.value = `${dataSize} B`
      }
    } catch (e) {
      cacheSize.value = '未知'
    }
  } catch (error) {
    cacheSize.value = '未知'
  }
}

function handleChangePassword() {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  showPasswordModal.value = true
}

function closePasswordModal() {
  showPasswordModal.value = false
}

async function handleConfirmPassword() {
  if (!passwordForm.old_password) {
    uni.showToast({ title: '请输入旧密码', icon: 'none' })
    return
  }

  if (!passwordForm.new_password) {
    uni.showToast({ title: '请输入新密码', icon: 'none' })
    return
  }

  if (passwordForm.new_password.length < 6) {
    uni.showToast({ title: '新密码至少6位', icon: 'none' })
    return
  }

  if (passwordForm.new_password !== passwordForm.confirm_password) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }

  passwordLoading.value = true
  try {
    await changeSpPassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })

    uni.showToast({ title: '密码修改成功', icon: 'success' })
    closePasswordModal()
  } catch (error: any) {
    console.error('修改密码失败:', error)
    uni.showToast({ title: error.message || '修改失败', icon: 'none' })
  } finally {
    passwordLoading.value = false
  }
}

function handleClearCache() {
  uni.showModal({
    title: '提示',
    content: '确定要清除缓存吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          uni.clearStorageSync()
          cacheSize.value = '0 B'
          uni.showToast({ title: '清除成功', icon: 'success' })

          setTimeout(() => {
            uni.reLaunch({ url: '/pages/sp/login' })
          }, 1000)
        } catch (error) {
          uni.showToast({ title: '清除失败', icon: 'none' })
        }
      }
    }
  })
}

async function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: async (res) => {
      if (res.confirm) {
        logoutLoading.value = true
        try {
          await spLogout()
        } catch (error) {
          console.error('退出登录失败:', error)
        } finally {
          spStore.logout()
          logoutLoading.value = false
        }
      }
    }
  })
}
</script>

<style scoped>
.settings-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.page-header {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  padding: 40rpx 32rpx;
  padding-top: 80rpx;
}

.header-title {
  font-size: 40rpx;
  font-weight: 600;
  color: #ffffff;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: #999999;
}

.section {
  background: #ffffff;
  border-radius: 16rpx;
  margin: 24rpx;
  overflow: hidden;
}

.section-title {
  font-size: 26rpx;
  color: #999999;
  padding: 24rpx 32rpx 16rpx;
  background: #fafafa;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28rpx 32rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 30rpx;
  color: #1a1a1a;
}

.info-value {
  font-size: 28rpx;
  color: #666666;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 32rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.setting-icon {
  width: 64rpx;
  height: 64rpx;
  background: #f0f5ff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20rpx;
  font-size: 32rpx;
}

.setting-info {
  flex: 1;
}

.setting-label {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.setting-value {
  font-size: 26rpx;
  color: #999999;
}

.arrow {
  font-size: 32rpx;
  color: #cccccc;
}

.logout-area {
  margin-top: 48rpx;
  padding: 0 32rpx;
}

.btn-logout {
  width: 100%;
  height: 96rpx;
  background: #ffffff;
  border-radius: 48rpx;
  font-size: 32rpx;
  color: #ff4d4f;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-logout[disabled] {
  background: #f5f5f5;
  color: #cccccc;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.modal-content {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.modal-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.modal-close {
  font-size: 48rpx;
  color: #999999;
  line-height: 1;
}

.modal-body {
  padding: 32rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 16rpx;
}

.form-input {
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #1a1a1a;
}

.form-input::placeholder {
  color: #cccccc;
}

.modal-footer {
  display: flex;
  padding: 32rpx;
  gap: 24rpx;
}

.btn-cancel {
  flex: 1;
  height: 88rpx;
  background: #f5f5f5;
  border-radius: 44rpx;
  font-size: 30rpx;
  color: #666666;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-confirm {
  flex: 1;
  height: 88rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 44rpx;
  font-size: 30rpx;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-confirm[disabled] {
  background: #cccccc;
}
</style>
