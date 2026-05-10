<template>
  <view class="announcement-list-container">
    <!-- 顶部导航栏 -->
    <view class="nav-bar">
      <text class="nav-title">系统公告</text>
      <text class="publish-btn" @click="goToPublish">发布</text>
    </view>

    <!-- 公告列表 -->
    <scroll-view
      class="announcement-list"
      scroll-y
      @scrolltolower="loadMore"
    >
      <view
        v-for="item in announcements"
        :key="item.id"
        class="announcement-card"
        @click="goToDetail(item.id)"
      >
        <view class="card-header">
          <text class="card-title">{{ item.title }}</text>
          <text class="card-arrow">›</text>
        </view>
        <text class="card-summary">{{ item.summary }}</text>
        <text class="card-time">{{ formatDate(item.published_at || '') }}</text>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && announcements.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && announcements.length === 0" class="empty">
        <text class="empty-icon">📢</text>
        <text class="empty-text">暂无公告</text>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAnnouncements } from '@api'
import type { Announcement } from '@types'

// 列表数据
const announcements = ref<Announcement[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

onShow(() => {
  loadAnnouncements(true)
})

async function loadAnnouncements(reset = false) {
  if (reset) {
    page.value = 1
    noMore.value = false
    announcements.value = []
  }

  if (noMore.value || loading.value) return

  loading.value = true

  try {
    const res = await getAnnouncements({
      page: page.value,
      page_size: pageSize
    })

    if (reset) {
      announcements.value = res.list
    } else {
      announcements.value.push(...res.list)
    }

    if (res.list.length < pageSize) {
      noMore.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载公告列表失败:', error)
  } finally {
    loading.value = false
  }
}

function loadMore() {
  loadAnnouncements()
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/sp/announcements/edit?id=${id}` })
}

function goToPublish() {
  uni.navigateTo({ url: '/pages/sp/announcements/edit' })
}
</script>

<style scoped>
.announcement-list-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.nav-bar {
  background: #ffffff;
  padding: 24rpx 32rpx;
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 1rpx solid #f0f0f0;
}

.nav-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.publish-btn {
  padding: 12rpx 32rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 32rpx;
  font-size: 28rpx;
  font-weight: 500;
}

.announcement-list {
  padding: 24rpx;
}

.announcement-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 28rpx;
  margin-bottom: 20rpx;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  flex: 1;
  margin-right: 16rpx;
}

.card-arrow {
  font-size: 36rpx;
  color: #cccccc;
  font-weight: 300;
}

.card-summary {
  font-size: 28rpx;
  color: #666666;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 16rpx;
}

.card-time {
  font-size: 24rpx;
  color: #999999;
}

.loading, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 32rpx;
  color: #333333;
  font-weight: 500;
}
</style>
