import { View } from '@tarojs/components'
import classnames from 'classnames'
import type { FilterType } from '@/api/types'
import styles from './index.module.scss'

const FILTERS: { key: FilterType; label: string }[] = [
  { key: 'unread', label: '未 读' },
  { key: 'all', label: '全 部' },
  { key: 'read', label: '已 读' },
  { key: 'today', label: '今 日' },
  { key: 'favorited', label: '收 藏' }
]

interface Props {
  value: FilterType
  onChange: (filter: FilterType) => void
}

export default function FilterBar({ value, onChange }: Props) {
  return (
    <View className={styles.bar}>
      {FILTERS.map((f) => (
        <View
          key={f.key}
          className={classnames(styles.chip, value === f.key && styles.chipActive)}
          onClick={() => onChange(f.key)}
        >
          {f.label}
        </View>
      ))}
    </View>
  )
}
