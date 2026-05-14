<template>
  <view class="merchant-detail-container">
    <view v-if="loading" class="state-card">加载中...</view>
    <view v-else-if="!merchant" class="state-card error-state">
      <text>加载失败，请重试</text>
      <button class="retry-btn" @click="loadMerchantDetail">重新加载</button>
    </view>

    <template v-else>
      <scroll-view class="detail-scroll" scroll-y>
        <view class="hero-card">
          <view class="hero-main">
            <view>
              <text class="merchant-name">{{ merchant.name }}</text>
              <text class="merchant-meta">{{ merchant.contact_name || '未设置联系人' }} · {{ merchant.contact_phone || '未设置电话' }}</text>
            </view>
            <view class="merchant-status" :class="getStatusClass(merchant.status)">
              {{ getStatusText(merchant.status) }}
            </view>
          </view>
          <view class="hero-summary">
            <text class="hero-item">收款商户号：{{ merchant.sub_mch_id || '未配置' }}</text>
            <text class="hero-item">支付配置：{{ getPaymentConfigText(merchant.payment_config_status) }}</text>
            <text class="hero-item">
              分账设置：{{ merchant.profit_sharing_enabled ? `已开启 ${formatRatio(merchant.profit_sharing_ratio)}` : '未开启' }}
            </text>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">图片资产</view>
          <view class="asset-grid">
            <view class="asset-card">
              <text class="asset-label">商家 Logo</text>
              <image
                v-if="merchant.logo"
                class="asset-image logo-image"
                :src="merchant.logo"
                mode="aspectFill"
                @click="previewImage(merchant.logo)"
              />
              <view v-else class="asset-placeholder">未上传</view>
              <button class="asset-btn" :disabled="uploadingField === 'logo'" @click="chooseAndUploadImage('logo')">
                {{ uploadingField === 'logo' ? '上传中...' : '更换 Logo' }}
              </button>
            </view>
            <view class="asset-card">
              <text class="asset-label">背景图</text>
              <image
                v-if="merchant.cover_image"
                class="asset-image cover-image"
                :src="merchant.cover_image"
                mode="aspectFill"
                @click="previewImage(merchant.cover_image)"
              />
              <view v-else class="asset-placeholder">未上传</view>
              <button class="asset-btn" :disabled="uploadingField === 'cover_image'" @click="chooseAndUploadImage('cover_image')">
                {{ uploadingField === 'cover_image' ? '上传中...' : '更换背景图' }}
              </button>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">商家资料</view>
          <view class="info-grid">
            <view class="info-item">
              <text class="info-label">行业分类</text>
              <text class="info-value">{{ merchant.business_category || '未设置' }}</text>
            </view>
            <view class="info-item">
              <text class="info-label">联系邮箱</text>
              <text class="info-value">{{ merchant.contact_email || '未设置' }}</text>
            </view>
            <view class="info-item full-width">
              <text class="info-label">营业时间</text>
              <text class="info-value">{{ merchant.business_hours || '未设置' }}</text>
            </view>
            <view class="info-item full-width">
              <text class="info-label">商家地址</text>
              <text class="info-value">{{ merchant.address || '未设置' }}</text>
            </view>
            <view class="info-item full-width">
              <text class="info-label">商家公告</text>
              <text class="info-value">{{ merchant.announcement || '未设置' }}</text>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">支付与分账配置</view>
          <view class="info-grid">
            <view class="info-item">
              <text class="info-label">子商户号</text>
              <text class="info-value">{{ merchant.sub_mch_id || '未配置' }}</text>
            </view>
            <view class="info-item">
              <text class="info-label">支付配置状态</text>
              <text class="info-value" :class="getPaymentConfigClass(merchant.payment_config_status)">
                {{ getPaymentConfigText(merchant.payment_config_status) }}
              </text>
            </view>
            <view class="info-item">
              <text class="info-label">是否分账</text>
              <text class="info-value">{{ merchant.profit_sharing_enabled ? '已开启' : '未开启' }}</text>
            </view>
            <view class="info-item">
              <text class="info-label">抽佣比例</text>
              <text class="info-value">{{ formatRatio(merchant.profit_sharing_ratio) }}</text>
            </view>
          </view>
        </view>

        <view class="section-card">
          <view class="section-title">经营数据</view>
          <view class="stats-grid">
            <view class="stat-card">
              <text class="stat-value">{{ merchant.total_users || 0 }}</text>
              <text class="stat-label">用户数</text>
            </view>
            <view class="stat-card">
              <text class="stat-value">{{ merchant.total_orders || 0 }}</text>
              <text class="stat-label">订单数</text>
            </view>
            <view class="stat-card">
              <text class="stat-value">¥{{ formatAmount(merchant.total_amount) }}</text>
              <text class="stat-label">累计金额</text>
            </view>
            <view class="stat-card">
              <text class="stat-value">{{ formatDate(merchant.created_at) }}</text>
              <text class="stat-label">创建时间</text>
            </view>
          </view>
        </view>
      </scroll-view>

      <view class="action-bar">
        <button class="action-btn" @click="goEditMerchant">编辑商家</button>
        <button class="action-btn" @click="goEditPaymentConfig">支付配置</button>
        <button class="action-btn" @click="goProfitSharingHistory">分账历史</button>
        <button class="action-btn primary" @click="openQrcode">商家二维码</button>
      </view>

      <view class="qrcode-modal" v-if="showQrcodeModal" @click="showQrcodeModal = false">
        <view class="qrcode-content" @click.stop>
          <text class="qrcode-title">商家二维码</text>
          <image v-if="merchant.qrcode_url" class="qrcode-image" :src="merchant.qrcode_url" mode="aspectFit" />
          <view v-else class="asset-placeholder">暂无二维码</view>
          <view class="qrcode-actions">
            <button class="modal-btn" @click="contactMerchant">联系商家</button>
            <button class="modal-btn primary" @click="showQrcodeModal = false">关闭</button>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getMerchantDetail, updateSpMerchantAssets, uploadImage } from '@api'
