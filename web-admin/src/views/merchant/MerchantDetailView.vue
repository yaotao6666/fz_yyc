<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMerchantDetail, getMerchantQRCode, updateSpMerchantAssets } from '@/api/sp'
import type { MerchantDetail } from '@/types/sp'
import { formatAmount, formatDateTime, formatPercent, getMerchantStatusText, getPaymentConfigText } from '@/utils/format'
import { uploadSpImage } from '@/utils/qiniu'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const merchant = ref<MerchantDetail | null>(null)
const qrcodeVisible = ref(false)
const qrcodeUrl = ref('')
const qrcodePagePath = ref('')

const merchantId = computed(() => Number(route.params.id || 0))

async function loadDetail() {
  if (!merchantId.value) return
  loading.value = true
  try {
    merchant.value = await getMerchantDetail(merchantId.value)
  } finally {
    loading.value = false
  }
}

async function openQrcode() {
  const result = await getMerchantQRCode(merchantId.value)
  qrcodeUrl.value = result.qrcode_url
  qrcodePagePath.value = result.page_path
  qrcodeVisible.value = true
}

async function updateAsset(field: 'logo' | 'cover_image', event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (!file || !merchant.value) return

  const uploadResult = await uploadSpImage(file)
  merchant.value = await updateSpMerchantAssets(merchantId.value, { [field]: uploadResult.url })
  ElMessage.success('图片更新成功')
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">商家详情</h1>
        <p class="page-subtitle">查看商家资料、支付配置、经营数据和图片资产。</p>
      </div>
      <el-space>
        <el-button @click="router.push('/merchants')">返回列表</el-button>
        <el-button type="primary" @click="router.push(`/merchants/${merchantId}/edit`)">编辑商家</el-button>
      </el-space>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
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
          <div class="metric-card">
            <div class="metric-label">用户数</div>
            <div class="metric-value">{{ merchant.total_users }}</div>
          </div>
          <div class="metric-card">
            <div class="metric-label">累计金额</div>
            <div class="metric-value">¥{{ formatAmount(merchant.total_amount) }}</div>
          </div>
        </div>

        <div class="section-grid">
          <el-card class="page-card" shadow="never">
            <template #header>
              <span>商家资料</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="商家名称">{{ merchant.name }}</el-descriptions-item>
              <el-descriptions-item label="联系人">{{ merchant.contact_name || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ merchant.contact_phone || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="联系邮箱">{{ merchant.contact_email || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="经营分类">{{ merchant.business_category || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="营业时间">{{ merchant.business_hours || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="子商户号">{{ merchant.sub_mch_id || '未配置' }}</el-descriptions-item>
              <el-descriptions-item label="抽佣比例">{{ merchant.profit_sharing_enabled ? formatPercent(merchant.profit_sharing_ratio) : '未开启' }}</el-descriptions-item>
              <el-descriptions-item label="商家地址" :span="2">{{ merchant.address || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="商家公告" :span="2">{{ merchant.announcement || '未设置' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(merchant.created_at) }}</el-descriptions-item>
            </el-descriptions>
            <el-space style="margin-top: 16px;">
              <el-button @click="router.push(`/profit-sharing?merchant_id=${merchant.id}`)">查看分账历史</el-button>
              <el-button type="primary" plain @click="openQrcode">查看商家二维码</el-button>
            </el-space>
          </el-card>

          <el-card class="page-card" shadow="never">
            <template #header>
              <span>图片资产</span>
            </template>
            <div class="asset-grid">
              <div class="asset-box">
                <div style="margin-bottom: 12px; font-weight: 600;">商家 Logo</div>
                <img v-if="merchant.logo" :src="merchant.logo" class="asset-preview" alt="商家 Logo" />
                <div v-else class="empty-asset">未上传</div>
                <label style="display: inline-block; margin-top: 12px;">
                  <input hidden type="file" accept="image/*" @change="updateAsset('logo', $event)" />
                  <el-button type="primary" plain>更换 Logo</el-button>
                </label>
              </div>
              <div class="asset-box">
                <div style="margin-bottom: 12px; font-weight: 600;">背景图</div>
                <img v-if="merchant.cover_image" :src="merchant.cover_image" class="asset-preview" alt="背景图" />
                <div v-else class="empty-asset">未上传</div>
                <label style="display: inline-block; margin-top: 12px;">
                  <input hidden type="file" accept="image/*" @change="updateAsset('cover_image', $event)" />
                  <el-button type="primary" plain>更换背景图</el-button>
                </label>
              </div>
            </div>
          </el-card>
        </div>
      </template>
    </el-skeleton>

    <el-dialog v-model="qrcodeVisible" title="商家二维码" width="420px">
      <div style="text-align: center;">
        <img v-if="qrcodeUrl" :src="qrcodeUrl" alt="商家二维码" style="max-width: 260px; width: 100%; border-radius: 12px;" />
        <el-empty v-else description="暂无二维码" />
        <div style="margin-top: 16px; color: #6b7280; word-break: break-all;">{{ qrcodePagePath }}</div>
      </div>
    </el-dialog>
  </div>
</template>
