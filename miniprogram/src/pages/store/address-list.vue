<template>
  <view class="address-list-page">
    <view v-if="addresses.length" class="address-list">
      <view
        v-for="item in addresses"
        :key="item.id"
        class="address-item"
        :class="{ active: item.id === selectedAddressId }"
        @click="handleSelect(item)"
      >
        <view class="address-header">
          <view class="address-user">
            <text class="address-name">{{ item.name }}</text>
            <text class="address-phone">{{ item.phone }}</text>
            <text v-if="item.is_default" class="default-tag">默认</text>
          </view>
          <text v-if="item.id === selectedAddressId" class="selected-tag">已选中</text>
        </view>
        <view class="address-detail">{{ formatUserAddress(item) }}</view>
        <view class="address-actions">
          <view class="action-btn" @click.stop="handleSetDefault(item)">
            {{ item.is_default ? '默认地址' : '设为默认' }}
          </view>
          <view class="action-btn" @click.stop="goEdit(item)">编辑</view>
          <view class="action-btn danger" @click.stop="handleDelete(item)">删除</view>
        </view>
      </view>
    </view>

    <view v-else class="empty-state">
      <view class="empty-title">暂无收货地址</view>
      <view class="empty-desc">你可以新增地址，或从微信快速导入</view>
    </view>

    <view class="bottom-actions">
      <view class="bottom-btn secondary" @click="handleImportWechatAddress">微信导入</view>
      <view class="bottom-btn" @click="goCreate">新增地址</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { createUserAddress, deleteUserAddress, getUserAddresses, updateUserAddress } from '@api'
import type { UserAddress } from '@types'
import { useAuth } from '../../utils/useAuth'
import {
  STORE_ADDRESS_LIST_REFRESH_KEY,
  STORE_EDIT_ADDRESS_KEY,
  STORE_SELECTED_ADDRESS_ID_KEY,
  buildUserAddressPayload,
  chooseWechatAddress,
  formatUserAddress,
  isChooseAddressCancel,
  mapWechatAddressToUserAddress
} from '../../utils/address'

const addresses = ref<UserAddress[]>([])
const selectedAddressId = ref<number | null>(null)
const source = ref('confirm')

onLoad(async (options: any) => {
  source.value = options?.source || 'confirm'
  const currentSelectedId = Number(options?.selected_id || 0)
  selectedAddressId.value = currentSelectedId > 0 ? currentSelectedId : null

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }
  await loadAddresses()
})

onShow(async () => {
  const shouldRefresh = uni.getStorageSync(STORE_ADDRESS_LIST_REFRESH_KEY)
  if (shouldRefresh) {
    uni.removeStorageSync(STORE_ADDRESS_LIST_REFRESH_KEY)
    await loadAddresses()
  }
})

async function loadAddresses() {
  try {
    uni.showLoading({ title: '加载中...' })
    addresses.value = await getUserAddresses()
    if (!selectedAddressId.value && addresses.value.length) {
      selectedAddressId.value = addresses.value.find((item) => item.is_default)?.id || addresses.value[0].id || null
    }
  } catch (error: any) {
    uni.showToast({ title: error.message || '获取地址失败', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

function goCreate() {
  uni.removeStorageSync(STORE_EDIT_ADDRESS_KEY)
  uni.navigateTo({
    url: '/pages/store/address-edit?source=list'
  })
}

function goEdit(address: UserAddress) {
  uni.setStorageSync(STORE_EDIT_ADDRESS_KEY, address)
  uni.navigateTo({
    url: `/pages/store/address-edit?id=${address.id}&source=list`
  })
}

function finishSelect(address: UserAddress) {
  if (address.id) {
    uni.setStorageSync(STORE_SELECTED_ADDRESS_ID_KEY, address.id)
  }
  uni.navigateBack()
}

function handleSelect(address: UserAddress) {
  selectedAddressId.value = address.id || null
  finishSelect(address)
}

async function handleSetDefault(address: UserAddress) {
  if (address.is_default) {
    return
  }
  try {
    await updateUserAddress(address.id!, buildUserAddressPayload({ ...address, is_default: true }))
    uni.showToast({ title: '已设为默认地址', icon: 'success' })
    await loadAddresses()
    selectedAddressId.value = address.id || null
  } catch (error: any) {
    uni.showToast({ title: error.message || '设置默认失败', icon: 'none' })
  }
}

async function handleDelete(address: UserAddress) {
  try {
    await new Promise((resolve, reject) => {
      uni.showModal({
        title: '删除地址',
        content: '确认删除这个收货地址吗？',
        success: (res) => (res.confirm ? resolve(true) : reject(new Error('cancel')))
      })
    })
  } catch {
    return
  }

  try {
    await deleteUserAddress(address.id!)
    if (selectedAddressId.value === address.id) {
      selectedAddressId.value = null
    }
    uni.showToast({ title: '删除成功', icon: 'success' })
    await loadAddresses()
  } catch (error: any) {
    uni.showToast({ title: error.message || '删除失败', icon: 'none' })
  }
}

async function handleImportWechatAddress() {
  try {
    const imported = mapWechatAddressToUserAddress(await chooseWechatAddress())
    const created = await createUserAddress(buildUserAddressPayload({
      ...imported,
      is_default: !addresses.value.some((item) => item.is_default)
    }))
    uni.showToast({ title: '导入成功', icon: 'success' })
    if (source.value === 'confirm') {
      finishSelect(created)
      return
    }
    uni.setStorageSync(STORE_ADDRESS_LIST_REFRESH_KEY, true)
    await loadAddresses()
    selectedAddressId.value = created.id || null
  } catch (error: any) {
    if (isChooseAddressCancel(error)) {
      return
    }
    uni.showToast({ title: error.message || '导入失败', icon: 'none' })
  }
}
</script>

<style scoped>
.address-list-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 180rpx;
  box-sizing: border-box;
}

.address-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.address-item {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  border: 2rpx solid transparent;
}

.address-item.active {
  border-color: #007AFF;
  box-shadow: 0 10rpx 24rpx rgba(0, 122, 255, 0.08);
}

.address-header {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.address-user {
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex-wrap: wrap;
}

.address-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.address-phone {
  font-size: 26rpx;
  color: #666666;
}

.default-tag,
.selected-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
}

.default-tag {
  background: #e6f0ff;
  color: #007AFF;
}

.selected-tag {
  background: #eefbf3;
  color: #18a058;
  white-space: nowrap;
}

.address-detail {
  font-size: 26rpx;
  line-height: 1.6;
  color: #333333;
  word-break: break-all;
}

.address-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.action-btn {
  flex: 1;
  height: 64rpx;
  border-radius: 999rpx;
  background: #f5f7fa;
  color: #333333;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn.danger {
  color: #ff4d4f;
  background: #fff1f0;
}

.empty-state {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 72rpx 32rpx;
  text-align: center;
}

.empty-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12rpx;
}

.empty-desc {
  font-size: 26rpx;
  color: #999999;
}

.bottom-actions {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(env(safe-area-inset-bottom) + 24rpx);
  display: flex;
  gap: 16rpx;
}

.bottom-btn {
  flex: 1;
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

.bottom-btn.secondary {
  background: #ffffff;
  color: #007AFF;
  border: 2rpx solid #d9e7ff;
}
</style>
