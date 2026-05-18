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

    <!-- 保存按钮 -->
    <view class="save-area">
      <button class="btn-save" :disabled="saving" @click="handleSave">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getDeliverySettings, updateDeliverySettings, updateMerchantSettings } from '@api'
import type { DeliverySettings, DistanceRule, MerchantDeliverySettings } from '@types'

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
