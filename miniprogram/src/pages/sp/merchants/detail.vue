<template>
  <view class="merchant-detail-container">
    <!-- 加载状态 -->
    <view v-if="loading" class="loading-container">
      <text>加载中...</text>
    </view>

    <!-- 商家详情 -->
    <view v-else-if="merchant" class="detail-content">
      <!-- 基本信息 -->
      <view class="section">
        <view class="section-title">基本信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">商家名称</text>
            <text class="info-value">{{ merchant.name || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">行业分类</text>
            <text class="info-value">{{ merchant.business_category || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">商家地址</text>
            <text class="info-value">{{ merchant.address || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">商家状态</text>
            <text class="info-value" :class="getStatusClass(merchant.status)">
              {{ getStatusText(merchant.status) }}
            </text>
          </view>
        </view>
      </view>

      <!-- 联系人信息 -->
      <view class="section">
        <view class="section-title">联系人信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">姓名</text>
            <text class="info-value">{{ merchant.contact_name || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">电话</text>
            <text class="info-value">{{ merchant.contact_phone || '-' }}</text>
          </view>
        </view>
      </view>

      <!-- 营业执照信息 -->
      <view class="section" v-if="licenseInfo">
        <view class="section-title">营业执照信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">营业执照号</text>
            <text class="info-value">{{ licenseInfo.license_no || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">营业执照名称</text>
            <text class="info-value">{{ licenseInfo.license_name || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">法人姓名</text>
            <text class="info-value">{{ licenseInfo.legal_person || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">法人身份证号</text>
            <text class="info-value">{{ licenseInfo.legal_person_id || '-' }}</text>
          </view>
          <view class="info-item" v-if="licenseInfo.valid_from || licenseInfo.valid_to">
            <text class="info-label">有效期</text>
            <text class="info-value">
              {{ licenseInfo.valid_from || '-' }} 至 {{ licenseInfo.valid_to || '长期' }}
            </text>
          </view>
        </view>

        <!-- 营业执照图片 -->
        <view class="image-section" v-if="licenseInfo.license_image">
          <text class="info-label">营业执照图片</text>
          <view class="image-grid">
            <image
              class="preview-image"
              :src="licenseInfo.license_image"
              mode="aspectFill"
              @click="previewImage(licenseInfo.license_image)"
            />
          </view>
        </view>

        <!-- 身份证照片 -->
        <view class="image-section" v-if="licenseInfo.legal_person_id_front || licenseInfo.legal_person_id_back">
          <text class="info-label">身份证照片</text>
          <view class="image-grid">
            <view class="image-item" v-if="licenseInfo.legal_person_id_front">
              <text class="image-label">正面</text>
              <image
                class="preview-image"
                :src="licenseInfo.legal_person_id_front"
                mode="aspectFill"
                @click="previewImage(licenseInfo.legal_person_id_front)"
              />
            </view>
            <view class="image-item" v-if="licenseInfo.legal_person_id_back">
              <text class="image-label">反面</text>
              <image
                class="preview-image"
                :src="licenseInfo.legal_person_id_back"
                mode="aspectFill"
                @click="previewImage(licenseInfo.legal_person_id_back)"
              />
            </view>
          </view>
        </view>
      </view>

      <!-- 银行账户信息 -->
      <view class="section" v-if="bankAccountInfo">
        <view class="section-title">银行账户信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">开户银行</text>
            <text class="info-value">{{ bankAccountInfo.bank_name || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">开户支行</text>
            <text class="info-value">{{ bankAccountInfo.bank_branch || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">银行账号</text>
            <text class="info-value">{{ bankAccountInfo.account_no || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">账户名称</text>
            <text class="info-value">{{ bankAccountInfo.account_name || '-' }}</text>
          </view>
        </view>
      </view>

      <!-- 门店信息 -->
      <view class="section" v-if="storeImages && storeImages.length > 0">
        <view class="section-title">门店信息</view>
        <view class="info-grid" v-if="merchant.store_name">
          <view class="info-item">
            <text class="info-label">门店名称</text>
            <text class="info-value">{{ merchant.store_name }}</text>
          </view>
        </view>

        <!-- 门头照 -->
        <view class="image-section" v-if="storeImages.length > 0">
          <text class="info-label">门店照片</text>
          <view class="image-grid">
            <image
              v-for="(image, index) in storeImages"
              :key="index"
              class="preview-image"
              :src="image"
              mode="aspectFill"
              @click="previewImages(storeImages, index)"
            />
          </view>
        </view>
      </view>

      <!-- 经营数据 -->
      <view class="section">
        <view class="section-title">经营数据</view>
        <view class="stats-grid">
          <view class="stat-item">
            <text class="stat-value">{{ merchant.total_orders || 0 }}</text>
            <text class="stat-label">累计订单数</text>
          </view>
          <view class="stat-item">
            <text class="stat-value">¥{{ formatAmount(merchant.total_amount) }}</text>
            <text class="stat-label">累计金额</text>
          </view>
        </view>
        <view class="info-grid" style="margin-top: 24rpx;">
          <view class="info-item">
            <text class="info-label">入驻时间</text>
            <text class="info-value">{{ merchant.created_at || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">当前状态</text>
            <text class="info-value" :class="getStatusClass(merchant.status)">
              {{ getStatusText(merchant.status) }}
            </text>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载失败 -->
    <view v-else class="error-container">
      <text>加载失败，请重试</text>
      <button class="retry-btn" @click="loadMerchantDetail">重新加载</button>
    </view>

    <!-- 底部操作按钮 -->
    <view class="action-bar" v-if="merchant">
      <button class="action-btn" @click="goToSetRate">
        <text class="action-icon">⚙</text>
        <text class="action-text">设置手续费率</text>
      </button>
      <button class="action-btn" @click="showQrcode">
        <text class="action-icon">⬡</text>
        <text class="action-text">商家二维码</text>
      </button>
      <button class="action-btn primary" @click="contactMerchant">
        <text class="action-icon">☎</text>
        <text class="action-text">联系商家</text>
      </button>
    </view>

    <!-- 二维码弹窗 -->
    <view class="qrcode-modal" v-if="showQrcodeModal" @click="showQrcodeModal = false">
      <view class="qrcode-content" @click.stop>
        <text class="qrcode-title">商家小程序码</text>
        <image
          v-if="merchant?.qrcode_url"
          class="qrcode-image"
          :src="merchant.qrcode_url"
          mode="aspectFit"
        />
        <view v-else class="qrcode-placeholder">
          <text>暂无二维码</text>
        </view>
        <button class="qrcode-close" @click="showQrcodeModal = false">关闭</button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getMerchantDetail } from '@api'
import type { MerchantDetail } from '@types'

const merchantId = ref<number>(0)
const merchant = ref<MerchantDetail | null>(null)
const loading = ref<boolean>(true)
const showQrcodeModal = ref<boolean>(false)

onLoad((options: any) => {
  if (options?.id) {
    merchantId.value = parseInt(options.id)
    loadMerchantDetail()
  }
})

const licenseInfo = computed(() => {
  return merchant.value?.license || null
})

const bankAccountInfo = computed(() => {
  return merchant.value?.settings?.bank_account || null
})

const storeImages = computed(() => {
  return merchant.value?.settings?.store_images || []
})

async function loadMerchantDetail() {
  loading.value = true
  try {
    const res = await getMerchantDetail(merchantId.value)
    merchant.value = res
  } catch (error) {
    console.error('获取商家详情失败:', error)
    uni.showToast({
      title: '获取详情失败',
      icon: 'none'
    })
  } finally {
    loading.value = false
  }
}

function previewImage(url: string) {
  if (!url) return
  uni.previewImage({
    urls: [url],
    current: 0
  })
}

function previewImages(urls: string[], currentIndex: number = 0) {
  if (!urls || urls.length === 0) return
  uni.previewImage({
    urls: urls,
    current: currentIndex
  })
}

function getStatusText(status: number): string {
  const statusMap: Record<number, string> = {
    1: '营业中',
    2: '休息中',
    3: '已关闭'
  }
  return statusMap[status] || '未知'
}

function getStatusClass(status: number): string {
  const classMap: Record<number, string> = {
    1: 'status-open',
    2: 'status-rest',
    3: 'status-closed'
  }
  return classMap[status] || ''
}

function formatAmount(amount: number | undefined): string {
  if (!amount) return '0.00'
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function goToSetRate() {
  uni.navigateTo({
    url: `/pages/sp/merchants/rate?id=${merchantId.value}`
  })
}

function showQrcode() {
  if (merchant.value?.qrcode_url) {
    showQrcodeModal.value = true
  } else {
    uni.showToast({
      title: '暂无可用二维码',
      icon: 'none'
    })
  }
}

function contactMerchant() {
  if (!merchant.value?.contact_phone) {
    uni.showToast({
      title: '暂无联系电话',
      icon: 'none'
    })
    return
  }

  uni.makePhoneCall({
    phoneNumber: merchant.value.contact_phone,
    fail: () => {
      uni.showToast({
        title: '拨打电话失败',
        icon: 'none'
      })
    }
  })
}
</script>

<style scoped>
.merchant-detail-container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 140rpx;
}

.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100rpx 0;
  color: #666;
}

.retry-btn {
  margin-top: 20rpx;
  padding: 10rpx 40rpx;
  background-color: #007aff;
  color: #fff;
  border-radius: 8rpx;
  font-size: 28rpx;
}

.detail-content {
  padding: 20rpx;
}

.section {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 24rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #007aff;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.info-label {
  font-size: 26rpx;
  color: #999;
}

.info-value {
  font-size: 28rpx;
  color: #333;
}

.status-open {
  color: #52c41a;
}

.status-rest {
  color: #fa8c16;
}

.status-closed {
  color: #ff4d4f;
}

.image-section {
  margin-top: 24rpx;
}

.image-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: 12rpx;
}

.image-item {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.image-label {
  font-size: 24rpx;
  color: #666;
  text-align: center;
}

.preview-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background-color: #f0f0f0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24rpx;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
  padding: 24rpx;
  background-color: #f8f8f8;
  border-radius: 12rpx;
}

.stat-value {
  font-size: 36rpx;
  font-weight: 600;
  color: #007aff;
}

.stat-label {
  font-size: 24rpx;
  color: #999;
}

.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 16rpx;
  padding: 20rpx 32rpx;
  background-color: #fff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.action-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  padding: 16rpx 0;
  background-color: #f5f5f5;
  border: none;
  border-radius: 12rpx;
  font-size: 24rpx;
}

.action-btn::after {
  border: none;
}

.action-btn.primary {
  background: linear-gradient(135deg, #007aff, #0056cc);
  color: #fff;
}

.action-icon {
  font-size: 36rpx;
}

.action-text {
  font-size: 24rpx;
}

.qrcode-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.qrcode-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32rpx;
  padding: 48rpx;
  background-color: #fff;
  border-radius: 24rpx;
  margin: 40rpx;
}

.qrcode-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}

.qrcode-image {
  width: 400rpx;
  height: 400rpx;
}

.qrcode-placeholder {
  width: 400rpx;
  height: 400rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  color: #999;
  font-size: 28rpx;
}

.qrcode-close {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background: linear-gradient(135deg, #007aff, #0056cc);
  color: #fff;
  border-radius: 44rpx;
  font-size: 32rpx;
  border: none;
}

.qrcode-close::after {
  border: none;
}
</style>
