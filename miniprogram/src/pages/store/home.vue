<template>
  <view class="store-home-container">
    <!-- 店铺头部 -->
    <view class="store-header">
      <image
        class="store-banner"
        :src="storeInfo?.merchant?.images?.[0] || '/static/store-banner.png'"
        mode="aspectFill"
      />
      <view class="store-mask"></view>
      <view class="store-info">
        <image
          class="store-logo"
          :src="storeInfo?.merchant?.logo || BrandAsset.DEFAULT_MERCHANT_LOGO"
          mode="aspectFill"
        />
        <view class="store-detail">
          <view class="store-name">{{ storeInfo?.merchant?.name || '加载中...' }}</view>
          <view class="store-meta">
            <view class="rating" v-if="storeInfo?.merchant?.rating">
              <text class="stars">★★★★★</text>
              <text class="rating-value">{{ storeInfo?.merchant?.rating }}</text>
            </view>
            <text class="sales-count">已售 {{ storeInfo?.merchant?.sales_count || 0 }}</text>
          </view>
          <view class="store-notice" v-if="storeInfo?.merchant?.announcement">
            <text class="notice-icon">📢</text>
            <text class="notice-text">{{ storeInfo.merchant.announcement }}</text>
          </view>
        </view>
      </view>
      <view class="store-status" :class="{ closed: storeInfo?.merchant?.status !== 1 }">
        {{ storeInfo?.merchant?.status === 1 ? '营业中' : '休息中' }}
      </view>
    </view>

    <view v-if="storeInfo?.merchant?.status !== 1" class="rest-tip">
      当前店铺休息中，可继续浏览商品；新订单暂不支持提交。
    </view>

    <view class="quick-entry-bar">
      <view class="quick-entry-card" @click="goMyOrders">
        <view class="quick-entry-icon">📋</view>
        <view class="quick-entry-content">
          <view class="quick-entry-title">我的订单</view>
          <view class="quick-entry-desc">查看当前店铺订单与退款进度</view>
        </view>
        <view class="quick-entry-arrow">›</view>
      </view>
    </view>

    <!-- 分类和商品 -->
    <view class="main-content">
      <!-- 左侧分类 -->
      <scroll-view class="category-sidebar" scroll-y>
        <view
          v-for="(category, index) in storeInfo?.categories"
          :key="category.id"
          class="category-item"
          :class="{ active: currentCategoryIndex === index }"
          @click="selectCategory(index)"
        >
          <text class="category-name">{{ category.name }}</text>
          <text class="category-count" v-if="category.product_count">{{ category.product_count }}</text>
        </view>
      </scroll-view>

      <!-- 右侧商品 -->
      <scroll-view class="product-list" scroll-y @scrolltolower="loadMoreProducts">
        <!-- 分类标题 -->
        <view class="category-title" v-if="currentCategory">
          {{ currentCategory.name }}
        </view>

        <!-- 热销推荐 -->
        <view class="hot-products" v-if="!currentCategoryIndex && storeInfo?.hot_products?.length">
          <view class="hot-title">🔥 热销推荐</view>
          <view class="product-grid">
            <view
              v-for="product in storeInfo.hot_products"
              :key="product.id"
              class="product-card"
              @click="goProductDetail(product.id)"
            >
              <image
                class="product-image"
                :src="getHotProductImage(product)"
                mode="aspectFill"
              />
              <view class="product-info">
                <view class="product-name">{{ product.name }}</view>
                <view class="product-bottom">
                  <view class="product-price">
                    <text class="price">¥{{ product.price.toFixed(2) }}</text>
                    <text v-if="product.original_price" class="original-price">
                      ¥{{ product.original_price.toFixed(2) }}
                    </text>
                  </view>
                  <view class="add-btn" @click.stop="addToCart(product)">+</view>
                </view>
                <view class="product-sales">已售 {{ product.sales || 0 }}</view>
              </view>
            </view>
          </view>
        </view>

        <!-- 分类商品列表 -->
        <view class="product-list-items">
          <view
            v-for="product in currentProducts"
            :key="product.id"
            class="product-list-item"
            @click="goProductDetail(product.id)"
          >
            <image
              class="item-image"
              :src="product.images?.[0] || '/static/default-product.png'"
              mode="aspectFill"
            />
            <view class="item-info">
              <view class="item-name">{{ product.name }}</view>
              <view class="item-desc" v-if="product.description">{{ product.description }}</view>
              <view class="item-bottom">
                <view class="item-price">
                  <text class="price">¥{{ product.price.toFixed(2) }}</text>
                  <text v-if="product.original_price" class="original-price">
                    ¥{{ product.original_price.toFixed(2) }}
                  </text>
                </view>
                <view class="add-btn" @click.stop="addToCart(product)">+</view>
              </view>
            </view>
          </view>
        </view>

        <view v-if="loadingProducts" class="loading">加载中...</view>
      </scroll-view>
    </view>

    <view class="cart-bar">
      <view class="cart-area" :class="{ disabled: isCartEmpty }" @click="goCart">
        <view :class="isCartEmpty ? 'cart-icon-empty' : 'cart-icon'">
          <text>🛒</text>
          <view class="cart-badge" v-if="cartCount > 0">{{ cartCount > 99 ? '99+' : cartCount }}</view>
        </view>
        <view class="cart-info">
          <view class="cart-amount">
            {{ isCartEmpty ? '未选购商品' : `¥${cartAmount.toFixed(2)}` }}
          </view>
          <view class="cart-desc">
            {{ isCartEmpty ? '点击商品上的 + 加入购物车' : '点击购物车查看已选商品' }}
          </view>
        </view>
      </view>

      <view
        class="cart-btn"
        :class="{ disabled: isCartEmpty }"
        @click="handlePrimaryAction"
      >
        {{ primaryActionText }}
      </view>
    </view>

    <view v-if="showAddDialog" class="add-dialog-mask" @click="closeAddDialog">
      <view class="add-dialog" @click.stop>
        <view class="add-dialog-header">
          <view class="add-dialog-title">{{ addDialogProduct?.name || '选择规格' }}</view>
          <view class="add-dialog-close" @click="closeAddDialog">×</view>
        </view>

        <view v-if="addDialogLoading" class="add-dialog-loading">加载中...</view>

        <template v-else>
          <view class="add-dialog-price">
            <text class="price-label">价格</text>
            <text class="price-value">¥{{ addDialogSelectedPrice.toFixed(2) }}</text>
          </view>

          <view class="add-dialog-specs" v-if="addDialogProduct?.specs?.length">
            <view
              v-for="spec in addDialogProduct.specs"
              :key="spec.name"
              class="add-spec-group"
            >
              <view class="add-spec-name">{{ spec.name }}</view>
              <view class="add-spec-options">
                <view
                  v-for="option in spec.options"
                  :key="option.name"
                  class="add-spec-option"
                  :class="{
                    selected: addDialogSelectedSpecs[spec.name] === option.name,
                    disabled: option.stock === 0
                  }"
                  @click="selectAddDialogSpec(spec.name, option)"
                >
                  <text class="option-name">{{ option.name }}</text>
                  <text class="option-price">+¥{{ option.price.toFixed(2) }}</text>
                </view>
              </view>
            </view>
          </view>

          <view class="add-dialog-quantity">
            <view class="quantity-title">数量</view>
            <view class="quantity-control">
              <view
                class="quantity-btn"
                :class="{ disabled: addDialogQuantity <= 1 }"
                @click="decreaseAddDialogQuantity"
              >-</view>
              <text class="quantity-value">{{ addDialogQuantity }}</text>
              <view
                class="quantity-btn"
                :class="{ disabled: addDialogQuantity >= addDialogSelectedStock }"
                @click="increaseAddDialogQuantity"
              >+</view>
            </view>
            <view class="stock-tip">库存：{{ addDialogSelectedStock }}</view>
          </view>

          <view class="add-dialog-footer">
            <view class="add-dialog-confirm" @click="confirmAddDialog">加入购物车</view>
          </view>
        </template>
      </view>
    </view>

    <!-- 操作指引弹窗 -->
    <view v-if="showGuide" class="guide-dialog" @click="closeGuide">
      <view class="guide-content" @click.stop>
        <view class="guide-title">🛒 购物指南</view>
        <view class="guide-list">
          <view class="guide-item">
            <view class="guide-step">1️⃣</view>
            <view class="guide-text">选择心仪的商品，点击「+」加入购物车</view>
          </view>
          <view class="guide-item">
            <view class="guide-step">2️⃣</view>
            <view class="guide-text">点击「去购物车」查看已选商品</view>
          </view>
          <view class="guide-item">
            <view class="guide-step">3️⃣</view>
            <view class="guide-text">确认订单并完成支付</view>
          </view>
          <view class="guide-item">
            <view class="guide-step">4️⃣</view>
            <view class="guide-text">到店出示核销码或等待配送</view>
          </view>
        </view>
        <view class="guide-footer">
          <view class="guide-tip">💡 有任何问题？点击「我的订单」联系商家</view>
          <view class="guide-close" @click="closeGuide">知道了</view>
        </view>
      </view>
    </view>

    <!-- 底部占位 -->
    <view class="bottom-placeholder"></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getStoreHome, getStoreProducts, getStoreProduct } from '@api'
