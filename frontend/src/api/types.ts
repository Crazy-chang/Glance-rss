// 后端接口响应类型，字段与后端 JSON tag 一一对应

export interface ApiResp<T> {
  code: number
  msg: string
  data: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  size: number
  hasMore: boolean
}

export type FilterType = 'all' | 'unread' | 'read' | 'today' | 'favorited'

export interface ArticleListItem {
  id: number
  source_id: number
  source_name: string
  title: string
  summary: string
  author: string
  link: string
  published_at: number // 秒级时间戳
  is_read: boolean
  is_favorited: boolean
}

export interface ArticleDetail extends ArticleListItem {
  content: string
}

export interface UserSetting {
  theme: string
  font_size: string
  max_items: number
  refresh_interval: string
  sort_order: string
}

export interface SourceItem {
  id: number
  title: string
  feed_url: string
  site_url: string
  favicon_url: string
  category: string
  description: string
  last_fetch_at: number
  last_error: string
  article_count: number
  unread_count: number
}

export interface SourceListResp {
  list: SourceItem[]
}

// 发现页：推荐源（附带当前用户是否已订阅）
export interface DiscoverSource {
  id: number
  title: string
  feed_url: string
  site_url: string
  category: string
  description: string
  subscribed: boolean
}

export interface DiscoverCategory {
  name: string
  sources: DiscoverSource[]
}

export interface DiscoverResult {
  hot: DiscoverSource[]
  categories: DiscoverCategory[]
}
