import React from 'react';
import { Card, Col, Row, Statistic, Typography } from 'antd';

const { Title, Paragraph } = Typography;

function SystemMonitoring() {
  return (
    <div style={{ padding: 24 }}>
      <Title level={3}>系统监控</Title>
      <Paragraph>实时查看系统运行状况与关键指标（占位页面）。</Paragraph>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12} lg={8}>
          <Card>
            <Statistic title="CPU 使用率" value={23.5} suffix="%" />
          </Card>
        </Col>
        <Col xs={24} md={12} lg={8}>
          <Card>
            <Statistic title="内存占用" value={58.2} suffix="%" />
          </Card>
        </Col>
        <Col xs={24} md={12} lg={8}>
          <Card>
            <Statistic title="请求错误率" value={0.7} suffix="%" />
          </Card>
        </Col>
      </Row>
    </div>
  );
}

export default SystemMonitoring;