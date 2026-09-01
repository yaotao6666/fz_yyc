<template>
  <view class="record-edit-page">
    <!-- 基本信息 -->
    <view class="form-card">
      <view class="card-title">基本信息</view>
      <view class="form-item">
        <view class="form-label">与本人关系</view>
        <view class="radio-group">
          <view
            class="radio-chip"
            :class="{ active: form.relation === 1 }"
            @click="form.relation = 1"
          >本人</view>
          <view
            class="radio-chip"
            :class="{ active: form.relation === 2 }"
            @click="form.relation = 2"
          >父母</view>
          <view
            class="radio-chip"
            :class="{ active: form.relation === 3 }"
            @click="form.relation = 3"
          >其他亲属</view>
        </view>
      </view>
      <view class="form-item">
        <view class="form-label">姓名</view>
        <input v-model="form.real_name" class="form-input" placeholder="请输入真实姓名" maxlength="30" />
      </view>

      <view class="form-item">
        <view class="form-label">性别</view>
        <view class="radio-group">
          <view
            class="radio-chip"
            :class="{ active: form.gender === 1 }"
            @click="form.gender = 1"
          >男</view>
          <view
            class="radio-chip"
            :class="{ active: form.gender === 2 }"
            @click="form.gender = 2"
          >女</view>
        </view>
      </view>

      <view class="form-item">
        <view class="form-label">出生日期</view>
        <picker mode="date" :value="form.birth_date" :end="today" @change="onBirthDateChange">
          <view class="picker-value" :class="{ placeholder: !form.birth_date }">
            {{ form.birth_date || '请选择出生日期' }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <view class="form-label">身份证号</view>
        <input v-model="form.id_card" class="form-input" placeholder="请输入身份证号" maxlength="18" />
      </view>

      <view class="form-item">
        <view class="form-label">手机号</view>
        <input v-model="form.phone" class="form-input" type="number" placeholder="请输入手机号" maxlength="11" />
      </view>

      <view class="form-item">
        <view class="form-label">紧急联系人</view>
        <input v-model="form.emergency_contact" class="form-input" placeholder="请输入紧急联系人姓名" maxlength="30" />
      </view>

      <view class="form-item">
        <view class="form-label">紧急联系电话</view>
        <input v-model="form.emergency_phone" class="form-input" type="number" placeholder="请输入紧急联系电话" maxlength="11" />
      </view>

      <view class="form-item">
        <view class="form-label">现住地址</view>
        <textarea
          v-model="form.address"
          class="form-textarea"
          maxlength="120"
          placeholder="请输入详细住址"
        />
      </view>
    </view>

    <!-- 体征信息 -->
    <view class="form-card">
      <view class="card-title">体征信息</view>
      <view class="form-item-row">
        <view class="form-item half">
          <view class="form-label">身高（cm）</view>
          <input v-model="form.height_cm" class="form-input" type="number" placeholder="如 170" maxlength="3" />
        </view>
        <view class="form-item half">
          <view class="form-label">体重（kg）</view>
          <input v-model="form.weight_kg" class="form-input" type="number" placeholder="如 60" maxlength="3" />
        </view>
      </view>

      <view class="form-item">
        <view class="form-label">血型</view>
        <picker mode="selector" :range="bloodTypeLabels" @change="onBloodTypeChange">
          <view class="picker-value" :class="{ placeholder: !form.blood_type }">
            {{ form.blood_type ? `${form.blood_type}型` : '请选择血型' }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <view class="form-label">吸烟情况</view>
        <picker mode="selector" :range="smokeOptions" @change="onSmokeChange">
          <view class="picker-value" :class="{ placeholder: !form.smoking }">
            {{ form.smoking || '请选择吸烟情况' }}
          </view>
        </picker>
      </view>

      <view class="form-item">
        <view class="form-label">饮酒情况</view>
        <picker mode="selector" :range="drinkOptions" @change="onDrinkChange">
          <view class="picker-value" :class="{ placeholder: !form.drinking }">
            {{ form.drinking || '请选择饮酒情况' }}
          </view>
        </picker>
      </view>
    </view>

    <!-- 慢病标签 -->
    <view class="form-card">
      <view class="card-title">慢性病标签</view>
      <view class="tag-chips">
        <view
          v-for="tag in chronicTagOptions"
          :key="tag"
          class="tag-chip"
          :class="{ active: form.chronic_tags.includes(tag) }"
          @click="toggleChronicTag(tag)"
        >{{ tag }}</view>
      </view>
      <view class="form-tip">可多选，用于健康管理与服务推荐</view>
    </view>

    <!-- 病史信息 -->
    <view class="form-card">
      <view class="card-title">病史信息</view>
      <view class="form-item">
        <view class="form-label">既往病史</view>
        <textarea
          v-model="form.past_history"
          class="form-textarea"
          maxlength="300"
          placeholder="多个病史请用逗号或分号分隔"
        />
      </view>
      <view class="form-item">
        <view class="form-label">过敏史</view>
        <textarea
          v-model="form.allergy_history"
          class="form-textarea"
          maxlength="300"
          placeholder="多个过敏原请用逗号或分号分隔"
        />
      </view>
      <view class="form-item">
        <view class="form-label">家族病史</view>
        <textarea
          v-model="form.family_history"
          class="form-textarea"
          maxlength="300"
          placeholder="家族相关疾病请用逗号或分号分隔"
        />
      </view>
      <view class="form-item">
        <view class="form-label">手术史</view>
        <textarea
          v-model="form.surgery_history"
          class="form-textarea"
          maxlength="300"
          placeholder="手术名称与时间，多个请用逗号或分号分隔"
        />
      </view>
      <view class="form-item">
        <view class="form-label">长期用药</view>
        <textarea
          v-model="form.medication_list"
          class="form-textarea"
          maxlength="300"
          placeholder="正在服用的药物，多个请用逗号或分号分隔"
        />
      </view>
      <view class="form-item">
        <view class="form-label">备注</view>
        <textarea
          v-model="form.remark"
          class="form-textarea"
          maxlength="300"
          placeholder="其他需要说明的健康情况"
        />
      </view>
    </view>

    <view class="bottom-bar">
      <view class="submit-btn" :class="{ disabled: saving }" @click="handleSubmit">
        {{ saving ? '保存中...' : '保存档案' }}
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { listUserHealthRecords, createUserHealthRecord, updateUserHealthRecord } from '../../api/health'
import type { HealthRecord } from '../../types'
import { useAuth } from '../../utils/useAuth'

interface HealthRecordForm {
  relation: number
  real_name: string
  gender: number | null
  birth_date: string
  id_card: string
  phone: string
  emergency_contact: string
  emergency_phone: string
  address: string
  height_cm: string
  weight_kg: string
  blood_type: string
  smoking: string
  drinking: string
  past_history: string
  allergy_history: string
  family_history: string
  surgery_history: string
  medication_list: string
  chronic_tags: string[]
  remark: string
}

const chronicTagOptions = [
  '高血压',
  '糖尿病',
  '冠心病',
  '脑卒中',
  '慢阻肺',
  '骨质疏松',
  '帕金森',
  '阿尔茨海默',
  '关节炎',
  '其他'
]
const bloodTypes = ['A', 'B', 'AB', 'O']
const bloodTypeLabels = bloodTypes.map(item => `${item}型`)
const smokeOptions = ['不吸烟', '偶尔吸烟', '经常吸烟', '已戒烟']
const drinkOptions = ['不饮酒', '偶尔饮酒', '经常饮酒', '已戒酒']

const form = reactive<HealthRecordForm>({
  relation: 1,
  real_name: '',
  gender: null,
  birth_date: '',
  id_card: '',
  phone: '',
  emergency_contact: '',
  emergency_phone: '',
  address: '',
  height_cm: '',
  weight_kg: '',
  blood_type: '',
  smoking: '',
  drinking: '',
  past_history: '',
  allergy_history: '',
  family_history: '',
  surgery_history: '',
  medication_list: '',
  chronic_tags: [],
  remark: ''
})

const saving = ref(false)
const today = getToday()

function getToday(): string {
  const date = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function onBirthDateChange(event: any) {
  form.birth_date = event.detail.value || ''
}

function onBloodTypeChange(event: any) {
  form.blood_type = bloodTypes[Number(event.detail.value)] || ''
}

function onSmokeChange(event: any) {
  form.smoking = smokeOptions[Number(event.detail.value)] || ''
}

function onDrinkChange(event: any) {
  form.drinking = drinkOptions[Number(event.detail.value)] || ''
}

function toggleChronicTag(tag: string) {
  const index = form.chronic_tags.indexOf(tag)
  if (index >= 0) {
    form.chronic_tags.splice(index, 1)
  } else {
    form.chronic_tags.push(tag)
  }
}

function splitStringArray(value: string): string[] {
  return value
    .split(/[,，;；]+/)
    .map(item => item.trim())
    .filter(Boolean)
}

let editId: number | null = null

onLoad(async (options: any) => {
  const { ensureAuth } = useAuth()
  const authed = await ensureAuth()
  if (!authed) {
    uni.showToast({ title: '登录失败，请重试', icon: 'none' })
    return
  }

  // 关系：新建时可由下单页通过 relation 参数预选（默认本人）
  const relation = Number(options?.relation)
  if (relation === 2 || relation === 3) {
    form.relation = relation
  }

  const id = Number(options?.id)
  if (!id) return

  try {
    const records = await listUserHealthRecords()
    const target = records.find(item => item.id === id)
    if (target) {
      editId = id
      fillForm(target)
    }
  } catch (error) {
    console.error('加载健康档案失败:', error)
  }
})

function fillForm(record: HealthRecord) {
  form.relation = record.relation === 2 || record.relation === 3 ? record.relation : 1
  form.real_name = record.real_name || ''
  form.gender = record.gender || null
  form.birth_date = record.birth_date || ''
  form.id_card = record.id_card || ''
  form.phone = record.phone || ''
  form.emergency_contact = record.emergency_contact || ''
  form.emergency_phone = record.emergency_phone || ''
  form.address = record.address || ''
  form.height_cm = record.height_cm ? String(record.height_cm) : ''
  form.weight_kg = record.weight_kg ? String(record.weight_kg) : ''
  form.blood_type = (record.blood_type || '').replace(/型$/, '')
  form.smoking = record.smoking || ''
  form.drinking = record.drinking || ''
  form.past_history = (record.past_history || []).join('，')
  form.allergy_history = (record.allergy_history || []).join('，')
  form.family_history = (record.family_history || []).join('，')
  form.surgery_history = (record.surgery_history || []).join('，')
  form.medication_list = (record.medication_list || []).join('，')
  form.chronic_tags = record.chronic_tags || []
  form.remark = record.remark || ''
}

function validateForm(): boolean {
  if (!form.real_name.trim()) {
    uni.showToast({ title: '请输入姓名', icon: 'none' })
    return false
  }
  if (form.phone && !/^1\d{10}$/.test(form.phone.trim())) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return false
  }
  if (form.id_card && !/^\d{17}[\dXx]$/.test(form.id_card.trim())) {
    uni.showToast({ title: '请输入正确的身份证号', icon: 'none' })
    return false
  }
  if (form.emergency_phone && !/^1\d{10}$/.test(form.emergency_phone.trim())) {
    uni.showToast({ title: '请输入正确的紧急联系电话', icon: 'none' })
    return false
  }
  return true
}

function buildPayload(): Partial<HealthRecord> {
  return {
    relation: form.relation,
    real_name: form.real_name.trim(),
    gender: form.gender ?? undefined,
    birth_date: form.birth_date || undefined,
    id_card: form.id_card.trim() || undefined,
    phone: form.phone.trim() || undefined,
    emergency_contact: form.emergency_contact.trim() || undefined,
    emergency_phone: form.emergency_phone.trim() || undefined,
    address: form.address.trim() || undefined,
    height_cm: form.height_cm ? Number(form.height_cm) : null,
    weight_kg: form.weight_kg ? Number(form.weight_kg) : null,
    blood_type: form.blood_type || undefined,
    smoking: form.smoking || undefined,
    drinking: form.drinking || undefined,
    past_history: splitStringArray(form.past_history),
    allergy_history: splitStringArray(form.allergy_history),
    family_history: splitStringArray(form.family_history),
    surgery_history: splitStringArray(form.surgery_history),
    medication_list: splitStringArray(form.medication_list),
    chronic_tags: form.chronic_tags,
    remark: form.remark.trim() || undefined
  }
}

async function handleSubmit() {
  if (saving.value || !validateForm()) {
    return
  }

  try {
    saving.value = true
    const payload = buildPayload()
    if (editId) {
      await updateUserHealthRecord(editId, payload)
    } else {
      await createUserHealthRecord(payload)
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 400)
  } catch (error: any) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.record-edit-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx 24rpx 180rpx;
  box-sizing: border-box;
}

.form-card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-item-row {
  display: flex;
  gap: 20rpx;
}

.form-item.half {
  flex: 1;
  min-width: 0;
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.form-input,
.picker-value,
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  background: #f8f9fa;
  border-radius: 14rpx;
  font-size: 30rpx;
  color: #1a1a1a;
}

.form-input,
.picker-value {
  height: 84rpx;
  padding: 0 24rpx;
  display: flex;
  align-items: center;
}

.picker-value.placeholder {
  color: #999999;
}

.form-textarea {
  min-height: 160rpx;
  padding: 20rpx 24rpx;
  line-height: 1.5;
}

.radio-group {
  display: flex;
  gap: 20rpx;
}

.radio-chip {
  flex: 1;
  height: 84rpx;
  border-radius: 14rpx;
  background: #f8f9fa;
  color: #666666;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2rpx solid transparent;
}

.radio-chip.active {
  background: rgba(0, 122, 255, 0.1);
  border-color: #007AFF;
  color: #007AFF;
  font-weight: 600;
}

.tag-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.tag-chip {
  padding: 16rpx 26rpx;
  border-radius: 999rpx;
  background: #f8f9fa;
  color: #666666;
  font-size: 26rpx;
  border: 2rpx solid transparent;
}

.tag-chip.active {
  background: rgba(0, 122, 255, 0.1);
  border-color: #007AFF;
  color: #007AFF;
  font-weight: 600;
}

.form-tip {
  margin-top: 16rpx;
  font-size: 22rpx;
  color: #999999;
}

.bottom-bar {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(env(safe-area-inset-bottom) + 24rpx);
}

.submit-btn {
  height: 88rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-btn.disabled {
  opacity: 0.72;
}
</style>
