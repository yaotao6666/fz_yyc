<template>
  <view class="address-edit-page">
    <view class="form-card">
      <view class="form-item">
        <view class="form-label">联系人</view>
        <input v-model="form.name" class="form-input" placeholder="请输入联系人姓名" />
      </view>

      <view class="form-item">
        <view class="form-label">联系电话</view>
        <input v-model="form.phone" class="form-input" type="number" maxlength="11" placeholder="请输入联系电话" />
      </view>

      <view class="form-item">
        <view class="form-label">省市区</view>
        <picker mode="region" :value="regionValue" @change="handleRegionChange">
          <view class="picker-value" :class="{ placeholder: !hasRegion }">
            {{ hasRegion ? regionValue.join(' ') : '请选择省市区' }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <view class="form-label">详细地址</view>
        <textarea
          v-model="form.address"
          class="form-textarea"
          maxlength="100"
          placeholder="请输入详细地址，如街道、楼栋、门牌号"
        />
      </view>

      <view class="switch-row">
        <view class="switch-info">
          <view class="switch-title">设为默认地址</view>
          <view class="switch-desc">下次下单将优先使用该地址</view>
        </view>
        <switch :checked="Boolean(form.is_default)" color="#007AFF" @change="handleDefaultChange" />
      </view>
    </view>

    <view class="bottom-bar">
      <view class="submit-btn" :class="{ disabled: saving }" @click="handleSubmit">
        {{ saving ? '保存中...' : isEditMode ? '保存修改' : '保存地址' }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { createUserAddress, updateUserAddress } from '@api'
import type { UserAddress } from '@types'
import { useAuth } from '../../utils/useAuth'
import {
  STORE_ADDRESS_LIST_REFRESH_KEY,
  STORE_EDIT_ADDRESS_KEY,
  STORE_SELECTED_ADDRESS_ID_KEY,
  buildUserAddressPayload
} from '../../utils/address'

const source = ref('list')
const addressId = ref<number | null>(null)
const saving = ref(false)
const form = reactive<Partial<UserAddress>>({
  name: '',
  phone: '',
  province: '',
  city: '',
  district: '',
  address: '',
  is_default: false
})

const isEditMode = computed(() => !!addressId.value)
const regionValue = computed(() => [form.province || '', form.city || '', form.district || ''])
const hasRegion = computed(() => Boolean(form.province || form.city || form.district))

onLoad(async (options: any) => {
  source.value = options?.source || 'list'
  addressId.value = Number(options?.id || 0) || null

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  if (isEditMode.value) {
    const cachedAddress = uni.getStorageSync(STORE_EDIT_ADDRESS_KEY) as UserAddress | undefined
    if (cachedAddress && cachedAddress.id === addressId.value) {
      fillForm(cachedAddress)
    }
  }
})

function fillForm(address: Partial<UserAddress>) {
  form.name = address.name || ''
  form.phone = address.phone || ''
  form.province = address.province || ''
  form.city = address.city || ''
  form.district = address.district || ''
  form.address = address.address || ''
  form.lat = address.lat
  form.lng = address.lng
  form.is_default = Boolean(address.is_default)
}

function handleRegionChange(event: any) {
  const [province, city, district] = event.detail.value || []
  form.province = province || ''
  form.city = city || ''
  form.district = district || ''
}

function handleDefaultChange(event: any) {
  form.is_default = Boolean(event.detail.value)
}

function validateForm() {
  if (!(form.name || '').trim()) {
    uni.showToast({ title: '请输入联系人', icon: 'none' })
    return false
  }
  if (!/^1\d{10}$/.test((form.phone || '').trim())) {
    uni.showToast({ title: '请输入正确的联系电话', icon: 'none' })
    return false
  }
  if (!(form.province || '').trim() || !(form.city || '').trim() || !(form.district || '').trim()) {
    uni.showToast({ title: '请选择省市区', icon: 'none' })
    return false
  }
  if (!(form.address || '').trim()) {
    uni.showToast({ title: '请输入详细地址', icon: 'none' })
    return false
  }
  return true
}

async function handleSubmit() {
  if (saving.value || !validateForm()) {
    return
  }

  try {
    saving.value = true
    const payload = buildUserAddressPayload(form)
    const savedAddress = isEditMode.value
      ? await updateUserAddress(addressId.value!, payload)
      : await createUserAddress(payload)

    uni.setStorageSync(STORE_ADDRESS_LIST_REFRESH_KEY, true)
    if (source.value === 'confirm' && savedAddress.id) {
      uni.setStorageSync(STORE_SELECTED_ADDRESS_ID_KEY, savedAddress.id)
    }

    uni.showToast({ title: isEditMode.value ? '保存成功' : '新增成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 400)
  } catch (error: any) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.address-edit-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 160rpx;
  box-sizing: border-box;
}

.form-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.form-input,
.picker-value,
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  background: #f8f9fa;
  border-radius: 14rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.form-input,
.picker-value {
  height: 84rpx;
  padding: 0 24rpx;
  display: flex;
  align-items: center;
}

.picker-value.placeholder {
  color: #999999;
}

.form-textarea {
  min-height: 200rpx;
  padding: 20rpx 24rpx;
  line-height: 1.5;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding-top: 8rpx;
}

.switch-title {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 600;
}

.switch-desc {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.bottom-bar {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(env(safe-area-inset-bottom) + 24rpx);
}

.submit-btn {
  height: 88rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-btn.disabled {
  opacity: 0.72;
}
</style>
