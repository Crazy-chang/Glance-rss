import { View } from '@tarojs/components'
import ArticleListView from '@/components/ArticleListView'

// 文章 Tab：全部订阅的文章流（搜索 / 筛选 / 分页 / 下拉刷新）
export default function ArticlePage() {
  return (
    <View>
      <ArticleListView withSourceProbe />
    </View>
  )
}
