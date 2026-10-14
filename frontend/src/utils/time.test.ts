import { describe, expect, it } from 'vitest'
import { formatRelative } from './time'

// 固定基准时间：2026-10-12 12:00:00（本地时区）
const NOW = Math.floor(new Date(2026, 9, 12, 12, 0, 0).getTime() / 1000)

describe('formatRelative', () => {
  it('1 分钟内显示刚刚', () => {
    expect(formatRelative(NOW - 30, NOW)).toBe('刚刚')
    expect(formatRelative(NOW, NOW)).toBe('刚刚')
  })

  it('未来时间按刚刚处理', () => {
    expect(formatRelative(NOW + 3600, NOW)).toBe('刚刚')
  })

  it('1 小时内显示分钟前', () => {
    expect(formatRelative(NOW - 60, NOW)).toBe('1分钟前')
    expect(formatRelative(NOW - 30 * 60, NOW)).toBe('30分钟前')
    expect(formatRelative(NOW - 59 * 60, NOW)).toBe('59分钟前')
  })

  it('24 小时内显示小时前', () => {
    expect(formatRelative(NOW - 3600, NOW)).toBe('1小时前')
    expect(formatRelative(NOW - 5 * 3600, NOW)).toBe('5小时前')
    expect(formatRelative(NOW - 23 * 3600 - 59 * 60, NOW)).toBe('23小时前')
  })

  it('7 天内显示天前', () => {
    expect(formatRelative(NOW - 86400, NOW)).toBe('1天前')
    expect(formatRelative(NOW - 3 * 86400, NOW)).toBe('3天前')
    expect(formatRelative(NOW - 6 * 86400, NOW)).toBe('6天前')
  })

  it('今年内显示 MM-DD', () => {
    const t = Math.floor(new Date(2026, 0, 5, 9, 30, 0).getTime() / 1000)
    expect(formatRelative(t, NOW)).toBe('01-05')
    const t2 = Math.floor(new Date(2026, 8, 1, 0, 0, 0).getTime() / 1000)
    expect(formatRelative(t2, NOW)).toBe('09-01')
  })

  it('跨年显示 YYYY-MM-DD', () => {
    const t = Math.floor(new Date(2025, 11, 31, 23, 59, 0).getTime() / 1000)
    expect(formatRelative(t, NOW)).toBe('2025-12-31')
    const t2 = Math.floor(new Date(2020, 2, 8, 8, 0, 0).getTime() / 1000)
    expect(formatRelative(t2, NOW)).toBe('2020-03-08')
  })
})
