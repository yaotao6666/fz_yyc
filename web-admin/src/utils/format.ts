export function formatAmount(value = 0): string {
  return Number(value || 0).toFixed(2)
}

export function formatPercent(value = 0): string {
  return `${Number(value || 0).toFixed(2)}%`
}

export function formatDateTime(value?: string): string {
  if (!value) return '-'
  return value.replace('T', ' ').slice(0, 19)
}

export function formatDate(value?: string): string {
  if (!value) return '-'
  return value.slice(0, 10)
}

export function getMerchantStatusText(status: number): string {
  return {
    1: '营业中',
    2: '休息中',
    3: '已关闭'
  }[status] || '未知状态'
}

export function getPaymentConfigText(status?: number): string {
  return Number(status || 0) === 1 ? '已完成' : '待完善'
}

export function getProductStatusText(status?: number): string {
  return Number(status || 0) === 1 ? '上架' : '下架'
}

export function getCategoryStatusText(status?: number): string {
  return Number(status || 0) === 1 ? '启用' : '停用'
}

export function getRentalUnitText(unit?: number): string {
  return { 1: '天', 2: '周', 3: '月' }[Number(unit || 0)] || ''
}

export function getDepositStatusText(status?: number): string {
  return {
    0: '无押金',
    1: '已收',
    2: '已退',
    3: '部分扣除'
  }[Number(status || 0)] || '未知'
}

// 商品类型：1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务
export const ProductTypeMap: Record<number, { text: string; tagType: string }> = {
  1: { text: '辅具零售', tagType: 'success' },
  2: { text: '辅具租赁', tagType: 'warning' },
  3: { text: '康养套餐', tagType: 'danger' },
  4: { text: '陪诊服务', tagType: 'primary' }
}

export function getProductTypeText(type?: number): string {
  return ProductTypeMap[Number(type || 1)]?.text || '未知类型'
}

export function getProductTypeTagType(type?: number): string {
  return ProductTypeMap[Number(type || 1)]?.tagType || 'info'
}

// Tab 筛选配置
export interface ProductTypeTab {
  key: string
  label: string
  product_type?: number
}

export const ProductTypeTabs: ProductTypeTab[] = [
  { key: 'all', label: '全部' },
  { key: 'retail', label: '辅具零售', product_type: 1 },
  { key: 'rental', label: '辅具租赁', product_type: 2 },
  { key: 'wellness', label: '康养套餐', product_type: 3 },
  { key: 'escort', label: '陪诊服务', product_type: 4 }
]
