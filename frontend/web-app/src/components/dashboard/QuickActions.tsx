import React from 'react';
import { Card, Row, Col, Button, Typography, Space } from 'antd';
import {
  MessageOutlined,
  BookOutlined,
  ProjectOutlined,
  SecurityScanOutlined,
  ReadOutlined,
  SettingOutlined,
  BarChartOutlined,
  ThunderboltOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

interface QuickAction {
  id: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  path: string;
}

interface QuickActionsProps {
  className?: string;
}

const QuickActions: React.FC<QuickActionsProps> = ({ className }) => {
  const navigate = useNavigate();

  const quickActions: QuickAction[] = [
    {
      id: 'ai-chat',
      title: 'AI智能对话',
      description: '与AI助手进行深度对话，获得智慧指导',
      icon: <MessageOutlined />,
      path: '/chat'
    },
    {
      id: 'ai-analysis',
      title: 'AI智能分析',
      description: '利用AI进行数据分析和洞察',
      icon: <BarChartOutlined />,
      path: '/ai/analysis'
    },
    {
      id: 'wisdom-library',
      title: '智慧宝库',
      description: '探索古今中外的智慧精华',
      icon: <BookOutlined />,
      path: '/wisdom'
    },
    {
      id: 'project-workspace',
      title: '项目工作台',
      description: '管理和协作您的项目',
      icon: <ProjectOutlined />,
      path: '/projects/workspace'
    },
    {
      id: 'security-center',
      title: '安全中心',
      description: '监控和保护系统安全',
      icon: <SecurityScanOutlined />,
      path: '/security'
    },
    {
      id: 'learning-courses',
      title: '学习课程',
      description: '智能化的学习体验',
      icon: <ReadOutlined />,
      path: '/learning/courses'
    },
    {
      id: 'system-settings',
      title: '系统设置',
      description: '配置和管理系统参数',
      icon: <SettingOutlined />,
      path: '/admin/settings'
    },
    {
      id: 'ai-assistant',
      title: 'AI助手',
      description: '智能助手为您提供全方位支持',
      icon: <ThunderboltOutlined />,
      path: '/ai/assistant'
    }
  ];

  const handleActionClick = (action: QuickAction) => {
    navigate(action.path);
  };

  return (
    <div className={className}>
      <Card 
        title={
          <Space>
            <Title level={4} style={{ margin: 0 }}>快捷操作</Title>
          </Space>
        }
      >
        <Row gutter={[16, 16]}>
          {quickActions.map((action) => (
            <Col xs={12} sm={8} lg={6} key={action.id}>
              <Button
                type="text"
                size="large"
                onClick={() => handleActionClick(action)}
                style={{
                  height: 'auto',
                  padding: '16px',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  width: '100%',
                  border: '1px solid #f0f0f0',
                  borderRadius: '8px'
                }}
                className="hover:border-primary hover:text-primary"
              >
                <div style={{ fontSize: '24px', marginBottom: '8px' }}>
                  {action.icon}
                </div>
                <Text style={{ fontSize: '12px', textAlign: 'center' }}>
                  {action.title}
                </Text>
              </Button>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );
};

export default QuickActions;