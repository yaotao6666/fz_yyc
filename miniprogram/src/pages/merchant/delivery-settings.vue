<template>
  <view class="delivery-settings-container">
    <!-- 配送开关 -->
    <view class="section">
      <view class="switch-row">
        <view class="switch-label">
          <text class="label-title">开启配送服务</text>
          <text class="label-desc">开启后用户可选择配送方式下单</text>
        </view>
        <switch
          :checked="formData.enabled"
          @change="onSwitchChange"
          color="#007AFF"
        />
      </view>
    </view>

    <!-- 配送设置 -->
    <view class="section" v-if="formData.enabled">
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
      </view>
    </view>

    <!-- 距离规则 -->
    <view class="section" v-if="formData.enabled">
      <view class="section-header">
        <view class="section-title">按距离收费</view>
        <view class="add-rule-btn" @click="addRule">添加规则</view>
      </view>

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
import { getDeliverySettings, updateDeliverySettings } from '@api'
import type { DeliverySettings, DistanceRule } from '@types'

const saving = ref(false)

const formData = reactive<DeliverySettings>({
  enabled: false,
  base_fee: 0,
  free_delivery_amount: 0,
  max_distance: 10,
  distance_rules: []
})

onShow(() => {
  loadSettings()
})

async function loadSettings() {
  try {
    const settings = await getDeliverySettings()
    Object.assign(formData, settings)
  } catch (error) {
    console.error('加载配送设置失败:', error)
  }
}

function onSwitchChange(e: any) {
  formData.enabled = e.detail.value
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

async function handleSave() {
  saving.value = true

  try {
    await updateDeliverySettings(formData)
    uni.showToast({ title: '保存成功', icon: 'success' })
    
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
