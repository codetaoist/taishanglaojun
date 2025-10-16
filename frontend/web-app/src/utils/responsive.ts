// 响应式断点配置
export const breakpoints = {
  xs: 480,    // 超小屏幕
  sm: 576,    // 小屏幕
  md: 768,    // 中等屏幕
  lg: 992,    // 大屏幕
  xl: 1200,   // 超大屏幕
  xxl: 1600,  // 超超大屏幕
} as const

export type Breakpoint = keyof typeof breakpoints

// 媒体查询字符串生成
export const mediaQueries = {
  xs: `(max-width: ${breakpoints.xs - 1}px)`,
  sm: `(min-width: ${breakpoints.xs}px) and (max-width: ${breakpoints.sm - 1}px)`,
  md: `(min-width: ${breakpoints.sm}px) and (max-width: ${breakpoints.md - 1}px)`,
  lg: `(min-width: ${breakpoints.md}px) and (max-width: ${breakpoints.lg - 1}px)`,
  xl: `(min-width: ${breakpoints.lg}px) and (max-width: ${breakpoints.xl - 1}px)`,
  xxl: `(min-width: ${breakpoints.xl}px)`,
  
  // 最小宽度查询
  minXs: `(min-width: ${breakpoints.xs}px)`,
  minSm: `(min-width: ${breakpoints.sm}px)`,
  minMd: `(min-width: ${breakpoints.md}px)`,
  minLg: `(min-width: ${breakpoints.lg}px)`,
  minXl: `(min-width: ${breakpoints.xl}px)`,
  minXxl: `(min-width: ${breakpoints.xxl}px)`,
  
  // 最大宽度查询
  maxXs: `(max-width: ${breakpoints.xs - 1}px)`,
  maxSm: `(max-width: ${breakpoints.sm - 1}px)`,
  maxMd: `(max-width: ${breakpoints.md - 1}px)`,
  maxLg: `(max-width: ${breakpoints.lg - 1}px)`,
  maxXl: `(max-width: ${breakpoints.xl - 1}px)`,
  maxXxl: `(max-width: ${breakpoints.xxl - 1}px)`,
} as const

// 获取当前屏幕断点
export const getCurrentBreakpoint = (): Breakpoint => {
  const width = window.innerWidth
  
  if (width < breakpoints.xs) return 'xs'
  if (width < breakpoints.sm) return 'sm'
  if (width < breakpoints.md) return 'md'
  if (width < breakpoints.lg) return 'lg'
  if (width < breakpoints.xl) return 'xl'
  return 'xxl'
}

// 检查是否为移动设备
export const isMobile = (): boolean => {
  return window.innerWidth < breakpoints.md
}

// 检查是否为平板设备
export const isTablet = (): boolean => {
  return window.innerWidth >= breakpoints.md && window.innerWidth < breakpoints.lg
}

// 检查是否为桌面设备
export const isDesktop = (): boolean => {
  return window.innerWidth >= breakpoints.lg
}

// 检查是否匹配指定断点
export const matchBreakpoint = (breakpoint: Breakpoint): boolean => {
  const width = window.innerWidth
  
  switch (breakpoint) {
    case 'xs':
      return width < breakpoints.xs
    case 'sm':
      return width >= breakpoints.xs && width < breakpoints.sm
    case 'md':
      return width >= breakpoints.sm && width < breakpoints.md
    case 'lg':
      return width >= breakpoints.md && width < breakpoints.lg
    case 'xl':
      return width >= breakpoints.lg && width < breakpoints.xl
    case 'xxl':
      return width >= breakpoints.xl
    default:
      return false
  }
}

// 检查是否大于等于指定断点
export const matchMinBreakpoint = (breakpoint: Breakpoint): boolean => {
  return window.innerWidth >= breakpoints[breakpoint]
}

// 检查是否小于指定断点
export const matchMaxBreakpoint = (breakpoint: Breakpoint): boolean => {
  return window.innerWidth < breakpoints[breakpoint]
}

// 响应式值配置类型
export interface ResponsiveValue<T> {
  xs?: T
  sm?: T
  md?: T
  lg?: T
  xl?: T
  xxl?: T
}

