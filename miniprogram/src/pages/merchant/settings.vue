<template>
  <view class="settings-container">
    <view class="hero-card">
      <view class="hero-header">
        <view class="hero-icon">🏪</view>
        <view class="hero-main">
          <view class="hero-title">{{ merchantInfo?.name || '未设置店铺名称' }}</view>
          <view class="hero-subtitle">{{ merchantInfo?.contact_phone || '未绑定联系电话' }}</view>
        </view>
      </view>
      <view class="hero-tags">
        <text class="hero-tag primary">{{ orderModeSummary }}</text>
        <text class="hero-tag">{{ wechatStatusText }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-title">经营工具</view>
      <view class="feature-grid">
        <view
          v-for="item in featureCards"
          :key="item.key"
          class="feature-card"
          @click="handleFeatureClick(item.key)"
        >
          <view class="feature-icon">{{ item.icon }}</view>
          <view class="feature-title">{{ item.title }}</view>
          <view class="feature-desc">{{ item.desc }}</view>
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section-title">账号与安全</view>
      <view class="security-grid">
        <view class="security-card" @click="openPasswordDialog">
          <view class="security-icon">🔑</view>
          <view class="security-main">
            <view class="security-title">修改密码</view>
            <view class="security-desc">修改成功后需重新登录</view>
          </view>
          <text class="security-arrow">›</text>
        </view>
        <view class="security-card" @click="handleWechatAction">
          <view class="security-icon">💬</view>
          <view class="security-main">
            <view class="security-title">微信快捷登录</view>
            <view class="security-desc">{{ wechatStatusText }}</view>
          </view>
          <text class="security-arrow">›</text>
        </view>
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

type FeatureKey =
  | 'delivery'
  | 'marketing'
  | 'printers'
  | 'notification'
  | 'staff'
  | 'profit-sharing'

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
const wechatStatusText = computed(() => {
  if (!settings.value?.wechat_bound) {
    return '未绑定微信快捷登录'
  }
  return formattedWechatBoundAt.value
    ? `已绑定 · ${formattedWechatBoundAt.value}`
    : '已绑定微信快捷登录'
})
const featureCards = computed(() => {
  const cards: Array<{ key: FeatureKey; icon: string; title: string; desc: string }> = [
    {
      key: 'delivery',
      icon: '🚚',
      title: '配送设置',
      desc: orderModeSummary.value
    },
    {
      key: 'marketing',
      icon: '🎯',
      title: '满减营销',
      desc: '配置满多少减多少'
    },
    {
      key: 'printers',
      icon: '🖨️',
      title: '打印机管理',
      desc: '飞鹅参数与打印开关'
    },
    {
      key: 'notification',
      icon: '🔔',
      title: '声音提醒',
      desc: notificationSummary.value
    },
    {
      key: 'profit-sharing',
      icon: '💰',
      title: '分账历史',
      desc: '查看抽佣日期与状态'
    }
  ]

  if (canManageStaff.value) {
    cards.splice(4, 0, {
      key: 'staff',
      icon: '👥',
      title: '员工管理',
      desc: '仅店主可管理账号'
    })
  }

  return cards
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

function goMarketing() {
  uni.navigateTo({ url: '/pages/merchant/marketing' })
}

function goPrinters() {
  uni.navigateTo({ url: '/pages/merchant/printers' })
}

function goNotificationSettings() {
  uni.navigateTo({ url: '/pages/merchant/notification-settings' })
}

function goProfitSharingHistory() {
  uni.navigateTo({ url: '/pages/merchant/settlements/history' })
}

function goStaffManagement() {
  uni.navigateTo({ url: '/pages/merchant/staff' })
}

function handleFeatureClick(key: FeatureKey) {
  switch (key) {
    case 'delivery':
      goDeliverySettings()
      return
    case 'marketing':
      goMarketing()
      return
    case 'printers':
      goPrinters()
      return
    case 'notification':
      goNotificationSettings()
      return
    case 'staff':
      goStaffManagement()
      return
    case 'profit-sharing':
      goProfitSharingHistory()
      return
  }
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
  background: linear-gradient(180deg, #eef5ff 0%, #f5f5f5 220rpx);
  padding: 24rpx 24rpx 200rpx;
}

.hero-card {
  background: linear-gradient(135deg, #0a60ff 0%, #3d8dff 100%);
  border-radius: 28rpx;
  padding: 32rpx;
  color: #ffffff;
  box-shadow: 0 20rpx 40rpx rgba(10, 96, 255, 0.18);
  margin-bottom: 24rpx;
}

.hero-header {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.hero-icon {
  width: 96rpx;
  height: 96rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.16);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 44rpx;
  flex-shrink: 0;
}

.hero-main {
  min-width: 0;
  flex: 1;
}

.hero-title {
  font-size: 36rpx;
  font-weight: 600;
  line-height: 1.4;
}

.hero-subtitle {
  margin-top: 10rpx;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.82);
  word-break: break-all;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: 28rpx;
}

.hero-tag {
  padding: 10rpx 20rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.14);
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.92);
}

.hero-tag.primary {
  background: rgba(255, 255, 255, 0.22);
}

.section {
  background: #ffffff;
  border-radius: 24rpx;
  margin-bottom: 24rpx;
  padding: 28rpx;
  box-shadow: 0 12rpx 32rpx rgba(17, 32, 68, 0.05);
}

.section-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20rpx;
}

.feature-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 18rpx;
}

.feature-card {
  width: calc((100% - 18rpx) / 2);
  min-height: 200rpx;
  border-radius: 22rpx;
  background: #f8fbff;
  border: 1rpx solid #e6efff;
  padding: 28rpx 24rpx;
  box-sizing: border-box;
}

.feature-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 18rpx;
  background: #e8f1ff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34rpx;
}

.feature-title {
  margin-top: 18rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.feature-desc {
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #7a7f89;
}

.security-grid {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.security-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f8f9fb;
}

.security-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 18rpx;
  background: #eef3ff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  flex-shrink: 0;
}

.security-main {
  min-width: 0;
  flex: 1;
}

.security-title {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 600;
}

.security-desc {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #7a7f89;
  line-height: 1.5;
}

.security-arrow {
  font-size: 34rpx;
  color: #c8ccd6;
}

.logout-area {
  margin-top: 16rpx;
}

.btn-logout {
  width: 100%;
  background: #ffffff;
  color: #ff4d4f;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 48rpx;
  font-size: 32rpx;
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
