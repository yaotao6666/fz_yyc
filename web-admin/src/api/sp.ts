import request, { unwrapApiResponse } from '@/utils/request'
import type {
  AmountTrendData,
  AnnouncementFormData,
  AnnouncementItem,
  AnnouncementListResponse,
  DashboardData,
  MerchantCategory,
  MerchantCategoryPayload,
  MerchantCategorySortItem,
  MerchantDetail,
  MerchantProduct,
  MerchantProductListResponse,
  MerchantProductQuery,
  MerchantProductSpecsPayload,
  MerchantProductSpecsResponse,
  MerchantProductUpsertPayload,
  MerchantDistributionData,
  MerchantListResponse,
  MerchantPaymentConfigFormData,
  MerchantPickupPoint,
  MerchantPickupPointPayload,
  OrderAnalyticsData,
  SpOrder,
  SpOrderListResponse,
  ProfitSharingRecordListResponse,
  ServiceProviderLoginRequest,
  ServiceProviderLoginResponse,
  SpMerchantFormData,
  SpSettings,
  TopMerchantRanking,
  UpdateSpMerchantFormData,
  UploadTokenResponse,
} from '@/types/sp'

export function spLogin(data: ServiceProviderLoginRequest) {
  return request.post('/api/v1/sp/auth/login', data).then(unwrapApiResponse<ServiceProviderLoginResponse>)
}

export function spLogout() {
  return request.post('/api/v1/sp/auth/logout', {}).then(unwrapApiResponse<{ message: string }>)
}

export function getDashboard() {
  return request.get('/api/v1/sp/dashboard').then(unwrapApiResponse<DashboardData>)
}

export function getMerchantList(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/merchants/list', { params }).then(unwrapApiResponse<MerchantListResponse>)
}

export function getMerchantDetail(merchantId: number) {
  return request.get(`/api/v1/sp/merchants/${merchantId}`).then(unwrapApiResponse<MerchantDetail>)
}

export function createSpMerchant(data: SpMerchantFormData) {
  return request.post('/api/v1/sp/merchants', data).then(unwrapApiResponse<MerchantDetail>)
}

export function updateSpMerchant(merchantId: number, data: UpdateSpMerchantFormData) {
  return request.put(`/api/v1/sp/merchants/${merchantId}`, data).then(unwrapApiResponse<MerchantDetail>)
}

export function updateSpMerchantPaymentConfig(merchantId: number, data: MerchantPaymentConfigFormData) {
  return request.put(`/api/v1/sp/merchants/${merchantId}/payment-config`, data).then(unwrapApiResponse<MerchantDetail>)
}

export function resetSpMerchantAdminPassword(merchantId: number, data: { new_password: string }) {
  return request.post(`/api/v1/sp/merchants/${merchantId}/admin/reset-password`, data).then(
    unwrapApiResponse<{ staff_id: number; username: string; message: string }>
  )
}

export function updateSpMerchantAssets(merchantId: number, data: { logo?: string; cover_image?: string }) {
  return request.put(`/api/v1/sp/merchants/${merchantId}/assets`, data).then(unwrapApiResponse<MerchantDetail>)
}

export function getMerchantPickupPoints(merchantId: number) {
  return request
    .get(`/api/v1/sp/merchants/${merchantId}/pickup-points`)
    .then(unwrapApiResponse<MerchantPickupPoint[]>)
}

export function createMerchantPickupPoint(merchantId: number, data: MerchantPickupPointPayload) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/pickup-points`, data)
    .then(unwrapApiResponse<MerchantPickupPoint>)
}

export function updateMerchantPickupPoint(merchantId: number, pickupPointId: number, data: MerchantPickupPointPayload) {
  return request
    .put(`/api/v1/sp/merchants/${merchantId}/pickup-points/${pickupPointId}`, data)
    .then(unwrapApiResponse<MerchantPickupPoint>)
}

export function deleteMerchantPickupPoint(merchantId: number, pickupPointId: number) {
  return request
    .delete(`/api/v1/sp/merchants/${merchantId}/pickup-points/${pickupPointId}`)
    .then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantCategories(merchantId: number) {
  return request
    .get(`/api/v1/sp/merchants/${merchantId}/categories`)
    .then(unwrapApiResponse<MerchantCategory[]>)
}

export function createMerchantCategory(merchantId: number, data: MerchantCategoryPayload) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/categories`, data)
    .then(unwrapApiResponse<MerchantCategory>)
}

export function updateMerchantCategory(merchantId: number, categoryId: number, data: MerchantCategoryPayload) {
  return request
    .put(`/api/v1/sp/merchants/${merchantId}/categories/${categoryId}`, data)
    .then(unwrapApiResponse<MerchantCategory>)
}

export function deleteMerchantCategory(merchantId: number, categoryId: number) {
  return request
    .delete(`/api/v1/sp/merchants/${merchantId}/categories/${categoryId}`)
    .then(unwrapApiResponse<{ message: string }>)
}

