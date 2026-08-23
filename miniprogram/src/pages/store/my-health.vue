<template>
  <view class="my-health-container">
    <!-- 顶部健康档案卡片 -->
    <view class="profile-card" :class="{ 'no-record': !record }">
      <template v-if="record">
        <view class="profile-header">
          <view class="profile-avatar">{{ profileAvatarText }}</view>
          <view class="profile-info">
            <view class="profile-name">
              {{ record.real_name || '未填写姓名' }}
              <text class="profile-tag" v-if="genderText">{{ genderText }}</text>
              <text class="profile-tag" v-if="ageText">{{ ageText }}岁</text>
            </view>
            <view class="profile-level" v-if="record.assessment_level">
              健康等级：{{ record.assessment_level }}
            </view>
            <view class="profile-level" v-else>尚未进行健康评估</view>
          </view>
          <view class="profile-edit" @click="goEditRecord">编辑 ›</view>
        </view>
        <view class="chronic-tags" v-if="chronicTags.length">
          <text v-for="tag in chronicTags" :key="tag" class="chronic-tag">{{ tag }}</text>
        </view>
        <view class="profile-meta" v-if="record.birth_date || record.blood_type">
          <text v-if="record.birth_date">出生 {{ record.birth_date }}</text>
          <text v-if="record.blood_type">血型 {{ record.blood_type }}型</text>
        </view>
      </template>
      <template v-else>
        <view class="no-record-icon">🩺</view>
        <view class="no-record-title">还未建立健康档案</view>
        <view class="no-record-desc">完善健康档案，获取更精准的健康服务</view>
        <view class="go-create-btn" @click="goEditRecord">去完善档案</view>
      </template>
    </view>

    <!-- 功能入口网格 -->
    <view class="entry-grid">
      <view class="entry-item" @click="goEditRecord">
        <view class="entry-icon record">📋</view>
        <view class="entry-title">健康档案</view>
        <view class="entry-desc">编辑个人健康信息</view>
      </view>
      <view class="entry-item" @click="goAssessment">
        <view class="entry-icon assess">📝</view>
        <view class="entry-title">自助评估</view>
        <view class="entry-desc">在线健康自测量表</view>
      </view>
      <view class="entry-item" @click="goRecords">
        <view class="entry-icon history">📊</view>
        <view class="entry-title">评估记录</view>
        <view class="entry-desc">查看历史评估结果</view>
      </view>
      <view class="entry-item" @click="goFitting">
        <view class="entry-icon fitting">🛒</view>
        <view class="entry-title">适配建议</view>
        <view class="entry-desc">查看辅具适配推荐</view>
      </view>
      <view class="entry-item" @click="goCarePlans">
        <view class="entry-icon care">🩺</view>
        <view class="entry-title">照护计划</view>
        <view class="entry-desc">查看照护方案与记录</view>
      </view>
      <view class="entry-item" @click="goFollowUps">
        <view class="entry-icon follow">🗓️</view>
        <view class="entry-title">康复随访</view>
        <view class="entry-desc">查看随访任务与结果</view>
      </view>
      <view class="entry-item" @click="goEducation">
        <view class="entry-icon edu">📖</view>
        <view class="entry-title">健康宣教</view>
        <view class="entry-desc">康复与慢病科普文章</view>
      </view>
      <view class="entry-item" @click="goMonitoring">
        <view class="entry-icon monitor">❤️</view>
        <view class="entry-title">体征记录</view>
        <view class="entry-desc">记录每日生命体征</view>
      </view>
    </view>

    <!-- 最近评估记录 -->
    <view class="recent-card">
      <view class="section-header">
        <text class="section-title">最近评估</text>
        <text class="section-more" @click="goRecords">全部 ›</text>
      </view>
      <view v-if="recentAssessments.length" class="assessment-list">
        <view
          v-for="item in recentAssessments"
          :key="item.id"
          class="assessment-item"
          @click="goRecords"
        >
          <view class="assessment-main">
            <view class="assessment-name">{{ item.form_name || '健康评估' }}</view>
            <view class="assessment-meta">{{ formatTime(item.created_at) }}</view>
          </view>
          <view class="assessment-right">
            <view class="assessment-score">{{ item.total_score ?? '--' }}</view>
            <view class="assessment-level" :class="getLevelClass(item.level)">
              {{ item.level || '未分级' }}
            </view>
          </view>
        </view>
      </view>
      <view v-else class="recent-empty">
        <view class="recent-empty-text">暂无评估记录</view>
        <view class="recent-empty-desc">去做一次健康评估，了解当前健康状态</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { getUserHealthRecord, getUserAssessments } from '../../api/health'
import type { HealthAssessment, HealthRecord } from '../../types'
import { useAuth } from '../../utils/useAuth'

const record = ref<HealthRecord | null>(null)
const recentAssessments = ref<HealthAssessment[]>([])
const loading = ref(false)

const hasRecord = computed(() => !!record.value)
const chronicTags = computed(() => record.value?.chronic_tags || [])
const genderText = computed(() => {
  const gender = record.value?.gender
  if (gender === 1) return '男'
  if (gender === 2) return '女'
  return ''
})
const ageText = computed(() => {
  const age = calcAge(record.value?.birth_date)
  return age === null ? '' : String(age)
})
const profileAvatarText = computed(() => {
  const name = record.value?.real_name?.trim()
  return name ? name.slice(0, 1) : '我'
})

