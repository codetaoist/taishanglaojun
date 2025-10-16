import { theme } from 'antd'
import type { ThemeConfig } from 'antd'

// 主色调配置
export const primaryColors = {
  gold: '#d4af37',      // 金色 - 太上老君主题色
  red: '#c41e3a',       // 中国红
  blue: '#1890ff',      // 蓝色
  green: '#52c41a',     // 绿色
  purple: '#722ed1',    // 紫色
}

// 浅色主题配置
export const lightTheme: ThemeConfig = {
  algorithm: theme.defaultAlgorithm,
  token: {
    // 主色调
    colorPrimary: primaryColors.gold,
    colorSuccess: primaryColors.green,
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: primaryColors.blue,
    
    // 背景色
    colorBgContainer: '#ffffff',
    colorBgElevated: '#ffffff',
    colorBgLayout: '#f5f5f5',
    colorBgSpotlight: '#ffffff',
    
    // 文字颜色
    colorText: '#000000d9',
    colorTextSecondary: '#00000073',
    colorTextTertiary: '#00000040',
    colorTextQuaternary: '#00000026',
    
    // 边框颜色
    colorBorder: '#d9d9d9',
    colorBorderSecondary: '#f0f0f0',
    
    // 链接颜色
    colorLink: primaryColors.gold,
    colorLinkHover: '#b8941f',
    colorLinkActive: '#9c7c1a',
    
    // 圆角
    borderRadius: 6,
    borderRadiusLG: 8,
    borderRadiusSM: 4,
    borderRadiusXS: 2,
    
    // 字体
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
    fontSize: 14,
    fontSizeLG: 16,
    fontSizeSM: 12,
    fontSizeXL: 20,
    
    // 阴影
    boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.02)',
    boxShadowSecondary: '0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
    
    // 动画
    motionDurationFast: '0.1s',
    motionDurationMid: '0.2s',
    motionDurationSlow: '0.3s',
    
    // 间距
    padding: 16,
    paddingLG: 24,
    paddingSM: 12,
    paddingXS: 8,
    paddingXXS: 4,
    
    margin: 16,
    marginLG: 24,
    marginSM: 12,
    marginXS: 8,
    marginXXS: 4,
    
    // 控件高度
    controlHeight: 32,
    controlHeightLG: 40,
    controlHeightSM: 24,
    controlHeightXS: 16,
    
    // 线条高度
    lineHeight: 1.5714285714285714,
    lineHeightLG: 1.5,
    lineHeightSM: 1.66,
    
    // Z-index
    zIndexBase: 0,
    zIndexPopupBase: 1000,
  },
  components: {
    Layout: {
      headerBg: '#ffffff',
      headerHeight: 64,
      headerPadding: '0 24px',
      siderBg: '#ffffff',
      triggerBg: '#ffffff',
      triggerColor: primaryColors.gold,
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: `${primaryColors.gold}1a`,
      itemSelectedColor: primaryColors.gold,
      itemHoverBg: `${primaryColors.gold}0d`,
      itemHoverColor: primaryColors.gold,
      itemActiveBg: `${primaryColors.gold}26`,
      subMenuItemBg: 'transparent',
      groupTitleColor: '#00000073',
    },
    Button: {
      primaryShadow: `0 2px 0 ${primaryColors.gold}26`,
      dangerShadow: '0 2px 0 rgba(255, 77, 79, 0.06)',
    },
    Card: {
      headerBg: 'transparent',
      actionsBg: '#fafafa',
    },
    Table: {
      headerBg: '#fafafa',
      headerSortActiveBg: '#f0f0f0',
      headerSortHoverBg: '#fafafa',
      bodySortBg: '#fafafa',
      rowHoverBg: '#fafafa',
      rowSelectedBg: `${primaryColors.gold}0d`,
      rowSelectedHoverBg: `${primaryColors.gold}1a`,
    },
    Form: {
      labelColor: '#000000d9',
      labelRequiredMarkColor: '#ff4d4f',
    },
    Input: {
      hoverBorderColor: primaryColors.gold,
      activeBorderColor: primaryColors.gold,
      activeShadow: `0 0 0 2px ${primaryColors.gold}26`,
    },
    Select: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionSelectedColor: primaryColors.gold,
      optionActiveBg: `${primaryColors.gold}0d`,
    },
    Tabs: {
      itemSelectedColor: primaryColors.gold,
      itemHoverColor: primaryColors.gold,
      itemActiveColor: primaryColors.gold,
      inkBarColor: primaryColors.gold,
    },
    Steps: {
      colorPrimary: primaryColors.gold,
      navArrowColor: primaryColors.gold,
    },
    Progress: {
      defaultColor: primaryColors.gold,
    },
    Badge: {
      colorBgContainer: primaryColors.gold,
    },
    Tag: {
      defaultBg: '#fafafa',
      defaultColor: '#000000d9',
    },
    Notification: {
      colorBgElevated: '#ffffff',
    },
    Message: {
      colorBgElevated: '#ffffff',
    },
    Modal: {
      headerBg: '#ffffff',
      contentBg: '#ffffff',
      footerBg: 'transparent',
    },
    Drawer: {
      colorBgElevated: '#ffffff',
    },
    Tooltip: {
      colorBgSpotlight: 'rgba(0, 0, 0, 0.85)',
    },
    Popover: {
      colorBgElevated: '#ffffff',
    },
    Calendar: {
      colorBgContainer: '#ffffff',
      itemActiveBg: `${primaryColors.gold}1a`,
    },
    DatePicker: {
      cellActiveWithRangeBg: `${primaryColors.gold}1a`,
      cellHoverWithRangeBg: `${primaryColors.gold}0d`,
    },
    TimePicker: {
      cellHoverBg: `${primaryColors.gold}0d`,
    },
    Upload: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#b8941f',
    },
    Switch: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#b8941f',
    },
    Radio: {
      colorPrimary: primaryColors.gold,
      buttonCheckedBg: primaryColors.gold,
    },
    Checkbox: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#b8941f',
    },
    Rate: {
      colorFillContent: primaryColors.gold,
    },
    Slider: {
      colorPrimary: primaryColors.gold,
      colorPrimaryBorder: primaryColors.gold,
      colorPrimaryBorderHover: '#b8941f',
    },
    Tree: {
      nodeSelectedBg: `${primaryColors.gold}1a`,
      nodeHoverBg: `${primaryColors.gold}0d`,
    },
    Transfer: {
      itemSelectedBg: `${primaryColors.gold}1a`,
      itemHoverBg: `${primaryColors.gold}0d`,
    },
    Cascader: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
    Mentions: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
    AutoComplete: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
  },
}

