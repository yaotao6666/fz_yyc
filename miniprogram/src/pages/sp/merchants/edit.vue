<template>
  <view class="merchant-edit-page">
    <scroll-view class="merchant-edit-container" scroll-y>
      <view class="merchant-edit-content">
        <view class="page-tip">
      <text class="page-tip-title">{{ isEditMode ? '编辑商家配置' : '新增商家' }}</text>
      <text class="page-tip-desc">服务商直接维护商家基础信息、管理员账号、收款商户号和分账比例。</text>
        </view>

        <view class="section-card">
      <view class="section-title">基础信息</view>
      <view class="form-item">
        <text class="form-label">商家名称</text>
        <input v-model="form.name" class="form-input" placeholder="请输入商家名称" />
      </view>
      <view class="form-item">
        <text class="form-label">联系人</text>
        <input v-model="form.contact_name" class="form-input" placeholder="请输入联系人姓名" />
      </view>
      <view class="form-item">
        <text class="form-label">联系电话</text>
        <input v-model="form.contact_phone" class="form-input" type="number" placeholder="请输入联系电话" />
      </view>
      <view class="form-item">
        <text class="form-label">联系邮箱</text>
        <input v-model="form.contact_email" class="form-input" placeholder="请输入联系邮箱" />
      </view>
      <view class="form-item">
        <text class="form-label">经营分类</text>
        <input v-model="form.business_category" class="form-input" placeholder="如：轻食简餐、茶饮甜品" />
      </view>
      <view class="form-item">
        <text class="form-label">营业时间</text>
        <input v-model="form.business_hours" class="form-input" placeholder="如：09:00-21:00" />
      </view>
      <view class="form-item">
        <text class="form-label">商家地址</text>
        <textarea v-model="form.address" class="form-textarea" placeholder="请输入商家地址" />
      </view>
      <view class="form-item">
        <text class="form-label">商家公告</text>
        <textarea v-model="form.announcement" class="form-textarea" placeholder="请输入商家公告" />
      </view>
      <button class="submit-btn" :loading="savingBasic" @click="submitBasicInfo">
        {{ isEditMode ? '保存基础信息' : '创建商家并生成管理员账号' }}
      </button>
    </view>

    <view class="section-card">
      <view class="section-title">管理员账号</view>
      <template v-if="!isEditMode">
        <view class="form-item">
          <text class="form-label">登录账号</text>
          <input v-model="form.username" class="form-input" placeholder="请输入登录账号" />
        </view>
        <view class="form-item">
          <text class="form-label">登录密码</text>
          <input v-model="form.password" class="form-input" password placeholder="请输入6位以上密码" />
        </view>
        <view class="form-item">
          <text class="form-label">员工姓名</text>
          <input v-model="form.staff_name" class="form-input" placeholder="默认同步联系人姓名" />
        </view>
        <view class="form-item">
          <text class="form-label">员工电话</text>
          <input v-model="form.staff_phone" class="form-input" type="number" placeholder="默认同步联系人电话" />
        </view>
      </template>
      <view v-else class="readonly-tip">
        商家管理员账号已创建完成，如需修改账号密码请到对应商家后台处理。
      </view>
    </view>

    <view class="section-card">
      <view class="section-title">支付与分账配置</view>
      <view class="form-item">
        <text class="form-label">收款子商户号</text>
        <input v-model="form.sub_mch_id" class="form-input" placeholder="请输入微信支付子商户号" />
      </view>
      <view class="switch-row">
        <view>
          <text class="form-label">开启分账抽佣</text>
          <text class="switch-desc">支付成功回调后按比例自动抽佣</text>
        </view>
        <switch :checked="form.profit_sharing_enabled" color="#1677ff" @change="onProfitSharingChange" />
      </view>
      <view class="form-item">
        <text class="form-label">分账比例（%）</text>
        <input
          v-model="profitSharingRatioInput"
          class="form-input"
          type="digit"
          placeholder="请输入分账比例"
        />
      </view>
      <view class="readonly-tip">
        当前状态：{{ paymentConfigStatusText }}
      </view>
      <button
        class="submit-btn secondary"
        :disabled="!isEditMode"
        :loading="savingPayment"
        @click="submitPaymentConfig"
      >
        {{ isEditMode ? '保存支付配置' : '请先创建商家后再保存支付配置' }}
      </button>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import {
  createSpMerchant,
  getMerchantDetail,
  updateSpMerchant,
  updateSpMerchantPaymentConfig
} from '@api'
import { PaymentConfigStatusText } from '@types'
import type {
  MerchantPaymentConfigFormData,
  MerchantDetail,
  SpMerchantFormData,
  UpdateSpMerchantFormData
} from '@types'

