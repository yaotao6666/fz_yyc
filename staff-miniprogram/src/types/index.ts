/**
 * 服务人员端全局类型定义（第0期占位，第2期按真实后端扩展）
 */

// 订单类型（与 orders.order_type 对齐）
export enum OrderType {
  /** 零售 */ Retail = 1,
  /** 租赁 */ Rental = 2,
  /** 康养上门 */ KangyangVisit = 3,
  /** 陪诊 */ Peizhen = 4,
  /** 科普体验 */ ScienceActivity = 5,
  /** 长护险服务 */ LongTermCare = 6
}

export const OrderTypeText: Record<OrderType, string> = {
  [OrderType.Retail]: '零售',
  [OrderType.Rental]: '租赁归还',
  [OrderType.KangyangVisit]: '康养上门',
  [OrderType.Peizhen]: '陪诊服务',
  [OrderType.ScienceActivity]: '科普体验',
  [OrderType.LongTermCare]: '长护险服务'
}

// 工单业务状态（biz_status，与后端 orders.biz_status 对齐）
export enum WorkorderBizStatus {
  /** 无（零售等无需派工的订单） */ None = 0,
  /** 待接单 */ Pending = 1,
  /** 已接单-待出发 */ Accepted = 2,
  /** 服务中-已签到 */ InService = 3,
  /** 待支付尾款 */ PendingPayment = 4,
  /** 已完成 */ Completed = 5,
  /** 已取消 */ Cancelled = 6
}

export interface StaffUser {
  id: number
  merchant_id: number
  name: string
  phone: string
  avatar?: string
  role?: string
  status?: number
  openid?: string
  created_at?: string
}

export interface WorkorderItem {
  id: number
  order_id: number
  order_no: string
  order_type: OrderType
  biz_status: WorkorderBizStatus
  /** 预约时间 */
  scheduled_at?: string
  /** 实际签到/签退 */
  actual_started_at?: string
  actual_ended_at?: string
  /** 客户信息 */
  contact_name?: string
  contact_phone?: string
  delivery_address?: string
  lat?: number
  lng?: number
  /** 服务内容快照 */
  service_content?: any
  remark?: string
}

export interface ScheduleDay {
  date: string // YYYY-MM-DD
  type: 'work' | 'rest' | 'duty' | 'leave'
  shift?: 'morning' | 'afternoon' | 'night' | 'full'
  note?: string
}
