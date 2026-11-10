import { Button, Image, Input, Text, View } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh, useRouter } from '@tarojs/taro'
import classnames from 'classnames'
import { useCallback, useEffect, useRef, useState } from 'react'
import { BASE_URL } from '@/api/config'
import { getToken } from '@/api/request'
import { deleteSource, getSourceList, addSource, renameSource } from '@/api/source'
import { refreshStart } from '@/api/common'
import type { SourceItem } from '@/api/types'
import EmptyState from '@/components/EmptyState'
import Spinner from '@/components/Spinner'
import { useHeaderMetrics } from '@/hooks/useHeaderMetrics'
import styles from './index.module.scss'

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// feed 地址规整：忽略大小写与尾部斜杠差异（与后端匹配口径一致）
const normalizeFeed = (url: string) => url.trim().toLowerCase().replace(/\/+$/, '')

// 给后端图片接口 URL 追加 token（<Image> 无法携带 Authorization 头）
function authImageUrl(url: string): string {
  if (!url) return ''
  const absolute = url.startsWith('http') ? url : `${BASE_URL}${url}`
  const token = getToken()
  if (!token) return absolute
  return absolute.includes('?')
    ? `${absolute}&token=${encodeURIComponent(token)}`
    : `${absolute}?token=${encodeURIComponent(token)}`
}

