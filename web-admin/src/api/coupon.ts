import request, { unwrapApiResponse } from '@/utils/request'

/* ============ 优惠券管理（PRD V2.0 阶段二） ============ */

export interface CouponTemplate {
  id: number
  name: string
  type: number // 1=满减券 2=折扣券
  threshold_amount: number
  discount_amount: number
  discount_rate: number
  total_count: number
  received_count: number
  per_user_limit: number
  valid_type: number // 1=固定期限 2=领取后N天
  valid_start_at: string | null
  valid_end_at: string | null
  valid_days: number | null
  apply_scope: number // 1=全场 2=指定分类 3=指定商品
  scope_ids: number[] | null
  status: number
  remark: string
  created_at: string
  updated_at: string
}

export interface CouponTemplateListResponse {
  list: CouponTemplate[]
  total: number
  page: number
  page_size: number
}

export interface CouponTemplatePayload {
  name: string
  type: number
  threshold_amount: number
  discount_amount: number
  discount_rate: number
  total_count: number
  per_user_limit: number
  valid_type: number
  valid_start_at: string | null
  valid_end_at: string | null
  valid_days: number
  apply_scope: number
  scope_ids: number[] | null
  remark: string
}

export interface UserCoupon {
  id: number
  user_id: number
  template_id: number
  status: number // 1=未使用 2=已使用 3=已过期 4=已作废
  source: number // 1=自主领取 2=系统发放 3=手动发放
  received_at: string
  expired_at: string
  used_at: string | null
  order_id: number | null
  order_no: string
  template_name?: string
  template_type?: number
  threshold_amount?: number
  discount_amount?: number
  discount_rate?: number
  nickname?: string
  openid?: string
}

export interface UserCouponListResponse {
  list: UserCoupon[]
  total: number
  page: number
  page_size: number
}

export function getCouponTemplates(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/coupon-templates', { params }).then(unwrapApiResponse<CouponTemplateListResponse>)
}

export function getCouponTemplate(id: number) {
  return request.get(`/api/v1/merchant/coupon-templates/${id}`).then(unwrapApiResponse<CouponTemplate>)
}

export function createCouponTemplate(data: CouponTemplatePayload) {
  return request.post('/api/v1/merchant/coupon-templates', data).then(unwrapApiResponse<CouponTemplate>)
}

export function updateCouponTemplate(id: number, data: CouponTemplatePayload) {
  return request.put(`/api/v1/merchant/coupon-templates/${id}`, data).then(unwrapApiResponse<CouponTemplate>)
}

export function updateCouponTemplateStatus(id: number, status: number) {
  return request.post(`/api/v1/merchant/coupon-templates/${id}/status`, { status }).then(unwrapApiResponse<CouponTemplate>)
}

export function deleteCouponTemplate(id: number) {
  return request.delete(`/api/v1/merchant/coupon-templates/${id}`).then(unwrapApiResponse<{ id: number }>)
}

export function grantCoupon(id: number, userId: number) {
  return request.post(`/api/v1/merchant/coupon-templates/${id}/grant`, { user_id: userId }).then(unwrapApiResponse<{ user_coupon_id: number }>)
}

export function getUserCoupons(params?: Record<string, unknown>) {
  return request.get('/api/v1/merchant/user-coupons', { params }).then(unwrapApiResponse<UserCouponListResponse>)
}
