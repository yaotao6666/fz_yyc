<template>
  <view class="printers-container">
    <view class="hero-card">
      <view class="hero-title">打印机管理</view>
      <view class="hero-desc">支持维护普通云打印机与飞鹅打印机参数，并管理自动打印和默认打印机。</view>
    </view>

    <view class="toolbar">
      <view class="toolbar-summary">当前共 {{ printers.length }} 台打印机</view>
      <view class="toolbar-btn" @click="openCreateDialog">新增打印机</view>
    </view>

    <view v-if="printers.length === 0" class="empty-card">
      <view class="empty-title">暂未配置打印机</view>
      <view class="empty-desc">支持新增普通云打印机或飞鹅打印机，并管理打印开关。</view>
    </view>

    <view v-for="printer in printers" :key="printer.id" class="printer-card">
      <view class="printer-header">
        <view class="printer-main">
          <view class="printer-title-row">
            <text class="printer-title">{{ printer.name }}</text>
            <text class="printer-tag" :class="{ active: printer.status === 1 }">
              {{ printer.status === 1 ? '已启用' : '已停用' }}
            </text>
            <text v-if="printer.is_default" class="printer-tag primary">默认</text>
          </view>
          <view class="printer-meta">
            <text>{{ getPrinterTypeLabel(printer.type) }}</text>
            <text>设备号：{{ printer.device_no }}</text>
            <text>自动打印：{{ printer.auto_print ? '开启' : '关闭' }}</text>
          </view>
        </view>
      </view>

      <view class="printer-stats">
        <view class="stat-item">
          <text class="stat-label">累计打印</text>
          <text class="stat-value">{{ printer.print_count }}</text>
        </view>
        <view class="stat-item">
          <text class="stat-label">上次打印</text>
          <text class="stat-value small">{{ formatDateTime(printer.last_print_at) }}</text>
        </view>
      </view>

      <view class="switch-list">
        <view class="switch-row">
          <view class="switch-main">
            <text class="switch-title">启用打印机</text>
            <text class="switch-desc">关闭后当前打印机不参与打印任务</text>
          </view>
          <switch :checked="printer.status === 1" color="#007AFF" @change="(e:any) => updatePrinterStatus(printer, e)" />
        </view>
        <view class="switch-row">
          <view class="switch-main">
            <text class="switch-title">自动打印</text>
            <text class="switch-desc">开启后下单时按打印场景自动发起打印</text>
          </view>
          <switch :checked="printer.auto_print" color="#007AFF" @change="(e:any) => updatePrinterAutoPrint(printer, e)" />
        </view>
        <view class="switch-row">
          <view class="switch-main">
            <text class="switch-title">设为默认打印机</text>
            <text class="switch-desc">默认打印机优先接收未指定设备的打印任务</text>
          </view>
          <switch :checked="printer.is_default" color="#007AFF" @change="(e:any) => updatePrinterDefault(printer, e)" />
        </view>
      </view>

      <view class="printer-actions">
        <view class="action-btn" @click="openEditDialog(printer)">编辑</view>
        <view class="action-btn secondary" @click="handleTestPrint(printer)">测试打印</view>
        <view class="action-btn danger" @click="handleDelete(printer)">删除</view>
      </view>
    </view>

    <view v-if="dialogVisible" class="dialog-mask" @click="closeDialog">
      <view class="dialog-card" @click.stop>
        <view class="dialog-title">{{ editingPrinterId ? '编辑打印机' : '新增打印机' }}</view>

        <view class="form-item">
          <view class="form-label">打印机名称</view>
          <input v-model="formData.name" class="form-input" placeholder="请输入打印机名称" />
        </view>

        <view class="form-item">
          <view class="form-label">打印机类型</view>
          <view class="type-list">
            <view
              v-for="item in printerTypeOptions"
              :key="item.value"
              class="type-item"
              :class="{ active: formData.type === item.value }"
              @click="formData.type = item.value"
            >
              {{ item.label }}
            </view>
          </view>
        </view>

        <view class="form-item">
          <view class="form-label">设备编号</view>
          <input v-model="formData.device_no" class="form-input" placeholder="请输入设备编号" />
        </view>

        <view class="form-item">
          <view class="form-label">接口地址</view>
          <input v-model="formData.api_url" class="form-input" placeholder="选填，留空则使用服务商默认配置" />
        </view>

        <view class="form-item">
          <view class="form-label">API Key</view>
          <input
            v-model="formData.api_key"
            class="form-input"
            password
            :placeholder="editingPrinterHasApiKey ? '已配置，可重新输入覆盖' : '请输入 API Key'"
          />
        </view>

        <view v-if="formData.type === 'feie'">
          <view class="form-item">
            <view class="form-label">飞鹅账号</view>
            <input v-model="formData.feie_user" class="form-input" placeholder="请输入飞鹅账号" />
          </view>

          <view class="form-item">
            <view class="form-label">飞鹅 UKey</view>
            <input
              v-model="formData.feie_ukey"
              class="form-input"
              password
              :placeholder="editingPrinterHasFeieUKey ? '已配置，可重新输入覆盖' : '请输入飞鹅 UKey'"
            />
          </view>

          <view class="form-item">
            <view class="form-label">飞鹅终端号</view>
            <input v-model="formData.feie_sn" class="form-input" placeholder="请输入飞鹅打印机终端号" />
          </view>
        </view>

        <view class="form-item">
          <view class="form-label">打印场景</view>
          <view class="print-type-list">
            <view
              v-for="item in printTypeOptions"
              :key="item.value"
              class="print-type-item"
              :class="{ active: formData.print_types.includes(item.value) }"
              @click="togglePrintType(item.value)"
            >
              {{ item.label }}
            </view>
          </view>
        </view>

        <view class="switch-list form-switch-list">
          <view class="switch-row">
            <view class="switch-main">
              <text class="switch-title">启用打印机</text>
              <text class="switch-desc">保存后立即生效</text>
            </view>
            <switch :checked="formData.status === 1" color="#007AFF" @change="(e:any) => formData.status = e.detail.value ? 1 : 0" />
          </view>
          <view class="switch-row">
            <view class="switch-main">
              <text class="switch-title">自动打印</text>
              <text class="switch-desc">新订单可按配置自动打印</text>
            </view>
            <switch :checked="formData.auto_print" color="#007AFF" @change="(e:any) => formData.auto_print = !!e.detail.value" />
          </view>
          <view class="switch-row">
            <view class="switch-main">
              <text class="switch-title">设为默认</text>
              <text class="switch-desc">同一商家仅保留一台默认打印机</text>
            </view>
            <switch :checked="formData.is_default" color="#007AFF" @change="(e:any) => formData.is_default = !!e.detail.value" />
          </view>
        </view>

        <view class="dialog-actions">
          <button class="dialog-btn secondary" @click="closeDialog">取消</button>
          <button class="dialog-btn primary" :disabled="saving" @click="handleSubmit">
            {{ saving ? '提交中...' : '保存打印机' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  createMerchantPrinter,
  deleteMerchantPrinter,
  getMerchantPrinters,
  testMerchantPrinter,
  updateMerchantPrinter
} from '@api'
import type { MerchantPrinter, MerchantPrinterPayload, MerchantPrinterPrintType, MerchantPrinterType } from '@types'

const printers = ref<MerchantPrinter[]>([])
const dialogVisible = ref(false)
const editingPrinterId = ref<number | null>(null)
const editingPrinterHasApiKey = ref(false)
const editingPrinterHasFeieUKey = ref(false)
const saving = ref(false)

const printerTypeOptions: Array<{ value: MerchantPrinterType; label: string }> = [
  { value: 'yilianyun', label: '普通云打印机' },
  { value: 'feie', label: '飞鹅打印机' }
]

const printTypeOptions: Array<{ value: MerchantPrinterPrintType; label: string }> = [
  { value: 'order', label: '下单小票' },
  { value: 'complete', label: '核销小票' },
  { value: 'refund', label: '退款小票' }
]

const formData = reactive({
  name: '',
  type: 'yilianyun' as MerchantPrinterType,
  device_no: '',
  api_key: '',
  api_url: '',
  feie_user: '',
  feie_ukey: '',
  feie_sn: '',
  print_types: ['order'] as MerchantPrinterPrintType[],
  status: 1,
  auto_print: true,
  is_default: false
})

onShow(() => {
  loadPrinters()
})

function resetForm() {
  formData.name = ''
  formData.type = 'yilianyun'
  formData.device_no = ''
  formData.api_key = ''
  formData.api_url = ''
  formData.feie_user = ''
  formData.feie_ukey = ''
  formData.feie_sn = ''
  formData.print_types = ['order']
  formData.status = 1
  formData.auto_print = true
  formData.is_default = false
  editingPrinterId.value = null
  editingPrinterHasApiKey.value = false
  editingPrinterHasFeieUKey.value = false
}

async function loadPrinters() {
  try {
    printers.value = await getMerchantPrinters()
  } catch (error) {
    console.error('加载打印机列表失败:', error)
  }
}

function getPrinterTypeLabel(type: string) {
  return type === 'feie' ? '飞鹅打印机' : '普通云打印机'
}

function formatDateTime(value?: string) {
  if (!value) {
    return '暂无记录'
  }
  return value.replace('T', ' ').slice(0, 16)
}

function togglePrintType(type: MerchantPrinterPrintType) {
  if (formData.print_types.includes(type)) {
    if (formData.print_types.length === 1) {
      return
    }
    formData.print_types = formData.print_types.filter((item) => item !== type)
    return
  }
  formData.print_types = [...formData.print_types, type]
}

function openCreateDialog() {
  resetForm()
  dialogVisible.value = true
}

function openEditDialog(printer: MerchantPrinter) {
  resetForm()
  editingPrinterId.value = printer.id
  editingPrinterHasApiKey.value = !!printer.has_api_key
  editingPrinterHasFeieUKey.value = !!printer.has_feie_ukey
  formData.name = printer.name
  formData.type = printer.type
  formData.device_no = printer.device_no
  formData.api_url = printer.api_url || ''
  formData.feie_user = printer.feie_user || ''
  formData.feie_sn = printer.feie_sn || ''
  formData.print_types = printer.print_types?.length ? [...printer.print_types] : ['order']
  formData.status = printer.status
  formData.auto_print = !!printer.auto_print
  formData.is_default = !!printer.is_default
  dialogVisible.value = true
}

function closeDialog() {
  dialogVisible.value = false
  resetForm()
}

function buildPayload(): MerchantPrinterPayload {
  const name = formData.name.trim()
  const deviceNo = formData.device_no.trim()

  if (!name) {
    throw new Error('请输入打印机名称')
  }
  if (!deviceNo) {
    throw new Error('请输入设备编号')
  }
  if (formData.type === 'feie' && !formData.feie_user.trim()) {
    throw new Error('请输入飞鹅账号')
  }
  if (formData.type === 'feie' && !editingPrinterHasFeieUKey.value && !formData.feie_ukey.trim()) {
    throw new Error('请输入飞鹅 UKey')
  }
  if (formData.type === 'feie' && !formData.feie_sn.trim()) {
    throw new Error('请输入飞鹅打印机终端号')
  }

  const payload: MerchantPrinterPayload = {
    name,
    type: formData.type,
    device_no: deviceNo,
    api_url: formData.api_url.trim(),
    print_types: formData.print_types,
    status: formData.status,
    auto_print: formData.auto_print,
    is_default: formData.is_default
  }

  if (formData.api_key.trim()) {
    payload.api_key = formData.api_key.trim()
  }
  if (formData.type === 'feie') {
    payload.feie_user = formData.feie_user.trim()
    payload.feie_sn = formData.feie_sn.trim()
    if (formData.feie_ukey.trim()) {
      payload.feie_ukey = formData.feie_ukey.trim()
    }
  }

  return payload
}

async function handleSubmit() {
  if (saving.value) {
    return
  }

  try {
    saving.value = true
    const payload = buildPayload()
    if (editingPrinterId.value) {
      await updateMerchantPrinter(editingPrinterId.value, payload)
    } else {
      await createMerchantPrinter(payload)
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    closeDialog()
    await loadPrinters()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '保存失败', icon: 'none' })
  } finally {
    saving.value = false
  }
}

