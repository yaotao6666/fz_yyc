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
  MerchantProductUpsertPayload,
  WellnessPackageContent
} from '@/types/sp'
import WellnessPackageEditor from './WellnessPackageEditor.vue'
import { uploadSpImage } from '@/utils/qiniu'

const props = withDefaults(defineProps<{
  modelValue: boolean
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
  product_type: 1 as number,
  service_content: null as WellnessPackageContent | Record<string, unknown> | null,
  sale_type: 1 as number,
  rental_unit: 0 as number,
  rental_price: 0 as number,
  deposit: 0 as number,
  max_rental_duration: 0 as number,
  sort: 0,
  sales: 0,
  specs: [] as MerchantProductEditableSpec[]
})

// 商品类型选项：1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务 5=科普资讯
const PRODUCT_TYPE_OPTIONS = [
  { label: '辅具零售', value: 1 },
  { label: '辅具租赁', value: 2 },
  { label: '康养套餐', value: 3 },
  { label: '陪诊服务', value: 4 },
  { label: '科普资讯', value: 5 }
]

// 商品类型切换时自动归一化 sale_type：租赁=2，其他=1
function handleProductTypeChange(value: number) {
  if (!value) return
  if (value === 2) {
    form.sale_type = 2
  } else {
    form.sale_type = 1
    // 切换到非租赁时清理租赁字段，避免脏数据
    form.rental_unit = 0
    form.rental_price = 0
    form.deposit = 0
    form.max_rental_duration = 0
  }
  // 康养套餐自动初始化 service_content 空结构
  if (value === 3 && !form.service_content) {
    form.service_content = {
      duration: '',
      items: [],
      notes: '',
      applicable_groups: ''
    }
  }
  // 非康养套餐清理 service_content
  if (value !== 3) {
    form.service_content = null
  }
}

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
  form.product_type = 1
  form.service_content = null
  form.sale_type = 1
  form.rental_unit = 0
  form.rental_price = 0
  form.deposit = 0
  form.max_rental_duration = 0
  form.sort = 0
  form.sales = 0
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
      const uploaded = await uploadSpImage(file)
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
    product_type: Number(form.product_type || 1),
    service_content: form.product_type === 3 ? form.service_content : null,
    sale_type: Number(form.sale_type || 1),
    rental_unit: Number(form.rental_unit || 0),
    rental_price: Number(form.rental_price || 0),
    deposit: Number(form.deposit || 0),
    max_rental_duration: Number(form.max_rental_duration || 0),
    sort: Number(form.sort || 0),
    sales: form.sales,
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
  if (payload.sale_type === 2) {
    if (!payload.rental_unit) return '租赁商品必须选择计费周期'
    const rentalPrice = Number(payload.rental_price ?? 0)
    if (!Number.isFinite(rentalPrice) || rentalPrice <= 0) {
      return '租赁商品单位租金必须大于 0'
    }
    const deposit = Number(payload.deposit ?? 0)
    if (!Number.isFinite(deposit) || deposit < 0) {
      return '押金不能小于 0'
    }
  }
  // 康养套餐校验
  if (payload.product_type === 3) {
    const content = payload.service_content as WellnessPackageContent | null | undefined
    if (content && Array.isArray(content.items)) {
      for (let i = 0; i < content.items.length; i++) {
        const item = content.items[i]
        if (!item?.name?.trim()) {
          return `服务项 ${i + 1} 请填写服务项名称`
        }
      }
    }
  }
  return ''
}

async function loadFormData() {
  if (!dialogVisible.value) return

  if (!props.productId) {
    resetForm()
    return
  }

  loading.value = true
  try {
    const [product, specPayload] = await Promise.all([
      getMerchantProduct(props.productId),
      getMerchantProductSpecs(props.productId)
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
    form.product_type = Number(product.product_type || 1)
    // 康养套餐读取 service_content，其他类型置空
    if (form.product_type === 3 && product.service_content) {
      form.service_content = product.service_content as WellnessPackageContent
    } else {
      form.service_content = null
    }
    form.sale_type = Number(product.sale_type || 1)
    form.rental_unit = Number(product.rental_unit || 0)
    form.rental_price = Number(product.rental_price || 0)
    form.deposit = Number(product.deposit || 0)
    form.max_rental_duration = Number(product.max_rental_duration || 0)
    form.sort = Number(product.sort || 0)
    form.sales = Number(product.sales || 0)
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
  if (submitting.value) return

  const payload = buildProductPayload()
  const message = validatePayload(payload)
  if (message) {
    ElMessage.warning(message)
    return
  }

  submitting.value = true
  try {
    const product = props.productId
      ? await updateMerchantProduct(props.productId, payload)
      : await createMerchantProduct(payload)

    await updateMerchantProductSpecs(product.id, normalizeSpecsPayload())
    ElMessage.success(props.productId ? '商品更新成功' : '商品创建成功')
    emit('success')
    closeDialog()
  } finally {
    submitting.value = false
  }
}

watch(
  () => [props.modelValue, props.productId],
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
          <el-cascader
            v-model="form.category_id"
            :options="categories"
            :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true, emitPath: false }"
            clearable
            placeholder="请选择商品分类（可任意层级）"
            style="width: 100%;"
          />
          <div class="form-tip">支持最多三级分类，商品可挂到任意层级。</div>
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
          <el-form-item label="商品类型" required>
            <el-select
              v-model="form.product_type"
              placeholder="请选择商品类型"
              style="width: 100%;"
              @change="handleProductTypeChange"
            >
              <el-option
                v-for="opt in PRODUCT_TYPE_OPTIONS"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="销售模式">
            <el-tag :type="form.sale_type === 2 ? 'warning' : 'success'" effect="light">
              {{ form.sale_type === 2 ? '租赁（由商品类型自动）' : '一口价（由商品类型自动）' }}
            </el-tag>
          </el-form-item>
          <template v-if="form.sale_type === 2">
            <el-form-item label="计费周期" required>
              <el-select v-model="form.rental_unit" placeholder="选择计费周期" style="width: 100%;">
                <el-option label="按天" :value="1" />
                <el-option label="按周" :value="2" />
                <el-option label="按月" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item label="单位租金" required>
              <el-input-number v-model="form.rental_price" :min="0" :precision="2" :step="1" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="押金">
              <el-input-number v-model="form.deposit" :min="0" :precision="2" :step="1" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="最大租赁时长">
              <el-input-number v-model="form.max_rental_duration" :min="0" :precision="0" :step="1" style="width: 100%;" placeholder="0=不限" />
            </el-form-item>
          </template>
          <el-form-item label="排序">
            <el-input-number v-model="form.sort" :min="0" :precision="0" :step="1" style="width: 100%;" />
          </el-form-item>
          <el-form-item label="销量">
            <el-input-number v-model="form.sales" :min="0" :precision="0" :step="1" style="width: 100%;" />
          </el-form-item>
        </div>

        <template v-if="form.product_type === 3">
          <el-form-item label="康养套餐配置" required>
            <WellnessPackageEditor v-model="form.service_content" />
          </el-form-item>
        </template>

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

.form-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
