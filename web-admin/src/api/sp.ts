import request, { unwrapApiResponse } from '@/utils/request'
import type {
  AssessmentForm,
  AssessmentFormQuestion,
  AssignRoleMenusRequest,
  CarePlan,
  CarePlanItem,
  CarePlanListResponse,
  CareVisit,
  CareVisitListResponse,
  DashboardData,
  EducationArticle,
  EducationArticleListResponse,
  FittingRecommendation,
  FittingRecommendationListResponse,
  FittingRecommendedProduct,
  FollowUpTask,
  FollowUpTaskListResponse,
  FollowUpTaskResult,
  HealthAssessmentListResponse,
  HealthMonitoring,
  HealthMonitoringListResponse,
  HealthRecord,
  HealthRecordListResponse,
  MerchantCategory,
  MerchantCategoryPayload,
  MerchantCategorySortItem,
  MerchantDetail,
  MerchantLoginRequest,
  MerchantLoginResponse,
  MerchantProduct,
  MerchantProductListResponse,
  MerchantProductQuery,
  MerchantProductSpecsPayload,
  MerchantProductSpecsResponse,
  MerchantProductUpsertPayload,
  MerchantPaymentConfigFormData,
  MerchantStaffInfo,
  MerchantStaffListResponse,
  OrderAnalyticsData,
  ProfitSharingConfig,
  ProfitSharingReceiver,
  ProfitSharingReceiverPayload,
  ProfitSharingRecord,
  ProfitSharingRecordListResponse,
  RbacPermissions,
  ServiceStaffListResponse,
  SpOrder,
  SpOrderListResponse,
  SysDepartment,
  SysMenu,
  SysRole,
  UpdateSpMerchantFormData,
  UploadTokenResponse,
} from '@/types/sp'

/* ============ 认证 ============ */

export function merchantLogin(data: MerchantLoginRequest) {
  return request.post('/api/v1/auth/merchant/login', data).then(unwrapApiResponse<MerchantLoginResponse>)
}

export function merchantLogout() {
  // 单商户模式无后端登出接口，直接清理本地登录态
  return Promise.resolve<{ message: string }>({ message: '已退出' })
}

export function getMerchantProfile() {
  return request.get('/api/v1/merchant/profile').then(unwrapApiResponse<{ staff: MerchantStaffInfo; merchant: MerchantDetail }>)
}

export function updateMerchantProfile(data: UpdateSpMerchantFormData) {
  return request.put('/api/v1/merchant/profile', data).then(unwrapApiResponse<MerchantDetail>)
}

export function updatePaymentConfig(data: MerchantPaymentConfigFormData) {
  return request.put('/api/v1/merchant/payment-config', data).then(unwrapApiResponse<MerchantDetail>)
}

