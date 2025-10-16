import React, { useState, useRef, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Upload, 
  Typography, 
  Space, 
  Image, 
  Spin, 
  message, 
  Tabs, 
  Tag, 
  Progress,
  List,
  Descriptions,
  Alert,
  Tooltip,
  Modal,
  Table,
  Badge,
  Divider
} from 'antd';
import { 
  UploadOutlined, 
  EyeOutlined, 
  DownloadOutlined, 
  ShareAltOutlined,
  ScanOutlined,
  BulbOutlined,
  BarChartOutlined,
  TagsOutlined,
  CameraOutlined,
  FileImageOutlined,
  SearchOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { aiService, type ImageAnalysisRequest, type ImageAnalysisResponse } from '../../services/aiService';

const { Title, Paragraph, Text } = Typography;
const { Dragger } = Upload;

interface AnalysisResult {
  id: string;
  imageUrl: string;
  fileName: string;
  fileSize: string;
  uploadTime: Date;
  status: 'analyzing' | 'completed' | 'failed';
  progress: number;
  results: {
    objects?: Array<{ name: string; confidence: number; bbox: number[] }>;
    faces?: Array<{ age: number; gender: string; emotion: string; confidence: number }>;
    text?: Array<{ text: string; confidence: number; bbox: number[] }>;
    colors?: Array<{ color: string; percentage: number; hex: string }>;
    tags?: Array<{ tag: string; confidence: number }>;
    description?: string;
    similarity?: Array<{ imageUrl: string; similarity: number; description: string }>;
  };
}

const ImageAnalysis: React.FC = () => {
  const [activeTab, setActiveTab] = useState('upload');
  const [analysisResults, setAnalysisResults] = useState<AnalysisResult[]>([]);
  const [selectedResult, setSelectedResult] = useState<AnalysisResult | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const fileInputRef = useRef<any>(null);

  // 分析类型选项
  const analysisTypes = [
    { key: 'objects', label: '物体识别', icon: <ScanOutlined />, description: '识别图像中的物体和场景' },
    { key: 'faces', label: '人脸分析', icon: <EyeOutlined />, description: '分析人脸特征、年龄、情绪等' },
    { key: 'text', label: '文字识别', icon: <FileImageOutlined />, description: '提取图像中的文字内容' },
    { key: 'colors', label: '色彩分析', icon: <BulbOutlined />, description: '分析图像的主要色彩构成' },
    { key: 'tags', label: '智能标签', icon: <TagsOutlined />, description: '生成描述性标签' },
    { key: 'similarity', label: '相似搜索', icon: <SearchOutlined />, description: '查找相似的图像' }
  ];

  // 加载分析历史记录
  useEffect(() => {
    const loadAnalysisHistory = async () => {
      try {
        const response = await aiService.getAnalysisHistory();
        if (response.code === 'SUCCESS' && response.data?.history) {
          // 转换API数据格式为组件所需格式
          const historyResults: AnalysisResult[] = response.data.history.map((item: any) => ({
            id: item.id,
            imageUrl: item.imageUrl,
            fileName: item.fileName,
            fileSize: item.fileSize,
            uploadTime: new Date(item.uploadTime),
            status: 'completed' as const,
            progress: 100,
            results: item.results || {}
          }));
          setAnalysisResults(historyResults);
        }
      } catch (error) {
        console.error('加载分析历史失败:', error);
        // 如果API失败，可以选择显示一些模拟的历史数据
      }
    };

    loadAnalysisHistory();
  }, []);

  // 处理文件上传
  const handleFileUpload = async (file: File) => {
    const fileSize = (file.size / 1024 / 1024).toFixed(2) + ' MB';
    const imageUrl = URL.createObjectURL(file);
    
    const newResult: AnalysisResult = {
      id: Date.now().toString(),
      imageUrl,
      fileName: file.name,
      fileSize,
      uploadTime: new Date(),
      status: 'analyzing',
      progress: 0,
      results: {}
    };

    setAnalysisResults(prev => [newResult, ...prev]);
    
    // 执行图像分析
    await performImageAnalysis(newResult.id, file);
    
    return false; // 阻止默认上传行为
  };

  // 执行图像分析
  const performImageAnalysis = async (resultId: string, file: File) => {
    const updateProgress = (progress: number, status?: 'analyzing' | 'completed' | 'failed') => {
      setAnalysisResults(prev => 
        prev.map(result => 
          result.id === resultId 
            ? { ...result, progress, status: status || result.status }
            : result
        )
      );
    };

    // 模拟分析进度
    const progressInterval = setInterval(() => {
      updateProgress(Math.min(90, Math.random() * 80 + 10));
    }, 500);

    try {
      const request: ImageAnalysisRequest = {
        imageFile: file,
        analysisTypes: ['objects', 'faces', 'text', 'colors', 'tags', 'similarity']
      };

      const response = await aiService.analyzeImage(request);
      
      if (response.code === 'SUCCESS') {
        const apiResult = response.data;
        
        setAnalysisResults(prev => 
          prev.map(result => 
            result.id === resultId 
              ? { 
                  ...result, 
                  status: 'completed', 
                  progress: 100,
                  results: apiResult.results 
                }
              : result
          )
        );
        
        message.success('图像分析完成！');
      } else {
        throw new Error(response.message || '分析失败');
      }
    } catch (error: any) {
      console.error('图像分析失败:', error);
      
      // 如果API失败，使用模拟数据作为后备
      const mockResults = {
        objects: [
          { name: '汽车', confidence: 0.95, bbox: [100, 100, 200, 200] },
          { name: '建筑物', confidence: 0.88, bbox: [50, 50, 300, 400] },
          { name: '树木', confidence: 0.76, bbox: [250, 80, 350, 300] }
        ],
        faces: [
          { age: 25, gender: '女性', emotion: '微笑', confidence: 0.92 },
          { age: 30, gender: '男性', emotion: '中性', confidence: 0.87 }
        ],
        text: [
          { text: 'WELCOME', confidence: 0.98, bbox: [120, 50, 280, 80] },
          { text: '欢迎光临', confidence: 0.94, bbox: [120, 90, 280, 120] }
        ],
        colors: [
          { color: '蓝色', percentage: 35, hex: '#4A90E2' },
          { color: '白色', percentage: 28, hex: '#FFFFFF' },
          { color: '绿色', percentage: 20, hex: '#7ED321' },
          { color: '灰色', percentage: 17, hex: '#9B9B9B' }
        ],
        tags: [
          { tag: '城市风景', confidence: 0.91 },
          { tag: '现代建筑', confidence: 0.85 },
          { tag: '交通工具', confidence: 0.79 },
          { tag: '户外场景', confidence: 0.73 }
        ],
        description: '这是一张现代城市街景图片，包含了高楼大厦、汽车和行人。图片色彩丰富，构图均衡，展现了繁华的都市生活场景。',
        similarity: [
          { imageUrl: 'https://picsum.photos/200/200?random=1', similarity: 0.89, description: '相似的城市街景' },
          { imageUrl: 'https://picsum.photos/200/200?random=2', similarity: 0.76, description: '类似的建筑风格' },
          { imageUrl: 'https://picsum.photos/200/200?random=3', similarity: 0.68, description: '相近的色彩搭配' }
        ]
      };

      setAnalysisResults(prev => 
        prev.map(result => 
          result.id === resultId 
            ? { ...result, status: 'completed', progress: 100, results: mockResults }
            : result
        )
      );

      message.warning('API调用失败，使用模拟数据作为示例');
    } finally {
      clearInterval(progressInterval);
    }
  };

  // 查看详细结果
  const viewDetails = (result: AnalysisResult) => {
    setSelectedResult(result);
    setModalVisible(true);
  };

  // 渲染分析结果卡片
  const renderResultCard = (result: AnalysisResult) => {
    const getStatusIcon = () => {
      switch (result.status) {
        case 'analyzing':
          return <Spin size="small" />;
        case 'completed':
          return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
        case 'failed':
          return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
        default:
          return null;
      }
    };

    return (
      <Card
        key={result.id}
        hoverable
        style={{ marginBottom: '16px' }}
        cover={
          <div style={{ position: 'relative', height: '200px', overflow: 'hidden' }}>
            <Image
              src={result.imageUrl}
              alt={result.fileName}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              preview={false}
            />
            {result.status === 'analyzing' && (
              <div style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                background: 'rgba(0,0,0,0.7)',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                color: 'white'
              }}>
                <Spin size="large" />
                <div style={{ marginTop: '16px' }}>
                  <Progress 
                    percent={result.progress} 
                    strokeColor="#1890ff"
                    trailColor="rgba(255,255,255,0.3)"
                  />
                  <Text style={{ color: 'white', fontSize: '12px' }}>
                    分析中... {result.progress}%
                  </Text>
                </div>
              </div>
            )}
          </div>
        }
        actions={[
          <Tooltip title="查看详情">
            <EyeOutlined onClick={() => viewDetails(result)} />
          </Tooltip>,
          <Tooltip title="下载结果">
            <DownloadOutlined />
          </Tooltip>,
          <Tooltip title="分享">
            <ShareAltOutlined />
          </Tooltip>
        ]}
      >
        <Card.Meta
          title={
            <Space>
              <Text ellipsis style={{ maxWidth: '200px' }}>
                {result.fileName}
              </Text>
              {getStatusIcon()}
            </Space>
          }
          description={
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                {result.fileSize} • {result.uploadTime.toLocaleString()}
              </Text>
              {result.status === 'completed' && result.results.tags && (
                <div>
                  {result.results.tags.slice(0, 3).map((tag, index) => (
                    <Tag key={index} size="small" color="blue">
                      {tag.tag}
                    </Tag>
                  ))}
                </div>
              )}
            </Space>
          }
        />
      </Card>
    );
  };

  // 渲染详细结果模态框
  const renderDetailsModal = () => {
    if (!selectedResult) return null;

    const { results } = selectedResult;

    return (
      <Modal
        title="分析结果详情"
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        width={800}
        footer={[
          <Button key="close" onClick={() => setModalVisible(false)}>
            关闭
          </Button>,
          <Button key="download" type="primary" icon={<DownloadOutlined />}>
            下载报告
          </Button>
        ]}
      >
        <Row gutter={[16, 16]}>
          <Col span={10}>
            <Image
              src={selectedResult.imageUrl}
              alt={selectedResult.fileName}
              style={{ width: '100%' }}
            />
          </Col>
          <Col span={14}>
            <Tabs 
              defaultActiveKey="objects"
              items={[
                ...(results.objects ? [{
                  key: 'objects',
                  label: '物体识别',
                  children: (
                    <List
                      size="small"
                      dataSource={results.objects}
                      renderItem={(item) => (
                        <List.Item>
                          <Space>
                            <Badge color="blue" />
                            <Text>{item.name}</Text>
                            <Tag color="green">{(item.confidence * 100).toFixed(1)}%</Tag>
                          </Space>
                        </List.Item>
                      )}
                    />
                  )
                }] : []),
                ...(results.faces ? [{
                  key: 'faces',
                  label: '人脸分析',
                  children: (
                    <>
                      {results.faces.map((face, index) => (
                        <Descriptions key={index} size="small" column={1} style={{ marginBottom: '16px' }}>
                          <Descriptions.Item label="年龄">{face.age}岁</Descriptions.Item>
                          <Descriptions.Item label="性别">{face.gender}</Descriptions.Item>
                          <Descriptions.Item label="情绪">{face.emotion}</Descriptions.Item>
                          <Descriptions.Item label="置信度">{(face.confidence * 100).toFixed(1)}%</Descriptions.Item>
                        </Descriptions>
                      ))}
                    </>
                  )
                }] : []),
                ...(results.text ? [{
                  key: 'text',
                  label: '文字识别',
                  children: (
                    <List
                      size="small"
                      dataSource={results.text}
                      renderItem={(item) => (
                        <List.Item>
                          <Space>
                            <Text code>{item.text}</Text>
                            <Tag color="orange">{(item.confidence * 100).toFixed(1)}%</Tag>
                          </Space>
                        </List.Item>
                      )}
                    />
                  )
                }] : []),
                ...(results.colors ? [{
                  key: 'colors',
                  label: '色彩分析',
                  children: (
                    <List
                      size="small"
                      dataSource={results.colors}
                      renderItem={(item) => (
                        <List.Item>
                          <Space>
                            <div
                              style={{
                                width: '20px',
                                height: '20px',
                                backgroundColor: item.hex,
                                borderRadius: '4px',
                                border: '1px solid #d9d9d9'
                              }}
                            />
                            <Text>{item.color}</Text>
                            <Tag>{item.percentage}%</Tag>
                          </Space>
                        </List.Item>
                      )}
                    />
                  )
                }] : []),
                ...(results.tags ? [{
                  key: 'tags',
                  label: '智能标签',
                  children: (
                    <Space wrap>
                      {results.tags.map((tag, index) => (
                        <Tag key={index} color="blue">
                          {tag.tag} ({(tag.confidence * 100).toFixed(1)}%)
                        </Tag>
                      ))}
                    </Space>
                  )
                }] : [])
              ]}
            />

            {results.description && (
              <div style={{ marginTop: '16px' }}>
                <Title level={5}>图像描述</Title>
                <Paragraph>{results.description}</Paragraph>
              </div>
            )}
          </Col>
        </Row>
      </Modal>
    );
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <ScanOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          AI图像分析
        </Title>
        <Paragraph>
          运用先进的计算机视觉技术，深度解析图像内容，提供全面的智能分析报告
        </Paragraph>
      </div>

      <Row gutter={[24, 24]}>
        {/* 左侧上传区域 */}
        <Col xs={24} lg={8}>
          <Card title="上传图像" extra={<CameraOutlined />}>
            <Dragger
              accept="image/*"
              beforeUpload={handleFileUpload}
              showUploadList={false}
              style={{ marginBottom: '16px' }}
            >
              <p className="ant-upload-drag-icon">
                <UploadOutlined style={{ fontSize: '48px', color: '#1890ff' }} />
              </p>
              <p className="ant-upload-text">点击或拖拽图像到此区域</p>
              <p className="ant-upload-hint">
                支持 JPG、PNG、GIF 等格式，单个文件不超过 10MB
              </p>
            </Dragger>

            <Alert
              message="分析功能"
              description="我们的AI可以识别物体、分析人脸、提取文字、解析色彩等，为您提供全面的图像洞察。"
              type="info"
              showIcon
              style={{ marginBottom: '16px' }}
            />

            {/* 分析类型说明 */}
            <div>
              <Title level={5}>支持的分析类型</Title>
              <List
                size="small"
                dataSource={analysisTypes}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={item.icon}
                      title={item.label}
                      description={item.description}
                    />
                  </List.Item>
                )}
              />
            </div>
          </Card>
        </Col>

        {/* 右侧结果展示 */}
        <Col xs={24} lg={16}>
          <Card 
            title="分析结果" 
            extra={
              <Space>
                <Text type="secondary">
                  共 {analysisResults.length} 个结果
                </Text>
                <Button icon={<BarChartOutlined />}>统计报告</Button>
              </Space>
            }
          >
            {analysisResults.length === 0 ? (
              <div style={{ 
                textAlign: 'center', 
                padding: '60px 20px',
                color: '#999'
              }}>
                <FileImageOutlined style={{ fontSize: '64px', marginBottom: '16px' }} />
                <div>
                  <Text>还没有分析结果</Text>
                  <br />
                  <Text type="secondary">上传图像开始智能分析</Text>
                </div>
              </div>
            ) : (
              <Row gutter={[16, 16]}>
                {analysisResults.map((result) => (
                  <Col xs={24} sm={12} lg={8} key={result.id}>
                    {renderResultCard(result)}
                  </Col>
                ))}
              </Row>
            )}
          </Card>
        </Col>
      </Row>

      {/* 详细结果模态框 */}
      {renderDetailsModal()}
    </div>
  );
};

export default ImageAnalysis;