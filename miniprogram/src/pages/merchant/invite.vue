<template>
  <view class="invite-container">
    <!-- 邀请码卡片 -->
    <view class="invite-card">
      <view class="card-header">
        <image
          class="merchant-logo"
          :src="merchantInfo?.logo || '/static/default-logo.png'"
          mode="aspectFill"
        />
        <view class="merchant-info">
          <view class="merchant-name">{{ merchantInfo?.name }}</view>
          <view class="invite-title">专属邀请码</view>
        </view>
      </view>

      <view class="invite-code-box">
        <text class="invite-code">{{ inviteInfo?.invite_code || '生成中...' }}</text>
      </view>

      <view class="invite-actions">
        <button class="btn-copy" @click="copyCode">复制邀请码</button>
        <button class="btn-share" @click="shareInvite">分享邀请</button>
      </view>

      <view class="qrcode-area" v-if="inviteInfo?.qrcode_url">
        <image
          class="qrcode-image"
          :src="inviteInfo.qrcode_url"
          mode="aspectFit"
        />
        <text class="qrcode-hint">扫码入驻</text>
      </view>
    </view>

    <!-- 邀请统计 -->
    <view class="stats-section">
      <view class="stats-title">邀请统计</view>
      <view class="stats-grid">
        <view class="stat-item">
          <view class="stat-value">{{ inviteInfoData?.total_invites || 0 }}</view>
          <view class="stat-label">累计邀请</view>
        </view>
        <view class="stat-item">
          <view class="stat-value">{{ inviteInfoData?.completed_invites || 0 }}</view>
          <view class="stat-label">已完成</view>
        </view>
        <view class="stat-item">
          <view class="stat-value">{{ inviteInfoData?.pending_invites || 0 }}</view>
          <view class="stat-label">待完成</view>
        </view>
      </view>

      <view class="rewards-info" v-if="inviteInfoData?.rewards">
        <view class="rewards-title">邀请奖励</view>
        <view class="reward-item">
          <text class="reward-icon">🎁</text>
          <text class="reward-text">免年费次数: {{ inviteInfoData.rewards.free_year_count }}</text>
        </view>
        <view class="reward-item">
          <text class="reward-icon">💰</text>
          <text class="reward-text">最低费率资格: {{ inviteInfoData.rewards.lowest_rate_qualified ? '已获得' : '未获得' }}</text>
        </view>
      </view>
    </view>

    <!-- 邀请记录 -->
    <view class="records-section">
      <view class="records-header">
        <view class="records-title">邀请记录</view>
        <text class="view-more" @click="viewAllRecords">查看全部 ›</text>
      </view>

      <view class="records-list">
        <view
          v-for="record in records"
          :key="record.id"
          class="record-item"
        >
          <view class="record-info">
            <view class="record-name">{{ record.invitee_name }}</view>
            <view class="record-phone">{{ record.invitee_phone }}</view>
          </view>
          <view class="record-status" :class="getRecordStatusClass(record.status)">
            {{ getRecordStatusText(record.status) }}
          </view>
        </view>

        <view v-if="records.length === 0" class="empty-records">
          <text>暂无邀请记录</text>
        </view>
      </view>
    </view>

    <!-- 邀请规则 -->
    <view class="rules-section">
      <view class="rules-title">邀请规则</view>
      <view class="rules-content">
        <view class="rule-item">1. 每成功邀请一位商家入驻，您可获得相应奖励</view>
        <view class="rule-item">2. 被邀请商家完成首次交易后，您可获得免年费资格</view>
        <view class="rule-item">3. 被邀请商家入驻后，您可申请最低0.2%交易费率</view>
        <view class="rule-item">4. 累计邀请多位商家可叠加享受更多权益</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onShow } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { generateInviteCode, getMyInviteInfo, getInviteRecords } from '../../api'
import type { InviteInfo, MyInviteInfo, InviteRecord } from '../../types/api'

const authStore = useAuthStore()
const merchantInfo = authStore.merchantInfo

const inviteInfo = ref<InviteInfo | null>(null)
const inviteInfoData = ref<MyInviteInfo | null>(null)
const records = ref<InviteRecord[]>([])

onShow(() => {
  loadData()
})

async function loadData() {
  await Promise.all([
    loadInviteCode(),
    loadInviteInfo(),
    loadRecords()
  ])
}

async function loadInviteCode() {
  try {
    const res = await generateInviteCode()
    inviteInfo.value = res
  } catch (error) {
    console.error('生成邀请码失败:', error)
  }
}

