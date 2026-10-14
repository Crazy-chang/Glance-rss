/**
 * 详情页"上一篇/下一篇"的本地序列上下文（不使用服务端 prev/next）。
 * 列表点击时写入当前有序 id 序列；列表项被本地移除（如未读筛选下标记已读）
 * 后同步更新 ids，使被移除文章的邻居仍然正确。
 *
 * 同时携带列表的查询参数 / 已加载页码 / hasMore，使详情页在翻到
 * 序列末尾时能自行调用 getArticleList 拉取下一页，无需依赖列表页存活。
 */

import type { FilterType } from '@/api/types'

interface ListQuery {
  filter: FilterType
  keyword: string
  sourceId?: number
}

interface NavContext {
  ids: number[]
  anchorIndex: number
  query: ListQuery
  loadedPage: number
  hasMore: boolean
  total: number
}

let ctx: NavContext = {
  ids: [],
  anchorIndex: -1,
  query: { filter: 'all', keyword: '' },
  loadedPage: 1,
  hasMore: false,
  total: 0
}

export interface SetNavContextParams {
  ids: number[]
  anchorId: number
  query: ListQuery
  page: number
  hasMore: boolean
  total: number
}

/** 进入详情：记录序列、入口文章位置与列表查询上下文 */
export function setNavContext(p: SetNavContextParams) {
  ctx = {
    ids: [...p.ids],
    anchorIndex: p.ids.indexOf(p.anchorId),
    query: { ...p.query },
    loadedPage: p.page,
    hasMore: p.hasMore,
    total: p.total
  }
}

/** 获取完整上下文（详情页用于续拉下一页） */
export function getNavContext(): NavContext {
  return ctx
}

/** 列表本地变化后同步序列（anchorIndex 与查询上下文保持不变） */
export function syncNavContextIds(ids: number[]) {
  ctx.ids = [...ids]
}

/** 详情页跳到上/下一篇时更新锚点位置 */
export function setNavAnchor(articleId: number) {
  const idx = ctx.ids.indexOf(articleId)
  if (idx >= 0) ctx.anchorIndex = idx
}

/**
 * 当前文章的邻居。文章已不在序列中（如未读列表里被标记已读）时
 * 退回使用锚点位置，保证"下一篇"跳过自身。
 */
export function getNavNeighbors(currentId: number): { prevId?: number; nextId?: number } {
  const { ids, anchorIndex } = ctx
  let idx = ids.indexOf(currentId)
  if (idx < 0) idx = anchorIndex
  return {
    prevId: idx > 0 ? ids[idx - 1] : undefined,
    nextId: idx >= 0 && idx < ids.length - 1 ? ids[idx + 1] : undefined
  }
}

/** 详情页末尾自动拉下一页后追加新 ids，并更新页码/hasMore/total */
export function appendNavPage(newIds: number[], page: number, hasMore: boolean, total: number) {
  const existed = new Set(ctx.ids)
  newIds.forEach((id) => {
    if (!existed.has(id)) {
      ctx.ids.push(id)
      existed.add(id)
    }
  })
  ctx.loadedPage = page
  ctx.hasMore = hasMore
  ctx.total = total
}
