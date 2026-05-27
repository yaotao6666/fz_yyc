<template>
  <view class="edit-container">
    <scroll-view class="form-scroll" scroll-y>
      <!-- 基本信息 -->
      <view class="section">
        <view class="section-title">基本信息</view>

        <view class="form-item">
          <view class="form-label">商品名称 <text class="required">*</text></view>
          <input
            v-model="formData.name"
            class="form-input"
            placeholder="请输入商品名称"
          />
        </view>

        <view class="form-item">
          <view class="form-label">商品分类 <text class="required">*</text></view>
          <picker
            :value="categoryIndex"
            :range="categories"
            range-key="name"
            @change="onCategoryChange"
          >
            <view class="picker-value">
              {{ selectedCategoryName || '请选择分类' }}
              <text class="arrow">›</text>
            </view>
          </picker>
        </view>

        <view class="form-item">
          <view class="form-label">商品描述</view>
          <textarea
            v-model="formData.description"
            class="form-textarea"
            placeholder="请输入商品描述（选填）"
            :maxlength="500"
          />
        </view>
      </view>

      <!-- 价格库存 -->
      <view class="section">
        <view class="section-title">价格库存</view>

        <view class="form-row">
          <view class="form-item half">
            <view class="form-label">售价 <text class="required">*</text></view>
            <input
              :value="formData.price"
              type="digit"
              class="form-input"
              placeholder="0.00"
              @input="onAmountInput('price', $event)"
              @blur="onAmountBlur('price')"
            />
          </view>
          <view class="form-item half">
            <view class="form-label">原价</view>
            <input
              :value="formData.original_price"
              type="digit"
              class="form-input"
              placeholder="0.00"
              @input="onAmountInput('original_price', $event)"
              @blur="onAmountBlur('original_price', true)"
            />
          </view>
        </view>

        <view class="form-row">
          <view class="form-item half">
            <view class="form-label">库存 <text class="required">*</text></view>
            <input
              v-model.number="formData.stock"
              type="number"
              class="form-input"
              placeholder="0"
            />
          </view>
          <view class="form-item half">
            <view class="form-label">单位</view>
            <input
              v-model="formData.unit"
              class="form-input"
              placeholder="份/个/盒"
            />
          </view>
        </view>
      </view>

      <!-- 商品图片 -->
      <view class="section">
        <view class="section-title">商品图片</view>
        <view class="image-list">
          <view
            v-for="(img, index) in formData.images"
            :key="index"
            class="image-item"
          >
            <image :src="getImagePreviewSrc(img)" mode="aspectFill" />
            <view class="delete-btn" @click="removeImage(index)">×</view>
          </view>
          <view
            v-if="formData.images.length < 9"
            class="add-image"
            @click="chooseImage"
          >
            <text class="icon">+</text>
            <text class="text">{{ formData.images.length }}/9</text>
          </view>
        </view>
      </view>

      <!-- 商品规格 -->
      <view class="section">
        <view class="section-header">
          <view class="section-title">商品规格</view>
          <view class="add-spec-btn" @click="addSpec">添加规格</view>
        </view>

        <view
          v-for="(spec, specIndex) in formData.specs"
          :key="specIndex"
          class="spec-item"
        >
          <view class="spec-header">
            <input
              v-model="spec.name"
              class="spec-name-input"
              placeholder="规格名称（如：份量）"
            />
            <text class="delete-spec" @click="removeSpec(specIndex)">删除</text>
          </view>
          <view
            v-for="(option, optIndex) in spec.options"
            :key="optIndex"
            class="spec-option"
          >
            <input
              v-model="option.name"
              class="option-name"
              placeholder="选项名称"
            />
            <input
              :value="option.price"
              type="digit"
              class="option-price"
              placeholder="0.00"
              @input="onOptionPriceInput(specIndex, optIndex, $event)"
              @blur="onOptionPriceBlur(specIndex, optIndex)"
            />
            <text class="delete-option" @click="removeOption(specIndex, optIndex)">×</text>
          </view>
          <view class="add-option" @click="addOption(specIndex)">+ 添加选项</view>
        </view>
      </view>

      <!-- 排序 -->
      <view class="section">
        <view class="section-title">排序</view>
        <view class="form-item">
          <input
            v-model.number="formData.sort"
            type="number"
            class="form-input"
            placeholder="数字越小排序越靠前"
          />
        </view>
      </view>
    </scroll-view>

    <!-- 底部按钮 -->
    <view class="bottom-bar">
      <button class="btn-draft" @click="saveDraft">保存草稿</button>
      <button class="btn-publish" :disabled="submitting" @click="handleSubmit">
        {{ submitting ? '发布中...' : '发布商品' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getProduct, createProduct, updateProduct, getCategories, uploadImage, ResponseCode } from '@api'
import type { Category, Product, ProductUpsertPayload } from '@types'

const productId = ref<number | null>(null)
const categories = ref<Category[]>([])
const categoryIndex = ref(-1)
const submitting = ref(false)
const imagePreviewMap = ref<Record<string, string>>({})

type ProductLoadError = {
  message?: string
  code?: number
  statusCode?: number
}

type ProductSpecFormOption = {
  name: string
  price: string
}

type ProductSpecForm = {
  name: string
  options: ProductSpecFormOption[]
}

type ProductAmountField = 'price' | 'original_price'

const formData = reactive({
  name: '',
  category_id: 0,
  description: '',
  price: '',
  original_price: '',
  stock: 0,
  unit: '份',
  images: [] as string[],
  specs: [] as ProductSpecForm[],
  sort: 0
})

const selectedCategoryName = computed(() => {
  if (categoryIndex.value >= 0 && categories.value[categoryIndex.value]) {
    return categories.value[categoryIndex.value].name
  }
  return ''
})

onLoad((options: any) => {
  initializePage(options)
})

async function initializePage(options: any) {
  if (options.category_id) {
    formData.category_id = Number(options.category_id)
  }

  await loadCategories()

  if (options.id) {
    productId.value = Number(options.id)
    await loadProduct(productId.value)
  } else {
    syncCategoryIndex()
  }
}

function syncCategoryIndex() {
  const index = categories.value.findIndex(c => c.id === formData.category_id)
  categoryIndex.value = index
}

async function loadCategories() {
  try {
    const res = await getCategories()
    categories.value = res || []
    syncCategoryIndex()
  } catch (error) {
    console.error('加载分类失败:', error)
    categories.value = []
    categoryIndex.value = -1
  }
}

function getProductLoadErrorDialog(error: ProductLoadError) {
  if (
    error.statusCode === 404
    || error.code === ResponseCode.NOT_FOUND
    || error.code === ResponseCode.PRODUCT_NOT_FOUND
  ) {
    return {
      title: '商品不存在',
      content: '该商品可能已删除，或当前账号已无权访问。'
    }
  }

  if (error.statusCode === 500 || error.code === ResponseCode.SERVER_ERROR) {
    return {
      title: '服务异常',
      content: error.message
        ? `商品详情读取失败：${error.message}`
        : '商品详情读取失败，请稍后重试。'
    }
  }

  return {
    title: '加载失败',
    content: error.message || '商品详情加载失败，请稍后重试。'
  }
}

async function loadProduct(id: number) {
  try {
    const product = await getProduct(id, { showErrorToast: false })
    
    formData.name = product.name
    formData.category_id = product.category_id
    formData.description = product.description || ''
    formData.price = formatAmountForInput(product.price)
    formData.original_price = formatAmountForInput(product.original_price, true)
    formData.stock = product.stock
    formData.unit = product.unit || '份'
    formData.images = product.images || []
    imagePreviewMap.value = {}
    formData.specs = normalizeProductSpecs(product.specs)
    formData.sort = product.sort || 0

    syncCategoryIndex()
  } catch (error) {
    const productLoadError = error as ProductLoadError
    const errorDialog = getProductLoadErrorDialog(productLoadError)
    console.error('加载商品失败:', productLoadError)
    uni.showModal({
      title: errorDialog.title,
      content: errorDialog.content,
      showCancel: false,
      success: () => {
        uni.navigateBack()
      }
    })
  }
}

function onCategoryChange(e: any) {
  const index = e.detail.value
  categoryIndex.value = index
  formData.category_id = categories.value[index].id
}

function chooseImage() {
  const count = 9 - formData.images.length
  uni.chooseImage({
    count,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      uni.showLoading({ title: '上传中...' })
      
      try {
        for (const tempFilePath of res.tempFilePaths) {
          const result = await uploadImage(tempFilePath)
          formData.images.push(result.url)
          imagePreviewMap.value[result.url] = tempFilePath
        }
      } catch (error) {
        uni.showToast({ title: '上传失败', icon: 'none' })
      } finally {
        uni.hideLoading()
      }
    }
  })
}

