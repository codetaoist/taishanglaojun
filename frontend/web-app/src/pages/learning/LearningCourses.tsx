import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Tag, Progress, Input, Select, Pagination, message } from 'antd';
import { 
  PlayCircleOutlined, 
  BookOutlined, 
  ClockCircleOutlined,
  UserOutlined,
  StarOutlined,
  SearchOutlined,
  FilterOutlined
} from '@ant-design/icons';

const { Search } = Input;
const { Option } = Select;

interface Course {
  id: string;
  title: string;
  description: string;
  instructor: string;
  duration: string;
  level: 'beginner' | 'intermediate' | 'advanced';
  category: string;
  rating: number;
  students: number;
  progress?: number;
  thumbnail: string;
  price: number;
  tags: string[];
  isEnrolled: boolean;
}

const LearningCourses: React.FC = () => {
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedLevel, setSelectedLevel] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(12);

  // 模拟课程数据
  const mockCourses: Course[] = [
    {
      id: '1',
      title: 'AI基础入门课程',
      description: '从零开始学习人工智能的基础概念和应用',
      instructor: '张教授',
      duration: '8小时',
      level: 'beginner',
      category: 'AI技术',
      rating: 4.8,
      students: 1250,
      progress: 65,
      thumbnail: '/api/placeholder/300/200',
      price: 299,
      tags: ['AI', '机器学习', '入门'],
      isEnrolled: true
    },
    {
      id: '2',
      title: '深度学习实战',
      description: '通过实际项目学习深度学习的核心技术',
      instructor: '李博士',
      duration: '12小时',
      level: 'advanced',
      category: 'AI技术',
      rating: 4.9,
      students: 890,
      thumbnail: '/api/placeholder/300/200',
      price: 599,
      tags: ['深度学习', 'TensorFlow', '实战'],
      isEnrolled: false
    },
    {
      id: '3',
      title: 'React开发进阶',
      description: '掌握React高级特性和最佳实践',
      instructor: '王工程师',
      duration: '10小时',
      level: 'intermediate',
      category: '前端开发',
      rating: 4.7,
      students: 2100,
      progress: 30,
      thumbnail: '/api/placeholder/300/200',
      price: 399,
      tags: ['React', 'JavaScript', '前端'],
      isEnrolled: true
    },
    {
      id: '4',
      title: '数据分析与可视化',
      description: '学习使用Python进行数据分析和可视化',
      instructor: '陈分析师',
      duration: '15小时',
      level: 'intermediate',
      category: '数据科学',
      rating: 4.6,
      students: 1680,
      thumbnail: '/api/placeholder/300/200',
      price: 459,
      tags: ['Python', '数据分析', '可视化'],
      isEnrolled: false
    },
    {
      id: '5',
      title: '云计算架构设计',
      description: '学习现代云计算架构的设计原则和实践',
      instructor: '刘架构师',
      duration: '20小时',
      level: 'advanced',
      category: '云计算',
      rating: 4.8,
      students: 756,
      thumbnail: '/api/placeholder/300/200',
      price: 799,
      tags: ['云计算', 'AWS', '架构设计'],
      isEnrolled: false
    },
    {
      id: '6',
      title: 'UI/UX设计基础',
      description: '掌握用户界面和用户体验设计的基本原理',
      instructor: '赵设计师',
      duration: '6小时',
      level: 'beginner',
      category: '设计',
      rating: 4.5,
      students: 1890,
      thumbnail: '/api/placeholder/300/200',
      price: 199,
      tags: ['UI设计', 'UX设计', '原型'],
      isEnrolled: false
    }
  ];

  useEffect(() => {
    loadCourses();
  }, []);

  const loadCourses = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      setCourses(mockCourses);
    } catch (error) {
      message.error('加载课程失败');
    } finally {
      setLoading(false);
    }
  };

  const handleEnroll = async (courseId: string) => {
    try {
      // 模拟报名API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      
      setCourses(prev => prev.map(course => 
        course.id === courseId 
          ? { ...course, isEnrolled: true, progress: 0 }
          : course
      ));
      
      message.success('报名成功！');
    } catch (error) {
      message.error('报名失败，请重试');
    }
  };

  const filteredCourses = courses.filter(course => {
    const matchesSearch = course.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         course.description.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory = selectedCategory === 'all' || course.category === selectedCategory;
    const matchesLevel = selectedLevel === 'all' || course.level === selectedLevel;
    
    return matchesSearch && matchesCategory && matchesLevel;
  });

  const paginatedCourses = filteredCourses.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'beginner': return 'green';
      case 'intermediate': return 'orange';
      case 'advanced': return 'red';
      default: return 'default';
    }
  };

  const getLevelText = (level: string) => {
    switch (level) {
      case 'beginner': return '初级';
      case 'intermediate': return '中级';
      case 'advanced': return '高级';
      default: return level;
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">学习课程</h1>
        <p className="text-gray-600">
          探索丰富的在线课程，提升您的专业技能
        </p>
      </div>

      {/* 搜索和筛选 */}
      <Card className="mb-6">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder="搜索课程..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              prefix={<SearchOutlined />}
            />
          </Col>
          
          <Col xs={12} sm={6} md={4}>
            <Select
              value={selectedCategory}
              onChange={setSelectedCategory}
              className="w-full"
              placeholder="选择分类"
            >
              <Option value="all">全部分类</Option>
              <Option value="AI技术">AI技术</Option>
              <Option value="前端开发">前端开发</Option>
              <Option value="数据科学">数据科学</Option>
              <Option value="云计算">云计算</Option>
              <Option value="设计">设计</Option>
            </Select>
          </Col>
          
          <Col xs={12} sm={6} md={4}>
            <Select
              value={selectedLevel}
              onChange={setSelectedLevel}
              className="w-full"
              placeholder="选择难度"
            >
              <Option value="all">全部难度</Option>
              <Option value="beginner">初级</Option>
              <Option value="intermediate">中级</Option>
              <Option value="advanced">高级</Option>
            </Select>
          </Col>
        </Row>
      </Card>

      {/* 课程列表 */}
      <Row gutter={[16, 16]}>
        {paginatedCourses.map(course => (
          <Col key={course.id} xs={24} sm={12} lg={8} xl={6}>
            <Card
              hoverable
              cover={
                <div className="h-48 bg-gradient-to-r from-blue-400 to-purple-500 flex items-center justify-center">
                  <BookOutlined className="text-4xl text-white" />
                </div>
              }
              actions={[
                course.isEnrolled ? (
                  <Button type="primary" icon={<PlayCircleOutlined />}>
                    继续学习
                  </Button>
                ) : (
                  <Button 
                    type="primary" 
                    onClick={() => handleEnroll(course.id)}
                  >
                    立即报名 ¥{course.price}
                  </Button>
                )
              ]}
            >
              <div className="mb-2">
                <h3 className="text-lg font-semibold mb-1 line-clamp-2">
                  {course.title}
                </h3>
                <p className="text-gray-600 text-sm line-clamp-2 mb-3">
                  {course.description}
                </p>
              </div>

              <div className="space-y-2 mb-3">
                <div className="flex items-center text-sm text-gray-500">
                  <UserOutlined className="mr-1" />
                  {course.instructor}
                </div>
                
                <div className="flex items-center text-sm text-gray-500">
                  <ClockCircleOutlined className="mr-1" />
                  {course.duration}
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="flex items-center text-sm">
                    <StarOutlined className="text-yellow-500 mr-1" />
                    {course.rating} ({course.students}人)
                  </div>
                  <Tag color={getLevelColor(course.level)}>
                    {getLevelText(course.level)}
                  </Tag>
                </div>
              </div>

              {course.isEnrolled && course.progress !== undefined && (
                <div className="mb-3">
                  <div className="flex justify-between text-sm mb-1">
                    <span>学习进度</span>
                    <span>{course.progress}%</span>
                  </div>
                  <Progress percent={course.progress} size="small" />
                </div>
              )}

              <div className="flex flex-wrap gap-1">
                {course.tags.map(tag => (
                  <Tag key={tag} size="small">{tag}</Tag>
                ))}
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 分页 */}
      {filteredCourses.length > pageSize && (
        <div className="mt-6 text-center">
          <Pagination
            current={currentPage}
            total={filteredCourses.length}
            pageSize={pageSize}
            onChange={setCurrentPage}
            showSizeChanger={false}
            showQuickJumper
            showTotal={(total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条课程`
            }
          />
        </div>
      )}
    </div>
  );
};

export default LearningCourses;