const merchantId = ref<number>(0)
const isEditMode = computed(() => merchantId.value > 0)
const savingBasic = ref(false)
const savingPayment = ref(false)
const profitSharingRatioInput = ref('0')

const form = reactive<SpMerchantFormData>({
  name: '',
  contact_name: '',
  contact_phone: '',
  contact_email: '',
  address: '',
  business_category: '',
  business_hours: '',
  announcement: '',
  username: '',
  password: '',
  staff_name: '',
  staff_phone: '',
  sub_mch_id: '',
  profit_sharing_enabled: false,
  profit_sharing_ratio: 0
})

const paymentConfigStatusText = computed(() => {
  const hasSubMchId = form.sub_mch_id.trim().length > 0
  if (!hasSubMchId) {
    return PaymentConfigStatusText[0]
  }
  if (form.profit_sharing_enabled && form.profit_sharing_ratio <= 0) {
    return PaymentConfigStatusText[0]
  }
  return PaymentConfigStatusText[1]
})

onLoad((options) => {
  const id = Number(options?.id || 0)
  if (!Number.isNaN(id) && id > 0) {
    merchantId.value = id
    loadMerchantDetail()
  }
})

async function loadMerchantDetail() {
  try {
    const detail = await getMerchantDetail(merchantId.value)
    fillForm(detail)
  } catch (requestError) {
    console.error('加载商家详情失败:', requestError)
    uni.showToast({ title: '加载商家详情失败', icon: 'none' })
  }
}

function fillForm(detail: MerchantDetail) {
  form.name = detail.name || ''
  form.contact_name = detail.contact_name || ''
  form.contact_phone = detail.contact_phone || ''
  form.contact_email = detail.contact_email || ''
  form.address = detail.address || ''
  form.business_category = detail.business_category || ''
  form.business_hours = detail.business_hours || ''
  form.announcement = detail.announcement || ''
  form.sub_mch_id = detail.sub_mch_id || ''
  form.profit_sharing_enabled = !!detail.profit_sharing_enabled
  form.profit_sharing_ratio = Number(detail.profit_sharing_ratio || 0)
  profitSharingRatioInput.value = String(form.profit_sharing_ratio || 0)
}

function normalizeRatio() {
  form.profit_sharing_ratio = Number(profitSharingRatioInput.value || 0)
}

function validateBasicInfo() {
  if (!form.name.trim()) {
    uni.showToast({ title: '请输入商家名称', icon: 'none' })
    return false
  }

  if (!isEditMode.value) {
    if (!form.username.trim()) {
      uni.showToast({ title: '请输入管理员账号', icon: 'none' })
      return false
    }
    if ((form.password || '').trim().length < 6) {
      uni.showToast({ title: '管理员密码至少6位', icon: 'none' })
      return false
    }
  }

  return true
}

function validatePaymentConfig() {
  normalizeRatio()
  if (!form.sub_mch_id.trim()) {
    uni.showToast({ title: '请输入收款子商户号', icon: 'none' })
    return false
  }
  if (form.profit_sharing_enabled && form.profit_sharing_ratio <= 0) {
    uni.showToast({ title: '开启分账时请输入大于0的比例', icon: 'none' })
    return false
  }
  return true
}

