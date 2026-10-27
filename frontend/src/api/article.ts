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