function removeImage(index: number) {
  const [removedImage] = formData.images.splice(index, 1)
  if (removedImage) {
    delete imagePreviewMap.value[removedImage]
  }
}

function getImagePreviewSrc(imageUrl: string) {
  return imagePreviewMap.value[imageUrl] || imageUrl
}

function getInputValue(event: any) {
  return String(event?.detail?.value ?? '')
}

// 金额输入先保留字符串中间态，避免小程序 digit 输入在录入小数时被 number 转换打断。
function sanitizeAmountInput(rawValue: string) {
  const value = rawValue.replace(/[^\d.]/g, '')
  if (!value) {
    return ''
  }

  const firstDotIndex = value.indexOf('.')
  if (firstDotIndex === -1) {
    return value
  }

  const integerPart = value.slice(0, firstDotIndex) || '0'
  const decimalPart = value.slice(firstDotIndex + 1).replace(/\./g, '').slice(0, 2)
  return `${integerPart}.${decimalPart}`
}

function roundAmount(value: number) {
  return Math.round(value * 100) / 100
}

function parseAmountValue(value: string) {
  const sanitizedValue = sanitizeAmountInput(value).replace(/\.$/, '')
  if (!sanitizedValue) {
    return undefined
  }

  const amount = Number(sanitizedValue)
  if (!Number.isFinite(amount)) {
    return undefined
  }

  return roundAmount(amount)
}

