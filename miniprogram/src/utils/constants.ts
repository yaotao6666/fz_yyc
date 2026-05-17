/**
 * 常量定义
 */

// 存储 Key
export const StorageKey = {
  TOKEN: 'token',
  USER_INFO: 'userInfo',
  MERCHANT_INFO: 'merchantInfo'
} as const

// 订单状态
export const OrderStatus = {
  PENDING_PAYMENT: 1,   // 待支付
  PAID: 2,              // 已支付
  COMPLETED: 3,         // 已完成
  CANCELLED: 4,         // 已取消
  REFUNDING: 5,         // 退款中
  REFUNDED: 6           // 已退款
} as const

export const OrderStatusText: Record<number, string> = {
  [OrderStatus.PENDING_PAYMENT]: '待支付',
  [OrderStatus.PAID]: '已支付',
  [OrderStatus.COMPLETED]: '已完成',
  [OrderStatus.CANCELLED]: '已取消',
  [OrderStatus.REFUNDING]: '退款中',
  [OrderStatus.REFUNDED]: '已退款'
}

// 商品状态
export const ProductStatus = {
  OFF_SALE: 0,  // 下架
  ON_SALE: 1    // 上架
} as const

// 配送类型
export const DeliveryType = {
  DELIVERY: 1,  // 配送
  DINE_IN: 2,   // 堂食
  PICKUP: 3     // 自提
} as const

// 邀请奖励类型
export const RewardType = {
  FREE_YEAR: 'free_year',     // 免年费
  LOWEST_RATE: 'lowest_rate'  // 最低费率
} as const

// 邀请记录状态
export const InviteStatus = {
  PENDING: 0,     // 待完成
  COMPLETED: 1,   // 已完成
  CANCELLED: 2    // 已取消
} as const

export const BrandAsset = {
  APP_LOGO: '/static/brand-icons/xunmeng-private-butler/icon-circle-aggregated-portal.svg',
  DEFAULT_MERCHANT_LOGO: '/static/brand-icons/xunmeng-private-butler/icon-circle-store-data-butler.svg',
  DEFAULT_PRODUCT_IMAGE: '/static/logo.png'
} as const

export default {
  StorageKey,
  OrderStatus,
  OrderStatusText,
  ProductStatus,
  DeliveryType,
  RewardType,
  InviteStatus,
  BrandAsset
}