// 获取响应式值
export const getResponsiveValue = <T>(
  responsiveValue: ResponsiveValue<T> | T,
  currentBreakpoint?: Breakpoint
): T | undefined => {
  if (typeof responsiveValue !== 'object' || responsiveValue === null) {
    return responsiveValue as T
  }
  
  const bp = currentBreakpoint || getCurrentBreakpoint()
  const breakpointOrder: Breakpoint[] = ['xs', 'sm', 'md', 'lg', 'xl', 'xxl']
  const currentIndex = breakpointOrder.indexOf(bp)
  
  // 从当前断点向下查找值
  for (let i = currentIndex; i >= 0; i--) {
    const key = breakpointOrder[i]
    if (responsiveValue[key] !== undefined) {
      return responsiveValue[key]
    }
  }
  
  // 如果没有找到，向上查找
  for (let i = currentIndex + 1; i < breakpointOrder.length; i++) {
    const key = breakpointOrder[i]
    if (responsiveValue[key] !== undefined) {
      return responsiveValue[key]
    }
  }
  
  return undefined
}

// 响应式网格配置
export interface GridConfig {
  gutter?: ResponsiveValue<number> | number
  span?: ResponsiveValue<number> | number
  offset?: ResponsiveValue<number> | number
  order?: ResponsiveValue<number> | number
  push?: ResponsiveValue<number> | number
  pull?: ResponsiveValue<number> | number
}

// 获取网格配置
export const getGridConfig = (config: GridConfig, currentBreakpoint?: Breakpoint) => {
  const bp = currentBreakpoint || getCurrentBreakpoint()
  
  return {
    gutter: getResponsiveValue(config.gutter, bp),
    span: getResponsiveValue(config.span, bp),
    offset: getResponsiveValue(config.offset, bp),
    order: getResponsiveValue(config.order, bp),
    push: getResponsiveValue(config.push, bp),
    pull: getResponsiveValue(config.pull, bp),
  }
}

// 响应式间距配置
export interface SpacingConfig {
  margin?: ResponsiveValue<number | string> | number | string
  marginTop?: ResponsiveValue<number | string> | number | string
  marginRight?: ResponsiveValue<number | string> | number | string
  marginBottom?: ResponsiveValue<number | string> | number | string
  marginLeft?: ResponsiveValue<number | string> | number | string
  padding?: ResponsiveValue<number | string> | number | string
  paddingTop?: ResponsiveValue<number | string> | number | string
  paddingRight?: ResponsiveValue<number | string> | number | string
  paddingBottom?: ResponsiveValue<number | string> | number | string
  paddingLeft?: ResponsiveValue<number | string> | number | string
}

// 获取间距配置
export const getSpacingConfig = (config: SpacingConfig, currentBreakpoint?: Breakpoint) => {
  const bp = currentBreakpoint || getCurrentBreakpoint()
  
  return {
    margin: getResponsiveValue(config.margin, bp),
    marginTop: getResponsiveValue(config.marginTop, bp),
    marginRight: getResponsiveValue(config.marginRight, bp),
    marginBottom: getResponsiveValue(config.marginBottom, bp),
    marginLeft: getResponsiveValue(config.marginLeft, bp),
    padding: getResponsiveValue(config.padding, bp),
    paddingTop: getResponsiveValue(config.paddingTop, bp),
    paddingRight: getResponsiveValue(config.paddingRight, bp),
    paddingBottom: getResponsiveValue(config.paddingBottom, bp),
    paddingLeft: getResponsiveValue(config.paddingLeft, bp),
  }
}

// 响应式字体大小配置
export const fontSizes = {
  xs: {
    h1: 20,
    h2: 18,
    h3: 16,
    h4: 14,
    h5: 12,
    h6: 11,
    body: 12,
    caption: 10,
  },
  sm: {
    h1: 24,
    h2: 20,
    h3: 18,
    h4: 16,
    h5: 14,
    h6: 12,
    body: 14,
    caption: 12,
  },
  md: {
    h1: 28,
    h2: 24,
    h3: 20,
    h4: 18,
    h5: 16,
    h6: 14,
    body: 14,
    caption: 12,
  },
  lg: {
    h1: 32,
    h2: 28,
    h3: 24,
    h4: 20,
    h5: 18,
    h6: 16,
    body: 16,
    caption: 14,
  },
  xl: {
    h1: 36,
    h2: 32,
    h3: 28,
    h4: 24,
    h5: 20,
    h6: 18,
    body: 16,
    caption: 14,
  },
  xxl: {
    h1: 40,
    h2: 36,
    h3: 32,
    h4: 28,
    h5: 24,
    h6: 20,
    body: 18,
    caption: 16,
  },
} as const

