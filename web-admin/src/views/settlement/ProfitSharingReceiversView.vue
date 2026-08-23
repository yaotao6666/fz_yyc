<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProfitSharingReceivers,
  createProfitSharingReceiver,
  updateProfitSharingReceiver,
  deleteProfitSharingReceiver,
  syncProfitSharingReceiver,
  getProfitSharingConfig,
  updateProfitSharingConfig
} from '@/api/sp'
import type { ProfitSharingReceiver, ProfitSharingReceiverPayload, ProfitSharingConfig } from '@/types/sp'

const loading = ref(false)
const list = ref<ProfitSharingReceiver[]>([])
const config = ref<ProfitSharingConfig>({ profit_sharing_enabled: false, sub_mch_id: '', max_ratio_pct: '30%' })

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const saving = ref(false)
const editingId = ref<number>(0)
const form = ref<ProfitSharingReceiverPayload>({
  receiver_type: 1,
  name: '',
  account: '',
  personal_name: '',
  relation_type: 'SERVICE_PROVIDER',
  default_ratio: 0,
  status: 1,
  sort: 0,
  remark: ''
})

const dialogTitle = computed(() => (dialogMode.value === 'create' ? '新增分账接收方' : '编辑分账接收方'))

function receiverTypeText(t: number) {
  return t === 2 ? '个人微信' : '商户号'
}

type TagType = 'success' | 'info' | 'warning' | 'danger' | 'primary'

function statusText(s: number) {
  return s === 1 ? '启用' : '停用'
}

function statusType(s: number): TagType {
  return s === 1 ? 'success' : 'info'
}

function boundText(b: number) {
  return b === 1 ? '已建立' : '未建立'
}

function boundType(b: number): TagType {
  return b === 1 ? 'success' : 'warning'
}

async function loadData() {
  loading.value = true
  try {
    list.value = (await getProfitSharingReceivers()) || []
  } catch (e) {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

async function loadConfig() {
  try {
    config.value = await getProfitSharingConfig()
  } catch (e) {
    // 忽略
  }
}

async function handleToggleShare(val: boolean) {
  try {
    await updateProfitSharingConfig(val)
    ElMessage.success(val ? '已开启自动分账' : '已关闭自动分账')
    loadConfig()
  } catch (e) {
    // 失败回滚开关
    config.value.profit_sharing_enabled = !val
  }
}

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = 0
  form.value = {
    receiver_type: 1,
    name: '',
    account: '',
    personal_name: '',
    relation_type: 'SERVICE_PROVIDER',
    default_ratio: 0,
    status: 1,
    sort: list.value.length,
    remark: ''
  }
  dialogVisible.value = true
}

function openEdit(row: ProfitSharingReceiver) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.value = {
    receiver_type: row.receiver_type,
    name: row.name,
    account: row.account,
    personal_name: row.personal_name || '',
    relation_type: row.relation_type || 'SERVICE_PROVIDER',
    default_ratio: row.default_ratio,
    status: row.status,
    sort: row.sort,
    remark: row.remark || ''
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.name || !form.value.account) {
    ElMessage.warning('请填写接收方名称与账号')
    return
  }
  if (form.value.default_ratio < 0 || form.value.default_ratio > 100) {
    ElMessage.warning('分账比例需在 0-100 之间')
    return
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createProfitSharingReceiver(form.value)
      ElMessage.success('分账接收方已创建')
    } else {
      await updateProfitSharingReceiver(editingId.value, form.value)
      ElMessage.success('分账接收方已更新')
    }
    dialogVisible.value = false
    loadData()
  } catch (e) {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: ProfitSharingReceiver) {
  try {
    await ElMessageBox.confirm(`确认删除分账接收方「${row.name}」？`, '删除', { type: 'warning' })
    await deleteProfitSharingReceiver(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    // 取消
  }
}

async function handleSync(row: ProfitSharingReceiver) {
  try {
    await syncProfitSharingReceiver(row.id)
    ElMessage.success('微信关系已同步')
    loadData()
  } catch (e) {
    // 拦截器已提示
  }
}

onMounted(() => {
  loadConfig()
  loadData()
})
</script>

<template>
  <div class="profit-receivers-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header">
          <span class="title">分账接收方管理</span>
          <el-button v-permission="'profit:receiver:manage'" type="primary" @click="openCreate">新增接收方</el-button>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        class="config-alert"
        show-icon
      >
        <template #title>
          <span>自动分账：</span>
          <el-switch
            v-model="config.profit_sharing_enabled"
            :disabled="!config.sub_mch_id"
            active-text="开启"
            inactive-text="关闭"
            @change="handleToggleShare"
          />
          <span v-if="!config.sub_mch_id" class="warn-text">（商家未配置子商户号，无法开启）</span>
          <span v-else class="muted-text">子商户号：{{ config.sub_mch_id }}，微信允许最大分账比例：{{ config.max_ratio_pct }}</span>
        </template>
      </el-alert>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">{{ receiverTypeText(row.receiver_type) }}</template>
        </el-table-column>
        <el-table-column prop="account" label="账号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="personal_name" label="真实姓名" min-width="100">
          <template #default="{ row }">{{ row.personal_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="分账比例" width="100">
          <template #default="{ row }">{{ row.default_ratio }}%</template>
        </el-table-column>
        <el-table-column label="微信关系" width="120">
          <template #default="{ row }">
            <el-tooltip v-if="row.wechat_bound === 0 && row.wechat_error" :content="row.wechat_error">
              <el-tag :type="boundType(row.wechat_bound)">{{ boundText(row.wechat_bound) }}</el-tag>
            </el-tooltip>
            <el-tag v-else :type="boundType(row.wechat_bound)">{{ boundText(row.wechat_bound) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'profit:receiver:manage'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'profit:receiver:manage'" link type="primary" @click="handleSync(row)">同步</el-button>
            <el-button v-permission="'profit:receiver:manage'" link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="接收方类型" required>
          <el-radio-group v-model="form.receiver_type">
            <el-radio :value="1">商户号</el-radio>
            <el-radio :value="2">个人微信</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="接收方名称" required>
          <el-input v-model="form.name" placeholder="显示名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="账号" required>
          <el-input v-model="form.account" :placeholder="form.receiver_type === 2 ? '个人 openid' : '商户号'" maxlength="64" />
        </el-form-item>
        <el-form-item v-if="form.receiver_type === 2" label="真实姓名">
          <el-input v-model="form.personal_name" placeholder="个人真实姓名（用于微信实名校验）" maxlength="64" />
        </el-form-item>
        <el-form-item label="分账比例(%)">
          <el-input-number v-model="form.default_ratio" :min="0" :max="100" :precision="2" :step="1" style="width: 200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" style="width: 200px" />
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
.config-alert {
  margin-bottom: 16px;
}
.muted-text {
  color: #909399;
  margin-left: 12px;
}
.warn-text {
  color: #e6a23c;
  margin-left: 8px;
}
</style>