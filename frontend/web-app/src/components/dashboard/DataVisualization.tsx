import React from 'react';
import { Card, Row, Col, Statistic, Typography, Space } from 'antd';
import {
  LineChartOutlined,
  BarChartOutlined,
  RiseOutlined,
  FallOutlined,
  ThunderboltOutlined,
  StarOutlined,
  HeartOutlined,
  SafetyOutlined
} from '@ant-design/icons';

const { Text, Title } = Typography;

interface DataVisualizationProps {
  className?: string;
}

const DataVisualization: React.FC<DataVisualizationProps> = ({ className }) => {
  // 模拟数据
  const performanceMetrics = [
    { label: 'AI响应速度', value: 95, icon: <ThunderboltOutlined />, trend: 12.5 },
    { label: '系统稳定性', value: 98, icon: <StarOutlined />, trend: 8.3 },
    { label: '用户满意度', value: 92, icon: <HeartOutlined />, trend: -2.1 },
    { label: '安全等级', value: 88, icon: <SafetyOutlined />, trend: 15.7 }
  ];

  const getTrendIcon = (trend: number) => {
    return trend > 0 ? <RiseOutlined style={{ color: '#52c41a' }} /> : <FallOutlined style={{ color: '#f5222d' }} />;
  };

  return (
    <div className={className}>
      <Row gutter={[16, 16]}>
        {/* 核心指标卡片 */}
        {performanceMetrics.map((metric, index) => (
          <Col xs={12} lg={6} key={index}>
            <Card>
              <Statistic
                title={metric.label}
                value={metric.value}
                suffix="%"
                prefix={metric.icon}
                valueStyle={{ 
                  color: metric.value >= 90 ? '#52c41a' : metric.value >= 80 ? '#faad14' : '#f5222d' 
                }}
              />
              <div style={{ marginTop: 8 }}>
                <Space>
                  {getTrendIcon(metric.trend)}
                  <Text type="secondary">{Math.abs(metric.trend)}%</Text>
                </Space>
              </div>
            </Card>
          </Col>
        ))}

        {/* 活动趋势图 */}
        <Col xs={24} lg={16}>
          <Card 
            title={
              <Space>
                <LineChartOutlined />
                <span>本周活动趋势</span>
              </Space>
            }
          >
            <div style={{ height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Text type="secondary">图表组件将在后续版本中集成</Text>
            </div>
          </Card>
        </Col>

        {/* 使用分布 */}
        <Col xs={24} lg={8}>
          <Card 
            title={
              <Space>
                <BarChartOutlined />
                <span>使用分布</span>
              </Space>
            }
          >
            <div style={{ height: 200 }}>
              <Row gutter={[0, 16]}>
                <Col span={24}>
                  <Statistic title="AI对话" value={45} suffix="%" valueStyle={{ color: '#1890ff' }} />
                </Col>
                <Col span={24}>
                  <Statistic title="智慧学习" value={28} suffix="%" valueStyle={{ color: '#52c41a' }} />
                </Col>
                <Col span={24}>
                  <Statistic title="项目管理" value={18} suffix="%" valueStyle={{ color: '#722ed1' }} />
                </Col>
                <Col span={24}>
                  <Statistic title="安全监控" value={9} suffix="%" valueStyle={{ color: '#f5222d' }} />
                </Col>
              </Row>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DataVisualization;