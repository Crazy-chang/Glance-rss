import { View, Text } from '@tarojs/components'
import type { ArticleListItem as ArticleListItemType } from '@/api/types'
import { formatRelative } from '@/utils/time'
import styles from './index.module.scss'

interface Props {
  item: ArticleListItemType
  onClick: (item: ArticleListItemType) => void
}

export default function ArticleListItem({ item, onClick }: Props) {
  return (
    <View className={styles.item} hoverClass={styles.itemHover} onClick={() => onClick(item)}>
      <View className={styles.titleRow}>
        {!item.is_read && <View className={styles.dot} />}
        <Text className={styles.title}>{item.title}</Text>
      </View>
      {!!item.summary && <Text className={styles.summary}>{item.summary}</Text>}
      <Text className={styles.meta}>
        {item.source_name} · {formatRelative(item.published_at)}
      </Text>
    </View>
  )
}