async function loadInviteInfo() {
  try {
    const res = await getMyInviteInfo()
    inviteInfoData.value = res
  } catch (error) {
    console.error('加载邀请信息失败:', error)
  }
}

async function loadRecords() {
  try {
    const res = await getInviteRecords({ page: 1, page_size: 5 })
    records.value = res.list
  } catch (error) {
    console.error('加载邀请记录失败:', error)
  }
}

function copyCode() {
  if (!inviteInfo.value?.invite_code) return
  
  uni.setClipboardData({
    data: inviteInfo.value.invite_code,
    success: () => uni.showToast({ title: '已复制', icon: 'success' })
  })
}

function shareInvite() {
  uni.showShareMenu({
    withShareTicket: true,
    menus: ['shareAppMessage', 'shareTimeline']
  })
}

function viewAllRecords() {
  uni.navigateTo({ url: '/pages/merchant/invite/records' })
}

function getRecordStatusClass(status: number): string {
  const classMap: Record<number, string> = {
    0: 'pending',
    1: 'completed',
    2: 'cancelled'
  }
  return classMap[status] || ''
}

function getRecordStatusText(status: number): string {
  const textMap: Record<number, string> = {
    0: '待完成',
    1: '已完成',
    2: '已取消'
  }
  return textMap[status] || '未知'
}
</script>

<style scoped>
.invite-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
  padding-bottom: 48rpx;
}

.invite-card {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 40rpx;
  color: #ffffff;
}

.card-header {
  display: flex;
  align-items: center;
  margin-bottom: 32rpx;
}

.merchant-logo {
  width: 80rpx;
  height: 80rpx;
  border-radius: 16rpx;
  background: #ffffff;
  margin-right: 20rpx;
}

.merchant-info {
  flex: 1;
}

.merchant-name {
  font-size: 30rpx;
  font-weight: 500;
  margin-bottom: 4rpx;
}

.invite-title {
  font-size: 26rpx;
  opacity: 0.8;
}

.invite-code-box {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 16rpx;
  padding: 32rpx;
  text-align: center;
  margin-bottom: 32rpx;
}

.invite-code {
  font-size: 48rpx;
  font-weight: 700;
  letter-spacing: 8rpx;
}

.invite-actions {
  display: flex;
  gap: 24rpx;
  margin-bottom: 32rpx;
}

.btn-copy, .btn-share {
  flex: 1;
  height: 80rpx;
  background: #ffffff;
  border-radius: 40rpx;
  font-size: 28rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-copy {
  color: #007AFF;
}

.btn-share {
  color: #007AFF;
}

.qrcode-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 24rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.2);
}

.qrcode-image {
  width: 240rpx;
  height: 240rpx;
  background: #ffffff;
  border-radius: 16rpx;
  margin-bottom: 12rpx;
}

.qrcode-hint {
  font-size: 24rpx;
  opacity: 0.8;
}

.stats-section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-top: 24rpx;
}

.stats-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.stats-grid {
  display: flex;
  justify-content: space-around;
  margin-bottom: 24rpx;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 40rpx;
  font-weight: 600;
  color: #007AFF;
  margin-bottom: 8rpx;
}

.stat-label {
  font-size: 26rpx;
  color: #999999;
}

.rewards-info {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
}

.rewards-title {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 16rpx;
}

.reward-item {
  display: flex;
  align-items: center;
  padding: 12rpx 0;
  font-size: 26rpx;
  color: #666666;
}

.reward-icon {
  margin-right: 12rpx;
}

.records-section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-top: 24rpx;
}

.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.records-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.view-more {
  font-size: 26rpx;
  color: #007AFF;
}

.record-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.record-item:last-child {
  border-bottom: none;
}

.record-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.record-phone {
  font-size: 24rpx;
  color: #999999;
}

.record-status {
  font-size: 26rpx;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.record-status.pending {
  background: #fff7e6;
  color: #fa8c16;
}

.record-status.completed {
  background: #f6ffed;
  color: #52c41a;
}

.record-status.cancelled {
  background: #f5f5f5;
  color: #999999;
}

.empty-records {
  text-align: center;
  padding: 48rpx 0;
  font-size: 28rpx;
  color: #999999;
}

.rules-section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-top: 24rpx;
}

.rules-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.rule-item {
  font-size: 28rpx;
  color: #666666;
  line-height: 2;
}
</style>
