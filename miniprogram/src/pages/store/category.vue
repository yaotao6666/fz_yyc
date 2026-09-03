<template>
  <view class="category-container">
    <view v-if="loading && productSections.length === 0" class="page-state">
      <text class="state-title">正在加载商品分类</text>
    </view>

    <view v-else-if="!categories.length" class="page-state">
      <text class="state-title">暂无商品分类</text>
      <text class="state-desc">商家尚未配置分类，可去首页逛逛～</text>
      <view class="state-btn" @click="goHome">返回首页</view>
    </view>

    <template v-else>
      <view class="category-body">
        <!-- 左侧分类 -->
        <scroll-view class="sidebar" scroll-y :scroll-into-view="sidebarScrollId">
          <view
            v-for="(category, index) in categories"
            :key="category.id"
            :id="`sidebar-cat-${category.id}`"
            class="sidebar-item"
            :class="{ active: currentIndex === index }"
            @click="selectCategory(index)"
          >
            <text class="sidebar-name">{{ category.name }}</text>
          </view>
        </scroll-view>

        <!-- 右侧商品 -->
        <scroll-view
          class="goods"
          scroll-y
          :scroll-top="scrollViewTop"
          @scrolltolower="loadMore"
        >
          <template v-for="(section, sIdx) in productSections" :key="section.category_id">
            <view v-if="sIdx !== 0 || section.page === 1" class="goods-category-title">
              <text class="goods-category-name">{{ section.category_name }}</text>
            </view>
            <view v-if="section.list.length === 0 && section.loaded" class="goods-empty">
              <text class="goods-empty-text">该分类暂无在售商品</text>
            </view>
            <view
              v-for="product in section.list"
              :key="product.id"
              class="goods-item"
              @click="goProduct(product.id)"
            >
              <image class="goods-image" :src="getProductCover(product)" mode="aspectFill" />
              <view class="goods-info">
                <view class="goods-name">{{ product.name }}</view>
                <view class="goods-desc" v-if="product.description">{{ product.description }}</view>
                <view class="goods-price-row">
                  <template v-if="Number(product.sale_type) === 2">
                    <text class="goods-price">¥{{ Number(product.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(product.rental_unit) }}</text>
                    <text class="goods-rental-tag">租赁</text>
                  </template>
                  <template v-else>
                    <text class="goods-price">¥{{ product.price.toFixed(2) }}</text>
                  </template>
                </view>
              </view>
              <view class="add-btn" @click.stop="onProductQuickAdd(product)">+</view>
            </view>
          </template>

          <view v-if="loadingMore" class="goods-loading">加载中...</view>
          <view v-else-if="noMore" class="goods-no-more">没有更多了</view>
        </scroll-view>
      </view>
    </template>

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
                  <text v-if="option.price > 0" class="option-price">+¥{{ option.price.toFixed(2) }}</text>
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

    <view v-if="showAddSuccessTip" class="add-success-tip">
      {{ addSuccessText }}
    </view>
  </view>
</template>

<script setup lang="ts">
import { nextTick, ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getStoreHome, getStoreProducts, getStoreProduct } from '../../api/store'
import { useCartStore } from '../../stores/cart'
import type { Product, SpecOption } from '../../types'
import { BrandAsset } from '../../utils/constants'

const cartStore = useCartStore()

const categories = ref<Array<{ id: number; name: string; product_count?: number }>>([])
type ProductSection = {
  category_id: number
  category_name: string
  list: Product[]
  page: number
  page_size: number
  total: number
  loaded: boolean
}
const productSections = ref<ProductSection[]>([])
const currentIndex = ref(0)
const loading = ref(false)
const loadingMore = ref(false)
const noMore = ref(false)
const sidebarScrollId = ref('')
const scrollViewTop = ref(0)

// 加购弹窗相关状态
const showAddDialog = ref(false)
const addDialogLoading = ref(false)
const addDialogProduct = ref<Product | null>(null)
const addDialogQuantity = ref(1)
const addDialogSelectedSpecs = reactive<Record<string, string>>({})
const showAddSuccessTip = ref(false)
const addSuccessText = ref('')
let addSuccessTipTimer: ReturnType<typeof setTimeout> | null = null

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
        return option.stock ?? addDialogProduct.value.stock
      }
    }
  }
  return addDialogProduct.value.stock
})