export function sortMerchantCategories(merchantId: number, categories: MerchantCategorySortItem[]) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/categories/sort`, { categories })
    .then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantProducts(merchantId: number, params?: MerchantProductQuery) {
  return request
    .get(`/api/v1/sp/merchants/${merchantId}/products`, { params })
    .then(unwrapApiResponse<MerchantProductListResponse>)
}

export function getMerchantProduct(merchantId: number, productId: number) {
  return request
    .get(`/api/v1/sp/merchants/${merchantId}/products/${productId}`)
    .then(unwrapApiResponse<MerchantProduct>)
}

export function createMerchantProduct(merchantId: number, data: MerchantProductUpsertPayload) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/products`, data)
    .then(unwrapApiResponse<MerchantProduct>)
}

export function updateMerchantProduct(merchantId: number, productId: number, data: MerchantProductUpsertPayload) {
  return request
    .put(`/api/v1/sp/merchants/${merchantId}/products/${productId}`, data)
    .then(unwrapApiResponse<MerchantProduct>)
}

export function merchantProductOnSale(merchantId: number, productId: number) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/products/${productId}/on-sale`, {})
    .then(unwrapApiResponse<{ message: string }>)
}

export function merchantProductOffSale(merchantId: number, productId: number) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/products/${productId}/off-sale`, {})
    .then(unwrapApiResponse<{ message: string }>)
}

export function batchUpdateMerchantProductStatus(merchantId: number, productIds: number[], status: number) {
  return request
    .post(`/api/v1/sp/merchants/${merchantId}/products/batch-status`, { product_ids: productIds, status })
    .then(unwrapApiResponse<{ message: string }>)
}

export function deleteMerchantProduct(merchantId: number, productId: number) {
  return request
    .delete(`/api/v1/sp/merchants/${merchantId}/products/${productId}`)
    .then(unwrapApiResponse<{ message: string }>)
}

export function updateMerchantProductStock(merchantId: number, productId: number, stock: number) {
  return request
    .put(`/api/v1/sp/merchants/${merchantId}/products/${productId}/stock`, { stock })
    .then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantProductSpecs(merchantId: number, productId: number) {
  return request
    .get(`/api/v1/sp/merchants/${merchantId}/products/${productId}/specs`)
    .then(unwrapApiResponse<MerchantProductSpecsResponse>)
}

export function updateMerchantProductSpecs(merchantId: number, productId: number, data: MerchantProductSpecsPayload) {
  return request
    .put(`/api/v1/sp/merchants/${merchantId}/products/${productId}/specs`, data)
    .then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantDistribution(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/merchants/analytics/distribution', { params }).then(unwrapApiResponse<MerchantDistributionData>)
}

export function getOrderAnalytics(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/orders/analytics', { params }).then(unwrapApiResponse<OrderAnalyticsData>)
}

export function getSpOrders(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/orders', { params }).then(unwrapApiResponse<SpOrderListResponse>)
}

export function getSpOrderDetail(orderId: number) {
  return request.get(`/api/v1/sp/orders/${orderId}`).then(unwrapApiResponse<SpOrder>)
}

export function getAmountAnalytics(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/amount/analytics', { params }).then(unwrapApiResponse<AmountTrendData>)
}

export function getTopMerchants(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/amount/top-merchants', { params }).then(unwrapApiResponse<TopMerchantRanking[]>)
}

export function getSpProfitSharingRecords(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/profit-sharing-records', { params }).then(unwrapApiResponse<ProfitSharingRecordListResponse>)
}

export function getSpSettings() {
  return request.get('/api/v1/sp/settings').then(unwrapApiResponse<SpSettings>)
}

export function updateSpSettings(data: { contact_name?: string; contact_phone?: string }) {
  return request.put('/api/v1/sp/settings', data).then(unwrapApiResponse<{ message: string }>)
}

export function changeSpPassword(data: { old_password: string; new_password: string }) {
  return request.post('/api/v1/sp/account/change-password', data).then(unwrapApiResponse<{ message: string }>)
}

export function getMerchantQRCode(merchantId: number) {
  return request.get(`/api/v1/sp/merchants/${merchantId}/qrcode`).then(
    unwrapApiResponse<{ merchant_id: number; merchant_name: string; qrcode_url: string; page_path: string }>
  )
}

export function getUploadToken(params?: { merchant_id?: number }) {
  return request.get('/api/v1/upload/token', { params }).then(unwrapApiResponse<UploadTokenResponse>)
}

export function getAnnouncements(params?: Record<string, unknown>) {
  return request.get('/api/v1/sp/announcements', { params }).then(unwrapApiResponse<AnnouncementListResponse>)
}

export function getAnnouncementDetail(announcementId: number) {
  return request.get(`/api/v1/sp/announcements/${announcementId}`).then(unwrapApiResponse<AnnouncementItem>)
}

export function createAnnouncement(data: AnnouncementFormData) {
  return request.post('/api/v1/sp/announcements', data).then(
    unwrapApiResponse<{ id: number; message: string }>
  )
}

export function updateAnnouncement(announcementId: number, data: AnnouncementFormData) {
  return request.put(`/api/v1/sp/announcements/${announcementId}`, data).then(unwrapApiResponse<AnnouncementItem>)
}

export function deleteAnnouncement(announcementId: number) {
  return request.delete(`/api/v1/sp/announcements/${announcementId}`).then(
    unwrapApiResponse<{ message: string }>
  )
}
