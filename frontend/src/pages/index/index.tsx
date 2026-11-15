import { useState } from 'react'
import { View } from '@tarojs/components'
import { usePageScroll, useShareAppMessage } from '@tarojs/taro'
import ArticleListView from '@/components/ArticleListView'
import { useHeaderMetrics } from '@/hooks/useHeaderMetrics'

// 文章 Tab：全部订阅的文章流（搜索 / 筛选 / 分页 / 下拉刷新）
// 头部折叠：上滑超过阈值收起（标题淡出、搜索框上移占位），回顶自动展开；带回滞避免边界抖动
export default function ArticlePage() {
  const { statusBarHeight, capsuleWidth } = useHeaderMetrics()
  const [collapsed, setCollapsed] = useState(false)

  // 整页分享：文案与订阅页保持一致
  useShareAppMessage(() => ({
    title: 'RSS订阅阅读器',
    path: '/pages/index/index'
  }))

  usePageScroll(({ scrollTop }) => {
    setCollapsed((prev) => (prev ? scrollTop > 40 : scrollTop > 80))
  })

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
      <ArticleListView withSourceProbe collapsible collapsed={collapsed} />
    </View>
  )
}
