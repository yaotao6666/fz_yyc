<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const now = new Date()
const currentYear = ref(now.getFullYear())
const currentMonth = ref(now.getMonth() + 1) // 1-12

const monthLabel = computed(() => `${currentYear.value}年${currentMonth.value}月`)

function daysInMonth(y: number, m: number) {
  return new Date(y, m, 0).getDate()
}
function firstWeekday(y: number, m: number) {
  // 0=周日 -> 对齐 6, 周一=0
  const w = new Date(y, m - 1, 1).getDay()
  return (w + 6) % 7
}

const gridCells = computed(() => {
  const total = daysInMonth(currentYear.value, currentMonth.value)
  const offset = firstWeekday(currentYear.value, currentMonth.value)
  const cells: Array<{ day: number | null; type?: 'work' | 'rest' | 'duty'; label?: string }> = []
  for (let i = 0; i < offset; i++) cells.push({ day: null })
  for (let d = 1; d <= total; d++) {
    // 第2期：从 staffScheduleApi.getScheduleList 取真实排班数据
    cells.push({ day: d })
  }
  return cells
})

function prevMonth() {
  if (currentMonth.value === 1) {
    currentMonth.value = 12
    currentYear.value--
  } else {
    currentMonth.value--
  }
}
function nextMonth() {
  if (currentMonth.value === 12) {
    currentMonth.value = 1
    currentYear.value++
  } else {
    currentMonth.value++
  }
}

onMounted(() => {
  // 第2期：拉取当前月排班数据
})
</script>

<template>
  <view class="container">
    <view class="card">
      <!-- 月份切换 -->
      <view class="month-bar">
        <view class="nav-btn" @tap="prevMonth">‹</view>
        <view class="month-label">{{ monthLabel }}</view>
        <view class="nav-btn" @tap="nextMonth">›</view>
      </view>

      <!-- 星期表头 -->
      <view class="week-row">
        <view class="week-cell" v-for="w in ['一','二','三','四','五','六','日']" :key="w">{{ w }}</view>
      </view>

      <!-- 日期网格 -->
      <view class="day-grid">
        <view
          v-for="(cell, idx) in gridCells"
          :key="idx"
          class="day-cell"
          :class="[
            cell.day === new Date().getDate() &&
            currentMonth.value === new Date().getMonth() + 1 &&
            currentYear.value === new Date().getFullYear() ? 'today' : ''
          ]"
        >
          <text v-if="cell.day !== null" class="day-num">{{ cell.day }}</text>
        </view>
      </view>

      <view class="legend">
        <view class="legend-item"><view class="dot today"></view><text>今天</text></view>
      </view>
    </view>

    <view class="card list-tip">
      <text class="tip-title">排班说明</text>
      <text class="tip-desc">第2期将实现：早班/晚班/休息/待命 状态标记、换班申请、请假申请等功能。</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.month-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 8rpx 20rpx;
  .nav-btn {
    width: 64rpx; height: 64rpx; line-height: 56rpx; text-align: center;
    font-size: 40rpx; color: var(--primary-color); background: #f3f7ed;
    border-radius: 50%; font-weight: 300;
  }
  .month-label { font-size: 34rpx; font-weight: 600; color: #333; }
}
.week-row, .day-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8rpx;
}
.week-row { margin-bottom: 16rpx; }
.week-cell {
  text-align: center; font-size: 24rpx; color: #999; padding: 12rpx 0;
}
.day-cell {
  aspect-ratio: 1 / 1;
  border-radius: 12rpx;
  display: flex; align-items: center; justify-content: center;
  background: #fafafa;
  .day-num { font-size: 28rpx; color: #333; }
  &.today {
    background: var(--primary-color);
    .day-num { color: #fff; font-weight: 600; }
  }
}
.legend {
  display: flex; gap: 32rpx;
  margin-top: 28rpx; padding-top: 24rpx;
  border-top: 2rpx solid var(--border-color);
  .legend-item { display: flex; align-items: center; gap: 10rpx; font-size: 24rpx; color: #666; }
  .dot { width: 20rpx; height: 20rpx; border-radius: 6rpx; &.today { background: var(--primary-color); } }
}
.list-tip {
  .tip-title { display: block; font-size: 30rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .tip-desc { display: block; font-size: 26rpx; color: #888; line-height: 1.7; }
}
</style>
