<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { staffProfileApi, staffHealthApi } from '@/api'

const profile = ref<any>(null)
const stats = ref({
  totalAccepted: 0,
  todayAccepted: 0,
  totalAmount: 0
})
const loading = ref(false)
// 待执行随访任务数（角标）
const pendingFollowUps = ref(0)

const menuGroups = computed(() => [
  {
    title: '服务',
    items: [
      { icon: '📋', label: '全部工单', path: '/pages/workorder/index' },
      { icon: '📊', label: '接单统计', path: '' },
      { icon: '🧰', label: '验机归还', path: '' },
      { icon: '🩺', label: '我的照护计划', path: '/pages/health/care-plans' },
      { icon: '📞', label: '随访任务', path: '/pages/health/follow-ups', badge: pendingFollowUps.value }
    ]
  },
  {
    title: '账户',
    items: [
      { icon: '🔔', label: '消息通知', path: '' },
      { icon: '⚙️', label: '设置', path: '' },
      { icon: '❓', label: '帮助与反馈', path: '' }
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

// 加载待执行随访任务数，用于入口角标
async function loadPendingFollowUps() {
  try {
    const res: any = await staffHealthApi.getMyFollowUpTasks({ status: 0, page: 1, page_size: 1 })
    pendingFollowUps.value = res?.total || 0
  } catch (e) {
    console.error('[Profile] loadPendingFollowUps error', e)
  }
}

function onTapItem(item: any) {
  if (!item.path) {
    uni.showToast({ title: '功能开发中', icon: 'none' })
    return
  }
  // tab 页用 switchTab，非 tab 页用 navigateTo
  const tabPaths = ['/pages/todo/index', '/pages/workorder/index', '/pages/schedule/index', '/pages/profile/index']
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

onMounted(() => {
  loadData()
  loadPendingFollowUps()
})
onShow(() => {
  // 每次显示时刷新统计数据与待执行随访任务数
  loadPendingFollowUps()
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
