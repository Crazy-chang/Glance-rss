import { Text, View } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import classnames from 'classnames'
import { useCallback, useEffect, useRef, useState } from 'react'
import { getDiscoverList, subscribeRecommend } from '@/api/discover'
import type { DiscoverResult, DiscoverSource } from '@/api/types'
import EmptyState from '@/components/EmptyState'
import Spinner from '@/components/Spinner'
import styles from './index.module.scss'

// 把指定推荐源在结果中标记为已订阅
function markSubscribed(result: DiscoverResult, id: number): DiscoverResult {
  const patch = (s: DiscoverSource): DiscoverSource => (s.id === id ? { ...s, subscribed: true } : s)
  return {
    hot: result.hot.map(patch),
    categories: result.categories.map((c) => ({ ...c, sources: c.sources.map(patch) }))
  }
}

export default function DiscoverPage() {
  const [result, setResult] = useState<DiscoverResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [subscribing, setSubscribing] = useState<Record<number, boolean>>({})
  const firstShowRef = useRef(true)

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

  // 切回 Tab 时静默刷新订阅状态
  useDidShow(() => {
    if (firstShowRef.current) {
      firstShowRef.current = false
      return
    }
    getDiscoverList().then(setResult).catch(() => {})
  })

  const onSubscribe = async (item: DiscoverSource) => {
    if (item.subscribed || subscribing[item.id]) return
    setSubscribing((prev) => ({ ...prev, [item.id]: true }))
    try {
      await subscribeRecommend(item.id)
      setResult((prev) => (prev ? markSubscribed(prev, item.id) : prev))
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
}

function SubscribeButton({ item, loading, onSubscribe }: SubscribeBtnProps) {
  if (item.subscribed) {
    return <Text className={styles.subBtnDone}>已订阅</Text>
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
}

function CategoryBlock({ category, subscribing, onSubscribe }: CategoryBlockProps) {
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
        <Text className={styles.catArrow}>{expanded ? '▾' : '▸'}</Text>
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
              />
            </View>
          ))}
        </View>
      )}
    </View>
  )
}
