<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  createMerchantProduct,
  getMerchantCategories,
  getMerchantProduct,
  updateMerchantProduct
} from '@/api/sp'
import type {
  MerchantCategory,
  MerchantProductSpec,
  MerchantProductUpsertPayload
} from '@/types/sp'
import { uploadSpImage } from '@/utils/qiniu'

type ProductSpecOptionForm = {
  name: string
  price: number
}

type ProductSpecForm = {
  name: string
  options: ProductSpecOptionForm[]
}

const route = useRoute()
const router = useRouter()
const merchantId = computed(() => Number(route.params.id || 0))
const productId = computed(() => Number(route.params.productId || 0))
const isEditMode = computed(() => productId.value > 0)
// web-admin 为服务商专属后台，登录者必为服务商
const isServiceProvider = true

const loading = ref(false)
const saving = ref(false)
const imageUploading = ref(false)
const categories = ref<MerchantCategory[]>([])

const form = reactive({
  name: '',
  category_id: 0,
  description: '',
  price: 0,
  original_price: 0,
  stock: 0,
  unit: '份',
  sort: 0,
  sales: 0,
  images: [] as string[],
  specs: [] as ProductSpecForm[]
})

function createEmptySpec(): ProductSpecForm {
  return {
    name: '',
    options: [{ name: '', price: 0 }]
  }
}

async function loadCategories() {
  categories.value = await getMerchantCategories(merchantId.value)
}

function normalizeSpecs(specs?: MerchantProductSpec[]) {
  return (specs || []).map((spec) => ({
    name: spec.name,
    options: (spec.options || []).map((option) => ({
      name: option.name,
      price: Number(option.price || 0)
    }))
  }))
}

async function loadProduct() {
  if (!isEditMode.value) return
  loading.value = true
  try {
    const product = await getMerchantProduct(merchantId.value, productId.value)
    form.name = product.name
    form.category_id = product.category_id
    form.description = product.description || ''
    form.price = Number(product.price || 0)
    form.original_price = Number(product.original_price || 0)
    form.stock = Number(product.stock || 0)
    form.unit = product.unit || '份'
    form.sort = Number(product.sort || 0)
    form.sales = Number(product.sales || 0)
    form.images = [...(product.images || [])]
    form.specs = normalizeSpecs(product.specs)
  } finally {
    loading.value = false
  }
}

function addSpec() {
  form.specs.push(createEmptySpec())
}

function removeSpec(index: number) {
  form.specs.splice(index, 1)
}

function addSpecOption(index: number) {
  form.specs[index].options.push({ name: '', price: 0 })
}

function removeSpecOption(specIndex: number, optionIndex: number) {
  form.specs[specIndex].options.splice(optionIndex, 1)
}

async function handleImageChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (!file) return

  imageUploading.value = true
  try {
    const result = await uploadSpImage(file)
    form.images.push(result.url)
    ElMessage.success('图片上传成功')
  } finally {
    imageUploading.value = false
  }
}

function removeImage(index: number) {
  form.images.splice(index, 1)
}

function validateForm() {
  if (!form.name.trim()) throw new Error('请输入商品名称')
  if (!form.category_id) throw new Error('请选择商品分类')
  if (form.price <= 0) throw new Error('请输入正确的商品售价')
  if (form.stock < 0) throw new Error('库存不能小于 0')

  for (const spec of form.specs) {
    if (!spec.name.trim()) throw new Error('规格组名称不能为空')
    if (spec.options.length === 0) throw new Error('规格组至少保留一个规格项')
    for (const option of spec.options) {
      if (!option.name.trim()) throw new Error('规格项名称不能为空')
    }
  }
}

function buildPayload(): MerchantProductUpsertPayload {
  return {
    name: form.name.trim(),
    category_id: form.category_id,
    description: form.description.trim(),
    price: Number(form.price || 0),
    original_price: Number(form.original_price || 0),
    stock: Number(form.stock || 0),
    unit: form.unit.trim() || '份',
    sort: Number(form.sort || 0),
    sales: form.sales,
    images: [...form.images],
    specs: form.specs.map((spec) => ({
      name: spec.name.trim(),
      options: spec.options.map((option) => ({
        name: option.name.trim(),
        price: Number(option.price || 0)
      }))
    }))
  }
}

