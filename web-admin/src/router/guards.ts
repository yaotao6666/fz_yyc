import type { Pinia } from 'pinia'
import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { APP_TITLE } from '@/config/env'

export function setupRouterGuards(router: Router, pinia: Pinia) {
  router.beforeEach((to) => {
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

    return true
  })

  router.afterEach((to) => {
    const title = typeof to.meta.title === 'string' ? to.meta.title : APP_TITLE
    document.title = `${title} - ${APP_TITLE}`
  })
}
