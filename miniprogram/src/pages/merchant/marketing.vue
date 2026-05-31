<template>
  <view class="marketing-container">
    <view class="hero-card">
      <view class="hero-title">满减营销</view>
      <view class="hero-desc">支持最多 5 档规则，下单时按最优优惠自动生效。</view>
    </view>

    <view class="section">
      <view class="section-header">
        <view class="section-title">规则配置</view>
        <view class="add-btn" :class="{ disabled: rules.length >= 5 }" @click="addRule">
          新增规则
        </view>
      </view>
      <view class="section-tip">启用后用户下单时会自动命中当前可用的最高优惠。</view>

      <view v-if="rules.length === 0" class="empty-card">
        <view class="empty-title">暂未配置满减规则</view>
        <view class="empty-desc">点击右上角新增，支持配置满多少减多少。</view>
      </view>

      <view v-for="(rule, index) in rules" :key="rule.local_id" class="rule-card">
        <view class="rule-header">
          <view class="rule-title">规则 {{ index + 1 }}</view>
          <view class="rule-actions">
            <text class="rule-status">{{ rule.status ? '启用中' : '已停用' }}</text>
            <text class="rule-delete" @click="removeRule(index)">删除</text>
          </view>
        </view>

        <view class="rule-row">
          <view class="field-item">
            <view class="field-label">满减门槛</view>
            <view class="input-row">
              <input
                v-model="rule.threshold_amount_text"
                type="digit"
                class="field-input"
                placeholder="例如 50"
              />
              <text class="input-suffix">元</text>
            </view>
          </view>
          <view class="field-item">
            <view class="field-label">减免金额</view>
            <view class="input-row">
              <input
                v-model="rule.discount_amount_text"
                type="digit"
                class="field-input"
                placeholder="例如 5"
              />
              <text class="input-suffix">元</text>
            </view>
          </view>
        </view>

        <view class="switch-row">
          <view class="switch-label">
            <text class="switch-title">启用此规则</text>
            <text class="switch-desc">停用后保留配置，但本档优惠不参与下单计算</text>
          </view>
          <switch :checked="rule.status === 1" color="#007AFF" @change="(e:any) => toggleRuleStatus(index, e)" />
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section-title">当前生效预览</view>
      <view v-if="enabledRules.length > 0" class="preview-list">
        <view v-for="item in enabledRules" :key="item.local_id" class="preview-item">
          满 {{ item.threshold_amount_text || '0' }} 元减 {{ item.discount_amount_text || '0' }} 元
        </view>
      </view>
      <view v-else class="preview-empty">当前没有启用中的满减规则</view>
    </view>

    <view class="save-area">
      <button class="save-btn" :disabled="saving" @click="handleSave">
        {{ saving ? '保存中...' : '保存规则' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getMerchantFullReductionRules, updateMerchantFullReductionRules } from '@api'
import type { MerchantFullReductionRule } from '@types'

interface EditableRule extends MerchantFullReductionRule {
  local_id: string
  threshold_amount_text: string
  discount_amount_text: string
}

const rules = ref<EditableRule[]>([])
const saving = ref(false)

const enabledRules = computed(() => {
  return rules.value
    .filter((item) => item.status === 1)
    .sort((prev, next) => Number(prev.threshold_amount_text || 0) - Number(next.threshold_amount_text || 0))
})

onShow(() => {
  loadRules()
})

function createLocalRule(rule?: Partial<MerchantFullReductionRule>): EditableRule {
  return {
    ...rule,
    status: Number(rule?.status ?? 1),
    local_id: `${Date.now()}-${Math.random()}`,
    threshold_amount_text: rule?.threshold_amount ? String(rule.threshold_amount) : '',
    discount_amount_text: rule?.discount_amount ? String(rule.discount_amount) : ''
  }
}

async function loadRules() {
  try {
    const result = await getMerchantFullReductionRules()
    rules.value = (result.rules || []).map((rule) => createLocalRule(rule))
  } catch (error) {
    console.error('加载满减规则失败:', error)
  }
}

function addRule() {
  if (rules.value.length >= 5) {
    uni.showToast({ title: '最多支持 5 档满减规则', icon: 'none' })
    return
  }
  rules.value.push(createLocalRule())
}

function removeRule(index: number) {
  rules.value.splice(index, 1)
}

function toggleRuleStatus(index: number, event: any) {
  rules.value[index].status = event?.detail?.value ? 1 : 0
}

function buildPayload() {
  const thresholdSet = new Set<number>()

  return rules.value.map((rule) => {
    const thresholdAmount = Number(rule.threshold_amount_text)
    const discountAmount = Number(rule.discount_amount_text)

    if (!Number.isFinite(thresholdAmount) || thresholdAmount <= 0) {
      throw new Error('满减门槛必须大于 0')
    }
    if (!Number.isFinite(discountAmount) || discountAmount <= 0) {
      throw new Error('减免金额必须大于 0')
    }
    if (discountAmount >= thresholdAmount) {
      throw new Error('减免金额必须小于满减门槛')
    }
    if (thresholdSet.has(thresholdAmount)) {
      throw new Error('满减门槛不能重复')
    }

    thresholdSet.add(thresholdAmount)

    return {
      threshold_amount: Number(thresholdAmount.toFixed(2)),
      discount_amount: Number(discountAmount.toFixed(2)),
      status: rule.status === 1 ? 1 : 0
    }
  })
}

async function handleSave() {
  if (saving.value) {
    return
  }

  try {
    saving.value = true
    const payload = buildPayload()
    await updateMerchantFullReductionRules({ rules: payload })
    uni.showToast({ title: '保存成功', icon: 'success' })
    await loadRules()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.marketing-container {
  min-height: 100vh;
  background: #f5f7fb;
  padding: 24rpx 24rpx 180rpx;
}

.hero-card {
  background: linear-gradient(135deg, #0a60ff 0%, #4b95ff 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
  margin-bottom: 24rpx;
}

.hero-title {
  font-size: 36rpx;
  font-weight: 600;
}

.hero-desc {
  margin-top: 12rpx;
  font-size: 26rpx;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.84);
}

.section {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.section-tip {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #8a9099;
  line-height: 1.5;
}

.add-btn {
  padding: 12rpx 22rpx;
  border-radius: 999rpx;
  background: #eef4ff;
  color: #007AFF;
  font-size: 24rpx;
}

.add-btn.disabled {
  opacity: 0.5;
}

.empty-card,
.preview-empty {
  margin-top: 24rpx;
  padding: 32rpx 24rpx;
  border-radius: 18rpx;
  background: #f8f9fb;
  text-align: center;
}

.empty-title {
  font-size: 28rpx;
  color: #1a1a1a;
  font-weight: 600;
}

.empty-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.rule-card {
  margin-top: 24rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f8fbff;
  border: 1rpx solid #e6efff;
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.rule-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.rule-actions {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.rule-status {
  font-size: 24rpx;
  color: #007AFF;
}

.rule-delete {
  font-size: 24rpx;
  color: #ff4d4f;
}

.rule-row {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}

.field-item {
  flex: 1;
  min-width: 0;
}

.field-label {
  margin-bottom: 12rpx;
  font-size: 24rpx;
  color: #666666;
}

.input-row {
  display: flex;
  align-items: center;
  height: 84rpx;
  border-radius: 16rpx;
  background: #ffffff;
  padding: 0 20rpx;
}

.field-input {
  flex: 1;
  min-width: 0;
  font-size: 30rpx;
  color: #1a1a1a;
}

.input-suffix {
  font-size: 24rpx;
  color: #8a9099;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 1rpx solid #e8edf5;
}

.switch-label {
  flex: 1;
  min-width: 0;
}

.switch-title {
  display: block;
  font-size: 28rpx;
  color: #1a1a1a;
}

.switch-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.preview-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 20rpx;
}

.preview-item {
  padding: 20rpx 24rpx;
  border-radius: 16rpx;
  background: #fff7e8;
  color: #d46b08;
  font-size: 26rpx;
}

.save-area {
  margin-top: 40rpx;
}

.save-btn {
  width: 100%;
  height: 96rpx;
  border-radius: 48rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