async function updatePrinterStatus(printer: MerchantPrinter, event: any) {
  try {
    await updateMerchantPrinter(printer.id, {
      status: event?.detail?.value ? 1 : 0
    })
    await loadPrinters()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '更新失败', icon: 'none' })
  }
}

async function updatePrinterAutoPrint(printer: MerchantPrinter, event: any) {
  try {
    await updateMerchantPrinter(printer.id, {
      auto_print: !!event?.detail?.value
    })
    await loadPrinters()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '更新失败', icon: 'none' })
  }
}

async function updatePrinterDefault(printer: MerchantPrinter, event: any) {
  const targetValue = !!event?.detail?.value
  if (!targetValue && printer.is_default) {
    uni.showToast({ title: '默认打印机需切换到其他设备', icon: 'none' })
    await loadPrinters()
    return
  }

  try {
    await updateMerchantPrinter(printer.id, {
      is_default: targetValue
    })
    await loadPrinters()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '设置失败', icon: 'none' })
  }
}

function handleDelete(printer: MerchantPrinter) {
  uni.showModal({
    title: '删除打印机',
    content: `确定删除“${printer.name}”吗？`,
    success: async (res) => {
      if (!res.confirm) {
        return
      }
      try {
        await deleteMerchantPrinter(printer.id)
        uni.showToast({ title: '删除成功', icon: 'success' })
        await loadPrinters()
      } catch (error: any) {
        uni.showToast({ title: error?.message || '删除失败', icon: 'none' })
      }
    }
  })
}

