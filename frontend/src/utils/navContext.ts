/**
 * 详情页"上一篇/下一篇"的本地序列上下文（不使用服务端 prev/next）。
 * 列表点击时写入当时可见顺序的 id 快照，本次详情浏览期间保持不变。
 * 注意：不能在标记已读 / 取消收藏后从 ids 中剔除条目——剔除后当前文章
 * 不在序列里，getNavNeighbors 会退回已经错位的 anchorIndex，
 * 导致"下一篇"跳过文章；保留快照才能保证翻页严格跟随点击时的可见顺序。
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

/** 详情页跳到上/下一篇时更新锚点位置 */
export function setNavAnchor(articleId: number) {
  const idx = ctx.ids.indexOf(articleId)
  if (idx >= 0) ctx.anchorIndex = idx
}

/**
 * 当前文章的邻居。文章不在序列中（非列表入口进入，如分享卡片）时
 * 退回使用锚点位置。
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
