<template>
  <view class="settings-container">
    <view class="section">
      <view class="section-title">商家信息</view>
      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon"><text>🏪</text></view>
          <view class="setting-info">
            <view class="setting-label">店铺名称</view>
            <view class="setting-value">{{ merchantInfo?.name || '未设置' }}</view>
          </view>
        </view>
      </view>
      <view class="setting-item">
        <view class="setting-left">
          <view class="setting-icon"><text>📞</text></view>
          <view class="setting-info">
            <view class="setting-label">联系电话</view>
            <view class="setting-value">{{ merchantInfo?.contact_phone || '未绑定' }}</view>
          </view>
        </view>
      </view>
      <view class="setting-item" @click="goDeliverySettings">
        <view class="setting-left">
          <view class="setting-icon"><text>🚚</text></view>
          <view class="setting-info">
            <view class="setting-label">配送设置</view>
            <view class="setting-value">{{ orderModeSummary }}</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view class="setting-item" @click="goQrcode">
        <view class="setting-left">
          <view class="setting-icon"><text>📱</text></view>
          <view class="setting-info">
            <view class="setting-label">店铺二维码</view>
            <view class="setting-value">前往工作台查看</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">账号安全</view>
      <view class="setting-item" @click="openPasswordDialog">
        <view class="setting-left">
          <view class="setting-icon"><text>🔑</text></view>
          <view class="setting-info">
            <view class="setting-label">修改密码</view>
            <view class="setting-value">修改成功后需重新登录</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view class="setting-item" @click="handleWechatAction">
        <view class="setting-left">
          <view class="setting-icon"><text>💬</text></view>
          <view class="setting-info">
            <view class="setting-label">微信快捷登录</view>
            <view class="setting-value">
              {{ settings?.wechat_bound ? `已绑定${formattedWechatBoundAt ? ` · ${formattedWechatBoundAt}` : ''}` : '未绑定' }}
            </view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">店铺运营</view>
      <view class="setting-item" @click="goNotificationSettings">
        <view class="setting-left">
          <view class="setting-icon"><text>🔔</text></view>
          <view class="setting-info">
            <view class="setting-label">声音提醒管理</view>
            <view class="setting-value">{{ notificationSummary }}</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view class="setting-item" @click="goAnalytics">
        <view class="setting-left">
          <view class="setting-icon"><text>📊</text></view>
          <view class="setting-info">
            <view class="setting-label">数据看板</view>
            <view class="setting-value">查看经营数据与访客趋势</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
      <view v-if="canManageStaff" class="setting-item" @click="goStaffManagement">
        <view class="setting-left">
          <view class="setting-icon"><text>👥</text></view>
          <view class="setting-info">
            <view class="setting-label">员工管理</view>
            <view class="setting-value">仅店主可管理账号与提醒</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="setting-item" @click="goProfitSharingHistory">
        <view class="setting-left">
          <view class="setting-icon"><text>💰</text></view>
          <view class="setting-info">
            <view class="setting-label">分账历史</view>
            <view class="setting-value">查看抽佣日期、金额、比例与状态</view>
          </view>
        </view>
        <text class="arrow">›</text>
      </view>
    </view>

    <view class="logout-area">
      <button class="btn-logout" @click="handleLogout">退出登录</button>
    </view>

    <view v-if="passwordDialogVisible" class="dialog-mask" @click="closePasswordDialog">
      <view class="dialog-card" @click.stop>
        <view class="dialog-title">修改密码</view>
        <view class="dialog-form">
          <input v-model="passwordForm.old_password" class="dialog-input" password placeholder="请输入原密码" />
          <input v-model="passwordForm.new_password" class="dialog-input" password placeholder="请输入新密码（至少6位）" />
          <input v-model="passwordForm.confirm_password" class="dialog-input" password placeholder="请再次输入新密码" />
        </view>
        <view class="dialog-actions">
          <button class="dialog-btn secondary" @click="closePasswordDialog">取消</button>
          <button class="dialog-btn primary" :disabled="passwordSaving" @click="submitPasswordChange">
            {{ passwordSaving ? '提交中...' : '确认修改' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '../../stores/auth'
import { bindMerchantWechat, changeMerchantPassword, getMerchantSettings, unbindMerchantWechat } from '@api'
import type { MerchantSettings } from '@types'
import { getMerchantWechatCode } from '../../utils/merchant_wechat'

const authStore = useAuthStore()

const merchantInfo = computed(() => authStore.merchantInfo)
const settings = ref<MerchantSettings | null>(null)
const passwordDialogVisible = ref(false)
const passwordSaving = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const orderModeSummary = computed(() => {
  const labels: string[] = []
  if (settings.value?.takeout_enabled) {
    labels.push('配送')
  }
  if (settings.value?.dine_in_enabled) {
    labels.push('堂食')
  }
  if (settings.value?.pickup_enabled) {
    labels.push('自提')
  }
  return labels.length > 0 ? labels.join(' / ') : '暂未开放下单方式'
})
const canManageStaff = computed(() => authStore.staff?.role === 'owner')
const notificationSummary = computed(() => {
  const labels: string[] = []
  if (authStore.orderSoundEnabled) {
    labels.push('下单提醒开')
  }
  if (authStore.browseSoundEnabled) {
    labels.push('浏览提醒开')
  }
  return labels.length > 0 ? labels.join(' / ') : '全部关闭'
})
const formattedWechatBoundAt = computed(() => {
  if (!settings.value?.wechat_bound_at) {
    return ''
  }
  return settings.value.wechat_bound_at.replace('T', ' ').slice(0, 16)
})

onShow(() => {
  loadSettings()
})

async function loadSettings() {
  try {
    const result = await getMerchantSettings()
    settings.value = result
    authStore.setOrderSoundEnabled(result.notify_enabled ?? true)
    authStore.setBrowseSoundEnabled(result.browse_notify_enabled ?? true)
  } catch (error) {
    console.error('加载商家设置失败:', error)
  }
}

function goDeliverySettings() {
  uni.navigateTo({ url: '/pages/merchant/delivery-settings' })
}

function goNotificationSettings() {
  uni.navigateTo({ url: '/pages/merchant/notification-settings' })
}

function goProfitSharingHistory() {
  uni.navigateTo({ url: '/pages/merchant/settlements/history' })
}

function goAnalytics() {
  uni.switchTab({ url: '/pages/merchant/analytics/index' })
}

function goQrcode() {
  uni.switchTab({ url: '/pages/merchant/home' })
}

function goStaffManagement() {
  uni.navigateTo({ url: '/pages/merchant/staff' })
}

function openPasswordDialog() {
  passwordDialogVisible.value = true
}

function closePasswordDialog() {
  passwordDialogVisible.value = false
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}

async function submitPasswordChange() {
  if (!passwordForm.old_password) {
    uni.showToast({ title: '请输入原密码', icon: 'none' })
    return
  }
  if (!passwordForm.new_password || passwordForm.new_password.length < 6) {
    uni.showToast({ title: '新密码至少 6 位', icon: 'none' })
    return
  }
  if (passwordForm.new_password === passwordForm.old_password) {
    uni.showToast({ title: '新旧密码不能相同', icon: 'none' })
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    uni.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
    return
  }

  passwordSaving.value = true
  try {
    await changeMerchantPassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    uni.showToast({ title: '密码修改成功，请重新登录', icon: 'success' })
    closePasswordDialog()
    setTimeout(() => {
      authStore.logout()
    }, 800)
  } catch (error: any) {
    uni.showToast({ title: error?.message || '修改失败', icon: 'none' })
  } finally {
    passwordSaving.value = false
  }
}

async function handleWechatAction() {
  if (settings.value?.wechat_bound) {
    uni.showModal({
      title: '解绑微信',
      content: '解绑后将无法使用微信快捷登录，是否继续？',
      success: async (res) => {
        if (!res.confirm) {
          return
        }
        try {
          await unbindMerchantWechat()
          if (authStore.staff) {
            authStore.updateStaffInfo({
              ...authStore.staff,
              openid: '',
              unionid: '',
              wechat_bound_at: ''
            })
          }
          await loadSettings()
          uni.showToast({ title: '解绑成功', icon: 'success' })
        } catch (error: any) {
          uni.showToast({ title: error?.message || '解绑失败', icon: 'none' })
        }
      }
    })
    return
  }

  try {
    const code = await getMerchantWechatCode()
    const result = await bindMerchantWechat({ code })
    settings.value = {
      ...(settings.value || {} as MerchantSettings),
      wechat_bound: true,
      unionid: result.unionid,
      wechat_bound_at: result.wechat_bound_at
    }
    if (authStore.staff) {
      authStore.updateStaffInfo({
        ...authStore.staff,
        openid: result.openid,
        unionid: result.unionid,
        wechat_bound_at: result.wechat_bound_at
      })
    }
    await loadSettings()
    uni.showToast({ title: '绑定成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error?.message || '绑定失败', icon: 'none' })
  }
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStore.logout()
      }
    }
  })
}
</script>

<style scoped>
.settings-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  padding-bottom: 200rpx;
}

.section {
  background: #ffffff;
  border-radius: 16rpx;
  margin-bottom: 24rpx;
  overflow: hidden;
}

.section-title {
  font-size: 26rpx;
  color: #999999;
  padding: 24rpx 32rpx 16rpx;
  background: #fafafa;
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
}

.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32rpx;
}

.dialog-card {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 40rpx 32rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.dialog-input {
  height: 88rpx;
  border-radius: 16rpx;
  background: #f8f9fa;
  padding: 0 24rpx;
  font-size: 30rpx;
}

.dialog-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 32rpx;
}

.dialog-btn {
  flex: 1;
  height: 84rpx;
  border-radius: 42rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-btn.secondary {
  background: #f5f5f5;
  color: #666666;
}

.dialog-btn.primary {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}
</style>
