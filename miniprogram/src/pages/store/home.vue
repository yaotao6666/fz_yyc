<template>
  <view class="store-home-container">
    <!-- 店铺头部 -->
    <view class="store-header">
      <image
        v-if="merchantCover"
        class="store-banner"
        :src="merchantCover"
        mode="aspectFill"
      />
      <view v-else class="store-banner store-banner-placeholder"></view>
      <view class="store-mask"></view>
      <view class="store-info">
        <image
          class="store-logo"
          :src="merchantLogo || BrandAsset.DEFAULT_MERCHANT_LOGO"
          mode="aspectFill"
        />
        <view class="store-detail">
          <view class="store-name">{{ storeInfo?.merchant?.name || '加载中...' }}</view>
          <view class="store-meta">
            <view class="rating" v-if="storeInfo?.merchant?.rating">
              <text class="stars">★★★★★</text>
              <text class="rating-value">{{ storeInfo?.merchant?.rating }}</text>
            </view>
            <text v-if="merchantPhone" class="merchant-phone" @click.stop="callMerchant">
              📞 {{ merchantPhone }}
            </text>
          </view>
          <view v-if="storeInfo?.merchant?.address" class="store-address">
            {{ storeInfo.merchant.address }}
          </view>
          <view
            class="store-notice"
            :class="{ expanded: isNoticeExpanded, clickable: canToggleNotice }"
            v-if="storeInfo?.merchant?.announcement"
            @click="toggleNotice"
          >
            <text class="notice-icon">📢</text>
            <text class="notice-text">{{ storeInfo.merchant.announcement }}</text>
            <text v-if="canToggleNotice" class="notice-toggle">{{ isNoticeExpanded ? '收起' : '展开' }}</text>
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

    <view v-if="isInitialLoading" class="page-state-card">
      <view class="page-state-title">正在加载店铺信息</view>
      <view class="page-state-desc">首次进入或下拉刷新时会重新拉取商家数据，请稍候。</view>
      <view class="page-state-loading">
        <view class="page-state-loading-bar"></view>
        <view class="page-state-loading-bar short"></view>
      </view>
    </view>

    <view v-else-if="showPageError" class="page-state-card error">
      <view class="page-state-title">店铺加载失败</view>
      <view class="page-state-desc">{{ pageErrorMessage }}</view>
      <view class="page-state-actions">
        <view class="page-state-btn primary" @click="retryLoadStoreHome">重新加载</view>
      </view>
    </view>

    <template v-else-if="storeInfo">
      <view v-if="showStickyHeader" class="sticky-mini-bar">
        <view class="sticky-mini-main">
          <image
            class="sticky-mini-logo"
            :src="merchantLogo || BrandAsset.DEFAULT_MERCHANT_LOGO"
            mode="aspectFill"
          />
          <view class="sticky-mini-info">
            <view class="sticky-mini-name">{{ storeInfo.merchant.name }}</view>
            <view class="sticky-mini-meta">
              <text class="sticky-mini-status" :class="{ closed: storeInfo.merchant.status !== 1 }">
                {{ storeInfo.merchant.status === 1 ? '营业中' : '休息中' }}
              </text>
              <text class="sticky-mini-divider">·</text>
              <text class="sticky-mini-category">{{ currentCategory?.name || '商品列表' }}</text>
            </view>
          </view>
        </view>
        <view class="sticky-mini-cart" @click="handlePrimaryAction">
          {{ isCartEmpty ? '去选购' : `¥${cartAmount.toFixed(2)}` }}
        </view>
      </view>

      <view class="quick-entry-bar">
        <view class="quick-entry-row">
          <view class="quick-entry-card" @click="goMyOrders">
            <view class="quick-entry-icon">📋</view>
            <view class="quick-entry-content">
              <view class="quick-entry-title">我的订单</view>
              <view class="quick-entry-desc">查看订单与退款进度</view>
            </view>
            <view class="quick-entry-arrow">›</view>
          </view>
          <view class="quick-entry-card" @click="goMyHealth">
            <view class="quick-entry-icon health">🩺</view>
            <view class="quick-entry-content">
              <view class="quick-entry-title">我的健康</view>
              <view class="quick-entry-desc">健康档案与自助评估</view>
            </view>
            <view class="quick-entry-arrow">›</view>
          </view>
        </view>
      </view>

      <!-- 分类和商品 -->
      <view class="main-content">
        <!-- 左侧分类 -->
        <scroll-view
          class="category-sidebar"
          scroll-y
          scroll-with-animation
          :scroll-into-view="categorySidebarScrollIntoView"
        >
          <view v-if="!hasCategories" class="category-empty">
            暂无分类
          </view>
          <view
            v-for="(category, index) in storeInfo.categories"
            :key="category.id"
            :id="getCategoryMenuId(category.id)"
            class="category-item"
            :class="{ active: currentCategoryIndex === index }"
            @click="selectCategory(index)"
          >
            <text class="category-name">{{ category.name }}</text>
            <text class="category-count" v-if="category.product_count">{{ category.product_count }}</text>
          </view>
        </scroll-view>

        <!-- 右侧商品 -->
        <scroll-view
          class="product-list"
          scroll-y
          scroll-with-animation
          :scroll-into-view="productScrollIntoView"
          @scroll="handleProductListScroll"
          @scrolltolower="loadMoreProducts"
        >
          <view class="product-list-top-anchor" id="product-list-top"></view>

          <!-- 热销推荐 -->
          <view class="hot-products" v-if="showHotProducts">
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
                  <view class="product-name">
                    {{ product.name }}
                    <text v-if="Number(product.sale_type) === 2" class="product-rental-tag">租赁</text>
                    <text v-else-if="Number(product.product_type) === 3" class="product-wellness-tag">套餐</text>
                    <text v-else-if="Number(product.product_type) === 4" class="product-escort-tag">陪诊</text>
                    <text v-else-if="Number(product.product_type) === 5" class="product-info-tag">资讯</text>
                  </view>
                  <view class="product-bottom">
                    <view class="product-price">
                      <template v-if="Number(product.sale_type) === 2">
                        <text class="price">¥{{ Number(product.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(product.rental_unit) }}</text>
                      </template>
                      <template v-else>
                        <text class="price">¥{{ product.price.toFixed(2) }}</text>
                        <text v-if="(product.original_price || 0) > 0" class="original-price">
                          ¥{{ (product.original_price || 0).toFixed(2) }}
                        </text>
                      </template>
                    </view>
                    <view class="add-btn" @click.stop="onProductQuickAdd(product)">+</view>
                  </view>
                  <view class="product-sales">已售 {{ product.sales || 0 }}</view>
                </view>
              </view>
            </view>
          </view>

          <!-- 按商品类型分区展示 -->
          <view class="type-sections" v-if="hasProductTypeSections">
            <view
              v-for="section in productTypeSections"
              :key="section.key"
              class="type-section"
            >
              <view class="type-section-header">
                <text class="type-section-icon">{{ section.icon }}</text>
                <text class="type-section-title">{{ section.title }}</text>
                <text class="type-section-count">{{ section.products.length }} 件</text>
              </view>
              <view class="type-section-grid">
                <view
                  v-for="product in section.products.slice(0, 6)"
                  :key="product.id"
                  class="type-section-card"
                  @click="goProductDetail(product.id)"
                >
                  <image
                    class="type-card-image"
                    :src="getProductImage(product)"
                    mode="aspectFill"
                  />
                  <view class="type-card-info">
                    <view class="type-card-name">
                      {{ product.name }}
                      <text v-if="Number(section.key) === 2" class="product-rental-tag">租赁</text>
                      <text v-else-if="Number(section.key) === 3" class="product-wellness-tag">套餐</text>
                      <text v-else-if="Number(section.key) === 4" class="product-escort-tag">陪诊</text>
                      <text v-else-if="Number(section.key) === 5" class="product-info-tag">资讯</text>
                    </view>
                    <view class="type-card-desc" v-if="product.description && Number(section.key) === 3">
                      {{ getWellnessPackagePreview(product) }}
                    </view>
                    <view class="type-card-bottom">
                      <view class="type-card-price">
                        <template v-if="Number(section.key) === 2">
                          <text class="price">¥{{ Number(product.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(product.rental_unit) }}</text>
                          <text class="rental-deposit-tip">押金 ¥{{ Number(product.deposit || 0).toFixed(2) }}</text>
                        </template>
                        <template v-else>
                          <text class="price">¥{{ product.price.toFixed(2) }}</text>
                          <text v-if="(product.original_price || 0) > 0" class="original-price">
                            ¥{{ (product.original_price || 0).toFixed(2) }}
                          </text>
                        </template>
                      </view>
                      <view class="add-btn" @click.stop="onProductQuickAdd(product)">+</view>
                    </view>
                  </view>
                </view>
              </view>
            </view>
          </view>

          <!-- 分类商品列表 -->
          <view v-if="showProductSkeleton" class="product-skeleton-list">
            <view v-for="item in 3" :key="item" class="product-skeleton-item">
              <view class="product-skeleton-image"></view>
              <view class="product-skeleton-content">
                <view class="product-skeleton-line primary"></view>
                <view class="product-skeleton-line secondary"></view>
                <view class="product-skeleton-line short"></view>
              </view>
            </view>
          </view>

          <view v-else class="product-list-items">
            <view
              v-for="section in productSections"
              :key="section.category.id"
              :id="getCategorySectionId(section.category.id)"
              class="product-section"
            >
              <view class="category-title">{{ section.category.name }}</view>
              <view v-if="section.products.length" class="product-section-list">
                <view
                  v-for="product in section.products"
                  :key="product.id"
                  class="product-list-item"
                  @click="goProductDetail(product.id)"
                >
                  <image
                    class="item-image"
                    :src="getProductImage(product)"
                    mode="aspectFill"
                  />
                  <view class="item-info">
                    <view class="item-name">
                      {{ product.name }}
                      <text v-if="Number(product.product_type) === 2 || Number(product.sale_type) === 2" class="product-rental-tag">租赁</text>
                      <text v-else-if="Number(product.product_type) === 3" class="product-wellness-tag">套餐</text>
                      <text v-else-if="Number(product.product_type) === 4" class="product-escort-tag">陪诊</text>
                      <text v-else-if="Number(product.product_type) === 5" class="product-info-tag">资讯</text>
                    </view>
                    <view class="item-desc" v-if="product.description">{{ product.description }}</view>
                    <view class="item-bottom">
                      <view class="item-price">
                        <template v-if="Number(product.sale_type) === 2">
                          <text class="price">¥{{ Number(product.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(product.rental_unit) }}</text>
                          <text class="rental-deposit-tip">押金 ¥{{ Number(product.deposit || 0).toFixed(2) }}</text>
                        </template>
                        <template v-else>
                          <text class="price">¥{{ product.price.toFixed(2) }}</text>
                          <text v-if="(product.original_price || 0) > 0" class="original-price">
                            ¥{{ (product.original_price || 0).toFixed(2) }}
                          </text>
                        </template>
                      </view>
                      <view class="add-btn" @click.stop="onProductQuickAdd(product)">+</view>
                    </view>
                  </view>
                </view>
              </view>
              <view v-else class="section-empty">当前分类暂无上架商品</view>
            </view>
          </view>

          <view v-if="loadingProducts && hasLoadedAnyProducts" class="loading">加载中...</view>
          <view v-else-if="showProductEmpty" class="product-empty">
            <view class="product-empty-title">当前暂无可展示商品</view>
            <view class="product-empty-desc">可以下拉刷新试试，或稍后再来看看商家上新。</view>
          </view>
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

      <view v-if="showAddSuccessTip" class="add-success-tip">
        {{ addSuccessText }}
      </view>

      <!-- 底部占位 -->
      <view class="bottom-placeholder"></view>
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
            <view class="guide-text">等待商品配送</view>
          </view>
        </view>
        <view class="guide-footer">
          <view class="guide-tip">💡 有任何问题？点击「我的订单」联系商家</view>
          <view class="guide-close" @click="closeGuide">知道了</view>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { ref, computed, reactive, nextTick, getCurrentInstance, watch } from 'vue'
import { onLoad, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { getStoreHome, getStoreProducts, getStoreProduct } from '../../api/store'
import { useCartStore } from '../../stores/cart'
import { useAnalytics } from '@utils/analytics'
import { getCachedImagePath, cacheImage } from '@utils/imageCache'
import { parseStoreEntryOptions } from '@utils/storeEntry'
import type { StoreHomeInfo, Product, SpecOption, StoreProductGroup, ProductType } from '@types'
import { BrandAsset } from '../../utils/constants'

// 商品类型分区配置
const PRODUCT_TYPE_SECTIONS: Array<{
  key: ProductType
  title: string
  icon: string
  tagClass: string
  tagText: string
}> = [
  { key: 1, title: '辅具零售', icon: '🛍️', tagClass: 'pt-retail-tag', tagText: '一口价' },
  { key: 2, title: '辅具租赁', icon: '🔑', tagClass: 'pt-rental-tag', tagText: '租赁' },
  { key: 3, title: '康养套餐', icon: '🌿', tagClass: 'pt-wellness-tag', tagText: '套餐' },
  { key: 4, title: '陪诊服务', icon: '🏥', tagClass: 'pt-escort-tag', tagText: '陪诊' },
  { key: 5, title: '科普资讯', icon: '📚', tagClass: 'pt-info-tag', tagText: '资讯' }
]

function getProductTypeSection(productType: ProductType | number) {
  return PRODUCT_TYPE_SECTIONS.find(s => s.key === productType)
}

const cartStore = useCartStore()
const { trackVisit, trackPageView } = useAnalytics()
const instance = getCurrentInstance()

const storeInfo = ref<StoreHomeInfo | null>(null)
const currentCategoryIndex = ref(0)
const loadingProducts = ref(false)
const merchantLogo = ref('')
const merchantCover = ref('')
const showGuide = ref(false) // 控制操作指引弹窗显示
const entrySource = ref('scan')
const isInitialLoading = ref(false)
const pageErrorMessage = ref('')
const isNoticeExpanded = ref(false)
const showAddSuccessTip = ref(false)
const addSuccessText = ref('')
const categoryProductsMap = reactive<Record<number, Product[]>>({})
const productSectionOffsets = ref<number[]>([])
const productScrollIntoView = ref('')
const categorySidebarScrollIntoView = ref('')
const productListScrollTop = ref(0)
const showStickyHeader = ref(false)

const showAddDialog = ref(false)
const addDialogLoading = ref(false)
const addDialogProduct = ref<Product | null>(null)
const addDialogQuantity = ref(1)
const addDialogSelectedSpecs = reactive<Record<string, string>>({})

const currentCategory = computed(() => {
  return storeInfo.value?.categories?.[currentCategoryIndex.value] || null
})
const hasCategories = computed(() => !!storeInfo.value?.categories?.length)
const showHotProducts = computed(() => !!storeInfo.value?.hot_products?.length)
const productSections = computed<StoreProductGroup[]>(() => {
  return (storeInfo.value?.categories || []).map(category => ({
    category: {
      id: category.id,
      name: category.name
    },
    products: categoryProductsMap[category.id] || []
  }))
})
// 所有已加载商品（跨分类聚合）
const allProducts = computed<Product[]>(() => {
  const list: Product[] = []
  Object.values(categoryProductsMap).forEach(items => list.push(...items))
  return list
})
// 按商品类型分区（用于首页按类型展示）
const productTypeSections = computed(() => {
  return PRODUCT_TYPE_SECTIONS.map(section => ({
    ...section,
    products: allProducts.value.filter(p => Number(p.product_type) === section.key)
  })).filter(section => section.products.length > 0)
})
const hasProductTypeSections = computed(() => productTypeSections.value.length > 0)
const hasLoadedAnyProducts = computed(() => productSections.value.some(section => section.products.length > 0))
const showProductEmpty = computed(() => !loadingProducts.value && !hasLoadedAnyProducts.value && !showHotProducts.value)
const showPageError = computed(() => !!pageErrorMessage.value && !storeInfo.value)
const showProductSkeleton = computed(() => loadingProducts.value && !hasLoadedAnyProducts.value)
const canToggleNotice = computed(() => {
  const notice = storeInfo.value?.merchant?.announcement || ''
  return notice.trim().length > 28
})
const merchantPhone = computed(() => {
  const merchant = storeInfo.value?.merchant as any
  return String((merchant?.contact_phone || merchant?.phone || '') ?? '').trim()
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
let _loadRetryCount = 0
let addSuccessTipTimer: ReturnType<typeof setTimeout> | null = null
let manualCategoryScrollTimer: ReturnType<typeof setTimeout> | null = null

function callMerchant() {
  const phone = merchantPhone.value
  const merchantName = storeInfo.value?.merchant?.name || '商家'

  if (!phone) {
    uni.showToast({ title: `暂无${merchantName}联系电话`, icon: 'none' })
    return
  }

  uni.makePhoneCall({
    phoneNumber: phone,
    fail: () => {
      uni.showToast({ title: `请联系${merchantName}`, icon: 'none' })
    }
  })
}
let ignoreScrollSync = false

function resetStoreHomeState() {
  storeInfo.value = null
  currentCategoryIndex.value = 0
  merchantLogo.value = ''
  merchantCover.value = ''
  pageErrorMessage.value = ''
  isNoticeExpanded.value = false
  showAddSuccessTip.value = false
  addSuccessText.value = ''
  productSectionOffsets.value = []
  productScrollIntoView.value = ''
  categorySidebarScrollIntoView.value = ''
  productListScrollTop.value = 0
  showStickyHeader.value = false
  showAddDialog.value = false
  addDialogLoading.value = false
  addDialogProduct.value = null
  Object.keys(categoryProductsMap).forEach((key) => {
    delete categoryProductsMap[Number(key)]
  })
}

watch(currentCategoryIndex, () => {
  const categoryId = currentCategory.value?.id
  if (!categoryId) {
    return
  }
  categorySidebarScrollIntoView.value = getCategoryMenuId(categoryId)
})

function parseEntryOptions(options?: Record<string, any>) {
  return parseStoreEntryOptions(options)
}

function applyEntryOptions(options?: Record<string, any>) {
  const { source } = parseEntryOptions(options)
  entrySource.value = source
}

onLoad((options) => {
  applyEntryOptions(options as Record<string, any> | undefined)
})

onPullDownRefresh(async () => {
  await refreshStoreHome()
})

onShow(() => {
  if (showPromise) return

  showPromise = (async () => {
    const pages = getCurrentPages()
    const currentPage = pages[pages.length - 1] as any
    applyEntryOptions(currentPage?.options)
    const source = entrySource.value

    showGuide.value = !uni.getStorageSync('storeHomeGuideShown')

    const loaded = await loadStoreHome()
    if (loaded) {
      void trackVisit({ source })
      void trackPageView('store_home', source)
    }
  })().finally(() => {
    showPromise = null
  })
})

async function loadStoreHome(silent = false) {
  if (!silent && !storeInfo.value) {
    isInitialLoading.value = true
  }
  pageErrorMessage.value = ''

  try {
    const res = await getStoreHome()
    _loadRetryCount = 0
    storeInfo.value = res

    cacheMerchantImages(res)

    if (res.categories?.length) {
      await loadAllCategoryProducts(res.categories, silent)
      await nextTick()
      measureProductSections()
    }
    return true
  } catch (error) {
    if (!silent) {
      console.error('加载店铺信息失败:', error)
      if (error instanceof TypeError && _loadRetryCount < 2) {
        _loadRetryCount++
        console.log(`加载店铺信息重试 (${_loadRetryCount}/2)`)
        await new Promise(resolve => setTimeout(resolve, 600))
        return loadStoreHome(silent)
      }
      pageErrorMessage.value = error instanceof Error ? error.message || '请重新进入后重试' : '请重新进入后重试'
      uni.showToast({ title: '加载失败，请重新进入', icon: 'none' })
    }
    return false
  } finally {
    isInitialLoading.value = false
  }
}

async function refreshStoreHome() {
  const loaded = await loadStoreHome(true)
  uni.stopPullDownRefresh()
  if (!loaded) {
    uni.showToast({ title: '刷新失败，请稍后重试', icon: 'none' })
  }
}

function retryLoadStoreHome() {
  void loadStoreHome()
}

function toggleNotice() {
  if (!canToggleNotice.value) {
    return
  }

  isNoticeExpanded.value = !isNoticeExpanded.value
}

async function loadAllCategoryProducts(
  categories: Array<{ id: number; name: string; sort: number; product_count: number }>,
  silent = false
) {
  if (!silent) {
    loadingProducts.value = true
  }

  Object.keys(categoryProductsMap).forEach((key) => {
    delete categoryProductsMap[Number(key)]
  })

  try {
    const results = await Promise.allSettled(categories.map(async (category) => {
      const res = await getStoreProducts({ category_id: category.id })
      return { categoryId: category.id, list: res.list || [] }
    }))

    results.forEach((result, index) => {
      const categoryId = categories[index].id
      if (result.status === 'fulfilled') {
        categoryProductsMap[categoryId] = result.value.list
      } else {
        categoryProductsMap[categoryId] = []
      }
    })
  } catch (error) {
    if (!silent) {
      console.error('加载商品失败:', error)
    }
  } finally {
    loadingProducts.value = false
  }
}

function getCategoryMenuId(categoryId: number) {
  return `category-menu-${categoryId}`
}

function getCategorySectionId(categoryId: number) {
  return `category-section-${categoryId}`
}

function cacheMerchantImages(info: StoreHomeInfo) {
  const logo = info?.merchant?.logo
  const cover = info?.merchant?.cover_image

  if (logo) {
    const cached = getCachedImagePath(logo)
    if (cached) {
      merchantLogo.value = cached
    } else {
      merchantLogo.value = logo
      cacheImage(logo).then((path) => {
        merchantLogo.value = path
      })
    }
  }

  if (cover) {
    const cached = getCachedImagePath(cover)
    if (cached) {
      merchantCover.value = cached
    } else {
      merchantCover.value = cover
      cacheImage(cover).then((path) => {
        merchantCover.value = path
      })
    }
  }
}

function measureProductSections() {
  if (!instance?.proxy || !productSections.value.length) {
    productSectionOffsets.value = []
    return
  }

  const currentScrollTop = productListScrollTop.value
  const query = uni.createSelectorQuery().in(instance.proxy)
  query.select('.product-list').boundingClientRect()
  query.selectAll('.product-section').boundingClientRect()
  query.exec((result) => {
    const containerRect = result?.[0] as { top: number } | undefined
    const sectionRects = (result?.[1] || []) as Array<{ top: number }>
    if (!containerRect || !sectionRects.length) {
      productSectionOffsets.value = []
      return
    }

    productSectionOffsets.value = sectionRects.map(rect => rect.top - containerRect.top + currentScrollTop)
  })
}

function setManualCategoryScrollLock() {
  ignoreScrollSync = true
  if (manualCategoryScrollTimer) {
    clearTimeout(manualCategoryScrollTimer)
  }
  manualCategoryScrollTimer = setTimeout(() => {
    ignoreScrollSync = false
  }, 420)
}

function scrollToCategory(categoryId: number) {
  setManualCategoryScrollLock()
  productScrollIntoView.value = ''
  nextTick(() => {
    productScrollIntoView.value = getCategorySectionId(categoryId)
  })
}

function selectCategory(index: number) {
  const category = storeInfo.value?.categories?.[index]
  if (!category) {
    return
  }

  currentCategoryIndex.value = index
  scrollToCategory(category.id)
}

function loadMoreProducts() {
  // 加载更多逻辑
}

function syncCurrentCategoryByScroll(scrollTop: number) {
  if (!productSectionOffsets.value.length || !storeInfo.value?.categories?.length) {
    return
  }

  const activeScrollTop = scrollTop + 24
  let nextIndex = 0

  for (let index = 0; index < productSectionOffsets.value.length; index++) {
    if (activeScrollTop >= productSectionOffsets.value[index]) {
      nextIndex = index
    } else {
      break
    }
  }

  currentCategoryIndex.value = nextIndex
}

function handleProductListScroll(event: any) {
  const scrollTop = Number(event?.detail?.scrollTop || 0)
  productListScrollTop.value = scrollTop
  showStickyHeader.value = scrollTop > 96

  if (!ignoreScrollSync) {
    syncCurrentCategoryByScroll(scrollTop)
  }
}

function getHotProductImage(product: any) {
  if (Array.isArray(product?.images) && product.images.length > 0) {
    return product.images[0]
  }

  return getProductImage(product)
}

function getProductImage(product: any) {
  if (Array.isArray(product?.images) && product.images.length > 0) {
    return product.images[0]
  }

  return product?.image || BrandAsset.DEFAULT_PRODUCT_IMAGE
}

function goProductDetail(productId: number) {
  uni.navigateTo({
    url: `/pages/store/product?product_id=${productId}`
  })
}

function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}

// 首页快捷加购：租赁/康养套餐/陪诊/资讯 跳详情页；普通一口价商品走原加购弹窗
function onProductQuickAdd(product: any) {
  const pt = Number(product.product_type ?? 0)
  if (pt === 2 || pt === 3 || pt === 4 || pt === 5 || Number(product.sale_type) === 2) {
    goProductDetail(product.id)
    return
  }
  addToCart(product)
}

// 康养套餐预览：展示前 N 个服务项名称
function getWellnessPackagePreview(product: Product): string {
  try {
    const sc = product.service_content
    if (sc && Array.isArray(sc.services) && sc.services.length > 0) {
      const names = sc.services.slice(0, 3).map((s: any) => s?.name || '').filter(Boolean)
      if (names.length) return '包含：' + names.join('、')
    }
  } catch (_) {}
  if (product.description) return product.description
  return '查看详情了解套餐内容'
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

  const merchantName = storeInfo.value?.merchant?.name || ''

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

function goCart() {
  if (isCartEmpty.value) {
    uni.showToast({ title: '请先选择商品', icon: 'none' })
    return
  }

  uni.navigateTo({
    url: `/pages/store/cart`
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

  uni.navigateTo({
    url: `/pages/store/confirm`
  })
}

function closeGuide() {
  showGuide.value = false

  uni.setStorageSync('storeHomeGuideShown', true)
}

function goMyOrders() {
  uni.navigateTo({
    url: `/pages/store/my-orders`
  })
}

function goMyHealth() {
  uni.navigateTo({
    url: `/pages/store/my-health`
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

.store-banner-placeholder {
  background:
    radial-gradient(circle at top right, rgba(255,255,255,0.24), transparent 38%),
    radial-gradient(circle at bottom left, rgba(255,255,255,0.18), transparent 30%),
    linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  flex-shrink: 0;
}

.store-detail {
  flex: 1;
  min-width: 0;
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

.merchant-phone {
  max-width: 280rpx;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  background: rgba(255, 255, 255, 0.18);
  border: 2rpx solid rgba(255, 255, 255, 0.22);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.store-address {
  font-size: 24rpx;
  opacity: 0.9;
  margin-bottom: 12rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stars {
  color: #ffd700;
  font-size: 24rpx;
}

.rating-value {
  font-size: 24rpx;
  margin-left: 8rpx;
}

.store-notice {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  opacity: 0.9;
  margin-top: 4rpx;
}

.store-notice.clickable {
  padding-right: 12rpx;
}

.store-notice.expanded {
  align-items: flex-start;
}

.notice-icon {
  margin-right: 8rpx;
  margin-top: 2rpx;
}

.notice-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.store-notice.expanded .notice-text {
  white-space: normal;
  line-height: 1.6;
}

.notice-toggle {
  margin-left: 12rpx;
  color: rgba(255, 255, 255, 0.92);
  font-size: 22rpx;
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

.sticky-mini-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 90;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 18rpx 24rpx;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 10rpx 24rpx rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(12rpx);
}

.sticky-mini-main {
  flex: 1;
  display: flex;
  align-items: center;
  min-width: 0;
}

.sticky-mini-logo {
  width: 68rpx;
  height: 68rpx;
  border-radius: 18rpx;
  margin-right: 18rpx;
  background: #f5f5f5;
}

.sticky-mini-info {
  flex: 1;
  min-width: 0;
}

.sticky-mini-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sticky-mini-meta {
  margin-top: 6rpx;
  display: flex;
  align-items: center;
  gap: 10rpx;
  font-size: 22rpx;
  color: #666666;
}

.sticky-mini-status {
  color: #18a058;
}

.sticky-mini-status.closed {
  color: #999999;
}

.sticky-mini-divider {
  color: #c7c7c7;
}

.sticky-mini-category {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sticky-mini-cart {
  min-width: 148rpx;
  height: 64rpx;
  padding: 0 22rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 24rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-state-card {
  margin: 24rpx;
  padding: 36rpx 32rpx;
  background: #ffffff;
  border-radius: 24rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.05);
}

.page-state-card.error {
  border: 1rpx solid rgba(255, 77, 79, 0.18);
}

.page-state-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.page-state-desc {
  margin-top: 14rpx;
  font-size: 26rpx;
  color: #666666;
  line-height: 1.6;
}

.page-state-loading {
  margin-top: 28rpx;
}

.page-state-loading-bar {
  height: 22rpx;
  border-radius: 12rpx;
  background: linear-gradient(90deg, #f2f3f5 0%, #e9ecef 50%, #f2f3f5 100%);
  background-size: 200% 100%;
  animation: loadingShimmer 1.2s linear infinite;
}

.page-state-loading-bar.short {
  width: 60%;
  margin-top: 18rpx;
}

.page-state-actions {
  margin-top: 28rpx;
}

.page-state-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 220rpx;
  height: 84rpx;
  padding: 0 28rpx;
  border-radius: 42rpx;
  font-size: 28rpx;
  font-weight: 500;
}

.page-state-btn.primary {
  color: #ffffff;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
}

.quick-entry-row {
  display: flex;
  gap: 20rpx;
}

.quick-entry-card {
  flex: 1;
  min-width: 0;
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
  flex-shrink: 0;
}

.quick-entry-icon.health {
  background: rgba(22, 163, 74, 0.1);
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

.category-empty {
  padding: 48rpx 20rpx;
  text-align: center;
  font-size: 24rpx;
  color: #999999;
  line-height: 1.5;
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
  width: calc(100% - 48rpx);
  padding: 24rpx;
}

.product-list-top-anchor {
  height: 2rpx;
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

.product-rental-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.product-wellness-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #22c55e;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.product-escort-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #6366f1;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.product-info-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #64748b;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
}

.rental-deposit-tip {
  display: block;
  font-size: 22rpx;
  color: #ff9500;
  margin-top: 4rpx;
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
  flex-shrink: 0;
}

.product-sales {
  font-size: 22rpx;
  color: #999999;
}

.product-list-items {
  display: flex;
  flex-direction: column;
}

.product-section {
  margin-bottom: 20rpx;
}

.product-section:last-child {
  margin-bottom: 0;
}

.product-section-list {
  display: flex;
  flex-direction: column;
}

.product-skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.product-skeleton-item {
  display: flex;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.product-skeleton-image {
  width: 200rpx;
  height: 200rpx;
  margin-right: 20rpx;
  border-radius: 12rpx;
  background: linear-gradient(90deg, #f2f3f5 0%, #e9ecef 50%, #f2f3f5 100%);
  background-size: 200% 100%;
  animation: loadingShimmer 1.2s linear infinite;
}

.product-skeleton-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 18rpx;
}

.product-skeleton-line {
  height: 24rpx;
  border-radius: 12rpx;
  background: linear-gradient(90deg, #f2f3f5 0%, #e9ecef 50%, #f2f3f5 100%);
  background-size: 200% 100%;
  animation: loadingShimmer 1.2s linear infinite;
}

.product-skeleton-line.primary {
  width: 68%;
}

.product-skeleton-line.secondary {
  width: 92%;
}

.product-skeleton-line.short {
  width: 38%;
}

.product-list-item {
  display: flex;
  align-items: flex-start;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.product-list-item:last-child {
  border-bottom: none;
}

.section-empty {
  padding: 24rpx 0 32rpx;
  font-size: 24rpx;
  color: #999999;
  text-align: center;
}

.item-image {
  width: 200rpx;
  min-width: 200rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f0f0f0;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.item-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
  line-height: 1.4;
  word-break: break-all;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.item-desc {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
  line-height: 1.5;
  word-break: break-all;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.item-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  margin-top: 16rpx;
}

.item-price {
  flex: 1;
  min-width: 0;
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

.product-empty {
  margin-top: 24rpx;
  padding: 48rpx 24rpx;
  background: #f8f9fa;
  border-radius: 20rpx;
  text-align: center;
}

.product-empty-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333333;
}

.product-empty-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #888888;
  line-height: 1.6;
}

/* 按商品类型分区展示 */
.type-sections {
  margin-bottom: 20rpx;
}

.type-section {
  margin-bottom: 28rpx;
  padding: 20rpx;
  background: #ffffff;
  border-radius: 20rpx;
  border: 1rpx solid #f0f0f0;
}

.type-section-header {
  display: flex;
  align-items: center;
  margin-bottom: 20rpx;
}

.type-section-icon {
  font-size: 36rpx;
  margin-right: 10rpx;
}

.type-section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  flex: 1;
}

.type-section-count {
  font-size: 22rpx;
  color: #999999;
}

.type-section-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
}

.type-section-card {
  background: #fafafa;
  border-radius: 16rpx;
  overflow: hidden;
}

.type-card-image {
  width: 100%;
  height: 220rpx;
  background: #e0e0e0;
}

.type-card-info {
  padding: 14rpx;
}

.type-card-name {
  font-size: 26rpx;
  color: #1a1a1a;
  line-height: 1.4;
  margin-bottom: 8rpx;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.type-card-desc {
  font-size: 22rpx;
  color: #666666;
  margin-bottom: 10rpx;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.type-card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.type-card-price {
  flex: 1;
  min-width: 0;
}

.type-card-price .price {
  font-size: 28rpx;
  font-weight: 600;
  color: #ff4d4f;
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

@keyframes loadingShimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
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