async function handleTestPrint(printer: MerchantPrinter) {
  try {
    await testMerchantPrinter(printer.id)
    uni.showToast({ title: '测试打印成功', icon: 'success' })
    await loadPrinters()
  } catch (error: any) {
    uni.showToast({ title: error?.message || '测试打印失败', icon: 'none' })
  }
}
</script>

<style scoped>
.printers-container {
  min-height: 100vh;
  background: #f5f7fb;
  padding: 24rpx 24rpx 180rpx;
}

.hero-card {
  background: linear-gradient(135deg, #222c59 0%, #47558f 100%);
  border-radius: 24rpx;
  padding: 32rpx;
  color: #ffffff;
}

.hero-title {
  font-size: 36rpx;
  font-weight: 600;
}

.hero-desc {
  margin-top: 12rpx;
  font-size: 26rpx;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.82);
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin: 24rpx 0;
}

.toolbar-summary {
  font-size: 26rpx;
  color: #5f6570;
}

.toolbar-btn {
  padding: 16rpx 24rpx;
  border-radius: 999rpx;
  background: #007AFF;
  color: #ffffff;
  font-size: 24rpx;
}

.empty-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 40rpx 32rpx;
  text-align: center;
}

.empty-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.empty-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.printer-card {
  background: #ffffff;
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.printer-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.printer-main {
  min-width: 0;
  flex: 1;
}

.printer-title-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  align-items: center;
}

