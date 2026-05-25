<script setup lang="ts">
import { computed, ref } from 'vue'
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getSpOrderDetail } from '@/api/sp'
import type { SpOrder, SpOrderItem } from '@/types/sp'
import { SpDeliveryTypeText, SpOrderStatusText } from '@/types/sp'
import { formatAmount, formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const order = ref<SpOrder | null>(null)

const orderId = computed(() => Number(route.params.id || 0))

const statusSummary = computed(() => {
  if (!order.value) {
    return '订单详情'
  }
  return SpOrderStatusText[order.value.status] || '未知状态'
})

function formatOptionalDateTime(value?: string) {
  return value ? formatDateTime(value) : '-'
}

function resolveSpecText(item: SpOrderItem) {
  return item.specs || item.spec_info || ''
}

function resolveDeliveryAddress() {
  return order.value?.delivery_info?.address || order.value?.delivery_address || '-'
}

function resolveContactInfo() {
  const contactName = order.value?.delivery_info?.contact_name || order.value?.contact_name
  const contactPhone = order.value?.delivery_info?.contact_phone || order.value?.contact_phone
  if (!contactName && !contactPhone) {
    return '-'
  }
  return [contactName, contactPhone].filter(Boolean).join(' ')
}

async function loadOrderDetail() {
  if (!orderId.value) {
    return
  }

  loading.value = true
  try {
    order.value = await getSpOrderDetail(orderId.value)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadOrderDetail()
})
</script>

<template>
  <div class="page-shell" v-loading="loading">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">订单详情</h1>
        <p class="page-subtitle">查看订单状态、商品明细、用户信息和支付核销信息。</p>
      </div>
      <el-button @click="goBack">返回列表</el-button>
    </div>

    <template v-if="order">
      <el-card class="page-card detail-summary-card" shadow="never">
        <div class="summary-title">{{ statusSummary }}</div>
        <div class="summary-subtitle">
          订单号：{{ order.order_no }}，下单时间：{{ formatDateTime(order.created_at) }}
        </div>
      </el-card>

      <div class="detail-grid">
        <el-card class="page-card" shadow="never">
          <template #header>订单信息</template>
          <div class="info-list">
            <div class="info-row"><span>订单编号</span><span>{{ order.order_no }}</span></div>
            <div class="info-row"><span>订单状态</span><span>{{ SpOrderStatusText[order.status] || '未知状态' }}</span></div>
            <div class="info-row"><span>下单时间</span><span>{{ formatDateTime(order.created_at) }}</span></div>
            <div class="info-row"><span>支付时间</span><span>{{ formatOptionalDateTime(order.paid_at) }}</span></div>
            <div class="info-row"><span>支付单号</span><span>{{ order.transaction_id || '-' }}</span></div>
            <div class="info-row"><span>核销码</span><span>{{ order.verify_code || '-' }}</span></div>
            <div class="info-row"><span>核销时间</span><span>{{ formatOptionalDateTime(order.completed_at) }}</span></div>
            <div class="info-row"><span>核销人</span><span>{{ order.completed_by_name || '-' }}</span></div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>商家与用户</template>
          <div class="info-list">
            <div class="info-row"><span>商家名称</span><span>{{ order.merchant?.name || '-' }}</span></div>
            <div class="info-row"><span>商家地址</span><span>{{ order.merchant?.address || '-' }}</span></div>
            <div class="info-row"><span>商家电话</span><span>{{ order.merchant?.contact_phone || order.merchant?.phone || '-' }}</span></div>
            <div class="info-row"><span>用户昵称</span><span>{{ order.user?.nickname || '匿名用户' }}</span></div>
            <div class="info-row"><span>用户手机号</span><span>{{ order.user?.phone || '-' }}</span></div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>配送信息</template>
          <div class="info-list">
            <div class="info-row"><span>配送方式</span><span>{{ SpDeliveryTypeText[order.delivery_type || 0] || '-' }}</span></div>
            <div class="info-row"><span>配送距离</span><span>{{ order.delivery_distance ? `${order.delivery_distance} km` : '-' }}</span></div>
            <div class="info-row"><span>收货地址</span><span>{{ resolveDeliveryAddress() }}</span></div>
            <div class="info-row"><span>联系人</span><span>{{ resolveContactInfo() }}</span></div>
            <div class="info-row"><span>订单备注</span><span>{{ order.remark || '-' }}</span></div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>金额明细</template>
          <div class="info-list">
            <div class="info-row"><span>商品金额</span><span>¥{{ formatAmount(order.total_amount) }}</span></div>
            <div class="info-row"><span>配送费</span><span>¥{{ formatAmount(order.delivery_fee) }}</span></div>
            <div class="info-row"><span>优惠金额</span><span>-¥{{ formatAmount(order.discount_amount) }}</span></div>
            <div class="info-row total-row"><span>实付金额</span><span>¥{{ formatAmount(order.pay_amount) }}</span></div>
          </div>
        </el-card>
      </div>

      <el-card class="page-card" shadow="never">
        <template #header>商品明细</template>
        <div class="item-list">
          <div v-for="item in order.items" :key="`${item.product_id}-${item.id || item.product_name}`" class="item-row">
            <el-image class="item-image" :src="item.image" fit="cover" />
            <div class="item-main">
              <div class="item-name">{{ item.product_name }}</div>
              <div v-if="resolveSpecText(item)" class="item-spec">{{ resolveSpecText(item) }}</div>
            </div>
            <div class="item-side">
              <div>¥{{ formatAmount(item.price) }}</div>
              <div>x{{ item.quantity }}</div>
              <div v-if="item.subtotal !== undefined">小计：¥{{ formatAmount(item.subtotal) }}</div>
            </div>
          </div>
        </div>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.detail-summary-card {
  margin-bottom: 20px;
}

.summary-title {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
}

.summary-subtitle {
  margin-top: 8px;
  font-size: 14px;
  color: #6b7280;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.info-list {
  display: grid;
  gap: 14px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  color: #374151;
}

.info-row span:first-child {
  color: #6b7280;
  white-space: nowrap;
}

.total-row {
  font-weight: 700;
  color: #111827;
}

.item-list {
  display: grid;
  gap: 16px;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid #eef2f7;
}

.item-row:last-child {
  border-bottom: none;
}

.item-image {
  width: 72px;
  height: 72px;
  border-radius: 12px;
  flex-shrink: 0;
  overflow: hidden;
}

.item-main {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-weight: 600;
  color: #111827;
}

.item-spec {
  margin-top: 6px;
  color: #6b7280;
  font-size: 13px;
}

.item-side {
  min-width: 120px;
  text-align: right;
  color: #374151;
}

@media (max-width: 1200px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
