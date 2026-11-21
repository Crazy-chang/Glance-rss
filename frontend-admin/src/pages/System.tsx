import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { AdminLog } from '../types';

function formatTime(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}

export default function System() {
  const [logs, setLogs] = useState<AdminLog[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setLogs(await api.logs(100));
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载日志失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const columns: ColumnsType<AdminLog> = [
    { title: '时间', dataIndex: 'created_at', width: 190, render: formatTime },
    { title: '管理员', dataIndex: 'username', width: 120 },
    { title: '操作', dataIndex: 'action', width: 140 },
    { title: '详情', dataIndex: 'detail', ellipsis: true },
  ];

  return (
    <Card
      title="操作日志"
      extra={
        <Button onClick={load} loading={loading}>
          刷新
        </Button>
      }
    >
      <Table
        rowKey="id"
        size="middle"
        loading={loading}
        columns={columns}
        dataSource={logs}
        pagination={{ pageSize: 20 }}
      />
    </Card>
  );
}
