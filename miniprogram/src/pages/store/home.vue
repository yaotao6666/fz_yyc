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
              <text class="sticky-mini-category">{{ currentCategory?.name || '商品列表' }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 金刚区：轮播图 + icon 分类宫格 整合卡片 -->
      <view class="kingkong-card">
        <!-- 轮播图 -->
        <swiper
          v-if="bannerImages.length > 0"
          class="banner-swiper"
          :indicator-dots="true"
          :autoplay="true"
          :interval="3500"
          :circular="true"
          indicator-active-color="#007AFF"
        >
          <swiper-item v-for="(img, index) in bannerImages" :key="index">
            <image class="banner-image" :src="img" mode="aspectFill" @click="onBannerTap(index)" />
          </swiper-item>
        </swiper>

        <!-- icon 分类宫格 -->
        <view class="icon-grid">
          <view
            v-for="type in PRODUCT_TYPE_SECTIONS"
            :key="type.key"
            class="icon-grid-item"
            @click="goTypeSection(type.key)"
          >
            <view class="icon-grid-icon" :class="'pt-' + type.key">{{ type.icon }}</view>
            <view class="icon-grid-text">{{ type.title }}</view>
          </view>
        </view>
      </view>

      <!-- 领券中心（PRD V2.0 阶段二） -->
      <view v-if="availableCoupons.length" class="coupon-strip">
        <view class="coupon-strip-header">
          <text class="coupon-strip-title">🎁 领券中心</text>
          <text class="coupon-strip-more" @click="goMyCoupons">我的券 ›</text>
        </view>
        <scroll-view class="coupon-strip-scroll" scroll-x>
          <view class="coupon-strip-list">
            <view v-for="coupon in availableCoupons" :key="coupon.id" class="coupon-strip-item">
              <view class="coupon-strip-left">
                <text class="coupon-strip-value" v-if="coupon.type === 1">¥{{ Number(coupon.discount_amount || 0).toFixed(0) }}</text>
                <text class="coupon-strip-value" v-else>{{ ((Number(coupon.discount_rate) || 0) * 10).toFixed(1) }}折</text>
                <text class="coupon-strip-threshold">
                  {{ Number(coupon.threshold_amount) > 0 ? `满${Number(coupon.threshold_amount).toFixed(0)}可用` : '无门槛' }}
                </text>
              </view>
              <view class="coupon-strip-right">
                <text class="coupon-strip-name">{{ coupon.name }}</text>
                <view
                  class="coupon-strip-btn"
                  :class="{ done: !coupon.can_receive }"
                  @click.stop="receiveCouponTap(coupon)"
                >
                  {{ !coupon.can_receive ? '已领取' : '领取' }}
                </view>
              </view>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 健康宣教（首页板块，轮播图高度） -->
      <view v-if="educationArticles.length" class="education-strip">
        <view class="education-strip-header">
          <text class="education-strip-title">📖 健康宣教</text>
          <text class="education-strip-more" @click="goEducationCenter">更多 ›</text>
        </view>
        <scroll-view class="education-strip-scroll" scroll-x>
          <view class="education-strip-list">
            <view
              v-for="item in educationArticles"
              :key="item.id"
              class="education-card"
              @click="goEducationDetail(item.id)"
            >
              <view class="education-card-title">{{ item.title }}</view>
              <view class="education-card-meta">
                <text v-if="item.category" class="education-card-chip">{{ item.category }}</text>
                <text v-if="item.views !== undefined" class="education-card-views">{{ item.views }} 次浏览</text>
              </view>
            </view>
          </view>
        </scroll-view>
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
                  </view>
                  <view class="product-sales">已售 {{ product.sales || 0 }}</view>
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

      </template>

  </view>
</template>

<script setup lang="ts">
import { ref, computed, reactive, nextTick, getCurrentInstance, watch } from 'vue'
import { onLoad, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { getStoreHome, getStoreProducts } from '../../api/store'
import { getAvailableCoupons, receiveCoupon } from '../../api/coupon'
import { getStoreEducationArticles } from '../../api/health'
import type { CouponTemplate } from '../../types/coupon'
import { useAnalytics } from '@utils/analytics'
import { getCachedImagePath, cacheImage } from '@utils/imageCache'
import { parseStoreEntryOptions } from '@utils/storeEntry'
import type { StoreHomeInfo, Product, StoreProductGroup, ProductType, EducationArticle } from '@types'
import { BrandAsset } from '../../utils/constants'

// 商品类型分区配置（首页 icon 宫格只展示 4 大业务分类）
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
  { key: 4, title: '陪诊服务', icon: '🏥', tagClass: 'pt-escort-tag', tagText: '陪诊' }
]

const { trackVisit, trackPageView } = useAnalytics()
const instance = getCurrentInstance()

const storeInfo = ref<StoreHomeInfo | null>(null)
const currentCategoryIndex = ref(0)
const loadingProducts = ref(false)
const merchantLogo = ref('')
const merchantCover = ref('')
const entrySource = ref('scan')
const isInitialLoading = ref(false)
const pageErrorMessage = ref('')
const isNoticeExpanded = ref(false)
const categoryProductsMap = reactive<Record<number, Product[]>>({})
const productSectionOffsets = ref<number[]>([])
const productScrollIntoView = ref('')
const categorySidebarScrollIntoView = ref('')
const productListScrollTop = ref(0)
const showStickyHeader = ref(false)

/* ============ 领券中心（PRD V2.0 阶段二） ============ */
const availableCoupons = ref<CouponTemplate[]>([])

/* ============ 健康宣教（首页板块，轮播图高度） ============ */
const educationArticles = ref<EducationArticle[]>([])

async function loadAvailableCoupons() {
  try {
    const res = await getAvailableCoupons()
    availableCoupons.value = res.list || []
  } catch (_e) {
    availableCoupons.value = []
  }
}

async function receiveCouponTap(coupon: CouponTemplate) {
  if (!coupon.can_receive) {
    uni.showToast({ title: '该券已领取', icon: 'none' })
    return
  }
  try {
    await receiveCoupon(coupon.id)
    uni.showToast({ title: '领取成功', icon: 'success' })
    loadAvailableCoupons()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '领取失败', icon: 'none' })
  }
}

