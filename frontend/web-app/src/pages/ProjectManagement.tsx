import React from 'react';
import { Card, Row, Col, Button, Statistic, Progress, Space, Avatar, Tag } from 'antd';
import { 
  ProjectOutlined, 
  TeamOutlined, 
  BarChartOutlined, 
  CheckCircleOutlined,
  ClockCircleOutlined,
  ArrowRightOutlined,
  TrophyOutlined,
  FileTextOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const ProjectManagement: React.FC = () => {
  const navigate = useNavigate();

  const quickStats = {
    totalProjects: 12,
    activeProjects: 8,
    completedTasks: 89,
    totalTasks: 156,
    teamMembers: 15,
    avgProgress: 67
  };

  const recentProjects = [
    {
      name: '太上老君文化研究项目',
      progress: 85,
      status: 'on-track',
      team: ['张三', '李四', '王五']
    },
    {
      name: '智慧学习平台开发',
      progress: 62,
      status: 'at-risk',
      team: ['赵六', '钱七', '孙八']
    },
    {
      name: '健康管理系统',
      progress: 45,
      status: 'delayed',
      team: ['周九', '吴十']
    }
  ];

  const getStatusColor = (status: string) => {
    const colors = {
      'on-track': 'success',
      'at-risk': 'warning',
      'delayed': 'error'
    };
    return colors[status as keyof typeof colors];
  };

  const getStatusText = (status: string) => {
    const texts = {
      'on-track': '正常',
      'at-risk': '风险',
      'delayed': '延期'
    };
    return texts[status as keyof typeof texts];
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ margin: 0, fontSize: '28px', fontWeight: 600 }}>项目管理中心</h1>
        <p style={{ margin: '8px 0 0 0', color: '#666', fontSize: '16px' }}>
          智能项目管理系统，提供全方位的项目规划、执行和监控功能
        </p>
      </div>

      {/* 统计概览 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总项目数"
              value={quickStats.totalProjects}
              prefix={<ProjectOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃项目"
              value={quickStats.activeProjects}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="任务完成率"
              value={Math.round((quickStats.completedTasks / quickStats.totalTasks) * 100)}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="团队成员"
              value={quickStats.teamMembers}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 功能模块导航 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card 
            hoverable
            style={{ textAlign: 'center', cursor: 'pointer' }}
            onClick={() => navigate('/projects/workspace')}
          >
            <ProjectOutlined style={{ fontSize: '48px', color: '#1890ff', marginBottom: '16px' }} />
            <h3>项目工作台</h3>
            <p style={{ color: '#666' }}>项目创建、配置和管理</p>
            <Button type="primary" icon={<ArrowRightOutlined />}>
              进入工作台
            </Button>
          </Card>
        </Col>
        <Col span={6}>
          <Card 
            hoverable
            style={{ textAlign: 'center', cursor: 'pointer' }}
            onClick={() => navigate('/projects/tasks')}
          >
            <CheckCircleOutlined style={{ fontSize: '48px', color: '#52c41a', marginBottom: '16px' }} />
            <h3>任务管理</h3>
            <p style={{ color: '#666' }}>任务分配、跟踪和执行</p>
            <Button type="primary" icon={<ArrowRightOutlined />}>
              管理任务
            </Button>
          </Card>
        </Col>
        <Col span={6}>
          <Card 
            hoverable
            style={{ textAlign: 'center', cursor: 'pointer' }}
            onClick={() => navigate('/projects/collaboration')}
          >
            <TeamOutlined style={{ fontSize: '48px', color: '#722ed1', marginBottom: '16px' }} />
            <h3>团队协作</h3>
            <p style={{ color: '#666' }}>团队管理和协作工具</p>
            <Button type="primary" icon={<ArrowRightOutlined />}>
              团队协作
            </Button>
          </Card>
        </Col>
        <Col span={6}>
          <Card 
            hoverable
            style={{ textAlign: 'center', cursor: 'pointer' }}
            onClick={() => navigate('/projects/analytics')}
          >
            <BarChartOutlined style={{ fontSize: '48px', color: '#fa8c16', marginBottom: '16px' }} />
            <h3>项目分析</h3>
            <p style={{ color: '#666' }}>数据分析和绩效洞察</p>
            <Button type="primary" icon={<ArrowRightOutlined />}>
              查看分析
            </Button>
          </Card>
        </Col>
      </Row>

      {/* 最近项目 */}
      <Row gutter={16}>
        <Col span={16}>
          <Card title="最近项目" extra={<Button type="link" onClick={() => navigate('/projects/workspace')}>查看全部</Button>}>
            <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
              {recentProjects.map((project, index) => (
                <div key={index} style={{ marginBottom: '16px', padding: '16px', border: '1px solid #f0f0f0', borderRadius: '8px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                    <h4 style={{ margin: 0 }}>{project.name}</h4>
                    <Tag color={getStatusColor(project.status)}>
                      {getStatusText(project.status)}
                    </Tag>
                  </div>
                  <div style={{ marginBottom: '12px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                      <span>进度</span>
                      <span>{project.progress}%</span>
                    </div>
                    <Progress 
                      percent={project.progress} 
                      size="small" 
                      status={project.progress < 50 ? 'exception' : project.progress < 80 ? 'active' : 'success'}
                    />
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space>
                      <span style={{ color: '#666' }}>团队:</span>
                      <Avatar.Group size="small" maxCount={3}>
                        {project.team.map((member, idx) => (
                          <Avatar key={idx} size="small">{member.charAt(0)}</Avatar>
                        ))}
                      </Avatar.Group>
                    </Space>
                    <Button type="link" size="small" onClick={() => navigate('/projects/workspace')}>
                      查看详情
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card title="快速操作">
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Button 
                type="primary" 
                block 
                icon={<ProjectOutlined />}
                onClick={() => navigate('/projects/workspace')}
              >
                创建新项目
              </Button>
              <Button 
                block 
                icon={<CheckCircleOutlined />}
                onClick={() => navigate('/projects/tasks')}
              >
                添加任务
              </Button>
              <Button 
                block 
                icon={<TeamOutlined />}
                onClick={() => navigate('/projects/collaboration')}
              >
                邀请成员
              </Button>
              <Button 
                block 
                icon={<FileTextOutlined />}
                onClick={() => navigate('/projects/collaboration')}
              >
                上传文档
              </Button>
              <Button 
                block 
                icon={<BarChartOutlined />}
                onClick={() => navigate('/projects/analytics')}
              >
                查看报告
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ProjectManagement;