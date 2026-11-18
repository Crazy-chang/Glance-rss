import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Flex, Modal, Space, Switch, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { AdminUser, AdminUserSource } from '../types';

function formatTime(value?: string | null) {
  if (!value) return '-';
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}

export default function Users() {
  const [list, setList] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [onlyWithSources, setOnlyWithSources] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [modalUser, setModalUser] = useState<AdminUser | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.users(page, size, onlyWithSources ? 1 : 0);
      setList(result.list);
      setTotal(result.total);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载用户失败');
    } finally {
      setLoading(false);
    }
  }, [page, size, onlyWithSources]);

  useEffect(() => {
    load();
  }, [load]);

  const openSources = (record: AdminUser) => {
    setModalUser(record);
    setModalOpen(true);
  };

  const copyFeed = async (source: AdminUserSource) => {
    try {
      await navigator.clipboard.writeText(source.feed_url);
      message.success(`已复制 Feed：${source.title}`);
    } catch {
      message.error('复制失败，请手动选择复制');
    }
  };

  const sourceColumns: ColumnsType<AdminUserSource> = [
    { title: '源 ID', dataIndex: 'id', width: 80 },
    { title: '名称', dataIndex: 'title', width: 140 },
    { title: '分类', dataIndex: 'category', width: 110, render: (value: string) => value || '-' },
    {
      title: 'Feed 地址',
      dataIndex: 'feed_url',
      width: 220,
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
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right',
      render: (_: unknown, record: AdminUserSource) => (
        <Button type="link" size="small" onClick={() => copyFeed(record)}>
          复制源
        </Button>
      ),
    },
  ];

  const columns: ColumnsType<AdminUser> = [
    { title: '用户 ID', dataIndex: 'id', width: 90 },
    { title: 'OpenID', dataIndex: 'open_id', width: 220 },
    { title: '订阅数', dataIndex: 'source_count', width: 90 },
    { title: '首次登录', dataIndex: 'created_at', width: 180, render: formatTime },
    { title: '最近更新', dataIndex: 'updated_at', width: 180, render: formatTime },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_: unknown, record) => (
        <Button
          type="link"
          size="small"
          disabled={record.source_count === 0}
          onClick={() => openSources(record)}
        >
          订阅 ({record.source_count})
        </Button>
      ),
    },
  ];

  return (
    <Card>
      <Flex justify="space-between" align="center" style={{ marginBottom: 16 }}>
        <Space size="large">
          <Space size={6}>
            <Switch
              checked={onlyWithSources}
              onChange={(checked) => {
                setOnlyWithSources(checked);
                setPage(1);
              }}
            />
            <Typography.Text>仅显示已订阅用户</Typography.Text>
          </Space>
          {/* 后续搜索条件挂这里：
              <Input.Search allowClear placeholder="搜索 OpenID" style={{ width: 220 }} />
          */}
        </Space>
        <Typography.Text type="secondary">共 {total} 个用户</Typography.Text>
      </Flex>

      <Table
        rowKey="id"
        size="middle"
        loading={loading}
        columns={columns}
        dataSource={list}
        scroll={{ x: 1000 }}
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
      <Modal
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        width={1000}
        title={
          modalUser
            ? `用户 #${modalUser.id}（${modalUser.open_id}）的订阅源 · ${modalUser.source_count} 个`
            : '订阅源'
        }
      >
        {modalUser && modalUser.sources.length > 0 ? (
          <Table
            rowKey="id"
            size="small"
            columns={sourceColumns}
            dataSource={modalUser.sources}
            pagination={false}
            scroll={{ y: 450 }}
          />
        ) : (
          <Typography.Text type="secondary">该用户暂未订阅任何源</Typography.Text>
        )}
      </Modal>
    </Card>
  );
}
