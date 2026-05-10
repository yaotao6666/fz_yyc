<template>
  <view class="staff-container">
    <view class="page-header">
      <view class="header-title">员工账号管理</view>
      <button class="add-btn" @click="openCreateDialog">新增员工</button>
    </view>

    <view v-if="staffList.length === 0" class="empty-card">
      <text>暂无员工，点击右上角新增</text>
    </view>

    <view v-for="item in staffList" :key="item.id" class="staff-card">
      <view class="staff-top">
        <view>
          <view class="staff-name">{{ item.name || item.username }}</view>
          <view class="staff-meta">{{ item.username }} · {{ formatRole(item.role) }}</view>
        </view>
        <view class="staff-status" :class="{ disabled: item.status !== 1 }">
          {{ item.status === 1 ? '启用中' : '已停用' }}
        </view>
      </view>

      <view class="staff-detail">
        <view>手机号：{{ item.phone || '未填写' }}</view>
        <view>下单提醒：{{ item.notify_enabled ? '开启' : '关闭' }}</view>
        <view>浏览提醒：{{ item.browse_notify_enabled ? '开启' : '关闭' }}</view>
        <view>微信快捷登录：{{ item.wechat_bound_at ? '已绑定' : '未绑定' }}</view>
      </view>

      <view class="staff-actions">
        <button class="action-btn" @click="openEditDialog(item)">编辑</button>
        <button class="action-btn" @click="openResetDialog(item)">重置密码</button>
        <button class="action-btn danger" @click="handleDelete(item)">删除</button>
      </view>
    </view>

    <view v-if="formDialogVisible" class="dialog-mask" @click="closeFormDialog">
      <view class="dialog-card" @click.stop>
        <view class="dialog-title">{{ editingStaffId ? '编辑员工' : '新增员工' }}</view>
        <view class="dialog-form">
          <input v-model="formData.name" class="dialog-input" placeholder="员工姓名" />
          <input v-model="formData.phone" class="dialog-input" type="number" placeholder="手机号" />
          <input v-model="formData.username" class="dialog-input" :disabled="!!editingStaffId" placeholder="登录账号" />
          <input
            v-if="!editingStaffId"
            v-model="formData.password"
            class="dialog-input"
            password
            placeholder="初始密码（至少6位）"
          />

          <picker :range="roleOptions" range-key="label" :value="roleIndex" @change="handleRoleChange">
            <view class="picker-row">{{ roleOptions[roleIndex].label }}</view>
          </picker>

          <picker :range="statusOptions" range-key="label" :value="statusIndex" @change="handleStatusChange">
            <view class="picker-row">{{ statusOptions[statusIndex].label }}</view>
          </picker>

          <view class="switch-row">
            <text>下单提醒</text>
            <switch :checked="formData.notify_enabled" color="#007AFF" @change="(e:any) => formData.notify_enabled = !!e.detail.value" />
          </view>
          <view class="switch-row">
            <text>浏览提醒</text>
            <switch :checked="formData.browse_notify_enabled" color="#007AFF" @change="(e:any) => formData.browse_notify_enabled = !!e.detail.value" />
          </view>
        </view>
        <view class="dialog-actions">
          <button class="dialog-btn secondary" @click="closeFormDialog">取消</button>
          <button class="dialog-btn primary" :disabled="formSaving" @click="submitForm">
            {{ formSaving ? '保存中...' : '保存' }}
          </button>
        </view>
      </view>
    </view>

    <view v-if="resetDialogVisible" class="dialog-mask" @click="closeResetDialog">
      <view class="dialog-card" @click.stop>
        <view class="dialog-title">重置密码</view>
        <view class="dialog-form">
          <view class="dialog-tip">将为 {{ resetTarget?.name || resetTarget?.username || '该员工' }} 设置新密码</view>
          <input v-model="resetPassword" class="dialog-input" password placeholder="请输入新密码（至少6位）" />
        </view>
        <view class="dialog-actions">
          <button class="dialog-btn secondary" @click="closeResetDialog">取消</button>
          <button class="dialog-btn primary" :disabled="resetSaving" @click="submitResetPassword">
            {{ resetSaving ? '提交中...' : '确认重置' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  createMerchantStaff,
  deleteMerchantStaff,
  getMerchantStaffList,
  resetMerchantStaffPassword,
  updateMerchantStaff
} from '@api'
import type { MerchantStaff } from '@types'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()

const staffList = ref<MerchantStaff[]>([])
const formDialogVisible = ref(false)
const resetDialogVisible = ref(false)
const formSaving = ref(false)
const resetSaving = ref(false)
const editingStaffId = ref<number | null>(null)
const resetTarget = ref<MerchantStaff | null>(null)
const resetPassword = ref('')

const roleOptions = [
  { label: '店主', value: 'owner' },
  { label: '店长', value: 'manager' },
  { label: '员工', value: 'staff' }
]
const statusOptions = [
  { label: '启用', value: 1 },
  { label: '停用', value: 0 }
]

const formData = reactive({
  name: '',
  phone: '',
  username: '',
  password: '',
  role: 'staff',
  status: 1,
  notify_enabled: true,
  browse_notify_enabled: true
})

const roleIndex = computed(() => {
  const index = roleOptions.findIndex((item) => item.value === formData.role)
  return index >= 0 ? index : 0
})
const statusIndex = computed(() => {
  const index = statusOptions.findIndex((item) => item.value === formData.status)
  return index >= 0 ? index : 0
})

onShow(() => {
  ensureOwnerAccess()
  loadStaffList()
})

function ensureOwnerAccess() {
  if (authStore.staff?.role === 'owner') {
    return
  }
  uni.showToast({ title: '仅店主可管理员工', icon: 'none' })
  setTimeout(() => {
    uni.navigateBack()
  }, 800)
}

async function loadStaffList() {
  try {
    const result = await getMerchantStaffList({ page: 1, page_size: 100 })
    staffList.value = result.list || []
  } catch (error) {
    console.error('加载员工列表失败:', error)
    uni.showToast({ title: '加载员工失败', icon: 'none' })
  }
}

function resetFormData() {
  formData.name = ''
  formData.phone = ''
  formData.username = ''
  formData.password = ''
  formData.role = 'staff'
  formData.status = 1
  formData.notify_enabled = true
  formData.browse_notify_enabled = true
}

function openCreateDialog() {
  editingStaffId.value = null
  resetFormData()
  formDialogVisible.value = true
}

function openEditDialog(staff: MerchantStaff) {
  editingStaffId.value = staff.id
  formData.name = staff.name || ''
  formData.phone = staff.phone || ''
  formData.username = staff.username || ''
  formData.password = ''
  formData.role = staff.role || 'staff'
  formData.status = staff.status ?? 1
  formData.notify_enabled = staff.notify_enabled ?? true
  formData.browse_notify_enabled = staff.browse_notify_enabled ?? true
  formDialogVisible.value = true
}

function closeFormDialog() {
  formDialogVisible.value = false
  editingStaffId.value = null
  resetFormData()
}

function openResetDialog(staff: MerchantStaff) {
  resetTarget.value = staff
  resetPassword.value = ''
  resetDialogVisible.value = true
}

function closeResetDialog() {
  resetDialogVisible.value = false
  resetTarget.value = null
  resetPassword.value = ''
}

function handleRoleChange(event: any) {
  const selected = roleOptions[Number(event.detail.value)]
  formData.role = selected?.value || 'staff'
}

function handleStatusChange(event: any) {
  const selected = statusOptions[Number(event.detail.value)]
  formData.status = selected?.value ?? 1
}

async function submitForm() {
  if (!formData.name) {
    uni.showToast({ title: '请输入员工姓名', icon: 'none' })
    return
  }
  if (!formData.phone) {
    uni.showToast({ title: '请输入手机号', icon: 'none' })
    return
  }
  if (!formData.username) {
    uni.showToast({ title: '请输入登录账号', icon: 'none' })
    return
  }
  if (!editingStaffId.value && formData.password.length < 6) {
    uni.showToast({ title: '初始密码至少 6 位', icon: 'none' })
    return
  }

  formSaving.value = true
  try {
    if (editingStaffId.value) {
      await updateMerchantStaff(editingStaffId.value, {
        name: formData.name,
        phone: formData.phone,
        role: formData.role,
        status: formData.status,
        notify_enabled: formData.notify_enabled,
        browse_notify_enabled: formData.browse_notify_enabled
      })
    } else {
      await createMerchantStaff({
        name: formData.name,
        phone: formData.phone,
        username: formData.username,
        password: formData.password,
        role: formData.role
      })
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    closeFormDialog()
    loadStaffList()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '保存失败', icon: 'none' })
  } finally {
    formSaving.value = false
  }
}

