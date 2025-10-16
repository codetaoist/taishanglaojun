import React, { useState } from 'react';
import { Card, Button, Space, Typography, Alert, Divider } from 'antd';
import { testPermissionAPI } from '../../utils/testPermissionAPI';

const { Title, Text, Paragraph } = Typography;

const APITest: React.FC = () => {
  const [testResult, setTestResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const handleTestAPI = async () => {
    setLoading(true);
    try {
      const result = await testPermissionAPI();
      setTestResult(result);
      console.log('API测试结果:', result);
    } catch (error) {
      console.error('测试失败:', error);
      setTestResult({
        success: false,
        error: error instanceof Error ? error.message : '未知错误'
      });
    } finally {
      setLoading(false);
    }
  };

  const renderResult = () => {
    if (!testResult) return null;

    if (testResult.success) {
      return (
        <Alert
          type="success"
          message="API调用成功"
          description={
            <div>
              <p><strong>状态码:</strong> {testResult.status}</p>
              <p><strong>数据:</strong></p>
              <pre style={{ background: '#f5f5f5', padding: '10px', borderRadius: '4px', overflow: 'auto' }}>
                {JSON.stringify(testResult.data, null, 2)}
              </pre>
            </div>
          }
        />
      );
    } else {
      return (
        <Alert
          type="error"
          message="API调用失败"
          description={
            <div>
              <p><strong>错误:</strong> {testResult.error}</p>
              {testResult.status && <p><strong>状态码:</strong> {testResult.status}</p>}
              {testResult.data && (
                <>
                  <p><strong>响应数据:</strong></p>
                  <pre style={{ background: '#f5f5f5', padding: '10px', borderRadius: '4px', overflow: 'auto' }}>
                    {JSON.stringify(testResult.data, null, 2)}
                  </pre>
                </>
              )}
              {testResult.needsAuth && (
                <Alert
                  type="warning"
                  message="需要登录"
                  description="请先登录系统后再测试API"
                  style={{ marginTop: '10px' }}
                />
              )}
            </div>
          }
        />
      );
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Title level={2}>权限API测试</Title>
        <Paragraph>
          此页面用于测试权限API的调用，帮助诊断菜单管理页面的权限加载问题。
        </Paragraph>
        
        <Divider />
        
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button 
            type="primary" 
            onClick={handleTestAPI}
            loading={loading}
            size="large"
          >
            测试权限API
          </Button>
          
          {renderResult()}
          
          <Divider />
          
          <div>
            <Title level={4}>检查项目:</Title>
            <ul>
              <li>检查是否已登录（localStorage中是否有token）</li>
              <li>检查API基础URL配置</li>
              <li>测试权限API调用（/api/v1/permissions）</li>
              <li>检查响应数据格式</li>
              <li>分析错误原因（认证、权限、网络等）</li>
            </ul>
          </div>
          
          <Alert
            type="info"
            message="提示"
            description="请打开浏览器开发者工具的控制台查看详细的测试日志。"
          />
        </Space>
      </Card>
    </div>
  );
};

export default APITest;