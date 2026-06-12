<template>
  <view class="delivery-settings-container">
    <!-- 下单方式 -->
    <view class="section">
      <view class="section-title">下单方式</view>

      <view class="switch-row mode-row">
        <view class="switch-label">
          <text class="label-title">开启堂食</text>
          <text class="label-desc">开启后用户可选择堂食方式下单</text>
        </view>
        <switch
          :checked="modeForm.dine_in_enabled"
          @change="onDineInSwitchChange"
          color="#007AFF"
        />
      </view>

      <view class="switch-row mode-row">
        <view class="switch-label">
          <text class="label-title">开启自提</text>
          <text class="label-desc">开启后用户可选择自提方式下单</text>
        </view>
        <switch
          :checked="modeForm.pickup_enabled"
          @change="onPickupSwitchChange"
          color="#007AFF"
        />
      </view>


      <view class="switch-row mode-row">
        <view class="switch-label">
          <text class="label-title">开启配送服务</text>
          <text class="label-desc">开启后用户可选择配送方式下单</text>
        </view>
        <switch
          :checked="modeForm.takeout_enabled"
          @change="onDeliverySwitchChange"
          color="#007AFF"
        />
      </view>

    </view>

    <!-- 配送设置 -->
    <view class="section" v-if="modeForm.takeout_enabled">
      <view class="section-title">配送费用</view>

      <view class="form-item">
        <view class="form-label">基础配送费</view>
        <view class="input-row">
          <input
            v-model.number="formData.base_fee"
            type="digit"
            class="form-input"
            placeholder="0.00"
          />
          <text class="input-suffix">元</text>
        </view>
      </view>

      <view class="form-item">
        <view class="form-label">满额免配送费</view>
        <view class="input-row">
          <input
            v-model.number="formData.free_delivery_amount"
            type="digit"
            class="form-input"
            placeholder="0.00"
          />
          <text class="input-suffix">元</text>
        </view>
        <view class="form-hint">订单金额达到此金额时免配送费</view>
      </view>

      <view class="form-item">
        <view class="form-label">最大配送距离</view>
        <view class="input-row">
          <input
            v-model.number="formData.max_distance"
            type="number"
            class="form-input"
            placeholder="10"
          />
          <text class="input-suffix">公里</text>
        </view>
        <view class="form-hint">仅用于用户选择配送档位，不进行真实定位计算</view>
      </view>
    </view>

    <!-- 距离规则 -->
    <view class="section" v-if="modeForm.takeout_enabled">
      <view class="section-header">
        <view class="section-title">按距离收费</view>
        <view class="add-rule-btn" @click="addRule">添加规则</view>
      </view>
      <view class="form-hint">用户下单时手动选择商家支持的距离档位，超出范围仅提示不可下单</view>

      <view
        v-for="(rule, index) in formData.distance_rules"
        :key="index"
        class="rule-item"
      >
        <view class="rule-header">
          <text class="rule-title">规则 {{ index + 1 }}</text>
          <text class="delete-rule" @click="deleteRule(index)">删除</text>
        </view>
        <view class="rule-content">
          <view class="rule-input">
            <input
              v-model.number="rule.min_distance"
              type="number"
              class="input"
              placeholder="0"
            />
            <text class="rule-unit">公里</text>
            <text class="rule-text">至</text>
            <input
              v-model.number="rule.max_distance"
              type="number"
              class="input"
              placeholder="5"
            />
            <text class="rule-unit">公里</text>
          </view>
          <view class="rule-fee">
            <input
              v-model.number="rule.fee"
              type="digit"
              class="input fee-input"
              placeholder="0"
            />
            <text class="fee-unit">元</text>
          </view>
        </view>
      </view>

      <view v-if="formData.distance_rules.length === 0" class="empty-rules">
        <text>暂无收费规则，点击上方按钮添加</text>
      </view>
    </view>

    <view class="section" v-if="modeForm.pickup_enabled">
      <view class="section-header">
        <view class="section-title">自提点</view>
        <view class="add-rule-btn" @click="openPickupDialog()">新增自提点</view>
      </view>

      <view v-if="pickupPoints.length === 0" class="empty-rules">
        <text>暂无自提点，点击右上角添加</text>
      </view>

      <view v-for="point in pickupPoints" :key="point.id" class="pickup-item">
        <view class="pickup-top">
          <view class="pickup-title">
            <text class="pickup-name">{{ point.name }}</text>
            <text v-if="point.is_default" class="pickup-tag">默认</text>
            <text v-if="point.status !== 1" class="pickup-tag disabled">停用</text>
          </view>
          <view class="pickup-actions">
            <text class="pickup-action" @click="editPickupPoint(point)">编辑</text>
            <text class="pickup-action danger" @click="removePickupPoint(point)">删除</text>
          </view>
        </view>
        <view class="pickup-address" @click="openPickupLocation(point)">{{ point.address }}</view>
        <view class="pickup-bottom">
          <text class="pickup-link" @click="setDefaultPickupPoint(point)" v-if="!point.is_default">设为默认</text>
          <text class="pickup-link" @click="togglePickupPointStatus(point)">
            {{ point.status === 1 ? '停用' : '启用' }}
          </text>
        </view>
      </view>
    </view>

    <!-- 保存按钮 -->
    <view class="save-area">
      <button class="btn-save" :disabled="saving" @click="handleSave">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </view>

    <view v-if="pickupDialogVisible" class="dialog-mask" @click="closePickupDialog">
      <view class="dialog-card" @click.stop>
        <view class="dialog-header">
          <view class="dialog-title">{{ pickupEditingId ? '编辑自提点' : '新增自提点' }}</view>
          <view class="dialog-close" @click="closePickupDialog">×</view>
        </view>

        <view class="form-item">
          <view class="form-label">自提点名称</view>
          <input v-model="pickupForm.name" class="form-input" placeholder="例如：XX店自提点" />
        </view>

        <view class="form-item">
          <view class="form-label">位置</view>
          <view class="location-row" @click="choosePickupLocation">
            <view class="location-text">
              {{ pickupForm.address ? pickupForm.address : '点击地图选点' }}
            </view>
            <view class="location-btn">选点</view>
          </view>
        </view>

        <view class="switch-row dialog-switch">
          <view class="switch-label">
            <text class="label-title">设为默认</text>
            <text class="label-desc">用户选择自提时默认选中该点</text>
          </view>
          <switch :checked="pickupForm.is_default" @change="onPickupDefaultChange" color="#007AFF" />
        </view>

        <view class="switch-row dialog-switch">
          <view class="switch-label">
            <text class="label-title">启用</text>
            <text class="label-desc">停用后用户无法选择该点</text>
          </view>
          <switch :checked="pickupForm.status === 1" @change="onPickupStatusChange" color="#007AFF" />
        </view>

        <view class="dialog-footer">
          <button class="dialog-btn cancel" @click="closePickupDialog">取消</button>
          <button class="dialog-btn primary" :disabled="pickupSubmitting" @click="submitPickupPoint">
            {{ pickupSubmitting ? '保存中...' : '保存' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { createPickupPoint, deletePickupPoint, getDeliverySettings, getPickupPoints, updateDeliverySettings, updateMerchantSettings, updatePickupPoint } from '@api'
import type { DeliverySettings, DistanceRule, MerchantDeliverySettings, PickupPoint } from '@types'

const saving = ref(false)
const modeForm = reactive({
  takeout_enabled: false,
  dine_in_enabled: false,
  pickup_enabled: false
})

const formData = reactive<DeliverySettings>({
  enabled: false,
  base_fee: 0,
  free_delivery_amount: 0,
  max_distance: 10,
  distance_rules: []
})

const pickupPoints = ref<PickupPoint[]>([])
const pickupDialogVisible = ref(false)
const pickupSubmitting = ref(false)
const pickupEditingId = ref<number | null>(null)
const pickupForm = reactive({
  name: '',
  address: '',
  lat: 0,
  lng: 0,
  is_default: false,
  status: 1,
  sort: 0
})

function fillModeSettings(settings: Pick<MerchantDeliverySettings, 'takeout_enabled' | 'dine_in_enabled' | 'pickup_enabled'>) {
  modeForm.takeout_enabled = !!settings.takeout_enabled
  modeForm.dine_in_enabled = !!settings.dine_in_enabled
  modeForm.pickup_enabled = !!settings.pickup_enabled
}

function fillFormData(settings: DeliverySettings) {
  formData.enabled = !!settings.enabled
  formData.base_fee = Number(settings.base_fee || 0)
  formData.free_delivery_amount = Number(settings.free_delivery_amount || 0)
  formData.max_distance = Number(settings.max_distance || 10)
  formData.distance_rules.splice(
    0,
    formData.distance_rules.length,
    ...(settings.distance_rules || []).map((rule) => ({
      min_distance: Number(rule.min_distance || 0),
      max_distance: Number(rule.max_distance || 0),
      fee: Number(rule.fee || 0)
    }))
  )
}

onShow(() => {
  loadSettings()
  loadPickupPoints()
})

async function loadSettings() {
  try {
    const deliverySettings = await getDeliverySettings()
    fillModeSettings(deliverySettings)
    fillFormData(deliverySettings)
    formData.enabled = !!deliverySettings.takeout_enabled
  } catch (error) {
    console.error('加载配送设置失败:', error)
  }
}

async function loadPickupPoints() {
  try {
    pickupPoints.value = await getPickupPoints()
  } catch (error) {
    console.error('加载自提点失败:', error)
    pickupPoints.value = []
  }
}

function onDeliverySwitchChange(e: any) {
  modeForm.takeout_enabled = !!e.detail.value
  formData.enabled = modeForm.takeout_enabled
}

function onDineInSwitchChange(e: any) {
  modeForm.dine_in_enabled = !!e.detail.value
}

function onPickupSwitchChange(e: any) {
  modeForm.pickup_enabled = !!e.detail.value
}

function openPickupDialog(point?: PickupPoint) {
  pickupEditingId.value = point ? point.id : null
  pickupForm.name = point?.name || ''
  pickupForm.address = point?.address || ''
  pickupForm.lat = point?.lat || 0
  pickupForm.lng = point?.lng || 0
  pickupForm.is_default = !!point?.is_default
  pickupForm.status = point?.status ?? 1
  pickupForm.sort = point?.sort ?? 0
  pickupDialogVisible.value = true
}

function closePickupDialog() {
  if (pickupSubmitting.value) return
  pickupDialogVisible.value = false
}

function editPickupPoint(point: PickupPoint) {
  openPickupDialog(point)
}

function onPickupDefaultChange(e: any) {
  pickupForm.is_default = !!e.detail.value
}

function onPickupStatusChange(e: any) {
  pickupForm.status = e.detail.value ? 1 : 0
}

async function choosePickupLocation() {
  try {
    const result = await uni.chooseLocation()
    pickupForm.address = result.address || ''
    pickupForm.lat = Number(result.latitude || 0)
    pickupForm.lng = Number(result.longitude || 0)
    if (!pickupForm.name) {
      pickupForm.name = result.name || ''
    }
  } catch (error: any) {
    const errMsg = String(error?.errMsg || '')
    if (!errMsg || errMsg.includes('cancel')) {
      return
    }

    if (
      errMsg.includes('auth deny') ||
      errMsg.includes('auth denied') ||
      errMsg.includes('authorize no response') ||
      errMsg.includes('system permission denied')
    ) {
      uni.showModal({
        title: '无法打开地图',
        content: '请在微信或系统设置中开启位置权限后重试。',
        confirmText: '去设置',
        success: (res) => {
          if (!res.confirm) return
          uni.openSetting()
        }
      })
      return
    }

    uni.showToast({
      title: '地图选点打开失败，请重试',
      icon: 'none'
    })
  }
}

function openPickupLocation(point: PickupPoint) {
  if (!point.lat || !point.lng) return
  uni.openLocation({
    latitude: point.lat,
    longitude: point.lng,
    name: point.name,
    address: point.address
  })
}

async function submitPickupPoint() {
  if (!pickupForm.name.trim()) {
    return uni.showToast({ title: '请输入自提点名称', icon: 'none' })
  }
  if (!pickupForm.address || !pickupForm.lat || !pickupForm.lng) {
    return uni.showToast({ title: '请选择自提点位置', icon: 'none' })
  }

  pickupSubmitting.value = true
  const payload = {
    name: pickupForm.name.trim(),
    address: pickupForm.address,
    lat: pickupForm.lat,
    lng: pickupForm.lng,
    is_default: pickupForm.is_default,
    status: pickupForm.status,
    sort: pickupForm.sort
  }

  try {
    if (pickupEditingId.value) {
      await updatePickupPoint(pickupEditingId.value, payload)
    } else {
      await createPickupPoint(payload)
    }
    await loadPickupPoints()
    pickupDialogVisible.value = false
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    pickupSubmitting.value = false
  }
}

async function setDefaultPickupPoint(point: PickupPoint) {
  try {
    await updatePickupPoint(point.id, {
      name: point.name,
      address: point.address,
      lat: point.lat,
      lng: point.lng,
      status: point.status,
      sort: point.sort,
      is_default: true
    })
    await loadPickupPoints()
  } catch (error: any) {
    uni.showToast({ title: error.message || '设置默认失败', icon: 'none' })
  }
}

async function togglePickupPointStatus(point: PickupPoint) {
  try {
    await updatePickupPoint(point.id, {
      name: point.name,
      address: point.address,
      lat: point.lat,
      lng: point.lng,
      status: point.status === 1 ? 0 : 1,
      sort: point.sort,
      is_default: point.is_default
    })
    await loadPickupPoints()
  } catch (error: any) {
    uni.showToast({ title: error.message || '更新失败', icon: 'none' })
  }
}

async function removePickupPoint(point: PickupPoint) {
  uni.showModal({
    title: '删除自提点',
    content: `确认删除「${point.name}」？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deletePickupPoint(point.id)
        await loadPickupPoints()
        uni.showToast({ title: '删除成功', icon: 'success' })
      } catch (error: any) {
        uni.showToast({ title: error.message || '删除失败', icon: 'none' })
      }
    }
  })
}

function addRule() {
  formData.distance_rules.push({
    min_distance: 0,
    max_distance: 5,
    fee: 0
  })
}

function deleteRule(index: number) {
  formData.distance_rules.splice(index, 1)
}

function normalizeDistanceRules(rules: DistanceRule[]) {
  return rules
    .map((rule) => ({
      min_distance: Number(rule.min_distance || 0),
      max_distance: Number(rule.max_distance || 0),
      fee: Number(rule.fee || 0)
    }))
    .sort((prev, next) => prev.min_distance - next.min_distance)
}

function validateFormData() {
  if (!formData.enabled) {
    return ''
  }

  if (formData.base_fee < 0) {
    return '基础配送费不能小于 0'
  }
  if (formData.free_delivery_amount < 0) {
    return '满额免配送费不能小于 0'
  }
  if (formData.max_distance <= 0) {
    return '最大配送距离必须大于 0'
  }

  const normalizedRules = normalizeDistanceRules(formData.distance_rules)
  for (let index = 0; index < normalizedRules.length; index += 1) {
    const rule = normalizedRules[index]
    if (rule.min_distance < 0) {
      return `第 ${index + 1} 条规则的起始距离不能小于 0`
    }
    if (rule.max_distance <= rule.min_distance) {
      return `第 ${index + 1} 条规则的结束距离必须大于起始距离`
    }
    if (rule.fee < 0) {
      return `第 ${index + 1} 条规则的配送费不能小于 0`
    }
    if (rule.max_distance > formData.max_distance) {
      return `第 ${index + 1} 条规则超出最大配送距离`
    }
    if (index > 0 && rule.min_distance < normalizedRules[index - 1].max_distance) {
      return `第 ${index + 1} 条规则与前一条规则区间重叠`
    }
  }

  formData.distance_rules.splice(0, formData.distance_rules.length, ...normalizedRules)
  return ''
}

async function handleSave() {
  const validationMessage = validateFormData()
  if (validationMessage) {
    uni.showToast({ title: validationMessage, icon: 'none' })
    return
  }

  saving.value = true

  try {
    await updateMerchantSettings({
      takeout_enabled: modeForm.takeout_enabled,
      dine_in_enabled: modeForm.dine_in_enabled,
      pickup_enabled: modeForm.pickup_enabled
    })

    try {
      const latestSettings = await updateDeliverySettings({
        ...formData,
        enabled: modeForm.takeout_enabled
      })
      fillModeSettings(latestSettings)
      fillFormData(latestSettings)
      formData.enabled = latestSettings.takeout_enabled
      uni.showToast({ title: '保存成功', icon: 'success' })
    } catch (error: any) {
      await loadSettings()
      uni.showToast({ title: error.message || '下单方式已保存，配送规则保存失败', icon: 'none' })
      return
    }
    
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error: any) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.delivery-settings-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 200rpx;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.switch-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mode-row {
  margin-top: 28rpx;
  padding-top: 28rpx;
  border-top: 1rpx solid #f0f0f0;
}

.switch-label {
  flex: 1;
}

.label-title {
  font-size: 32rpx;
  color: #1a1a1a;
  font-weight: 500;
  display: block;
  margin-bottom: 8rpx;
}

.label-desc {
  font-size: 26rpx;
  color: #999999;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.add-rule-btn {
  font-size: 28rpx;
  color: #007AFF;
}

.form-item {
  margin-bottom: 28rpx;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.input-row {
  display: flex;
  align-items: center;
}

.form-input {
  flex: 1;
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
}

.input-suffix {
  font-size: 28rpx;
  color: #666666;
  margin-left: 16rpx;
}

.form-hint {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.rule-item {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.rule-title {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 500;
}

.delete-rule {
  font-size: 26rpx;
  color: #ff4d4f;
}

.rule-content {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.rule-input {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.rule-input .input {
  width: 100rpx;
  height: 64rpx;
  background: #ffffff;
  border-radius: 8rpx;
  text-align: center;
  font-size: 28rpx;
}

.rule-unit {
  font-size: 26rpx;
  color: #666666;
}

.rule-text {
  font-size: 26rpx;
  color: #999999;
  margin: 0 8rpx;
}

.rule-fee {
  display: flex;
  align-items: center;
}

.fee-input {
  width: 120rpx !important;
}

.fee-unit {
  font-size: 26rpx;
  color: #666666;
  margin-left: 8rpx;
}

.empty-rules {
  text-align: center;
  padding: 48rpx 0;
  font-size: 28rpx;
  color: #999999;
}

.pickup-item {
  padding: 24rpx 0;
  border-top: 1rpx solid #f0f0f0;
}

.pickup-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16rpx;
}

.pickup-title {
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex: 1;
  min-width: 0;
}

.pickup-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pickup-tag {
  font-size: 22rpx;
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: rgba(0, 122, 255, 0.12);
  color: #007AFF;
}

.pickup-tag.disabled {
  background: rgba(153, 153, 153, 0.15);
  color: #999999;
}

.pickup-actions {
  display: flex;
  gap: 16rpx;
  flex-shrink: 0;
}

.pickup-action {
  font-size: 26rpx;
  color: #007AFF;
}

.pickup-action.danger {
  color: #ff4d4f;
}

.pickup-address {
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #666666;
  line-height: 1.6;
}

.pickup-bottom {
  margin-top: 12rpx;
  display: flex;
  gap: 24rpx;
}

.pickup-link {
  font-size: 26rpx;
  color: #007AFF;
}

.dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 999;
}

.dialog-card {
  width: 88%;
  background: #ffffff;
  border-radius: 16rpx;
  padding: 28rpx;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.dialog-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.dialog-close {
  font-size: 44rpx;
  color: #999999;
  line-height: 1;
  padding: 8rpx;
}

.location-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 18rpx 20rpx;
  gap: 16rpx;
}

.location-text {
  flex: 1;
  min-width: 0;
  font-size: 26rpx;
  color: #333333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.location-btn {
  flex-shrink: 0;
  font-size: 26rpx;
  color: #007AFF;
}

.dialog-switch {
  margin-top: 18rpx;
}

.dialog-footer {
  display: flex;
  gap: 16rpx;
  margin-top: 28rpx;
}

.dialog-btn {
  flex: 1;
  height: 80rpx;
  border-radius: 12rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-btn.cancel {
  background: #f5f5f5;
  color: #333333;
}

.dialog-btn.primary {
  background: #007AFF;
  color: #ffffff;
}

.dialog-btn[disabled] {
  background: #cccccc;
}

.save-area {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-save {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-save[disabled] {
  background: #cccccc;
}
</style>
