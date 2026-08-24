<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox, type UploadRawFile, type UploadProps } from 'element-plus'
import {
  createMiniProgramBanner,
  deleteMiniProgramBanner,
  getMiniProgramBanners,
  updateMiniProgramBanner,
  updateMiniProgramBannerStatus
} from '@/api/sp'
import type { MiniProgramBanner } from '@/types/sp'
import { uploadSpImage } from '@/utils/qiniu'

const loading = ref(false)
const list = ref<MiniProgramBanner[]>([])

const linkTypeOptions = [
  { label: '无跳转', value: 'none' },
  { label: '商品详情', value: 'product' },
  { label: '分类页', value: 'category' },
  { label: '外部链接', value: 'url' }
]

function linkTypeText(t: string) {
  return linkTypeOptions.find(o => o.value === t)?.label || t
}

function statusText(s: number) {
  return s === 1 ? '启用' : '禁用'
}
function statusType(s: number) {
  return s === 1 ? 'success' : 'info'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getMiniProgramBanners()
    list.value = res.list || []
  } catch (_e) {
    // 拦截器提示
  } finally {
    loading.value = false
  }
}

/* ------ 新增/编辑弹窗 ------ */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)
const saving = ref(false)
const form = ref({
  title: '',
  image: '',
  link_type: 'none' as 'none' | 'product' | 'category' | 'url',
  link_value: '',
  sort: 0,
  status: 1
})

const linkPlaceholder = computed(() => {
  switch (form.value.link_type) {
    case 'product':
      return '请输入商品 ID'
    case 'category':
      return '跳转分类 Tab，无需填写'
    case 'url':
      return '请输入完整 URL，如 https://example.com'
    default:
      return '无需跳转'
  }
})

function openCreate() {
  dialogMode.value = 'create'
  editingId.value = null
  form.value = { title: '', image: '', link_type: 'none', link_value: '', sort: 0, status: 1 }
  dialogVisible.value = true
}

function openEdit(row: MiniProgramBanner) {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.value = {
    title: row.title || '',
    image: row.image,
    link_type: row.link_type || 'none',
    link_value: row.link_value || '',
    sort: row.sort ?? 0,
    status: row.status ?? 1
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.image) {
    ElMessage.warning('请上传轮播图图片')
    return
  }
  if (form.value.link_type === 'product' && !form.value.link_value) {
    ElMessage.warning('请输入商品 ID')
    return
  }
  if (form.value.link_type === 'url' && !form.value.link_value) {
    ElMessage.warning('请输入跳转链接')
    return
  }
  const payload = { ...form.value }
  if (form.value.link_type === 'none' || form.value.link_type === 'category') {
    payload.link_value = ''
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createMiniProgramBanner(payload as any)
      ElMessage.success('新增成功')
    } else if (editingId.value != null) {
      await updateMiniProgramBanner(editingId.value, payload as any)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 拦截器提示
  } finally {
    saving.value = false
  }
}

/* ------ 上传 ------ */
const imageUploading = ref(false)
const beforeBannerUpload: UploadProps['beforeUpload'] = (rawFile: UploadRawFile) => {
  if (!/^image\/(png|jpe?g|gif|webp)$/i.test(rawFile.type)) {
    ElMessage.warning('仅支持 JPG/PNG/GIF/WebP 图片')
    return false
  }
  if (rawFile.size / 1024 / 1024 > 5) {
    ElMessage.warning('图片大小不能超过 5MB')
    return false
  }
  return true
}
const customBannerUpload: UploadProps['httpRequest'] = async (options) => {
  const file = options.file as File
  imageUploading.value = true
  try {
    const { key } = await uploadSpImage(file)
    form.value.image = key
    ElMessage.success('上传成功')
    options.onSuccess?.({ key })
  } catch (e: any) {
    const msg = e?.message || '上传失败，请重试'
    ElMessage.error(msg)
    options.onError?.({ name: 'UploadError', message: msg, status: 0, method: 'POST', url: '' } as any)
  } finally {
    imageUploading.value = false
  }
}