export function changePassword(data: { old_password: string; new_password: string }) {
  return request.post('/api/v1/merchant/account/change-password', data).then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantQRCode() {
  return request.get('/api/v1/merchant/qrcode').then(
    unwrapApiResponse<{ qrcode_url: string; page?: string; scene?: string }>
  )
}

/* ============ 商品分类 ============ */

export function getMerchantCategories() {
  return request.get('/api/v1/merchant/categories').then(unwrapApiResponse<MerchantCategory[]>)
}

export function createMerchantCategory(data: MerchantCategoryPayload) {
  return request.post('/api/v1/merchant/categories', data).then(unwrapApiResponse<MerchantCategory>)
}

export function updateMerchantCategory(categoryId: number, data: MerchantCategoryPayload) {
  return request.put(`/api/v1/merchant/categories/${categoryId}`, data).then(unwrapApiResponse<MerchantCategory>)
}

export function deleteMerchantCategory(categoryId: number) {
  return request.delete(`/api/v1/merchant/categories/${categoryId}`).then(unwrapApiResponse<{ message: string }>)
}

export function sortMerchantCategories(categories: MerchantCategorySortItem[]) {
  return request.post('/api/v1/merchant/categories/sort', { categories }).then(unwrapApiResponse<{ message: string }>)
}

/* ============ 商品管理 ============ */

export function getMerchantProducts(params?: MerchantProductQuery) {
  return request.get('/api/v1/merchant/products', { params }).then(unwrapApiResponse<MerchantProductListResponse>)
}

export function getMerchantProduct(productId: number) {
  return request.get(`/api/v1/merchant/products/${productId}`).then(unwrapApiResponse<MerchantProduct>)
}

export function createMerchantProduct(data: MerchantProductUpsertPayload) {
  return request.post('/api/v1/merchant/products', data).then(unwrapApiResponse<MerchantProduct>)
}

export function updateMerchantProduct(productId: number, data: MerchantProductUpsertPayload) {
  return request.put(`/api/v1/merchant/products/${productId}`, data).then(unwrapApiResponse<MerchantProduct>)
}

export function merchantProductOnSale(productId: number) {
  return request.post(`/api/v1/merchant/products/${productId}/on-sale`, {}).then(unwrapApiResponse<{ message: string }>)
}

export function merchantProductOffSale(productId: number) {
  return request.post(`/api/v1/merchant/products/${productId}/off-sale`, {}).then(unwrapApiResponse<{ message: string }>)
}

export function batchUpdateMerchantProductStatus(productIds: number[], status: number) {
  return request.post('/api/v1/merchant/products/batch-status', { product_ids: productIds, status }).then(unwrapApiResponse<{ message: string }>)
}

export function deleteMerchantProduct(productId: number) {
  return request.delete(`/api/v1/merchant/products/${productId}`).then(unwrapApiResponse<{ message: string }>)
}

export function updateMerchantProductStock(productId: number, stock: number) {
  return request.put(`/api/v1/merchant/products/${productId}/stock`, { stock }).then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantProductSpecs(productId: number) {
  return request.get(`/api/v1/merchant/products/${productId}/specs`).then(unwrapApiResponse<MerchantProductSpecsResponse>)
}

export function updateMerchantProductSpecs(productId: number, data: MerchantProductSpecsPayload) {
  return request.put(`/api/v1/merchant/products/${productId}/specs`, data).then(unwrapApiResponse<{ message: string }>)
}

/* ============ 订单管理 ============ */

export function getOrders(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/orders', { params }).then(unwrapApiResponse<SpOrderListResponse>)
}

export function getOrderDetail(orderId: number) {
  return request.get(`/api/v1/merchant/orders/${orderId}`).then(unwrapApiResponse<SpOrder>)
}

export function completeOrder(orderId: number, verifyCode: string) {
  return request.post(`/api/v1/merchant/orders/${orderId}/complete`, { verify_code: verifyCode }).then(unwrapApiResponse<SpOrder>)
}

export function quickCompleteOrder(orderId: number) {
  return request.post('/api/v1/merchant/orders/quick-complete', { order_id: orderId }).then(unwrapApiResponse<SpOrder>)
}

export function refundOrder(orderId: number, data: { refund_amount?: number; reason?: string }) {
  return request.post(`/api/v1/merchant/orders/${orderId}/refund`, data).then(unwrapApiResponse<unknown>)
}

export function returnRentalOrder(orderId: number, data: { deduct_amount?: number; remark?: string }) {
  return request.post(`/api/v1/merchant/orders/${orderId}/return`, data).then(unwrapApiResponse<SpOrder>)
}

/* ============ 服务人员管理 ============ */

export function getServiceStaffList(params?: { status?: number; keyword?: string; page?: number; page_size?: number }) {
  return request.get('/api/v1/merchant/service-staff', { params }).then(unwrapApiResponse<ServiceStaffListResponse>)
}

export function updateServiceStaffStatus(id: number, status: number) {
  return request.put(`/api/v1/merchant/service-staff/${id}/status`, { status }).then(unwrapApiResponse<{ id: number; status: number }>)
}

export function resetServiceStaffPassword(id: number, new_password: string) {
  return request.post(`/api/v1/merchant/service-staff/${id}/reset-password`, { new_password }).then(unwrapApiResponse<{ id: number }>)
}

export function deleteServiceStaff(id: number) {
  return request.delete(`/api/v1/merchant/service-staff/${id}`).then(unwrapApiResponse<{ id: number }>)
}

/* ============ 数据分析 ============ */

export function getDashboard() {
  return request.get('/api/v1/merchant/orders/statistics').then(unwrapApiResponse<DashboardData>)
}

export function getAnalyticsOverview(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/analytics/overview', { params }).then(unwrapApiResponse<Record<string, unknown>>)
}

export function getSalesTrend(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/analytics/sales-trend', { params }).then(unwrapApiResponse<Array<Record<string, unknown>>>)
}

export function getProductRanking(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/analytics/product-ranking', { params }).then(unwrapApiResponse<Array<Record<string, unknown>>>)
}

export function getOrderAnalytics(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/analytics/sales-trend', { params }).then(unwrapApiResponse<OrderAnalyticsData>)
}

/* ============ 上传 ============ */

export function getUploadToken() {
  return request.get('/api/v1/upload/token').then(unwrapApiResponse<UploadTokenResponse>)
}

/* ============ RBAC 系统管理 ============ */

// 获取当前登录员工的可见菜单树与权限码（登录后刷新权限）
export function getRbacPermissions() {
  return request.get('/api/v1/merchant/rbac/permissions').then(unwrapApiResponse<RbacPermissions>)
}

/* ---------- 菜单管理 ---------- */

export function getMenuTree() {
  return request.get('/api/v1/merchant/rbac/menus').then(unwrapApiResponse<SysMenu[]>)
}

export function createMenu(data: Partial<SysMenu>) {
  return request.post('/api/v1/merchant/rbac/menus', data).then(unwrapApiResponse<SysMenu>)
}

export function updateMenu(menuId: number, data: Partial<SysMenu>) {
  return request.put(`/api/v1/merchant/rbac/menus/${menuId}`, data).then(unwrapApiResponse<SysMenu>)
}

export function deleteMenu(menuId: number) {
  return request.delete(`/api/v1/merchant/rbac/menus/${menuId}`).then(unwrapApiResponse<{ message: string }>)
}

/* ---------- 角色管理 ---------- */

export function getRoleList(params?: { page?: number; page_size?: number; keyword?: string; status?: number | string }) {
  return request.get('/api/v1/merchant/rbac/roles', { params }).then(
    unwrapApiResponse<{ list: SysRole[]; pagination: { total: number; page: number; page_size: number } }>
  )
}

export function getAllRoles() {
  return request.get('/api/v1/merchant/rbac/roles/all').then(unwrapApiResponse<SysRole[]>)
}

export function createRole(data: Partial<SysRole>) {
  return request.post('/api/v1/merchant/rbac/roles', data).then(unwrapApiResponse<SysRole>)
}

export function updateRole(roleId: number, data: Partial<SysRole>) {
  return request.put(`/api/v1/merchant/rbac/roles/${roleId}`, data).then(unwrapApiResponse<SysRole>)
}

export function deleteRole(roleId: number) {
  return request.delete(`/api/v1/merchant/rbac/roles/${roleId}`).then(unwrapApiResponse<{ message: string }>)
}

export function getRoleMenus(roleId: number) {
  return request.get(`/api/v1/merchant/rbac/roles/${roleId}/menus`).then(
    unwrapApiResponse<{ role_id: number; menu_ids: number[] }>
  )
}

export function assignRoleMenus(roleId: number, data: AssignRoleMenusRequest) {
  return request.put(`/api/v1/merchant/rbac/roles/${roleId}/menus`, data).then(unwrapApiResponse<{ message: string }>)
}

/* ---------- 部门管理 ---------- */

export function getDepartmentTree() {
  return request.get('/api/v1/merchant/rbac/departments').then(unwrapApiResponse<SysDepartment[]>)
}

export function createDepartment(data: Partial<SysDepartment>) {
  return request.post('/api/v1/merchant/rbac/departments', data).then(unwrapApiResponse<SysDepartment>)
}

export function updateDepartment(departmentId: number, data: Partial<SysDepartment>) {
  return request.put(`/api/v1/merchant/rbac/departments/${departmentId}`, data).then(unwrapApiResponse<SysDepartment>)
}

export function deleteDepartment(departmentId: number) {
  return request.delete(`/api/v1/merchant/rbac/departments/${departmentId}`).then(unwrapApiResponse<{ message: string }>)
}

/* ---------- 员工管理 ---------- */

export function getMerchantStaffList(params?: { page?: number; page_size?: number; keyword?: string }) {
  return request.get('/api/v1/merchant/staff', { params }).then(unwrapApiResponse<MerchantStaffListResponse>)
}

export function createMerchantStaff(data: {
  name: string
  phone: string
  username: string
  password: string
  role?: string
  department_id?: number | null
  role_ids?: number[]
}) {
  return request.post('/api/v1/merchant/staff', data).then(unwrapApiResponse<{ id: number; message: string }>)
}

export function updateMerchantStaff(
  staffId: number,
  data: {
    name?: string
    phone?: string
    role?: string
    department_id?: number | null
    role_ids?: number[]
    status?: number
  }
) {
  return request.put(`/api/v1/merchant/staff/${staffId}`, data).then(unwrapApiResponse<MerchantStaffInfo>)
}

export function deleteMerchantStaff(staffId: number) {
  return request.delete(`/api/v1/merchant/staff/${staffId}`).then(unwrapApiResponse<{ message: string }>)
}

export function resetMerchantStaffPassword(staffId: number, newPassword: string) {
  return request
    .post(`/api/v1/merchant/staff/${staffId}/reset-password`, { new_password: newPassword })
    .then(unwrapApiResponse<{ message: string }>)
}

/* ============ 健康服务 ============ */

// 后端 JSON 字段可能返回字符串或数组，统一规范化为字符串数组
function normalizeStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map((item) => String(item))
  }
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) {
        return parsed.map((item) => String(item))
      }
    } catch (_e) {
      // 非 JSON 字符串，按逗号拆分
    }
    return value
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  }
  return []
}

