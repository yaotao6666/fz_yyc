<template>
  <view class="product-list-container">
    <!-- 筛选栏 -->
    <view class="filter-bar">
      <view class="filter-row">
        <picker
          mode="selector"
          :range="categoryOptions"
          range-key="name"
          :value="selectedCategoryIndex"
          @change="onCategoryChange"
          class="category-picker"
        >
          <view class="picker-value">
            <text>{{ selectedCategoryLabel }}</text>
            <text class="arrow">▼</text>
          </view>
        </picker>
        <view class="search-box">
          <input
            v-model="keyword"
            class="search-input"
            placeholder="搜索商品名称"
            @confirm="handleSearch"
          />
          <text class="search-btn" @click="handleSearch">搜索</text>
        </view>
      </view>
      <view class="filter-tabs">
        <view
          class="tab-item"
          :class="{ active: filterStatus === '' }"
          @click="changeFilter('')"
        >
          全部
        </view>
        <view
          class="tab-item"
          :class="{ active: filterStatus === 'on_sale' }"
          @click="changeFilter('on_sale')"
        >
          上架
        </view>
        <view
          class="tab-item"
          :class="{ active: filterStatus === 'off_sale' }"
          @click="changeFilter('off_sale')"
        >
          下架
        </view>
      </view>
    </view>

    <!-- 商品列表 -->
    <scroll-view
      class="product-list"
      scroll-y
      :lower-threshold="80"
      @scrolltolower="loadMore"
    >
      <checkbox-group @change="onCheckboxChange">
        <view
          v-for="product in products"
          :key="product.id"
          class="product-item"
        >
          <checkbox
            :value="String(product.id)"
            :checked="selectedIds.includes(product.id)"
            class="product-checkbox"
          />
          <image
            class="product-image"
            :src="product.images?.[0] || '/static/default-product.png'"
            mode="aspectFill"
            @click="goDetail(product.id)"
          />
          <view class="product-content" @click="goEdit(product.id)">
            <view class="product-name">{{ product.name }}</view>
            <view class="product-price">
              <text class="current-price">¥{{ product.price.toFixed(2) }}</text>
              <text v-if="Number(product.original_price || 0) > 0" class="original-price">
                ¥{{ Number(product.original_price || 0).toFixed(2) }}
              </text>
            </view>
            <view class="product-meta">
              <text>库存: {{ product.stock }}</text>
              <text>销量: {{ product.sales || 0 }}</text>
            </view>
          </view>
          <view class="product-actions">
            <view
              class="action-btn"
              :class="{ 'off-line': product.status === 1 }"
              @click.stop="toggleStatus(product)"
            >
              {{ product.status === 1 ? '下架' : '上架' }}
            </view>
            <view class="action-btn delete" @click.stop="deleteProduct(product)">删除</view>
          </view>
        </view>
      </checkbox-group>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && products.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && products.length === 0" class="empty">
        <text class="empty-icon">📦</text>
        <text class="empty-text">暂无商品</text>
        <button class="btn-add" @click="goAdd">添加商品</button>
      </view>
      <view v-if="products.length > 0" class="list-bottom-spacer"></view>
    </scroll-view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar" v-if="products.length > 0">
      <view class="select-all" @click="toggleSelectAll">
        <checkbox :checked="isAllSelected" />
        <text>全选</text>
      </view>
      <view class="batch-actions">
        <view class="batch-btn" :class="{ disabled: selectedIds.length === 0 }" @click="batchOnSale">
          批量上架
        </view>
        <view class="batch-btn" :class="{ disabled: selectedIds.length === 0 }" @click="batchOffSale">
          批量下架
        </view>
      </view>
      <button class="btn-add-product" @click="goAdd">添加</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getProducts, getCategories, productOnSale, productOffSale, deleteProduct as deleteProductApi, batchUpdateProductStatus } from '@api'
import type { Product, Category } from '@types'

