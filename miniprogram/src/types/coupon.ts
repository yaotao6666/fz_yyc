/* ============ C端优惠券类型（PRD V2.0 阶段二：优惠券营销闭环） ============ */

/** 券模板（C端展示） */
export interface CouponTemplate {
  id: number
  name: string
  type: number // 1=满减券 2=折扣券
  threshold_amount: number
  discount_amount: number
  discount_rate: number
  total_count: number
  received_count: number
  remain_count: number // -1 表示不限量
  per_user_limit: number
  valid_type: number // 1=固定期限 2=领取后N天
  valid_start_at?: string | null
  valid_end_at?: string | null
  valid_days?: number | null
  my_received_count: number
  can_receive: boolean
}

/** 我的用户券 */
export interface MyCoupon {
  id: number
  template_id: number
  status: number // 1=未使用 2=已使用 3=已过期 4=已作废
  source: number // 1=自主领取 2=系统发放 3=手动发放
  received_at: string
  expired_at: string
  used_at?: string | null
  order_no?: string
  name?: string
  type?: number
  threshold_amount?: number
  discount_amount?: number
  discount_rate?: number
}

/** 下单可用券预览 */
export interface OrderUsableCoupon {
  user_coupon_id: number
  template_id: number
  name: string
  type: number // 1=满减券 2=折扣券
  threshold_amount: number
  discount_amount: number
  discount_rate: number
  discount: number // 本单预计抵扣金额
  expired_at: string
}
