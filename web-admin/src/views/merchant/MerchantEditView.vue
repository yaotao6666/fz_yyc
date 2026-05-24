<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  createSpMerchant,
  getMerchantDetail,
  resetSpMerchantAdminPassword,
  updateSpMerchant,
  updateSpMerchantPaymentConfig,
} from '@/api/sp'
import type {
  MerchantDetail,
  MerchantPaymentConfigFormData,
  SpMerchantFormData,
  UpdateSpMerchantFormData,
} from '@/types/sp'
import { getPaymentConfigText } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const savingBasic = ref(false)
const savingPayment = ref(false)
const savingAdminPassword = ref(false)

const merchantId = computed(() => Number(route.params.id || 0))
const isEditMode = computed(() => merchantId.value > 0)

const basicForm = reactive<SpMerchantFormData>({
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
  profit_sharing_ratio: 0,
})

const paymentForm = reactive<MerchantPaymentConfigFormData>({
  sub_mch_id: '',
  profit_sharing_enabled: false,
  profit_sharing_ratio: 0,
})

const adminForm = reactive({
  new_password: '',
})

const paymentConfigText = computed(() => getPaymentConfigText(
  paymentForm.sub_mch_id && (!paymentForm.profit_sharing_enabled || paymentForm.profit_sharing_ratio > 0) ? 1 : 0
))

function fillForm(detail: MerchantDetail) {
  basicForm.name = detail.name || ''
  basicForm.contact_name = detail.contact_name || ''
  basicForm.contact_phone = detail.contact_phone || ''
  basicForm.contact_email = detail.contact_email || ''
  basicForm.address = detail.address || ''
  basicForm.business_category = detail.business_category || ''
  basicForm.business_hours = detail.business_hours || ''
  basicForm.announcement = detail.announcement || ''
  basicForm.username = detail.admin_username || ''
  basicForm.staff_name = detail.admin_name || ''
  basicForm.staff_phone = detail.admin_phone || ''
  basicForm.password = ''
  paymentForm.sub_mch_id = detail.sub_mch_id || ''
  paymentForm.profit_sharing_enabled = Boolean(detail.profit_sharing_enabled)
  paymentForm.profit_sharing_ratio = Number(detail.profit_sharing_ratio || 0)
}

async function loadDetail() {
  if (!isEditMode.value) return
  loading.value = true
  try {
    const detail = await getMerchantDetail(merchantId.value)
    fillForm(detail)
  } finally {
    loading.value = false
  }
}

function validateBasic() {
  if (!basicForm.name.trim()) {
    throw new Error('请输入商家名称')
  }
  if (!isEditMode.value) {
    if (!basicForm.username.trim()) {
      throw new Error('请输入管理员账号')
    }
    if (basicForm.password.trim().length < 6) {
      throw new Error('管理员密码至少 6 位')
    }
  }
}

function validatePayment() {
  if (!paymentForm.sub_mch_id.trim()) {
    throw new Error('请输入收款子商户号')
  }
  if (paymentForm.profit_sharing_enabled && paymentForm.profit_sharing_ratio <= 0) {
    throw new Error('开启分账时请输入大于 0 的比例')
  }
}

function validateAdminPassword() {
  if (adminForm.new_password.trim().length < 6) {
    throw new Error('管理员新密码至少 6 位')
  }
}

async function submitBasic() {
  try {
    validateBasic()
  } catch (error: any) {
    ElMessage.warning(error.message)
    return
  }

  savingBasic.value = true
  try {
    if (isEditMode.value) {
      const payload: UpdateSpMerchantFormData = {
        name: basicForm.name.trim(),
        contact_name: basicForm.contact_name?.trim(),
        contact_phone: basicForm.contact_phone?.trim(),
        contact_email: basicForm.contact_email?.trim(),
        address: basicForm.address?.trim(),
        business_category: basicForm.business_category?.trim(),
        business_hours: basicForm.business_hours?.trim(),
        announcement: basicForm.announcement?.trim(),
      }
      await updateSpMerchant(merchantId.value, payload)
      ElMessage.success('基础信息已保存')
      return
    }

    const created = await createSpMerchant({
      ...basicForm,
      name: basicForm.name.trim(),
      contact_name: basicForm.contact_name?.trim(),
      contact_phone: basicForm.contact_phone?.trim(),
      contact_email: basicForm.contact_email?.trim(),
      address: basicForm.address?.trim(),
      business_category: basicForm.business_category?.trim(),
      business_hours: basicForm.business_hours?.trim(),
      announcement: basicForm.announcement?.trim(),
      username: basicForm.username.trim(),
      password: basicForm.password.trim(),
      staff_name: basicForm.staff_name?.trim(),
      staff_phone: basicForm.staff_phone?.trim(),
      sub_mch_id: paymentForm.sub_mch_id.trim(),
      profit_sharing_enabled: paymentForm.profit_sharing_enabled,
      profit_sharing_ratio: Number(paymentForm.profit_sharing_ratio || 0),
    })
    ElMessage.success('商家创建成功')
    await router.replace(`/merchants/${created.id}`)
  } finally {
    savingBasic.value = false
  }
}

async function submitPayment() {
  if (!isEditMode.value) return

  try {
    validatePayment()
  } catch (error: any) {
    ElMessage.warning(error.message)
    return
  }

  savingPayment.value = true
  try {
    const detail = await updateSpMerchantPaymentConfig(merchantId.value, {
      sub_mch_id: paymentForm.sub_mch_id.trim(),
      profit_sharing_enabled: paymentForm.profit_sharing_enabled,
      profit_sharing_ratio: Number(paymentForm.profit_sharing_ratio || 0),
    })
    fillForm(detail)
    ElMessage.success('支付配置已保存')
  } finally {
    savingPayment.value = false
  }
}