const keyword = ref('')
const filterStatus = ref('')
const products = ref<Product[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10
const total = ref(0)

// 分类相关
const categories = ref<Category[]>([])
const selectedCategoryId = ref<number | null>(null)
const selectedCategoryIndex = ref(0)

const categoryOptions = computed(() => {
  return [
    { id: null, name: '全部分类' },
    ...categories.value.map(c => ({ id: c.id, name: c.name }))
  ]
})

const selectedCategoryLabel = computed(() => {
  return categoryOptions.value[selectedCategoryIndex.value]?.name || '全部分类'
})

const isAllSelected = computed(() => {
  return products.value.length > 0 && selectedIds.value.length === products.value.length
})

onLoad((options: any) => {
  const status = String(options?.status || '')
  if (status === 'on_sale' || status === 'off_sale') {
    filterStatus.value = status
  }
})

onShow(() => {
  loadCategories()
  loadProducts(true)
})

async function loadCategories() {
  try {
    const data = await getCategories()
    categories.value = data || []
  } catch (error) {
    console.error('加载分类失败:', error)
    categories.value = []
  }
}

async function loadProducts(reset = false) {
  if (reset) {
    page.value = 1
    noMore.value = false
    products.value = []
    total.value = 0
  }

  if (noMore.value || loading.value) return

  loading.value = true

  try {
    const params: any = {
      page: page.value,
      page_size: pageSize,
    }
    
    if (selectedCategoryId.value !== null) {
      params.category_id = selectedCategoryId.value
    }
    
    if (filterStatus.value) {
      params.status = filterStatus.value === 'on_sale' ? 1 : 2
    }
    
    if (keyword.value) {
      params.keyword = keyword.value
    }

    const res = await getProducts(params)

    if (reset) {
      products.value = res.list
    } else {
      products.value.push(...res.list)
    }

    total.value = Number(res.pagination?.total || 0)
    const pageSizeValue = Number(res.pagination?.page_size || pageSize)
    const loadedCount = products.value.length
    const reachedEndByTotal = total.value > 0 && loadedCount >= total.value
    const reachedEndByPageSize = res.list.length < pageSizeValue

    if (reachedEndByTotal || reachedEndByPageSize) {
      noMore.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载商品失败:', error)
  } finally {
    loading.value = false
  }
}

function loadMore() {
  loadProducts()
}

function handleSearch() {
  loadProducts(true)
}

function changeFilter(status: string) {
  filterStatus.value = status
  loadProducts(true)
}

function onCategoryChange(e: any) {
  const index = e.detail.value
  selectedCategoryIndex.value = index
  selectedCategoryId.value = categoryOptions.value[index]?.id || null
  loadProducts(true)
}

function onCheckboxChange(e: any) {
  selectedIds.value = e.detail.value.map((id: string) => Number(id))
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = products.value.map(p => p.id)
  }
}

async function toggleStatus(product: Product) {
  try {
    if (product.status === 1) {
      await productOffSale(product.id)
      product.status = 2
    } else {
      await productOnSale(product.id)
      product.status = 1
    }
    uni.showToast({ title: '操作成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

async function deleteProduct(product: Product) {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除商品"${product.name}"吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteProductApi(product.id)
          products.value = products.value.filter(p => p.id !== product.id)
          selectedIds.value = selectedIds.value.filter(id => id !== product.id)
          uni.showToast({ title: '删除成功', icon: 'success' })
        } catch (error: any) {
          if (error.message?.includes('不存在') || error.message?.includes('已删除')) {
            loadProducts(true)
          }
          uni.showToast({ title: error.message || '删除失败', icon: 'none' })
        }
      }
    }
  })
}

