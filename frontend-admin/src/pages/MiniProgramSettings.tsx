import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Card, InputNumber, Popconfirm, Space, Typography, message } from 'antd';
import { api } from '../api';

// 小程序端功能设置：一键阅读条数、恢复内置推荐数据
export default function MiniProgramSettings() {
  const [batchReadLimit, setBatchReadLimit] = useState(5);
  const [configLoading, setConfigLoading] = useState(false);
  const [configSaving, setConfigSaving] = useState(false);
  const [resetting, setResetting] = useState(false);

  const loadConfig = useCallback(async () => {
    setConfigLoading(true);
    try {
      const cfg = await api.getConfig();
      setBatchReadLimit(cfg.batch_read_limit);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载配置失败');
    } finally {
      setConfigLoading(false);
    }
  }, []);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const saveConfig = async () => {
    setConfigSaving(true);
    try {
      await api.updateConfig(batchReadLimit);
      message.success('配置已保存');
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存失败');
    } finally {
      setConfigSaving(false);
    }
  };

  const reset = async () => {
    setResetting(true);
    try {
      await api.resetAll();
      message.success('已恢复初始设置');
    } catch (err) {
      message.error(err instanceof Error ? err.message : '恢复失败');
    } finally {
      setResetting(false);
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card
        title="一键阅读"
        extra={
          <Button onClick={loadConfig} loading={configLoading}>
            刷新
          </Button>
        }
      >
        <Space direction="vertical" size={12}>
          <Typography.Text type="secondary">
            控制小程序首页「一键阅读」按钮一次标记已读的条数，默认 5 条；填 0 表示标记全部未读文章。
          </Typography.Text>
          <Space>
            <Typography.Text>每次标记条数</Typography.Text>
            <InputNumber
              min={0}
              max={100}
              value={batchReadLimit}
              onChange={(value) => setBatchReadLimit(typeof value === 'number' ? value : 0)}
              style={{ width: 120 }}
            />
            <Button type="primary" loading={configSaving} onClick={saveConfig}>
              保存
            </Button>
          </Space>
        </Space>
      </Card>

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
    </Space>
  );
}
