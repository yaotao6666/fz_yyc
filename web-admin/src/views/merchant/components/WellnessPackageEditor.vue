<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { WellnessPackageContent, WellnessPackageItem } from '@/types/sp'

const props = defineProps<{
  modelValue?: WellnessPackageContent | Record<string, unknown> | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: WellnessPackageContent): void
}>()

// 本地 form 数据
const form = reactive<WellnessPackageContent>({
  duration: '',
  items: [],
  notes: '',
  applicable_groups: ''
})

// 生成临时 ID
function genTempId(): string {
  return `tmp_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

// 初始化：从 props.modelValue 反序列化
function hydrate(value: WellnessPackageContent | Record<string, unknown> | null | undefined) {
  if (!value || typeof value !== 'object') {
    form.duration = ''
    form.items = []
    form.notes = ''
    form.applicable_groups = ''
    return
  }
  const raw = value as WellnessPackageContent
  form.duration = raw.duration || ''
  form.notes = raw.notes || ''
  form.applicable_groups = raw.applicable_groups || ''
  // 规范化 items
  const items: WellnessPackageItem[] = Array.isArray(raw.items)
    ? raw.items.map((item) => ({
        id: item.id || genTempId(),
        name: item.name || '',
        description: item.description || '',
        count: Number(item.count || 0),
        unit: item.unit || '次'
      }))
    : []
  form.items = items
}

hydrate(props.modelValue)

watch(
  () => props.modelValue,
  (next) => {
    hydrate(next)
  },
  { deep: true }
)

// 变化时向上 emit
function emitChange() {
  const payload: WellnessPackageContent = {
    duration: form.duration || undefined,
    items: (form.items || []).map((item) => ({
      id: item.id,
      name: item.name,
      description: item.description || undefined,
      count: Number(item.count || 0),
      unit: item.unit || '次'
    })),
    notes: form.notes || undefined,
    applicable_groups: form.applicable_groups || undefined
  }
  emit('update:modelValue', payload)
}

watch(
  () => [form.duration, form.items, form.notes, form.applicable_groups],
  () => {
    emitChange()
  },
  { deep: true }
)

function addItem() {
  form.items?.push({
    id: genTempId(),
    name: '',
    description: '',
    count: 1,
    unit: '次'
  })
}

function removeItem(index: number) {
  form.items?.splice(index, 1)
}
</script>

<template>
  <div class="wellness-editor">
    <div class="editor-grid">
      <el-form-item label="套餐周期">
        <el-input
          v-model="form.duration"
          maxlength="64"
          placeholder="如：一个月、三个月、半年"
          clearable
        />
      </el-form-item>
      <el-form-item label="适用人群">
        <el-input
          v-model="form.applicable_groups"
          maxlength="128"
          placeholder="如：高血压人群、术后康复者"
          clearable
        />
      </el-form-item>
    </div>

    <el-form-item label="服务项列表">
      <div class="items-editor">
        <div v-if="!form.items || form.items.length === 0" class="empty-items">
          未配置服务项，请点击下方按钮添加。
        </div>
        <div
          v-for="(item, index) in form.items"
          :key="item.id ?? index"
          class="item-card"
        >
          <div class="item-header">
            <span class="item-index">服务项 {{ index + 1 }}</span>
            <el-button link type="danger" @click="removeItem(index)">删除</el-button>
          </div>
          <div class="item-grid">
            <el-input
              v-model="item.name"
              placeholder="服务项名称，如：血压测量"
              maxlength="64"
            />
            <el-input
              v-model="item.description"
              placeholder="服务项描述（选填）"
              maxlength="256"
              clearable
            />
            <div class="count-wrap">
              <el-input-number
                v-model="item.count"
                :min="0"
                :precision="0"
                :step="1"
                style="width: 100%;"
              />
            </div>
            <el-select v-model="item.unit" placeholder="单位">
              <el-option label="次" value="次" />
              <el-option label="小时" value="小时" />
              <el-option label="天" value="天" />
              <el-option label="项" value="项" />
              <el-option label="个" value="个" />
            </el-select>
          </div>
        </div>
        <el-button plain type="primary" @click="addItem">新增服务项</el-button>
      </div>
    </el-form-item>

    <el-form-item label="套餐备注">
      <el-input
        v-model="form.notes"
        type="textarea"
        :rows="3"
        maxlength="500"
        show-word-limit
        placeholder="套餐使用说明、注意事项等"
      />
    </el-form-item>
  </div>
</template>

<style scoped>
.wellness-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
}

.editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.items-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.empty-items {
  color: #6b7280;
  padding: 12px;
  text-align: center;
  background: #fff;
  border-radius: 8px;
  border: 1px dashed #e5e7eb;
}

.item-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-index {
  font-weight: 600;
  color: #111827;
  font-size: 14px;
}

.item-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 1.2fr) minmax(0, 0.8fr) minmax(0, 0.6fr);
  gap: 10px;
  align-items: center;
}

.count-wrap {
  width: 100%;
}

/* 嵌套在 dialog 内的 form-item 与外层 label 宽度协调 */
.wellness-editor :deep(.el-form-item) {
  margin-bottom: 0;
}
</style>
