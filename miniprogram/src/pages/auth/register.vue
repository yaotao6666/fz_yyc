<template>
  <view class="register-container">
    <view class="form-card">
      <!-- 基本信息 -->
      <view class="section">
        <view class="section-title">基本信息</view>
        
        <view class="form-item">
          <view class="form-label">商家名称 <text class="required">*</text></view>
          <input
            v-model="formData.name"
            class="form-input"
            placeholder="请输入商家名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">联系人 <text class="required">*</text></view>
          <input
            v-model="formData.contact_name"
            class="form-input"
            placeholder="请输入联系人姓名"
          />
        </view>

        <view class="form-item">
          <view class="form-label">联系电话 <text class="required">*</text></view>
          <input
            v-model="formData.contact_phone"
            class="form-input"
            type="number"
            placeholder="请输入手机号码"
            maxlength="11"
          />
        </view>

        <view class="form-item">
          <view class="form-label">联系邮箱</view>
          <input
            v-model="formData.contact_email"
            class="form-input"
            type="text"
            placeholder="请输入邮箱（选填）"
          />
        </view>

        <view class="form-item">
          <view class="form-label">店铺地址 <text class="required">*</text></view>
          <input
            v-model="formData.address"
            class="form-input"
            placeholder="请输入详细地址"
          />
        </view>

        <view class="form-item">
          <view class="form-label">经营类目 <text class="required">*</text></view>
          <picker
            :value="categoryIndex"
            :range="categories"
            range-key="name"
            @change="onCategoryChange"
          >
            <view class="picker-value">
              {{ formData.business_category || '请选择经营类目' }}
              <text class="arrow">›</text>
            </view>
          </picker>
        </view>

        <view class="form-item">
          <view class="form-label">邀请码</view>
          <input
            v-model="formData.invite_code"
            class="form-input"
            placeholder="请输入邀请码（选填）"
          />
        </view>
      </view>

      <!-- 营业执照信息 -->
      <view class="section">
        <view class="section-title">营业执照</view>
        
        <view class="form-item">
          <view class="form-label">营业执照号 <text class="required">*</text></view>
          <input
            v-model="formData.license.license_no"
            class="form-input"
            placeholder="请输入统一社会信用代码"
          />
        </view>

        <view class="form-item">
          <view class="form-label">执照名称 <text class="required">*</text></view>
          <input
            v-model="formData.license.license_name"
            class="form-input"
            placeholder="请输入营业执照名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">法人姓名 <text class="required">*</text></view>
          <input
            v-model="formData.license.legal_person"
            class="form-input"
            placeholder="请输入法人姓名"
          />
        </view>

        <view class="form-item">
          <view class="form-label">法人身份证号 <text class="required">*</text></view>
          <input
            v-model="formData.license.legal_person_id"
            class="form-input"
            placeholder="请输入法人身份证号"
            maxlength="18"
          />
        </view>

        <view class="form-item">
          <view class="form-label">营业执照照片 <text class="required">*</text></view>
          <view class="upload-area" @click="chooseImage('license_image')">
            <image
              v-if="formData.license.license_image"
              :src="formData.license.license_image"
              class="upload-image"
              mode="aspectFill"
            />
            <view v-else class="upload-placeholder">
              <text class="icon">+</text>
              <text class="text">上传营业执照</text>
            </view>
          </view>
        </view>

        <view class="form-item">
          <view class="form-label">身份证正面 <text class="required">*</text></view>
          <view class="upload-area" @click="chooseImage('legal_person_id_front')">
            <image
              v-if="formData.license.legal_person_id_front"
              :src="formData.license.legal_person_id_front"
              class="upload-image"
              mode="aspectFill"
            />
            <view v-else class="upload-placeholder">
              <text class="icon">+</text>
              <text class="text">上传身份证正面</text>
            </view>
          </view>
        </view>

        <view class="form-item">
          <view class="form-label">身份证反面 <text class="required">*</text></view>
          <view class="upload-area" @click="chooseImage('legal_person_id_back')">
            <image
              v-if="formData.license.legal_person_id_back"
              :src="formData.license.legal_person_id_back"
              class="upload-image"
              mode="aspectFill"
            />
            <view v-else class="upload-placeholder">
              <text class="icon">+</text>
              <text class="text">上传身份证反面</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 银行卡信息 -->
      <view class="section">
        <view class="section-title">结算银行卡</view>
        
        <view class="form-item">
          <view class="form-label">开户银行 <text class="required">*</text></view>
          <input
            v-model="formData.bank_account.bank_name"
            class="form-input"
            placeholder="请输入开户银行名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">开户支行 <text class="required">*</text></view>
          <input
            v-model="formData.bank_account.bank_branch"
            class="form-input"
            placeholder="请输入开户支行"
          />
        </view>

        <view class="form-item">
          <view class="form-label">银行卡号 <text class="required">*</text></view>
          <input
            v-model="formData.bank_account.account_no"
            class="form-input"
            type="number"
            placeholder="请输入银行卡号"
          />
        </view>

        <view class="form-item">
          <view class="form-label">账户名称 <text class="required">*</text></view>
          <input
            v-model="formData.bank_account.account_name"
            class="form-input"
            placeholder="请输入账户名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">账户类型 <text class="required">*</text></view>
          <picker
            :value="accountTypeIndex"
            :range="accountTypes"
            range-key="name"
            @change="onAccountTypeChange"
          >
            <view class="picker-value">
              {{ formData.bank_account.account_type === 1 ? '对公账户' : formData.bank_account.account_type === 2 ? '对私账户' : '请选择账户类型' }}
              <text class="arrow">›</text>
            </view>
          </picker>
        </view>
      </view>

      <!-- 门店信息 -->
      <view class="section">
        <view class="section-title">门店信息</view>
        
        <view class="form-item">
          <view class="form-label">门店名称 <text class="required">*</text></view>
          <input
            v-model="formData.store_info.store_name"
            class="form-input"
            placeholder="请输入门店名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">门店照片</view>
          <view class="image-list">
            <view
              v-for="(img, index) in formData.store_info.store_images"
              :key="index"
              class="image-item"
            >
              <image :src="img" mode="aspectFill" />
              <view class="delete-btn" @click="removeStoreImage(index)">×</view>
            </view>
            <view
              v-if="formData.store_info.store_images.length < 4"
              class="add-image"
              @click="chooseStoreImage"
            >
              <text class="icon">+</text>
            </view>
          </view>
        </view>
      </view>

      <view class="submit-area">
        <button
          class="btn-submit"
          :disabled="submitting"
          @click="handleSubmit"
        >
          {{ submitting ? '提交中...' : '提交入驻申请' }}
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { merchantRegister } from '@api'
import { uploadImage } from '@api'

