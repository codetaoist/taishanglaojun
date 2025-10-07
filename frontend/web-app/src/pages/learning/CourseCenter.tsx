import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Tag, 
  Input, 
  Select, 
  Pagination,
  Rate,
  Avatar,
  Badge,
  Tooltip,
  Modal,
  Form,
  Slider,
  Checkbox,
  Divider,
  Empty,
  Spin,
  message,
  Tabs,
  List,
  Progress,
  Carousel,
  Affix,
  BackTop,
  Breadcrumb,
  Dropdown,
  Menu
} from 'antd';
import { 
  SearchOutlined, 
  FilterOutlined, 
  BookOutlined, 
  PlayCircleOutlined,
  StarOutlined,
  ClockCircleOutlined,
  UserOutlined,
  EyeOutlined,
  HeartOutlined,
  ShareAltOutlined,
  DownloadOutlined,
  TrophyOutlined,
  FireOutlined,
  ThunderboltOutlined,
  CrownOutlined,
  GiftOutlined,
  TeamOutlined,
  CalendarOutlined,
  TagOutlined,
  SortAscendingOutlined,
  AppstoreOutlined,
  BarsOutlined,
  MoreOutlined,
  LikeOutlined,
  MessageOutlined,
  ShoppingCartOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import moment from 'moment';

const { Title, Paragraph, Text } = Typography;
const { Search } = Input;
const { Option } = Select;
const { TabPane } = Tabs;
const { Meta } = Card;

interface Course {
  id: string;
  title: string;
  description: string;
  instructor: {
    id: string;
    name: string;
    avatar: string;
    title: string;
    rating: number;
  };
  category: string;
  subcategory: string;
  level: 'beginner' | 'intermediate' | 'advanced';
  duration: number; // 总时长（分钟）
  lessonsCount: number;
  rating: number;
  reviewsCount: number;
  studentsCount: number;
  price: number;
  originalPrice?: number;
  isFree: boolean;
  isNew: boolean;
  isHot: boolean;
  isFeatured: boolean;
  thumbnail: string;
  tags: string[];
  skills: string[];
  prerequisites: string[];
  learningOutcomes: string[];
  createdAt: Date;
  updatedAt: Date;
  language: string;
  hasSubtitles: boolean;
  hasCertificate: boolean;
  difficulty: number; // 1-5
  completionRate: number;
  enrollmentStatus?: 'not_enrolled' | 'enrolled' | 'completed';
  progress?: number;
  lastWatched?: Date;
  chapters: {
    id: string;
    title: string;
    duration: number;
    lessonsCount: number;
    isFree: boolean;
  }[];
  reviews: {
    id: string;
    user: {
      name: string;
      avatar: string;
    };
    rating: number;
    comment: string;
    date: Date;
    helpful: number;
  }[];
}

interface Category {
  id: string;
  name: string;
  icon: string;
  count: number;
  subcategories: {
    id: string;
    name: string;
    count: number;
  }[];
}

interface Instructor {
  id: string;
  name: string;
  avatar: string;
  title: string;
  bio: string;
  rating: number;
  studentsCount: number;
  coursesCount: number;
  specialties: string[];
}

const CourseCenter: React.FC = () => {
  const [courses, setCourses] = useState<Course[]>([]);
  const [filteredCourses, setFilteredCourses] = useState<Course[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [instructors, setInstructors] = useState<Instructor[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedLevel, setSelectedLevel] = useState<string>('all');
  const [selectedPrice, setSelectedPrice] = useState<string>('all');
  const [sortBy, setSortBy] = useState<string>('popularity');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(12);
  const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);
  const [courseDetailVisible, setCourseDetailVisible] = useState(false);
  const [filterVisible, setFilterVisible] = useState(false);
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 1000]);
  const [durationRange, setDurationRange] = useState<[number, number]>([0, 50]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [activeTab, setActiveTab] = useState('all');

  useEffect(() => {
    loadCoursesData();
  }, []);

  useEffect(() => {
    filterCourses();
  }, [courses, searchText, selectedCategory, selectedLevel, selectedPrice, sortBy, priceRange, durationRange, selectedTags, activeTab]);

  const loadCoursesData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockCategories: Category[] = [
        {
          id: 'philosophy',
          name: '哲学思想',
          icon: '🧠',
          count: 45,
          subcategories: [
            { id: 'taoism', name: '道家思想', count: 15 },
            { id: 'confucianism', name: '儒家思想', count: 18 },
            { id: 'buddhism', name: '佛家思想', count: 12 }
          ]
        },
        {
          id: 'classics',
          name: '经典文献',
          icon: '📚',
          count: 38,
          subcategories: [
            { id: 'daodejing', name: '道德经', count: 8 },
            { id: 'analects', name: '论语', count: 10 },
            { id: 'yijing', name: '易经', count: 12 },
            { id: 'buddhist', name: '佛经', count: 8 }
          ]
        },
        {
          id: 'practice',
          name: '修行实践',
          icon: '🧘',
          count: 32,
          subcategories: [
            { id: 'meditation', name: '冥想禅修', count: 15 },
            { id: 'qigong', name: '气功养生', count: 10 },
            { id: 'cultivation', name: '修身养性', count: 7 }
          ]
        },
        {
          id: 'modern',
          name: '现代应用',
          icon: '💡',
          count: 28,
          subcategories: [
            { id: 'ai', name: 'AI与传统文化', count: 8 },
            { id: 'management', name: '管理智慧', count: 12 },
            { id: 'psychology', name: '心理学应用', count: 8 }
          ]
        }
      ];

      const mockInstructors: Instructor[] = [
        {
          id: '1',
          name: '张教授',
          avatar: '/api/placeholder/64/64',
          title: '道德经研究专家',
          bio: '从事道家思想研究30年，著有多部相关著作',
          rating: 4.9,
          studentsCount: 15420,
          coursesCount: 12,
          specialties: ['道德经', '庄子', '道家哲学']
        },
        {
          id: '2',
          name: '李教授',
          avatar: '/api/placeholder/64/64',
          title: '儒学大师',
          bio: '儒家思想传承人，致力于传统文化现代化传播',
          rating: 4.8,
          studentsCount: 12350,
          coursesCount: 15,
          specialties: ['论语', '孟子', '儒家礼制']
        },
        {
          id: '3',
          name: '王法师',
          avatar: '/api/placeholder/64/64',
          title: '禅修导师',
          bio: '佛学院毕业，专注禅修指导和心经解读',
          rating: 4.7,
          studentsCount: 8920,
          coursesCount: 8,
          specialties: ['心经', '禅修', '佛学基础']
        }
      ];

      const mockCourses: Course[] = [
        {
          id: '1',
          title: '道德经深度解读',
          description: '深入解读道德经八十一章，领悟老子的智慧精髓，学习道家哲学的核心思想',
          instructor: mockInstructors[0],
          category: 'philosophy',
          subcategory: 'taoism',
          level: 'intermediate',
          duration: 720, // 12小时
          lessonsCount: 24,
          rating: 4.8,
          reviewsCount: 1250,
          studentsCount: 8420,
          price: 299,
          originalPrice: 399,
          isFree: false,
          isNew: false,
          isHot: true,
          isFeatured: true,
          thumbnail: '/api/placeholder/300/200',
          tags: ['道德经', '老子', '道家哲学', '人生智慧'],
          skills: ['哲学思辨', '人生感悟', '智慧应用'],
          prerequisites: ['基础文言文阅读能力'],
          learningOutcomes: [
            '深入理解道德经核心思想',
            '掌握道家哲学精髓',
            '提升人生智慧和修养',
            '学会运用道家思想解决现实问题'
          ],
          createdAt: new Date('2023-06-01'),
          updatedAt: new Date('2024-01-15'),
          language: '中文',
          hasSubtitles: true,
          hasCertificate: true,
          difficulty: 3,
          completionRate: 85,
          enrollmentStatus: 'enrolled',
          progress: 75,
          lastWatched: new Date('2024-02-01'),
          chapters: [
            { id: '1', title: '道可道，非常道', duration: 180, lessonsCount: 6, isFree: true },
            { id: '2', title: '无为而治的智慧', duration: 240, lessonsCount: 8, isFree: false },
            { id: '3', title: '柔弱胜刚强', duration: 300, lessonsCount: 10, isFree: false }
          ],
          reviews: [
            {
              id: '1',
              user: { name: '学员A', avatar: '/api/placeholder/32/32' },
              rating: 5,
              comment: '张教授讲解得非常深入，让我对道德经有了全新的理解',
              date: new Date('2024-01-20'),
              helpful: 45
            }
          ]
        },
        {
          id: '2',
          title: '论语精讲',
          description: '系统学习论语经典篇章，理解孔子的教育思想和人生哲学',
          instructor: mockInstructors[1],
          category: 'classics',
          subcategory: 'analects',
          level: 'beginner',
          duration: 900, // 15小时
          lessonsCount: 30,
          rating: 4.9,
          reviewsCount: 2100,
          studentsCount: 12350,
          price: 0,
          isFree: true,
          isNew: false,
          isHot: true,
          isFeatured: false,
          thumbnail: '/api/placeholder/300/200',
          tags: ['论语', '孔子', '儒家思想', '修身'],
          skills: ['儒家礼仪', '修身养性', '教育智慧'],
          prerequisites: [],
          learningOutcomes: [
            '掌握论语核心思想',
            '理解儒家教育理念',
            '提升个人修养',
            '学会君子之道'
          ],
          createdAt: new Date('2023-03-01'),
          updatedAt: new Date('2024-01-10'),
          language: '中文',
          hasSubtitles: true,
          hasCertificate: true,
          difficulty: 2,
          completionRate: 92,
          enrollmentStatus: 'completed',
          progress: 100,
          chapters: [
            { id: '1', title: '学而时习之', duration: 300, lessonsCount: 10, isFree: true },
            { id: '2', title: '为政以德', duration: 300, lessonsCount: 10, isFree: true },
            { id: '3', title: '君子之道', duration: 300, lessonsCount: 10, isFree: true }
          ],
          reviews: []
        },
        {
          id: '3',
          title: '心经禅修指导',
          description: '通过心经学习禅修方法，体验内心的宁静与智慧',
          instructor: mockInstructors[2],
          category: 'practice',
          subcategory: 'meditation',
          level: 'advanced',
          duration: 360, // 6小时
          lessonsCount: 12,
          rating: 4.7,
          reviewsCount: 680,
          studentsCount: 3420,
          price: 199,
          originalPrice: 299,
          isFree: false,
          isNew: true,
          isHot: false,
          isFeatured: false,
          thumbnail: '/api/placeholder/300/200',
          tags: ['心经', '禅修', '冥想', '佛学'],
          skills: ['禅修技巧', '内观方法', '心性修养'],
          prerequisites: ['基础佛学知识', '冥想经验'],
          learningOutcomes: [
            '掌握心经要义',
            '学会禅修方法',
            '提升专注力',
            '获得内心平静'
          ],
          createdAt: new Date('2024-01-01'),
          updatedAt: new Date('2024-02-01'),
          language: '中文',
          hasSubtitles: true,
          hasCertificate: false,
          difficulty: 4,
          completionRate: 78,
          enrollmentStatus: 'not_enrolled',
          chapters: [
            { id: '1', title: '观自在菩萨', duration: 120, lessonsCount: 4, isFree: true },
            { id: '2', title: '色即是空', duration: 120, lessonsCount: 4, isFree: false },
            { id: '3', title: '禅修实践', duration: 120, lessonsCount: 4, isFree: false }
          ],
          reviews: []
        },
        {
          id: '4',
          title: 'AI与传统文化融合',
          description: '探索人工智能技术如何与传统文化相结合，创造新的学习体验',
          instructor: {
            id: '4',
            name: '陈博士',
            avatar: '/api/placeholder/64/64',
            title: 'AI研究专家',
            rating: 4.6
          },
          category: 'modern',
          subcategory: 'ai',
          level: 'intermediate',
          duration: 480, // 8小时
          lessonsCount: 16,
          rating: 4.6,
          reviewsCount: 320,
          studentsCount: 1850,
          price: 399,
          isFree: false,
          isNew: true,
          isHot: false,
          isFeatured: true,
          thumbnail: '/api/placeholder/300/200',
          tags: ['AI', '传统文化', '创新', '技术'],
          skills: ['AI应用', '文化创新', '技术融合'],
          prerequisites: ['基础AI知识', '传统文化基础'],
          learningOutcomes: [
            '理解AI与传统文化的结合点',
            '掌握文化数字化方法',
            '学会创新应用开发',
            '提升跨领域思维'
          ],
          createdAt: new Date('2024-01-15'),
          updatedAt: new Date('2024-02-01'),
          language: '中文',
          hasSubtitles: true,
          hasCertificate: true,
          difficulty: 3,
          completionRate: 72,
          enrollmentStatus: 'not_enrolled',
          chapters: [
            { id: '1', title: 'AI技术概述', duration: 120, lessonsCount: 4, isFree: true },
            { id: '2', title: '传统文化数字化', duration: 180, lessonsCount: 6, isFree: false },
            { id: '3', title: '融合应用案例', duration: 180, lessonsCount: 6, isFree: false }
          ],
          reviews: []
        }
      ];

      // 添加更多模拟课程数据
      for (let i = 5; i <= 20; i++) {
        const randomCategory = mockCategories[Math.floor(Math.random() * mockCategories.length)];
        const randomInstructor = mockInstructors[Math.floor(Math.random() * mockInstructors.length)];
        const randomLevel = ['beginner', 'intermediate', 'advanced'][Math.floor(Math.random() * 3)] as 'beginner' | 'intermediate' | 'advanced';
        
        mockCourses.push({
          id: i.toString(),
          title: `课程 ${i}`,
          description: `这是第 ${i} 门课程的描述`,
          instructor: randomInstructor,
          category: randomCategory.id,
          subcategory: randomCategory.subcategories[0].id,
          level: randomLevel,
          duration: Math.floor(Math.random() * 600) + 300,
          lessonsCount: Math.floor(Math.random() * 20) + 10,
          rating: Math.round((Math.random() * 1.5 + 3.5) * 10) / 10,
          reviewsCount: Math.floor(Math.random() * 1000) + 100,
          studentsCount: Math.floor(Math.random() * 5000) + 500,
          price: Math.random() > 0.3 ? Math.floor(Math.random() * 400) + 100 : 0,
          isFree: Math.random() > 0.7,
          isNew: Math.random() > 0.8,
          isHot: Math.random() > 0.7,
          isFeatured: Math.random() > 0.9,
          thumbnail: '/api/placeholder/300/200',
          tags: ['标签1', '标签2', '标签3'],
          skills: ['技能1', '技能2'],
          prerequisites: [],
          learningOutcomes: ['学习成果1', '学习成果2'],
          createdAt: new Date(),
          updatedAt: new Date(),
          language: '中文',
          hasSubtitles: true,
          hasCertificate: Math.random() > 0.5,
          difficulty: Math.floor(Math.random() * 5) + 1,
          completionRate: Math.floor(Math.random() * 40) + 60,
          enrollmentStatus: ['not_enrolled', 'enrolled', 'completed'][Math.floor(Math.random() * 3)] as any,
          chapters: [],
          reviews: []
        });
      }

      setCourses(mockCourses);
      setCategories(mockCategories);
      setInstructors(mockInstructors);
      setLoading(false);
    }, 1000);
  };

  const filterCourses = () => {
    let filtered = [...courses];

    // 根据活动标签过滤
    if (activeTab !== 'all') {
      switch (activeTab) {
        case 'featured':
          filtered = filtered.filter(course => course.isFeatured);
          break;
        case 'new':
          filtered = filtered.filter(course => course.isNew);
          break;
        case 'hot':
          filtered = filtered.filter(course => course.isHot);
          break;
        case 'free':
          filtered = filtered.filter(course => course.isFree);
          break;
        case 'enrolled':
          filtered = filtered.filter(course => course.enrollmentStatus === 'enrolled');
          break;
      }
    }

    // 搜索过滤
    if (searchText) {
      filtered = filtered.filter(course =>
        course.title.toLowerCase().includes(searchText.toLowerCase()) ||
        course.description.toLowerCase().includes(searchText.toLowerCase()) ||
        course.instructor.name.toLowerCase().includes(searchText.toLowerCase()) ||
        course.tags.some(tag => tag.toLowerCase().includes(searchText.toLowerCase()))
      );
    }

    // 分类过滤
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(course => course.category === selectedCategory);
    }

    // 难度过滤
    if (selectedLevel !== 'all') {
      filtered = filtered.filter(course => course.level === selectedLevel);
    }

    // 价格过滤
    if (selectedPrice !== 'all') {
      switch (selectedPrice) {
        case 'free':
          filtered = filtered.filter(course => course.isFree);
          break;
        case 'paid':
          filtered = filtered.filter(course => !course.isFree);
          break;
        case 'low':
          filtered = filtered.filter(course => course.price <= 200);
          break;
        case 'medium':
          filtered = filtered.filter(course => course.price > 200 && course.price <= 500);
          break;
        case 'high':
          filtered = filtered.filter(course => course.price > 500);
          break;
      }
    }

    // 价格范围过滤
    filtered = filtered.filter(course => 
      course.price >= priceRange[0] && course.price <= priceRange[1]
    );

    // 时长范围过滤
    filtered = filtered.filter(course => {
      const hours = course.duration / 60;
      return hours >= durationRange[0] && hours <= durationRange[1];
    });

    // 标签过滤
    if (selectedTags.length > 0) {
      filtered = filtered.filter(course =>
        selectedTags.some(tag => course.tags.includes(tag))
      );
    }

    // 排序
    switch (sortBy) {
      case 'popularity':
        filtered.sort((a, b) => b.studentsCount - a.studentsCount);
        break;
      case 'rating':
        filtered.sort((a, b) => b.rating - a.rating);
        break;
      case 'newest':
        filtered.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        break;
      case 'price_low':
        filtered.sort((a, b) => a.price - b.price);
        break;
      case 'price_high':
        filtered.sort((a, b) => b.price - a.price);
        break;
      case 'duration':
        filtered.sort((a, b) => a.duration - b.duration);
        break;
    }

    setFilteredCourses(filtered);
  };

  // 获取难度标签颜色
  const getLevelColor = (level: string) => {
    switch (level) {
      case 'beginner': return 'green';
      case 'intermediate': return 'orange';
      case 'advanced': return 'red';
      default: return 'default';
    }
  };

  // 获取难度中文名称
  const getLevelName = (level: string) => {
    switch (level) {
      case 'beginner': return '初级';
      case 'intermediate': return '中级';
      case 'advanced': return '高级';
      default: return level;
    }
  };

  // 格式化时长
  const formatDuration = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    if (hours > 0) {
      return `${hours}小时${mins > 0 ? mins + '分钟' : ''}`;
    }
    return `${mins}分钟`;
  };

  // 格式化价格
  const formatPrice = (price: number, isFree: boolean) => {
    if (isFree) return '免费';
    return `¥${price}`;
  };

  // 渲染课程卡片
  const renderCourseCard = (course: Course) => (
    <Card
      key={course.id}
      hoverable
      style={{ marginBottom: '16px' }}
      cover={
        <div style={{ position: 'relative' }}>
          <img
            alt={course.title}
            src={course.thumbnail}
            style={{ height: '200px', objectFit: 'cover' }}
          />
          {course.isNew && (
            <Tag color="red" style={{ position: 'absolute', top: '8px', left: '8px' }}>
              新课程
            </Tag>
          )}
          {course.isHot && (
            <Tag color="orange" style={{ position: 'absolute', top: '8px', right: '8px' }}>
              热门
            </Tag>
          )}
          {course.isFeatured && (
            <Tag color="gold" style={{ position: 'absolute', top: course.isNew ? '40px' : '8px', left: '8px' }}>
              精选
            </Tag>
          )}
          <div style={{ 
            position: 'absolute', 
            bottom: '8px', 
            right: '8px',
            background: 'rgba(0,0,0,0.7)',
            color: 'white',
            padding: '4px 8px',
            borderRadius: '4px',
            fontSize: '12px'
          }}>
            {formatDuration(course.duration)}
          </div>
        </div>
      }
      actions={[
        <Tooltip title="查看详情">
          <EyeOutlined 
            onClick={() => {
              setSelectedCourse(course);
              setCourseDetailVisible(true);
            }}
          />
        </Tooltip>,
        <Tooltip title="收藏">
          <HeartOutlined />
        </Tooltip>,
        <Tooltip title="分享">
          <ShareAltOutlined />
        </Tooltip>,
        <Dropdown
          overlay={
            <Menu>
              <Menu.Item key="1" icon={<ShoppingCartOutlined />}>
                加入购物车
              </Menu.Item>
              <Menu.Item key="2" icon={<DownloadOutlined />}>
                下载资料
              </Menu.Item>
            </Menu>
          }
        >
          <MoreOutlined />
        </Dropdown>
      ]}
    >
      <Meta
        title={
          <div>
            <Text strong style={{ fontSize: '16px' }}>{course.title}</Text>
            {course.enrollmentStatus === 'enrolled' && (
              <Tag color="blue" style={{ marginLeft: '8px', fontSize: '10px' }}>
                已报名
              </Tag>
            )}
            {course.enrollmentStatus === 'completed' && (
              <Tag color="green" style={{ marginLeft: '8px', fontSize: '10px' }}>
                已完成
              </Tag>
            )}
          </div>
        }
        description={
          <div>
            <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: '8px' }}>
              {course.description}
            </Paragraph>
            
            <Space direction="vertical" style={{ width: '100%' }} size="small">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <Avatar size="small" src={course.instructor.avatar} />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {course.instructor.name}
                  </Text>
                </Space>
                <Tag color={getLevelColor(course.level)} size="small">
                  {getLevelName(course.level)}
                </Tag>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <Rate disabled value={course.rating} style={{ fontSize: '12px' }} />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    ({course.reviewsCount})
                  </Text>
                </Space>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  <UserOutlined /> {course.studentsCount}
                </Text>
              </div>

              {course.enrollmentStatus === 'enrolled' && course.progress !== undefined && (
                <Progress percent={course.progress} size="small" />
              )}

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space wrap>
                  {course.tags.slice(0, 2).map((tag, index) => (
                    <Tag key={index} size="small">{tag}</Tag>
                  ))}
                </Space>
                <div>
                  {course.originalPrice && (
                    <Text delete type="secondary" style={{ fontSize: '12px', marginRight: '4px' }}>
                      ¥{course.originalPrice}
                    </Text>
                  )}
                  <Text strong style={{ color: course.isFree ? '#52c41a' : '#ff4d4f' }}>
                    {formatPrice(course.price, course.isFree)}
                  </Text>
                </div>
              </div>

              <Button 
                type="primary" 
                block 
                icon={
                  course.enrollmentStatus === 'enrolled' ? <PlayCircleOutlined /> :
                  course.enrollmentStatus === 'completed' ? <CheckCircleOutlined /> :
                  <ShoppingCartOutlined />
                }
              >
                {course.enrollmentStatus === 'enrolled' ? '继续学习' :
                 course.enrollmentStatus === 'completed' ? '重新学习' :
                 course.isFree ? '免费学习' : '立即购买'}
              </Button>
            </Space>
          </div>
        }
      />
    </Card>
  );

  // 渲染列表视图
  const renderCourseList = (course: Course) => (
    <Card key={course.id} style={{ marginBottom: '16px' }}>
      <Row gutter={16}>
        <Col xs={24} sm={8} md={6}>
          <div style={{ position: 'relative' }}>
            <img
              alt={course.title}
              src={course.thumbnail}
              style={{ width: '100%', height: '120px', objectFit: 'cover', borderRadius: '4px' }}
            />
            {course.isNew && (
              <Tag color="red" style={{ position: 'absolute', top: '4px', left: '4px' }}>
                新
              </Tag>
            )}
            {course.isHot && (
              <Tag color="orange" style={{ position: 'absolute', top: '4px', right: '4px' }}>
                热
              </Tag>
            )}
          </div>
        </Col>
        <Col xs={24} sm={16} md={18}>
          <Space direction="vertical" style={{ width: '100%' }} size="small">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <Title level={5} style={{ marginBottom: '4px' }}>
                  {course.title}
                  {course.enrollmentStatus === 'enrolled' && (
                    <Tag color="blue" style={{ marginLeft: '8px' }}>已报名</Tag>
                  )}
                </Title>
                <Text type="secondary">{course.description}</Text>
              </div>
              <div style={{ textAlign: 'right' }}>
                {course.originalPrice && (
                  <Text delete type="secondary" style={{ display: 'block', fontSize: '12px' }}>
                    ¥{course.originalPrice}
                  </Text>
                )}
                <Text strong style={{ color: course.isFree ? '#52c41a' : '#ff4d4f', fontSize: '18px' }}>
                  {formatPrice(course.price, course.isFree)}
                </Text>
              </div>
            </div>

            <Space>
              <Avatar size="small" src={course.instructor.avatar} />
              <Text type="secondary">{course.instructor.name}</Text>
              <Tag color={getLevelColor(course.level)} size="small">
                {getLevelName(course.level)}
              </Tag>
              <Text type="secondary">
                <ClockCircleOutlined /> {formatDuration(course.duration)}
              </Text>
              <Text type="secondary">
                <BookOutlined /> {course.lessonsCount}课时
              </Text>
            </Space>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Space>
                <Rate disabled value={course.rating} style={{ fontSize: '12px' }} />
                <Text type="secondary">({course.reviewsCount})</Text>
                <Text type="secondary">
                  <UserOutlined /> {course.studentsCount}人学习
                </Text>
              </Space>
              <Space>
                <Button size="small" icon={<HeartOutlined />}>收藏</Button>
                <Button 
                  type="primary" 
                  size="small"
                  icon={
                    course.enrollmentStatus === 'enrolled' ? <PlayCircleOutlined /> :
                    <ShoppingCartOutlined />
                  }
                >
                  {course.enrollmentStatus === 'enrolled' ? '继续学习' :
                   course.isFree ? '免费学习' : '立即购买'}
                </Button>
              </Space>
            </div>

            {course.enrollmentStatus === 'enrolled' && course.progress !== undefined && (
              <Progress percent={course.progress} size="small" />
            )}
          </Space>
        </Col>
      </Row>
    </Card>
  );

  // 获取所有标签
  const getAllTags = () => {
    const allTags = new Set<string>();
    courses.forEach(course => {
      course.tags.forEach(tag => allTags.add(tag));
    });
    return Array.from(allTags);
  };

  const paginatedCourses = filteredCourses.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BookOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          课程中心
        </Title>
        <Paragraph>
          发现优质课程，开启智慧学习之旅
        </Paragraph>
      </div>

      {/* 精选课程轮播 */}
      <Card style={{ marginBottom: '24px' }}>
        <Carousel autoplay>
          {courses.filter(course => course.isFeatured).map(course => (
            <div key={course.id}>
              <div style={{ 
                height: '200px', 
                background: `linear-gradient(rgba(0,0,0,0.4), rgba(0,0,0,0.4)), url(${course.thumbnail})`,
                backgroundSize: 'cover',
                backgroundPosition: 'center',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                textAlign: 'center'
              }}>
                <div>
                  <Title level={3} style={{ color: 'white', marginBottom: '8px' }}>
                    {course.title}
                  </Title>
                  <Paragraph style={{ color: 'white', marginBottom: '16px' }}>
                    {course.description}
                  </Paragraph>
                  <Button type="primary" size="large">
                    立即学习
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </Carousel>
      </Card>

      {/* 课程分类标签 */}
      <Card style={{ marginBottom: '24px' }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="全部课程" key="all" />
          <TabPane tab={<Badge count={courses.filter(c => c.isFeatured).length}><span>精选课程</span></Badge>} key="featured" />
          <TabPane tab={<Badge count={courses.filter(c => c.isNew).length}><span>最新课程</span></Badge>} key="new" />
          <TabPane tab={<Badge count={courses.filter(c => c.isHot).length}><span>热门课程</span></Badge>} key="hot" />
          <TabPane tab={<Badge count={courses.filter(c => c.isFree).length}><span>免费课程</span></Badge>} key="free" />
          <TabPane tab={<Badge count={courses.filter(c => c.enrollmentStatus === 'enrolled').length}><span>我的课程</span></Badge>} key="enrolled" />
        </Tabs>
      </Card>

      <Row gutter={24}>
        {/* 侧边栏过滤器 */}
        <Col xs={24} lg={6}>
          <Affix offsetTop={24}>
            <Card title="筛选条件" size="small">
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                {/* 搜索 */}
                <div>
                  <Text strong>搜索课程</Text>
                  <Search
                    placeholder="搜索课程、讲师、标签"
                    value={searchText}
                    onChange={(e) => setSearchText(e.target.value)}
                    style={{ marginTop: '8px' }}
                  />
                </div>

                <Divider style={{ margin: '12px 0' }} />

                {/* 课程分类 */}
                <div>
                  <Text strong>课程分类</Text>
                  <Select
                    value={selectedCategory}
                    onChange={setSelectedCategory}
                    style={{ width: '100%', marginTop: '8px' }}
                  >
                    <Option value="all">全部分类</Option>
                    {categories.map(category => (
                      <Option key={category.id} value={category.id}>
                        {category.icon} {category.name} ({category.count})
                      </Option>
                    ))}
                  </Select>
                </div>

                {/* 难度等级 */}
                <div>
                  <Text strong>难度等级</Text>
                  <Select
                    value={selectedLevel}
                    onChange={setSelectedLevel}
                    style={{ width: '100%', marginTop: '8px' }}
                  >
                    <Option value="all">全部等级</Option>
                    <Option value="beginner">初级</Option>
                    <Option value="intermediate">中级</Option>
                    <Option value="advanced">高级</Option>
                  </Select>
                </div>

                {/* 价格范围 */}
                <div>
                  <Text strong>价格范围</Text>
                  <Select
                    value={selectedPrice}
                    onChange={setSelectedPrice}
                    style={{ width: '100%', marginTop: '8px' }}
                  >
                    <Option value="all">全部价格</Option>
                    <Option value="free">免费</Option>
                    <Option value="low">¥1-200</Option>
                    <Option value="medium">¥201-500</Option>
                    <Option value="high">¥500+</Option>
                  </Select>
                  <Slider
                    range
                    min={0}
                    max={1000}
                    value={priceRange}
                    onChange={setPriceRange}
                    style={{ marginTop: '8px' }}
                  />
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: '#999' }}>
                    <span>¥{priceRange[0]}</span>
                    <span>¥{priceRange[1]}</span>
                  </div>
                </div>

                {/* 课程时长 */}
                <div>
                  <Text strong>课程时长</Text>
                  <Slider
                    range
                    min={0}
                    max={50}
                    value={durationRange}
                    onChange={setDurationRange}
                    style={{ marginTop: '8px' }}
                  />
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: '#999' }}>
                    <span>{durationRange[0]}小时</span>
                    <span>{durationRange[1]}小时</span>
                  </div>
                </div>

                {/* 热门标签 */}
                <div>
                  <Text strong>热门标签</Text>
                  <div style={{ marginTop: '8px' }}>
                    <Checkbox.Group
                      value={selectedTags}
                      onChange={setSelectedTags}
                    >
                      <Space direction="vertical">
                        {getAllTags().slice(0, 8).map(tag => (
                          <Checkbox key={tag} value={tag}>
                            {tag}
                          </Checkbox>
                        ))}
                      </Space>
                    </Checkbox.Group>
                  </div>
                </div>

                <Button 
                  block 
                  onClick={() => {
                    setSearchText('');
                    setSelectedCategory('all');
                    setSelectedLevel('all');
                    setSelectedPrice('all');
                    setPriceRange([0, 1000]);
                    setDurationRange([0, 50]);
                    setSelectedTags([]);
                  }}
                >
                  清除筛选
                </Button>
              </Space>
            </Card>
          </Affix>
        </Col>

        {/* 主要内容区域 */}
        <Col xs={24} lg={18}>
          {/* 工具栏 */}
          <Card style={{ marginBottom: '16px' }}>
            <Row justify="space-between" align="middle">
              <Col>
                <Space>
                  <Text>共找到 {filteredCourses.length} 门课程</Text>
                  <Select
                    value={sortBy}
                    onChange={setSortBy}
                    style={{ width: '120px' }}
                  >
                    <Option value="popularity">最受欢迎</Option>
                    <Option value="rating">评分最高</Option>
                    <Option value="newest">最新发布</Option>
                    <Option value="price_low">价格从低到高</Option>
                    <Option value="price_high">价格从高到低</Option>
                    <Option value="duration">时长最短</Option>
                  </Select>
                </Space>
              </Col>
              <Col>
                <Space>
                  <Button.Group>
                    <Button
                      type={viewMode === 'grid' ? 'primary' : 'default'}
                      icon={<AppstoreOutlined />}
                      onClick={() => setViewMode('grid')}
                    />
                    <Button
                      type={viewMode === 'list' ? 'primary' : 'default'}
                      icon={<BarsOutlined />}
                      onClick={() => setViewMode('list')}
                    />
                  </Button.Group>
                </Space>
              </Col>
            </Row>
          </Card>

          {/* 课程列表 */}
          <Spin spinning={loading}>
            {filteredCourses.length === 0 ? (
              <Card>
                <Empty description="没有找到符合条件的课程" />
              </Card>
            ) : (
              <>
                {viewMode === 'grid' ? (
                  <Row gutter={[16, 16]}>
                    {paginatedCourses.map(course => (
                      <Col xs={24} sm={12} lg={8} key={course.id}>
                        {renderCourseCard(course)}
                      </Col>
                    ))}
                  </Row>
                ) : (
                  <div>
                    {paginatedCourses.map(course => renderCourseList(course))}
                  </div>
                )}

                {/* 分页 */}
                <Card style={{ marginTop: '16px', textAlign: 'center' }}>
                  <Pagination
                    current={currentPage}
                    pageSize={pageSize}
                    total={filteredCourses.length}
                    onChange={(page, size) => {
                      setCurrentPage(page);
                      setPageSize(size || pageSize);
                    }}
                    showSizeChanger
                    showQuickJumper
                    showTotal={(total, range) =>
                      `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                    }
                  />
                </Card>
              </>
            )}
          </Spin>
        </Col>
      </Row>

      {/* 课程详情模态框 */}
      <Modal
        title="课程详情"
        visible={courseDetailVisible}
        onCancel={() => setCourseDetailVisible(false)}
        width={800}
        footer={null}
      >
        {selectedCourse && (
          <div>
            <Row gutter={24}>
              <Col span={10}>
                <img
                  src={selectedCourse.thumbnail}
                  alt={selectedCourse.title}
                  style={{ width: '100%', borderRadius: '8px' }}
                />
              </Col>
              <Col span={14}>
                <Space direction="vertical" style={{ width: '100%' }} size="middle">
                  <div>
                    <Title level={3}>{selectedCourse.title}</Title>
                    <Text type="secondary">{selectedCourse.description}</Text>
                  </div>

                  <Space>
                    <Avatar src={selectedCourse.instructor.avatar} />
                    <div>
                      <Text strong>{selectedCourse.instructor.name}</Text>
                      <br />
                      <Text type="secondary">{selectedCourse.instructor.title}</Text>
                    </div>
                  </Space>

                  <div>
                    <Space>
                      <Rate disabled value={selectedCourse.rating} />
                      <Text>({selectedCourse.reviewsCount} 评价)</Text>
                    </Space>
                    <br />
                    <Text type="secondary">
                      {selectedCourse.studentsCount} 人已学习
                    </Text>
                  </div>

                  <div>
                    <Space wrap>
                      <Tag color={getLevelColor(selectedCourse.level)}>
                        {getLevelName(selectedCourse.level)}
                      </Tag>
                      <Tag>
                        <ClockCircleOutlined /> {formatDuration(selectedCourse.duration)}
                      </Tag>
                      <Tag>
                        <BookOutlined /> {selectedCourse.lessonsCount} 课时
                      </Tag>
                      {selectedCourse.hasCertificate && (
                        <Tag color="gold">
                          <TrophyOutlined /> 可获得证书
                        </Tag>
                      )}
                    </Space>
                  </div>

                  <div>
                    <Text strong style={{ fontSize: '24px', color: selectedCourse.isFree ? '#52c41a' : '#ff4d4f' }}>
                      {formatPrice(selectedCourse.price, selectedCourse.isFree)}
                    </Text>
                    {selectedCourse.originalPrice && (
                      <Text delete type="secondary" style={{ marginLeft: '8px' }}>
                        ¥{selectedCourse.originalPrice}
                      </Text>
                    )}
                  </div>

                  <Space>
                    <Button 
                      type="primary" 
                      size="large"
                      icon={
                        selectedCourse.enrollmentStatus === 'enrolled' ? <PlayCircleOutlined /> :
                        <ShoppingCartOutlined />
                      }
                    >
                      {selectedCourse.enrollmentStatus === 'enrolled' ? '继续学习' :
                       selectedCourse.isFree ? '免费学习' : '立即购买'}
                    </Button>
                    <Button icon={<HeartOutlined />}>收藏</Button>
                    <Button icon={<ShareAltOutlined />}>分享</Button>
                  </Space>
                </Space>
              </Col>
            </Row>

            <Divider />

            <Tabs defaultActiveKey="1">
              <TabPane tab="课程介绍" key="1">
                <div>
                  <Title level={4}>学习成果</Title>
                  <List
                    dataSource={selectedCourse.learningOutcomes}
                    renderItem={item => (
                      <List.Item>
                        <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                        {item}
                      </List.Item>
                    )}
                  />

                  {selectedCourse.prerequisites.length > 0 && (
                    <>
                      <Title level={4}>前置要求</Title>
                      <List
                        dataSource={selectedCourse.prerequisites}
                        renderItem={item => (
                          <List.Item>
                            <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                            {item}
                          </List.Item>
                        )}
                      />
                    </>
                  )}

                  <Title level={4}>课程标签</Title>
                  <Space wrap>
                    {selectedCourse.tags.map((tag, index) => (
                      <Tag key={index}>{tag}</Tag>
                    ))}
                  </Space>
                </div>
              </TabPane>

              <TabPane tab="课程目录" key="2">
                <List
                  dataSource={selectedCourse.chapters}
                  renderItem={(chapter, index) => (
                    <List.Item>
                      <List.Item.Meta
                        title={`第${index + 1}章 ${chapter.title}`}
                        description={
                          <Space>
                            <Text type="secondary">
                              <ClockCircleOutlined /> {formatDuration(chapter.duration)}
                            </Text>
                            <Text type="secondary">
                              <BookOutlined /> {chapter.lessonsCount} 课时
                            </Text>
                            {chapter.isFree && (
                              <Tag color="green" size="small">免费试看</Tag>
                            )}
                          </Space>
                        }
                      />
                      <Button type="link" icon={<PlayCircleOutlined />}>
                        {chapter.isFree ? '免费试看' : '开始学习'}
                      </Button>
                    </List.Item>
                  )}
                />
              </TabPane>

              <TabPane tab="讲师介绍" key="3">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Avatar size={64} src={selectedCourse.instructor.avatar} />
                    <div style={{ marginLeft: '16px' }}>
                      <Title level={4} style={{ marginBottom: '4px' }}>
                        {selectedCourse.instructor.name}
                      </Title>
                      <Text type="secondary">{selectedCourse.instructor.title}</Text>
                      <br />
                      <Rate disabled value={selectedCourse.instructor.rating} style={{ fontSize: '14px' }} />
                    </div>
                  </div>
                  
                  <Row gutter={16}>
                    <Col span={8}>
                      <Statistic title="学生数量" value={selectedCourse.instructor.studentsCount} />
                    </Col>
                    <Col span={8}>
                      <Statistic title="课程数量" value={selectedCourse.instructor.coursesCount} />
                    </Col>
                    <Col span={8}>
                      <Statistic title="评分" value={selectedCourse.instructor.rating} precision={1} />
                    </Col>
                  </Row>
                </Space>
              </TabPane>

              <TabPane tab="学员评价" key="4">
                {selectedCourse.reviews.length > 0 ? (
                  <List
                    dataSource={selectedCourse.reviews}
                    renderItem={review => (
                      <List.Item>
                        <List.Item.Meta
                          avatar={<Avatar src={review.user.avatar} />}
                          title={
                            <Space>
                              <Text strong>{review.user.name}</Text>
                              <Rate disabled value={review.rating} style={{ fontSize: '12px' }} />
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {moment(review.date).format('YYYY-MM-DD')}
                              </Text>
                            </Space>
                          }
                          description={review.comment}
                        />
                        <div>
                          <Button type="link" size="small" icon={<LikeOutlined />}>
                            有用 ({review.helpful})
                          </Button>
                        </div>
                      </List.Item>
                    )}
                  />
                ) : (
                  <Empty description="暂无评价" />
                )}
              </TabPane>
            </Tabs>
          </div>
        )}
      </Modal>

      <BackTop />
    </div>
  );
};

export default CourseCenter;