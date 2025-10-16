// 简单的环境检测工具
// 避免循环依赖，只包含基本的环境检测功能

export const isTauriEnvironment = (): boolean => {
  return typeof window !== 'undefined' && '__TAURI__' in window;
};

export const isWebEnvironment = (): boolean => {
  return !isTauriEnvironment();
};

export const isDevelopment = (): boolean => {
  return process.env.NODE_ENV === 'development';
};

export const isProduction = (): boolean => {
  return process.env.NODE_ENV === 'production';
};