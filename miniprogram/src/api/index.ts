/**
 * API接口封装 - 基于PRD文档
 * 统一管理所有API接口调用
 */

import { get, post, put, del, upload } from '../utils/request'
import type {
  MerchantLoginRequest,
  MerchantLoginResponse,
  MerchantRegisterRequest,
  MerchantRegisterResponse,
  MerchantInfo,
  MerchantApplicationStatus,
  MerchantSettings,
  DeliverySettings,
  Category,
  Product,
  ProductListResponse,
  Order,
  OrderListResponse,
  OrderStatistics,
  CreateOrderRequest,
  CreateOrderResponse,
  SalesOverview,
  SalesTrend,
  ProductRanking,
  HourlyAnalysis,
  StockAlert,
  InviteInfo,
  MyInviteInfo,
  InviteRecordListResponse,
  StoreHomeInfo,
  StoreProductGroup,
  PaginationParams
} from '../types/api'

// ============ 认证相关 ============

/**
 * 商家登录
 */
export function merchantLogin(data: MerchantLoginRequest) {
  return post<MerchantLoginResponse>('/api/v1/merchant/auth/login', data)
}

/**
 * 获取当前商家信息
 */
export function getMerchantProfile() {
  return get<MerchantInfo>('/api/v1/merchant/profile')
}

// ============ 商家入驻相关 ============

/**
 * 商家入驻申请
 */
export function merchantRegister(data: MerchantRegisterRequest) {
  return post<MerchantRegisterResponse>('/api/v1/merchant/register', data)
}

/**
 * 获取商家进件状态
 */
export function getMerchantApplicationStatus() {
  return get<MerchantApplicationStatus>('/api/v1/merchant/application/status')
}

// ============ 商家设置相关 ============

/**
 * 更新商家基本信息
 */
export function updateMerchantProfile(data: Partial<MerchantInfo>) {
  return put<null>('/api/v1/merchant/profile', data)
}

/**
 * 获取商家设置
 */
export function getMerchantSettings() {
  return get<MerchantSettings>('/api/v1/merchant/settings')
}

/**
 * 更新商家设置
 */
export function updateMerchantSettings(data: Partial<MerchantSettings>) {
  return put<null>('/api/v1/merchant/settings', data)
}

/**
 * 获取配送设置
 */
export function getDeliverySettings() {
  return get<DeliverySettings>('/api/v1/merchant/delivery-settings')
}

/**
 * 更新配送设置
 */
export function updateDeliverySettings(data: Partial<DeliverySettings>) {
  return put<DeliverySettings>('/api/v1/merchant/delivery-settings', data)
}

/**
 * 开启/关闭店铺
 */
export function updateMerchantStatus(status: 'active' | 'inactive') {
  return post<null>('/api/v1/merchant/status', { status })
}

/**
 * 获取商家小程序码
 */
export function getMerchantQrcode(params?: { page?: string; width?: number }) {
  return get<{ qrcode_url: string; expire_time?: string }>('/api/v1/merchant/qrcode', params)
}

// ============ 商品分类相关 ============

/**
 * 获取分类列表
 */
export function getCategories() {
  return get<Category[]>('/api/v1/merchant/categories')
}

/**
 * 创建分类
 */
export function createCategory(data: { name: string; sort?: number }) {
  return post<Category>('/api/v1/merchant/categories', data)
}

/**
 * 更新分类
 */
export function updateCategory(categoryId: number, data: { name?: string; sort?: number }) {
  return put<Category>(`/api/v1/merchant/categories/${categoryId}`, data)
}

/**
 * 删除分类
 */
export function deleteCategory(categoryId: number) {
  return del<null>(`/api/v1/merchant/categories/${categoryId}`)
}

/**
 * 批量排序分类
 */
export function sortCategories(orders: { id: number; sort: number }[]) {
  return post<null>('/api/v1/merchant/categories/sort', { orders })
}

// ============ 商品管理相关 ============

/**
 * 获取商品列表
 */
export function getProducts(params?: {
  page?: number
  page_size?: number
  category_id?: number
  status?: string
  keyword?: string
}) {
  return get<ProductListResponse>('/api/v1/merchant/products', params)
}

