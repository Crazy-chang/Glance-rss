import { request } from './request'
import type { SourceItem, SourceListResp } from './types'

export function getSourceList() {
  return request<SourceListResp>('/source/list')
}

// 手动添加订阅源：校验 RSS -> 查重 -> 创建 -> 异步首抓
export function addSource(feedUrl: string) {
  return request<SourceItem>('/source/add', { method: 'POST', data: { feed_url: feedUrl } })
}

// 删除订阅源：事务级联删除文章与状态
export function deleteSource(id: number) {
  return request(`/source/delete/${id}`, { method: 'POST' })
}

// 重命名订阅源标题
export function renameSource(id: number, title: string) {
  return request(`/source/rename/${id}`, { method: 'POST', data: { title } })
}
