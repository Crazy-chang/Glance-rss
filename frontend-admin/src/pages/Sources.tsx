import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Typography,
  Upload,
  message,
} from 'antd';
import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  DownloadOutlined,
  PlusOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { Category, ExportData, Source, SourceInput } from '../types';

export default function Sources() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryId, setCategoryId] = useState<number>();
  const [list, setList] = useState<Source[]>([]);
  const [loading, setLoading] = useState(false);

  const [modalOpen, setModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<Source | null>(null);
  const [form] = Form.useForm<SourceInput>();

  const [importOpen, setImportOpen] = useState(false);
  const [importText, setImportText] = useState('');
  const [importing, setImporting] = useState(false);

  const loadCategories = useCallback(async () => {
    try {
      const data = await api.categories();
      setCategories(data);
      setCategoryId((current) => {
        if (current && data.some((item) => item.id === current)) {
          return current;
        }
        return data[0]?.id;
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载分类失败');
    }
  }, []);

  const loadSources = useCallback(async (id?: number) => {
    if (!id) {
      setList([]);
      return;
    }
    setLoading(true);
    try {
      setList(await api.sources(id));
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载推荐源失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  useEffect(() => {
    loadSources(categoryId);
  }, [categoryId, loadSources]);

  const openCreate = () => {
    if (!categoryId) {
      message.warning('请先在分类管理中创建分类');
      return;
    }
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (item: Source) => {
    setEditing(item);
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editing) {
        await api.updateSource(editing.id, values);
      } else {
        await api.createSource(values);
      }
      message.success('已保存');
      setModalOpen(false);
      if (values.category_id !== categoryId) {
        setCategoryId(values.category_id);
      } else {
        loadSources(categoryId);
      }
      loadCategories();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const move = async (index: number, offset: number) => {
    const target = index + offset;
    if (target < 0 || target >= list.length) {
      return;
    }
    const next = [...list];
    [next[index], next[target]] = [next[target], next[index]];
    setList(next);
    try {
      await api.reorderSources(next.map((item) => item.id));
    } catch (err) {
      message.error(err instanceof Error ? err.message : '排序失败');
      loadSources(categoryId);
    }
  };

  const remove = async (item: Source) => {
    try {
      await api.deleteSource(item.id);
      message.success('已删除');
      loadSources(categoryId);
      loadCategories();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '删除失败');
    }
  };

  const handleExport = async () => {
    try {
      const data = await api.exportData();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `glance-recommend-${new Date().toISOString().slice(0, 10)}.json`;
      link.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '导出失败');
    }
  };

  const handleImport = async () => {
    let data: ExportData;
    try {
      data = JSON.parse(importText);
    } catch {
      message.error('JSON 格式不正确');
      return;
    }
    if (!data || !Array.isArray(data.categories)) {
      message.error('数据缺少 categories 数组');
      return;
    }
    setImporting(true);
    try {
      await api.importData(data);
      message.success('导入成功，已整体替换推荐数据');
      setImportOpen(false);
      setImportText('');
      // 导入后分类会整体重建、ID 变化：刷新分类即可，列表由 categoryId 变化触发的 effect 重新加载
      await loadCategories();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '导入失败');
    } finally {
      setImporting(false);
    }
  };

  const handleFile = (file: File) => {
    const reader = new FileReader();
    reader.onload = () => setImportText(String(reader.result || ''));
    reader.readAsText(file);
    return false;
  };

  const columns: ColumnsType<Source> = [
    {
      title: '标题',
      dataIndex: 'title',
      width: 220,
      render: (value: string) => <Typography.Text strong>{value}</Typography.Text>,
    },
    {
      title: 'RSS 地址',
      dataIndex: 'feed_url',
      ellipsis: true,
      render: (value: string) => (
        <Typography.Text copyable={{ text: value }} ellipsis>
          {value}
        </Typography.Text>
      ),
    },
    {
      title: '热门',
      dataIndex: 'hot_rank',
      width: 80,
      render: (value: number) => (value > 0 ? value : '-'),
    },
    {
      title: '操作',
      width: 250,
      render: (_, item, index) => (
        <Space size={4}>
          <Button
            size="small"
            type="text"
            icon={<ArrowUpOutlined />}
            disabled={index === 0}
            onClick={() => move(index, -1)}
          />
          <Button
            size="small"
            type="text"
            icon={<ArrowDownOutlined />}
            disabled={index === list.length - 1}
            onClick={() => move(index, 1)}
          />
          <Button size="small" onClick={() => openEdit(item)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该推荐源吗？"
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => remove(item)}
          >
            <Button size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="推荐源管理"
      extra={
        <Space>
          <Button icon={<UploadOutlined />} onClick={() => setImportOpen(true)}>
            导入
          </Button>
          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            导出
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
            新增推荐源
          </Button>
        </Space>
      }
    >
      {categories.length === 0 ? (
        <Alert type="warning" showIcon message="还没有分类，请先到「分类管理」创建分类" />
      ) : (
        <>
          <Space style={{ marginBottom: 16 }}>
            <span>分类：</span>
            <Select
              style={{ width: 200 }}
              value={categoryId}
              options={categories.map((item) => ({ label: item.name, value: item.id }))}
              onChange={(value) => setCategoryId(value)}
            />
            <Typography.Text type="secondary">当前分类共 {list.length} 个推荐源</Typography.Text>
          </Space>
          <Table
            rowKey="id"
            size="middle"
            loading={loading}
            columns={columns}
            dataSource={list}
            pagination={false}
          />
        </>
      )}

      <Modal
        title={editing ? '编辑推荐源' : '新增推荐源'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={submit}
        confirmLoading={saving}
        width={640}
        destroyOnClose
      >
        <Form
          key={editing ? `edit-${editing.id}` : 'create'}
          form={form}
          layout="vertical"
          preserve={false}
          initialValues={
            editing
              ? {
                  title: editing.title,
                  feed_url: editing.feed_url,
                  site_url: editing.site_url,
                  description: editing.description,
                  category_id: editing.category_id,
                  hot_rank: editing.hot_rank,
                }
              : {
                  title: '',
                  feed_url: '',
                  site_url: '',
                  description: '',
                  category_id: categoryId,
                  hot_rank: 0,
                }
          }
        >
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input maxLength={200} placeholder="例如：36氪" />
          </Form.Item>
          <Form.Item
            name="feed_url"
            label="RSS 地址"
            rules={[
              { required: true, message: '请输入 RSS 地址' },
              { pattern: /^https?:\/\/.+/i, message: '需以 http:// 或 https:// 开头' },
            ]}
          >
            <Input placeholder="https://example.com/feed" />
          </Form.Item>
          <Form.Item
            name="site_url"
            label="站点地址"
            rules={[{ pattern: /^https?:\/\/.+/i, message: '需以 http:// 或 https:// 开头' }]}
          >
            <Input placeholder="https://example.com" />
          </Form.Item>
          <Form.Item
            name="category_id"
            label="所属分类"
            rules={[{ required: true, message: '请选择分类' }]}
          >
            <Select options={categories.map((item) => ({ label: item.name, value: item.id }))} />
          </Form.Item>
          <Form.Item
            name="hot_rank"
            label="热门排序"
            extra="大于 0 时进入发现页热门区，数字越小越靠前；填 0 表示不进入热门"
          >
            <InputNumber min={0} max={999} style={{ width: 160 }} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea maxLength={500} rows={3} placeholder="选填" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="导入推荐数据"
        open={importOpen}
        onCancel={() => setImportOpen(false)}
        onOk={handleImport}
        okText="导入并整体替换"
        okButtonProps={{ danger: true }}
        confirmLoading={importing}
        width={720}
        destroyOnClose
      >
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 12 }}
          message="导入会清空现有全部分类与推荐源，再按 JSON 内容重建，请先导出备份。"
        />
        <Space style={{ marginBottom: 12 }}>
          <Upload accept=".json,application/json" showUploadList={false} beforeUpload={handleFile}>
            <Button icon={<UploadOutlined />}>选择 JSON 文件</Button>
          </Upload>
          <Typography.Text type="secondary">
            格式：{'{ "categories": [{ "name": "", "sources": [{ "title": "", "feed_url": "" }] }] }'}
          </Typography.Text>
        </Space>
        <Input.TextArea
          rows={14}
          value={importText}
          onChange={(event) => setImportText(event.target.value)}
          placeholder="在这里粘贴 JSON"
        />
      </Modal>
    </Card>
  );
}
