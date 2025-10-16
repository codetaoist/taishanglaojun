import React from 'react';
import { Drawer, List, Avatar, Button, Empty, Tag, Typography, Space, Tooltip } from 'antd';
import {
  BellOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  DeleteOutlined,
  CheckOutlined,
  ClearOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useNotifications } from '../../hooks';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';

dayjs.extend(relativeTime);
dayjs.locale('zh-cn');

const { Text, Paragraph } = Typography;

interface NotificationCenterProps {
  visible: boolean;
  onClose: () => void;
}

const typeIcons = {
  info: <InfoCircleOutlined className="text-blue-500" />,
  success: <CheckCircleOutlined className="text-green-500" />,
  warning: <ExclamationCircleOutlined className="text-orange-500" />,
  error: <CloseCircleOutlined className="text-red-500" />
};

const typeColors = {
  info: 'blue',
  success: 'green',
  warning: 'orange',
  error: 'red'
};

export const NotificationCenter: React.FC<NotificationCenterProps> = ({
  visible,
  onClose
}) => {
  const navigate = useNavigate();
  const { notifications, unreadCount, markRead, markAllRead, remove, clear } = useNotifications();

  // 处理通知点击
  const handleNotificationClick = (notification: any) => {
    if (!notification.read) {
      markRead(notification.id);
    }
    
    // 如果有操作按钮，可以在这里处理
    if (notification.actions && notification.actions.length > 0) {
      // 处理默认操作
      const defaultAction = notification.actions[0];
      if (defaultAction.action === 'navigate') {
        // 导航到指定页面
        navigate(defaultAction.url || '/');
      }
    }
  };

  // 处理操作按钮点击
  const handleActionClick = (action: any, notificationId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    
    switch (action.action) {
      case 'navigate':
        navigate(action.url || '/');
        break;
      case 'dismiss':
        remove(notificationId);
        break;
      case 'markRead':
        markRead(notificationId);
        break;
      default:
        console.log('Unknown action:', action);
    }
  };

  // 格式化时间
  const formatTime = (timestamp: number) => {
    try {
      return dayjs(timestamp).fromNow();
    } catch (error) {
      return '刚刚';
    }
  };

  // 渲染通知项
  const renderNotificationItem = (notification: any) => (
    <List.Item
      key={notification.id}
      className={`cursor-pointer transition-colors border-l-4 ${
        notification.read 
          ? 'border-l-gray-200 bg-white hover:bg-gray-50' 
          : `border-l-${typeColors[notification.type]}-500 bg-blue-50 hover:bg-blue-100`
      }`}
      onClick={() => handleNotificationClick(notification)}
      actions={[
        <Tooltip title="标记为已读" key="read">
          <Button
            type="text"
            size="small"
            icon={<CheckOutlined />}
            onClick={(e) => {
              e.stopPropagation();
              markRead(notification.id);
            }}
            disabled={notification.read}
          />
        </Tooltip>,
        <Tooltip title="删除" key="delete">
          <Button
            type="text"
            size="small"
            icon={<DeleteOutlined />}
            onClick={(e) => {
              e.stopPropagation();
              remove(notification.id);
            }}
            danger
          />
        </Tooltip>
      ]}
    >
      <List.Item.Meta
        avatar={
          <Avatar
            icon={typeIcons[notification.type]}
            style={{ 
              backgroundColor: notification.read ? '#f5f5f5' : '#fff',
              border: `2px solid ${notification.read ? '#d9d9d9' : typeColors[notification.type]}`
            }}
          />
        }
        title={
          <div className="flex items-center justify-between">
            <Text strong={!notification.read} className={notification.read ? 'text-gray-600' : 'text-gray-900'}>
              {notification.title}
            </Text>
            <div className="flex items-center gap-2">
              {!notification.read && (
                <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
              )}
              <Text type="secondary" className="text-xs">
                {formatTime(notification.timestamp)}
              </Text>
            </div>
          </div>
        }
        description={
          <div>
            <Paragraph
              className={`mb-2 ${notification.read ? 'text-gray-500' : 'text-gray-700'}`}
              ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
            >
              {notification.message}
            </Paragraph>
            
            {notification.actions && notification.actions.length > 0 && (
              <Space size="small" className="mt-2">
                {notification.actions.map((action: any, index: number) => (
                  <Button
                    key={index}
                    size="small"
                    type={index === 0 ? 'primary' : 'default'}
                    onClick={(e) => handleActionClick(action, notification.id, e)}
                  >
                    {action.label}
                  </Button>
                ))}
              </Space>
            )}
          </div>
        }
      />
    </List.Item>
  );

  return (
    <Drawer
      title={
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <BellOutlined />
            <span>通知中心</span>
            {unreadCount > 0 && (
              <Tag color="red" className="ml-2">
                {unreadCount} 条未读
              </Tag>
            )}
          </div>
        </div>
      }
      placement="right"
      onClose={onClose}
      open={visible}
      width={400}
      extra={
        <Space>
          {unreadCount > 0 && (
            <Button
              type="text"
              size="small"
              icon={<CheckOutlined />}
              onClick={markAllRead}
            >
              全部已读
            </Button>
          )}
          {notifications.length > 0 && (
            <Button
              type="text"
              size="small"
              icon={<ClearOutlined />}
              onClick={clear}
              danger
            >
              清空
            </Button>
          )}
        </Space>
      }
    >
      {notifications.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无通知"
          className="mt-8"
        />
      ) : (
        <List
          dataSource={notifications}
          renderItem={renderNotificationItem}
          className="notification-list"
          split={false}
        />
      )}
    </Drawer>
  );
};

export default NotificationCenter;