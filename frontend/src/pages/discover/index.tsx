import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import classnames from 'classnames'
import { useCallback, useEffect, useState } from 'react'
import { getDiscoverList, subscribeRecommend, unsubscribeRecommend } from '@/api/discover'
import type { DiscoverResult, DiscoverSource } from '@/api/types'
import EmptyState from '@/components/EmptyState'
import Spinner from '@/components/Spinner'
import styles from './index.module.scss'

// 把指定推荐源在结果中标记为已订阅/未订阅
function patchSubscribed(result: DiscoverResult, id: number, subscribed: boolean): DiscoverResult {
  const patch = (s: DiscoverSource): DiscoverSource => (s.id === id ? { ...s, subscribed } : s)
  return {
    hot: result.hot.map(patch),
    categories: result.categories.map((c) => ({ ...c, sources: c.sources.map(patch) }))
  }
}

export default function DiscoverPage() {
  const [result, setResult] = useState<DiscoverResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [subscribing, setSubscribing] = useState<Record<number, boolean>>({})

  const load = useCallback(async () => {
    try {
      setResult(await getDiscoverList())
    } catch (err) {
      console.error('[Discover] 加载失败', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const onSubscribe = async (item: DiscoverSource) => {
    if (item.subscribed || subscribing[item.id]) return
    setSubscribing((prev) => ({ ...prev, [item.id]: true }))
    try {
      await subscribeRecommend(item.id)
      setResult((prev) => (prev ? patchSubscribed(prev, item.id, true) : prev))
      Taro.showToast({ title: '订阅成功', icon: 'success' })
    } catch {
      // request 层已 toast
    } finally {
      setSubscribing((prev) => {
        const next = { ...prev }
        delete next[item.id]
        return next
      })
    }
  }

  // 已订阅再次点击：二次确认后取消订阅（后端会级联删除该源文章）
  const onUnsubscribe = async (item: DiscoverSource) => {
    if (subscribing[item.id]) return
    const { confirm } = await Taro.showModal({
      title: '删除订阅',
      content: `确定取消订阅「${item.title}」吗？该源及其文章将被删除。`,
      confirmText: '删除订阅',
      cancelText: '再想想',
      confirmColor: '#FA5151'
    })
    if (!confirm) return
    setSubscribing((prev) => ({ ...prev, [item.id]: true }))
    try {
      await unsubscribeRecommend(item.id)
      setResult((prev) => (prev ? patchSubscribed(prev, item.id, false) : prev))
      Taro.showToast({ title: '已取消订阅', icon: 'success' })
    } catch {
      // request 层已 toast
    } finally {
      setSubscribing((prev) => {
        const next = { ...prev }
        delete next[item.id]
        return next
      })
    }
  }

  const empty = !result || result.categories.length === 0

  return (
    <View className={styles.page}>
      {loading && !result ? (
        <View className={styles.firstLoading}>
          <Spinner />
        </View>
      ) : empty ? (
        <EmptyState text="暂无推荐内容" />
      ) : (
        <View className={styles.body}>
          {result!.categories.map((cat) => (
            <CategoryBlock
              key={cat.name}
              category={cat}
              subscribing={subscribing}
              onSubscribe={onSubscribe}
              onUnsubscribe={onUnsubscribe}
            />
          ))}
        </View>
      )}
    </View>
  )
}

interface SubscribeBtnProps {
  item: DiscoverSource
  loading: boolean
  onSubscribe: (item: DiscoverSource) => void
  onUnsubscribe: (item: DiscoverSource) => void
}

function SubscribeButton({ item, loading, onSubscribe, onUnsubscribe }: SubscribeBtnProps) {
  if (item.subscribed) {
    return (
      <Text
        className={classnames(styles.subBtnDone, loading && styles.subBtnLoading)}
        onClick={(e) => {
          e.stopPropagation()
          onUnsubscribe(item)
        }}
      >
        {loading ? '处理中' : '已订阅'}
      </Text>
    )
  }
  return (
    <Text
      className={classnames(styles.subBtn, loading && styles.subBtnLoading)}
      onClick={(e) => {
        e.stopPropagation()
        onSubscribe(item)
      }}
    >
      {loading ? '订阅中' : '订阅'}
    </Text>
  )
}

interface CategoryBlockProps {
  category: { name: string; sources: DiscoverSource[] }
  subscribing: Record<number, boolean>
  onSubscribe: (item: DiscoverSource) => void
  onUnsubscribe: (item: DiscoverSource) => void
}

function CategoryBlock({ category, subscribing, onSubscribe, onUnsubscribe }: CategoryBlockProps) {
  const [expanded, setExpanded] = useState(false)
  return (
    <View className={styles.catBlock}>
      <View
        className={classnames(styles.catHeader, expanded && styles.catHeaderSticky)}
        hoverClass={styles.catHeaderHover}
        onClick={() => setExpanded((v) => !v)}
      >
        <Text className={styles.catName}>{category.name}</Text>
        <Text className={styles.catCount}>{category.sources.length}</Text>
        <View
          className={classnames(styles.catArrow, expanded && styles.catArrowExpanded)}
        >
          <Text className={styles.catArrowIcon}>{'›'}</Text>
        </View>
      </View>
      {expanded && (
        <View className={styles.catList}>
          {category.sources.map((item) => (
            <View key={item.id} className={styles.catItem}>
              <View className={styles.catItemMain}>
                <Text className={styles.catItemTitle}>{item.title}</Text>
                {!!item.description && (
                  <Text className={styles.catItemDesc}>{item.description}</Text>
                )}
              </View>
              <SubscribeButton
                item={item}
                loading={!!subscribing[item.id]}
                onSubscribe={onSubscribe}
                onUnsubscribe={onUnsubscribe}
              />
            </View>
          ))}
        </View>
      )}
    </View>
  )
}
