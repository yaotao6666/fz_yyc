<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { APP_TITLE } from '@/config/env'
import { useAuthStore } from '@/stores/auth'
import { ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const menuItems = [
  { index: '/dashboard', label: '工作台' },
  { index: '/orders', label: '订单管理' },
  { index: '/products', label: '商品管理' },
  { index: '/categories', label: '分类管理' },
  { index: '/staff', label: '服务人员' },
  { index: '/analytics', label: '数据分析' },
  { index: '/profile', label: '商家资料' }
]

const activeMenu = computed(() => {
  if (route.path.startsWith('/orders')) {
    return '/orders'
  }
  return route.path
})

async function handleLogout() {
  try {
    await ElMessageBox.confirm('退出后需要重新登录，是否继续？', '退出登录', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }

  await authStore.logout()
  await router.replace('/login')
}
</script>

<template>
  <el-container class="layout-shell">
    <el-aside class="layout-aside" width="240px">
      <div class="brand-block">
        <div class="brand-title">{{ APP_TITLE }}</div>
        <div class="brand-subtitle">商家后台管理</div>
      </div>
      <el-menu router :default-active="activeMenu" class="side-menu">
        <el-menu-item v-for="item in menuItems" :key="item.index" :index="item.index">
          {{ item.label }}
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div>
          <div class="header-page-title">{{ route.meta.title || '商家后台' }}</div>
          <div class="header-page-subtitle">{{ authStore.staffName }}</div>
        </div>
        <div class="header-actions">
          <span class="operator-name">{{ authStore.staffName }}</span>
          <el-button type="primary" plain @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-shell {
  min-height: 100vh;
}

.layout-aside {
  background: linear-gradient(180deg, #0f172a 0%, #16213c 100%);
  color: #fff;
  border-right: none;
}

.brand-block {
  padding: 28px 24px 20px;
}

.brand-title {
  font-size: 22px;
  font-weight: 700;
}

.brand-subtitle {
  margin-top: 8px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.72);
}

:deep(.side-menu) {
  border-right: none;
  background: transparent;
}

:deep(.side-menu .el-menu-item) {
  margin: 6px 12px;
  border-radius: 12px;
  color: rgba(255, 255, 255, 0.8);
}

:deep(.side-menu .el-menu-item.is-active) {
  background: rgba(59, 130, 246, 0.16);
  color: #ffffff;
}

:deep(.side-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.08);
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  height: 76px;
  padding: 0 24px;
  background: rgba(255, 255, 255, 0.92);
  border-bottom: 1px solid #e5e7eb;
}

.header-page-title {
  font-size: 20px;
  font-weight: 700;
  color: #111827;
}

.header-page-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: #6b7280;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.operator-name {
  font-size: 14px;
  color: #4b5563;
}

.layout-main {
  background: #f3f5f9;
}
</style>
