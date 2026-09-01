<template>
  <view class="review-submit-container">
    <!-- 服务人员信息 -->
    <view class="staff-card">
      <view class="staff-avatar">👤</view>
      <view class="staff-info">
        <view class="staff-name">{{ assignedStaffName || '服务人员' }}</view>
        <view class="staff-tip">请根据本次服务体验，如实填写评价</view>
      </view>
    </view>

    <!-- 评分维度 -->
    <view class="score-card">
      <view v-for="dim in dimensions" :key="dim.key" class="score-dim">
        <view class="dim-label">{{ dim.label }}</view>
        <view class="dim-stars">
          <text
            v-for="n in 5"
            :key="n"
            class="star"
            :class="{ active: n <= dim.value }"
            @click="setScore(dim.key, n)"
          >★</text>
        </view>
        <view class="dim-text">{{ scoreText(dim.value) }}</view>
      </view>
    </view>

    <!-- 评价内容 -->
    <view class="content-card">
      <view class="content-title">评价内容</view>
      <textarea
        v-model="content"
        class="content-input"
        :maxlength="512"
        placeholder="说说本次服务的体验感受吧（选填）"
        placeholder-class="content-placeholder"
      />
      <view class="content-count">{{ content.length }}/512</view>
    </view>

    <!-- 图片上传 -->
    <view class="image-card">
      <view class="image-title">上传图片（选填，最多 9 张）</view>
      <view class="image-grid">
        <view v-for="(img, index) in images" :key="index" class="image-item">
          <image class="preview-image" :src="img" mode="aspectFill" @click="previewImage(index)" />
          <view class="image-remove" @click="removeImage(index)">×</view>
        </view>
        <view v-if="images.length < 9" class="image-add" @click="chooseImages">+</view>
      </view>
    </view>

    <!-- 提交 -->
    <view class="submit-bar">
      <button class="submit-btn" :disabled="submitting || !allScored" @click="onSubmit">
        {{ submitting ? '提交中...' : '提交评价' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getMyOrderDetail, submitReview, uploadImage } from '@api'
import { useAuth } from '../../utils/useAuth'

const orderId = ref(0)
const assignedStaffName = ref('')
const submitting = ref(false)

const dimensions = ref([
  { key: 'score', label: '总体评分', value: 0 },
  { key: 'attitude_score', label: '服务态度', value: 0 },
  { key: 'professional_score', label: '专业技能', value: 0 },
  { key: 'punctual_score', label: '准时守约', value: 0 }
])

const content = ref('')
const images = ref<string[]>([])

const allScored = computed(() => dimensions.value.every(d => d.value > 0))

function scoreText(value: number): string {
  return ['', '非常差', '较差', '一般', '满意', '非常满意'][value] || ''
}

function setScore(key: string, n: number) {
  const dim = dimensions.value.find(d => d.key === key)
  if (dim) dim.value = n
}

onLoad(async (options: any) => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }

  orderId.value = Number(options?.id || 0)
  if (orderId.value <= 0) {
    uni.showToast({ title: '订单参数错误', icon: 'none' })
    return
  }

  try {
    const order = await getMyOrderDetail(orderId.value)
    assignedStaffName.value = order.assigned_staff_name || ''
  } catch (error: any) {
    uni.showToast({ title: error?.message || '加载失败', icon: 'none' })
  }
})

function chooseImages() {
  const remain = 9 - images.value.length
  uni.chooseImage({
    count: remain,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const paths = res.tempFilePaths || []
      await uploadAll(paths)
    }
  })
}

async function uploadAll(paths: string[]) {
  for (const path of paths) {
    try {
      const uploaded = await uploadImage(path)
      if (uploaded?.url) {
        images.value.push(uploaded.url)
      }
    } catch (error) {
      uni.showToast({ title: '图片上传失败', icon: 'none' })
    }
  }
}

function removeImage(index: number) {
  images.value.splice(index, 1)
}

function previewImage(index: number) {
  uni.previewImage({
    current: images.value[index],
    urls: images.value
  })
}

async function onSubmit() {
  if (!allScored.value) {
    uni.showToast({ title: '请完成所有评分', icon: 'none' })
    return
  }
  if (submitting.value) return

  submitting.value = true
  try {
    await submitReview(orderId.value, {
      score: dimensions.value[0].value,
      attitude_score: dimensions.value[1].value,
      professional_score: dimensions.value[2].value,
      punctual_score: dimensions.value[3].value,
      content: content.value.trim() || undefined,
      images: images.value.length ? images.value : undefined
    })
    uni.showToast({ title: '评价成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 1200)
  } catch (error: any) {
    uni.showToast({ title: error?.message || '提交失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.review-submit-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 160rpx;
}

.staff-card {
  display: flex;
  align-items: center;
  margin: 24rpx;
  padding: 32rpx;
  border-radius: 20rpx;
  background: linear-gradient(135deg, #1677ff 0%, #0b57d0 100%);
  color: #ffffff;
}

.staff-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.22);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
}

.staff-info {
  flex: 1;
  min-width: 0;
}

.staff-name {
  font-size: 34rpx;
  font-weight: 600;
  margin-bottom: 8rpx;
}

.staff-tip {
  font-size: 24rpx;
  opacity: 0.9;
}

.score-card {
  margin: 24rpx;
  padding: 12rpx 32rpx;
  border-radius: 20rpx;
  background: #ffffff;
}

.score-dim {
  display: flex;
  align-items: center;
  padding: 26rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.score-dim:last-child {
  border-bottom: none;
}

.dim-label {
  width: 160rpx;
  font-size: 28rpx;
  color: #1f2329;
}

.dim-stars {
  flex: 1;
  display: flex;
  gap: 8rpx;
}

.star {
  font-size: 48rpx;
  color: #e5e6eb;
  transition: color 0.1s;
}

.star.active {
  color: #ff9500;
}

.dim-text {
  width: 120rpx;
  text-align: right;
  font-size: 24rpx;
  color: #86909c;
}

.content-card {
  margin: 24rpx;
  padding: 32rpx;
  border-radius: 20rpx;
  background: #ffffff;
}

.content-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 20rpx;
}

.content-input {
  width: 100%;
  height: 180rpx;
  font-size: 28rpx;
  color: #1f2329;
  line-height: 1.6;
}

.content-placeholder {
  color: #c0c4cc;
}

.content-count {
  margin-top: 12rpx;
  text-align: right;
  font-size: 22rpx;
  color: #c0c4cc;
}

.image-card {
  margin: 24rpx;
  padding: 32rpx;
  border-radius: 20rpx;
  background: #ffffff;
}

.image-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 20rpx;
}

.image-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
}

.preview-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f2f3f5;
}

.image-remove {
  position: absolute;
  top: -16rpx;
  right: -16rpx;
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.6);
  color: #ffffff;
  font-size: 32rpx;
  line-height: 44rpx;
  text-align: center;
}

.image-add {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f7f8fa;
  border: 2rpx dashed #d9d9d9;
  font-size: 64rpx;
  color: #c0c4cc;
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -8rpx 24rpx rgba(0, 0, 0, 0.06);
}

.submit-btn {
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 999rpx;
  font-size: 30rpx;
  font-weight: 600;
  border: none;
  background: #1677ff;
  color: #ffffff;
}

.submit-btn[disabled] {
  background: #c0c4cc;
  color: #ffffff;
}
</style>