import { useCartStore } from '../../stores/cart'
import { useAnalytics } from '@utils/analytics'
import { useAuth } from '../../utils/useAuth'
import type { StoreHomeInfo, Product, SpecOption } from '@types'
import { BrandAsset } from '../../utils/constants'

const cartStore = useCartStore()
const { trackVisit, trackPageView } = useAnalytics()

const storeInfo = ref<StoreHomeInfo | null>(null)
const currentMerchantId = ref(1)
const currentCategoryIndex = ref(0)
const currentProducts = ref<Product[]>([])
const loadingProducts = ref(false)
const showGuide = ref(false) // 控制操作指引弹窗显示

const showAddDialog = ref(false)
const addDialogLoading = ref(false)
const addDialogProduct = ref<Product | null>(null)
const addDialogQuantity = ref(1)
const addDialogSelectedSpecs = reactive<Record<string, string>>({})

const currentCategory = computed(() => {
  return storeInfo.value?.categories?.[currentCategoryIndex.value] || null
})

const cartCount = computed(() => cartStore.totalCount)
const cartAmount = computed(() => cartStore.totalAmount)
const isCartEmpty = computed(() => cartStore.isEmpty)
const isStoreOpen = computed(() => storeInfo.value?.merchant?.status === 1)
const primaryActionText = computed(() => {
  if (isCartEmpty.value) {
    return '请选择商品'
  }

  return isStoreOpen.value ? '去结算' : '去购物车'
})

