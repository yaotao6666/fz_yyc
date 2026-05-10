<template>
  <view class="audit-detail-container">
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
      <view class="section">
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
      <view class="section">
        <view class="section-title">银行账户信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">开户银行</text>
            <text class="info-value">{{ bankInfo.bank_name || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">开户支行</text>
            <text class="info-value">{{ bankInfo.bank_branch || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">银行账号</text>
            <text class="info-value">{{ bankInfo.account_no || '-' }}</text>
          </view>
          <view class="info-item">
            <text class="info-label">账户名称</text>
            <text class="info-value">{{ bankInfo.account_name || '-' }}</text>
          </view>
        </view>
      </view>

      <!-- 门店信息 -->
      <view class="section">
        <view class="section-title">门店信息</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">门店名称</text>
            <text class="info-value">{{ storeInfo.store_name || '-' }}</text>
          </view>
        </view>

        <!-- 门店图片 -->
        <view class="image-section" v-if="storeInfo.store_images && storeInfo.store_images.length > 0">
          <text class="info-label">门店照片</text>
          <view class="image-grid">
            <image
              v-for="(image, index) in storeInfo.store_images"
              :key="index"
              class="preview-image"
              :src="image"
              mode="aspectFill"
              @click="previewImages(storeInfo.store_images, index)"
            />
          </view>
        </view>
      </view>

      <!-- 审核操作区域 - 仅待审核状态显示 -->
      <view class="section audit-section" v-if="isPending">
        <view class="section-title">审核操作</view>

        <!-- 审核状态 -->
        <view class="audit-status">
          <text class="status-badge pending">待审核</text>
        </view>

        <!-- 备注输入 -->
        <view class="remark-input">
          <text class="input-label">审核备注</text>
          <textarea
            v-model="auditRemark"
            class="remark-textarea"
            placeholder="请输入审核备注（选填）"
            maxlength="200"
          />
        </view>

        <!-- 操作按钮 -->
        <view class="action-buttons">
          <button
            class="btn btn-reject"
            :disabled="submitting"
            @click="handleAudit('rejected')"
          >
            {{ submitting ? '提交中...' : '拒绝' }}
          </button>
          <button
            class="btn btn-approve"
            :disabled="submitting"
            @click="handleAudit('approved')"
          >
            {{ submitting ? '提交中...' : '通过' }}
          </button>
        </view>
      </view>

      <!-- 审核结果 - 非待审核状态显示 -->
      <view class="section audit-result" v-else>
        <view class="section-title">审核结果</view>
        <view class="info-grid">
          <view class="info-item">
            <text class="info-label">审核状态</text>
            <text class="info-value" :class="getAuditStatusClass()">
              {{ getAuditStatusText() }}
            </text>
          </view>
          <view class="info-item" v-if="merchant.audit_remark">
            <text class="info-label">审核备注</text>
            <text class="info-value">{{ merchant.audit_remark }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载失败 -->
    <view v-else class="error-container">
      <text>加载失败，请重试</text>
      <button class="retry-btn" @click="loadMerchantDetail">重新加载</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { get, post } from '@api'

interface MerchantDetail {
  id: number
  name: string
  logo?: string
  contact_name?: string
  contact_phone?: string
  address?: string
  business_category?: string
  audit_status: number
  audit_remark?: string
  application?: {
    business_license_info?: any
    legal_person_info?: any
    bank_account_info?: any
    store_info?: any
  }
}

const merchantId = ref<number>(0)
const merchant = ref<MerchantDetail | null>(null)
const loading = ref<boolean>(true)
const submitting = ref<boolean>(false)
const auditRemark = ref<string>('')

onLoad((options: any) => {
  if (options?.id) {
    merchantId.value = parseInt(options.id)
    loadMerchantDetail()
  }
})

const isPending = computed(() => {
  return merchant.value?.audit_status === 0
})

const licenseInfo = computed(() => {
  return merchant.value?.application?.business_license_info || {}
})

const bankInfo = computed(() => {
  return merchant.value?.application?.bank_account_info || {}
})

const storeInfo = computed(() => {
  const store = merchant.value?.application?.store_info
  if (store) {
    return {
      store_name: store.store_name || '',
      store_images: store.store_images || []
    }
  }
  return { store_name: '', store_images: [] }
})

async function loadMerchantDetail() {
  loading.value = true
  try {
    const res = await get<MerchantDetail>(`/api/v1/sp/merchants/${merchantId.value}`)
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

function getAuditStatusText(): string {
  if (!merchant.value) return ''
  const statusMap: Record<number, string> = {
    0: '待审核',
    1: '已通过',
    2: '已拒绝'
  }
  return statusMap[merchant.value.audit_status] || '未知'
}

function getAuditStatusClass(): string {
  if (!merchant.value) return ''
  const classMap: Record<number, string> = {
    0: 'status-pending',
    1: 'status-approved',
    2: 'status-rejected'
  }
  return classMap[merchant.value.audit_status] || ''
}

async function handleAudit(action: 'approved' | 'rejected') {
  if (submitting.value) return

  const confirmText = action === 'approved' ? '确认通过审核？' : '确认拒绝审核？'

  uni.showModal({
    title: '审核确认',
    content: confirmText,
    success: async (res) => {
      if (res.confirm) {
        await performAudit(action)
      }
    }
  })
}

async function performAudit(action: 'approved' | 'rejected') {
  submitting.value = true

  try {
    const auditStatus = action === 'approved' ? 1 : 2
    const res = await post(`/api/v1/sp/merchants/${merchantId.value}/approve`, {
      audit_status: auditStatus,
      audit_remark: auditRemark.value
    })

    uni.showToast({
      title: action === 'approved' ? '审核通过' : '审核拒绝',
      icon: 'success'
    })

    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error) {
    console.error('审核操作失败:', error)
    uni.showToast({
      title: '审核操作失败',
      icon: 'none'
    })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.audit-detail-container {
  min-height: 100vh;
  background-color: #f5f5f5;
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

.audit-section {
  position: sticky;
  bottom: 0;
  z-index: 100;
}

.audit-status {
  margin-bottom: 24rpx;
}

.status-badge {
  display: inline-block;
  padding: 8rpx 24rpx;
  border-radius: 8rpx;
  font-size: 28rpx;
  font-weight: 500;
}

.status-badge.pending {
  background-color: #fff7e6;
  color: #fa8c16;
}

.status-pending {
  color: #fa8c16;
}

.status-approved {
  color: #52c41a;
}

.status-rejected {
  color: #ff4d4f;
}

.remark-input {
  margin-bottom: 24rpx;
}

.input-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 12rpx;
}

.remark-textarea {
  width: 100%;
  min-height: 120rpx;
  padding: 16rpx;
  border: 1rpx solid #e8e8e8;
  border-radius: 8rpx;
  font-size: 28rpx;
  box-sizing: border-box;
  background-color: #fafafa;
}

.action-buttons {
  display: flex;
  gap: 24rpx;
}

.btn {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  border: none;
}

.btn-reject {
  background-color: #fff;
  color: #ff4d4f;
  border: 2rpx solid #ff4d4f;
}

.btn-reject:active {
  background-color: #fff1f0;
}

.btn-approve {
  background: linear-gradient(135deg, #007aff, #0056cc);
  color: #fff;
}

.btn-approve:active {
  opacity: 0.9;
}

.btn[disabled] {
  opacity: 0.6;
}

.audit-result {
  background-color: #fff;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20rpx;
}
</style>
