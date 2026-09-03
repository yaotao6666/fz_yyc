<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import type { HealthRecord, HealthAssessment, AssessmentFormQuestion, FittingRecommendation } from '@/types'
import { formatDate } from '@/utils/format'

const userId = ref<string>('')
// 外部（如订单详情）指定档案时，评估记录仅显示该档案；为空则回退到当前最近档案
const targetRecordId = ref<number | ''>('')
const loading = ref(false)
const record = ref<HealthRecord | null>(null)
const recordLoaded = ref(false)
// 档案加载是否失败（无权限等情况，区别于未建档返回 null）
const recordError = ref(false)
const assessments = ref<HealthAssessment[]>([])
const total = ref(0)
// 适配建议列表
const fittingList = ref<FittingRecommendation[]>([])
const fittingTotal = ref(0)

// 评估详情弹层
const detailItem = ref<HealthAssessment | null>(null)
// 适配建议详情弹层
const fittingDetail = ref<FittingRecommendation | null>(null)
// 病史分段是否展开全部
const showHistory = ref(false)

// 按出生日期计算年龄
function calcAge(birthDate?: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate.replace(/-/g, '/'))
  if (isNaN(birth.getTime())) return ''
  const now = new Date()
  let age = now.getFullYear() - birth.getFullYear()
  const monthDiff = now.getMonth() - birth.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && now.getDate() < birth.getDate())) age--
  return age > 0 ? `${age}岁` : ''
}

function genderText(gender?: number) {
  if (gender === 1) return '男'
  if (gender === 2) return '女'
  return ''
}

// 适配建议状态文案
function fittingStatusText(status?: number) {
  const map: Record<number, string> = { 0: '草稿', 1: '已确认', 2: '已下单' }
  return status == null ? '' : (map[status] || '未知')
}

// 销售类型文案：2 租赁，其余一口价
function saleTypeText(saleType?: number) {
  return saleType === 2 ? '租赁' : '一口价'
}

// 推荐商品名称拼接（空名称自动忽略）
function productNamesText(item: FittingRecommendation) {
  return (item.recommended_products || []).map(p => p.name).filter(Boolean).join('、')
}

// 后端 JSON 列可能返回字符串或数组，统一规范化为字符串数组
function toArray(value: any): string[] {
  if (Array.isArray(value)) return value.filter((v: any) => typeof v === 'string')
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed.filter((v: any) => typeof v === 'string')
    } catch {
      // 非 JSON 字符串按单元素处理
    }
    return [value]
  }
  return []
}

// 后端 JSON 列可能返回字符串，规范化为题目数组
function toQuestionArray(value: any): AssessmentFormQuestion[] {
  if (Array.isArray(value)) return value as AssessmentFormQuestion[]
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed as AssessmentFormQuestion[]
    } catch {
      // 解析失败视为无题目
    }
  }
  return []
}

// 评估答案可能为 JSON 字符串，规范化为对象
function normalizeAnswers(value: any): Record<string, string> {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value
  if (typeof value === 'string' && value.trim()) {
    try {
      const parsed = JSON.parse(value)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return parsed
    } catch {
      // 解析失败视为空答案
    }
  }
  return {}
}

const chronicTags = computed(() => (record.value ? toArray(record.value.chronic_tags) : []))

// 病史分段（仅保留非空字段）
const historySections = computed(() => {
  const r = record.value
  if (!r) return []
  const fieldList: { label: string; value: any }[] = [
    { label: '既往病史', value: r.past_history },
    { label: '过敏史', value: r.allergy_history },
    { label: '家族病史', value: r.family_history },
    { label: '手术史', value: r.surgery_history },
    { label: '长期用药', value: r.medication_list },
    { label: '吸烟情况', value: r.smoking },
    { label: '饮酒情况', value: r.drinking }
  ]
  const sections: { label: string; items: string[] }[] = []
  fieldList.forEach(f => {
    const items = toArray(f.value)
    if (items.length) sections.push({ label: f.label, items })
  })
  return sections
})

const hasHistory = computed(() => historySections.value.length > 0)

// 展开前只展示前两段，避免页面过长
const displayHistorySections = computed(() =>
  showHistory.value ? historySections.value : historySections.value.slice(0, 2)
)

