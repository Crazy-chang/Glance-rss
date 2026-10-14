import { View } from '@tarojs/components'
import styles from './index.module.scss'

interface Props {
  size?: number
}

// 纯 CSS 旋转加载，Taro 4 已无 ActivityIndicator 组件
export default function Spinner({ size = 36 }: Props) {
  return <View className={styles.spinner} style={{ width: `${size}px`, height: `${size}px` }} />
}
