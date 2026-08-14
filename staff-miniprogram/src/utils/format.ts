/**
 * 日期时间工具
 */
export function pad2(n: number) { return n < 10 ? '0' + n : '' + n }

export function formatDate(d: Date | string | number, pattern = 'YYYY-MM-DD HH:mm:ss'): string {
  const date = d instanceof Date ? d : new Date(d)
  if (isNaN(date.getTime())) return ''
  return pattern
    .replace('YYYY', String(date.getFullYear()))
    .replace('MM', pad2(date.getMonth() + 1))
    .replace('DD', pad2(date.getDate()))
    .replace('HH', pad2(date.getHours()))
    .replace('mm', pad2(date.getMinutes()))
    .replace('ss', pad2(date.getSeconds()))
}

export function fromNow(d: Date | string | number): string {
  const date = d instanceof Date ? d : new Date(d)
  const diff = (Date.now() - date.getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + '分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + '小时前'
  if (diff < 86400 * 7) return Math.floor(diff / 86400) + '天前'
  return formatDate(date, 'YYYY-MM-DD')
}
