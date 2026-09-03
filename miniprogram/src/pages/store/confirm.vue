<template>
  <view class="confirm-container">
    <!-- 收货信息（服务订单为「服务地址」） -->
    <view class="section delivery-section">
      <view class="section-title">{{ hasServiceItems ? '服务地址' : '收货信息' }}</view>

      <view class="address-card" :class="{ empty: !selectedAddress }" @click="goAddressList">
        <view v-if="selectedAddress" class="address-main">
          <view class="address-top">
            <text class="address-name">{{ selectedAddress.name }}</text>
            <text class="address-phone">{{ selectedAddress.phone }}</text>
            <text v-if="selectedAddress.is_default" class="address-default-tag">默认</text>
          </view>
          <view class="address-detail">{{ selectedAddressText }}</view>
        </view>
        <view v-else class="address-empty">
          <text class="address-empty-title">请选择收货地址</text>
          <text class="address-empty-desc">支持从微信快速导入，也可选择已保存地址</text>
        </view>
        <text class="address-arrow">></text>
      </view>

      <view class="address-actions">
        <view class="address-action-btn" @click.stop="goAddressList">选择地址</view>
        <view class="address-action-btn secondary" @click.stop="goAddressEdit">新增地址</view>
        <view class="address-action-btn secondary" @click.stop="importWechatAddress">微信导入</view>
      </view>

      <view class="remark-form">
        <view class="form-label">备注</view>
        <input
          v-model="remark"
          class="form-input"
          placeholder="下单备注信息（选填）"
        />
      </view>
    </view>

    <!-- 服务对象（服务订单必选健康档案） -->
    <view v-if="hasServiceItems" class="section record-section">
      <view class="section-title-row">
        <view class="section-title">服务对象</view>
        <text v-if="healthRecordList.length > 1" class="section-meta">共 {{ healthRecordList.length }} 份档案</text>
      </view>
      <view class="record-card" :class="{ empty: !healthRecord }" @click="goHealthRecord">
        <template v-if="healthRecord">
          <view class="record-main">
            <view class="record-top">
              <text class="record-name">{{ healthRecord.real_name || '未命名' }}</text>
              <text class="record-rel-tag">{{ relationText(healthRecord.relation) }}</text>
              <text class="record-tag">服务档案</text>
            </view>
            <view class="record-detail">
              {{ recordSummaryText }}
            </view>
          </view>
          <text class="address-arrow">></text>
        </template>
        <view v-else class="record-empty">
          <text class="record-empty-title">{{ loadingRecord ? '档案加载中...' : '请先创建健康档案' }}</text>
          <text class="record-empty-desc">服务订单需指定服务对象，点击前往建档</text>
        </view>
      </view>
    </view>

    <!-- 商品信息 -->
    <view class="section goods-section">
      <view class="section-title">商品信息</view>
      <view class="goods-list">
        <view
          v-for="item in displayItems"
          :key="`${item.product_id}-${item.specs}-${item.rental_duration || 0}`"
          class="goods-item"
          :class="{ 'goods-item-rental': Number(item.sale_type) === 2 }"
        >
          <image
            class="goods-image"
            :src="item.image || BrandAsset.DEFAULT_PRODUCT_IMAGE"
            mode="aspectFill"
          />
          <view class="goods-info">
            <view class="goods-name">
              {{ item.product_name }}
              <text v-if="Number(item.product_type) === 2 || Number(item.sale_type) === 2" class="goods-rental-tag">租赁</text>
              <text v-else-if="Number(item.product_type) === 3" class="goods-wellness-tag">套餐</text>
              <text v-else-if="Number(item.product_type) === 4" class="goods-escort-tag">陪诊</text>
            </view>
            <view class="goods-spec" v-if="item.specs">{{ item.specs }}</view>
            <view class="goods-rental-info" v-if="Number(item.sale_type) === 2">
              <text class="rental-info-text">¥{{ Number(item.rental_price || 0).toFixed(2) }}/{{ getRentalUnitText(item.rental_unit) }} × {{ item.rental_duration || 0 }}{{ getRentalUnitText(item.rental_unit) }}</text>
              <text class="rental-info-deposit">押金 ¥{{ Number(item.deposit || 0).toFixed(2) }}</text>
            </view>
          </view>
          <view class="goods-right">
            <view class="goods-price">¥{{ getItemPayAmount(item).toFixed(2) }}</view>
            <view class="goods-quantity">x{{ item.quantity }}</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 金额明细 -->
    <view class="section amount-section">
      <view class="amount-caption" :class="{ final: hasFinalPricing }">
        {{ amountCaption }}
      </view>
      <view class="amount-row">
        <text class="amount-label">{{ hasFinalPricing ? '商品金额' : '预估商品金额' }}</text>
        <text class="amount-value">¥{{ displayGoodsAmount.toFixed(2) }}</text>
      </view>
      <view v-if="displayDepositAmount > 0" class="amount-row deposit">
        <text class="amount-label">{{ hasFinalPricing ? '押金' : '预估押金' }}</text>
        <text class="amount-value">¥{{ displayDepositAmount.toFixed(2) }}</text>
      </view>
      <!-- 服务订单不计配送费，隐藏配送费明细行 -->
      <view v-if="!hasServiceItems" class="amount-row">
        <text class="amount-label">{{ hasFinalPricing ? '配送费' : '预估配送费' }}</text>
        <text class="amount-value">¥{{ displayDeliveryFee.toFixed(2) }}</text>
      </view>
      <view class="amount-row coupon" @click="openCouponPopup">
        <text class="amount-label">优惠券</text>
        <text class="amount-value" :class="{ 'coupon-active': couponSummaryText.startsWith('-') }">
          {{ couponSummaryText }} <text class="coupon-arrow">›</text>
        </text>
      </view>
      <view v-if="displayDiscountAmount > 0" class="amount-row discount">
        <text class="amount-label">券抵扣</text>
        <text class="amount-value">-¥{{ displayDiscountAmount.toFixed(2) }}</text>
      </view>
      <view class="amount-row total">
        <text class="amount-label">{{ hasFinalPricing ? '合计' : '预估合计' }}</text>
        <text class="amount-value">¥{{ displayTotalAmount.toFixed(2) }}</text>
      </view>
    </view>

    <!-- 优惠券选择弹窗 -->
    <view v-if="couponPopupVisible" class="coupon-selector-mask" @click="couponPopupVisible = false">
      <view class="coupon-selector-panel" @click.stop>
        <view class="coupon-selector-header">
          <text class="coupon-selector-title">选择优惠券</text>
          <text class="coupon-selector-close" @click="couponPopupVisible = false">×</text>
        </view>
        <scroll-view class="coupon-selector-list" scroll-y>
          <view v-if="usableCoupons.length === 0" class="coupon-empty">
            <text class="coupon-empty-icon">🎟️</text>
            <text class="coupon-empty-text">暂无可用优惠券</text>
          </view>
          <view
            v-for="coupon in usableCoupons"
            :key="coupon.user_coupon_id"
            class="coupon-option"
            :class="{ selected: selectedCoupon?.user_coupon_id === coupon.user_coupon_id }"
            @click="selectCoupon(coupon)"
          >
            <view class="coupon-option-left">
              <text class="coupon-option-value">¥{{ Number(coupon.discount).toFixed(2) }}</text>
              <text class="coupon-option-threshold">
                {{ Number(coupon.threshold_amount) > 0 ? `满${Number(coupon.threshold_amount).toFixed(0)}可用` : '无门槛' }}
              </text>
            </view>
            <view class="coupon-option-right">
              <text class="coupon-option-name">{{ coupon.name }}</text>
              <text class="coupon-option-validity">有效期至 {{ String(coupon.expired_at || '').replace('T', ' ').slice(0, 10) }}</text>
            </view>
            <view class="coupon-option-check" :class="{ checked: selectedCoupon?.user_coupon_id === coupon.user_coupon_id }"></view>
          </view>
          <view
            v-if="usableCoupons.length > 0"
            class="coupon-option none"
            :class="{ selected: !selectedCoupon }"
            @click="selectCoupon(null)"
          >
            <text class="coupon-option-name">不使用优惠券</text>
            <view class="coupon-option-check" :class="{ checked: !selectedCoupon }"></view>
          </view>
        </scroll-view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <view class="total-info">
        <text class="total-label">{{ hasFinalPricing ? '最终实付' : '预计实付' }}</text>
        <text class="total-amount">¥{{ displayPayAmount.toFixed(2) }}</text>
        <text v-if="displayDepositAmount > 0" class="total-deposit-tip">(含押金 ¥{{ displayDepositAmount.toFixed(2) }})</text>
      </view>
      <view class="submit-btn" :class="{ disabled: submitting }" @click="submitOrder">
        {{ submitting ? '处理中...' : '提交订单' }}
      </view>
    </view>

    <view v-if="hasRentalItems" class="rental-notice">
      租赁订单：押金将在归还商品并由商家验机后原路退还，若有损坏将扣除相应费用
    </view>

    <!-- 服务对象：多档案选择弹层 -->
    <view v-if="recordPickerVisible" class="picker-mask" @click="recordPickerVisible = false">
      <view class="picker-panel" @click.stop>
        <view class="picker-header">
          <text class="picker-title">选择服务对象</text>
          <text class="picker-close" @click="recordPickerVisible = false">×</text>
        </view>
        <scroll-view scroll-y class="picker-body">
          <view
            v-for="r in healthRecordList"
            :key="r.id"
            class="picker-item"
            :class="{ active: healthRecord?.id === r.id }"
            @tap="selectRecordFromList(r)"
          >
            <view class="picker-item-main">
              <view class="picker-item-top">
                <text class="picker-item-name">{{ r.real_name || '未命名档案' }}</text>
                <text class="picker-item-rel">{{ relationText(r.relation) }}</text>
              </view>
              <view class="picker-item-sub">{{ buildRecordSummary(r) }}</view>
            </view>
            <text v-if="healthRecord?.id === r.id" class="picker-item-check">✓</text>
          </view>
          <view class="picker-create" @tap="goCreateRecordFromPicker">
            ＋ 新增健康档案
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { createOrder, getStoreDeliveryRules } from '../../api/store'
import { createUserAddress, getUserAddresses } from '../../api'
import { getUserHealthRecord, listUserHealthRecords } from '../../api/health'
import { getOrderUsableCoupons } from '../../api/coupon'
import type { OrderUsableCoupon } from '../../types/coupon'
import { useCartStore, getItemPayAmount, getItemDeposit, getRentalUnitText } from '../../stores/cart'
import type { CartItem } from '../../stores/cart'
import { useAnalytics } from '@utils/analytics'
import { useAuth } from '../../utils/useAuth'
import { parseStoreEntryOptions } from '@utils/storeEntry'
import type { CreateOrderRequest, HealthRecord, Order, StoreDeliveryRules, UserAddress } from '@types'
import { BrandAsset } from '../../utils/constants'
import {
  STORE_SELECTED_ADDRESS_ID_KEY,
  buildUserAddressPayload,
  chooseWechatAddress,
  formatUserAddress,
  getPreferredAddress,
  isChooseAddressCancel,
  mapWechatAddressToUserAddress
} from '../../utils/address'

