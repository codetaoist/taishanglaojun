import React, { useState, useRef, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Input, 
  Select, 
  Slider, 
  Typography, 
  Space, 
  Upload, 
  Image, 
  Spin, 
  message, 
  Tabs, 
  Tag, 
  Tooltip,
  Progress,
  Modal,
  Divider
} from 'antd';
import { 
  PictureOutlined, 
  UploadOutlined, 
  DownloadOutlined, 
  ShareAltOutlined,
  ThunderboltOutlined,
  SettingOutlined,
  HistoryOutlined,
  StarOutlined,
  ReloadOutlined,
  CopyOutlined,
  DeleteOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { aiService, type ImageGenerationRequest, type ImageGenerationResponse } from '../../services/aiService';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;


interface GeneratedImage {
  id: string;
  url: string;
  prompt: string;
  style: string;
  timestamp: Date;
  liked: boolean;
  metadata?: {
    steps: number;
    guidance: number;
    quality: number;
    seed: number;
  };
}

const ImageGeneration: React.FC = () => {
  const [activeTab, setActiveTab] = useState('text-to-image');
  const [prompt, setPrompt] = useState('');
  const [style, setStyle] = useState('realistic');
  const [size, setSize] = useState('1024x1024');
  const [quality, setQuality] = useState(80);
  const [steps, setSteps] = useState(30);
  const [guidance, setGuidance] = useState(7.5);
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [generatedImages, setGeneratedImages] = useState<GeneratedImage[]>([]);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const fileInputRef = useRef<any>(null);

  // 加载历史记录
  useEffect(() => {
    loadGenerationHistory();
  }, []);

  const loadGenerationHistory = async () => {
    setHistoryLoading(true);
    try {
      const response = await aiService.getGenerationHistory(20);
      if (response.code === 'SUCCESS') {
        const historyImages: GeneratedImage[] = response.data.map(item => ({
          id: item.id,
          url: item.url,
          prompt: item.prompt,
          style: item.style,
          timestamp: new Date(item.timestamp),
          liked: false,
          metadata: item.metadata
        }));
        setGeneratedImages(historyImages);
      }
    } catch (error) {
      console.error('加载历史记录失败:', error);
      // 如果API失败，使用一些模拟数据
      const mockImages: GeneratedImage[] = [
        {
          id: '1',
          url: 'https://picsum.photos/1024/1024?random=1',
          prompt: '一幅宁静的山水画',
          style: 'realistic',
          timestamp: new Date(Date.now() - 3600000),
          liked: false,
          metadata: { steps: 30, guidance: 7.5, quality: 80, seed: 123456 }
        },
        {
          id: '2',
          url: 'https://picsum.photos/1024/1024?random=2',
          prompt: '现代都市夜景',
          style: 'realistic',
          timestamp: new Date(Date.now() - 7200000),
          liked: true,
          metadata: { steps: 30, guidance: 7.5, quality: 80, seed: 789012 }
        }
      ];
      setGeneratedImages(mockImages);
    } finally {
      setHistoryLoading(false);
    }
  };

  // 预设风格选项
  const styleOptions = [
    { label: '写实风格', value: 'realistic' },
    { label: '动漫风格', value: 'anime' },
    { label: '油画风格', value: 'oil-painting' },
    { label: '水彩风格', value: 'watercolor' },
    { label: '素描风格', value: 'sketch' },
    { label: '科幻风格', value: 'sci-fi' },
    { label: '古典风格', value: 'classical' },
    { label: '现代艺术', value: 'modern-art' }
  ];

  // 尺寸选项
  const sizeOptions = [
    { label: '正方形 (1024×1024)', value: '1024x1024' },
    { label: '横向 (1344×768)', value: '1344x768' },
    { label: '纵向 (768×1344)', value: '768x1344' },
    { label: '宽屏 (1536×640)', value: '1536x640' }
  ];

  // 预设提示词
  const promptTemplates = [
    '一幅宁静的山水画，远山如黛，近水如镜',
    '现代都市夜景，霓虹灯闪烁，车水马龙',
    '古典园林，亭台楼阁，小桥流水',
    '科幻城市，未来建筑，飞行器穿梭',
    '温馨的咖啡厅，阳光透过窗户洒在桌上',
    '神秘的森林，阳光透过树叶形成光斑'
  ];

  // 图像生成
  const handleGenerate = async () => {
    if (!prompt.trim()) {
      message.warning('请输入图像描述');
      return;
    }

    setLoading(true);
    setProgress(0);

    // 模拟生成进度
    const progressInterval = setInterval(() => {
      setProgress(prev => {
        if (prev >= 90) {
          clearInterval(progressInterval);
          return 90;
        }
        return prev + Math.random() * 15;
      });
    }, 200);

    try {
      const request: ImageGenerationRequest = {
        prompt: prompt.trim(),
        style,
        size,
        quality,
        steps,
        guidance,
      };

      const response = await aiService.generateImage(request);
      
      if (response.code === 'SUCCESS') {
        const apiResult = response.data;
        const newImage: GeneratedImage = {
          id: apiResult.id,
          url: apiResult.url,
          prompt: apiResult.prompt,
          style: apiResult.style,
          timestamp: new Date(apiResult.timestamp),
          liked: false,
          metadata: apiResult.metadata
        };

        setGeneratedImages(prev => [newImage, ...prev]);
        setProgress(100);
        message.success('图像生成成功！');
      } else {
        throw new Error(response.message || '生成失败');
      }
    } catch (error: any) {
      console.error('图像生成失败:', error);
      message.error(error.message || '生成失败，请重试');
      
      // 如果API失败，使用模拟数据作为后备
      const fallbackImage: GeneratedImage = {
        id: Date.now().toString(),
        url: `https://picsum.photos/1024/1024?random=${Date.now()}`,
        prompt,
        style,
        timestamp: new Date(),
        liked: false,
        metadata: {
          steps,
          guidance,
          quality,
          seed: Math.floor(Math.random() * 1000000)
        }
      };
      
      setGeneratedImages(prev => [fallbackImage, ...prev]);
      setProgress(100);
      message.warning('API调用失败，使用模拟数据作为示例');
    } finally {
      clearInterval(progressInterval);
      setLoading(false);
      setTimeout(() => setProgress(0), 1000);
    }
  };

  // 处理图像上传
  const handleImageUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      // 处理上传的图像
      message.success('图像上传成功，可以开始编辑');
    };
    reader.readAsDataURL(file);
    return false; // 阻止默认上传行为
  };

  // 切换喜欢状态
  const toggleLike = (imageId: string) => {
    setGeneratedImages(prev => 
      prev.map(img => 
        img.id === imageId ? { ...img, liked: !img.liked } : img
      )
    );
  };

  // 下载图像
  const downloadImage = (imageUrl: string, prompt: string) => {
    const link = document.createElement('a');
    link.href = imageUrl;
    link.download = `generated-image-${Date.now()}.png`;
    link.click();
    message.success('图像下载成功');
  };

  // 复制提示词
  const copyPrompt = (prompt: string) => {
    navigator.clipboard.writeText(prompt);
    message.success('提示词已复制到剪贴板');
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <ThunderboltOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          AI图像生成
        </Title>
        <Paragraph>
          使用先进的AI技术，将您的创意想法转化为精美的图像作品
        </Paragraph>
      </div>

      <Row gutter={[24, 24]}>
        {/* 左侧控制面板 */}
        <Col xs={24} lg={8}>
          <Card title="生成设置" extra={<SettingOutlined />}>
            <Tabs 
              activeKey={activeTab} 
              onChange={setActiveTab}
              items={[
                {
                  key: 'text-to-image',
                  label: '文本生成',
                  children: (
                    <Space direction="vertical" style={{ width: '100%' }} size="middle">
                      <div>
                        <Text strong>图像描述</Text>
                        <TextArea
                          value={prompt}
                          onChange={(e) => setPrompt(e.target.value)}
                          placeholder="描述您想要生成的图像..."
                          rows={4}
                          maxLength={500}
                          showCount
                        />
                      </div>

                      <div>
                        <Text strong>快速模板</Text>
                        <div style={{ marginTop: '8px' }}>
                          <Space wrap>
                            {promptTemplates.map((template, index) => (
                              <Tag
                                key={index}
                                style={{ cursor: 'pointer', marginBottom: '4px' }}
                                onClick={() => setPrompt(template)}
                              >
                                {template.substring(0, 15)}...
                              </Tag>
                            ))}
                          </Space>
                        </div>
                      </div>

                      <div>
                        <Text strong>艺术风格</Text>
                        <Select
                          value={style}
                          onChange={setStyle}
                          style={{ width: '100%', marginTop: '8px' }}
                          options={styleOptions}
                        />
                      </div>

                      <div>
                        <Text strong>图像尺寸</Text>
                        <Select
                          value={size}
                          onChange={setSize}
                          style={{ width: '100%', marginTop: '8px' }}
                          options={sizeOptions}
                        />
                      </div>
                    </Space>
                  )
                },
                {
                  key: 'image-edit',
                  label: '图像编辑',
                  children: (
                    <Space direction="vertical" style={{ width: '100%' }} size="middle">
                      <div>
                        <Text strong>上传图像</Text>
                        <Upload
                          accept="image/*"
                          beforeUpload={handleImageUpload}
                          showUploadList={false}
                          style={{ width: '100%', marginTop: '8px' }}
                        >
                          <Button icon={<UploadOutlined />} block>
                            选择图像文件
                          </Button>
                        </Upload>
                      </div>

                      <div>
                        <Text strong>编辑描述</Text>
                        <TextArea
                          value={prompt}
                          onChange={(e) => setPrompt(e.target.value)}
                          placeholder="描述您想要的修改..."
                          rows={3}
                        />
                      </div>

                      <div>
                        <Text strong>编辑强度</Text>
                        <Slider
                          value={quality}
                          onChange={setQuality}
                          min={10}
                          max={100}
                          marks={{ 10: '轻微', 50: '中等', 100: '强烈' }}
                          style={{ marginTop: '16px' }}
                        />
                      </div>
                    </Space>
                  )
                }
              ]}
            />

            <Divider />

            {/* 高级设置 */}
            <div>
              <Text strong>高级设置</Text>
              <div style={{ marginTop: '16px' }}>
                <div style={{ marginBottom: '16px' }}>
                  <Text>生成步数: {steps}</Text>
                  <Slider
                    value={steps}
                    onChange={setSteps}
                    min={10}
                    max={50}
                    style={{ marginTop: '8px' }}
                  />
                </div>

                <div style={{ marginBottom: '16px' }}>
                  <Text>引导强度: {guidance}</Text>
                  <Slider
                    value={guidance}
                    onChange={setGuidance}
                    min={1}
                    max={20}
                    step={0.5}
                    style={{ marginTop: '8px' }}
                  />
                </div>
              </div>
            </div>

            <Divider />

            {/* 生成按钮 */}
            <Button
              type="primary"
              size="large"
              block
              icon={<ThunderboltOutlined />}
              loading={loading}
              onClick={handleGenerate}
              disabled={!prompt.trim()}
            >
              {loading ? '生成中...' : '开始生成'}
            </Button>

            {loading && (
              <div style={{ marginTop: '16px' }}>
                <Progress percent={Math.round(progress)} status="active" />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  正在生成您的专属图像...
                </Text>
              </div>
            )}
          </Card>
        </Col>

        {/* 右侧结果展示 */}
        <Col xs={24} lg={16}>
          <Card 
            title="生成结果" 
            extra={
              <Space>
                <Button icon={<HistoryOutlined />}>历史记录</Button>
                <Button icon={<ReloadOutlined />}>刷新</Button>
              </Space>
            }
          >
            {generatedImages.length === 0 ? (
              <div style={{ 
                textAlign: 'center', 
                padding: '60px 20px',
                color: '#999'
              }}>
                <PictureOutlined style={{ fontSize: '64px', marginBottom: '16px' }} />
                <div>
                  <Text>还没有生成的图像</Text>
                  <br />
                  <Text type="secondary">输入描述并点击生成按钮开始创作</Text>
                </div>
              </div>
            ) : (
              <Row gutter={[16, 16]}>
                {generatedImages.map((image) => (
                  <Col xs={24} sm={12} lg={8} key={image.id}>
                    <Card
                      hoverable
                      cover={
                        <div style={{ position: 'relative' }}>
                          <Image
                            src={image.url}
                            alt={image.prompt}
                            style={{ width: '100%', height: '200px', objectFit: 'cover' }}
                            preview={{
                              mask: <EyeOutlined style={{ fontSize: '20px' }} />
                            }}
                          />
                          <div style={{
                            position: 'absolute',
                            top: '8px',
                            right: '8px',
                            background: 'rgba(0,0,0,0.6)',
                            borderRadius: '4px',
                            padding: '4px'
                          }}>
                            <Tag color={image.style === 'realistic' ? 'blue' : 'green'}>
                              {styleOptions.find(s => s.value === image.style)?.label}
                            </Tag>
                          </div>
                        </div>
                      }
                      actions={[
                        <Tooltip title={image.liked ? '取消收藏' : '收藏'}>
                          <StarOutlined
                            style={{ color: image.liked ? '#faad14' : '#999' }}
                            onClick={() => toggleLike(image.id)}
                          />
                        </Tooltip>,
                        <Tooltip title="下载">
                          <DownloadOutlined
                            onClick={() => downloadImage(image.url, image.prompt)}
                          />
                        </Tooltip>,
                        <Tooltip title="分享">
                          <ShareAltOutlined />
                        </Tooltip>,
                        <Tooltip title="复制提示词">
                          <CopyOutlined
                            onClick={() => copyPrompt(image.prompt)}
                          />
                        </Tooltip>
                      ]}
                    >
                      <Card.Meta
                        title={
                          <Text ellipsis style={{ fontSize: '14px' }}>
                            {image.prompt}
                          </Text>
                        }
                        description={
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            {image.timestamp.toLocaleString()}
                          </Text>
                        }
                      />
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ImageGeneration;