// 规范化健康档案中的 JSON 数组字段
function normalizeHealthRecord(record: HealthRecord): HealthRecord {
  const jsonFields = [
    'past_history',
    'allergy_history',
    'family_history',
    'surgery_history',
    'medication_list',
    'chronic_tags'
  ] as const
  const normalized: HealthRecord = { ...record }
  for (const field of jsonFields) {
    ;(normalized as unknown as Record<string, unknown>)[field] = normalizeStringArray(record[field])
  }
  return normalized
}

export function getHealthRecords(params?: { keyword?: string; assessment_level?: string; status?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/health-records', { params })
    .then(unwrapApiResponse<HealthRecordListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeHealthRecord) }))
}

export function getHealthRecord(id: number) {
  return request
    .get(`/api/v1/merchant/health-records/${id}`)
    .then(unwrapApiResponse<HealthRecord>)
    .then(normalizeHealthRecord)
}

export function updateHealthRecord(
  id: number,
  data: {
    real_name?: string
    gender?: number
    height_cm?: number | null
    weight_kg?: number | null
    blood_type?: string
    chronic_tags?: string[]
    smoking?: string
    drinking?: string
    assessment_level?: string
    remark?: string
  }
) {
  return request
    .put(`/api/v1/merchant/health-records/${id}`, data)
    .then(unwrapApiResponse<HealthRecord>)
    .then(normalizeHealthRecord)
}

