import type { Pinia } from 'pinia'
import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { APP_TITLE } from '@/config/env'

// 会话恢复：页面加载内仅尝试一次，避免旧会话/旧缓存反复刷新
let sessionRestored = false

export function setupRouterGuards(router: Router, pinia: Pinia) {
  router.beforeEach(async (to) => {
    const authStore = useAuthStore(pinia)

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      return {
        path: '/login',
        query: { redirect: to.fullPath }
      }
    }

    if (to.path === '/login' && authStore.isAuthenticated) {
      return '/dashboard'
    }

    // 已登录但本地菜单/权限缓存为空（旧会话/旧缓存）：自动刷新一次权限
    if (authStore.isAuthenticated && !sessionRestored && authStore.menus.length === 0 && authStore.permissions.length === 0) {
      sessionRestored = true
      try {
        await authStore.fetchPermissions()
      } catch (_error) {
        // 刷新失败不阻断导航，保持登录态即可
      }
    }

    // 页面级权限控制：meta.permission 定义所需权限码（可为 string 或 string[]，数组时任一命中即放行）
    const required = to.meta.permission
    let hasRequired = false
    if (typeof required === 'string') {
      hasRequired = required ? authStore.hasPermission(required) : true
    } else if (Array.isArray(required)) {
      hasRequired = required.length === 0 || required.some((code) => authStore.hasPermission(code))
    } else {
      hasRequired = true
    }
    if (required && authStore.isAuthenticated && !hasRequired && to.path !== '/dashboard') {
      return '/dashboard'
    }

    return true
  })

  router.afterEach((to) => {
    const title = typeof to.meta.title === 'string' ? to.meta.title : APP_TITLE
    document.title = `${title} - ${APP_TITLE}`
  })
}
