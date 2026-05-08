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
          :src="storeInfo?.merchant?.logo || '/static/default-logo.png'"
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
      <view class="store-status" :class="{ closed: storeInfo?.merchant?.status === 'closed' }">
        {{ storeInfo?.merchant?.status === 'open' ? '营业中' : '休息中' }}
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
                :src="product.image || '/static/default-product.png'"
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

    <!-- 底部购物栏 -->
    <view class="cart-bar" v-if="cartStore.totalCount > 0" @click="goCart">
      <view class="cart-icon">
        <text class="cart-badge">{{ cartStore.totalCount }}</text>
      </view>
      <view class="cart-info">
        <text class="cart-amount">¥{{ cartStore.totalAmount.toFixed(2) }}</text>
      </view>
      <view class="cart-btn">去购物车</view>
    </view>

    <!-- 底部占位 -->
    <view class="bottom-placeholder"></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getStoreHome, getStoreProducts } from '../../api'
import { useCartStore } from '../../stores/cart'
import type { StoreHomeInfo, Product } from '../../types/index'

const cartStore = useCartStore()

const storeInfo = ref<StoreHomeInfo | null>(null)
const currentCategoryIndex = ref(0)
const currentProducts = ref<Product[]>([])
const loadingProducts = ref(false)

const currentCategory = computed(() => {
  return storeInfo.value?.categories?.[currentCategoryIndex.value] || null
})

onShow(() => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const merchantId = currentPage?.options?.merchant_id || 1

  loadStoreHome(merchantId)
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
    
    const categoryData = res.find(item => item.category.id === categoryId)
    currentProducts.value = categoryData?.products || []
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

function goProductDetail(productId: number) {
  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/product?merchant_id=${merchantId}&product_id=${productId}`
  })
}

function addToCart(product: any) {
  cartStore.addItem({
    product_id: product.id,
    product_name: product.name,
    image: product.images?.[0] || '',
    price: product.price,
    quantity: 1,
    max_stock: product.stock
  })

  uni.showToast({
    title: '已加入购物车',
    icon: 'success'
  })
}

function goCart() {
  const merchantId = storeInfo.value?.merchant?.id || 1
  uni.navigateTo({
    url: `/pages/store/cart?merchant_id=${merchantId}`
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

.main-content {
  display: flex;
  height: calc(100vh - 400rpx - 120rpx);
  background: #ffffff;
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
  height: 100rpx;
  background: #1a1a1a;
  display: flex;
  align-items: center;
  padding: 0 32rpx;
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 100;
}

.cart-icon {
  width: 100rpx;
  height: 100rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -40rpx;
  position: relative;
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
  flex: 1;
  margin-left: 24rpx;
}

.cart-amount {
  font-size: 36rpx;
  font-weight: 600;
  color: #ffffff;
}

.cart-btn {
  padding: 16rpx 40rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 40rpx;
  font-size: 28rpx;
  color: #ffffff;
}

.bottom-placeholder {
  height: 120rpx;
}
</style>
