import { Text, View } from '@tarojs/components'
import classnames from 'classnames'
import { useSetting } from '@/store/setting'
import type { UserSetting } from '@/api/types'
import { FONT_SIZES, THEME_KEYS, themeChrome } from '@/utils/theme'
import styles from './index.module.scss'

// 每源最大条数选项（对齐后端 validItems 白名单）
const MAX_ITEM_OPTIONS: { key: UserSetting['max_items']; label: string }[] = [
  { key: 50, label: '50 篇' },
  { key: 100, label: '100 篇' },
  { key: 200, label: '200 篇' },
  { key: 500, label: '500 篇' }
]

// 刷新间隔选项（对齐后端 model 定义的枚举值）
const REFRESH_INTERVAL_OPTIONS: { key: UserSetting['refresh_interval']; label: string }[] = [
  { key: '30m', label: '30 分钟' },
  { key: '1h', label: '1 小时' },
  { key: '2h', label: '2 小时' },
  { key: 'manual', label: '手动' }
]

// 排序选项
const SORT_OPTIONS: { key: UserSetting['sort_order']; label: string }[] = [
  { key: 'desc', label: '最新' },
  { key: 'asc', label: '最早' }
]

// 设置 Tab：承载阅读主题 / 字号 / 每源条数 / 刷新间隔 / 排序
export default function SettingPage() {
  const { setting, updateSetting } = useSetting()

  const theme = setting?.theme || 'white'
  const fontSize = setting?.font_size || 'medium'
  const maxItems = setting?.max_items ?? 100
  const refreshInterval = setting?.refresh_interval || '30m'
  const sortOrder = setting?.sort_order || 'desc'

  const change = async (patch: Partial<UserSetting>) => {
    try {
      await updateSetting(patch)
    } catch {
      // request 层已 toast；store 已回滚为服务端数据
    }
  }

  /** 渲染一张分段控件卡片（浅灰槽 + 选中项白底滑块） */
  const renderSegCard = (
    title: string,
    options: { key: any; label?: string; name?: string }[],
    current: any,
    onPick: (k: any) => void
  ) => (
    <View className={styles.card}>
      <Text className={styles.cardTitle}>{title}</Text>
      <View className={styles.segRow}>
        {options.map((o) => {
          const active = current === o.key
          return (
            <View
              key={String(o.key)}
              className={classnames(styles.seg, active && styles.segActive)}
              hoverClass={active ? undefined : styles.segHover}
              onClick={() => !active && onPick(o.key)}
            >
              <Text>{o.label ?? o.name ?? String(o.key)}</Text>
            </View>
          )
        })}
      </View>
    </View>
  )

  return (
    <View className={styles.page}>
      {/* 阅读主题 */}
      <View className={styles.card}>
        <Text className={styles.cardTitle}>阅读主题</Text>
        <View className={styles.swatchRow}>
          {THEME_KEYS.map((key) => {
            const chrome = themeChrome(key)
            const active = theme === key
            return (
              <View
                key={key}
                className={styles.swatchItem}
                hoverClass={styles.swatchHover}
                onClick={() => !active && change({ theme: key })}
              >
                <View
                  className={classnames(styles.swatch, active && styles.swatchActive)}
                  style={{ background: chrome.swatch }}
                >
                  {active && <Text className={styles.swatchCheck}>✓</Text>}
                </View>
              </View>
            )
          })}
        </View>
      </View>

      {renderSegCard('正文字号', FONT_SIZES, fontSize, (k) => change({ font_size: k }))}
      {renderSegCard('每源最大条数', MAX_ITEM_OPTIONS, maxItems, (k) => change({ max_items: k }))}
      {renderSegCard('静默刷新间隔', REFRESH_INTERVAL_OPTIONS, refreshInterval, (k) => change({ refresh_interval: k }))}
      {renderSegCard('排序', SORT_OPTIONS, sortOrder, (k) => change({ sort_order: k }))}
    </View>
  )
}
