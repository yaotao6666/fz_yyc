<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSystemConfigs, updateSystemConfigs, deleteSystemConfig, type SystemConfigEntry } from '@/api/safety'

const loading = ref(false)
const list = ref<SystemConfigEntry[]>([])

// 编辑/新增对话框
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const saving = ref(false)
const editingIndex = ref(-1) // 列表下标，-1 表示新增
const form = reactive({ config_key: '', config_value: '', remark: '' })

async function loadData() {
  loading.value = true
  try {
    list.value = (await getSystemConfigs()) || []
  } catch (e) {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialogMode.value = 'create'
  editingIndex.value = -1
  form.config_key = ''
  form.config_value = ''
  form.remark = ''
  dialogVisible.value = true
}

function openEdit(row: SystemConfigEntry, index: number) {
  dialogMode.value = 'edit'
  editingIndex.value = index
  form.config_key = row.config_key
  form.config_value = row.config_value
  form.remark = row.remark || ''
  dialogVisible.value = true
}

function isValidConfigKey(key: string) {
  return /^[a-zA-Z0-9_.\-]{1,100}$/.test(key)
}

async function handleSave() {
  const key = form.config_key.trim()
  if (!key) {
    ElMessage.warning('请填写配置键(config_key)')
    return
  }
  if (!isValidConfigKey(key)) {
    ElMessage.warning('配置键仅允许字母数字及 _ . -，长度不超过 100')
    return
  }
  const value = form.config_value.trim()
  if (!value) {
    ElMessage.warning('请填写配置值(config_value)')
    return
  }
  // config_value 需为合法 JSON
  try {
    JSON.parse(value)
  } catch (e) {
    ElMessage.warning('配置值必须是合法的 JSON 字符串，例如：30 或 "{\\"key\\": 1}"')
    return
  }
  saving.value = true
  try {
    await updateSystemConfigs([{ config_key: key, config_value: value, remark: form.remark.trim() }])
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: SystemConfigEntry) {
  try {
    await ElMessageBox.confirm(`确认永久删除配置项「${row.config_key}」？删除后业务将按默认值回退。`, '删除配置', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch (e) {
    return // 取消
  }
  try {
    await deleteSystemConfig(row.config_key)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    // 拦截器已提示
  }
}

function formatTime(t?: string) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

onMounted(loadData)
</script>

<template>
  <div class="system-config-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header">
          <span class="title">系统配置</span>
          <el-button v-permission="'systemconfig:update'" type="primary" @click="openCreate">新增配置</el-button>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="通用键值配置存储（system_configs），config_value 为 JSON 字符串"
        class="top-alert"
      />

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="config_key" label="配置键" min-width="200" show-overflow-tooltip />
        <el-table-column prop="config_value" label="配置值(JSON)" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <code class="cfg-value">{{ row.config_value }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row, $index }">
            <el-button v-permission="'systemconfig:update'" link type="primary" @click="openEdit(row, $index)">编辑</el-button>
            <el-button v-permission="'systemconfig:update'" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增配置' : '编辑配置'"
      width="560px"
      @closed="form.config_key = ''"
    >
      <el-form :model="form" label-width="120px">
        <el-form-item label="配置键" required>
          <el-input
            v-model="form.config_key"
            placeholder="例如 alert.goods_unverified_hours"
            :disabled="dialogMode === 'edit'"
            maxlength="100"
          />
          <div class="form-tip">仅允许字母数字及 _ . -</div>
        </el-form-item>
        <el-form-item label="配置值(JSON)" required>
          <el-input
            v-model="form.config_value"
            type="textarea"
            :rows="3"
            placeholder='例如：30 或 {"enabled": true} 或 "文本"'
          />
          <div class="form-tip">必须是合法 JSON 字符串</div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" placeholder="配置项说明" />
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
.top-alert {
  margin-bottom: 16px;
}
.cfg-value {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
  margin-top: 2px;
}
</style>