/**
 * 获取商品详情
 */
export function getProduct(productId: number) {
  return get<Product>(`/api/v1/merchant/products/${productId}`)
}

/**
 * 创建商品
 */
export function createProduct(data: {
  name: string
  description?: string
  images: string[]
  category_id: number
  price: number
  original_price?: number
  stock?: number
  unit?: string
  sort?: number
  specs?: { name: string; options: { name: string; price: number; stock?: number }[] }[]
}) {
  return post<Product>('/api/v1/merchant/products', data)
}

/**
 * 更新商品
 */
export function updateProduct(productId: number, data: Partial<Product>) {
  return put<Product>(`/api/v1/merchant/products/${productId}`, data)
}

/**
 * 商品上架
 */
export function productOnSale(productId: number) {
  return post<null>(`/api/v1/merchant/products/${productId}/on-sale`)
}

/**
 * 商品下架
 */
export function productOffSale(productId: number) {
  return post<null>(`/api/v1/merchant/products/${productId}/off-sale`)
}

/**
 * 批量更新商品状态
 */
export function batchUpdateProductStatus(productIds: number[], status: 'on_sale' | 'off_sale') {
  return post<null>('/api/v1/merchant/products/batch-status', { product_ids: productIds, status })
}

/**
 * 删除商品
 */
export function deleteProduct(productId: number) {
  return del<null>(`/api/v1/merchant/products/${productId}`)
}

/**
 * 更新库存
 */
export function updateProductStock(productId: number, data: { stock: number; action: 'set' | 'add' | 'subtract' }) {
  return put<null>(`/api/v1/merchant/products/${productId}/stock`, data)
}

// ============ 订单管理相关 ============

/**
 * 获取订单列表
 */
export function getOrders(params?: {
  page?: number
  page_size?: number
  status?: number
  start_date?: string
  end_date?: string
  order_no?: string
}) {
  return get<OrderListResponse>('/api/v1/merchant/orders', params)
}

/**
 * 获取订单详情
 */
export function getOrder(orderId: number) {
  return get<Order>(`/api/v1/merchant/orders/${orderId}`)
}

/**
 * 订单核销
 */
export function completeOrder(orderId: number, verifyCode: string) {
  return post<{ order_id: number; order_no: string; completed_at: string }>(
    `/api/v1/merchant/orders/${orderId}/complete`,
    { verify_code: verifyCode }
  )
}

/**
 * 订单退款
 */
export function refundOrder(orderId: number, data: { refund_amount: number; refund_reason: string }) {
  return post<{ refund_id: number; refund_no: string; status: string }>(
    `/api/v1/merchant/orders/${orderId}/refund`,
    data
  )
}

/**
 * 获取订单统计
 */
export function getOrderStatistics(params?: { start_date?: string; end_date?: string }) {
  return get<OrderStatistics>('/api/v1/merchant/orders/statistics', params)
}

// ============ 数据分析相关 ============

/**
 * 获取销售概览
 */
export function getSalesOverview(params?: { period?: 'today' | 'week' | 'month' | 'year' }) {
  return get<SalesOverview>('/api/v1/merchant/analytics/overview', params)
}

/**
 * 获取销售趋势
 */
export function getSalesTrend(params: { start_date: string; end_date: string; granularity?: 'day' | 'week' | 'month' }) {
  return get<SalesTrend[]>('/api/v1/merchant/analytics/sales-trend', params)
}

/**
 * 获取商品销量排行
 */
export function getProductRanking(params?: {
  start_date?: string
  end_date?: string
  limit?: number
  sort_by?: 'sales' | 'amount'
}) {
  return get<ProductRanking[]>('/api/v1/merchant/analytics/product-ranking', params)
}

/**
 * 获取时段分析
 */
export function getHourlyAnalysis(params?: { date?: string }) {
  return get<HourlyAnalysis[]>('/api/v1/merchant/analytics/hourly', params)
}

/**
 * 获取库存预警
 */
export function getStockAlert(params?: { threshold?: number }) {
  return get<StockAlert[]>('/api/v1/merchant/analytics/stock-alert', params)
}

// ============ 邀请入驻相关 ============

/**
 * 生成邀请码
 */
