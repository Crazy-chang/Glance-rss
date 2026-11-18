import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  Typography,
  message,
} from 'antd';
import { ArrowDownOutlined, ArrowUpOutlined, PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { api } from '../api';
import type { Category } from '../types';

export default function Categories() {
  const [list, setList] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [form] = Form.useForm<{ name: string; description: string }>();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setList(await api.categories());
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (item: Category) => {
    setEditing(item);
    setModalOpen(true);
  };

  const submit = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editing) {
        await api.updateCategory({ id: editing.id, ...values });
      } else {
        await api.createCategory(values);
      }
      message.success('已保存');
      setModalOpen(false);
      load();
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
      await api.reorderCategories(next.map((item) => item.id));
    } catch (err) {
      message.error(err instanceof Error ? err.message : '排序失败');
      load();
    }
  };

  const remove = async (item: Category) => {
    try {
      await api.deleteCategory(item.id);
      message.success('已删除');
      load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '删除失败');
    }
  };

  const columns: ColumnsType<Category> = [
    { title: '排序', dataIndex: 'sort_order', width: 80 },
    { title: '分类名称', dataIndex: 'name', width: 160 },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '推荐源数量', dataIndex: 'source_count', width: 110 },
    {
      title: '操作',
      width: 260,
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
            title="确定删除该分类吗？"
            description="分类下还有推荐源时无法删除"
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
      title="分类管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新增分类
        </Button>
      }
    >
      <Typography.Paragraph type="secondary">
        分类顺序即发现页展示顺序；分类下仍有推荐源时不能删除分类。
      </Typography.Paragraph>
      <Table
        rowKey="id"
        size="middle"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={false}
      />

      <Modal
        title={editing ? '编辑分类' : '新增分类'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={submit}
        confirmLoading={saving}
        destroyOnClose
      >
        <Form
          key={editing ? `edit-${editing.id}` : 'create'}
          form={form}
          layout="vertical"
          preserve={false}
          initialValues={
            editing
              ? { name: editing.name, description: editing.description }
              : { name: '', description: '' }
          }
        >
          <Form.Item
            name="name"
            label="分类名称"
            rules={[{ required: true, message: '请输入分类名称' }]}
          >
            <Input maxLength={50} placeholder="例如：科技资讯" />
          </Form.Item>
          <Form.Item name="description" label="分类描述">
            <Input.TextArea maxLength={500} rows={3} placeholder="选填" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
