import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Select,
  DatePicker,
  Space,
  Statistic,
  Progress,
  Table,
  Tag,
  Avatar,
  Tooltip,
  Button,
  Divider
} from 'antd';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  TrophyOutlined,
  TeamOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Option } = Select;
const { RangePicker } = DatePicker;

interface ProjectStats {
  totalProjects: number;
  activeProjects: number;
  completedProjects: number;
  overdueProjects: number;
  totalTasks: number;
  completedTasks: number;
  teamMembers: number;
  avgProgress: number;
}

interface ProjectPerformance {
  id: string;
  name: string;
  progress: number;
  tasksCompleted: number;
  totalTasks: number;
  teamSize: number;
  daysRemaining: number;
  status: 'on-track' | 'at-risk' | 'delayed';
  priority: 'high' | 'medium' | 'low';
}

interface TimelineData {
  date: string;
  completed: number;
  created: number;
  inProgress: number;
}

interface TeamPerformance {
  name: string;
  tasksCompleted: number;
  efficiency: number;
  projects: number;
}

const ProjectAnalytics: React.FC = () => {
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const [projects, setProjects] = useState<ProjectPerformance[]>([]);
  const [timelineData, setTimelineData] = useState<TimelineData[]>([]);
  const [teamData, setTeamData] = useState<TeamPerformance[]>([]);
  const [loading, setLoading] = useState(false);
  const [timeRange, setTimeRange] = useState<string>('30');
  const [selectedProject, setSelectedProject] = useState<string>('all');

  // 模拟数据
  useEffect(() => {
    setLoading(true);
    setTimeout(() => {
      setStats({
        totalProjects: 12,
        activeProjects: 8,
        completedProjects: 4,
        overdueProjects: 2,
        totalTasks: 156,
        completedTasks: 89,
        teamMembers: 15,
        avgProgress: 67
      });

      setProjects([
        {
          id: '1',
          name: '太上老君文化研究项目',
          progress: 85,
          tasksCompleted: 34,
          totalTasks: 40,
          teamSize: 6,
          daysRemaining: 15,
          status: 'on-track',
          priority: 'high'
        },
        {
          id: '2',
          name: '智慧学习平台开发',
          progress: 62,
          tasksCompleted: 25,
          totalTasks: 45,
          teamSize: 8,
          daysRemaining: 30,
          status: 'at-risk',
          priority: 'high'
        },
        {
          id: '3',
          name: '健康管理系统',
          progress: 45,
          tasksCompleted: 18,
          totalTasks: 35,
          teamSize: 4,
          daysRemaining: -5,
          status: 'delayed',
          priority: 'medium'
        },
        {
          id: '4',
          name: '社区交流平台',
          progress: 78,
          tasksCompleted: 28,
          totalTasks: 36,
          teamSize: 5,
          daysRemaining: 20,
          status: 'on-track',
          priority: 'medium'
        }
      ]);

      setTimelineData([
        { date: '01-25', completed: 12, created: 8, inProgress: 15 },
        { date: '01-26', completed: 15, created: 10, inProgress: 18 },
        { date: '01-27', completed: 18, created: 12, inProgress: 20 },
        { date: '01-28', completed: 22, created: 15, inProgress: 25 },
        { date: '01-29', completed: 25, created: 18, inProgress: 28 },
        { date: '01-30', completed: 28, created: 20, inProgress: 30 },
        { date: '01-31', completed: 32, created: 22, inProgress: 35 },
        { date: '02-01', completed: 35, created: 25, inProgress: 38 }
      ]);

      setTeamData([
        { name: '张三', tasksCompleted: 24, efficiency: 92, projects: 3 },
        { name: '李四', tasksCompleted: 18, efficiency: 88, projects: 2 },
        { name: '王五', tasksCompleted: 22, efficiency: 85, projects: 3 },
        { name: '赵六', tasksCompleted: 16, efficiency: 82, projects: 2 },
        { name: '钱七', tasksCompleted: 20, efficiency: 90, projects: 2 }
      ]);

      setLoading(false);
    }, 1000);
  }, [timeRange, selectedProject]);

  const getStatusColor = (status: ProjectPerformance['status']) => {
    const colors = {
      'on-track': 'success',
      'at-risk': 'warning',
      'delayed': 'error'
    };
    return colors[status];
  };

  const getStatusText = (status: ProjectPerformance['status']) => {
    const texts = {
      'on-track': '正常',
      'at-risk': '风险',
      'delayed': '延期'
    };
    return texts[status];
  };

  const getPriorityColor = (priority: ProjectPerformance['priority']) => {
    const colors = {
      high: 'red',
      medium: 'orange',
      low: 'blue'
    };
    return colors[priority];
  };

  const getPriorityText = (priority: ProjectPerformance['priority']) => {
    const texts = {
      high: '高',
      medium: '中',
      low: '低'
    };
    return texts[priority];
  };

  const projectColumns: ColumnsType<ProjectPerformance> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (name, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          <Space size={4}>
            <Tag color={getStatusColor(record.status)}>
              {getStatusText(record.status)}
            </Tag>
            <Tag color={getPriorityColor(record.priority)}>
              {getPriorityText(record.priority)}
            </Tag>
          </Space>
        </div>
      )
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      render: (progress) => (
        <div style={{ width: 120 }}>
          <Progress 
            percent={progress} 
            size="small" 
            status={progress < 50 ? 'exception' : progress < 80 ? 'active' : 'success'}
          />
        </div>
      )
    },
    {
      title: '任务完成',
      key: 'tasks',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>
            {record.tasksCompleted}/{record.totalTasks}
          </div>
          <div style={{ fontSize: '12px', color: '#666' }}>
            完成率 {Math.round((record.tasksCompleted / record.totalTasks) * 100)}%
          </div>
        </div>
      )
    },
    {
      title: '团队规模',
      dataIndex: 'teamSize',
      key: 'teamSize',
      render: (size) => (
        <Space>
          <TeamOutlined />
          {size}人
        </Space>
      )
    },
    {
      title: '剩余天数',
      dataIndex: 'daysRemaining',
      key: 'daysRemaining',
      render: (days) => (
        <Space>
          <ClockCircleOutlined style={{ color: days < 0 ? '#f5222d' : days < 7 ? '#fa8c16' : '#52c41a' }} />
          <span style={{ color: days < 0 ? '#f5222d' : days < 7 ? '#fa8c16' : '#52c41a' }}>
            {days < 0 ? `逾期${Math.abs(days)}天` : `${days}天`}
          </span>
        </Space>
      )
    }
  ];

  const pieData = [
    { name: '已完成', value: stats?.completedProjects || 0, color: '#52c41a' },
    { name: '进行中', value: stats?.activeProjects || 0, color: '#1890ff' },
    { name: '逾期', value: stats?.overdueProjects || 0, color: '#f5222d' }
  ];

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题和筛选器 */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>项目分析</h1>
          <p style={{ margin: '8px 0 0 0', color: '#666' }}>
            项目数据分析和绩效洞察
          </p>
        </div>
        <Space>
          <Select
            value={selectedProject}
            style={{ width: 200 }}
            onChange={setSelectedProject}
          >
            <Option value="all">全部项目</Option>
            <Option value="1">太上老君文化研究项目</Option>
            <Option value="2">智慧学习平台开发</Option>
            <Option value="3">健康管理系统</Option>
            <Option value="4">社区交流平台</Option>
          </Select>
          <Select
            value={timeRange}
            style={{ width: 120 }}
            onChange={setTimeRange}
          >
            <Option value="7">近7天</Option>
            <Option value="30">近30天</Option>
            <Option value="90">近90天</Option>
          </Select>
          <Button icon={<DownloadOutlined />}>导出报告</Button>
        </Space>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总项目数"
              value={stats?.totalProjects || 0}
              prefix={<BarChartOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃项目"
              value={stats?.activeProjects || 0}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#52c41a' }}
              suffix={`/ ${stats?.totalProjects || 0}`}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="任务完成率"
              value={stats ? Math.round((stats.completedTasks / stats.totalTasks) * 100) : 0}
              prefix={<CheckCircleOutlined />}
              suffix="%"
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均进度"
              value={stats?.avgProgress || 0}
              prefix={<TrophyOutlined />}
              suffix="%"
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={16}>
          <Card title="项目进度趋势" extra={<LineChartOutlined />}>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={timelineData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <RechartsTooltip />
                <Legend />
                <Line 
                  type="monotone" 
                  dataKey="completed" 
                  stroke="#52c41a" 
                  name="已完成"
                  strokeWidth={2}
                />
                <Line 
                  type="monotone" 
                  dataKey="inProgress" 
                  stroke="#1890ff" 
                  name="进行中"
                  strokeWidth={2}
                />
                <Line 
                  type="monotone" 
                  dataKey="created" 
                  stroke="#fa8c16" 
                  name="新创建"
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col span={8}>
          <Card title="项目状态分布" extra={<PieChartOutlined />}>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <RechartsTooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* 项目绩效和团队绩效 */}
      <Row gutter={16}>
        <Col span={16}>
          <Card title="项目绩效排行">
            <Table
              columns={projectColumns}
              dataSource={projects}
              rowKey="id"
              loading={loading}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="团队绩效" extra={<TeamOutlined />}>
            <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
              {teamData.map((member, index) => (
                <div key={member.name} style={{ marginBottom: '16px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                    <Space>
                      <Avatar size="small">{member.name.charAt(0)}</Avatar>
                      <span style={{ fontWeight: 500 }}>{member.name}</span>
                    </Space>
                    <Tag color={member.efficiency >= 90 ? 'green' : member.efficiency >= 80 ? 'orange' : 'red'}>
                      {member.efficiency}%
                    </Tag>
                  </div>
                  <div style={{ fontSize: '12px', color: '#666', marginBottom: '4px' }}>
                    完成任务: {member.tasksCompleted} | 参与项目: {member.projects}
                  </div>
                  <Progress 
                    percent={member.efficiency} 
                    size="small" 
                    showInfo={false}
                    strokeColor={member.efficiency >= 90 ? '#52c41a' : member.efficiency >= 80 ? '#fa8c16' : '#f5222d'}
                  />
                  {index < teamData.length - 1 && <Divider style={{ margin: '16px 0' }} />}
                </div>
              ))}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ProjectAnalytics;