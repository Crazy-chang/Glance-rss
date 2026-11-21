import { request } from './request'
import type { ArticleDetail, ArticleListItem, FilterType, PageResult } from './types'

// 每页条数，后端固定上限 50
export const PAGE_SIZE = 50

export interface ArticleListQuery {
  page?: number
  size?: number
  filter?: FilterType
  search?: string
  source_id?: number
}

export function getArticleList(query: ArticleListQuery = {}) {
  return request<PageResult<ArticleListItem>>('/article/list', { query })
}

export function getArticleDetail(id: number, query?: { share?: string; sig?: string }) {
  return request<ArticleDetail>(`/article/detail/${id}`, { query })
}

export function markRead(id: number) {
  return request<{ id: number; is_read: boolean }>(`/article/read/${id}`, { method: 'POST', silent: true })
}

export function markUnread(id: number) {
  return request<{ id: number; is_read: boolean }>(`/article/unread/${id}`, { method: 'POST', silent: true })
}

export function favorite(id: number) {
  return request<{ id: number; is_favorited: boolean }>(`/article/favorite/${id}`, { method: 'POST', silent: true })
}

export function unfavorite(id: number) {
  return request<{ id: number; is_favorited: boolean }>(`/article/unfavorite/${id}`, {
    method: 'POST',
    silent: true
  })
}

// 清理已读且未收藏的文章，返回删除数量
export function clearReadArticles() {
  return request<{ deleted: number }>('/article/clear-read', { method: 'POST' })
}

// 当前用户资料（一键阅读开关与每次条数）
export function getUserProfile() {
  return request<{ batch_read_enabled: boolean; batch_read_limit: number }>('/user/profile', {
    silent: true
  })
}

// 一键阅读：按后台配置批量把未读标记为已读，返回实际条数
export function readBatch() {
  return request<{ count: number }>('/article/read-batch', { method: 'POST' })
}
