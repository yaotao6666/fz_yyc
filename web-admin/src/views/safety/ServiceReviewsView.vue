<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getServiceReviews, hideServiceReview, type ServiceReview } from '@/api/review'

/* ============ 字典 ============ */
const statusOptions = [
  { label: '正常展示', value: 1 },
  { label: '已隐藏', value: 0 }
]

const scoreOptions = [5, 4, 3, 2, 1].map((n) => ({ label: `${n} 星`, value: n }))

function statusTagType(s: number) {
  return s === 1 ? 'success' : 'info'
}

function formatDateTime(t: string | null | undefined) {
  if (!t) return '-'
  return t.replace('T', ' ').slice(0, 19)
}

/* ============ 列表 ============ */
const loading = ref(false)
const list = ref<ServiceReview[]>([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 10, status: '', score: '', keyword: '' })

async function loadList() {
  loading.value = true
  try {
    const res = await getServiceReviews({
      page: query.page,
      page_size: query.page_size,
      status: query.status || undefined,
      score: query.score || undefined,
      keyword: query.keyword || undefined
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch (_e) {
    // 拦截器提示
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  query.page = 1
  loadList()
}

function handlePageChange(p: number) {
  query.page = p
  loadList()
}

/* ============ 隐藏违规评价 ============ */
const detailVisible = ref(false)
const currentReview = ref<ServiceReview | null>(null)

function openDetail(row: ServiceReview) {
  currentReview.value = row
  detailVisible.value = true
}

async function handleHide(row: ServiceReview) {
  try {
    await ElMessageBox.confirm(
      `确认隐藏该条评价？隐藏后将从服务质量分统计中剔除。`,
      '隐藏评价',
      { confirmButtonText: '确认隐藏', cancelButtonText: '取消', type: 'warning' }
    )
    await hideServiceReview(row.id)
    ElMessage.success('已隐藏')
    loadList()
  } catch (_e) {
    // 取消或拦截器提示
  }
}

onMounted(loadList)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">服务评价</h1>
        <p class="page-subtitle">服务订单完结后的用户评价，可隐藏违规评价并同步重算服务质量分。</p>
      </div>
    </div>

    <div class="page-card">
      <div class="filter-bar">
        <el-select v-model="query.status" placeholder="评价状态" clearable style="width: 150px">
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
        </el-select>
        <el-select v-model="query.score" placeholder="总体评分" clearable style="width: 130px">
          <el-option v-for="o in scoreOptions" :key="o.value" :label="o.label" :value="String(o.value)" />
        </el-select>
        <el-input v-model="query.keyword" placeholder="服务人员姓名/手机号" clearable style="width: 200px" @keyup.enter="handleSearch" />
        <el-button @click="handleSearch">查询</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="服务人员" min-width="130">
          <template #default="{ row }">
            <template v-if="row.staff">
              <div>{{ row.staff.name }}</div>
              <div class="sub-info">{{ row.staff.phone }}</div>
            </template>
            <span v-else class="empty-tip">ID: {{ row.staff_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="关联订单" min-width="150">
          <template #default="{ row }">
            <template v-if="row.order">
              {{ row.order.order_no }}
            </template>
            <span v-else class="empty-tip">ID: {{ row.order_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="总体评分" width="110">
          <template #default="{ row }">
            <span class="score-text">★ {{ row.score }}</span>
          </template>
        </el-table-column>
        <el-table-column label="分项（态度/专业/准时）" width="170">
          <template #default="{ row }">
            <span class="sub-info">{{ row.attitude_score }} / {{ row.professional_score }} / {{ row.punctual_score }}</span>
          </template>
        </el-table-column>
        <el-table-column label="评价内容" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status_cn }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="评价时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 1"
              size="small"
              type="danger"
              v-permission="'service-reviews:update'"
              @click="handleHide(row)"
            >
              隐藏
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap" v-if="total > query.page_size">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="query.page_size"
          :current-page="query.page"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 评价详情弹窗 -->
    <el-dialog v-model="detailVisible" title="评价详情" width="560px">
      <template v-if="currentReview">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="服务人员">
            {{ currentReview.staff ? `${currentReview.staff.name}（${currentReview.staff.phone}）` : `ID: ${currentReview.staff_id}` }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentReview.status)">{{ currentReview.status_cn }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="关联订单" :span="2">
            {{ currentReview.order ? currentReview.order.order_no : `ID: ${currentReview.order_id}` }}
          </el-descriptions-item>
          <el-descriptions-item label="总体评分">★ {{ currentReview.score }}</el-descriptions-item>
          <el-descriptions-item label="分项评分">
            态度 {{ currentReview.attitude_score }} / 专业 {{ currentReview.professional_score }} / 准时 {{ currentReview.punctual_score }}
          </el-descriptions-item>
          <el-descriptions-item label="评价时间" :span="2">{{ formatDateTime(currentReview.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="评价内容" :span="2">
            <span style="white-space: pre-wrap">{{ currentReview.content || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="评价图片" :span="2" v-if="(currentReview.images || []).length">
            <div class="review-images">
              <el-image
                v-for="(img, idx) in currentReview.images"
                :key="idx"
                :src="img"
                :preview-src-list="currentReview.images"
                preview-teleported
                fit="cover"
                class="review-image-item"
              />
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.score-text {
  color: #ff9500;
  font-weight: 600;
}

.review-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.review-image-item {
  width: 84px;
  height: 84px;
  border-radius: 6px;
  overflow: hidden;
}
</style>