const cartStore = useCartStore()
const { trackPageView, trackPayment } = useAnalytics()

const remark = ref('')
const entrySource = ref('scan')
const submitting = ref(false)
const finalOrder = ref<Order | null>(null)
const finalAmountAdjusted = ref(false)
const selectedAddress = ref<UserAddress | null>(null)
const addressList = ref<UserAddress[]>([])
const loadingAddresses = ref(false)
const isBuyNow = ref(false)
const deliveryConfig = ref<StoreDeliveryRules>({
  enabled: false,
  base_fee: 0,
  free_delivery_amount: 0,
  max_distance: 0,
  distance_rules: [] as { min_distance: number; max_distance: number; fee: number }[]
})

const deliveryRules = ref<{ distance: number; fee: number; label: string }[]>([])
const deliveryDistance = ref(0)
const deliveryDistanceIndex = ref(0)

/* ============ 优惠券（PRD V2.0 阶段二） ============ */
const usableCoupons = ref<OrderUsableCoupon[]>([])
const selectedCoupon = ref<OrderUsableCoupon | null>(null)
const couponPopupVisible = ref(false)
const couponsLoaded = ref(false)

const couponDiscount = computed(() => Number(selectedCoupon.value?.discount || 0))

const couponSummaryText = computed(() => {
  if (finalOrder.value) {
    return Number(finalOrder.value.discount_amount) > 0 ? `-¥${Number(finalOrder.value.discount_amount).toFixed(2)}` : '未使用'
  }
  if (selectedCoupon.value) {
    return `-¥${couponDiscount.value.toFixed(2)}`
  }
  if (usableCoupons.value.length > 0) {
    return `${usableCoupons.value.length} 张可用`
  }
  return couponsLoaded.value ? '暂无可用' : '暂无可用'
})

