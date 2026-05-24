<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createAnnouncement, getAnnouncementDetail, updateAnnouncement } from '@/api/sp'
import type { AnnouncementFormData } from '@/types/sp'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const saving = ref(false)

const announcementId = computed(() => Number(route.params.id || 0))
const isEditMode = computed(() => announcementId.value > 0)

const form = reactive<AnnouncementFormData>({
  title: '',
  content: '',
  status: 1
})

const statusOptions = [
  { label: '已发布', value: 1 },
  { label: '草稿/停用', value: 0 }
]

function fillForm(data: AnnouncementFormData) {
  form.title = data.title || ''
  form.content = data.content || ''
  form.status = Number(data.status ?? 1)
}

async function loadDetail() {
  if (!isEditMode.value) return
  loading.value = true
  try {
    const detail = await getAnnouncementDetail(announcementId.value)
    fillForm({
      title: detail.title,
      content: detail.content,
      status: detail.status
    })
  } finally {
    loading.value = false
  }
}

function buildPayload(): AnnouncementFormData {
  const payload: AnnouncementFormData = {
    title: form.title.trim(),
    content: form.content.trim(),
    status: Number(form.status)
  }

  if (!payload.title) {
    throw new Error('请输入公告标题')
  }
  if (!payload.content) {
    throw new Error('请输入公告内容')
  }

  return payload
}

async function handleSubmit() {
  let payload: AnnouncementFormData
  try {
    payload = buildPayload()
  } catch (error: any) {
    ElMessage.warning(error.message)
    return
  }

  saving.value = true
  try {
    if (isEditMode.value) {
      const detail = await updateAnnouncement(announcementId.value, payload)
      fillForm({
        title: detail.title,
        content: detail.content,
        status: detail.status
      })
      ElMessage.success('公告已保存')
      return
    }

    await createAnnouncement(payload)
    ElMessage.success('公告创建成功')
    await router.replace('/announcements')
  } finally {
    saving.value = false
  }
}

onMounted(loadDetail)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">{{ isEditMode ? '编辑公告' : '新增公告' }}</h1>
        <p class="page-subtitle">服务商可在此维护公告标题、内容和发布状态，供商家端读取展示。</p>
      </div>
      <el-button @click="router.push('/announcements')">返回公告列表</el-button>
    </div>

    <el-skeleton :rows="8" animated :loading="loading">
      <el-card class="page-card" shadow="never">
        <el-form label-position="top">
          <el-form-item label="公告标题" required>
            <el-input v-model="form.title" maxlength="128" show-word-limit placeholder="请输入公告标题" />
          </el-form-item>
          <el-form-item label="公告内容" required>
            <el-input
              v-model="form.content"
              type="textarea"
              :rows="10"
              maxlength="5000"
              show-word-limit
              placeholder="请输入公告内容"
            />
          </el-form-item>
          <el-form-item label="发布状态">
            <el-radio-group v-model="form.status">
              <el-radio v-for="item in statusOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </el-radio>
            </el-radio-group>
          </el-form-item>
          <el-alert
            :title="form.status === 1 ? '当前将作为已发布公告对商家展示' : '当前将作为草稿/停用状态保存'"
            :type="form.status === 1 ? 'success' : 'info'"
            :closable="false"
            show-icon
          />
          <div class="form-actions">
            <el-button @click="router.push('/announcements')">取消</el-button>
            <el-button type="primary" :loading="saving" @click="handleSubmit">
              {{ isEditMode ? '保存公告' : '创建公告' }}
            </el-button>
          </div>
        </el-form>
      </el-card>
    </el-skeleton>
  </div>
</template>

<style scoped>
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}
</style>
