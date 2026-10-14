import { useEffect, useMemo } from 'react'
import Taro, { useRouter } from '@tarojs/taro'
import ArticleListView from '@/components/ArticleListView'

// 订阅文章页：单个订阅源的文章流，交互与文章 Tab 一致
export default function SourceArticlePage() {
  const router = useRouter()

  const sourceId = useMemo(() => {
    const id = Number(router.params.id)
    return Number.isFinite(id) && id > 0 ? id : 0
  }, [router.params.id])

  useEffect(() => {
    const name = router.params.name ? decodeURIComponent(router.params.name) : '订阅文章'
    Taro.setNavigationBarTitle({ title: name }).catch(() => {})
  }, [router.params.name])

  if (!sourceId) return null
  return <ArticleListView sourceId={sourceId} />
}
