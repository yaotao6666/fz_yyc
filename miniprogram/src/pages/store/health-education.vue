<template>
  <view class="education-page">
    <!-- 顶部说明 -->
    <view class="intro-banner">
      <view class="intro-title">📖 健康宣教</view>
      <view class="intro-desc">为您推荐的康复与慢病健康科普文章</view>
    </view>

    <!-- 分类筛选（按分类表拉取，含子级归属） -->
    <view v-if="categories.length" class="category-bar">
      <scroll-view scroll-x class="category-scroll" :show-scrollbar="false">
        <view class="category-list">
          <view
            class="category-tag"
            :class="{ active: !activeCategoryId }"
            @click="switchCategory(0)"
          >
            全部
          </view>
          <view
            v-for="category in topCategories"
            :key="category.id"
            class="category-tag"
            :class="{ active: activeCategoryId === category.id }"
            @click="switchCategory(category.id)"
          >
            {{ category.name }}
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 文章列表 -->
    <view v-if="loading && !articles.length" class="loading">加载中...</view>
    <view v-else-if="articles.length" class="article-list">
      <view
        v-for="item in filteredArticles"
        :key="item.id"
        class="article-card"
        @click="goDetail(item.id)"
      >
        <view class="article-title">{{ item.title || '未命名文章' }}</view>
        <view class="article-meta">
          <text v-if="categoryName(item)" class="category-chip">{{ categoryName(item) }}</text>
          <text v-if="item.publish_at" class="meta-text">{{ formatDate(item.publish_at) }}</text>
          <text v-if="item.views !== undefined" class="meta-text">{{ item.views }} 次浏览</text>
        </view>
      </view>
      <view v-if="!filteredArticles.length" class="empty">
        <view class="empty-icon">📖</view>
        <view class="empty-text">该分类暂无文章</view>
      </view>
    </view>
    <view v-else class="empty">
      <view class="empty-icon">📖</view>
      <view class="empty-text">暂无宣教文章</view>
      <view class="empty-desc">完成健康档案与评估后，将为您推送更匹配的健康科普</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { getStoreEducationCategories, getUserEducationArticles } from '../../api/health'
import type { EducationArticle, HealthEducationCategory } from '../../types'
import { useAuth } from '../../utils/useAuth'

const articles = ref<EducationArticle[]>([])
const categories = ref<HealthEducationCategory[]>([])
const loading = ref(false)
const activeCategoryId = ref(0) // 0 = 全部

/** 一级分类（parent_id=0，按 sort 升序） */
const topCategories = computed<HealthEducationCategory[]>(() =>
  categories.value
    .filter(c => c.parent_id === 0)
    .sort((a, b) => a.sort - b.sort)
)

/** 一级分类及其全部子级 id 集合：选中一级时，同时命中其下子级文章 */
function categoryIdSet(topId: number): Set<number> {
  const ids = new Set<number>([topId])
  categories.value.forEach(c => {
    if (c.parent_id === topId) ids.add(c.id)
  })
  return ids
}

/** 按当前分类客户端过滤（选中一级分类，包含其子级文章） */
const filteredArticles = computed(() => {
  if (!activeCategoryId.value) {
    return articles.value
  }
  const ids = categoryIdSet(activeCategoryId.value)
  return articles.value.filter(item => item.category_id !== undefined && item.category_id !== null && ids.has(item.category_id))
})

/** 通过 category_id 解析分类名（优先新分类，回退旧 category 字符串） */
function categoryName(item: EducationArticle): string {
  if (item.category_id) {
    const cat = categories.value.find(c => c.id === item.category_id)
    if (cat) return cat.name
  }
  return item.category || ''
}

onLoad(async () => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  loadCategories()
  loadList()
})

function switchCategory(id: number) {
  activeCategoryId.value = id
}

function formatDate(date?: string): string {
  if (!date) return ''
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}`
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/store/education-detail?id=${id}` })
}

async function loadCategories() {
  try {
    categories.value = await getStoreEducationCategories()
  } catch (error) {
    console.error('加载宣教分类失败:', error)
  }
}

async function loadList() {
  if (loading.value) return
  loading.value = true
  try {
    const data = await getUserEducationArticles()
    articles.value = data || []
  } catch (error) {
    console.error('加载宣教文章失败:', error)
  } finally {
    loading.value = false
  }
}

onPullDownRefresh(async () => {
  await Promise.all([loadCategories(), loadList()])
  uni.stopPullDownRefresh()
})
</script>

<style scoped>
.education-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 60rpx;
  box-sizing: border-box;
}

/* 顶部说明 */
.intro-banner {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.intro-title {
  font-size: 34rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.intro-desc {
  font-size: 24rpx;
  opacity: 0.85;
}

/* 分类筛选 */
.category-bar {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 20rpx 0;
  margin-bottom: 24rpx;
}

.category-scroll {
  width: 100%;
  white-space: nowrap;
}

.category-list {
  display: inline-flex;
  padding: 0 20rpx;
  gap: 16rpx;
}

.category-tag {
  flex-shrink: 0;
  font-size: 26rpx;
  color: #666666;
  background: #f5f5f5;
  padding: 10rpx 28rpx;
  border-radius: 999rpx;
}

.category-tag.active {
  color: #ffffff;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  font-weight: 600;
}

/* 文章列表 */
.article-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.article-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.article-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.5;
  margin-bottom: 20rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.category-chip {
  font-size: 22rpx;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 4rpx 16rpx;
  border-radius: 999rpx;
}

.meta-text {
  font-size: 22rpx;
  color: #999999;
}

.loading,
.empty {
  text-align: center;
  padding: 40rpx 0;
  font-size: 26rpx;
  color: #999999;
}

.empty-icon {
  font-size: 80rpx;
  margin-bottom: 16rpx;
}

.empty-text {
  font-size: 30rpx;
  color: #666666;
  margin-bottom: 10rpx;
}

.empty-desc {
  font-size: 24rpx;
  color: #999999;
}
</style>
