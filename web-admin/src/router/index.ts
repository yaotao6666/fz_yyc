import type { Pinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import LoginView from '@/views/login/LoginView.vue'
import DashboardView from '@/views/dashboard/DashboardView.vue'
import ProductsView from '@/views/product/ProductsView.vue'
import CategoriesView from '@/views/product/CategoriesView.vue'
import OrderListView from '@/views/order/OrderListView.vue'
import OrderDetailView from '@/views/order/OrderDetailView.vue'
import ServiceStaffListView from '@/views/staff/ServiceStaffListView.vue'
import MerchantProfileView from '@/views/merchant/MerchantProfileView.vue'
import MerchantStatsView from '@/views/analytics/MerchantStatsView.vue'
import { setupRouterGuards } from './guards'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: AppLayout,
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: '/dashboard',
        name: 'dashboard',
        component: DashboardView,
        meta: { title: '工作台', requiresAuth: true }
      },
      {
        path: '/orders',
        name: 'orders',
        component: OrderListView,
        meta: { title: '订单管理', requiresAuth: true }
      },
      {
        path: '/orders/:id',
        name: 'order-detail',
        component: OrderDetailView,
        meta: { title: '订单详情', requiresAuth: true }
      },
      {
        path: '/products',
        name: 'products',
        component: ProductsView,
        meta: { title: '商品管理', requiresAuth: true }
      },
      {
        path: '/categories',
        name: 'categories',
        component: CategoriesView,
        meta: { title: '分类管理', requiresAuth: true }
      },
      {
        path: '/staff',
        name: 'service-staff',
        component: ServiceStaffListView,
        meta: { title: '服务人员', requiresAuth: true }
      },
      {
        path: '/profile',
        name: 'merchant-profile',
        component: MerchantProfileView,
        meta: { title: '商家资料', requiresAuth: true }
      },
      {
        path: '/analytics',
        name: 'analytics',
        component: MerchantStatsView,
        meta: { title: '数据分析', requiresAuth: true }
      }
    ]
  }
]

export function setupRouter(pinia: Pinia) {
  const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
    scrollBehavior: () => ({ top: 0 })
  })

  setupRouterGuards(router, pinia)
  return router
}
