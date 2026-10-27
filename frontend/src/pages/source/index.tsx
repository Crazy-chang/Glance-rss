import { Button, Image, Input, Text, View } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
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
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [menu, setMenu] = useState<{ item: SourceItem } | null>(null)
  const [renameState, setRenameState] = useState<{ id: number; value: string; submitting: boolean } | null>(null)
  const shareItemRef = useRef<SourceItem | null>(null)

  // 分享当前操作的订阅源：分享到该源的文章列表页
  Taro.useShareAppMessage(() => {
    const item = shareItemRef.current
    return {
      title: item ? `推荐订阅：${item.title}` : 'Glance',
      path: item
        ? `/pages/source-article/index?id=${item.id}&name=${encodeURIComponent(item.title)}`
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

  const load = useCallback(async () => {
    try {
      const res = await getSourceList()
      setList(res.list)
    } catch (err) {
      console.error('[Source] 加载失败', err)
    } finally {
      setLoading(false)
    }
  }, [])

  // 每次进入（含从推荐订阅页返回）都静默刷新列表
  useDidShow(() => {
    load()
  })

  usePullDownRefresh(async () => {
    try {
      await refreshStart()
    } catch (err) {
      console.error('[Source] refresh/start 失败', err)
    }
    await sleep(1500)
    await load()
    Taro.stopPullDownRefresh()
  })

  const goArticles = (item: SourceItem) => {
    Taro.navigateTo({
      url: `/pages/source-article/index?id=${item.id}&name=${encodeURIComponent(item.title)}`
    })
  }

  const openAdd = () => {
    setFeedUrl('')
    setDialogVisible(true)
  }

  const onConfirmAdd = async () => {
    const url = feedUrl.trim()
    if (!url) {
      Taro.showToast({ title: '请输入 RSS 地址', icon: 'none' })
      return
    }
    setSubmitting(true)
    try {
      await addSource(url)
      setDialogVisible(false)
      Taro.showToast({ title: '添加成功', icon: 'success' })
      await load()
    } catch {
      // request 层已 toast（RSS 校验失败、重复订阅等）
    } finally {
      setSubmitting(false)
    }
  }

  const onDelete = async (item: SourceItem) => {
    const { confirm } = await Taro.showModal({
      title: '删除订阅',
      content: `确定删除「${item.title}」吗？该操作将删除其下所有文章，不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
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
            text="还没有订阅源"
            actionText="推荐订阅源"
            onAction={() => Taro.navigateTo({ url: '/pages/discover/index' })}
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
        onChange={setFeedUrl}
        onCancel={() => setDialogVisible(false)}
        onConfirm={onConfirmAdd}
      />
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
  onChange: (v: string) => void
  onCancel: () => void
  onConfirm: () => void
}

function AddDialog({ visible, value, submitting, onChange, onCancel, onConfirm }: AddDialogProps) {
  if (!visible) return null
  return (
    <View className={styles.mask} onClick={onCancel} catchMove>
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
        <View className={styles.dialogActions}>
          <Text className={styles.dialogBtn} onClick={onCancel}>取消</Text>
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
