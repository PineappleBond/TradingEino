import { useState } from 'react';
import { Layout, Menu, theme } from 'antd';
import { useNavigate } from 'react-router-dom';
import {
  ClockCircleOutlined,
  FileTextOutlined,
  DashboardOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import './MainLayout.css';

const { Header, Sider, Content } = Layout;

type MenuItem = Required<MenuProps>['items'][number];

/**
 * 菜单项配置
 */
function getItem(
  label: string,
  key: string,
  icon?: React.ReactNode,
  children?: MenuItem[]
): MenuItem {
  return {
    key,
    icon,
    children,
    label,
  } as MenuItem;
}

/**
 * 菜单项定义 - 分组结构
 */
const menuItems: MenuItem[] = [
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: '仪表盘',
  },
  getItem('任务管理', 'task', <ClockCircleOutlined />, [
    {
      key: 'task/list',
      label: '定时任务',
    },
    {
      key: 'task/execution',
      label: '执行记录',
    },
  ]),
  getItem('日志中心', 'log', <FileTextOutlined />, [
    {
      key: 'log/execution',
      label: '执行日志',
    },
    {
      key: 'log/system',
      label: '系统日志',
    },
  ]),
];

/**
 * 主布局组件
 */
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  /**
   * 菜单点击处理
   */
  const handleMenuClick: MenuProps['onClick'] = (e) => {
    const key = e.key;
    // 路由跳转
    if (key === 'dashboard') {
      navigate('/task'); // 仪表盘重定向到任务列表
    } else if (key === 'task/list') {
      navigate('/task');
    } else if (key === 'task/execution') {
      navigate('/task/execution');
    } else if (key === 'log/execution') {
      navigate('/log/execution');
    } else if (key === 'log/system') {
      navigate('/log/system');
    }
  };

  /**
   * 根据当前路径获取默认选中的菜单项
   */
  const getSelectedKeys = (): string[] => {
    const pathname = window.location.pathname;
    if (pathname === '/' || pathname === '/task') {
      return ['task/list'];
    }
    if (pathname.startsWith('/task/execution')) {
      return ['task/execution'];
    }
    if (pathname.startsWith('/log/execution')) {
      return ['log/execution'];
    }
    if (pathname.startsWith('/log/system')) {
      return ['log/system'];
    }
    return [];
  };

  /**
   * 根据当前路径获取默认展开的菜单组
   */
  const getOpenKeys = (): string[] => {
    const pathname = window.location.pathname;
    if (pathname.startsWith('/task')) {
      return ['task'];
    }
    if (pathname.startsWith('/log')) {
      return ['log'];
    }
    return [];
  };

  return (
    <Layout className="main-layout">
      <Header className="main-header">
        <div className="header-logo">
          <span className="logo-text cyber-gradient-text">
            {collapsed ? 'TE' : 'TradingEino'}
          </span>
        </div>
      </Header>
      <Layout>
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          className="main-sider"
          width={220}
        >
          <div className="sider-menu-wrapper" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            <div style={{ padding: '16px', textAlign: 'right' }}>
              {collapsed ? (
                <MenuUnfoldOutlined
                  onClick={() => setCollapsed(!collapsed)}
                  style={{ fontSize: 18, cursor: 'pointer', color: '#bf00ff' }}
                />
              ) : (
                <MenuFoldOutlined
                  onClick={() => setCollapsed(!collapsed)}
                  style={{ fontSize: 18, cursor: 'pointer', color: '#bf00ff' }}
                />
              )}
            </div>
            <Menu
              theme="dark"
              mode="inline"
              defaultSelectedKeys={getSelectedKeys()}
              defaultOpenKeys={getOpenKeys()}
              items={menuItems}
              onClick={handleMenuClick}
              style={{ flex: 1, borderRight: 0 }}
            />
          </div>
        </Sider>
        <Content className="main-content">
          <div
            className="content-wrapper"
            style={{
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
            }}
          >
            {children}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
