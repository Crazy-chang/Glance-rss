import { View, Text } from '@tarojs/components'
import styles from './index.module.scss'

interface Props {
  text: string
  subText?: string
  actionText?: string
  onAction?: () => void
}

export default function EmptyState({ text, subText, actionText, onAction }: Props) {
  return (
    <View className={styles.wrap}>
      <View className={styles.textGroup}>
        <Text className={styles.text}>{text}</Text>
        {subText && <Text className={styles.sub}>{subText}</Text>}
      </View>
      {actionText && (
        <View className={styles.action} onClick={onAction}>
          {actionText}
        </View>
      )}
    </View>
  )
}
