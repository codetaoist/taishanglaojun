import React from 'react';
import { Result, Button, Card, Typography, Space, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import { 
  RocketOutlined, 
  ToolOutlined, 
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

interface PlaceholderPageProps {
  title: string;
  description?: string;
  status: 'planned' | 'development' | 'partial' | 'completed';
  features?: string[];
  estimatedCompletion?: string;
  relatedPages?: Array<{
    title: string;
    path: string;
  }>;
}

const PlaceholderPage: React.FC<PlaceholderPageProps> = ({
  title,
  description,
  status,
  features = [],
  estimatedCompletion,
  relatedPages = []
}) => {
  const navigate = useNavigate();

  const getStatusConfig = () => {
    switch (status) {
      case 'completed':
        return {
          icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
          color: 'success',
          text: '已完成',
          subTitle: '此功能已完全开发完成'
        };
      case 'partial':
        return {
          icon: <ExclamationCircleOutlined style={{ color: '#faad14' }} />,
          color: 'warning',
          text: '部分完成',
          subTitle: '此功能已部分实现，正在持续完善中'
        };
      case 'development':
        return {
          icon: <ToolOutlined style={{ color: '#1890ff' }} />,
          color: 'processing',
          text: '开发中',
          subTitle: '此功能正在积极开发中'
        };
      case 'planned':
      default:
        return {
          icon: <ClockCircleOutlined style={{ color: '#8c8c8c' }} />,
          color: 'default',
          text: '规划中',
          subTitle: '此功能已列入开发计划'
        };
    }
  };

  const statusConfig = getStatusConfig();

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto">
        <Result
          icon={<RocketOutlined style={{ color: '#1890ff', fontSize: '72px' }} />}
          title={
            <Space direction="vertical" size="small">
              <Title level={2}>{title}</Title>
              <Tag color={statusConfig.color} icon={statusConfig.icon}>
                {statusConfig.text}
              </Tag>
            </Space>
          }
          subTitle={statusConfig.subTitle}
          extra={[
            <Button type="primary" key="home" onClick={() => navigate('/')}>
              返回首页
            </Button>,
            <Button key="back" onClick={() => navigate(-1)}>
              返回上页
            </Button>
          ]}
        />

        <div className="mt-8 space-y-6">
          {description && (
            <Card title="功能描述" className="shadow-sm">
              <Paragraph>{description}</Paragraph>
            </Card>
          )}

          {features.length > 0 && (
            <Card title="计划功能" className="shadow-sm">
              <ul className="space-y-2">
                {features.map((feature, index) => (
                  <li key={index} className="flex items-center space-x-2">
                    <span className="w-2 h-2 bg-blue-500 rounded-full"></span>
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
            </Card>
          )}

          {estimatedCompletion && (
            <Card title="预计完成时间" className="shadow-sm">
              <Paragraph>
                <ClockCircleOutlined className="mr-2" />
                {estimatedCompletion}
              </Paragraph>
            </Card>
          )}

          {relatedPages.length > 0 && (
            <Card title="相关功能" className="shadow-sm">
              <Space wrap>
                {relatedPages.map((page, index) => (
                  <Button 
                    key={index}
                    type="link" 
                    onClick={() => navigate(page.path)}
                  >
                    {page.title}
                  </Button>
                ))}
              </Space>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
};

export default PlaceholderPage;