import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Result, Button, Typography, Card, Space, Collapse } from 'antd';
import { getNotificationInstance } from '../../services/notificationService';
import { 
  BugOutlined, 
  ReloadOutlined, 
  HomeOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  CopyOutlined
} from '@ant-design/icons';

const { Text, Paragraph } = Typography;
const { Panel } = Collapse;

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  errorId: string;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    };
  }

  public static getDerivedStateFromError(error: Error): Partial<State> {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return {
      hasError: true,
      error,
      errorId: `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 记录错误信息
    this.setState({
      error,
      errorInfo
    });

    // 调用外部错误处理函数
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // 发送错误报告到监控服务
    this.reportError(error, errorInfo);
  }

  private reportError = (error: Error, errorInfo: ErrorInfo) => {
    // 这里可以集成错误监控服务，如 Sentry
    const errorReport = {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      errorId: this.state.errorId,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
      userId: localStorage.getItem('userId') || 'anonymous'
    };

    console.error('Error Boundary caught an error:', errorReport);

    // 可以发送到后端API
    // apiManager.reportError(errorReport);
  };

  private handleReload = () => {
    window.location.reload();
  };

  private handleGoHome = () => {
    window.location.href = '/';
  };

  private handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    });
  };

  private copyErrorInfo = () => {
    const errorText = `
错误ID: ${this.state.errorId}
时间: ${new Date().toLocaleString()}
错误信息: ${this.state.error?.message}
错误堆栈: ${this.state.error?.stack}
组件堆栈: ${this.state.errorInfo?.componentStack}
用户代理: ${navigator.userAgent}
页面URL: ${window.location.href}
    `.trim();

    navigator.clipboard.writeText(errorText).then(() => {
      const notification = getNotificationInstance();
      notification.success({
        message: '复制成功',
        description: '错误信息已复制到剪贴板',
        duration: 2
      });
    }).catch(() => {
      const notification = getNotificationInstance();
      notification.error({
        message: '复制失败',
        description: '无法复制到剪贴板，请手动选择文本复制',
        duration: 3
      });
    });
  };

  public render() {
    if (this.state.hasError) {
      // 如果有自定义的fallback组件，使用它
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // 默认的错误UI
      return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
          <Card className="max-w-2xl w-full shadow-lg">
            <Result
              status="error"
              icon={<BugOutlined className="text-red-500" />}
              title="应用程序遇到了错误"
              subTitle={
                <Space direction="vertical" className="text-left">
                  <Text type="secondary">
                    很抱歉，应用程序遇到了意外错误。我们已经记录了这个问题，并会尽快修复。
                  </Text>
                  <Text code className="text-xs">
                    错误ID: {this.state.errorId}
                  </Text>
                </Space>
              }
              extra={
                <Space wrap>
                  <Button 
                    type="primary" 
                    icon={<ReloadOutlined />}
                    onClick={this.handleReset}
                  >
                    重试
                  </Button>
                  <Button 
                    icon={<ReloadOutlined />}
                    onClick={this.handleReload}
                  >
                    刷新页面
                  </Button>
                  <Button 
                    icon={<HomeOutlined />}
                    onClick={this.handleGoHome}
                  >
                    返回首页
                  </Button>
                </Space>
              }
            />

            {/* 错误详情（开发环境或调试模式下显示） */}
            {(process.env.NODE_ENV === 'development' || window.location.search.includes('debug=true')) && (
              <div className="mt-6">
                <Collapse ghost>
                  <Panel 
                    header={
                      <Space>
                        <InfoCircleOutlined />
                        <span>错误详情（开发模式）</span>
                      </Space>
                    } 
                    key="error-details"
                  >
                    <Space direction="vertical" className="w-full">
                      <div>
                        <Text strong>错误信息:</Text>
                        <Paragraph 
                          code 
                          className="bg-red-50 p-3 rounded border border-red-200 mt-2"
                        >
                          {this.state.error?.message}
                        </Paragraph>
                      </div>

                      {this.state.error?.stack && (
                        <div>
                          <Text strong>错误堆栈:</Text>
                          <Paragraph 
                            code 
                            className="bg-gray-50 p-3 rounded border mt-2 text-xs max-h-40 overflow-auto"
                          >
                            {this.state.error.stack}
                          </Paragraph>
                        </div>
                      )}

                      {this.state.errorInfo?.componentStack && (
                        <div>
                          <Text strong>组件堆栈:</Text>
                          <Paragraph 
                            code 
                            className="bg-gray-50 p-3 rounded border mt-2 text-xs max-h-40 overflow-auto"
                          >
                            {this.state.errorInfo.componentStack}
                          </Paragraph>
                        </div>
                      )}

                      <Button 
                        size="small" 
                        icon={<CopyOutlined />}
                        onClick={this.copyErrorInfo}
                        className="mt-2"
                      >
                        复制错误信息
                      </Button>
                    </Space>
                  </Panel>
                </Collapse>
              </div>
            )}

            {/* 用户反馈区域 */}
            <div className="mt-6 p-4 bg-blue-50 rounded border border-blue-200">
              <Space>
                <WarningOutlined className="text-blue-500" />
                <div>
                  <Text strong className="text-blue-700">遇到问题？</Text>
                  <br />
                  <Text className="text-blue-600 text-sm">
                    如果问题持续存在，请联系技术支持团队，并提供错误ID: {this.state.errorId}
                  </Text>
                </div>
              </Space>
            </div>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

// 高阶组件，用于包装其他组件
export const withErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<Props, 'children'>
) => {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  return WrappedComponent;
};

// 错误边界Hook（用于函数组件中的错误处理）
export const useErrorHandler = () => {
  const [error, setError] = React.useState<Error | null>(null);

  const resetError = React.useCallback(() => {
    setError(null);
  }, []);

  const captureError = React.useCallback((error: Error) => {
    setError(error);
  }, []);

  React.useEffect(() => {
    if (error) {
      throw error;
    }
  }, [error]);

  return { captureError, resetError };
};

export default ErrorBoundary;