// 深色主题配置
export const darkTheme: ThemeConfig = {
  algorithm: theme.darkAlgorithm,
  token: {
    // 主色调
    colorPrimary: primaryColors.gold,
    colorSuccess: primaryColors.green,
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: primaryColors.blue,
    
    // 背景色
    colorBgContainer: '#141414',
    colorBgElevated: '#1f1f1f',
    colorBgLayout: '#000000',
    colorBgSpotlight: '#262626',
    
    // 文字颜色
    colorText: '#ffffffd9',
    colorTextSecondary: '#ffffff73',
    colorTextTertiary: '#ffffff40',
    colorTextQuaternary: '#ffffff26',
    
    // 边框颜色
    colorBorder: '#434343',
    colorBorderSecondary: '#303030',
    
    // 链接颜色
    colorLink: primaryColors.gold,
    colorLinkHover: '#e6c547',
    colorLinkActive: '#f0d157',
    
    // 圆角
    borderRadius: 6,
    borderRadiusLG: 8,
    borderRadiusSM: 4,
    borderRadiusXS: 2,
    
    // 字体
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
    fontSize: 14,
    fontSizeLG: 16,
    fontSizeSM: 12,
    fontSizeXL: 20,
    
    // 阴影
    boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.02)',
    boxShadowSecondary: '0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
    
    // 动画
    motionDurationFast: '0.1s',
    motionDurationMid: '0.2s',
    motionDurationSlow: '0.3s',
    
    // 间距
    padding: 16,
    paddingLG: 24,
    paddingSM: 12,
    paddingXS: 8,
    paddingXXS: 4,
    
    margin: 16,
    marginLG: 24,
    marginSM: 12,
    marginXS: 8,
    marginXXS: 4,
    
    // 控件高度
    controlHeight: 32,
    controlHeightLG: 40,
    controlHeightSM: 24,
    controlHeightXS: 16,
    
    // 线条高度
    lineHeight: 1.5714285714285714,
    lineHeightLG: 1.5,
    lineHeightSM: 1.66,
    
    // Z-index
    zIndexBase: 0,
    zIndexPopupBase: 1000,
  },
  components: {
    Layout: {
      headerBg: '#141414',
      headerHeight: 64,
      headerPadding: '0 24px',
      siderBg: '#141414',
      triggerBg: '#141414',
      triggerColor: primaryColors.gold,
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: `${primaryColors.gold}1a`,
      itemSelectedColor: primaryColors.gold,
      itemHoverBg: `${primaryColors.gold}0d`,
      itemHoverColor: primaryColors.gold,
      itemActiveBg: `${primaryColors.gold}26`,
      subMenuItemBg: 'transparent',
      groupTitleColor: '#ffffff73',
      darkItemBg: 'transparent',
      darkItemSelectedBg: `${primaryColors.gold}1a`,
      darkItemSelectedColor: primaryColors.gold,
      darkItemHoverBg: `${primaryColors.gold}0d`,
      darkItemHoverColor: primaryColors.gold,
      darkSubMenuItemBg: 'transparent',
      darkGroupTitleColor: '#ffffff73',
    },
    Button: {
      primaryShadow: `0 2px 0 ${primaryColors.gold}26`,
      dangerShadow: '0 2px 0 rgba(255, 77, 79, 0.06)',
    },
    Card: {
      headerBg: 'transparent',
      actionsBg: '#262626',
    },
    Table: {
      headerBg: '#262626',
      headerSortActiveBg: '#303030',
      headerSortHoverBg: '#262626',
      bodySortBg: '#262626',
      rowHoverBg: '#262626',
      rowSelectedBg: `${primaryColors.gold}0d`,
      rowSelectedHoverBg: `${primaryColors.gold}1a`,
    },
    Form: {
      labelColor: '#ffffffd9',
      labelRequiredMarkColor: '#ff4d4f',
    },
    Input: {
      hoverBorderColor: primaryColors.gold,
      activeBorderColor: primaryColors.gold,
      activeShadow: `0 0 0 2px ${primaryColors.gold}26`,
    },
    Select: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionSelectedColor: primaryColors.gold,
      optionActiveBg: `${primaryColors.gold}0d`,
    },
    Tabs: {
      itemSelectedColor: primaryColors.gold,
      itemHoverColor: primaryColors.gold,
      itemActiveColor: primaryColors.gold,
      inkBarColor: primaryColors.gold,
    },
    Steps: {
      colorPrimary: primaryColors.gold,
      navArrowColor: primaryColors.gold,
    },
    Progress: {
      defaultColor: primaryColors.gold,
    },
    Badge: {
      colorBgContainer: primaryColors.gold,
    },
    Tag: {
      defaultBg: '#262626',
      defaultColor: '#ffffffd9',
    },
    Notification: {
      colorBgElevated: '#1f1f1f',
    },
    Message: {
      colorBgElevated: '#1f1f1f',
    },
    Modal: {
      headerBg: '#1f1f1f',
      contentBg: '#1f1f1f',
      footerBg: 'transparent',
    },
    Drawer: {
      colorBgElevated: '#1f1f1f',
    },
    Tooltip: {
      colorBgSpotlight: 'rgba(0, 0, 0, 0.85)',
    },
    Popover: {
      colorBgElevated: '#1f1f1f',
    },
    Calendar: {
      colorBgContainer: '#141414',
      itemActiveBg: `${primaryColors.gold}1a`,
    },
    DatePicker: {
      cellActiveWithRangeBg: `${primaryColors.gold}1a`,
      cellHoverWithRangeBg: `${primaryColors.gold}0d`,
    },
    TimePicker: {
      cellHoverBg: `${primaryColors.gold}0d`,
    },
    Upload: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#e6c547',
    },
    Switch: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#e6c547',
    },
    Radio: {
      colorPrimary: primaryColors.gold,
      buttonCheckedBg: primaryColors.gold,
    },
    Checkbox: {
      colorPrimary: primaryColors.gold,
      colorPrimaryHover: '#e6c547',
    },
    Rate: {
      colorFillContent: primaryColors.gold,
    },
    Slider: {
      colorPrimary: primaryColors.gold,
      colorPrimaryBorder: primaryColors.gold,
      colorPrimaryBorderHover: '#e6c547',
    },
    Tree: {
      nodeSelectedBg: `${primaryColors.gold}1a`,
      nodeHoverBg: `${primaryColors.gold}0d`,
    },
    Transfer: {
      itemSelectedBg: `${primaryColors.gold}1a`,
      itemHoverBg: `${primaryColors.gold}0d`,
    },
    Cascader: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
    Mentions: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
    AutoComplete: {
      optionSelectedBg: `${primaryColors.gold}1a`,
      optionHoverBg: `${primaryColors.gold}0d`,
    },
  },
}

