<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  changePassword,
  getMerchantProfile,
  getMerchantQRCode,
  updateMerchantProfile,
  updatePaymentConfig
} from '@/api/sp'
import type { MerchantDetail } from '@/types/sp'
import { formatDateTime, getMerchantStatusText, getPaymentConfigText } from '@/utils/format'
import { uploadSpImage } from '@/utils/qiniu'

const loading = ref(false)
const savingProfile = ref(false)
const savingPayment = ref(false)
const merchant = ref<MerchantDetail | null>(null)

const logoInputRef = ref<HTMLInputElement | null>(null)
const coverInputRef = ref<HTMLInputElement | null>(null)
const assetUploading = ref<'logo' | 'cover_image' | ''>('')

const profileForm = reactive({
  name: '',
  contact_name: '',
  contact_phone: '',
  contact_email: '',
  address: '',
  business_hours: '',
  announcement: ''
})

const paymentForm = reactive({
  sub_mch_id: ''
})

const passwordDialogVisible = ref(false)
const passwordForm = reactive({ old_password: '', new_password: '', confirm_password: '' })
const changingPassword = ref(false)

const qrcodeDialogVisible = ref(false)
const qrcodeUrl = ref('')

async function loadProfile() {
  loading.value = true
  try {
    const result = await getMerchantProfile()
    merchant.value = result.merchant
    syncProfileForm(result.merchant)
    syncPaymentForm(result.merchant)
  } finally {
    loading.value = false
  }
}

function syncProfileForm(m: MerchantDetail) {
  profileForm.name = m.name || ''
  profileForm.contact_name = m.contact_name || ''
  profileForm.contact_phone = m.contact_phone || ''
  profileForm.contact_email = m.contact_email || ''
  profileForm.address = m.address || ''
  profileForm.business_hours = m.business_hours || ''
  profileForm.announcement = m.announcement || ''
}

function syncPaymentForm(m: MerchantDetail) {
  paymentForm.sub_mch_id = m.sub_mch_id || ''
}

async function submitProfile() {
  if (savingProfile.value) return
  savingProfile.value = true
  try {
    const updated = await updateMerchantProfile({ ...profileForm })
    merchant.value = updated
    ElMessage.success('商家资料已更新')
  } finally {
    savingProfile.value = false
  }
}

async function submitPayment() {
  if (savingPayment.value) return
  if (!paymentForm.sub_mch_id.trim()) {
    ElMessage.warning('子商户号不能为空')
    return
  }

  savingPayment.value = true
  try {
    const updated = await updatePaymentConfig({
      sub_mch_id: paymentForm.sub_mch_id.trim()
    })
    merchant.value = updated
    syncPaymentForm(updated)
    ElMessage.success('支付配置已更新')
  } finally {
    savingPayment.value = false
  }
}

function triggerAssetInput(field: 'logo' | 'cover_image') {
  if (assetUploading.value) return
  const inputRef = field === 'logo' ? logoInputRef.value : coverInputRef.value
  inputRef?.click()
}

async function handleAssetChange(field: 'logo' | 'cover_image', event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (!file) return

  assetUploading.value = field
  try {
    const uploaded = await uploadSpImage(file)
    const updated = await updateMerchantProfile({ [field]: uploaded.url })
    merchant.value = updated
    ElMessage.success('图片更新成功')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '图片更新失败')
  } finally {
    assetUploading.value = ''
  }
}

function openPasswordDialog() {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  passwordDialogVisible.value = true
}

async function submitPassword() {
  if (changingPassword.value) return
  if (!passwordForm.old_password || !passwordForm.new_password) {
    ElMessage.warning('请填写原密码与新密码')
    return
  }
  if (passwordForm.new_password.length < 6) {
    ElMessage.warning('新密码至少 6 位')
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }

  changingPassword.value = true
  try {
    await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
  } finally {
    changingPassword.value = false
  }
}

async function openQrcode() {
  try {
    const result = await getMerchantQRCode()
    qrcodeUrl.value = result.qrcode_url || ''
    qrcodeDialogVisible.value = true
  } catch (error) {
    ElMessage.error('生成二维码失败')
  }
}

onMounted(loadProfile)
</script>