async function submitAdminPassword() {
  if (!isEditMode.value) return

  try {
    validateAdminPassword()
  } catch (error: any) {
    ElMessage.warning(error.message)
    return
  }

  savingAdminPassword.value = true
  try {
    const result = await resetSpMerchantAdminPassword(merchantId.value, {
      new_password: adminForm.new_password.trim(),
    })
    adminForm.new_password = ''
    ElMessage.success(`${result.username} 的登录密码已重置`)
  } finally {
    savingAdminPassword.value = false
  }
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">{{ isEditMode ? '编辑商家配置' : '新增商家' }}</h1>
        <p class="page-subtitle">服务商直接维护商家基础信息、管理员账号、收款账户与分账配置。</p>
      </div>
      <el-button @click="router.push(isEditMode ? `/merchants/${merchantId}` : '/merchants')">返回</el-button>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <div class="section-grid" style="align-items: start;">
        <el-card class="page-card" shadow="never">
          <template #header>
            <span>基础信息</span>
          </template>
          <el-form label-position="top">
            <el-form-item label="商家名称">
              <el-input v-model="basicForm.name" placeholder="请输入商家名称" />
            </el-form-item>
            <el-form-item label="联系人">
              <el-input v-model="basicForm.contact_name" placeholder="请输入联系人姓名" />
            </el-form-item>
            <el-form-item label="联系电话">
              <el-input v-model="basicForm.contact_phone" placeholder="请输入联系电话" />
            </el-form-item>
            <el-form-item label="联系邮箱">
              <el-input v-model="basicForm.contact_email" placeholder="请输入联系邮箱" />
            </el-form-item>
            <el-form-item label="经营分类">
              <el-input v-model="basicForm.business_category" placeholder="如：轻食简餐、茶饮甜品" />
            </el-form-item>
            <el-form-item label="营业时间">
              <el-input v-model="basicForm.business_hours" placeholder="如：09:00-21:00" />
            </el-form-item>
            <el-form-item label="商家地址">
              <el-input v-model="basicForm.address" type="textarea" :rows="3" placeholder="请输入商家地址" />
            </el-form-item>
            <el-form-item label="商家公告">
              <el-input v-model="basicForm.announcement" type="textarea" :rows="3" placeholder="请输入商家公告" />
            </el-form-item>
            <template v-if="!isEditMode">
              <el-divider>管理员账号</el-divider>
              <el-form-item label="登录账号">
                <el-input v-model="basicForm.username" placeholder="请输入登录账号" />
              </el-form-item>
              <el-form-item label="登录密码">
                <el-input v-model="basicForm.password" show-password type="password" placeholder="请输入 6 位以上密码" />
              </el-form-item>
              <el-form-item label="员工姓名">
                <el-input v-model="basicForm.staff_name" placeholder="默认同步联系人姓名" />
              </el-form-item>
              <el-form-item label="员工电话">
                <el-input v-model="basicForm.staff_phone" placeholder="默认同步联系人电话" />
              </el-form-item>
            </template>
            <template v-else>
              <el-divider>管理员账号</el-divider>
              <el-form-item label="登录账号">
                <el-input :model-value="basicForm.username || '未配置管理员账号'" disabled />
              </el-form-item>
              <el-form-item label="员工姓名">
                <el-input :model-value="basicForm.staff_name || '-'" disabled />
              </el-form-item>
              <el-form-item label="员工电话">
                <el-input :model-value="basicForm.staff_phone || '-'" disabled />
              </el-form-item>
              <el-form-item label="新登录密码">
                <el-input
                  v-model="adminForm.new_password"
                  show-password
                  type="password"
                  placeholder="请输入 6 位以上新密码"
                />
              </el-form-item>
              <el-alert
                title="重置后将直接覆盖当前商家管理员账号密码，仅影响负责人 owner 账号。"
                type="info"
                :closable="false"
                show-icon
              />
              <el-button
                type="warning"
                plain
                :loading="savingAdminPassword"
                :disabled="!basicForm.username"
                style="margin-top: 12px;"
                @click="submitAdminPassword"
              >
                重置管理员密码
              </el-button>
            </template>
            <el-button type="primary" :loading="savingBasic" style="margin-top: 12px;" @click="submitBasic">
              {{ isEditMode ? '保存基础信息' : '创建商家并生成管理员账号' }}
            </el-button>
          </el-form>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <span>支付与分账配置</span>
          </template>
          <el-form label-position="top">
            <el-form-item label="收款子商户号">
              <el-input v-model="paymentForm.sub_mch_id" placeholder="请输入微信支付子商户号" />
            </el-form-item>
            <el-form-item label="开启分账抽佣">
              <el-switch v-model="paymentForm.profit_sharing_enabled" />
            </el-form-item>
            <el-form-item label="分账比例（%）">
              <el-input-number v-model="paymentForm.profit_sharing_ratio" :min="0" :precision="2" :step="0.1" style="width: 100%;" />
            </el-form-item>
            <el-alert :title="`当前状态：${paymentConfigText}`" type="success" :closable="false" show-icon />
            <el-button type="primary" :loading="savingPayment" :disabled="!isEditMode" style="margin-top: 16px;" @click="submitPayment">
              {{ isEditMode ? '保存支付配置' : '请先创建商家后再保存支付配置' }}
            </el-button>
          </el-form>
        </el-card>
      </div>
    </el-skeleton>
  </div>
</template>