function formatAmountForInput(amount?: number, emptyWhenZero = false) {
  if (amount === undefined || amount === null) {
    return ''
  }

  const normalizedAmount = roundAmount(Number(amount))
  if (!Number.isFinite(normalizedAmount)) {
    return ''
  }

  if (emptyWhenZero && normalizedAmount <= 0) {
    return ''
  }

  return normalizedAmount.toFixed(2)
}

function normalizeAmountValue(value: string, emptyWhenZero = false) {
  const amount = parseAmountValue(value)
  if (amount === undefined) {
    return ''
  }

  if (emptyWhenZero && amount <= 0) {
    return ''
  }

  return amount.toFixed(2)
}

function onAmountInput(field: ProductAmountField, event: any) {
  formData[field] = sanitizeAmountInput(getInputValue(event))
}

function onAmountBlur(field: ProductAmountField, emptyWhenZero = false) {
  formData[field] = normalizeAmountValue(formData[field], emptyWhenZero)
}

function onOptionPriceInput(specIndex: number, optIndex: number, event: any) {
  formData.specs[specIndex].options[optIndex].price = sanitizeAmountInput(getInputValue(event))
}

function onOptionPriceBlur(specIndex: number, optIndex: number) {
  formData.specs[specIndex].options[optIndex].price = normalizeAmountValue(
    formData.specs[specIndex].options[optIndex].price
  )
}

function normalizeProductSpecs(specs?: Product['specs']): ProductSpecForm[] {
  if (!Array.isArray(specs)) {
    return []
  }

  return specs.map(spec => ({
    name: spec.name || '',
    options: Array.isArray(spec.options)
      ? spec.options.map(option => ({
          name: option.name || '',
          price: formatAmountForInput(option.price)
        }))
      : []
  }))
}

function normalizeFormAmounts() {
  formData.price = normalizeAmountValue(formData.price)
  formData.original_price = normalizeAmountValue(formData.original_price, true)
  formData.specs.forEach((spec) => {
    spec.options.forEach((option) => {
      option.price = normalizeAmountValue(option.price)
    })
  })
}

function addSpec() {
  formData.specs.push({
    name: '',
    options: [{ name: '', price: '' }]
  })
}

function removeSpec(index: number) {
  formData.specs.splice(index, 1)
}

function addOption(specIndex: number) {
  formData.specs[specIndex].options.push({ name: '', price: '' })
}