// 主题工具函数
export const getThemeConfig = (isDark: boolean): ThemeConfig => {
  return isDark ? darkTheme : lightTheme
}

// 主题颜色工具函数
export const getThemeColors = (isDark: boolean) => {
  return {
    primary: primaryColors.gold,
    success: primaryColors.green,
    warning: '#faad14',
    error: '#ff4d4f',
    info: primaryColors.blue,
    background: isDark ? '#000000' : '#ffffff',
    surface: isDark ? '#141414' : '#ffffff',
    text: isDark ? '#ffffffd9' : '#000000d9',
    textSecondary: isDark ? '#ffffff73' : '#00000073',
    border: isDark ? '#434343' : '#d9d9d9',
  }
}

// CSS 变量生成
export const generateCSSVariables = (isDark: boolean) => {
  const colors = getThemeColors(isDark)
  return {
    '--color-primary': colors.primary,
    '--color-success': colors.success,
    '--color-warning': colors.warning,
    '--color-error': colors.error,
    '--color-info': colors.info,
    '--color-background': colors.background,
    '--color-surface': colors.surface,
    '--color-text': colors.text,
    '--color-text-secondary': colors.textSecondary,
    '--color-border': colors.border,
  }
}

export default {
  lightTheme,
  darkTheme,
  primaryColors,
  getThemeConfig,
  getThemeColors,
  generateCSSVariables,
}