import { get, post } from '../utils/request'
import type { CouponTemplate, MyCoupon, OrderUsableCoupon } from '../types/coupon'

/* ============ C端优惠券（PRD V2.0 阶段二：优惠券营销闭环） ============ */

/** 可领券列表（首页领券入口） */
export function getAvailableCoupons() {
  return get<{ list: CouponTemplate[] }>(`/api/v1/store/coupons/available`, undefined, {
    loading: false,
    showErrorToast: false
  }).then((res) => ({
    list: Array.isArray(res?.list) ? res.list : []
  }))
}

/** 领取优惠券 */
export function receiveCoupon(templateId: number) {
  return post<{ user_coupon_id: number; template_id: number; expired_at: string }>(
    `/api/v1/store/coupons/${templateId}/receive`
  )
}

/** 我的券列表（status: 1=未使用 2=已使用 3=已过期） */
export function getMyCoupons(status: number, page = 1, pageSize = 10) {
  return get<{ list: MyCoupon[]; total: number }>(`/api/v1/store/my-coupons`, {
    status,
    page,
    page_size: pageSize
  }).then((res) => ({
    list: Array.isArray(res?.list) ? res.list : [],
    total: res?.total || 0
  }))
}

/** 下单可用券预览（结算页选券） */
export function getOrderUsableCoupons(totalAmount: number, productIds: number[], categoryIds: number[] = []) {
  const params: Record<string, unknown> = {
    total_amount: totalAmount
  }
  if (productIds.length > 0) {
    params.product_ids = productIds.join(',')
  }
  if (categoryIds.length > 0) {
    params.category_ids = categoryIds.join(',')
  }
  return get<{ list: OrderUsableCoupon[] }>(`/api/v1/store/coupons/usable`, params, {
    loading: false,
    showErrorToast: false
  }).then((res) => ({
    list: Array.isArray(res?.list) ? res.list : []
  }))
}
