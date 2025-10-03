import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Tag, 
  Space, 
  Button, 
  Divider, 
  Spin,
  message,
  Row,
  Col,
  Breadcrumb,
  BackTop,
  Tabs
} from 'antd';
import { 
  ArrowLeftOutlined,
  BookOutlined,
  EyeOutlined,
  HeartOutlined,
  ShareAltOutlined,
  StarOutlined,
  CalendarOutlined,
  UserOutlined,
  RobotOutlined,
  BulbOutlined,
  BarChartOutlined,
  EditOutlined
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';
import type { CulturalWisdom } from '../types';
import WisdomInterpretation from '../components/ai/WisdomInterpretation';
import WisdomRecommendation from '../components/ai/WisdomRecommendation';
import WisdomAnalysis from '../components/ai/WisdomAnalysis';
import WisdomRecommendations from '../components/WisdomRecommendations';
import NoteModal from '../components/notes/NoteModal';
import behaviorService from '../services/behaviorService';

const { Title, Paragraph } = Typography;

const WisdomDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [wisdom, setWisdom] = useState<CulturalWisdom | null>(null);
  const [loading, setLoading] = useState(false);
  const [liked, setLiked] = useState(false);
  const [favorited, setFavorited] = useState(false);
  const [noteModalVisible, setNoteModalVisible] = useState(false);
  const [existingNote, setExistingNote] = useState<any>(null);

  // 加载智慧详情
  const loadWisdomDetail = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const response = await apiClient.getWisdomById(id);
      
      if (response.success && response.data) {
        setWisdom(response.data);
        
        // 检查收藏状态
        try {
          const favoriteResponse = await apiClient.checkFavoriteStatus(id);
          if (favoriteResponse.success) {
            setFavorited(favoriteResponse.data.is_favorited);
          }
        } catch (error) {
          console.error('检查收藏状态失败:', error);
        }

        // 检查是否有笔记
        try {
          const noteResponse = await apiClient.getNote(id);
          if (noteResponse.success && noteResponse.data) {
            setExistingNote(noteResponse.data);
          }
        } catch (error) {
          // 没有笔记是正常情况，不需要显示错误
          console.log('用户暂无笔记');
        }
        
        // 记录浏览行为
        await behaviorService.recordWisdomView(id, {
          category: response.data.category,
          source: response.data.source,
          dynasty: response.data.dynasty,
          tags: response.data.tags,
          viewTime: new Date().toISOString()
        });
      } else {
        message.error('获取智慧详情失败');
        navigate('/wisdom');
      }
    } catch (error) {
      console.error('加载智慧详情失败:', error);
      message.error('网络错误，请稍后重试');
      navigate('/wisdom');
    }
    setLoading(false);
  };

  // 点赞
  const handleLike = async () => {
    if (!wisdom) return;
    
    try {
      // 记录点赞行为
      await behaviorService.recordWisdomLike(wisdom.id, !liked);
      
      // 这里应该调用点赞API
      setLiked(!liked);
      message.success(liked ? '取消点赞' : '点赞成功');
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 收藏
  const handleFavorite = async () => {
    if (!wisdom) return;
    
    try {
      if (favorited) {
        await apiClient.removeFavorite(wisdom.id);
        setFavorited(false);
        message.success('取消收藏');
      } else {
        await apiClient.addFavorite(wisdom.id);
        setFavorited(true);
        message.success('收藏成功');
      }
      
      // 记录收藏行为
      await behaviorService.recordBookmark(wisdom.id, !favorited);
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 分享
  const handleShare = async () => {
    if (!wisdom) return;
    
    try {
      // 记录分享行为
      await behaviorService.recordWisdomShare(wisdom.id, navigator.share ? 'native' : 'clipboard');
      
      if (navigator.share) {
        await navigator.share({
          title: wisdom?.title,
          text: wisdom?.content?.substring(0, 100) + '...',
          url: window.location.href,
        });
      } else {
        // 复制链接到剪贴板
        await navigator.clipboard.writeText(window.location.href);
        message.success('链接已复制到剪贴板');
      }
    } catch (error) {
      console.error('分享失败:', error);
      message.error('分享失败');
    }
  };

  // 处理笔记操作
  const handleNoteSuccess = () => {
    // 重新加载笔记状态
    if (id) {
      apiClient.getNote(id).then(response => {
        if (response.success && response.data) {
          setExistingNote(response.data);
        } else {
          setExistingNote(null);
        }
      }).catch(() => {
        setExistingNote(null);
      });
    }
  };

  // 组件挂载时加载数据
  useEffect(() => {
    loadWisdomDetail();
  }, [id]);

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <Spin size="large" />
      </div>
    );
  }

  if (!wisdom) {
    return (
      <div className="text-center py-12">
        <BookOutlined className="text-6xl text-gray-300 mb-4" />
        <Title level={3} type="secondary">智慧内容不存在</Title>
        <Button type="primary" onClick={() => navigate('/wisdom')}>
          返回智慧库
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 面包屑导航 */}
      <Breadcrumb
        items={[
          {
            title: (
              <span onClick={() => navigate('/')} className="cursor-pointer hover:text-primary-500">
                首页
              </span>
            ),
          },
          {
            title: (
              <span onClick={() => navigate('/wisdom')} className="cursor-pointer hover:text-primary-500">
                文化智慧
              </span>
            ),
          },
          {
            title: wisdom.title,
          },
        ]}
      />

      {/* 返回按钮 */}
      <Button 
        icon={<ArrowLeftOutlined />} 
        onClick={() => navigate('/wisdom')}
        className="mb-4"
      >
        返回智慧库
      </Button>

      <Row gutter={[24, 24]}>
        {/* 主要内容 */}
        <Col xs={24} lg={18}>
          <Card className="shadow-lg">
            {/* 标题和基本信息 */}
            <div className="space-y-6">
              <div>
                <Title level={1} className="mb-4 text-slate-800">
                  {wisdom.title}
                </Title>
                
                <div className="flex flex-wrap items-center gap-3 mb-6">
                  <Tag color="gold" className="text-sm px-3 py-1">
                    <BookOutlined className="mr-1" />
                    {wisdom.category}
                  </Tag>
                  {wisdom.source && (
                    <Tag color="blue" className="text-sm px-3 py-1">
                      <UserOutlined className="mr-1" />
                      {wisdom.source}
                    </Tag>
                  )}
                  {wisdom.dynasty && (
                    <Tag color="green" className="text-sm px-3 py-1">
                      <CalendarOutlined className="mr-1" />
                      {wisdom.dynasty}
                    </Tag>
                  )}
                </div>
              </div>

              <Divider />

              {/* 正文内容 */}
              <div className="prose max-w-none">
                <Paragraph className="text-lg leading-relaxed text-gray-700 whitespace-pre-wrap">
                  {wisdom.content}
                </Paragraph>
              </div>

              {/* 解释说明 */}
              {wisdom.explanation && (
                <>
                  <Divider>
                    <span className="text-gray-500">智慧解读</span>
                  </Divider>
                  <Card size="small" className="bg-blue-50 border-blue-200">
                    <Paragraph className="text-gray-700 leading-relaxed whitespace-pre-wrap mb-0">
                      {wisdom.explanation}
                    </Paragraph>
                  </Card>
                </>
              )}

              {/* AI功能区域 */}
              <Divider>
                <Space>
                  <RobotOutlined className="text-blue-500" />
                  <span className="text-gray-500">AI智能功能</span>
                </Space>
              </Divider>
              
              <Tabs
                defaultActiveKey="interpretation"
                items={[
                  {
                    key: 'interpretation',
                    label: (
                      <Space>
                        <BulbOutlined />
                        AI解读
                      </Space>
                    ),
                    children: <WisdomInterpretation wisdomId={wisdom.id} />,
                  },
                  {
                    key: 'analysis',
                    label: (
                      <Space>
                        <BarChartOutlined />
                        深度分析
                      </Space>
                    ),
                    children: <WisdomAnalysis wisdomId={wisdom.id} />,
                  },
                  {
                    key: 'recommendation',
                    label: (
                      <Space>
                        <BookOutlined />
                        相关推荐
                      </Space>
                    ),
                    children: <WisdomRecommendations wisdomId={wisdom.id} onWisdomClick={(id) => navigate(`/wisdom/${id}`)} />,
                  },
                ]}
                className="mb-6"
              />

              {/* 标签 */}
              {wisdom.tags && wisdom.tags.length > 0 && (
                <>
                  <Divider>
                    <span className="text-gray-500">相关标签</span>
                  </Divider>
                  <div className="flex flex-wrap gap-2">
                    {wisdom.tags.map((tag, index) => (
                      <Tag 
                        key={index} 
                        color="processing" 
                        className="cursor-pointer hover:bg-blue-100"
                        onClick={() => navigate(`/wisdom?tag=${tag}`)}
                      >
                        {tag}
                      </Tag>
                    ))}
                  </div>
                </>
              )}
            </div>
          </Card>
        </Col>

        {/* 侧边栏 */}
        <Col xs={24} lg={6}>
          <div className="space-y-4 sticky top-4">
            {/* 操作按钮 */}
            <Card title="操作" size="small">
              <Space direction="vertical" className="w-full">
                <Button 
                  type={liked ? "primary" : "default"}
                  icon={<HeartOutlined />}
                  onClick={handleLike}
                  block
                  className={liked ? "bg-red-500 border-red-500" : ""}
                >
                  {liked ? '已点赞' : '点赞'} ({wisdom.likes || 0})
                </Button>
                
                <Button 
                  type={favorited ? "primary" : "default"}
                  icon={<StarOutlined />}
                  onClick={handleFavorite}
                  block
                  className={favorited ? "bg-yellow-500 border-yellow-500" : ""}
                >
                  {favorited ? '已收藏' : '收藏'}
                </Button>
                
                <Button 
                  icon={<ShareAltOutlined />}
                  onClick={handleShare}
                  block
                >
                  分享
                </Button>
                
                <Button 
                  type={existingNote ? "primary" : "default"}
                  icon={<EditOutlined />}
                  onClick={() => setNoteModalVisible(true)}
                  block
                  className={existingNote ? "bg-green-500 border-green-500" : ""}
                >
                  {existingNote ? '编辑笔记' : '添加笔记'}
                </Button>
              </Space>
            </Card>

            {/* 统计信息 */}
            <Card title="统计信息" size="small">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Space>
                    <EyeOutlined className="text-gray-500" />
                    <span className="text-gray-500">阅读量</span>
                  </Space>
                  <span className="font-semibold">{wisdom.views || 0}</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <Space>
                    <HeartOutlined className="text-red-500" />
                    <span className="text-gray-500">点赞数</span>
                  </Space>
                  <span className="font-semibold">{wisdom.likes || 0}</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <Space>
                    <StarOutlined className="text-yellow-500" />
                    <span className="text-gray-500">收藏数</span>
                  </Space>
                  <span className="font-semibold">{wisdom.favorites || 0}</span>
                </div>
              </div>
            </Card>

            {/* 相关信息 */}
            <Card title="相关信息" size="small">
              <div className="space-y-2">
                {wisdom.author && (
                  <div>
                    <span className="text-gray-500">作者：</span>
                    <span>{wisdom.author}</span>
                  </div>
                )}
                
                {wisdom.source && (
                  <div>
                    <span className="text-gray-500">出处：</span>
                    <span>{wisdom.source}</span>
                  </div>
                )}
                
                {wisdom.dynasty && (
                  <div>
                    <span className="text-gray-500">朝代：</span>
                    <span>{wisdom.dynasty}</span>
                  </div>
                )}
                
                {wisdom.createdAt && (
                  <div>
                    <span className="text-gray-500">收录时间：</span>
                    <span>{new Date(wisdom.createdAt).toLocaleDateString()}</span>
                  </div>
                )}
              </div>
            </Card>
          </div>
        </Col>
      </Row>

      {/* 笔记模态框 */}
      <NoteModal
        visible={noteModalVisible}
        onCancel={() => setNoteModalVisible(false)}
        onSuccess={handleNoteSuccess}
        wisdomId={wisdom.id}
        wisdomTitle={wisdom.title}
        existingNote={existingNote}
      />

      <BackTop />
    </div>
  );
};

export default WisdomDetail;