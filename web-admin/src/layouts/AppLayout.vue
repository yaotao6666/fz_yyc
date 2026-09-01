<script setup lang="ts">
import { computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { APP_TITLE } from '@/config/env'
import { useAuthStore } from '@/stores/auth'
import { ElMessageBox } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import TabsBar from '@/components/TabsBar.vue'
import type { SysMenu } from '@/types/sp'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 根据菜单图标名动态渲染 Element Plus 图标
function renderIcon(icon?: string) {
  if (!icon) return null
  const iconComponent = (ElementPlusIconsVue as Record<string, unknown>)[icon]
  return iconComponent ? () => h('el-icon', {}, [h(iconComponent as never)]) : null
}

// 递归生成菜单渲染数据：仅保留「菜单/目录」类型，剔除按钮(menu_type=2)
function buildMenuNodes(nodes: SysMenu[]): SysMenu[] {
  return nodes
    .filter((node) => node.menu_type === 1)
    .map((node) => {
      const children = buildMenuNodes(node.children || [])
      if (children.length > 0) {
        return { ...node, children }
      }
      if (node.path) {
        return node
      }
      return null
    })
    .filter((node): node is SysMenu => Boolean(node))
}

const menuNodes = computed<SysMenu[]>(() => buildMenuNodes(authStore.menus))

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
      <el-menu router :default-active="activeMenu" class="side-menu" unique-opened>
        <template v-for="menu in menuNodes" :key="menu.id">
          <!-- 含子菜单的目录：一级目录/二级子菜单 -->
          <el-sub-menu v-if="menu.children && menu.children.length" :index="String(menu.id)">
            <template #title>
              <el-icon v-if="renderIcon(menu.icon)">
                <component :is="renderIcon(menu.icon)" />
              </el-icon>
              <span>{{ menu.name }}</span>
            </template>
            <el-menu-item v-for="child in menu.children" :key="child.id" :index="child.path || ''">
              <el-icon v-if="child.icon && renderIcon(child.icon)">
                <component :is="renderIcon(child.icon)" />
              </el-icon>
              <span>{{ child.name }}</span>
            </el-menu-item>
          </el-sub-menu>
          <!-- 一级菜单 -->
          <el-menu-item v-else-if="menu.path" :index="menu.path">
            <el-icon v-if="renderIcon(menu.icon)">
              <component :is="renderIcon(menu.icon)" />
            </el-icon>
            <span>{{ menu.name }}</span>
          </el-menu-item>
        </template>
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
      <TabsBar />
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

/* 二级子菜单容器：继承透明背景，避免 Element Plus 默认白底导致白色文字不可见 */
:deep(.side-menu .el-menu--inline) {
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

:deep(.side-menu .el-sub-menu__title) {
  margin: 6px 12px;
  border-radius: 12px;
  color: rgba(255, 255, 255, 0.8);
}

:deep(.side-menu .el-sub-menu__title:hover) {
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
