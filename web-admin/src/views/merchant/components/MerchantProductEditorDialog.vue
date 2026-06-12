<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createMerchantProduct,
  getMerchantProduct,
  getMerchantProductSpecs,
  updateMerchantProduct,
  updateMerchantProductSpecs
} from '@/api/sp'
import type {
  MerchantCategory,
  MerchantProductEditableSpec,
  MerchantProductSpecsPayload,
  MerchantProductUpsertPayload
} from '@/types/sp'
import { uploadSpImage } from '@/utils/qiniu'

const props = withDefaults(defineProps<{
  modelValue: boolean
  merchantId: number
  categories: MerchantCategory[]
  productId?: number | null
}>(), {
  productId: null
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value)
})

const loading = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const imageInputRef = ref<HTMLInputElement | null>(null)

interface ProductImageItem {
  url: string
  previewUrl: string
}

const form = reactive({
  name: '',
  category_id: undefined as number | undefined,
  description: '',
  images: [] as ProductImageItem[],
  price: 0,
  original_price: 0,
  stock: 0,
  unit: '',
  sort: 0,
  specs: [] as MerchantProductEditableSpec[]
})

const dialogTitle = computed(() => props.productId ? '编辑商品' : '新增商品')

function revokeImagePreview(image: ProductImageItem) {
  if (image.previewUrl.startsWith('blob:')) {
    URL.revokeObjectURL(image.previewUrl)
  }
}

function replaceImages(nextImages: ProductImageItem[]) {
  for (const image of form.images) {
    revokeImagePreview(image)
  }
  form.images = nextImages
}

function resetForm() {
  form.name = ''
  form.category_id = undefined
  form.description = ''
  replaceImages([])
  form.price = 0
  form.original_price = 0
  form.stock = 0
  form.unit = ''
  form.sort = 0
  form.specs = []
}

function closeDialog() {
  dialogVisible.value = false
}

function addSpec() {
  form.specs.push({
    name: '',
    values: ['']
  })
}

function removeSpec(index: number) {
  form.specs.splice(index, 1)
}

function addSpecValue(specIndex: number) {
  form.specs[specIndex]?.values.push('')
}

function removeSpecValue(specIndex: number, valueIndex: number) {
  const values = form.specs[specIndex]?.values
  if (!values) return
  values.splice(valueIndex, 1)
  if (values.length === 0) {
    values.push('')
  }
}

function removeImage(index: number) {
  const image = form.images[index]
  if (!image) return
  revokeImagePreview(image)
  form.images.splice(index, 1)
}

function triggerImageUpload() {
  if (uploading.value) return
  imageInputRef.value?.click()
}

async function handleImageChange(event: Event) {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])
  target.value = ''
  if (files.length === 0) return

  uploading.value = true
  try {
    for (const file of files) {
      const uploaded = await uploadSpImage(file, props.merchantId)
      form.images.push({
        url: uploaded.url,
        previewUrl: URL.createObjectURL(file)
      })
    }
    ElMessage.success('图片上传成功')
  } finally {
    uploading.value = false
  }
}

function normalizeSpecsPayload(): MerchantProductSpecsPayload {
  const specs = form.specs
    .map((spec) => ({
      id: spec.id,
      name: spec.name.trim(),
      values: spec.values.map((value) => value.trim()).filter(Boolean)
    }))
    .filter((spec) => spec.name)

  return {
    specs,
    skus: []
  }
}

function buildProductPayload(): MerchantProductUpsertPayload {
  return {
    category_id: form.category_id || undefined,
    name: form.name.trim(),
    description: form.description.trim(),
    images: form.images.map((image) => image.url),
    price: Number(form.price || 0),
    original_price: Number(form.original_price || 0),
    stock: Number(form.stock || 0),
    unit: form.unit.trim(),
    sort: Number(form.sort || 0),
    specs: []
  }
}

function validatePayload(payload: MerchantProductUpsertPayload) {
  if (!payload.name) {
    return '请输入商品名称'
  }
  if (!Number.isFinite(payload.price) || payload.price < 0) {
    return '请输入正确的售价'
  }
  return ''
}