async function submitBasicInfo() {
  if (!validateBasicInfo()) {
    return
  }

  savingBasic.value = true
  try {
    if (isEditMode.value) {
      const payload: UpdateSpMerchantFormData = {
        name: form.name.trim(),
        contact_name: form.contact_name.trim(),
        contact_phone: form.contact_phone.trim(),
        contact_email: form.contact_email.trim(),
        address: form.address.trim(),
        business_category: form.business_category.trim(),
        business_hours: form.business_hours.trim(),
        announcement: form.announcement.trim()
      }
      await updateSpMerchant(merchantId.value, payload)
      uni.showToast({ title: '基础信息已保存', icon: 'success' })
      return
    }

    normalizeRatio()
    const created = await createSpMerchant({
      ...form,
      name: form.name.trim(),
      contact_name: form.contact_name.trim(),
      contact_phone: form.contact_phone.trim(),
      contact_email: form.contact_email.trim(),
      address: form.address.trim(),
      business_category: form.business_category.trim(),
      business_hours: form.business_hours.trim(),
      announcement: form.announcement.trim(),
      username: form.username.trim(),
      password: form.password.trim(),
      staff_name: form.staff_name.trim(),
      staff_phone: form.staff_phone.trim(),
      sub_mch_id: form.sub_mch_id.trim()
    })

    merchantId.value = created.id
    fillForm(created)
    uni.showToast({ title: '商家创建成功', icon: 'success' })
    setTimeout(() => {
      uni.redirectTo({ url: `/pages/sp/merchants/detail?id=${created.id}` })
    }, 400)
  } catch (requestError: any) {
    console.error('保存商家失败:', requestError)
    uni.showToast({ title: requestError?.message || '保存失败', icon: 'none' })
  } finally {
    savingBasic.value = false
  }
}

async function submitPaymentConfig() {
  if (!isEditMode.value) {
    return
  }
  if (!validatePaymentConfig()) {
    return
  }

  savingPayment.value = true
  try {
    const payload: MerchantPaymentConfigFormData = {
      sub_mch_id: form.sub_mch_id.trim(),
      profit_sharing_enabled: form.profit_sharing_enabled,
      profit_sharing_ratio: form.profit_sharing_ratio
    }
    const detail = await updateSpMerchantPaymentConfig(merchantId.value, payload)
    fillForm(detail)
    uni.showToast({ title: '支付配置已保存', icon: 'success' })
  } catch (requestError: any) {
    console.error('保存支付配置失败:', requestError)
    uni.showToast({ title: requestError?.message || '保存失败', icon: 'none' })
  } finally {
    savingPayment.value = false
  }
}

function onProfitSharingChange(event: any) {
  form.profit_sharing_enabled = !!event?.detail?.value
}
</script>

<style scoped>
.merchant-edit-page {
  min-height: 100vh;
  height: 100vh;
  background: #f5f5f5;
}

.merchant-edit-container {
  height: 100%;
}

.merchant-edit-content {
  min-height: 100%;
  background: #f5f5f5;
  padding: 24rpx;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}

.page-tip {
  padding: 32rpx;
  border-radius: 24rpx;
  background: linear-gradient(135deg, #3f7cff 0%, #635bff 100%);
  color: #ffffff;
  margin-bottom: 24rpx;
}

.page-tip-title {
  display: block;
  font-size: 34rpx;
  font-weight: 600;
}

.page-tip-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 25rpx;
  line-height: 1.6;
  opacity: 0.95;
}

.section-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 24rpx;
}

.form-item + .form-item {
  margin-top: 20rpx;
}

.form-label {
  display: block;
  font-size: 25rpx;
  color: #4e5969;
  margin-bottom: 12rpx;
}

.form-input,
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 22rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  font-size: 28rpx;
  color: #1f2329;
}

.form-textarea {
  min-height: 160rpx;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 20rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  margin-bottom: 20rpx;
}

.switch-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: #86909c;
}

.readonly-tip {
  padding: 22rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
  font-size: 24rpx;
  line-height: 1.7;
  color: #4e5969;
}

.submit-btn {
  margin-top: 24rpx;
  width: 100%;
  height: 84rpx;
  line-height: 84rpx;
  border: none;
  border-radius: 20rpx;
  background: #1677ff;
  color: #ffffff;
  font-size: 28rpx;
}

.submit-btn.secondary {
  background: #eef3ff;
  color: #1677ff;
}

.submit-btn[disabled] {
  opacity: 0.5;
}
</style>
