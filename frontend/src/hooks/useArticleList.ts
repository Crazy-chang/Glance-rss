import { useCallback, useEffect, useRef, useState } from 'react'
import Taro, { useDidShow, usePullDownRefresh, useReachBottom } from '@tarojs/taro'
import { getArticleList } from '@/api/article'
import { getSourceList } from '@/api/source'
import { refreshStart } from '@/api/common'
import type { ArticleListItem, FilterType } from '@/api/types'
import { onArticleState } from '@/utils/articleEvents'
import { syncNavContextIds } from '@/utils/navContext'

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// 后端定时抓取间隔固定 1 小时，前端静默刷新节奏与此保持一致
const SILENT_REFRESH_MS = 60 * 60 * 1000

interface Options {
  /** 限定单个订阅源（订阅文章页），不传为全部订阅 */
  sourceId?: number
  /** 空列表时是否探测"完全没有订阅源"（文章 Tab 用于引导去发现页） */
  withSourceProbe?: boolean
}

export function useArticleList(options: Options = {}) {
  const { sourceId = 0, withSourceProbe = false } = options

  const [list, setList] = useState<ArticleListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [filter, setFilter] = useState<FilterType>('unread')
  const [searchInput, setSearchInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [hasMore, setHasMore] = useState(false)
  const [total, setTotal] = useState(0)
  const [noSource, setNoSource] = useState(false)

  // 固定分页大小（与 max_items 无关——max_items 是后端每源抓取截断数）
  const PAGE_SIZE = 50

  // 静默刷新间隔（ref 存，避免触发重渲染）
  const silentDelayRef = useRef(SILENT_REFRESH_MS)

  // 回调里读取最新值的 refs
  const listRef = useRef<ArticleListItem[]>([])
  const pageRef = useRef(1)
  const hasMoreRef = useRef(false)
  const loadingMoreRef = useRef(false)
  const totalRef = useRef(0)
  const paramsRef = useRef<{ filter: FilterType; keyword: string }>({ filter: 'unread', keyword: '' })
  const reqSeqRef = useRef(0)
  const firstShowRef = useRef(true)
  const mountedRef = useRef(true)
  const silentTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // 分页页码只走 ref（不触发渲染）
  const applyList = useCallback((items: ArticleListItem[], page: number, more: boolean, count: number) => {
    listRef.current = items
    setList(items)
    pageRef.current = page
    hasMoreRef.current = more
    totalRef.current = count
    setHasMore(more)
    setTotal(count)
  }, [])

  const probeSourceEmpty = useCallback(
    async (seq: number) => {
      if (!withSourceProbe) return
      try {
        const sources = await getSourceList()
        if (seq === reqSeqRef.current) setNoSource(sources.list.length === 0)
      } catch (err) {
        console.error('[ArticleList] 订阅源探测失败', err)
      }
    },
    [withSourceProbe]
  )

  const replaceLoad = useCallback(
    async (showLoading: boolean): Promise<boolean> => {
      if (showLoading) setLoading(true)
      const seq = ++reqSeqRef.current
      try {
        const res = await getArticleList({
          page: 1,
          size: PAGE_SIZE,
          filter: paramsRef.current.filter,
          search: paramsRef.current.keyword,
          source_id: sourceId || undefined
        })
        if (seq !== reqSeqRef.current) return true
        applyList(res.list, 1, res.hasMore, res.total)
        const isEmptyAll =
          res.list.length === 0 && paramsRef.current.filter === 'all' && !paramsRef.current.keyword
        if (isEmptyAll) {
          probeSourceEmpty(seq)
        } else {
          setNoSource(false)
        }
        return true
      } catch (err) {
        console.error('[ArticleList] 首页加载失败', err)
        return false
      } finally {
        if (seq === reqSeqRef.current) setLoading(false)
      }
    },
    [sourceId, applyList, probeSourceEmpty]
  )

  const loadMore = useCallback(async () => {
    if (loadingMoreRef.current || !hasMoreRef.current) return
    loadingMoreRef.current = true
    setLoadingMore(true)
    const seq = ++reqSeqRef.current
    const targetPage = pageRef.current + 1
    try {
      const res = await getArticleList({
        page: targetPage,
        size: PAGE_SIZE,
        filter: paramsRef.current.filter,
        search: paramsRef.current.keyword,
        source_id: sourceId || undefined
      })
      if (seq !== reqSeqRef.current) return
      const existed = new Set(listRef.current.map((a) => a.id))
      const merged = [...listRef.current]
      res.list.forEach((item) => {
        if (!existed.has(item.id)) merged.push(item)
      })
      applyList(merged, targetPage, res.hasMore, res.total)
    } catch (err) {
      console.error('[ArticleList] 加载更多失败', err)
    } finally {
      if (seq === reqSeqRef.current) {
        loadingMoreRef.current = false
        setLoadingMore(false)
      }
    }
  }, [sourceId, applyList])

  const changeFilter = useCallback(
    (next: FilterType) => {
      if (next === paramsRef.current.filter) return
      paramsRef.current.filter = next
      setFilter(next)
      replaceLoad(true)
    },
    [replaceLoad]
  )

  const applyKeyword = useCallback(
    (kw: string) => {
      if (kw === paramsRef.current.keyword) return
      paramsRef.current.keyword = kw
      setKeyword(kw)
      replaceLoad(true)
    },
    [replaceLoad]
  )

  const onSearchInput = useCallback(
    (value: string) => {
      setSearchInput(value)
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
      searchTimerRef.current = setTimeout(() => applyKeyword(value.trim()), 400)
    },
    [applyKeyword]
  )

  const clearSearch = useCallback(() => {
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    setSearchInput('')
    applyKeyword('')
  }, [applyKeyword])

  // 首屏加载 + 卸载清理
  useEffect(() => {
    replaceLoad(true)
    return () => {
      mountedRef.current = false
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
      if (silentTimerRef.current) clearTimeout(silentTimerRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 详情页状态变更广播 -> 本地更新（不 refetch）
  useEffect(
    () =>
      onArticleState(({ id, patch }) => {
        const prev = listRef.current
        if (!prev.some((a) => a.id === id)) return
        const f = paramsRef.current.filter
        const shouldRemove =
          (patch.is_read === true && f === 'unread') ||
          (patch.is_favorited === false && f === 'favorited')
        if (shouldRemove) {
          const next = prev.filter((a) => a.id !== id)
          listRef.current = next
          setList(next)
          syncNavContextIds(next.map((a) => a.id))
          setTotal((t) => Math.max(0, t - 1))
        } else {
          const next = prev.map((a) => (a.id === id ? { ...a, ...patch } : a))
          listRef.current = next
          setList(next)
        }
      }),
    []
  )

  // 下拉刷新：立即收起原生下拉动画，自绘转圈；轮询首页直到有更新（最多 6s）
  usePullDownRefresh(async () => {
    const startedAt = Date.now()
    setRefreshing(true)
    // 原生下拉动画会和列表回弹抢节奏，立刻收掉，改用自己的转圈
    Taro.stopPullDownRefresh()
    try {
      await refreshStart()
    } catch (err) {
      console.error('[ArticleList] refresh/start 失败', err)
    }
    // 后端抓取是异步的，固定等待容易"白等"：改为轮询，总数变化即提前结束
    const beforeTotal = totalRef.current
    const deadline = Date.now() + 6000
    try {
      while (Date.now() < deadline) {
        await sleep(800)
        const ok = await replaceLoad(false)
        // 请求失败（网络/服务异常）不再继续轮询，避免反复打同一个接口
        if (!ok) break
        if (totalRef.current !== beforeTotal) break
      }
    } finally {
      // 至少展示 600ms，避免提示一闪而过
      const remain = 600 - (Date.now() - startedAt)
      if (remain > 0) await sleep(remain)
      setRefreshing(false)
    }
  })

  // 切回 Tab / 返回页面：后台触发刷新，固定间隔后静默替换"第一页且非搜索态"
  useDidShow(() => {
    if (firstShowRef.current) {
      firstShowRef.current = false
      return
    }
    const delay = silentDelayRef.current
    refreshStart().catch((err) => console.error('[ArticleList] 静默刷新触发失败', err))
    silentTimerRef.current = setTimeout(() => {
      if (
        mountedRef.current &&
        pageRef.current === 1 &&
        !paramsRef.current.keyword &&
        !loadingMoreRef.current
      ) {
        replaceLoad(false)
      }
    }, delay)
  })

  useReachBottom(() => {
    loadMore()
  })

  // 构造详情页本地序列上下文（含列表查询参数，用于详情页末尾续拉下一页）
  const buildNavContext = useCallback(
    () => ({
      ids: listRef.current.map((a) => a.id),
      page: pageRef.current,
      hasMore: hasMoreRef.current,
      total: totalRef.current,
      query: {
        filter: paramsRef.current.filter,
        keyword: paramsRef.current.keyword,
        sourceId: sourceId || undefined
      }
    }),
    [sourceId]
  )

  return {
    list,
    loading,
    loadingMore,
    refreshing,
    filter,
    keyword,
    searchInput,
    hasMore,
    total,
    noSource,
    changeFilter,
    onSearchInput,
    clearSearch,
    buildNavContext
  }
}
