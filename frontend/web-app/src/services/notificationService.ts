import React from 'react';
import { App } from 'antd';

export interface NotificationService {
  success: (config: { message: string; description?: string; duration?: number }) => void;
  error: (config: { message: string; description?: string; duration?: number }) => void;
  warning: (config: { message: string; description?: string; duration?: number }) => void;
  info: (config: { message: string; description?: string; duration?: number }) => void;
}

let notificationInstance: NotificationService | null = null;

export const setNotificationInstance = (instance: NotificationService) => {
  notificationInstance = instance;
};

export const getNotificationInstance = (): NotificationService => {
  if (!notificationInstance) {
    // 如果没有设置实例，返回一个空的实现，避免错误
    console.warn('Notification service not initialized');
    return {
      success: (config) => console.log('Success:', config.message),
      error: (config) => console.error('Error:', config.message),
      warning: (config) => console.warn('Warning:', config.message),
      info: (config) => console.info('Info:', config.message),
    };
  }
  return notificationInstance;
};

// Hook to initialize notification service
export const useNotificationService = () => {
  const { notification } = App.useApp();
  
  // 设置notification实例
  React.useEffect(() => {
    setNotificationInstance(notification);
  }, [notification]);
  
  return notification;
};