export function getAssessmentForms() {
  return request
    .get('/api/v1/merchant/assessment-forms')
    .then(unwrapApiResponse<AssessmentForm[]>)
    .then((list) => (list || []).map(normalizeAssessmentForm))
}

export function createAssessmentForm(data: Partial<AssessmentForm>) {
  return request
    .post('/api/v1/merchant/assessment-forms', data)
    .then(unwrapApiResponse<AssessmentForm>)
    .then(normalizeAssessmentForm)
}

export function updateAssessmentForm(id: number, data: Partial<AssessmentForm>) {
  return request
    .put(`/api/v1/merchant/assessment-forms/${id}`, data)
    .then(unwrapApiResponse<AssessmentForm>)
    .then(normalizeAssessmentForm)
}

export function setAssessmentFormStatus(id: number, status: number) {
  return request
    .patch(`/api/v1/merchant/assessment-forms/${id}/status`, { status })
    .then(unwrapApiResponse<{ id: number; status: number }>)
}

export function deleteAssessmentForm(id: number) {
  return request.delete(`/api/v1/merchant/assessment-forms/${id}`).then(unwrapApiResponse<{ message: string }>)
}

export function getHealthAssessments(params?: { keyword?: string; form_id?: number | string; page?: number; page_size?: number }) {
  return request.get('/api/v1/merchant/health-assessments', { params }).then(unwrapApiResponse<HealthAssessmentListResponse>)
}

