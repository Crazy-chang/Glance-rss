import Taro from '@tarojs/taro'
import type { ArticleListItem } from '@/api/types'

// 文章状态在详情页变更后广播，列表页据此本地更新（不重新请求）
export const ARTICLE_STATE_EVENT = 'articleStateChanged'

export type ArticleStatePatch = Partial<Pick<ArticleListItem, 'is_read' | 'is_favorited'>>

export function emitArticleState(id: number, patch: ArticleStatePatch) {
  Taro.eventCenter.trigger(ARTICLE_STATE_EVENT, { id, patch })
}

export function onArticleState(handler: (payload: { id: number; patch: ArticleStatePatch }) => void) {
  Taro.eventCenter.on(ARTICLE_STATE_EVENT, handler)
  return () => {
    Taro.eventCenter.off(ARTICLE_STATE_EVENT, handler)
  }
}
