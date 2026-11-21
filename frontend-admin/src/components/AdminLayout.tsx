import {
  AppstoreOutlined,
  LogoutOutlined,
  MobileOutlined,
  SettingOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { Button, Layout, Menu, Typography } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { api, clearToken } from '../api';

const { Header, Sider, Content } = Layout;

const menuItems = [
  {
    key: 'recommend',
    icon: <AppstoreOutlined />,
    label: '推荐管理',
    children: [
      { key: 'recommend/categories', label: '分类管理' },
      { key: 'recommend/sources', label: '推荐源管理' },
    ],
  },
  { key: 'users', icon: <TeamOutlined />, label: '用户订阅' },
  { key: 'miniapp', icon: <MobileOutlined />, label: '小程序设置' },
  { key: 'system', icon: <SettingOutlined />, label: '系统管理' },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const selected = location.pathname.replace(/^\//, '') || 'recommend/categories';

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
          defaultOpenKeys={['recommend']}
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
