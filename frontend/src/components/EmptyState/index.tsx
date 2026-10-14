import { View, Text } from '@tarojs/components'
import styles from './index.module.scss'

interface Props {
  text: string
  actionText?: string
  onAction?: () => void
}

export default function EmptyState({ text, actionText, onAction }: Props) {
  return (
    <View className={styles.wrap}>
      <Text className={styles.text}>{text}</Text>
      {actionText && (
        <View className={styles.action} onClick={onAction}>
          {actionText}
        </View>
      )}
    </View>
  )
}
