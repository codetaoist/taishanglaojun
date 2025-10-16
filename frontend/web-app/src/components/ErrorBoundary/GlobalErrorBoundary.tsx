import React from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Result, Button, Typography } from 'antd';
import { ReloadOutlined, HomeOutlined } from '@ant-design/icons';

const { Paragraph, Text } = Typography;

interface ErrorFallbackProps {
  error: Error;
  resetErrorBoundary: () => void;
}

const ErrorFallback: React.FC<ErrorFallbackProps> = ({ error, resetErrorBoundary }) => {
  const handleGoHome = () => {
    window.location.href = '/';
  };

  const handleReload = () => {
    window.location.reload();
  };

  return (
    <div style={{ 
      minHeight: '100vh', 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'center',
      padding: '20px'
    }}>
      <Result
        status="error"
        title="页面出现错误"
        subTitle="抱歉，页面遇到了一些问题。请尝试刷新页面或返回首页。"
        extra={[
          <Button 
            type="primary" 
            icon={<ReloadOutlined />} 
            onClick={resetErrorBoundary}
            key="retry"
          >
            重试
          </Button>,
          <Button 
            icon={<ReloadOutlined />} 
            onClick={handleReload}
            key="reload"
          >
            刷新页面
          </Button>,
          <Button 
            icon={<HomeOutlined />} 
            onClick={handleGoHome}
            key="home"
          >
            返回首页
          </Button>,
        ]}
      >
        <div style={{ textAlign: 'left', maxWidth: '600px' }}>
          <Paragraph>
            <Text strong>错误详情：</Text>
          </Paragraph>
          <Paragraph>
            <Text code>{error.message}</Text>
          </Paragraph>
          {process.env.NODE_ENV === 'development' && (
            <details style={{ marginTop: '16px' }}>
              <summary style={{ cursor: 'pointer', marginBottom: '8px' }}>
                <Text type="secondary">查看错误堆栈（开发模式）</Text>
              </summary>
              <pre style={{ 
                background: '#f5f5f5', 
                padding: '12px', 
                borderRadius: '4px',
                fontSize: '12px',
                overflow: 'auto',
                maxHeight: '200px'
              }}>
                {error.stack}
              </pre>
            </details>
          )}
        </div>
      </Result>
    </div>
  );
};

interface GlobalErrorBoundaryProps {
  children: React.ReactNode;
}

const GlobalErrorBoundary: React.FC<GlobalErrorBoundaryProps> = ({ children }) => {
  const handleError = (error: Error, errorInfo: { componentStack: string }) => {
    // 在生产环境中，可以将错误发送到错误监控服务
    console.error('Global Error Boundary caught an error:', error, errorInfo);
    
    // 这里可以集成错误监控服务，如 Sentry
    // if (process.env.NODE_ENV === 'production') {
    //   Sentry.captureException(error, {
    //     contexts: {
    //       react: {
    //         componentStack: errorInfo.componentStack,
    //       },
    //     },
    //   });
    // }
  };

  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={handleError}
      onReset={() => {
        // 清除可能的错误状态
        window.location.reload();
      }}
    >
      {children}
    </ErrorBoundary>
  );
};

export default GlobalErrorBoundary;