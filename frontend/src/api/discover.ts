import { request } from './request'
import type { DiscoverResult, SourceItem } from './types'

// 发现页：热门订阅 + 分类推荐
export function getDiscoverList() {
  return request<DiscoverResult>('/discover/list')
}

// 订阅推荐源：返回创建后的订阅源信息
export function subscribeRecommend(id: number) {
  return request<SourceItem>(`/discover/subscribe/${id}`, { method: 'POST' })
}
