import React, { useState } from 'react';
import { Card, Row, Col, Tabs, Button, Upload, Spin, Progress } from 'antd';
import { getNotificationInstance } from '../../services/notificationService';
import { 
  BarChartOutlined, 
  LineChartOutlined, 
  PieChartOutlined,
  UploadOutlined,
  FileTextOutlined,
  DatabaseOutlined
} from '@ant-design/icons';

const { TabPane } = Tabs;

interface AnalysisResult {
  type: string;
  data: any;
  insights: string[];
  confidence: number;
}

const AIAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [analysisResults, setAnalysisResults] = useState<AnalysisResult[]>([]);
  const [activeTab, setActiveTab] = useState('data');

  const handleDataUpload = async (file: File) => {
    setLoading(true);
    try {
      // 模拟数据分析过程
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const mockResult: AnalysisResult = {
        type: 'data_analysis',
        data: {
          trends: [65, 78, 82, 91, 88, 95],
          categories: ['A', 'B', 'C', 'D', 'E', 'F'],
          summary: '数据显示整体上升趋势'
        },
        insights: [
          '数据质量良好，完整性达到95%',
          '发现3个异常值需要进一步验证',
          '趋势分析显示持续增长模式',
          '建议关注E类别的波动情况'
        ],
        confidence: 0.92
      };
      
      setAnalysisResults([mockResult]);
      getNotificationInstance().success({
        message: '数据分析完成'
      });
    } catch (error) {
      getNotificationInstance().error({
        message: '分析失败，请重试'
      });
    } finally {
      setLoading(false);
    }
  };

  const handleTrendAnalysis = async () => {
    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      const mockResult: AnalysisResult = {
        type: 'trend_analysis',
        data: {
          predictions: [98, 102, 105, 108, 112],
          timeframe: ['下周', '下月', '下季度', '下半年', '明年'],
          accuracy: 0.87
        },
        insights: [
          '预测模型显示持续增长趋势',
          '季节性因素影响较小',
          '建议在下季度加大投入',
          '风险评估：低风险'
        ],
        confidence: 0.87
      };
      
      setAnalysisResults(prev => [...prev, mockResult]);
      getNotificationInstance().success({
        message: '趋势分析完成'
      });
    } catch (error) {
      getNotificationInstance().error({
        message: '分析失败，请重试'
      });
    } finally {
      setLoading(false);
    }
  };

  const renderAnalysisResults = () => {
    if (analysisResults.length === 0) {
      return (
        <Card>
          <div className="text-center py-8 text-gray-500">
            暂无分析结果，请上传数据或开始分析
          </div>
        </Card>
      );
    }

    return analysisResults.map((result, index) => (
      <Card key={index} className="mb-4">
        <div className="mb-4">
          <h3 className="text-lg font-semibold mb-2">
            {result.type === 'data_analysis' ? '数据分析结果' : '趋势分析结果'}
          </h3>
          <Progress 
            percent={Math.round(result.confidence * 100)} 
            status="active"
            format={percent => `置信度: ${percent}%`}
          />
        </div>
        
        <div className="mb-4">
          <h4 className="font-medium mb-2">关键洞察：</h4>
          <ul className="list-disc list-inside space-y-1">
            {result.insights.map((insight, idx) => (
              <li key={idx} className="text-gray-700">{insight}</li>
            ))}
          </ul>
        </div>
      </Card>
    ));
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">AI智能分析</h1>
        <p className="text-gray-600">
          利用先进的AI算法对数据进行深度分析，发现隐藏的模式和趋势
        </p>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab={
          <span>
            <DatabaseOutlined />
            数据分析
          </span>
        } key="data">
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={8}>
              <Card title="数据上传" className="h-full">
                <Upload.Dragger
                  accept=".csv,.xlsx,.json"
                  beforeUpload={(file) => {
                    handleDataUpload(file);
                    return false;
                  }}
                  disabled={loading}
                >
                  <p className="ant-upload-drag-icon">
                    <UploadOutlined />
                  </p>
                  <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
                  <p className="ant-upload-hint">
                    支持 CSV、Excel、JSON 格式
                  </p>
                </Upload.Dragger>
              </Card>
            </Col>
            
            <Col xs={24} lg={16}>
              <Spin spinning={loading}>
                {renderAnalysisResults()}
              </Spin>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab={
          <span>
            <LineChartOutlined />
            趋势分析
          </span>
        } key="trends">
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={8}>
              <Card title="趋势分析工具" className="h-full">
                <div className="space-y-4">
                  <Button 
                    type="primary" 
                    icon={<LineChartOutlined />}
                    onClick={handleTrendAnalysis}
                    loading={loading}
                    block
                  >
                    开始趋势分析
                  </Button>
                  
                  <div className="text-sm text-gray-600">
                    <p>• 时间序列分析</p>
                    <p>• 季节性检测</p>
                    <p>• 异常值识别</p>
                    <p>• 预测建模</p>
                  </div>
                </div>
              </Card>
            </Col>
            
            <Col xs={24} lg={16}>
              <Spin spinning={loading}>
                {renderAnalysisResults()}
              </Spin>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab={
          <span>
            <BarChartOutlined />
            报告生成
          </span>
        } key="reports">
          <Card>
            <div className="text-center py-8">
              <FileTextOutlined className="text-4xl text-gray-400 mb-4" />
              <h3 className="text-lg font-medium mb-2">智能报告生成</h3>
              <p className="text-gray-600 mb-4">
                基于分析结果自动生成专业的分析报告
              </p>
              <Button type="primary" disabled={analysisResults.length === 0}>
                生成报告
              </Button>
            </div>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default AIAnalysis;