async function loadData() {
  if (!userId.value) return
  loading.value = true
  try {
    // 档案与评估记录独立加载，其中一个失败不影响另一个展示
    try {
      record.value = await staffHealthApi.getResidentHealthRecord(userId.value)
    } catch (e) {
      recordError.value = true
      console.error('[HealthResident] load record error', e)
    }
    recordLoaded.value = true
    // 评估记录按档案隔离：优先外部指定档案，其次当前最近档案；无档案时不同步全部评估，避免串档
    const rid = targetRecordId.value || record.value?.id || 0
    if (rid) {
      try {
        const assessRes: any = await staffHealthApi.getResidentAssessments(userId.value, { page: 1, page_size: 20, record_id: rid })
        const list = assessRes?.list || []
        assessments.value = list.map((a: any) => ({
          ...a,
          answers: normalizeAnswers(a.answers)
        }))
        total.value = assessRes?.total || 0
      } catch (e) {
        console.error('[HealthResident] load assessments error', e)
      }
    } else {
      assessments.value = []
      total.value = 0
    }
    // 适配建议独立加载，失败不影响档案与评估展示
    try {
      const fitRes: any = await staffHealthApi.getResidentFittingRecommendations(userId.value, { page: 1, page_size: 20 })
      fittingList.value = fitRes?.list || []
      fittingTotal.value = fitRes?.total || 0
    } catch (e) {
      console.error('[HealthResident] load fitting recommendations error', e)
    }
  } finally {
    loading.value = false
  }
}

// 评估详情：按题目顺序映射答案，缺失题目的答案按原 key 补充展示
function buildDetailAnswers(item: HealthAssessment) {
  const answers = item.answers || {}
  const questions = toQuestionArray((item as any).form?.questions)
  const rows: { title: string; answer: string }[] = []
  const visited = new Set<string>()
  questions.forEach((q: AssessmentFormQuestion) => {
    visited.add(q.key)
    if (answers[q.key]) rows.push({ title: q.title, answer: answers[q.key] })
  })
  Object.keys(answers).forEach(key => {
    if (!visited.has(key) && answers[key]) rows.push({ title: key, answer: answers[key] })
  })
  return rows
}

const detailAnswers = computed(() => (detailItem.value ? buildDetailAnswers(detailItem.value) : []))

function openDetail(item: HealthAssessment) {
  detailItem.value = item
}

function closeDetail() {
  detailItem.value = null
}

function goAssess() {
  uni.navigateTo({ url: `/pages/health/assess?userId=${userId.value}` })
}

function openFittingDetail(item: FittingRecommendation) {
  fittingDetail.value = item
}

function closeFittingDetail() {
  fittingDetail.value = null
}

function goFitting() {
  uni.navigateTo({ url: `/pages/health/fitting-form?userId=${userId.value}` })
}

