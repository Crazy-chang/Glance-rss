import { useCallback, useEffect, useRef, useState } from 'react'
import Taro from '@tarojs/taro'
import { Text, View } from '@tarojs/components'
import { getUserProfile, readBatch } from '@/api/article'
import styles from './index.module.scss'

interface Props {
  /** 批量标记完成后的回调（通常用于刷新列表） */
  onDone?: () => void
}

interface Pos {
  x: number
  y: number
}

// 与样式保持一致：750 设计稿下的尺寸（scss px 会被转换成 rpx）
const FAB_W = 200
const FAB_H = 96
const EDGE_GAP = 8
const POS_KEY = 'batch_read_fab_pos'

function getWindowSize() {
  try {
    const info = Taro.getWindowInfo ? Taro.getWindowInfo() : Taro.getSystemInfoSync()
    return { width: info.windowWidth, height: info.windowHeight }
  } catch {
    return { width: 375, height: 667 }
  }
}

// 设计稿尺寸换算成当前设备的物理像素
const WIN = getWindowSize()
const SCALE = WIN.width / 750
const FAB_W_PX = FAB_W * SCALE
const FAB_H_PX = FAB_H * SCALE
const MIN_TOP = 80

function clampPos(p: Pos): Pos {
  const maxX = Math.max(EDGE_GAP, WIN.width - FAB_W_PX - EDGE_GAP)
  const maxY = Math.max(MIN_TOP, WIN.height - FAB_H_PX - EDGE_GAP)
  return {
    x: Math.min(Math.max(p.x, EDGE_GAP), maxX),
    y: Math.min(Math.max(p.y, MIN_TOP), maxY)
  }
}

function defaultPos(): Pos {
  return clampPos({ x: WIN.width - FAB_W_PX - EDGE_GAP, y: WIN.height * 0.62 })
}

/**
 * 一键阅读悬浮按钮：
 * - 显隐由组件内部根据用户资料自行判断，不依赖页面传参
 * - 可拖动，松手自动吸附到左右边缘，位置持久化到本地存储
 */
export default function BatchReadFab({ onDone }: Props) {
  const [visible, setVisible] = useState(false)
  const [limit, setLimit] = useState(0)
  const [pos, setPos] = useState<Pos>(defaultPos)
  const [dragging, setDragging] = useState(false)
  const [busy, setBusy] = useState(false)

  const posRef = useRef<Pos>(pos)
  const busyRef = useRef(false)
  const dragRef = useRef({ startX: 0, startY: 0, originX: 0, originY: 0, moved: false })

  const updatePos = useCallback((next: Pos) => {
    posRef.current = next
    setPos(next)
  }, [])

  // 拉取用户资料：未开启一键阅读时组件不渲染任何内容
  useEffect(() => {
    let alive = true
    getUserProfile()
      .then((profile) => {
        if (!alive || !profile.batch_read_enabled) return
        setLimit(profile.batch_read_limit)
        setVisible(true)
        const stored = Taro.getStorageSync(POS_KEY) as Pos | '' | null
        const init =
          stored && typeof stored.x === 'number' && typeof stored.y === 'number'
            ? clampPos(stored)
            : defaultPos()
        updatePos(init)
      })
      .catch(() => {
        // 静默失败：拿不到资料就不展示按钮
      })
    return () => {
      alive = false
    }
  }, [updatePos])

  const onTouchStart = (e: any) => {
    const touch = e.touches && e.touches[0]
    if (!touch) return
    dragRef.current = {
      startX: touch.clientX,
      startY: touch.clientY,
      originX: posRef.current.x,
      originY: posRef.current.y,
      moved: false
    }
    setDragging(true)
  }

  const onTouchMove = (e: any) => {
    const touch = e.touches && e.touches[0]
    if (!touch) return
    const dx = touch.clientX - dragRef.current.startX
    const dy = touch.clientY - dragRef.current.startY
    if (Math.abs(dx) > 4 || Math.abs(dy) > 4) dragRef.current.moved = true
    updatePos(
      clampPos({
        x: dragRef.current.originX + dx,
        y: dragRef.current.originY + dy
      })
    )
  }

  const onTouchEnd = () => {
    setDragging(false)
    if (!dragRef.current.moved) return
    // 松手吸附到更近的一侧
    const centerX = posRef.current.x + FAB_W_PX / 2
    const snappedX =
      centerX < WIN.width / 2 ? EDGE_GAP : WIN.width - FAB_W_PX - EDGE_GAP
    const next = clampPos({ x: snappedX, y: posRef.current.y })
    updatePos(next)
    Taro.setStorageSync(POS_KEY, next)
  }

  const runBatchRead = async () => {
    if (busyRef.current) return
    const content =
      limit > 0
        ? `将最新的 ${limit} 篇未读文章标记为已读？`
        : '将全部未读文章标记为已读？'
    const modal = await Taro.showModal({
      title: '一键阅读',
      content,
      confirmText: '标记已读',
      confirmColor: '#5b9df9'
    })
    if (!modal.confirm) return

    busyRef.current = true
    setBusy(true)
    Taro.showLoading({ title: '处理中', mask: true })
    try {
      const data = await readBatch()
      Taro.hideLoading()
      Taro.showToast({
        title: data.count > 0 ? `已标记 ${data.count} 篇` : '没有未读文章',
        icon: 'none'
      })
      if (data.count > 0) onDone?.()
    } catch {
      // 失败提示由请求层统一 toast
      Taro.hideLoading()
    } finally {
      busyRef.current = false
      setBusy(false)
    }
  }

  if (!visible) return null

  return (
    <View
      className={`${styles.fab} ${dragging ? styles.dragging : ''}`}
      style={{ left: `${pos.x}px`, top: `${pos.y}px` }}
      catchMove
      onTouchStart={onTouchStart}
      onTouchMove={onTouchMove}
      onTouchEnd={onTouchEnd}
      onTouchCancel={onTouchEnd}
      onClick={() => {
        if (dragRef.current.moved) {
          dragRef.current.moved = false
          return
        }
        if (!busy) runBatchRead()
      }}
    >
      <Text className={styles.label}>{busy ? '处理中' : '一键阅读'}</Text>
    </View>
  )
}
