<template>
  <view class="my-health-container">
    <!-- 成员选择区（最多 5 个） -->
    <view v-if="records.length" class="member-bar">
      <scroll-view class="member-scroll" scroll-x>
        <view class="member-list">
          <view
            v-for="item in records"
            :key="item.id"
            class="member-item"
            :class="{ active: item.id === activeId }"
            @click="switchMember(item.id)"
          >
            <view class="member-avatar">{{ memberAvatarText(item) }}</view>
            <view class="member-name">{{ item.real_name || '未填写' }}</view>
            <view class="member-relation">{{ relationText(item.relation) }}</view>
          </view>
          <view v-if="records.length < 5" class="member-item member-add" @click="goAddMember">
            <view class="member-add-icon">＋</view>
            <view class="member-name">添加成员</view>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 当前成员健康档案卡片 -->
    <view class="profile-card" :class="{ 'no-record': !activeRecord }">
      <template v-if="activeRecord">
        <view class="profile-header">
          <view class="profile-avatar">{{ activeAvatarText }}</view>
          <view class="profile-info">
            <view class="profile-name">
              {{ activeRecord.real_name || '未填写姓名' }}
              <text v-if="relationText(activeRecord.relation)" class="profile-tag">{{ relationText(activeRecord.relation) }}</text>
              <text v-if="genderText" class="profile-tag">{{ genderText }}</text>
              <text v-if="ageText" class="profile-tag">{{ ageText }}岁</text>
            </view>
            <view class="profile-level" v-if="activeRecord.assessment_level">
              健康等级：{{ activeRecord.assessment_level }}
            </view>
            <view class="profile-level" v-else>尚未进行健康评估</view>
          </view>
          <view class="profile-actions">
            <view class="profile-action" @click="goEditRecord">编辑</view>
            <view class="profile-action danger" @click="deleteRecord">删除</view>
          </view>
        </view>
        <view class="chronic-tags" v-if="chronicTags.length">
          <text v-for="tag in chronicTags" :key="tag" class="chronic-tag">{{ tag }}</text>
        </view>
        <view class="profile-meta" v-if="activeRecord.birth_date || activeRecord.blood_type">
          <text v-if="activeRecord.birth_date">出生 {{ activeRecord.birth_date }}</text>
          <text v-if="activeRecord.blood_type">血型 {{ activeRecord.blood_type }}型</text>
        </view>
      </template>
      <template v-else>
        <view class="no-record-icon">🩺</view>
        <view class="no-record-title">还没有健康档案</view>
        <view class="no-record-desc">支持本人 / 父母 / 其他亲属分别建档，最多 5 个</view>
        <view class="go-create-btn" @click="goAddMember">去添加成员</view>
      </template>
    </view>

    <!-- 功能入口网格 -->
    <view class="entry-grid">
      <view class="entry-item" @click="goEditRecord">
        <view class="entry-icon record">📋</view>
        <view class="entry-title">编辑档案</view>
        <view class="entry-desc">完善当前成员信息</view>
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
    </view>

    <!-- 最近评估记录 -->
    <view class="recent-card" v-if="activeRecord">
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
import { listUserHealthRecords, deleteUserHealthRecord, getUserAssessments } from '../../api/health'
import type { HealthAssessment, HealthRecord } from '../../types'
import { useAuth } from '../../utils/useAuth'

const records = ref<HealthRecord[]>([])
const activeId = ref<number | null>(null)
const recentAssessments = ref<HealthAssessment[]>([])
const loading = ref(false)

const activeRecord = computed(() => records.value.find(item => item.id === activeId.value) || null)
const chronicTags = computed(() => activeRecord.value?.chronic_tags || [])
const genderText = computed(() => {
  const gender = activeRecord.value?.gender
  if (gender === 1) return '男'
  if (gender === 2) return '女'
  return ''
})
const ageText = computed(() => {
  const age = calcAge(activeRecord.value?.birth_date)
  return age === null ? '' : String(age)
})
const activeAvatarText = computed(() => {
  const name = activeRecord.value?.real_name?.trim()
  return name ? name.slice(0, 1) : '我'
})