// 规范化评估量表中的 JSON 字段（题目/评分规则可能以字符串返回）
function normalizeAssessmentForm(form: AssessmentForm): AssessmentForm {
  const normalized: AssessmentForm = { ...form }
  const raw = form as unknown as Record<string, unknown>

  // 题目
  const questions = raw.questions
  if (Array.isArray(questions)) {
    normalized.questions = questions as AssessmentFormQuestion[]
  } else if (typeof questions === 'string') {
    try {
      const parsed = JSON.parse(questions)
      normalized.questions = Array.isArray(parsed) ? parsed : []
    } catch (_e) {
      normalized.questions = []
    }
  } else {
    normalized.questions = []
  }

  // 评分规则
  const scoreRule = raw.score_rule
  if (Array.isArray(scoreRule)) {
    normalized.score_rule = scoreRule as AssessmentForm['score_rule']
  } else if (typeof scoreRule === 'string') {
    try {
      const parsed = JSON.parse(scoreRule)
      normalized.score_rule = Array.isArray(parsed) ? parsed : []
    } catch (_e) {
      normalized.score_rule = []
    }
  } else {
    normalized.score_rule = []
  }

  return normalized
}

/* ---------- 适配建议 ---------- */

// 规范化适配建议中的 JSON 数组字段（推荐商品可能以字符串返回）
function normalizeFittingRecommendation(record: FittingRecommendation): FittingRecommendation {
  const normalized: FittingRecommendation = { ...record }
  const raw = record as unknown as Record<string, unknown>
  const products = raw.recommended_products
  if (Array.isArray(products)) {
    normalized.recommended_products = products as FittingRecommendedProduct[]
  } else if (typeof products === 'string') {
    try {
      const parsed = JSON.parse(products)
      normalized.recommended_products = Array.isArray(parsed) ? parsed : []
    } catch (_e) {
      normalized.recommended_products = []
    }
  } else {
    normalized.recommended_products = []
  }
  return normalized
}

export function getFittingRecommendations(params?: { keyword?: string; status?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/fitting-recommendations', { params })
    .then(unwrapApiResponse<FittingRecommendationListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeFittingRecommendation) }))
}

export function getFittingRecommendation(id: number) {
  return request
    .get(`/api/v1/merchant/fitting-recommendations/${id}`)
    .then(unwrapApiResponse<FittingRecommendation>)
    .then(normalizeFittingRecommendation)
}

export function updateFittingRecommendation(id: number, data: Partial<FittingRecommendation>) {
  return request
    .put(`/api/v1/merchant/fitting-recommendations/${id}`, data)
    .then(unwrapApiResponse<FittingRecommendation>)
    .then(normalizeFittingRecommendation)
}

export function deleteFittingRecommendation(id: number) {
  return request.delete(`/api/v1/merchant/fitting-recommendations/${id}`).then(unwrapApiResponse<{ message: string }>)
}

/* ---------- 照护计划 ---------- */

// 规范化照护记录中的 JSON 字段（护理项/生命体征/照片可能以字符串返回）
function normalizeCareVisit(visit: CareVisit): CareVisit {
  const normalized: CareVisit = { ...visit }
  const raw = visit as unknown as Record<string, unknown>

  // 护理项列表
  const nursingItems = raw.nursing_items
  if (Array.isArray(nursingItems)) {
    normalized.nursing_items = nursingItems as CareVisit['nursing_items']
  } else if (typeof nursingItems === 'string') {
    try {
      const parsed = JSON.parse(nursingItems)
      normalized.nursing_items = Array.isArray(parsed) ? parsed : []
    } catch (_e) {
      normalized.nursing_items = []
    }
  } else {
    normalized.nursing_items = []
  }

  // 生命体征对象
  const vitals = raw.vitals
  if (vitals && typeof vitals === 'object') {
    normalized.vitals = vitals as CareVisit['vitals']
  } else if (typeof vitals === 'string') {
    try {
      normalized.vitals = JSON.parse(vitals) || {}
    } catch (_e) {
      normalized.vitals = {}
    }
  } else {
    normalized.vitals = {}
  }

  // 照片数组
  normalized.photos = normalizeStringArray(raw.photos)

  return normalized
}

