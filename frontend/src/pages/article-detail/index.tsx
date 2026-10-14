import { Button, RichText, Text, View } from '@tarojs/components'
import Taro, { useShareAppMessage, useReady, useRouter } from '@tarojs/taro'
import classnames from 'classnames'
import { useCallback, useEffect, useRef, useState } from 'react'
import { BASE_URL } from '@/api/config'
import {
  favorite,
  getArticleDetail,
  getArticleList,
  markRead,
  unfavorite
} from '@/api/article'
import { getToken } from '@/api/request'
import type { ArticleDetail as ArticleDetailType } from '@/api/types'
import { useSetting } from '@/store/setting'
import { emitArticleState } from '@/utils/articleEvents'
import {
  appendNavPage,
  getNavContext,
  getNavNeighbors,
  setNavAnchor
} from '@/utils/navContext'
import { formatRelative } from '@/utils/time'
import Spinner from '@/components/Spinner'
import styles from './index.module.scss'

/**
 * 把正文中的 img-proxy URL 补全为带 token 的绝对地址，并给所有 img 标签加
 * max-width:100%;height:auto;display:block; 防止超宽图片横向溢出。
 * <Image> 标签无法携带 Authorization 头，必须走 query 鉴权；
 * rich-text 内部节点不受外层 WXSS 作用域影响，必须用 inline style 控制图片宽度。
 */