function memberAvatarText(record: HealthRecord): string {
  const name = record.real_name?.trim()
  return name ? name.slice(0, 1) : '?'
}

function relationText(relation?: number): string {
  if (relation === 1) return '本人'
  if (relation === 2) return '父母'
  if (relation === 3) return '其他亲属'
  return ''
}

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
    const list = await listUserHealthRecords()
    records.value = list
    if (!list.length) {
      activeId.value = null
      recentAssessments.value = []
    } else {
      const stillExists = activeId.value !== null && list.some(item => item.id === activeId.value)
      if (!stillExists) {
        activeId.value = list[0].id
      }
      await loadRecentAssessments()
    }
  } catch (error) {
    console.error('加载健康档案失败:', error)
  } finally {
    loading.value = false
  }
}

async function loadRecentAssessments() {
  if (!activeId.value) {
    recentAssessments.value = []
    return
  }
  try {
    const res = await getUserAssessments({ page: 1, page_size: 3, record_id: activeId.value })
    recentAssessments.value = res?.list || []
  } catch (error) {
    recentAssessments.value = []
  }
}

function switchMember(id: number) {
  if (activeId.value === id) return
  activeId.value = id
  loadRecentAssessments()
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

function goAddMember() {
  uni.navigateTo({ url: '/pages/store/health-record-edit' })
}

function requireActiveRecord(): boolean {
  if (activeId.value) return true
  uni.showToast({ title: '请先添加健康档案', icon: 'none' })
  goAddMember()
  return false
}

function goEditRecord() {
  if (activeId.value) {
    uni.navigateTo({ url: `/pages/store/health-record-edit?id=${activeId.value}` })
  } else {
    goAddMember()
  }
}

function goAssessment() {
  if (!requireActiveRecord()) return
  uni.navigateTo({ url: `/pages/store/health-assessment?mode=assess&record_id=${activeId.value}` })
}

function goRecords() {
  if (!requireActiveRecord()) return
  uni.navigateTo({ url: `/pages/store/health-assessment?mode=records&record_id=${activeId.value}` })
}

function goFitting() {
  uni.navigateTo({ url: '/pages/store/my-fitting' })
}

function deleteRecord() {
  const record = activeRecord.value
  if (!record) return
  uni.showModal({
    title: '删除档案',
    content: `确定删除「${record.real_name || '该成员'}」的健康档案吗？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteUserHealthRecord(record.id)
        uni.showToast({ title: '删除成功', icon: 'success' })
        await loadData()
      } catch (error: any) {
        uni.showToast({ title: error.message || '删除失败', icon: 'none' })
      }
    }
  })
}
</script>

<style scoped>
.my-health-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

/* 成员选择区 */
.member-bar {
  margin-bottom: 24rpx;
}

.member-scroll {
  width: 100%;
  white-space: nowrap;
}

.member-list {
  display: inline-flex;
  gap: 16rpx;
}

.member-item {
  width: 140rpx;
  flex-shrink: 0;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 8rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 2rpx solid transparent;
}

.member-item.active {
  border-color: #007AFF;
  background: rgba(0, 122, 255, 0.06);
}

.member-avatar {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: rgba(0, 122, 255, 0.12);
  color: #007AFF;
  font-size: 32rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12rpx;
}

.member-name {
  font-size: 24rpx;
  color: #1a1a1a;
  max-width: 120rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4rpx;
}

.member-relation {
  font-size: 20rpx;
  color: #999999;
}

.member-add {
  justify-content: center;
}

.member-add-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: #f0f2f5;
  color: #666666;
  font-size: 40rpx;
  font-weight: 400;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12rpx;
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

.profile-actions {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  flex-shrink: 0;
  margin-left: 16rpx;
}

.profile-action {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.18);
  text-align: center;
}

.profile-action.danger {
  background: rgba(255, 255, 255, 0.12);
  color: #ffd9d9;
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