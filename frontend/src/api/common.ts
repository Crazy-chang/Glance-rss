import { request } from './request'

// 触发一次全量刷新（后端异步抓取），后台静默调用
export function refreshStart() {
  return request('/refresh/start', { method: 'POST', silent: true })
}
