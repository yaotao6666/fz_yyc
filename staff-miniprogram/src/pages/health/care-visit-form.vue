<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { staffHealthApi } from '@/api'
import { formatDate } from '@/utils/format'

// 来源标识：二选一（planId 或 orderId）
const planId = ref<number | null>(null)
const orderId = ref<number | null>(null)
// 是否有计划（决定护理项是勾选式还是自由输入式）
const hasPlan = ref(false)
const loading = ref(false)
const submitting = ref(false)

// 到访时间（默认当前时间），拆分为日期与时间两段便于 uni picker
const now = new Date()
const visitDate = ref(formatDate(now, 'YYYY-MM-DD'))
const visitTime = ref(formatDate(now, 'HH:mm'))

// 护理项表单状态：有计划时来自计划 items，无计划时自由添加
interface ItemForm {
  name: string
  desc?: string
  done: boolean
  remark: string
}
const itemForms = ref<ItemForm[]>([])
// 无计划时的护理项名称输入
const customItemName = ref('')

// 生命体征输入
const vitals = reactive({
  blood_pressure: '',
  blood_glucose: '',
  heart_rate: '',
  oxygen: '',
  weight: ''
})

// 照片 URL 文本（逗号/换行分隔）
const photosText = ref('')
const remark = ref('')
const followUpAdvice = ref('')

const pageTitle = computed(() => (hasPlan.value ? '录入照护记录' : '录入照护记录（无计划）'))
const visitAt = computed(() => `${visitDate.value} ${visitTime.value}`)

// 已勾选的护理项数量
const doneCount = computed(() => itemForms.value.filter(it => it.done).length)

function onDateChange(e: any) {
  visitDate.value = e?.detail?.value || visitDate.value
}
function onTimeChange(e: any) {
  visitTime.value = e?.detail?.value || visitTime.value
}

function addCustomItem() {
  const name = customItemName.value.trim()
  if (!name) {
    uni.showToast({ title: '请输入护理项名称', icon: 'none' })
    return
  }
  itemForms.value.push({ name, done: false, remark: '' })
  customItemName.value = ''
}

function removeItem(idx: number) {
  itemForms.value.splice(idx, 1)
}

// 照片文本转数组（逗号/换行分隔，过滤空值）
function parsePhotos(text: string): string[] {
  return text
    .split(/[,，\n]/)
    .map(s => s.trim())
    .filter(Boolean)
}