const productIds = computed(() => displayItems.value.map(item => Number(item.product_id)).filter(id => id > 0))

async function loadUsableCoupons() {
  if (goodsAmount.value <= 0 || productIds.value.length === 0) {
    usableCoupons.value = []
    couponsLoaded.value = true
    return
  }
  try {
    const res = await getOrderUsableCoupons(Number(goodsAmount.value.toFixed(2)), productIds.value)
    usableCoupons.value = res.list || []
    // 已选中的券失效时自动清除
    if (selectedCoupon.value) {
      const matched = usableCoupons.value.find(c => c.user_coupon_id === selectedCoupon.value?.user_coupon_id)
      selectedCoupon.value = matched || null
    }
  } catch (_e) {
    usableCoupons.value = []
  } finally {
    couponsLoaded.value = true
  }
}

function openCouponPopup() {
  if (finalOrder.value) {
    uni.showToast({ title: '订单已创建，无法修改优惠券', icon: 'none' })
    return
  }
  couponPopupVisible.value = true
}

function selectCoupon(coupon: OrderUsableCoupon | null) {
  if (coupon && Number(coupon.discount) <= 0) {
    return
  }
  selectedCoupon.value = coupon
  couponPopupVisible.value = false
}

const displayItems = computed<CartItem[]>(() => {
  if (isBuyNow.value && cartStore.buyNowItem) {
    return [cartStore.buyNowItem]
  }
  return cartStore.items
})

const hasRentalItems = computed(() => displayItems.value.some(item => Number(item.sale_type) === 2))

// 服务订单（康养套餐/陪诊服务）：必须绑定服务对象（健康档案）
const hasServiceItems = computed(() => displayItems.value.some(item => [3, 4].includes(Number(item.product_type))))
// 实物订单（普通商品/实物租赁）：与服务商品互斥，禁止混单支付
const hasGoodsItems = computed(() => displayItems.value.some(item => [1, 2].includes(Number(item.product_type))))

// 支付成功/取消后跳转对应的订单列表页（订单含服务项→服务订单页，否则→实物订单页）
const orderListUrl = computed(() =>
  hasServiceItems.value ? '/pages/store/my-orders-service' : '/pages/store/my-orders-goods'
)

const healthRecord = ref<HealthRecord | null>(null)
const loadingRecord = ref(false)
// 多档案列表支持（账号下的全部健康档案）
const healthRecordList = ref<HealthRecord[]>([])
const recordPickerVisible = ref(false)

const recordSummaryText = computed(() => {
  const record = healthRecord.value
  if (!record) return ''
  const gender = Number(record.gender) === 2 ? '女' : '男'
  const birth = record.birth_date ? String(record.birth_date).slice(0, 10) : ''
  const age = birth ? `${new Date().getFullYear() - Number(birth.slice(0, 4))}岁` : '年龄未知'
  const tags = (record.chronic_tags || []).slice(0, 2).join('、')
  return [gender, age, tags].filter(Boolean).join(' · ')
})

function relationText(relation?: number) {
  const map: Record<number, string> = { 1: '本人', 2: '父母', 3: '其他亲属' }
  return relation == null ? '未设置' : (map[relation] || '其他亲属')
}

function buildRecordSummary(record: HealthRecord) {
  const gender = Number(record.gender) === 2 ? '女' : '男'
  const birth = record.birth_date ? String(record.birth_date).slice(0, 10) : ''
  const age = birth ? `${new Date().getFullYear() - Number(birth.slice(0, 4))}岁` : '年龄未知'
  const tags = (record.chronic_tags || []).slice(0, 2).join('、')
  return [gender, age, tags].filter(Boolean).join(' · ')
}

