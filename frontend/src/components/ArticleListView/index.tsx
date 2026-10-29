import { Input, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import type { ArticleListItem as ArticleListItemType, FilterType } from '@/api/types'
import { useArticleList } from '@/hooks/useArticleList'
import { setNavContext } from '@/utils/navContext'
import ArticleListItem from '../ArticleListItem'
import EmptyState from '../EmptyState'
import FilterBar from '../FilterBar'
import LoadingMore from '../LoadingMore'
import Spinner from '../Spinner'
import styles from './index.module.scss'

interface Props {
  /** 限定单个订阅源；不传为全部订阅（文章 Tab） */
  sourceId?: number
  /** 是否探测"无任何订阅"空态（仅文章 Tab） */
  withSourceProbe?: boolean
  /** 首页折叠头部：标题行显示当前筛选总数，滚动收起后搜索框上移占位 */
  collapsible?: boolean
  /** 收起状态（由页面 onPageScroll 驱动） */
  collapsed?: boolean
}

function emptyText(filter: FilterType, keyword: string, noSource: boolean, withSourceProbe: boolean): string {
  if (keyword) return '未找到相关文章'
  switch (filter) {
    case 'unread':
      return '暂无未读文章'
    case 'read':
      return '暂无已读文章'
    case 'today':
      return '今天还没有更新'
    case 'favorited':
      return '还没有收藏文章'
    default:
      return withSourceProbe && noSource ? '还没有订阅内容' : '暂无文章'
  }
}

export default function ArticleListView({
  sourceId,
  withSourceProbe = false,
  collapsible = false,
  collapsed = false
}: Props) {
  const {
    list,
    loading,
    loadingMore,
    refreshing,
    filter,
    keyword,
    searchInput,
    noSource,
    total,
    changeFilter,
    onSearchInput,
    clearSearch,
    buildNavContext
  } = useArticleList({ sourceId, withSourceProbe })

  const goDetail = (item: ArticleListItemType) => {
    setNavContext({
      ...buildNavContext(),
      anchorId: item.id
    })
    Taro.navigateTo({ url: `/pages/article-detail/index?id=${item.id}` })
  }

  const showDiscoverEntry = withSourceProbe && noSource && !keyword && filter === 'all'

  return (
    <View
      className={`${styles.page} ${collapsible ? styles.collapsible : ''} ${collapsed ? styles.collapsed : ''}`}
    >
      <View className={styles.header}>
        {/* 首页折叠头部：标题行与胶囊齐平，滚动收起时淡出让位给搜索框 */}
        {collapsible && (
          <View className={styles.titleRow}>
            <Text className={styles.pageTitle}>文章</Text>
            <Text className={styles.pageCount}>（{total}）</Text>
          </View>
        )}
        {/* source-article 无标题行：搜索行与胶囊齐平让出胶囊宽度；首页展开时在标题行下全宽 */}
        <View className={styles.searchRow}>
          <View className={styles.searchBox}>
            <Input
              className={styles.searchInput}
              value={searchInput}
              placeholder="搜索文章"
              placeholderClass={styles.searchPlaceholder}
              confirmType="search"
              onInput={(e) => onSearchInput(e.detail.value)}
            />
            {!!searchInput && (
              <View className={styles.clearBtn} onClick={clearSearch}>
                <Text className={styles.clearIcon}>×</Text>
              </View>
            )}
          </View>
        </View>
        <FilterBar value={filter} onChange={changeFilter} />
      </View>

      <View className={styles.body}>
        {loading && list.length === 0 ? (
          <View className={styles.firstLoading}>
            <Spinner size={36} />
          </View>
        ) : list.length === 0 ? (
          <EmptyState
            text={emptyText(filter, keyword, noSource, withSourceProbe)}
            actionText={showDiscoverEntry ? '推荐订阅源' : undefined}
            onAction={() => Taro.navigateTo({ url: '/pages/discover/index' })}
          />
        ) : (
          <View className={styles.listCard}>
            {list.map((item) => (
              <ArticleListItem key={item.id} item={item} onClick={goDetail} />
            ))}
          </View>
        )}
        <LoadingMore loading={loadingMore} />
      </View>
      {/* 下拉刷新：只保留转圈，原生下拉动画在 onPullDownRefresh 里已立即收起 */}
      {refreshing && (
        <View className={styles.refreshTip}>
          <Spinner size={30} />
        </View>
      )}
    </View>
  )
}