import { PaymentConfigStatusText } from '@types'
import type { MerchantDetail, PaymentConfigStatus } from '@types'

const merchantId = ref(0)
const merchant = ref<MerchantDetail | null>(null)
const loading = ref(true)
const showQrcodeModal = ref(false)
const uploadingField = ref<'logo' | 'cover_image' | ''>('')

onLoad((options: any) => {
  const id = Number(options?.id || 0)
  if (!Number.isNaN(id) && id > 0) {
    merchantId.value = id
    loadMerchantDetail()
  }
})

async function loadMerchantDetail() {
  loading.value = true
  try {
    merchant.value = await getMerchantDetail(merchantId.value)
  } catch (requestError) {
    console.error('获取商家详情失败:', requestError)
    uni.showToast({ title: '获取商家详情失败', icon: 'none' })
    merchant.value = null
  } finally {
    loading.value = false
  }
}

function previewImage(url?: string) {
  if (!url) {
    return
  }
  uni.previewImage({ urls: [url], current: 0 })
}

async function chooseAndUploadImage(field: 'logo' | 'cover_image') {
  if (!merchant.value) {
    return
  }

  try {
    const chooseResult = await uni.chooseImage({
      count: 1,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera']
    })
    const filePath = chooseResult.tempFilePaths?.[0]
    if (!filePath) {
      return
    }

    uploadingField.value = field
    const uploadResult = await uploadImage(filePath)
    merchant.value = await updateSpMerchantAssets(merchantId.value, {
      [field]: uploadResult.url
    })
    uni.showToast({ title: '图片更新成功', icon: 'success' })
  } catch (requestError) {
    console.error('更新图片失败:', requestError)
    uni.showToast({ title: '图片更新失败', icon: 'none' })
  } finally {
    uploadingField.value = ''
  }
}

function getStatusText(status: number) {
  const statusMap: Record<number, string> = {
    1: '营业中',
    2: '休息中',
    3: '已关闭'
  }
  return statusMap[status] || '未知状态'
}

function getStatusClass(status: number) {
  const classMap: Record<number, string> = {
    1: 'open',
    2: 'rest',
    3: 'closed'
  }
  return classMap[status] || ''
}

function getPaymentConfigText(status?: number) {
  return PaymentConfigStatusText[(status ?? 0) as PaymentConfigStatus] || '待完善'
}

function getPaymentConfigClass(status?: number) {
  return Number(status || 0) === 1 ? 'success' : 'warning'
}

function formatAmount(amount = 0) {
  return Number(amount || 0).toFixed(2)
}

function formatRatio(ratio = 0) {
  return `${Number(ratio || 0).toFixed(2)}%`
}

function formatDate(value?: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').slice(0, 19)
}

function goEditMerchant() {
  uni.navigateTo({ url: `/pages/sp/merchants/edit?id=${merchantId.value}` })
}

