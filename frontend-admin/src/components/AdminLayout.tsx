import { AppstoreOutlined, LinkOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import { Button, Layout, Menu, Typography } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { api, clearToken } from '../api';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: 'categories', icon: <AppstoreOutlined />, label: '分类管理' },
  { key: 'sources', icon: <LinkOutlined />, label: '推荐源管理' },
  { key: 'system', icon: <SettingOutlined />, label: '系统管理' },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const selected = location.pathname.split('/')[1] || 'categories';

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch {
      // 退出接口失败也清除本地令牌
    }
    clearToken();
    navigate('/login', { replace: true });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={208} theme="light">
        <div className="brand">Glance 管理后台</div>
        <Menu
          mode="inline"
          selectedKeys={[selected]}
          items={menuItems}
          onClick={({ key }) => navigate(`/${key}`)}
        />
      </Sider>
      <Layout>
        <Header className="admin-header">
          <Typography.Text type="secondary">推荐分类与推荐源维护</Typography.Text>
          <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Button>
        </Header>
        <Content className="admin-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