function goMyCoupons() {
  uni.navigateTo({ url: '/pages/store/my-coupons' })
}

async function loadEducationArticles() {
  try {
    const res = await getStoreEducationArticles({ page: 1, page_size: 6 })
    educationArticles.value = res.list || []
  } catch (_e) {
    educationArticles.value = []
  }
}

function goEducationDetail(id: number) {
  uni.navigateTo({ url: `/pages/store/education-detail?id=${id}` })
}

function goEducationCenter() {
  uni.navigateTo({ url: '/pages/store/health-education' })
}

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
// 首页轮播图：仅展示接口 banners，为空时不回退商家封面占位
const bannerImages = computed<string[]>(() => {
  const banners = storeInfo.value?.banners
  if (Array.isArray(banners) && banners.length) {
    return banners.map((banner) => banner.image).filter(Boolean)
  }
  return []
})
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

let showPromise: Promise<void> | null = null
let _loadRetryCount = 0
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
  productSectionOffsets.value = []
  productScrollIntoView.value = ''
  categorySidebarScrollIntoView.value = ''
  productListScrollTop.value = 0
  showStickyHeader.value = false
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

    const loaded = await loadStoreHome()
    if (loaded) {
      void trackVisit({ source })
      void trackPageView('store_home', source)
    }
    void loadAvailableCoupons()
    void loadEducationArticles()
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
    updatePendingReviewBadge(res.pending_review_count)

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

