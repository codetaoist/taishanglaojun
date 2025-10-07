import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, Progress, Row, Col, Statistic, Descriptions, Tabs, List, Avatar } from 'antd';
import { BookOutlined, PlayCircleOutlined, TrophyOutlined, UserOutlined, EyeOutlined, EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;
const { TabPane } = Tabs;

interface SecurityCourse {
  id: string;
  title: string;
  category: string;
  level: 'beginner' | 'intermediate' | 'advanced';
  duration: number; // 分钟
  description: string;
  instructor: string;
  enrolledCount: number;
  completionRate: number;
  status: 'active' | 'draft' | 'archived';
  createdAt: string;
}

interface LabEnvironment {
  id: string;
  name: string;
  type: 'web' | 'network' | 'mobile' | 'cloud';
  difficulty: 'easy' | 'medium' | 'hard';
  description: string;
  objectives: string[];
  estimatedTime: number;
  completedCount: number;
  status: 'available' | 'maintenance' | 'unavailable';
}

interface SecurityCertification {
  id: string;
  name: string;
  description: string;
  requirements: string[];
  validityPeriod: number; // 月
  passScore: number;
  totalQuestions: number;
  timeLimit: number; // 分钟
  issuedCount: number;
  status: 'active' | 'inactive';
}

const SecurityEducation: React.FC = () => {
  const [courses, setCourses] = useState<SecurityCourse[]>([]);
  const [labs, setLabs] = useState<LabEnvironment[]>([]);
  const [certifications, setCertifications] = useState<SecurityCertification[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [courseModalVisible, setCourseModalVisible] = useState(false);
  const [selectedCourse, setSelectedCourse] = useState<SecurityCourse | null>(null);
  const [selectedLab, setSelectedLab] = useState<LabEnvironment | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadCourses();
    loadLabs();
    loadCertifications();
  }, []);

  const loadCourses = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        setCourses([
          {
            id: '1',
            title: 'Web应用安全基础',
            category: 'Web安全',
            level: 'beginner',
            duration: 120,
            description: '学习Web应用常见安全漏洞及防护措施',
            instructor: '张安全',
            enrolledCount: 156,
            completionRate: 85,
            status: 'active',
            createdAt: '2024-01-01'
          },
          {
            id: '2',
            title: '网络安全渗透测试',
            category: '渗透测试',
            level: 'advanced',
            duration: 240,
            description: '深入学习渗透测试方法和工具使用',
            instructor: '李专家',
            enrolledCount: 89,
            completionRate: 72,
            status: 'active',
            createdAt: '2024-01-05'
          },
          {
            id: '3',
            title: '安全编码实践',
            category: '安全开发',
            level: 'intermediate',
            duration: 180,
            description: '学习安全编码规范和最佳实践',
            instructor: '王开发',
            enrolledCount: 203,
            completionRate: 91,
            status: 'active',
            createdAt: '2024-01-10'
          }
        ]);
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load courses:', error);
      setLoading(false);
    }
  };

  const loadLabs = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setLabs([
          {
            id: '1',
            name: 'SQL注入攻击实验',
            type: 'web',
            difficulty: 'medium',
            description: '通过实际操作学习SQL注入攻击和防护',
            objectives: ['理解SQL注入原理', '掌握攻击技巧', '学习防护方法'],
            estimatedTime: 60,
            completedCount: 234,
            status: 'available'
          },
          {
            id: '2',
            name: '网络扫描与侦察',
            type: 'network',
            difficulty: 'easy',
            description: '学习网络扫描工具的使用和信息收集技术',
            objectives: ['掌握Nmap使用', '学习端口扫描', '了解服务识别'],
            estimatedTime: 45,
            completedCount: 189,
            status: 'available'
          },
          {
            id: '3',
            name: '移动应用逆向分析',
            type: 'mobile',
            difficulty: 'hard',
            description: '学习Android应用的逆向分析技术',
            objectives: ['掌握APK分析', '学习代码混淆', '了解动态调试'],
            estimatedTime: 90,
            completedCount: 67,
            status: 'maintenance'
          }
        ]);
      }, 500);
    } catch (error) {
      console.error('Failed to load labs:', error);
    }
  };

  const loadCertifications = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setCertifications([
          {
            id: '1',
            name: '网络安全基础认证',
            description: '验证网络安全基础知识和技能',
            requirements: ['完成基础课程', '通过实验考核', '参加在线考试'],
            validityPeriod: 24,
            passScore: 80,
            totalQuestions: 100,
            timeLimit: 120,
            issuedCount: 145,
            status: 'active'
          },
          {
            id: '2',
            name: 'Web安全专家认证',
            description: '验证Web应用安全专业技能',
            requirements: ['完成高级课程', '完成项目实战', '通过专家评审'],
            validityPeriod: 36,
            passScore: 85,
            totalQuestions: 150,
            timeLimit: 180,
            issuedCount: 67,
            status: 'active'
          }
        ]);
      }, 300);
    } catch (error) {
      console.error('Failed to load certifications:', error);
    }
  };

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'beginner': return 'green';
      case 'intermediate': return 'orange';
      case 'advanced': return 'red';
      default: return 'default';
    }
  };

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case 'easy': return 'green';
      case 'medium': return 'orange';
      case 'hard': return 'red';
      default: return 'default';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': case 'available': return 'green';
      case 'draft': case 'maintenance': return 'orange';
      case 'archived': case 'unavailable': return 'red';
      case 'inactive': return 'gray';
      default: return 'default';
    }
  };

  const courseColumns: ColumnsType<SecurityCourse> = [
    {
      title: '课程名称',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
    },
    {
      title: '难度等级',
      dataIndex: 'level',
      key: 'level',
      render: (level) => (
        <Tag color={getLevelColor(level)}>
          {level === 'beginner' ? '初级' :
           level === 'intermediate' ? '中级' : '高级'}
        </Tag>
      ),
    },
    {
      title: '时长',
      dataIndex: 'duration',
      key: 'duration',
      render: (duration) => `${duration} 分钟`,
    },
    {
      title: '讲师',
      dataIndex: 'instructor',
      key: 'instructor',
    },
    {
      title: '报名人数',
      dataIndex: 'enrolledCount',
      key: 'enrolledCount',
    },
    {
      title: '完成率',
      dataIndex: 'completionRate',
      key: 'completionRate',
      render: (rate) => (
        <Progress percent={rate} size="small" />
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'active' ? '活跃' :
           status === 'draft' ? '草稿' : '已归档'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedCourse(record);
              setCourseModalVisible(true);
            }}
          >
            查看
          </Button>
          <Button type="link" icon={<EditOutlined />}>编辑</Button>
        </Space>
      ),
    },
  ];

  const labColumns: ColumnsType<LabEnvironment> = [
    {
      title: '实验名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag>
          {type === 'web' ? 'Web安全' :
           type === 'network' ? '网络安全' :
           type === 'mobile' ? '移动安全' : '云安全'}
        </Tag>
      ),
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      render: (difficulty) => (
        <Tag color={getDifficultyColor(difficulty)}>
          {difficulty === 'easy' ? '简单' :
           difficulty === 'medium' ? '中等' : '困难'}
        </Tag>
      ),
    },
    {
      title: '预计时长',
      dataIndex: 'estimatedTime',
      key: 'estimatedTime',
      render: (time) => `${time} 分钟`,
    },
    {
      title: '完成人数',
      dataIndex: 'completedCount',
      key: 'completedCount',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'available' ? '可用' :
           status === 'maintenance' ? '维护中' : '不可用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<PlayCircleOutlined />}
            disabled={record.status !== 'available'}
          >
            开始实验
          </Button>
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedLab(record);
              setModalVisible(true);
            }}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  const handleCreateCourse = (values: any) => {
    console.log('Creating course with values:', values);
    setCourseModalVisible(false);
    form.resetFields();
    // 这里会调用API创建课程
  };

  return (
    <div>
      {/* 安全教育统计 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="在线课程"
              value={12}
              prefix={<BookOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="实验环境"
              value={8}
              prefix={<PlayCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="学习人数"
              value={1256}
              prefix={<UserOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="颁发证书"
              value={234}
              prefix={<TrophyOutlined style={{ color: '#722ed1' }} />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Tabs defaultActiveKey="courses">
        <TabPane tab="在线课程" key="courses">
          <Card 
            title="安全课程管理" 
            extra={
              <Space>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={() => setCourseModalVisible(true)}
                >
                  新建课程
                </Button>
                <Button onClick={loadCourses}>刷新</Button>
              </Space>
            }
          >
            <Table
              columns={courseColumns}
              dataSource={courses}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
            />
          </Card>
        </TabPane>

        <TabPane tab="实验环境" key="labs">
          <Card 
            title="实验环境管理" 
            extra={
              <Space>
                <Button type="primary" icon={<PlusOutlined />}>
                  新建实验
                </Button>
                <Button onClick={loadLabs}>刷新</Button>
              </Space>
            }
          >
            <Table
              columns={labColumns}
              dataSource={labs}
              rowKey="id"
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
            />
          </Card>
        </TabPane>

        <TabPane tab="认证管理" key="certifications">
          <Card 
            title="安全认证管理" 
            extra={
              <Space>
                <Button type="primary" icon={<PlusOutlined />}>
                  新建认证
                </Button>
                <Button>刷新</Button>
              </Space>
            }
          >
            <List
              grid={{ gutter: 16, column: 2 }}
              dataSource={certifications}
              renderItem={(cert) => (
                <List.Item>
                  <Card
                    title={cert.name}
                    extra={
                      <Tag color={getStatusColor(cert.status)}>
                        {cert.status === 'active' ? '活跃' : '停用'}
                      </Tag>
                    }
                    actions={[
                      <Button type="link" icon={<EyeOutlined />}>查看</Button>,
                      <Button type="link" icon={<EditOutlined />}>编辑</Button>,
                    ]}
                  >
                    <p>{cert.description}</p>
                    <Row gutter={[16, 8]}>
                      <Col span={12}>
                        <strong>通过分数:</strong> {cert.passScore}%
                      </Col>
                      <Col span={12}>
                        <strong>题目数量:</strong> {cert.totalQuestions}
                      </Col>
                      <Col span={12}>
                        <strong>考试时长:</strong> {cert.timeLimit}分钟
                      </Col>
                      <Col span={12}>
                        <strong>已颁发:</strong> {cert.issuedCount}份
                      </Col>
                    </Row>
                  </Card>
                </List.Item>
              )}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* 课程详情模态框 */}
      <Modal
        title="课程详情"
        open={courseModalVisible && selectedCourse !== null}
        onCancel={() => setCourseModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setCourseModalVisible(false)}>
            关闭
          </Button>,
          <Button key="edit" type="primary" icon={<EditOutlined />}>
            编辑课程
          </Button>,
        ]}
        width={800}
      >
        {selectedCourse && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="课程名称" span={2}>
                {selectedCourse.title}
              </Descriptions.Item>
              <Descriptions.Item label="类别">
                {selectedCourse.category}
              </Descriptions.Item>
              <Descriptions.Item label="难度等级">
                <Tag color={getLevelColor(selectedCourse.level)}>
                  {selectedCourse.level === 'beginner' ? '初级' :
                   selectedCourse.level === 'intermediate' ? '中级' : '高级'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="课程时长">
                {selectedCourse.duration} 分钟
              </Descriptions.Item>
              <Descriptions.Item label="讲师">
                {selectedCourse.instructor}
              </Descriptions.Item>
              <Descriptions.Item label="报名人数">
                {selectedCourse.enrolledCount}
              </Descriptions.Item>
              <Descriptions.Item label="完成率">
                <Progress percent={selectedCourse.completionRate} />
              </Descriptions.Item>
              <Descriptions.Item label="课程描述" span={2}>
                {selectedCourse.description}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* 实验详情模态框 */}
      <Modal
        title="实验详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setModalVisible(false)}>
            关闭
          </Button>,
          <Button 
            key="start" 
            type="primary" 
            icon={<PlayCircleOutlined />}
            disabled={selectedLab?.status !== 'available'}
          >
            开始实验
          </Button>,
        ]}
        width={700}
      >
        {selectedLab && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="实验名称" span={2}>
                {selectedLab.name}
              </Descriptions.Item>
              <Descriptions.Item label="类型">
                <Tag>
                  {selectedLab.type === 'web' ? 'Web安全' :
                   selectedLab.type === 'network' ? '网络安全' :
                   selectedLab.type === 'mobile' ? '移动安全' : '云安全'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="难度">
                <Tag color={getDifficultyColor(selectedLab.difficulty)}>
                  {selectedLab.difficulty === 'easy' ? '简单' :
                   selectedLab.difficulty === 'medium' ? '中等' : '困难'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="预计时长">
                {selectedLab.estimatedTime} 分钟
              </Descriptions.Item>
              <Descriptions.Item label="完成人数">
                {selectedLab.completedCount}
              </Descriptions.Item>
              <Descriptions.Item label="实验描述" span={2}>
                {selectedLab.description}
              </Descriptions.Item>
              <Descriptions.Item label="学习目标" span={2}>
                <ul>
                  {selectedLab.objectives.map((objective, index) => (
                    <li key={index}>{objective}</li>
                  ))}
                </ul>
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* 新建课程模态框 */}
      <Modal
        title="新建安全课程"
        open={courseModalVisible && selectedCourse === null}
        onCancel={() => setCourseModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateCourse}
        >
          <Form.Item
            name="title"
            label="课程名称"
            rules={[{ required: true, message: '请输入课程名称' }]}
          >
            <Input placeholder="请输入课程名称" />
          </Form.Item>
          
          <Form.Item
            name="category"
            label="课程类别"
            rules={[{ required: true, message: '请选择课程类别' }]}
          >
            <Select placeholder="请选择课程类别">
              <Option value="Web安全">Web安全</Option>
              <Option value="网络安全">网络安全</Option>
              <Option value="移动安全">移动安全</Option>
              <Option value="云安全">云安全</Option>
              <Option value="安全开发">安全开发</Option>
              <Option value="渗透测试">渗透测试</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="level"
            label="难度等级"
            rules={[{ required: true, message: '请选择难度等级' }]}
          >
            <Select placeholder="请选择难度等级">
              <Option value="beginner">初级</Option>
              <Option value="intermediate">中级</Option>
              <Option value="advanced">高级</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="duration"
            label="课程时长（分钟）"
            rules={[{ required: true, message: '请输入课程时长' }]}
          >
            <Input type="number" placeholder="请输入课程时长" />
          </Form.Item>

          <Form.Item
            name="instructor"
            label="讲师"
            rules={[{ required: true, message: '请输入讲师姓名' }]}
          >
            <Input placeholder="请输入讲师姓名" />
          </Form.Item>

          <Form.Item
            name="description"
            label="课程描述"
            rules={[{ required: true, message: '请输入课程描述' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入课程描述" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SecurityEducation;