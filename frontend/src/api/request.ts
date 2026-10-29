import Taro from '@tarojs/taro'
import { BASE_URL } from './config'
import type { ApiResp } from './types'

const TOKEN_KEY = 'glance_token'

export function getToken(): string {
  return Taro.getStorageSync(TOKEN_KEY) || ''
}

// loginPromise 保证全应用只有一个登录请求在飞
let loginPromise: Promise<string> | null = null
// 登录失败冷却时间（ms）：断网时避免每个请求都再打一次 /auth/login
const LOGIN_FAIL_COOLDOWN = 3000
let loginFailedAt = 0

/**
 * 静默登录：wx.login 取 code 换 JWT。
 * force=true 时（401 重试）无论本地是否有 token 都重新换发。
 */
export function login(force = false): Promise<string> {
  if (!force) {
    if (loginPromise) return loginPromise
    if (getToken()) return Promise.resolve(getToken())
  }

  const task = (async () => {
    const { code } = await Taro.login()
    const res = await Taro.request({
      url: `${BASE_URL}/api/auth/login`,
      method: 'POST',
      data: { code },
      timeout: 10000
    })
    const body = res.data as ApiResp<{ token: string }>
    if (res.statusCode !== 200 || body.code !== 0 || !body.data?.token) {
      throw new Error(body?.msg || '登录失败')
    }
    Taro.setStorageSync(TOKEN_KEY, body.data.token)
    console.info('[Auth] 登录成功')
    return body.data.token
  })()

  // 不用 Promise.finally（项目 TS target 为 es2017）
  loginPromise = task.then(
    (v) => {
      loginPromise = null
      return v
    },
    (e) => {
      loginPromise = null
      throw e
    }
  )
  return loginPromise
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, any> | string
  query?: Record<string, any>
  silent?: boolean // true 时不弹错误 toast（如下拉刷新这类后台动作）
}

function buildQuery(query?: RequestOptions['query']): string {
  if (!query) return ''
  const parts: string[] = []
  Object.keys(query).forEach((k) => {
    const v = query[k]
    if (v !== undefined && v !== null && v !== '') {
      parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    }
  })
  return parts.length ? `?${parts.join('&')}` : ''
}

/**
 * 统一请求出口：
 * - 自动等待静默登录、注入 Bearer token
 * - 401 自动重新登录并重试一次
 * - 业务 code 非 0 统一 toast，返回 reject
 */
export async function request<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
  if (!getToken()) {
    if (Date.now() - loginFailedAt < LOGIN_FAIL_COOLDOWN) {
      throw new Error('网络异常，请稍后再试')
    }
    try {
      await login()
    } catch (err) {
      loginFailedAt = Date.now()
      throw err
    }
  }

  const url = `${BASE_URL}/api${path}${buildQuery(options.query)}`

  const doFetch = (token: string) =>
    Taro.request({
      url,
      method: options.method || 'GET',
      data: options.data,
      header: { Authorization: `Bearer ${token}` },
      timeout: 15000
    })

  let res = await doFetch(getToken())

  // token 过期：重新登录后重试一次
  if (res.statusCode === 401) {
    const freshToken = await login(true)
    res = await doFetch(freshToken)
  }

  const body = res.data as ApiResp<T>
  if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
    return body.data
  }

  console.error('[request] 请求失败', path, res.statusCode, body)
  const msg = body?.msg || '网络异常，请稍后再试'
  if (!options.silent) {
    Taro.showToast({ title: msg, icon: 'none' })
  }
  throw Object.assign(new Error(msg), { code: body?.code, statusCode: res.statusCode })
}
