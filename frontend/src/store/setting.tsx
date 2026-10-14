import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { getSetting, updateSetting as apiUpdateSetting } from '@/api/setting'
import { login } from '@/api/request'
import type { UserSetting } from '@/api/types'

interface SettingContextValue {
  setting: UserSetting | null
  updateSetting: (patch: Partial<UserSetting>) => Promise<void>
}

const SettingContext = createContext<SettingContextValue>({
  setting: null,
  updateSetting: async () => {}
})

export const DEFAULT_SETTING: UserSetting = {
  theme: 'white',
  font_size: 'medium',
  max_items: 100,
  refresh_interval: '30m',
  sort_order: 'desc'
}

export function SettingProvider({ children }: PropsWithChildren<any>) {
  const [setting, setSetting] = useState<UserSetting | null>(null)

  useEffect(() => {
    const init = async () => {
      try {
        await login()
        setSetting(await getSetting())
      } catch (err) {
        // 拉取失败不阻塞页面，使用默认主题；request 层已有错误提示
        console.error('[Setting] 初始化失败', err)
      }
    }
    init()
  }, [])

  const updateSetting = useCallback(
    async (patch: Partial<UserSetting>) => {
      setSetting((prev) => ({ ...DEFAULT_SETTING, ...prev, ...patch }))
      try {
        setSetting(await apiUpdateSetting(patch))
      } catch (err) {
        console.error('[Setting] 更新失败', err)
        // 失败时回滚为服务端数据
        setSetting(await getSetting().catch(() => null))
        throw err
      }
    },
    []
  )

  const value = useMemo(() => ({ setting, updateSetting }), [setting, updateSetting])

  return <SettingContext.Provider value={value}>{children}</SettingContext.Provider>
}

export function useSetting() {
  return useContext(SettingContext)
}