async function loadHealthRecord() {
  try {
    loadingRecord.value = true
    // 优先拉取账号下的档案列表（多档案），失败或为空时再退回单档案接口
    try {
      const list = await listUserHealthRecords()
      healthRecordList.value = Array.isArray(list) ? list : []
      // 默认选中：当前 healthRecord（若在列表中）、否则取列表第 1 条
      const currentId = healthRecord.value?.id
      const matched = healthRecordList.value.find(r => r.id === currentId)
      if (matched) {
        healthRecord.value = matched
      } else if (healthRecordList.value.length > 0) {
        healthRecord.value = healthRecordList.value[0]
      } else {
        healthRecord.value = null
      }
      return
    } catch (err) {
      console.warn('[Confirm] listUserHealthRecords failed, fallback to single record', err)
    }
    healthRecord.value = await getUserHealthRecord()
    if (healthRecord.value) {
      healthRecordList.value = [healthRecord.value]
    }
  } catch (error) {
    console.error('获取健康档案失败:', error)
    healthRecord.value = null
    healthRecordList.value = []
  } finally {
    loadingRecord.value = false
  }
}

function goHealthRecord() {
  // 若已有多档案，展示选择弹层；否则直接跳档案页
  if (healthRecordList.value.length > 0) {
    recordPickerVisible.value = true
  } else {
    uni.navigateTo({ url: '/pages/store/my-health' })
  }
}

function selectRecordFromList(record: HealthRecord) {
  healthRecord.value = record
  recordPickerVisible.value = false
}

function goCreateRecordFromPicker() {
  recordPickerVisible.value = false
  uni.navigateTo({ url: '/pages/store/health-record-edit' })
}

const totalDeposit = computed(() => displayItems.value.reduce((sum, item) => sum + getItemDeposit(item), 0))

function applyEntryOptions(options?: Record<string, any>) {
  const { source } = parseStoreEntryOptions(options)
  entrySource.value = source
  isBuyNow.value = !!(options && Number(options.buy_now) === 1)
}

onLoad((options) => {
  applyEntryOptions(options as Record<string, any> | undefined)
})

onShow(async () => {
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  applyEntryOptions(currentPage?.options)
  cartStore.restoreFromStorage()

  // 服务订单不计配送费，也无需配送规则/档位；仅在实物订单时加载配送规则
  if (!hasServiceItems.value) {
    await loadDeliveryRules()
  }

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (authed) {
    await loadAddresses(readSelectedAddressId())
    if (hasServiceItems.value) {
      await loadHealthRecord()
    }
    void loadUsableCoupons()
  }
  void trackPageView('store_confirm', entrySource.value)
})

async function loadDeliveryRules() {
  try {
    const rules = await getStoreDeliveryRules()
    deliveryConfig.value = rules
    deliveryRules.value = (rules.distance_rules || []).map((item) => ({
      distance: item.max_distance,
      fee: item.fee,
      label: `${item.min_distance}-${item.max_distance}km · 配送费 ¥${item.fee.toFixed(2)}`
    }))

    if (deliveryRules.value.length === 0 && rules.max_distance > 0) {
      deliveryRules.value = [{
        distance: rules.max_distance,
        fee: rules.base_fee,
        label: `0-${rules.max_distance}km · 配送费 ¥${rules.base_fee.toFixed(2)}`
      }]
    }

    if (deliveryRules.value.length > 0) {
      deliveryDistanceIndex.value = 0
      deliveryDistance.value = deliveryRules.value[0].distance
    }
  } catch (error) {
    console.error('获取配送规则失败:', error)
    deliveryRules.value = []
  }
}

function readSelectedAddressId() {
  const value = uni.getStorageSync(STORE_SELECTED_ADDRESS_ID_KEY)
  if (!value) return undefined
  uni.removeStorageSync(STORE_SELECTED_ADDRESS_ID_KEY)
  const addressId = Number(value)
  return Number.isFinite(addressId) && addressId > 0 ? addressId : undefined
}

async function loadAddresses(preferredAddressId?: number) {
  try {
    loadingAddresses.value = true
    const list = await getUserAddresses()
    addressList.value = list
    selectedAddress.value = getPreferredAddress(
      list,
      preferredAddressId || selectedAddress.value?.id
    )
  } catch (error) {
    console.error('获取地址列表失败:', error)
    addressList.value = []
    selectedAddress.value = null
  } finally {
    loadingAddresses.value = false
  }
}

async function goAddressList() {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    return uni.showToast({ title: '登录失败，请重试', icon: 'none' })
  }
  uni.navigateTo({
    url: `/pages/store/address-list?selected_id=${selectedAddress.value?.id || ''}`
  })
}

async function goAddressEdit() {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    return uni.showToast({ title: '登录失败，请重试', icon: 'none' })
  }
  uni.navigateTo({
    url: '/pages/store/address-edit'
  })
}

async function importWechatAddress() {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    return uni.showToast({ title: '登录失败，请重试', icon: 'none' })
  }

  try {
    const importedAddress = mapWechatAddressToUserAddress(await chooseWechatAddress())
    const createdAddress = await createUserAddress(buildUserAddressPayload({
      ...importedAddress,
      is_default: !addressList.value.some((item) => item.is_default)
    }))
    selectedAddress.value = createdAddress
    await loadAddresses(createdAddress.id)
    uni.showToast({ title: '导入地址成功', icon: 'success' })
  } catch (error: any) {
    if (isChooseAddressCancel(error)) {
      return
    }
    uni.showToast({ title: error.message || '导入地址失败', icon: 'none' })
  }
}

const selectedDeliveryRule = computed(() => deliveryRules.value[deliveryDistanceIndex.value] || null)
const selectedAddressText = computed(() => formatUserAddress(selectedAddress.value))

