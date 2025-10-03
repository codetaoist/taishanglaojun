import React, { useState } from 'react';
import { 
  Card, 
  Button, 
  Input, 
  Typography, 
  Space, 
  message,
  Divider,
  Tag,
  Alert
} from 'antd';
import { 
  RobotOutlined, 
  QuestionCircleOutlined,
  BulbOutlined,
  ReloadOutlined,
  SendOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { TextArea } = Input;
const { Paragraph } = Typography;

interface WisdomInterpretationProps {
  wisdomId: string;
  wisdomTitle: string;
  wisdomContent: string;
}

interface InterpretationResponse {
  wisdom_id: string;
  title: string;
  content: string;
  interpretation: string;
}

interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
}

const WisdomInterpretation: React.FC<WisdomInterpretationProps> = ({
  wisdomId,
  wisdomTitle,
  wisdomContent
}) => {
  const [loading, setLoading] = useState(false);
  const [interpretation, setInterpretation] = useState<string>('');
  const [question, setQuestion] = useState<string>('');
  const [hasInterpretation, setHasInterpretation] = useState(false);

  // 获取AI智慧解读
  const handleGetInterpretation = async (customQuestion?: string) => {
    if (!wisdomId) {
      message.error('智慧ID不能为空');
      return;
    }

    setLoading(true);
    try {
      const requestBody = customQuestion ? { question: customQuestion } : {};
      const response = await apiClient.post(`/cultural-wisdom/ai/${wisdomId}/interpret`, requestBody);
      
      if (response.data.code === 'SUCCESS') {
        const data: InterpretationResponse = response.data.data;
        setInterpretation(data.interpretation);
        setHasInterpretation(true);
        message.success('AI解读获取成功');
        
        // 清空问题输入框
        if (customQuestion) {
          setQuestion('');
        }
      } else {
        message.error(response.data.message || 'AI解读获取失败');
      }
    } catch (error) {
       const apiError = error as ApiError;
       console.error('AI解读请求失败:', error);
       message.error(apiError.response?.data?.message || 'AI解读请求失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 提交自定义问题
  const handleSubmitQuestion = () => {
    if (!question.trim()) {
      message.warning('请输入您的问题');
      return;
    }
    handleGetInterpretation(question.trim());
  };

  // 重新获取解读
  const handleRefresh = () => {
    setInterpretation('');
    setHasInterpretation(false);
    handleGetInterpretation();
  };

  return (
    <Card 
       title={
         <Space>
           <RobotOutlined className="text-cultural-gold" />
           <span>AI智慧解读</span>
         </Space>
       }
      extra={
        hasInterpretation && (
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleRefresh}
            disabled={loading}
            size="small"
          >
            重新解读
          </Button>
        )
      }
      className="shadow-lg"
    >
      <div className="space-y-4">
        {/* 智慧信息展示 */}
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 p-4 rounded-lg border border-blue-200">
          <div className="flex items-start space-x-3">
            <BulbOutlined className="text-blue-500 text-lg mt-1" />
            <div className="flex-1">
              <span className="font-semibold text-blue-800">{wisdomTitle}</span>
              <Paragraph className="text-blue-700 mt-2 mb-0 text-sm leading-relaxed">
                {wisdomContent.length > 200 ? `${wisdomContent.substring(0, 200)}...` : wisdomContent}
              </Paragraph>
            </div>
          </div>
        </div>

        {/* AI解读结果 */}
        {hasInterpretation && interpretation && (
          <div className="bg-gradient-to-r from-green-50 to-emerald-50 p-4 rounded-lg border border-green-200">
            <div className="flex items-start space-x-3">
              <RobotOutlined className="text-green-500 text-lg mt-1" />
              <div className="flex-1">
                <span className="font-semibold text-green-800">AI智慧解读</span>
                <Paragraph className="text-green-700 mt-2 mb-0 leading-relaxed whitespace-pre-wrap">
                  {interpretation}
                </Paragraph>
              </div>
            </div>
          </div>
        )}

        {/* 操作区域 */}
        <div className="space-y-3">
          {!hasInterpretation && (
            <Button 
              type="primary" 
              icon={<RobotOutlined />}
              onClick={() => handleGetInterpretation()}
              loading={loading}
              size="large"
              block
              className="bg-gradient-to-r from-blue-500 to-indigo-500 border-0 shadow-lg hover:shadow-xl transition-all duration-300"
            >
              {loading ? '正在生成AI解读...' : '获取AI智慧解读'}
            </Button>
          )}

          <Divider>
             <span className="text-gray-500 text-sm">或者提出您的具体问题</span>
           </Divider>

          {/* 自定义问题输入 */}
          <div className="space-y-2">
            <TextArea
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="请输入您想了解的具体问题，例如：这句话在现代生活中有什么指导意义？"
              rows={3}
              maxLength={500}
              showCount
              disabled={loading}
            />
            <Button 
              type="primary"
              icon={<SendOutlined />}
              onClick={handleSubmitQuestion}
              loading={loading}
              disabled={!question.trim()}
              className="bg-cultural-gold border-cultural-gold hover:bg-yellow-600 hover:border-yellow-600"
            >
              提问AI助手
            </Button>
          </div>
        </div>

        {/* 使用提示 */}
        <Alert
          message="AI解读说明"
          description={
            <div className="text-sm space-y-1">
              <div>• AI将从多个角度为您解读这段智慧的深层含义</div>
              <div>• 您也可以提出具体问题，获得针对性的解答</div>
              <div>• AI解读仅供参考，建议结合个人理解和实践</div>
            </div>
          }
          type="info"
          showIcon
          icon={<QuestionCircleOutlined />}
          className="border-blue-200 bg-blue-50"
        />

        {/* 推荐标签 */}
        <div>
          <span className="text-gray-500 text-sm mb-2 block">常见问题示例：</span>
          <div className="flex flex-wrap gap-2">
            {[
              '这句话的含义是什么？',
              '如何生活在应用？',
              '有什么现代启示？',
              '相关的历史背景？'
            ].map((tag, index) => (
              <Tag 
                key={index}
                className="cursor-pointer hover:bg-blue-100 border-blue-300 text-blue-600"
                onClick={() => setQuestion(tag)}
              >
                {tag}
              </Tag>
            ))}
          </div>
        </div>
      </div>
    </Card>
  );
};

export default WisdomInterpretation;