<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createMerchantPickupPoint,
  deleteMerchantPickupPoint,
  getMerchantDetail,
  getMerchantPickupPoints,
  updateMerchantPickupPoint
} from '@/api/sp'
import type { MerchantDetail, MerchantPickupPoint, MerchantPickupPointPayload } from '@/types/sp'

const route = useRoute()
const router = useRouter()

const merchantId = computed(() => Number(route.params.id || 0))
const merchant = ref<MerchantDetail | null>(null)
const points = ref<MerchantPickupPoint[]>([])

const loading = ref(false)
const dialogVisible = ref(false)
const dialogSubmitting = ref(false)
const deletingId = ref<number | null>(null)
const editingPoint = ref<MerchantPickupPoint | null>(null)

const form = reactive({
  name: '',
  address: '',
  lat: 0,
  lng: 0,
  sort: 0,
  is_default: false,
  status: 1
})

function resetForm() {
  form.name = ''
  form.address = ''
  form.lat = 0
  form.lng = 0
  form.sort = 0
  form.is_default = false
  form.status = 1
}

function openCreate() {
  editingPoint.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(point: MerchantPickupPoint) {
  editingPoint.value = point
  form.name = point.name
  form.address = point.address
  form.lat = point.lat
  form.lng = point.lng
  form.sort = point.sort
  form.is_default = point.is_default
  form.status = point.status
  dialogVisible.value = true
}

function buildPayload(overrides?: Partial<MerchantPickupPointPayload>): MerchantPickupPointPayload {
  return {
    name: String(overrides?.name ?? form.name).trim(),
    address: String(overrides?.address ?? form.address).trim(),
    lat: Number(overrides?.lat ?? form.lat),
    lng: Number(overrides?.lng ?? form.lng),
    sort: Number(overrides?.sort ?? form.sort),
    status: Number(overrides?.status ?? form.status),
    is_default: overrides?.is_default ?? form.is_default
  }
}

function validatePayload(payload: MerchantPickupPointPayload) {
  if (!payload.name) {
    return '请输入自提点名称'
  }
  if (!payload.address) {
    return '请输入自提点地址'
  }
  if (!Number.isFinite(payload.lat) || !Number.isFinite(payload.lng) || payload.lat === 0 || payload.lng === 0) {
    return '请输入正确的经纬度'
  }
  return ''
}

async function loadData() {
  if (!merchantId.value) return
  loading.value = true
  try {
    const [merchantDetail, pickupPoints] = await Promise.all([
      getMerchantDetail(merchantId.value),
      getMerchantPickupPoints(merchantId.value)
    ])
    merchant.value = merchantDetail
    points.value = pickupPoints
  } finally {
    loading.value = false
  }
}

async function submitForm() {
  if (!merchantId.value || dialogSubmitting.value) return

  const payload = buildPayload()
  const message = validatePayload(payload)
  if (message) {
    return ElMessage.warning(message)
  }

  dialogSubmitting.value = true
  try {
    if (editingPoint.value) {
      await updateMerchantPickupPoint(merchantId.value, editingPoint.value.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createMerchantPickupPoint(merchantId.value, payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    await loadData()
  } finally {
    dialogSubmitting.value = false
  }
}

async function handleDelete(point: MerchantPickupPoint) {
  if (!merchantId.value) return
  try {
    await ElMessageBox.confirm(`确认删除自提点「${point.name}」？`, '删除自提点', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  deletingId.value = point.id
  try {
    await deleteMerchantPickupPoint(merchantId.value, point.id)
    ElMessage.success('删除成功')
    await loadData()
  } finally {
    deletingId.value = null
  }
}

async function setDefault(point: MerchantPickupPoint) {
  if (!merchantId.value) return
  await updateMerchantPickupPoint(merchantId.value, point.id, {
    name: point.name,
    address: point.address,
    lat: point.lat,
    lng: point.lng,
    sort: point.sort,
    status: point.status,
    is_default: true
  })
  ElMessage.success('已设为默认')
  await loadData()
}

async function toggleStatus(point: MerchantPickupPoint, nextStatus: number) {
  if (!merchantId.value) return
  await updateMerchantPickupPoint(merchantId.value, point.id, {
    name: point.name,
    address: point.address,
    lat: point.lat,
    lng: point.lng,
    sort: point.sort,
    status: nextStatus,
    is_default: point.is_default
  })
  ElMessage.success(nextStatus === 1 ? '已启用' : '已停用')
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">自提点管理</h1>
        <p class="page-subtitle">
          {{ merchant ? `当前商家：${merchant.name}` : '维护商家自提点（默认/启用状态会影响用户下单选择）' }}
        </p>
      </div>
      <el-space>
        <el-button @click="router.push(`/merchants/${merchantId}`)">返回商家</el-button>
        <el-button type="primary" @click="openCreate">新增自提点</el-button>
      </el-space>
    </div>

    <el-card class="page-card" shadow="never">
      <el-table :data="points" v-loading="loading" style="width: 100%;">
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="address" label="地址" min-width="240" />
        <el-table-column label="经纬度" width="220">
          <template #default="scope">
            {{ scope.row.lat }}, {{ scope.row.lng }}
          </template>
        </el-table-column>
        <el-table-column label="默认" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.is_default ? 'success' : 'info'">
              {{ scope.row.is_default ? '默认' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'info'">
              {{ scope.row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="100" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="scope">
            <el-space>
              <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
              <el-button
                v-if="!scope.row.is_default"
                link
                type="primary"
                @click="setDefault(scope.row)"
              >
                设为默认
              </el-button>
              <el-button
                link
                type="primary"
                @click="toggleStatus(scope.row, scope.row.status === 1 ? 0 : 1)"
              >
                {{ scope.row.status === 1 ? '停用' : '启用' }}
              </el-button>
              <el-button
                link
                type="danger"
                :loading="deletingId === scope.row.id"
                @click="handleDelete(scope.row)"
              >
                删除
              </el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && points.length === 0" description="暂无自提点，请先新增" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingPoint ? '编辑自提点' : '新增自提点'" width="520px">
      <el-form label-width="92px" style="padding-right: 12px;">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="例如：XX店自提点" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" placeholder="请输入详细地址" />
        </el-form-item>
        <el-form-item label="纬度">
          <el-input-number v-model="form.lat" :precision="6" :step="0.000001" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="经度">
          <el-input-number v-model="form.lng" :precision="6" :step="0.000001" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="默认">
          <el-switch v-model="form.is_default" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-space>
          <el-button :disabled="dialogSubmitting" @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="dialogSubmitting" @click="submitForm">保存</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