const goodsAmount = computed(() => displayItems.value.reduce((sum, item) => {
  if (Number(item.sale_type) === 2) {
    return sum + Number(item.rental_price || 0) * Number(item.rental_duration || 0) * Number(item.quantity || 1)
  }
  return sum + Number(item.price || 0) * Number(item.quantity || 1)
}, 0))

const hasFinalPricing = computed(() => !!finalOrder.value)
const displayGoodsAmount = computed(() => finalOrder.value?.total_amount ?? goodsAmount.value)
const displayDeliveryFee = computed(() => finalOrder.value?.delivery_fee ?? deliveryFee.value)
const displayDepositAmount = computed(() => finalOrder.value?.total_deposit ?? totalDeposit.value)
const displayDiscountAmount = computed(() => Number(finalOrder.value?.discount_amount ?? couponDiscount.value) || 0)
const displayTotalAmount = computed(() => Math.max(0, displayGoodsAmount.value + displayDepositAmount.value + displayDeliveryFee.value - displayDiscountAmount.value))
const displayPayAmount = computed(() => finalOrder.value?.pay_amount ?? payableAmount.value)

const amountCaption = computed(() => {
  if (hasFinalPricing.value) {
    return finalAmountAdjusted.value
      ? '订单已按后端最终结算更新，支付将以以下金额为准'
      : '订单已创建，以下金额以后端订单返回为准'
  }

  return '以下金额为提交前预估，创建订单后会自动切换为最终支付金额'
})

const deliveryFee = computed(() => {
  // 服务订单不计配送费（不受商家配送参数约束）
  if (hasServiceItems.value) {
    return 0
  }

  if (goodsAmount.value >= deliveryConfig.value.free_delivery_amount && deliveryConfig.value.free_delivery_amount > 0) {
    return 0
  }

  if (!selectedDeliveryRule.value) {
    return deliveryConfig.value.base_fee || 0
  }

  return selectedDeliveryRule.value.fee
})

const totalAmount = computed(() => {
  return goodsAmount.value + totalDeposit.value + deliveryFee.value
})

const payableAmount = computed(() => Math.max(0, totalAmount.value - couponDiscount.value))

// 商品金额变化时刷新可用券（订单创建前）
watch(goodsAmount, () => {
  if (!finalOrder.value) {
    void loadUsableCoupons()
  }
})

watch(
  [
    () => selectedAddress.value?.id,
    () => cartStore.totalAmount,
    () => cartStore.buyNowItem,
    isBuyNow
  ],
  () => {
    if (!finalOrder.value) {
      return
    }

    finalOrder.value = null
    finalAmountAdjusted.value = false
  }
)

function isAmountDifferent(currentValue: number, nextValue: number) {
  return Math.abs(currentValue - nextValue) > 0.009
}

function applyFinalOrderAmount(order: Order) {
  finalAmountAdjusted.value = (
    isAmountDifferent(goodsAmount.value, order.total_amount) ||
    isAmountDifferent(deliveryFee.value, order.delivery_fee) ||
    isAmountDifferent(payableAmount.value, order.pay_amount)
  )
  finalOrder.value = order
}

