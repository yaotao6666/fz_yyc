<template>
  <view class="category-container">
    <!-- 分类列表 -->
    <view class="category-list">
      <view
        v-for="(category, index) in categories"
        :key="category.id"
        class="category-item"
        :class="{ active: currentIndex === index }"
        @click="selectCategory(index)"
      >
        <text class="category-name">{{ category.name }}</text>
        <text class="product-count">({{ category.product_count || 0 }})</text>
      </view>
      <view class="category-item add" @click="showAddDialog">
        <text class="add-icon">+</text>
        <text class="category-name">添加分类</text>
      </view>
    </view>

    <!-- 商品列表 -->
    <scroll-view class="product-scroll" scroll-y @scrolltolower="loadMore">
      <view class="product-grid">
        <view
          v-for="product in products"
          :key="product.id"
          class="product-card"
          @click="goEdit(product.id)"
        >
          <image
            class="product-image"
            :src="product.images?.[0] || '/static/default-product.png'"
            mode="aspectFill"
          />
          <view class="product-info">
            <view class="product-name">{{ product.name }}</view>
            <view class="product-price">¥{{ product.price.toFixed(2) }}</view>
            <view class="product-meta">
              <text class="stock">库存: {{ product.stock }}</text>
              <text class="sales">销量: {{ product.sales || 0 }}</text>
            </view>
          </view>
          <view class="product-status" :class="{ off: product.status === 0 }">
            {{ product.status === 1 ? '上架' : '下架' }}
          </view>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="noMore && products.length > 0" class="no-more">没有更多了</view>
      <view v-if="!loading && products.length === 0" class="empty">
        <text class="empty-icon">📦</text>
        <text class="empty-text">暂无商品</text>
        <button class="btn-add-product" @click="goAddProduct">添加商品</button>
      </view>
    </scroll-view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <button class="btn-add" @click="goAddProduct">添加商品</button>
    </view>

    <!-- 添加/编辑分类弹窗 -->
    <view v-if="showDialog" class="dialog-mask" @click="closeDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">{{ editingCategory ? '编辑分类' : '添加分类' }}</view>
        <input
          v-model="categoryName"
          class="dialog-input"
          placeholder="请输入分类名称"
          focus
        />
        <view class="dialog-actions">
          <button class="btn-cancel" @click="closeDialog">取消</button>
          <button class="btn-confirm" @click="saveCategory">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive, onShow } from 'vue'
import { getCategories, createCategory, updateCategory, deleteCategory, getProducts } from '../../api'
import type { Category, Product } from '../../types/api'

const categories = ref<Category[]>([])
const products = ref<Product[]>([])
const currentIndex = ref(0)
const loading = ref(false)
const noMore = ref(false)
const page = ref(1)
const pageSize = 10

const showDialog = ref(false)
const categoryName = ref('')
const editingCategory = ref<Category | null>(null)

onShow(() => {
  loadCategories()
})

async function loadCategories() {
  try {
    const res = await getCategories()
    categories.value = res
    if (res.length > 0) {
      loadProducts(res[0].id)
    }
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

async function loadProducts(categoryId?: number, reset = false) {
  if (reset) {
    page.value = 1
    noMore.value = false
    products.value = []
  }

  if (noMore.value || loading.value) return

  loading.value = true

  try {
    const res = await getProducts({
      page: page.value,
      page_size: pageSize,
      category_id: categoryId
    })

    if (reset) {
      products.value = res.list
    } else {
      products.value.push(...res.list)
    }

    if (res.list.length < pageSize) {
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

function selectCategory(index: number) {
  currentIndex.value = index
  loadProducts(categories.value[index].id, true)
}

function loadMore() {
  if (categories.value[currentIndex.value]) {
    loadProducts(categories.value[currentIndex.value].id)
  }
}

function showAddDialog() {
  editingCategory.value = null
  categoryName.value = ''
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  categoryName.value = ''
  editingCategory.value = null
}

async function saveCategory() {
  if (!categoryName.value.trim()) {
    uni.showToast({ title: '请输入分类名称', icon: 'none' })
    return
  }

  try {
    if (editingCategory.value) {
      await updateCategory(editingCategory.value.id, { name: categoryName.value })
      const index = categories.value.findIndex(c => c.id === editingCategory.value!.id)
      if (index !== -1) {
        categories.value[index].name = categoryName.value
      }
      uni.showToast({ title: '修改成功', icon: 'success' })
    } else {
      const newCategory = await createCategory({ name: categoryName.value })
      categories.value.push(newCategory)
      uni.showToast({ title: '添加成功', icon: 'success' })
    }
    closeDialog()
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

function goEdit(productId: number) {
  uni.navigateTo({ url: `/pages/merchant/products/edit?id=${productId}` })
}

function goAddProduct() {
  const categoryId = categories.value[currentIndex.value]?.id
  uni.navigateTo({ url: `/pages/merchant/products/edit?category_id=${categoryId}` })
}
</script>

<style scoped>
.category-container {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
}

.category-list {
  width: 200rpx;
  background: #ffffff;
  padding: 24rpx 0;
}

.category-item {
  padding: 28rpx 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #666666;
  border-left: 6rpx solid transparent;
}

.category-item.active {
  background: #f0f5ff;
  color: #007AFF;
  border-left-color: #007AFF;
}

.category-item.add {
  color: #999999;
  margin-top: 24rpx;
  border-top: 1rpx solid #f0f0f0;
  padding-top: 28rpx;
}

.add-icon {
  font-size: 32rpx;
  margin-right: 8rpx;
}

.product-scroll {
  flex: 1;
  padding: 24rpx;
}

.product-grid {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.product-card {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 20rpx;
  display: flex;
  position: relative;
}

.product-image {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-price {
  font-size: 32rpx;
  color: #ff4d4f;
  font-weight: 600;
}

.product-meta {
  display: flex;
  gap: 16rpx;
  font-size: 24rpx;
  color: #999999;
}

.product-status {
  position: absolute;
  top: 20rpx;
  right: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  background: #f6ffed;
  color: #52c41a;
}

.product-status.off {
  background: #f5f5f5;
  color: #999999;
}

.loading, .no-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
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

.btn-add-product {
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
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-add {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.dialog-content {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 32rpx;
  text-align: center;
}

.dialog-input {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  margin-bottom: 32rpx;
}

.dialog-actions {
  display: flex;
  gap: 24rpx;
}

.btn-cancel, .btn-confirm {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666666;
}

.btn-confirm {
  background: #007AFF;
  color: #ffffff;
}
</style>