async function submitForm() {
  try {
    validateForm()
  } catch (error: any) {
    ElMessage.warning(error.message)
    return
  }

  saving.value = true
  try {
    const payload = buildPayload()
    if (isEditMode.value) {
      await updateMerchantProduct(merchantId.value, productId.value, payload)
      ElMessage.success('商品已更新')
    } else {
      await createMerchantProduct(merchantId.value, payload)
      ElMessage.success('商品已创建')
    }
    router.push(`/merchants/${merchantId.value}`)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadCategories()
  await loadProduct()
})
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">{{ isEditMode ? '编辑商品' : '新增商品' }}</h1>
        <p class="page-subtitle">服务商可直接代管商家的商品基础信息、图片和规格。</p>
      </div>
      <el-button @click="router.push(`/merchants/${merchantId}`)">返回商家详情</el-button>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <div class="section-grid single-column">
        <el-card class="page-card" shadow="never">
          <template #header>
            <span>商品基础信息</span>
          </template>
          <el-form label-position="top">
            <el-form-item label="商品名称">
              <el-input v-model="form.name" maxlength="64" placeholder="请输入商品名称" />
            </el-form-item>
            <el-form-item label="商品分类">
              <el-select v-model="form.category_id" placeholder="选择分类" style="width: 100%;">
                <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="商品描述">
              <el-input v-model="form.description" type="textarea" :rows="4" placeholder="请输入商品描述" />
            </el-form-item>

            <div class="inline-grid">
              <el-form-item label="售价">
                <el-input-number v-model="form.price" :min="0" :precision="2" style="width: 100%;" />
              </el-form-item>
              <el-form-item label="划线价">
                <el-input-number v-model="form.original_price" :min="0" :precision="2" style="width: 100%;" />
              </el-form-item>
              <el-form-item label="库存">
                <el-input-number v-model="form.stock" :min="0" style="width: 100%;" />
              </el-form-item>
              <el-form-item label="排序值">
                <el-input-number v-model="form.sort" :min="0" :max="9999" style="width: 100%;" />
              </el-form-item>
              <el-form-item v-if="isServiceProvider" label="销量">
                <el-input-number v-model="form.sales" :min="0" style="width: 100%;" />
              </el-form-item>
            </div>

            <el-form-item label="单位">
              <el-input v-model="form.unit" maxlength="8" placeholder="如：份、杯、盒" />
            </el-form-item>

            <el-form-item label="商品图片">
              <div class="image-grid">
                <div v-for="(image, index) in form.images" :key="`${image}-${index}`" class="image-card">
                  <img :src="image" alt="商品图片" />
                  <el-button link type="danger" @click="removeImage(index)">删除</el-button>
                </div>
                <label class="image-upload-trigger">
                  <input hidden type="file" accept="image/*" @change="handleImageChange" />
                  <el-button type="primary" plain :loading="imageUploading">上传图片</el-button>
                </label>
              </div>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <div class="spec-header">
              <span>商品规格</span>
              <el-button type="primary" plain @click="addSpec">新增规格组</el-button>
            </div>
          </template>

          <div v-if="form.specs.length === 0" class="empty-block">暂无规格，商品将按单一价格售卖。</div>

          <div v-for="(spec, specIndex) in form.specs" :key="specIndex" class="spec-card">
            <div class="spec-card-header">
              <el-input v-model="spec.name" placeholder="规格组名称，如：温度 / 尺寸" />
              <el-button link type="danger" @click="removeSpec(specIndex)">删除规格组</el-button>
            </div>

            <div v-for="(option, optionIndex) in spec.options" :key="optionIndex" class="spec-option-row">
              <el-input v-model="option.name" placeholder="规格项名称，如：大杯 / 热" />
              <el-input-number v-model="option.price" :min="0" :precision="2" style="width: 180px;" />
              <el-button
                link
                type="danger"
                :disabled="spec.options.length === 1"
                @click="removeSpecOption(specIndex, optionIndex)"
              >
                删除
              </el-button>
            </div>

            <el-button plain @click="addSpecOption(specIndex)">新增规格项</el-button>
          </div>
        </el-card>

        <div class="page-actions">
          <el-button @click="router.push(`/merchants/${merchantId}`)">取消</el-button>
          <el-button type="primary" :loading="saving" @click="submitForm">保存商品</el-button>
        </div>
      </div>
    </el-skeleton>
  </div>
</template>

<style scoped>
.single-column {
  grid-template-columns: 1fr;
}

.inline-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.image-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.image-card {
  width: 120px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.image-card img {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 16px;
  border: 1px solid #e5e7eb;
}

.image-upload-trigger {
  display: inline-flex;
  align-items: center;
}

.spec-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.empty-block {
  padding: 16px;
  background: #f8fafc;
  border-radius: 16px;
  color: #6b7280;
}

.spec-card {
  margin-bottom: 20px;
  padding: 20px;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  background: #fff;
}

.spec-card-header {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 16px;
}

.spec-option-row {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 12px;
}

.page-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
