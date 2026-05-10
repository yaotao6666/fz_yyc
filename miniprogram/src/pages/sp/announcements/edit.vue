<template>
  <view class="edit-container">
    <scroll-view class="form-scroll" scroll-y>
      <view class="section">
        <view class="form-item">
          <view class="form-label">
            公告标题 <text class="required">*</text>
          </view>
          <input
            v-model="formData.title"
            class="form-input"
            placeholder="请输入公告标题"
            :maxlength="50"
          />
          <view class="char-count">{{ formData.title.length }}/50</view>
        </view>

        <view class="form-item">
          <view class="form-label">
            公告内容 <text class="required">*</text>
          </view>
          <textarea
            v-model="formData.content"
            class="form-textarea"
            placeholder="请输入公告内容"
            :maxlength="2000"
          />
          <view class="char-count">{{ formData.content.length }}/2000</view>
        </view>
      </view>
    </scroll-view>

    <view class="bottom-bar">
      <button
        class="btn-draft"
        :disabled="submitting"
        @click="handleSaveDraft"
      >
        保存草稿
      </button>
      <button
        class="btn-publish"
        :disabled="submitting"
        @click="handlePublish"
      >
        {{ submitting ? '发布中...' : '立即发布' }}
      </button>
    </view>

    <!-- 确认对话框 -->
    <view v-if="showConfirm" class="confirm-overlay" @click="cancelConfirm">
      <view class="confirm-dialog" @click.stop>
        <view class="confirm-title">确认{{ confirmAction === 'publish' ? '发布' : '保存草稿' }}公告</view>
        <view class="confirm-content">
          <template v-if="confirmAction === 'publish'">
            <view>确定要立即发布此公告吗？</view>
            <view class="confirm-sub">发布后，所有商家将收到此公告通知。</view>
          </template>
          <template v-else>
            <view>确定要将此公告保存为草稿吗？</view>
          </template>
        </view>
        <view class="confirm-buttons">
          <button class="confirm-btn cancel" @click="cancelConfirm">取消</button>
          <button class="confirm-btn confirm" @click="confirmSubmit">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAnnouncement, createAnnouncement, updateAnnouncement } from '@api'
import type { AnnouncementStatus } from '@types'

const announcementId = ref<number | null>(null)
const submitting = ref(false)
const loading = ref(false)
const showConfirm = ref(false)
const confirmAction = ref<'publish' | 'draft'>('publish')

const formData = reactive({
  title: '',
  content: ''
})

onLoad((options: any) => {
  if (options.id) {
    announcementId.value = Number(options.id)
    uni.setNavigationBarTitle({ title: '编辑公告' })
    loadAnnouncement(announcementId.value)
  }
})

async function loadAnnouncement(id: number) {
  loading.value = true
  try {
    const announcement = await getAnnouncement(id)
    formData.title = announcement.title
    formData.content = announcement.content
  } catch (error: any) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function validateForm(): boolean {
  if (!formData.title.trim()) {
    uni.showToast({ title: '请输入公告标题', icon: 'none' })
    return false
  }
  if (!formData.content.trim()) {
    uni.showToast({ title: '请输入公告内容', icon: 'none' })
    return false
  }
  return true
}

function handleSaveDraft() {
  if (!validateForm()) return
  confirmAction.value = 'draft'
  showConfirm.value = true
}

function handlePublish() {
  if (!validateForm()) return
  confirmAction.value = 'publish'
  showConfirm.value = true
}

function cancelConfirm() {
  showConfirm.value = false
}

async function confirmSubmit() {
  showConfirm.value = false
  await submitForm(confirmAction.value === 'publish')
}

async function submitForm(publish: boolean) {
  submitting.value = true

  try {
    const data = {
      title: formData.title.trim(),
      content: formData.content.trim(),
      status: publish ? 1 : 0 as AnnouncementStatus
    }

    if (announcementId.value) {
      await updateAnnouncement(announcementId.value, data)
    } else {
      await createAnnouncement(data)
    }

    uni.showToast({
      title: publish ? '发布成功' : '保存成功',
      icon: 'success'
    })

    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.edit-container {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.form-scroll {
  flex: 1;
  padding-bottom: 140rpx;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.form-item {
  margin-bottom: 32rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 16rpx;
}

.required {
  color: #ff4d4f;
}

.form-input {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.form-textarea {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 30rpx;
  color: #1a1a1a;
  width: 100%;
  height: 400rpx;
  box-sizing: border-box;
}

.char-count {
  font-size: 24rpx;
  color: #999999;
  text-align: right;
  margin-top: 8rpx;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 24rpx;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-draft,
.btn-publish {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-draft {
  background: #f5f5f5;
  color: #666666;
}

.btn-publish {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
}

.btn-draft[disabled],
.btn-publish[disabled] {
  background: #cccccc;
  color: #ffffff;
}

.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.confirm-dialog {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  overflow: hidden;
}

.confirm-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  text-align: center;
  padding: 32rpx 24rpx 24rpx;
}

.confirm-content {
  padding: 0 32rpx 24rpx;
  font-size: 28rpx;
  color: #666666;
  line-height: 1.6;
}

.confirm-sub {
  font-size: 24rpx;
  color: #999999;
  margin-top: 12rpx;
}

.confirm-buttons {
  display: flex;
  border-top: 1rpx solid #eeeeee;
}

.confirm-btn {
  flex: 1;
  height: 96rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  border: none;
  border-radius: 0;
  background: transparent;
}

.confirm-btn.cancel {
  color: #666666;
  border-right: 1rpx solid #eeeeee;
}

.confirm-btn.confirm {
  color: #667eea;
  font-weight: 500;
}
</style>