function removeOption(specIndex: number, optIndex: number) {
  formData.specs[specIndex].options.splice(optIndex, 1)
}

function validateForm(): boolean {
  normalizeFormAmounts()

  if (!formData.name.trim()) {
    uni.showToast({ title: '请输入商品名称', icon: 'none' })
    return false
  }
  if (!formData.category_id) {
    uni.showToast({ title: '请选择商品分类', icon: 'none' })
    return false
  }
  const price = parseAmountValue(formData.price)
  if (price === undefined || price <= 0) {
    uni.showToast({ title: '请输入正确的售价', icon: 'none' })
    return false
  }
  const originalPrice = parseAmountValue(formData.original_price)
  if (formData.original_price && originalPrice === undefined) {
    uni.showToast({ title: '请输入正确的原价', icon: 'none' })
    return false
  }
  for (const spec of formData.specs) {
    for (const option of spec.options) {
      if (option.price && parseAmountValue(option.price) === undefined) {
        uni.showToast({ title: '请输入正确的规格加价', icon: 'none' })
        return false
      }
    }
  }
  if (formData.stock < 0) {
    uni.showToast({ title: '库存不能为负数', icon: 'none' })
    return false
  }
  return true
}

async function saveDraft() {
  if (!validateForm()) return
  await submitForm(false)
}

async function handleSubmit() {
  if (!validateForm()) return
  await submitForm(true)
}

async function submitForm(publish: boolean) {
  submitting.value = true

  try {
    const price = parseAmountValue(formData.price)
    if (price === undefined) {
      throw new Error('请输入正确的售价')
    }

    const originalPrice = parseAmountValue(formData.original_price)
    const specs = formData.specs
      .map(spec => ({
        name: spec.name.trim(),
        options: spec.options
          .map(option => ({
            name: option.name.trim(),
            price: parseAmountValue(option.price) ?? 0
          }))
          .filter(option => option.name)
      }))
      .filter(spec => spec.name && spec.options.length > 0)

    const data: ProductUpsertPayload = {
      name: formData.name.trim(),
      category_id: formData.category_id,
      description: formData.description.trim(),
      price,
      original_price: originalPrice,
      stock: formData.stock,
      unit: formData.unit.trim(),
      images: formData.images,
      specs,
      sort: formData.sort
    }

    if (productId.value) {
      await updateProduct(productId.value, data)
    } else {
      await createProduct(data)
    }

    uni.showToast({ title: publish ? '发布成功' : '保存成功', icon: 'success' })
    
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.edit-container {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.form-scroll {
  flex: 1;
  padding-bottom: 140rpx;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.section-header .section-title {
  margin-bottom: 0;
}

.add-spec-btn {
  font-size: 28rpx;
  color: #007AFF;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-row {
  display: flex;
  gap: 24rpx;
}

.form-item.half {
  flex: 1;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.required {
  color: #ff4d4f;
}

.form-input, .form-textarea {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.form-textarea {
  height: 160rpx;
  width: 100%;
  box-sizing: border-box;
}

.picker-value {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
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

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.image-item {
  position: relative;
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
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
  border-radius: 12rpx;
  border: 2rpx dashed #dddddd;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.add-image .icon {
  font-size: 64rpx;
  color: #cccccc;
  line-height: 1;
}

.add-image .text {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.spec-item {
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.spec-header {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}

.spec-name-input {
  flex: 1;
  height: 72rpx;
  background: #ffffff;
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
}

.delete-spec {
  margin-left: 16rpx;
  font-size: 26rpx;
  color: #ff4d4f;
}

.spec-option {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.option-name {
  flex: 1;
  height: 72rpx;
  background: #ffffff;
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  margin-right: 12rpx;
}

.option-price {
  width: 160rpx;
  height: 72rpx;
  background: #ffffff;
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  margin-right: 12rpx;
}

.delete-option {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ff4d4f;
  font-size: 32rpx;
}

.add-option {
  font-size: 28rpx;
  color: #007AFF;
  padding: 12rpx 0;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 24rpx;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-draft, .btn-publish {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-draft {
  background: #f5f5f5;
  color: #666666;
}

.btn-publish {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}

.btn-publish[disabled] {
  background: #cccccc;
}
</style>
