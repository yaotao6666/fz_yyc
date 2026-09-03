import request, { unwrapApiResponse } from '@/utils/request'
import type {
  AssessmentForm,
  AssessmentFormQuestion,
  AssignRoleMenusRequest,
  DashboardData,
  EducationArticle,
  EducationArticleListResponse,
  HealthEducationCategory,
  FittingRecommendation,
  FittingRecommendationListResponse,
  FittingRecommendedProduct,
  HealthAssessment,
  HealthAssessmentListResponse,
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
  StaffAuditDetailResponse,
  StaffAuditListResponse,
  SpOrder,
  SpOrderListResponse,
  SysDepartment,
  SysMenu,
  SysRole,
  UpdateSpMerchantFormData,
  MiniProgramBanner,
  MiniProgramBannerListResponse,
  MiniProgramBannerPayload,
  HomeRecommend,
  HomeRecommendListResponse,
  HomeRecommendPayload,
  UploadTokenResponse,
  MerchantSettings,
  Printer,
  PrinterListResponse,
  PrinterPayload,
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

/* ============ 商家运营设置 ============ */

// 获取商家营业设置（下单方式/配送/公告/营业时间等）
export function getMerchantSettings() {
  return request.get('/api/v1/merchant/settings').then(unwrapApiResponse<MerchantSettings>)
}

// 更新商家营业设置（下单方式开关等）
export function updateMerchantSettings(payload: Partial<MerchantSettings>) {
  return request.put('/api/v1/merchant/settings', payload).then(unwrapApiResponse<MerchantSettings>)
}

// 更新营业状态：status 0=休息中 1=营业中
export function updateMerchantStatus(status: number) {
  return request.post('/api/v1/merchant/status', { status }).then(unwrapApiResponse<{ status: number }>)
}

/* ============ 打印机管理 ============ */

export function getPrinters() {
  return request.get('/api/v1/merchant/printers').then(unwrapApiResponse<PrinterListResponse>)
}

export function createPrinter(data: PrinterPayload) {
  return request.post('/api/v1/merchant/printers', data).then(unwrapApiResponse<Printer>)
}

export function updatePrinter(printerId: number, data: Partial<PrinterPayload> & { is_default?: number }) {
  return request.put(`/api/v1/merchant/printers/${printerId}`, data).then(unwrapApiResponse<Printer>)
}

export function deletePrinter(printerId: number) {
  return request.delete(`/api/v1/merchant/printers/${printerId}`).then(unwrapApiResponse<{ message: string }>)
}

export function testPrinter(printerId: number) {
  return request.post(`/api/v1/merchant/printers/${printerId}/test`, {}).then(unwrapApiResponse<{ message: string }>)
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

/* ============ 订单派单 / 租赁到期 / 续租 ============ */
export function getDispatchableStaffList() {
  return request.get('/api/v1/merchant/orders/dispatchable-staff').then(unwrapApiResponse<{ id: number; name: string; phone?: string }[]>)
}

export function dispatchOrder(orderId: number, staff_id: number) {
  return request.post(`/api/v1/merchant/orders/${orderId}/dispatch`, { staff_id }).then(unwrapApiResponse<{ id: number; assigned_staff_id: number; biz_status: number }>)
}

export function getRentalDueOrders(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/orders/rental-due', { params }).then(unwrapApiResponse<{
    list: { order: SpOrder; rental_end_at?: string; is_overdue?: boolean; days_left?: number }[]
    total: number
    pagination: { total: number; page: number; page_size: number }
  }>)
}

export function renewOrder(orderId: number, duration?: number) {
  return request.post(`/api/v1/merchant/orders/${orderId}/renew`, { duration }).then(unwrapApiResponse<{ order: SpOrder }>)
}

/* ============ 服务人员管理 ============ */

export function getServiceStaffList(params?: { status?: number; keyword?: string; page?: number; page_size?: number }) {
  return request.get('/api/v1/merchant/service-staff', { params }).then(unwrapApiResponse<ServiceStaffListResponse>)
}

