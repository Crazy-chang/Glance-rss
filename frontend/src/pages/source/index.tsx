import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import classnames from 'classnames'
import { useCallback, useState } from 'react'
import { BASE_URL } from '@/api/config'
import { getToken } from '@/api/request'
import { deleteSource, getSourceList, addSource } from '@/api/source'
import { refreshStart } from '@/api/common'
import type { SourceItem } from '@/api/types'
import EmptyState from '@/components/EmptyState'
import Spinner from '@/components/Spinner'
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
  const [list, setList] = useState<SourceItem[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogVisible, setDialogVisible] = useState(false)
  const [feedUrl, setFeedUrl] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)

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

  return (
    <View className={styles.page}>
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
                onDelete={onDelete}
                onClick={goArticles}
              />
            ))}
          </View>
        )}
      </View>

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
  onDelete: (item: SourceItem) => void
  onClick: (item: SourceItem) => void
}

function SourceRow({ item, deleting, onDelete, onClick }: SourceRowProps) {
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
        <Text className={styles.title}>{item.title}</Text>
        <Text className={styles.sub}>
          {item.unread_count > 0 ? `${item.unread_count} 未读 · ` : ''}
          {item.article_count} 篇{item.category ? ` · ${item.category}` : ''}
        </Text>
      </View>
      <Text
        className={classnames(styles.deleteBtn, deleting && styles.deleteBtnLoading)}
        onClick={(e) => {
          e.stopPropagation()
          if (!deleting) onDelete(item)
        }}
      >
        {deleting ? '删除中' : '删除'}
      </Text>
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
