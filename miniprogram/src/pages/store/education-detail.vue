<template>
  <view class="education-detail-page">
    <view v-if="loading && !article" class="loading">加载中...</view>

    <template v-else-if="article">
      <!-- 标题与元信息 -->
      <view class="article-header">
        <view class="article-title">{{ article.title || '未命名文章' }}</view>
        <view class="article-meta">
          <text v-if="article.category" class="category-chip">{{ article.category }}</text>
          <text v-if="article.publish_at" class="meta-text">{{ formatDate(article.publish_at) }}</text>
          <text v-if="article.views !== undefined" class="meta-text">{{ article.views }} 次浏览</text>
        </view>
      </view>

      <!-- 定向标签 -->
      <view class="tag-bar" v-if="article.tags?.length">
        <text v-for="tag in article.tags" :key="tag" class="tag-chip"># {{ tag }}</text>
      </view>

      <!-- 封面图 -->
      <image v-if="article.cover" class="article-cover" :src="article.cover" mode="widthFix" />

      <!-- 正文：包含 HTML 标签时用 rich-text，否则按纯文本换行渲染 -->
      <view class="article-body">
        <rich-text v-if="isHtmlContent" class="article-rich" :nodes="article.content || ''" />
        <text v-else class="article-text">{{ article.content }}</text>
      </view>
    </template>

    <view v-else class="empty">
      <view class="empty-icon">📖</view>
      <view class="empty-text">文章不存在或已下线</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getUserEducationArticle } from '../../api/health'
import type { EducationArticle } from '../../types'
import { useAuth } from '../../utils/useAuth'

const article = ref<EducationArticle | null>(null)
const loading = ref(false)

/** 内容包含 HTML 标签时用 rich-text 渲染，否则按纯文本换行渲染 */
const isHtmlContent = computed(() => {
  const content = article.value?.content || ''
  return content.includes('<') && content.includes('>')
})

onLoad(async (query) => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  const id = Number(query?.id)
  if (!id) {
    uni.showToast({ title: '文章参数错误', icon: 'none' })
    return
  }
  loadArticle(id)
})

function formatDate(date?: string): string {
  if (!date) return ''
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}`
}

async function loadArticle(id: number) {
  loading.value = true
  try {
    article.value = await getUserEducationArticle(id)
  } catch (error) {
    console.error('加载文章详情失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.education-detail-page {
  min-height: 100vh;
  background: #ffffff;
  padding: 32rpx 32rpx 60rpx;
  box-sizing: border-box;
}

.loading,
.empty {
  text-align: center;
  padding: 80rpx 0;
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
}

/* 标题与元信息 */
.article-header {
  padding-bottom: 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
  margin-bottom: 24rpx;
}

.article-title {
  font-size: 38rpx;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.5;
  margin-bottom: 20rpx;
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

/* 定向标签 */
.tag-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 24rpx;
}

.tag-chip {
  font-size: 22rpx;
  color: #666666;
  background: #f5f5f5;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
}

/* 封面图 */
.article-cover {
  width: 100%;
  border-radius: 16rpx;
  margin-bottom: 24rpx;
  background: #f0f0f0;
}

/* 正文 */
.article-text {
  font-size: 30rpx;
  color: #333333;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-all;
}

.article-rich {
  font-size: 30rpx;
  color: #333333;
  line-height: 1.8;
  word-break: break-all;
}
</style>
