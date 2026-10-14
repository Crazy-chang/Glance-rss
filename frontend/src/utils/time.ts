/**
 * 相对时间格式化：
 * - 1 小时内：x 分钟前（不足 1 分钟显示"刚刚"）
 * - 24 小时内：x 小时前
 * - 7 天内：x 天前
 * - 今年内：MM-DD
 * - 跨年：YYYY-MM-DD
 * @param unixSeconds 秒级时间戳
 * @param nowSec 当前时间（秒），默认取本机时间，测试可显式传入
 */
export function formatRelative(unixSeconds: number, nowSec: number = Date.now() / 1000): string {
  const diff = Math.max(0, nowSec - unixSeconds)

  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  if (diff < 7 * 86400) return `${Math.floor(diff / 86400)}天前`

  const d = new Date(unixSeconds * 1000)
  const now = new Date(nowSec * 1000)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')

  if (d.getFullYear() === now.getFullYear()) {
    return `${mm}-${dd}`
  }
  return `${d.getFullYear()}-${mm}-${dd}`
}
