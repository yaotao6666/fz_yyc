<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { staffProfileApi, staffHealthApi, staffSafetyApi } from '@/api'
import { formatDate } from '@/utils/format'

const profile = ref<any>(null)
const stats = ref({
  totalAccepted: 0,
  todayAccepted: 0,
  totalAmount: 0
})
const loading = ref(false)
// 我的服务区域（阶段三）
const myRegion = ref<any>(null)
// 我的质量分与近期评价（阶段四）
const quality = ref<any>(null)

const menuGroups = computed(() => [
  {
    title: '服务',
    items: [
      { icon: '📋', label: '全部工单', path: '/pages/workorder/index' }
    ]
  },
  {
    title: '账户',
    items: [
      { icon: '📝', label: '资料变更（需审核）', path: '/pages/profile/edit' },
      { icon: '📋', label: '我的审核记录', path: '/pages/profile/my-audits' }
    ]
  }
])

async function loadData() {
  loading.value = true
  try {
    const [profileRes, statsRes] = await Promise.all([
      staffProfileApi.getProfile(),
      staffProfileApi.getStatistics()
    ]) as any[]

    if (profileRes.code === 0) {
      profile.value = profileRes.data
    }
    if (statsRes.code === 0) {
      const s = statsRes.data || {}
      stats.value.totalAccepted = s.total_accepted || 0
      stats.value.todayAccepted = s.today_accepted || 0
      stats.value.totalAmount = Number(s.total_amount || 0)
    }
  } catch (e) {
    console.error('[Profile] loadData error', e)
  } finally {
    loading.value = false
  }
}

// 加载我的服务区域（阶段三：接单池按区域过滤）
async function loadMyRegion() {
  try {
    const res: any = await staffSafetyApi.getMyRegion()
    if (res?.code === 0) {
      myRegion.value = res.data
    }
  } catch (e) {
    console.error('[Profile] loadMyRegion error', e)
  }
}

// 加载我的质量分与近期评价（阶段四）
async function loadMyQualityScore() {
  try {
    const res: any = await staffProfileApi.getMyQualityScore()
    if (res?.code === 0) {
      quality.value = res.data || null
    }
  } catch (e) {
    console.error('[Profile] loadMyQualityScore error', e)
  }
}

function onTapItem(item: any) {
  if (!item.path) {
    uni.showToast({ title: '功能开发中', icon: 'none' })
    return
  }
  // tab 页用 switchTab，非 tab 页用 navigateTo
  const tabPaths = ['/pages/todo/index', '/pages/workorder/index', '/pages/profile/index']
  if (tabPaths.includes(item.path)) {
    uni.switchTab({ url: item.path })
  } else {
    uni.navigateTo({ url: item.path })
  }
}

function onLogout() {
  uni.showModal({
    title: '确认退出',
    content: '退出登录后需重新登录',
    success: (res) => {
      if (res.confirm) {
        uni.removeStorageSync('staff_token')
        uni.reLaunch({ url: '/pages/login/index' })
      }
    }
  })
}

// 服务区域详情：区域变更需联系商家后台维护
function showRegionDetail() {
  const limitless = myRegion.value?.region_limitless
  const list = myRegion.value?.region_list || []
  uni.showModal({
    title: '我的服务区域',
    content: limitless
      ? '当前未限定服务区域，可接全城订单。'
      : `当前服务区域：${list.join('、')}。待接订单池将按区域过滤，区域变更请联系商家后台维护。`,
    showCancel: false,
    confirmText: '知道了'
  })
}

onMounted(() => {
  loadData()
  loadMyRegion()
  loadMyQualityScore()
})
onShow(() => {
  // 每次显示时刷新统计数据与质量分
  loadMyQualityScore()
  if (profile.value) {
    staffProfileApi.getStatistics().then((res: any) => {
      if (res.code === 0) {
        const s = res.data || {}
        stats.value.totalAccepted = s.total_accepted || 0
        stats.value.todayAccepted = s.today_accepted || 0
        stats.value.totalAmount = Number(s.total_amount || 0)
      }
    }).catch(() => {})
  }
})
</script>

