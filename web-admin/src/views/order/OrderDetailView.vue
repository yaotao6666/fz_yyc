<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  completeOrder,
  getOrderDetail,
  quickCompleteOrder,
  returnRentalOrder
} from '@/api/sp'
import type { SpOrder, SpOrderItem } from '@/types/sp'
import { SpOrderStatusText, OrderTypeText, BizStatusText } from '@/types/sp'
import { formatAmount, formatDateTime, getDepositStatusText, getRentalUnitText } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const order = ref<SpOrder | null>(null)
const returning = ref(false)
const returnDialogVisible = ref(false)
const returnForm = ref({ deduct_amount: 0, remark: '' })

const completing = ref(false)
const verifyDialogVisible = ref(false)
const verifyCode = ref('')

const orderId = computed(() => Number(route.params.id || 0))

const statusSummary = computed(() => {
  if (!order.value) {
    return '订单详情'
  }
  return SpOrderStatusText[order.value.status] || '未知状态'
})

const isRentalOrder = computed(() => Number(order.value?.total_deposit || 0) > 0)

const isServiceOrder = computed(() => {
  const t = Number(order.value?.order_type || 0)
  return t === 3 || t === 4 || t === 6
})

const canReturnRental = computed(() => {
  const o = order.value
  if (!o) return false
  return o.status === 2 && Number(o.total_deposit || 0) > 0 && Number(o.deposit_status || 0) === 1
})

const canComplete = computed(() => {
  const o = order.value
  if (!o) return false
  // 已支付且未完成时可核销
  return o.status === 2
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
    order.value = await getOrderDetail(orderId.value)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}

function openReturnDialog() {
  returnForm.value = { deduct_amount: 0, remark: '' }
  returnDialogVisible.value = true
}

async function confirmReturn() {
  if (!order.value) return
  returning.value = true
  try {
    order.value = await returnRentalOrder(orderId.value, {
      deduct_amount: Number(returnForm.value.deduct_amount || 0),
      remark: returnForm.value.remark.trim()
    })
    ElMessage.success('归还成功，押金已退还')
    returnDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '归还失败')
  } finally {
    returning.value = false
  }
}

function openVerifyDialog() {
  verifyCode.value = ''
  verifyDialogVisible.value = true
}