function calcAge(birthDate?: string): number | null {
  if (!birthDate) return null
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return null
  const today = new Date()
  let age = today.getFullYear() - birth.getFullYear()
  const monthDiff = today.getMonth() - birth.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--
  }
  return age
}

function formatTime(time?: string): string {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function getLevelClass(level?: string): string {
  if (!level) return ''
  if (level.includes('高')) return 'high'
  if (level.includes('中')) return 'medium'
  return 'low'
}

async function loadData() {
  if (loading.value) return
  loading.value = true
  try {
    const [recordData, assessments] = await Promise.all([
      getUserHealthRecord(),
      getUserAssessments({ page: 1, page_size: 3 })
    ])
    record.value = recordData
    recentAssessments.value = assessments?.list || []
  } catch (error) {
    console.error('加载我的健康数据失败:', error)
  } finally {
    loading.value = false
  }
}

onShow(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  loadData()
})

onPullDownRefresh(async () => {
  await loadData()
  uni.stopPullDownRefresh()
})

function goEditRecord() {
  uni.navigateTo({ url: '/pages/store/health-record-edit' })
}

function goAssessment() {
  uni.navigateTo({ url: '/pages/store/health-assessment?mode=assess' })
}

function goRecords() {
  uni.navigateTo({ url: '/pages/store/health-assessment?mode=records' })
}

function goFitting() {
  uni.navigateTo({ url: '/pages/store/my-fitting' })
}

function goCarePlans() {
  uni.navigateTo({ url: '/pages/store/my-care-plans' })
}

function goFollowUps() {
  uni.navigateTo({ url: '/pages/store/my-follow-ups' })
}

function goEducation() {
  uni.navigateTo({ url: '/pages/store/health-education' })
}

function goMonitoring() {
  uni.navigateTo({ url: '/pages/store/my-monitoring' })
}
</script>

<style scoped>
.my-health-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

/* 健康档案卡片 */
.profile-card {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 36rpx 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.profile-card.no-record {
  text-align: center;
  padding: 48rpx 32rpx;
}

.profile-header {
  display: flex;
  align-items: center;
}

.profile-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 44rpx;
  font-weight: 600;
  margin-right: 24rpx;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-size: 36rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12rpx;
}

.profile-tag {
  font-size: 22rpx;
  font-weight: 400;
  padding: 2rpx 14rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
}

.profile-level {
  margin-top: 12rpx;
  font-size: 24rpx;
  opacity: 0.9;
}

.profile-edit {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  flex-shrink: 0;
}

.chronic-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 24rpx;
}

.chronic-tag {
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.2);
}

.profile-meta {
  margin-top: 20rpx;
  font-size: 22rpx;
  opacity: 0.85;
  display: flex;
  gap: 24rpx;
}

.no-record-icon {
  font-size: 88rpx;
  margin-bottom: 20rpx;
}

.no-record-title {
  font-size: 34rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.no-record-desc {
  font-size: 24rpx;
  opacity: 0.85;
  margin-bottom: 32rpx;
}

.go-create-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 240rpx;
  height: 76rpx;
  padding: 0 40rpx;
  border-radius: 999rpx;
  background: #ffffff;
  color: #007AFF;
  font-size: 28rpx;
  font-weight: 600;
}

/* 功能入口网格 */
.entry-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
  margin-bottom: 24rpx;
}

.entry-item {
  flex: 1;
  min-width: 190rpx;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx 16rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.entry-icon {
  width: 80rpx;
  height: 80rpx;
  border-radius: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 38rpx;
  margin-bottom: 16rpx;
}

.entry-icon.record {
  background: rgba(0, 122, 255, 0.1);
}

.entry-icon.assess {
  background: rgba(22, 163, 74, 0.1);
}

.entry-icon.history {
  background: rgba(255, 149, 0, 0.1);
}

.entry-icon.fitting {
  background: rgba(99, 102, 241, 0.1);
}

.entry-icon.care {
  background: rgba(0, 191, 166, 0.1);
}

.entry-icon.follow {
  background: rgba(255, 59, 48, 0.1);
}

.entry-icon.edu {
  background: rgba(94, 92, 230, 0.1);
}

.entry-icon.monitor {
  background: rgba(255, 45, 85, 0.1);
}

.entry-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.entry-desc {
  font-size: 20rpx;
  color: #999999;
}

/* 最近评估 */
.recent-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.section-more {
  font-size: 24rpx;
  color: #007AFF;
}

.assessment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.assessment-item:last-child {
  border-bottom: none;
}

.assessment-main {
  flex: 1;
  min-width: 0;
}

.assessment-name {
  font-size: 30rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.assessment-meta {
  font-size: 22rpx;
  color: #999999;
}

.assessment-right {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-left: 20rpx;
}

.assessment-score {
  font-size: 34rpx;
  font-weight: 700;
  color: #1a1a1a;
}

.assessment-level {
  font-size: 20rpx;
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
}

.assessment-level.high {
  background: #fff1f0;
  color: #ff4d4f;
}

.assessment-level.medium {
  background: #fff7e6;
  color: #fa8c16;
}

.assessment-level.low {
  background: #f6ffed;
  color: #52c41a;
}

.recent-empty {
  text-align: center;
  padding: 40rpx 0;
}

.recent-empty-text {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 10rpx;
}

.recent-empty-desc {
  font-size: 24rpx;
  color: #999999;
}
</style>
