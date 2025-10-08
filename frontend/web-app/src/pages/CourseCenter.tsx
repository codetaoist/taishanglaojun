import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Input,
  Select,
  Button,
  Tag,
  Rate,
  Progress,
  Avatar,
  Tabs,
  Badge,
  Space,
  Typography,
  Divider,
  Empty,
  Spin,
  Modal,
  List,
  Statistic,
  Timeline,
  Tooltip,
} from 'antd';
import {
  SearchOutlined,
  FilterOutlined,
  BookOutlined,
  PlayCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
  StarOutlined,
  HeartOutlined,
  ShareAltOutlined,
  DownloadOutlined,
  EyeOutlined,
  TrophyOutlined,
  FireOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;
const { Search } = Input;

// 模拟课程数据
const generateMockCourses = () => {
  const categories = ['前端开发', '后端开发', '数据科学', '人工智能', '移动开发', '云计算', '网络安全', '产品设计'];
  const levels = ['初级', '中级', '高级', '专家'];
  const instructors = ['张教授', '李老师', '王博士', '陈专家', '刘导师'];
  const tags = ['热门', '新课', '精品', '实战', '理论', '项目'];

  return Array.from({ length: 24 }, (_, index) => ({
    id: index + 1,
    title: `${categories[index % categories.length]}实战课程 ${index + 1}`,
    description: '这是一门深入浅出的课程，通过实际项目案例帮助学员掌握核心技能，提升实战能力。',
    category: categories[index % categories.length],
    level: levels[index % levels.length],
    instructor: instructors[index % instructors.length],
    rating: 4.0 + Math.random() * 1.0,
    students: Math.floor(Math.random() * 5000) + 100,
    duration: Math.floor(Math.random() * 40) + 10,
    lessons: Math.floor(Math.random() * 30) + 5,
    price: Math.floor(Math.random() * 500) + 99,
    originalPrice: Math.floor(Math.random() * 200) + 199,
    tags: tags.slice(0, Math.floor(Math.random() * 3) + 1),
    progress: Math.floor(Math.random() * 100),
    isEnrolled: Math.random() > 0.7,
    isFavorite: Math.random() > 0.8,
    thumbnail: `https://picsum.photos/300/200?random=${index}`,
    lastUpdated: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toLocaleDateString(),
    skills: ['技能1', '技能2', '技能3'].slice(0, Math.floor(Math.random() * 3) + 1),
  }));
};

// 模拟课程详情数据
const generateCourseDetail = (courseId: number) => ({
  id: courseId,
  title: `课程详情 ${courseId}`,
  description: '这是一门全面的课程，涵盖了从基础到高级的所有内容。通过理论学习和实践项目，学员将掌握核心技能。',
  fullDescription: `
    本课程是一门综合性的学习课程，旨在帮助学员从零基础开始，逐步掌握相关技能。

    课程特色：
    • 理论与实践相结合
    • 项目驱动的学习方式
    • 一对一指导和答疑
    • 完整的学习路径规划

    学习收获：
    • 掌握核心理论知识
    • 具备实际项目开发能力
    • 获得行业认可的技能证书
    • 建立完整的知识体系
  `,
  outline: [
    { title: '第一章：基础入门', lessons: 5, duration: '2小时' },
    { title: '第二章：核心概念', lessons: 8, duration: '4小时' },
    { title: '第三章：实战项目', lessons: 10, duration: '6小时' },
    { title: '第四章：高级应用', lessons: 6, duration: '3小时' },
    { title: '第五章：项目实战', lessons: 4, duration: '2小时' },
  ],
  reviews: [
    { user: '学员A', rating: 5, comment: '课程内容很棒，老师讲解清晰！', date: '2024-01-15' },
    { user: '学员B', rating: 4, comment: '实战项目很有价值，学到了很多。', date: '2024-01-10' },
    { user: '学员C', rating: 5, comment: '推荐给所有想学习的同学！', date: '2024-01-05' },
  ],
  prerequisites: ['基础编程知识', '计算机基础', '学习热情'],
  targetAudience: ['初学者', '在职人员', '学生', '转行人员'],
  certificate: true,
  downloadable: true,
  lifetime: true,
});

const CourseCenter: React.FC = () => {
  const navigate = useNavigate();
  const [courses, setCourses] = useState<any[]>([]);
  const [filteredCourses, setFilteredCourses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedLevel, setSelectedLevel] = useState('all');
  const [sortBy, setSortBy] = useState('popular');
  const [viewMode, setViewMode] = useState('grid');
  const [courseDetailVisible, setCourseDetailVisible] = useState(false);
  const [selectedCourse, setSelectedCourse] = useState<any>(null);

  const categories = ['all', '前端开发', '后端开发', '数据科学', '人工智能', '移动开发', '云计算', '网络安全', '产品设计'];
  const levels = ['all', '初级', '中级', '高级', '专家'];

  useEffect(() => {
    // 模拟加载数据
    setTimeout(() => {
      const mockCourses = generateMockCourses();
      setCourses(mockCourses);
      setFilteredCourses(mockCourses);
      setLoading(false);
    }, 1000);
  }, []);

  useEffect(() => {
    let filtered = courses;

    // 搜索过滤
    if (searchText) {
      filtered = filtered.filter(course =>
        course.title.toLowerCase().includes(searchText.toLowerCase()) ||
        course.description.toLowerCase().includes(searchText.toLowerCase()) ||
        course.instructor.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    // 分类过滤
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(course => course.category === selectedCategory);
    }

    // 级别过滤
    if (selectedLevel !== 'all') {
      filtered = filtered.filter(course => course.level === selectedLevel);
    }

    // 排序
    switch (sortBy) {
      case 'popular':
        filtered.sort((a, b) => b.students - a.students);
        break;
      case 'rating':
        filtered.sort((a, b) => b.rating - a.rating);
        break;
      case 'newest':
        filtered.sort((a, b) => new Date(b.lastUpdated).getTime() - new Date(a.lastUpdated).getTime());
        break;
      case 'price':
        filtered.sort((a, b) => a.price - b.price);
        break;
      default:
        break;
    }

    setFilteredCourses(filtered);
  }, [courses, searchText, selectedCategory, selectedLevel, sortBy]);

  const handleCourseClick = (course: any) => {
    setSelectedCourse({ ...course, ...generateCourseDetail(course.id) });
    setCourseDetailVisible(true);
  };

  const handleEnroll = (courseId: number) => {
    setCourses(prev => prev.map(course =>
      course.id === courseId ? { ...course, isEnrolled: true, progress: 0 } : course
    ));
    Modal.success({
      title: '报名成功！',
      content: '您已成功报名该课程，可以开始学习了。',
    });
  };

  const handleFavorite = (courseId: number) => {
    setCourses(prev => prev.map(course =>
      course.id === courseId ? { ...course, isFavorite: !course.isFavorite } : course
    ));
  };

  const renderCourseCard = (course: any) => (
    <Card
      key={course.id}
      hoverable
      cover={
        <div style={{ position: 'relative' }}>
          <img
            alt={course.title}
            src={course.thumbnail}
            style={{ height: 200, objectFit: 'cover', width: '100%' }}
          />
          <div style={{
            position: 'absolute',
            top: 8,
            right: 8,
            display: 'flex',
            gap: 4,
          }}>
            {course.tags.map((tag: string) => (
              <Tag key={tag} color={tag === '热门' ? 'red' : tag === '新课' ? 'green' : 'blue'}>
                {tag}
              </Tag>
            ))}
          </div>
          {course.isEnrolled && (
            <div style={{
              position: 'absolute',
              bottom: 8,
              left: 8,
              right: 8,
            }}>
              <Progress
                percent={course.progress}
                size="small"
                strokeColor="#52c41a"
                trailColor="rgba(255,255,255,0.3)"
              />
            </div>
          )}
        </div>
      }
      actions={[
        <Tooltip title={course.isFavorite ? '取消收藏' : '收藏'}>
          <HeartOutlined
            style={{ color: course.isFavorite ? '#ff4d4f' : undefined }}
            onClick={(e) => {
              e.stopPropagation();
              handleFavorite(course.id);
            }}
          />
        </Tooltip>,
        <Tooltip title="分享">
          <ShareAltOutlined />
        </Tooltip>,
        <Tooltip title="查看详情">
          <EyeOutlined onClick={() => handleCourseClick(course)} />
        </Tooltip>,
      ]}
      onClick={() => handleCourseClick(course)}
    >
      <Card.Meta
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text strong ellipsis style={{ flex: 1 }}>
              {course.title}
            </Text>
            <Tag color="blue">{course.level}</Tag>
          </div>
        }
        description={
          <div>
            <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 8 }}>
              {course.description}
            </Paragraph>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
              <Space>
                <Avatar size="small" icon={<UserOutlined />} />
                <Text type="secondary">{course.instructor}</Text>
              </Space>
              <Space>
                <Rate disabled defaultValue={course.rating} style={{ fontSize: 12 }} />
                <Text type="secondary">({course.students})</Text>
              </Space>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Space>
                <ClockCircleOutlined />
                <Text type="secondary">{course.duration}小时</Text>
                <BookOutlined />
                <Text type="secondary">{course.lessons}课时</Text>
              </Space>
              <div>
                <Text delete type="secondary" style={{ marginRight: 8 }}>
                  ¥{course.originalPrice}
                </Text>
                <Text strong style={{ color: '#ff4d4f', fontSize: 16 }}>
                  ¥{course.price}
                </Text>
              </div>
            </div>
          </div>
        }
      />
    </Card>
  );

  const renderCourseDetail = () => (
    <Modal
      title={selectedCourse?.title}
      open={courseDetailVisible}
      onCancel={() => setCourseDetailVisible(false)}
      width={800}
      footer={[
        <Button key="cancel" onClick={() => setCourseDetailVisible(false)}>
          关闭
        </Button>,
        selectedCourse?.isEnrolled ? (
          <Button key="continue" type="primary" icon={<PlayCircleOutlined />}>
            继续学习
          </Button>
        ) : (
          <Button
            key="enroll"
            type="primary"
            icon={<BookOutlined />}
            onClick={() => handleEnroll(selectedCourse?.id)}
          >
            立即报名 ¥{selectedCourse?.price}
          </Button>
        ),
      ]}
    >
      {selectedCourse && (
        <div>
          <Row gutter={[16, 16]}>
            <Col span={24}>
              <img
                src={selectedCourse.thumbnail}
                alt={selectedCourse.title}
                style={{ width: '100%', height: 200, objectFit: 'cover', borderRadius: 8 }}
              />
            </Col>
            <Col span={24}>
              <Space size="large">
                <Statistic title="学员数量" value={selectedCourse.students} prefix={<UserOutlined />} />
                <Statistic title="课程时长" value={selectedCourse.duration} suffix="小时" prefix={<ClockCircleOutlined />} />
                <Statistic title="课时数量" value={selectedCourse.lessons} suffix="节" prefix={<BookOutlined />} />
                <Statistic title="评分" value={selectedCourse.rating} precision={1} prefix={<StarOutlined />} />
              </Space>
            </Col>
          </Row>

          <Divider />

          <Tabs defaultActiveKey="description">
            <TabPane tab="课程介绍" key="description">
              <Paragraph>{selectedCourse.fullDescription}</Paragraph>
              <Title level={5}>适合人群：</Title>
              <div style={{ marginBottom: 16 }}>
                {selectedCourse.targetAudience.map((audience: string) => (
                  <Tag key={audience} color="blue">{audience}</Tag>
                ))}
              </div>
              <Title level={5}>前置要求：</Title>
              <List
                size="small"
                dataSource={selectedCourse.prerequisites}
                renderItem={(item: string) => <List.Item>• {item}</List.Item>}
              />
            </TabPane>
            <TabPane tab="课程大纲" key="outline">
              <Timeline>
                {selectedCourse.outline.map((chapter: any, index: number) => (
                  <Timeline.Item key={index} color="blue">
                    <div>
                      <Text strong>{chapter.title}</Text>
                      <div style={{ marginTop: 4 }}>
                        <Space>
                          <Tag>{chapter.lessons}课时</Tag>
                          <Tag>{chapter.duration}</Tag>
                        </Space>
                      </div>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            </TabPane>
            <TabPane tab="学员评价" key="reviews">
              <List
                dataSource={selectedCourse.reviews}
                renderItem={(review: any) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<UserOutlined />} />}
                      title={
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <Text strong>{review.user}</Text>
                          <Rate disabled defaultValue={review.rating} style={{ fontSize: 12 }} />
                        </div>
                      }
                      description={
                        <div>
                          <Paragraph>{review.comment}</Paragraph>
                          <Text type="secondary">{review.date}</Text>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </TabPane>
          </Tabs>
        </div>
      )}
    </Modal>
  );

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <BookOutlined style={{ marginRight: 8 }} />
          课程中心
        </Title>
        <Text type="secondary">发现优质课程，提升专业技能</Text>
      </div>

      {/* 搜索和筛选区域 */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder="搜索课程、讲师或关键词"
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
            />
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              placeholder="选择分类"
              style={{ width: '100%' }}
              value={selectedCategory}
              onChange={setSelectedCategory}
            >
              {categories.map(category => (
                <Option key={category} value={category}>
                  {category === 'all' ? '全部分类' : category}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              placeholder="选择级别"
              style={{ width: '100%' }}
              value={selectedLevel}
              onChange={setSelectedLevel}
            >
              {levels.map(level => (
                <Option key={level} value={level}>
                  {level === 'all' ? '全部级别' : level}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              placeholder="排序方式"
              style={{ width: '100%' }}
              value={sortBy}
              onChange={setSortBy}
            >
              <Option value="popular">最受欢迎</Option>
              <Option value="rating">评分最高</Option>
              <Option value="newest">最新发布</Option>
              <Option value="price">价格最低</Option>
            </Select>
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Button.Group style={{ width: '100%' }}>
              <Button
                type={viewMode === 'grid' ? 'primary' : 'default'}
                onClick={() => setViewMode('grid')}
                style={{ width: '50%' }}
              >
                网格
              </Button>
              <Button
                type={viewMode === 'list' ? 'primary' : 'default'}
                onClick={() => setViewMode('list')}
                style={{ width: '50%' }}
              >
                列表
              </Button>
            </Button.Group>
          </Col>
        </Row>
      </Card>

      {/* 课程统计 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="全部课程"
              value={courses.length}
              prefix={<BookOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="已报名课程"
              value={courses.filter(c => c.isEnrolled).length}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="收藏课程"
              value={courses.filter(c => c.isFavorite).length}
              prefix={<HeartOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="筛选结果"
              value={filteredCourses.length}
              prefix={<FilterOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 课程列表 */}
      <Spin spinning={loading}>
        {filteredCourses.length > 0 ? (
          <Row gutter={[16, 16]}>
            {filteredCourses.map(course => (
              <Col key={course.id} xs={24} sm={12} md={8} lg={6}>
                {renderCourseCard(course)}
              </Col>
            ))}
          </Row>
        ) : (
          <Empty
            description="暂无符合条件的课程"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        )}
      </Spin>

      {/* 课程详情弹窗 */}
      {renderCourseDetail()}
    </div>
  );
};

export default CourseCenter;