export default function SourcePage() {
  const { statusBarHeight } = useHeaderMetrics()
  const [list, setList] = useState<SourceItem[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogVisible, setDialogVisible] = useState(false)
  const [feedUrl, setFeedUrl] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [addError, setAddError] = useState('')
  const [refreshing, setRefreshing] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [menu, setMenu] = useState<{ item: SourceItem } | null>(null)
  const [renameState, setRenameState] = useState<{ id: number; value: string; submitting: boolean } | null>(null)
  const shareItemRef = useRef<SourceItem | null>(null)

  // 分享当前操作的订阅源：带 feed_url 供接收方匹配/一键订阅（id 仅分享者本人账号有效）
  Taro.useShareAppMessage(() => {
    const item = shareItemRef.current
    return {
      title: item ? `推荐订阅：${item.title}` : 'RSS订阅阅读器',
      path: item
        ? `/pages/source/index?share=1&name=${encodeURIComponent(
            item.title
          )}&feed_url=${encodeURIComponent(item.feed_url)}`
        : '/pages/index/index'
    }
  })

  // 任一弹窗打开时隐藏原生 tabBar（原生 tabBar 在 webview 之上，z-index 无法遮盖）；
  // 关闭时恢复。React 批量更新保证「关菜单→开重命名」不会闪烁。
  const anyModalOpen = !!menu || !!renameState || dialogVisible
  useEffect(() => {
    if (anyModalOpen) {
      Taro.hideTabBar().catch(() => {})
    } else {
      Taro.showTabBar().catch(() => {})
    }
    return () => {
      Taro.showTabBar().catch(() => {})
    }
  }, [anyModalOpen])

  // 新源首抓由后端异步执行：last_fetch_at 仍为 0 说明还在抓取中。
  // 轮询到抓取结束（成功或失败）再刷新，避免列表文章数一直显示 0。
  // 请求失败立即终止；重新进入页面会开启新一轮并让旧轮询退出，避免请求叠加。
  const pollSeq = useRef(0)
  const pollPendingSources = useCallback(async (items: SourceItem[]) => {
    if (!items.some((s) => s.last_fetch_at === 0)) return
    const seq = ++pollSeq.current
    for (let i = 0; i < 6; i++) {
      await sleep(1000)
      if (pollSeq.current !== seq) return
      try {
        const latest = await getSourceList()
        if (pollSeq.current !== seq) return
        setList(latest.list)
        if (!latest.list.some((s) => s.last_fetch_at === 0)) return
      } catch (err) {
        console.error('[Source] 首抓轮询失败', err)
        return
      }
    }
  }, [])

  const load = useCallback(async () => {
    try {
      const res = await getSourceList()
      setList(res.list)
      void pollPendingSources(res.list)
    } catch (err) {
      console.error('[Source] 加载失败', err)
    } finally {
      setLoading(false)
    }
  }, [pollPendingSources])

  // 每次进入（含从推荐订阅页返回）都静默刷新列表
  useDidShow(() => {
    load()
  })

  // 分享落地（share=1&feed_url=xxx）：已订阅则停留当前页；
  // 未订阅则在本页弹窗引导添加，成功后刷新列表
  const router = useRouter()
  const shareHandledRef = useRef(false)
  useEffect(() => {
    if (shareHandledRef.current) return
    const { share, feed_url, name } = router.params
    if (share !== '1' || !feed_url) return
    shareHandledRef.current = true
    const feed = decodeURIComponent(feed_url)
    const shareName = name ? decodeURIComponent(name) : ''
    void handleShareLanding(feed, shareName)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleShareLanding = async (feed: string, shareName: string) => {
    try {
      const { list: sources } = await getSourceList()
      const target = normalizeFeed(feed)
      if (sources.some((s) => normalizeFeed(s.feed_url) === target)) return
    } catch {
      // 列表查询失败也继续引导，添加接口会兜底报错
    }
    const { confirm } = await Taro.showModal({
      title: '推荐订阅',
      content: `好友向你推荐了订阅源《${shareName}》，添加后可在订阅页持续查看更新。`,
      confirmText: '添加订阅',
      cancelText: '暂不'
    })
    if (confirm) void addSharedFeed(feed)
  }

  const addSharedFeed = async (feed: string) => {
    Taro.showLoading({ title: '订阅中...', mask: true }).catch(() => {})
    try {
      await addSource(feed)
      Taro.hideLoading()
      Taro.showToast({ title: '订阅成功', icon: 'success' })
      void load()
    } catch (err: any) {
      Taro.hideLoading()
      const { confirm } = await Taro.showModal({
        title: '订阅失败',
        content: err?.message || '请稍后重试',
        confirmText: '重试',
        cancelText: '取消'
      })
      if (confirm) void addSharedFeed(feed)
    }
  }

  usePullDownRefresh(async () => {
    const startedAt = Date.now()
    setRefreshing(true)
    // 原生下拉动画会和列表回弹抢节奏，立刻收掉，改用自己的转圈
    Taro.stopPullDownRefresh()
    try {
      await refreshStart()
    } catch (err) {
      console.error('[Source] refresh/start 失败', err)
    }
    await sleep(800)
    try {
      await load()
    } finally {
      const remain = 600 - (Date.now() - startedAt)
      if (remain > 0) await sleep(remain)
      setRefreshing(false)
    }
  })

  const goArticles = (item: SourceItem) => {
    Taro.navigateTo({
      url: `/pages/source-article/index?id=${item.id}&name=${encodeURIComponent(item.title)}`
    })
  }

  const openAdd = () => {
    setFeedUrl('')
    setAddError('')
    setDialogVisible(true)
  }

  const onConfirmAdd = async () => {
    const url = feedUrl.trim()
    if (!url) {
      Taro.showToast({ title: '请输入 RSS 地址', icon: 'none' })
      return
    }
    setAddError('')
    setSubmitting(true)
    try {
      await addSource(url)
      setDialogVisible(false)
      Taro.showToast({ title: '添加成功', icon: 'success' })
      await load()
    } catch (err: any) {
      // 失败原因留在弹窗内红色展示，输入内容保留
      setAddError(err?.message || '添加失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  const onDelete = async (item: SourceItem) => {
    const { confirm } = await Taro.showModal({
      title: '删除订阅',
      content: `确定取消订阅「${item.title}」吗？该源及其文章将被删除。`,
      confirmText: '删除订阅',
      cancelText: '再想想',
      confirmColor: '#FA5151'
    })
    if (!confirm) return
    setDeletingId(item.id)
    try {
      await deleteSource(item.id)
      setList((prev) => prev.filter((s) => s.id !== item.id))
      Taro.showToast({ title: '已删除', icon: 'success' })
    } catch {
      // request 层已 toast
    } finally {
      setDeletingId(null)
    }
  }

  // 打开底部操作菜单
  const openMenu = (item: SourceItem, e: any) => {
    e.stopPropagation()
    setMenu({ item })
  }

  const closeMenu = () => setMenu(null)

  const copyFeed = (item: SourceItem) => {
    Taro.setClipboardData({ data: item.feed_url })
    Taro.showToast({ title: '源地址已复制', icon: 'none' })
  }

  const openRename = (item: SourceItem) => {
    setRenameState({ id: item.id, value: item.title, submitting: false })
  }

  const closeRename = () => setRenameState(null)

  const confirmRename = async () => {
    if (!renameState) return
    const title = renameState.value.trim()
    if (!title) {
      Taro.showToast({ title: '请输入名称', icon: 'none' })
      return
    }
    setRenameState({ ...renameState, submitting: true })
    try {
      await renameSource(renameState.id, title)
      setList((prev) =>
        prev.map((s) => (s.id === renameState.id ? { ...s, title } : s))
      )
      closeRename()
      Taro.showToast({ title: '已重命名', icon: 'success' })
    } catch {
      // request 层已 toast
    } finally {
      setRenameState((s) => (s ? { ...s, submitting: false } : null))
    }
  }

  return (
    <View
      className={styles.page}
      style={{ '--status-bar-h': `${statusBarHeight}px` } as any}
    >
      {/* 固定头部：标题行与胶囊齐平 + 操作按钮行，白底铺满状态栏区域 */}
      <View className={styles.header}>
        <View className={styles.titleRow}>
          <Text className={styles.title}>订阅</Text>
          {!loading && <Text className={styles.count}>（{list.length}）</Text>}
        </View>
        <View className={styles.topBar}>
          <View className={styles.addBtn} hoverClass={styles.addBtnHover} onClick={openAdd}>
            <Text>添加订阅</Text>
          </View>
          <View
            className={styles.recommendBtn}
            hoverClass={styles.recommendBtnHover}
            onClick={() => Taro.navigateTo({ url: '/pages/discover/index' })}
          >
            <Text>推荐订阅源</Text>
          </View>
        </View>
      </View>

      <View className={styles.body}>
        {loading && list.length === 0 ? (
          <View className={styles.firstLoading}>
            <Spinner />
          </View>
        ) : list.length === 0 ? (
          <EmptyState
            text="暂无订阅源"
            subText="添加后，新文章会进入「文章」页，也会归到该订阅源下"
          />
        ) : (
          <View className={styles.list}>
            {list.map((item) => (
              <SourceRow
                key={item.id}
                item={item}
                deleting={deletingId === item.id}
                onMore={openMenu}
                onClick={goArticles}
              />
            ))}
          </View>
        )}
      </View>

      {/* 操作下拉菜单：分享 / 重命名 / 复制源 / 删除 / 取消 */}
      {menu && (
        <View className={styles.menuMask} onClick={closeMenu} catchMove>
          <View
            className={styles.menu}
            onClick={(e) => e.stopPropagation()}
          >
            <Button
              openType="share"
              className={styles.menuItem}
              onClick={() => {
                shareItemRef.current = menu.item
                closeMenu()
              }}
            >
              分享
            </Button>
            <View
              className={styles.menuItem}
              onClick={() => {
                const item = menu.item
                closeMenu()
                openRename(item)
              }}
            >
              重命名
            </View>
            <View
              className={styles.menuItem}
              onClick={() => {
                const item = menu.item
                closeMenu()
                copyFeed(item)
              }}
            >
              复制源
            </View>
            <View
              className={classnames(styles.menuItem, styles.menuItemDanger)}
              onClick={() => {
                const item = menu.item
                closeMenu()
                onDelete(item)
              }}
            >
              删除
            </View>
            <View className={styles.menuDivider} />
            <View className={styles.menuItem} onClick={closeMenu}>
              取消
            </View>
          </View>
        </View>
      )}

      {/* 重命名弹窗 */}
      {renameState && (
        <View className={styles.mask} onClick={closeRename} catchMove>
          <View className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <Text className={styles.dialogTitle}>重命名</Text>
            <Input
              className={styles.dialogInput}
              value={renameState.value}
              placeholder="输入新的名称"
              placeholderClass={styles.dialogPlaceholder}
              focus
              onInput={(e) => setRenameState({ ...renameState, value: e.detail.value })}
            />
            <View className={styles.dialogActions}>
              <Text className={styles.dialogBtn} onClick={closeRename}>
                取消
              </Text>
              <Text
                className={classnames(
                  styles.dialogBtn,
                  styles.dialogBtnPrimary,
                  renameState.submitting && styles.dialogBtnDisabled
                )}
                onClick={() => {
                  if (!renameState.submitting) confirmRename()
                }}
              >
                {renameState.submitting ? '保存中' : '确认'}
              </Text>
            </View>
          </View>
        </View>
      )}

      <AddDialog
        visible={dialogVisible}
        value={feedUrl}
        submitting={submitting}
        error={addError}
        onChange={(v) => {
          setFeedUrl(v)
          setAddError('')
        }}
        onCancel={() => setDialogVisible(false)}
        onConfirm={onConfirmAdd}
      />

      {refreshing && (
        <View className={styles.refreshTip}>
          <Spinner size={30} />
        </View>
      )}
    </View>
  )
}

interface SourceRowProps {
  item: SourceItem
  deleting: boolean
  onMore: (item: SourceItem, e: any) => void
  onClick: (item: SourceItem) => void
}

function SourceRow({ item, deleting, onMore, onClick }: SourceRowProps) {
  const faviconUrl = item.favicon_url ? authImageUrl(item.favicon_url) : ''
  return (
    <View className={styles.row} hoverClass={styles.rowHover} onClick={() => onClick(item)}>
      <View className={styles.favicon}>
        {faviconUrl ? (
          <Image className={styles.faviconImg} src={faviconUrl} mode="aspectFit" lazyLoad />
        ) : (
          <Text className={styles.faviconText}>{(item.title || '?').slice(0, 1)}</Text>
        )}
      </View>
      <View className={styles.rowMain}>
        <View className={styles.titleLine}>
          <Text className={styles.rowTitle}>{item.title}</Text>
          <View className={styles.chevron} />
        </View>
        <Text className={styles.sub}>
          {item.unread_count > 0 ? `${item.unread_count} 未读 · ` : ''}
          {item.article_count} 篇{item.category ? ` · ${item.category}` : ''}
        </Text>
      </View>
      {/* 三个竖点：打开操作菜单 */}
      <View
        className={classnames(styles.moreBtn, deleting && styles.deleteBtnLoading)}
        onClick={(e) => {
          e.stopPropagation()
          if (!deleting) onMore(item, e)
        }}
      >
        <View className={styles.moreDot} />
        <View className={styles.moreDot} />
        <View className={styles.moreDot} />
      </View>
    </View>
  )
}

interface AddDialogProps {
  visible: boolean
  value: string
  submitting: boolean
  error: string
  onChange: (v: string) => void
  onCancel: () => void
  onConfirm: () => void
}

function AddDialog({
  visible,
  value,
  submitting,
  error,
  onChange,
  onCancel,
  onConfirm
}: AddDialogProps) {
  if (!visible) return null
  const close = () => {
    if (!submitting) onCancel()
  }
  return (
    <View className={styles.mask} onClick={close} catchMove>
      <View className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <Text className={styles.dialogTitle}>添加订阅</Text>
        <Input
          className={styles.dialogInput}
          value={value}
          placeholder="输入 RSS 地址，如 https://example.com/feed"
          placeholderClass={styles.dialogPlaceholder}
          focus
          onInput={(e) => onChange(e.detail.value)}
        />
        {/* 状态区固定高度，无论显示加载还是错误都不改变弹窗高度 */}
        <View className={styles.dialogStatus}>
          {submitting && (
            <View className={styles.dialogStatusRow}>
              <Spinner size={28} />
              <Text className={styles.dialogStatusText}>正在解析订阅源…</Text>
            </View>
          )}
          {!submitting && !!error && (
            <Text className={styles.dialogStatusError}>{error}</Text>
          )}
        </View>
        <View className={classnames(styles.dialogActions, styles.dialogActionsTight)}>
          <Text
            className={classnames(
              styles.dialogBtn,
              submitting && styles.dialogBtnDisabled
            )}
            onClick={close}
          >
            取消
          </Text>
          <Text
            className={classnames(
              styles.dialogBtn,
              styles.dialogBtnPrimary,
              submitting && styles.dialogBtnDisabled
            )}
            onClick={() => {
              if (!submitting) onConfirm()
            }}
          >
            {submitting ? '添加中' : '确认'}
          </Text>
        </View>
      </View>
    </View>
  )
}
