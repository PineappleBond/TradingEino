import type { ThemeConfig } from 'antd';

/**
 * 赛博朋克主题配置 - 经典霓虹紫蓝配色
 */
export const cyberpunkTheme: ThemeConfig = {
  token: {
    // 主色 - 霓虹紫
    colorPrimary: '#bf00ff',
    // 成功色 - 青色
    colorSuccess: '#00ffff',
    // 警告色 - 亮橙
    colorWarning: '#ff6b00',
    // 错误色 - 霓虹粉
    colorError: '#ff0080',
    // 信息色 - 电光蓝
    colorInfo: '#0080ff',
    // 背景色 - 紫黑（已提亮）
    colorBgLayout: '#1a1a25',
    // 容器背景 - 紫黑（已提亮）
    colorBgContainer: '#252535',
    // 弹出层背景（已提亮）
    colorBgElevated: '#303045',
    // 文本色 - 纯白
    colorText: '#ffffff',
    // 次要文本色 - 亮灰
    colorTextSecondary: '#c0c0c0',
    // 禁用文本色
    colorTextDisabled: '#808080',
    // 边框色 - 紫灰（已提亮）
    colorBorder: '#454555',
    // 圆角
    borderRadius: 6,
    // 字体
    fontFamily: '"JetBrains Mono", "Fira Code", Consolas, Monaco, monospace',
    // 字体大小
    fontSize: 14,
    // 链接色
    colorLink: '#00ffff',
    // 链接悬停色
    colorLinkHover: '#bf00ff',
    // 占位符颜色
    colorFillContent: '#252535',
    colorFillContentHover: '#303045',
  },
  components: {
    Layout: {
      headerBg: '#1a1a25',
      siderBg: '#1a1a25',
      footerBg: '#1a1a25',
      triggerBg: '#252535',
      triggerColor: '#bf00ff',
    },
    Menu: {
      darkItemBg: '#1a1a25',
      darkPopupBg: '#1a1a25',
      darkItemSelectedBg: 'rgba(191, 0, 255, 0.15)',
      darkItemColor: '#ffffff',
      darkItemHoverColor: '#00ffff',
      darkItemSelectedColor: '#bf00ff',
      darkSubMenuItemBg: '#1a1a25',
      itemMarginInline: 8,
      itemBorderRadius: 6,
    },
    Button: {
      primaryShadow: '0 0 10px rgba(191, 0, 255, 0.5)',
    },
    Table: {
      headerBg: '#252535',
      headerColor: '#bf00ff',
      borderColor: '#454555',
      colorText: '#ffffff',
      colorTextSecondary: '#c0c0c0',
      colorFillContent: '#252535',
      rowHoverBg: 'rgba(191, 0, 255, 0.08)',
    },
    Input: {
      activeBorderColor: '#bf00ff',
      activeShadow: '0 0 8px rgba(191, 0, 255, 0.3)',
      hoverBorderColor: '#00ffff',
      activeBg: '#252535',
      hoverBg: '#252535',
      colorText: '#ffffff',
      colorTextDisabled: '#808080',
      colorBgContainer: '#252535',
    },
    Select: {
      selectorBg: '#252535',
      optionActiveBg: '#303045',
      colorText: '#ffffff',
      colorTextSecondary: '#c0c0c0',
      colorTextDisabled: '#808080',
      optionSelectedBg: 'rgba(191, 0, 255, 0.15)',
      optionSelectedColor: '#ffffff',
      colorBorder: '#454555',
      colorBorderSecondary: '#454555',
    },
    Dropdown: {
      colorBgContainer: '#303045',
      colorText: '#ffffff',
    },
    Modal: {
      contentBg: '#303045',
      headerBg: '#252535',
      colorText: '#ffffff',
      colorTextSecondary: '#c0c0c0',
    },
    Card: {
      colorBgContainer: '#252535',
      colorText: '#ffffff',
      colorTextSecondary: '#c0c0c0',
    },
    Form: {
      colorTextHeading: '#ffffff',
      colorTextLabel: '#c0c0c0',
    },
    Tag: {
      colorText: '#ffffff',
    },
    Pagination: {
      colorText: '#ffffff',
      colorTextDisabled: '#808080',
      itemBg: '#252535',
      itemInputBg: '#252535',
    },
    Radio: {
      colorText: '#ffffff',
    },
    Checkbox: {
      colorText: '#ffffff',
    },
    DatePicker: {
      colorBgContainer: '#252535',
      colorText: '#ffffff',
    },
    Tree: {
      colorBgContainer: '#252535',
    },
  },
};