const categories = [
  { name: '餐饮', value: 'catering' },
  { name: '零售', value: 'retail' },
  { name: '水果生鲜', value: 'fresh' },
  { name: '鲜花礼品', value: 'flower' },
  { name: '服装鞋帽', value: 'clothing' },
  { name: '其他', value: 'other' }
]

const accountTypes = [
  { name: '对公账户', value: 1 },
  { name: '对私账户', value: 2 }
]

const formData = reactive({
  name: '',
  contact_name: '',
  contact_phone: '',
  contact_email: '',
  address: '',
  business_category: '',
  invite_code: '',
  license: {
    license_no: '',
    license_name: '',
    license_image: '',
    legal_person: '',
    legal_person_id: '',
    legal_person_id_front: '',
    legal_person_id_back: '',
    valid_from: '',
    valid_to: ''
  },
  bank_account: {
    bank_name: '',
    bank_branch: '',
    account_no: '',
    account_name: '',
    account_type: 0
  },
  store_info: {
    store_name: '',
    store_images: [] as string[]
  }
})

const categoryIndex = ref(-1)
const accountTypeIndex = ref(-1)
const submitting = ref(false)

const categoryIndexComputed = computed(() => {
  return categories.findIndex(c => c.value === formData.business_category)
})

function onCategoryChange(e: any) {
  const index = e.detail.value
  categoryIndex.value = index
  formData.business_category = categories[index].value
}

