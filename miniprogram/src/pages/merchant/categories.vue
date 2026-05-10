<template>
  <view class="category-container">
    <!-- 分类列表 -->
    <view class="category-list">
      <view
        v-for="category in categories"
        :key="category.id"
        class="category-item"
      >
        <view class="category-info" @click="editCategory(category)">
          <text class="category-name">{{ category.name }}</text>
          <text class="product-count">({{ category.product_count || 0 }}个商品)</text>
        </view>
        <view class="category-actions">
          <view class="action-icon edit" @click.stop="editCategory(category)">✏️</view>
          <view class="action-icon delete" @click.stop="deleteCategory(category)">🗑️</view>
        </view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-if="!loading && categories.length === 0" class="empty">
        <text class="empty-icon">📂</text>
        <text class="empty-text">暂无分类</text>
      </view>
    </view>

    <!-- 底部添加按钮 -->
    <view class="bottom-bar">
      <button class="btn-add" @click="showAddDialog">添加分类</button>
    </view>

    <!-- 添加/编辑分类弹窗 -->
    <view v-if="showDialog" class="dialog-mask" @click="closeDialog">
      <view class="dialog-content" @click.stop>
        <view class="dialog-title">{{ editingCategory ? '编辑分类' : '添加分类' }}</view>
        <input
          v-model="categoryName"
          class="dialog-input"
          placeholder="请输入分类名称"
          focus
        />
        <view class="dialog-actions">
          <button class="btn-cancel" @click="closeDialog">取消</button>
          <button class="btn-confirm" @click="saveCategory">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCategories, createCategory, updateCategory, deleteCategory } from '@api'
import type { Category } from '@types'

const categories = ref<Category[]>([])
const loading = ref(false)

const showDialog = ref(false)
const categoryName = ref('')
const editingCategory = ref<Category | null>(null)

onShow(() => {
  loadCategories()
})

async function loadCategories() {
  loading.value = true
  try {
    const res = await getCategories()
    categories.value = res
  } catch (error) {
    console.error('加载分类失败:', error)
    uni.showToast({ title: '加载分类失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function showAddDialog() {
  editingCategory.value = null
  categoryName.value = ''
  showDialog.value = true
}

function editCategory(category: Category) {
  editingCategory.value = category
  categoryName.value = category.name
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  categoryName.value = ''
  editingCategory.value = null
}

async function saveCategory() {
  if (!categoryName.value.trim()) {
    uni.showToast({ title: '请输入分类名称', icon: 'none' })
    return
  }

  try {
    if (editingCategory.value) {
      await updateCategory(editingCategory.value.id, { name: categoryName.value.trim() })
      const index = categories.value.findIndex(c => c.id === editingCategory.value!.id)
      if (index !== -1) {
        categories.value[index].name = categoryName.value.trim()
      }
      uni.showToast({ title: '修改成功', icon: 'success' })
    } else {
      const newCategory = await createCategory({ name: categoryName.value.trim() })
      categories.value.push(newCategory)
      uni.showToast({ title: '添加成功', icon: 'success' })
    }
    closeDialog()
  } catch (error: any) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

async function deleteCategory(category: Category) {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除分类"${category.name}"吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteCategory(category.id)
          categories.value = categories.value.filter(c => c.id !== category.id)
          uni.showToast({ title: '删除成功', icon: 'success' })
        } catch (error: any) {
          uni.showToast({ title: error.message || '删除失败', icon: 'none' })
        }
      }
    }
  })
}
</script>

<style scoped>
.category-container {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.category-list {
  padding: 24rpx;
}

.category-item {
  background: #ffffff;
  border-radius: 16rpx;
  padding: 28rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.category-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.category-name {
  font-size: 30rpx;
  color: #1a1a1a;
  font-weight: 500;
}

.product-count {
  font-size: 24rpx;
  color: #999999;
}

.category-actions {
  display: flex;
  gap: 16rpx;
}

.action-icon {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  font-size: 28rpx;
}

.action-icon.edit {
  background: #e6f0ff;
}

.action-icon.delete {
  background: #fff1f0;
}

.loading {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 100rpx 0;
}

.empty-icon {
  font-size: 120rpx;
  margin-bottom: 24rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999999;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16rpx 32rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #ffffff;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.btn-add {
  width: 100%;
  height: 88rpx;
  background: linear-gradient(135deg, #007AFF 0%, #0056CC 100%);
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 32rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.dialog-content {
  width: 600rpx;
  background: #ffffff;
  border-radius: 24rpx;
  padding: 48rpx;
}

.dialog-title {
  font-size: 34rpx;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 32rpx;
  text-align: center;
}

.dialog-input {
  height: 88rpx;
  background: #f8f9fa;
  border-radius: 16rpx;
  padding: 0 24rpx;
  font-size: 30rpx;
  margin-bottom: 32rpx;
}

.dialog-actions {
  display: flex;
  gap: 24rpx;
}

.btn-cancel, .btn-confirm {
  flex: 1;
  height: 88rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666666;
}

.btn-confirm {
  background: #007AFF;
  color: #ffffff;
}
</style>