<template>
  <div class="page-shell" v-loading="loading">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">商家资料</h1>
        <p class="page-subtitle">维护商家基础信息与支付配置，并管理登录密码。</p>
      </div>
      <el-space>
        <el-button @click="openPasswordDialog">修改密码</el-button>
        <el-button type="primary" plain @click="openQrcode">商家二维码</el-button>
      </el-space>
    </div>

    <template v-if="merchant">
      <div class="metric-grid">
        <div class="metric-card">
          <div class="metric-label">商家状态</div>
          <div class="metric-value" style="font-size: 22px;">{{ getMerchantStatusText(merchant.status) }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">支付配置</div>
          <div class="metric-value" style="font-size: 22px;">{{ getPaymentConfigText(merchant.payment_config_status) }}</div>
        </div>
      </div>

      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>商家资料</template>
          <el-form label-width="100px">
            <el-form-item label="商家名称" required>
              <el-input v-model="profileForm.name" maxlength="50" placeholder="请输入商家名称" />
            </el-form-item>
            <el-form-item label="联系人">
              <el-input v-model="profileForm.contact_name" placeholder="联系人姓名" />
            </el-form-item>
            <el-form-item label="联系电话">
              <el-input v-model="profileForm.contact_phone" placeholder="联系电话" />
            </el-form-item>
            <el-form-item label="联系邮箱">
              <el-input v-model="profileForm.contact_email" placeholder="联系邮箱" />
            </el-form-item>
            <el-form-item label="商家地址">
              <el-input v-model="profileForm.address" placeholder="商家地址" />
            </el-form-item>
            <el-form-item label="营业时间">
              <el-input v-model="profileForm.business_hours" placeholder="如：09:00-18:00" />
            </el-form-item>
            <el-form-item label="商家公告">
              <el-input v-model="profileForm.announcement" type="textarea" :rows="3" maxlength="200" show-word-limit placeholder="商家公告" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingProfile" @click="submitProfile">保存资料</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <div class="right-column">
          <el-card class="page-card" shadow="never">
            <template #header>图片资产</template>
            <div class="asset-grid">
              <div class="asset-box">
                <div class="asset-label">商家 Logo</div>
                <img v-if="merchant.logo" :src="merchant.logo" class="asset-preview" alt="Logo" />
                <div v-else class="empty-asset">未上传</div>
                <input ref="logoInputRef" hidden type="file" accept="image/*" @change="handleAssetChange('logo', $event)" />
                <el-button type="primary" plain :loading="assetUploading === 'logo'" @click="triggerAssetInput('logo')">更换 Logo</el-button>
              </div>
              <div class="asset-box">
                <div class="asset-label">背景图</div>
                <img v-if="merchant.cover_image" :src="merchant.cover_image" class="asset-preview" alt="背景图" />
                <div v-else class="empty-asset">未上传</div>
                <input ref="coverInputRef" hidden type="file" accept="image/*" @change="handleAssetChange('cover_image', $event)" />
                <el-button type="primary" plain :loading="assetUploading === 'cover_image'" @click="triggerAssetInput('cover_image')">更换背景图</el-button>
              </div>
            </div>
            <div class="profile-meta">
              <div>创建时间：{{ formatDateTime(merchant.created_at) }}</div>
            </div>
          </el-card>

          <el-card class="page-card" shadow="never">
            <template #header>支付配置</template>
            <el-form label-width="100px">
              <el-form-item label="子商户号" required>
                <el-input v-model="paymentForm.sub_mch_id" placeholder="微信支付子商户号" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="savingPayment" @click="submitPayment">保存配置</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </div>
      </div>
    </template>

    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="440px">
      <el-form label-width="90px">
        <el-form-item label="原密码" required>
          <el-input v-model="passwordForm.old_password" show-password type="password" placeholder="请输入原密码" />
        </el-form-item>
        <el-form-item label="新密码" required>
          <el-input v-model="passwordForm.new_password" show-password type="password" placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="确认密码" required>
          <el-input v-model="passwordForm.confirm_password" show-password type="password" placeholder="再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="changingPassword" @click="submitPassword">确认修改</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="qrcodeDialogVisible" title="商家二维码" width="420px">
      <div class="qrcode-wrap">
        <img v-if="qrcodeUrl" :src="qrcodeUrl" alt="商家二维码" class="qrcode-img" />
        <el-empty v-else description="暂无二维码" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.right-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.asset-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.asset-label {
  font-weight: 600;
  color: #374151;
}

.asset-preview {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.empty-asset {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
  color: #9ca3af;
  font-size: 13px;
}

.profile-meta {
  margin-top: 16px;
  color: #6b7280;
  font-size: 13px;
}

.qrcode-wrap {
  text-align: center;
}

.qrcode-img {
  max-width: 260px;
  width: 100%;
  border-radius: 12px;
}
</style>