let showPromise: Promise<void> | null = null

onShow(() => {
  if (showPromise) return

  showPromise = (async () => {
    const pages = getCurrentPages()
    const currentPage = pages[pages.length - 1] as any
    const merchantId = Number(currentPage?.options?.merchant_id) || 1
    const source = currentPage?.options?.scene || 'scan'

    currentMerchantId.value = merchantId

    const guideKey = `storeHomeGuideShown:${merchantId}`
    showGuide.value = !uni.getStorageSync(guideKey)

    const { ensureAuth } = useAuth()
    await ensureAuth()

    await trackVisit({ merchant_id: merchantId, source })
    await trackPageView('store_home', merchantId, source)

    loadStoreHome(merchantId)
  })().finally(() => {
    showPromise = null
  })
})

async function loadStoreHome(merchantId: number) {
  try {
    const res = await getStoreHome(merchantId)
    storeInfo.value = res
    
    if (res.categories?.length) {
      loadCategoryProducts(res.categories[0].id)
    }
  } catch (error) {
    console.error('加载店铺信息失败:', error)
    uni.showToast({ title: '加载失败', icon: 'none' })
  }
}

async function loadCategoryProducts(categoryId: number) {
  loadingProducts.value = true

  try {
    const merchantId = storeInfo.value?.merchant?.id || 1
    const res = await getStoreProducts(merchantId, { category_id: categoryId })
    currentProducts.value = res.list || []
  } catch (error) {
    console.error('加载商品失败:', error)
  } finally {
    loadingProducts.value = false
  }
}

function selectCategory(index: number) {
  currentCategoryIndex.value = index
  
  if (storeInfo.value?.categories?.[index]) {
    loadCategoryProducts(storeInfo.value.categories[index].id)
  }
}

function loadMoreProducts() {
  // 加载更多逻辑
}

function getHotProductImage(product: any) {
  if (Array.isArray(product?.images) && product.images.length > 0) {
    return product.images[0]
  }

  return product?.image || '/static/default-product.png'
}

