<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { staffWorkorderApi } from '@/api'
import { OrderTypeText } from '@/types'
import { fromNow, formatDate, calcAge } from '@/utils/format'

const tabs = [
  { key: 'all', label: '全部', bizStatus: 0 },
  { key: 'pending', label: '待接单', bizStatus: 1 },
  { key: 'accepted', label: '待出发', bizStatus: 2 },
  { key: 'ongoing', label: '服务中', bizStatus: 3 },
  { key: 'done', label: '已完成', bizStatus: 5 }
]
const activeTab = ref('all')
const categories = [
  { key: 0, label: '全部' },
  { key: 1, label: '实物' },
  { key: 2, label: '服务' }
]
const activeCategory = ref(0)
const loading = ref(false)
const list = ref<any[]>([])

async function refresh() {
  loading.value = true
  try {
    const tab = tabs.find(t => t.key === activeTab.value)
    const bizStatus = tab?.bizStatus || 0
    const category = activeCategory.value || undefined

    if (activeTab.value === 'pending') {
      // 待接单：调用 pending 接口
      const res: any = await staffWorkorderApi.getPendingOrders({ category })
      if (res.code === 0) {
        list.value = res.data?.list || []
      }
    } else {
      // 已接订单：调用 accepted 接口
      const params: any = { category }
      if (bizStatus) params.biz_status = bizStatus
      const res: any = await staffWorkorderApi.getAcceptedOrders(params)
      if (res.code === 0) {
        list.value = res.data?.list || []
      }
    }
  } catch (e) {
    console.error('[Workorder] refresh error', e)
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function changeTab(key: string) {
  activeTab.value = key
  refresh()
}

function changeCategory(key: number) {
  activeCategory.value = key
  refresh()
}

function getOrderTypeText(type: number) {
  return OrderTypeText[type as keyof typeof OrderTypeText] || '未知'
}

// 服务对象性别文案（1男 2女）
function getGenderText(gender?: number) {
  if (gender === 1) return '男'
  if (gender === 2) return '女'
  return ''
}

// 服务对象展示文案：{姓名}{性别}{年龄}岁（有 real_name 才返回）
function getServiceObjectText(item: any) {
  const record = item?.customer?.record
  if (!record?.real_name) return ''
  return `${record.real_name}${getGenderText(record.gender)}${calcAge(record.birth_date)}岁`
}

function getBizStatusText(status: number) {
  const map: Record<number, string> = {
    1: '待接单', 2: '待出发', 3: '服务中', 5: '已完成', 6: '已取消'
  }
  return map[status] || '未知'
}

function getBizStatusClass(status: number) {
  const map: Record<number, string> = {
    1: 'status-pending', 2: 'status-warn', 3: 'status-info', 5: 'status-success', 6: 'status-cancel'
  }
  return map[status] || ''
}

async function handleAccept(id: number) {
  uni.showModal({
    title: '确认接单',
    content: '接单后将出现在您的待办中，请及时前往服务',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const result: any = await staffWorkorderApi.acceptOrder(id)
        if (result.code === 0) {
          uni.showToast({ title: '接单成功', icon: 'success' })
          refresh()
        } else {
          uni.showToast({ title: result.message || '接单失败', icon: 'none' })
        }
      } catch (e: any) {
        const msg = e?.data?.message || '接单失败'
        uni.showToast({ title: msg, icon: 'none' })
      }
    }
  })
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/workorder/detail?id=${id}` })
}

onPullDownRefresh(refresh)
onMounted(refresh)
</script>

<template>
  <view>
    <!-- 分类维度 -->
    <scroll-view scroll-x class="tabs-wrap tabs-category" show-scrollbar="false">
      <view class="tabs">
        <view
          v-for="c in categories"
          :key="c.key"
          class="tab-item"
          :class="{ active: activeCategory === c.key }"
          @tap="changeCategory(c.key)"
        >
          <text>{{ c.label }}</text>
          <view v-if="activeCategory === c.key" class="tab-underline"></view>
        </view>
      </view>
    </scroll-view>

    <!-- 顶部 Tabs -->
    <scroll-view scroll-x class="tabs-wrap" show-scrollbar="false">
      <view class="tabs">
        <view
          v-for="t in tabs"
          :key="t.key"
          class="tab-item"
          :class="{ active: activeTab === t.key }"
          @tap="changeTab(t.key)"
        >
          <text>{{ t.label }}</text>
          <view v-if="activeTab === t.key" class="tab-underline"></view>
        </view>
      </view>
    </scroll-view>

    <view class="container">
      <view v-if="list.length === 0 && !loading" class="empty card">
        <text class="empty-title">暂无工单</text>
        <text class="empty-desc">切换 Tab 查看其他状态，下拉可刷新</text>
      </view>

      <view
        v-for="item in list"
        :key="item.id"
        class="card work-item"
        @tap="goDetail(item.id)"
      >
        <view class="item-header">
          <text class="item-type">{{ getOrderTypeText(item.order_type) }}</text>
          <text :class="['item-status', getBizStatusClass(item.biz_status)]">
            {{ getBizStatusText(item.biz_status) }}
          </text>
        </view>
        <view class="item-no">单号：{{ item.order_no }}</view>
        <view class="item-info" v-if="getServiceObjectText(item)">
          <text class="item-service">服务对象</text>
          <text>{{ getServiceObjectText(item) }}</text>
        </view>
        <view class="item-info" v-if="item.customer?.record?.real_name && item.contact_name">
          <text class="item-service">下单人</text>
          <text>{{ item.contact_name }}</text>
          <text v-if="item.contact_phone" class="item-phone">{{ item.contact_phone }}</text>
        </view>
        <view class="item-info" v-else-if="item.contact_name">
          <text>{{ item.contact_name }}</text>
          <text v-if="item.contact_phone" class="item-phone">{{ item.contact_phone }}</text>
        </view>
        <view class="item-addr" v-if="item.delivery_address">{{ item.delivery_address }}</view>
        <view class="item-extra" v-if="item.delivery_district">{{ item.delivery_district }}</view>
        <view class="item-extra" v-if="item.scheduled_at">预约时间：{{ formatDate(item.scheduled_at) }}</view>
        <view class="item-remark" v-if="item.remark">备注：{{ item.remark }}</view>
        <view class="item-footer">
          <view class="item-amount">¥{{ Number(item.pay_amount || 0).toFixed(2) }}</view>
          <view class="item-actions">
            <text class="item-time">{{ fromNow(item.paid_at || item.created_at) }}</text>
            <button
              v-if="item.biz_status === 1"
              class="btn-accept"
              @tap.stop="handleAccept(item.id)"
            >立即接单</button>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.tabs-wrap {
  background: #fff;
  border-bottom: 2rpx solid var(--border-color);
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 10;
}
.tabs-category {
  position: static;
  border-bottom: none;
  background: #f7f8fa;
  .tab-item {
    padding: 20rpx 28rpx;
    font-size: 28rpx;
  }
  .tab-underline {
    bottom: 8rpx;
  }
}
.tabs { display: inline-flex; padding: 0 16rpx; }
.tab-item {
  position: relative;
  padding: 28rpx 28rpx;
  font-size: 30rpx;
  color: #666;
  &.active { color: var(--primary-color); font-weight: 600; }
}
.tab-underline {
  position: absolute;
  bottom: 12rpx; left: 50%;
  transform: translateX(-50%);
  width: 48rpx; height: 6rpx;
  border-radius: 4rpx;
  background: var(--primary-color);
}
.empty {
  text-align: center;
  padding: 120rpx 32rpx !important;
  margin-top: 24rpx;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}

.work-item {
  padding: 28rpx 24rpx;
  margin-bottom: 20rpx;
}
.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.item-type { font-size: 30rpx; font-weight: 600; color: #333; }
.item-status {
  font-size: 24rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  &.status-pending { background: #fff7e6; color: #fa8c16; }
  &.status-warn { background: #fff2e8; color: #ff9500; }
  &.status-info { background: #e6f4ff; color: #1989fa; }
  &.status-success { background: #e8f8ee; color: #07c160; }
  &.status-cancel { background: #f5f5f5; color: #999; }
}
.item-no { font-size: 24rpx; color: #999; margin-bottom: 8rpx; }
.item-info { display: flex; gap: 16rpx; font-size: 28rpx; color: #333; margin-bottom: 4rpx; }
.item-service {
  flex-shrink: 0;
  font-size: 22rpx;
  color: #fff;
  background: var(--primary-color);
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
  align-self: center;
}
.item-phone { color: #666; }
.item-addr { font-size: 26rpx; color: #666; margin-bottom: 4rpx; }
.item-extra { font-size: 24rpx; color: #999; margin-bottom: 4rpx; }
.item-remark { font-size: 26rpx; color: #999; margin-bottom: 8rpx; }
.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item-amount { font-size: 32rpx; font-weight: 600; color: #ff4d4f; }
.item-actions {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.item-time { font-size: 24rpx; color: #999; }
.btn-accept {
  font-size: 26rpx;
  color: #fff;
  background: var(--primary-color);
  padding: 8rpx 24rpx;
  border-radius: 8rpx;
  line-height: 1.6;
  margin: 0;
  &::after { border: none; }
}
</style>
