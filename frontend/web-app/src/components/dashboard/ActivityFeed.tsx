import React from 'react';
import { Card, List, Avatar, Typography, Space, Tag } from 'antd';
import {
  ClockCircleOutlined,
  UserOutlined,
  MessageOutlined,
  BookOutlined,
  ProjectOutlined,
  SecurityScanOutlined
} from '@ant-design/icons';

const { Text, Title } = Typography;

interface Activity {
  id: string;
  type: 'ai' | 'wisdom' | 'project' | 'security';
  user: string;
  action: string;
  target: string;
  time: string;
}

interface ActivityFeedProps {
  className?: string;
}

const ActivityFeed: React.FC<ActivityFeedProps> = ({ className }) => {
  const activities: Activity[] = [
    {
      id: '1',
      type: 'ai',
      user: '张三',
      action: '完成了AI对话',
      target: '智能助手咨询',
      time: '刚刚'
    },
    {
      id: '2',
      type: 'project',
      user: '李四',
      action: '更新了项目',
      target: '太上老君系统',
      time: '5分钟前'
    },
    {
      id: '3',
      type: 'wisdom',
      user: '王五',
      action: '学习了课程',
      target: '智慧学习平台',
      time: '10分钟前'
    },
    {
      id: '4',
      type: 'system',
      user: '系统',
      action: '系统更新完成',
      target: 'AI模块 v2.1.0',
      time: '1小时前'
    }
  ];

  const getActivityIcon = (type: string) => {
    const iconMap = {
      ai: <MessageOutlined />,
      wisdom: <BookOutlined />,
      project: <ProjectOutlined />,
      system: <BellOutlined />
    };
    return iconMap[type as keyof typeof iconMap] || <InfoCircleOutlined />;
  };



  return (
    <div className={className}>
      <Card title="最近活动">
        <List
          dataSource={activities}
          renderItem={(activity) => (
            <List.Item>
              <List.Item.Meta
                avatar={<Avatar icon={getActivityIcon(activity.type)} />}
                title={
                  <Space>
                    <Text strong>{activity.user}</Text>
                    <Text type="secondary">{activity.action}</Text>
                    <Text>{activity.target}</Text>
                  </Space>
                }
                description={activity.time}
              />
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
};

export default ActivityFeed;