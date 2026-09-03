<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createPrinter,
  deletePrinter,
  getPrinters,
  testPrinter,
  updatePrinter
} from '@/api/sp'
import type { Printer, PrinterPayload } from '@/types/sp'
import { PrinterTypeText } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const list = ref<Printer[]>([])

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const saving = ref(false)
const editingId = ref<number>(0)
const hasUkey = ref(false) // 编辑时后端已保存密钥，避免重复提交
const form = reactive<PrinterPayload>({
  name: '',
  type: 1,
  feie_user: '',
  feie_ukey: '',
  feie_sn: '',
  status: 1,
  auto_print: 1,
  remark: ''
})

const dialogTitle = computed(() => (dialogMode.value === 'create' ? '新增打印机' : '编辑打印机'))

function printerTypeText(type: number) {
  return PrinterTypeText[type] || '未知'
}

type TagType = 'success' | 'info' | 'warning' | 'danger' | 'primary'

function statusText(s: number) {
  return s === 1 ? '启用' : '禁用'
}

function statusType(s: number): TagType {
  return s === 1 ? 'success' : 'info'
}

function autoPrintText(v: number) {
  return v === 1 ? '自动' : '手动'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getPrinters()
    list.value = res.list || []
  } catch (e) {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = 0
  hasUkey.value = false
  Object.assign(form, {
    name: '',
    type: 1,
    feie_user: '',
    feie_ukey: '',
    feie_sn: '',
    status: 1,
    auto_print: 1,
    remark: ''
  })
  dialogVisible.value = true
}

function openEdit(row: Printer) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  hasUkey.value = Boolean(row.has_feie_ukey)
  Object.assign(form, {
    name: row.name,
    type: row.type,
    feie_user: row.feie_user || '',
    feie_ukey: '',
    feie_sn: row.feie_sn || '',
    status: row.status,
    auto_print: row.auto_print,
    remark: row.remark || ''
  })
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写打印机名称')
    return
  }
  if (form.type === 1 && !(form.feie_sn || '').trim()) {
    ElMessage.warning('请填写飞鹅打印机编号(sn)')
    return
  }
  saving.value = true
  try {
    const payload: Partial<PrinterPayload> = { ...form }
    // 编辑时若后端已保存密钥且用户未填写新值，则不提交 feie_ukey，避免覆盖
    if (dialogMode.value === 'edit' && hasUkey.value && !payload.feie_ukey) {
      delete payload.feie_ukey
    }
    if (dialogMode.value === 'create') {
      await createPrinter(payload as PrinterPayload)
      ElMessage.success('打印机已创建')
    } else {
      await updatePrinter(editingId.value, payload)
      ElMessage.success('打印机已更新')
    }
    dialogVisible.value = false
    loadData()
  } catch (e) {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

// 状态/自动打印等属性变更：直接调用更新接口
async function handleSettingsChange(row: Printer, patch: Partial<PrinterPayload>) {
  try {
    await updatePrinter(row.id, patch)
    ElMessage.success('已保存')
    loadData()
  } catch (e) {
    // 失败时回滚界面，重新拉取数据恢复原值
    loadData()
  }
}

// 设置默认打印机（设为默认后后端会自动把其它打印机置为非默认）
async function handleSetDefault(row: Printer) {
  if (row.is_default === 1) return
  try {
    await updatePrinter(row.id, { is_default: 1 })
    ElMessage.success(`已将「${row.name}」设为默认打印机`)
    loadData()
  } catch (e) {
    loadData()
  }
}

async function handleTest(row: Printer) {
  try {
    await testPrinter(row.id)
    ElMessage.success('测试打印成功，请检查打印机是否出单')
  } catch (e) {
    ElMessage.error('测试打印失败')
  }
}

async function handleDelete(row: Printer) {
  const tip = row.is_default === 1 ? '该打印机为默认打印机，删除后系统将自动把另一台打印机设为默认。是否继续？' : ''
  try {
    await ElMessageBox.confirm(tip || `确认删除打印机「${row.name}」？`, '删除', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await deletePrinter(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    // 取消
  }
}

onMounted(loadData)
</script>

<template>
  <div class="printer-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header">
          <span class="title">打印机管理</span>
          <el-button v-permission="'printers:create'" type="primary" @click="openCreate">新增打印机</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">{{ printerTypeText(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="feie_sn" label="打印机编号" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.feie_sn || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch
              v-permission="'printers:update'"
              :model-value="row.status === 1"
              :active-value="1"
              :inactive-value="0"
              @change="(val: number) => handleSettingsChange(row, { status: val })"
            />
            <el-tag v-if="row.status === 1" :type="statusType(row.status)" size="small" class="status-tag">{{
              statusText(row.status)
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认打印机" width="110" align="center">
          <template #default="{ row }">
            <el-radio
              v-permission="'printers:update'"
              :model-value="row.is_default === 1"
              :value="1"
              @change="handleSetDefault(row)"
            >
              <el-tag v-if="row.is_default === 1" type="warning" size="small">默认</el-tag>
            </el-radio>
          </template>
        </el-table-column>
        <el-table-column label="自动打印" width="90">
          <template #default="{ row }">{{ autoPrintText(row.auto_print) }}</template>
        </el-table-column>
        <el-table-column prop="print_count" label="已打印次数" width="100">
          <template #default="{ row }">{{ row.print_count ?? 0 }} 次</template>
        </el-table-column>
        <el-table-column label="最近打印时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.last_print_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'printers:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'printers:update'" link type="primary" @click="handleTest(row)">测试打印</el-button>
            <el-button v-permission="'printers:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="打印机名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.type" style="width: 100%">
            <el-option :value="1" label="飞鹅" />
            <el-option :value="2" label="通用" />
          </el-select>
        </el-form-item>
        <template v-if="form.type === 1">
          <el-form-item label="飞鹅用户名">
            <el-input v-model="form.feie_user" placeholder="飞鹅云平台用户名" maxlength="64" />
          </el-form-item>
          <el-form-item label="飞鹅密钥">
            <el-input
              v-model="form.feie_ukey"
              type="password"
              show-password
              :placeholder="hasUkey ? '已保存，留空则不修改' : '飞鹅云平台密钥'"
              maxlength="128"
            />
          </el-form-item>
          <el-form-item label="打印机编号" required>
            <el-input v-model="form.feie_sn" placeholder="飞鹅打印机 sn 号码" maxlength="64" />
          </el-form-item>
        </template>
        <el-form-item label="自动打印">
          <el-switch v-model="form.auto_print" :active-value="1" :inactive-value="0" active-text="自动" inactive-text="手动" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="256" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title {
  font-weight: 600;
}
.status-tag {
  margin-left: 8px;
}
</style>