export type FontSizeType = keyof typeof fontSizes.md

// 获取响应式字体大小
export const getResponsiveFontSize = (
  type: FontSizeType,
  currentBreakpoint?: Breakpoint
): number => {
  const bp = currentBreakpoint || getCurrentBreakpoint()
  return fontSizes[bp][type]
}

// 响应式容器最大宽度
export const containerMaxWidths = {
  xs: '100%',
  sm: '540px',
  md: '720px',
  lg: '960px',
  xl: '1140px',
  xxl: '1320px',
} as const

// 获取容器最大宽度
export const getContainerMaxWidth = (currentBreakpoint?: Breakpoint): string => {
  const bp = currentBreakpoint || getCurrentBreakpoint()
  return containerMaxWidths[bp]
}

// 响应式工具类生成器
export const generateResponsiveClasses = () => {
  const classes: Record<string, string> = {}
  
  Object.entries(breakpoints).forEach(([key, value]) => {
    // 隐藏类
    classes[`hidden-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { display: none !important; }`
    classes[`visible-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { display: block !important; }`
    
    // 文本对齐
    classes[`text-left-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { text-align: left !important; }`
    classes[`text-center-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { text-align: center !important; }`
    classes[`text-right-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { text-align: right !important; }`
    
    // 浮动
    classes[`float-left-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { float: left !important; }`
    classes[`float-right-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { float: right !important; }`
    classes[`float-none-${key}`] = `@media ${mediaQueries[key as Breakpoint]} { float: none !important; }`
  })
  
  return classes
}

// 设备检测
export const deviceDetection = {
  // 检测是否为触摸设备
  isTouchDevice: (): boolean => {
    return 'ontouchstart' in window || navigator.maxTouchPoints > 0
  },
  
  // 检测是否为移动设备（基于用户代理）
  isMobileDevice: (): boolean => {
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
  },
  
  // 检测是否为iOS设备
  isIOS: (): boolean => {
    return /iPad|iPhone|iPod/.test(navigator.userAgent)
  },
  
  // 检测是否为Android设备
  isAndroid: (): boolean => {
    return /Android/.test(navigator.userAgent)
  },
  
  // 检测是否支持Retina显示
  isRetina: (): boolean => {
    return window.devicePixelRatio > 1
  },
  
  // 获取设备像素比
  getDevicePixelRatio: (): number => {
    return window.devicePixelRatio || 1
  },
  
  // 检测是否支持WebP
  supportsWebP: (): Promise<boolean> => {
    return new Promise((resolve) => {
      const webP = new Image()
      webP.onload = webP.onerror = () => {
        resolve(webP.height === 2)
      }
      webP.src = 'data:image/webp;base64,UklGRjoAAABXRUJQVlA4IC4AAACyAgCdASoCAAIALmk0mk0iIiIiIgBoSygABc6WWgAA/veff/0PP8bA//LwYAAA'
    })
  },
  
  // 检测网络连接类型
  getConnectionType: (): string => {
    const connection = (navigator as any).connection || (navigator as any).mozConnection || (navigator as any).webkitConnection
    return connection ? connection.effectiveType || connection.type || 'unknown' : 'unknown'
  },
  
  // 检测是否为慢速网络
  isSlowNetwork: (): boolean => {
    const connection = (navigator as any).connection || (navigator as any).mozConnection || (navigator as any).webkitConnection
    if (!connection) return false
    
    const slowConnections = ['slow-2g', '2g', '3g']
    return slowConnections.includes(connection.effectiveType)
  },
}

export default {
  breakpoints,
  mediaQueries,
  getCurrentBreakpoint,
  isMobile,
  isTablet,
  isDesktop,
  matchBreakpoint,
  matchMinBreakpoint,
  matchMaxBreakpoint,
  getResponsiveValue,
  getGridConfig,
  getSpacingConfig,
  fontSizes,
  getResponsiveFontSize,
  containerMaxWidths,
  getContainerMaxWidth,
  generateResponsiveClasses,
  deviceDetection,
}