async function submitResetPassword() {
  if (!resetTarget.value) {
    return
  }
  if (resetPassword.value.length < 6) {
    uni.showToast({ title: '新密码至少 6 位', icon: 'none' })
    return
  }

  resetSaving.value = true
  try {
    await resetMerchantStaffPassword(resetTarget.value.id, resetPassword.value)
    uni.showToast({ title: '重置成功', icon: 'success' })
    closeResetDialog()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '重置失败', icon: 'none' })
  } finally {
    resetSaving.value = false
  }
}

function handleDelete(staff: MerchantStaff) {
  uni.showModal({
    title: '删除员工',
    content: `确认删除 ${staff.name || staff.username} 吗？`,
    success: async (res) => {
      if (!res.confirm) {
        return
      }
      try {
        await deleteMerchantStaff(staff.id)
        uni.showToast({ title: '删除成功', icon: 'success' })
        loadStaffList()
      } catch (error: any) {
        uni.showToast({ title: error?.message || '删除失败', icon: 'none' })
      }
    }
  })
}

function formatRole(role: string) {
  const roleMap: Record<string, string> = {
    owner: '店主',
    manager: '店长',
    staff: '员工'
  }
  return roleMap[role] || role
}
</script>

<style scoped>
.staff-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.header-title {
  font-size: 36rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.add-btn {
  margin: 0;
  padding: 0 28rpx;
  height: 72rpx;
  border-radius: 36rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-card,
.staff-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 20rpx;
}

.empty-card {
  text-align: center;
  color: #999999;
  font-size: 28rpx;
}

.staff-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.staff-name {
  font-size: 32rpx;
  color: #1a1a1a;
  font-weight: 600;
}

.staff-meta,
.staff-detail {
  font-size: 24rpx;
  color: #666666;
}

.staff-meta {
  margin-top: 8rpx;
}

.staff-status {
  padding: 8rpx 18rpx;
  background: #f6ffed;
  color: #389e0d;
  border-radius: 24rpx;
  font-size: 24rpx;
}

.staff-status.disabled {
  background: #fff1f0;
  color: #cf1322;
}

.staff-detail {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.staff-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.action-btn {
  flex: 1;
  height: 72rpx;
  border-radius: 36rpx;
  background: #f0f5ff;
  color: #0056CC;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn.danger {
  background: #fff2f0;
  color: #cf1322;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
}

.dialog-card {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.dialog-input,
.picker-row {
  height: 84rpx;
  border-radius: 16rpx;
  background: #f8f9fa;
  padding: 0 24rpx;
  font-size: 28rpx;
  display: flex;
  align-items: center;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8rpx 8rpx 8rpx 16rpx;
  border-radius: 16rpx;
  background: #f8f9fa;
  font-size: 28rpx;
  color: #1a1a1a;
}

.dialog-tip {
  font-size: 26rpx;
  color: #666666;
}

.dialog-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 28rpx;
}

.dialog-btn {
  flex: 1;
  height: 84rpx;
  border-radius: 42rpx;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-btn.secondary {
  background: #f5f5f5;
  color: #666666;
}

.dialog-btn.primary {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}
</style>