function onAccountTypeChange(e: any) {
  const index = e.detail.value
  accountTypeIndex.value = index
  formData.bank_account.account_type = accountTypes[index].value
}

function chooseImage(field: string) {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const tempFilePath = res.tempFilePaths[0]
      uni.showLoading({ title: '上传中...' })
      
      try {
        const result = await uploadImage(tempFilePath)
        ;(formData.license as any)[field] = result.url
        uni.hideLoading()
      } catch (error) {
        uni.hideLoading()
        uni.showToast({ title: '上传失败', icon: 'none' })
      }
    }
  })
}

function chooseStoreImage() {
  uni.chooseImage({
    count: 4 - formData.store_info.store_images.length,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      uni.showLoading({ title: '上传中...' })
      
      try {
        for (const tempFilePath of res.tempFilePaths) {
          const result = await uploadImage(tempFilePath)
          formData.store_info.store_images.push(result.url)
        }
        uni.hideLoading()
      } catch (error) {
        uni.hideLoading()
        uni.showToast({ title: '上传失败', icon: 'none' })
      }
    }
  })
}

function removeStoreImage(index: number) {
  formData.store_info.store_images.splice(index, 1)
}

async function handleSubmit() {
  // 表单验证
  if (!formData.name) {
    return uni.showToast({ title: '请输入商家名称', icon: 'none' })
  }
  if (!formData.contact_name) {
    return uni.showToast({ title: '请输入联系人', icon: 'none' })
  }
  if (!formData.contact_phone) {
    return uni.showToast({ title: '请输入联系电话', icon: 'none' })
  }
  if (!formData.address) {
    return uni.showToast({ title: '请输入店铺地址', icon: 'none' })
  }
  if (!formData.business_category) {
    return uni.showToast({ title: '请选择经营类目', icon: 'none' })
  }

  submitting.value = true

  try {
    const res = await merchantRegister(formData)
    
    uni.showToast({ title: '提交成功', icon: 'success' })
    
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error: any) {
    uni.showToast({ title: error.message || '提交失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}

.form-card {
  background: #ffffff;
  border-radius: 24rpx;
  overflow: hidden;
}

.section {
  padding: 32rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.section:last-child {
  border-bottom: none;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 32rpx;
  padding-left: 16rpx;
  border-left: 6rpx solid #007AFF;
}

.form-item {
  margin-bottom: 28rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.required {
  color: #ff4d4f;
}

.form-input {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.picker-value {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 30rpx;
  color: #1a1a1a;
}

.arrow {
  font-size: 32rpx;
  color: #cccccc;
}

.upload-area {
  width: 200rpx;
  height: 200rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  overflow: hidden;
}

.upload-image {
  width: 100%;
  height: 100%;
}

.upload-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-placeholder .icon {
  font-size: 64rpx;
  color: #cccccc;
  line-height: 1;
}

.upload-placeholder .text {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
  border-radius: 16rpx;
  overflow: hidden;
}

.image-item image {
  width: 100%;
  height: 100%;
}

.delete-btn {
  position: absolute;
  top: 8rpx;
  right: 8rpx;
  width: 40rpx;
  height: 40rpx;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 28rpx;
}

.add-image {
  width: 200rpx;
  height: 200rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.add-image .icon {
  font-size: 64rpx;
  color: #cccccc;
}

.submit-area {
  padding: 32rpx;
}

.btn-submit {
  width: 100%;
  height: 96rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 48rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.btn-submit[disabled] {
  background: #cccccc;
}
</style>