onLoad((options: any) => {
  userId.value = options?.userId || ''
  const pRid = options?.recordId || options?.record_id
  targetRecordId.value = pRid ? Number(pRid) || '' : ''
  if (!userId.value) {
    uni.showToast({ title: '缺少客户ID', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  loadData()
})
</script>

<template>
  <view class="page">
    <view class="container" v-if="!loading">
      <!-- 未建档 / 加载失败 -->
      <view v-if="recordLoaded && !record" class="empty card">
        <template v-if="recordError">
          <text class="empty-title">档案加载失败</text>
          <text class="empty-desc">可能无权限查看该客户档案，请返回确认服务关系</text>
        </template>
        <template v-else>
          <text class="empty-title">该客户暂未建立健康档案</text>
          <text class="empty-desc">可先为客户完成健康评估</text>
        </template>
      </view>

      <template v-else-if="record">
        <!-- 客户档案卡 -->
        <view class="card profile-card">
          <view class="profile-header">
            <view class="avatar">
              <text class="avatar-text">{{ record.real_name?.[0] || '客' }}</text>
            </view>
            <view class="profile-main">
              <view class="profile-name">
                <text class="name">{{ record.real_name || '未填写姓名' }}</text>
                <text v-if="record.gender" class="gender">{{ genderText(record.gender) }}</text>
                <text v-if="calcAge(record.birth_date)" class="age">{{ calcAge(record.birth_date) }}</text>
              </view>
              <text v-if="record.phone" class="profile-phone">{{ record.phone }}</text>
            </view>
            <text v-if="record.assessment_level" class="level-badge">{{ record.assessment_level }}</text>
          </view>

          <view class="profile-body">
            <view class="profile-row" v-if="record.address">
              <text class="label">住址</text>
              <text class="value">{{ record.address }}</text>
            </view>
            <view class="profile-row" v-if="record.blood_type || record.height_cm || record.weight_kg">
              <text class="label">体征</text>
              <text class="value">
                <template v-if="record.blood_type">血型 {{ record.blood_type }}</template>
                <template v-if="record.height_cm"> · {{ record.height_cm }}cm</template>
                <template v-if="record.weight_kg"> · {{ record.weight_kg }}kg</template>
              </text>
            </view>
            <view class="profile-row" v-if="record.emergency_contact || record.emergency_phone">
              <text class="label">紧急联系</text>
              <text class="value">{{ record.emergency_contact || '' }}{{ record.emergency_phone ? ' ' + record.emergency_phone : '' }}</text>
            </view>
          </view>
        </view>

        <!-- 慢病标签 -->
        <view class="card tags-card" v-if="chronicTags.length">
          <view class="card-title">慢病标签</view>
          <view class="chips">
            <text v-for="tag in chronicTags" :key="tag" class="chip">{{ tag }}</text>
          </view>
        </view>

        <!-- 病史分段 -->
        <view class="card section-card">
          <view class="card-title">健康病史</view>
          <view v-if="!hasHistory" class="no-data">暂无病史信息</view>
          <template v-else>
            <view v-for="section in displayHistorySections" :key="section.label" class="history-section">
              <text class="history-label">{{ section.label }}</text>
              <view class="chips">
                <text v-for="item in section.items" :key="item" class="chip">{{ item }}</text>
              </view>
            </view>
            <view v-if="historySections.length > 2" class="toggle-btn" @tap="showHistory = !showHistory">
              {{ showHistory ? '收起' : '展开全部' }}
            </view>
          </template>
        </view>

        <!-- 评估记录 -->
        <view class="card section-card">
          <view class="card-title-row">
            <text class="card-title">评估记录</text>
            <text v-if="total" class="record-total">共 {{ total }} 条</text>
          </view>
          <view v-if="assessments.length === 0" class="no-data">暂无评估记录</view>
          <view
            v-for="item in assessments"
            :key="item.id"
            class="assess-item"
            @tap="openDetail(item)"
          >
            <view class="assess-header">
              <text class="assess-name">{{ item.form_name || '健康评估' }}</text>
              <text class="assess-score">{{ item.total_score }} 分</text>
            </view>
            <view class="assess-footer">
              <text v-if="item.level" class="assess-level">{{ item.level }}</text>
              <text class="assess-time">{{ formatDate(item.created_at, 'YYYY-MM-DD HH:mm') }}</text>
            </view>
          </view>
        </view>

        <!-- 适配建议 -->
        <view class="card section-card">
          <view class="card-title-row">
            <text class="card-title">适配建议</text>
            <text v-if="fittingTotal" class="record-total">共 {{ fittingTotal }} 条</text>
          </view>
          <view v-if="fittingList.length === 0" class="no-data">暂无适配建议</view>
          <view
            v-for="item in fittingList"
            :key="item.id"
            class="assess-item"
            @tap="openFittingDetail(item)"
          >
            <view class="assess-header">
              <text class="assess-name">{{ item.fitting_result || '适配建议' }}</text>
              <text v-if="item.status != null" class="fit-status">{{ fittingStatusText(item.status) }}</text>
            </view>
            <view v-if="productNamesText(item)" class="fit-products">{{ productNamesText(item) }}</view>
            <view class="assess-footer">
              <text class="assess-time">{{ formatDate(item.created_at, 'YYYY-MM-DD HH:mm') }}</text>
            </view>
          </view>
        </view>
      </template>
    </view>

    <view v-else class="loading-wrap">
      <text>加载中...</text>
    </view>

    <!-- 底部操作栏 -->
    <view class="footer-bar" v-if="recordLoaded">
      <button class="btn-outline" @tap="goFitting">生成适配建议</button>
      <button class="btn-assess" @tap="goAssess">为客户评估 ›</button>
    </view>

    <!-- 评估详情弹层 -->
    <view v-if="detailItem" class="modal-mask" @tap="closeDetail">
      <view class="modal-content" @tap.stop>
        <view class="modal-title">{{ detailItem.form_name || '健康评估' }}</view>
        <view class="modal-meta">
          <text class="meta-score">总分 {{ detailItem.total_score }}</text>
          <text v-if="detailItem.level" class="meta-level">{{ detailItem.level }}</text>
          <text class="meta-time">{{ formatDate(detailItem.created_at, 'YYYY-MM-DD HH:mm') }}</text>
        </view>
        <scroll-view scroll-y class="modal-body">
          <view v-if="detailItem.symptom_desc" class="modal-block">
            <text class="modal-block-title">主诉/需求描述</text>
            <text class="modal-block-text">{{ detailItem.symptom_desc }}</text>
          </view>
          <view v-if="detailAnswers.length" class="modal-block">
            <text class="modal-block-title">逐题答案</text>
            <view v-for="row in detailAnswers" :key="row.title" class="answer-row">
              <text class="answer-q">{{ row.title }}</text>
              <text class="answer-a">{{ row.answer }}</text>
            </view>
          </view>
          <view v-if="detailItem.conclusion" class="modal-block">
            <text class="modal-block-title">评估结论</text>
            <text class="modal-block-text">{{ detailItem.conclusion }}</text>
          </view>
        </scroll-view>
        <button class="modal-close" @tap="closeDetail">关闭</button>
      </view>
    </view>

    <!-- 适配建议详情弹层 -->
    <view v-if="fittingDetail" class="modal-mask" @tap="closeFittingDetail">
      <view class="modal-content" @tap.stop>
        <view class="modal-title">适配建议</view>
        <view class="modal-meta">
          <text v-if="fittingDetail.status != null" class="fit-status">{{ fittingStatusText(fittingDetail.status) }}</text>
          <text class="meta-time">{{ formatDate(fittingDetail.created_at, 'YYYY-MM-DD HH:mm') }}</text>
        </view>
        <scroll-view scroll-y class="modal-body">
          <view v-if="fittingDetail.symptom_desc" class="modal-block">
            <text class="modal-block-title">症状/需求描述</text>
            <text class="modal-block-text">{{ fittingDetail.symptom_desc }}</text>
          </view>
          <view v-if="fittingDetail.fitting_result" class="modal-block">
            <text class="modal-block-title">适配结论</text>
            <text class="modal-block-text">{{ fittingDetail.fitting_result }}</text>
          </view>
          <view v-if="fittingDetail.recommended_products.length" class="modal-block">
            <text class="modal-block-title">推荐商品</text>
            <view v-for="p in fittingDetail.recommended_products" :key="p.product_id" class="fit-product-row">
              <view class="fit-product-head">
                <text class="fit-product-name">{{ p.name }}</text>
                <text class="fit-product-type">{{ saleTypeText(p.sale_type) }}</text>
              </view>
              <text v-if="p.reason" class="fit-product-reason">理由：{{ p.reason }}</text>
            </view>
          </view>
        </scroll-view>
        <button class="modal-close" @tap="closeFittingDetail">关闭</button>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page { min-height: 100vh; }
.container {
  padding-bottom: 160rpx;
}
.empty {
  text-align: center;
  padding: 120rpx 32rpx !important;
  .empty-title { display: block; font-size: 32rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .empty-desc { display: block; font-size: 26rpx; color: #999; }
}
.loading-wrap {
  text-align: center;
  padding: 200rpx 0;
  color: #999;
  font-size: 28rpx;
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
  .record-total { font-size: 24rpx; color: #999; }
}

/* 客户档案卡 */
.profile-card {
  padding: 28rpx 24rpx;
}
.profile-header {
  display: flex;
  align-items: center;
  margin-bottom: 24rpx;
}
.avatar {
  width: 104rpx; height: 104rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #517528 0%, #7a9e4f 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  .avatar-text { font-size: 44rpx; font-weight: 600; color: #fff; }
}
.profile-main {
  flex: 1;
  margin-left: 20rpx;
  min-width: 0;
}
.profile-name {
  display: flex;
  align-items: center;
  .name { font-size: 34rpx; font-weight: 600; color: #333; }
  .gender {
    margin-left: 16rpx;
    font-size: 22rpx;
    color: var(--info-color);
    background: #e6f4ff;
    padding: 2rpx 12rpx;
    border-radius: 8rpx;
  }
  .age {
    margin-left: 8rpx;
    font-size: 22rpx;
    color: #666;
  }
}
.profile-phone {
  display: block;
  font-size: 26rpx;
  color: #999;
  margin-top: 8rpx;
}
.level-badge {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #fff;
  background: var(--warning-color);
  padding: 8rpx 20rpx;
  border-radius: 32rpx;
}
.profile-row {
  display: flex;
  padding: 10rpx 0;
  .label { width: 140rpx; font-size: 26rpx; color: #999; flex-shrink: 0; }
  .value { flex: 1; font-size: 26rpx; color: #333; text-align: right; line-height: 1.5; }
}

/* 标签 chips */
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.chip {
  font-size: 24rpx;
  color: var(--primary-color);
  background: rgba(81, 117, 40, 0.1);
  padding: 8rpx 24rpx;
  border-radius: 32rpx;
}

/* 病史分段 */
.history-section {
  padding: 16rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
  .history-label { display: block; font-size: 26rpx; color: #666; margin-bottom: 12rpx; }
}
.toggle-btn {
  margin-top: 16rpx;
  text-align: center;
  font-size: 26rpx;
  color: var(--primary-color);
  padding: 12rpx 0;
}
.no-data {
  font-size: 26rpx;
  color: #999;
  padding: 16rpx 0;
}

/* 评估记录 */
.assess-item {
  padding: 20rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
}
.assess-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
  .assess-name { font-size: 28rpx; color: #333; font-weight: 500; }
  .assess-score { font-size: 28rpx; color: var(--primary-color); font-weight: 600; }
}
.assess-footer {
  display: flex;
  align-items: center;
  gap: 16rpx;
  .assess-level {
    font-size: 22rpx;
    color: var(--warning-color);
    background: #fff7e6;
    padding: 2rpx 16rpx;
    border-radius: 8rpx;
  }
  .assess-time { font-size: 24rpx; color: #999; }
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
.btn-outline {
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
.btn-assess {
  flex: 1.4;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
}

/* 适配建议 */
.fit-status {
  font-size: 22rpx;
  color: var(--warning-color);
  background: #fff7e6;
  padding: 2rpx 16rpx;
  border-radius: 8rpx;
  flex-shrink: 0;
}
.fit-products {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-bottom: 10rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.fit-product-row {
  padding: 12rpx 0;
  .fit-product-head { display: flex; align-items: center; gap: 12rpx; margin-bottom: 6rpx; }
  .fit-product-name { font-size: 26rpx; color: #333; font-weight: 500; }
  .fit-product-type { font-size: 22rpx; color: var(--info-color); background: #e6f4ff; padding: 2rpx 12rpx; border-radius: 8rpx; }
  .fit-product-reason { display: block; font-size: 24rpx; color: #999; line-height: 1.5; }
}

/* 评估详情弹层 */
.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal-content {
  width: 620rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx 32rpx 32rpx;
}
.modal-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
  text-align: center;
  margin-bottom: 12rpx;
}
.modal-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
  .meta-score { font-size: 26rpx; color: var(--primary-color); font-weight: 600; }
  .meta-level { font-size: 22rpx; color: var(--warning-color); background: #fff7e6; padding: 2rpx 16rpx; border-radius: 8rpx; }
  .meta-time { font-size: 22rpx; color: #999; }
}
.modal-body {
  height: 520rpx;
  border-top: 2rpx solid var(--border-color);
  border-bottom: 2rpx solid var(--border-color);
  box-sizing: border-box;
}
.modal-block {
  padding: 20rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
  .modal-block-title { display: block; font-size: 26rpx; font-weight: 600; color: #333; margin-bottom: 12rpx; }
  .modal-block-text { display: block; font-size: 26rpx; color: #555; line-height: 1.6; }
}
.answer-row {
  display: flex;
  justify-content: space-between;
  gap: 24rpx;
  padding: 8rpx 0;
  .answer-q { flex: 1; font-size: 26rpx; color: #666; }
  .answer-a { flex-shrink: 0; max-width: 50%; font-size: 26rpx; color: #333; text-align: right; }
}
.modal-close {
  margin-top: 24rpx;
  width: 100%;
  font-size: 30rpx;
  color: #fff;
  background: var(--primary-color);
  border-radius: 12rpx;
  line-height: 2.2;
  &::after { border: none; }
}
</style>