function goProductDetail(productId: number) {
  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/product?merchant_id=${merchantId}&product_id=${productId}`
  })
}

async function addToCart(product: any) {
  const merchantId = storeInfo.value?.merchant?.id || currentMerchantId.value
  currentMerchantId.value = merchantId

  showAddDialog.value = true
  addDialogLoading.value = true
  addDialogProduct.value = null
  addDialogQuantity.value = 1
  Object.keys(addDialogSelectedSpecs).forEach((key) => {
    delete addDialogSelectedSpecs[key]
  })

  try {
    const detail = await getStoreProduct(merchantId, product.id)
    addDialogProduct.value = detail

    if (detail.specs?.length) {
      for (const spec of detail.specs) {
        if (spec.options?.length) {
          addDialogSelectedSpecs[spec.name] = spec.options[0].name
        }
      }
    }
  } catch (error) {
    uni.showToast({ title: '加载商品失败', icon: 'none' })
    closeAddDialog()
  } finally {
    addDialogLoading.value = false
  }
}

const addDialogSelectedPrice = computed(() => {
  if (!addDialogProduct.value) return 0

  let price = addDialogProduct.value.price

  if (addDialogProduct.value.specs?.length) {
    for (const spec of addDialogProduct.value.specs) {
      const selectedName = addDialogSelectedSpecs[spec.name]
      const option = spec.options?.find(o => o.name === selectedName)
      if (option) {
        price += option.price
      }
    }
  }

  return price
})

const addDialogSelectedStock = computed(() => {
  if (!addDialogProduct.value) return 0

  if (addDialogProduct.value.specs?.length) {
    for (const spec of addDialogProduct.value.specs) {
      const selectedName = addDialogSelectedSpecs[spec.name]
      const option = spec.options?.find(o => o.name === selectedName)
      if (option) {
        return option.stock || addDialogProduct.value.stock
      }
    }
  }

  return addDialogProduct.value.stock
})

function selectAddDialogSpec(specName: string, option: SpecOption) {
  if (option.stock === 0) return
  addDialogSelectedSpecs[specName] = option.name
  if (addDialogQuantity.value > addDialogSelectedStock.value) {
    addDialogQuantity.value = addDialogSelectedStock.value
  }
}

function decreaseAddDialogQuantity() {
  if (addDialogQuantity.value > 1) {
    addDialogQuantity.value -= 1
  }
}

function increaseAddDialogQuantity() {
  if (addDialogQuantity.value < addDialogSelectedStock.value) {
    addDialogQuantity.value += 1
  }
}

function getAddDialogSpecString(): string {
  const specs: string[] = []
  for (const spec of addDialogProduct.value?.specs || []) {
    if (addDialogSelectedSpecs[spec.name]) {
      specs.push(addDialogSelectedSpecs[spec.name])
    }
  }
  return specs.join('/')
}

function confirmAddDialog() {
  if (!addDialogProduct.value) return
  if (addDialogSelectedStock.value <= 0) {
    return uni.showToast({ title: '库存不足', icon: 'none' })
  }

  const merchantId = storeInfo.value?.merchant?.id || currentMerchantId.value
  const merchantName = storeInfo.value?.merchant?.name || ''

  cartStore.addItem({
    merchant_id: merchantId,
    merchant_name: merchantName,
    product_id: addDialogProduct.value.id,
    product_name: addDialogProduct.value.name,
    image: addDialogProduct.value.images?.[0] || '',
    price: addDialogSelectedPrice.value,
    quantity: addDialogQuantity.value,
    specs: getAddDialogSpecString(),
    max_stock: addDialogSelectedStock.value
  })

  uni.showToast({ title: '已加入购物车', icon: 'success' })
  closeAddDialog()
}

function closeAddDialog() {
  showAddDialog.value = false
}

function goCart() {
  if (isCartEmpty.value) {
    uni.showToast({ title: '请先选择商品', icon: 'none' })
    return
  }

  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/cart?merchant_id=${merchantId}`
  })
}

function handlePrimaryAction() {
  if (isCartEmpty.value) {
    uni.showToast({ title: '请先选择商品', icon: 'none' })
    return
  }

  if (!isStoreOpen.value) {
    goCart()
    return
  }

  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/confirm?merchant_id=${merchantId}`
  })
}

function closeGuide() {
  showGuide.value = false

  const guideKey = `storeHomeGuideShown:${currentMerchantId.value}`
  uni.setStorageSync(guideKey, true)
}

function goMyOrders() {
  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/my-orders?merchant_id=${merchantId}`
  })
}
</script>