.printer-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.printer-tag {
  padding: 6rpx 16rpx;
  border-radius: 999rpx;
  background: #f2f3f5;
  color: #8a9099;
  font-size: 22rpx;
}

.printer-tag.active {
  background: #e8f7ed;
  color: #1f8f47;
}

.printer-tag.primary {
  background: #eef4ff;
  color: #007AFF;
}

.printer-meta {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #70757f;
}

.printer-stats {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}

.stat-item {
  flex: 1;
  border-radius: 18rpx;
  background: #f8f9fb;
  padding: 20rpx;
}

.stat-label {
  display: block;
  font-size: 24rpx;
  color: #8a9099;
}

.stat-value {
  display: block;
  margin-top: 10rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-value.small {
  font-size: 24rpx;
  line-height: 1.5;
}

.switch-list {
  margin-top: 24rpx;
  border-top: 1rpx solid #eef1f5;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #eef1f5;
}

.switch-row:last-child {
  border-bottom: none;
}

.switch-main {
  flex: 1;
  min-width: 0;
}

.switch-title {
  display: block;
  font-size: 28rpx;
  color: #1a1a1a;
}

.switch-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.5;
  color: #8a9099;
}

.printer-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.action-btn {
  flex: 1;
  height: 76rpx;
  border-radius: 999rpx;
  background: #eef4ff;
  color: #007AFF;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn.secondary {
  background: #eef7f1;
  color: #1f8f47;
}

.action-btn.danger {
  background: #fff2f0;
  color: #ff4d4f;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
  z-index: 30;
}

.dialog-card {
  width: 100%;
  max-height: 88vh;
  overflow-y: auto;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 32rpx;
  box-sizing: border-box;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-label {
  margin-bottom: 12rpx;
  font-size: 26rpx;
  color: #4f5560;
}

.form-input {
  width: 100%;
  height: 84rpx;
  border-radius: 16rpx;
  background: #f8f9fb;
  padding: 0 22rpx;
  font-size: 30rpx;
  box-sizing: border-box;
}

.type-list,
.print-type-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.type-item,
.print-type-item {
  padding: 14rpx 22rpx;
  border-radius: 16rpx;
  background: #f4f6fa;
  color: #5f6570;
  font-size: 24rpx;
}

.type-item.active,
.print-type-item.active {
  background: #e8f1ff;
  color: #007AFF;
}

.form-switch-list {
  margin-top: 8rpx;
}

.dialog-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 32rpx;
}

.dialog-btn {
  flex: 1;
  height: 84rpx;
  border-radius: 42rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-btn.secondary {
  background: #f3f4f6;
  color: #666666;
}

.dialog-btn.primary {
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
}
</style>
