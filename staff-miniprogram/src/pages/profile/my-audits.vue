<template>
  <view class="page">
    <view class="container">
      <view v-if="records.length" class="audit-list">
        <view v-for="r in records" :key="r.id" class="audit-card">
          <view class="audit-header">
            <text class="audit-type">{{ auditTypeText(r.audit_type) }}</text>
            <text class="audit-status" :class="'s' + r.status">{{ statusText(r.status) }}</text>
          </view>
          <view class="audit-meta">
            <text>申请时间：{{ formatTime(r.created_at) }}</text>
          </view>
          <view v-if="r.qualifications && qualsCount(r.qualifications)" class="audit-quals">
            <text class="quals-title">提交资质材料 {{ qualsCount(r.qualifications) }} 项</text>
          </view>
          <view v-if="r.review_remark" class="audit-remark">审核备注：{{ r.review_remark }}</view>
          <view v-if="r.review_at" class="audit-meta">
            <text>审核时间：{{ formatTime(r.review_at) }}</text>
          </view>
        </view>
      </view>
      <view v-else class="empty">
        <text class="empty-text">暂无审核记录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffProfileApi } from '@/api'

const records = ref<any[]>([])

function auditTypeText(t: number) {
  return { 1: '注册申请', 2: '信息变更', 3: '资质提交', 4: '状态变更' }[t] || '未知'
}
function statusText(s: number) {
  return { 0: '待审核', 1: '已通过', 2: '已驳回' }[s] || '未知'
}
function qualsCount(quals: any): number {
  if (Array.isArray(quals)) return quals.length
  if (typeof quals === 'string' && quals.trim()) { try { return JSON.parse(quals).length } catch {} }
  return 0
}
function formatTime(t?: string) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

onLoad(async () => {
  const res: any = await staffProfileApi.getMyAuditList()
  if (res.code === 0) {
    records.value = res.data?.list || []
  }
})
</script>

<style lang="scss" scoped>
.page { min-height: 100vh; background: var(--bg-color); }
.container { padding: 24rpx 32rpx; }
.audit-card { background: #fff; border-radius: 16rpx; padding: 28rpx; margin-bottom: 24rpx; }
.audit-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.audit-type { font-size: 30rpx; font-weight: 600; color: #333; }
.audit-status { font-size: 24rpx; padding: 6rpx 20rpx; border-radius: 24rpx; }
.audit-status.s0 { background: #fff3e0; color: #ed6c00; }
.audit-status.s1 { background: #e8f5e9; color: #2e7d32; }
.audit-status.s2 { background: #fdecea; color: #c62828; }
.audit-meta { font-size: 24rpx; color: #999; margin-top: 8rpx; }
.audit-quals { margin-top: 12rpx; }
.quals-title { font-size: 24rpx; color: #517528; }
.audit-remark { margin-top: 12rpx; font-size: 26rpx; color: #666; background: #f8f8f8; border-radius: 8rpx; padding: 12rpx; }
.empty { text-align: center; padding-top: 120rpx; }
.empty-text { color: #999; font-size: 28rpx; }
</style>