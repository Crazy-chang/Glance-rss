import { useState } from 'react';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Navigate, useNavigate } from 'react-router-dom';
import { api, getToken, setToken } from '../api';

export default function Login() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  if (getToken()) {
    return <Navigate to="/categories" replace />;
  }

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const result = await api.login(values.username, values.password);
      setToken(result.token);
      message.success('登录成功');
      navigate('/categories', { replace: true });
    } catch (err) {
      message.error(err instanceof Error ? err.message : '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <Card className="login-card">
        <Typography.Title level={4} style={{ textAlign: 'center', marginBottom: 24 }}>
          Glance 管理后台
        </Typography.Title>
        <Form layout="vertical" onFinish={handleSubmit} initialValues={{ username: 'admin' }}>
          <Form.Item
            name="username"
            label="账号"
            rules={[{ required: true, message: '请输入账号' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="请输入账号" autoComplete="username" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              autoComplete="current-password"
            />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            登录
          </Button>
        </Form>
        <Typography.Text type="secondary" className="login-hint">
          默认账号：admin　密码：666666
        </Typography.Text>
      </Card>
    </div>
  );
}
