import React, { useState } from 'react';
import { Card, Row, Col, Tabs, Button, Input, Select, message, Spin, Typography } from 'antd';
import { 
  EditOutlined, 
  CodeOutlined, 
  PictureOutlined,
  FileTextOutlined,
  BulbOutlined,
  DownloadOutlined
} from '@ant-design/icons';

const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;
const { Paragraph } = Typography;

interface GenerationResult {
  type: string;
  content: string;
  metadata?: any;
}

const AIGeneration: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('text');
  const [textPrompt, setTextPrompt] = useState('');
  const [codePrompt, setCodePrompt] = useState('');
  const [designPrompt, setDesignPrompt] = useState('');
  const [language, setLanguage] = useState('javascript');
  const [textType, setTextType] = useState('article');
  const [results, setResults] = useState<GenerationResult[]>([]);

  const generateText = async () => {
    if (!textPrompt.trim()) {
      message.warning('请输入文本生成提示');
      return;
    }

    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const mockContent = `基于您的提示"${textPrompt}"，我为您生成了以下内容：

这是一个关于${textPrompt}的详细分析。在当今快速发展的技术环境中，我们需要深入理解这个主题的各个方面。

首先，让我们从基础概念开始。${textPrompt}不仅仅是一个简单的概念，它涉及到多个层面的理解和应用。通过深入分析，我们可以发现其中蕴含的深层价值。

其次，实际应用方面也值得我们关注。在实践中，${textPrompt}的应用场景非常广泛，从基础的日常使用到高级的专业应用，都有其独特的价值。

最后，展望未来，${textPrompt}的发展前景十分广阔。随着技术的不断进步，我们有理由相信它将在更多领域发挥重要作用。

总结而言，${textPrompt}是一个值得深入研究和应用的重要主题，它将为我们带来更多的机遇和可能性。`;

      const result: GenerationResult = {
        type: 'text',
        content: mockContent,
        metadata: {
          wordCount: mockContent.length,
          type: textType,
          prompt: textPrompt
        }
      };

      setResults(prev => [result, ...prev]);
      message.success('文本生成完成');
    } catch (error) {
      message.error('生成失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const generateCode = async () => {
    if (!codePrompt.trim()) {
      message.warning('请输入代码生成提示');
      return;
    }

    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      let mockCode = '';
      if (language === 'javascript') {
        mockCode = `// ${codePrompt}
function ${codePrompt.replace(/\s+/g, '')}() {
  // 实现${codePrompt}的功能
  const result = {
    success: true,
    data: null,
    message: '${codePrompt}执行成功'
  };
  
  try {
    // 核心逻辑实现
    console.log('开始执行${codePrompt}');
    
    // 处理业务逻辑
    result.data = processData();
    
    return result;
  } catch (error) {
    console.error('执行失败:', error);
    return {
      success: false,
      error: error.message
    };
  }
}

function processData() {
  // 数据处理逻辑
  return {
    timestamp: new Date().toISOString(),
    processed: true
  };
}

// 使用示例
const result = ${codePrompt.replace(/\s+/g, '')}();
console.log(result);`;
      } else if (language === 'python') {
        mockCode = `# ${codePrompt}
def ${codePrompt.replace(/\s+/g, '_').toLowerCase()}():
    """
    实现${codePrompt}的功能
    """
    try:
        print(f"开始执行${codePrompt}")
        
        # 核心逻辑实现
        result = process_data()
        
        return {
            'success': True,
            'data': result,
            'message': '${codePrompt}执行成功'
        }
    except Exception as e:
        print(f"执行失败: {e}")
        return {
            'success': False,
            'error': str(e)
        }

def process_data():
    """数据处理逻辑"""
    import datetime
    return {
        'timestamp': datetime.datetime.now().isoformat(),
        'processed': True
    }

# 使用示例
if __name__ == "__main__":
    result = ${codePrompt.replace(/\s+/g, '_').toLowerCase()}()
    print(result)`;
      }

      const result: GenerationResult = {
        type: 'code',
        content: mockCode,
        metadata: {
          language,
          prompt: codePrompt,
          lines: mockCode.split('\n').length
        }
      };

      setResults(prev => [result, ...prev]);
      message.success('代码生成完成');
    } catch (error) {
      message.error('生成失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const generateDesign = async () => {
    if (!designPrompt.trim()) {
      message.warning('请输入设计生成提示');
      return;
    }

    setLoading(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const mockDesign = `设计方案：${designPrompt}

设计理念：
基于现代简约风格，注重用户体验和视觉美感的统一。

色彩搭配：
- 主色调：#1890ff (科技蓝)
- 辅助色：#52c41a (活力绿)
- 背景色：#f5f5f5 (浅灰)
- 文字色：#262626 (深灰)

布局结构：
1. 顶部导航区域 - 简洁明了的导航菜单
2. 主内容区域 - 核心功能展示
3. 侧边栏区域 - 辅助功能和快捷操作
4. 底部信息区域 - 版权和链接信息

交互设计：
- 响应式设计，适配多种设备
- 流畅的动画过渡效果
- 直观的操作反馈
- 无障碍访问支持

技术实现：
- React + TypeScript
- Ant Design 组件库
- CSS-in-JS 样式方案
- 移动端优先的响应式设计`;

      const result: GenerationResult = {
        type: 'design',
        content: mockDesign,
        metadata: {
          prompt: designPrompt,
          style: 'modern',
          components: ['navigation', 'content', 'sidebar', 'footer']
        }
      };

      setResults(prev => [result, ...prev]);
      message.success('设计方案生成完成');
    } catch (error) {
      message.error('生成失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const downloadResult = (result: GenerationResult, index: number) => {
    const element = document.createElement('a');
    const file = new Blob([result.content], { type: 'text/plain' });
    element.href = URL.createObjectURL(file);
    element.download = `ai_generated_${result.type}_${index + 1}.txt`;
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
    message.success('文件下载成功');
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">AI智能生成</h1>
        <p className="text-gray-600">
          利用先进的AI技术生成高质量的文本、代码和设计方案
        </p>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card>
            <Tabs activeKey={activeTab} onChange={setActiveTab}>
              <TabPane tab={
                <span>
                  <EditOutlined />
                  文本生成
                </span>
              } key="text">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">内容类型</label>
                    <Select
                      value={textType}
                      onChange={setTextType}
                      className="w-full"
                    >
                      <Option value="article">文章</Option>
                      <Option value="summary">摘要</Option>
                      <Option value="email">邮件</Option>
                      <Option value="report">报告</Option>
                    </Select>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium mb-2">生成提示</label>
                    <TextArea
                      value={textPrompt}
                      onChange={(e) => setTextPrompt(e.target.value)}
                      placeholder="请描述您想要生成的内容..."
                      rows={4}
                    />
                  </div>
                  
                  <Button
                    type="primary"
                    icon={<BulbOutlined />}
                    onClick={generateText}
                    loading={loading}
                    block
                  >
                    生成文本
                  </Button>
                </div>
              </TabPane>

              <TabPane tab={
                <span>
                  <CodeOutlined />
                  代码生成
                </span>
              } key="code">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">编程语言</label>
                    <Select
                      value={language}
                      onChange={setLanguage}
                      className="w-full"
                    >
                      <Option value="javascript">JavaScript</Option>
                      <Option value="python">Python</Option>
                      <Option value="java">Java</Option>
                      <Option value="cpp">C++</Option>
                    </Select>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium mb-2">功能描述</label>
                    <TextArea
                      value={codePrompt}
                      onChange={(e) => setCodePrompt(e.target.value)}
                      placeholder="请描述您想要实现的功能..."
                      rows={4}
                    />
                  </div>
                  
                  <Button
                    type="primary"
                    icon={<CodeOutlined />}
                    onClick={generateCode}
                    loading={loading}
                    block
                  >
                    生成代码
                  </Button>
                </div>
              </TabPane>

              <TabPane tab={
                <span>
                  <PictureOutlined />
                  设计生成
                </span>
              } key="design">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">设计需求</label>
                    <TextArea
                      value={designPrompt}
                      onChange={(e) => setDesignPrompt(e.target.value)}
                      placeholder="请描述您的设计需求..."
                      rows={4}
                    />
                  </div>
                  
                  <Button
                    type="primary"
                    icon={<PictureOutlined />}
                    onClick={generateDesign}
                    loading={loading}
                    block
                  >
                    生成设计方案
                  </Button>
                </div>
              </TabPane>
            </Tabs>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="生成结果" className="h-full">
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {results.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <FileTextOutlined className="text-4xl mb-4" />
                  <p>暂无生成结果</p>
                </div>
              ) : (
                results.map((result, index) => (
                  <Card key={index} size="small" className="mb-4">
                    <div className="flex justify-between items-start mb-2">
                      <span className="font-medium">
                        {result.type === 'text' && '文本生成'}
                        {result.type === 'code' && '代码生成'}
                        {result.type === 'design' && '设计方案'}
                      </span>
                      <Button
                        size="small"
                        icon={<DownloadOutlined />}
                        onClick={() => downloadResult(result, index)}
                      >
                        下载
                      </Button>
                    </div>
                    <Paragraph
                      ellipsis={{ rows: 6, expandable: true }}
                      className="mb-0"
                    >
                      <pre className="whitespace-pre-wrap text-sm">
                        {result.content}
                      </pre>
                    </Paragraph>
                  </Card>
                ))
              )}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default AIGeneration;