// 待评价红点：通过「我的」Tab 角标展示（tabBar 第 4 项 index=3）
function updatePendingReviewBadge(count?: number) {
  const n = Number(count || 0)
  if (n > 0) {
    uni.setTabBarBadge({ index: 3, text: n > 99 ? '99+' : String(n) })
  } else {
    uni.removeTabBarBadge({ index: 3 })
  }
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

function onBannerTap(index: number) {
  const banner = storeInfo.value?.banners?.[index]
  if (!banner) return
  if (banner.link_type === 'product' && banner.link_value) {
    goProductDetail(Number(banner.link_value))
    return
  }
  if (banner.link_type === 'category') {
    uni.switchTab({ url: `/pages/store/category` })
  }
}

// 点击 icon 分类宫格：统一跳转对应内页（零售/租赁/康养套餐/陪诊服务）
function goTypeSection(productType: number) {
  if (productType === 3 || productType === 4) {
    const mode = productType === 3 ? 'wellness' : 'escort'
    uni.navigateTo({ url: `/pages/store/service-selection?mode=${mode}` })
    return
  }
  uni.navigateTo({ url: `/pages/store/product-list?type=${productType}` })
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

/* 首页轮播图 */
.kingkong-card {
  margin: 24rpx 24rpx 0;
  background: #ffffff;
  border-radius: 20rpx;
  overflow: hidden;
}

/* 领券中心（PRD V2.0 阶段二） */
.coupon-strip {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 24rpx 8rpx;
}

.coupon-strip-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.coupon-strip-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.coupon-strip-more {
  font-size: 24rpx;
  color: #007AFF;
}

.coupon-strip-scroll {
  width: 100%;
  white-space: nowrap;
}

.coupon-strip-list {
  display: inline-flex;
  gap: 16rpx;
  padding-bottom: 20rpx;
}

.coupon-strip-item {
  display: inline-flex;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx;
  border-radius: 16rpx;
  background: linear-gradient(135deg, #fff3ee 0%, #fff8f0 100%);
  border: 2rpx solid rgba(255, 107, 59, 0.25);
  flex-shrink: 0;
}

.coupon-strip-left {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 120rpx;
  padding-right: 16rpx;
  border-right: 2rpx dashed rgba(255, 107, 59, 0.35);
}

.coupon-strip-value {
  font-size: 36rpx;
  font-weight: 700;
  color: #ff3b30;
  line-height: 1.2;
}

.coupon-strip-threshold {
  font-size: 20rpx;
  color: #ff7a45;
  margin-top: 4rpx;
}

.coupon-strip-right {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  max-width: 160rpx;
}

.coupon-strip-name {
  font-size: 24rpx;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160rpx;
}

.coupon-strip-btn {
  font-size: 22rpx;
  color: #ffffff;
  background: linear-gradient(135deg, #ff6b3b 0%, #ff3b30 100%);
  border-radius: 999rpx;
  padding: 6rpx 28rpx;
}

.coupon-strip-btn.done {
  background: #cccccc;
}

/* 健康宣教（首页板块，轮播图高度） */
.education-strip {
  margin: 20rpx 24rpx 0;
  background: #ffffff;
  border-radius: 20rpx;
  padding: 24rpx 24rpx 8rpx;
}

.education-strip-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.education-strip-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.education-strip-more {
  font-size: 24rpx;
  color: #007AFF;
}

.education-strip-scroll {
  width: 100%;
  white-space: nowrap;
}

.education-strip-list {
  display: inline-flex;
  gap: 16rpx;
  padding-bottom: 20rpx;
}

.education-card {
  width: 400rpx;
  height: 240rpx;
  box-sizing: border-box;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 24rpx;
  border-radius: 16rpx;
  background: linear-gradient(135deg, #eaf3ff 0%, #f4f8ff 100%);
  border: 2rpx solid rgba(0, 122, 255, 0.12);
}

.education-card-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.45;
  white-space: normal;
  word-break: break-all;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  overflow: hidden;
}

.education-card-meta {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.education-card-chip {
  font-size: 20rpx;
  color: #007AFF;
  background: rgba(0, 122, 255, 0.1);
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  max-width: 180rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.education-card-views {
  font-size: 20rpx;
  color: #999999;
}

.banner-swiper {
  height: 300rpx;
  width: 100%;
}

.banner-image {
  width: 100%;
  height: 300rpx;
  background: #e8eaf0;
}

/* icon 分类宫格 */
.icon-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 28rpx 8rpx 16rpx;
}

.icon-grid-item {
  flex: 1;
  min-width: 25%;
  max-width: 25%;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 12rpx;
}

.icon-grid-icon {
  width: 88rpx;
  height: 88rpx;
  border-radius: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 42rpx;
  margin-bottom: 12rpx;
}

.icon-grid-icon.pt-1 {
  background: rgba(0, 122, 255, 0.1);
}

.icon-grid-icon.pt-2 {
  background: rgba(255, 149, 0, 0.12);
}

.icon-grid-icon.pt-3 {
  background: rgba(34, 197, 94, 0.12);
}

.icon-grid-icon.pt-4 {
  background: rgba(99, 102, 241, 0.12);
}

.icon-grid-icon.pt-5 {
  background: rgba(100, 116, 139, 0.14);
}

.icon-grid-text {
  font-size: 24rpx;
  color: #333333;
  text-align: center;
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

.sticky-mini-category {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.main-content {
  display: flex;
  /* header(400rpx) + margin-top(24rpx) 正好填满 tabBar 之上可视区，
     商品列表在内部 scroll-view 滚动，页面本身不滚动，避免 tabBar 上方出现空白 */
  height: calc(100vh - 424rpx);
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

@keyframes loadingShimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
