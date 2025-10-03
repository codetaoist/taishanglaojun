import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Typography, 
  Space, 
  Spin, 
  message,
  Row,
  Col,
  Statistic,
  List,
  Tag,
  Empty
} from 'antd';
import { 
  RobotOutlined, 
  BulbOutlined,
  BookOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  BarChartOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { Title, Paragraph } = Typography;

interface WisdomAnalysisProps {
  wisdomId: string;
}

interface AnalysisData {
  analysis_summary: {
    total_key_points: number;
    total_recommendations: number;
    total_related_concepts: number;
    has_practical_advice: boolean;
  };
  key_points: string[];
  modern_relevance: string;
  recommendations: string[];
  related_concepts: string[];
  practical_applications: string[];
}

interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
}

const WisdomAnalysis: React.FC<WisdomAnalysisProps> = ({ wisdomId }) => {
  const [loading, setLoading] = useState(false);
  const [analysis, setAnalysis] = useState<AnalysisData | null>(null);
  const [hasAnalysis, setHasAnalysis] = useState(false);

  // 获取AI智慧分析
  const loadAnalysis = async () => {
    if (!wisdomId) {
      message.error('智慧ID不能为空');
      return;
    }

    setLoading(true);
    try {
      const response = await apiClient.get(`/cultural-wisdom/ai/${wisdomId}/analysis`);
      
      if (response.data.code === 'SUCCESS') {
        const data: AnalysisData = response.data.data;
        setAnalysis(data);
        setHasAnalysis(true);
        message.success('AI分析获取成功');
      } else {
        message.error(response.data.message || 'AI分析获取失败');
      }
    } catch (error) {
      const apiError = error as ApiError;
      console.error('AI分析请求失败:', error);
      message.error(apiError.response?.data?.message || 'AI分析请求失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 重新获取分析
  const handleRefresh = () => {
    setAnalysis(null);
    setHasAnalysis(false);
    loadAnalysis();
  };

  // 组件挂载时自动加载分析
  useEffect(() => {
    if (wisdomId) {
      loadAnalysis();
    }
  }, [wisdomId]);

  return (
    <Card 
      title={
        <Space>
          <BarChartOutlined className="text-cultural-gold" />
          <span>AI深度分析</span>
        </Space>
      }
      extra={
        hasAnalysis && (
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleRefresh}
            disabled={loading}
            size="small"
          >
            重新分析
          </Button>
        )
      }
      className="shadow-lg"
    >
      <Spin spinning={loading}>
        {!hasAnalysis && !loading && (
          <div className="text-center py-8">
            <BarChartOutlined className="text-6xl text-gray-300 mb-4" />
            <Button 
              type="primary" 
              icon={<BarChartOutlined />}
              onClick={loadAnalysis}
              size="large"
              className="bg-gradient-to-r from-purple-500 to-pink-500 border-0 shadow-lg hover:shadow-xl transition-all duration-300"
            >
              开始AI深度分析
            </Button>
            <div className="mt-3 text-gray-500 text-sm">
              AI将从多个维度深度分析这段智慧
            </div>
          </div>
        )}

        {hasAnalysis && analysis && (
          <div className="space-y-6">
            {/* 分析概览 */}
            <Card size="small" className="bg-gradient-to-r from-purple-50 to-pink-50 border-purple-200">
              <Title level={5} className="text-purple-800 mb-4">
                <BarChartOutlined className="mr-2" />
                分析概览
              </Title>
              <Row gutter={[16, 16]}>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="关键要点"
                    value={analysis.analysis_summary.total_key_points}
                    suffix="个"
                    valueStyle={{ color: '#722ed1' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="实践建议"
                    value={analysis.analysis_summary.total_recommendations}
                    suffix="条"
                    valueStyle={{ color: '#eb2f96' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <Statistic
                    title="相关概念"
                    value={analysis.analysis_summary.total_related_concepts}
                    suffix="个"
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Col>
                <Col xs={12} sm={6}>
                  <div className="text-center">
                    <div className="text-sm text-gray-500 mb-1">实用性</div>
                    <div className="flex items-center justify-center">
                      {analysis.analysis_summary.has_practical_advice ? (
                        <Tag color="success" icon={<CheckCircleOutlined />}>
                          高实用性
                        </Tag>
                      ) : (
                        <Tag color="warning" icon={<ExclamationCircleOutlined />}>
                          理论性
                        </Tag>
                      )}
                    </div>
                  </div>
                </Col>
              </Row>
            </Card>

            {/* 关键要点 */}
            {analysis.key_points && analysis.key_points.length > 0 && (
              <Card size="small" title={
                <Space>
                  <BulbOutlined className="text-yellow-500" />
                  <span>关键要点</span>
                </Space>
              }>
                <List
                  dataSource={analysis.key_points}
                  renderItem={(item, index) => (
                    <List.Item className="border-0 py-2">
                      <div className="flex items-start space-x-3 w-full">
                        <div className="w-6 h-6 bg-yellow-100 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                          <span className="text-yellow-600 text-xs font-bold">{index + 1}</span>
                        </div>
                        <span className="flex-1 leading-relaxed">{item}</span>
                      </div>
                    </List.Item>
                  )}
                />
              </Card>
            )}

            {/* 现代意义 */}
            {analysis.modern_relevance && (
              <Card size="small" title={
                <Space>
                  <BookOutlined className="text-blue-500" />
                  <span>现代意义</span>
                </Space>
              }>
                <Paragraph className="text-gray-700 leading-relaxed whitespace-pre-wrap mb-0">
                  {analysis.modern_relevance}
                </Paragraph>
              </Card>
            )}

            {/* 实践建议 */}
            {analysis.recommendations && analysis.recommendations.length > 0 && (
              <Card size="small" title={
                <Space>
                  <CheckCircleOutlined className="text-green-500" />
                  <span>实践建议</span>
                </Space>
              }>
                <List
                  dataSource={analysis.recommendations}
                  renderItem={(item) => (
                    <List.Item className="border-0 py-2">
                      <div className="flex items-start space-x-3 w-full">
                        <CheckCircleOutlined className="text-green-500 mt-1 flex-shrink-0" />
                        <span className="flex-1 leading-relaxed">{item}</span>
                      </div>
                    </List.Item>
                  )}
                />
              </Card>
            )}

            {/* 相关概念 */}
            {analysis.related_concepts && analysis.related_concepts.length > 0 && (
              <Card size="small" title={
                <Space>
                  <BookOutlined className="text-indigo-500" />
                  <span>相关概念</span>
                </Space>
              }>
                <div className="flex flex-wrap gap-2">
                  {analysis.related_concepts.map((concept, index) => (
                    <Tag 
                      key={index} 
                      color="processing" 
                      className="cursor-pointer hover:bg-blue-100"
                    >
                      {concept}
                    </Tag>
                  ))}
                </div>
              </Card>
            )}

            {/* 实际应用 */}
            {analysis.practical_applications && analysis.practical_applications.length > 0 && (
              <Card size="small" title={
                <Space>
                  <RobotOutlined className="text-orange-500" />
                  <span>实际应用</span>
                </Space>
              }>
                <List
                  dataSource={analysis.practical_applications}
                  renderItem={(item, index) => (
                    <List.Item className="border-0 py-2">
                      <div className="flex items-start space-x-3 w-full">
                        <div className="w-6 h-6 bg-orange-100 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                          <span className="text-orange-600 text-xs font-bold">{index + 1}</span>
                        </div>
                        <span className="flex-1 leading-relaxed">{item}</span>
                      </div>
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </div>
        )}

        {hasAnalysis && !analysis && (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="分析数据为空"
            className="py-8"
          >
            <Button 
              type="primary" 
              onClick={handleRefresh}
              icon={<ReloadOutlined />}
            >
              重新分析
            </Button>
          </Empty>
        )}
      </Spin>
    </Card>
  );
};

export default WisdomAnalysis;