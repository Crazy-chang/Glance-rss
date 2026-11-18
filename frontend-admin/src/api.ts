import type { AdminLog, Category, ExportData, Source, SourceInput, UserListResult } from './types';

const TOKEN_KEY = 'glance_admin_token';

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || '';
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

interface ApiEnvelope<T> {
  code: number;
  msg: string;
  data: T;
}

async function request<T>(path: string, method: 'GET' | 'POST' = 'GET', body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }
  const token = getToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`/api/admin${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  let payload: ApiEnvelope<T> | null = null;
  try {
    payload = (await res.json()) as ApiEnvelope<T>;
  } catch {
    // 非 JSON 响应（如网关错误页）
  }

  if (!res.ok || !payload || payload.code !== 0) {
    if (res.status === 401) {
      clearToken();
      if (!window.location.hash.startsWith('#/login')) {
        window.location.hash = '#/login';
      }
    }
    throw new Error(payload?.msg || `请求失败（${res.status}）`);
  }
  return payload.data;
}

export const api = {
  login(username: string, password: string) {
    return request<{ token: string; username: string }>('/login', 'POST', { username, password });
  },
  logout() {
    return request<{ ok: boolean }>('/logout', 'POST', {});
  },

  categories() {
    return request<Category[]>('/category/list');
  },
  createCategory(data: { name: string; description: string }) {
    return request<Category>('/category/create', 'POST', data);
  },
  updateCategory(data: { id: number; name: string; description: string }) {
    return request<{ ok: boolean }>('/category/update', 'POST', data);
  },
  deleteCategory(id: number) {
    return request<{ ok: boolean }>('/category/delete', 'POST', { id });
  },
  reorderCategories(ids: number[]) {
    return request<{ ok: boolean }>('/category/reorder', 'POST', { ids });
  },

  sources(categoryId?: number) {
    const query = categoryId ? `?category_id=${categoryId}` : '';
    return request<Source[]>(`/recommend/list${query}`);
  },
  createSource(data: SourceInput) {
    return request<Source>('/recommend/create', 'POST', data);
  },
  updateSource(id: number, data: SourceInput) {
    return request<{ ok: boolean }>('/recommend/update', 'POST', { id, ...data });
  },
  deleteSource(id: number) {
    return request<{ ok: boolean }>('/recommend/delete', 'POST', { id });
  },
  reorderSources(ids: number[]) {
    return request<{ ok: boolean }>('/recommend/reorder', 'POST', { ids });
  },
  exportData() {
    return request<ExportData>('/recommend/export');
  },
  importData(data: ExportData) {
    return request<{ ok: boolean }>('/recommend/import', 'POST', data);
  },

  resetAll() {
    return request<{ ok: boolean }>('/system/reset', 'POST', {});
  },
  users(page = 1, size = 20) {
    return request<UserListResult>(`/user/list?page=${page}&size=${size}`);
  },
  logs(limit = 100) {
    return request<AdminLog[]>(`/system/logs?limit=${limit}`);
  },
};
