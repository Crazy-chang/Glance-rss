import Taro, { useReady } from '@tarojs/taro'
import { useState } from 'react'

/** 自定义导航栏行高（物理 px）：容纳右上角胶囊按钮，内容不得进入该区域 */
const NAV_BAR_ROW = 44
/** 内容与胶囊左边缘的最小间隙（物理 px） */
const CAPSULE_GAP = 8

export interface HeaderMetrics {
  /** 头部总留白：状态栏 + 导航栏行（物理 px） */
  navHeight: number
  /** 状态栏高度（物理 px） */
  statusBarHeight: number
  /** 搜索框右侧需让出的宽度：屏幕右边缘到胶囊左边缘 + 间隙（物理 px） */
  capsuleWidth: number
}

const DEFAULT_METRICS: HeaderMetrics = {
  navHeight: NAV_BAR_ROW + 20,
  statusBarHeight: 20,
  capsuleWidth: 96
}

/**
 * 自定义导航栏页面用：读取状态栏高度与胶囊按钮位置。
 * 页面根节点通过 inline style 注入 CSS 变量：
 *   --nav-h          头部总留白（吸顶/固定元素用）
 *   --status-bar-h   状态栏高度（头部从屏幕顶端铺满时用）
 *   --capsule-w      右侧需让出的胶囊宽度
 */
export function useHeaderMetrics() {
  const [metrics, setMetrics] = useState<HeaderMetrics>(DEFAULT_METRICS)

  useReady(() => {
    try {
      const info = Taro.getSystemInfoSync()
      const statusBarHeight = info.statusBarHeight || 20
      let capsuleWidth = DEFAULT_METRICS.capsuleWidth
      const rect = Taro.getMenuButtonBoundingClientRect()
      if (rect && rect.left > 0 && info.windowWidth) {
        capsuleWidth = info.windowWidth - rect.left + CAPSULE_GAP
      }
      setMetrics({
        navHeight: statusBarHeight + NAV_BAR_ROW,
        statusBarHeight,
        capsuleWidth
      })
    } catch (err) {
      console.error('[useHeaderMetrics] 获取头部尺寸失败', err)
    }
  })

  return metrics
}