async function submitOrder() {
  if (submitting.value) {
    return
  }

  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    return uni.showToast({ title: '登录失败，请重试', icon: 'none' })
  }

  if (!selectedAddress.value) {
    return uni.showToast({ title: loadingAddresses.value ? '地址加载中，请稍后' : '请选择收货地址', icon: 'none' })
  }

  // 禁止实物商品与服务商品混单支付
  if (hasServiceItems.value && hasGoodsItems.value) {
    return uni.showToast({ title: '订单不能同时包含实物商品与服务商品，请拆分下单', icon: 'none' })
  }

  // 服务订单不受商家配送参数约束，跳过配送档位/距离校验
  if (!hasServiceItems.value) {
    if (!deliveryRules.value.length) {
      return uni.showToast({ title: '当前暂无可选配送档位', icon: 'none' })
    }
    if (!selectedDeliveryRule.value) {
      return uni.showToast({ title: '请选择配送距离档位', icon: 'none' })
    }
    if (deliveryDistance.value > deliveryConfig.value.max_distance) {
      return uni.showToast({ title: '已超出商家配送范围', icon: 'none' })
    }
  }

  if (!/^1\d{10}$/.test(selectedAddress.value.phone || '')) {
    return uni.showToast({ title: '请输入正确的联系电话', icon: 'none' })
  }

  if (displayItems.value.length === 0) {
    return uni.showToast({ title: '商品列表为空', icon: 'none' })
  }

  // 服务订单硬性规则：必须指定服务对象（健康档案）
  if (hasServiceItems.value) {
    if (loadingRecord.value) {
      return uni.showToast({ title: '健康档案加载中，请稍后', icon: 'none' })
    }
    if (!healthRecord.value) {
      uni.showModal({
        title: '请先创建健康档案',
        content: '服务订单需指定服务对象，是否前往创建健康档案？',
        confirmText: '去建档',
        confirmColor: '#2B5CE3',
        success: (res) => {
          if (res.confirm) {
            uni.navigateTo({ url: '/pages/store/my-health' })
          }
        }
      })
      return
    }
  }

  const orderData: CreateOrderRequest = {
    items: displayItems.value.map(item => ({
      product_id: item.product_id,
      spec_info: item.specs,
      quantity: item.quantity,
      rental_duration: Number(item.sale_type) === 2 ? Number(item.rental_duration || 0) : undefined
    })),
    delivery_distance: deliveryDistance.value,
    user_coupon_id: selectedCoupon.value?.user_coupon_id || undefined,
    record_id: hasServiceItems.value && healthRecord.value ? healthRecord.value.id : undefined,
    address_id: selectedAddress.value?.id,
    delivery_address: selectedAddressText.value,
    contact_name: selectedAddress.value?.name || '',
    contact_phone: selectedAddress.value?.phone || '',
    remark: remark.value
  }

  try {
    submitting.value = true
    uni.showLoading({ title: '创建订单中...' })

    const res = await createOrder(orderData)
    applyFinalOrderAmount(res.order)

    if (res.pay_params) {
      uni.requestPayment({
        provider: 'wxpay',
        timeStamp: res.pay_params.timeStamp,
        nonceStr: res.pay_params.nonceStr,
        package: res.pay_params.package,
        signType: res.pay_params.signType,
        paySign: res.pay_params.paySign,
        success: async () => {
          uni.hideLoading()
          clearOrderCartAfterSuccess()
          uni.showToast({ title: '支付成功', icon: 'success' })
          await trackPayment(res.order.id, res.order.pay_amount)

          setTimeout(() => {
            uni.redirectTo({
              url: `${orderListUrl.value}?status=2`
            })
          }, 1500)
        },
        fail: (err) => {
          uni.hideLoading()
          if (err.errMsg?.includes('cancel')) {
            uni.showToast({ title: '支付已取消', icon: 'none' })

            setTimeout(() => {
              uni.redirectTo({
                url: `${orderListUrl.value}?status=1`
              })
            }, 1500)
          } else {
            uni.showToast({ title: '支付失败', icon: 'none' })
            setTimeout(() => {
              uni.redirectTo({
                url: `${orderListUrl.value}?status=1`
              })
            }, 1500)
          }
        },
        complete: () => {
          submitting.value = false
        }
      })
    } else {
      uni.hideLoading()
      clearOrderCartAfterSuccess()
      await trackPayment(res.order.id, res.order.pay_amount)

      if ((res.order.pay_amount ?? 0) > 0) {
        // 需要支付，但当前没有可用支付参数（凭证未配置或商家未完成支付设置）
        const hint = res.pay_hint || '商家支付配置未完成，请稍后在「待付款」中重新发起支付，或联系商家确认'
        uni.showModal({
          title: '订单已创建，请稍后支付',
          content: hint,
          showCancel: false,
          confirmText: '查看订单',
          confirmColor: '#2B5CE3',
          success: () => {
            setTimeout(() => {
              uni.redirectTo({
                url: `${orderListUrl.value}?status=1`
              })
            }, 300)
          }
        })
      } else {
        uni.showToast({ title: '订单创建成功', icon: 'success' })
        setTimeout(() => {
          uni.redirectTo({
            url: orderListUrl.value
          })
        }, 1500)
      }
      submitting.value = false
    }
  } catch (error: any) {
    uni.hideLoading()
    submitting.value = false
    uni.showToast({ title: error.message || '创建订单失败', icon: 'none' })
  }
}

function clearOrderCartAfterSuccess() {
  if (isBuyNow.value) {
    cartStore.clearBuyNowItem()
    return
  }
  cartStore.clearCart()
}
</script>

<style scoped>
.confirm-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}

.section {
  background: #ffffff;
  margin: 24rpx;
  border-radius: 16rpx;
  padding: 32rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.section-title-row .section-title {
  margin-bottom: 0;
}
.section-meta {
  font-size: 24rpx;
  color: #64748b;
}

.delivery-type-selector {
  display: flex;
  gap: 24rpx;
  margin-bottom: 24rpx;
}

.delivery-mode-empty {
  padding: 24rpx;
  border-radius: 12rpx;
  background: #fff7e6;
  color: #d46b08;
  font-size: 28rpx;
}

.type-item {
  flex: 1;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 2rpx solid transparent;
}

.type-item.selected {
  background: #e6f0ff;
  border-color: #007AFF;
}

.type-icon {
  font-size: 48rpx;
  margin-bottom: 8rpx;
}

.type-name {
  font-size: 28rpx;
  color: #1a1a1a;
}

.delivery-form {
  margin-bottom: 24rpx;
}

.address-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
}

.address-card.empty {
  border: 2rpx dashed #d9e7ff;
  background: #f7fbff;
}

.address-main,
.address-empty {
  flex: 1;
  min-width: 0;
}

.address-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 10rpx;
  flex-wrap: wrap;
}

.address-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.address-phone {
  font-size: 26rpx;
  color: #666666;
}

.address-default-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: #e6f0ff;
  color: #007AFF;
  font-size: 22rpx;
}

.address-detail {
  font-size: 26rpx;
  line-height: 1.5;
  color: #333333;
  word-break: break-all;
}

.address-empty-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.address-empty-desc {
  display: block;
  font-size: 24rpx;
  color: #999999;
  line-height: 1.5;
}

.address-arrow {
  font-size: 28rpx;
  color: #c0c4cc;
  flex-shrink: 0;
}

.record-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
}

.record-card.empty {
  border: 2rpx dashed #d9e7ff;
  background: #f7fbff;
}

.record-main {
  flex: 1;
  min-width: 0;
}

.record-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 10rpx;
}

.record-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.record-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: #e6f9f0;
  color: #0bab5e;
  font-size: 22rpx;
}

.record-rel-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
}

.record-detail {
  font-size: 26rpx;
  line-height: 1.5;
  color: #333333;
  word-break: break-all;
}

.record-empty {
  flex: 1;
  min-width: 0;
}