/* ------ 操作 ------ */
async function handleStatusChange(row: MiniProgramBanner) {
  const targetStatus = row.status === 1 ? 0 : 1
  const action = targetStatus === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确认${action}该轮播图？`, action, { type: 'warning' })
    await updateMiniProgramBannerStatus(row.id, targetStatus)
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (_e) {
    // 取消或失败
  }
}

async function handleDelete(row: MiniProgramBanner) {
  try {
    await ElMessageBox.confirm('确认删除该轮播图？删除后不可恢复。', '删除', { type: 'warning' })
    await deleteMiniProgramBanner(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-shell">
    <div class="page-header">
      <div class="page-title-wrap">
        <h1 class="page-title">小程序轮播图</h1>
        <p class="page-subtitle">配置 C 端小程序首页金刚区展示的轮播图、跳转方式与排序。</p>
      </div>
      <el-button type="primary" v-permission="'banners:create'" @click="openCreate">+ 新增轮播图</el-button>
    </div>

    <div class="page-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column label="预览" width="200">
          <template #default="{ row }">
            <el-image
              v-if="row.image"
              :src="row.image"
              fit="cover"
              style="width: 160px; height: 90px; border-radius: 8px"
              :preview-src-list="[row.image]"
            />
            <span v-else class="empty-tip">无图</span>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.title">{{ row.title }}</span>
            <span v-else class="empty-tip">-</span>
          </template>
        </el-table-column>
        <el-table-column label="跳转方式" width="140">
          <template #default="{ row }">
            <span>{{ linkTypeText(row.link_type) }}</span>
            <span v-if="row.link_type === 'product' && row.link_value" class="sub-info">(ID: {{ row.link_value }})</span>
            <span v-else-if="row.link_type === 'url' && row.link_value" class="sub-info wrap-url">{{ row.link_value }}</span>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="90">
          <template #default="{ row }">{{ row.sort ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" v-permission="'banners:update'" @click="openEdit(row)">编辑</el-button>
            <el-button
              size="small"
              :type="row.status === 1 ? 'warning' : 'primary'"
              v-permission="'banners:status'"
              @click="handleStatusChange(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button size="small" type="danger" v-permission="'banners:delete'" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!loading && list.length === 0" class="empty-state">
        <p>还未配置任何轮播图，点击右上角「新增轮播图」开始配置。</p>
      </div>
    </div>

    <!-- 新增/编辑 弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增轮播图' : '编辑轮播图'" width="640px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="仅后台识别（可选）" maxlength="50" show-word-limit />
        </el-form-item>

        <el-form-item label="轮播图" required>
          <el-upload
            :show-file-list="false"
            accept="image/*"
            :before-upload="beforeBannerUpload"
            :http-request="customBannerUpload"
          >
            <div v-if="form.image" class="banner-upload-preview">
              <el-image :src="form.image" fit="cover" class="preview-image" />
              <div class="preview-mask">点击更换</div>
            </div>
            <div v-else class="banner-upload-placeholder">
              <el-icon class="upload-icon"><Plus /></el-icon>
              <p>上传轮播图（建议 750×360，≤5MB）</p>
            </div>
          </el-upload>
          <p v-if="imageUploading" class="upload-tip">上传中…</p>
        </el-form-item>

        <el-form-item label="跳转方式" required>
          <el-radio-group v-model="form.link_type">
            <el-radio v-for="o in linkTypeOptions" :key="o.value" :value="o.value">
              {{ o.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.link_type === 'product'" label="商品 ID" required>
          <el-input v-model="form.link_value" placeholder="商品 ID，点击轮播跳转到商品详情" />
        </el-form-item>
        <el-form-item v-else-if="form.link_type === 'url'" label="跳转链接" required>
          <el-input v-model="form.link_value" placeholder="完整 URL，如 https://example.com" />
        </el-form-item>
        <el-form-item v-else label="跳转值">
          <el-input v-model="form.link_value" disabled :placeholder="linkPlaceholder" />
        </el-form-item>

        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
          <span class="form-tip">值越小越靠前，留空默认为 0。</span>
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ dialogMode === 'create' ? '确认新增' : '保存修改' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts">
import { Plus } from '@element-plus/icons-vue'
export default { components: { Plus } }
</script>

<style scoped>
.page-shell { display: flex; flex-direction: column; gap: 16px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-end; }
.page-title { margin: 0 0 4px; font-size: 22px; }
.page-subtitle { margin: 0; font-size: 13px; color: #6b7280; }
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.empty-tip { color: #9ca3af; }
.sub-info { color: #6b7280; font-size: 12px; margin-left: 4px; }
.sub-info.wrap-url { display: block; margin-left: 0; word-break: break-all; }
.empty-state { padding: 60px 20px; text-align: center; color: #9ca3af; }
.empty-state p { margin: 0; }

.banner-upload-preview,
.banner-upload-placeholder {
  width: 480px;
  height: 240px;
  border-radius: 12px;
  border: 2px dashed #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  transition: border-color .15s;
}
.banner-upload-preview:hover { border-color: #007AFF; }
.banner-upload-preview .preview-image { width: 100%; height: 100%; }
.banner-upload-preview .preview-mask {
  position: absolute; inset: 0;
  background: rgba(0,0,0,0.45); color: #fff;
  display: flex; align-items: center; justify-content: center;
  opacity: 0; transition: opacity .15s;
}
.banner-upload-preview:hover .preview-mask { opacity: 1; }

.banner-upload-placeholder { flex-direction: column; background: #fafafa; }
.banner-upload-placeholder:hover { border-color: #007AFF; }
.upload-icon { font-size: 40px; color: #9ca3af; margin-bottom: 8px; }
.banner-upload-placeholder p { color: #9ca3af; font-size: 13px; margin: 0; }

.upload-tip { margin: 6px 0 0; font-size: 12px; color: #6b7280; }
.form-tip { margin-left: 12px; font-size: 12px; color: #9ca3af; }
</style>
