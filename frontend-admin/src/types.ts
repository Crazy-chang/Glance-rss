// 后台接口返回类型
export interface Category {
  id: number;
  name: string;
  description: string;
  sort_order: number;
  source_count: number;
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
}

export interface ExportSource {
  title: string;
  feed_url: string;
  site_url: string;
  description: string;
  hot_rank: number;
}

export interface ExportCategory {
  name: string;
  description: string;
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
