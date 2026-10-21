import { Text, View } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useSetting } from '@/store/setting'
import { useHeaderMetrics } from '@/hooks/useHeaderMetrics'
import { clearReadArticles } from '@/api/article'
import type { UserSetting } from '@/api/types'
import { FONT_SIZES, THEME_KEYS, THEME_CHROME, themeChrome } from '@/utils/theme'
import styles from './index.module.scss'

// 每源最大条数选项（对齐后端 validItems 白名单）
const MAX_ITEM_OPTIONS: { key: UserSetting['max_items']; label: string }[] = [
  { key: 100, label: '100 篇' },
  { key: 200, label: '200 篇' },
  { key: 300, label: '300 篇' },
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

// 版本号（与 package.json version 保持一致）
const APP_VERSION = 'v0.1.0'

// 设置 Tab：承载阅读主题 / 字号 / 每源条数 / 刷新间隔 / 排序
export default function SettingPage() {
  const { navHeight } = useHeaderMetrics()
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

  /** 清理缓存：删除已读且未收藏的文章（二次确认） */
  const onClearCache = async () => {
    const { confirm } = await Taro.showModal({
      title: '清理已读文章',
      content: '将删除所有已读且未收藏的文章，释放存储空间。收藏的文章会保留，此操作不可恢复。',
      confirmText: '清理',
      cancelText: '取消',
      confirmColor: '#FA5151'
    })
    if (!confirm) return
    try {
      const res = await clearReadArticles()
      if (res.deleted > 0) {
        Taro.showToast({ title: `已清理 ${res.deleted} 篇文章`, icon: 'success' })
      } else {
        Taro.showToast({ title: '暂无可清理的文章', icon: 'none' })
      }
    } catch {
      // request 层已 toast
    }
  }

  /** 通用 action sheet 选择 */
  const pick = async <T,>(
    options: { key: T; label: string }[],
    current: T,
    onPick: (k: T) => void
  ) => {
    const res = await Taro.showActionSheet({ itemList: options.map((o) => o.label) })
    const picked = options[res.tapIndex]
    if (picked && picked.key !== current) onPick(picked.key)
  }

  return (
    <View
      className={styles.page}
      style={{ '--nav-h': `${navHeight}px` } as any}
    >
      <View className={styles.card}>
        {/* 阅读主题 */}
        <View
          className={styles.row}
          onClick={() =>
            pick(
              THEME_KEYS.map((k) => ({ key: k, label: THEME_CHROME[k].name })),
              theme,
              (k) => change({ theme: k })
            )
          }
        >
          <Text className={styles.rowLabel}>阅读主题</Text>
          <View className={styles.rowValue}>
            <View className={styles.themeDot} style={{ background: themeChrome(theme).swatch }} />
            <Text className={styles.rowValueText}>{themeChrome(theme).name}</Text>
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>

        {/* 正文字号 */}
        <View
          className={styles.row}
          onClick={() =>
            pick(
              FONT_SIZES.map((f) => ({ key: f.key, label: f.name })),
              fontSize,
              (k) => change({ font_size: k })
            )
          }
        >
          <Text className={styles.rowLabel}>正文字号</Text>
          <View className={styles.rowValue}>
            <Text className={styles.rowValueText}>
              {FONT_SIZES.find((f) => f.key === fontSize)?.name}
            </Text>
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>

        {/* 每源最大条数 */}
        <View
          className={styles.row}
          onClick={() => pick(MAX_ITEM_OPTIONS, maxItems, (k) => change({ max_items: k }))}
        >
          <Text className={styles.rowLabel}>每源最大条数</Text>
          <View className={styles.rowValue}>
            <Text className={styles.rowValueText}>
              {MAX_ITEM_OPTIONS.find((o) => o.key === maxItems)?.label}
            </Text>
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>

        {/* 刷新间隔 */}
        <View
          className={styles.row}
          onClick={() =>
            pick(REFRESH_INTERVAL_OPTIONS, refreshInterval, (k) => change({ refresh_interval: k }))
          }
        >
          <Text className={styles.rowLabel}>刷新间隔</Text>
          <View className={styles.rowValue}>
            <Text className={styles.rowValueText}>
              {REFRESH_INTERVAL_OPTIONS.find((o) => o.key === refreshInterval)?.label}
            </Text>
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>

        {/* 排序 */}
        <View
          className={styles.row}
          onClick={() => pick(SORT_OPTIONS, sortOrder, (k) => change({ sort_order: k }))}
        >
          <Text className={styles.rowLabel}>按时间倒序排序</Text>
          <View className={styles.rowValue}>
            <Text className={styles.rowValueText}>
              {SORT_OPTIONS.find((o) => o.key === sortOrder)?.label}
            </Text>
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>

        {/* 清理缓存：删除已读且未收藏的文章 */}
        <View className={styles.rowLast} onClick={onClearCache}>
          <Text className={styles.rowLabel}>清理缓存</Text>
          <View className={styles.rowValue}>
            {/* <Text className={styles.rowValueHint}>已读文章</Text> */}
            <Text className={styles.chevron}>›</Text>
          </View>
        </View>
      </View>

      <Text className={styles.version}>Glance RSS {APP_VERSION}</Text>
    </View>
  )
}