<template>
  <view class="page">
    <!-- 顶部信息 -->
    <view class="header">
      <view class="avatar-wrap">
        <text class="avatar-placeholder">{{ profile?.name?.[0] || '服' }}</text>
      </view>
      <view class="info">
        <view class="name">{{ profile?.name || '未登录' }}</view>
        <view class="sub">{{ profile?.phone || '服务人员端 v0.1.0' }}</view>
      </view>
      <view class="stats-row">
        <view class="stat">
          <view class="num">{{ stats.totalAccepted }}</view>
          <view class="lbl">累计接单</view>
        </view>
        <view class="stat">
          <view class="num">{{ stats.todayAccepted }}</view>
          <view class="lbl">今日接单</view>
        </view>
        <view class="stat">
          <view class="num">{{ stats.totalAmount.toFixed(0) }}</view>
          <view class="lbl">累计金额</view>
        </view>
      </view>
    </view>

    <view class="container" style="margin-top: -32rpx;">
      <!-- 我的服务区域（阶段三：接单池按区域过滤） -->
      <view class="card region-card" v-if="myRegion">
        <view class="region-row" @tap="showRegionDetail">
          <view class="region-info">
            <text class="region-title">我的服务区域</text>
            <text class="region-value">
              {{ myRegion.region_limitless ? '不限区域（可接全城订单）' : (myRegion.region_list || []).join('、') }}
            </text>
          </view>
          <text class="arrow">›</text>
        </view>
      </view>

      <!-- 我的质量分与近期评价（阶段四） -->
      <view class="card quality-card" v-if="quality">
        <view class="quality-head">
          <view class="quality-main">
            <text class="quality-score">{{ quality.quality_score ?? '5.0' }}</text>
            <text class="quality-unit">分</text>
          </view>
          <view class="quality-meta">
            <text class="quality-count">共 {{ quality.review_count }} 条评价</text>
            <text class="quality-sub">态度 {{ quality.avg_attitude }} · 专业 {{ quality.avg_professional }} · 准时 {{ quality.avg_punctual }}</text>
          </view>
        </view>
        <view class="quality-review-list" v-if="quality.recent_reviews?.length">
          <view v-for="r in quality.recent_reviews" :key="r.id" class="quality-review-item">
            <view class="quality-review-top">
              <text class="quality-review-stars">{{ '★★★★★'.slice(0, r.score) }}</text>
              <text class="quality-review-time">{{ formatDate(r.created_at) }}</text>
            </view>
            <text v-if="r.content" class="quality-review-content">{{ r.content }}</text>
          </view>
        </view>
        <text v-else class="quality-empty">暂无评价</text>
      </view>

      <view
        v-for="group in menuGroups"
        :key="group.title"
        class="card menu-card"
      >
        <view class="group-title">{{ group.title }}</view>
        <view
          v-for="(item, idx) in group.items"
          :key="item.label"
          class="menu-item"
          :class="{ 'border-top': idx > 0 }"
          @tap="onTapItem(item)"
        >
          <text class="icon">{{ item.icon }}</text>
          <text class="label">{{ item.label }}</text>
          <text v-if="item.badge" class="badge">{{ item.badge }}</text>
          <text class="arrow">›</text>
        </view>
      </view>

      <button class="logout-btn" @tap="onLogout">退出登录</button>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { min-height: 100vh; background: var(--bg-color); }
.header {
  background: linear-gradient(135deg, #517528 0%, #7a9e4f 100%);
  padding: 64rpx 32rpx 96rpx;
  color: #fff;
}
.avatar-wrap {
  width: 120rpx; height: 120rpx; border-radius: 50%;
  background: rgba(255,255,255,0.2);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 24rpx;
}
.avatar-placeholder { font-size: 48rpx; font-weight: 600; color: #fff; }
.info {
  .name { font-size: 40rpx; font-weight: 700; }
  .sub { font-size: 26rpx; opacity: 0.85; margin-top: 8rpx; }
}
.stats-row {
  display: flex; justify-content: space-around;
  margin-top: 48rpx;
  padding: 28rpx 0;
  background: rgba(255,255,255,0.12);
  border-radius: 20rpx;
  backdrop-filter: blur(10px);
  .stat { text-align: center;
    .num { font-size: 40rpx; font-weight: 700; line-height: 1.1; }
    .lbl { font-size: 22rpx; opacity: 0.85; margin-top: 8rpx; }
  }
}

.menu-card { padding: 0 32rpx !important; }

/* 我的服务区域卡片（阶段三） */
.region-card { padding: 24rpx 32rpx !important; }
.region-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.region-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  flex: 1;
}
.region-title { font-size: 30rpx; font-weight: 600; color: #333; }
.region-value { font-size: 26rpx; color: #666; line-height: 1.5; }
.region-card .arrow { font-size: 36rpx; color: #ccc; }

/* 我的质量分卡片（阶段四） */
.quality-card { padding: 24rpx 32rpx !important; }
.quality-head {
  display: flex;
  align-items: center;
  gap: 24rpx;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid var(--border-color);
}
.quality-main {
  display: flex;
  align-items: baseline;
}
.quality-score {
  font-size: 64rpx;
  font-weight: 700;
  color: #517528;
  line-height: 1;
}
.quality-unit { font-size: 24rpx; color: #999; margin-left: 8rpx; }
.quality-meta {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  flex: 1;
}
.quality-count { font-size: 26rpx; color: #333; }
.quality-sub { font-size: 22rpx; color: #999; }
.quality-review-list { margin-top: 8rpx; }
.quality-review-item {
  padding: 20rpx 0;
  border-bottom: 2rpx solid var(--border-color);
}
.quality-review-item:last-child { border-bottom: none; }
.quality-review-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.quality-review-stars { font-size: 28rpx; color: #ff9500; }
.quality-review-time { font-size: 22rpx; color: #999; }
.quality-review-content { font-size: 26rpx; color: #555; line-height: 1.5; margin-top: 8rpx; }
.quality-empty { font-size: 26rpx; color: #999; padding: 24rpx 0; text-align: center; }

.group-title {
  font-size: 24rpx; color: #999;
  padding: 24rpx 0 16rpx;
}
.menu-item {
  display: flex; align-items: center;
  padding: 28rpx 0;
  &.border-top { border-top: 2rpx solid var(--border-color); }
  .icon { font-size: 36rpx; margin-right: 20rpx; }
  .label { flex: 1; font-size: 30rpx; color: #333; }
  .badge {
    min-width: 36rpx;
    height: 36rpx;
    line-height: 36rpx;
    text-align: center;
    padding: 0 12rpx;
    border-radius: 18rpx;
    background: var(--danger-color);
    color: #fff;
    font-size: 22rpx;
    box-sizing: border-box;
  }
  .arrow { font-size: 36rpx; color: #ccc; }
}

.logout-btn {
  margin-top: 48rpx; margin-bottom: 48rpx;
  width: 100%;
  background: #fff; color: var(--danger-color);
  border: 2rpx solid #ffd1d3;
  border-radius: 48rpx;
  padding: 22rpx;
  font-size: 30rpx;
  line-height: 1.4;
}
</style>
