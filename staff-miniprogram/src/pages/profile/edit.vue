<template>
  <view class="page">
    <view class="container">
      <view class="card">
        <view class="form-item">
          <view class="form-label">姓名</view>
          <input v-model="form.name" placeholder="请输入姓名" class="form-input" />
        </view>
        <view class="form-item">
          <view class="form-label">手机号</view>
          <input v-model="form.phone" placeholder="请输入手机号" type="number" maxlength="11" class="form-input" />
        </view>
        <view class="form-item">
          <view class="form-label">资质材料（可更新）</view>
          <view class="qual-list">
            <view v-for="(q, i) in form.qualifications" :key="i" class="qual-item">
              <input v-model="q.name" placeholder="材料名称（如执业证书）" class="qual-input" />
              <image v-if="q.url" :src="q.url" class="qual-img" mode="aspectFill" @click="previewQual(i)" />
              <text class="qual-remove" @click="form.qualifications.splice(i, 1)">移除</text>
            </view>
            <view class="qual-add" @click="pickQualification">+ 添加材料</view>
          </view>
        </view>
      </view>
      <button class="btn-primary" :loading="saving" @click="handleSubmit">提交变更（需审核）</button>
      <text class="tip">提交后需管理员审核通过才会生效，不影响当前账号启用状态。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffProfileApi, uploadQualificationFile } from '@/api'

const saving = ref(false)
const form = reactive({
  name: '',
  phone: '',
  avatar: '',
  qualifications: [] as { type: string; name: string; url: string }[]
})

function previewQual(index: number) {
  const q = form.qualifications[index]
  if (q?.url) uni.previewImage({ urls: [q.url] })
}

function pickQualification() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: (res) => {
      uploadQualificationFile(res.tempFilePaths[0])
        .then(({ url }) => form.qualifications.push({ type: 'certificate', name: '', url }))
        .catch(() => {})
    }
  })
}

async function loadProfile() {
  const res: any = await staffProfileApi.getProfile()
  if (res.code === 0 && res.data) {
    form.name = res.data.name || ''
    form.phone = res.data.phone || ''
    form.avatar = res.data.avatar || ''
    const raw = res.data.qualifications
    let quals: any[] = []
    if (Array.isArray(raw)) quals = raw
    else if (typeof raw === 'string' && raw.trim()) { try { quals = JSON.parse(raw) } catch {} }
    form.qualifications = quals.map((q: any) => ({ type: q.type || '', name: q.name || '', url: q.url || '' }))
  }
}

async function handleSubmit() {
  if (!form.name || !form.phone) {
    uni.showToast({ title: '请填写姓名和手机号', icon: 'none' })
    return
  }
  saving.value = true
  try {
    const res: any = await staffProfileApi.requestProfileChange({
      name: form.name,
      phone: form.phone,
      avatar: form.avatar,
      qualifications: form.qualifications
    })
    if (res.code === 0) {
      uni.showToast({ title: '变更已提交审核', icon: 'success' })
      setTimeout(() => uni.navigateBack(), 1500)
    } else {
      uni.showToast({ title: res.message || '提交失败', icon: 'none' })
    }
  } catch (err: any) {
    uni.showToast({ title: err?.data?.message || '提交失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}

onLoad(() => {
  loadProfile()
})
</script>

<style lang="scss" scoped>
.page { min-height: 100vh; background: var(--bg-color); }
.container { padding: 24rpx 32rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 32rpx; }
.form-item { margin-bottom: 28rpx; }
.form-label { font-size: 28rpx; color: #333; margin-bottom: 12rpx; }
.form-input {
  width: 100%; height: 88rpx; background: #f8f8f8; border: 2rpx solid #eee;
  border-radius: 12rpx; padding: 0 24rpx; font-size: 30rpx; box-sizing: border-box;
}
.btn-primary {
  margin-top: 32rpx; width: 100%; height: 92rpx; line-height: 92rpx;
  background: #517528; color: #fff; font-size: 32rpx; font-weight: 600;
  border-radius: 12rpx; border: none;
}
.btn-primary::after { border: none; }
.tip { display: block; text-align: center; margin-top: 24rpx; font-size: 24rpx; color: #999; }
.qual-list { display: flex; flex-wrap: wrap; gap: 16rpx; }
.qual-item { width: 200rpx; display: flex; flex-direction: column; align-items: center; }
.qual-input { width: 100%; height: 60rpx; background: #f8f8f8; border: 2rpx solid #eee; border-radius: 10rpx; padding: 0 16rpx; font-size: 22rpx; box-sizing: border-box; margin-bottom: 10rpx; }
.qual-img { width: 200rpx; height: 150rpx; border-radius: 12rpx; background: #f0f0f0; }
.qual-remove { font-size: 22rpx; color: #f56c6c; margin-top: 8rpx; }
.qual-add { width: 200rpx; height: 150rpx; border: 2rpx dashed #ccc; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; color: #999; font-size: 26rpx; background: #fafafa; }
</style>