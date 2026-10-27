import { View } from '@tarojs/components'
import ArticleListView from '@/components/ArticleListView'
import { useHeaderMetrics } from '@/hooks/useHeaderMetrics'

// 文章 Tab：全部订阅的文章流（搜索 / 筛选 / 分页 / 下拉刷新）
export default function ArticlePage() {
  const { statusBarHeight, capsuleWidth } = useHeaderMetrics()
  return (
    <View
      style={
        {
          '--status-bar-h': `${statusBarHeight}px`,
          '--capsule-w': `${capsuleWidth}px`,
          '--header-bg': 'var(--card)'
        } as any
      }
    >
      <ArticleListView withSourceProbe />
    </View>
  )
}