.record-empty-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.record-empty-desc {
  display: block;
  font-size: 24rpx;
  color: #999999;
  line-height: 1.5;
}

.address-actions {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.address-action-btn {
  flex: 1;
  height: 72rpx;
  border-radius: 999rpx;
  background: #007AFF;
  color: #ffffff;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.address-action-btn.secondary {
  background: #eef3ff;
  color: #007AFF;
}

.form-item {
  margin-bottom: 20rpx;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.form-input {
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
}

.picker-value {
  height: 80rpx;
  background: #f8f9fa;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

.picker-value.disabled {
  color: #999999;
}

.distance-tip {
  font-size: 24rpx;
  color: #ff4d4f;
  margin-top: 12rpx;
  padding: 8rpx 16rpx;
  background: #fff2f0;
  border-radius: 8rpx;
}

.pickup-form {
  margin-bottom: 24rpx;
}

.pickup-card {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
}

.pickup-card.empty {
  border: 2rpx dashed #d9e7ff;
  background: #f7fbff;
}

.pickup-main,
.pickup-empty {
  flex: 1;
  min-width: 0;
}

.pickup-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 10rpx;
  flex-wrap: wrap;
}

.pickup-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.pickup-default-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: #e6f0ff;
  color: #007AFF;
  font-size: 22rpx;
}

.pickup-address {
  font-size: 26rpx;
  line-height: 1.5;
  color: #333333;
  word-break: break-all;
}

.pickup-empty-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.pickup-empty-desc {
  display: block;
  font-size: 24rpx;
  color: #999999;
  line-height: 1.5;
}

.pickup-selector-mask {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  display: flex;
  justify-content: center;
  align-items: flex-end;
}

.pickup-selector-panel {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx 24rpx 0 0;
  padding: 24rpx;
  max-height: 70vh;
}

.pickup-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.pickup-selector-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.pickup-selector-close {
  font-size: 44rpx;
  color: #999999;
  line-height: 1;
  padding: 8rpx 12rpx;
}

.pickup-selector-list {
  max-height: 58vh;
}

.pickup-selector-item {
  padding: 20rpx 20rpx;
  border-radius: 16rpx;
  background: #f8f9fb;
  margin-bottom: 16rpx;
}

.pickup-selector-item.selected {
  background: #e6f0ff;
}

.pickup-selector-item-top {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
}

.pickup-selector-item-title {
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex: 1;
  min-width: 0;
}

.pickup-selector-item-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pickup-selector-tag {
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: rgba(0, 122, 255, 0.12);
  color: #007AFF;
  font-size: 22rpx;
  flex-shrink: 0;
}

.pickup-selector-nav {
  font-size: 26rpx;
  color: #007AFF;
  flex-shrink: 0;
}

.pickup-selector-item-address {
  margin-top: 10rpx;
  font-size: 26rpx;
  color: #666666;
  line-height: 1.5;
}

.remark-form {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
}

.goods-list {
  display: flex;
  flex-direction: column;
}

.promo-banner {
  border-radius: 18rpx;
  background: #f8f9fb;
  padding: 24rpx;
}

.promo-banner.active {
  background: #fff7e8;
}

.promo-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.promo-desc {
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.promo-empty {
  border-radius: 18rpx;
  background: #f8f9fb;
  padding: 24rpx;
  font-size: 24rpx;
  color: #8a9099;
}

.promo-rule-list {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
  margin-top: 20rpx;
}

.promo-rule-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 20rpx 24rpx;
  border-radius: 16rpx;
  background: #f8f9fb;
  font-size: 24rpx;
  color: #5f6570;
}

.promo-rule-tag {
  flex-shrink: 0;
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  background: #eef1f5;
  color: #8a9099;
}

.promo-rule-tag.active {
  background: #ffe7ba;
  color: #d46b08;
}

.goods-item {
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.goods-item:last-child {
  border-bottom: none;
}

.goods-image {
  width: 140rpx;
  height: 140rpx;
  flex-shrink: 0;
  border-radius: 12rpx;
  background: #f0f0f0;
}

.goods-info {
  flex: 1;
  min-width: 0;
  padding-top: 4rpx;
}

.goods-name {
  font-size: 30rpx;
  line-height: 1.4;
  color: #1a1a1a;
  margin-bottom: 8rpx;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.goods-spec {
  font-size: 26rpx;
  line-height: 1.4;
  color: #999999;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.goods-rental-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #ff9500;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
  -webkit-line-clamp: unset;
}

.goods-wellness-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #22c55e;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
  -webkit-line-clamp: unset;
}

.goods-escort-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #6366f1;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
  -webkit-line-clamp: unset;
}

.goods-info-tag {
  display: inline-block;
  font-size: 20rpx;
  color: #ffffff;
  background: #64748b;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 8rpx;
  vertical-align: middle;
  -webkit-line-clamp: unset;
}

.goods-rental-info {
  display: flex;
  flex-direction: column;
  margin-top: 8rpx;
  padding: 8rpx 12rpx;
  background: #fff7e6;
  border-radius: 8rpx;
  font-size: 24rpx;
  overflow: visible;
  -webkit-line-clamp: unset;
  -webkit-box-orient: unset;
}

.rental-info-text {
  color: #ff9500;
}

.rental-info-deposit {
  color: #666666;
  margin-top: 4rpx;
}

.goods-right {
  width: 148rpx;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
  text-align: right;
  padding-top: 4rpx;
}

.goods-price {
  font-size: 28rpx;
  color: #1a1a1a;
  margin-bottom: 8rpx;
}

.goods-quantity {
  font-size: 26rpx;
  color: #999999;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  padding: 16rpx 0;
  font-size: 28rpx;
}

.amount-caption {
  margin-bottom: 12rpx;
  padding: 18rpx 20rpx;
  border-radius: 12rpx;
  background: #f5f7fa;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.amount-caption.final {
  background: #eef6ff;
  color: #0056cc;
}

.amount-label {
  color: #666666;
}

.amount-value {
  color: #1a1a1a;
}

.amount-row.discount .amount-value {
  color: #ff4d4f;
}

.amount-row.deposit .amount-label {
  color: #ff9500;
}

.amount-row.coupon .amount-value {
  color: #ff4d4f;
}

.amount-row.coupon .coupon-active {
  color: #ff4d4f;
  font-weight: 600;
}

.coupon-arrow {
  color: #c0c4cc;
  margin-left: 4rpx;
}

.coupon-selector-mask {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  display: flex;
  justify-content: center;
  align-items: flex-end;
}

.coupon-selector-panel {
  width: 100%;
  background: #ffffff;
  border-radius: 24rpx 24rpx 0 0;
  padding: 24rpx;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
}

.coupon-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.coupon-selector-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.coupon-selector-close {
  font-size: 44rpx;
  color: #999999;
  line-height: 1;
  padding: 8rpx 12rpx;
}

.coupon-selector-list {
  max-height: 58vh;
}

.coupon-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 0;
}

