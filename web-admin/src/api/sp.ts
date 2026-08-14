import request, { unwrapApiResponse } from '@/utils/request'
import type {
  DashboardData,
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
  OrderAnalyticsData,
  SpOrder,
  SpOrderListResponse,
  ServiceStaffListResponse,
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
