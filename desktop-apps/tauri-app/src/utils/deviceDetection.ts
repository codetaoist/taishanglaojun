import { DeviceType } from '../types/menu';

// 设备信息接口
export interface DeviceInfo {
  type: DeviceType;
  screenSize: {
    width: number;
    height: number;
  };
  userAgent: string;
  isTouchDevice: boolean;
  capabilities: string[];
  orientation?: 'portrait' | 'landscape';
}

// 检测设备类型
export function detectDeviceType(): DeviceType {
  const userAgent = navigator.userAgent.toLowerCase();
  const screenWidth = window.screen.width;
  const screenHeight = window.screen.height;
  const minDimension = Math.min(screenWidth, screenHeight);
  const maxDimension = Math.max(screenWidth, screenHeight);

  // 检测是否为手表设备
  if (minDimension <= 300 && maxDimension <= 400) {
    return DeviceType.WATCH;
  }

  // 检测移动设备
  if (/android|iphone|ipod|blackberry|iemobile|opera mini/i.test(userAgent)) {
    return DeviceType.MOBILE;
  }

  // 检测平板设备
  if (/ipad|android(?!.*mobile)|tablet/i.test(userAgent) || 
      (minDimension >= 768 && minDimension <= 1024)) {
    return DeviceType.TABLET;
  }

  // 默认为桌面设备
  return DeviceType.DESKTOP;
}

// 获取完整设备信息
export function getDeviceInfo(): DeviceInfo {
  const deviceType = detectDeviceType();
  const screenWidth = window.screen.width;
  const screenHeight = window.screen.height;
  const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0;

  // 根据设备类型确定功能支持
  const capabilities: string[] = ['basic_ui'];

  switch (deviceType) {
    case DeviceType.DESKTOP:
      capabilities.push('keyboard_shortcuts', 'multi_window', 'drag_drop', 'right_click');
      break;
    case DeviceType.TABLET:
      capabilities.push('touch_gestures', 'rotation', 'split_screen');
      break;
    case DeviceType.MOBILE:
      capabilities.push('touch_gestures', 'rotation', 'camera', 'gps');
      break;
    case DeviceType.WATCH:
      capabilities.push('touch_gestures', 'voice_input', 'haptic_feedback');
      break;
  }

  if (isTouchDevice) {
    capabilities.push('touch_input');
  }

  return {
    type: deviceType,
    screenSize: {
      width: screenWidth,
      height: screenHeight
    },
    userAgent: navigator.userAgent,
    isTouchDevice,
    capabilities,
    orientation: screenWidth > screenHeight ? 'landscape' : 'portrait'
  };
}

// 检查设备是否支持特定功能
export function supportsCapability(capability: string): boolean {
  const deviceInfo = getDeviceInfo();
  return deviceInfo.capabilities.includes(capability);
}

// 获取设备特定的CSS类名
export function getDeviceClasses(): string[] {
  const deviceInfo = getDeviceInfo();
  const classes = [`device-${deviceInfo.type}`];

  if (deviceInfo.isTouchDevice) {
    classes.push('touch-device');
  }

  if (deviceInfo.orientation) {
    classes.push(`orientation-${deviceInfo.orientation}`);
  }

  // 根据屏幕尺寸添加类名
  const { width } = deviceInfo.screenSize;
  if (width <= 480) {
    classes.push('screen-xs');
  } else if (width <= 768) {
    classes.push('screen-sm');
  } else if (width <= 1024) {
    classes.push('screen-md');
  } else if (width <= 1440) {
    classes.push('screen-lg');
  } else {
    classes.push('screen-xl');
  }

  return classes;
}

// 监听设备方向变化
export function onOrientationChange(callback: (orientation: 'portrait' | 'landscape') => void): () => void {
  const handleOrientationChange = () => {
    const newOrientation = window.innerWidth > window.innerHeight ? 'landscape' : 'portrait';
    callback(newOrientation);
  };

  window.addEventListener('resize', handleOrientationChange);
  window.addEventListener('orientationchange', handleOrientationChange);

  return () => {
    window.removeEventListener('resize', handleOrientationChange);
    window.removeEventListener('orientationchange', handleOrientationChange);
  };
}

// 获取推荐的UI配置
export function getRecommendedUIConfig(deviceType: DeviceType) {
  const configs = {
    [DeviceType.DESKTOP]: {
      sidebarWidth: 280,
      showSidebar: true,
      compactMode: false,
      maxMenuItems: 20,
      showIcons: true,
      showLabels: true,
      fontSize: 'medium'
    },
    [DeviceType.TABLET]: {
      sidebarWidth: 240,
      showSidebar: true,
      compactMode: false,
      maxMenuItems: 15,
      showIcons: true,
      showLabels: true,
      fontSize: 'medium'
    },
    [DeviceType.MOBILE]: {
      sidebarWidth: 0,
      showSidebar: false,
      compactMode: true,
      maxMenuItems: 10,
      showIcons: true,
      showLabels: false,
      fontSize: 'small'
    },
    [DeviceType.WATCH]: {
      sidebarWidth: 0,
      showSidebar: false,
      compactMode: true,
      maxMenuItems: 5,
      showIcons: true,
      showLabels: false,
      fontSize: 'small'
    }
  };

  return configs[deviceType];
}