<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { staffWorkorderApi } from '@/api'
import { OrderTypeText, WorkorderBizStatus } from '@/types'
import { formatDate, fromNow, calcAge } from '@/utils/format'

const loading = ref(false)
const stats = ref({
  todayAssigned: 0,
  pending: 0,
  inProgress: 0,
  todayDone: 0
})
const todoList = ref<any[]>([])

async function refresh() {
  loading.value = true
  try {
    const res: any = await staffWorkorderApi.getTodoList()
    if (res.code === 0) {
      todoList.value = res.data?.list || []
      const s = res.data?.stats || {}
      stats.value.todayAssigned = s.today_assigned || 0
      stats.value.todayDone = s.today_completed || 0
      stats.value.pending = todoList.value.filter((o: any) => o.biz_status === 2).length
      stats.value.inProgress = todoList.value.filter((o: any) => o.biz_status === 3).length
    }
  } catch (e) {
    console.error('[Todo] refresh error', e)
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function getOrderTypeText(type: number) {
  return OrderTypeText[type as keyof typeof OrderTypeText] || '未知'
}

// 服务对象展示文案：{姓名}{性别}{年龄}岁（有 real_name 才返回）
function getServiceObjectText(item: any) {
  const record = item?.customer?.record
  if (!record?.real_name) return ''
  const gender = record.gender === 1 ? '男' : record.gender === 2 ? '女' : ''
  return `${record.real_name}${gender}${calcAge(record.birth_date)}岁`
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

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/workorder/detail?id=${id}` })
}

onPullDownRefresh(refresh)
onMounted(refresh)
</script>

<template>
  <view class="container">
    <!-- 数据概览卡 -->
    <view class="stats-grid">
      <view class="stat-card">
        <view class="stat-num">{{ stats.todayAssigned }}</view>
        <view class="stat-label">今日派单</view>
      </view>
      <view class="stat-card stat-warn">
        <view class="stat-num">{{ stats.pending }}</view>
        <view class="stat-label">待出发</view>
      </view>
      <view class="stat-card stat-info">
        <view class="stat-num">{{ stats.inProgress }}</view>
        <view class="stat-label">服务中</view>
      </view>
      <view class="stat-card stat-success">
        <view class="stat-num">{{ stats.todayDone }}</view>
        <view class="stat-label">今日完成</view>
      </view>
    </view>

    <view class="section-header">
      <text class="title">我的待办</text>
      <text class="more" @tap="() => uni.switchTab({ url: '/pages/workorder/index' })">全部工单 ›</text>
    </view>

    <view v-if="todoList.length === 0 && !loading" class="empty card">
      <text class="empty-title">暂无待办</text>
      <text class="empty-desc">您今天暂无待处理的工单，下拉可刷新</text>
    </view>

    <view
      v-for="item in todoList"
      :key="item.id"
      class="card todo-item"
      @tap="goDetail(item.id)"
    >
      <view class="todo-header">
        <text class="todo-type">{{ getOrderTypeText(item.order_type) }}</text>
        <text :class="['todo-status', getBizStatusClass(item.biz_status)]">
          {{ getBizStatusText(item.biz_status) }}
        </text>
      </view>
      <view class="todo-no">单号：{{ item.order_no }}</view>
      <view class="todo-info" v-if="getServiceObjectText(item)">
        <text class="todo-service">服务对象</text>
        <text>{{ getServiceObjectText(item) }}</text>
      </view>
      <view class="todo-info" v-if="item.customer?.record?.real_name && item.contact_name">
        <text class="todo-service">下单人</text>
        <text>{{ item.contact_name }}</text>
        <text v-if="item.contact_phone" class="todo-phone">{{ item.contact_phone }}</text>
      </view>
      <view class="todo-info" v-else-if="item.contact_name">
        <text>{{ item.contact_name }}</text>
        <text v-if="item.contact_phone" class="todo-phone">{{ item.contact_phone }}</text>
      </view>
      <view class="todo-addr" v-if="item.delivery_address">{{ item.delivery_address }}</view>
      <view class="todo-addr" v-if="item.scheduled_at">预约时间：{{ formatDate(item.scheduled_at) }}</view>
      <view class="todo-footer">
        <text class="todo-amount">¥{{ Number(item.pay_amount || 0).toFixed(2) }}</text>
        <text class="todo-time">{{ fromNow(item.actual_started_at || item.paid_at || item.created_at) }}</text>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20rpx;
  margin-bottom: 32rpx;
}
.stat-card {
  background: linear-gradient(135deg, #517528 0%, #7a9e4f 100%);
  color: #fff;
  border-radius: 20rpx;
  padding: 32rpx 28rpx;
  box-shadow: 0 6rpx 20rpx rgba(81, 117, 40, 0.2);
  &.stat-warn { background: linear-gradient(135deg, #ff9500 0%, #ffb347 100%); box-shadow: 0 6rpx 20rpx rgba(255,149,0,0.2); }
  &.stat-info { background: linear-gradient(135deg, #1989fa 0%, #54a6ff 100%); box-shadow: 0 6rpx 20rpx rgba(25,137,250,0.2); }
  &.stat-success { background: linear-gradient(135deg, #07c160 0%, #4fd18a 100%); box-shadow: 0 6rpx 20rpx rgba(7,193,96,0.2); }
}
.stat-num { font-size: 52rpx; font-weight: 700; line-height: 1.1; }
.stat-label { font-size: 24rpx; opacity: 0.85; margin-top: 8rpx; }

.section-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8rpx 4rpx 20rpx;
  .title { font-size: 32rpx; font-weight: 600; color: #333; }
  .more { font-size: 26rpx; color: var(--primary-color); }
}

.empty {
  text-align: center;
  padding: 80rpx 32rpx !important;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}

.todo-item {
  padding: 28rpx 24rpx;
  margin-bottom: 20rpx;
}
.todo-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.todo-type {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}
.todo-status {
  font-size: 24rpx;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  &.status-pending { background: #fff7e6; color: #fa8c16; }
  &.status-warn { background: #fff2e8; color: #ff9500; }
  &.status-info { background: #e6f4ff; color: #1989fa; }
  &.status-success { background: #e8f8ee; color: #07c160; }
  &.status-cancel { background: #f5f5f5; color: #999; }
}
.todo-no {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 8rpx;
}
.todo-info {
  display: flex;
  gap: 16rpx;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 4rpx;
  .todo-service {
    flex-shrink: 0;
    font-size: 22rpx;
    color: #fff;
    background: var(--primary-color);
    padding: 2rpx 12rpx;
    border-radius: 6rpx;
    align-self: center;
  }
  .todo-phone { color: #666; }
}
.todo-addr {
  font-size: 26rpx;
  color: #666;
  margin-bottom: 12rpx;
}
.todo-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  .todo-amount { font-size: 32rpx; font-weight: 600; color: #ff4d4f; }
  .todo-time { font-size: 24rpx; color: #999; }
}
</style>