onLoad(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const home = await getStoreHome()
    // 左侧只展示一级分类；右侧商品由后端按该分类子树过滤（含子孙分类），
    // 未设置分类的商品/服务天然被排除
    categories.value = (home.categories || []).filter((cat: any) => Number(cat.level) === 1)
    if (categories.value.length) {
      currentIndex.value = 0
      noMore.value = false
      productSections.value = []
      await loadCategoryPage(0, 1, true)
    }
  } catch (_error) {
    uni.showToast({ title: '加载失败，请重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function categoryHasMore(sec: ProductSection): boolean {
  return sec.page * sec.page_size < sec.total
}

async function loadCategoryPage(categoryIdx: number, page: number, isFresh: boolean) {
  const category = categories.value[categoryIdx]
  if (!category) {
    noMore.value = true
    return
  }
  loadingMore.value = !isFresh
  try {
    const PAGE_SIZE = 20
    const res = await getStoreProducts({ category_id: category.id, page, page_size: PAGE_SIZE } as any)
    const list = res.list || []
    const total = res.pagination?.total ?? list.length

    if (isFresh || page === 1) {
      const existsIndex = productSections.value.findIndex(s => s.category_id === category.id)
      const newSection: ProductSection = {
        category_id: category.id,
        category_name: category.name,
        list,
        page,
        page_size: PAGE_SIZE,
        total,
        loaded: true
      }
      if (existsIndex >= 0) {
        productSections.value.splice(existsIndex, 1, newSection)
      } else {
        productSections.value.push(newSection)
      }
    } else {
      // Append next page to existing section
      const section = productSections.value.find(s => s.category_id === category.id)
      if (section) {
        section.list = section.list.concat(list)
        section.page = page
      }
    }
  } catch (_error) {
    uni.showToast({ title: '加载商品失败', icon: 'none' })
  } finally {
    loadingMore.value = false
  }
}

async function selectCategory(index: number) {
  if (loadingMore.value) return
  currentIndex.value = index
  sidebarScrollId.value = `sidebar-cat-${categories.value[index]?.id}`

  // 如果该分类已经加载过，跳转到它所在的位置；否则替换掉后续已加载内容并从头加载
  const categoryID = categories.value[index].id
  const existingIdx = productSections.value.findIndex(s => s.category_id === categoryID)
  if (existingIdx >= 0) {
    // 截断到该分类，回滚滚动条
    productSections.value = productSections.value.slice(0, existingIdx + 1)
    noMore.value = false
    scrollViewTop.value = scrollViewTop.value === 0 ? 1 : 0
    await nextTick()
    scrollViewTop.value = 0
  } else {
    productSections.value = []
    noMore.value = false
    await loadCategoryPage(index, 1, true)
  }
}

async function loadMore() {
  if (loadingMore.value || noMore.value) return

  // 当前分类还有更多分页：加载下一页
  const curCategory = categories.value[currentIndex.value]
  if (curCategory) {
    const curSection = productSections.value.find(s => s.category_id === curCategory.id)
    if (curSection && categoryHasMore(curSection)) {
      await loadCategoryPage(currentIndex.value, curSection.page + 1, false)
      return
    }
  }

  // 当前分类已加载完毕：自动加载下一个分类
  const nextIndex = currentIndex.value + 1
  if (nextIndex >= categories.value.length) {
    noMore.value = true
    return
  }

  currentIndex.value = nextIndex
  sidebarScrollId.value = `sidebar-cat-${categories.value[nextIndex].id}`
  await loadCategoryPage(nextIndex, 1, false)
}

function getProductCover(product: Product): string {
  if (Array.isArray(product.images) && product.images.length) {
    return product.images[0]
  }
  return BrandAsset.DEFAULT_PRODUCT_IMAGE
}

function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}

function goProduct(productId: number) {
  uni.navigateTo({ url: `/pages/store/product?product_id=${productId}` })
}

function goHome() {
  uni.switchTab({ url: `/pages/store/home` })
}

// 加购相关方法
function onProductQuickAdd(product: any) {
  const pt = Number(product.product_type ?? 0)
  if (pt === 2 || pt === 3 || pt === 4 || pt === 5 || Number(product.sale_type) === 2) {
    goProduct(product.id)
    return
  }
  addToCart(product)
}

async function addToCart(product: any) {
  showAddDialog.value = true
  addDialogLoading.value = true
  addDialogProduct.value = null
  addDialogQuantity.value = 1
  Object.keys(addDialogSelectedSpecs).forEach((key) => {
    delete addDialogSelectedSpecs[key]
  })

  try {
    const detail = await getStoreProduct(product.id)
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

  const merchantName = '' // 分类页暂未获取商户名称，可后续补充

  cartStore.addItem({
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
  showAddToCartFeedback(addDialogProduct.value.name, addDialogQuantity.value)
  closeAddDialog()
}

function showAddToCartFeedback(productName: string, quantity: number) {
  addSuccessText.value = `${productName} x${quantity} 已加入购物车`
  showAddSuccessTip.value = true

  if (addSuccessTipTimer) {
    clearTimeout(addSuccessTipTimer)
  }

  addSuccessTipTimer = setTimeout(() => {
    showAddSuccessTip.value = false
  }, 1600)
}

function closeAddDialog() {
  showAddDialog.value = false
}
</script>

<style scoped>
.category-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.page-state {
  padding: 160rpx 48rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.state-title {
  font-size: 32rpx;
  color: #333333;
  font-weight: 600;
}

.state-desc {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #999999;
}

.state-btn {
  margin-top: 40rpx;
  padding: 20rpx 56rpx;
  border-radius: 999rpx;
  background: #007AFF;
  color: #ffffff;
  font-size: 28rpx;
}

.category-body {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 180rpx;
  background: #f8f9fa;
  height: 100%;
}

.sidebar-item {
  padding: 32rpx 20rpx;
  font-size: 26rpx;
  color: #666666;
  border-left: 6rpx solid transparent;
  text-align: center;
}

.sidebar-item.active {
  background: #ffffff;
  color: #007AFF;
  border-left-color: #007AFF;
  font-weight: 600;
}

.goods {
  flex: 1;
  height: 100%;
  background: #ffffff;
  padding: 20rpx 24rpx;
  box-sizing: border-box;
}

.goods-category-title {
  padding: 8rpx 0 16rpx;
  display: flex;
  align-items: center;
}
.goods-category-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #1a1a1a;
  position: relative;
  padding-left: 20rpx;
}
.goods-category-name::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 6rpx;
  height: 28rpx;
  border-radius: 6rpx;
  background: #007AFF;
}

.goods-empty {
  padding: 80rpx 0;
  text-align: center;
}

.goods-empty-text {
  font-size: 26rpx;
  color: #999999;
}

.goods-item {
  display: flex;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.goods-image {
  width: 200rpx;
  min-width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
}

.goods-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.goods-name {
  font-size: 28rpx;
  color: #1a1a1a;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.goods-desc {
  font-size: 22rpx;
  color: #999999;
  margin-top: 6rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  overflow: hidden;
}

.goods-price-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.goods-price {
  font-size: 30rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.goods-rental-tag {
  font-size: 20rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
}

.goods-loading {
  padding: 20rpx 0 4rpx;
  text-align: center;
  font-size: 24rpx;
  color: #999999;
}

.goods-no-more {
  padding: 24rpx 0 12rpx;
  text-align: center;
  font-size: 22rpx;
  color: #999999;
}

/* 加购按钮 */
.add-btn {
  width: 48rpx;
  height: 48rpx;
  background: #007AFF;
  border-radius: 50%;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 300;
  margin-left: auto;
}

/* 加购成功提示 */
.add-success-tip {
  position: fixed;
  left: 50%;
  bottom: calc(160rpx + env(safe-area-inset-bottom));
  transform: translateX(-50%);
  max-width: 620rpx;
  padding: 18rpx 28rpx;
  border-radius: 999rpx;
  background: rgba(26, 26, 26, 0.86);
  color: #ffffff;
  font-size: 26rpx;
  line-height: 1.5;
  text-align: center;
  z-index: 120;
  box-shadow: 0 12rpx 28rpx rgba(0, 0, 0, 0.18);
}

/* 加购弹窗 */
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

.add-dialog-confirm:active {
  opacity: 0.9;
}
</style>
