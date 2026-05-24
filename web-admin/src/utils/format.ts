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