export function generateInviteCode() {
  return get<InviteInfo>('/api/v1/merchant/invite/generate')
}

/**
 * 获取我的邀请信息
 */
export function getMyInviteInfo() {
  return get<MyInviteInfo>('/api/v1/merchant/invite/info')
}

/**
 * 获取邀请记录
 */
export function getInviteRecords(params?: PaginationParams & { status?: string }) {
  return get<InviteRecordListResponse>('/api/v1/merchant/invite/records', params)
}

// ============ C端店铺相关 ============

/**
 * 获取店铺首页信息
 */
export function getStoreHome(merchantId: number) {
  return get<StoreHomeInfo>(`/api/v1/store/${merchantId}/home`)
}

/**
 * 获取店铺商品列表
 */
export function getStoreProducts(merchantId: number, params?: { category_id?: number }) {
  return get<StoreProductGroup[]>(`/api/v1/store/${merchantId}/products`, params)
}

/**
 * 获取店铺商品详情
 */
export function getStoreProduct(merchantId: number, productId: number) {
  return get<Product>(`/api/v1/store/${merchantId}/products/${productId}`)
}

/**
 * 获取配送费规则
 */
export function getStoreDeliveryRules(merchantId: number) {
  return get<{
    enabled: boolean
    base_fee: number
    free_delivery_amount: number
    max_distance: number
    rules: { min_distance: number; max_distance: number; fee: number }[]
  }>(`/api/v1/store/${merchantId}/delivery-rules`)
}

// ============ C端用户订单相关 ============

/**
 * 创建订单
 */
export function createOrder(merchantId: number, data: CreateOrderRequest) {
  return post<CreateOrderResponse>(`/api/v1/store/${merchantId}/orders`, data)
}

/**
 * 获取我的订单列表
 */
export function getMyOrders(params?: {
  page?: number
  page_size?: number
  merchant_id?: number
  status?: number
}) {
  return get<OrderListResponse>('/api/v1/user/orders', params)
}

/**
 * 获取我的订单详情
 */
export function getMyOrder(orderId: number) {
  return get<Order>(`/api/v1/user/orders/${orderId}`)
}

/**
 * 取消订单
 */
export function cancelMyOrder(orderId: number) {
  return post<null>(`/api/v1/user/orders/${orderId}/cancel`)
}

/**
 * 申请退款
 */
export function applyRefund(orderId: number, data: { refund_reason: string }) {
  return post<{ refund_id: number; refund_no: string; status: string }>(
    `/api/v1/user/orders/${orderId}/refund`,
    data
  )
}

// ============ 文件上传相关 ============

/**
 * 获取上传凭证
 */
export function getUploadToken() {
  return get<{ token: string; domain: string }>('/api/v1/upload/token')
}

/**
 * 上传图片
 */
export function uploadImage(filePath: string) {
  return upload<{ url: string }>('/api/v1/upload/image', filePath, 'image')
}

export default {
  // 认证
  merchantLogin,
  getMerchantProfile,
  merchantRegister,
  getMerchantApplicationStatus,
  // 商家设置
  updateMerchantProfile,
  getMerchantSettings,
  updateMerchantSettings,
  getDeliverySettings,
  updateDeliverySettings,
  updateMerchantStatus,
  getMerchantQrcode,
  // 商品分类
  getCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  sortCategories,
  // 商品管理
  getProducts,
  getProduct,
  createProduct,
  updateProduct,
  productOnSale,
  productOffSale,
  batchUpdateProductStatus,
  deleteProduct,
  updateProductStock,
  // 订单管理
  getOrders,
  getOrder,
  completeOrder,
  refundOrder,
  getOrderStatistics,
  // 数据分析
  getSalesOverview,
  getSalesTrend,
  getProductRanking,
  getHourlyAnalysis,
  getStockAlert,
  // 邀请入驻
  generateInviteCode,
  getMyInviteInfo,
  getInviteRecords,
  // C端店铺
  getStoreHome,
  getStoreProducts,
  getStoreProduct,
  getStoreDeliveryRules,
  // C端订单
  createOrder,
  getMyOrders,
  getMyOrder,
  cancelMyOrder,
  applyRefund,
  // 文件上传
  getUploadToken,
  uploadImage
}
