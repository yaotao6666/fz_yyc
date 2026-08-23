<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi, staffStoreApi } from '@/api'
import type { StoreProduct } from '@/types'

const userId = ref<string>('')
// 可选的关联评估记录 ID（URL query 传入）
const assessmentId = ref<number | null>(null)
const symptomDesc = ref('')
const fittingResult = ref('')
const keyword = ref('')
const products = ref<StoreProduct[]>([])
const loading = ref(false)
const submitting = ref(false)
// 已选中商品 id -> 推荐理由（值为 undefined 表示未选中）
const selectedMap = reactive<Record<number, string>>({})

// 按关键字本地过滤商品列表，提升搜索响应速度
const filteredProducts = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return products.value
  return products.value.filter(p => (p.name || '').toLowerCase().includes(kw))
})

const selectedCount = computed(() => Object.keys(selectedMap).length)

// 症状描述、适配结论与推荐商品均需填写完整方可提交
const canSubmit = computed(
  () => symptomDesc.value.trim() !== '' && fittingResult.value.trim() !== '' && selectedCount.value > 0
)

// 销售类型文案：2 租赁，其余一口价
function saleTypeText(saleType?: number) {
  return saleType === 2 ? '租赁' : '一口价'
}

// 商品价格文案：租赁展示租金，一口价展示售价
function priceText(p: StoreProduct) {
  const value = p.sale_type === 2 ? p.rental_price : p.price
  const num = Number(value)
  if (!num) return '价格电询'
  return p.sale_type === 2 ? `¥${num}/期` : `¥${num}`
}

function toggleSelect(p: StoreProduct) {
  if (selectedMap[p.id] !== undefined) {
    delete selectedMap[p.id]
  } else {
    selectedMap[p.id] = ''
  }
}

async function loadProducts() {
  loading.value = true
  try {
    const data: any = await staffStoreApi.getProducts({ page_size: 100 })
    products.value = data?.list || []
  } catch (e: any) {
    console.error('[FittingForm] loadProducts error', e)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!canSubmit.value) {
    uni.showToast({ title: '请填写完整并选择推荐商品', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    await staffHealthApi.createResidentFittingRecommendation(userId.value, {
      assessment_id: assessmentId.value,
      symptom_desc: symptomDesc.value.trim(),
      fitting_result: fittingResult.value.trim(),
      recommended_products: Object.keys(selectedMap).map(id => ({
        product_id: Number(id),
        reason: selectedMap[Number(id)]?.trim() || undefined
      }))
    })
    uni.showToast({ title: '适配建议已生成', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (e: any) {
    console.error('[FittingForm] handleSubmit error', e)
  } finally {
    submitting.value = false
  }
}

function goBack() {
  uni.navigateBack()
}

onLoad((options: any) => {
  userId.value = options?.userId || ''
  if (!userId.value) {
    uni.showToast({ title: '缺少客户ID', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  if (options?.assessmentId) assessmentId.value = Number(options.assessmentId) || null
  loadProducts()
})
</script>

<template>
  <view class="container">
    <!-- 症状/需求描述 -->
    <view class="card section-card">
      <view class="card-title">症状/需求描述</view>
      <textarea
        v-model="symptomDesc"
        class="desc-input"
        placeholder="请描述客户症状、身体状况与辅具使用需求"
        maxlength="500"
      />
    </view>

    <!-- 适配结论 -->
    <view class="card section-card">
      <view class="card-title">适配结论</view>
      <textarea
        v-model="fittingResult"
        class="desc-input"
        placeholder="请给出适配建议与结论（如：建议使用四脚拐杖辅助行走）"
        maxlength="500"
      />
    </view>

    <!-- 推荐商品选择 -->
    <view class="card section-card">
      <view class="card-title-row">
        <text class="card-title">推荐商品</text>
        <text class="selected-count">已选 {{ selectedCount }}</text>
      </view>
      <view class="search-bar">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商品名称"
          confirm-type="search"
        />
      </view>
      <view v-if="loading" class="no-data">商品加载中...</view>
      <view v-else-if="products.length === 0" class="no-data">暂无商品</view>
      <view v-else-if="filteredProducts.length === 0" class="no-data">未找到匹配商品</view>
      <view v-else class="product-list">
        <view v-for="p in filteredProducts" :key="p.id" class="product-item">
          <view class="product-main" @tap="toggleSelect(p)">
            <view class="product-check" :class="{ checked: selectedMap[p.id] !== undefined }">
              <text v-if="selectedMap[p.id] !== undefined" class="check-mark">✓</text>
            </view>
            <view class="product-info">
              <text class="product-name">{{ p.name }}</text>
              <view class="product-meta">
                <text class="product-price">{{ priceText(p) }}</text>
                <text class="product-type" :class="{ rental: p.sale_type === 2 }">{{ saleTypeText(p.sale_type) }}</text>
              </view>
            </view>
          </view>
          <!-- 选中商品后可填写推荐理由 -->
          <view v-if="selectedMap[p.id] !== undefined" class="reason-box">
            <textarea
              v-model="selectedMap[p.id]"
              class="reason-input"
              placeholder="推荐理由（选填）"
              maxlength="200"
            />
          </view>
        </view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="footer-bar">
      <button class="btn-outline-bar" @tap="goBack">返回</button>
      <button class="btn-submit" :disabled="submitting" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '提交适配建议' }}
      </button>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.container {
  padding-bottom: 140rpx;
}
.card-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 20rpx;
}
.card-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
  .card-title { margin-bottom: 0; }
  .selected-count { font-size: 24rpx; color: var(--primary-color); }
}

.section-card {
  margin-bottom: 20rpx;
  padding: 28rpx 24rpx;
}
.desc-input {
  width: 100%;
  height: 200rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}

/* 商品搜索 */
.search-bar {
  margin-bottom: 16rpx;
}
.search-input {
  width: 100%;
  height: 72rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}

.no-data {
  font-size: 26rpx;
  color: #999;
  padding: 16rpx 0;
}

/* 商品列表 */
.product-item {
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
}
.product-main {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
}
.product-check {
  width: 36rpx; height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  margin-right: 16rpx;
  flex-shrink: 0;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  &.checked {
    border-color: var(--primary-color);
    background: var(--primary-color);
  }
  .check-mark { font-size: 24rpx; color: #fff; }
}
.product-info {
  flex: 1;
  min-width: 0;
  .product-name { display: block; font-size: 28rpx; color: #333; }
}
.product-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 8rpx;
  .product-price { font-size: 26rpx; color: var(--primary-color); font-weight: 600; }
  .product-type {
    font-size: 22rpx;
    color: var(--info-color);
    background: #e6f4ff;
    padding: 2rpx 12rpx;
    border-radius: 8rpx;
    &.rental {
      color: var(--warning-color);
      background: #fff7e6;
    }
  }
}
.reason-box {
  padding-bottom: 20rpx;
}
.reason-input {
  width: 100%;
  height: 120rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  font-size: 26rpx;
  box-sizing: border-box;
}

/* 底部操作栏 */
.footer-bar {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  display: flex;
  gap: 20rpx;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}
.btn-submit {
  flex: 1.4;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &[disabled] { opacity: 0.5; }
}
.btn-outline-bar {
  flex: 1;
  font-size: 30rpx;
  color: var(--primary-color);
  background: #fff;
  border: 2rpx solid var(--primary-color);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
}
</style>