function goEditPaymentConfig() {
  uni.navigateTo({ url: `/pages/sp/merchants/edit?id=${merchantId.value}` })
}

function goProfitSharingHistory() {
  uni.navigateTo({ url: `/pages/sp/settlements/history?merchant_id=${merchantId.value}` })
}

function openQrcode() {
  showQrcodeModal.value = true
}

function contactMerchant() {
  if (!merchant.value?.contact_phone) {
    uni.showToast({ title: '暂无联系电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: merchant.value.contact_phone })
}
</script>

<style scoped>
.merchant-detail-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.detail-scroll {
  height: calc(100vh - 136rpx);
  padding: 24rpx;
  box-sizing: border-box;
}

.state-card {
  margin: 24rpx;
  padding: 48rpx 32rpx;
  border-radius: 24rpx;
  background: #ffffff;
  text-align: center;
  color: #4e5969;
}

.error-state {
  color: #cf1322;
}

.retry-btn {
  margin-top: 24rpx;
  width: 220rpx;
  height: 76rpx;
  line-height: 76rpx;
  border: none;
  border-radius: 999rpx;
  background: #1677ff;
  color: #ffffff;
}

.hero-card {
  padding: 32rpx;
  border-radius: 28rpx;
  background: linear-gradient(135deg, #3f7cff 0%, #635bff 100%);
  color: #ffffff;
  margin-bottom: 24rpx;
}

.hero-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.merchant-name {
  display: block;
  font-size: 38rpx;
  font-weight: 600;
}

.merchant-meta {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  opacity: 0.92;
}

.merchant-status {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.16);
  font-size: 22rpx;
}

.hero-summary {
  display: grid;
  gap: 12rpx;
  margin-top: 24rpx;
}

.hero-item {
  font-size: 24rpx;
  line-height: 1.6;
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

.asset-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.asset-card {
  padding: 24rpx;
  border-radius: 20rpx;
  background: #f7f8fa;
}

.asset-label {
  display: block;
  font-size: 24rpx;
  color: #4e5969;
  margin-bottom: 16rpx;
}

.asset-image {
  width: 100%;
  border-radius: 18rpx;
  background: #ffffff;
}

.logo-image,
.cover-image,
.asset-placeholder {
  height: 220rpx;
}

.asset-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 18rpx;
  background: #ffffff;
  color: #86909c;
  font-size: 24rpx;
}

.asset-btn {
  margin-top: 16rpx;
  width: 100%;
  height: 72rpx;
  line-height: 72rpx;
  border: none;
  border-radius: 16rpx;
  background: #eef3ff;
  color: #1677ff;
  font-size: 24rpx;
}

.info-grid,
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.info-item,
.stat-card {
  padding: 22rpx 24rpx;
  border-radius: 18rpx;
  background: #f7f8fa;
}

.info-item.full-width {
  grid-column: 1 / -1;
}

.info-label,
.stat-label {
  display: block;
  font-size: 22rpx;
  color: #86909c;
}

.info-value,
.stat-value {
  display: block;
  margin-top: 10rpx;
  font-size: 26rpx;
  line-height: 1.6;
  color: #1f2329;
  word-break: break-all;
}

.info-value.success {
  color: #389e0d;
}

.info-value.warning {
  color: #d48806;
}

.action-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12rpx;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -8rpx 24rpx rgba(0, 0, 0, 0.06);
}

.action-btn {
  margin: 0;
  height: 76rpx;
  line-height: 76rpx;
  border: none;
  border-radius: 18rpx;
  background: #f2f3f5;
  color: #1f2329;
  font-size: 24rpx;
}

.action-btn.primary {
  background: #1677ff;
  color: #ffffff;
}

.qrcode-modal {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32rpx;
  background: rgba(0, 0, 0, 0.45);
}

.qrcode-content {
  width: 100%;
  max-width: 560rpx;
  padding: 32rpx;
  border-radius: 24rpx;
  background: #ffffff;
}

.qrcode-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2329;
  text-align: center;
}

.qrcode-image {
  width: 360rpx;
  height: 360rpx;
  display: block;
  margin: 32rpx auto 0;
}

.qrcode-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 28rpx;
}

.modal-btn {
  margin: 0;
  height: 76rpx;
  line-height: 76rpx;
  border: none;
  border-radius: 18rpx;
  background: #f2f3f5;
  color: #1f2329;
  font-size: 26rpx;
}

.modal-btn.primary {
  background: #1677ff;
  color: #ffffff;
}
</style>
