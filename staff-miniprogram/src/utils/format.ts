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

/**
 * 由出生日期字符串计算周岁年龄
 * 无效日期返回空字符串，避免在页面展示 NaN
 */
export function calcAge(birthDate?: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (isNaN(birth.getTime())) return ''
  const now = new Date()
  let age = now.getFullYear() - birth.getFullYear()
  const monthDiff = now.getMonth() - birth.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && now.getDate() < birth.getDate())) {
    age -= 1
  }
  return age >= 0 ? String(age) : ''
}