async function batchOnSale() {
  if (selectedIds.value.length === 0) return

  try {
    // on_sale 对应数字 1，off_sale 对应数字 2
    await batchUpdateProductStatus(selectedIds.value, 1)
    products.value.forEach(p => {
      if (selectedIds.value.includes(p.id)) {
        p.status = 1
      }
    })
    selectedIds.value = []
    uni.showToast({ title: '上架成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

async function batchOffSale() {
  if (selectedIds.value.length === 0) return

  try {
    await batchUpdateProductStatus(selectedIds.value, 2)
    products.value.forEach(p => {
      if (selectedIds.value.includes(p.id)) {
        p.status = 2
      }
    })
    selectedIds.value = []
    uni.showToast({ title: '下架成功', icon: 'success' })
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/merchant/products/edit?id=${id}` })
}

function goEdit(id: number) {
  uni.navigateTo({ url: `/pages/merchant/products/edit?id=${id}` })
}

function goAdd() {
  uni.navigateTo({ url: '/pages/merchant/products/edit' })
}
</script>

<style scoped>
.product-list-container {
  height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.filter-bar {
  background: #ffffff;
  padding: 24rpx;
  flex-shrink: 0;
}

.filter-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.category-picker {
  width: 200rpx;
  height: 72rpx;
  background: #f8f9fa;
  border-radius: 36rpx;
  display: flex;
  align-items: center;
  padding: 0 24rpx;
}

.picker-value {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  font-size: 28rpx;
  color: #1a1a1a;
}

.picker-value .arrow {
  font-size: 20rpx;
  color: #999999;
  margin-left: 8rpx;
}

.search-box {
  flex: 1;
  display: flex;
  gap: 16rpx;
}

.search-input {
  flex: 1;
  height: 72rpx;
  background: #f8f9fa;
  border-radius: 36rpx;
  padding: 0 32rpx;
  font-size: 28rpx;
}

.search-btn {
  width: 120rpx;
  height: 72rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 36rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
}

.filter-tabs {
  display: flex;
  gap: 24rpx;
}

.tab-item {
  padding: 12rpx 24rpx;
  font-size: 28rpx;
  color: #666666;
  border-radius: 24rpx;
}

.tab-item.active {
  background: #e6f0ff;
  color: #007AFF;
}

.product-list {
  flex: 1;
  min-height: 0;
  width: 100%;
  box-sizing: border-box;
  padding: 24rpx;
}

.product-item {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: flex-start;
}

.product-checkbox {
  margin-right: 16rpx;
  margin-top: 60rpx;
}

.product-image {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.product-content {
  flex: 1;
}

.product-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  margin-bottom: 12rpx;
}

.product-price {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-bottom: 12rpx;
}

.current-price {
  font-size: 32rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.original-price {
  font-size: 24rpx;
  color: #999999;
  text-decoration: line-through;
}

.product-meta {
  display: flex;
  gap: 16rpx;
  font-size: 24rpx;
  color: #999999;
}

.product-actions {
  display: flex;
  gap: 12rpx;
  margin-left: 16rpx;
}

.action-btn {
  padding: 8rpx 20rpx;
  border-radius: 8rpx;
  font-size: 24rpx;
  background: #f5f5f5;
  color: #666666;
}

.action-btn.off-line {
  background: #fff7e6;
  color: #fa8c16;
}

.action-btn.delete {
  background: #fff1f0;
  color: #ff4d4f;
}

.loading, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.list-bottom-spacer {
  height: calc(140rpx + env(safe-area-inset-bottom));
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 100rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999999;
  margin-bottom: 32rpx;
}

.btn-add {
  padding: 20rpx 48rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 40rpx;
  font-size: 28rpx;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 100rpx;
  background: #ffffff;
  display: flex;
  align-items: center;
  padding: 0 24rpx;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.select-all {
  display: flex;
  align-items: center;
  gap: 8rpx;
  font-size: 28rpx;
  color: #666666;
}

.batch-actions {
  flex: 1;
  display: flex;
  gap: 16rpx;
  margin-left: 32rpx;
}

.batch-btn {
  padding: 12rpx 24rpx;
  border-radius: 8rpx;
  font-size: 26rpx;
  background: #f5f5f5;
  color: #666666;
}

.batch-btn.disabled {
  opacity: 0.5;
}

.btn-add-product {
  width: 160rpx;
  height: 72rpx;
  background: #007AFF;
  color: #ffffff;
  border-radius: 36rpx;
  font-size: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
