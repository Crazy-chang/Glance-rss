import { useCallback, useEffect, useRef, useState } from 'react'
import Taro, { useDidShow, usePullDownRefresh, useReachBottom } from '@tarojs/taro'
import { getArticleList } from '@/api/article'
import { getSourceList } from '@/api/source'
import { refreshStart } from '@/api/common'
import { useSetting } from '@/store/setting'
import type { ArticleListItem, FilterType } from '@/api/types'
import { onArticleState } from '@/utils/articleEvents'
import { syncNavContextIds } from '@/utils/navContext'

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

/** refresh_interval 字符串 → 毫秒，manual 表示不静默刷新（返回 0 时跳过），非法值回退 30 分钟 */
function intervalToMs(interval?: string): number {
  if (interval === 'manual') return 0
  const m = /^(\d+)([smh])$/.exec(interval || '')
  if (!m) return 30 * 60 * 1000
  const num = Number(m[1])
  const unit = m[2]
  if (unit === 's') return num * 1000
  if (unit === 'm') return num * 60 * 1000
  if (unit === 'h') return num * 60 * 60 * 1000
  return 30 * 60 * 1000
}

interface Options {
  /** 限定单个订阅源（订阅文章页），不传为全部订阅 */
  sourceId?: number
  /** 空列表时是否探测"完全没有订阅源"（文章 Tab 用于引导去发现页） */
  withSourceProbe?: boolean
}

export function useArticleList(options: Options = {}) {
  const { sourceId = 0, withSourceProbe = false } = options
  const { setting } = useSetting()

  const [list, setList] = useState<ArticleListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [filter, setFilter] = useState<FilterType>('unread')
  const [searchInput, setSearchInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [hasMore, setHasMore] = useState(false)
  const [total, setTotal] = useState(0)
  const [noSource, setNoSource] = useState(false)

  // 固定分页大小（与 max_items 无关——max_items 是后端每源抓取截断数）
  const PAGE_SIZE = 50

  // 从用户设置读取的动态值（ref 存，避免触发重渲染；变化时通过单独 useEffect 同步）
  const silentDelayRef = useRef(intervalToMs(setting?.refresh_interval))

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
    async (showLoading: boolean) => {
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
        if (seq !== reqSeqRef.current) return
        applyList(res.list, 1, res.hasMore, res.total)
        const isEmptyAll =
          res.list.length === 0 && paramsRef.current.filter === 'all' && !paramsRef.current.keyword
        if (isEmptyAll) {
          probeSourceEmpty(seq)
        } else {
          setNoSource(false)
        }
      } catch (err) {
        console.error('[ArticleList] 首页加载失败', err)
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

  // 用户设置变化：refresh_interval 实时生效；sort_order 后端从 setting 读，变化后触发 reload
  useEffect(() => {
    silentDelayRef.current = intervalToMs(setting?.refresh_interval)
  }, [setting?.refresh_interval])

  useEffect(() => {
    // sort_order 后端从用户 setting 读，不需要前端传参；设置变化后重拉
    replaceLoad(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setting?.sort_order])

  // max_items 变化不触发前端 reload——它是后端抓取截断数，改完等下次 refresh/start 生效即可

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

  // 下拉刷新：触发后端抓取 -> 等待 -> 重拉第一页
  usePullDownRefresh(async () => {
    try {
      await refreshStart()
    } catch (err) {
      console.error('[ArticleList] refresh/start 失败', err)
    }
    await sleep(2500)
    await replaceLoad(false)
    Taro.stopPullDownRefresh()
  })

  // 切回 Tab / 返回页面：后台触发刷新，间隔由用户设置决定，"第一页且非搜索态"静默替换
  useDidShow(() => {
    if (firstShowRef.current) {
      firstShowRef.current = false
      return
    }
    const delay = silentDelayRef.current
    if (delay === 0) return // manual 模式：不静默刷新
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
