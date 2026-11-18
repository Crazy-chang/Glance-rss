import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, Popconfirm, Space, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { AdminLog } from '../types';

function formatTime(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}

export default function System() {
  const [logs, setLogs] = useState<AdminLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [resetting, setResetting] = useState(false);

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

  const reset = async () => {
    setResetting(true);
    try {
      await api.resetAll();
      message.success('已恢复初始设置');
      load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '恢复失败');
    } finally {
      setResetting(false);
    }
  };

  const columns: ColumnsType<AdminLog> = [
    { title: '时间', dataIndex: 'created_at', width: 190, render: formatTime },
    { title: '管理员', dataIndex: 'username', width: 120 },
    { title: '操作', dataIndex: 'action', width: 140 },
    { title: '详情', dataIndex: 'detail', ellipsis: true },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card title="恢复初始设置">
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="将清空现有全部分类与推荐源，并按内置数据重建，操作不可撤销。"
        />
        <Popconfirm
          title="确定恢复初始设置吗？"
          description="现有推荐分类与推荐源会被清空并重建"
          okText="确定恢复"
          cancelText="取消"
          okButtonProps={{ danger: true }}
          onConfirm={reset}
        >
          <Button danger loading={resetting}>
            恢复初始设置
          </Button>
        </Popconfirm>
      </Card>

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
    </Space>
  );
}
