// 主题/字号的静态配置：运行时接线用（原生导航栏、tabBar、窗口背景）。
// 随主题变化的页面内颜色见 app.scss 的 CSS 自定义属性，两处取值保持一致。

export interface ThemeChrome {
  name: string
  swatch: string // 设置页色块颜色
  navFront: '#000000' | '#ffffff' // 导航栏文字颜色（仅支持黑/白）
  navBg: string // 导航栏背景 = 页面背景
  tabBg: string // tabBar 背景 = 卡片色
  tabColor: string // tabBar 未选中文字
  tabSelected: string // tabBar 选中文字（品牌色恒定）
  tabBorder: 'black' | 'white'
}

export const DEFAULT_THEME = 'white'
export const ACCENT = '#5B9DF9'

export const THEME_CHROME: Record<string, ThemeChrome> = {
  white: {
    name: '浅色',
    swatch: '#FFFFFF',
    navFront: '#000000',
    navBg: '#F7F8FA',
    tabBg: '#FFFFFF',
    tabColor: '#999999',
    tabSelected: ACCENT,
    tabBorder: 'white'
  },
  black: {
    name: '深色',
    swatch: '#121212',
    navFront: '#ffffff',
    navBg: '#121212',
    tabBg: '#1E1E1E',
    tabColor: '#888888',
    tabSelected: ACCENT,
    tabBorder: 'black'
  },
  yellow: {
    name: '黄色',
    swatch: '#F5ECD7',
    navFront: '#000000',
    navBg: '#F5ECD7',
    tabBg: '#FBF5E3',
    tabColor: '#8C7B5E',
    tabSelected: ACCENT,
    tabBorder: 'white'
  },
  green: {
    name: '绿色',
    swatch: '#D9E9DC',
    navFront: '#000000',
    navBg: '#D9E9DC',
    tabBg: '#EFF6EF',
    tabColor: '#6B7F70',
    tabSelected: ACCENT,
    tabBorder: 'white'
  }
}

export const THEME_KEYS = ['white', 'black', 'yellow', 'green']

export function themeChrome(theme: string): ThemeChrome {
  return THEME_CHROME[theme] || THEME_CHROME[DEFAULT_THEME]
}

export const FONT_SIZES = [
  { key: 'small', name: '小' },
  { key: 'medium', name: '中' },
  { key: 'large', name: '大' }
]
