import type { Pinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import LoginView from '@/views/login/LoginView.vue'
import DashboardView from '@/views/dashboard/DashboardView.vue'
import ProductsView from '@/views/product/ProductsView.vue'
import ServiceProductsView from '@/views/product/ServiceProductsView.vue'
import CategoriesView from '@/views/product/CategoriesView.vue'
import OrderListView from '@/views/order/OrderListView.vue'
import OrderDetailView from '@/views/order/OrderDetailView.vue'
import RentalDueListView from '@/views/order/RentalDueListView.vue'
import ServiceStaffListView from '@/views/staff/ServiceStaffListView.vue'
import StaffAuditListView from '@/views/staff/StaffAuditListView.vue'
import MerchantProfileView from '@/views/merchant/MerchantProfileView.vue'
import MerchantStatsView from '@/views/analytics/MerchantStatsView.vue'
import MenuManagementView from '@/views/system/MenuManagementView.vue'
import RoleManagementView from '@/views/system/RoleManagementView.vue'
import DepartmentManagementView from '@/views/system/DepartmentManagementView.vue'
import StaffManagementView from '@/views/system/StaffManagementView.vue'
import HealthRecordsView from '@/views/health/HealthRecordsView.vue'
import HealthAssessmentFormsView from '@/views/health/HealthAssessmentFormsView.vue'
import HealthAssessmentsView from '@/views/health/HealthAssessmentsView.vue'
import FittingRecommendationsView from '@/views/health/FittingRecommendationsView.vue'
import MiniProgramBannersView from '@/views/miniprogram/MiniProgramBannersView.vue'
import CouponTemplatesView from '@/views/marketing/CouponTemplatesView.vue'
// 服务过程安全（PRD V2.0 阶段三）
import AlertEventsView from '@/views/safety/AlertEventsView.vue'
import AgreementsView from '@/views/system/AgreementsView.vue'
// 服务评价管理（PRD V2.0 阶段四）
import ServiceReviewsView from '@/views/safety/ServiceReviewsView.vue'
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
        meta: { title: '工作台', requiresAuth: true, permission: 'dashboard:view' }
      },
      {
        path: '/orders',
        name: 'orders',
        component: OrderListView,
        meta: { title: '订单管理', requiresAuth: true, permission: 'orders:view' }
      },
      {
        path: '/orders/rental-due',
        name: 'rental-due',
        component: RentalDueListView,
        meta: { title: '租赁到期提醒', requiresAuth: true, permission: 'orderrental:view' }
      },
      {
        path: '/orders/:id',
        name: 'order-detail',
        component: OrderDetailView,
        meta: { title: '订单详情', requiresAuth: true, permission: 'orders:view' }
      },
      {
        path: '/products',
        name: 'products',
        component: ProductsView,
        meta: { title: '商品管理', requiresAuth: true, permission: 'products:view' }
      },
      {
        path: '/service-products',
        name: 'service-products',
        component: ServiceProductsView,
        meta: { title: '服务管理', requiresAuth: true, permission: 'service-products:view' }
      },
      {
        path: '/categories',
        name: 'categories',
        component: CategoriesView,
        meta: { title: '分类管理', requiresAuth: true, permission: 'categories:view' }
      },
      // 优惠券管理（PRD V2.0 阶段二）
      {
        path: '/coupon-templates',
        name: 'coupon-templates',
        component: CouponTemplatesView,
        meta: { title: '优惠券管理', requiresAuth: true, permission: 'coupon-templates:view' }
      },
      // 服务过程安全（PRD V2.0 阶段三）
      {
        path: '/alert-events',
        name: 'alert-events',
        component: AlertEventsView,
        meta: { title: '预警中心', requiresAuth: true, permission: 'alert-events:view' }
      },
      {
        path: '/system/agreements',
        name: 'system-agreements',
        component: AgreementsView,
        meta: { title: '协议管理', requiresAuth: true, permission: 'agreements:view' }
      },
      // 服务评价管理（PRD V2.0 阶段四）
      {
        path: '/service-reviews',
        name: 'service-reviews',
        component: ServiceReviewsView,
        meta: { title: '服务评价', requiresAuth: true, permission: 'service-reviews:view' }
      },
      {
        path: '/staff',
        name: 'service-staff',
        component: ServiceStaffListView,
        meta: { title: '服务人员', requiresAuth: true, permission: 'staff:view' }
      },
      {
        path: '/staff/audits',
        name: 'staff-audits',
        component: StaffAuditListView,
        meta: { title: '服务人员审核', requiresAuth: true, permission: 'staffaudit:view' }
      },
      {
        path: '/profile',
        name: 'merchant-profile',
        component: MerchantProfileView,
        meta: { title: '商家资料', requiresAuth: true, permission: 'profile:view' }
      },
      {
        path: '/analytics',
        name: 'analytics',
        component: MerchantStatsView,
        meta: { title: '数据分析', requiresAuth: true, permission: 'analytics:view' }
      },
      // 小程序配置
      {
        path: '/miniprogram-banners',
        name: 'miniprogram-banners',
        component: MiniProgramBannersView,
        meta: { title: '小程序轮播图', requiresAuth: true, permission: 'banners:view' }
      },
      // 系统管理（RBAC）
      {
        path: '/system/menus',
        name: 'system-menus',
        component: MenuManagementView,
        meta: { title: '菜单管理', requiresAuth: true, permission: 'system:menu:view' }
      },
      {
        path: '/system/roles',
        name: 'system-roles',
        component: RoleManagementView,
        meta: { title: '角色管理', requiresAuth: true, permission: 'system:role:view' }
      },
      {
        path: '/system/departments',
        name: 'system-departments',
        component: DepartmentManagementView,
        meta: { title: '部门管理', requiresAuth: true, permission: 'system:dept:view' }
      },
      {
        path: '/system/staff',
        name: 'system-staff',
        component: StaffManagementView,
        meta: { title: '员工管理', requiresAuth: true, permission: 'system:staff:view' }
      },
      // 健康服务（基层健康服务闭环）
      {
        path: '/health/records',
        name: 'health-records',
        component: HealthRecordsView,
        meta: { title: '健康档案管理', requiresAuth: true, permission: 'health:view' }
      },
      {
        path: '/health/assessment-forms',
        name: 'health-assessment-forms',
        component: HealthAssessmentFormsView,
        meta: { title: '评估量表管理', requiresAuth: true, permission: 'assessment:view' }
      },
      {
        path: '/health/assessments',
        name: 'health-assessments',
        component: HealthAssessmentsView,
        meta: { title: '评估记录', requiresAuth: true, permission: 'assessment:view' }
      },
      {
        path: '/health/fitting',
        name: 'health-fitting',
        component: FittingRecommendationsView,
        meta: { title: '适配建议管理', requiresAuth: true, permission: 'fitting:view' }
      },
      {
        path: '/health/education',
        name: 'health-education',
        component: () => import('@/views/health/EducationArticlesView.vue'),
        meta: { title: '健康宣教管理', requiresAuth: true, permission: 'education:view' }
      },
      {
        path: '/health/education-categories',
        name: 'health-education-categories',
        component: () => import('@/views/health/HealthEducationCategoriesView.vue'),
        meta: { title: '宣教分类管理', requiresAuth: true, permission: 'education-categories:view' }
      },
      // 分账管理（服务商分账）
      {
        path: '/settlement/receivers',
        name: 'profit-receivers',
        component: () => import('@/views/settlement/ProfitSharingReceiversView.vue'),
        meta: { title: '分账接收方', requiresAuth: true, permission: 'profit:view' }
      },
      {
        path: '/settlement/records',
        name: 'profit-records',
        component: () => import('@/views/settlement/ProfitSharingRecordsView.vue'),
        meta: { title: '分账记录', requiresAuth: true, permission: 'profit:view' }
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
