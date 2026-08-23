import type { Pinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import LoginView from '@/views/login/LoginView.vue'
import DashboardView from '@/views/dashboard/DashboardView.vue'
import ProductsView from '@/views/product/ProductsView.vue'
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
import CarePlansView from '@/views/health/CarePlansView.vue'
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
        path: '/categories',
        name: 'categories',
        component: CategoriesView,
        meta: { title: '分类管理', requiresAuth: true, permission: 'categories:view' }
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
        path: '/health/care-plans',
        name: 'health-care-plans',
        component: CarePlansView,
        meta: { title: '照护计划管理', requiresAuth: true, permission: 'care:view' }
      },
      {
        path: '/health/follow-ups',
        name: 'health-follow-ups',
        component: () => import('@/views/health/FollowUpTasksView.vue'),
        meta: { title: '随访任务管理', requiresAuth: true, permission: 'followup:view' }
      },
      {
        path: '/health/monitoring',
        name: 'health-monitoring',
        component: () => import('@/views/health/HealthMonitoringView.vue'),
        meta: { title: '生命体征监测', requiresAuth: true, permission: 'monitor:view' }
      },
      {
        path: '/health/education',
        name: 'health-education',
        component: () => import('@/views/health/EducationArticlesView.vue'),
        meta: { title: '健康宣教管理', requiresAuth: true, permission: 'education:view' }
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
