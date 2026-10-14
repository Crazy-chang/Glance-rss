import { View, Text } from '@tarojs/components'
import Spinner from '../Spinner'
import styles from './index.module.scss'

interface Props {
  loading: boolean
}

export default function LoadingMore({ loading }: Props) {
  if (!loading) return null
  return (
    <View className={styles.wrap}>
      <Spinner size={28} />
      <Text className={styles.text}>加载中…</Text>
    </View>
  )
}