async function loadPlanItems() {
  if (!planId.value) return
  loading.value = true
  try {
    const data: any = await staffHealthApi.getCarePlanDetail(planId.value)
    const items = (data?.items || []) as { name: string; desc?: string }[]
    // 将计划护理项初始化为未勾选，逐项可填写备注
    itemForms.value = items.map(it => ({ name: it.name, desc: it.desc, done: false, remark: '' }))
  } catch (e) {
    console.error('[CareVisitForm] loadPlanItems error', e)
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!planId.value && !orderId.value) {
    uni.showToast({ title: '缺少计划或订单信息', icon: 'none' })
    return
  }
  if (!visitAt.value) {
    uni.showToast({ title: '请选择到访时间', icon: 'none' })
    return
  }
  // 无计划场景至少需要填写一个护理项
  if (!hasPlan.value && itemForms.value.filter(it => it.name.trim()).length === 0) {
    uni.showToast({ title: '请至少填写一个护理项', icon: 'none' })
    return
  }

  // 仅保留非空生命体征
  const vitalsPayload: Record<string, string> = {}
  Object.keys(vitals).forEach(key => {
    const value = (vitals as any)[key].trim()
    if (value) vitalsPayload[key] = value
  })

  submitting.value = true
  try {
    await staffHealthApi.createCareVisit({
      plan_id: planId.value,
      order_id: orderId.value,
      visit_at: visitAt.value,
      nursing_items: itemForms.value
        .filter(it => it.name.trim() !== '')
        .map(it => ({
          name: it.name.trim(),
          done: it.done,
          remark: it.remark.trim() || undefined
        })),
      vitals: Object.keys(vitalsPayload).length ? vitalsPayload : undefined,
      photos: parsePhotos(photosText.value),
      remark: remark.value.trim() || undefined,
      follow_up_advice: followUpAdvice.value.trim() || undefined
    })
    uni.showToast({ title: '照护记录已保存', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (e) {
    console.error('[CareVisitForm] handleSubmit error', e)
  } finally {
    submitting.value = false
  }
}

onLoad((options: any) => {
  if (options?.planId) {
    planId.value = Number(options.planId) || null
    hasPlan.value = !!planId.value
    loadPlanItems()
  } else if (options?.orderId) {
    orderId.value = Number(options.orderId) || null
    hasPlan.value = false
  } else {
    uni.showToast({ title: '缺少计划或订单信息', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
  }
})
</script>

<template>
  <view class="container">
    <!-- 到访时间 -->
    <view class="card section">
      <view class="section-title">到访时间</view>
      <view class="time-pickers">
        <picker mode="date" :value="visitDate" start="2020-01-01" end="2035-12-31" @change="onDateChange">
          <view class="time-picker">{{ visitDate }}</view>
        </picker>
        <picker mode="time" :value="visitTime" @change="onTimeChange">
          <view class="time-picker">{{ visitTime }}</view>
        </picker>
      </view>
    </view>

    <!-- 护理项 -->
    <view class="card section">
      <view class="section-title-row">
        <text class="section-title">护理项</text>
        <text v-if="hasPlan && itemForms.length" class="done-count">已完成 {{ doneCount }}/{{ itemForms.length }}</text>
      </view>

      <!-- 有计划：逐项勾选完成情况并可填备注 -->
      <view v-if="hasPlan">
        <view v-if="loading" class="no-data">护理项加载中...</view>
        <view v-else-if="itemForms.length === 0" class="no-data">该计划暂无护理项</view>
        <view v-else>
          <view v-for="(item, idx) in itemForms" :key="idx" class="nursing-item">
            <view class="nursing-head" @tap="item.done = !item.done">
              <view class="nursing-check" :class="{ checked: item.done }">
                <text v-if="item.done" class="check-mark">✓</text>
              </view>
              <view class="nursing-info">
                <text class="nursing-name">{{ item.name }}</text>
                <text v-if="item.desc" class="nursing-desc">{{ item.desc }}</text>
              </view>
            </view>
            <input
              v-model="item.remark"
              class="remark-input"
              placeholder="完成情况备注（选填）"
              maxlength="200"
            />
          </view>
        </view>
      </view>

      <!-- 无计划：自由添加护理项 -->
      <view v-else>
        <view class="add-row">
          <input
            v-model="customItemName"
            class="add-input"
            placeholder="输入护理项名称"
            maxlength="50"
            confirm-type="done"
            @confirm="addCustomItem"
          />
          <view class="add-btn" @tap="addCustomItem">添加</view>
        </view>
        <view v-if="itemForms.length === 0" class="no-data">请添加本次提供的护理服务项</view>
        <view v-else>
          <view v-for="(item, idx) in itemForms" :key="idx" class="nursing-item">
            <view class="nursing-head" @tap="item.done = !item.done">
              <view class="nursing-check" :class="{ checked: item.done }">
                <text v-if="item.done" class="check-mark">✓</text>
              </view>
              <text class="nursing-name">{{ item.name }}</text>
              <text class="remove-btn" @tap.stop="removeItem(idx)">删除</text>
            </view>
            <input
              v-model="item.remark"
              class="remark-input"
              placeholder="完成情况备注（选填）"
              maxlength="200"
            />
          </view>
        </view>
      </view>
    </view>

    <!-- 生命体征 -->
    <view class="card section">
      <view class="section-title">生命体征（选填）</view>
      <view class="vitals-grid">
        <view class="vital-field">
          <text class="vital-label">血压(mmHg)</text>
          <input v-model="vitals.blood_pressure" class="vital-input" placeholder="如 120/80" maxlength="20" />
        </view>
        <view class="vital-field">
          <text class="vital-label">血糖(mmol/L)</text>
          <input v-model="vitals.blood_glucose" class="vital-input" placeholder="如 6.1" maxlength="20" />
        </view>
        <view class="vital-field">
          <text class="vital-label">心率(次/分)</text>
          <input v-model="vitals.heart_rate" class="vital-input" placeholder="如 72" maxlength="20" />
        </view>
        <view class="vital-field">
          <text class="vital-label">血氧(%)</text>
          <input v-model="vitals.oxygen" class="vital-input" placeholder="如 98" maxlength="20" />
        </view>
        <view class="vital-field">
          <text class="vital-label">体重(kg)</text>
          <input v-model="vitals.weight" class="vital-input" placeholder="如 60" maxlength="20" />
        </view>
      </view>
    </view>

    <!-- 照片 -->
    <view class="card section">
      <view class="section-title">照片（选填）</view>
      <textarea
        v-model="photosText"
        class="photo-input"
        placeholder="请输入照片URL，多个地址用逗号或换行分隔"
        maxlength="1000"
      />
      <view class="photo-tip">暂未接入图片上传，可直接粘贴图片链接</view>
    </view>

    <!-- 备注与随访建议 -->
    <view class="card section">
      <view class="section-title">备注（选填）</view>
      <textarea
        v-model="remark"
        class="desc-input"
        placeholder="本次照护过程中的其他情况说明"
        maxlength="500"
      />
    </view>
    <view class="card section">
      <view class="section-title">下次随访建议（选填）</view>
      <textarea
        v-model="followUpAdvice"
        class="desc-input"
        placeholder="如：建议三天后复测血压，注意饮食清淡"
        maxlength="500"
      />
    </view>

    <!-- 底部操作栏 -->
    <view class="footer-bar">
      <button class="btn-submit" :disabled="submitting" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '保存照护记录' }}
      </button>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.container {
  padding-bottom: 140rpx;
}
.section {
  margin-bottom: 20rpx;
  padding: 28rpx 24rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 20rpx;
}
.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
  .section-title { margin-bottom: 0; }
  .done-count { font-size: 24rpx; color: var(--primary-color); }
}
.no-data { font-size: 26rpx; color: #999; padding: 12rpx 0; }

/* 到访时间 */
.time-pickers {
  display: flex;
  gap: 20rpx;
}
.time-picker {
  flex: 1;
  height: 76rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #333;
}

/* 护理项 */
.nursing-item {
  padding: 16rpx 0;
  border-bottom: 2rpx solid var(--border-color);
  &:last-child { border-bottom: none; }
}
.nursing-head {
  display: flex;
  align-items: center;
}
.nursing-check {
  width: 36rpx; height: 36rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  margin-right: 16rpx;
  flex-shrink: 0;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  &.checked { border-color: var(--primary-color); background: var(--primary-color); }
  .check-mark { font-size: 24rpx; color: #fff; }
}
.nursing-info {
  flex: 1;
  min-width: 0;
  .nursing-name { font-size: 28rpx; color: #333; }
  .nursing-desc { display: block; font-size: 24rpx; color: #999; margin-top: 2rpx; }
}
.nursing-name { font-size: 28rpx; color: #333; flex: 1; min-width: 0; }
.remove-btn {
  font-size: 24rpx;
  color: var(--danger-color);
  padding: 4rpx 16rpx;
  flex-shrink: 0;
}
.remark-input {
  margin-top: 12rpx;
  height: 64rpx;
  background: var(--bg-color);
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 26rpx;
}

/* 无计划添加护理项 */
.add-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 16rpx;
  .add-input {
    flex: 1;
    height: 72rpx;
    background: var(--bg-color);
    border-radius: 12rpx;
    padding: 0 24rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }
  .add-btn {
    width: 128rpx;
    height: 72rpx;
    background: var(--primary-color);
    color: #fff;
    border-radius: 12rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28rpx;
    flex-shrink: 0;
  }
}

/* 生命体征 */
.vitals-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}
.vital-field {
  width: calc(50% - 10rpx);
  box-sizing: border-box;
  .vital-label { display: block; font-size: 24rpx; color: #999; margin-bottom: 8rpx; }
  .vital-input {
    height: 72rpx;
    background: var(--bg-color);
    border-radius: 12rpx;
    padding: 0 20rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }
}

/* 照片 */
.photo-input {
  width: 100%;
  height: 140rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 26rpx;
  box-sizing: border-box;
}
.photo-tip {
  font-size: 22rpx;
  color: #999;
  margin-top: 12rpx;
}

/* 描述输入 */
.desc-input {
  width: 100%;
  height: 160rpx;
  background: var(--bg-color);
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}

/* 底部操作栏 */
.footer-bar {
  position: fixed;
  bottom: 0; left: 0; right: 0;
  padding: 20rpx 32rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}
.btn-submit {
  width: 100%;
  font-size: 30rpx;
  color: #fff;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-light) 100%);
  border-radius: 12rpx;
  line-height: 2.4;
  margin: 0;
  &::after { border: none; }
  &[disabled] { opacity: 0.5; }
}
</style>