.coupon-empty-icon {
  font-size: 72rpx;
}

.coupon-empty-text {
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #999999;
}

.coupon-option {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  border-radius: 16rpx;
  background: #f8f9fb;
  margin-bottom: 16rpx;
  border: 2rpx solid transparent;
}

.coupon-option.selected {
  background: #e6f0ff;
  border-color: #007AFF;
}

.coupon-option.none {
  justify-content: space-between;
  padding: 20rpx 24rpx;
}

.coupon-option-left {
  width: 160rpx;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12rpx 0;
  border-right: 2rpx dashed #e0e0e0;
}

.coupon-option-value {
  font-size: 40rpx;
  font-weight: 700;
  color: #ff3b30;
  line-height: 1.2;
}

.coupon-option-threshold {
  font-size: 20rpx;
  color: #ff7a45;
  margin-top: 4rpx;
}

.coupon-option-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.coupon-option-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coupon-option-validity {
  font-size: 22rpx;
  color: #999999;
}

.coupon-option-check {
  width: 36rpx;
  height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #d9d9d9;
  flex-shrink: 0;
  box-sizing: border-box;
}

.coupon-option-check.checked {
  border: none;
  background: #007AFF;
  position: relative;
}

.coupon-option-check.checked::after {
  content: '';
  position: absolute;
  left: 12rpx;
  top: 6rpx;
  width: 8rpx;
  height: 16rpx;
  border: solid #ffffff;
  border-width: 0 4rpx 4rpx 0;
  transform: rotate(45deg);
}

.amount-row.deposit .amount-value {
  color: #ff9500;
}

.amount-row.total {
  border-top: 1rpx solid #f0f0f0;
  padding-top: 24rpx;
  margin-top: 8rpx;
}

.amount-row.total .amount-label {
  font-weight: 600;
  color: #1a1a1a;
}

.amount-row.total .amount-value {
  font-size: 36rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.bottom-bar {
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
}

.total-info {
  flex: 1;
}

.total-label {
  font-size: 26rpx;
  color: #666666;
  margin-right: 8rpx;
}

.total-amount {
  font-size: 40rpx;
  font-weight: 600;
  color: #ff4d4f;
}

.total-deposit-tip {
  display: inline-block;
  font-size: 22rpx;
  color: #ff9500;
  margin-left: 8rpx;
  vertical-align: middle;
}

.submit-btn {
  padding: 24rpx 64rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  color: #ffffff;
}

.submit-btn.disabled {
  opacity: 0.72;
}

.rental-notice {
  margin: 0 24rpx 24rpx;
  padding: 16rpx 24rpx;
  background: #fff7e6;
  color: #d46b08;
  font-size: 24rpx;
  border-radius: 12rpx;
  line-height: 1.5;
}

/* 服务对象多档案选择弹层 */
.picker-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 999;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}
.picker-panel {
  width: 100%;
  max-height: 75vh;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  display: flex;
  flex-direction: column;
}
.picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 32rpx;
  border-bottom: 2rpx solid #f1f5f9;
}
.picker-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}
.picker-close {
  font-size: 44rpx;
  color: #94a3b8;
  line-height: 1;
  padding: 0 16rpx;
}
.picker-body {
  flex: 1;
  min-height: 0;
  padding: 16rpx 24rpx 40rpx;
}
.picker-item {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 24rpx;
  border-radius: 16rpx;
  border: 2rpx solid #e2e8f0;
  margin-bottom: 16rpx;
  background: #fff;
}
.picker-item.active {
  border-color: #2b5ce3;
  background: rgba(43, 92, 227, 0.05);
}
.picker-item-main { flex: 1; min-width: 0; }
.picker-item-top {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 8rpx;
}
.picker-item-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}
.picker-item-rel {
  font-size: 22rpx;
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
}
.picker-item-sub {
  font-size: 26rpx;
  color: #475569;
  line-height: 1.4;
}
.picker-item-check {
  color: #2b5ce3;
  font-size: 32rpx;
  font-weight: 700;
}
.picker-create {
  margin-top: 8rpx;
  padding: 24rpx;
  text-align: center;
  border-radius: 16rpx;
  border: 2rpx dashed #2b5ce3;
  color: #2b5ce3;
  font-size: 28rpx;
  background: rgba(43, 92, 227, 0.04);
}
</style>
