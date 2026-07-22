import type { Pinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import LoginView from '@/views/login/LoginView.vue'
import DashboardView from '@/views/dashboard/DashboardView.vue'
import MerchantListView from '@/views/merchant/MerchantListView.vue'
import MerchantDetailView from '@/views/merchant/MerchantDetailView.vue'
import MerchantEditView from '@/views/merchant/MerchantEditView.vue'
import MerchantPickupPointsView from '@/views/merchant/MerchantPickupPointsView.vue'
import MerchantProductEditView from '@/views/merchant/MerchantProductEditView.vue'
import OrderListView from '@/views/order/OrderListView.vue'
import OrderDetailView from '@/views/order/OrderDetailView.vue'
import AnnouncementListView from '@/views/announcement/AnnouncementListView.vue'
import AnnouncementEditView from '@/views/announcement/AnnouncementEditView.vue'
import ProfitSharingHistoryView from '@/views/settlement/ProfitSharingHistoryView.vue'
import MerchantStatsView from '@/views/analytics/MerchantStatsView.vue'
import SpSettingsView from '@/views/settings/SpSettingsView.vue'
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
        path: '/merchants',
        name: 'merchants',
        component: MerchantListView,
        meta: { title: '商家列表', requiresAuth: true }
      },
      {
        path: '/merchants/new',
        name: 'merchant-create',
        component: MerchantEditView,
        meta: { title: '新增商家', requiresAuth: true }
      },
      {
        path: '/merchants/:id',
        name: 'merchant-detail',
        component: MerchantDetailView,
        meta: { title: '商家详情', requiresAuth: true }
      },
      {
        path: '/merchants/:id/pickup-points',
        name: 'merchant-pickup-points',
        component: MerchantPickupPointsView,
        meta: { title: '自提点管理', requiresAuth: true }
      },
      {
        path: '/merchants/:id/edit',
        name: 'merchant-edit',
        component: MerchantEditView,
        meta: { title: '编辑商家', requiresAuth: true }
      },
      {
        path: '/merchants/:id/products/new',
        name: 'merchant-product-create',
        component: MerchantProductEditView,
        meta: { title: '新增商品', requiresAuth: true }
      },
      {
        path: '/merchants/:id/products/:productId/edit',
        name: 'merchant-product-edit',
        component: MerchantProductEditView,
        meta: { title: '编辑商品', requiresAuth: true }
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
        path: '/profit-sharing',
        name: 'profit-sharing',
        component: ProfitSharingHistoryView,
        meta: { title: '分账历史', requiresAuth: true }
      },
      {
        path: '/announcements',
        name: 'announcements',
        component: AnnouncementListView,
        meta: { title: '公告管理', requiresAuth: true }
      },
      {
        path: '/announcements/new',
        name: 'announcement-create',
        component: AnnouncementEditView,
        meta: { title: '新增公告', requiresAuth: true }
      },
      {
        path: '/announcements/:id/edit',
        name: 'announcement-edit',
        component: AnnouncementEditView,
        meta: { title: '编辑公告', requiresAuth: true }
      },
      {
        path: '/analytics',
        name: 'analytics',
        component: MerchantStatsView,
        meta: { title: '数据分析', requiresAuth: true }
      },
      {
        path: '/settings',
        name: 'settings',
        component: SpSettingsView,
        meta: { title: '服务商设置', requiresAuth: true }
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