function rewriteContentImages(content: string): string {
  if (!content) return ''
  let html = content

  // 1) 给所有 img 标签补 max-width:100%;height:auto;display:block;
  html = html.replace(/<img\b([^>]*?)(\s*\/?)>/gi, (match, attrs: string, slash: string) => {
    const styleMatch = attrs.match(/style\s*=\s*["']([^"']*)["']/i)
    const rule = 'max-width:100%;height:auto;display:block;'
    if (styleMatch) {
      if (styleMatch[1].includes('max-width')) return match
      const merged = `${rule}${styleMatch[1]}`
      const newAttrs = attrs.replace(/style\s*=\s*["'][^"']*["']/i, `style="${merged}"`)
      return `<img${newAttrs}${slash}>`
    }
    return `<img${attrs} style="${rule}"${slash}>`
  })

  // 2) 给 img-proxy URL 补 token（<Image> 无法带 Authorization 头）
  const token = getToken()
  if (!token) return html
  return html.replace(
    /(src=["'])([^"']*\/api\/img-proxy\?url=[^"']*)(["'])/gi,
    (match, prefix: string, url: string, suffix: string) => {
      if (url.includes('token=')) return match
      const absolute = url.startsWith('http') ? url : `${BASE_URL}${url}`
      return `${prefix}${absolute}&token=${encodeURIComponent(token)}${suffix}`
    }
  )
}

export default function ArticleDetailPage() {
  const router = useRouter()
  const { setting } = useSetting()
  const theme = setting?.theme || 'white'
  const fontSize = setting?.font_size || 'medium'

  const [currentId, setCurrentId] = useState(() => Number(router.params.id) || 0)
  const [detail, setDetail] = useState<ArticleDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [faving, setFaving] = useState(false)
  const [navVersion, setNavVersion] = useState(0)
  const [loadingMoreNav, setLoadingMoreNav] = useState(false)
  const [statusBarHeight, setStatusBarHeight] = useState(20)
  const reqSeqRef = useRef(0)

  // 取状态栏高度，用于自定义导航栏的顶部留白
  useReady(() => {
    try {
      const info = Taro.getSystemInfoSync()
      setStatusBarHeight(info.statusBarHeight || 20)
    } catch (err) {
      console.error('[ArticleDetail] getSystemInfoSync 失败', err)
    }
  })

  const loadDetail = useCallback(
    async (id: number) => {
      if (!id) return
      setLoading(true)
      setError('')
      const seq = ++reqSeqRef.current
      try {
        const data = await getArticleDetail(id)
        if (seq !== reqSeqRef.current) return
        setDetail(data)
        setNavAnchor(id)
        // 打开自动标记已读并广播给列表做本地更新
        if (!data.is_read) {
          markRead(id).catch((e) => console.error('[ArticleDetail] 标记已读失败', e))
          const patched = { ...data, is_read: true }
          setDetail(patched)
          emitArticleState(id, { is_read: true })
        }
      } catch (err: any) {
        if (seq !== reqSeqRef.current) return
        setError(err?.message || '加载失败')
      } finally {
        if (seq === reqSeqRef.current) setLoading(false)
      }
    },
    []
  )

  useEffect(() => {
    loadDetail(currentId)
  }, [currentId, loadDetail])

  // 详情页末尾自动拉下一页：当无下一篇且列表仍有 hasMore 时
  const loadNextPageForNav = useCallback(async () => {
    const navCtx = getNavContext()
    if (!navCtx.hasMore || loadingMoreNav) return
    setLoadingMoreNav(true)
    const targetPage = navCtx.loadedPage + 1
    try {
      const res = await getArticleList({
        page: targetPage,
        size: 50,
        filter: navCtx.query.filter,
        search: navCtx.query.keyword,
        source_id: navCtx.query.sourceId
      })
      appendNavPage(
        res.list.map((a) => a.id),
        targetPage,
        res.hasMore,
        res.total
      )
      setNavVersion((v) => v + 1)
    } catch (err) {
      console.error('[ArticleDetail] 续拉下一页失败', err)
    } finally {
      setLoadingMoreNav(false)
    }
  }, [loadingMoreNav])

  const neighbors = getNavNeighbors(currentId)
  useEffect(() => {
    if (!neighbors.nextId && !loading && !loadingMoreNav) {
      const navCtx = getNavContext()
      if (navCtx.hasMore) {
        loadNextPageForNav()
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [neighbors.nextId, loading, loadingMoreNav, navVersion, loadNextPageForNav])

  const scrollTop = () => Taro.pageScrollTo({ scrollTop: 0, duration: 0 })

  const onPrev = () => {
    if (!neighbors.prevId) return
    setCurrentId(neighbors.prevId)
    scrollTop()
  }
  const onNext = () => {
    if (!neighbors.nextId) {
      loadNextPageForNav()
      return
    }
    setCurrentId(neighbors.nextId)
    scrollTop()
  }

  const onToggleFavorite = async () => {
    if (!detail || faving) return
    setFaving(true)
    const next = !detail.is_favorited
    const prev = detail
    setDetail({ ...detail, is_favorited: next })
    emitArticleState(detail.id, { is_favorited: next })
    try {
      if (next) await favorite(detail.id)
      else await unfavorite(detail.id)
    } catch (err) {
      console.error('[ArticleDetail] 收藏切换失败', err)
      setDetail(prev)
      emitArticleState(detail.id, { is_favorited: prev.is_favorited })
    } finally {
      setFaving(false)
    }
  }

  const goBack = () => Taro.navigateBack()

  useShareAppMessage(() => {
    if (!detail) return { title: 'Glance 阅读', path: '/pages/article/index' }
    return {
      title: detail.title,
      path: `/pages/article-detail/index?id=${detail.id}`
    }
  })

  const navCtx = getNavContext()
  const hasPrev = !!neighbors.prevId
  const hasNext = !!neighbors.nextId || navCtx.hasMore
  const contentHtml = rewriteContentImages(detail?.content || detail?.summary || '')

  return (
    <View
      className={`${styles.page} theme-${theme} fs-${fontSize}`}
      style={{ '--status-bar-h': `${statusBarHeight}px` } as any}
    >
      {/* 自定义导航栏 */}
      <View className={styles.navBar}>
        <View className={styles.navContent}>
          <View className={styles.backBtn} onClick={goBack}>
            <Text className={styles.backIcon}>‹</Text>
          </View>
          <View className={styles.navTitle}>
            <Text className={styles.navTitleText}>
              {detail?.title || '加载中…'}
            </Text>
          </View>
          <View className={styles.navPlaceholder} />
        </View>
      </View>

      {/* 主体 */}
      <View className={styles.body}>
        {loading && !detail ? (
          <View className={styles.loading}>
            <Spinner size={36} />
          </View>
        ) : error ? (
          <View className={styles.errorBox}>
            <Text className={styles.errorText}>{error}</Text>
            <View className={styles.retryBtn} onClick={() => loadDetail(currentId)}>
              重试
            </View>
          </View>
        ) : detail ? (
          <>
            <View className={styles.header}>
              <Text className={styles.title}>{detail.title}</Text>
              <View className={styles.metaRow}>
                <Text className={styles.author}>
                  {[detail.author, detail.source_name].filter(Boolean).join(' · ')}
                </Text>
                <Text className={styles.published}>
                  {formatRelative(detail.published_at)}
                </Text>
              </View>
            </View>
            <View className={styles.content}>
              {contentHtml ? (
                <RichText nodes={contentHtml} />
              ) : (
                <Text className={styles.empty}>暂无正文</Text>
              )}
            </View>
          </>
        ) : null}
      </View>

      {/* 底部操作栏 */}
      <View className={styles.bottomBar}>
        <View
          className={classnames(styles.actionBtn, { [styles.actionDisabled]: !hasPrev })}
          onClick={onPrev}
        >
          上一篇
        </View>
        <View
          className={classnames(styles.actionBtn, styles.favoriteBtn, {
            [styles.actionActive]: detail?.is_favorited
          })}
          onClick={onToggleFavorite}
        >
          {detail?.is_favorited ? '已收藏' : '收藏'}
        </View>
        <Button
          className={classnames(styles.actionBtn, styles.shareBtn)}
          openType="share"
        >
          分享
        </Button>
        <View
          className={classnames(styles.actionBtn, {
            [styles.actionDisabled]: !hasNext,
            [styles.actionLoading]: loadingMoreNav
          })}
          onClick={onNext}
        >
          {loadingMoreNav ? '加载中' : '下一篇'}
        </View>
      </View>
    </View>
  )
}
