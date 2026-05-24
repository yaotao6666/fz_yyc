<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteAnnouncement, getAnnouncements } from '@/api/sp'
import type { AnnouncementItem } from '@/types/sp'
import { formatDateTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const deletingId = ref<number | null>(null)
const announcements = ref<AnnouncementItem[]>([])
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

function getStatusText(status: number) {
  return Number(status) === 1 ? '已发布' : '草稿/已停用'
}

function getStatusTagType(status: number) {
  return Number(status) === 1 ? 'success' : 'info'
}

async function loadAnnouncements() {
  loading.value = true
  try {
    const response = await getAnnouncements({
      page: pagination.page,
      page_size: pagination.page_size
    })
    announcements.value = response.list
    pagination.total = response.pagination.total
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadAnnouncements()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadAnnouncements()
}

async function handleDelete(row: AnnouncementItem) {
  try {
    await ElMessageBox.confirm(
      `删除后该公告将变为停用状态，是否继续删除“${row.title}”？`,
      '删除公告',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }

  deletingId.value = row.id
  try {
    await deleteAnnouncement(row.id)
    ElMessage.success('删除成功，公告已停用')
    await loadAnnouncements()
  } finally {
    deletingId.value = null
  }
}

onMounted(loadAnnouncements)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">公告管理</h1>
        <p class="page-subtitle">统一维护服务商面向商家展示的系统公告内容与发布状态。</p>
      </div>
      <el-button type="primary" @click="router.push('/announcements/new')">新增公告</el-button>
    </div>

    <el-card class="page-card" shadow="never">
      <el-table :data="announcements" v-loading="loading" style="width: 100%;">
        <el-table-column prop="title" label="公告标题" min-width="220" />
        <el-table-column label="公告内容" min-width="320">
          <template #default="scope">
            <div class="content-preview">{{ scope.row.content || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="getStatusTagType(scope.row.status)">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="180">
          <template #default="scope">
            {{ formatDateTime(scope.row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-space>
              <el-button link type="primary" @click="router.push(`/announcements/${scope.row.id}/edit`)">编辑</el-button>
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

      <div style="display: flex; justify-content: flex-end; margin-top: 20px;">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.page_size"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.content-preview {
  display: -webkit-box;
  overflow: hidden;
  color: #4b5563;
  line-height: 1.6;
  text-overflow: ellipsis;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
