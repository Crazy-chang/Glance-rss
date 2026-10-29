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
import type { ArticleDetail as ArticleDetailType, UserSetting } from '@/api/types'
import { useSetting } from '@/store/setting'
import {
  ACCENT,
  FONT_SIZES,
  LINE_HEIGHTS,
  TEXT_ALIGNS,
  THEME_CHROME,
  THEME_KEYS
} from '@/utils/theme'
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
 * max-width:100%;height:auto;display:block; 防止超宽图片横向溢出，
 * 同时加 loading="lazy" decoding="async" 让未滚到的图片延迟加载、异步解码。
 * <Image> 标签无法携带 Authorization 头，必须走 query 鉴权；
 * rich-text 内部节点不受外层 WXSS 作用域影响，必须用 inline style 控制图片宽度。
 */
function rewriteContentImages(content: string): string {
  if (!content) return ''
  let html = content

  // 1) 给所有 img 标签补 max-width + 懒加载属性
  html = html.replace(/<img\b([^>]*?)(\s*\/?)>/gi, (_match, attrs: string, slash: string) => {
    const styleMatch = attrs.match(/style\s*=\s*["']([^"']*)["']/i)
    const rule = 'max-width:100%;height:auto;display:block;'
    let newAttrs = attrs
    if (styleMatch) {
      if (!styleMatch[1].includes('max-width')) {
        const merged = `${rule}${styleMatch[1]}`
        newAttrs = attrs.replace(/style\s*=\s*["'][^"']*["']/i, `style="${merged}"`)
      }
    } else {
      newAttrs = `${attrs} style="${rule}"`
    }
    // 懒加载：已有 loading 属性则不覆盖
    if (!/\bloading\s*=/i.test(newAttrs)) {
      newAttrs += ' loading="lazy" decoding="async"'
    }
    return `<img${newAttrs}${slash}>`
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

// ========== 底部操作栏图标（内联 SVG data URI） ==========
// 小程序 WXML 不支持 svg 标签，WXSS background-image 对 SVG data URI 兼容最好；
// 颜色随主题在 JS 侧按 hex 传入，避免 currentColor 在背景图中失效。
type ActionIconName = 'settings' | 'bookmark' | 'share' | 'prev' | 'next'

const ICON_BODY: Record<ActionIconName, (color: string, filled?: boolean) => string> = {
  settings: (c) =>
    `<g fill="none" stroke="${c}" stroke-width="1.8" stroke-linecap="round">` +
    `<path d="M6 11h28M6 20h28M6 29h28"/>` +
    `<circle cx="15" cy="11" r="3.4" fill="${c}" stroke="none"/>` +
    `<circle cx="26" cy="20" r="3.4" fill="${c}" stroke="none"/>` +
    `<circle cx="18" cy="29" r="3.4" fill="${c}" stroke="none"/></g>`,
  bookmark: (c, filled) =>
    `<path d="M10 7h20a2 2 0 0 1 2 2v25l-12-7-12 7V9a2 2 0 0 1 2-2z" ` +
    (filled
      ? `fill="${c}"/>`
      : `fill="none" stroke="${c}" stroke-width="1.8" stroke-linejoin="round"/>`),
  share: (c) =>
    `<g fill="none" stroke="${c}" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">` +
    `<path d="M12 24v9h16v-9"/><path d="M20 26V6"/><path d="M14 12l6-6 6 6"/></g>`,
  prev: (c) =>
    `<path d="M24 8L13 20l11 12" fill="none" stroke="${c}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`,
  next: (c) =>
    `<path d="M16 8l11 12-11 12" fill="none" stroke="${c}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`
}

function actionIcon(name: ActionIconName, color: string, filled = false): string {
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 40">` +
    ICON_BODY[name](color, filled) +
    `</svg>`
  return `url("data:image/svg+xml,${encodeURIComponent(svg)}")`
}

export default function ArticleDetailPage() {
  const router = useRouter()
  // 分享卡片进入时路径带 share=1&sig=xxx：底部操作栏整体隐藏，只能阅读
  const isShareMode = router.params.share === '1'
  const shareSig = router.params.sig || ''
  const { setting, updateSetting } = useSetting()
  const theme = setting?.theme || 'white'
  const fontSize = setting?.font_size || 'medium'
  const lineHeight = setting?.line_height || 'normal'
  const textAlign = setting?.text_align || 'left'
  // 图标颜色跟随主题（SVG data URI 无法使用 CSS 变量）
  const iconColor = theme === 'black' ? '#F2F3F5' : '#1C1C1E'
  const [panelOpen, setPanelOpen] = useState(false)

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
        const data = isShareMode
          ? await getArticleDetail(id, { share: '1', sig: shareSig })
          : await getArticleDetail(id)
        if (seq !== reqSeqRef.current) return
        setDetail(data)
        setNavAnchor(id)
        // 打开自动标记已读并广播给列表做本地更新（分享模式无个人状态，跳过）
        if (!isShareMode && !data.is_read) {
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
    [isShareMode, shareSig]
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

  /** 阅读设置：主题 / 字号（即时生效，失败由 request 层提示并回滚） */
  const changeSetting = async (patch: Partial<UserSetting>) => {
    try {
      await updateSetting(patch)
    } catch (err) {
      console.error('[ArticleDetail] 更新阅读设置失败', err)
    }
  }

  const goBack = () => {
    if (Taro.getCurrentPages().length > 1) {
      Taro.navigateBack()
    } else {
      Taro.switchTab({ url: '/pages/index/index' })
    }
  }

  useShareAppMessage(() => {
    if (!detail) return { title: 'Glance 阅读', path: '/pages/index/index' }
    return {
      title: detail.title,
      path: `/pages/article-detail/index?id=${detail.id}&share=1&sig=${detail.share_sig || ''}`
    }
  })

  const navCtx = getNavContext()
  const hasPrev = !!neighbors.prevId
  const hasNext = !!neighbors.nextId || navCtx.hasMore
  const contentHtml = rewriteContentImages(detail?.content || detail?.summary || '')

  return (
    <View
      className={`${styles.page} theme-${theme} fs-${fontSize} lh-${lineHeight} align-${textAlign}`}
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

      {/* 主体：分享模式无底部操作栏，收紧底部留白 */}
      <View className={classnames(styles.body, { [styles.bodyShare]: isShareMode })}>
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

      {/* 底部操作栏：分享模式整体隐藏，只能阅读 */}
      {!isShareMode && (
        <View
          className={classnames(styles.bottomBar, { [styles.barDisabled]: !!error })}
        >
          <View className={styles.actionBtn} onClick={() => setPanelOpen(true)}>
            <View
              className={styles.actionIcon}
              style={{ backgroundImage: actionIcon('settings', panelOpen ? ACCENT : iconColor) }}
            />
            <Text
              className={classnames(styles.actionLabel, {
                [styles.actionLabelActive]: panelOpen
              })}
            >
              设置
            </Text>
          </View>
          <View className={styles.actionBtn} onClick={onToggleFavorite}>
            <View
              className={styles.actionIcon}
              style={{
                backgroundImage: actionIcon(
                  'bookmark',
                  detail?.is_favorited ? ACCENT : iconColor,
                  !!detail?.is_favorited
                )
              }}
            />
            <Text
              className={classnames(styles.actionLabel, {
                [styles.actionLabelActive]: detail?.is_favorited
              })}
            >
              {detail?.is_favorited ? '已收藏' : '收藏'}
            </Text>
          </View>
          <Button className={classnames(styles.shareBtn)} openType="share">
            <View
              className={styles.actionIcon}
              style={{ backgroundImage: actionIcon('share', iconColor) }}
            />
            <Text className={styles.actionLabel}>分享</Text>
          </Button>
          <View
            className={classnames(styles.actionBtn, { [styles.actionDisabled]: !hasPrev })}
            onClick={onPrev}
          >
            <View
              className={styles.actionIcon}
              style={{ backgroundImage: actionIcon('prev', iconColor) }}
            />
            <Text className={styles.actionLabel}>上一篇</Text>
          </View>
          <View
            className={classnames(styles.actionBtn, {
              [styles.actionDisabled]: !hasNext || loadingMoreNav,
              [styles.actionLoading]: loadingMoreNav
            })}
            onClick={onNext}
          >
            <View
              className={styles.actionIcon}
              style={{ backgroundImage: actionIcon('next', iconColor) }}
            />
            <Text className={styles.actionLabel}>{loadingMoreNav ? '加载中' : '下一篇'}</Text>
          </View>
        </View>
      )}

      {/* 阅读设置面板：主题 + 字号，点击蒙层关闭 */}
      {panelOpen && (
        <View className={styles.mask} onClick={() => setPanelOpen(false)}>
          <View className={styles.sheet} onClick={(e) => e.stopPropagation()}>
            <View className={styles.sheetSection}>
              <Text className={styles.sheetLabel}>主题</Text>
              <View className={styles.themeRow}>
                {THEME_KEYS.map((key) => (
                  <View
                    key={key}
                    className={styles.themeItem}
                    onClick={() => changeSetting({ theme: key })}
                  >
                    <View
                      className={classnames(styles.themeSwatch, {
                        [styles.themeSwatchActive]: theme === key
                      })}
                      style={{ background: THEME_CHROME[key].swatch }}
                    />
                    <Text
                      className={classnames(styles.themeName, {
                        [styles.themeNameActive]: theme === key
                      })}
                    >
                      {THEME_CHROME[key].name}
                    </Text>
                  </View>
                ))}
              </View>
            </View>

            <View className={styles.sheetSection}>
              <Text className={styles.sheetLabel}>字号</Text>
              <View className={styles.segment}>
                {FONT_SIZES.map((item) => (
                  <View
                    key={item.key}
                    className={classnames(styles.segmentItem, {
                      [styles.segmentItemActive]: fontSize === item.key
                    })}
                    onClick={() => changeSetting({ font_size: item.key })}
                  >
                    {item.name}
                  </View>
                ))}
              </View>
            </View>

            <View className={styles.sheetSection}>
              <Text className={styles.sheetLabel}>行距</Text>
              <View className={styles.segment}>
                {LINE_HEIGHTS.map((item) => (
                  <View
                    key={item.key}
                    className={classnames(styles.segmentItem, {
                      [styles.segmentItemActive]: lineHeight === item.key
                    })}
                    onClick={() => changeSetting({ line_height: item.key })}
                  >
                    {item.name}
                  </View>
                ))}
              </View>
            </View>

            <View className={styles.sheetSection}>
              <Text className={styles.sheetLabel}>对齐</Text>
              <View className={styles.segment}>
                {TEXT_ALIGNS.map((item) => (
                  <View
                    key={item.key}
                    className={classnames(styles.segmentItem, {
                      [styles.segmentItemActive]: textAlign === item.key
                    })}
                    onClick={() => changeSetting({ text_align: item.key })}
                  >
                    {item.name}
                  </View>
                ))}
              </View>
            </View>

            <View className={styles.sheetDone} onClick={() => setPanelOpen(false)}>
              完成
            </View>
          </View>
        </View>
      )}
    </View>
  )
}
