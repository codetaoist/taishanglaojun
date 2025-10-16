import React, { useState } from 'react';
import { 
  Card, 
  Input, 
  List, 
  Typography, 
  Collapse, 
  Space, 
  Tag, 
  Button, 
  Row, 
  Col,
  Divider,
  Avatar,
  Rate
} from 'antd';
import { getNotificationInstance } from '../services/notificationService';
import {
  SearchOutlined,
  QuestionCircleOutlined,
  BookOutlined,
  CustomerServiceOutlined,
  PhoneOutlined,
  MailOutlined,
  WechatOutlined,
  RightOutlined
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;
const { Panel } = Collapse;

interface FAQItem {
  id: string;
  question: string;
  answer: string;
  category: string;
  tags: string[];
  helpful: number;
}

interface GuideItem {
  id: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  link: string;
}

const Help: React.FC = () => {
  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');

  // 常见问题数据
  const faqData: FAQItem[] = [
    {
      id: '1',
      question: '如何开始使用太上老君AI平台？',
      answer: '首先注册账号并完成邮箱验证，然后登录平台。您可以从仪表板开始，探索各种AI功能，如智能对话、文化智慧等。建议先查看新手指南了解平台的主要功能。',
      category: '入门指南',
      tags: ['注册', '登录', '新手'],
      helpful: 156
    },
    {
      id: '2',
      question: '如何使用AI智能对话功能？',
      answer: '点击左侧菜单的"AI智能服务" > "智能对话"，进入对话界面。您可以输入问题或话题，AI会根据您的输入提供智能回复。支持文本、图片等多种输入方式。',
      category: 'AI功能',
      tags: ['对话', 'AI', '聊天'],
      helpful: 203
    },
    {
      id: '3',
      question: '如何管理我的学习计划？',
      answer: '在"智能学习"模块中，您可以创建个性化学习计划，设置学习目标，跟踪学习进度。系统会根据您的学习情况推荐合适的课程和资源。',
      category: '学习功能',
      tags: ['学习', '计划', '进度'],
      helpful: 89
    },
    {
      id: '4',
      question: '如何联系客服支持？',
      answer: '您可以通过多种方式联系我们：1) 点击右下角的客服图标进行在线咨询；2) 发送邮件至 support@taishanglaojun.com；3) 拨打客服热线 400-123-4567。',
      category: '客服支持',
      tags: ['客服', '联系', '支持'],
      helpful: 67
    },
    {
      id: '5',
      question: '如何保护我的账号安全？',
      answer: '建议您：1) 设置强密码并定期更换；2) 开启两步验证；3) 不要在公共设备上保存登录信息；4) 定期检查登录记录。如发现异常请及时联系客服。',
      category: '安全设置',
      tags: ['安全', '密码', '验证'],
      helpful: 134
    }
  ];

  // 使用指南数据
  const guideData: GuideItem[] = [
    {
      id: '1',
      title: '快速入门指南',
      description: '了解平台基本功能和操作流程',
      icon: <BookOutlined />,
      link: '/guide/getting-started'
    },
    {
      id: '2',
      title: 'AI功能详解',
      description: '深入了解各种AI智能服务',
      icon: <QuestionCircleOutlined />,
      link: '/guide/ai-features'
    },
    {
      id: '3',
      title: '学习系统使用',
      description: '充分利用智能学习功能',
      icon: <BookOutlined />,
      link: '/guide/learning-system'
    },
    {
      id: '4',
      title: '项目管理工具',
      description: '高效管理您的项目和任务',
      icon: <BookOutlined />,
      link: '/guide/project-management'
    }
  ];

  // 分类列表
  const categories = ['all', '入门指南', 'AI功能', '学习功能', '客服支持', '安全设置'];

  // 过滤FAQ数据
  const filteredFAQ = faqData.filter(item => {
    const matchesSearch = item.question.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.answer.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.tags.some(tag => tag.toLowerCase().includes(searchText.toLowerCase()));
    const matchesCategory = selectedCategory === 'all' || item.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  // 处理有用评价
  const handleHelpful = (id: string) => {
    const notification = getNotificationInstance();
    notification.success({
      message: '感谢您的反馈！'
    });
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      {/* 页面标题 */}
      <div className="mb-8">
        <Title level={2} className="mb-2">
          <QuestionCircleOutlined className="mr-3" />
          帮助中心
        </Title>
        <Paragraph className="text-gray-600">
          欢迎来到太上老君AI平台帮助中心，这里有您需要的所有帮助信息
        </Paragraph>
      </div>

      {/* 搜索栏 */}
      <Card className="mb-6">
        <Input
          size="large"
          placeholder="搜索帮助内容..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className="mb-4"
        />
        
        {/* 分类标签 */}
        <Space wrap>
          {categories.map(category => (
            <Tag
              key={category}
              color={selectedCategory === category ? 'blue' : 'default'}
              style={{ cursor: 'pointer' }}
              onClick={() => setSelectedCategory(category)}
            >
              {category === 'all' ? '全部' : category}
            </Tag>
          ))}
        </Space>
      </Card>

      <Row gutter={[24, 24]}>
        {/* 左侧：常见问题 */}
        <Col xs={24} lg={16}>
          <Card title="常见问题" className="mb-6">
            <Collapse ghost>
              {filteredFAQ.map(item => (
                <Panel
                  key={item.id}
                  header={
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{item.question}</span>
                      <Space>
                        {item.tags.map(tag => (
                          <Tag key={tag} size="small">{tag}</Tag>
                        ))}
                      </Space>
                    </div>
                  }
                >
                  <div className="pl-4">
                    <Paragraph className="mb-4">{item.answer}</Paragraph>
                    <div className="flex items-center justify-between">
                      <Space>
                        <Text type="secondary">这个回答对您有帮助吗？</Text>
                        <Button 
                          type="link" 
                          size="small"
                          onClick={() => handleHelpful(item.id)}
                        >
                          有用 ({item.helpful})
                        </Button>
                      </Space>
                      <Text type="secondary">分类：{item.category}</Text>
                    </div>
                  </div>
                </Panel>
              ))}
            </Collapse>
            
            {filteredFAQ.length === 0 && (
              <div className="text-center py-8">
                <Text type="secondary">没有找到相关问题，请尝试其他关键词</Text>
              </div>
            )}
          </Card>
        </Col>

        {/* 右侧：使用指南和联系方式 */}
        <Col xs={24} lg={8}>
          {/* 使用指南 */}
          <Card title="使用指南" className="mb-6">
            <List
              dataSource={guideData}
              renderItem={item => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={item.icon} />}
                    title={
                      <Button type="link" className="p-0 h-auto">
                        {item.title}
                        <RightOutlined className="ml-2" />
                      </Button>
                    }
                    description={item.description}
                  />
                </List.Item>
              )}
            />
          </Card>

          {/* 联系我们 */}
          <Card title="联系我们">
            <Space direction="vertical" className="w-full">
              <div className="flex items-center">
                <CustomerServiceOutlined className="mr-3 text-blue-500" />
                <div>
                  <div className="font-medium">在线客服</div>
                  <Text type="secondary">7×24小时在线支持</Text>
                </div>
              </div>
              
              <Divider className="my-3" />
              
              <div className="flex items-center">
                <PhoneOutlined className="mr-3 text-green-500" />
                <div>
                  <div className="font-medium">客服热线</div>
                  <Text type="secondary">400-123-4567</Text>
                </div>
              </div>
              
              <Divider className="my-3" />
              
              <div className="flex items-center">
                <MailOutlined className="mr-3 text-orange-500" />
                <div>
                  <div className="font-medium">邮箱支持</div>
                  <Text type="secondary">support@taishanglaojun.com</Text>
                </div>
              </div>
              
              <Divider className="my-3" />
              
              <div className="flex items-center">
                <WechatOutlined className="mr-3 text-green-600" />
                <div>
                  <div className="font-medium">微信客服</div>
                  <Text type="secondary">扫码添加客服微信</Text>
                </div>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Help;