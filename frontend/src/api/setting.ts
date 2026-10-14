import { request } from './request'
import type { UserSetting } from './types'

export function getSetting() {
  return request<UserSetting>('/setting/info')
}

// 后端按字段白名单校验，仅传需要变更的字段
export function updateSetting(patch: Partial<UserSetting>) {
  return request<UserSetting>('/setting/update', { method: 'POST', data: patch })
}
