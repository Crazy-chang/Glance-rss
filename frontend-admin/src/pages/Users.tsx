import { useCallback, useEffect, useState } from 'react';
import { Card, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { AdminUser, AdminUserSource } from '../types';

function formatTime(value?: string | null) {
  if (!value) return '-';
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}

const sourceColumns: ColumnsType<AdminUserSource> = [
  { title: '源 ID', dataIndex: 'id', width: 80 },
  { title: '名称', dataIndex: 'title', width: 200 },
  { title: '分类', dataIndex: 'category', width: 110, render: (value: string) => value || '-' },
  {
    title: 'Feed 地址',
    dataIndex: 'feed_url',
    ellipsis: true,
    render: (value: string) => (
      <a href={value} target="_blank" rel="noreferrer">
        {value}
      </a>
    ),
  },
  { title: '最近抓取', dataIndex: 'last_fetch_at', width: 180, render: formatTime },
  {
    title: '状态',
    dataIndex: 'last_error',
    width: 110,
    render: (value: string) =>
      value ? <Tag color="red">抓取失败</Tag> : <Tag color="green">正常</Tag>,
  },
];

export default function Users() {
  const [list, setList] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.users(page, size);
      setList(result.list);
      setTotal(result.total);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载用户失败');
    } finally {
      setLoading(false);
    }
  }, [page, size]);

  useEffect(() => {
    load();
  }, [load]);

  const columns: ColumnsType<AdminUser> = [
    { title: '用户 ID', dataIndex: 'id', width: 90 },
    {
      title: 'OpenID',
      dataIndex: 'open_id',
      render: (value: string) => <Typography.Text copyable>{value}</Typography.Text>,
    },
    { title: '订阅数', dataIndex: 'source_count', width: 90 },
    { title: '首次登录', dataIndex: 'created_at', width: 180, render: formatTime },
    { title: '最近更新', dataIndex: 'updated_at', width: 180, render: formatTime },
  ];

  return (
    <Card
      title="用户订阅"
      extra={
        <Typography.Text type="secondary">
          {total} 个用户，点击行展开查看订阅源
        </Typography.Text>
      }
    >
      <Table
        rowKey="id"
        size="middle"
        loading={loading}
        columns={columns}
        dataSource={list}
        expandable={{
          rowExpandable: (record) => record.source_count > 0,
          expandedRowRender: (record) => (
            <Table
              rowKey="id"
              size="small"
              columns={sourceColumns}
              dataSource={record.sources}
              pagination={false}
            />
          ),
        }}
        pagination={{
          current: page,
          pageSize: size,
          total,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          onChange: (nextPage, nextSize) => {
            setPage(nextPage);
            setSize(nextSize);
          },
        }}
      />
    </Card>
  );
}