async function loadFormData() {
  if (!dialogVisible.value || !props.merchantId) return

  if (!props.productId) {
    resetForm()
    return
  }

  loading.value = true
  try {
    const [product, specPayload] = await Promise.all([
      getMerchantProduct(props.merchantId, props.productId),
      getMerchantProductSpecs(props.merchantId, props.productId)
    ])

    form.name = product.name || ''
    form.category_id = product.category_id || undefined
    form.description = product.description || ''
    replaceImages(
      Array.isArray(product.images)
        ? product.images.map((image) => ({
            url: image,
            previewUrl: image
          }))
        : []
    )
    form.price = Number(product.price || 0)
    form.original_price = Number(product.original_price || 0)
    form.stock = Number(product.stock || 0)
    form.unit = product.unit || ''
    form.sort = Number(product.sort || 0)
    form.specs = Array.isArray(specPayload.specs)
      ? specPayload.specs.map((spec) => ({
          id: spec.id,
          name: spec.name || '',
          values: Array.isArray(spec.values) && spec.values.length > 0 ? [...spec.values] : ['']
        }))
      : []
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!props.merchantId || submitting.value) return

  const payload = buildProductPayload()
  const message = validatePayload(payload)
  if (message) {
    ElMessage.warning(message)
    return
  }

  submitting.value = true
  try {
    const product = props.productId
      ? await updateMerchantProduct(props.merchantId, props.productId, payload)
      : await createMerchantProduct(props.merchantId, payload)

    await updateMerchantProductSpecs(props.merchantId, product.id, normalizeSpecsPayload())
    ElMessage.success(props.productId ? '商品更新成功' : '商品创建成功')
    emit('success')
    closeDialog()
  } finally {
    submitting.value = false
  }
}

watch(
  () => [props.modelValue, props.productId, props.merchantId],
  ([visible]) => {
    if (visible) {
      void loadFormData()
    }
  }
)
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="dialogTitle" width="880px" destroy-on-close>
    <div v-loading="loading">
      <el-form label-width="96px">
        <el-form-item label="商品名称" required>
          <el-input v-model="form.name" maxlength="50" show-word-limit placeholder="请输入商品名称" />
        </el-form-item>

        <el-form-item label="商品分类">
          <el-select v-model="form.category_id" clearable placeholder="请选择商品分类" style="width: 100%;">
            <el-option
              v-for="category in categories"
              :key="category.id"
              :label="category.name"
              :value="category.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="商品描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请输入商品描述"
          />
        </el-form-item>

        <el-form-item label="商品图片">
          <div class="image-editor">
            <div v-for="(image, index) in form.images" :key="`${image.url}-${index}`" class="image-card">
              <img :src="image.previewUrl || image.url" alt="商品图片" class="image-preview" />
              <el-button link type="danger" @click="removeImage(index)">删除</el-button>
            </div>
            <div class="image-actions">
              <input
                ref="imageInputRef"
                hidden
                type="file"
                accept="image/*"
                multiple
                @change="handleImageChange"
              />
              <el-button :loading="uploading" @click="triggerImageUpload">
                {{ uploading ? '上传中...' : '上传图片' }}
              </el-button>
            </div>
          </div>
        </el-form-item>

        <div class="form-grid">
          <el-form-item label="售价" required>
            <el-input-number v-model="form.price" :min="0" :precision="2" :step="1" style="width: 100%;" />
          </el-form-item>
          <el-form-item label="原价">
            <el-input-number v-model="form.original_price" :min="0" :precision="2" :step="1" style="width: 100%;" />
          </el-form-item>
          <el-form-item label="库存">
            <el-input-number v-model="form.stock" :min="0" :precision="0" :step="1" style="width: 100%;" />
          </el-form-item>
          <el-form-item label="单位">
            <el-input v-model="form.unit" placeholder="如：份、件、杯" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="form.sort" :min="0" :precision="0" :step="1" style="width: 100%;" />
          </el-form-item>
        </div>

        <el-form-item label="商品规格">
          <div class="spec-editor">
            <div v-if="form.specs.length === 0" class="empty-spec">未配置规格，保存后将按单规格商品处理。</div>
            <div v-for="(spec, specIndex) in form.specs" :key="spec.id ?? specIndex" class="spec-card">
              <div class="spec-card-header">
                <el-input v-model="spec.name" placeholder="规格名，如：口味、温度" />
                <el-button link type="danger" @click="removeSpec(specIndex)">删除规格</el-button>
              </div>
              <div v-for="(_, valueIndex) in spec.values" :key="valueIndex" class="spec-value-row">
                <el-input v-model="spec.values[valueIndex]" placeholder="规格值，如：大杯、少冰" />
                <el-button link type="danger" @click="removeSpecValue(specIndex, valueIndex)">删除</el-button>
              </div>
              <el-button link type="primary" @click="addSpecValue(specIndex)">新增规格值</el-button>
            </div>
            <el-button plain type="primary" @click="addSpec">新增规格</el-button>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <el-space>
        <el-button @click="closeDialog">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </el-space>
    </template>
  </el-dialog>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.image-editor {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  width: 100%;
}

.image-card,
.image-actions {
  width: 120px;
}

.image-preview {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  display: block;
  margin-bottom: 8px;
}

.image-actions {
  display: flex;
  align-items: center;
}

.spec-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.spec-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: #f9fafb;
}

.spec-card-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.spec-value-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  margin-bottom: 8px;
}

.empty-spec {
  color: #6b7280;
}
</style>
