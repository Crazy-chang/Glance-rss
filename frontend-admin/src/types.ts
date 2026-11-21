// 后台接口返回类型
export interface Category {
  id: number;
  name: string;
  description: string;
  sort_order: number;
  source_count: number;
  enabled: boolean;
}

export interface Source {
  id: number;
  title: string;
  feed_url: string;
  site_url: string;
  category: string;
  category_id: number;
  description: string;
  hot_rank: number;
  sort_order: number;
  enabled: boolean;
}

export interface ExportSource {
  title: string;
  feed_url: string;
  site_url: string;
  description: string;
  hot_rank: number;
  enabled?: boolean;
}

export interface ExportCategory {
  name: string;
  description: string;
  enabled?: boolean;
  sources: ExportSource[];
}

export interface ExportData {
  categories: ExportCategory[];
}

export interface AdminLog {
  id: number;
  admin_id: number;
  username: string;
  action: string;
  detail: string;
  created_at: string;
}

export interface SourceInput {
  title: string;
  feed_url: string;
  site_url: string;
  description: string;
  category_id: number;
  hot_rank: number;
}

export interface AdminUserSource {
  id: number;
  title: string;
  feed_url: string;
  site_url: string;
  category: string;
  last_fetch_at: string | null;
  last_error: string;
  created_at: string;
}

export interface AdminUser {
  id: number;
  open_id: string;
  batch_read_enabled: boolean;
  created_at: string;
  updated_at: string;
  source_count: number;
  sources: AdminUserSource[];
}

export interface AppConfig {
  batch_read_limit: number;
}

export interface UserListResult {
  list: AdminUser[];
  total: number;
  page: number;
  size: number;
  hasMore: boolean;
}
