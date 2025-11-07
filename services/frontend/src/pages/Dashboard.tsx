import React from 'react';
import { Card, Row, Col, Statistic, Progress } from 'antd';
import { 
  DatabaseOutlined, 
  AppstoreOutlined, 
  FileTextOutlined,
  ExperimentOutlined
} from '@ant-design/icons';

const Dashboard: React.FC = () => {
  return (
    <div>
      <h1>仪表盘</h1>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic
              title="已安装插件"
              value={12}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行中插件"
              value={8}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已注册模型"
              value={5}
              prefix={<ExperimentOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="向量集合"
              value={3}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>
      </Row>
      
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="系统资源使用情况">
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>CPU使用率</span>
                <span>45%</span>
              </div>
              <Progress percent={45} status="active" />
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>内存使用率</span>
                <span>68%</span>
              </div>
              <Progress percent={68} status="active" />
            </div>
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>磁盘使用率</span>
                <span>32%</span>
              </div>
              <Progress percent={32} />
            </div>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="最近任务">
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>索引重建</span>
                <span>进行中</span>
              </div>
              <Progress percent={75} status="active" />
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>模型训练</span>
                <span>已完成</span>
              </div>
              <Progress percent={100} />
            </div>
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>数据预处理</span>
                <span>等待中</span>
              </div>
              <Progress percent={0} />
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;