// 规范化照护计划中的 JSON 字段（护理项/内嵌照护记录可能以字符串返回）
function normalizeCarePlan(record: CarePlan): CarePlan {
  const normalized: CarePlan = { ...record }
  const raw = record as unknown as Record<string, unknown>

  // 护理项列表
  const items = raw.items
  if (Array.isArray(items)) {
    normalized.items = items as CarePlanItem[]
  } else if (typeof items === 'string') {
    try {
      const parsed = JSON.parse(items)
      normalized.items = Array.isArray(parsed) ? parsed : []
    } catch (_e) {
      normalized.items = []
    }
  } else {
    normalized.items = []
  }

  // 内嵌照护记录列表
  const visits = raw.visits
  if (Array.isArray(visits)) {
    normalized.visits = visits.map(normalizeCareVisit)
  } else if (typeof visits === 'string') {
    try {
      const parsed = JSON.parse(visits)
      normalized.visits = Array.isArray(parsed) ? parsed.map(normalizeCareVisit) : []
    } catch (_e) {
      normalized.visits = []
    }
  } else {
    normalized.visits = []
  }

  return normalized
}

export function getCarePlans(params?: { keyword?: string; plan_type?: number | string; status?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/care-plans', { params })
    .then(unwrapApiResponse<CarePlanListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeCarePlan) }))
}

export function getCarePlan(id: number) {
  return request
    .get(`/api/v1/merchant/care-plans/${id}`)
    .then(unwrapApiResponse<CarePlan>)
    .then(normalizeCarePlan)
}

export function createCarePlan(data: Partial<CarePlan>) {
  return request
    .post('/api/v1/merchant/care-plans', data)
    .then(unwrapApiResponse<CarePlan>)
    .then(normalizeCarePlan)
}

export function updateCarePlan(id: number, data: Partial<CarePlan>) {
  return request
    .put(`/api/v1/merchant/care-plans/${id}`, data)
    .then(unwrapApiResponse<CarePlan>)
    .then(normalizeCarePlan)
}

export function deleteCarePlan(id: number) {
  return request.delete(`/api/v1/merchant/care-plans/${id}`).then(unwrapApiResponse<{ message: string }>)
}

export function getCareVisits(params?: { plan_id?: number | string; user_id?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/care-visits', { params })
    .then(unwrapApiResponse<CareVisitListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeCareVisit) }))
}

/* ---------- 随访任务 ---------- */

// 规范化随访任务中的 JSON 字段（执行结果可能以字符串返回）
function normalizeFollowUpTask(task: FollowUpTask): FollowUpTask {
  const normalized: FollowUpTask = { ...task }
  const raw = task as unknown as Record<string, unknown>

  // 执行结果对象
  const result = raw.result
  if (result && typeof result === 'object') {
    normalized.result = result as FollowUpTaskResult
  } else if (typeof result === 'string') {
    try {
      normalized.result = JSON.parse(result) || {}
    } catch (_e) {
      normalized.result = {}
    }
  } else {
    normalized.result = {}
  }

  // 结果中的宣教文章 ID 列表（可能以字符串数组返回）
  if (normalized.result) {
    const ids = normalized.result.education_article_ids
    if (Array.isArray(ids)) {
      normalized.result.education_article_ids = ids
        .map((id) => Number(id))
        .filter((id) => !Number.isNaN(id))
    } else {
      normalized.result.education_article_ids = []
    }
  }

  return normalized
}

export function getFollowUpTasks(params?: { keyword?: string; task_type?: number | string; status?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/follow-up-tasks', { params })
    .then(unwrapApiResponse<FollowUpTaskListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeFollowUpTask) }))
}

export function getFollowUpTask(id: number) {
  return request
    .get(`/api/v1/merchant/follow-up-tasks/${id}`)
    .then(unwrapApiResponse<FollowUpTask>)
    .then(normalizeFollowUpTask)
}

export function createFollowUpTask(data: Partial<FollowUpTask>) {
  return request
    .post('/api/v1/merchant/follow-up-tasks', data)
    .then(unwrapApiResponse<FollowUpTask>)
    .then(normalizeFollowUpTask)
}

export function completeFollowUpTask(id: number, data: { result: FollowUpTaskResult }) {
  return request
    .post(`/api/v1/merchant/follow-up-tasks/${id}/complete`, data)
    .then(unwrapApiResponse<FollowUpTask>)
    .then(normalizeFollowUpTask)
}

/* ---------- 生命体征监测 ---------- */

