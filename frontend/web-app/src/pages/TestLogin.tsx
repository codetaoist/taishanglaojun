import React, { useState } from 'react';
import { Button, Input, Card, Typography, Space, Alert } from 'antd';
import { useAuth } from '../hooks/useAuth';

const { Title } = Typography;

const TestLogin: React.FC = () => {
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('admin123');
  const [testResult, setTestResult] = useState<string>('');
  const [networkLogs, setNetworkLogs] = useState<string[]>([]);
  
  const { login, user, isAuthenticated, isLoading, checkAuthStatus } = useAuth();

  // 监听网络请求
  const logNetworkRequest = (message: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setNetworkLogs(prev => [...prev, `[${timestamp}] ${message}`]);
  };

  const handleTestLogin = async () => {
    setTestResult('');
    setNetworkLogs([]);
    
    try {
      logNetworkRequest('开始登录测试...');
      
      const result = await login(email, password);
      
      if (result.success) {
        logNetworkRequest('登录成功');
        setTestResult('登录成功！');
      } else {
        logNetworkRequest(`登录失败: ${result.error}`);
        setTestResult(`登录失败: ${result.error}`);
      }
    } catch (error: any) {
      logNetworkRequest(`登录异常: ${error.message}`);
      setTestResult(`登录异常: ${error.message}`);
    }
  };

  const handleTestAuth = async () => {
    setNetworkLogs([]);
    logNetworkRequest('开始认证状态检查...');
    
    try {
      await checkAuthStatus();
      logNetworkRequest('认证状态检查完成');
    } catch (error: any) {
      logNetworkRequest(`认证状态检查失败: ${error.message}`);
    }
  };

  const clearLogs = () => {
    setNetworkLogs([]);
    setTestResult('');
  };

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <Title level={2}>登录功能测试页面</Title>
      
      <Card title="当前认证状态" style={{ marginBottom: '20px' }}>
        <Space direction="vertical">
          <span className="text-gray-600">认证状态: {isAuthenticated ? '已认证' : '未认证'}</span>
          <span className="text-gray-600">加载状态: {isLoading ? '加载中' : '已完成'}</span>
          <span className="text-gray-600">用户信息: {user ? `${user.username} (${user.email})` : '无'}</span>
        </Space>
      </Card>

      <Card title="登录测试" style={{ marginBottom: '20px' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input
            placeholder="邮箱"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <Input.Password
            placeholder="密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <Space>
            <Button type="primary" onClick={handleTestLogin} loading={isLoading}>
              测试登录
            </Button>
            <Button onClick={handleTestAuth}>
              测试认证状态检查
            </Button>
            <Button onClick={clearLogs}>
              清除日志
            </Button>
          </Space>
          {testResult && (
            <Alert
              message={testResult}
              type={testResult.includes('成功') ? 'success' : 'error'}
            />
          )}
        </Space>
      </Card>

      <Card title="网络请求日志">
        <div style={{ 
          backgroundColor: '#f5f5f5', 
          padding: '10px', 
          borderRadius: '4px',
          maxHeight: '300px',
          overflowY: 'auto',
          fontFamily: 'monospace',
          fontSize: '12px'
        }}>
          {networkLogs.length === 0 ? (
            <span className="text-gray-500">暂无日志</span>
          ) : (
            networkLogs.map((log, index) => (
              <div key={index}>{log}</div>
            ))
          )}
        </div>
      </Card>
    </div>
  );
};

export default TestLogin;