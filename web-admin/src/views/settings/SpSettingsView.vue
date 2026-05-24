<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { changeSpPassword, getSpSettings, updateSpSettings } from '@/api/sp'
import type { SpSettings } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const saving = ref(false)
const passwordSaving = ref(false)

const info = ref<SpSettings>({
  name: '',
  sp_name: '',
  contact_phone: '',
  contact_email: '',
  created_at: ''
})

const settingForm = reactive({
  contact_name: '',
  contact_phone: ''
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

async function loadSettings() {
  loading.value = true
  try {
    const result = await getSpSettings()
    info.value = result
    settingForm.contact_name = result.sp_name || ''
    settingForm.contact_phone = result.contact_phone || ''
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await updateSpSettings({
      contact_name: settingForm.contact_name.trim(),
      contact_phone: settingForm.contact_phone.trim(),
    })
    ElMessage.success('设置已保存')
    await loadSettings()
  } finally {
    saving.value = false
  }
}

async function savePassword() {
  if (!passwordForm.old_password.trim()) {
    ElMessage.warning('请输入旧密码')
    return
  }
  if (passwordForm.new_password.trim().length < 6) {
    ElMessage.warning('新密码至少 6 位')
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    ElMessage.warning('两次密码不一致')
    return
  }

  passwordSaving.value = true
  try {
    await changeSpPassword({
      old_password: passwordForm.old_password.trim(),
      new_password: passwordForm.new_password.trim()
    })
    ElMessage.success('密码修改成功')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
  } finally {
    passwordSaving.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">服务商设置</h1>
        <p class="page-subtitle">维护服务商联系方式并调整登录密码。</p>
      </div>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <div class="section-grid">
        <el-card class="page-card" shadow="never">
          <template #header>
            <span>服务商信息</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="服务商名称">{{ info.name || '未设置' }}</el-descriptions-item>
            <el-descriptions-item label="服务商姓名">{{ info.sp_name || '未设置' }}</el-descriptions-item>
            <el-descriptions-item label="联系方式">{{ info.contact_phone || '未设置' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDateTime(info.created_at) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card class="page-card" shadow="never">
          <template #header>
            <span>联系信息维护</span>
          </template>
          <el-form label-position="top">
            <el-form-item label="服务商姓名">
              <el-input v-model="settingForm.contact_name" placeholder="请输入服务商姓名" />
            </el-form-item>
            <el-form-item label="联系方式">
              <el-input v-model="settingForm.contact_phone" placeholder="请输入联系方式" />
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="saveSettings">保存设置</el-button>
          </el-form>
        </el-card>
      </div>

      <el-card class="page-card" shadow="never" style="margin-top: 20px;">
        <template #header>
          <span>修改密码</span>
        </template>
        <div class="detail-grid">
          <el-form label-position="top">
            <el-form-item label="旧密码">
              <el-input v-model="passwordForm.old_password" show-password type="password" placeholder="请输入旧密码" />
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="passwordForm.new_password" show-password type="password" placeholder="请输入新密码" />
            </el-form-item>
            <el-form-item label="确认新密码">
              <el-input v-model="passwordForm.confirm_password" show-password type="password" placeholder="请再次输入新密码" />
            </el-form-item>
            <el-button type="primary" :loading="passwordSaving" @click="savePassword">确认修改</el-button>
          </el-form>
        </div>
      </el-card>
    </el-skeleton>
  </div>
</template>