<style scoped>
.store-home-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.store-header {
  position: relative;
  height: 400rpx;
}

.store-banner {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.store-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to bottom, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0.7) 100%);
}

.store-info {
  position: absolute;
  bottom: 60rpx;
  left: 32rpx;
  right: 32rpx;
  display: flex;
  color: #ffffff;
}

.store-logo {
  width: 140rpx;
  height: 140rpx;
  border-radius: 16rpx;
  background: #ffffff;
  margin-right: 24rpx;
  border: 4rpx solid #ffffff;
}

.store-detail {
  flex: 1;
}

.store-name {
  font-size: 40rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.store-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.stars {
  color: #ffd700;
  font-size: 24rpx;
}

.rating-value {
  font-size: 24rpx;
  margin-left: 8rpx;
}

.sales-count {
  font-size: 24rpx;
  opacity: 0.9;
}

.store-notice {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  opacity: 0.9;
}

.notice-icon {
  margin-right: 8rpx;
}

.store-status {
  position: absolute;
  top: 32rpx;
  right: 32rpx;
  padding: 8rpx 24rpx;
  background: #52c41a;
  border-radius: 32rpx;
  font-size: 24rpx;
  color: #ffffff;
}

.store-status.closed {
  background: #999999;
}

.rest-tip {
  margin: 24rpx 24rpx 0;
  padding: 20rpx 24rpx;
  background: #fff7e6;
  border-radius: 16rpx;
  color: #d46b08;
  font-size: 26rpx;
}

.quick-entry-bar {
  padding: 20rpx 24rpx 0;
}

.quick-entry-card {
  display: flex;
  align-items: center;
  padding: 24rpx;
  background: #ffffff;
  border-radius: 24rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.05);
}

.quick-entry-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: 20rpx;
  background: rgba(0, 122, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34rpx;
  margin-right: 20rpx;
}

.quick-entry-content {
  flex: 1;
}

.quick-entry-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 6rpx;
}

.quick-entry-desc {
  font-size: 24rpx;
  color: #666666;
}

.quick-entry-arrow {
  font-size: 40rpx;
  color: #c0c4cc;
  margin-left: 16rpx;
}

.main-content {
  display: flex;
  height: calc(100vh - 400rpx - 120rpx - 116rpx);
  background: #ffffff;
  margin-top: 24rpx;
}

.category-sidebar {
  width: 180rpx;
  background: #f8f9fa;
}

.category-item {
  padding: 32rpx 24rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 26rpx;
  color: #666666;
  border-left: 6rpx solid transparent;
}

.category-item.active {
  background: #ffffff;
  color: #007AFF;
  border-left-color: #007AFF;
}

.category-name {
  margin-bottom: 8rpx;
}

.category-count {
  font-size: 22rpx;
  background: #f0f0f0;
  padding: 2rpx 12rpx;
  border-radius: 12rpx;
}

.product-list {
  flex: 1;
  padding: 24rpx;
}

.category-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.hot-title {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 20rpx;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
  margin-bottom: 32rpx;
}

.product-card {
  background: #f8f9fa;
  border-radius: 16rpx;
  overflow: hidden;
}

.product-image {
  width: 100%;
  height: 300rpx;
  background: #e0e0e0;
}

.product-info {
  padding: 16rpx;
}

.product-name {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 12rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8rpx;
}