async function handleQuickComplete() {
  if (!order.value) return
  try {
    await ElMessageBox.confirm('确认快速核销该订单？核销后订单将标记为已完成。', '快速核销', {
      type: 'warning',
      confirmButtonText: '确认核销',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  completing.value = true
  try {
    order.value = await quickCompleteOrder(orderId.value)
    ElMessage.success('订单已核销完成')
  } catch (error: any) {
    ElMessage.error(error?.message || '核销失败')
  } finally {
    completing.value = false
  }
}

async function confirmVerifyComplete() {
  if (!verifyCode.value.trim()) {
    ElMessage.warning('请输入核销码')
    return
  }

  completing.value = true
  try {
    order.value = await completeOrder(orderId.value, verifyCode.value.trim())
    ElMessage.success('订单已核销完成')
    verifyDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '核销失败')
  } finally {
    completing.value = false
  }
}

onMounted(loadOrderDetail)
</script>

<template>
  <div class="page-shell" v-loading="loading">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">订单详情</h1>
        <p class="page-subtitle">查看订单状态、商品明细、工单进度与核销信息。</p>
      </div>
      <el-button @click="goBack">返回列表</el-button>
    </div>

    <template v-if="order">
      <el-card class="page-card detail-summary-card" shadow="never">
        <div class="summary-title">
          {{ statusSummary }}
          <el-tag v-if="order.order_type" size="small" style="margin-left: 12px;">{{ OrderTypeText[order.order_type] || '-' }}</el-tag>
          <el-tag v-if="isRentalOrder" type="warning" size="small" style="margin-left: 8px;">租赁订单</el-tag>
        </div>
        <div class="summary-subtitle">
          订单号：{{ order.order_no }}，下单时间：{{ formatDateTime(order.created_at) }}
        </div>
        <div v-if="canComplete || canReturnRental" class="summary-actions">
          <el-button v-if="canComplete" type="primary" :loading="completing" @click="handleQuickComplete">快速核销</el-button>
          <el-button v-if="canComplete" plain @click="openVerifyDialog">核销码核销</el-button>
          <el-button v-if="canReturnRental" type="primary" @click="openReturnDialog">归还退押金</el-button>
        </div>
      </el-card>

      <div class="detail-grid">
        <el-card class="page-card" shadow="never">
          <template #header>订单信息</template>
          <div class="info-list">
            <div class="info-row"><span>订单编号</span><span>{{ order.order_no }}</span></div>
            <div class="info-row"><span>订单类型</span><span>{{ order.order_type ? (OrderTypeText[order.order_type] || '-') : '-' }}</span></div>
            <div class="info-row"><span>订单状态</span><span>{{ SpOrderStatusText[order.status] || '未知状态' }}</span></div>
            <div class="info-row"><span>下单时间</span><span>{{ formatDateTime(order.created_at) }}</span></div>
            <div class="info-row"><span>支付时间</span><span>{{ formatOptionalDateTime(order.paid_at) }}</span></div>
            <div class="info-row"><span>支付单号</span><span>{{ order.transaction_id || '-' }}</span></div>
            <div class="info-row"><span>完成时间</span><span>{{ formatOptionalDateTime(order.completed_at) }}</span></div>
            <div class="info-row"><span>完成人</span><span>{{ order.completed_by_name || '-' }}</span></div>
          </div>
        </el-card>

        <el-card v-if="isServiceOrder" class="page-card" shadow="never">
          <template #header>工单信息</template>
          <div class="info-list">
            <div class="info-row"><span>工单状态</span><span>{{ order.biz_status ? (BizStatusText[order.biz_status] || '-') : '-' }}</span></div>
            <div class="info-row"><span>服务人员ID</span><span>{{ order.assigned_staff_id || '未指派' }}</span></div>
            <div class="info-row"><span>预约时间</span><span>{{ formatOptionalDateTime(order.scheduled_at) }}</span></div>
            <div class="info-row"><span>开始服务</span><span>{{ formatOptionalDateTime(order.actual_started_at) }}</span></div>
            <div class="info-row"><span>结束服务</span><span>{{ formatOptionalDateTime(order.actual_ended_at) }}</span></div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>用户与配送</template>
          <div class="info-list">
            <div class="info-row"><span>用户昵称</span><span>{{ order.user?.nickname || '匿名用户' }}</span></div>
            <div class="info-row"><span>用户手机号</span><span>{{ order.user?.phone || '-' }}</span></div>
            <div class="info-row"><span>收货地址</span><span>{{ resolveDeliveryAddress() }}</span></div>
            <div class="info-row"><span>联系人</span><span>{{ resolveContactInfo() }}</span></div>
            <div class="info-row"><span>订单备注</span><span>{{ order.remark || '-' }}</span></div>
          </div>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>金额明细</template>
          <div class="info-list">
            <div class="info-row"><span>商品金额</span><span>¥{{ formatAmount(order.total_amount) }}</span></div>
            <div v-if="isRentalOrder" class="info-row"><span>押金</span><span>¥{{ formatAmount(order.total_deposit) }}</span></div>
            <div class="info-row"><span>配送费</span><span>¥{{ formatAmount(order.delivery_fee) }}</span></div>
            <div class="info-row"><span>优惠金额</span><span>-¥{{ formatAmount(order.discount_amount) }}</span></div>
            <div class="info-row total-row"><span>实付金额</span><span>¥{{ formatAmount(order.pay_amount) }}</span></div>
            <template v-if="isRentalOrder">
              <div class="info-row"><span>押金状态</span><span>{{ getDepositStatusText(order.deposit_status) }}</span></div>
              <div v-if="Number(order.deposit_deduct_amount || 0) > 0" class="info-row"><span>押金扣除</span><span>¥{{ formatAmount(order.deposit_deduct_amount) }}</span></div>
              <div v-if="Number(order.deposit_refund_amount || 0) > 0" class="info-row"><span>押金退还</span><span>¥{{ formatAmount(order.deposit_refund_amount) }}</span></div>
              <div v-if="order.deposit_refunded_at" class="info-row"><span>退还时间</span><span>{{ formatDateTime(order.deposit_refunded_at) }}</span></div>
              <div v-if="order.rental_returned_at" class="info-row"><span>归还时间</span><span>{{ formatDateTime(order.rental_returned_at) }}</span></div>
              <div v-if="order.rental_return_remark" class="info-row"><span>归还备注</span><span>{{ order.rental_return_remark }}</span></div>
            </template>
          </div>
        </el-card>
      </div>

      <el-card class="page-card" shadow="never">
        <template #header>商品明细</template>
        <div class="item-list">
          <div v-for="item in order.items" :key="`${item.product_id}-${item.id || item.product_name}`" class="item-row">
            <el-image class="item-image" :src="item.image" fit="cover" />
            <div class="item-main">
              <div class="item-name">
                {{ item.product_name }}
                <el-tag v-if="Number(item.sale_type) === 2" type="warning" size="small" style="margin-left: 8px;">租赁</el-tag>
              </div>
              <div v-if="resolveSpecText(item)" class="item-spec">{{ resolveSpecText(item) }}</div>
              <div v-if="Number(item.sale_type) === 2" class="item-spec">
                租赁时长：{{ item.rental_duration }}{{ getRentalUnitText(item.rental_unit) }}
                <span v-if="Number(item.deposit || 0) > 0">，押金：¥{{ formatAmount(item.deposit) }}</span>
              </div>
            </div>
            <div class="item-side">
              <template v-if="Number(item.sale_type) === 2">
                <div>¥{{ formatAmount(item.unit_rental_price) }}/{{ getRentalUnitText(item.rental_unit) }}</div>
                <div>x{{ item.quantity }}</div>
                <div>租金：¥{{ formatAmount(item.rental_subtotal) }}</div>
              </template>
              <template v-else>
                <div>¥{{ formatAmount(item.price) }}</div>
                <div>x{{ item.quantity }}</div>
                <div v-if="item.subtotal !== undefined">小计：¥{{ formatAmount(item.subtotal) }}</div>
              </template>
            </div>
          </div>
        </div>
      </el-card>
    </template>

    <el-dialog v-model="returnDialogVisible" title="归还租赁商品" width="480px">
      <el-form label-width="100px">
        <el-form-item label="押金总额">
          <span>¥{{ formatAmount(order?.total_deposit || 0) }}</span>
        </el-form-item>
        <el-form-item label="扣除金额">
          <el-input-number v-model="returnForm.deduct_amount" :min="0" :max="Number(order?.total_deposit || 0)" :precision="2" :step="1" style="width: 100%;" placeholder="损坏赔偿金额，0=全额退还" />
        </el-form-item>
        <el-form-item label="验机备注">
          <el-input v-model="returnForm.remark" type="textarea" :rows="3" maxlength="200" show-word-limit placeholder="验机情况说明（可选）" />
        </el-form-item>
        <div class="return-summary">
          退还押金：¥{{ formatAmount(Math.max(0, Number(order?.total_deposit || 0) - Number(returnForm.deduct_amount || 0))) }}
        </div>
      </el-form>
      <template #footer>
        <el-button @click="returnDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="returning" @click="confirmReturn">确认归还</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="verifyDialogVisible" title="核销码核销" width="420px">
      <el-form label-width="80px">
        <el-form-item label="核销码" required>
          <el-input v-model="verifyCode" placeholder="请输入用户提供的核销码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="verifyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="completing" @click="confirmVerifyComplete">确认核销</el-button>
      </template>
    </el-dialog>
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

.summary-actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.return-summary {
  margin-top: 12px;
  padding: 12px 16px;
  background: #f0f9ff;
  border-radius: 8px;
  color: #0369a1;
  font-weight: 600;
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