export function createServiceStaff(data: { username: string; password: string; name: string; phone: string }) {
  return request.post('/api/v1/merchant/service-staff', data).then(unwrapApiResponse<{ id: number }>)
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

/* ============ 服务人员审核中心 ============ */
export function listStaffAudits(params?: { audit_type?: number; status?: number; keyword?: string; page?: number; page_size?: number }) {
  return request.get('/api/v1/merchant/staff-audits', { params }).then(unwrapApiResponse<StaffAuditListResponse>)
}

export function getStaffAuditDetail(id: number) {
  return request.get(`/api/v1/merchant/staff-audits/${id}`, {}).then(unwrapApiResponse<StaffAuditDetailResponse>)
}

export function approveStaffAudit(id: number, remark?: string) {
  return request.post(`/api/v1/merchant/staff-audits/${id}/approve`, { remark }).then(unwrapApiResponse<{ id: number; status: number }>)
}

export function rejectStaffAudit(id: number, remark?: string) {
  return request.post(`/api/v1/merchant/staff-audits/${id}/reject`, { remark }).then(unwrapApiResponse<{ id: number; status: number }>)
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

// 对单个七牛资源地址进行私有签名，供富文本编辑器实时预览粘贴/上传的图片使用
export function signMediaUrl(url: string) {
  return request
    .post('/api/v1/upload/sign', { url })
    .then(unwrapApiResponse<{ url: string }>)
    .then((data) => data?.url || url)
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
    .then(unwrapApiResponse<{ record: HealthRecord; assessments: HealthAssessment[] }>)
    .then((res) => normalizeHealthRecord({ ...(res.record || {}), assessments: res.assessments ?? [] }))
}

export function getRecordAssessments(id: number) {
  return request.get(`/api/v1/merchant/health-records/${id}/assessments`).then(unwrapApiResponse<HealthAssessment[]>)
}

export function updateHealthRecord(
  id: number,
  data: {
    relation?: number
    real_name?: string
    gender?: number
    birth_date?: string
    id_card?: string
    phone?: string
    emergency_contact?: string
    emergency_phone?: string
    address?: string
    height_cm?: number | null
    weight_kg?: number | null
    blood_type?: string
    past_history?: string[]
    allergy_history?: string[]
    family_history?: string[]
    surgery_history?: string[]
    medication_list?: string[]
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

// ---- 健康宣教分类 ----

export function getHealthEducationCategories() {
  return request
    .get('/api/v1/merchant/education-categories')
    .then(unwrapApiResponse<HealthEducationCategory[]>)
}

export function createHealthEducationCategory(data: { parent_id?: number; name: string; sort?: number; status?: number }) {
  return request
    .post('/api/v1/merchant/education-categories', data)
    .then(unwrapApiResponse<HealthEducationCategory>)
}

export function updateHealthEducationCategory(id: number, data: { parent_id?: number; name: string; sort?: number; status?: number }) {
  return request
    .put(`/api/v1/merchant/education-categories/${id}`, data)
    .then(unwrapApiResponse<{ message?: string }>)
}

export function deleteHealthEducationCategory(id: number) {
  return request
    .delete(`/api/v1/merchant/education-categories/${id}`)
    .then(unwrapApiResponse<{ message: string }>)
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

/* ============ 小程序轮播图配置 ============ */

export function getMiniProgramBanners() {
  return request
    .get('/api/v1/merchant/miniprogram-banners')
    .then(unwrapApiResponse<MiniProgramBannerListResponse>)
}

export function getMiniProgramBanner(id: number) {
  return request
    .get(`/api/v1/merchant/miniprogram-banners/${id}`)
    .then(unwrapApiResponse<MiniProgramBanner>)
}

export function createMiniProgramBanner(data: MiniProgramBannerPayload) {
  return request
    .post('/api/v1/merchant/miniprogram-banners', data)
    .then(unwrapApiResponse<MiniProgramBanner>)
}

export function updateMiniProgramBanner(id: number, data: Partial<MiniProgramBannerPayload>) {
  return request
    .put(`/api/v1/merchant/miniprogram-banners/${id}`, data)
    .then(unwrapApiResponse<MiniProgramBanner>)
}

export function updateMiniProgramBannerStatus(id: number, status: number) {
  return request
    .patch(`/api/v1/merchant/miniprogram-banners/${id}/status`, { status })
    .then(unwrapApiResponse<{ message?: string }>)
}

export function deleteMiniProgramBanner(id: number) {
  return request
    .delete(`/api/v1/merchant/miniprogram-banners/${id}`)
    .then(unwrapApiResponse<{ message: string }>)
}

/* ============ 小程序首页推荐配置 ============ */

export function getHomeRecommends() {
  return request
    .get('/api/v1/merchant/home-recommends')
    .then(unwrapApiResponse<HomeRecommendListResponse>)
}

export function createHomeRecommend(data: HomeRecommendPayload) {
  return request
    .post('/api/v1/merchant/home-recommends', data)
    .then(unwrapApiResponse<HomeRecommend>)
}

export function updateHomeRecommend(id: number, data: Partial<HomeRecommendPayload>) {
  return request
    .put(`/api/v1/merchant/home-recommends/${id}`, data)
    .then(unwrapApiResponse<HomeRecommend>)
}

export function updateHomeRecommendStatus(id: number, status: number) {
  return request
    .patch(`/api/v1/merchant/home-recommends/${id}/status`, { status })
    .then(unwrapApiResponse<{ message?: string }>)
}

export function deleteHomeRecommend(id: number) {
  return request
    .delete(`/api/v1/merchant/home-recommends/${id}`)
    .then(unwrapApiResponse<{ message: string }>)
}
