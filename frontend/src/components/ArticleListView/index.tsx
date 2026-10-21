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

export default function ArticleListView({ sourceId, withSourceProbe = false }: Props) {
  const {
    list,
    loading,
    loadingMore,
    filter,
    keyword,
    searchInput,
    noSource,
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
    <View className={styles.page}>
      <View className={styles.header}>
        {/* 搜索行与胶囊按钮齐平，右侧让出胶囊宽度；筛选栏保持全宽 */}
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
    </View>
  )
}