// 规范化生命体征记录中的 JSON 字段（附加信息可能以字符串返回）
function normalizeHealthMonitoring(record: HealthMonitoring): HealthMonitoring {
  const normalized: HealthMonitoring = { ...record }
  const raw = record as unknown as Record<string, unknown>

  // 附加信息对象
  const extra = raw.extra
  if (extra && typeof extra === 'object') {
    normalized.extra = extra as HealthMonitoring['extra']
  } else if (typeof extra === 'string') {
    try {
      normalized.extra = JSON.parse(extra) || {}
    } catch (_e) {
      normalized.extra = {}
    }
  } else {
    normalized.extra = {}
  }

  return normalized
}

export function getMonitoring(params?: { keyword?: string; record_type?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/monitoring', { params })
    .then(unwrapApiResponse<HealthMonitoringListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeHealthMonitoring) }))
}

/* ---------- 健康宣教 ---------- */

// 规范化宣教文章中的 JSON 字段（定向标签可能以字符串返回）
function normalizeEducationArticle(article: EducationArticle): EducationArticle {
  const normalized: EducationArticle = { ...article }
  normalized.tags = normalizeStringArray(article.tags)
  return normalized
}

export function getEducationArticles(params?: { keyword?: string; category?: string; status?: number | string; page?: number; page_size?: number }) {
  return request
    .get('/api/v1/merchant/health-education', { params })
    .then(unwrapApiResponse<EducationArticleListResponse>)
    .then((res) => ({ ...res, list: (res.list || []).map(normalizeEducationArticle) }))
}

export function createEducationArticle(data: Partial<EducationArticle>) {
  return request
    .post('/api/v1/merchant/health-education', data)
    .then(unwrapApiResponse<EducationArticle>)
    .then(normalizeEducationArticle)
}

export function updateEducationArticle(id: number, data: Partial<EducationArticle>) {
  return request
    .put(`/api/v1/merchant/health-education/${id}`, data)
    .then(unwrapApiResponse<EducationArticle>)
    .then(normalizeEducationArticle)
}

export function deleteEducationArticle(id: number) {
  return request.delete(`/api/v1/merchant/health-education/${id}`).then(unwrapApiResponse<{ message: string }>)
}

/* ============ 服务商分账 ============ */

// 分账接收方
export function getProfitSharingReceivers() {
  return request.get('/api/v1/merchant/profit-sharing/receivers').then(unwrapApiResponse<ProfitSharingReceiver[]>)
}

export function createProfitSharingReceiver(data: ProfitSharingReceiverPayload) {
  return request
    .post('/api/v1/merchant/profit-sharing/receivers', data)
    .then(unwrapApiResponse<ProfitSharingReceiver>)
}

export function updateProfitSharingReceiver(id: number, data: Partial<ProfitSharingReceiverPayload>) {
  return request
    .put(`/api/v1/merchant/profit-sharing/receivers/${id}`, data)
    .then(unwrapApiResponse<ProfitSharingReceiver>)
}

export function deleteProfitSharingReceiver(id: number) {
  return request
    .delete(`/api/v1/merchant/profit-sharing/receivers/${id}`)
    .then(unwrapApiResponse<{ message: string }>)
}

export function syncProfitSharingReceiver(id: number) {
  return request
    .post(`/api/v1/merchant/profit-sharing/receivers/${id}/sync`)
    .then(unwrapApiResponse<ProfitSharingReceiver>)
}

// 分账配置
export function getProfitSharingConfig() {
  return request.get('/api/v1/merchant/profit-sharing/config').then(unwrapApiResponse<ProfitSharingConfig>)
}

export function updateProfitSharingConfig(enabled: boolean) {
  return request
    .put('/api/v1/merchant/profit-sharing/config', { profit_sharing_enabled: enabled })
    .then(unwrapApiResponse<boolean>)
}

// 分账记录
export function getProfitSharingRecords(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/profit-sharing/records', { params }).then(
    unwrapApiResponse<ProfitSharingRecordListResponse>
  )
}

export function getProfitSharingRecord(id: number) {
  return request.get(`/api/v1/merchant/profit-sharing/records/${id}`).then(unwrapApiResponse<ProfitSharingRecord>)
}

export function retryProfitSharingRecord(id: number) {
  return request
    .post(`/api/v1/merchant/profit-sharing/records/${id}/retry`)
    .then(unwrapApiResponse<ProfitSharingRecord>)
}