.product-price .price {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.original-price {
  font-size: 22rpx;
  color: #999999;
  text-decoration: line-through;
  margin-left: 8rpx;
}

.add-btn {
  width: 48rpx;
  height: 48rpx;
  background: #007AFF;
  border-radius: 50%;
  color: #ffffff;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-sales {
  font-size: 22rpx;
  color: #999999;
}

.product-list-items {
  display: flex;
  flex-direction: column;
}

.product-list-item {
  display: flex;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.product-list-item:last-child {
  border-bottom: none;
}

.item-image {
  width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.item-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
}

.item-desc {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-price .price {
  font-size: 32rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.loading {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.cart-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 120rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  padding: 0 32rpx;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
  z-index: 100;
}

.cart-area {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex: 1;
}

.cart-area.disabled {
  opacity: 0.7;
}

.cart-icon {
  width: 100rpx;
  height: 100rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  color: #ffffff;
  font-size: 48rpx;
}

.cart-icon-empty {
  width: 100rpx;
  height: 100rpx;
  background: #f0f2f5;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  color: #8c8c8c;
  font-size: 48rpx;
}

.cart-badge {
  position: absolute;
  top: -10rpx;
  right: -10rpx;
  min-width: 36rpx;
  height: 36rpx;
  background: #ff4d4f;
  color: #ffffff;
  border-radius: 18rpx;
  font-size: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8rpx;
}

.cart-info {
  margin-left: 8rpx;
}

.cart-amount {
  font-size: 36rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.cart-desc {
  font-size: 22rpx;
  color: #666666;
  margin-top: 4rpx;
}

.cart-btn {
  min-width: 200rpx;
  height: 80rpx;
  padding: 0 32rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 40rpx;
  font-size: 30rpx;
  font-weight: 500;
  color: #ffffff;
  margin-left: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cart-btn.disabled {
  background: #d9d9d9;
  color: #ffffff;
}

.add-dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 1100;
}

.add-dialog {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx 24rpx 0 0;
  padding: 28rpx 28rpx calc(28rpx + env(safe-area-inset-bottom));
}

.add-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.add-dialog-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  flex: 1;
  padding-right: 24rpx;
}

.add-dialog-close {
  width: 56rpx;
  height: 56rpx;
  border-radius: 28rpx;
  background: #f0f2f5;
  color: #333333;
  font-size: 40rpx;
  line-height: 56rpx;
  text-align: center;
}

.add-dialog-loading {
  padding: 40rpx 0;
  text-align: center;
  color: #666666;
  font-size: 28rpx;
}

.add-dialog-price {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 0;
}

.price-label {
  font-size: 26rpx;
  color: #666666;
}

.price-value {
  font-size: 36rpx;
  font-weight: 700;
  color: #ff4d4f;
}

.add-dialog-specs {
  margin-top: 8rpx;
}

.add-spec-group {
  margin-top: 20rpx;
}

.add-spec-name {
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 14rpx;
}

.add-spec-options {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}

.add-spec-option {
  padding: 16rpx 22rpx;
  background: #f5f5f5;
  border-radius: 14rpx;
  border: 2rpx solid transparent;
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.add-spec-option.selected {
  background: rgba(0, 122, 255, 0.1);
  border-color: #007AFF;
}

.add-spec-option.disabled {
  opacity: 0.5;
}

.option-name {
  font-size: 26rpx;
  color: #1a1a1a;
}

.option-price {
  font-size: 22rpx;
  color: #666666;
}

.add-dialog-quantity {
  margin-top: 28rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.quantity-title {
  font-size: 28rpx;
  color: #333333;
}

.quantity-control {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.quantity-btn {
  width: 64rpx;
  height: 64rpx;
  background: #f5f5f5;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36rpx;
  color: #666666;
}

.quantity-btn.disabled {
  opacity: 0.5;
}

.quantity-value {
  min-width: 60rpx;
  text-align: center;
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.stock-tip {
  font-size: 22rpx;
  color: #999999;
}

.add-dialog-footer {
  margin-top: 28rpx;
}

.add-dialog-confirm {
  height: 88rpx;
  border-radius: 44rpx;
  background: linear-gradient(135deg, #ff9500 0%, #ff5e3a 100%);
  color: #ffffff;
  font-size: 32rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 操作指引弹窗样式 */
.guide-dialog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.guide-content {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 40rpx;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { transform: translateY(50rpx); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.guide-title {
  font-size: 36rpx;
  font-weight: 600;
  color: #1a1a1a;
  text-align: center;
  margin-bottom: 32rpx;
}

.guide-list {
  margin-bottom: 32rpx;
}

.guide-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: 24rpx;
  padding: 20rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
}

.guide-step {
  font-size: 40rpx;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.guide-text {
  font-size: 28rpx;
  color: #333333;
  line-height: 1.6;
  flex: 1;
}

.guide-footer {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
}

.guide-tip {
  font-size: 24rpx;
  color: #666666;
  text-align: center;
  margin-bottom: 24rpx;
}

.guide-close {
  padding: 24rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  border-radius: 12rpx;
  font-size: 32rpx;
  font-weight: 500;
  text-align: center;
}

.bottom-placeholder {
  height: 160rpx;
}
</style>
